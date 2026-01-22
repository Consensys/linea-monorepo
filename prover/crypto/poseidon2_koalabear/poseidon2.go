package poseidon2_koalabear

import (
	"fmt"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const BlockSize = 8

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {
	hash.StateStorer
	maxValue types.Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state field.Octuplet

	// data to hash
	buffer         []field.Element
	bufferPosition int
}

// NewMDHasher creates a new MDHasher with the given options.
func NewMDHasher() *MDHasher {
	var maxVal field.Octuplet
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1")
	}
	h := &MDHasher{
		StateStorer: gnarkposeidon2.NewMerkleDamgardHasher(),
		maxValue:    types.HashToBytes32(maxVal),
		state:       field.Octuplet{},
	}

	return h
}

// Compress calls the compression function of Poseidon2 over state and block.
func Compress(state, block field.Octuplet) field.Octuplet {
	return vortex.CompressPoseidon2(state, block)
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.bufferPosition = 0
	d.state = field.Octuplet{}
}

// SetStateOctuplet modifies the state (??)
func (d *MDHasher) SetStateOctuplet(state field.Octuplet) {
	for i := 0; i < 8; i++ {
		d.state[i].Set(&state[i])
	}
}

func (d *MDHasher) GetStateOctuplet() field.Octuplet {
	return d.state
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts ...field.Element) {

	d.buffer = append(d.buffer, elmts[:]...)

}

func (d *MDHasher) SumElement() field.Octuplet {
	for len(d.buffer) != 0 {
		var buf [BlockSize]field.Element

		// in this case we left pad by zeroes
		if len(d.buffer) < BlockSize {
			copy(buf[BlockSize-len(d.buffer):], d.buffer)
			d.buffer = d.buffer[:0]
		} else {
			copy(buf[:], d.buffer)
			d.buffer = d.buffer[BlockSize:]
		}

		d.state = vortex.CompressPoseidon2(d.state, buf)
	}

	return d.state
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) Write(p []byte) (int, error) {

	elemByteSize := field.Bytes // 4 bytes = 1 field element

	if len(p)%elemByteSize != 0 {
		return 0, fmt.Errorf("input length is not a multiple of 4 byte size")
	}
	elems := make([]field.Element, 0, len(p)/elemByteSize)

	for start := 0; start < len(p); start += elemByteSize {
		chunk := p[start : start+elemByteSize]

		var elem field.Element
		elem.SetBytes(chunk)
		elems = append(elems, elem)

	}
	d.WriteElements(elems...)
	return len(p), nil
}
func (h *MDHasher) State() [BlockSize]field.Element {
	return h.state
}

// Sum computes the poseidon2 hash of msg
func (d *MDHasher) Sum(msg []byte) []byte {
	d.Write(msg)
	h := d.SumElement()
	bytes := types.HashToBytes32(h)
	return bytes[:]
}
func (d MDHasher) MaxBytes32() types.Bytes32 {
	return d.maxValue
}

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompressionMekle(api frontend.API, oldState, block [BlockSize]koalagnark.Element) (newState [BlockSize]koalagnark.Element) {
	panic("unimplemented")
}

// HashVec hashes a vector of field elements
func HashVec(v ...field.Element) (h field.Octuplet) {
	state := NewMDHasher()
	state.WriteElements(v...)
	return state.SumElement()
}
