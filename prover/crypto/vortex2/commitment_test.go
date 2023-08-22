package vortex2

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/stretchr/testify/require"
)

func TestLinearCombination(t *testing.T) {

	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)

	params := NewParams(blowUpFactor, polySize, nPolys, ringsis.StdParams)

	// Polynomials to commit to
	polys := make([]smartvectors.SmartVector, nPolys)
	ys := make([]field.Element, nPolys)
	for i := range polys {
		polys[i] = smartvectors.Rand(polySize)
		ys[i] = smartvectors.Interpolate(polys[i], x)
	}

	// Make a linear combination of the poly
	lc := smartvectors.PolyEval(polys, randomCoin)

	// Generate the proof
	proof := params.OpenWithLC(polys, randomCoin)

	// Evaluate the two on a random-ish point. Should
	// yield the same result.
	y0 := smartvectors.Interpolate(lc, x)
	y1 := smartvectors.Interpolate(proof.LinearCombination, x)

	require.Equal(t, y0, y1)
}

func TestVortexOneCommitment(t *testing.T) {

	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys, ringsis.StdParams)

	// Polynomials to commit to
	polys := make([]smartvectors.SmartVector, nPolys)
	ys := make([]field.Element, nPolys)
	for i := range polys {
		polys[i] = smartvectors.Rand(polySize)
		ys[i] = smartvectors.Interpolate(polys[i], x)
	}

	// Commits to it
	commitment, committedMatrix := params.Commit(polys)

	// Generate the proof
	proof := params.OpenWithLC(polys, randomCoin)
	proof.WithEntryList([]CommittedMatrix{committedMatrix}, entryList)

	// Check the proof
	err := params.VerifyOpening([]Commitment{commitment}, proof, x, [][]field.Element{ys}, randomCoin, entryList)
	require.NoError(t, err)
}

func TestVortexNCommitment(t *testing.T) {

	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists := make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	commitments := make([]Commitment, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range commitments {
		commitments[j], committedMatrices[j] = params.Commit(polyLists[j])
	}

	// Generate the proof
	proof := params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)

	// Check the proof
	err := params.VerifyOpening(commitments, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)
}

func TestVortexNCommitmentWithMerkle(t *testing.T) {

	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)
	params.WithMerkleMode(mimc.NewMiMC)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists := make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	roots := make([]hashtypes.Digest, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range trees {
		committedMatrices[j], trees[j], _ = params.CommitMerkle(polyLists[j])
		roots[j] = trees[j].Root
	}

	// Generate the proof
	proof := params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)
	proof.WithMerkleProof(trees, entryList)

	// Check the proof
	err := params.VerifyMerkle(roots, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)
}

func TestVortexOneCommitmentNoSis(t *testing.T) {

	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys, ringsis.StdParams)
	params.RemoveSis(mimc.NewMiMC)

	require.True(t, params.HasSisReplacement())

	// Polynomials to commit to
	polys := make([]smartvectors.SmartVector, nPolys)
	ys := make([]field.Element, nPolys)
	for i := range polys {
		polys[i] = smartvectors.Rand(polySize)
		ys[i] = smartvectors.Interpolate(polys[i], x)
	}

	// Commits to it
	commitment, committedMatrix := params.Commit(polys)

	// Generate the proof
	proof := params.OpenWithLC(polys, randomCoin)
	proof.WithEntryList([]CommittedMatrix{committedMatrix}, entryList)

	// Check the proof
	err := params.VerifyOpening([]Commitment{commitment}, proof, x, [][]field.Element{ys}, randomCoin, entryList)
	require.NoError(t, err)
}

func TestVortexNCommitmentNoSis(t *testing.T) {

	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)
	params.RemoveSis(mimc.NewMiMC)

	require.True(t, params.HasSisReplacement())

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists := make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	commitments := make([]Commitment, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range commitments {
		commitments[j], committedMatrices[j] = params.Commit(polyLists[j])
	}

	// Generate the proof
	proof := params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)

	// Check the proof
	err := params.VerifyOpening(commitments, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)
}

func TestVortexNCommitmentWithMerkleNoSis(t *testing.T) {

	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2

	x := field.NewElement(478)
	randomCoin := field.NewElement(1523)
	entryList := []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)
	params.WithMerkleMode(mimc.NewMiMC)
	params.RemoveSis(mimc.NewMiMC)

	require.True(t, params.HasSisReplacement())

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists := make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	roots := make([]hashtypes.Digest, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range trees {
		committedMatrices[j], trees[j], _ = params.CommitMerkle(polyLists[j])
		roots[j] = trees[j].Root
	}

	// Generate the proof
	proof := params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)
	proof.WithMerkleProof(trees, entryList)

	// Check the proof
	err := params.VerifyMerkle(roots, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)
}
