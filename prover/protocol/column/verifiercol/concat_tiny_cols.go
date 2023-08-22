package verifiercol

import (
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = ConcatTinyColumns{}

// Represents a column obtained by concatenating the
// values of several proof elements.
type ConcatTinyColumns struct {
	columns    []ifaces.Column
	paddedSize int
	paddingVal field.Element
	round      int
}

// NewConcatTinyColumns creates a new ConcatTinyColumns.
// The columns must all have a length of "1"
func NewConcatTinyColumns(
	comp *wizard.CompiledIOP,
	paddedSize int,
	paddingVal field.Element,
	cols ...ifaces.Column,
) ifaces.Column {

	// Compute the interaction rounds
	round := 0

	// Check the length of the columns
	for _, col := range cols {
		// sanity-check
		col.MustExists()

		// sanity check the publicity of the column
		AssertIsPublicCol(comp, col)

		// sanity-check : we only support length 1 tiny columns
		if col.Size() != 1 {
			utils.Panic("expected column to have length 1, but got %v for %v", col.Size(), col.GetColID())
		}

		// update the round
		round = utils.Max(round, col.Round())
	}

	// Then, the total length must not exceed the the PaddedSize
	if paddedSize < len(cols) {
		utils.Panic("the target length (=%v) is smaller than the given columns (=%v)", paddedSize, len(cols))
	}

	return ConcatTinyColumns{
		columns:    cols,
		paddedSize: paddedSize,
		paddingVal: paddingVal,
		round:      round,
	}
}

// Round returns the round ID of the column
func (c ConcatTinyColumns) Round() int {
	return c.round
}

// GetColID returns the col id
func (c ConcatTinyColumns) GetColID() ifaces.ColID {
	colIDs := make([]string, len(c.columns))
	for i := range colIDs {
		colIDs[i] = string(c.columns[i].GetColID())
	}
	return ifaces.ColIDf(
		"CTC_%v_PAD_%v_UPTO_%v",
		strings.Join(colIDs, "_"),
		c.paddingVal.String(),
		c.paddedSize,
	)
}

// MustExists always pass
func (c ConcatTinyColumns) MustExists() {}

// Returns the size of the colum,
func (c ConcatTinyColumns) Size() int {
	return c.paddedSize
}

// GetColAssignment returns the assignment of the current column
func (c ConcatTinyColumns) GetColAssignment(run ifaces.Runtime) (res ifaces.ColAssignment) {

	vec := make([]field.Element, len(c.columns))
	for i, col := range c.columns {
		vec[i] = col.GetColAssignmentAt(run, 0)
	}

	return smartvectors.RightPadded(vec, c.paddingVal, c.paddedSize)
}

// GetColAssignment returns a gnark assignment of the current column
func (c ConcatTinyColumns) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {

	res := make([]frontend.Variable, c.paddedSize)
	for i, col := range c.columns {
		res[i] = col.GetColAssignmentGnarkAt(run, 0)
	}

	// fill the remaining values with the padding value
	for i := len(c.columns); i < c.paddedSize; i++ {
		res[i] = c.paddingVal
	}

	return res
}

// Returns a particular position of the coin value
func (c ConcatTinyColumns) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return c.GetColAssignment(run).Get(pos)
}

// Returns a particular position of the coin value
func (c ConcatTinyColumns) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return c.GetColAssignmentGnark(run)[pos]
}

func (c ConcatTinyColumns) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (c ConcatTinyColumns) String() string {
	return string(c.GetColID())
}

// Split the FromYs by restricting to a range
func (c ConcatTinyColumns) Split(comp *wizard.CompiledIOP, from, to int) (split ifaces.Column) {

	if to < from || to-from > c.Size() {
		utils.Panic("Can't split %++v into [%v, %v]", c, from, to)
	}

	// edge-case, the slice if fully in the padding
	if from > len(c.columns) {
		return ConcatTinyColumns{
			columns:    []ifaces.Column{},
			paddingVal: c.paddingVal,
			paddedSize: to - from,
		}
	}

	// else, the slice is fully on the columns
	if to < len(c.columns) {
		return ConcatTinyColumns{
			columns:    c.columns[from:to],
			paddedSize: to - from,
		}
	}

	// in the average case, there is an overlap on both side
	return ConcatTinyColumns{
		columns:    c.columns[from:],
		paddingVal: c.paddingVal,
		paddedSize: to - from,
	}
}
