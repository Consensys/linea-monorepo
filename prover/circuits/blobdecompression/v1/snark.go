package v1

import (
	"errors"
	"math/big"
	"math/bits"
	"slices"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	"github.com/consensys/gnark/std/compress/lzss"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/internal"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/internal/plonk"
	public_input "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/public-input"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1"
)

const (
	checkSumSize   = 32
	MaxNbBatches   = 100 // TODO ensure this is a reasonable maximum
	maxBlobNbBytes = 128 * 1024 * blob.PackingSizeU256 / 256
)

func combine(api frontend.API, bytes []frontend.Variable, perNewWord int) []frontend.Variable {
	res := make([]frontend.Variable, len(bytes)/perNewWord)
	for i := range res {
		res[i] = readNum(api, bytes[i*perNewWord:(i+1)*perNewWord])
	}
	return res
}

func readNum(api frontend.API, bytes []frontend.Variable) frontend.Variable {
	return compress.ReadNum(api, bytes, big.NewInt(256))
}

// parseHeader takes in a blob and returns the header length, the number of batches, and the length of each batch
// it assumes that the blob is already range-checked
// it ignores the dict checksum
// past nbBatches, the lengths are considered zero
// all lengths l are guaranteed to be within 0 ≤ l - 31 ≤ nextPowerOfTwo(maxPayloadBytes - 31)
func parseHeader(api frontend.API, blobBytes []frontend.Variable, blobLen frontend.Variable) (headerLen frontend.Variable, dictHash frontend.Variable, nbBatches frontend.Variable, bytesPerBatch []frontend.Variable, err error) {
	if len(blobBytes) < 2+checkSumSize+blob.NbElemsEncodingBytes { // version + checksum + nbBatches
		return 0, 0, 0, nil, errors.New("blob too short - no room for header")
	}

	api.AssertIsEqual(blobBytes[0], 255)
	api.AssertIsEqual(blobBytes[1], 255)
	blobBytes = blobBytes[2:]

	dictHash = compress.ReadNum(api, blobBytes[:checkSumSize], big.NewInt(256))
	blobBytes = blobBytes[checkSumSize:]

	// header structure: checksum, *nbBatches*, [batch lengths]
	nbBatches = compress.ReadNum(api, blobBytes[:blob.NbElemsEncodingBytes], big.NewInt(256))

	// no longer need the checksum or nbBatches encodings
	blobBytes = blobBytes[blob.NbElemsEncodingBytes:]

	// read MaxNbBatches 24-bit numbers
	blobWords := combine(api, blobBytes[:min(MaxNbBatches*blob.ByteLenEncodingBytes, len(blobBytes))], blob.ByteLenEncodingBytes)

	// header length is the length of the checksum and nbBatches, plus the length of the batch lengths
	headerLen = api.Add(2+checkSumSize+blob.NbElemsEncodingBytes, api.Mul(blob.ByteLenEncodingBytes, nbBatches)) // length in words
	bytesPerBatch = internal.Truncate(api, blobWords[:MaxNbBatches], nbBatches)                                  // zero out the "length" of the batches that don't exist

	// range checks for the batch lengths
	rc := rangecheck.New(api)
	const maxLMinus31 = blob.MaxUncompressedBytes - 31
	maxLMinus31Bits := bits.Len(uint(maxLMinus31))
	iterateInRange(api, nbBatches, MaxNbBatches, func(i int, inRange frontend.Variable) { // TODO-perf decide whether or not to merge this "loop" with the truncation above. PROBABLY NOT WORTH IT: currently this entire function is not even showing up in the profile graph
		rc.Check(api.MulAcc(api.Mul(-31, inRange), inRange, bytesPerBatch[i]), maxLMinus31Bits) // check for inRange * (bytesPerBatch[i] - 31). i.e. don't check past nbBatches
	})

	api.AssertIsLessOrEqual(headerLen, blobLen)
	api.AssertIsLessOrEqual(blobLen, len(blobBytes)) // redundant (considering how it's used in the zkevm)

	return
}

