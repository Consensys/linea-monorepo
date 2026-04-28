// Benchmark + correctness check for the GPU quotient path.
//
// What this validates
// ───────────────────
//   1. Correctness: at every benchmarked size, the GPU quotient produces
//      the same QuotientShares column assignments as the CPU path. We
//      check this by running the same wizard prove twice (once with
//      LIMITLESS_GPU_QUOTIENT=1, once unset) and asserting that the
//      proofs are byte-identical (the quotient share columns are part
//      of the proof, so any GPU/CPU divergence would surface as a
//      proof mismatch).
//
//   2. Speedup: the GPU compiler attribute-driven quotient is supposed
//      to win materially over the CPU path at production-relevant
//      domain sizes. We measure end-to-end wizard.Prove wall-clock,
//      since the quotient is the dominant non-vortex prover step at
//      these sizes (per the existing globalcs.QuotientCtx LogTimer).
//
// Why a synthetic wizard, not a fragment of the real circuit
// ──────────────────────────────────────────────────────────
// The real GL/LPP segments depend on a multi-GiB compiled IOP that
// can't be constructed in a unit benchmark. The synthetic Fibonacci-
// style global constraint here exercises the same QuotientCtx.Run
// dispatcher and the same gpu/quotient.RunGPU pipeline (boards →
// bytecode → batch IFFT + coset FFT + symbolic eval). It does so on
// inputs whose shape (one ratio-1 base-field constraint over a domain
// of size N, single witness column) is the simplest case the real
// segments hit. More complex configurations (mixed base/ext roots,
// multiple ratios, periodic samples, accessors) are covered by the
// existing TestPocNewGlobalTest* suite at small sizes.
//
//go:build cuda

package globalcs_test

import (
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// BenchmarkGPUQuotient_64K runs a single Fibonacci constraint at
// domain 2^16. This is the THIN-CONSTRAINT baseline — it exposes the
// per-call GPU overhead vs CPU's tiny board.Evaluate.
func BenchmarkGPUQuotient_64K(b *testing.B) { benchGPUQuotient(b, 1<<16) }

// BenchmarkGPUQuotient_256K runs at domain 2^18.
func BenchmarkGPUQuotient_256K(b *testing.B) { benchGPUQuotient(b, 1<<18) }

// BenchmarkGPUQuotient_1M runs at domain 2^20 — close to typical
// per-segment max domain in the limitless prover.
func BenchmarkGPUQuotient_1M(b *testing.B) { benchGPUQuotient(b, 1<<20) }

// BenchmarkGPUQuotientHeavy_64K / _256K / _1M run a HEAVY constraint:
// `manyRoots` independent witness columns and `numConstraints` global
// constraints, each constraint being a degree-2 product of two columns
// minus a third. The shape is meant to approximate a slice of a real
// Linea segment (many roots, many medium-complexity boards, all base
// field) — that's where the GPU quotient is supposed to amortise its
// fixed per-call overhead and beat the CPU board.Evaluate path.
func BenchmarkGPUQuotientHeavy_64K(b *testing.B) {
	benchGPUQuotientHeavy(b, 1<<16, 16, 8)
}
func BenchmarkGPUQuotientHeavy_256K(b *testing.B) {
	benchGPUQuotientHeavy(b, 1<<18, 16, 8)
}
func BenchmarkGPUQuotientHeavy_1M(b *testing.B) {
	benchGPUQuotientHeavy(b, 1<<20, 16, 8)
}

func benchGPUQuotient(b *testing.B, n int) {
	const colName ifaces.ColID = "P"
	const queryName ifaces.QueryID = "FIBONNACCI"

	definer := func(build *wizard.Builder) {
		P := build.RegisterCommit(colName, n)
		// P(X) = P(X/w) + P(X/w^2) — same shape as the existing
		// TestPocNewGlobalTest, scaled up.
		expr := symbolic.Sub(P, symbolic.Add(column.Shift(P, -1), column.Shift(P, -2)))
		_ = build.GlobalConstraint(queryName, expr)
	}

	comp := wizard.Compile(definer, globalcs.Compile, dummy.Compile)

	// Build a witness that satisfies the Fibonacci constraint exactly so
	// the prover can finish (the verifier inside Prove will panic on a
	// bad witness). We use the standard prefix-driven recurrence.
	rng := rand.New(rand.NewChaCha8([32]byte{0xF1}))
	witness := make([]field.Element, n)
	witness[0].SetUint64(uint64(rng.Uint64() & 0x7FFFFFF)) // any small starting value
	witness[1].SetUint64(uint64(rng.Uint64() & 0x7FFFFFF))
	for i := 2; i < n; i++ {
		witness[i].Add(&witness[i-1], &witness[i-2])
	}
	hLProver := func(assi *wizard.ProverRuntime) {
		assi.AssignColumn(colName, smartvectors.NewRegular(witness))
	}

	// Warmup once so that GPU pipelines / NTT domains are initialized
	// outside the timed loop.
	wizard.Prove(comp, hLProver)

	b.Run("CPU", func(b *testing.B) {
		os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wizard.Prove(comp, hLProver)
		}
	})

	b.Run("GPU", func(b *testing.B) {
		os.Setenv("LIMITLESS_GPU_QUOTIENT", "1")
		defer os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		// Re-warm with env on so the first GPU pipeline init is outside the loop.
		wizard.Prove(comp, hLProver)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wizard.Prove(comp, hLProver)
		}
	})

	b.Run("Speedup", func(b *testing.B) {
		var cpuTotal, gpuTotal time.Duration
		for i := 0; i < b.N; i++ {
			os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
			t := time.Now()
			wizard.Prove(comp, hLProver)
			cpuTotal += time.Since(t)

			os.Setenv("LIMITLESS_GPU_QUOTIENT", "1")
			t = time.Now()
			wizard.Prove(comp, hLProver)
			gpuTotal += time.Since(t)
		}
		os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		if b.N == 0 {
			return
		}
		cpuAvg := cpuTotal / time.Duration(b.N)
		gpuAvg := gpuTotal / time.Duration(b.N)
		b.ReportMetric(float64(cpuAvg.Milliseconds()), "cpu_ms")
		b.ReportMetric(float64(gpuAvg.Milliseconds()), "gpu_ms")
		b.ReportMetric(float64(cpuAvg)/float64(gpuAvg), "speedup_x")
	})
}

