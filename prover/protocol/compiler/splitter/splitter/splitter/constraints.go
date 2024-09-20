package splitter

import (
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	alliance "github.com/consensys/linea-monorepo/prover/protocol/compiler/splitter/splitter"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

func (ctx splitterContext) constraints() {
	ctx.LocalOpening()
	ctx.LocalGlobalConstraints()
}

func (ctx splitterContext) LocalOpening() {
	for _, qName := range ctx.comp.QueriesParams.AllUnignoredKeys() {
		// Filters out only the LocalOpening
		q, ok := ctx.comp.QueriesParams.Data(qName).(query.LocalOpening)
		if !ok {
			utils.Panic("got an uncompilable query %v", qName)
		}

		round := ctx.comp.QueriesParams.Round(q.ID)

		if !isColEligible(ctx.Splittings, q.Pol) {
			continue
		}
		// mark the query as ignored
		ctx.comp.QueriesParams.MarkAsIgnored(qName)
		// Get the sub column
		subCol := getSubColForLocal(ctx, q.Pol)
		// apply the local constrain over the subCol
		newQ := ctx.comp.InsertLocalOpening(round, queryName(q.ID), subCol)

		// Registers the prover's step responsible for assigning the new query
		ctx.comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
			y := run.QueriesParams.MustGet(q.ID).(query.LocalOpeningParams).Y
			run.AssignLocalPoint(newQ.ID, y)
		})

	}
}

func (ctx splitterContext) LocalGlobalConstraints() {
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
			if !alliance.IsExprEligible(isColEligible, ctx.Splittings, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the sub columns
			ctx.comp.InsertLocal(round, queryName(qName), ctx.adjustExpressionForLocal(q.Expression, 0))

		case query.GlobalConstraint:
			board = q.Board()
			// detect if the expression is over the eligible columns.
			if !alliance.IsExprEligible(isColEligible, ctx.Splittings, board) {
				continue
			}

			// if the associated expression is eligible to the stitching, mark the query, over the sub columns, as ignored.
			ctx.comp.QueriesNoParams.MarkAsIgnored(qName)

			// adjust the query over the sub columns
			numSlots := q.DomainSize / ctx.size
			for slot := 0; slot < numSlots; slot++ {

				ctx.comp.InsertGlobal(round,
					ifaces.QueryIDf("SPLITTER_GLOBALQ_SLOT_%v", slot),
					ctx.adjustExpressionForGlobal(q.Expression, slot),
					q.NoBoundCancel)

				ctx.localQueriesForGapsInGlobal(q, slot, numSlots)
			}

		default:
			utils.Panic("got an uncompilable query %++v", qName)
		}
	}
}

// it checks if a column registered in the compiler has the proper size and state for splitting.
func isColEligible(splittings alliance.MultiSummary, col ifaces.Column) bool {
	natural := column.RootParents(col)[0]
	_, found := splittings[col.Round()].ByBigCol[natural.GetColID()]
	return found
}

// It finds the subCol containing the first row of col,
// it then shifts the subCol so that its first row equals with the first row of col.
func getSubColForLocal(ctx splitterContext, col ifaces.Column) ifaces.Column {
	// TBD verifierCol

	// find the subCol that contain the first row of col
	position := utils.PositiveMod(column.StackOffsets(col), col.Size())
	subColID, posInSubCol := position/ctx.size, position%ctx.size

	// The subCol is linked to the "root" of q.Pol (i.e., natural column associated with col)
	natural := column.RootParents(col)[0]
	round := col.Round()

	subColNat := ctx.Splittings[round].ByBigCol[natural.GetColID()][subColID]
	return column.Shift(subColNat, posInSubCol)
}

