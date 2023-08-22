package hash_proof

import (
	"bytes"

	keccakHash "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/hashing/hash"
	keccak "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

type Bytes []byte
type MultiHash struct {
	// the input and output of all the hashes (without padding)
	InputHash  []Bytes
	OutputHash []Bytes
}

// it extracts all the provoked keccakf (submitted via different hashes),
// then assigns the columns for keccakFModule
func (t *MultiHash) Prover(run *wizard.ProverRuntime, keccakFModule keccakf.KeccakFModule) {
	h := keccakHash.NewLegacyKeccak256()
	var inputBatchPerm, outputBatchPerm [][5][5]uint64
	for i, u := range t.InputHash {
		h.Reset()

		if u == nil {
			utils.Panic("the message is empty for the %v hash", i)
		}
		// write extracts and assigns the input/output of permutations to keccakf module
		h.Write(u)
		out := h.Sum(nil)
		if !bytes.Equal(t.OutputHash[i], out) {
			panic("hash outputs is not consistent")
		}
		// append all the permutation provoked by the current hash
		s, _ := h.(*keccakHash.State)
		inputBatchPerm = append(inputBatchPerm, s.InputPermUint64...)
		outputBatchPerm = append(outputBatchPerm, s.OutputPermUint64...)

	}
	ctx := keccakFModule
	numPerm := keccakFModule.NP
	m := NumberOfKeccakf(*t)

	// check that the number of provoked permutations is correct and smaller that numPerm in the keccakf module
	if m != len(inputBatchPerm) {
		utils.Panic("expected to provoke %v permutation but it provoked %v permutations  ", m, len(inputBatchPerm))
	}
	if len(inputBatchPerm) > numPerm {
		utils.Panic("numPerm supported by the keccakModule is %v while %v permutations are provoked ", numPerm, len(inputBatchPerm))
	}

	// if a small number of permutations is provoked, add fake permutations to the keccakf module
	if len(inputBatchPerm) < numPerm {
		inputBatchPerm, outputBatchPerm = AddFakePerm(inputBatchPerm, outputBatchPerm, numPerm)
	}

	//fill up public inputs for the keccakf module
	for l := range inputBatchPerm {
		ctx.InputPI[l] = keccakf.ConvertState(inputBatchPerm[l], keccakf.First)
		ctx.OutputPI[l] = keccakf.ConvertState(outputBatchPerm[l], keccakf.First)
	}
	// assigns all the witness for keccakf module
	ctx.ProverAssign(run)

}

// adds fake permutation in order to fill up the keccakf module
func AddFakePerm(inputOldBatch, outputOldBatch [][5][5]uint64, numPerm int) (in, out [][5][5]uint64) {
	n := numPerm - len(inputOldBatch)
	in = inputOldBatch
	out = outputOldBatch
	var zeroState [5][5]uint64
	resultZeroState := keccak.KeccakF1600Original(zeroState)
	for i := 0; i < n; i++ {
		in = append(in, zeroState)
		out = append(out, resultZeroState)
	}
	return in, out
}

// it extract the total number of keccakf in a multihash
func NumberOfKeccakf(h MultiHash) int {
	s := 0
	msg := h.InputHash
	for i := range msg {
		n := len(msg[i])
		m := n/keccakHash.Rate + 1
		s = s + m
	}
	return s
}
