package column

import (
	"reflect"

	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

const (
	shift       string = "SHIFT"
	interleaved string = "INTERLEAVED"
	// Generalizes the concept of Natural to
	// the case of verifier defined columns
	nonComposite string = "NONCOMPOSITE"
)

// Constructs a natural column. input validation. Not exported, use
// [store.AddToRound] which will output a well-formed object. This ensures the
// invariant that a `Natural` always have a store.
func newNatural(name ifaces.ColID, position columnPosition, store *Store) Natural {
	if len(name) == 0 {
		utils.Panic("empty name")
	}
	if store == nil {
		utils.Panic("null store (%v)", name)
	}
	return Natural{ID: name, position: position, store: store}
}

// RootParents returns the underlying base [Natural] of the current handle. If
// the provided [Column] `h` is an [Interleaved] or a derivative of an
// [Interleaved], the function returns the list of all the underlying [Natural]
// columns.
func RootParents(h ifaces.Column) []ifaces.Column {

	if !h.IsComposite() {
		return []ifaces.Column{h}
	}

	switch inner := h.(type) {
	case Natural:
		// No changes
		return []ifaces.Column{h}
	case Shifted:
		return RootParents(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// StackOffset sums all the offsets contained in the handle and return the result
//
// If `h` is an [Interleaved] or derive from an [Interleaved] column, it will
// expect that the stacked offset of its parent is zero. (i.e, we should always
// shift an interleave but never interleave a shift. In practice, this does not
// cause issues as we do not have that in the arithmetization). This restriction
// is motivated by the fact that the "offset" would not be defined in this
// situation.
func StackOffsets(h ifaces.Column) int {

	if !h.IsComposite() {
		return 0
	}

	switch inner := h.(type) {
	case Natural:
		// No changes
		return 0
	case Shifted:
		return inner.Offset + StackOffsets(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// NbLeaves returns the number of underlying [Natural] columns for `h`. If `h`
// is neither an [Interleaved] nor derived from an [Interleaved], the function
// returns 1.
func NbLeaves(h ifaces.Column) int {
	switch inner := h.(type) {
	case Natural:
		// No changes
		return 1
	case Shifted:
		return NbLeaves(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}