// For the column 'col' and the given 'posInCol',
// it returns the subColumn from the natural column located in position 'posInNatural'.
// where the posInNatural is calculated via the offset in Col.
func getSubColForGlobal(ctx splitterContext, col ifaces.Column, posInCol int) ifaces.Column {
	// Sanity-check : only for the edge-case h.Size() < ctx.size
	round := col.Round()
	if col.Size() < ctx.size {
		if posInCol > 0 {
			utils.Panic(
				"tried to get share #%v of column %v, but this is an undersized column %v",
				posInCol, col.GetColID(), col.Size(),
			)
		}

		// Not a split  column : returns the input directly
		return col
	}

	if !col.IsComposite() {
		switch m := col.(type) {
		case verifiercol.VerifierCol:
			// Create the split in live
			return m.Split(ctx.comp, posInCol*ctx.size, (posInCol+1)*ctx.size)
		default:
			// No changes
			return ctx.Splittings[round].ByBigCol[col.GetColID()][posInCol]
		}
	}

	switch inner := col.(type) {
	// Shift the subparent, if the offset is larger than the subparent
	// we repercute it on the num
	case column.Shifted:
		// This works fine assuming h.Size() > ctx.size
		var (
			offset                  = inner.Offset
			posInNatural            = posInCol + (offset / ctx.size)
			offsetOfSubColInNatural = utils.PositiveMod(offset, ctx.size)
		)
		// This indicates that the offset is so large
		if posInNatural < 0 {
			posInNatural = (col.Size() / ctx.size) + posInNatural
		}
		// The resulting offset should keep the same sign as the old one. This is
		// because the sign indicates which range of position is touched by
		// bound cancellation.
		if offsetOfSubColInNatural*offset < 0 {
			offsetOfSubColInNatural -= ctx.size
		}
		parent := ctx.Splittings[round].ByBigCol[inner.Parent.GetColID()][posInNatural]
		return column.Shift(parent, offsetOfSubColInNatural)

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

func queryName(oldQ ifaces.QueryID) ifaces.QueryID {
	return ifaces.QueryIDf("%v_SPLITTER", oldQ)
}

// it shift all the columns inside the expression by shift and then applies the local constraints.
func (ctx splitterContext) adjustExpressionForLocal(
	expr *symbolic.Expression, shift int,
) *symbolic.Expression {

	board := expr.Board()
	metadatas := board.ListVariableMetadata()
	translationMap := collection.NewMapping[string, *symbolic.Expression]()

	for _, metadata := range metadatas {
		// Replace the expression by the one

		switch m := metadata.(type) {
		case ifaces.Column:
			subCol := getSubColForLocal(ctx, column.Shift(m, shift))
			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(subCol))
		// @Azam why we need these cases?
		case coin.Info, ifaces.Accessor:
			translationMap.InsertNew(m.String(), symbolic.NewVariable(m))
		case variables.X:
			panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// @Azam why Positive? what about shift
			newSample := variables.NewPeriodicSample(m.T, utils.PositiveMod(m.Offset, m.T))
			translationMap.InsertNew(m.String(), newSample)
		default:
			// Repass the same variable
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}
	}

	newExpr := expr.Replay(translationMap)
	return newExpr
}

func (ctx splitterContext) adjustExpressionForGlobal(
	expr *symbolic.Expression, slot int,
) *symbolic.Expression {
	board := expr.Board()
	metadatas := board.ListVariableMetadata()
	translationMap := collection.NewMapping[string, *symbolic.Expression]()

	for _, metadata := range metadatas {

		// For each slot, get the expression obtained by replacing the commitment
		// by the appropriated column.

		switch m := metadata.(type) {
		case ifaces.Column:
			// Pass the same variable
			subCol := getSubColForGlobal(ctx, m, slot)
			// Sanity-check : the subHandle should have the target size
			if subCol.Size() != ctx.size {
				utils.Panic(
					"outgoing column %v should have size %v but has size %v (ingoing column was %v, with size %v)",
					subCol.GetColID(), ctx.size, subCol.Size(), m.GetColID(), m.Size(),
				)
			}
			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(subCol))
		case variables.X:
			utils.Panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// Check that the period is not larger than the domain size. If
			// the period is smaller this is a no-op because the period does
			// not change.
			translated := symbolic.NewVariable(metadata)

			if m.T > ctx.size {

				// Here, there are two possibilities. (1) The current slot is
				// on a portion of the Periodic sample where everything is
				// zero or (2) the current slot matchs a portion of the
				// periodic sampling containing a 1. To determine which is
				// the current situation, we need to find out where the slot
				// is located compared to the period.
				var (
					slotStartAt = (slot * ctx.size) % m.T
					slotStopAt  = slotStartAt + ctx.size
				)

				if m.Offset >= slotStartAt && m.Offset < slotStopAt {
					translated = variables.NewPeriodicSample(ctx.size, m.Offset%ctx.size)
				} else {
					translated = symbolic.NewConstant(0)
				}
			}

			// And we can just pass it over because the period does not change
			translationMap.InsertNew(m.String(), translated)
		default:
			// Repass the same variable (for coins or other types of single-valued variable)
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}

	}
	return expr.Replay(translationMap)
}

func (ctx splitterContext) localQueriesForGapsInGlobal(q query.GlobalConstraint, slot, numSlots int) {

	// Now, we need to cancel the expression at the beginning and/or the end
	// For the first one, only cancel the end. For the last one, only cancel
	// the beginning.
	offsetRange := q.MinMaxOffset()
	round := ctx.comp.QueriesNoParams.Round(q.ID)

	if offsetRange.Min < 0 {
		for i := 0; i < offsetRange.Min; i-- {
			// And fill the gap with a local constraint
			if slot > 0 || q.NoBoundCancel {
				// adjust the query over the sub columns
				ctx.comp.InsertLocal(round,
					ifaces.QueryIDf("%v_LOCAL_GAPS_NEG_OFFSET_%v", q.ID, i),
					ctx.adjustExpressionForLocal(q.Expression, slot*ctx.size-i))
			}
		}
	}

	if offsetRange.Max > 0 {
		for i := 0; i < offsetRange.Max; i++ {
			point := ctx.size - i - 1 // point at which we want to cancel the constraint
			// And fill the gap with a local constraint
			if slot < numSlots-1 || q.NoBoundCancel {
				shift := slot*ctx.size + point
				ctx.comp.InsertLocal(round,
					ifaces.QueryIDf("%v_LOCAL_GAPS_POS_OFFSET_%v", q.ID, i),
					ctx.adjustExpressionForLocal(q.Expression, shift))
			}
		}
	}
}
