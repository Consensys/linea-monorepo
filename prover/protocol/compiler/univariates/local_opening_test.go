package univariates

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

func TestBasicLocalOpening(t *testing.T) {
	testLocalOpening(t, basicLocalOpening)
}

func basicLocalOpening() (wizard.DefineFunc, wizard.MainProverStep) {

	n := 16

	definer := func(build *wizard.Builder) {
		p1 := build.RegisterCommit("P1", n)
		p2 := build.RegisterCommit("P2", n)

		build.LocalOpening("O1", column.Shift(p1, 1))
		build.LocalOpening("O2", column.Shift(p2, -1))
	}

	prover := func(run *wizard.ProverRuntime) {
		p1 := smartvectors.Rand(n)
		p2 := smartvectors.Rand(n)

		run.AssignColumn("P1", p1)
		run.AssignColumn("P2", p2)

		run.AssignLocalPoint("O1", p1.Get(1))
		run.AssignLocalPoint("O2", p2.Get(n-1))
	}

	return definer, prover
}

func testLocalOpening(t *testing.T, gen func() (wizard.DefineFunc, wizard.MainProverStep)) {

	builder, prover := gen()
	comp := wizard.Compile(builder, CompileLocalOpening, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)

	assert.NoError(t, err)

	for _, qName := range comp.QueriesNoParams.AllKeysAt(0) {
		switch q := comp.QueriesParams.Data(qName).(type) {
		case query.LocalOpening:
			t.Logf("query %v - with pol %v", q.ID, q.Pol.GetColID())
		case query.UnivariateEval:
			t.Logf("query %v - with pols %v", q.QueryID, q.Pols)
		}
	}
}
