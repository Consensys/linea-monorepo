package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// asFromAccessors is an ad-hoc interface that serves to identify [verifiercol.FromAccessors]
// without creating a cyclic dependency.
type asFromAccessors[T zk.Element] interface {
	GetFromAccessorsFields() (accs []ifaces.Accessor[T], padding field.Element)
}

// FromPublicColumn refers to a position of a public column
type FromPublicColumn[T zk.Element] struct {
	// Info points to the underlying coin on which the accessor points to.
	Col column.Natural[T]
	// Pos indexes the pointed position in the coin.
	Pos int
}

func (c *FromPublicColumn[T]) IsBase() bool {
	//TODO implement me
	panic("implement me")
}

func (c *FromPublicColumn[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (c *FromPublicColumn[T]) GetValExt(run ifaces.Runtime) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (c *FromPublicColumn[T]) GetFrontendVariableBase(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (c *FromPublicColumn[T]) GetFrontendVariableExt(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewFromPublicColumn constructs an [ifaces.Accessor] refering to the row #pos
// of the column `col`. The provided column must be public and the position must
// exist. Note that if the compiler later decide to make the column internal to
// the prover, the accessor will become invalid and this will panic at verifier
// time with a "column does not exist" error.
//
// The function accepts only Natural columns and panics if they are not of this
// type.
func NewFromPublicColumn[T zk.Element](col ifaces.Column[T], pos int) ifaces.Accessor[T] {

	if faccs, ok := col.(asFromAccessors[T]); ok {
		accs, pad := faccs.GetFromAccessorsFields()

		if pos >= len(accs) {
			return NewConstant[T](pad)
		}
		return accs[pos]
	}

	nat, isNat := col.(column.Natural[T])
	if !isNat {
		utils.Panic("expected a [%T], got [%T]", column.Natural[T]{}, col)
	}

	if !nat.Status().IsPublic() {
		panic("expected a public column")
	}
	if nat.Size() <= pos {
		utils.Panic("the column has size %v, but requested position %v", nat.Size(), pos)
	}
	return &FromPublicColumn[T]{
		Col: nat,
		Pos: pos,
	}
}

// Name implements [ifaces.Accessor]
func (c *FromPublicColumn[T]) Name() string {
	return fmt.Sprintf("FROM_COLUMN_POSITION_ACCESSOR_%v_%v", c.Col.ID, c.Pos)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (c *FromPublicColumn[T]) String() string {
	return c.Name()
}

// GetVal implements [ifaces.Accessor]
func (c *FromPublicColumn[T]) GetVal(run ifaces.Runtime) field.Element {
	return run.GetColumnAt(c.Col.ID, c.Pos)
}

// GetFrontendVariable implements [ifaces.Accessor]
func (c *FromPublicColumn[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	return circ.GetColumnAt(c.Col.ID, c.Pos)
}

// AsVariable implements the [ifaces.Accessor] interface
func (c *FromPublicColumn[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](c)
}

// Round implements the [ifaces.Accessor] interface
func (c *FromPublicColumn[T]) Round() int {
	return c.Col.Round()
}
