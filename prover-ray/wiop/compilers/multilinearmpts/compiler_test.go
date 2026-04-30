package multilinearmpts

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// buildLagrangeTable returns the Lagrange-form evaluation table of a
// random degree-(n-1) polynomial at the standard n-th roots of unity.
func buildLagrangeTable(rng *rand.Rand, n int) []field.Element {
	coeffs := make([]field.Element, n)
	for i := range coeffs {
		coeffs[i] = field.PseudoRand(rng)
	}
	// NTT: coefficient form → Lagrange form.
	d := fft.NewDomain(uint64(n))
	d.FFT(coeffs, fft.DIT)
	return coeffs
}

// TestMPTSSinglePoly runs Compile on a single LagrangeEval and verifies that
// prover and verifier actions are both consistent.
func TestMPTSSinglePoly(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xdeadbeef, 0))
	const n = 32

	sys := wiop.NewSystemf("test")
	round0 := sys.NewRound()
	round1 := sys.NewRound()

	mod := sys.NewSizedModule(sys.Context.Childf("mod"), n, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, round0)
	evalCoin := round1.NewCoinField(sys.Context.Childf("z"))

	sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, evalCoin)

	// Run the compiler.
	Compile(sys)

	// Build the Lagrange table.
	lagrange := buildLagrangeTable(rng, n)

	// Derive coefficient table for ground-truth verification.
	coeffs := make([]field.Element, n)
	copy(coeffs, lagrange)
	fft.NewDomain(uint64(n)).FFTInverse(coeffs, fft.DIF)

	// Create runtime and execute round 0.
	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col, &wiop.ConcreteVector{
		Plain: field.VecFromBase(lagrange),
	})

	// Advance to round 1 (derives the eval coin).
	rt.AdvanceRound()

	// Assign the eval coin claim for the (now-reduced) LagrangeEval.
	// Even though the LagrangeEval is reduced, its claim cell still lives in the
	// runtime and must be assigned before advancing.
	for _, le := range sys.LagrangeEvals {
		if !le.IsAlreadyAssigned(rt) {
			le.SelfAssign(rt)
		}
	}

	// Advance to round 2 (the MPTS round).
	rt.AdvanceRound()

	// Run prover actions for round 2 (the MPTS round).
	for _, a := range rt.CurrentRound().ProverActions {
		a.Run(rt)
	}

	// Run verifier actions.
	for _, va := range rt.CurrentRound().VerifierActions {
		if err := va.Check(rt); err != nil {
			t.Fatalf("verifier check: %v", err)
		}
	}

	// Check the MultilinearEval query.
	for _, me := range sys.MultilinearEvals {
		if err := me.Check(rt); err != nil {
			t.Fatalf("MultilinearEval check: %v", err)
		}
	}
}

// TestMPTSMultiPoly tests Compile with multiple polynomials in one LagrangeEval.
func TestMPTSMultiPoly(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xc0ffee, 0))
	const n = 16
	const m = 4

	sys := wiop.NewSystemf("test-multi")
	round0 := sys.NewRound()
	round1 := sys.NewRound()

	mod := sys.NewSizedModule(sys.Context.Childf("mod"), n, wiop.PaddingDirectionNone)
	cols := make([]*wiop.Column, m)
	views := make([]*wiop.ColumnView, m)
	lagranges := make([][]field.Element, m)

	for j := range m {
		cols[j] = mod.NewColumn(sys.Context.Childf("col[%d]", j), wiop.VisibilityOracle, round0)
		views[j] = cols[j].View()
		lagranges[j] = buildLagrangeTable(rng, n)
	}

	evalCoin := round1.NewCoinField(sys.Context.Childf("z"))
	sys.NewLagrangeEval(sys.Context.Childf("le"), views, evalCoin)

	Compile(sys)

	rt := wiop.NewRuntime(sys)
	for j := range m {
		rt.AssignColumn(cols[j], &wiop.ConcreteVector{
			Plain: field.VecFromBase(lagranges[j]),
		})
	}

	rt.AdvanceRound()
	for _, le := range sys.LagrangeEvals {
		if !le.IsAlreadyAssigned(rt) {
			le.SelfAssign(rt)
		}
	}
	rt.AdvanceRound()

	for _, a := range rt.CurrentRound().ProverActions {
		a.Run(rt)
	}
	for _, va := range rt.CurrentRound().VerifierActions {
		if err := va.Check(rt); err != nil {
			t.Fatalf("verifier check: %v", err)
		}
	}
	for _, me := range sys.MultilinearEvals {
		if err := me.Check(rt); err != nil {
			t.Fatalf("MultilinearEval check: %v", err)
		}
	}
}

// TestMPTSMultipleQueries tests two separate LagrangeEval queries in the same group.
func TestMPTSMultipleQueries(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xbeef, 0))
	const n = 8

	sys := wiop.NewSystemf("test-multi-queries")
	round0 := sys.NewRound()
	round1 := sys.NewRound()

	mod := sys.NewSizedModule(sys.Context.Childf("mod"), n, wiop.PaddingDirectionNone)
	col0 := mod.NewColumn(sys.Context.Childf("col0"), wiop.VisibilityOracle, round0)
	col1 := mod.NewColumn(sys.Context.Childf("col1"), wiop.VisibilityOracle, round0)

	evalCoin := round1.NewCoinField(sys.Context.Childf("z"))
	sys.NewLagrangeEval(sys.Context.Childf("le0"), []*wiop.ColumnView{col0.View()}, evalCoin)
	sys.NewLagrangeEval(sys.Context.Childf("le1"), []*wiop.ColumnView{col1.View()}, evalCoin)

	Compile(sys)

	lag0 := buildLagrangeTable(rng, n)
	lag1 := buildLagrangeTable(rng, n)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(col0, &wiop.ConcreteVector{Plain: field.VecFromBase(lag0)})
	rt.AssignColumn(col1, &wiop.ConcreteVector{Plain: field.VecFromBase(lag1)})

	rt.AdvanceRound()
	for _, le := range sys.LagrangeEvals {
		if !le.IsAlreadyAssigned(rt) {
			le.SelfAssign(rt)
		}
	}
	rt.AdvanceRound()

	for _, a := range rt.CurrentRound().ProverActions {
		a.Run(rt)
	}
	for _, va := range rt.CurrentRound().VerifierActions {
		if err := va.Check(rt); err != nil {
			t.Fatalf("verifier check: %v", err)
		}
	}
	for _, me := range sys.MultilinearEvals {
		if err := me.Check(rt); err != nil {
			t.Fatalf("MultilinearEval check: %v", err)
		}
	}
}
