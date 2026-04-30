//go:build cuda

package plonk2

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	blscs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/gpu"
	oldplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

func BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating GPU device should succeed")
	defer dev.Close()

	for _, constraints := range plonkBenchConstraintCounts(b) {
		setup := newFullProverBLS12377Setup(b, constraints)
		defer setup.gpk.Close()

		b.Run(fmt.Sprintf("bls12-377/constraints=%d/cpu", constraints), func(b *testing.B) {
			setup.benchmarkCPU(b)
		})
		b.Run(fmt.Sprintf("bls12-377/constraints=%d/current-gpu", constraints), func(b *testing.B) {
			setup.benchmarkCurrentGPU(b, dev)
		})
	}
}

func BenchmarkPlonk2EnabledFullProverBLS12377_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating GPU device should succeed")
	defer dev.Close()

	for _, constraints := range plonkBenchConstraintCounts(b) {
		setup := newFullProverBLS12377Setup(b, constraints)
		defer setup.gpk.Close()

		b.Run(fmt.Sprintf("bls12-377/constraints=%d/plonk2-enabled", constraints), func(b *testing.B) {
			prover, err := NewProver(
				dev,
				setup.spr,
				setup.pk,
				WithEnabled(true),
				WithStrictMode(true),
			)
			require.NoError(b, err, "creating enabled plonk2 prover should succeed")
			defer prover.Close()

			proof, err := prover.Prove(setup.fullWitness)
			require.NoError(b, err, "warmup plonk2 GPU prove should succeed")
			require.NoError(b, gnarkplonk.Verify(proof, setup.vk, setup.publicWitness), "warmup proof should verify")

			b.ReportMetric(float64(setup.spr.GetNbConstraints()), "constraints")
			b.ReportMetric(float64(prover.MemoryPlan().DomainSize), "domain")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				proof, err := prover.Prove(setup.fullWitness)
				require.NoError(b, err, "plonk2 GPU prove should succeed")
				if proof == nil {
					b.Fatal("plonk2 GPU prove returned nil proof")
				}
			}
		})
	}
}

func BenchmarkPlonk2EnabledFullProverAllCurves_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating GPU device should succeed")
	defer dev.Close()

	for _, curve := range benchPlonkCurves {
		for _, constraints := range plonkBenchConstraintCounts(b) {
			setup := newFullProverSetup(b, curve.id, constraints)
			b.Run(
				fmt.Sprintf("%s/constraints=%d/%s", curve.name, constraints, setup.backendLabel()),
				func(b *testing.B) {
					setup.benchmarkPlonk2Enabled(b, dev)
				},
			)
		}
	}
}

type fullProverBLS12377Setup struct {
	spr           *blscs.SparseR1CS
	pk            *blsplonk.ProvingKey
	vk            *blsplonk.VerifyingKey
	gpk           *oldplonk.GPUProvingKey
	fullWitness   witness.Witness
	publicWitness witness.Witness
}

func newFullProverBLS12377Setup(tb testing.TB, constraints int) *fullProverBLS12377Setup {
	tb.Helper()

	ccs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(tb, err, "compiling benchmark circuit should succeed")

	spr, ok := ccs.(*blscs.SparseR1CS)
	require.True(tb, ok, "BLS12-377 scs compile should return SparseR1CS")

	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "PlonK setup should succeed")

	pk := pkI.(*blsplonk.ProvingKey)
	vk := vkI.(*blsplonk.VerifyingKey)
	store, err := oldplonk.NewSRSStore(testSRSRootDir)
	require.NoError(tb, err, "creating SRS store should succeed")
	pinned, err := store.GetSRSGPUPinned(context.Background(), ccs)
	require.NoError(tb, err, "loading pinned BLS12-377 GPU SRS should succeed")
	gpk := oldplonk.NewGPUProvingKeyFromPinned(pinned, vk)

	assignment := newBenchChainAssignment(ecc.BLS12_377, constraints)
	fullWitness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
	require.NoError(tb, err, "creating full witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(tb, err, "extracting public witness should succeed")

	return &fullProverBLS12377Setup{
		spr:           spr,
		pk:            pk,
		vk:            vk,
		gpk:           gpk,
		fullWitness:   fullWitness,
		publicWitness: publicWitness,
	}
}

