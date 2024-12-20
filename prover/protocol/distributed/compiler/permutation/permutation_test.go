package dist_permutation_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	dist_permutation "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestDistPermutationNoMultiColumnNoFragment(t *testing.T) {
	var (
		runS    *wizard.ProverRuntime
		G       ifaces.Query
		permCtx *dist_permutation.PermutationIntoGrandProductCtx
	)
	permCtx = dist_permutation.NewPermutationIntoGrandProductCtx(dist_permutation.Settings{MaxNumOfQueryPerModule: 4})
	initialDefine := func(builder *wizard.Builder) {
		A := []ifaces.Column{
			builder.RegisterCommit("MODULE_A.A0", 4),
		}
		B := []ifaces.Column{
			builder.RegisterCommit("MODULE_B.B0", 4),
		}
		C := []ifaces.Column{
			builder.RegisterCommit("MODULE_C.C0", 4),
		}
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_A_MOD_B", A, B)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_C_MOD_A", C, A)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_B_MOD_C", B, C)
	}

	moduleADefine := func(builder *wizard.Builder) {
		builder.RegisterCommit("MODULE_A.A0", 4)
	}

	initialComp := wizard.Compile(initialDefine)
	moduleAComp := wizard.Compile(moduleADefine)

	moduleAProve := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("MODULE_A.A0", smartvectors.ForTest(1, 2, 3, 4))
		G = permCtx.AddGdProductQuery(initialComp, moduleAComp, "MODULE_A", run)
	}
	_ = wizard.Prove(moduleAComp, moduleAProve)
	errG := G.Check(runS)

	if errG != nil {
		t.Fatalf("error verifying the grand product: %v", errG.Error())
	}
}

func TestDistPermutationNoFragment(t *testing.T) {
	var (
		runS    *wizard.ProverRuntime
		G       ifaces.Query
		permCtx *dist_permutation.PermutationIntoGrandProductCtx
	)
	permCtx = dist_permutation.NewPermutationIntoGrandProductCtx(dist_permutation.Settings{MaxNumOfQueryPerModule: 4})
	initialDefine := func(builder *wizard.Builder) {
		A := []ifaces.Column{
			builder.RegisterCommit("MODULE_A.A0", 4),
			builder.RegisterCommit("MODULE_A.A1", 4),
			builder.RegisterCommit("MODULE_A.A2", 4),
		}
		B := []ifaces.Column{
			builder.RegisterCommit("MODULE_B.B0", 4),
			builder.RegisterCommit("MODULE_B.B1", 4),
			builder.RegisterCommit("MODULE_B.B2", 4),
		}
		C := []ifaces.Column{
			builder.RegisterCommit("MODULE_C.C0", 4),
			builder.RegisterCommit("MODULE_C.C1", 4),
			builder.RegisterCommit("MODULE_C.C2", 4),
		}
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_A_MOD_B", A, B)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_C_MOD_A", C, A)
		_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_B_MOD_C", B, C)
	}

	moduleADefine := func(builder *wizard.Builder) {
		builder.RegisterCommit("MODULE_A.A0", 4)
		builder.RegisterCommit("MODULE_A.A1", 4)
		builder.RegisterCommit("MODULE_A.A2", 4)
	}

	initialComp := wizard.Compile(initialDefine)
	moduleAComp := wizard.Compile(moduleADefine)

	moduleAProve := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("MODULE_A.A0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A2", smartvectors.ForTest(1, 2, 3, 4))
		G = permCtx.AddGdProductQuery(initialComp, moduleAComp, "MODULE_A", run)
	}
	_ = wizard.Prove(moduleAComp, moduleAProve)
	errG := G.Check(runS)

	if errG != nil {
		t.Fatalf("error verifying the grand product: %v", errG.Error())
	}
}

func TestDistPermutationFragment(t *testing.T) {
	var (
		runS    *wizard.ProverRuntime
		G       ifaces.Query
		permCtx *dist_permutation.PermutationIntoGrandProductCtx
	)
	permCtx = dist_permutation.NewPermutationIntoGrandProductCtx(dist_permutation.Settings{MaxNumOfQueryPerModule: 4})
	initialDefine := func(builder *wizard.Builder) {
		A := [][]ifaces.Column{
			{builder.RegisterCommit("MODULE_A.A00", 4),
				builder.RegisterCommit("MODULE_A.A10", 4),
				builder.RegisterCommit("MODULE_A.A20", 4)},
			{builder.RegisterCommit("MODULE_A.A01", 4),
				builder.RegisterCommit("MODULE_A.A11", 4),
				builder.RegisterCommit("MODULE_A.A21", 4)},
		}
		B := [][]ifaces.Column{
			{builder.RegisterCommit("MODULE_B.B00", 4),
				builder.RegisterCommit("MODULE_B.B10", 4),
				builder.RegisterCommit("MODULE_B.B20", 4)},
			{builder.RegisterCommit("MODULE_B.B01", 4),
				builder.RegisterCommit("MODULE_B.B11", 4),
				builder.RegisterCommit("MODULE_B.B21", 4)},
		}
		C := [][]ifaces.Column{
			{builder.RegisterCommit("MODULE_C.C00", 4),
				builder.RegisterCommit("MODULE_C.C10", 4),
				builder.RegisterCommit("MODULE_C.C20", 4)},
			{builder.RegisterCommit("MODULE_C.C01", 4),
				builder.RegisterCommit("MODULE_C.C11", 4),
				builder.RegisterCommit("MODULE_C.C21", 4)},
		}
		_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_A_MOD_B", A, B)
		_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_C_MOD_A", C, A)
		_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_B_MOD_C", B, C)
	}

	moduleADefine := func(builder *wizard.Builder) {
		builder.RegisterCommit("MODULE_A.A00", 4)
		builder.RegisterCommit("MODULE_A.A10", 4)
		builder.RegisterCommit("MODULE_A.A20", 4)
		builder.RegisterCommit("MODULE_A.A01", 4)
		builder.RegisterCommit("MODULE_A.A11", 4)
		builder.RegisterCommit("MODULE_A.A21", 4)
	}

	initialComp := wizard.Compile(initialDefine)
	moduleAComp := wizard.Compile(moduleADefine)

	moduleAProve := func(run *wizard.ProverRuntime) {
		runS = run
		run.AssignColumn("MODULE_A.A00", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A10", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A20", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A01", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A11", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("MODULE_A.A21", smartvectors.ForTest(1, 2, 3, 4))
		G = permCtx.AddGdProductQuery(initialComp, moduleAComp, "MODULE_A", run)
	}
	_ = wizard.Prove(moduleAComp, moduleAProve)
	errG := G.Check(runS)

	if errG != nil {
		t.Fatalf("error verifying the grand product: %v", errG.Error())
	}
}
