//go:build cuda

package plonk2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func BenchmarkGenericSolvedWireCommitment_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curve := range benchPlonkCurves {
		for _, constraints := range plonkBenchConstraintCounts(b) {
			b.Run(benchPlonkCaseName(curve.name, constraints), func(b *testing.B) {
				benchmarkGenericSolvedWireCommitment(b, dev, curve.id, constraints)
			})
		}
	}
}

func BenchmarkGenericSolvedWireCommitmentWave_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curve := range benchPlonkCurves {
		for _, constraints := range plonkBenchConstraintCounts(b) {
			b.Run(benchPlonkCaseName(curve.name, constraints), func(b *testing.B) {
				benchmarkGenericSolvedWireCommitmentWave(b, dev, curve.id, constraints)
			})
		}
	}
}

func benchmarkGenericSolvedWireCommitment(
	b *testing.B,
	dev *gpu.Device,
	curveID ecc.ID,
	constraints int,
) {
	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(b, err, "compiling benchmark circuit should succeed")
	srs, srsLagrange := testSRSAssets(b).loadForCCS(b, ccs)
	pk, _, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(b, err, "PlonK setup should succeed")

	state, err := newGenericProverState(dev, ccs, pk)
	require.NoError(b, err, "preparing generic GPU prover state should succeed")
	defer state.Close()

	fullWitness, err := frontend.NewWitness(
		newBenchChainAssignment(curveID, constraints),
		curveID.ScalarField(),
	)
	require.NoError(b, err, "creating witness should succeed")
	lRaw := solvedLWireRaw(b, ccs, fullWitness)
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.ReportMetric(float64(ccs.GetNbConstraints()), "constraints")
	b.ReportMetric(float64(state.n), "domain_size")
	b.SetBytes(int64(len(lRaw)) * 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := state.commitLagrangeRaw(lRaw)
		require.NoError(b, err, "generic GPU L-wire commitment should succeed")
	}
}

func benchmarkGenericSolvedWireCommitmentWave(
	b *testing.B,
	dev *gpu.Device,
	curveID ecc.ID,
	constraints int,
) {
	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(b, err, "compiling benchmark circuit should succeed")
	srs, srsLagrange := testSRSAssets(b).loadForCCS(b, ccs)
	pk, _, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(b, err, "PlonK setup should succeed")

	prover, err := NewProver(dev, ccs, pk, WithEnabled(true))
	require.NoError(b, err, "preparing generic GPU prover state should succeed")
	defer prover.Close()
	state := prover.genericState
	require.NotNil(b, state, "generic state should be prepared")

	fullWitness, err := frontend.NewWitness(
		newBenchChainAssignment(curveID, constraints),
		curveID.ScalarField(),
	)
	require.NoError(b, err, "creating witness should succeed")
	lRaw, rRaw, oRaw := solvedLROWireRaw(b, ccs, fullWitness)
	wave := [][]uint64{lRaw, rRaw, oRaw}

	require.NoError(b, state.pinLagrangeCommitmentWave(), "pinning Lagrange MSM wave buffers should succeed")
	_, err = state.commitLagrangeWaveRaw(wave)
	require.NoError(b, err, "warmup generic GPU LRO commitment wave should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	plan := prover.MemoryPlan()
	b.SetBytes(int64(len(lRaw)+len(rRaw)+len(oRaw)) * 8)
	b.ResetTimer()
	var lTimings, rTimings, oTimings [MSMPhaseCount]float32
	for i := 0; i < b.N; i++ {
		_, err := state.commitLagrangeRaw(lRaw)
		require.NoError(b, err, "generic GPU L commitment should succeed")
		lTimings = state.lagKzg.LastPhaseTimings()
		_, err = state.commitLagrangeRaw(rRaw)
		require.NoError(b, err, "generic GPU R commitment should succeed")
		rTimings = state.lagKzg.LastPhaseTimings()
		_, err = state.commitLagrangeRaw(oRaw)
		require.NoError(b, err, "generic GPU O commitment should succeed")
		oTimings = state.lagKzg.LastPhaseTimings()
	}
	reportMSMTimings(b, "l", lTimings)
	reportMSMTimings(b, "r", rTimings)
	reportMSMTimings(b, "o", oTimings)
	b.ReportMetric(float64(ccs.GetNbConstraints()), "constraints")
	b.ReportMetric(float64(plan.DomainSize), "domain_size")
	b.ReportMetric(float64(plan.PersistentBytes()), "persistent_bytes")
	b.ReportMetric(float64(plan.PerWaveBytes()), "per_wave_bytes")
	b.ReportMetric(float64(plan.EstimatedPeakBytes()), "peak_bytes")
	b.StopTimer()
	require.NoError(b, state.releaseLagrangeCommitmentWave(), "releasing Lagrange MSM wave buffers should succeed")
}

func reportMSMTimings(b *testing.B, prefix string, timings [MSMPhaseCount]float32) {
	b.Helper()
	for phase := MSMPhase(0); phase < MSMPhaseCount; phase++ {
		b.ReportMetric(float64(timings[phase]*1000), prefix+"_"+phase.String()+"_us")
	}
}

func solvedLWireRaw(tb testing.TB, ccs any, fullWitness witness.Witness) []uint64 {
	tb.Helper()
	l, _, _ := solvedLROWireRaw(tb, ccs, fullWitness)
	return l
}

func solvedLROWireRaw(tb testing.TB, ccs any, fullWitness witness.Witness) ([]uint64, []uint64, []uint64) {
	tb.Helper()
	switch spr := ccs.(type) {
	case *csbn254.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(tb, err, "solving BN254 witness should succeed")
		typed := solution.(*csbn254.SparseR1CSSolution)
		return genericRawBN254Fr([]bnfr.Element(typed.L)),
			genericRawBN254Fr([]bnfr.Element(typed.R)),
			genericRawBN254Fr([]bnfr.Element(typed.O))
	case *csbls12377.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(tb, err, "solving BLS12-377 witness should succeed")
		typed := solution.(*csbls12377.SparseR1CSSolution)
		return genericRawBLS12377Fr([]blsfr.Element(typed.L)),
			genericRawBLS12377Fr([]blsfr.Element(typed.R)),
			genericRawBLS12377Fr([]blsfr.Element(typed.O))
	case *csbw6761.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(tb, err, "solving BW6-761 witness should succeed")
		typed := solution.(*csbw6761.SparseR1CSSolution)
		return genericRawBW6761Fr([]bwfr.Element(typed.L)),
			genericRawBW6761Fr([]bwfr.Element(typed.R)),
			genericRawBW6761Fr([]bwfr.Element(typed.O))
	default:
		tb.Fatalf("unexpected constraint system type %T", ccs)
		return nil, nil, nil
	}
}
