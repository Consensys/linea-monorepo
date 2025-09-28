package column

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Shifted represents a column that is obtains by (cyclic)-shifting the parent
// column by an Offset. This is useful to implement [github.com/consensys/linea-monorepo/prover/protocol/query.GlobalConstraint] constraints
// to express the "next" value of a column or the previous value of a column.
// For instance, say we want enforce that a column is assigned to a Fibonacci
// sequence: a[i+2] - a[i+1] - a[i] == 0. This can be done using the following
// constraint:
//
//	var (
//		aNextNext = ifaces.Column[T]AsVariable(column.Shift(a, 2))
//		aNext = ifaces.Column[T]AsVariable(column.Shift(a, 1))
//		a = ifaces.Column[T]AsVariable(a)
//	)
//
//	expr := aNextNext.Sub(aNext).Sub(a)
//
// The user should not directly instantiate the struct and should instead use
// the constructor [Shift]
type Shifted[T zk.Element] struct {
	Parent ifaces.Column[T]
	Offset int
}

// Shift constructs a [Shifted] column. The function performs a few checks and
// normalization before instantiating the column. If the user provides an offset
// of zero, the function is a no-op and returns the parent column. Since, the
// shift is cyclic, the offset is normalized. If the column is already a shift
// then the function fuses the two shifts into a single one for simplification.
func Shift[T zk.Element](parent ifaces.Column[T], offset int) ifaces.Column[T] {
	// input validation : in theory it is ok, but it is strange
	if offset <= -parent.Size() || offset >= parent.Size() {
		reducedOffset := utils.PositiveMod(offset, parent.Size())
		logrus.Debugf(
			"`Shift` : the offset is %v, but the size is %v. This is legit"+
				"in a context of splitting but otherwise it probably is not."+
				"wrapping the offset around -> %v",
			offset, parent.Size(), reducedOffset)
		offset = reducedOffset
	}

	if offset == 0 {
		// Skip zero shifts
		return parent
	}

	/*
		Special case, the parent is already shiifted, in that case
		we normalize the handle by using a single shift. This is not
		only an optimization : the splitting compiler does not support
		having Shift(Repeat(Shift)).
	*/
	if parentShift, ok := parent.(Shifted[T]); ok {
		totalOffset := parentShift.Offset + offset

		// This can happen when combining offsets within the splitter compiler
		if totalOffset >= parent.Size() {
			totalOffset -= parent.Size()
		}
		return Shift(parentShift.Parent, totalOffset)
	}

	return Shifted[T]{
		Parent: parent,
		Offset: offset,
	}
}

// Size returns the size of the column, as required by the [ifaces.Column[T]]
// interface.
func (s Shifted[T]) Size() int {
	return s.Parent.Size()
}

// GetColID implements the [ifaces.Column[T]] interface and returns the string
// identifier of the column. The ColID is obtained as SHIFT_<Offset>_<ParentID>.
func (s Shifted[T]) GetColID() ifaces.ColID {
	return ifaces.ColIDf("%v_%v", getNodeRepr(s), s.Parent.GetColID())
}

// MustExists validates the construction of the natural handle and implements
// the [ifaces.Column[T]] interface.
func (s Shifted[T]) MustExists() {
	s.Parent.MustExists()
}

// Round retuns the round of definition of the column. See [ifaces.Column[T]] as
// method implements the interface.
func (s Shifted[T]) Round() int {
	return s.Parent.Round()
}

// IsComposite implements [ifaces.Column[T]], by definition, it is not a
// composite thus it shall always return true
func (s Shifted[T]) IsComposite() bool { return true }

// GetColAssignment implements [ifaces.Column[T]]. The function resolves the
// assignment of the parent column and rotates it according to the offset.
func (s Shifted[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return s.Parent.GetColAssignment(run).RotateRight(-s.Offset)
}

// GetColAssignmentGnark implements [ifaces.Column[T]] and works like
// GetColAssignment.
func (s Shifted[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {
	parent := s.Parent.GetColAssignmentGnark(run) // [a b c d e f g h]
	res := make([]T, len(parent))
	for i := range res {
		posParent := utils.PositiveMod(i+s.Offset, len(parent))
		res[i] = parent[posParent]
	}
	return res
}

func (s Shifted[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	if s.IsBase() {
		parent := s.Parent.GetColAssignmentGnark(run) // [a b c d e f g h]
		res := make([]T, len(parent))
		for i := range res {
			posParent := utils.PositiveMod(i+s.Offset, len(parent))
			res[i] = parent[posParent]
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("requested base elements but column is defined over the extension")
	}
}

func (s Shifted[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	parent := s.Parent.GetColAssignmentGnarkExt(run) // [a b c d e f g h]
	res := make([]gnarkfext.E4Gen[T], len(parent))
	for i := range res {
		posParent := utils.PositiveMod(i+s.Offset, len(parent))
		res[i] = parent[posParent]
	}
	return res
}

// GetColAssignmentAt gets a particular entry of the shifted column. The query
// is delegated to the underlying column and the requested position is shifted
// according to the offset. This implements the [ifaces.Column[T]] interface.
func (s Shifted[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return s.Parent.GetColAssignmentAt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

func (s Shifted[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	return s.Parent.GetColAssignmentAtBase(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

func (s Shifted[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	return s.Parent.GetColAssignmentAtExt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

// GetColAssignmentGnarkAt gets the witness from the parent and performs a shift in the gnark circuit
// setting. The method implements the [ifaces.Column[T]] interface.
func (s Shifted[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	return s.Parent.GetColAssignmentGnarkAt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

func (s Shifted[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	return s.Parent.GetColAssignmentGnarkAtBase(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

func (s Shifted[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	return s.Parent.GetColAssignmentGnarkAtExt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

// String returns the ID of the column as a string and implements [ifaces.Column[T]]
// and [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]. It returns the same as [GetColID] but as a string
// (required by Metadata).
func (s Shifted[T]) String() string {
	return string(s.GetColID())
}

func (s Shifted[T]) IsBase() bool {
	return s.Parent.IsBase()
}
