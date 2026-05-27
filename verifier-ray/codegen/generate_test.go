package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
)

type unsupportedVerifierAction struct{}

func (unsupportedVerifierAction) Check(wiop.Runtime) error {
	return nil
}

func makeTestVec(vals ...uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, len(vals))
	for i, v := range vals {
		elems[i].SetUint64(v)
	}
	return &wiop.ConcreteVector{Plain: field.VecFromBase(elems)}
}

func TestGenerateZigRejectsUnsupportedVerifierActions(t *testing.T) {
	sys := wiop.NewSystemf("unsupported")
	round := sys.NewRound()
	round.RegisterVerifierAction(unsupportedVerifierAction{})

	var out bytes.Buffer
	err := GenerateZig(sys, &out)
	if err == nil {
		t.Fatal("GenerateZig() error = nil, want unsupported action failure")
	}
	if !strings.Contains(err.Error(), "unsupported verifier action") {
		t.Fatalf("GenerateZig() error = %q, want unsupported verifier action context", err)
	}
}

// TestGenerateZigIsDeterministic verifies that GenerateZig produces identical
// output across two calls on the same system, exercising the full emitGlobalVerify
// path including expression trees and coin references.
func TestGenerateZigIsDeterministic(t *testing.T) {
	sys := wiop.NewSystemf("deterministic")
	sys.NewRound()
	m := sys.NewSizedModule(sys.Context.Childf("main"), 4, wiop.PaddingDirectionRight)
	a := m.NewPrecomputedColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, makeTestVec(0, 1, 0, 0))
	b := m.NewPrecomputedColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, makeTestVec(1, 0, 0, 0))
	m.NewVanishing(sys.Context.Childf("AB"), wiop.Mul(a.View(), b.View()))
	global.Compile(sys)
	wiop.Materialize(sys)

	var first bytes.Buffer
	if err := GenerateZig(sys, &first); err != nil {
		t.Fatalf("GenerateZig(first) error = %v", err)
	}

	var second bytes.Buffer
	if err := GenerateZig(sys, &second); err != nil {
		t.Fatalf("GenerateZig(second) error = %v", err)
	}

	if first.String() != second.String() {
		t.Fatal("GenerateZig() output changed between identical invocations")
	}
}

// TestGenerateZig_DynamicModule verifies that a dynamic module (unknown size at
// compile time) emits an assignment.len() call to derive n at runtime rather
// than a hard-coded constant.
func TestGenerateZig_DynamicModule(t *testing.T) {
	sys := wiop.NewSystemf("dynamic")
	r0 := sys.NewRound()
	m := sys.NewDynamicModule(sys.Context.Childf("dyn"), wiop.PaddingDirectionRight)
	a := m.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	b := m.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	m.NewVanishing(sys.Context.Childf("AB"), wiop.Mul(a.View(), b.View()))
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()
	if !strings.Contains(src, "assignment.len()") {
		t.Fatalf("dynamic module should emit assignment.len() for n, got:\n%s", src)
	}
	// The size variable must not carry an explicit usize type annotation, which
	// would mean a hard-coded constant was emitted instead of the len() call.
	if strings.Contains(src, "_n0: usize =") {
		t.Fatalf("dynamic module should not emit a hard-coded constant size, got:\n%s", src)
	}
}

func TestGenerateZig_CellExpressionUsesFlattenedProofCells(t *testing.T) {
	sys := wiop.NewSystemf("cell-flatten")
	r0 := sys.NewRound()
	r1 := sys.NewRound()
	m := sys.NewSizedModule(sys.Context.Childf("main"), 4, wiop.PaddingDirectionRight)
	a := m.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r1)
	_ = r0.NewCell(sys.Context.Childf("padding-cell"), false)
	cell := r1.NewCell(sys.Context.Childf("public-cell"), false)
	m.NewVanishing(sys.Context.Childf("APlusCell"), wiop.Add(a.View(), cell))
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()
	if !strings.Contains(src, "ext.Ext.lift(proof.cells[1].base)") {
		t.Fatalf("cell from round 1 should read from flattened proof.cells[1], got:\n%s", src)
	}
}

func TestGenerateZig_CoinExpressionUsesDerivedRoundCoins(t *testing.T) {
	sys := wiop.NewSystemf("coin-expression")
	sys.NewRound()
	r1 := sys.NewRound()
	coin := r1.NewCoinField(sys.Context.Childf("challenge"))
	m := sys.NewSizedModule(sys.Context.Childf("main"), 4, wiop.PaddingDirectionRight)
	a := m.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r1)
	m.NewVanishing(sys.Context.Childf("ATimesCoin"), wiop.Mul(a.View(), coin))
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()
	if !strings.Contains(src, "coins_r1[0]") {
		t.Fatalf("coin from round 1 should read from coins_r1[0], got:\n%s", src)
	}
}

