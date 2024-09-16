package stitcher

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

func (ctx stitchingContext) constraints() {
	ctx.LocalOpening()
	ctx.LocalGlobalConstraints()
}

func (ctx stitchingContext) LocalOpening() {
	// Ignore the LocalOpening queries over the subColumns.
	for _, qName := range ctx.comp.QueriesParams.AllUnignoredKeys() {
		// Filters out only the LocalOpening
		q, ok := ctx.comp.QueriesParams.Data(qName).(query.LocalOpening)
		if !ok {
			utils.Panic("got an uncompilable query %v", qName)
		}

		if !isColEligible(ctx, q.Pol) {
			continue
		}
		// mark the query as ignored
		ctx.comp.QueriesParams.MarkAsIgnored(qName)

		// get the stitching column associated with the sub column q.Poly.
		stitchingCol := getStitchingCol(ctx, q.Pol)
		round := q.Pol.Round()

		newQ := ctx.comp.InsertLocalOpening(round, queryName(q.ID), stitchingCol)

		// Registers the prover's step responsible for assigning the new query
		ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
			y := run.QueriesParams.MustGet(q.ID).(query.LocalOpeningParams).Y
			run.AssignLocalPoint(newQ.ID, y)
		})
	}

}

func (ctx stitchingContext) LocalGlobalConstraints() {
	for _, qName := range ctx.comp.QueriesNoParams.AllUnignoredKeys() {

		q := ctx.comp.QueriesNoParams.Data(qName)
		// round of definition of the query to compile
		round := ctx.comp.QueriesNoParams.Round(qName)

		var board symbolic.ExpressionBoard

		switch q := q.(type) {
		case query.LocalConstraint:
			board = q.Board()
			// detect if the expression is eligible;
			// i.e., it contains columns of proper size with status Precomputed, committed, or verifiercol.
			if !isExprEligible(ctx, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the stitching columns
			ctx.comp.InsertLocal(round, queryName(qName), ctx.adjustExpression(q.Expression, false))

		case query.GlobalConstraint:
			board = q.Board()
			// detect if the expression is over the eligible columns.
			if !isExprEligible(ctx, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the stitching columns
			ctx.comp.InsertGlobal(round, queryName(qName),
				ctx.adjustExpression(q.Expression, true),
				q.NoBoundCancel)

		default:
			utils.Panic("got an uncompilable query %++v", qName)
		}
	}
}

// Takes a sub column and returns the stitching column.
// the stitching column is shifted such that the first row agrees with the first row of the sub column.
// more detailed, such stitching column agrees with the the sub column up to a subsampling with offset zero.
func getStitchingCol(ctx stitchingContext, col ifaces.Column) ifaces.Column {

	switch m := col.(type) {
	case verifiercol.VerifierCol:
		scaling := ctx.MaxSize / col.Size()
		return verifiercol.ExpandedVerifCol{
			Verifiercol: m,
			Expansion:   scaling,
		}
	}

	// Extract the assumedly single col
	natural := column.RootParents(col)[0]

	round := col.Round()
	subColInfo := ctx.Stitchings[round].BySubCol[natural.GetColID()]
	stitchingCol := ctx.comp.Columns.GetHandle(subColInfo.NameStitching)

	// Shift the stitching column by the right position
	position := column.StackOffsets(col)

	scaling := stitchingCol.Size() / natural.Size()
	newPosition := scaling*position + subColInfo.PosInNew

	return column.Shift(stitchingCol, newPosition)
}

func queryName(oldQ ifaces.QueryID) ifaces.QueryID {
	return ifaces.QueryIDf("%v_STITCHER", oldQ)
}

// it adjusts the expression, that is among sub columns, by replacing the sub columns with their stitching columns.
// for the verfiercol instead of stitching, they are expanded to reach the proper size.
// This is due to the fact that the verifiercols are not tracked by the compiler and can not be stitched
// via [scanAndClassifyEligibleColumns].
func (ctx *stitchingContext) adjustExpression(
	expr *symbolic.Expression,
	isGlobalConstraint bool,
) (
	newExpr *symbolic.Expression,
) {

	board := expr.Board()
	metadata := board.ListVariableMetadata()
	replaceMap := collection.NewMapping[string, *symbolic.Expression]()
	domainSize := 0

	for i := range metadata {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			// it's always a compiled column
			domainSize = m.Size()
			stitchingCol := getStitchingCol(*ctx, m)
			replaceMap.InsertNew(m.String(), ifaces.ColumnAsVariable(stitchingCol))
		case coin.Info, ifaces.Accessor:
			replaceMap.InsertNew(m.String(), symbolic.NewVariable(m))
		case variables.X:
			panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// there, we need to inflate the period and the offset
			scaling := ctx.MaxSize / domainSize
			replaceMap.InsertNew(m.String(), variables.NewPeriodicSample(m.T*scaling, m.Offset*scaling))
		}
	}

	newExpr = expr.Replay(replaceMap)
	if isGlobalConstraint {
		// for the global constraints, check the constraint only over the subSampeling of the columns.
		newExpr = symbolic.Mul(newExpr, variables.NewPeriodicSample(ctx.MaxSize/domainSize, 0))
	}

	return newExpr
}
