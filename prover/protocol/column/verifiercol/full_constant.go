package verifiercol

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// Represents a constant column
type ConstCol struct {
	f    field.Element
	size int
}

// NewConstCol creates a new ConstCol column
func NewConstantCol(f field.Element, size int) ifaces.Column {
	return ConstCol{f: f, size: size}
}

// Returns the round of definition of the column (always zero)
// Even though this is more of a convention than a meaningful
// return value.
func (cc ConstCol) Round() int {
	return 0
}

// Returns a generic name from the column. Defined from the coin's.
func (cc ConstCol) GetColID() ifaces.ColID {
	return ifaces.ColIDf("CONSTCOL_%v_%v", cc.f.String(), cc.size)
}

// Always return true
func (cc ConstCol) MustExists() {}

// Returns the size of the column
func (cc ConstCol) Size() int {
	return cc.size
}

// Returns a constant smart-vector
func (cc ConstCol) GetColAssignment(_ ifaces.Runtime) ifaces.ColAssignment {
	return smartvectors.NewConstant(cc.f, cc.size)
}

// Returns the column as a list of gnark constants
func (cc ConstCol) GetColAssignmentGnark(_ ifaces.GnarkRuntime) []frontend.Variable {
	res := make([]frontend.Variable, cc.size)
	for i := range res {
		res[i] = cc.f
	}
	return res
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return cc.GetColAssignment(run).Get(pos)
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return cc.GetColAssignmentGnark(run)[pos]
}

// Since the column is directly defined from the
// values of a random coin it does not count as a
// composite column.
func (cc ConstCol) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (cc ConstCol) String() string {
	return string(cc.GetColID())
}

// Splits the column and return a handle of it
func (cc ConstCol) Split(comp *wizard.CompiledIOP, from, to int) ifaces.Column {

	if to < from || to-from > cc.Size() {
		utils.Panic("Can't split %++v into [%v, %v]", cc, from, to)
	}

	// Copy the underlying cc, and assigns the new from and to
	return NewConstantCol(cc.f, to-from)
}
