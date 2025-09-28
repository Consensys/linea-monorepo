package column

import (
	"fmt"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// var _ ifaces.Column[T] = &Natural{}

// Natural represents a [ifaces.Column[T]] that has been directly declared in the
// corresponding [github.com/consensys/linea-monorepo/prover/protocol/wizard.CompiledIOP] or [github.com/consensys/linea-monorepo/prover/protocol/wizard.Builder]
// object. The struct should not be constructed directly and should be
// constructed from the [github.com/consensys/linea-monorepo/prover/protocol/wizard.CompiledIOP]
type Natural[T zk.Element] struct {
	// The ID of the column
	ID ifaces.ColID
	// position contains the indexes of the column in the store.
	position columnPosition
	// store points to the Store[T] where the column is registered. It is accessed
	// to fetch static informations about the column such as its size or its
	// status.
	store  *Store[T]
	isBase bool
}

// Size returns the size of the column, as required by the [ifaces.Column[T]]
// interface.
func (n Natural[T]) Size() int {
	return n.store.GetSize(n.ID)
}

// GetColID implements the [ifaces.Column[T]] interface and returns the string
// identifier of the column.
func (n Natural[T]) GetColID() ifaces.ColID {
	return n.ID
}

// MustExists validates the construction of the natural handle and implements
// the [ifaces.Column[T]] interface.
func (n Natural[T]) MustExists() {
	// store pointer is set
	if n.store == nil {
		utils.Panic("no entry for store in %v", n.GetColID())
	}

	s := n.store

	// check the positions matches
	storedPos := s.indicesByNames.MustGet(n.ID)
	if n.position != storedPos {
		utils.Panic("mismatched position has %v, but stored was %v", n.position, storedPos)
	}
}

// Round retuns the round of definition of the column. See [ifaces.Column[T]] as
// method implements the interface.
func (n Natural[T]) Round() int {
	return n.position.round
}

// IsComposite implements [ifaces.Column[T]], by definition, it is not a
// composite thus it shall always return true
func (n Natural[T]) IsComposite() bool { return false }

// GetColAssignment implements [ifaces.Column[T]]
func (n Natural[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return run.GetColumn(n.ID)
}

// GetColAssignmentAt implements [ifaces.Column[T]]
func (n Natural[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return run.GetColumnAt(n.ID, pos)
}

func (n Natural[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	return run.GetColumnAtBase(n.ID, pos)
}

func (n Natural[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	return run.GetColumnAtExt(n.ID, pos)
}

// GetColAssignmentGnark implements [ifaces.Column[T]]
func (n Natural[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {
	return run.GetColumn(n.ID)
}

func (n Natural[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	if n.isBase {
		return run.GetColumn(n.ID), nil
	} else {
		return nil, fmt.Errorf("requested base elements but column defined over field extensions")
	}
}

func (n Natural[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	return run.GetColumnExt(n.ID)
}

// GetColAssignmentGnarkAt implements [ifaces.Column[T]]
func (n Natural[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	return run.GetColumnAt(n.ID, utils.PositiveMod(pos, n.Size()))
}

func (n Natural[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	if n.isBase {
		return run.GetColumnAt(n.ID, utils.PositiveMod(pos, n.Size())), nil
	} else {
		var a T
		return a, fmt.Errorf("requested base elements but column defined over field extensions")
	}

}

func (n Natural[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	return run.GetColumnAtExt(n.ID, utils.PositiveMod(pos, n.Size()))
}

// String returns the ID of the column as a string and implements [ifaces.Column[T]]
// and [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (n Natural[T]) String() string {
	return string(n.GetColID()) + "_" + strconv.Itoa(n.Round()) + "_" + strconv.Itoa(n.Size())
}

// Status returns the status of the column. It is only implemented for Natural
// and not by the other composite columns.
func (n Natural[T]) Status() Status {
	return n.store.Status(n.ID)
}

func (n Natural[T]) IsBase() bool {
	return n.isBase
}

// SetPragma sets the pragma for a given column name.
func (n Natural[T]) SetPragma(pragma string, data any) {
	n.store.SetPragma(n.ID, pragma, data)
}

// GetPragma returns the pragma for a given column name.
func (n Natural[T]) GetPragma(pragma string) (any, bool) {
	return n.store.GetPragma(n.ID, pragma)
}

// GetStore[T]Unsafe returns the internal store pointer of the column. It is unsafe to
// use.
func (n Natural[T]) GetStoreUnsafe() *Store[T] {
	return n.store
}
