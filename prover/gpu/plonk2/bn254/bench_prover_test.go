//go:build cuda

package bn254_test

import (
	"context"
	"encoding/binary"
	"sync"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bn254"
	kzgbn254 "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	curplonk "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/backend/witness"
	cs "github.com/consensys/gnark/constraint/bn254"
	gnarkfrontend "github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bn254"
	"github.com/stretchr/testify/require"
)

// ─── ECMul circuit ────────────────────────────────────────────────────────────

type ecMulCircuitBN struct {
	Points   []sw_emulated.AffinePoint[emulated.BN254Fp]
	Scalars  []emulated.Element[emulated.BN254Fr]
	Expected []sw_emulated.AffinePoint[emulated.BN254Fp]
	N        int `gnark:"-"`
}

func (c *ecMulCircuitBN) Define(api gnarkfrontend.API) error {
	cv, err := sw_emulated.New[emulated.BN254Fp, emulated.BN254Fr](api, sw_emulated.GetBN254Params())
	if err != nil {
		return err
	}
	for i := 0; i < c.N; i++ {
		res := cv.ScalarMul(&c.Points[i], &c.Scalars[i])
		cv.AssertIsEqual(res, &c.Expected[i])
	}
	return nil
}

func makeECMulCircuitBN(n int) *ecMulCircuitBN {
	return &ecMulCircuitBN{
		Points:   make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		Scalars:  make([]emulated.Element[emulated.BN254Fr], n),
		Expected: make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		N:        n,
	}
}

func deterministicScalarBN(seed, index uint64) bn254fr.Element {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[0:8], seed)
	binary.LittleEndian.PutUint64(buf[8:16], index)
	binary.LittleEndian.PutUint64(buf[16:24], seed^0xdeadbeefcafebabe)
	binary.LittleEndian.PutUint64(buf[24:32], index^0x0123456789abcdef)
	var e bn254fr.Element
	e.SetBytes(buf[:])
	return e
}

func computeECMulWitnessRawBN(n int) ([]curve.G1Affine, []bn254fr.Element, []curve.G1Affine) {
	_, _, G, _ := curve.Generators()
	pts := make([]curve.G1Affine, n)
	scsArr := make([]bn254fr.Element, n)
	exp := make([]curve.G1Affine, n)
	for i := range n {
		u := deterministicScalarBN(0x42, uint64(i))
		v := deterministicScalarBN(0x43, uint64(i))
		pts[i].ScalarMultiplication(&G, u.BigInt(new(big.Int)))
		exp[i].ScalarMultiplication(&pts[i], v.BigInt(new(big.Int)))
		scsArr[i] = v
	}
	return pts, scsArr, exp
}

func buildECMulWitnessBN(n int) (witness.Witness, error) {
	pts, scsArr, exp := computeECMulWitnessRawBN(n)
	w := &ecMulCircuitBN{
		Points:   make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		Scalars:  make([]emulated.Element[emulated.BN254Fr], n),
		Expected: make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		N:        n,
	}
	for i := range n {
		w.Points[i] = sw_emulated.AffinePoint[emulated.BN254Fp]{
			X: emulated.ValueOf[emulated.BN254Fp](pts[i].X),
			Y: emulated.ValueOf[emulated.BN254Fp](pts[i].Y),
		}
		w.Scalars[i] = emulated.ValueOf[emulated.BN254Fr](scsArr[i])
		w.Expected[i] = sw_emulated.AffinePoint[emulated.BN254Fp]{
			X: emulated.ValueOf[emulated.BN254Fp](exp[i].X),
			Y: emulated.ValueOf[emulated.BN254Fp](exp[i].Y),
		}
	}
	return gnarkfrontend.NewWitness(w, ecc.BN254.ScalarField())
}

// ─── Cache helpers ────────────────────────────────────────────────────────────

const plonk2BNCacheDir = "tmp/plonk2_ecmul_cache"

func plonk2BNVKCachePath(n int) string {
	return filepath.Join(plonk2BNCacheDir, fmt.Sprintf("ecmul_bn254_n%d.vk.bin", n))
}
var bnBenchSetupCache sync.Map

func loadCachedSPRBN(path string) (*cs.SparseR1CS, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	spr := cs.NewSparseR1CS(0)
	if _, err = spr.ReadFrom(f); err != nil {
		return nil, err
	}
	return spr, nil
}

func saveCachedSPRBN(path string, spr *cs.SparseR1CS) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = spr.WriteTo(f)
	return err
}

func loadCachedVKBN(path string) (*curplonk.VerifyingKey, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	vk := new(curplonk.VerifyingKey)
	if _, err = vk.ReadFrom(f); err != nil {
		return nil, err
	}
	return vk, nil
}

