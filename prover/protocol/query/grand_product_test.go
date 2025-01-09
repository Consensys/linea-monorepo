package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
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
		alpha := builder.RegisterRandomCoin("ALPHA", coin.Field)
		A := []*symbolic.Expression{
			symbolic.Add(A0, symbolic.Mul(B0, alpha)),
			symbolic.Add(B0, symbolic.Mul(C0, alpha)),
			symbolic.Add(C0, symbolic.Mul(A0, alpha)),
		}
		B := []*symbolic.Expression{
			symbolic.Add(C0, symbolic.Mul(A0, alpha)),
			symbolic.Add(B0, symbolic.Mul(C0, alpha)),
			symbolic.Add(A0, symbolic.Mul(B0, alpha)),
		}
		key := [2]int{0, 0}
		zCat1 := map[[2]int]*query.GrandProductInput{}
		zCat1[key] = &query.GrandProductInput{
			Numerators:   A,
			Denominators: B,
		}
		G = builder.CompiledIOP.InsertGrandProduct(0, "GrandProductTest", zCat1)
	}

	prove := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("A0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("B0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("C0", smartvectors.ForTest(1, 2, 3, 4))
		runS.AssignGrandProduct("GrandProductTest", field.One())
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
