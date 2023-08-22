package column

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

/*
Natural directly represents a committed vector without
modification.
*/
type Natural struct {
	ID       ifaces.ColID
	position commitPosition
	store    *Store
}

/*
Accesses the size of the column. The natural column stores its own size
*/
func (n Natural) Size() int {
	return n.store.GetSize(n.ID)
}

/*
Just return the name as a string repr of the handle
*/
func (n Natural) GetColID() ifaces.ColID {
	return n.ID
}

/*
Validates the construction of the natural handle
*/
func (n Natural) MustExists() {
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

/*
Get the round from the inner state of the handle
*/
func (n Natural) Round() int {
	return n.position.round
}

// Implements InnerHandle, by definition, it is not a composite
// thus it shall always return true
func (n Natural) IsComposite() bool { return false }

// Directly get the witness from the name
func (n Natural) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return run.GetColumn(n.ID)
}

// Directly get the witness from the name
func (n Natural) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return run.GetColumnAt(n.ID, pos)
}

// Returns a gnark assignment to the column
func (n Natural) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {
	return run.GetColumn(n.ID)
}

// Directly get the witness from the name
func (n Natural) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return run.GetColumnAt(n.ID, utils.PositiveMod(pos, n.Size()))
}

// Returns the name of the column as a string
func (n Natural) String() string {
	return string(n.GetColID())
}
