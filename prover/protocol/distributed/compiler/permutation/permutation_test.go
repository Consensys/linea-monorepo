package dist_permutation_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	dist_permutation "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestPermutation(t *testing.T) {

	var (
		moduleAName = "moduleA"
	)

	testcases := []struct {
		Name       string
		DefineFunc func(builder *wizard.Builder)
	}{
		{
			Name: "single-column-no-fragment",
			DefineFunc: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("moduleA.A0", 4),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("moduleB.B0", 4),
				}
				c := []ifaces.Column{
					builder.RegisterCommit("moduleC.C0", 4),
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
					builder.RegisterCommit("moduleA.A0", 4),
					builder.RegisterCommit("moduleA.A1", 4),
					builder.RegisterCommit("moduleA.A2", 4),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("moduleB.B0", 4),
					builder.RegisterCommit("moduleB.B1", 4),
					builder.RegisterCommit("moduleB.B2", 4),
				}
				c := []ifaces.Column{
					builder.RegisterCommit("moduleC.C0", 4),
					builder.RegisterCommit("moduleC.C1", 4),
					builder.RegisterCommit("moduleC.C2", 4),
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
						builder.RegisterCommit("moduleA.A00", 4),
						builder.RegisterCommit("moduleA.A10", 4),
						builder.RegisterCommit("moduleA.A20", 4),
					},
					{
						builder.RegisterCommit("moduleA.A01", 4),
						builder.RegisterCommit("moduleA.A11", 4),
						builder.RegisterCommit("moduleA.A21", 4),
					},
				}
				b := [][]ifaces.Column{
					{
						builder.RegisterCommit("moduleB.B00", 4),
						builder.RegisterCommit("moduleB.B10", 4),
						builder.RegisterCommit("moduleB.B20", 4),
					},
					{
						builder.RegisterCommit("moduleB.B01", 4),
						builder.RegisterCommit("moduleB.B11", 4),
						builder.RegisterCommit("moduleB.B21", 4),
					},
				}
				c := [][]ifaces.Column{
					{
						builder.RegisterCommit("moduleC.C00", 4),
						builder.RegisterCommit("moduleC.C10", 4),
						builder.RegisterCommit("moduleC.C20", 4),
					},
					{
						builder.RegisterCommit("moduleC.C01", 4),
						builder.RegisterCommit("moduleC.C11", 4),
						builder.RegisterCommit("moduleC.C21", 4),
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

			// This function assigns the initial module and is aimed at working
			// for all test-case.
			initialProve := func(run *wizard.ProverRuntime) {
				for _, colName := range run.Spec.Columns.AllKeys() {
					run.AssignColumn(colName, smartvectors.ForTest(1, 2, 3, 4))
				}
			}

			// initialComp is defined according to the define function provided by the
			// test-case.
			initialComp := wizard.Compile(tc.DefineFunc)

			disc := namebaseddiscoverer.PeriodSeperatingModuleDiscoverer{}
			disc.Analyze(initialComp)

			// This declares a compiled IOP with only the columns of the module A
			moduleAComp := distributed.GetFreshModuleComp(initialComp, &disc, moduleAName)

			// This distributes the permutation queries
			dist_permutation.NewPermutationIntoGrandProductCtx(
				dist_permutation.Settings{TargetModuleName: moduleAName},
				initialComp, moduleAComp, &disc,
			)

			// This runs the initial prover
			initialRuntime := wizard.RunProver(initialComp, initialProve)

			proof := wizard.Prove(moduleAComp, func(run *wizard.ProverRuntime) {
				run.ParentRuntime = initialRuntime
			})
			valid := wizard.Verify(moduleAComp, proof)
			require.NoError(t, valid)

		})

	}

}
