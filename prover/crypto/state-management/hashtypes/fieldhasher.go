package hashtypes

import (
	"github.com/consensys/gnark-crypto/hash"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

// Poseidon2Hasher implements a Poseidon2-based hasher that works with both field elements and bytes
type Poseidon2Hasher struct {
	hash.StateStorer
	maxValue Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state  field.Octuplet
	buffer []field.Element // data to hash
}

// WriteElement adds a field element to the running hash.
func (d Poseidon2Hasher) WriteElement(e field.Element) {
	d.buffer = append(d.buffer, e)
}

// WriteElements adds a slice of field elements to the running hash.
func (d Poseidon2Hasher) WriteElements(elems []field.Element) {
	d.buffer = append(d.buffer, elems...)
}

func (p Poseidon2Hasher) SumElement() field.Octuplet {
	p.state = poseidon2.Poseidon2Sponge(p.buffer)
	p.buffer = p.buffer[:0]

	return p.state
}

func (p Poseidon2Hasher) SumElements(xs []field.Element) field.Octuplet {
	p.state = poseidon2.Poseidon2Sponge(xs)

	return p.state
}

func (p Poseidon2Hasher) MaxBytes32() Bytes32 {
	return p.maxValue
}

// ///// Constructor for Poseidon2Hasher /////
func Poseidon2() Poseidon2Hasher {
	var maxVal field.Octuplet
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1")
	}
	return Poseidon2Hasher{
		StateStorer: gnarkposeidon2.NewMerkleDamgardHasher(),
		maxValue:    HashToBytes32(maxVal),
		state:       field.Octuplet{},
		buffer:      make([]field.Element, 0),
	}
}
