//go:build cuda

package bw6761_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	bn254crypto "github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	curve "github.com/consensys/gnark-crypto/ecc/bw6-761"
	kzgbw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	curplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"
	cs "github.com/consensys/gnark/constraint/bw6-761"
	gnarkfrontend "github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bw6761"
	"github.com/stretchr/testify/require"
)

// ─── ECMul circuit ────────────────────────────────────────────────────────────

type ecMulCircuitBW struct {
	Points   []sw_emulated.AffinePoint[emulated.BN254Fp]
	Scalars  []emulated.Element[emulated.BN254Fr]
	Expected []sw_emulated.AffinePoint[emulated.BN254Fp]
	N        int `gnark:"-"`
}

func (c *ecMulCircuitBW) Define(api gnarkfrontend.API) error {
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

func makeECMulCircuitBW(n int) *ecMulCircuitBW {
	return &ecMulCircuitBW{
		Points:   make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		Scalars:  make([]emulated.Element[emulated.BN254Fr], n),
		Expected: make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		N:        n,
	}
}

func deterministicScalarBW(seed, index uint64) bn254fr.Element {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[0:8], seed)
	binary.LittleEndian.PutUint64(buf[8:16], index)
	binary.LittleEndian.PutUint64(buf[16:24], seed^0xdeadbeefcafebabe)
	binary.LittleEndian.PutUint64(buf[24:32], index^0x0123456789abcdef)
	var e bn254fr.Element
	e.SetBytes(buf[:])
	return e
}

func computeECMulWitnessRawBW(n int) ([]bn254crypto.G1Affine, []bn254fr.Element, []bn254crypto.G1Affine) {
	_, _, G, _ := bn254crypto.Generators()
	pts := make([]bn254crypto.G1Affine, n)
	scsArr := make([]bn254fr.Element, n)
	exp := make([]bn254crypto.G1Affine, n)
	for i := range n {
		u := deterministicScalarBW(0x42, uint64(i))
		v := deterministicScalarBW(0x43, uint64(i))
		pts[i].ScalarMultiplication(&G, u.BigInt(new(big.Int)))
		exp[i].ScalarMultiplication(&pts[i], v.BigInt(new(big.Int)))
		scsArr[i] = v
	}
	return pts, scsArr, exp
}

