package dist_projection_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/distributedprojection"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	dist_projection "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestDistributeProjection(t *testing.T) {
	var (
		moduleAName                    = "moduleA"
		flagSizeA                      = 512
		flagSizeB                      = 256
		flagA, flagB, columnA, columnB ifaces.Column
	)
	testcases := []struct {
		Name       string
		DefineFunc func(builder *wizard.Builder)
	}{
		{
			Name: "distribute-projection-both-A-and-B",
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("moduleA.FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("moduleA.FliterB"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnB"), flagSizeB)
				_ = builder.InsertProjection("ProjectionTest-both-A-and-B",
					query.ProjectionInput{ColumnA: []ifaces.Column{columnA}, ColumnB: []ifaces.Column{columnB}, FilterA: flagA, FilterB: flagB})

			},
		},
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			// This function assigns the initial module and is aimed at working
			// for all test-case.
			initialProve := func(run *wizard.ProverRuntime) {
				// assign filters and columns
				var (
					flagAWit   = make([]field.Element, flagSizeA)
					columnAWit = make([]field.Element, flagSizeA)
					flagBWit   = make([]field.Element, flagSizeB)
					columnBWit = make([]field.Element, flagSizeB)
				)
				for i := 0; i < 10; i++ {
					flagAWit[i] = field.One()
					columnAWit[i] = field.NewElement(uint64(i))
				}
				for i := flagSizeB - 10; i < flagSizeB; i++ {
					flagBWit[i] = field.One()
					columnBWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
				}
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(columnB.GetColID(), smartvectors.RightZeroPadded(columnBWit, flagSizeB))
				run.AssignColumn(columnA.GetColID(), smartvectors.RightZeroPadded(columnAWit, flagSizeA))
			}

			// initialComp is defined according to the define function provided by the
			// test-case.
			initialComp := wizard.Compile(tc.DefineFunc)

			disc := namebaseddiscoverer.PeriodSeperatingModuleDiscoverer{}
			disc.Analyze(initialComp)

			// This declares a compiled IOP with only the columns of the module A
			moduleAComp := distributed.GetFreshModuleComp(initialComp, &disc, moduleAName)
			dist_projection.NewDistributeProjectionCtx(moduleAName, initialComp, moduleAComp, &disc)

			// Compile the distributed projection query
			distributedprojection.CompileDistributedProjection(moduleAComp)

			// This adds a dummy compilation step
			dummy.CompileAtProverLvl(moduleAComp)

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
