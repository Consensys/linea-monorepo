package column

import (
	"reflect"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
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

// EvalExprColumn resolves an expression to a column assignment. The expression
// must be converted to a board prior to evaluating the expression.
//
//   - If the expression does not uses ifaces.Column as metadata, the function
//     will panic.
//
//   - If the expression contains several columns and they don't contain all
//     have the same size.
func EvalExprColumn(run ifaces.Runtime, board symbolic.ExpressionBoard) smartvectors.SmartVector {

	var (
		metadata = board.ListVariableMetadata()
		inputs   = make([]smartvectors.SmartVector, len(metadata))
		length   = ExprIsOnSameLengthHandles(&board)
	)

	// Attempt to recover the size of the
	for i := range inputs {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			inputs[i] = m.GetColAssignment(run)
		case coin.Info:
			v := run.GetRandomCoinField(m.Name)
			inputs[i] = smartvectors.NewConstant(v, length)
		case ifaces.Accessor:
			v := m.GetVal(run)
			inputs[i] = smartvectors.NewConstant(v, length)
		case variables.PeriodicSample:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		case variables.X:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		}
	}

	return board.Evaluate(inputs)
}

// ExprIsOnSameLengthHandles checks that all the variables of the expression
// that are [ifaces.Column] have the same size (and panics if it does not), then
// returns the match.
func ExprIsOnSameLengthHandles(board *symbolic.ExpressionBoard) int {

	var (
		metadatas = board.ListVariableMetadata()
		length    = 0
	)

	for _, m := range metadatas {
		switch metadata := m.(type) {
		case ifaces.Column:
			// Initialize the length with the first commitment
			if length == 0 {
				length = metadata.Size()
			}

			// Sanity-check the vector should all have the same length
			if length != metadata.Size() {
				utils.Panic("Inconsistent length for %v (has size %v, but expected %v)", metadata.GetColID(), metadata.Size(), length)
			}
		// The expression can involve random coins
		case coin.Info, variables.X, variables.PeriodicSample, ifaces.Accessor:
			// Do nothing
		default:
			utils.Panic("unknown type %T", metadata)
		}
	}

	// No commitment were found in the metadata, thus this call is broken
	if length == 0 {
		utils.Panic("declared a handle from an expression which does not contains any handle")
	}

	return length
}
