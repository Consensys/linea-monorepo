package poseidon2_koalabear

import (
	"fmt"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const BlockSize = 8

// TODO @thomas fixme arbitrary max size of the buffer
const maxSizeBuf = 1024

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {
	hash.StateStorer
	maxValue types.Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state field.Octuplet

	// data to hash
	buffer []field.Element
}

// Constructor for Poseidon2MDHasher
func NewMDHasher() *MDHasher {
	var maxVal field.Octuplet
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1")
	}
	return &MDHasher{
		StateStorer: gnarkposeidon2.NewMerkleDamgardHasher(),
		maxValue:    types.HashToBytes32(maxVal),
		state:       field.Octuplet{},
		buffer:      make([]field.Element, 0),
	}
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.buffer = d.buffer[:0]
	d.state = field.Octuplet{}
}

// SetStateOctuplet modifies the state (??)
func (d *MDHasher) SetStateOctuplet(state field.Octuplet) {
	for i := 0; i < 8; i++ {
		d.state[i].Set(&state[i])
	}
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts []field.Element) {
	quo := (len(d.buffer) + len(elmts)) / maxSizeBuf
	rem := (len(d.buffer) + len(elmts)) % maxSizeBuf
	off := len(d.buffer)
	for i := 0; i < quo; i++ {
		d.buffer = append(d.buffer, elmts[:maxSizeBuf-off]...)
		_ = d.SumElement()
		d.buffer = d.buffer[:0] // flush the buffer once maxSizeBuf is reached
		off = len(d.buffer)
	}
	d.buffer = append(d.buffer, elmts[:rem-off]...)
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
	d.WriteElements(elems)
	return len(p), nil
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
func GnarkBlockCompressionMekle(api frontend.API, oldState, block [BlockSize]zk.WrappedVariable) (newState [BlockSize]zk.WrappedVariable) {
	panic("unimplemented")
}
