package recursion

import (
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// innerCircuitSize is the per-column row count for the inner wizard.
const innerCircuitSize = 1 << 14

// defineInnerRealCircuit registers a non-trivial inner wizard with three
// committed columns and two inclusion queries, mirroring the shape of a small
// zkEVM-style module.
func defineInnerRealCircuit(bui *wizard.Builder) {
	a := bui.RegisterCommit("A", innerCircuitSize)
	b := bui.RegisterCommit("B", innerCircuitSize)
	c := bui.RegisterCommit("C", innerCircuitSize)

	bui.Inclusion("Q_AB", []ifaces.Column{a}, []ifaces.Column{b})
	bui.Inclusion("Q_AC", []ifaces.Column{a}, []ifaces.Column{c})
}

// proveInnerRealCircuit assigns zero columns — the inclusion queries above hold
// trivially. The work in the prover comes from the compilation, not the data.
func proveInnerRealCircuit(run *wizard.ProverRuntime) {
	zero := smartvectors.NewConstant(field.Zero(), innerCircuitSize)
	run.AssignColumn("A", zero)
	run.AssignColumn("B", zero)
	run.AssignColumn("C", zero)
}

// productionOuterSuite is the sequence of compilation steps applied to the
// outer recursion wizard *after* DefineRecursionOf(MaxNumProof: 2). It mirrors
// the conglomeration-segment compilation in distributed/segment_compilation.go
// (numComs=16, ForceNumOpenedColumns=64, two Vortex rounds with self-recursion
// in between, final round PremarkAsSelfRecursed) so that outer-prove actually
// produces a single succinct outer wizard proof, not a dummy in-process check.
//
// Skipped vs. the production conglomeration path:
//   - no plonkinwizard.Compile (DefineRecursionOf already lowers Plonk-in-wizard
//     queries via plonkinternal.PlonkCheck);
//   - no AddPrecomputedMerkleRootToPublicInputs / VerifyingKeyPublicInput
//     plumbing (those require segment-specific public inputs we don't declare);
//   - no final makeRecursion(BLS) wrapper that the conglomeration applies on top.
func productionOuterSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
			compiler.WithStitcherMinSize(2),
		),
		vortex.Compile(
			16,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(512),
		),
		selfrecursion.SelfRecurse,
		poseidon2.CompilePoseidon2,
		cleanup.CleanUp,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<14),
			compiler.WithStitcherMinSize(2),
		),
		vortex.Compile(
			16,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.PremarkAsSelfRecursed(),
			vortex.WithOptionalSISHashingThreshold(512),
		),
	}
}

// realisticInnerSuite mirrors the heavy "case-1" suite from TestLookup:
// arithmetic compilation → first vortex round → self-recursion → second vortex
// round (premarked as self-recursed so the outer recursion can consume it).
func realisticInnerSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(compiler.WithTargetColSize(1 << 13)),
		vortex.Compile(
			16,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(compiler.WithTargetColSize(1 << 13)),
		vortex.Compile(
			16,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
			vortex.PremarkAsSelfRecursed(),
		),
	}
}

// aggTimings records per-phase wall-clock time for a single 2→1 aggregation.
type aggTimings struct {
	innerCompile  time.Duration
	innerProveAB  time.Duration
	outerCompile  time.Duration
	outerProve    time.Duration
	outerVerify   time.Duration
	totalEndToEnd time.Duration
}

func runAggregate2To1(tb testing.TB, suite []func(*wizard.CompiledIOP)) aggTimings {
	tb.Helper()

	overallStart := time.Now()
	var t aggTimings

	t0 := time.Now()
	comp1 := wizard.Compile(defineInnerRealCircuit, suite...)
	t.innerCompile = time.Since(t0)

	var recCtx *Recursion
	define2 := func(build2 *wizard.Builder) {
		recCtx = DefineRecursionOf(build2.CompiledIOP, comp1, Parameters{
			Name:        "agg2to1",
			MaxNumProof: 2,
		})
	}

	t0 = time.Now()
	comp2 := wizard.Compile(define2, productionOuterSuite()...)
	t.outerCompile = time.Since(t0)

	stopRound := recCtx.GetStoppingRound() + 1

	t0 = time.Now()
	rt1 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound, false)
	witness1 := ExtractWitness(rt1)
	rt2 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound, false)
	witness2 := ExtractWitness(rt2)
	t.innerProveAB = time.Since(t0)

	prove2 := func(run *wizard.ProverRuntime) {
		recCtx.Assign(run, []Witness{witness1, witness2}, nil)
	}

	t0 = time.Now()
	proof2 := wizard.Prove(comp2, prove2)
	t.outerProve = time.Since(t0)

	t0 = time.Now()
	require.NoError(tb, wizard.Verify(comp2, proof2), "aggregated proof failed verification")
	t.outerVerify = time.Since(t0)

	t.totalEndToEnd = time.Since(overallStart)
	return t
}

