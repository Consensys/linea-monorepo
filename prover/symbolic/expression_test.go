package symbolic

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/stretchr/testify/require"
)

func TestReplayExpression(t *testing.T) {

	// Original dummy variable
	a, b, c, x := NewDummyVar("a"), NewDummyVar("b"), NewDummyVar("c"), NewDummyVar("x")
	a_, b_, c_, x_ := NewDummyVar("a_"), NewDummyVar("b_"), NewDummyVar("c_"), NewDummyVar("x_")

	expressions := []*Expression{
		a.Add(b),
		a.Add(a).Mul(b),
		a.Neg().Add(b).Neg().Mul(c).Add(a),
		a.Sub(b).Mul(c),
		a.Mul(a).Mul(b).Mul(a).Mul(c).Mul(c),
		a.Mul(NewPolyEval(x, []*Expression{a, b, a, a.Add(c)})),
	}

	// Random constants for the polyEvals
	var r field.Element
	r.SetRandom()

	witnesses := map[string]smartvectors.SmartVector{
		"a": smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8),
		"b": smartvectors.ForTest(8, 16, 32, 64, 128, 256, 512, 1024),
		"c": smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21),
		"x": smartvectors.NewConstant(r, 8),
	}

	witnesses_ := map[string]smartvectors.SmartVector{
		"a_": witnesses["a"].SubVector(0, 4),
		"b_": witnesses["b"].SubVector(0, 4),
		"c_": witnesses["c"].SubVector(0, 4),
		"x_": witnesses["x"].SubVector(0, 4),
	}

	translationMap := collection.NewMapping[string, *Expression]()
	translationMap.InsertNew("a", a_)
	translationMap.InsertNew("b", b_)
	translationMap.InsertNew("c", c_)
	translationMap.InsertNew("x", x_)

	for _, expr := range expressions {
		board := expr.Board()
		metadatas := board.ListVariableMetadata()
		inputs := make([]smartvectors.SmartVector, len(metadatas))
		for i := range metadatas {
			inputs[i] = witnesses[metadatas[i].String()]
		}

		replayed := expr.Replay(translationMap)

		// Check that the expression is valid
		replayed.AssertValid()

		replayedBoard := replayed.Board()
		replayedMetadata := replayedBoard.ListVariableMetadata()
		require.Len(t, replayedMetadata, len(metadatas))

		inputs_ := make([]smartvectors.SmartVector, len(metadatas))
		for i := range replayedMetadata {
			inputs_[i] = witnesses_[replayedMetadata[i].String()]
		}

		// Should be the same thing as the old board
		oldBoard := expr.Board()
		oldMetadata := oldBoard.ListVariableMetadata()
		require.Len(t, oldMetadata, len(metadatas))

		for i := range metadatas {
			require.Equal(t, metadatas[i].String(), oldMetadata[i].String())
			require.Equal(t, metadatas[i].String()+"_", replayedMetadata[i].String())
		}

		eval := board.Evaluate(inputs)
		replayedEval := replayedBoard.Evaluate(inputs_)
		oldEval := oldBoard.Evaluate(inputs)

		// The oldEval and eval should be consistent
		require.Equal(t, eval.Pretty(), oldEval.Pretty())
		require.Equal(t, eval.SubVector(0, 4).Pretty(), replayedEval.Pretty())
	}

}
