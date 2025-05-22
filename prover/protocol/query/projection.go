package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Projection represents a projection query. A projection query enforces that
// two sides A and B contains the same values on the same order at positions
// marked with a 1 in the corresponding filter. The query supports multi-ary
// projections: meaning that the A and B sides can have multiple columns and
// the query will read both sides left-to-right then top-to-bottom (e.g. like
// reading text in english). In that context, the two sides can have different
// "width": this allows the query to be used for flattening an array of columns
// or widening it.
//
// And on top of that, the query can also "tables": the query look over vector
// of values instead of just values. In that case, all parts of the two sides
// of multi-ary projection must have the same number of columns.
type Projection struct {
	Round int
	ID    ifaces.QueryID
	Inp   ProjectionMultiAryInput
}

// ProjectionInput is a collection of parameters to provide to a [Projection]
// query. It corresponds to the case where the projection query is "unary".
type ProjectionInput struct {
	// ColumnA and ColumnB are the columns of the left and right side. Each
	// entry of either corresponds to a column in the projected table.
	ColumnA, ColumnB []ifaces.Column
	// FilterA and FilterB are the filters of the ColumnA and ColumnB
	FilterA, FilterB ifaces.Column
}

// ProjectionMultiAryInput is a collection of parameters to provide to a
// [Projection] query in the general case.
type ProjectionMultiAryInput struct {
	A ProjectionSideMultiAry
	B ProjectionSideMultiAry
}

// ProjectionSideMultiAry is a collection of parameters to provide to a
// [Projection] query in the general case.
type ProjectionSideMultiAry struct {
	Columns [][]ifaces.Column
	Filters []ifaces.Column
}

// NewProjection constructs a projection. Will panic if it is mal-formed
func NewProjection(round int, id ifaces.QueryID, inp ProjectionInput) Projection {
	return NewProjectionMultiAry(round, id, ProjectionMultiAryInput{
		A: ProjectionSideMultiAry{
			Columns: [][]ifaces.Column{inp.ColumnA},
			Filters: []ifaces.Column{inp.FilterA},
		},
		B: ProjectionSideMultiAry{
			Columns: [][]ifaces.Column{inp.ColumnB},
			Filters: []ifaces.Column{inp.FilterB},
		},
	})
}

// NewProjectionMultiAry returns a new [Projection] query object.
func NewProjectionMultiAry(
	round int,
	id ifaces.QueryID,
	inp ProjectionMultiAryInput,
) Projection {

	if len(inp.A.Columns) == 0 || len(inp.B.Columns) == 0 {
		utils.Panic("A and B must have at least one table: len(A)=%v, len(B)=%v", len(inp.A.Columns), len(inp.B.Columns))
	}

	if len(inp.A.Columns[0]) == 0 || len(inp.B.Columns[0]) == 0 {
		utils.Panic("A and B must have at least one column: len(A[0])=%v, len(B[0])=%v", len(inp.A.Columns[0]), len(inp.B.Columns[0]))
	}

	var (
		numCols   = len(inp.A.Columns[0])
		numPartsA = len(inp.A.Columns)
		numPartsB = len(inp.B.Columns)
		sizeA     = inp.A.Columns[0][0].Size()
		sizeB     = inp.B.Columns[0][0].Size()
	)

	if len(inp.A.Filters) != numPartsA || len(inp.B.Filters) != numPartsB {
		utils.Panic("A and B must have the same number of filters: len(A)=%v, len(B)=%v", len(inp.A.Filters), len(inp.B.Filters))
	}

	for i := range inp.A.Columns {
		if len(inp.A.Columns[i]) != numCols {
			utils.Panic("All table must have the same number of columns: len(A[%v])=%v, len(A[0])=%v", i, len(inp.A.Columns[i]), numCols)
		}

		size := ifaces.AssertSameLength(inp.A.Columns[i]...)

		if i == 0 {
			sizeA = size
		}

		if size != sizeA {
			utils.Panic("All table must have the same number of columns: len(A[%v])=%v, len(A[0])=%v", i, len(inp.A.Columns), numCols)
		}

		if sizeA != inp.A.Filters[i].Size() {
			utils.Panic("A[%v] and its filter do not have the same column sizes", i)
		}
	}

	for i := range inp.B.Columns {
		if len(inp.B.Columns[i]) != numCols {
			utils.Panic("All table must have the same number of columns: len(B[%v])=%v, numCols=%v", i, len(inp.B.Columns[i]), numCols)
		}

		size := ifaces.AssertSameLength(inp.B.Columns[i]...)

		if i == 0 {
			sizeB = size
		}

		if size != sizeB {
			utils.Panic("All table must have the same number of columns: len(B[%v])=%v, len(B[0])=%v", i, len(inp.B.Columns), numCols)
		}

		if sizeB != inp.B.Filters[i].Size() {
			utils.Panic("B[%v] and its filter do not have the same column sizes", i)
		}
	}

	return Projection{Round: round, ID: id, Inp: inp}
}

