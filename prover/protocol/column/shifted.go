package column

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

/*
Shifted is the representation of a shifted commitment.
Let Parent = (a, b, c, d, e, f) (in evaluation form),
then Shifted(Parent, offset=1) would be  (b, c, d, e, f, a).
As the example suggests, this is a circular shift.
*/
type Shifted struct {
	Parent ifaces.Column
	Offset int
}

/*
Shift returns a shifted version of the column
*/
func Shift(parent ifaces.Column, offset int) ifaces.Column {
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
		logrus.Debugf("warning : attempted to shift %v by zero. Ignoring\n", parent.GetColID())
		return parent
	}

	/*
		Special case, the parent is already shiifted, in that case
		we normalize the handle by using a single shift. This is not
		only an optimization : the splitting compiler does not support
		having Shift(Repeat(Shift)).
	*/
	if parentShift, ok := parent.(Shifted); ok {
		totalOffset := parentShift.Offset + offset

		// This can happen when combining offsets within the splitter compiler
		if totalOffset >= parent.Size() {
			logrus.Debugf("Combining shifts results in an overflow %v + %v >= %v. Wrapping around.", parentShift.Offset, offset, parent.Size())
			totalOffset -= parent.Size()
		}
		return Shift(parentShift.Parent, totalOffset)
	}

	/*
		Special case, the parent is a REPEAT. In that case, we can
		intervert the two. for simplification. Although, it is not
		immediately trivial this actually works. This is not only an
		optimization : the splitting compiler does not support having
		Shift(Repeat(Shift)).
	*/
	if parentRepeat, ok := parent.(Repeated); ok {
		nbRepeat := parentRepeat.Nb
		return Repeat(Shift(parentRepeat.Parent, offset), nbRepeat)
	}

	return Shifted{
		Parent: parent,
		Offset: offset,
	}
}

/*
Does not change the size
*/
func (s Shifted) Size() int {
	return s.Parent.Size()
}

/*
String repr of a shifted handle
*/
func (s Shifted) GetColID() ifaces.ColID {
	return ifaces.ColIDf("%v_%v", getNodeRepr(s), s.Parent.GetColID())
}

/*
Defers to the parent
*/
func (s Shifted) MustExists() {
	s.Parent.MustExists()
}

/*
Defers to the parent
*/
func (s Shifted) Round() int {
	return s.Parent.Round()
}

func (s Shifted) IsComposite() bool { return true }

/*
Get the witness from the parent and performs a shift
*/
func (s Shifted) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return s.Parent.GetColAssignment(run).RotateRight(-s.Offset)
}

/*
Get the witness from the parent and performs a shift
*/
func (s Shifted) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {
	parent := s.Parent.GetColAssignmentGnark(run) // [a b c d e f g h]
	res := make([]frontend.Variable, len(parent))
	for i := range res {
		posParent := utils.PositiveMod(i+s.Offset, len(parent))
		res[i] = parent[posParent]
	}
	return res
}

/*
Get a particular entry of the shifted column. The query is
delegated to the underlying column and the requested position
is shifted according to the offset.
*/
func (s Shifted) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return s.Parent.GetColAssignmentAt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

/*
Get the witness from the parent and performs a shift in the gnark circuit setting
*/
func (s Shifted) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return s.Parent.GetColAssignmentGnarkAt(run, utils.PositiveMod(pos+s.Offset, s.Parent.Size()))
}

/*
Returns the name of the column as a string
*/
func (s Shifted) String() string {
	return string(s.GetColID())
}
