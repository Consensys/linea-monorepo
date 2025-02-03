package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
)

func TestDistributedProjection(t *testing.T) {
	var (
		runS                           *wizard.ProverRuntime
		DP                             ifaces.Query
		round                          = 0
		flagSizeA                      = 512
		flagSizeB                      = 256
		flagA, flagB, columnA, columnB ifaces.Column
		evalRand                       field.Element
		hornerA, hornerB, hornerBoth   field.Element
		flagAWit                       = make([]field.Element, flagSizeA)
		columnAWit                     = make([]field.Element, flagSizeA)
		flagBWit                       = make([]field.Element, flagSizeB)
		columnBWit                     = make([]field.Element, flagSizeB)
		hornerAArray                   = make([]field.Element, flagSizeA)
		hornerBArray                   = make([]field.Element, flagSizeB)
		queryNameBothAAndB             = ifaces.QueryID("DistributedProjectionTestBothAAndB")
		queryNameOnlyA                 = ifaces.QueryID("DistributedProjectionTestOnlyA")
		queryNameOnlyB                 = ifaces.QueryID("DistributedProjectionTestOnlyB")
	)
	// Computing common test data
	_, errBeta := evalRand.SetRandom()
	if errBeta != nil {
		// Cannot happen unless the entropy was exhausted
		panic(errBeta)
	}
	// assign filters and columns
	for i := 0; i < 10; i++ {
		flagAWit[i] = field.One()
		columnAWit[i] = field.NewElement(uint64(i))
	}
	for i := flagSizeB - 10; i < flagSizeB; i++ {
		flagBWit[i] = field.One()
		columnBWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
	}
	hornerBoth = field.Zero()
	hornerAArray = poly.CmptHorner(columnAWit, flagAWit, evalRand)
	hornerBArray = poly.CmptHorner(columnBWit, flagBWit, evalRand)
	hornerA = hornerAArray[0]
	hornerB = hornerBArray[0]
	hornerB.Neg(&hornerB)

	testcases := []struct {
		Name        string
		HornerParam field.Element
		QueryName   ifaces.QueryID
		DefineFunc  func(builder *wizard.Builder)
	}{
		{
			Name:        "distributed-projection-both-A-and-B",
			HornerParam: hornerBoth,
			QueryName:   queryNameBothAAndB,
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("FliterB"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("ColumnB"), flagSizeB)
				var (
					colA, _, _ = wizardutils.AsExpr(columnA)
					colB, _, _ = wizardutils.AsExpr(columnB)
					fA, _, _   = wizardutils.AsExpr(flagA)
					fB, _, _   = wizardutils.AsExpr(flagB)
				)
				DP = builder.CompiledIOP.InsertDistributedProjection(round, queryNameBothAAndB,
					[]*query.DistributedProjectionInput{
						{ColumnA: colA, ColumnB: colB, FilterA: fA, FilterB: fB, IsAInModule: true, IsBInModule: true},
					})
			},
		},
		{
			Name:        "distributed-projection-only-A",
			HornerParam: hornerA,
			QueryName:   queryNameOnlyA,
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("FliterB"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("ColumnB"), flagSizeB)
				var (
					colA, _, _ = wizardutils.AsExpr(columnA)
					colB, _, _ = wizardutils.AsExpr(columnB)
					fA, _, _   = wizardutils.AsExpr(flagA)
					fB, _, _   = wizardutils.AsExpr(flagB)
				)
				DP = builder.CompiledIOP.InsertDistributedProjection(round, queryNameOnlyA,
					[]*query.DistributedProjectionInput{
						{ColumnA: colA, ColumnB: colB, FilterA: fA, FilterB: fB, IsAInModule: true, IsBInModule: false},
					})
			},
		},
		{
			Name:        "distributed-projection-only-B",
			HornerParam: hornerB,
			QueryName:   queryNameOnlyB,
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("FliterB"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("ColumnB"), flagSizeB)
				var (
					colA, _, _ = wizardutils.AsExpr(columnA)
					colB, _, _ = wizardutils.AsExpr(columnB)
					fA, _, _   = wizardutils.AsExpr(flagA)
					fB, _, _   = wizardutils.AsExpr(flagB)
				)
				DP = builder.CompiledIOP.InsertDistributedProjection(round, queryNameOnlyB,
					[]*query.DistributedProjectionInput{
						{ColumnA: colA, ColumnB: colB, FilterA: fA, FilterB: fB, IsAInModule: false, IsBInModule: true},
					})
			},
		},
	}

	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			prover := func(run *wizard.ProverRuntime) {
				runS = run
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(columnB.GetColID(), smartvectors.RightZeroPadded(columnBWit, flagSizeB))
				run.AssignColumn(columnA.GetColID(), smartvectors.RightZeroPadded(columnAWit, flagSizeA))

				runS.AssignDistributedProjection(tc.QueryName, query.DistributedProjectionParams{HornerVal: tc.HornerParam, EvalRands: []field.Element{evalRand}})
			}
			var (
				comp  = wizard.Compile(tc.DefineFunc)
				_     = wizard.Prove(comp, prover)
				errDP = DP.Check(runS)
			)

			if errDP != nil {
				t.Fatalf("error verifying the distributed projection query: %v", errDP.Error())
			}

		})
	}

}