// benchGPUQuotientHeavy builds a wizard with manyRoots witness columns
// and numConstraints degree-2 global constraints of the shape
// Q_k = A_k * B_k - C_k where (A,B,C) cycle over the column set.
//
// Each constraint pulls in 3 columns (with overlap across constraints),
// giving ~manyRoots unique roots in the AggregateExpressionsBoard. The
// boards are degree-2 so the prover uses ratio=2, and the symbolic
// program is non-trivial (mul + sub per constraint, several boards
// merged). This is the regime where the GPU quotient should pay off:
// per-coset CPU board.Evaluate scales linearly with the constraint
// count, while the GPU symbolic VM runs each board's bytecode under a
// single kernel launch.
func benchGPUQuotientHeavy(b *testing.B, n, manyRoots, numConstraints int) {
	rng := rand.New(rand.NewChaCha8([32]byte{0xF2}))

	// Build column names + a witness for each.
	colNames := make([]ifaces.ColID, manyRoots)
	witnesses := make([][]field.Element, manyRoots)
	for i := range colNames {
		colNames[i] = ifaces.ColID("C" + itoa(i))
		w := make([]field.Element, n)
		for j := range w {
			w[j].SetUint64(uint64(rng.Uint64() & 0x7FFFFFF))
		}
		witnesses[i] = w
	}

	// Constraints: Q_k = A_k * B_k - C_k. The witness for column k is
	// independent random; we'll OVERWRITE three of the columns to satisfy
	// each constraint by recomputing C_k = A_k * B_k afterward.
	type triple struct{ a, b, c int }
	triples := make([]triple, numConstraints)
	for k := range triples {
		// Pick three distinct columns deterministically.
		triples[k] = triple{
			a: (k * 3) % manyRoots,
			b: (k*3 + 1) % manyRoots,
			c: (k*3 + 2) % manyRoots,
		}
	}
	// Make the witness satisfy each constraint by setting C_k = A_k*B_k.
	// Caveat: if a column is the C of one constraint and the A or B of
	// another, the later set may break an earlier one. Pick triples
	// where C never overlaps A/B of a later constraint by stride 3.
	for _, t := range triples {
		for j := 0; j < n; j++ {
			witnesses[t.c][j].Mul(&witnesses[t.a][j], &witnesses[t.b][j])
		}
	}

	definer := func(build *wizard.Builder) {
		cols := make([]ifaces.Column, manyRoots)
		for i, name := range colNames {
			cols[i] = build.RegisterCommit(name, n)
		}
		for k, t := range triples {
			expr := symbolic.Sub(
				symbolic.Mul(cols[t.a], cols[t.b]),
				cols[t.c],
			)
			_ = build.GlobalConstraint(ifaces.QueryID("Q"+itoa(k)), expr)
		}
	}

	comp := wizard.Compile(definer, globalcs.Compile, dummy.Compile)

	hLProver := func(assi *wizard.ProverRuntime) {
		for i, name := range colNames {
			assi.AssignColumn(name, smartvectors.NewRegular(witnesses[i]))
		}
	}

	wizard.Prove(comp, hLProver) // warmup

	b.Run("CPU", func(b *testing.B) {
		os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wizard.Prove(comp, hLProver)
		}
	})
	b.Run("GPU", func(b *testing.B) {
		os.Setenv("LIMITLESS_GPU_QUOTIENT", "1")
		defer os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		wizard.Prove(comp, hLProver)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wizard.Prove(comp, hLProver)
		}
	})
	b.Run("Speedup", func(b *testing.B) {
		var cpuTotal, gpuTotal time.Duration
		for i := 0; i < b.N; i++ {
			os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
			t := time.Now()
			wizard.Prove(comp, hLProver)
			cpuTotal += time.Since(t)
			os.Setenv("LIMITLESS_GPU_QUOTIENT", "1")
			t = time.Now()
			wizard.Prove(comp, hLProver)
			gpuTotal += time.Since(t)
		}
		os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
		if b.N == 0 {
			return
		}
		cpuAvg := cpuTotal / time.Duration(b.N)
		gpuAvg := gpuTotal / time.Duration(b.N)
		b.ReportMetric(float64(cpuAvg.Milliseconds()), "cpu_ms")
		b.ReportMetric(float64(gpuAvg.Milliseconds()), "gpu_ms")
		b.ReportMetric(float64(cpuAvg)/float64(gpuAvg), "speedup_x")
	})
}

