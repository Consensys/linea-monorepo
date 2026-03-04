package poseidon2_bn254

import (
	"errors"
	"fmt"
	"sync"

	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

const BlockSize = 1

// TODO @thomas fixme arbitrary max size of the buffer
const maxSizeBuf = 1024

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {

	// Sponge construction state
	state bn254fr.Element

	// data to hash
	buffer []bn254fr.Element

	verbose bool
}

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
	compressor           poseidon2.Permutation
	once                 sync.Once
)

func init() {
	once.Do(func() {
		// BN254 Poseidon2 parameters: t=2, rf=6, rp=50
		compressor = *poseidon2.NewPermutation(2, 6, 50)
	})
}

// Constructor for Poseidon2MDHasher
func NewMDHasher(verbose ...bool) *MDHasher {
	return &MDHasher{
		state:   bn254fr.Element{},
		buffer:  make([]bn254fr.Element, 0),
		verbose: len(verbose) > 0 && verbose[0],
	}
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.buffer = d.buffer[:0]
	d.state = bn254fr.Element{}
}

func (d *MDHasher) SetStateFrElement(s bn254fr.Element) {
	d.Reset()
	d.state.Set(&s)
}

// State returns the state of the hasher. The function must not mutate the
// MDHasher. If it needs to flush the buffer to access the state, it will clone
// the struct and flush there.
func (d *MDHasher) State() bn254fr.Element {

	// If the buffer is clean, we can short-path the execution and directly
	// return the state.
	if len(d.buffer) == 0 {
		return d.state
	}

	// If the buffer is not clean, we cannot clean it locally as it would modify
	// the state of the hasher locally. Instead, we clone the buffer and flush
	// the buffer on the clone.
	clone := NewMDHasher()
	clone.buffer = make([]bn254fr.Element, len(d.buffer))
	copy(clone.buffer, d.buffer)
	clone.state = d.state
	_ = clone.SumElement()
	return clone.state
}

// WriteKoalabearElements encodes Koalabear elements into BN254 field elements and adds them to the buffer.
func (d *MDHasher) WriteKoalabearElements(elmts ...field.Element) {
	_elmts := encoding.EncodeKoalabearsToBN254FrElement(elmts)
	d.WriteElements(_elmts...)
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts ...bn254fr.Element) {
	d.buffer = append(d.buffer, elmts[:]...)
}

func compress(left, right bn254fr.Element) bn254fr.Element {
	var x [2]bn254fr.Element
	x[0].Set(&left)
	x[1].Set(&right)
	res := x[1] // save right to feed forward later
	compressor.Permutation(x[:])
	res.Add(&res, &x[1])
	return res
}

func (d *MDHasher) SumElement() bn254fr.Element {
	if d.verbose {
		fmt.Printf("[native fs flush] oldState %v, buffer = %v\n", d.state.String(), bn254fr.Vector(d.buffer).String())
	}
	for i := 0; i < len(d.buffer); i++ {
		d.state = compress(d.state, d.buffer[i])
	}
	d.buffer = d.buffer[:0]
	return d.state
}
