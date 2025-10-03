package verifiercol

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// compile check to enforce the struct to belong to the corresponding interface
// var _ VerifierCol = FromAccessors{}

// FromAccessors is a [VerifierCol] implementation that is built by stacking the
// values of [ifaces.Accessor] on top of each others. It is the most type of
// column and subsumes almost all the other types.
type FromAccessors[T zk.Element] struct {
	// Accessors stores the list of accessors building the column.
	Accessors []ifaces.Accessor[T]
	Padding   fext.Element
	Size_     int
	// Round_ caches the round value of the column.
	Round_ int
}

func (f FromAccessors[T]) IsBase() bool {
	return f.Accessors[0].IsBase()
}

func (f FromAccessors[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (f FromAccessors[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {

	if pos >= f.Size_ {
		utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
	}

	if pos >= len(f.Accessors) {
		return f.Padding
	}

	return f.Accessors[pos].GetValExt(run)
}

func (f FromAccessors[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	//TODO implement me
	panic("implement me")
}

func (f FromAccessors[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

func (f FromAccessors[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (f FromAccessors[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewFromAccessors instantiates a [FromAccessors] column from a list of
// [ifaces.Accessor]. The column's size must be a power of 2.
//
// You should not pass accessors of type [expressionAsAccessor] as their
// evaluation within a gnark circuit requires using the frontend.API which we
// can't access in the context currently.
func NewFromAccessors[T zk.Element](accessors []ifaces.Accessor[T], padding fext.Element, size int) ifaces.Column[T] {
	if !utils.IsPowerOfTwo(size) {
		utils.Panic("the column must be a power of two (size=%v)", size)
	}
	round := 0
	for i := range accessors {
		round = max(round, accessors[i].Round())
	}
	return FromAccessors[T]{Accessors: accessors, Round_: round, Padding: padding, Size_: size}
}

// Round returns the round ID of the column and implements the [ifaces.Column[T]]
// interface.
func (f FromAccessors[T]) Round() int {
	return f.Round_
}

// GetColID returns the column ID
func (f FromAccessors[T]) GetColID() ifaces.ColID {
	accessorNames := make([]string, len(f.Accessors))
	for i := range f.Accessors {
		accessorNames[i] = f.Accessors[i].Name()
	}
	return ifaces.ColIDf("FROM_ACCESSORS_%v_PADDING=%v_SIZE=%v", strings.Join(accessorNames, "_"), f.Padding.String(), f.Size_)
}

// MustExists implements the [ifaces.Column[T]] interface and always returns true.
func (f FromAccessors[T]) MustExists() {}

// Size returns the size of the colum and implements the [ifaces.Column[T]]
// interface.
func (f FromAccessors[T]) Size() int {
	return f.Size_
}

// GetColAssignment returns the assignment of the current column
func (f FromAccessors[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	res := make([]fext.Element, len(f.Accessors))
	for i := range res {
		res[i] = f.Accessors[i].GetValExt(run)
	}
	return smartvectors.RightPaddedExt(res, f.Padding, f.Size_)
}

// GetColAssignment returns a gnark assignment of the current column
func (f FromAccessors[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {

	res := make([]T, f.Size_)
	for i := range f.Accessors {
		res[i] = f.Accessors[i].GetFrontendVariable(nil, run)
	}

	for i := len(f.Accessors); i < f.Size_; i++ {
		// TODO @thomas mixed ext base, fix it
		// res[i] = f.Padding
	}

	return res
}

// GetColAssignmentAt returns a particular position of the column
func (f FromAccessors[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	//TODO implement me
	panic("implement me")
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (f FromAccessors[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	if pos >= f.Size_ {
		utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
	}

	// if pos >= len(f.Accessors) {
	// 	return f.Padding
	// }

	return f.Accessors[pos].GetFrontendVariable(nil, run)
}

// IsComposite implements the [ifaces.Column[T]] interface
func (f FromAccessors[T]) IsComposite() bool {
	return false
}

// String implements the [symbolic.Metadata] interface
func (f FromAccessors[T]) String() string {
	return string(f.GetColID())
}

// Split implements the [VerifierCol] interface
func (f FromAccessors[T]) Split(_ *wizard.CompiledIOP[T], from, to int) ifaces.Column[T] {

	if from >= len(f.Accessors) {
		// The reason we don't want to remove the size from the name here is that
		// these columns tend to only exist as compilation artefacts.
		return NewConstantColExt[T](f.Padding, to-from, "")
	}

	var subAccessors = f.Accessors[from:]

	if to < len(f.Accessors) {
		subAccessors = f.Accessors[from:to]
	}

	// We don't call the accessor to ensure that the segment has the same round
	// definition as the original column.
	return FromAccessors[T]{
		Accessors: subAccessors,
		Round_:    f.Round_,
		Padding:   f.Padding,
		Size_:     to - from,
	}
}

func (f FromAccessors[T]) GetFromAccessorsFields() (accs []ifaces.Accessor[T], padding fext.Element) {
	return f.Accessors, f.Padding
}
