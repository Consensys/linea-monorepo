package global

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// GetBoundaries extracts the boundaries for all the columns present in the expression.
func GetBoundaries(expr *symbolic.Expression, numSegments int, run *wizard.ProverRuntime,
) collection.Mapping[ifaces.ColID, []field.Element] {

	var (
		maxshift      = GetMaxShift(expr)
		board         = expr.Board()
		metadata      = board.ListVariableMetadata()
		boundaries    = collection.NewMapping[ifaces.ColID, []field.Element]()
		numBoundaries int
	)
	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:

			var (
				segmentSize   = t.Size() / numSegments
				tWitness      = t.GetColAssignment(run).IntoRegVecSaveAlloc()
				colBoundaries []field.Element
			)

			if !utils.IsPowerOfTwo(segmentSize) {
				panic("the segmentSize is not power of two")
			}

			if shifted, ok := t.(column.Shifted); ok {
				numBoundaries = maxshift - shifted.Offset

			} else {
				numBoundaries = maxshift
			}

			for segmentID := 0; segmentID < numSegments; segmentID++ {
				segmentLastRow := (segmentID + 1) * segmentSize
				for i := segmentLastRow - numBoundaries; i < segmentLastRow; i++ {
					colBoundaries = append(colBoundaries, tWitness[i])
				}
			}

			boundaries.InsertNew(t.GetColID(), colBoundaries)
		}

	}
	return boundaries
}

func GetMaxShift(expr *symbolic.Expression) int {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		maxshift = 0
	)
	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:
			if shifted, ok := t.(column.Shifted); ok {
				maxshift = max(maxshift, shifted.Offset)
			}
		}

	}
	return maxshift
}

// GetSegmentProvider returns the provider of the segment
func GetSegmentProvider(exprs []*symbolic.Expression, numSegments int, run *wizard.ProverRuntime,
) collection.Mapping[int, []field.Element] {

	var (
		providers = collection.NewMapping[int, []field.Element]()
	)
	for segmentID := 0; segmentID < numSegments; segmentID++ {
		segmentProvider := []field.Element{}
		for _, expr := range exprs {
			exprBoundaries := GetBoundaries(expr, numSegments, run)
			for _, col := range exprBoundaries.ListAllKeys() {
				colBoundaries := exprBoundaries.MustGet(col)
				numBoundariesPerSegment := len(colBoundaries) / numSegments
				firstIndex := numBoundariesPerSegment * segmentID
				segmentProvider = append(segmentProvider, colBoundaries[firstIndex:numBoundariesPerSegment]...)
			}
		}
		providers.InsertNew(segmentID, segmentProvider)
	}
	return providers
}

func AdjustExpressionForGlobal(comp *wizard.CompiledIOP,
	expr *symbolic.Expression, segID, numSegments int,
) *symbolic.Expression {

	var (
		board          = expr.Board()
		metadatas      = board.ListVariableMetadata()
		translationMap = collection.NewMapping[string, *symbolic.Expression]()
		colTranslation ifaces.Column
		// column is split fairly among segments.
		size    = column.ExprIsOnSameLengthHandles(&board)
		segSize = size / numSegments
	)

	for _, metadata := range metadatas {

		// For each slot, get the expression obtained by replacing the commitment
		// by the appropriated column.

		switch m := metadata.(type) {
		case ifaces.Column:

			switch col := m.(type) {
			case column.Natural:
				colTranslation = comp.Columns.GetHandle(m.GetColID())

			case verifiercol.VerifierCol:
				// Create the split in live
				colTranslation = col.Split(comp, segID*segSize, (segID+1)*segSize)

			// Shift the subparent, if the offset is larger than the subparent
			// we repercute it on the num
			case column.Shifted:
				colTranslation = column.Shift(comp.Columns.GetHandle(col.Parent.GetColID()), col.Offset)

			}

			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(colTranslation))
		case variables.X:
			utils.Panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// Check that the period is not larger than the domain size. If
			// the period is smaller this is a no-op because the period does
			// not change.
			translated := symbolic.NewVariable(metadata)

			if m.T > segSize {

				// Here, there are two possibilities. (1) The current slot is
				// on a portion of the Periodic sample where everything is
				// zero or (2) the current slot matches a portion of the
				// periodic sampling containing a 1. To determine which is
				// the current situation, we need to find out where the slot
				// is located compared to the period.
				var (
					slotStartAt = (segID * segSize) % m.T
					slotStopAt  = slotStartAt + segSize
				)

				if m.Offset >= slotStartAt && m.Offset < slotStopAt {
					translated = variables.NewPeriodicSample(segSize, m.Offset%segSize)
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
