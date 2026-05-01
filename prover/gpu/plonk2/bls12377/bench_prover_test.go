//go:build cuda

package bls12377_test

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
	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"
	kzgbls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	bn254crypto "github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	curplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	gnarkfrontend "github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bls12377"
	"github.com/stretchr/testify/require"
)

// ─── ECMul circuit ────────────────────────────────────────────────────────────

type ecMulCircuitBLS struct {
	Points   []sw_emulated.AffinePoint[emulated.BN254Fp]
	Scalars  []emulated.Element[emulated.BN254Fr]
	Expected []sw_emulated.AffinePoint[emulated.BN254Fp]
	N        int `gnark:"-"`
}

func (c *ecMulCircuitBLS) Define(api gnarkfrontend.API) error {
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

func makeECMulCircuitBLS(n int) *ecMulCircuitBLS {
	return &ecMulCircuitBLS{
		Points:   make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		Scalars:  make([]emulated.Element[emulated.BN254Fr], n),
		Expected: make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		N:        n,
	}
}

func deterministicScalarBLS(seed, index uint64) bn254fr.Element {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[0:8], seed)
	binary.LittleEndian.PutUint64(buf[8:16], index)
	binary.LittleEndian.PutUint64(buf[16:24], seed^0xdeadbeefcafebabe)
	binary.LittleEndian.PutUint64(buf[24:32], index^0x0123456789abcdef)
	var e bn254fr.Element
	e.SetBytes(buf[:])
	return e
}

type ecMulWitnessRawBLS struct {
	Points   []bn254crypto.G1Affine
	Scalars  []bn254fr.Element
	Expected []bn254crypto.G1Affine
}

func computeECMulWitnessRawBLS(n int) *ecMulWitnessRawBLS {
	_, _, G, _ := bn254crypto.Generators()
	pts := make([]bn254crypto.G1Affine, n)
	scsArr := make([]bn254fr.Element, n)
	exp := make([]bn254crypto.G1Affine, n)
	for i := range n {
		u := deterministicScalarBLS(0x42, uint64(i))
		v := deterministicScalarBLS(0x43, uint64(i))
		pts[i].ScalarMultiplication(&G, u.BigInt(new(big.Int)))
		exp[i].ScalarMultiplication(&pts[i], v.BigInt(new(big.Int)))
		scsArr[i] = v
	}
	return &ecMulWitnessRawBLS{Points: pts, Scalars: scsArr, Expected: exp}
}

func makeECMulWitnessBLS(n int) (*ecMulCircuitBLS, error) {
	raw := computeECMulWitnessRawBLS(n)
	points := make([]sw_emulated.AffinePoint[emulated.BN254Fp], n)
	scalars := make([]emulated.Element[emulated.BN254Fr], n)
	expected := make([]sw_emulated.AffinePoint[emulated.BN254Fp], n)
	for i := range n {
		points[i] = sw_emulated.AffinePoint[emulated.BN254Fp]{
			X: emulated.ValueOf[emulated.BN254Fp](raw.Points[i].X),
			Y: emulated.ValueOf[emulated.BN254Fp](raw.Points[i].Y),
		}
		scalars[i] = emulated.ValueOf[emulated.BN254Fr](raw.Scalars[i])
		expected[i] = sw_emulated.AffinePoint[emulated.BN254Fp]{
			X: emulated.ValueOf[emulated.BN254Fp](raw.Expected[i].X),
			Y: emulated.ValueOf[emulated.BN254Fp](raw.Expected[i].Y),
		}
	}
	return &ecMulCircuitBLS{Points: points, Scalars: scalars, Expected: expected, N: n}, nil
}

func buildECMulWitnessBLS(n int) (witness.Witness, error) {
	w, err := makeECMulWitnessBLS(n)
	if err != nil {
		return nil, err
	}
	return gnarkfrontend.NewWitness(w, ecc.BLS12_377.ScalarField())
}

// ─── Cache helpers ────────────────────────────────────────────────────────────

const plonk2CacheDir = "tmp/plonk2_ecmul_cache"

func plonk2BLSVKCachePath(n int) string {
	return filepath.Join(plonk2CacheDir, fmt.Sprintf("ecmul_bls12377_n%d.vk.bin", n))
}

