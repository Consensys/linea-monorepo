package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestPermutation(t *testing.T) {

	var (
		runS *wizard.ProverRuntime
		P    ifaces.Query
	)

	define := func(builder *wizard.Builder) {
		A := []ifaces.Column{
			builder.RegisterCommit("A0", 4),
			builder.RegisterCommit("A1", 4),
			builder.RegisterCommit("A2", 4),
		}
		B := []ifaces.Column{
			builder.RegisterCommit("B0", 4),
			builder.RegisterCommit("B1", 4),
			builder.RegisterCommit("B2", 4),
		}
		P = builder.CompiledIOP.InsertPermutation(0, "P", A, B)
	}

	prove := func(run *wizard.ProverRuntime) {
		run.AssignColumn("A0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("A1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("B0", smartvectors.ForTest(4, 3, 2, 1))
		run.AssignColumn("B1", smartvectors.ForTest(4, 3, 2, 1))
		run.AssignColumn("B2", smartvectors.ForTest(4, 3, 2, 1))
		runS = run
	}

	var (
		comp = wizard.Compile(define)
		_    = wizard.Prove(comp, prove)
		errP = P.Check(runS)
	)

	if errP != nil {
		t.Fatalf("error verifying the permutation: %v", errP.Error())
	}
}
