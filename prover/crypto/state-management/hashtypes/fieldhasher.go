package hashtypes

import (
	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/hash"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// Poseidon2FieldHasher is an interface for a hash function that operates on both field elements and bytes
type Poseidon2FieldHasher interface {
	hash.StateStorer
	WriteElement(e field.Element)
	WriteElements(elems []field.Element)
	SumElements(elems []field.Element) field.Octuplet

	MaxBytes32() types.Bytes32 // Returns the maximum representable value as Bytes32
}

// Poseidon2FieldHasherDigest implements a Poseidon2-based hasher that works with field elements
type Poseidon2FieldHasherDigest struct {
	hash.StateStorer

	maxValue types.Bytes32

	// Sponge construction state
	h    field.Octuplet
	data []field.Element // data to hash
}

// Poseidon2 returns a Poseidon2FieldHasher (works with typed field elements and bytes)
func Poseidon2() Poseidon2FieldHasher {
	var maxVal field.Octuplet // This stores the maximal value for each element
	for i := range maxVal {
		maxVal[i] = field.NewFromString("-1") // Initialize max field value (field modulus - 1)
	}
	poseidon2FieldHasherDigest := &Poseidon2FieldHasherDigest{
		maxValue:    types.HashToBytes32(maxVal),
		StateStorer: gnarkposeidon2.NewMerkleDamgardHasher(),
		h:           field.Octuplet{},
		data:        []field.Element{},
	}

	return poseidon2FieldHasherDigest
}

// Reset resets the Hash to its initial state.
func (d *Poseidon2FieldHasherDigest) Reset() {
	d.data = d.data[:0]
	d.h = field.Octuplet{}
}

// WriteElement adds a field element to the running hash.
func (d *Poseidon2FieldHasherDigest) WriteElement(e field.Element) {
	d.data = append(d.data, e)
}

// WriteElements adds a slice of field elements to the running hash.
func (d *Poseidon2FieldHasherDigest) WriteElements(elems []field.Element) {
	d.data = append(d.data, elems...)
}

// SumElements returns the current hash as a field Octuplet.
func (d *Poseidon2FieldHasherDigest) SumElements(elems []field.Element) field.Octuplet {
	h1 := poseidon2.Poseidon2Sponge(d.data) // Poseidon2Sponge include the feedforward process
	vector.Add(d.h[:], h1[:], d.h[:])

	if elems != nil {
		h2 := poseidon2.Poseidon2Sponge(elems) // Poseidon2Sponge include the feedforward process
		vector.Add(d.h[:], h2[:], d.h[:])
	}

	d.data = d.data[:0]
	return d.h
}
func (d *Poseidon2FieldHasherDigest) Write(p []byte) (int, error) {

	elemByteSize := field.Bytes // 4 bytes = 1 field element

	start := 0
	for start < len(p) {
		end := start + elemByteSize
		var chunk []byte

		if end < len(p) {
			// full chunk, take as-is
			chunk = p[start:end]
		} else {
			// last chunk is short, left-pad with zeroes
			chunk = make([]byte, elemByteSize)
			copy(chunk[elemByteSize-(len(p)-start):], p[start:]) // left-pad
		}

		var elem field.Element
		elem.SetBytes(chunk)
		d.data = append(d.data, elem)

		start += elemByteSize
	}
	return len(p), nil
}

// Sum computes the poseidon2 hash of msg
func (d *Poseidon2FieldHasherDigest) Sum(msg []byte) []byte {
	h := d.SumElements(nil)
	bytes := types.HashToBytes32(h)
	return bytes[:]
}

// MaxBytes32 returns the maximal field value that Poseidon2 can work with
func (p Poseidon2FieldHasherDigest) MaxBytes32() types.Bytes32 {
	return p.maxValue
}