// it is the caller's responsibility to incorporate an indicator of the batch lengths into the checksum
// all batches must be at least 31 bytes long
func ChecksumBatches(api frontend.API, nbBatches frontend.Variable, blobPayload []frontend.Variable, batchEnds []frontend.Variable) (checksums [MaxNbBatches]frontend.Variable, err error) {

	qV, rV, err := divBy31(api, batchEnds[:], bits.Len(uint(len(blobPayload))+31)) // side effect: batchEnds are range checked to be within a reasonable factor of the maximum blob payload length
	if err != nil {
		return
	}
	q, r := internal.SliceToTable(api, qV), internal.SliceToTable(api, rV)

	// building a function that compares 0 ≤ a,b < 31
	// the function is built using a lookup table
	// since 0 ≤ a,b < 31, we always get -30 ≤ a - b ≤ 30 ⇒ 0 ≤ a - b + 30 ≤ 60
	// now we further have a ≥ b ⇔ a - b + 30 ≥ 30
	rComparisonTable := logderivlookup.New(api)
	for i := 0; i < 61; i++ {
		if i >= 30 {
			rComparisonTable.Insert(1)
		} else {
			rComparisonTable.Insert(0)
		}
	}
	rGte := func(a, b frontend.Variable) frontend.Variable { // compares 0 ≤ a,b < 31
		diff := api.Add(a, 30, api.Neg(b)) // b ≤ a ⇔ a - b ≥ 0 ⇔ a - b + 30 ≥ 30
		return rComparisonTable.Lookup(diff)[0]
	}

	nbHashes := 1 + (len(blobPayload)+30)/31 // one extra iteration of the main loop to simplify eof handling

	// let the payload be p₀ p₁ ... pₙ₋₁
	// then packedShifted[i] = (pᵢ pᵢ₊₁ ... pᵢ₊₃₀)₃₁      (with zero padding, if necessary)
	// i.e. a full word to incorporate into the checksum, starting at the i-th byte
	packedShifted := logderivlookup.New(api)
	nr := compress.NewNumReader(api, blobPayload, 31*8, 8)
	for i := 0; i < 31*nbHashes; i++ {
		packedShifted.Insert(nr.Next())
	}

	hsh, err := mimc.NewMiMC(api)
	if err != nil {
		return
	}

	// create a table of claimed sums and prove their correctness as we go through the payload
	{
		hintIn := make([]frontend.Variable, 1, 1+MaxNbBatches+len(blobPayload))
		hintIn[0] = nbBatches
		hintIn = append(hintIn, batchEnds[:]...)
		hintIn = append(hintIn, blobPayload...)
		var res []frontend.Variable
		if res, err = api.Compiler().NewHint(checksumBatchesPackedHint, MaxNbBatches, hintIn...); err != nil {
			return
		}
		copy(checksums[:], res)
	}

	batchSums := internal.SliceToTable(api, checksums[:])

	batchSum := packedShifted.Lookup(0)[0] // normally this should be taken care of by the api.Select(currAlreadyOver,... line. But the very first batch doesn't get this treatment because we know it starts at 0
	batchI := frontend.Variable(0)         // index of the current batch
	// each 31 byte block partially belongs to one or two batches (guaranteed by rejecting batches smaller than 31 bytes)
	currStartR := frontend.Variable(0) // the remainder by 31 of where the current batch starts

	for i := 0; i < nbHashes; i++ {

		currEndQ, currEndR := q.Lookup(batchI)[0], r.Lookup(batchI)[0]
		currEnds := api.IsZero(api.Sub(currEndQ, i)) // does the current batch end somewhere in the current 31-byte word?

		currLengthAligned := rGte(currStartR, currEndR) // do we need to incorporate the current word past the currStartR'th byte into the current batch?
		// if currEndR \leq currStartR we don't.

		if i != 0 {
			hsh.Reset()
			hsh.Write(batchSum, packedShifted.Lookup(api.Add(currStartR, 31*i))[0])
			// if we've already incorporated the parts of this word belonging to the current batch, disregard the hash we just computed
			batchSum = api.Select(api.Mul(currLengthAligned, currEnds), batchSum, hsh.Sum())
			assertEqualIf(api, currEnds, batchSum, batchSums.Lookup(batchI)[0])                         // if we're done with the current checksum, check that the claimed one from the table is equal to it
			batchSum = api.Select(currEnds, packedShifted.Lookup(api.Add(currEndR, 31*i))[0], batchSum) // if the next one starts, the sum is the 31 byte "prefix" of the next batch
		}

		currStartR = api.Select(currEnds, currEndR, currStartR) // if the next one starts, update the current start R
		batchI = api.Add(batchI, currEnds)                      // advance the batch index iff the current batch ends
	}
	return
}

func toBytes(ins []*big.Int) []byte {
	res := make([]byte, len(ins))
	for i := range ins {
		res[i] = byte(ins[i].Uint64())
	}
	return res
}

