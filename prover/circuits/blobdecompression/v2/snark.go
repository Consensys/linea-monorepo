package v2

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	snarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
)

// calculatePartialSums hint returns a function that consumes the input in the form
// nbBatches, batchLens[0], ..., batchLens[maxNbBatches-1],
// evaluationPoint[0], ..., evaluationPoint[maxNbBatches-1],
// blobPayloadByte[0], ...
// and outputs batchSums0[0], batchSums0[1], ..., batchSums0[maxNbBatches-1],
// batchSums1[0], batchSums1[1], ..., batchSums1[maxNbBatches-1].
func calculatePartialSumsHint(maxNbBatches int) solver.Hint {
	return func(_ *big.Int, ins, outs []*big.Int) error {
		if len(ins) < 1+2*maxNbBatches {
			return fmt.Errorf("input must be at least of length 1 + %d to accomodate metadata", maxNbBatches)
		}
		if !ins[0].IsUint64() || ins[0].Uint64() > uint64(maxNbBatches) {
			return fmt.Errorf("%s batches exceed the maximum %d allowed", ins[0], maxNbBatches)
		}
		if len(outs) != 2*maxNbBatches {
			return fmt.Errorf("expected %d outputs, got %d", 2*maxNbBatches, len(outs))
		}

		evaluationPoints := ins[1+maxNbBatches : 1+2*maxNbBatches]
		evaluationCalculators := make([]func([][]byte) []byte, maxNbBatches)
		for i := range evaluationPoints {
			evaluationCalculators[i] = evaluateBatch(evaluationPoints[i])
		}

		payload := ins[1+2*maxNbBatches:]
		var (
			batch       bytes.Buffer
			packedBatch [][]byte
		)
		for i := range ins[0].Uint64() {

			if !ins[1+i].IsUint64() {
				return fmt.Errorf("bacth #%d of length %s is too large", i, ins[1+i])
			}
			batchLen := ins[i+1].Uint64()
			if uint64(len(payload)) < batchLen {
				return fmt.Errorf("unexpected end of payload at batch #%d", i)
			}

			// prepare input to partial sum calculators
			batch.Reset()
			for j := range batchLen {
				if j%31 == 0 {
					batch.WriteByte(0)
				}
				if !payload[j].IsUint64() || payload[j].Uint64() > 255 {
					return fmt.Errorf("payload element is not a byte: %s", payload[j])
				}
				batch.WriteByte(byte(payload[j].Uint64()))
			}
			for batch.Len()%32 != 0 {
				batch.WriteByte(0)
			}
			packedBatch = packedBatch[:0]
			for j := range batch.Len() / 32 {
				packedBatch = append(packedBatch, batch.Bytes()[j*32:j*32+32])
			}
			payload = payload[batchLen:]

			// call partial sum calculators
			for j, f := range []func([][]byte) []byte{poseidon2PartialHash, evaluationCalculators[i]} {
				outs[uint64(j*maxNbBatches)+i].SetBytes(f(packedBatch))
			}
		}

		return nil
	}
}

func poseidon2PartialHash(batch [][]byte) []byte {
	var err error
	hsh := poseidon2.NewDefaultPermutation()
	res := batch[0]
	for _, b := range batch[1:] {
		if res, err = hsh.Compress(res, b); err != nil {
			panic(err) // if an error occurs, it is catastrophic
		}
	}
	return res
}

func evaluateBatch(evaluationPoint *big.Int) func([][]byte) []byte {
	return func(batch [][]byte) []byte {
		var x, evaluation, c fr.Element
		x.SetBigInt(evaluationPoint)
		evaluation.SetBytes(batch[0])
		for i := 1; i < len(batch); i++ {
			evaluation.Mul(&evaluation, &x)
			c.SetBytes(batch[i])
			evaluation.Add(&evaluation, &c)
		}
		res := evaluation.Bytes()
		return res[:]
	}
}

