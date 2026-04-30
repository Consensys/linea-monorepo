//go:build cuda

package plonk2

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	bn254crypto "github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

const (
	plonk2ECMulDefaultTargetDomain = 1 << 24
	plonk2ECMulCacheDir            = "tmp/plonk2_ecmul_cache"
)

// plonk2ECMulCircuit proves N independent BN254 ECMul checks in the selected
// outer curve and includes one explicit BSB22 commitment.
type plonk2ECMulCircuit struct {
	Points       []sw_emulated.AffinePoint[emulated.BN254Fp]
	Scalars      []emulated.Element[emulated.BN254Fr]
	Expected     []sw_emulated.AffinePoint[emulated.BN254Fp]
	CommitInputs []frontend.Variable
	N            int `gnark:"-"`
}

func (c *plonk2ECMulCircuit) Define(api frontend.API) error {
	if len(c.CommitInputs) > 0 {
		committer, ok := api.(frontend.Committer)
		if !ok {
			return fmt.Errorf("frontend does not implement Committer")
		}
		commitment, err := committer.Commit(c.CommitInputs...)
		if err != nil {
			return err
		}
		api.AssertIsEqual(api.Sub(commitment, commitment), 0)
	}

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

func BenchmarkPlonk2ECMulTargetDomainAllCurves_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	targetDomain := plonk2ECMulTargetDomain(b)
	for _, curve := range plonk2ECMulBenchCurves(b) {
		setup := newPlonk2ECMulSetup(b, curve.id, targetDomain)
		b.Run(
			fmt.Sprintf("%s/domain=%s/instances=%d/generic-gpu", curve.name, benchPlonkSizeLabel(targetDomain), setup.instances),
			func(b *testing.B) {
				setup.benchmark(b, dev)
			},
		)
	}
}

type plonk2ECMulSetup struct {
	curveID        ecc.ID
	ccs            constraint.ConstraintSystem
	pk             gnarkplonk.ProvingKey
	vk             gnarkplonk.VerifyingKey
	fullWitness    witness.Witness
	publicWitness  witness.Witness
	instances      int
	targetDomain   int
	actualDomain   int
	constraints    int
	commitments    int
	compileTime    time.Duration
	setupTime      time.Duration
	witnessTime    time.Duration
	cacheWasLoaded bool
}

func newPlonk2ECMulSetup(tb testing.TB, curveID ecc.ID, targetDomain int) *plonk2ECMulSetup {
	tb.Helper()
	require.True(tb, isPowerOfTwo(targetDomain), "target ECMul domain should be a power of two")

	instances := plonk2ECMulInstancesForTargetDomain(tb, curveID, targetDomain)
	compileStart := time.Now()
	ccs, cacheWasLoaded := loadOrCompilePlonk2ECMulCCS(tb, curveID, instances, targetDomain)
	compileTime := time.Since(compileStart)
	actualDomain := plonk2ECMulDomainSize(ccs)
	require.Equal(tb, targetDomain, actualDomain, "calibrated ECMul circuit should use target domain")
	commitments := plonkCommitmentCountForCCS(ccs)
	require.Positive(tb, commitments, "ECMul benchmark circuit should exercise BSB22 commitments")

	setupStart := time.Now()
	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "PlonK setup should succeed")
	setupTime := time.Since(setupStart)

	witnessStart := time.Now()
	fullWitness, err := frontend.NewWitness(
		newPlonk2ECMulAssignment(instances),
		curveID.ScalarField(),
	)
	require.NoError(tb, err, "creating ECMul witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(tb, err, "extracting ECMul public witness should succeed")
	witnessTime := time.Since(witnessStart)

	return &plonk2ECMulSetup{
		curveID:        curveID,
		ccs:            ccs,
		pk:             pk,
		vk:             vk,
		fullWitness:    fullWitness,
		publicWitness:  publicWitness,
		instances:      instances,
		targetDomain:   targetDomain,
		actualDomain:   actualDomain,
		constraints:    ccs.GetNbConstraints(),
		commitments:    commitments,
		compileTime:    compileTime,
		setupTime:      setupTime,
		witnessTime:    witnessTime,
		cacheWasLoaded: cacheWasLoaded,
	}
}

