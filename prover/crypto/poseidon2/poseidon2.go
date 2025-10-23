package poseidon2

import (
	"fmt"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

const blockSize = 8

// TODO @thomas fixme arbitrary max size of the buffer
const maxSizeBuf = 1024

// Poseidon2FieldHasherDigest implements a Poseidon2-based hasher that works with both field elements and bytes
type Poseidon2FieldHasherDigest struct {
	hash.StateStorer
	maxValue Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state  field.Octuplet
	buffer []field.Element // data to hash
}

// Reset clears the buffer, and reset state to iv
func (d *Poseidon2FieldHasherDigest) Reset() {
	d.buffer = d.buffer[:0]
	d.state = field.Octuplet{}
}

// WriteElements adds a slice of field elements to the running hash.
func (d *Poseidon2FieldHasherDigest) WriteElements(elmts []field.Element) {
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
func (d *Poseidon2FieldHasherDigest) Write(p []byte) (int, error) {

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

func (d *Poseidon2FieldHasherDigest) SumElement() field.Octuplet {
	for len(d.buffer) != 0 {
		var buf [blockSize]field.Element
		// in this case we left pad by zeroes
		if len(d.buffer) < blockSize {
			copy(buf[blockSize-len(d.buffer):], d.buffer)
			d.buffer = d.buffer[:0]
		} else {
			copy(buf[:], d.buffer)
			d.buffer = d.buffer[blockSize:]
		}
		d.state = vortex.CompressPoseidon2(d.state, buf)
	}
	return d.state
}

// Sum computes the poseidon2 hash of msg
func (d *Poseidon2FieldHasherDigest) Sum(msg []byte) []byte {
	d.Write(msg)
	h := d.SumElement()
	bytes := HashToBytes32(h)
	return bytes[:]
}
func (d Poseidon2FieldHasherDigest) MaxBytes32() Bytes32 {
	return d.maxValue
}

// ///// Constructor for Poseidon2Hasher /////
func Poseidon2() *Poseidon2FieldHasherDigest {
	var maxVal field.Octuplet
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1")
	}
	return &Poseidon2FieldHasherDigest{
		StateStorer: gnarkposeidon2.NewMerkleDamgardHasher(),
		maxValue:    HashToBytes32(maxVal),
		state:       field.Octuplet{},
		buffer:      make([]field.Element, 0),
	}
}

// Poseidon2Sponge returns a Poseidon2 hash of an array of field elements
func Poseidon2Sponge(x []field.Element) (newState field.Octuplet) {
	var state, xBlock field.Octuplet
	for len(x) != 0 {
		if len(x) < blockSize {
			padded := make([]field.Element, blockSize)
			copy(padded[blockSize-len(x):], x) // left-padding
			x = padded
		}

		copy(xBlock[:], x[:])
		state = vortex.CompressPoseidon2(state, xBlock)
		x = x[blockSize:]
	}
	return state
}

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompressionMekle(api frontend.API, oldState, block [blockSize]frontend.Variable) (newState [blockSize]frontend.Variable) {
	panic("unimplemented")
}