// CheckBatchesSums checks batch checksums consisting of H(batchLen, contentSum) where contentSum = Blocks[0] if len(Blocks) == 1 and H(Blocks...) otherwise. Blocks are consecutive 31-byte chunks of the data in the batch, with the last one right-padded with zeros if necessary.
// All batches must be at least 31 bytes long. The function performs this range check.
// It is also checked that the batches are all within the MAXIMUM range of the blob. CheckBatchesSums does not have access to the actual blob size, so it remains the caller's responsibility to check that the batches are within the confines of the ACTUAL blob size.
// The expected checksums are not checked beyond nbBatches
// CheckBatchesSums checks batch checksums consisting of H(batchLen, contentSum) where contentSum = Blocks[0] if len(Blocks) == 1 and H(Blocks...) otherwise. Blocks are consecutive 31-byte chunks of the data in the batch, with the last one right-padded with zeros if necessary.
// All batches must be at least 31 bytes long. The function performs this range check.
// It is also checked that the batches are all within the MAXIMUM range of the blob. CheckBatchesSums does not have access to the actual blob size, so it remains the caller's responsibility to check that the batches are within the confines of the ACTUAL blob size.
// The expected checksums are not checked beyond nbBatches
func CheckBatchesSums(api frontend.API, maxNbBatches int, compressor snarkHash.Compressor, nbBatches frontend.Variable, blobPayload, batchLengths []frontend.Variable, evaluationPoints []frontend.Variable, expectedSums [2][]frontend.Variable) error {
	hashSums := expectedSums[0]
	pointEvalSums := expectedSums[1]

	if len(hashSums) != len(batchLengths) {
		return fmt.Errorf("given checksums and batch lengths don't match in number %d≠%d", len(hashSums), len(batchLengths))
	}
	if len(pointEvalSums) != len(evaluationPoints) {
		return fmt.Errorf("given evaluations and evaluation points don't match in number %d≠%d", len(pointEvalSums), len(evaluationPoints))
	}
	if len(hashSums) != len(pointEvalSums) {
		return fmt.Errorf("given hashes and evaluations don't match in number %d≠%d", len(hashSums), len(pointEvalSums))
	}

	packedBatches, err := v1.PackBatches(api, nbBatches, blobPayload, batchLengths)
	if err != nil {
		return err
	}

	var partialSums []frontend.Variable

	// create a table of claimed sums and prove their correctness as we go through the payload
	{
		hintIn := make([]frontend.Variable, 1, 1+1*len(packedBatches.BatchEnds)+len(blobPayload))
		hintIn[0] = nbBatches
		hintIn = append(hintIn, packedBatches.BatchEnds...)
		hintIn = append(hintIn, evaluationPoints...)
		hintIn = append(hintIn, blobPayload...)
		if partialSums, err = api.Compiler().NewHint(calculatePartialSumsHint(maxNbBatches), len(packedBatches.BatchEnds), hintIn...); err != nil {
			return err
		}
	}
	const (
		HASH = 0
		EVAL = 1
	)
	var partialsTables [2]*logderivlookup.Table
	partialsTables[HASH] = internal.SliceToTable(api, partialSums[:maxNbBatches])
	partialsTables[HASH].Insert(0) // dummy in case of maximum nbBatches
	partialsTables[EVAL] = internal.SliceToTable(api, partialSums[maxNbBatches:])
	partialsTables[EVAL].Insert(0)
	evaluationPointsT := internal.SliceToTable(api, evaluationPoints)

	iterations := packedBatches.Iterations

	updateFunctions := []func(int, frontend.Variable, frontend.Variable) frontend.Variable{
		func(_ int, state, current frontend.Variable) frontend.Variable { return compressor.Compress(state, current) },
		func(i int, state, current frontend.Variable) frontend.Variable { return api.MulAcc(current, state, evaluationPointsT.Lookup(iterations[i].BatchI)[0]) },
	}

	// normally this should be taken care of by the api.Select(currAlreadyOver,... line. But the very first batch doesn't get this treatment because we know it starts at 0
	state := [2]frontend.Variable{iterations[0].Current, iterations[0].Current}

	for i := 1; i < len(iterations); i++ {
		for j, f := range updateFunctions {
			state[j] = f(i, state[j], iterations[i].Current)
			internal.AssertEqualIf(api, iterations[i].NextStarts, state[j], partialsTables[j].Lookup(iterations[i].BatchI)[0]) // if we're done with the current checksum, check that the claimed one from the table is equal to it
			// THIS STEP REQUIRES THAT NO BATCH SHOULD BE SMALLER THAN 31 BYTES
			//
			// This is always the case in practice since a batch contains
			// at least one block which in turn contains a 32 byte root hash.
			state[j] = api.Select(iterations[i].NextStarts, iterations[i].Next, state[j]) // if the next one starts, the checksum is the 31 byte "prefix" of the next batch; assumes any batch is at least 31 bytes long
		}
	}

	// hash along the lengths and compare with expected values
	for i := range batchLengths {
		for j, f := range updateFunctions {
			packedBatches.Range.AssertEqualI(i, expectedSums[j][i], f(i, partialSums[j*maxNbBatches+i], batchLengths[i]))
		}
	}

	return nil
}