func (s *plonk2ECMulSetup) benchmark(b *testing.B, dev *gpu.Device) {
	proverStart := time.Now()
	prover, err := NewProver(dev, s.ccs, s.pk, WithEnabled(true), WithStrictMode(true))
	require.NoError(b, err, "creating strict generic ECMul prover should succeed")
	defer prover.Close()
	proverPrepareTime := time.Since(proverStart)

	var proof gnarkplonk.Proof
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err = prover.Prove(s.fullWitness)
		require.NoError(b, err, "strict generic ECMul GPU proof should succeed")
		if proof == nil {
			b.Fatal("strict generic ECMul GPU proof returned nil proof")
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(s.instances), "ecmul_instances")
	b.ReportMetric(float64(s.constraints), "constraints")
	b.ReportMetric(float64(s.actualDomain), "domain")
	b.ReportMetric(float64(s.commitments), "bsb22_commitments")
	b.ReportMetric(s.compileTime.Seconds(), "compile_s")
	b.ReportMetric(s.setupTime.Seconds(), "setup_s")
	b.ReportMetric(s.witnessTime.Seconds(), "witness_s")
	b.ReportMetric(proverPrepareTime.Seconds(), "gpu_prepare_s")
	if s.cacheWasLoaded {
		b.ReportMetric(1, "ccs_cache_hit")
	} else {
		b.ReportMetric(0, "ccs_cache_hit")
	}
	require.NoError(b, gnarkplonk.Verify(proof, s.vk, s.publicWitness), "ECMul GPU proof should verify")
}

func plonk2ECMulTargetDomain(tb testing.TB) int {
	tb.Helper()
	raw := os.Getenv("PLONK2_ECMUL_TARGET_DOMAIN")
	if raw == "" {
		return plonk2ECMulDefaultTargetDomain
	}
	target, err := parsePlonkBenchConstraintCount(raw)
	require.NoError(tb, err, "parsing PLONK2_ECMUL_TARGET_DOMAIN should succeed")
	require.True(tb, isPowerOfTwo(target), "PLONK2_ECMUL_TARGET_DOMAIN should be a power of two")
	return target
}

func plonk2ECMulBenchCurves(tb testing.TB) []benchPlonkCurve {
	tb.Helper()
	raw := os.Getenv("PLONK2_ECMUL_CURVES")
	if raw == "" || strings.EqualFold(raw, "all") {
		return benchPlonkCurves
	}
	wanted := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		wanted[strings.ToLower(name)] = struct{}{}
	}
	var out []benchPlonkCurve
	for _, curve := range benchPlonkCurves {
		if _, ok := wanted[strings.ToLower(curve.name)]; ok {
			out = append(out, curve)
		}
	}
	require.NotEmpty(tb, out, "PLONK2_ECMUL_CURVES should select at least one supported curve")
	return out
}

func plonk2ECMulInstancesForTargetDomain(tb testing.TB, curveID ecc.ID, targetDomain int) int {
	tb.Helper()
	if raw := os.Getenv("PLONK2_ECMUL_INSTANCES"); raw != "" {
		instances, err := parsePlonkBenchConstraintCount(raw)
		require.NoError(tb, err, "parsing PLONK2_ECMUL_INSTANCES should succeed")
		require.Positive(tb, instances, "PLONK2_ECMUL_INSTANCES should be positive")
		return instances
	}
	size1 := plonk2ECMulCalibratedSize(tb, curveID, 1)
	size2 := plonk2ECMulCalibratedSize(tb, curveID, 2)
	perInstance := size2 - size1
	require.Positive(tb, perInstance, "ECMul system size should grow with instances")
	base := size1 - perInstance
	instances := (targetDomain - base) / perInstance
	require.Positive(tb, instances, "target domain should fit at least one ECMul instance")

	for instances > 1 && plonk2ECMulDomainSize(mustCompilePlonk2ECMulCCS(tb, curveID, instances)) > targetDomain {
		instances--
	}
	for {
		nextDomain := plonk2ECMulDomainSize(mustCompilePlonk2ECMulCCS(tb, curveID, instances+1))
		if nextDomain > targetDomain {
			break
		}
		instances++
	}
	return instances
}

func plonk2ECMulCalibratedSize(tb testing.TB, curveID ecc.ID, instances int) int {
	tb.Helper()
	return plonk2ECMulSystemSize(mustCompilePlonk2ECMulCCS(tb, curveID, instances))
}

func loadOrCompilePlonk2ECMulCCS(
	tb testing.TB,
	curveID ecc.ID,
	instances, targetDomain int,
) (constraint.ConstraintSystem, bool) {
	tb.Helper()
	path := plonk2ECMulCCSCachePath(curveID, instances, targetDomain)
	ccs, err := loadCachedPlonk2ECMulCCS(curveID, path)
	if err == nil {
		return ccs, true
	}
	_ = os.Remove(path)

	ccs = mustCompilePlonk2ECMulCCS(tb, curveID, instances)
	require.NoError(tb, saveCachedPlonk2ECMulCCS(path, ccs), "saving ECMul compiled circuit cache should succeed")
	return ccs, false
}

