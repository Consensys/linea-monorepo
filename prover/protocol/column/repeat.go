package column

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

// Repeat refers to an underlying column repeated it several time : useful for lookups.
// For instance, if P = (1, 2, 3, 4) then Repeat(P, 2) would be (1, 2, 3, 4, 1, 2, 3, 4)
type Repeated struct {
	Nb     int
	Parent ifaces.Column
}

// Create a repeat, check the inputs
func Repeat(parent ifaces.Column, nb int) ifaces.Column {

	if nb == 1 {
		logrus.Debugf("attempted to repeat column %v with nb=1, skipping", parent.GetColID())
		return parent
	}

	if nb < 1 {
		utils.Panic("can't repeat %v, %v times", parent.GetColID(), nb)
	}

	// Input validation
	if !utils.IsPowerOfTwo(nb) {
		utils.Panic("expected a power of two, got %v", nb)
	}

	/*
		In case, the parent is itself a repeat, we can simplify it
		with a single repeat.
	*/
	if parentRepeat, ok := parent.(Repeated); ok {
		totalNb := parentRepeat.Nb * nb
		return Repeat(parentRepeat.Parent, totalNb)
	}

	return Repeated{Nb: nb, Parent: parent}
}

/*
Returns the number of rows of the derivated poly
*/
func (r Repeated) Size() int {
	return r.Parent.Size() * r.Nb
}

/*
Returns a string representative
*/
func (r Repeated) GetColID() ifaces.ColID {
	return ifaces.ColIDf("%v_%v", getNodeRepr(r), r.Parent.GetColID())
}

/*
Defers to the parent
*/
func (r Repeated) MustExists() {
	r.Parent.MustExists()
}

/*
Defers to the parent
*/
func (r Repeated) Round() int {
	return r.Parent.Round()
}

// Get the witness from the parent and recopy it several time
func (r Repeated) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {

	if !utils.IsPowerOfTwo(r.Nb) {
		utils.Panic("not a power of two %v", r.Nb)
	}

	res := make([]field.Element, r.Size())
	p := r.Parent.GetColAssignment(run)
	parSize := p.Len()
	p.WriteInSlice(res[:p.Len()])

	// The below algorithm assumes that Nb is a power of two
	for i := 1; i < r.Nb; i *= 2 {
		copy(res[i*parSize:2*i*parSize], res[:i*parSize])
	}

	return smartvectors.NewRegular(res)
}

// Get a particular entry of the column
func (r Repeated) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return r.Parent.GetColAssignmentAt(run, pos%r.Parent.Size())
}

// Get the witness from the parent and recopy it several time
func (r Repeated) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {

	if !utils.IsPowerOfTwo(r.Nb) {
		utils.Panic("not a power of two %v", r.Nb)
	}

	res := make([]frontend.Variable, r.Size())
	copy(res, r.Parent.GetColAssignmentGnark(run))
	parSize := r.Parent.Size()

	// The below algorithm assumes that Nb is a power of two
	for i := 1; i < r.Nb; i *= 2 {
		copy(res[i*parSize:2*i*parSize], res[:i*parSize])
	}

	return res
}

// Get a particular entry of the column
func (r Repeated) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return r.Parent.GetColAssignmentGnarkAt(run, utils.PositiveMod(pos, r.Parent.Size()))
}

func (r Repeated) IsComposite() bool { return true }

/*
Returns the name of the column as a string
*/
func (r Repeated) String() string {
	return string(r.GetColID())
}
