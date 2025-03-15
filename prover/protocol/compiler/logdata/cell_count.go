package logdata

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CellCount is a struct storing numerous metrics pertaining to a
// compiled-IOP.
type CellCount struct {
	NumColumnsCommitted        int
	NumColumnsProof            int
	NumColumnsPrecomputed      int
	NumCellsCommitted          int
	NumCellsProof              int
	NumCellsPrecomputed        int
	NumResultsInnerProduct     int
	NumResultUnivariate        int
	NumResultLocalOpening      int
	NumResultLogDerivativeSum  int
	NumResultGrandProduct      int
	NumResultHorner            int
	NumQueriesInnerProduct     int
	NumQueriesUnivariate       int
	NumQueriesLocalOpening     int
	NumQueriesLogDerivativeSum int
	NumQueriesGrandProduct     int
	NumQueriesHorner           int
}

// CountCells counts the cells occuring in the provided wizard-IOP.
func CountCells(comp *wizard.CompiledIOP) *CellCount {

	var (
		listOfColumns = comp.Columns.AllKeys()
		listOfQueries = comp.QueriesParams.AllKeys()
		cellCount     = &CellCount{}
	)

	for _, colName := range listOfColumns {

		var (
			col    = comp.Columns.GetHandle(colName)
			status = comp.Columns.Status(colName)
			size   = col.Size()
		)

		switch status {
		case column.Committed:
			cellCount.NumCellsCommitted += size
			cellCount.NumColumnsCommitted++
		case column.Proof:
			cellCount.NumCellsProof += size
			cellCount.NumColumnsProof++
		case column.Precomputed:
			cellCount.NumCellsPrecomputed += size
			cellCount.NumColumnsPrecomputed++
		}
	}

	for _, qName := range listOfQueries {

		q := comp.QueriesParams.Data(qName)

		switch qParams := q.(type) {
		case query.UnivariateEval:
			cellCount.NumQueriesUnivariate++
			cellCount.NumResultUnivariate += len(qParams.Pols)
		case query.InnerProduct:
			cellCount.NumQueriesInnerProduct++
			cellCount.NumResultsInnerProduct += len(qParams.Bs)
		case query.LocalOpening:
			cellCount.NumQueriesLocalOpening++
			cellCount.NumResultLocalOpening++
		case query.LogDerivativeSum:
			cellCount.NumQueriesLogDerivativeSum++
			cellCount.NumResultLogDerivativeSum++
		case query.GrandProduct:
			cellCount.NumQueriesGrandProduct++
			cellCount.NumResultGrandProduct++
		case *query.Horner:
			cellCount.NumQueriesHorner++
			cellCount.NumResultHorner++
		}
	}

	return cellCount
}

// TotalCells returns the total number of cells recorded in the wizard.
func (c CellCount) TotalCells() int {
	return c.NumCellsCommitted +
		c.NumCellsPrecomputed +
		c.NumCellsProof
}
