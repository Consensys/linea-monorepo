package column_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/stretchr/testify/require"
)

func TestVariableMetaData(t *testing.T) {
	store := column.NewStore()
	V := store.AddToRound(0, ifaces.ColIDf("V"), 4, column.Committed)
	v := ifaces.ColumnAsVariable(V)
	vBoard := v.Board()

	// This will panic if the casting does not work
	var _ ifaces.Column = vBoard.ListVariableMetadata()[0].(ifaces.Column)

	require.True(t, v.IsVariable())

	a := column.Shift(V, 1)
	b := column.Shift(V, 1)
	require.True(t, a == b)

	a = column.Shift(column.Repeat(V, 2), 1)
	b = column.Shift(column.Repeat(V, 2), 1)
	require.True(t, a == b)

	a = column.Shift(column.Repeat(V, 2), 2)
	b = column.Shift(column.Repeat(V, 2), 1)
	require.True(t, a != b)

	a = column.Shift(column.Repeat(V, 2), 1)
	b = column.Shift(column.Repeat(V, 4), 1)
	require.True(t, a != b)
}

func TestReprAndDerivation(t *testing.T) {

	store := column.NewStore()
	V := store.AddToRound(0, ifaces.ColID("V"), 4, column.Committed)
	V2 := store.AddToRound(0, ifaces.ColID("V2"), 4, column.Committed)
	v := smartvectors.ForTest(1, 2, 3, 4)
	v2 := smartvectors.ForTest(11, 12, 13, 14)
	x := field.NewElement(546)

	// Test for shifting
	{
		shifted1 := smartvectors.ForTest(2, 3, 4, 1)
		Shifted1 := column.Shift(V, 1)

		expectedY := smartvectors.Interpolate(shifted1, x)

		cachedXs := collection.NewMapping[string, field.Element]()
		cachedXs.InsertNew("", x)

		x_ := column.DeriveEvaluationPoint(Shifted1, "", cachedXs, x)

		allDSBranches := column.AllDownStreamBranches(Shifted1)
		roots := column.RootParents(Shifted1)

		require.Len(t, allDSBranches, 1)
		require.Len(t, x_, 1)
		require.Len(t, roots, 1)

		// Should find the downstreams in the cached map
		for j, ds := range allDSBranches {
			require.Equal(t, x_[j], cachedXs.MustGet(ds))
		}
		// Evaluate the derived claim : should equal the expected Y
		derivedY := smartvectors.Interpolate(v, x_[0])

		finalYs := collection.NewMapping[string, field.Element]()
		finalYs.InsertNew(column.DerivedYRepr(allDSBranches[0], V), derivedY)

		// Test that we recovered
		recoveredY := column.VerifyYConsistency(Shifted1, "", cachedXs, finalYs)
		require.Equal(t, expectedY.String(), recoveredY.String())
		require.Equal(t, expectedY.String(), derivedY.String())
	}

	// Test for repeat
	{
		repeat2 := smartvectors.ForTest(1, 2, 3, 4, 1, 2, 3, 4)
		Repeat2 := column.Repeat(V, 2)

		expectedY := smartvectors.Interpolate(repeat2, x)
		cachedXs := collection.NewMapping[string, field.Element]()
		cachedXs.InsertNew("", x)

		x_ := column.DeriveEvaluationPoint(Repeat2, "", cachedXs, x)

		allDSBranches := column.AllDownStreamBranches(Repeat2)
		roots := column.RootParents(Repeat2)

		require.Len(t, allDSBranches, 1)
		require.Len(t, x_, 1)
		require.Len(t, roots, 1)

		// Should find the downstreams in the cached map
		for j, ds := range allDSBranches {
			require.Equal(t, x_[j], cachedXs.MustGet(ds))
		}

		// Evaluate the derived claim : should equal the expected Y
		derivedY := smartvectors.Interpolate(v, x_[0])

		finalYs := collection.NewMapping[string, field.Element]()
		finalYs.InsertNew(column.DerivedYRepr(allDSBranches[0], V), derivedY)

		// Test that we recovered
		recoveredY := column.VerifyYConsistency(Repeat2, "", cachedXs, finalYs)
		require.Equal(t, expectedY.String(), recoveredY.String())
		require.Equal(t, expectedY.String(), derivedY.String())

	}

	// Test for interleaving
	{
		interleaved := smartvectors.ForTest(1, 11, 2, 12, 3, 13, 4, 14)
		Interleaved := column.Interleave(V, V2)
		expectedY := smartvectors.Interpolate(interleaved, x)

		cachedXs := collection.NewMapping[string, field.Element]()
		cachedXs.InsertNew("", x)
		xs := column.DeriveEvaluationPoint(Interleaved, "", cachedXs, x)

		allDSBranches := column.AllDownStreamBranches(Interleaved)
		roots := column.RootParents(Interleaved)

		require.Len(t, allDSBranches, 2)
		require.Len(t, xs, 2)
		require.Len(t, roots, 2)

		// Should find the downstreams in the cached map
		// Should find the downstreams in the cached map
		for j, ds := range allDSBranches {
			require.Equal(t, xs[j], cachedXs.MustGet(ds))
		}

		// plugin the derived evaluations whose consistency is to be checked
		// with `expectedY`
		finalYs := collection.NewMapping[string, field.Element]()

		derivedYV := smartvectors.Interpolate(v, xs[0])
		finalYs.InsertNew(column.DerivedYRepr(allDSBranches[0], V), derivedYV)
		derivedYV2 := smartvectors.Interpolate(v2, xs[1])
		finalYs.InsertNew(column.DerivedYRepr(allDSBranches[1], V2), derivedYV2)

		// Test that we recovered the right value
		recoveredY := column.VerifyYConsistency(Interleaved, "", cachedXs, finalYs)
		require.Equal(t, expectedY.String(), recoveredY.String())

	}
}