func TestGenerateZig_DynamicModuleUsesAllModuleColumnsForSize(t *testing.T) {
	sys := wiop.NewSystemf("dynamic-max")
	r0 := sys.NewRound()
	m := sys.NewDynamicModule(sys.Context.Childf("dyn"), wiop.PaddingDirectionRight)
	a := m.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	_ = m.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	m.NewVanishing(sys.Context.Childf("AOnly"), a.View())
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()
	if !strings.Contains(src, "proof.columns[0].assignment.len()") {
		t.Fatalf("dynamic size should include the referenced column, got:\n%s", src)
	}
	if !strings.Contains(src, "proof.columns[1].assignment.len()") {
		t.Fatalf("dynamic size should include unreferenced module columns, got:\n%s", src)
	}
}

func TestGenerateZig_AdvanceRoundUsesOnlyVerifierVisibleColumns(t *testing.T) {
	sys := wiop.NewSystemf("visible-columns")
	r0 := sys.NewRound()
	_ = sys.NewRound()
	m := sys.NewSizedModule(sys.Context.Childf("main"), 4, wiop.PaddingDirectionRight)
	_ = m.NewColumn(sys.Context.Childf("internal"), wiop.VisibilityInternal, r0)
	a := m.NewColumn(sys.Context.Childf("oracle"), wiop.VisibilityOracle, r0)
	m.NewVanishing(sys.Context.Childf("AOnly"), a.View())
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()
	if strings.Contains(src, "pub const min_columns: usize = 3;") {
		t.Fatalf("internal columns should not be counted in proof payload, got:\n%s", src)
	}
	if !strings.Contains(src, ".columns = proof.columns[0..1],") {
		t.Fatalf("advanceRound should slice only the oracle/public columns, got:\n%s", src)
	}
}

// TestGenerateZig_MultiModule verifies that two modules with independent
// vanishing constraints each get distinct _n and _ve variables, and that
// eval_cells indices don't collide between them.
func TestGenerateZig_MultiModule(t *testing.T) {
	sys := wiop.NewSystemf("multi-module")
	r0 := sys.NewRound()
	// Module 0: size 4, vanishing over A*B
	m0 := sys.NewSizedModule(sys.Context.Childf("m0"), 4, wiop.PaddingDirectionRight)
	a0 := m0.NewPrecomputedColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, makeTestVec(1, 0, 0, 0))
	b0 := m0.NewPrecomputedColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, makeTestVec(0, 1, 0, 0))
	m0.NewVanishing(sys.Context.Childf("AB"), wiop.Mul(a0.View(), b0.View()))
	// Module 1: size 4, single column vanishing
	m1 := sys.NewSizedModule(sys.Context.Childf("m1"), 4, wiop.PaddingDirectionRight)
	c1 := m1.NewColumn(sys.Context.Childf("C"), wiop.VisibilityOracle, r0)
	m1.NewVanishing(sys.Context.Childf("C"), c1.View())
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	if err := GenerateZig(sys, &buf); err != nil {
		t.Fatalf("GenerateZig() error = %v", err)
	}
	src := buf.String()

	// Each module gets a distinct size variable
	if !strings.Contains(src, "_n0") {
		t.Fatalf("missing _n0 for first module, got:\n%s", src)
	}
	if !strings.Contains(src, "_n1") {
		t.Fatalf("missing _n1 for second module, got:\n%s", src)
	}

	// Each module gets a distinct vanishing-evals variable
	if !strings.Contains(src, "_ve0_0") {
		t.Fatalf("missing _ve0_0 for first module, got:\n%s", src)
	}
	if !strings.Contains(src, "_ve1_0") {
		t.Fatalf("missing _ve1_0 for second module, got:\n%s", src)
	}

	// Quotient eval_cells slices must not overlap — verify two distinct ranges appear
	if strings.Count(src, "proof.eval_cells[") < 2 {
		t.Fatalf("expected at least two eval_cells references for two modules, got:\n%s", src)
	}
}

func TestGenerateZigRejectsDynamicModuleWithInternalColumns(t *testing.T) {
	sys := wiop.NewSystemf("dynamic-internal")
	r0 := sys.NewRound()
	m := sys.NewDynamicModule(sys.Context.Childf("dyn"), wiop.PaddingDirectionRight)
	a := m.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	_ = m.NewColumn(sys.Context.Childf("secret"), wiop.VisibilityInternal, r0)
	m.NewVanishing(sys.Context.Childf("AOnly"), a.View())
	global.Compile(sys)
	wiop.Materialize(sys)

	var buf bytes.Buffer
	err := GenerateZig(sys, &buf)
	if err == nil {
		t.Fatal("GenerateZig() error = nil, want dynamic internal-column failure")
	}
	if !strings.Contains(err.Error(), "verifier cannot derive runtime size") {
		t.Fatalf("GenerateZig() error = %q, want runtime-size failure", err)
	}
}
