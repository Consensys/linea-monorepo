package wizard

import (
	"strconv"

	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

var _ Column = &ColumnConstant{}

// ColumnConstant refers to an abstract variable X over which all polynomials are defined. It
// implements the [Column] interface and represents a column storing the
// successive powers of the generator root of unity.
//
// This column is used to cancel expressions at specific points.
type ColumnConstant struct {
	size int
	v    field.Element
}

// Construct a new variable for coin
func NewColumnConstant(size int, v field.Element) ColumnConstant {
	return ColumnConstant{size: size, v: v}
}

// to implement symbolic.Metadata
func (c ColumnConstant) String() string {
	return "column-constant/size=" + strconv.Itoa(c.size) + "/v=" + c.v.String()
}

// Returns an evaluation of the X, possibly over a coset. Pass
// `GetAssignment(size, 0, 0, false)` to directly evaluate over a coset
func (c ColumnConstant) GetAssignment(run Runtime) sv.SmartVector {
	return sv.NewConstant(c.v, c.size)
}

// Evaluate the variable, but not over a coset
func (c ColumnConstant) GetAssignmentGnark(_ frontend.API, _ RuntimeGnark) []frontend.Variable {
	res := make([]frontend.Variable, c.size)
	for i := range res {
		res[i] = c.v
	}
	return res
}

func (c ColumnConstant) Round() int {
	return 0
}

func (c ColumnConstant) Size() int {
	return c.size
}

func (c ColumnConstant) Shift(n int) Column {
	return c
}