// Name implements the [ifaces.Query] interface
func (p Projection) Name() ifaces.QueryID {
	return p.ID
}

// Check implements the [ifaces.Query] interface
func (p Projection) Check(run ifaces.Runtime) error {

	// The function is implemented by creating two iterator functions and
	// checking that they yield the same values.
	nextA := p.Inp.A.NextIterator(run)
	nextB := p.Inp.B.NextIterator(run)

	for {
		a, aOk := nextA()
		b, bOk := nextB()

		if !aOk && !bOk {
			return nil
		}

		if aOk != bOk {
			return fmt.Errorf("a and b must yield the same number of rows, a %v b %v", aOk, bOk)
		}

		if len(a) != len(b) {
			// Note: this is redundant with the constructor's check, this should
			// not be a runtime error.
			panic("A and B must yield the same number of columns")
		}

		for i := range a {
			if !a[i].Equal(&b[i]) {
				return fmt.Errorf("a and b must yield the same values, a=%v b=%v", vector.Prettify(a), vector.Prettify(b))
			}
		}
	}
}

// GnarkCheck implements the [ifaces.Query] interface. It will panic in this
// construction because we do not have a good way to check the query within a
// circuit
func (i Projection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an Projection query directly into the circuit")
}

// GetShiftedRelatedColumns returns the list of the [HornerParts.Selectors] found
// in the query. This is used to check if the query is compatible with
// Wizard distribution.
//
// Note: the fact that this method is implemented makes [Inclusion] satisfy
// an anonymous interface that is matched to detect queries that are
// incompatible with wizard distribution. So we should not rename or remove
// this implementation without doing the corresponding changes in the
// distributed package. Otherwise, this will silence the checks that we are
// doing.
func (p Projection) GetShiftedRelatedColumns() []ifaces.Column {

	res := []ifaces.Column{}

	for _, f := range p.Inp.A.Filters {
		if f.IsComposite() {
			res = append(res, f)
		}
	}

	for _, f := range p.Inp.B.Filters {
		if f.IsComposite() {
			res = append(res, f)
		}
	}

	for i := range p.Inp.A.Columns {
		for _, col := range p.Inp.A.Columns[i] {
			if col.IsComposite() {
				res = append(res, col)
			}
		}
	}

	for i := range p.Inp.B.Columns {
		for _, col := range p.Inp.B.Columns[i] {
			if col.IsComposite() {
				res = append(res, col)
			}
		}
	}

	return res
}

// NextIterator returns an iterator over the selected values of A in the
// assignment of the runtime. This function is useful as it addresses a common
// pattern for assignments.
func (p ProjectionSideMultiAry) NextIterator(run ifaces.Runtime) (next func() ([]field.Element, bool)) {

	var (
		currRow  = 0
		currPart = 0
	)

	return func() ([]field.Element, bool) {
		for {

			if currRow >= p.Filters[currPart].Size() {
				return nil, false
			}

			f := p.Filters[currPart].GetColAssignmentAt(run, currRow)

			if f.IsZero() {
				currPart++
				if currPart == len(p.Columns) {
					currPart = 0
					currRow++
				}
				continue
			}

			var res []field.Element
			for _, col := range p.Columns[currPart] {
				res = append(res, col.GetColAssignmentAt(run, currRow))
			}

			currPart++
			if currPart == len(p.Columns) {
				currPart = 0
				currRow++
			}

			return res, true
		}
	}
}
