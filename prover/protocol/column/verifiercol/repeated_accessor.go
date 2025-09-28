package verifiercol

import (
	"errors"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// RepeatedAccessor is a [VerifierCol] implementation that is built by
// repeated Size times the same accessor. It will not work if the provided
// accessor require "api" to be evaluated in a gnark circuit.
type RepeatedAccessor[T zk.Element] struct {
	Accessor ifaces.Accessor[T]
	ColSize  int
}

func (f RepeatedAccessor[T]) IsBase() bool {
	return f.Accessor.IsBase()
}

func (f RepeatedAccessor[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	if !f.IsBase() {
		return field.Element{}, errors.New("not base")
	}

	return f.Accessor.GetVal(run), nil
}

func (f RepeatedAccessor[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	if f.IsBase() {
		v, err := f.Accessor.GetValBase(run)
		if err != nil {
			panic(err)
		}
		return fext.Lift(v)
	}

	return f.Accessor.GetValExt(run)
}

func (f RepeatedAccessor[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	//TODO implement me
	panic("implement me")
}

func (f RepeatedAccessor[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

func (f RepeatedAccessor[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (f RepeatedAccessor[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewRepeatedAccessor instantiates a [RepeatedAccessor] column from an
// [ifaces.Accessor] and a size.
func NewRepeatedAccessor[T zk.Element](accessor ifaces.Accessor[T], size int) ifaces.Column[T] {
	return RepeatedAccessor[T]{Accessor: accessor, ColSize: size}
}

// Round returns the round ID of the column and implements the [ifaces.Column[T]]
// interface.
func (f RepeatedAccessor[T]) Round() int {
	return f.Accessor.Round()
}

// GetColID returns the column ID
func (f RepeatedAccessor[T]) GetColID() ifaces.ColID {
	return ifaces.ColIDf("REPEATED_%v_SIZE=%v", f.Accessor.Name(), f.ColSize)
}

// MustExists implements the [ifaces.Column[T]] interface and always returns true.
func (f RepeatedAccessor[T]) MustExists() {}

// Size returns the size of the colum and implements the [ifaces.Column[T]]
// interface.
func (f RepeatedAccessor[T]) Size() int {
	return f.ColSize
}

// Split implements the [VerifierCol] interface
func (f RepeatedAccessor[T]) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column[T] {
	return NewRepeatedAccessor(f.Accessor, to-from)
}

// GetColAssignment returns the assignment of the current column
func (f RepeatedAccessor[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return smartvectors.NewConstant(f.Accessor.GetVal(run), f.Size())
}

// GetColAssignmentGnark returns a gnark assignment of the current column
func (f RepeatedAccessor[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {
	res := make([]T, f.Size())
	x := f.Accessor.GetFrontendVariable(nil, run)
	for i := range res {
		res[i] = x
	}
	return res
}

// GetColAssignmentAt returns a particular position of the column
func (f RepeatedAccessor[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return f.Accessor.GetVal(run)
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (f RepeatedAccessor[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	return f.Accessor.GetFrontendVariable(nil, run)
}

// IsComposite implements the [ifaces.Column[T]] interface
func (f RepeatedAccessor[T]) IsComposite() bool {
	return false
}

// String implements the [symbolic.Metadata] interface
func (f RepeatedAccessor[T]) String() string {
	return f.Accessor.Name()
}
