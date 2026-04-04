// Package wioptest provides reusable test scenarios for the wiop protocol.
//
// Each scenario bundles a System, the query under test, and two runtime
// population callbacks:
//   - RunHonest: builds a valid witness; Query.Check must pass.
//   - RunInvalid: builds an invalid witness; Query.Check must return an error.
//
// VanishingScenarios returns analogous fixtures for compiler integration tests.
// Because each factory function closes over its own freshly-allocated System,
// calling the factory twice yields two completely independent scenarios.
package wioptest

import (
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/wiop"
)

// Scenario is a self-contained test fixture for one query type.
//
// Call the factory function to obtain an instance; each call returns a new
// Scenario backed by its own System so tests never share mutable state.
type Scenario struct {
	// Name identifies the scenario in test output.
	Name string
	// Sys is the protocol system. Create fresh Runtimes from it; never mutate
	// it directly (e.g. do not compile it — use VanishingScenario for that).
	Sys *wiop.System
	// Query is the registered query whose Check method is under test.
	Query wiop.Query
	// RunHonest populates rt with a fully valid witness: oracle column
	// assignments, round advancement, and SelfAssign where applicable.
	// After this returns, Query.Check(rt) must return nil.
	RunHonest func(rt *wiop.Runtime)
	// RunInvalid populates rt with an invalid witness that Query.Check must
	// reject. For claim-based queries the column assignment is honest but the
	// result cell holds a wrong value; for relation queries the column
	// assignments directly violate the predicate.
	RunInvalid func(rt *wiop.Runtime)
}

// All returns factory functions for every built-in query scenario.
// Call each factory once per test to obtain an independent Scenario.
func All() []func() *Scenario {
	return []func() *Scenario{
		NewLocalOpeningScenario,
		NewLagrangeEvalScenario,
		NewRationalReductionScenario,
		NewPermutationScenario,
		NewInclusionScenario,
	}
}

// ---- LocalOpening ----

// NewLocalOpeningScenario returns a scenario for LocalOpening.
//
//   - Column: [7, 7, 7, 7]; opening at position 2 → claimed value 7.
//   - Invalid: column honest, Result cell set to 0 (≠ 7).
func NewLocalOpeningScenario() *Scenario {
	sys := wiop.NewSystemf("lo-sc")
	r0 := sys.NewRound()
	sys.NewRound() // r1 needed so LagrangeEval / runtime construction is valid
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	lo := col.At(2).Open(sys.Context.Childf("lo"))

	return &Scenario{
		Name:  "LocalOpening",
		Sys:   sys,
		Query: lo,
		RunHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 7))
			lo.SelfAssign(*rt)
		},
		RunInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 7))
			// Correct value is 7; assign 0 instead.
			rt.AssignCell(lo.Result, field.ElemZero())
		},
	}
}

// ---- LagrangeEval ----

// NewLagrangeEvalScenario returns a scenario for LagrangeEval.
//
//   - Column: [3, 3, 3, 3] (constant 3); evaluation point: verifier coin.
//   - The Lagrange evaluation of a constant-3 polynomial is 3 everywhere.
//   - Invalid: column honest, claim cell set to 0 (≠ 3).
func NewLagrangeEvalScenario() *Scenario {
	sys := wiop.NewSystemf("le-sc")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	coin := r1.NewCoinField(sys.Context.Childf("coin"))
	le := sys.NewLagrangeEval(sys.Context.Childf("le"), []*wiop.ColumnView{col.View()}, coin)

	return &Scenario{
		Name:  "LagrangeEval",
		Sys:   sys,
		Query: le,
		RunHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 3))
			rt.AdvanceRound() // sample coin
			le.SelfAssign(*rt)
		},
		RunInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 3))
			rt.AdvanceRound() // sample coin
			// Real evaluation is 3; claim 0 instead.
			rt.AssignCell(le.EvaluationClaims[0], field.ElemZero())
		},
	}
}

// ---- RationalReduction ----

// NewRationalReductionScenario returns a scenario for RationalReduction.
//
//   - Column: [2, 2, 2, 2]; denominator: constant 1; sum = 4 × 2 = 8.
//   - Invalid: column honest, Result cell set to 0 (≠ 8).
func NewRationalReductionScenario() *Scenario {
	sys := wiop.NewSystemf("rr-sc")
	r0 := sys.NewRound()
	sys.NewRound() // result cell goes here
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)

	den := wiop.NewConstantVector(mod, field.NewFromString("1"))
	rr := sys.NewRationalReduction(
		sys.Context.Childf("rr"),
		[]wiop.Fraction{{Numerator: col.View(), Denominator: den}},
	)

	return &Scenario{
		Name:  "RationalReduction",
		Sys:   sys,
		Query: rr,
		RunHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 2))
			rt.AdvanceRound()
			rr.SelfAssign(*rt)
		},
		RunInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, baseVec(4, 2))
			rt.AdvanceRound()
			// Real result is 8; claim 0 instead.
			rt.AssignCell(rr.Result, field.ElemZero())
		},
	}
}

// ---- Permutation ----

// NewPermutationScenario returns a scenario for a Permutation TableRelation.
//
//   - Valid: A = [1, 2, 3, 4], B = [2, 1, 4, 3] (same multiset, rows shuffled).
//   - Invalid: A = [1, 2, 3, 4], B = [1, 2, 3, 5] (5 ≠ 4; different multisets).
func NewPermutationScenario() *Scenario {
	sys := wiop.NewSystemf("perm-sc")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)
	perm := sys.NewPermutation(
		sys.Context.Childf("perm"),
		[]wiop.Table{wiop.NewTable(colA.View())},
		[]wiop.Table{wiop.NewTable(colB.View())},
	)

	return &Scenario{
		Name:  "Permutation",
		Sys:   sys,
		Query: perm,
		RunHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(colA, makeVec(1, 2, 3, 4))
			rt.AssignColumn(colB, makeVec(2, 1, 4, 3)) // same multiset, different order
		},
		RunInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(colA, makeVec(1, 2, 3, 4))
			rt.AssignColumn(colB, makeVec(1, 2, 3, 5)) // 5 replaces 4 → different multisets
		},
	}
}

