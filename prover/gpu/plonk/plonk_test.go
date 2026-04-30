//go:build cuda

package plonk_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bn254crypto "github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	plonk377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	gnarkcs "github.com/consensys/gnark/constraint/bls12-377"
	gnarkfrontend "github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// ecMulCircuit proves N independent BN254 ECMul checks inside one PLONK circuit.
type ecMulCircuit struct {
	Points   []sw_emulated.AffinePoint[emulated.BN254Fp]
	Scalars  []emulated.Element[emulated.BN254Fr]
	Expected []sw_emulated.AffinePoint[emulated.BN254Fp]
	N        int `gnark:"-"`
}

func (c *ecMulCircuit) Define(api gnarkfrontend.API) error {
	curve, err := sw_emulated.New[emulated.BN254Fp, emulated.BN254Fr](api, sw_emulated.GetBN254Params())
	if err != nil {
		return err
	}
	for i := 0; i < c.N; i++ {
		res := curve.ScalarMul(&c.Points[i], &c.Scalars[i])
		curve.AssertIsEqual(res, &c.Expected[i])
	}
	return nil
}

func makeECMulCircuit(n int) *ecMulCircuit {
	return &ecMulCircuit{
		Points:   make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		Scalars:  make([]emulated.Element[emulated.BN254Fr], n),
		Expected: make([]sw_emulated.AffinePoint[emulated.BN254Fp], n),
		N:        n,
	}
}

func deterministicScalar(seed, index uint64) bn254fr.Element {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[0:8], seed)
	binary.LittleEndian.PutUint64(buf[8:16], index)
	binary.LittleEndian.PutUint64(buf[16:24], seed^0xdeadbeefcafebabe)
	binary.LittleEndian.PutUint64(buf[24:32], index^0x0123456789abcdef)
	var e bn254fr.Element
	e.SetBytes(buf[:])
	return e
}

type ecMulWitnessRaw struct {
	Points   []bn254crypto.G1Affine
	Scalars  []bn254fr.Element
	Expected []bn254crypto.G1Affine
}

func computeECMulWitnessRaw(n int) *ecMulWitnessRaw {
	_, _, G, _ := bn254crypto.Generators()
	pts := make([]bn254crypto.G1Affine, n)
	scs := make([]bn254fr.Element, n)
	exp := make([]bn254crypto.G1Affine, n)
	for i := range n {
		u := deterministicScalar(0x42, uint64(i))
		v := deterministicScalar(0x43, uint64(i))
		pts[i].ScalarMultiplication(&G, u.BigInt(new(big.Int)))
		exp[i].ScalarMultiplication(&pts[i], v.BigInt(new(big.Int)))
		scs[i] = v
	}
	return &ecMulWitnessRaw{Points: pts, Scalars: scs, Expected: exp}
}

func makeECMulWitness(t testing.TB, n int) *ecMulCircuit {
	t.Helper()
	raw := computeECMulWitnessRaw(n)
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
	return &ecMulCircuit{Points: points, Scalars: scalars, Expected: expected, N: n}
}

func buildECMulWitness(t testing.TB, n int) (witness.Witness, fr.Vector) {
	t.Helper()
	w := makeECMulWitness(t, n)
	full, err := gnarkfrontend.NewWitness(w, ecc.BLS12_377.ScalarField())
	require.NoError(t, err, "witness")
	pub, err := full.Public()
	require.NoError(t, err, "public witness")
	return full, pub.Vector().(fr.Vector)
}

func buildInvalidECMulWitness(t testing.TB, n int) witness.Witness {
	t.Helper()
	w := makeECMulWitness(t, n)
	w.Expected[0] = w.Points[0]
	full, err := gnarkfrontend.NewWitness(w, ecc.BLS12_377.ScalarField())
	require.NoError(t, err, "invalid witness")
	return full
}

type plonkECMulSetup struct {
	spr          *gnarkcs.SparseR1CS
	vk           *plonk377.VerifyingKey
	gpk          *plonk.GPUProvingKey
	constraints  int
	compileTime  time.Duration
	setupTime    time.Duration
	prepareGPKAt time.Duration
}

const plonkCacheDir = "tmp/plonk_cache"

func plonkECMulCachePaths(n int) (sprPath, vkPath string) {
	base := fmt.Sprintf("ecmul_bls12_377_n%d", n)
	return filepath.Join(plonkCacheDir, base+".spr.bin"), filepath.Join(plonkCacheDir, base+".vk.bin")
}