type fullProverSetup struct {
	curveID       ecc.ID
	ccs           constraint.ConstraintSystem
	pk            gnarkplonk.ProvingKey
	vk            gnarkplonk.VerifyingKey
	fullWitness   witness.Witness
	publicWitness witness.Witness
}

func newFullProverSetup(tb testing.TB, curveID ecc.ID, constraints int) *fullProverSetup {
	tb.Helper()

	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(tb, err, "compiling benchmark circuit should succeed")

	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "PlonK setup should succeed")

	fullWitness, err := frontend.NewWitness(
		newBenchChainAssignment(curveID, constraints),
		curveID.ScalarField(),
	)
	require.NoError(tb, err, "creating full witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(tb, err, "extracting public witness should succeed")

	return &fullProverSetup{
		curveID:       curveID,
		ccs:           ccs,
		pk:            pk,
		vk:            vk,
		fullWitness:   fullWitness,
		publicWitness: publicWitness,
	}
}

func (s *fullProverSetup) backendLabel() string {
	return "generic-gpu"
}

func (s *fullProverSetup) benchmarkPlonk2Enabled(b *testing.B, dev *gpu.Device) {
	prover, err := NewProver(dev, s.ccs, s.pk, WithEnabled(true), WithStrictMode(true))
	require.NoError(b, err, "creating enabled plonk2 prover should succeed")
	defer prover.Close()

	proof, err := prover.Prove(s.fullWitness)
	require.NoError(b, err, "warmup plonk2 prove should succeed")
	require.NoError(b, gnarkplonk.Verify(proof, s.vk, s.publicWitness), "warmup proof should verify")

	if constraintSystem, ok := s.ccs.(interface{ GetNbConstraints() int }); ok {
		b.ReportMetric(float64(constraintSystem.GetNbConstraints()), "constraints")
	}
	b.ReportMetric(float64(prover.MemoryPlan().DomainSize), "domain")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err := prover.Prove(s.fullWitness)
		require.NoError(b, err, "plonk2 prove should succeed")
		if proof == nil {
			b.Fatal("plonk2 prove returned nil proof")
		}
	}
}

func (s *fullProverBLS12377Setup) benchmarkCPU(b *testing.B) {
	proof, err := gnarkplonk.Prove(s.spr, s.pk, s.fullWitness)
	require.NoError(b, err, "warmup CPU prove should succeed")
	require.NoError(b, gnarkplonk.Verify(proof, s.vk, s.publicWitness), "warmup CPU proof should verify")

	b.ReportMetric(float64(s.spr.GetNbConstraints()), "constraints")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err := gnarkplonk.Prove(s.spr, s.pk, s.fullWitness)
		require.NoError(b, err, "CPU prove should succeed")
		if proof == nil {
			b.Fatal("CPU prove returned nil proof")
		}
	}
}

func (s *fullProverBLS12377Setup) benchmarkCurrentGPU(b *testing.B, dev *gpu.Device) {
	require.NoError(b, s.gpk.Prepare(dev, s.spr), "GPU proving key preparation should succeed")
	proof, err := oldplonk.GPUProve(dev, s.gpk, s.spr, s.fullWitness)
	require.NoError(b, err, "warmup GPU prove should succeed")
	require.NoError(b, gnarkplonk.Verify(proof, s.vk, s.publicWitness), "warmup GPU proof should verify")

	b.ReportMetric(float64(s.spr.GetNbConstraints()), "constraints")
	b.ReportMetric(float64(s.gpk.Size()), "domain")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err := oldplonk.GPUProve(dev, s.gpk, s.spr, s.fullWitness)
		require.NoError(b, err, "GPU prove should succeed")
		if proof == nil {
			b.Fatal("GPU prove returned nil proof")
		}
	}
}