func logTimings(tb testing.TB, label string, t aggTimings) {
	tb.Helper()
	tb.Logf("[%s] inner-compile=%s  inner-prove-x2=%s  outer-compile=%s  outer-prove=%s  outer-verify=%s  total=%s",
		label,
		t.innerCompile.Round(time.Millisecond),
		t.innerProveAB.Round(time.Millisecond),
		t.outerCompile.Round(time.Millisecond),
		t.outerProve.Round(time.Millisecond),
		t.outerVerify.Round(time.Millisecond),
		t.totalEndToEnd.Round(time.Millisecond),
	)
}

// BenchmarkAggregate2To1Real benchmarks the 2→1 outer prove on the realistic
// inner suite. Inner compilation/witness generation are reused across
// iterations; only the outer prove+verify is measured.
func BenchmarkAggregate2To1Real(b *testing.B) {
	benchmarkAggregate2To1(b, realisticInnerSuite())
}

func benchmarkAggregate2To1(b *testing.B, suite []func(*wizard.CompiledIOP)) {
	b.StopTimer()

	comp1 := wizard.Compile(defineInnerRealCircuit, suite...)

	var recCtx *Recursion
	define2 := func(build2 *wizard.Builder) {
		recCtx = DefineRecursionOf(build2.CompiledIOP, comp1, Parameters{
			Name:        "agg2to1",
			MaxNumProof: 2,
		})
	}
	comp2 := wizard.Compile(define2, productionOuterSuite()...)

	stopRound := recCtx.GetStoppingRound() + 1
	rt1 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound, false)
	witness1 := ExtractWitness(rt1)
	rt2 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound, false)
	witness2 := ExtractWitness(rt2)

	prove2 := func(run *wizard.ProverRuntime) {
		recCtx.Assign(run, []Witness{witness1, witness2}, nil)
	}

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		proof := wizard.Prove(comp2, prove2)
		if err := wizard.Verify(comp2, proof); err != nil {
			b.Fatalf("aggregated proof failed verification: %v", err)
		}
	}
}

// fullPipelineRecursionSuite mirrors makeRecursion(false) in
// distributed/segment_compilation.go. It is applied to the *outermost* wizard
// (which wraps the 2→1 outer wizard via DefineRecursionOf(MaxNumProof: 1)) and
// produces the final succinct conglomeration-node-equivalent proof.
//
// Skipped vs. production:
//   - no plonkinwizard.Compile (DefineRecursionOf already lowers Plonk-in-wizard);
//   - no AddPrecomputedMerkleRootToPublicInputs / VerifyingKey2PublicInput plumbing.
func fullPipelineRecursionSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<19),
			compiler.WithStitcherMinSize(2),
		),
		vortex.Compile(
			2,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(512),
		),
		selfrecursion.SelfRecurse,
		poseidon2.CompilePoseidon2,
		cleanup.CleanUp,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<14),
			compiler.WithStitcherMinSize(2),
		),
		vortex.Compile(
			16,
			false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.PremarkAsSelfRecursed(),
			vortex.WithOptionalSISHashingThreshold(512),
		),
	}
}

// fullPipelineTimings extends aggTimings with the second recursion-wrap layer
// (rec2*) that mirrors production's makeRecursion(false) pass.
type fullPipelineTimings struct {
	innerCompile  time.Duration
	innerProveAB  time.Duration
	rec1Compile   time.Duration
	rec2Compile   time.Duration
	rec1Prove     time.Duration
	rec2Prove     time.Duration
	rec2Verify    time.Duration
	totalEndToEnd time.Duration
}

func logFullPipelineTimings(tb testing.TB, t fullPipelineTimings) {
	tb.Helper()
	tb.Logf("[full-pipeline] inner-compile=%s  inner-prove-x2=%s  rec1-compile=%s  rec2-compile=%s  rec1-prove=%s  rec2-prove=%s  rec2-verify=%s  total=%s",
		t.innerCompile.Round(time.Millisecond),
		t.innerProveAB.Round(time.Millisecond),
		t.rec1Compile.Round(time.Millisecond),
		t.rec2Compile.Round(time.Millisecond),
		t.rec1Prove.Round(time.Millisecond),
		t.rec2Prove.Round(time.Millisecond),
		t.rec2Verify.Round(time.Millisecond),
		t.totalEndToEnd.Round(time.Millisecond),
	)
}

