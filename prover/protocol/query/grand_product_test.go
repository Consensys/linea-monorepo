package query_test

import (
	"testing"

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
		A := []*symbolic.Expression{
			symbolic.NewConstant(1),
			symbolic.NewConstant(2),
			symbolic.NewConstant(3),
			symbolic.NewConstant(4),
		}
		B := []*symbolic.Expression{
			symbolic.NewConstant(4),
			symbolic.NewConstant(3),
			symbolic.NewConstant(2),
			symbolic.NewConstant(1),
		}
		G = builder.CompiledIOP.InsertGrandProduct(0, "G", A, B)
	}

	prove := func(run *wizard.ProverRuntime) {
		runS = run
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
