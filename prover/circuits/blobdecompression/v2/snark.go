package v2

import (
	"math/big"
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	snarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
)

// CheckBatchesSums checks batch checksums consisting of H(batchLen, contentSum) where contentSum = Blocks[0] if len(Blocks) == 1 and H(Blocks...) otherwise. Blocks are consecutive 31-byte chunks of the data in the batch, with the last one right-padded with zeros if necessary.
// All batches must be at least 31 bytes long. The function performs this range check.
// It is also checked that the batches are all within the MAXIMUM range of the blob. CheckBatchesSums does not have access to the actual blob size, so it remains the caller's responsibility to check that the batches are within the confines of the ACTUAL blob size.
// The expected checksums are not checked beyond nbBatches
func CheckBatchesSums(api frontend.API, hasher snarkHash.FieldHasher, nbBatches frontend.Variable, blobPayload []frontend.Variable, batchLengths []frontend.Variable, expectedChecksums []frontend.Variable) error {

	batchEnds := internal.PartialSums(api, batchLengths)

	if len(batchEnds) > len(expectedChecksums) {
		return errors.New("more batches than checksums")
	}

	if len(blobPayload) < 31 { // edge case
		api.AssertIsEqual(nbBatches, 0)
		for i := range batchEnds {
			api.AssertIsEqual(batchEnds[i], 0)
		}
		return nil
	}

	api.AssertIsDifferent(nbBatches, 0)

	cappedRamp := logderivlookup.New(api)
	const cappedRampNegMin = 31
	for i := -cappedRampNegMin; i < len(blobPayload)+93; i++ {
		cappedRamp.Insert(min(31, max(i, 0)))
	}
	// min0Max31(x) = min(31, max(x, 0)), i.e. it returns 0 if x < 0, 31 if x > 31, and x otherwise
	min0Max31 := func(x frontend.Variable) frontend.Variable {
		return cappedRamp.Lookup(api.Add(x, cappedRampNegMin))[0]
	}

	nbHashes := 1 + len(blobPayload)/31 // one extra iteration of the main loop to simplify eof handling
	// every batch is only sealed when the next is about to begin. So we need to start creating the dummy batch when the loop ends.
	// to ensure that happens in the case of a full blob, we will need a dummy iteration as well

	// a practically infinite dummy batch at the end to prevent index overflows
	// and to make sure the dummy batch still doesn't "end" on the last iteration, we must give it and extra 31 bytes on top of that
	// nbHashes*31+1 is JUST beyond what the loop will reach so that the dummy batch is never sealed.
	dummyBatchEnd := nbHashes*31 + 1

	batchesRange := internal.NewRange(api, nbBatches, len(batchEnds)) // this also range-checks nbBatches
	for i := range batchEnds {                                        // check that the size of every batch is at least 31
		// in particular this ensures that for ⌊ end[i] / 31 ⌋ != ⌊ end[i+1] / 31 ⌋ for all applicable i

		internal.AssertEqualIf(api, batchesRange.IsFirstBeyond[i], 31, // "Select" to avoid going out of range
			min0Max31(api.Select(batchesRange.InRange[i], batchLengths[i], 31)))

		batchEnds[i] = api.Select(batchesRange.IsFirstBeyond[i], dummyBatchEnd, batchEnds[i])
	}

	// side effect: batchEnds are range checked to be within a reasonable factor of the maximum blob payload length; useless because we will have to perform a stronger check in the end
	endQsV, endRsV, err := v1.DivBy31(api, batchEnds, bits.Len(uint(len(blobPayload))+31))
	if err != nil {
		return err
	}
	endQs, endRs := internal.SliceToTable(api, endQsV), internal.SliceToTable(api, endRsV)

	// another practically infinite dummy batch in case nbBatches == len(batchEnds)
	endQs.Insert(dummyBatchEnd / 31)
	endRs.Insert(dummyBatchEnd % 31)
	// we need an extra dummy input element past the end of the dummy batch, because the loop is always considering
	// sealing the dummy batch and starting yet another one after it, though it never actually happens.
	// still, the circuit computes the 31-byte prefix of the next batch.
	inputExt := make([]frontend.Variable, dummyBatchEnd+31)
	for n := copy(inputExt, blobPayload); n < len(inputExt); n++ {
		inputExt[n] = 0
	}
	inputT := internal.SliceToTable(api, inputExt)
	// inputAt returns a packed, zero-padded substring of length min(l,31) starting at i
	inputAt := func(i, l frontend.Variable) frontend.Variable {
		out := make([]frontend.Variable, 31)
		r := internal.NewRange(api, l, len(out))
		for j := range out {
			out[j] = api.Mul(r.InRange[j], inputT.Lookup(api.Add(i, j))[0]) // Perf note this enables substrings of length 0 which we never use
		}
		return compress.ReadNum(api, out, big.NewInt(256))
	}

	// let the payload be p₀ p₁ ... pₙ₋₁
	// then inputAt_31B[i] = (pᵢ pᵢ₊₁ ... pᵢ₊₃₀)₃₁      (with zero padding, if necessary)
	// i.e. a full word to incorporate into the checksum, starting at the i-th byte
	inputAt31B := logderivlookup.New(api)
	nr := compress.NewNumReader(api, inputExt, 31*8, 8)
	for i := 0; i < 31*nbHashes+2; i++ { // TODO justify the +2
		inputAt31B.Insert(nr.Next())
	}

	_hsh := func(a, b frontend.Variable) frontend.Variable {
		hasher.Reset()
		hasher.Write(a, b)
		res := hasher.Sum()
		return res
	}

	var (
		partialSumsT *logderivlookup.Table
		partialSums  []frontend.Variable
	)
	// create a table of claimed sums and prove their correctness as we go through the payload
	{
		hintIn := make([]frontend.Variable, 1, 1+len(batchEnds)+len(blobPayload))
		hintIn[0] = nbBatches
		hintIn = append(hintIn, batchEnds[:]...)
		hintIn = append(hintIn, blobPayload...)
		if partialSums, err = api.Compiler().NewHint(partialChecksumBatchesPackedHint, len(batchEnds), hintIn...); err != nil {
			return err
		}
	}
	partialSumsT = internal.SliceToTable(api, partialSums)
	partialSumsT.Insert(0) // dummy in case of maximum nbBatches

	batchSum := inputAt(0, 31)     // normally this should be taken care of by the api.Select(currAlreadyOver,... line. But the very first batch doesn't get this treatment because we know it starts at 0
	batchI := frontend.Variable(0) // index of the current batch
	// each 31 byte block partially belongs to one or two batches (guaranteed by rejecting batches smaller than 31 bytes)
	startR := frontend.Variable(0) // the remainder by 31 of where the current batch starts

	// each iteration is able to process at most one new batch. This dictates that end[i] % 31 != end[i+1] % 31 for any applicable i
	for i := 0; i < nbHashes; i++ {

		endQ, endR := endQs.Lookup(batchI)[0], endRs.Lookup(batchI)[0]
		end := api.Add(api.Mul(31, endQ), endR)

		currNbBytesRemaining := api.Sub(end, 31*i, startR) // 31i + startR is the location of the "head"
		hashLen := min0Max31(currNbBytesRemaining)
		startNext := api.IsZero(api.Sub(endQ, i))
		noHash := api.IsZero(hashLen) // or equivalently, isZero(currNbBytesRemaining)

		if i != 0 {
			batchSum = api.Select(
				noHash,
				batchSum,
				_hsh(batchSum, inputAt(api.Add(31*i, startR), hashLen)),
			)

			internal.AssertEqualIf(api, startNext, batchSum, partialSumsT.Lookup(batchI)[0]) // if we're done with the current checksum, check that the claimed one from the table is equal to it
			// THIS STEP REQUIRES THAT NO BATCH SHOULD BE SMALLER THAN 31 BYTES
			//
			// This is always the case in practice since a batch contains
			// at least one block which in turn contains a 32 byte root hash.
			batchSum = api.Select(startNext, inputAt31B.Lookup(end)[0], batchSum) // if the next one starts, the sum is the 31 byte "prefix" of the next batch; assumes any batch is at least 31 bytes long

		}

		startR = api.Select(startNext, endR, startR) // if the next one starts, update the current start R
		batchI = api.Add(batchI, startNext)
	}

	api.AssertIsEqual(batchI, nbBatches) // check that we're actually done

	// hash along the lengths and compare with expected values
	for i := range batchEnds {
		hasher.Reset()
		hasher.Write(batchLengths[i], partialSums[i])
		batchesRange.AssertEqualI(i, expectedChecksums[i], hasher.Sum())
	}

	return nil
}

func partialChecksumBatchesPackedHint(_ *big.Int, ins, outs []*big.Int) error {

}
