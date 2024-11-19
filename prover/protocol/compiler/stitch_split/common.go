package alliance

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// It stores the information regarding an alliance between a BigCol and a set of SubColumns.
// The subject of alliance can be either stitching or splitting.
type Alliance struct {
	// the  bigCol in the alliance;
	// - for stitching, it is the result of stitching of the subColumns.
	// - for splitting, it is split to the sub Columns.
	BigCol ifaces.Column
	// sub columns allied with the bigCol
	SubCols []ifaces.Column
	// the Round in which the alliance is created.
	Round int
	// Status of the sub columns
	// the only valid Status for the eligible sub columns are;
	// committed, Precomputed, VerifierDefined
	Status column.Status
}

// It summarizes the information about all the alliances in a single round of PIOP.
type SummerizedAlliances struct {
	// associate a group of the sub columns to their splitting
	ByBigCol map[ifaces.ColID][]ifaces.Column
	// for a sub column, it indicates its splitting column and its position in the splitting.
	BySubCol map[ifaces.ColID]struct {
		NameBigCol  ifaces.ColID
		PosInBigCol int
	}
}

// The summary of all the alliances round-by-round.
type MultiSummary []SummerizedAlliances

// It inserts the new alliance into the summary.
func (summary MultiSummary) InsertNew(s Alliance) {
	// Initialize the bySubCol if necessary
	if summary[s.Round].BySubCol == nil {
		summary[s.Round].BySubCol = map[ifaces.ColID]struct {
			NameBigCol  ifaces.ColID
			PosInBigCol int
		}{}
	}

	// Populate the bySubCol
	for posInNew, c := range s.SubCols {
		summary[s.Round].BySubCol[c.GetColID()] = struct {
			NameBigCol  ifaces.ColID
			PosInBigCol int
		}{
			NameBigCol:  s.BigCol.GetColID(),
			PosInBigCol: posInNew,
		}
	}

	// Initialize ByBigCol if necessary
	if summary[s.Round].ByBigCol == nil {
		summary[s.Round].ByBigCol = make(map[ifaces.ColID][]ifaces.Column)
	}
	// populate ByBigCol
	summary[s.Round].ByBigCol[s.BigCol.GetColID()] = s.SubCols
}

// It checks if the expression is over a set of the columns eligible to the stitching.
// Namely, it contains columns of proper size with status Precomputed, Committed, or Verifiercol.
// It panics if the expression includes a mixture of eligible columns and columns with status Proof/VerifiyingKey/Ignored.
//
// If all the columns are verifierCol the expression is not eligible to the compilation.
// This is an expected behavior, since the verifier checks such expression by itself.
func IsExprEligible(
	isColEligible func(MultiSummary, ifaces.Column) bool,
	stitchings MultiSummary,
	board symbolic.ExpressionBoard,
) bool {
	metadata := board.ListVariableMetadata()
	hasAtLeastOneEligible := false
	allAreEligible := true
	allAreVeriferCol := true
	for i := range metadata {
		switch m := metadata[i].(type) {
		// reminder: [verifiercol.VerifierCol] , [column.Natural] and [column.Shifted]
		// all implement [ifaces.Column]
		case ifaces.Column: // it is a Committed, Precomputed or verifierCol
			natural := column.RootParents(m)[0]
			switch natural.(type) {
			case column.Natural: // then it is not a verifiercol
				allAreVeriferCol = false
				b := isColEligible(stitchings, m)

				hasAtLeastOneEligible = hasAtLeastOneEligible || b
				allAreEligible = allAreEligible && b
				if m.Size() == 0 {
					panic("found no columns in the expression")
				}
			}

		}

	}

	if hasAtLeastOneEligible && !allAreEligible {
		// 1. we expect no expression including Proof columns
		// 2. we expect no expression over ignored columns
		// 3. we expect no VerifiyingKey withing the stitching range.
		panic("the expression is not valid, it is mixed with invalid columns of status Proof/Ingnored/verifierKey")
	}
	if allAreVeriferCol {
		// 4. we expect no expression involving only and only the verifierCols.
		// We expect that this case wont happen.
		// Otherwise should be handled in the [github.com/consensys/linea-monorepo/prover/protocol/query] package.
		// Namely, Local/Global queries should be checked directly by the verifer.
		panic("all the columns in the expression are verifierCols, unsupported by the compiler")
	}

	return hasAtLeastOneEligible
}
