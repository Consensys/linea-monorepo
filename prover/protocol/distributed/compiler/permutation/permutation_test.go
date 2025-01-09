package dist_permutation_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	dist_permutation "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/permutation"
	modulediscoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestPermutationAlex(t *testing.T) {

	var (
		moduleAName = "MODULE_A"
		// Initialise the period separating module discoverer
		disc = modulediscoverer.PeriodSeperatingModuleDiscoverer{}
	)

	testcases := []struct {
		Name       string
		DefineFunc func(builder *wizard.Builder)
	}{
		{
			Name: "single-column-no-fragment",
			DefineFunc: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("MODULE_A.A0", 4),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("MODULE_B.B0", 4),
				}
				c := []ifaces.Column{
					builder.RegisterCommit("MODULE_C.C0", 4),
				}
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_A_MOD_B", a, b)
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_C_MOD_A", c, a)
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_B_MOD_C", b, c)
			},
		},
		{
			Name: "multi-column-no-fragment",
			DefineFunc: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("MODULE_A.A0", 4),
					builder.RegisterCommit("MODULE_A.A1", 4),
					builder.RegisterCommit("MODULE_A.A2", 4),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("MODULE_B.B0", 4),
					builder.RegisterCommit("MODULE_B.B1", 4),
					builder.RegisterCommit("MODULE_B.B2", 4),
				}
				c := []ifaces.Column{
					builder.RegisterCommit("MODULE_C.C0", 4),
					builder.RegisterCommit("MODULE_C.C1", 4),
					builder.RegisterCommit("MODULE_C.C2", 4),
				}
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_A_MOD_B", a, b)
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_C_MOD_A", c, a)
				_ = builder.CompiledIOP.InsertPermutation(0, "P_MOD_B_MOD_C", b, c)
			},
		},
		{
			Name: "multi-column-multi-fragment",
			DefineFunc: func(builder *wizard.Builder) {
				a := [][]ifaces.Column{
					{
						builder.RegisterCommit("MODULE_A.A00", 4),
						builder.RegisterCommit("MODULE_A.A10", 4),
						builder.RegisterCommit("MODULE_A.A20", 4),
					},
					{
						builder.RegisterCommit("MODULE_A.A01", 4),
						builder.RegisterCommit("MODULE_A.A11", 4),
						builder.RegisterCommit("MODULE_A.A21", 4),
					},
				}
				b := [][]ifaces.Column{
					{
						builder.RegisterCommit("MODULE_B.B00", 4),
						builder.RegisterCommit("MODULE_B.B10", 4),
						builder.RegisterCommit("MODULE_B.B20", 4),
					},
					{
						builder.RegisterCommit("MODULE_B.B01", 4),
						builder.RegisterCommit("MODULE_B.B11", 4),
						builder.RegisterCommit("MODULE_B.B21", 4),
					},
				}
				c := [][]ifaces.Column{
					{
						builder.RegisterCommit("MODULE_C.C00", 4),
						builder.RegisterCommit("MODULE_C.C10", 4),
						builder.RegisterCommit("MODULE_C.C20", 4),
					},
					{
						builder.RegisterCommit("MODULE_C.C01", 4),
						builder.RegisterCommit("MODULE_C.C11", 4),
						builder.RegisterCommit("MODULE_C.C21", 4),
					},
				}
				_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_A_MOD_B", a, b)
				_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_C_MOD_A", c, a)
				_ = builder.CompiledIOP.InsertFragmentedPermutation(0, "P_MOD_B_MOD_C", b, c)
			},
		},
	}

	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {

			initialComp := wizard.Compile(tc.DefineFunc)

			disc.Analyze(initialComp)

			moduleAComp := wizard.Compile(func(build *wizard.Builder) {

				for _, colName := range initialComp.Columns.AllKeys() {

					col := initialComp.Columns.GetHandle(colName)
					if !disc.ColumnIsInModule(col, moduleAName) {
						continue
					}

					build.RegisterCommit(col.GetColID(), col.Size())
				}
			}, dummy.CompileAtProverLvl)

			var (
				_ = dist_permutation.NewPermutationIntoGrandProductCtx(
					dist_permutation.Settings{TargetModuleName: moduleAName},
					initialComp, moduleAComp, &disc,
				)
				initialRun *wizard.ProverRuntime
			)

			initialProve := func(run *wizard.ProverRuntime) {
				for _, colName := range run.Spec.Columns.AllKeys() {
					run.AssignColumn(colName, smartvectors.ForTest(1, 2, 3, 4))
				}

				initialRun = run
			}

			_ = wizard.Prove(initialComp, initialProve)

			moduleAProve := func(run *wizard.ProverRuntime) {
				for _, colName := range initialComp.Columns.AllKeys() {

					col := initialComp.Columns.GetHandle(colName)
					if !disc.ColumnIsInModule(col, moduleAName) {
						continue
					}

					c := initialRun.GetColumn(colName)
					run.AssignColumn(colName, c)
				}
			}

			_ = wizard.Prove(moduleAComp, moduleAProve)

		})

	}

}