func loadCachedSparseR1CS(path string) (spr *gnarkcs.SparseR1CS, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	defer func() {
		if r := recover(); r != nil {
			spr = nil
			err = fmt.Errorf("read sparse r1cs cache: %v", r)
		}
	}()

	spr = gnarkcs.NewSparseR1CS(0)
	if _, err = spr.ReadFrom(f); err != nil {
		return nil, err
	}
	return spr, nil
}

func saveCachedSparseR1CS(path string, spr *gnarkcs.SparseR1CS) error {
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

func loadCachedVK(path string) (vk *plonk377.VerifyingKey, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	defer func() {
		if r := recover(); r != nil {
			vk = nil
			err = fmt.Errorf("read verifying key cache: %v", r)
		}
	}()

	vk = new(plonk377.VerifyingKey)
	if _, err = vk.ReadFrom(f); err != nil {
		return nil, err
	}
	return vk, nil
}

func saveCachedVK(path string, vk *plonk377.VerifyingKey) error {
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

func setupPlonkECMul(t testing.TB, n int) *plonkECMulSetup {
	t.Helper()
	require := require.New(t)
	ctx := context.Background()
	store := getTestSRSStore(t)
	sprCachePath, vkCachePath := plonkECMulCachePaths(n)

	compileStart := time.Now()
	spr, err := loadCachedSparseR1CS(sprCachePath)
	if err != nil {
		ccs, compileErr := gnarkfrontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder[constraint.U64], makeECMulCircuit(n))
		require.NoError(compileErr, "compile")
		spr = ccs.(*gnarkcs.SparseR1CS)
		require.NoError(saveCachedSparseR1CS(sprCachePath, spr), "save compiled circuit cache")
	}
	compileTime := time.Since(compileStart)

	setupStart := time.Now()
	vk, err := loadCachedVK(vkCachePath)
	if err != nil {
		srsI, srsLagI, srsErr := store.GetSRSCPU(ctx, spr)
		require.NoError(srsErr, "load SRS from store")
		_, vkI, setupErr := gnarkplonk.Setup(spr, srsI, srsLagI)
		require.NoError(setupErr, "plonk setup")
		vk = vkI.(*plonk377.VerifyingKey)
		require.NoError(saveCachedVK(vkCachePath, vk), "save vk cache")
	}
	setupTime := time.Since(setupStart)

	gpkStart := time.Now()
	canonicalTE, err := store.GetSRSGPU(ctx, spr)
	require.NoError(err, "load gpu SRS from store")
	gpk := plonk.NewGPUProvingKey(canonicalTE, vk)
	prepareTime := time.Since(gpkStart)

	return &plonkECMulSetup{
		spr:          spr,
		vk:           vk,
		gpk:          gpk,
		constraints:  spr.GetNbConstraints(),
		compileTime:  compileTime,
		setupTime:    setupTime,
		prepareGPKAt: prepareTime,
	}
}

func TestPlonkECMulBasicValidation(t *testing.T) {
	require := require.New(t)

	const nInstances = 1
	setup := setupPlonkECMul(t, nInstances)
	defer setup.gpk.Close()

	dev, err := gpu.New()
	require.NoError(err)
	defer dev.Close()

	require.NoError(setup.gpk.Prepare(dev, setup.spr), "prepare gpk")

	full, pubVec := buildECMulWitness(t, nInstances)
	proof, err := plonk.GPUProve(dev, setup.gpk, setup.spr, full)
	require.NoError(err, "GPU prove")
	require.NoError(plonk377.Verify(proof, setup.vk, pubVec), "verify")
}

func TestPlonkECMulLazyPrepareValidation(t *testing.T) {
	require := require.New(t)

	const nInstances = 1
	setup := setupPlonkECMul(t, nInstances)
	defer setup.gpk.Close()

	dev, err := gpu.New()
	require.NoError(err)
	defer dev.Close()

	full, pubVec := buildECMulWitness(t, nInstances)
	proof, err := plonk.GPUProve(dev, setup.gpk, setup.spr, full)
	require.NoError(err, "GPU prove")
	require.NoError(plonk377.Verify(proof, setup.vk, pubVec), "verify")
}

func TestPlonkECMulNegativeCases(t *testing.T) {
	require := require.New(t)

	const nInstances = 1
	setup := setupPlonkECMul(t, nInstances)
	defer setup.gpk.Close()

	dev, err := gpu.New()
	require.NoError(err)
	defer dev.Close()

	require.NoError(setup.gpk.Prepare(dev, setup.spr), "prepare gpk")

	full, pubVec := buildECMulWitness(t, nInstances)
	proof, err := plonk.GPUProve(dev, setup.gpk, setup.spr, full)
	require.NoError(err, "GPU prove")

	proof.LRO[0].X.SetZero()
	require.Error(plonk377.Verify(proof, setup.vk, pubVec), "tampered proof should not verify")

	invalidWitness := buildInvalidECMulWitness(t, nInstances)
	_, err = plonk.GPUProve(dev, setup.gpk, setup.spr, invalidWitness)
	require.Error(err, "invalid witness should not produce a proof")
}

func TestPlonkECMulConcurrentProofsSameKey(t *testing.T) {
	require := require.New(t)

	const nInstances = 1
	setup := setupPlonkECMul(t, nInstances)
	defer setup.gpk.Close()

	dev, err := gpu.New()
	require.NoError(err)
	defer dev.Close()

	require.NoError(setup.gpk.Prepare(dev, setup.spr), "prepare gpk")

	var g errgroup.Group
	for range 2 {
		g.Go(func() error {
			full, pubVec := buildECMulWitness(t, nInstances)
			proof, err := plonk.GPUProve(dev, setup.gpk, setup.spr, full)
			if err != nil {
				return err
			}
			return plonk377.Verify(proof, setup.vk, pubVec)
		})
	}

	require.NoError(g.Wait(), "concurrent proofs")
}

func BenchmarkPlonkECMul10(b *testing.B)  { benchmarkPlonkECMul(b, 10) }
func BenchmarkPlonkECMul30(b *testing.B)  { benchmarkPlonkECMul(b, 30) }
func BenchmarkPlonkECMul121(b *testing.B) { benchmarkPlonkECMul(b, 121) }
func BenchmarkPlonkECMul353(b *testing.B) { benchmarkPlonkECMul(b, 353) }
func BenchmarkPlonkECMul750(b *testing.B) { benchmarkPlonkECMul(b, 750) }

func benchmarkPlonkECMul(b *testing.B, nInstances int) {
	if testing.Short() && nInstances > 10 {
		b.Skip("skipping medium/large benchmark in short mode")
	}

	require := require.New(b)
	setup := setupPlonkECMul(b, nInstances)
	defer setup.gpk.Close()

	dev, err := gpu.New()
	require.NoError(err)
	defer dev.Close()
	prepareGPKStart := time.Now()
	require.NoError(setup.gpk.Prepare(dev, setup.spr), "prepare gpk")
	setup.prepareGPKAt = time.Since(prepareGPKStart)

	fullWarmup, pubWarmup := buildECMulWitness(b, nInstances)
	firstProveStart := time.Now()
	proofWarmup, err := plonk.GPUProve(dev, setup.gpk, setup.spr, fullWarmup)
	firstProveTime := time.Since(firstProveStart)
	require.NoError(err, "warmup prove")
	require.NoError(plonk377.Verify(proofWarmup, setup.vk, pubWarmup), "warmup verify")

	var totalProve time.Duration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		full, pubVec := buildECMulWitness(b, nInstances)
		pStart := time.Now()
		proof, err := plonk.GPUProve(dev, setup.gpk, setup.spr, full)
		totalProve += time.Since(pStart)
		require.NoError(err, "prove")
		require.NoError(plonk377.Verify(proof, setup.vk, pubVec), "verify")
	}
	b.StopTimer()

	b.ReportMetric(float64(setup.constraints), "constraints")
	b.ReportMetric(float64(setup.gpk.Size()), "domain")
	b.ReportMetric(setup.compileTime.Seconds(), "compile_s")
	b.ReportMetric(setup.setupTime.Seconds(), "setup_s")
	b.ReportMetric(setup.prepareGPKAt.Seconds(), "gpu_pk_s")
	b.ReportMetric(firstProveTime.Seconds(), "first_prove_s")

	if b.N > 0 {
		b.ReportMetric(float64(totalProve.Nanoseconds())/float64(b.N), "prove_ns/op")
	}
}
