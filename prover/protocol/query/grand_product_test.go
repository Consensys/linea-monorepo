package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func TestGrandProduct(t *testing.T) {

	var (
		runS *wizard.ProverRuntime
		G    ifaces.Query
	)

	define := func(builder *wizard.Builder) {
		A0 := builder.RegisterCommit("A0", 4)
		B0 := builder.RegisterCommit("B0", 4)
		C0 := builder.RegisterCommit("C0", 4)
		A := []*symbolic.Expression{
			symbolic.Add(A0, B0),
			symbolic.Add(B0, C0),
			symbolic.Add(C0, A0),
		}
		B := []*symbolic.Expression{
			symbolic.Add(C0, A0),
			symbolic.Add(B0, C0),
			symbolic.Add(A0, B0),
		}
		G = builder.CompiledIOP.InsertGrandProduct(0, "G", A, B)
	}

	prove := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("A0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("B0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("C0", smartvectors.ForTest(1, 2, 3, 4))
		runS.AssignGrandProduct("G", field.One())
	}

	var (
		comp = wizard.Compile(define)
		_    = wizard.Prove(comp, prove)
		errG = G.Check(runS)
	)

	if errG != nil {
		t.Fatalf("error verifying the grand product: %v", errG.Error())
	}
}
