//go:build cuda

package plonk2

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	blscs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/gpu"
	oldplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

func BenchmarkPlonkReferenceFullProverCPUvsCurrentGPU_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating GPU device should succeed")
	defer dev.Close()

	for _, constraints := range benchPlonkConstraintCounts {
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

	srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
	require.NoError(tb, err, "creating unsafe test SRS should succeed")
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "PlonK setup should succeed")

	pk := pkI.(*blsplonk.ProvingKey)
	vk := vkI.(*blsplonk.VerifyingKey)
	gpk := oldplonk.NewGPUProvingKey(oldplonk.ConvertG1AffineToTE(pk.Kzg.G1), vk)

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