// ins: nbBatches, [end byte positions], payload...
// the result for each batch is <data (31 bytes)> ... <data (31 bytes)>
// for soundness some kind of length indicator must later be incorporated.
func checksumBatchesPackedHint(_ *big.Int, ins []*big.Int, outs []*big.Int) error {
	if len(outs) != MaxNbBatches {
		return errors.New("expected exactly MaxNbBatches outputs")
	}

	nbBatches := int(ins[0].Uint64())
	ends := toInts(ins[1 : 1+MaxNbBatches])
	in := append(toBytes(ins[1+MaxNbBatches:]), make([]byte, 31)...) // pad with 31 bytes to avoid out of range panic

	hsh := hash.MIMC_BLS12_377.New()

	batchStart := 0
	for i := range outs[:nbBatches] {
		res := in[batchStart : batchStart+31]
		nbWords := (ends[i] - batchStart + 30) / 31 // take as few 31-byte words as possible to cover the batch
		batchSumEnd := nbWords*31 + batchStart
		for j := batchStart + 31; j < batchSumEnd; j += 31 {
			hsh.Reset()
			hsh.Write(res)
			hsh.Write(in[j : j+31])
			res = hsh.Sum(nil)
		}
		outs[i].SetBytes(res)

		batchStart = ends[i]
	}

	return nil
}

func toInts(ints []*big.Int) []int {
	res := make([]int, len(ints))
	for i := range ints {
		res[i] = int(ints[i].Uint64())
	}
	return res
}

func registerHints() {
	lzss.RegisterHints()
	solver.RegisterHint(checksumBatchesPackedHint, divBy31Hint)
}

// side effect: ensures 0 ≤ v[i] < 2ᵇⁱᵗˢ⁺² for all i
func divBy31(api frontend.API, v []frontend.Variable, bits int) (q, r []frontend.Variable, err error) {
	qNbBits := bits - 4

	if hintOut, err := api.Compiler().NewHint(divBy31Hint, 2*len(v), v...); err != nil {
		return nil, nil, err
	} else {
		q, r = hintOut[:len(v)], hintOut[len(v):]
	}

	rChecker := rangecheck.New(api)

	for i := range v { // TODO See if lookups or api.AssertIsLte would be more efficient
		rChecker.Check(r[i], 5)
		api.AssertIsDifferent(r[i], 31)
		rChecker.Check(q[i], qNbBits)
		api.AssertIsEqual(v[i], api.Add(api.Mul(q[i], 31), r[i])) // 31 × q < 2ᵇⁱᵗˢ⁻⁴ 2⁵ ⇒ v < 2ᵇⁱᵗˢ⁺¹ + 31 < 2ᵇⁱᵗˢ⁺²
	}
	return q, r, nil
}

// outs: [quotients], [remainders]
func divBy31Hint(_ *big.Int, ins []*big.Int, outs []*big.Int) error {
	if len(outs) != 2*len(ins) {
		return errors.New("expected output layout: [quotients][remainders]")
	}

	q := outs[:len(ins)]
	r := outs[len(ins):]
	for i := range ins {
		v := ins[i].Uint64()
		q[i].SetUint64(v / 31)
		r[i].SetUint64(v % 31)
	}

	return nil
}

// assert cond ⇒ (a == b)
func assertEqualIf(api frontend.API, cond, a, b frontend.Variable) {
	api.AssertIsEqual(0, api.Mul(cond, api.Sub(a, b)))
}

// iterateInRange runs f(i, inRange) for 0 ≤ i < staticRange where inRange is 1 if i < dynamicRange and 0 otherwise
func iterateInRange(api frontend.API, dynamicRange frontend.Variable, staticRange int, f func(i int, inRange frontend.Variable)) {
	inRange := frontend.Variable(1)
	for i := 0; i < staticRange; i++ {
		inRange = api.Sub(inRange, api.IsZero(api.Sub(i, dynamicRange)))
		f(i, inRange)
	}
}

