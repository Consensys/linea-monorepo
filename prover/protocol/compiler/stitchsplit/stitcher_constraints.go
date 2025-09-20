package stitchsplit

import (
	"strconv"

	"github.com/consensys/gnark/frontend"
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
	"github.com/sirupsen/logrus"
)

func (ctx StitchingContext) constraints() {
	ctx.LocalOpening()
	ctx.LocalGlobalConstraints()
}

func (ctx StitchingContext) LocalOpening() {

	// Ignore the LocalOpening queries over the subColumns.
	for _, qName := range ctx.Comp.QueriesParams.AllUnignoredKeys() {
		// Filters out only the LocalOpening
		q, ok := ctx.Comp.QueriesParams.Data(qName).(query.LocalOpening)
		if !ok {
			utils.Panic("got an uncompilable query %v", qName)
		}

		round := ctx.Comp.QueriesParams.Round(q.ID)

		if q.Pol.Size() < ctx.MinSize {
			//sanity-check: column should be public
			verifiercol.AssertIsPublicCol(ctx.Comp, q.Pol)
			// Ask the verifier to directly check the query
			insertVerifier(ctx.Comp, q, round)
			// mark the query as ignored
			ctx.Comp.QueriesParams.MarkAsIgnored(q.ID)

			// And skip the rest of the compilation : we are done
			continue
		}

		if !isColEligibleStitching(ctx.Stitchings, q.Pol) {
			continue
		}

		switch m := q.Pol.(type) {
		case verifiercol.VerifierCol:
			utils.Panic("unsupported, received a localOpening over the verifier column %v", m.GetColID())
		}
		// mark the query as ignored
		ctx.Comp.QueriesParams.MarkAsIgnored(qName)

		// get the stitching column associated with the sub column q.Poly.
		stitchingCol := getStitchingCol(ctx, q.Pol)

		newQ := ctx.Comp.InsertLocalOpening(round, queryNameStitcher(q.ID), stitchingCol)

		// Registers the prover's step responsible for assigning the new query
		ctx.Comp.RegisterProverAction(round, &AssignLocalPointProverAction{
			QID:  q.ID,
			NewQ: newQ.ID,
		})
	}

}

func (ctx StitchingContext) LocalGlobalConstraints() {
	for _, qName := range ctx.Comp.QueriesNoParams.AllUnignoredKeys() {

		q := ctx.Comp.QueriesNoParams.Data(qName)
		// round of definition of the query to compile
		round := ctx.Comp.QueriesNoParams.Round(qName)

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
						verifiercol.AssertIsPublicCol(ctx.Comp, h)
					}
				}
				insertVerifier(ctx.Comp, q, round)
				// mark the query as ignored
				ctx.Comp.QueriesNoParams.MarkAsIgnored(qName)
				continue
			}

			// detect if the expression is eligible;
			// i.e., it contains columns of proper size with status Precomputed, committed, or verifiercol.
			isEligible, unSupported := IsExprEligibleForStitching(isColEligibleStitching, ctx.Stitchings, board)
			if !isEligible && !unSupported {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.Comp.QueriesNoParams.MarkAsIgnored(qName)

			if unSupported {
				continue
			}

			// adjust the query over the stitching columns
			ctx.Comp.InsertLocal(round, queryNameStitcher(qName), ctx.adjustExpression(q.Expression, q.DomainSize, false))

		case query.GlobalConstraint:
			board = q.Board()
			if q.DomainSize < ctx.MinSize {

				// Sanity-check : at this point all the parameters of the query
				// should have a public status. Indeed, prior to compiling the
				// constraints to work
				metadatas := board.ListVariableMetadata()
				for _, metadata := range metadatas {
					if h, ok := metadata.(ifaces.Column); ok {
						verifiercol.AssertIsPublicCol(ctx.Comp, h)
					}
				}
				insertVerifier(ctx.Comp, q, round)
				// mark the query as ignored
				ctx.Comp.QueriesNoParams.MarkAsIgnored(qName)
				continue
			}
			// detect if the expression is over the eligible columns.
			isEligible, unSupported := IsExprEligibleForStitching(isColEligibleStitching, ctx.Stitchings, board)
			if !isEligible && !unSupported {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.Comp.QueriesNoParams.MarkAsIgnored(qName)

			if unSupported {
				continue
			}

			// adjust the query over the stitching columns
			ctx.Comp.InsertGlobal(round, queryNameStitcher(qName),
				ctx.adjustExpression(q.Expression, q.DomainSize, true),
				q.NoBoundCancel)

		default:
			utils.Panic("got an uncompilable query %++v", qName)
		}
	}
}

// Takes a sub column and returns the stitching column.
// The stitching column is shifted in such a way that the first row agrees with the first row of the sub column.
// i.e., such stitching column agrees with the sub column up to a subsampling with offset zero.
// The col should only be either verifiercol or eligible col.
// option is always empty, and used only for the recursive calls over the shifted columns.
func getStitchingCol(ctx StitchingContext, col ifaces.Column, option ...int) ifaces.Column {
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
		stitchingCol = ctx.Comp.Columns.GetHandle(subColInfo.NameBigCol)
		scaling := stitchingCol.Size() / col.Size()
		if len(option) != 0 {
			newOffset = scaling * option[0]
		}
		newOffset = newOffset + subColInfo.PosInBigCol
		return column.Shift(stitchingCol, newOffset)

	case column.Shifted:
		// Shift the stitching column by the right position
		offset := column.StackOffsets(col)
		col = column.RootParents(col)
		res := getStitchingCol(ctx, col, offset)
		return res

	default:

		panic("unsupported")

	}
}

