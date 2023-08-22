package wizardutils

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Parse the metadata of an expression and returns the first round where this function can be evaluated
func LastRoundToEval(comp *wizard.CompiledIOP, expr *symbolic.Expression) int {
	board := expr.Board()
	metadatas := board.ListVariableMetadata()

	maxRound := 0

	for _, m := range metadatas {
		switch metadata := m.(type) {
		case ifaces.Column:
			maxRound = utils.Max(maxRound, metadata.Round())
		// The expression can involve random coins
		case coin.Info:
			maxRound = utils.Max(maxRound, metadata.Round)
			// assert the coin is an expression
			if metadata.Type != coin.Field {
				utils.Panic("The coin %v should be of type `Field`", metadata.Name)
			}
		case variables.X, variables.PeriodicSample:
			// Do nothing
		case *ifaces.Accessor:
			maxRound = utils.Max(maxRound, metadata.Round)
		default:
			panic("unreachable")
		}
	}

	return maxRound
}

// All coms must have the same length, returns the size
func ExprIsOnSameLengthHandles(comp *wizard.CompiledIOP, expr *symbolic.Expression) int {

	board := expr.Board()
	metadatas := board.ListVariableMetadata()

	length := 0
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
		case coin.Info, *variables.X, *variables.PeriodicSample, *ifaces.Accessor:
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

// maximal round of declaration for a list of commitment
func MaxRound(handles ...ifaces.Column) int {
	res := 0
	for _, handle := range handles {
		res = utils.Max(res, handle.Round())
	}
	return res
}

// Assert that all the handles have the same length
func AssertAllHandleSameLength(list ...ifaces.Column) int {
	if len(list) == 0 {
		panic("passed an empty leaf")
	}

	res := list[0].Size()
	for i := range list {
		if list[i].Size() != res {
			utils.Panic("the column %v (size %v) does not have the same size as column 0 (size %v)", i, list[i].Size(), res)
		}
	}

	return res
}
