/*
Package projection implements the utilities for the projection query.

A projection query between sets (columnsA,filterA) and (columnsB,filterB) asserts whether
the columnsA filtered by filterA is the same as columnsB filtered by filterB, preserving the order.

Example:

FilterA = (1,0,0,1,1), ColumnA := (aO,a1,a2,a3,a4)

FiletrB := (0,0,1,0,0,0,0,0,1,1), ColumnB :=(b0,b1,b2,b3,b4,b5,b6,b7,b8,b9)

Thus we have,

ColumnA filtered by FilterA = (a0,a3,a4)

ColumnB filtered by FilterB  = (b2,b8,b9)

The projection query checks if a0 = b2, a3 = b8, a4 = b9

Note that the query imposes that:
  - the number of 1 in the filters are equal
  - the order of filtered elements is preserved
*/
package projection

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// The projection struct presents the new columns required for the query.
// Particularly, the new columns should impose a fix order.
// The imposed order would guarantee the order-preserving property of the query.
type projection struct {
	// for column Col, the accumulator accCol is built via the relations;
	// accCol[0]=Col[0] and
	// accCol[i] = col[i]+ accCol[i-1] where i stands for the index of the row.
	orderColA ifaces.Column // The accumulator for filterA
	orderColB ifaces.Column // The accumulator for filterB
}

// InsertProjection applies a projection query between sets (columnsA, filterA) and (columnsB,filterB).
//
// Note: The filters are supposed to be binary.
// These binary constraints are not handled here
// and should have been imposed before calling the projection query
func InsertProjection(
	comp *wizard.CompiledIOP, round int,
	queryName ifaces.QueryID,
	columnsA, columnsB []ifaces.Column,
	filterA, filterB ifaces.Column,
) {
	i := projection{}
	sizeA := filterA.Size()
	sizeB := filterB.Size()
	queryNameA := ifaces.QueryIDf(string(queryName), "A")
	queryNameB := ifaces.QueryIDf(string(queryName), "B")

	// Declare the columns
	i.orderColA = comp.InsertCommit(round, deriveName(string(queryName), "orderColA"), sizeA)
	i.orderColB = comp.InsertCommit(round, deriveName(string(queryName), "orderColB"), sizeB)

	// Declare the constrains on the order columns;  orderCol = filter +  shift(orderCol,-1)
	csOrderCol(comp, round, queryNameA, i.orderColA, filterA)
	csOrderCol(comp, round, queryNameB, i.orderColB, filterB)

	// Lookup queries for two sides inclusion (preserving the order via orderCol)
	comp.InsertInclusionDoubleConditional(round, queryNameA, columnsA, columnsB, filterA, filterB)
	comp.InsertInclusionDoubleConditional(round, queryNameB, columnsB, columnsA, filterB, filterA)

	comp.SubProvers.AppendToInner(
		round,
		func(run *wizard.ProverRuntime) {
			i.assignOrder(run, []ifaces.Column{filterA, filterB})
		},
	)

}

// It imposes the constraints between Filters and OrderColumns.
// Namely, the orderColumns should be the accumulators of filters
func csOrderCol(comp *wizard.CompiledIOP, round int, queryName ifaces.QueryID, orderCol, filter ifaces.Column) {
	// orderCol = filter +  shift(orderCol,-1)
	expr := ifaces.ColumnAsVariable(orderCol).
		Sub(ifaces.ColumnAsVariable(filter).Add(ifaces.ColumnAsVariable(column.Shift(orderCol, -1))))
	comp.InsertGlobal(round, ifaces.QueryIDf(string(queryName), "OrderImposing"), expr)
}

// It assigns the OrderColumns via the filters
func (order projection) assignOrder(run *wizard.ProverRuntime, filters []ifaces.Column) {
	for j := range filters {
		size := filters[j].Size()
		orderCol := make([]field.Element, size)
		filterWit := filters[j].GetColAssignment(run).IntoRegVecSaveAlloc()

		//sanity check
		for i := range filterWit {
			if filterWit[i] != field.One() && filterWit[i] != field.Zero() {
				utils.Panic("filters should be binary")

			}
		}

		orderCol[0] = filterWit[0]
		for i := 1; i < len(filterWit); i++ {
			orderCol[i].Add(&filterWit[i], &orderCol[i-1])
		}

		if j == 0 {
			run.AssignColumn(order.orderColA.GetColID(), smartvectors.RightZeroPadded(orderCol, size))
		} else if j == 1 {
			run.AssignColumn(order.orderColB.GetColID(), smartvectors.RightZeroPadded(orderCol, size))
		}

	}

}

// It derives column names
func deriveName(prefix, mainName string) ifaces.ColID {
	return ifaces.ColIDf("%v_%v", prefix, mainName)
}
