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
		moduleAName                                                   = "moduleA"
		flagSizeA                                                     = 512
		flagSizeB                                                     = 256
		flagA, flagB, columnA, columnB, flagC, columnC                ifaces.Column
		colA0, colA1, colA2, colB0, colB1, colB2, colC0, colC1, colC2 ifaces.Column
	)
	testcases := []struct {
		Name              string
		DefineFunc        func(builder *wizard.Builder)
		InitialProverFunc func(run *wizard.ProverRuntime)
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
			InitialProverFunc: func(run *wizard.ProverRuntime) {
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
			},
		},
		{
			Name: "distribute-projection-multiple_projections",
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("moduleA.FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("moduleB.FliterB"), flagSizeB)
				flagC = builder.RegisterCommit(ifaces.ColID("moduleC.FliterB"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB"), flagSizeB)
				columnC = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC"), flagSizeB)
				_ = builder.InsertProjection("ProjectionTest-A-B",
					query.ProjectionInput{ColumnA: []ifaces.Column{columnA}, ColumnB: []ifaces.Column{columnB}, FilterA: flagA, FilterB: flagB})
				_ = builder.InsertProjection("ProjectionTest-C-A",
					query.ProjectionInput{ColumnA: []ifaces.Column{columnC}, ColumnB: []ifaces.Column{columnA}, FilterA: flagC, FilterB: flagA})

			},
			InitialProverFunc: func(run *wizard.ProverRuntime) {
				// assign filters and columns
				var (
					flagAWit   = make([]field.Element, flagSizeA)
					columnAWit = make([]field.Element, flagSizeA)
					flagBWit   = make([]field.Element, flagSizeB)
					columnBWit = make([]field.Element, flagSizeB)
					flagCWit   = make([]field.Element, flagSizeB)
					columnCWit = make([]field.Element, flagSizeB)
				)
				for i := 0; i < 10; i++ {
					flagAWit[i] = field.One()
					columnAWit[i] = field.NewElement(uint64(i))
				}
				for i := flagSizeB - 10; i < flagSizeB; i++ {
					flagBWit[i] = field.One()
					flagCWit[i] = field.One()
					columnBWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					columnCWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
				}
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(flagC.GetColID(), smartvectors.RightZeroPadded(flagCWit, flagSizeB))
				run.AssignColumn(columnB.GetColID(), smartvectors.RightZeroPadded(columnBWit, flagSizeB))
				run.AssignColumn(columnA.GetColID(), smartvectors.RightZeroPadded(columnAWit, flagSizeA))
				run.AssignColumn(columnC.GetColID(), smartvectors.RightZeroPadded(columnCWit, flagSizeB))
			},
		},
		{
			Name: "distribute-projection-multiple_projections-multi-columns",
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("moduleA.FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("moduleB.FliterB"), flagSizeB)
				flagC = builder.RegisterCommit(ifaces.ColID("moduleC.FliterB"), flagSizeB)
				colA0 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA0"), flagSizeA)
				colA1 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA1"), flagSizeA)
				colA2 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA2"), flagSizeA)
				colB0 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB0"), flagSizeB)
				colB1 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB1"), flagSizeB)
				colB2 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB2"), flagSizeB)
				colC0 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC0"), flagSizeB)
				colC1 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC1"), flagSizeB)
				colC2 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC2"), flagSizeB)
				_ = builder.InsertProjection("ProjectionTest-A-B-Multicolum",
					query.ProjectionInput{ColumnA: []ifaces.Column{colA0, colA1, colA2}, ColumnB: []ifaces.Column{colB0, colB1, colB2}, FilterA: flagA, FilterB: flagB})
				_ = builder.InsertProjection("ProjectionTest-C-A-Multicolumn",
					query.ProjectionInput{ColumnA: []ifaces.Column{colC0, colC1, colC2}, ColumnB: []ifaces.Column{colA0, colA1, colA2}, FilterA: flagC, FilterB: flagA})

			},
			InitialProverFunc: func(run *wizard.ProverRuntime) {
				// assign filters and columns
				var (
					flagAWit = make([]field.Element, flagSizeA)
					flagBWit = make([]field.Element, flagSizeB)
					flagCWit = make([]field.Element, flagSizeB)
					colA0Wit = make([]field.Element, flagSizeA)
					colA1Wit = make([]field.Element, flagSizeA)
					colA2Wit = make([]field.Element, flagSizeA)
					colB0Wit = make([]field.Element, flagSizeB)
					colB1Wit = make([]field.Element, flagSizeB)
					colB2Wit = make([]field.Element, flagSizeB)
					colC0Wit = make([]field.Element, flagSizeB)
					colC1Wit = make([]field.Element, flagSizeB)
					colC2Wit = make([]field.Element, flagSizeB)
				)
				for i := 0; i < 10; i++ {
					flagAWit[i] = field.One()
					colA0Wit[i] = field.NewElement(uint64(i))
					colA1Wit[i] = field.NewElement(uint64(i + 1))
					colA2Wit[i] = field.NewElement(uint64(i + 2))
				}
				for i := flagSizeB - 10; i < flagSizeB; i++ {
					flagBWit[i] = field.One()
					flagCWit[i] = field.One()
					colB0Wit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					colC0Wit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					colB1Wit[i] = field.NewElement(uint64(i + 1 - (flagSizeB - 10)))
					colC1Wit[i] = field.NewElement(uint64(i + 1 - (flagSizeB - 10)))
					colB2Wit[i] = field.NewElement(uint64(i + 2 - (flagSizeB - 10)))
					colC2Wit[i] = field.NewElement(uint64(i + 2 - (flagSizeB - 10)))
				}
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(flagC.GetColID(), smartvectors.RightZeroPadded(flagCWit, flagSizeB))
				run.AssignColumn(colA0.GetColID(), smartvectors.RightZeroPadded(colA0Wit, flagSizeA))
				run.AssignColumn(colA1.GetColID(), smartvectors.RightZeroPadded(colA1Wit, flagSizeA))
				run.AssignColumn(colA2.GetColID(), smartvectors.RightZeroPadded(colA2Wit, flagSizeA))
				run.AssignColumn(colB0.GetColID(), smartvectors.RightZeroPadded(colB0Wit, flagSizeB))
				run.AssignColumn(colB1.GetColID(), smartvectors.RightZeroPadded(colB1Wit, flagSizeB))
				run.AssignColumn(colB2.GetColID(), smartvectors.RightZeroPadded(colB2Wit, flagSizeB))
				run.AssignColumn(colC0.GetColID(), smartvectors.RightZeroPadded(colC0Wit, flagSizeB))
				run.AssignColumn(colC1.GetColID(), smartvectors.RightZeroPadded(colC1Wit, flagSizeB))
				run.AssignColumn(colC2.GetColID(), smartvectors.RightZeroPadded(colC2Wit, flagSizeB))
			},
		},
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			// This function assigns the initial module and is aimed at working
			// for all test-case.
			initialProve := tc.InitialProverFunc

			// initialComp is defined according to the define function provided by the
			// test-case.
			initialComp := wizard.Compile(tc.DefineFunc)

			disc := namebaseddiscoverer.PeriodSeperatingModuleDiscoverer{}
			disc.Analyze(initialComp)

			// This declares a compiled IOP with only the columns of the module A
			moduleAComp := distributed.GetFreshModuleComp(initialComp, &disc, moduleAName)
			dist_projection.NewDistributeProjectionCtx(moduleAName, initialComp, moduleAComp, &disc)

			wizard.ContinueCompilation(moduleAComp, distributedprojection.CompileDistributedProjection, dummy.CompileAtProverLvl)

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
