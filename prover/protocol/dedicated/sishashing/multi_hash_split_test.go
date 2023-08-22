package sishashing

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiSisHashSplit(t *testing.T) {

	// dimensions of the test
	var (
		NUM_INPUTS           int = 8
		NUM_HASHES_PER_SHARE int = 2
		NUM_SHARES           int = 2
		proverRuntime        *wizard.ProverRuntime
	)

	key := ringsis.StdParams.GenerateKey(NUM_INPUTS)

	// the test-vector we use has

	splitIndices := make([]int, NUM_SHARES)
	for i := range splitIndices {
		splitIndices[i] = i * NUM_INPUTS
	}

	testvecs := make([][][]field.Element, NUM_SHARES)

	for share := range testvecs {
		testvecs[share] = make([][]field.Element, NUM_HASHES_PER_SHARE)
		for i := range testvecs[share] {
			testvecs[share][i] = vector.Rand(NUM_INPUTS)
		}
	}

	digest := [][]field.Element{}

	for share := range testvecs {
		shareDigest := []field.Element{}
		for _, testvec := range testvecs[share] {
			h := key.Hash(testvec)
			shareDigest = append(shareDigest, h...)
		}
		digest = append(digest, shareDigest)
	}

	// Concatenate the chunks to get the "full" columns
	preimageFullCols := make([][]field.Element, NUM_HASHES_PER_SHARE)

	for i := range preimageFullCols {
		toConcat := []field.Element{}
		for share := range testvecs {
			toConcat = append(toConcat, testvecs[share][i]...)
		}
		preimageFullCols[i] = toConcat
	}

	compiled := wizard.Compile(func(build *wizard.Builder) {
		// the wizard simply calls RingSISCheck over prover-defined columns
		preimages := make([]ifaces.Column, len(testvecs))
		for i := range testvecs {
			preimages[i] = build.RegisterCommit(limbPreimageName(i), NUM_SHARES*NUM_INPUTS*key.NumLimbs())
		}

		allegedSisHashes := make([]ifaces.Column, NUM_SHARES)
		for share := range digest {
			allegedSisHashes[share] = build.InsertPublicInput(0, allegedSisHashName(share), len(digest[share]))
		}

		MultiSplitSplitCheck("MULTI_SIS_SPLIT", build.CompiledIOP, &key, preimages, allegedSisHashes, splitIndices)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {
		// Save a pointer to the runtime so that we
		// can inspect it later on in the tests
		proverRuntime = assi

		// Inject the precomputed values into it
		for i := range testvecs {
			assi.AssignColumn(limbPreimageName(i), smartvectors.NewRegular(key.LimbSplit(preimageFullCols[i])))
		}

		for share := range digest {
			assi.AssignColumn(allegedSisHashName(share), smartvectors.NewRegular(digest[share]))
		}
	})

	mergerCoin := proverRuntime.GetRandomCoinField("MULTI_SIS_SPLIT_MERGER")
	mergedKey := splitAndMergedKey(&key, splitIndices, NUM_INPUTS*NUM_SHARES, mergerCoin)

	// Checks that the mergedSisHashes was correctly computed
	{

		mergedLaidOutKeyFromRuntime := smartvectors.IntoRegVec(proverRuntime.GetColumn("MULTI_SIS_SPLIT_MERGED_KEY"))
		mergedLaidOutKeyRecomputed := mergedKey.LaidOutKey()

		assert.Equalf(t,
			vector.Prettify(mergedLaidOutKeyRecomputed),
			vector.Prettify(mergedLaidOutKeyFromRuntime),
			"the laid out key did not match",
		)

		mergedSisHashesFromRuntime := smartvectors.IntoRegVec(proverRuntime.GetColumn("MULTI_SIS_SPLIT_MERGED_HASHES"))
		mergedDualSisHashesFromRuntime := smartvectors.IntoRegVec(proverRuntime.GetColumn("MULTI_SIS_SPLIT_DUAL"))

		mergedSisHashRecomputed := []field.Element{}
		mergedDualSisHashRecomputed := []field.Element{}

		for i := range preimageFullCols {
			h := mergedKey.Hash(preimageFullCols[i])
			mergedSisHashRecomputed = append(mergedSisHashRecomputed, h...)
			dual := mergedKey.HashModXnMinus1(mergedKey.LimbSplit(preimageFullCols[i]))
			mergedDualSisHashRecomputed = append(mergedDualSisHashRecomputed, dual...)
		}

		assert.Equal(t,
			vector.Prettify(mergedSisHashRecomputed),
			vector.Prettify(mergedSisHashesFromRuntime),
			"the merged sis hash did not match",
		)

		assert.Equalf(t,
			vector.Prettify(mergedDualSisHashRecomputed),
			vector.Prettify(mergedDualSisHashesFromRuntime),
			"the dual hashes did not match",
		)
	}

	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)

}