// ---- Inclusion ----

// NewInclusionScenario returns a scenario for an Inclusion TableRelation.
//
//   - Valid: A = [1, 1, 2, 2], B = [1, 2, 3, 4] (every A value is in B).
//   - Invalid: A = [1, 1, 2, 5], B = [1, 2, 3, 4] (5 ∉ B).
func NewInclusionScenario() *Scenario {
	sys := wiop.NewSystemf("inc-sc")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	colA := mod.NewColumn(sys.Context.Childf("colA"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("colB"), wiop.VisibilityOracle, r0)
	inc := sys.NewInclusion(
		sys.Context.Childf("inc"),
		[]wiop.Table{wiop.NewTable(colA.View())},
		[]wiop.Table{wiop.NewTable(colB.View())},
	)

	return &Scenario{
		Name:  "Inclusion",
		Sys:   sys,
		Query: inc,
		RunHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(colA, makeVec(1, 1, 2, 2))
			rt.AssignColumn(colB, makeVec(1, 2, 3, 4))
		},
		RunInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(colA, makeVec(1, 1, 2, 5)) // 5 not in B
			rt.AssignColumn(colB, makeVec(1, 2, 3, 4))
		},
	}
}

// ---------------------------------------------------------------------------
// Vanishing scenarios (for compiler integration tests)
// ---------------------------------------------------------------------------

// VanishingScenario is a fixture for testing the global-quotient compiler.
//
// Sys holds the pre-compilation system; AssignHonest / AssignInvalid assign
// only the oracle columns from the initial round. The compiler's prover and
// verifier actions are registered on the rounds that Compile adds.
type VanishingScenario struct {
	// Name identifies the scenario in test output.
	Name string
	// Sys is the pre-compilation System. Call compilers.Compile(sc.Sys) before
	// creating a Runtime; each factory call returns an independent Sys.
	Sys *wiop.System
	// AssignHonest assigns oracle columns with a witness that satisfies all
	// vanishing constraints. After Compile and RunAndVerify, must return nil.
	AssignHonest func(rt *wiop.Runtime)
	// AssignInvalid assigns oracle columns that violate at least one vanishing
	// constraint. After Compile and RunAndVerify, must return an error.
	AssignInvalid func(rt *wiop.Runtime)
}

// VanishingScenarios returns factory functions for the built-in vanishing
// scenarios. The returned Sys is always pre-compilation; call
// compilers.Compile(sc.Sys) in the test before creating a Runtime.
func VanishingScenarios() []func() *VanishingScenario {
	return []func() *VanishingScenario{
		NewBooleanColumnVanishingScenario,
	}
}

// NewBooleanColumnVanishingScenario returns a scenario for the constraint
// col * col − col = 0 (boolean column).
//
//   - Valid: column = [0, 1, 0, 1] — all values are 0 or 1.
//   - Invalid: column = [0, 2, 0, 1] — row 1 is 2, violating col² = col.
func NewBooleanColumnVanishingScenario() *VanishingScenario {
	sys := wiop.NewSystemf("bool-col")
	r0 := sys.NewRound()
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	// col² − col = 0  ⟺  col ∈ {0, 1} on every row.
	mod.NewVanishing(
		sys.Context.Childf("bool"),
		wiop.Sub(wiop.Mul(col.View(), col.View()), col.View()),
	)

	return &VanishingScenario{
		Name: "BooleanColumn",
		Sys:  sys,
		AssignHonest: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, makeVec(0, 1, 0, 1))
		},
		AssignInvalid: func(rt *wiop.Runtime) {
			rt.AssignColumn(col, makeVec(0, 2, 0, 1)) // row 1 is 2: 2²−2 = 2 ≠ 0
		},
	}
}

// ---------------------------------------------------------------------------
// RunAndVerify helper (used by compiler integration tests)
// ---------------------------------------------------------------------------

// RunAndVerify advances a Runtime through all rounds after r0, running prover
// actions immediately after entering each new round. Once at the final round,
// it runs all registered verifier actions across every round and returns the
// first error encountered, or nil.
//
// The caller must assign all r0 oracle columns before calling RunAndVerify.
func RunAndVerify(rt *wiop.Runtime) error {
	sys := rt.System
	for rt.CurrentRound().ID < len(sys.Rounds)-1 {
		rt.AdvanceRound()
		for _, a := range rt.CurrentRound().Actions {
			a.Run(*rt)
		}
	}
	for _, r := range sys.Rounds {
		for _, va := range r.VerifierActions {
			if err := va.Check(*rt); err != nil {
				return err
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// baseVec returns a ConcreteVector of length n where every element equals val.
func baseVec(n int, val uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, n)
	var e field.Element
	e.SetUint64(val)
	for i := range elems {
		elems[i] = e
	}
	return &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
}

// makeVec returns a ConcreteVector from a varargs list of uint64 values.
func makeVec(vals ...uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, len(vals))
	for i, v := range vals {
		elems[i].SetUint64(v)
	}
	return &wiop.ConcreteVector{Plain: []field.FieldVec{field.VecFromBase(elems)}}
}