// bitStreamToByteStream converts a slice of bits into a slice of bytes, taking the last non-zero byte as signifying the end of the data
func bitStreamToByteStream(api frontend.API, bits []frontend.Variable) (bytes []frontend.Variable, nbBytes frontend.Variable) {
	bytes = make([]frontend.Variable, 0, len(bits)/8)
	for i := 0; i < len(bits); i += 8 {
		var curr [8]frontend.Variable
		nbBits := min(8, len(bits)-i)
		copy(curr[:], bits[i:i+nbBits])
		for j := nbBits; j < 8; j++ {
			curr[j] = 0
		}
		slices.Reverse(curr[:])
		bytes = append(bytes, api.FromBinary(curr[:]...))
	}

	found := frontend.Variable(0)
	nbBytes = frontend.Variable(0)
	for i := len(bytes) - 1; i >= 0; i-- {

		z := api.IsZero(bytes[i])

		lastNonZero := plonk.EvaluateExpression(api, z, found, -1, -1, 1, 1)   // nz - found
		nbBytes = api.Add(nbBytes, api.Mul(lastNonZero, frontend.Variable(i))) // the last nonzero byte itself is useless

		//api.AssertIsEqual(api.Mul(api.Sub(bytesPerElem-i%bytesPerElem, unpacked[i]), lastNonZero), 0) // sanity check, technically unnecessary TODO @Tabaie make sure it's one constraint only or better yet, remove

		found = plonk.EvaluateExpression(api, z, found, -1, 0, 1, 1) // found ? 1 : nz = nz + found (1 - nz) = 1 - z + found z
	}

	return
}

// ProcessBlob takes in a blob, an evaluation challenge, and a decompression dictionary. It returns a hash of the blob data along with its "evaluation" at the challenge point and a hash of all the batches in the blob payload
func ProcessBlob(api frontend.API, blobBytes []frontend.Variable, evaluationChallenge [2]frontend.Variable, eip4844Enabled frontend.Variable, dict []frontend.Variable) (blobSum frontend.Variable, evaluation [2]frontend.Variable, batchesSum frontend.Variable, err error) {

	blobBits := internal.PackedBytesToBits(api, blobBytes, blob.PackingSizeU256)

	hsh, err := mimc.NewMiMC(api)
	if err != nil {
		return
	}

	blobPacked377 := internal.PackNative(api, blobBits, 1) // repack into bls12-377 elements to compute a checksum
	hsh.Write(blobPacked377...)
	blobSum = hsh.Sum()

	// EIP-4844 stuff
	if evaluation, err = public_input.VerifyBlobConsistency(api, blobBits, evaluationChallenge, eip4844Enabled); err != nil {
		return
	}

	// repack into bytes TODO possible optimization: pass bits directly to decompressor
	// unpack into bytes
	blobUnpackedBytes, blobUnpackedNbBytes := bitStreamToByteStream(api, blobBits)

	// get header length, number of batches, and length of each batch
	headerLen, dictChecksum, nbBatches, bytesPerBatch, err := parseHeader(api, blobUnpackedBytes[:maxBlobNbBytes], blobUnpackedNbBytes)
	if err != nil {
		return
	}

	// check if the decompression dictionary checksum matches
	if err = CheckDictChecksum(api, dictChecksum, dict); err != nil {
		return
	}

	// decompress the batches
	payload := make([]frontend.Variable, blob.MaxUncompressedBytes)
	payloadLen, err := lzss.Decompress(
		api,
		compress.ShiftLeft(api, blobUnpackedBytes[:maxBlobNbBytes], headerLen), // TODO Signal to the decompressor that the input is zero padded; to reduce constraint numbers
		api.Sub(blobUnpackedNbBytes, headerLen),
		payload,
		dict,
	)
	if err != nil {
		return
	}
	api.AssertIsDifferent(payloadLen, -1) // decompression should not fail

	// convert batch length array to batch ends array
	batchEnds := make([]frontend.Variable, MaxNbBatches)
	batchEnds[0] = bytesPerBatch[0]
	for i := 1; i < len(bytesPerBatch); i++ {
		batchEnds[i] = api.Add(batchEnds[i-1], bytesPerBatch[i])
	}

	// compute checksum for each batch
	batchSums, err := ChecksumBatches(api, nbBatches, payload, batchEnds)
	if err != nil {
		return
	}

	// hash the checksums together, along with the number of batches
	hsh.Reset()
	hsh.Write(nbBatches)
	hsh.Write(batchSums[:]...)
	batchesSum = hsh.Sum()

	return
}

func CheckDictChecksum(api frontend.API, checksum frontend.Variable, dict []frontend.Variable) error {
	// decompose into bits TODO possible optimization: decompose into crumbs instead of bits
	dictBits := internal.PackedBytesToBits(api, dict, 8) // basically just turn bytes into bits
	dictBits = append(dictBits, 1, 1, 1, 1, 1, 1, 1, 1)  // add the 0xff end-of-stream marker
	dictPacked := internal.PackNative(api, dictBits, 1)
	return compress.AssertChecksumEquals(api, dictPacked, checksum)
}
