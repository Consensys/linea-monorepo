package column

import (
	"reflect"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

/*
Constructs a natural column. input validation. Not exported, use `store.AddToRound`
which will output a well-formed object. This ensures the invariant that a `Natural`
always have a store.
*/
func newNatural(name ifaces.ColID, position commitPosition, store *Store) Natural {
	if len(name) == 0 {
		utils.Panic("empty name")
	}
	if store == nil {
		utils.Panic("null store (%v)", name)
	}
	return Natural{ID: name, position: position, store: store}
}

/*
Returns the underlying base "natural" of the current handle.
*/
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
	case Repeated:
		return RootParents(inner.Parent)
	case Interleaved:
		res := []ifaces.Column{}
		for _, parent := range inner.Parents {
			res = append(res, RootParents(parent)...)
		}
		return res
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

/*
Sums all the offsets contained in the handle and return the result

If there is an interleaving, it will expect that the stacked offset
of its parent is zero. (i.e, we should always shift an interleave but
never interleave a shift. In practice, this does not pause issues
as we do not have that in the arithmetization.)
*/
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
	case Repeated:
		// Return the shift of the parent
		return StackOffsets(inner.Parent)
	case Interleaved:
		/*
			Assume that all the subcolumns return zero. Otherwise,
			we do not know how to handle it anyway. It is fine to
			do this because we are in the context of the zkevm
		*/
		for _, parent := range inner.Parents {
			if StackOffsets(parent) != 0 {
				panic("Assumption broken : encountered an interleaving of shifted columns")
			}
		}
		return 0
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

/*
Gives the number of leaves for the given handle
*/
func NbLeaves(h ifaces.Column) int {
	switch inner := h.(type) {
	case Natural:
		// No changes
		return 1
	case Shifted:
		return NbLeaves(inner.Parent)
	case Repeated:
		return NbLeaves(inner.Parent)
	case Interleaved:
		res := 0
		for _, parent := range inner.Parents {
			res += NbLeaves(parent)
		}
		return res
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}
