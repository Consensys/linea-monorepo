package verifiercol

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// compile check to enforce the struct to belong to the corresponding interface
// var _ VerifierCol = FromYs{}

// Represents a column populated by alleged evaluations of arrange of columns
type FromYs[T zk.Element] struct {
	// The list of the evaluated column in the same order
	// as we like to layout the currently-described column
	Ranges []ifaces.ColID
	// The Query from which we shall select the evaluations
	Query query.UnivariateEval[T]
	// Remember the round in which the query was made
	Round_ int
}

func (fys FromYs[T]) IsBase() bool {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (fys FromYs[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// Construct a new column from a univariate query and a list of of ifaces.ColID
// If passed a column that is not part of the query. It will not panic but it will
// return a zero entry. This is the expected behavior when given a shadow column
// from the vortex compiler but otherwise this is a bug.
func NewFromYs[T zk.Element](comp *wizard.CompiledIOP[T], q query.UnivariateEval[T], ranges []ifaces.ColID) ifaces.Column[T] {

	// All the names in the range should also be part of the query.
	// To make sure of this, we build the following map.
	nameMap := map[ifaces.ColID]struct{}{}
	for _, polName := range q.Pols {
		nameMap[polName.GetColID()] = struct{}{}
	}

	for _, rangeName := range ranges {
		if _, ok := nameMap[rangeName]; !ok && !strings.Contains(string(rangeName), "SHADOW") {
			utils.Panic("NewFromYs : %v is not part of the query %v", rangeName, q.QueryID)
		}
	}

	// Make sure that the query is indeed registered in the current wizard.
	comp.QueriesParams.MustExists(q.QueryID)
	round := comp.QueriesParams.Round(q.QueryID)

	res := FromYs[T]{
		Ranges: ranges,
		Query:  q,
		Round_: round,
	}

	return res
}

// Returns the round of definition of the column
func (fys FromYs[T]) Round() int {
	return fys.Round_
}

// Returns a generic name from the column. Defined from the coin's.
func (fys FromYs[T]) GetColID() ifaces.ColID {
	return ifaces.ColIDf("FYS_%v", fys.Query.QueryID)
}

// Always return true. We sanity-check the existence of the
// random coin prior to constructing the object.
func (fys FromYs[T]) MustExists() {}

// Return the size of the fys
func (fys FromYs[T]) Size() int {
	return len(fys.Ranges)
}

// Returns the coin's value as a column assignment
func (fys FromYs[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {

	queryParams := run.GetParams(fys.Query.QueryID).(query.UnivariateEvalParams[T])

	// Map the alleged evaluations to their respective commitment names
	yMap := map[ifaces.ColID]fext.Element{} // TODO@yao:check fext or field?
	for i, polName := range fys.Query.Pols {
		yMap[polName.GetColID()] = queryParams.ExtYs[i]
	}

	// This will leaves the columns missing from the query to zero.
	res := make([]fext.Element, len(fys.Ranges))
	for i, name := range fys.Ranges {
		res[i] = yMap[name]
	}

	return smartvectors.NewRegularExt(res)
}

// Returns the coin's value as a column assignment
func (fys FromYs[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {

	queryParams := run.GetParams(fys.Query.QueryID).(query.GnarkUnivariateEvalParams[T])

	// Map the alleged evaluations to their respective commitment names
	yMap := map[ifaces.ColID]T{}
	for i, polName := range fys.Query.Pols {
		yMap[polName.GetColID()] = queryParams.Ys[i]
	}

	// This will leave some of the columns to nil
	res := make([]T, len(fys.Ranges))
	for i, name := range fys.Ranges {
		if y, found := yMap[name]; found {
			res[i] = y
		} else {
			// Set it to zero explicitly
			res[i] = *zk.ValueOf[T](0)
		}
	}

	return res
}

// Returns a particular position of the coin value
func (fys FromYs[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return fys.GetColAssignment(run).Get(pos)
}

// Returns a particular position of the coin value
func (fys FromYs[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	return fys.GetColAssignmentGnark(run)[pos]
}

func (fys FromYs[T]) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (fys FromYs[T]) String() string {
	return string(fys.GetColID())
}

// Split the FromYs by restricting to a range
func (fys FromYs[T]) Split(comp *wizard.CompiledIOP[T], from, to int) ifaces.Column[T] {
	return NewFromYs[T](comp, fys.Query, fys.Ranges[from:to])
}