func loadCachedSPRBLS(path string) (*cs.SparseR1CS, error) {
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

func saveCachedSPRBLS(path string, spr *cs.SparseR1CS) error {
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

func loadCachedVKBLS(path string) (*curplonk.VerifyingKey, error) {
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

func saveCachedVKBLS(path string, vk *curplonk.VerifyingKey) error {
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

const benchSRSRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"

var benchSRSStore *plonk.SRSStore

func getBenchSRSStore(b *testing.B) *plonk.SRSStore {
	b.Helper()
	if benchSRSStore != nil {
		return benchSRSStore
	}
	store, err := plonk.NewSRSStore(benchSRSRootDir)
	require.NoError(b, err, "NewSRSStore")
	benchSRSStore = store
	return store
}

// ─── Setup helper ─────────────────────────────────────────────────────────────

type blsBenchSetup struct {
	spr         *cs.SparseR1CS
	vk          *curplonk.VerifyingKey
	srsCanon    kzg.SRS
	srsLag      kzg.SRS
	g1Points    []curve.G1Affine
	constraints int
	domainSize  int
}

// blsBenchSetupCache holds per-N compiled setups (compiled once per process).
var blsBenchSetupCache sync.Map // map[int]*blsBenchSetup

func setupBLSBench(b *testing.B, n int) *blsBenchSetup {
	b.Helper()
	if v, ok := blsBenchSetupCache.Load(n); ok {
		return v.(*blsBenchSetup)
	}

	b.Logf("setting up BLS12-377 ECMul n=%d (first time, may take a while)…", n)
	ctx := context.Background()
	store := getBenchSRSStore(b)
	vkPath := plonk2BLSVKCachePath(n)

	// Always compile circuit fresh (SPR file cache has CBOR re-read bug in this gnark version)
	ccs, compileErr := gnarkfrontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, makeECMulCircuitBLS(n))
	require.NoError(b, compileErr, "compile circuit n=%d", n)
	spr := ccs.(*cs.SparseR1CS)

	sizeCanonical, sizeLagrange := gnarkplonk.SRSSize(spr)

	// Load SRS
	srsCanon, srsLag, err := store.GetSRSCPUForCurve(ctx, ecc.BLS12_377, sizeCanonical, sizeLagrange)
	require.NoError(b, err, "load BLS12-377 SRS n=%d", n)

	// Load or compute VK (VK cache is stable binary format, no CBOR issue)
	vk, err := loadCachedVKBLS(vkPath)
	if err != nil {
		_, vkIface, setupErr := gnarkplonk.Setup(spr, srsCanon, srsLag)
		require.NoError(b, setupErr, "plonk setup n=%d", n)
		vk = vkIface.(*curplonk.VerifyingKey)
		require.NoError(b, saveCachedVKBLS(vkPath, vk), "save vk cache")
	}

	g1Points := srsCanon.(*kzgbls12377.SRS).Pk.G1
	setup := &blsBenchSetup{
		spr:         spr,
		vk:          vk,
		srsCanon:    srsCanon,
		srsLag:      srsLag,
		g1Points:    g1Points,
		constraints: spr.GetNbConstraints(),
		domainSize:  int(vk.Size),
	}
	blsBenchSetupCache.Store(n, setup)
	return setup
}

// ─── New prover benchmarks ────────────────────────────────────────────────────

func BenchmarkNewProver_BLS12377_N1(b *testing.B)   { benchmarkNewProverBLS(b, 1) }
func BenchmarkNewProver_BLS12377_N10(b *testing.B)  { benchmarkNewProverBLS(b, 10) }
func BenchmarkNewProver_BLS12377_N30(b *testing.B)  { benchmarkNewProverBLS(b, 30) }
func BenchmarkNewProver_BLS12377_N121(b *testing.B) { benchmarkNewProverBLS(b, 121) }
func BenchmarkNewProver_BLS12377_N353(b *testing.B) { benchmarkNewProverBLS(b, 353) }
func BenchmarkNewProver_BLS12377_N750(b *testing.B) { benchmarkNewProverBLS(b, 750) }

func benchmarkNewProverBLS(b *testing.B, n int) {
	b.Helper()
	setup := setupBLSBench(b, n)

	gpk := bls12377.NewGPUProvingKey(setup.g1Points, setup.vk)
	defer gpk.Close()

	dev, err := gpu.New()
	require.NoError(b, err)
	defer dev.Close()

	// Warmup
	fullW, err := buildECMulWitnessBLS(n)
	require.NoError(b, err)
	_, err = bls12377.GPUProve(dev, gpk, setup.spr, fullW)
	require.NoError(b, err, "warmup prove")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, werr := buildECMulWitnessBLS(n)
		require.NoError(b, werr)
		pStart := time.Now()
		_, perr := bls12377.GPUProve(dev, gpk, setup.spr, w)
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

// ─── Legacy prover benchmarks ─────────────────────────────────────────────────

func BenchmarkLegacyProver_BLS12377_N1(b *testing.B)   { benchmarkLegacyProverBLS(b, 1) }
func BenchmarkLegacyProver_BLS12377_N10(b *testing.B)  { benchmarkLegacyProverBLS(b, 10) }
func BenchmarkLegacyProver_BLS12377_N30(b *testing.B)  { benchmarkLegacyProverBLS(b, 30) }
func BenchmarkLegacyProver_BLS12377_N121(b *testing.B) { benchmarkLegacyProverBLS(b, 121) }
func BenchmarkLegacyProver_BLS12377_N353(b *testing.B) { benchmarkLegacyProverBLS(b, 353) }
func BenchmarkLegacyProver_BLS12377_N750(b *testing.B) { benchmarkLegacyProverBLS(b, 750) }

func benchmarkLegacyProverBLS(b *testing.B, n int) {
	b.Helper()
	setup := setupBLSBench(b, n)

	// Convert G1Affine points to TE points for legacy prover
	tePoints := plonk.ConvertG1AffineToTE(setup.g1Points)
	gpk := plonk.NewGPUProvingKey(tePoints, setup.vk)
	defer gpk.Close()

	dev, err := gpu.New()
	require.NoError(b, err)
	defer dev.Close()

	// Warmup
	fullW, err := buildECMulWitnessBLS(n)
	require.NoError(b, err)
	_, err = plonk.GPUProve(dev, gpk, setup.spr, fullW)
	require.NoError(b, err, "warmup prove")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, werr := buildECMulWitnessBLS(n)
		require.NoError(b, werr)
		pStart := time.Now()
		_, perr := plonk.GPUProve(dev, gpk, setup.spr, w)
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