// runFullPipeline2To1 runs the production-shape 2→1 conglomeration node:
// inner wizard → DefineRecursionOf(MaxNumProof:2) + productionOuterSuite (rec1)
// → DefineRecursionOf(MaxNumProof:1) + fullPipelineRecursionSuite (rec2) → verify.
func runFullPipeline2To1(tb testing.TB) fullPipelineTimings {
	tb.Helper()

	overallStart := time.Now()
	var t fullPipelineTimings

	t0 := time.Now()
	comp1 := wizard.Compile(defineInnerRealCircuit, realisticInnerSuite()...)
	t.innerCompile = time.Since(t0)

	var recCtx2 *Recursion
	define2 := func(build *wizard.Builder) {
		recCtx2 = DefineRecursionOf(build.CompiledIOP, comp1, Parameters{
			Name:        "agg2to1",
			MaxNumProof: 2,
		})
	}

	t0 = time.Now()
	comp2 := wizard.Compile(define2, productionOuterSuite()...)
	t.rec1Compile = time.Since(t0)

	var recCtx3 *Recursion
	define3 := func(build *wizard.Builder) {
		recCtx3 = DefineRecursionOf(build.CompiledIOP, comp2, Parameters{
			Name:        "outer-recursion",
			MaxNumProof: 1,
		})
	}

	t0 = time.Now()
	comp3 := wizard.Compile(define3, fullPipelineRecursionSuite()...)
	t.rec2Compile = time.Since(t0)

	stopRound2 := recCtx2.GetStoppingRound() + 1

	t0 = time.Now()
	rt1 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound2, false)
	witness1 := ExtractWitness(rt1)
	rt2 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound2, false)
	witness2 := ExtractWitness(rt2)
	t.innerProveAB = time.Since(t0)

	prove2 := func(run *wizard.ProverRuntime) {
		recCtx2.Assign(run, []Witness{witness1, witness2}, nil)
	}

	stopRound3 := recCtx3.GetStoppingRound() + 1

	t0 = time.Now()
	rtOuter := wizard.RunProverUntilRound(comp2, prove2, stopRound3, false)
	witnessOuter := ExtractWitness(rtOuter)
	t.rec1Prove = time.Since(t0)

	prove3 := func(run *wizard.ProverRuntime) {
		recCtx3.Assign(run, []Witness{witnessOuter}, nil)
	}

	t0 = time.Now()
	finalProof := wizard.Prove(comp3, prove3)
	t.rec2Prove = time.Since(t0)

	t0 = time.Now()
	require.NoError(tb, wizard.Verify(comp3, finalProof), "full-pipeline aggregated proof failed verification")
	t.rec2Verify = time.Since(t0)

	t.totalEndToEnd = time.Since(overallStart)
	return t
}

// BenchmarkAggregate2To1FullPipeline benchmarks only the rec2 prove+verify
// step (the outermost wrap). All upstream compilation, inner proofs and the
// rec1 prove that produces the witness for rec2 are reused across iterations.
func BenchmarkAggregate2To1FullPipeline(b *testing.B) {
	b.StopTimer()

	comp1 := wizard.Compile(defineInnerRealCircuit, realisticInnerSuite()...)

	var recCtx2 *Recursion
	define2 := func(build *wizard.Builder) {
		recCtx2 = DefineRecursionOf(build.CompiledIOP, comp1, Parameters{
			Name:        "agg2to1",
			MaxNumProof: 2,
		})
	}
	comp2 := wizard.Compile(define2, productionOuterSuite()...)

	var recCtx3 *Recursion
	define3 := func(build *wizard.Builder) {
		recCtx3 = DefineRecursionOf(build.CompiledIOP, comp2, Parameters{
			Name:        "outer-recursion",
			MaxNumProof: 1,
		})
	}
	comp3 := wizard.Compile(define3, fullPipelineRecursionSuite()...)

	stopRound2 := recCtx2.GetStoppingRound() + 1
	rt1 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound2, false)
	witness1 := ExtractWitness(rt1)
	rt2 := wizard.RunProverUntilRound(comp1, proveInnerRealCircuit, stopRound2, false)
	witness2 := ExtractWitness(rt2)

	prove2 := func(run *wizard.ProverRuntime) {
		recCtx2.Assign(run, []Witness{witness1, witness2}, nil)
	}

	stopRound3 := recCtx3.GetStoppingRound() + 1

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// rec1 prove — produces the witness consumed by rec3 (the outermost wrap).
		// Run inside the loop so each iteration measures the real per-aggregation-
		// node cost: rec1 prove + rec2 prove + rec2 verify.
		rtOuter := wizard.RunProverUntilRound(comp2, prove2, stopRound3, false)
		witnessOuter := ExtractWitness(rtOuter)

		prove3 := func(run *wizard.ProverRuntime) {
			recCtx3.Assign(run, []Witness{witnessOuter}, nil)
		}

		proof := wizard.Prove(comp3, prove3)
		if err := wizard.Verify(comp3, proof); err != nil {
			b.Fatalf("full-pipeline aggregated proof failed verification: %v", err)
		}
	}
}
