package stitcher

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	alliance "github.com/consensys/linea-monorepo/prover/protocol/compiler/stitch_split"
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

		round := ctx.comp.QueriesParams.Round(q.ID)

		if q.Pol.Size() < ctx.MinSize {
			//sanity-check: column should be public
			verifiercol.AssertIsPublicCol(ctx.comp, q.Pol)
			// Ask the verifier to directly check the query
			insertVerifier(ctx.comp, q, round)
			// mark the query as ignored
			ctx.comp.QueriesParams.MarkAsIgnored(q.ID)

			// And skip the rest of the compilation : we are done
			continue
		}

		if !isColEligible(ctx.Stitchings, q.Pol) {
			continue
		}

		switch m := q.Pol.(type) {
		case verifiercol.VerifierCol:
			utils.Panic("unsupported, received a localOpening over the verifier column %v", m.GetColID())
		}
		// mark the query as ignored
		ctx.comp.QueriesParams.MarkAsIgnored(qName)

		// get the stitching column associated with the sub column q.Poly.
		stitchingCol := getStitchingCol(ctx, q.Pol)

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
			if q.DomainSize < ctx.MinSize {
				// Sanity-check : at this point all the parameters of the query
				// should have a public status. Indeed, prior to compiling the
				// constraints to work
				metadatas := board.ListVariableMetadata()
				for _, metadata := range metadatas {
					if h, ok := metadata.(ifaces.Column); ok {
						verifiercol.AssertIsPublicCol(ctx.comp, h)
					}
				}
				insertVerifier(ctx.comp, q, round)
				// mark the query as ignored
				ctx.comp.QueriesNoParams.MarkAsIgnored(qName)
				continue
			}
			// detect if the expression is eligible;
			// i.e., it contains columns of proper size with status Precomputed, committed, or verifiercol.
			if !alliance.IsExprEligible(isColEligible, ctx.Stitchings, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the stitching columns
			ctx.comp.InsertLocal(round, queryName(qName), ctx.adjustExpression(q.Expression, q.DomainSize, false))

		case query.GlobalConstraint:
			board = q.Board()
			if q.DomainSize < ctx.MinSize {

				// Sanity-check : at this point all the parameters of the query
				// should have a public status. Indeed, prior to compiling the
				// constraints to work
				metadatas := board.ListVariableMetadata()
				for _, metadata := range metadatas {
					if h, ok := metadata.(ifaces.Column); ok {
						verifiercol.AssertIsPublicCol(ctx.comp, h)
					}
				}
				insertVerifier(ctx.comp, q, round)
				// mark the query as ignored
				ctx.comp.QueriesNoParams.MarkAsIgnored(qName)
				continue
			}
			// detect if the expression is over the eligible columns.
			if !alliance.IsExprEligible(isColEligible, ctx.Stitchings, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the stitching columns
			ctx.comp.InsertGlobal(round, queryName(qName),
				ctx.adjustExpression(q.Expression, q.DomainSize, true),
				q.NoBoundCancel)

		default:
			utils.Panic("got an uncompilable query %++v", qName)
		}
	}
}

// Takes a sub column and returns the stitching column.
// the stitching column is shifted such that the first row agrees with the first row of the sub column.
// more detailed, such stitching column agrees with the the sub column up to a subsampling with offset zero.
// the col should only be either verifiercol or eligible col.
// option is always empty, and used only for the recursive calls over the shifted columns.
func getStitchingCol(ctx stitchingContext, col ifaces.Column, option ...int) ifaces.Column {
	var (
		stitchingCol ifaces.Column
		newOffset    int
		round        = col.Round()
	)

	switch m := col.(type) {
	// case: verifier columns without shift
	case verifiercol.VerifierCol:
		scaling := ctx.MaxSize / col.Size()
		// expand the veriferCol
		stitchingCol = verifiercol.ExpandedVerifCol{
			Verifiercol: m,
			Expansion:   scaling,
		}
		if len(option) != 0 {
			// if it is a shifted veriferCol, set the offset for shifting the expanded column
			newOffset = option[0] * col.Size()
		}
		return column.Shift(stitchingCol, newOffset)
	case column.Natural:
		// find the stitching column
		subColInfo := ctx.Stitchings[round].BySubCol[col.GetColID()]
		stitchingCol = ctx.comp.Columns.GetHandle(subColInfo.NameBigCol)
		scaling := stitchingCol.Size() / col.Size()
		if len(option) != 0 {
			newOffset = scaling * option[0]
		}
		newOffset = newOffset + subColInfo.PosInBigCol
		return column.Shift(stitchingCol, newOffset)

	case column.Shifted:
		// Shift the stitching column by the right position
		offset := column.StackOffsets(col)
		col = column.RootParents(col)[0]
		res := getStitchingCol(ctx, col, offset)
		return res

	default:

		panic("unsupported")

	}
}

func queryName(oldQ ifaces.QueryID) ifaces.QueryID {
	return ifaces.QueryIDf("%v_STITCHER", oldQ)
}

// it adjusts the expression, that is among sub columns, by replacing the sub columns with their stitching columns.
// for the verfiercol instead of stitching, they are expanded to reach the proper size.
// This is due to the fact that the verifiercols are not tracked by the compiler and can not be stitched
// via [scanAndClassifyEligibleColumns].
func (ctx *stitchingContext) adjustExpression(
	expr *symbolic.Expression, domainSize int,
	isGlobalConstraint bool,
) (
	newExpr *symbolic.Expression,
) {

	board := expr.Board()
	metadata := board.ListVariableMetadata()
	replaceMap := collection.NewMapping[string, *symbolic.Expression]()

	for i := range metadata {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			// it's always a compiled column
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

func insertVerifier(
	comp *wizard.CompiledIOP,
	q ifaces.Query,
	round int,
) {

	// Requires the verifier to verify the query itself
	comp.InsertVerifier(round, func(vr *wizard.VerifierRuntime) error {
		return q.Check(vr)
	}, func(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
		q.CheckGnark(api, wvc)
	})

}