func limbPreimageName(i int) ifaces.ColID {
	return ifaces.ColIDf("PREIMAGE_LIMBS_%v", i)
}

func allegedSisHashName(chunkNo int) ifaces.ColID {
	return ifaces.ColIDf("ALLEGED_HASH_%v", chunkNo)
}

func splitAndMergedKey(
	sisKey *ringsis.Key,
	splitIndicesField []int,
	totalSizeField int,
	mergerCoin field.Element,
) ringsis.Key {

	multiplier := field.One()

	// this generates a key with a larger buffer. This occasionate a bunch of
	// unnecessary code run but this is ok enough for testing. The content of
	// the key get overwritten before returning.
	res := sisKey.GenerateKey(totalSizeField)

	// sanity-check on the parameters of the lattice instance
	if res.NumLimbs() < res.Degree {
		utils.Panic("lattice instance does not allow for self-recursion")
	}

	// compute the number of sis polynomials needed to hash a field element
	numPolyToHashOneField := res.NumLimbs() / res.Degree

	for chunkNo, startAtField := range splitIndicesField {
		startAtPoly := startAtField * numPolyToHashOneField

		// Case for the last chunk
		var stopAtPoly int
		switch {
		case chunkNo+1 < len(splitIndicesField):
			stopAtPoly = splitIndicesField[chunkNo+1] * numPolyToHashOneField
		case chunkNo+1 == len(splitIndicesField):
			stopAtPoly = totalSizeField * numPolyToHashOneField
		default:
			panic("unreachable")
		}

		for polID := startAtPoly; polID < stopAtPoly; polID++ {
			vector.ScalarMul(res.A[polID], sisKey.A[polID-startAtPoly], multiplier)
			vector.ScalarMul(res.Ag[polID], sisKey.Ag[polID-startAtPoly], multiplier)
		}

		multiplier.Mul(&multiplier, &mergerCoin)
	}

	return res
}

func TestLaidOutKeyChunk(t *testing.T) {

	params := ringsis.Params{LogTwoBound_: 32, LogTwoDegree: 2}
	chunkSize := 8
	key := params.GenerateKey(2 * chunkSize)
	splitIndices := []int{0, chunkSize}

	laidout := key.LaidOutKey()
	chunk0 := smartvectors.IntoRegVec(LaidOutKeyChunk(&key, splitIndices, 0, 2*chunkSize))
	chunk1 := smartvectors.IntoRegVec(LaidOutKeyChunk(&key, splitIndices, 1, 2*chunkSize))

	require.Equalf(
		t,
		vector.Prettify(laidout[:chunkSize*params.NumLimbs()]),
		vector.Prettify(chunk0[:chunkSize*params.NumLimbs()]),
		"chunk 0 does not match with laid out key",
	)

	require.Equalf(
		t,
		vector.Prettify(laidout[:chunkSize*params.NumLimbs()]),
		vector.Prettify(chunk1[chunkSize*params.NumLimbs():]),
		"chunk 1 does not match with the laid out key",
	)

	// This should compute the sum of the two
	sumMergedKey := splitAndMergedKey(&key, splitIndices, 2*chunkSize, field.One())
	sumLaidOutKey := sumMergedKey.LaidOutKey()

	require.Equalf(t,
		vector.Prettify(sumLaidOutKey[:chunkSize*params.NumLimbs()]),
		vector.Prettify(chunk0[:chunkSize*params.NumLimbs()]),
		"chunk 0 does not match with merged key",
	)

	require.Equalf(t,
		vector.Prettify(sumLaidOutKey[chunkSize*params.NumLimbs():]),
		vector.Prettify(chunk1[chunkSize*params.NumLimbs():]),
		"chunk 1 does not match with merged key",
	)
}
