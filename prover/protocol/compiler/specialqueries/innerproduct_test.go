package specialqueries

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func TestSimpleInnerProduct(t *testing.T) {

	testcases := []map[string]smartvectors.SmartVector{
		{
			"A": smartvectors.NewConstant(field.NewElement(5), 16),
			"B": smartvectors.NewConstant(field.NewElement(4), 16),
		},
		{
			"A": smartvectors.Rand(16),
			"B": smartvectors.Rand(16),
		},
	}

	definer := func(build *wizard.Builder) {
		a := build.RegisterCommit("A", 16)
		b := build.RegisterCommit("B", 16)
		_ = build.InnerProduct("IP", a, b)
	}

	for i, testcase := range testcases {

		t.Logf("Running test #%v", i)
		expectedValue := smartvectors.InnerProduct(testcase["A"], testcase["B"])

		prover := func(run *wizard.ProverRuntime) {
			run.AssignColumn("A", testcase["A"])
			run.AssignColumn("B", testcase["B"])
			run.AssignInnerProduct("IP", expectedValue)
		}

		compiled := wizard.Compile(definer, CompileInnerProduct, dummy.Compile)
		proof := wizard.Prove(compiled, prover)
		if err := wizard.Verify(compiled, proof); err != nil {
			panic(err)
		}
	}
}

func TestTwoInnerProductSameSize(t *testing.T) {

	definer := func(build *wizard.Builder) {
		a1 := build.RegisterCommit("A1", 16)
		b1 := build.RegisterCommit("B1", 16)
		a2 := build.RegisterCommit("A2", 16)
		b2 := build.RegisterCommit("B2", 16)
		_ = build.InnerProduct("IP1", a1, b1)
		_ = build.InnerProduct("IP2", a2, b2)
	}

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("A1", smartvectors.NewConstant(field.NewElement(5), 16))
		run.AssignColumn("B1", smartvectors.NewConstant(field.NewElement(4), 16))
		run.AssignColumn("A2", smartvectors.NewConstant(field.NewElement(3), 16))
		run.AssignColumn("B2", smartvectors.NewConstant(field.NewElement(2), 16))
		run.AssignInnerProduct("IP1", field.NewElement(16*5*4))
		run.AssignInnerProduct("IP2", field.NewElement(16*3*2))
	}

	compiled := wizard.Compile(definer, CompileInnerProduct, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	if err := wizard.Verify(compiled, proof); err != nil {
		panic(err)
	}
}

func TestTwoInnerProductMixedSize(t *testing.T) {

	definer := func(build *wizard.Builder) {
		a1 := build.RegisterCommit("A1", 16)
		b1 := build.RegisterCommit("B1", 16)
		a2 := build.RegisterCommit("A2", 32)
		b2 := build.RegisterCommit("B2", 32)
		_ = build.InnerProduct("IP1", a1, b1)
		_ = build.InnerProduct("IP2", a2, b2)
	}

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("A1", smartvectors.NewConstant(field.NewElement(5), 16))
		run.AssignColumn("B1", smartvectors.NewConstant(field.NewElement(4), 16))
		run.AssignColumn("A2", smartvectors.NewConstant(field.NewElement(3), 32))
		run.AssignColumn("B2", smartvectors.NewConstant(field.NewElement(2), 32))
		run.AssignInnerProduct("IP1", field.NewElement(16*5*4))
		run.AssignInnerProduct("IP2", field.NewElement(32*3*2))
	}

	compiled := wizard.Compile(definer, CompileInnerProduct, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	if err := wizard.Verify(compiled, proof); err != nil {
		panic(err)
	}
}

func TestInnerProductMixedSize(t *testing.T) {

	definer := func(build *wizard.Builder) {
		a1_16 := build.RegisterCommit("A1_16", 16)
		b1_16 := build.RegisterCommit("B1_16", 16)
		a2_16 := build.RegisterCommit("A2_16", 16)
		b2_16 := build.RegisterCommit("B2_16", 16)
		a1_32 := build.RegisterCommit("A1_32", 32)
		b1_32 := build.RegisterCommit("B1_32", 32)
		a2_32 := build.RegisterCommit("A2_32", 32)
		b2_32 := build.RegisterCommit("B2_32", 32)
		_ = build.InnerProduct("IP1_16", a1_16, b1_16)
		_ = build.InnerProduct("IP2_16", a2_16, b2_16)
		_ = build.InnerProduct("IP1_32", a1_32, b1_32)
		_ = build.InnerProduct("IP2_32", a2_32, b2_32)
	}

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("A1_16", smartvectors.NewConstant(field.NewElement(5), 16))
		run.AssignColumn("B1_16", smartvectors.NewConstant(field.NewElement(4), 16))
		run.AssignColumn("A2_16", smartvectors.NewConstant(field.NewElement(3), 16))
		run.AssignColumn("B2_16", smartvectors.NewConstant(field.NewElement(2), 16))
		run.AssignColumn("A1_32", smartvectors.NewConstant(field.NewElement(5), 32))
		run.AssignColumn("B1_32", smartvectors.NewConstant(field.NewElement(4), 32))
		run.AssignColumn("A2_32", smartvectors.NewConstant(field.NewElement(3), 32))
		run.AssignColumn("B2_32", smartvectors.NewConstant(field.NewElement(2), 32))
		run.AssignInnerProduct("IP1_16", field.NewElement(16*5*4))
		run.AssignInnerProduct("IP2_16", field.NewElement(16*3*2))
		run.AssignInnerProduct("IP1_32", field.NewElement(32*5*4))
		run.AssignInnerProduct("IP2_32", field.NewElement(32*3*2))
	}

	compiled := wizard.Compile(definer, CompileInnerProduct, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	if err := wizard.Verify(compiled, proof); err != nil {
		panic(err)
	}
}