// It checks if the expression is over a set of the columns eligible to the stitching.
// Namely, it contains columns of proper size with status Precomputed, Committed, Verifiercol, Proof, and Verifying key.
// It panics if the expression includes a mixture of eligible columns and columns with status Ignored.
// If all the columns are verifierCol the expression is not eligible to the compilation.
// This is an expected behavior, since the verifier checks such expression by itself.
func IsExprEligibleForStitching(
	isColEligible func(MultiSummary, ifaces.Column) bool,
	stitchings MultiSummary,
	board symbolic.ExpressionBoard,
) (isEligible bool, isUnsupported bool) {

	var (
		metadata              = board.ListVariableMetadata()
		hasAtLeastOneEligible = false
		allAreEligible        = true
		allAreVeriferCol      = true
		statusMap             = map[ifaces.ColID]string{}
		b                     = true
	)

	for i := range metadata {
		switch m := metadata[i].(type) {
		// reminder: [verifiercol.VerifierCol] , [column.Natural] and [column.Shifted]
		// all implement [ifaces.Column]
		case ifaces.Column: // it is a Committed, Precomputed or verifierCol
			rootColumn := column.RootParents(m)

			switch nat := rootColumn.(type) {
			case column.Natural: // then it is not a verifiercol
				switch nat.Status() {
				case column.Proof:
					// proof columns are eligible,
					// we already checked that their size > minSize
					b = true
				case column.VerifyingKey:
					// verifying key columns are eligible,
					// we already checked that their size > minSize
					b = true
				case column.Committed, column.Precomputed:
					b = isColEligible(stitchings, m)
				default:
					utils.Panic("unsupported column status %v", nat.Status())
				}
				statusMap[rootColumn.GetColID()] = nat.Status().String() + "/" + strconv.Itoa(nat.Size())
				allAreVeriferCol = false

				hasAtLeastOneEligible = hasAtLeastOneEligible || b
				allAreEligible = allAreEligible && b
				if m.Size() == 0 {
					panic("found a column with a size of 0")
				}
			case verifiercol.VerifierCol:
				statusMap[rootColumn.GetColID()] = column.VerifierDefined.String() + "/" + strconv.Itoa(nat.Size())
			}
		case variables.PeriodicSample:
			// periodic samples are always eligible
		default:
			// unsupported column type
			utils.Panic("unsupported column type %T", m)
		}
	}

	if hasAtLeastOneEligible && !allAreEligible {
		// We expect no expression over ignored columns
		logrus.Errorf("the expression is not valid, it is mixed with invalid columns of status Ignored, %v", statusMap)
		return false, true
	}

	if allAreVeriferCol {
		// 4. we expect no expression involving only and only the verifierCols.
		// We expect that this case wont happen.
		// Otherwise should be handled in the [github.com/consensys/linea-monorepo/prover/protocol/query] package.
		// Namely, Local/Global queries should be checked directly by the verifer.
		panic("all the columns in the expression are verifierCols, unsupported by the compiler")
	}

	return hasAtLeastOneEligible, false
}

func queryNameStitcher(oldQ ifaces.QueryID) ifaces.QueryID {
	return ifaces.QueryIDf("%v_STITCHER", oldQ)
}

// it adjusts the expression, that is among sub columns, by replacing the sub columns with their stitching columns.
// for the verfiercol instead of stitching, they are expanded to reach the proper size.
// This is due to the fact that the verifiercols are not tracked by the compiler and can not be stitched
// via [scanAndClassifyEligibleColumns].
func (ctx *StitchingContext) adjustExpression(
	expr *symbolic.Expression, domainSize int,
	isGlobalConstraint bool,
) (
	newExpr *symbolic.Expression,
) {

	var (
		board        = expr.Board()
		metadata     = board.ListVariableMetadata()
		replaceMap   = collection.NewMapping[string, *symbolic.Expression]()
		stitchingCol ifaces.Column
	)

	for i := range metadata {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			// it's always a compiled column
			rootColumn := column.RootParents(m)
			switch nat := rootColumn.(type) {
			case column.Natural: // then it is not a verifiercol
				switch nat.Status() {
				case column.Proof, column.VerifyingKey:
					stitchingCol = verifiercol.ExpandedProofOrVerifyingKeyColWithZero{
						Col:       rootColumn,
						Expansion: ctx.MaxSize / rootColumn.Size(),
					}
				case column.Committed, column.Precomputed:
					stitchingCol = getStitchingCol(*ctx, m)
				default:
					utils.Panic("unsupported column status %v", nat.Status())
				}
				replaceMap.InsertNew(m.String(), ifaces.ColumnAsVariable(stitchingCol))
			}
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

type QueryVerifierAction struct {
	Q ifaces.Query
}

func (a *QueryVerifierAction) Run(vr wizard.Runtime) error {
	return a.Q.Check(vr)
}

func (a *QueryVerifierAction) RunGnark(api frontend.API, wvc wizard.GnarkRuntime) {
	a.Q.CheckGnark(api, wvc)
}

func insertVerifier(
	comp *wizard.CompiledIOP,
	q ifaces.Query,
	round int,
) {
	// Register the VerifierAction instead of using a closure
	comp.RegisterVerifierAction(round, &QueryVerifierAction{
		Q: q,
	})
}
