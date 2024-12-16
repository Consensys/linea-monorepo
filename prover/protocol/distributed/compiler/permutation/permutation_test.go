package dist_permutation_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	dist_permutation "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestDistPermutation(t *testing.T) {
	var (
		runS *wizard.ProverRuntime
		G    ifaces.Query
	)

	define := func(builder *wizard.Builder) {
		A := []ifaces.Column{
			builder.RegisterCommit("modExp.A0", 4),
			builder.RegisterCommit("modExp.A1", 4),
			builder.RegisterCommit("modExp.A2", 4),
		}
		B := []ifaces.Column{
			builder.RegisterCommit("rom.B0", 4),
			builder.RegisterCommit("rom.B1", 4),
			builder.RegisterCommit("rom.B2", 4),
		}
		C := []ifaces.Column{
			builder.RegisterCommit("romLex.C0", 4),
			builder.RegisterCommit("romLex.C1", 4),
			builder.RegisterCommit("romLex.C2", 4),
		}
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MODEXP_ROM", A, B)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_ROMLEX_MODEXP", C, A)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_ROM_ROMLEX", B, C)
	}

	initialComp := wizard.Compile(define)
	moduleCompModExp := wizard.Compile(define)
	G = dist_permutation.AddGdProductQuery(initialComp, moduleCompModExp, "modExp", dist_permutation.Settings{MaxNumOfQueryPerModule: 4})
	prove := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("modExp.A0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("modExp.A1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("modExp.A2", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("rom.B0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("rom.B1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("rom.B2", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("romLex.C0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("romLex.C1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("romLex.C2", smartvectors.ForTest(1, 2, 3, 4))
		runS.AssignGrandProduct("modExp_GRAND_PRODUCT", field.One())
	}
	_ = wizard.Prove(initialComp, prove)
	_ = wizard.Prove(moduleCompModExp, prove)
	errG := G.Check(runS)

	if errG != nil {
		t.Fatalf("error verifying the grand product: %v", errG.Error())
	}

}
