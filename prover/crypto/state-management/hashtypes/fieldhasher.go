package hashtypes

import (
	"github.com/consensys/gnark-crypto/hash"

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