func saveCachedVKBN(path string, vk *curplonk.VerifyingKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = vk.WriteTo(f)
	return err
}

// ─── SRS store singleton ──────────────────────────────────────────────────────

const benchBNSRSRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"

var benchBNSRSStore *plonk.SRSStore

func getBenchBNSRSStore(b *testing.B) *plonk.SRSStore {
	b.Helper()
	if benchBNSRSStore != nil {
		return benchBNSRSStore
	}
	store, err := plonk.NewSRSStore(benchBNSRSRootDir)
	require.NoError(b, err, "NewSRSStore")
	benchBNSRSStore = store
	return store
}

// ─── Setup helper ─────────────────────────────────────────────────────────────

type bnBenchSetup struct {
	spr         *cs.SparseR1CS
	vk          *curplonk.VerifyingKey
	g1Points    []curve.G1Affine
	constraints int
	domainSize  int
}

func setupBNBench(b *testing.B, n int) *bnBenchSetup {
	b.Helper()
	if v, ok := bnBenchSetupCache.Load(n); ok {
		return v.(*bnBenchSetup)
	}
	b.Logf("setting up BN254 ECMul n=%d (first time, may take a while)…", n)
	ctx := context.Background()
	store := getBenchBNSRSStore(b)
	vkPath := plonk2BNVKCachePath(n)

	ccs, compileErr := gnarkfrontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, makeECMulCircuitBN(n))
	require.NoError(b, compileErr, "compile circuit n=%d", n)
	spr := ccs.(*cs.SparseR1CS)

	sizeCanonical, sizeLagrange := gnarkplonk.SRSSize(spr)

	srsCanon, srsLag, err := store.GetSRSCPUForCurve(ctx, ecc.BN254, sizeCanonical, sizeLagrange)
	require.NoError(b, err, "load BN254 SRS n=%d", n)

	vk, err := loadCachedVKBN(vkPath)
	if err != nil {
		_, vkIface, setupErr := gnarkplonk.Setup(spr, srsCanon, srsLag)
		require.NoError(b, setupErr, "plonk setup n=%d", n)
		vk = vkIface.(*curplonk.VerifyingKey)
		require.NoError(b, saveCachedVKBN(vkPath, vk), "save vk cache")
	}

	g1Points := srsCanon.(*kzgbn254.SRS).Pk.G1
	setup := &bnBenchSetup{spr: spr, vk: vk, g1Points: g1Points,
		constraints: spr.GetNbConstraints(), domainSize: int(vk.Size)}
	bnBenchSetupCache.Store(n, setup)
	return setup
}

// ─── Benchmarks ───────────────────────────────────────────────────────────────

func BenchmarkNewProver_BN254_N1(b *testing.B)   { benchmarkNewProverBN(b, 1) }
func BenchmarkNewProver_BN254_N10(b *testing.B)  { benchmarkNewProverBN(b, 10) }
func BenchmarkNewProver_BN254_N30(b *testing.B)  { benchmarkNewProverBN(b, 30) }
func BenchmarkNewProver_BN254_N121(b *testing.B) { benchmarkNewProverBN(b, 121) }
func BenchmarkNewProver_BN254_N353(b *testing.B) { benchmarkNewProverBN(b, 353) }
func BenchmarkNewProver_BN254_N750(b *testing.B) { benchmarkNewProverBN(b, 750) }

func benchmarkNewProverBN(b *testing.B, n int) {
	b.Helper()
	setup := setupBNBench(b, n)

	gpk := bn254.NewGPUProvingKey(setup.g1Points, setup.vk)
	defer gpk.Close()

	dev, err := gpu.New()
	require.NoError(b, err)
	defer dev.Close()

	// Warmup
	fullW, err := buildECMulWitnessBN(n)
	require.NoError(b, err)
	_, err = bn254.GPUProve(dev, gpk, setup.spr, fullW)
	require.NoError(b, err, "warmup prove")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, werr := buildECMulWitnessBN(n)
		require.NoError(b, werr)
		pStart := time.Now()
		_, perr := bn254.GPUProve(dev, gpk, setup.spr, w)
		totalProve += time.Since(pStart)
		require.NoError(b, perr, "prove")
	}
	b.StopTimer()

	b.ReportMetric(float64(setup.constraints), "constraints")
	b.ReportMetric(float64(setup.domainSize), "domain_size")
	if b.N > 0 {
		b.ReportMetric(float64(totalProve.Milliseconds())/float64(b.N), "prove_ms")
	}
}