func mustCompilePlonk2ECMulCCS(tb testing.TB, curveID ecc.ID, instances int) constraint.ConstraintSystem {
	tb.Helper()
	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder[constraint.U64],
		newPlonk2ECMulCircuit(instances),
	)
	require.NoError(tb, err, "compiling ECMul benchmark circuit should succeed")
	return ccs
}

func newPlonk2ECMulCircuit(instances int) *plonk2ECMulCircuit {
	return &plonk2ECMulCircuit{
		Points:       make([]sw_emulated.AffinePoint[emulated.BN254Fp], instances),
		Scalars:      make([]emulated.Element[emulated.BN254Fr], instances),
		Expected:     make([]sw_emulated.AffinePoint[emulated.BN254Fp], instances),
		CommitInputs: make([]frontend.Variable, 2),
		N:            instances,
	}
}

func newPlonk2ECMulAssignment(instances int) *plonk2ECMulCircuit {
	raw := computePlonk2ECMulWitnessRaw(instances)
	points := make([]sw_emulated.AffinePoint[emulated.BN254Fp], instances)
	scalars := make([]emulated.Element[emulated.BN254Fr], instances)
	expected := make([]sw_emulated.AffinePoint[emulated.BN254Fp], instances)
	for i := range instances {
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
	return &plonk2ECMulCircuit{
		Points:       points,
		Scalars:      scalars,
		Expected:     expected,
		CommitInputs: []frontend.Variable{17, 19},
		N:            instances,
	}
}

type plonk2ECMulWitnessRaw struct {
	Points   []bn254crypto.G1Affine
	Scalars  []bn254fr.Element
	Expected []bn254crypto.G1Affine
}

func computePlonk2ECMulWitnessRaw(instances int) *plonk2ECMulWitnessRaw {
	_, _, generator, _ := bn254crypto.Generators()
	points := make([]bn254crypto.G1Affine, instances)
	scalars := make([]bn254fr.Element, instances)
	expected := make([]bn254crypto.G1Affine, instances)
	for i := range instances {
		u := deterministicPlonk2ECMulScalar(0x42, uint64(i))
		v := deterministicPlonk2ECMulScalar(0x43, uint64(i))
		points[i].ScalarMultiplication(&generator, u.BigInt(new(big.Int)))
		expected[i].ScalarMultiplication(&points[i], v.BigInt(new(big.Int)))
		scalars[i] = v
	}
	return &plonk2ECMulWitnessRaw{Points: points, Scalars: scalars, Expected: expected}
}

func deterministicPlonk2ECMulScalar(seed, index uint64) bn254fr.Element {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[0:8], seed)
	binary.LittleEndian.PutUint64(buf[8:16], index)
	binary.LittleEndian.PutUint64(buf[16:24], seed^0xdeadbeefcafebabe)
	binary.LittleEndian.PutUint64(buf[24:32], index^0x0123456789abcdef)
	var e bn254fr.Element
	e.SetBytes(buf[:])
	return e
}

func plonk2ECMulCCSCachePath(curveID ecc.ID, instances, targetDomain int) string {
	return filepath.Join(
		plonk2ECMulCacheDir,
		fmt.Sprintf("ecmul_%s_domain_%d_n%d.spr.bin", curveID.String(), targetDomain, instances),
	)
}

func loadCachedPlonk2ECMulCCS(curveID ecc.ID, path string) (ccs constraint.ConstraintSystem, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defer func() {
		if r := recover(); r != nil {
			ccs = nil
			err = fmt.Errorf("read ECMul sparse R1CS cache: %v", r)
		}
	}()

	ccs, err = newPlonk2SparseR1CS(curveID)
	if err != nil {
		return nil, err
	}
	if _, err = ccs.ReadFrom(file); err != nil {
		return nil, err
	}
	return ccs, nil
}

func saveCachedPlonk2ECMulCCS(path string, ccs constraint.ConstraintSystem) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = ccs.WriteTo(file)
	return err
}

func newPlonk2SparseR1CS(curveID ecc.ID) (constraint.ConstraintSystem, error) {
	switch curveID {
	case ecc.BN254:
		return csbn254.NewSparseR1CS(0), nil
	case ecc.BLS12_377:
		return csbls12377.NewSparseR1CS(0), nil
	case ecc.BW6_761:
		return csbw6761.NewSparseR1CS(0), nil
	default:
		return nil, fmt.Errorf("unsupported ECMul benchmark curve %s", curveID)
	}
}

func plonk2ECMulDomainSize(ccs constraint.ConstraintSystem) int {
	size := plonk2ECMulSystemSize(ccs)
	domain := 1
	for domain < size {
		domain <<= 1
	}
	return domain
}

func plonk2ECMulSystemSize(ccs constraint.ConstraintSystem) int {
	return ccs.GetNbConstraints() + ccs.GetNbPublicVariables()
}
