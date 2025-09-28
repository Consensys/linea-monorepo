package column

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// var _ ifaces.Column[T] = (*FakeColumn[T])(nil)

// FakeColumn is a dummy implementation of the [ifaces.Column[T]] interface
// it only implements the following methods:
//
//   - [ifaces.Column[T].GetColID]
//   - [ifaces.Column[T].String]
type FakeColumn[T zk.Element] struct {
	ID ifaces.ColID
}

func (fc *FakeColumn[T]) GetColID() ifaces.ColID {
	return fc.ID
}

func (fc *FakeColumn[T]) String() string {
	return string(fc.ID)
}

func (fc *FakeColumn[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentBase(run ifaces.Runtime) ifaces.ColAssignment {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentExt(run ifaces.Runtime) ifaces.ColAssignment {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) Round() int {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) Size() int {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) MustExists() {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	panic("unimplemented")
}

func (fc *FakeColumn[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	panic("unimplemented")
}

// The fake column is interpreted as composite column so that
func (fc *FakeColumn[T]) IsComposite() bool {
	return true
}

func (fc *FakeColumn[T]) IsBase() bool {
	panic("unimplemented")
}
