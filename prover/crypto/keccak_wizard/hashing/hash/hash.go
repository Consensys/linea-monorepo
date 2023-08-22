package hash

import (
	"encoding/binary"
	"hash"

	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
)

// this implements keccak hash while it keeps track of provoked permutations, it does not support multihash or proof
func NewLegacyKeccak256() hash.Hash {
	var a [5][5]uint64
	return &State{aUint64: a}
}

const (
	//rate in byte
	Rate = 136
	//output length in byte
	OutputLen = 32
)

// hash state
type State struct {
	aUint64                           [5][5]uint64   // original form of state
	output                            []byte         // the output
	InputPermUint64, OutputPermUint64 [][5][5]uint64 // all  the input/output to the  keccakf provoked by the hash
}

// BlockSize returns the rate of sponge
func (d *State) BlockSize() int { return Rate }

// Size returns the output size of the hash function in bytes.
func (d *State) Size() int { return OutputLen }

// Reset clears the internal state by zeroing the sponge state and
func (d *State) Reset() {
	// set the permutation state to Zero
	//d.stateZero()
	var a [5][5]uint64
	d.aUint64 = a

}

// Write absorbs more data into the hash's state.
func (d *State) Write(t []byte) (written int, err error) {
	//sanity check, state should be zero at the beginning
	if !isStateZero(d.aUint64) {
		panic(" state should be zero at the beginning")
	}
	//pad it to get multiple of Rate
	t = pad(t)

	// cut the input to blocks
	blockBytes := messageBlocks(t)

	//sanity check
	if len(blockBytes) != len(t)/Rate {
		panic("cutting the message to blocks is not done correctly")
	}
	written = len(blockBytes)
	d.InputPermUint64 = make([][5][5]uint64, len(blockBytes))
	d.OutputPermUint64 = make([][5][5]uint64, len(blockBytes))
	for l := range blockBytes {
		//sanity check
		if len(blockBytes[l]) != Rate {
			panic("each block should be of size 'Rate'")

		}
		//absorb the block and update the state
		d.absorbBlock(blockBytes[l])

		out := keccakf.KeccakF1600Original(d.aUint64)
		//add the input/output of permutation to the list of provoked keccakf
		d.InputPermUint64[l] = d.aUint64
		d.OutputPermUint64[l] = out

		// update the state
		d.aUint64 = out

	}

	return
}

// Read squeezes "OutputLen=32" number of bytes from the sponge.
func (d *State) Read() (n int, err error) {
	// convert the output to slice of bytes
	var hashByte []byte
	for i := 0; i < OutputLen/8; i++ {
		s := d.aUint64[i][0]
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, s)
		hashByte = append(hashByte, b...)
	}
	n = len(hashByte)
	d.output = hashByte

	return
}

// Sum  squeezes out the desired number of output.
func (d *State) Sum(in []byte) []byte {
	d.Read()
	return append(in, d.output...)
}
