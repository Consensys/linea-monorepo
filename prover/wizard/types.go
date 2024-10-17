package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// id is a general identifier type for every Wizard object (column, query, etc..).
// The id can be parsed as: <Obj type (1 byte) || Obj indentifier (48 bytes)>.
//
// The id is purely internal
type id uint64

// Item is an interface requiring the minimal methods for a struct to be
// considered a component of a wizard protocol. They can be an integer, a vector,
// a field element, a query or a constraint or an action from the prover or the
// verifier.
type Item interface {
	// Round returns the index of the interaction round in which the Item
	// is defined: meaning the round during which the Item's value becomes
	// available to the verifier.
	Round() int
	// id returns an internal and unique identified representing the item
	id() id
	// Explain returns a string description of the item and the context of its
	// creation. This can be used for debugging
	Explain() string
	// Tags returns the list of the user/compiler provided tags attached to the
	// Item.
	Tags() []string
}

// Accessor is an interface that represents a backed field element value which
// can be part of a protocol. The [Accessor] value are always visible to the
// verifier. For instance, a random challenge coin is an instance of Accessor
// because it represents a field element. Another example would be the result
// of a polynomial query opening.
type Accessor interface {
	Item
	// The requirement of Metadata ensures that an Accessor can be used as part
	// of a symbolic expression as a Variable.
	symbolic.Metadata
	// GetVal returns the field element value of the accessor. It should not be
	// called before the value is available otherwise the method implementation
	// will panic.
	GetVal(run Runtime) field.Element
	// GetValGnark is as GetVal but returns the Accessor value in the context
	// of a gnark verifier circuit.
	GetValGnark(api frontend.API, run RuntimeGnark) frontend.Variable
}

// Coin is an interface representing public random coins as part of a protocol.
// Coins can be of any sort (field elements, vector of small integers representing
// positions in a vector of polynomials from a specific set etc...)
type Coin interface {
	Item
	symbolic.Metadata
	// Round returns the
	sample(fs *fiatshamir.State) any
}
