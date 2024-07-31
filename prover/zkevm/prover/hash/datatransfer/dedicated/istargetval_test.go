package dedicated

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

func makeTestCaseIsTargetValue() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	size := 8
	var colA, colB ifaces.Column
	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		colA = comp.InsertCommit(round, ifaces.ColIDf("ColA"), size)
		colB = comp.InsertCommit(round, ifaces.ColIDf("ColB"), size)
		InsertIsTargetValue(comp, round, ifaces.QueryIDf("IsTarget"), field.NewElement(3), colA, colB)
	}
	prover = func(run *wizard.ProverRuntime) {
		cola := vector.ForTest(3, 4, 5, 9, 1, 9, 7, 3)
		colb := vector.ForTest(1, 0, 0, 0, 0, 0, 0, 1)
		run.AssignColumn(colA.GetColID(), smartvectors.NewRegular(cola))
		run.AssignColumn(colB.GetColID(), smartvectors.NewRegular(colb))
	}
	return define, prover
}
func TestIsZeroIff(t *testing.T) {
	define, prover := makeTestCaseIsTargetValue()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