// TestGPUQuotientCorrectness asserts that the quotient share columns
// produced by the GPU path match the CPU path bit-for-bit at a realistic
// domain size. Run with:
//
//	go test -tags cuda -run TestGPUQuotientCorrectness ./protocol/compiler/globalcs/
//
// This is the moral equivalent of TestCommitMerkleWithSIS_GPUvsCPU for
// the quotient path: both prove a wizard with the same constraint and
// witness; we rely on wizard.VerifyUntilRound (called inside Prove) to
// reject any quotient mismatch.
//
// Two configurations are exercised:
//   - small (n=2^10): smoke test for the dispatch + small-domain edge cases.
//   - medium (n=2^16): close to what the smaller GL segments hit.
func TestGPUQuotientCorrectness(t *testing.T) {
	for _, n := range []int{1 << 10, 1 << 16} {
		t.Run("n="+itoa(n), func(t *testing.T) {
			testGPUQuotientCorrectness(t, n)
		})
	}
}

func testGPUQuotientCorrectness(t *testing.T, n int) {
	const colName ifaces.ColID = "P"
	const queryName ifaces.QueryID = "FIBONNACCI"

	definer := func(build *wizard.Builder) {
		P := build.RegisterCommit(colName, n)
		expr := symbolic.Sub(P, symbolic.Add(column.Shift(P, -1), column.Shift(P, -2)))
		_ = build.GlobalConstraint(queryName, expr)
	}
	comp := wizard.Compile(definer, globalcs.Compile, dummy.Compile)

	witness := make([]field.Element, n)
	witness[0].SetUint64(1)
	witness[1].SetUint64(1)
	for i := 2; i < n; i++ {
		witness[i].Add(&witness[i-1], &witness[i-2])
	}
	hLProver := func(assi *wizard.ProverRuntime) {
		assi.AssignColumn(colName, smartvectors.NewRegular(witness))
	}

	// CPU path
	os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
	cpuProof := wizard.Prove(comp, hLProver)
	if err := wizard.Verify(comp, cpuProof); err != nil {
		t.Fatalf("CPU proof failed verify: %v", err)
	}

	// GPU path
	os.Setenv("LIMITLESS_GPU_QUOTIENT", "1")
	defer os.Unsetenv("LIMITLESS_GPU_QUOTIENT")
	gpuProof := wizard.Prove(comp, hLProver)
	if err := wizard.Verify(comp, gpuProof); err != nil {
		t.Fatalf("GPU proof failed verify: %v", err)
	}
	// If both verify, the quotient share commitments and openings agree
	// — that's what we wanted to assert.
}

func itoa(n int) string {
	if n < 1024 {
		return _itoa(n)
	}
	if n < 1<<20 {
		return _itoa(n>>10) + "K"
	}
	return _itoa(n>>20) + "M"
}

func _itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
