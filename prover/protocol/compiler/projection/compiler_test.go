package projection_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

func makeTestCaseProjection() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	flagSizeA := 512
	flagSizeB := 256
	var flagA, flagB, columnA, columnB ifaces.Column
	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		flagA = comp.InsertCommit(round, ifaces.ColID("FilterA"), flagSizeA)
		flagB = comp.InsertCommit(round, ifaces.ColID("FliterB"), flagSizeB)
		columnA = comp.InsertCommit(round, ifaces.ColID("ColumnA"), flagSizeA)
		columnB = comp.InsertCommit(round, ifaces.ColID("ColumnB"), flagSizeB)
		comp.InsertProjection("Projection_Compilation_Test", query.ProjectionInput{ColumnA: []ifaces.Column{columnA}, ColumnB: []ifaces.Column{columnB}, FilterA: flagA, FilterB: flagB})

	}
	prover = func(run *wizard.ProverRuntime) {
		// assign filters and columns
		flagAWit := make([]field.Element, flagSizeA)
		columnAWit := make([]field.Element, flagSizeA)
		flagBWit := make([]field.Element, flagSizeB)
		columnBWit := make([]field.Element, flagSizeB)
		for i := 0; i < 10; i++ {
			flagAWit[i] = field.One()
			columnAWit[i] = field.NewElement(uint64(i))
		}
		for i := flagSizeB - 10; i < flagSizeB; i++ {
			flagBWit[i] = field.One()
			columnBWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
		}
		run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
		run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
		run.AssignColumn(columnB.GetColID(), smartvectors.RightZeroPadded(columnBWit, flagSizeB))
		run.AssignColumn(columnA.GetColID(), smartvectors.RightZeroPadded(columnAWit, flagSizeA))
	}
	return define, prover
}

func TestProjectionQuery(t *testing.T) {

	define, prover := makeTestCaseProjection()
	comp := wizard.Compile(define, projection.CompileProjection, dummy.CompileAtProverLvl)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
