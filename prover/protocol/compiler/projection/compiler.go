package projection

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// CompileProjection compiles [query.Projection] queries
func CompileProjection(comp *wizard.CompiledIOP) {

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {
		// Filter out non projection queries
		projection, ok := comp.QueriesNoParams.Data(qName).(query.Projection)
		if !ok {
			continue
		}

		// This ensures that the projection query is not used again in the
		// compilation process. We know that the query was not already ignored at the beginning
		// because we are iterating over the unignored keys.
		comp.QueriesNoParams.MarkAsIgnored(qName)
		round := comp.QueriesNoParams.Round(qName)
		compile(comp, round, projection)
	}

}

func compile(comp *wizard.CompiledIOP, round int, projection query.Projection) {
	var (
		compRound    = round
		sizeA        = projection.Inp.FilterA.Size()
		sizeB        = projection.Inp.FilterB.Size()
		numCol       = len(projection.Inp.ColumnA)
		a, b, af, bf any
		queryName    = projection.ID
		columnsA     = projection.Inp.ColumnA
		columnsB     = projection.Inp.ColumnB
		filterA      = projection.Inp.FilterA
		filterB      = projection.Inp.FilterB
	)

	// a and b are storing the columns used to compute the linear combination
	// of the columnsA and columnsB. The initial assignment is for the case
	// where there is only a single column. If there is more than one column
	// then they will store an expression computing a random linear
	// combination of the columns.
	a, b = projection.Inp.ColumnA[0], projection.Inp.ColumnB[0]

	// af and bf are as a and b but shifted by -1. They are initially
	// assigned assuming the case where the number of column is 1 and
	// replaced later by a random linear combination if not. They are meant
	// to be used in the local constraints to point to the last entry of the
	// "a" and "b".
	af, bf = column.Shift(projection.Inp.ColumnA[0], -1), column.Shift(projection.Inp.ColumnB[0], -1)

	if numCol > 0 {
		compRound++
		alpha := comp.InsertCoin(compRound, coin.Namef("%v_MERGING_COIN", queryName), coin.Field)
		a = wizardutils.RandLinCombColSymbolic(alpha, columnsA)
		b = wizardutils.RandLinCombColSymbolic(alpha, columnsB)

		afs := make([]ifaces.Column, numCol)
		bfs := make([]ifaces.Column, numCol)

		for i := range afs {
			afs[i] = column.Shift(columnsA[i], -1)
			bfs[i] = column.Shift(columnsB[i], -1)
		}

		af = wizardutils.RandLinCombColSymbolic(alpha, afs)
		bf = wizardutils.RandLinCombColSymbolic(alpha, bfs)
	}

	var (
		aExpr, _, _ = wizardutils.AsExpr(a)
		bExpr, _, _ = wizardutils.AsExpr(b)
		pa          = projectionProverAction{
			Name:     queryName,
			EvalCoin: comp.InsertCoin(compRound, coin.Namef("%v_EVAL_COIN", queryName), coin.Field),
			FilterA:  filterA,
			FilterB:  filterB,
			ColA:     columnsA,
			ColB:     columnsB,
			ABoard:   aExpr.Board(),
			BBoard:   bExpr.Board(),
			HornerA:  comp.InsertCommit(compRound, ifaces.ColIDf("%v_HORNER_A", queryName), sizeA),
			HornerB:  comp.InsertCommit(compRound, ifaces.ColIDf("%v_HORNER_B", queryName), sizeB),
		}
	)

	comp.InsertGlobal(
		compRound,
		ifaces.QueryIDf("%v_HORNER_A_GLOBAL", queryName),
		sym.Sub(
			pa.HornerA,
			sym.Mul(
				sym.Sub(1, pa.FilterA),
				column.Shift(pa.HornerA, 1),
			),
			sym.Mul(
				pa.FilterA,
				sym.Add(
					a,
					sym.Mul(
						pa.EvalCoin,
						column.Shift(pa.HornerA, 1),
					),
				),
			),
		),
	)

	comp.InsertGlobal(
		compRound,
		ifaces.QueryIDf("%v_HORNER_B_GLOBAL", queryName),
		sym.Sub(
			pa.HornerB,
			sym.Mul(
				sym.Sub(1, pa.FilterB),
				column.Shift(pa.HornerB, 1),
			),
			sym.Mul(
				pa.FilterB,
				sym.Add(b, sym.Mul(pa.EvalCoin, column.Shift(pa.HornerB, 1))),
			),
		),
	)

	comp.InsertLocal(
		compRound,
		ifaces.QueryIDf("%v_HORNER_A_LOCAL_END", queryName),
		sym.Sub(
			column.Shift(pa.HornerA, -1),
			sym.Mul(column.Shift(pa.FilterA, -1), af),
		),
	)

	comp.InsertLocal(
		compRound,
		ifaces.QueryIDf("%v_HORNER_B_LOCAL_END", queryName),
		sym.Sub(
			column.Shift(pa.HornerB, -1),
			sym.Mul(column.Shift(pa.FilterB, -1), bf),
		),
	)

	pa.HornerA0 = comp.InsertLocalOpening(compRound, ifaces.QueryIDf("%v_HORNER_A0", queryName), pa.HornerA)
	pa.HornerB0 = comp.InsertLocalOpening(compRound, ifaces.QueryIDf("%v_HORNER_B0", queryName), pa.HornerB)

	comp.RegisterProverAction(compRound, pa)
	comp.RegisterVerifierAction(compRound, &projectionVerifierAction{HornerA0: pa.HornerA0, HornerB0: pa.HornerB0, Name: queryName})

}
