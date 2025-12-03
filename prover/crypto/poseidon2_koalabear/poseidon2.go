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

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {
	hash.StateStorer
	maxValue types.Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state field.Octuplet

	// data to hash
	buffer         [BlockSize]field.Element
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
	// d.buffer has BlockSize slots. Some may already be filled (indicated by d.bufferPosition).
	// We fill up d.buffer, and whenever it gets full, we compress it and reset it.
	// We repeat this until all elmts are consumed.
	// At the end, d.buffer may be partially filled.
	for _, e := range elmts {
		d.buffer[d.bufferPosition] = e
		d.bufferPosition++
		if d.bufferPosition == BlockSize {
			// buffer full, compress
			d.state = vortex.CompressPoseidon2(d.state, d.buffer)
			d.bufferPosition = 0
		}
	}
}

func (d *MDHasher) SumElement() field.Octuplet {
	if d.bufferPosition == 0 {
		return d.state
	}
	// pad the buffer and compress
	// we need to pad on the left
	var buf [BlockSize]field.Element
	copy(buf[BlockSize-d.bufferPosition:], d.buffer[:d.bufferPosition])
	d.state = vortex.CompressPoseidon2(d.state, buf)
	d.bufferPosition = 0
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
