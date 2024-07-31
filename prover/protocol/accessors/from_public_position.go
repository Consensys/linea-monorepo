package accessors

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// FromPublicColumn refers to a position of a public column
type FromPublicColumn struct {
	// Info points to the underlying coin on which the accessor points to.
	Col column.Natural
	// Pos indexes the pointed position in the coin.
	Pos int
}

// NewFromPublicColumn constructs an [ifaces.Accessor] refering to the row #pos
// of the column `col`. The provided column must be public and the position must
// exist. Note that if the compiler later decide to make the column internal to
// the prover, the accessor will become invalid and this will panic at verifier
// time with a "column does not exist" error.
//
// The function accepts only Natural columns and panics if they are not of this
// type.
func NewFromPublicColumn(col ifaces.Column, pos int) ifaces.Accessor {

	nat, isNat := col.(column.Natural)
	if !isNat {
		utils.Panic("expected a [%T], got [%T]", column.Natural{}, col)
	}

	if !nat.Status().IsPublic() {
		panic("expected an coin.IntegerVec")
	}
	if nat.Size() <= pos {
		utils.Panic("the column has size %v, but requested position %v", nat.Size(), pos)
	}
	return &FromPublicColumn{
		Col: nat,
		Pos: pos,
	}
}

// Name implements [ifaces.Accessor]
func (c *FromPublicColumn) Name() string {
	return fmt.Sprintf("FROM_COLUMN_POSITION_ACCESSOR_%v_%v", c.Col.ID, c.Pos)
}

// String implements [github.com/consensys/zkevm-monorepo/prover/symbolic.Metadata]
func (c *FromPublicColumn) String() string {
	return c.Name()
}

// GetVal implements [ifaces.Accessor]
func (c *FromPublicColumn) GetVal(run ifaces.Runtime) field.Element {
	return run.GetColumnAt(c.Col.ID, c.Pos)
}

// GetFrontendVariable implements [ifaces.Accessor]
func (c *FromPublicColumn) GetFrontendVariable(_ frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {
	return circ.GetColumnAt(c.Col.ID, c.Pos)
}

// AsVariable implements the [ifaces.Accessor] interface
func (c *FromPublicColumn) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(c)
}

// Round implements the [ifaces.Accessor] interface
func (c *FromPublicColumn) Round() int {
	return c.Col.Round()
}