func buildECMulWitnessBW(n int) (witness.Witness, error) {
	pts, scsArr, exp := computeECMulWitnessRawBW(n)
	w := &ecMulCircuitBW{
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
	return gnarkfrontend.NewWitness(w, ecc.BW6_761.ScalarField())
}

// ─── Cache helpers ────────────────────────────────────────────────────────────

const plonk2BWCacheDir = "tmp/plonk2_ecmul_cache"

func plonk2BWVKCachePath(n int) string {
	return filepath.Join(plonk2BWCacheDir, fmt.Sprintf("ecmul_bw6761_n%d.vk.bin", n))
}

var bwBenchSetupCache sync.Map

func _unused_plonk2BWCachePaths(n int) (sprPath, vkPath string) {
	base := fmt.Sprintf("ecmul_bw6761_n%d", n)
	return filepath.Join(plonk2BWCacheDir, base+".spr.bin"),
		filepath.Join(plonk2BWCacheDir, base+".vk.bin")
}

func loadCachedSPRBW(path string) (*cs.SparseR1CS, error) {
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

func saveCachedSPRBW(path string, spr *cs.SparseR1CS) error {
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

func loadCachedVKBW(path string) (*curplonk.VerifyingKey, error) {
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

func saveCachedVKBW(path string, vk *curplonk.VerifyingKey) error {
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

const benchBWSRSRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"

var benchBWSRSStore *plonk.SRSStore

func getBenchBWSRSStore(b *testing.B) *plonk.SRSStore {
	b.Helper()
	if benchBWSRSStore != nil {
		return benchBWSRSStore
	}
	store, err := plonk.NewSRSStore(benchBWSRSRootDir)
	require.NoError(b, err, "NewSRSStore")
	benchBWSRSStore = store
	return store
}

// ─── Setup helper ─────────────────────────────────────────────────────────────

type bwBenchSetup struct {
	spr         *cs.SparseR1CS
	vk          *curplonk.VerifyingKey
	g1Points    []curve.G1Affine
	constraints int
	domainSize  int
}

func setupBWBench(b *testing.B, n int) *bwBenchSetup {
	b.Helper()
	if v, ok := bwBenchSetupCache.Load(n); ok {
		return v.(*bwBenchSetup)
	}
	b.Logf("setting up BW6-761 ECMul n=%d (first time)…", n)
	ctx := context.Background()
	store := getBenchBWSRSStore(b)
	vkPath := plonk2BWVKCachePath(n)

	ccs, compileErr := gnarkfrontend.Compile(ecc.BW6_761.ScalarField(), scs.NewBuilder, makeECMulCircuitBW(n))
	require.NoError(b, compileErr, "compile circuit n=%d", n)
	spr := ccs.(*cs.SparseR1CS)

	sizeCanonical, sizeLagrange := gnarkplonk.SRSSize(spr)

	srsCanon, srsLag, err := store.GetSRSCPUForCurve(ctx, ecc.BW6_761, sizeCanonical, sizeLagrange)
	require.NoError(b, err, "load BW6-761 SRS n=%d", n)

	vk, err := loadCachedVKBW(vkPath)
	if err != nil {
		_, vkIface, setupErr := gnarkplonk.Setup(spr, srsCanon, srsLag)
		require.NoError(b, setupErr, "plonk setup n=%d", n)
		vk = vkIface.(*curplonk.VerifyingKey)
		require.NoError(b, saveCachedVKBW(vkPath, vk), "save vk cache")
	}

	g1Points := srsCanon.(*kzgbw6761.SRS).Pk.G1
	setup := &bwBenchSetup{spr: spr, vk: vk, g1Points: g1Points,
		constraints: spr.GetNbConstraints(), domainSize: int(vk.Size)}
	bwBenchSetupCache.Store(n, setup)
	return setup
}

// ─── CPU setup helper ─────────────────────────────────────────────────────────

// bwCPUBenchSetup extends bwBenchSetup with the PLONK proving key, which is
// required for CPU-only proving but not for GPU proving.
type bwCPUBenchSetup struct {
	*bwBenchSetup
	pk gnarkplonk.ProvingKey
}

var bwCPUBenchSetupCache sync.Map

// setupBWCPUBench builds (and caches) the full PLONK setup including the
// proving key. The SRS is loaded fresh; setup is run once per n value.
func setupBWCPUBench(b *testing.B, n int) *bwCPUBenchSetup {
	b.Helper()
	if v, ok := bwCPUBenchSetupCache.Load(n); ok {
		return v.(*bwCPUBenchSetup)
	}
	base := setupBWBench(b, n)
	ctx := context.Background()
	store := getBenchBWSRSStore(b)
	sizeCanonical, sizeLagrange := gnarkplonk.SRSSize(base.spr)
	srsCanon, srsLag, err := store.GetSRSCPUForCurve(ctx, ecc.BW6_761, sizeCanonical, sizeLagrange)
	require.NoError(b, err, "load BW6-761 SRS for CPU bench n=%d", n)
	pkIface, _, err := gnarkplonk.Setup(base.spr, srsCanon, srsLag)
	require.NoError(b, err, "plonk setup (CPU) n=%d", n)
	setup := &bwCPUBenchSetup{bwBenchSetup: base, pk: pkIface}
	bwCPUBenchSetupCache.Store(n, setup)
	return setup
}

// ─── Benchmarks ───────────────────────────────────────────────────────────────

func BenchmarkNewProver_BW6761_N1(b *testing.B)   { benchmarkNewProverBW(b, 1) }
func BenchmarkNewProver_BW6761_N10(b *testing.B)  { benchmarkNewProverBW(b, 10) }
func BenchmarkNewProver_BW6761_N30(b *testing.B)  { benchmarkNewProverBW(b, 30) }
func BenchmarkNewProver_BW6761_N121(b *testing.B) { benchmarkNewProverBW(b, 121) }
func BenchmarkNewProver_BW6761_N353(b *testing.B) { benchmarkNewProverBW(b, 353) }

func BenchmarkCPUProver_BW6761_N1(b *testing.B)   { benchmarkCPUProverBW(b, 1) }
func BenchmarkCPUProver_BW6761_N10(b *testing.B)  { benchmarkCPUProverBW(b, 10) }
func BenchmarkCPUProver_BW6761_N30(b *testing.B)  { benchmarkCPUProverBW(b, 30) }
func BenchmarkCPUProver_BW6761_N121(b *testing.B) { benchmarkCPUProverBW(b, 121) }
func BenchmarkCPUProver_BW6761_N353(b *testing.B) { benchmarkCPUProverBW(b, 353) }

func benchmarkNewProverBW(b *testing.B, n int) {
	b.Helper()
	setup := setupBWBench(b, n)

	gpk := bw6761.NewGPUProvingKey(setup.g1Points, setup.vk)
	defer gpk.Close()

	dev, err := gpu.New()
	require.NoError(b, err)
	defer dev.Close()

	// Warmup
	fullW, err := buildECMulWitnessBW(n)
	require.NoError(b, err)
	_, err = bw6761.GPUProve(dev, gpk, setup.spr, fullW)
	require.NoError(b, err, "warmup prove")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, werr := buildECMulWitnessBW(n)
		require.NoError(b, werr)
		pStart := time.Now()
		_, perr := bw6761.GPUProve(dev, gpk, setup.spr, w)
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

func benchmarkCPUProverBW(b *testing.B, n int) {
	b.Helper()
	setup := setupBWCPUBench(b, n)

	// Warmup
	fullW, err := buildECMulWitnessBW(n)
	require.NoError(b, err)
	_, err = gnarkplonk.Prove(setup.spr, setup.pk, fullW)
	require.NoError(b, err, "warmup prove")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, werr := buildECMulWitnessBW(n)
		require.NoError(b, werr)
		pStart := time.Now()
		_, perr := gnarkplonk.Prove(setup.spr, setup.pk, w)
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
