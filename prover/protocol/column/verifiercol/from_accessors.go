package verifiercol

import (
	"strings"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = FromAccessors{}

// FromAccessors is a [VerifierCol] implementation that is built by stacking the
// values of [ifaces.Accessor] on top of each others. It is the most type of
// column and subsumes almost all the other types.
type FromAccessors struct {
	// Accessors stores the list of accessors building the column.
	Accessors []ifaces.Accessor
	// Round_ caches the round value of the column.
	Round_ int
}

// NewFromAccessors instantiates a [FromAccessors] column from a list of
// [ifaces.Accessor]. The column's size must be a power of 2.
//
// You should not pass accessors of type [expressionAsAccessor] as their
// evaluation within a gnark circuit requires using the frontend.API which we
// can't access in the context currently.
func NewFromAccessors(accessors []ifaces.Accessor) ifaces.Column {
	if !utils.IsPowerOfTwo(len(accessors)) {
		panic("the column must be a power of two")
	}
	round := 0
	for i := range accessors {
		round = max(round, accessors[i].Round())
	}
	return FromAccessors{Accessors: accessors, Round_: round}
}

// Round returns the round ID of the column and implements the [ifaces.Column]
// interface.
func (f FromAccessors) Round() int {
	return f.Round_
}

// GetColID returns the column ID
func (f FromAccessors) GetColID() ifaces.ColID {
	accessorNames := make([]string, len(f.Accessors))
	for i := range f.Accessors {
		accessorNames[i] = f.Accessors[i].Name()
	}
	return ifaces.ColIDf("FROM_ACCESSORS_%v", strings.Join(accessorNames, "_"))
}

// MustExists implements the [ifaces.Column] interface and always returns true.
func (f FromAccessors) MustExists() {}

// Size returns the size of the colum and implements the [ifaces.Column]
// interface.
func (f FromAccessors) Size() int {
	return len(f.Accessors)
}

// GetColAssignment returns the assignment of the current column
func (f FromAccessors) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	res := make([]field.Element, len(f.Accessors))
	for i := range res {
		res[i] = f.Accessors[i].GetVal(run)
	}
	return smartvectors.NewRegular(res)
}

// GetColAssignment returns a gnark assignment of the current column
func (f FromAccessors) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {
	res := make([]frontend.Variable, len(f.Accessors))
	for i := range res {
		res[i] = f.Accessors[i].GetFrontendVariable(nil, run)
	}
	return res
}

// GetColAssignmentAt returns a particular position of the column
func (f FromAccessors) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return f.Accessors[pos].GetVal(run)
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (f FromAccessors) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return f.Accessors[pos].GetFrontendVariable(nil, run)
}

// IsComposite implements the [ifaces.Column] interface
func (f FromAccessors) IsComposite() bool {
	return false
}

// IsComposite implements the [symbolic.Metadata] interface
func (f FromAccessors) String() string {
	return string(f.GetColID())
}

// Split implements the [VerifierCol] interface
func (f FromAccessors) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column {
	// We don't call the accessor to ensure that the segment has the same round
	// definition as the original column.
	return FromAccessors{
		Accessors: f.Accessors[from:to],
		Round_:    f.Round_,
	}
}
