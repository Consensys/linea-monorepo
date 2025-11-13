package poseidon2_bls12377

import (
	"errors"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const BlockSize = 1

// TODO @thomas fixme arbitrary max size of the buffer
const maxSizeBuf = 1024

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {
	hash.StateStorer
	maxValue types.Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction state
	state fr.Element

	// data to hash
	buffer []fr.Element
}

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
	compressor           poseidon2.Permutation
	once                 sync.Once
)

func init() {
	once.Do(func() {
		// default parameters
		compressor = *poseidon2.NewPermutation(2, 6, 26)
	})
}

// Constructor for Poseidon2MDHasher
func NewMDHasher() *MDHasher {
	// var maxVal fr.Element
	// maxVal.SetOne().Neg(&maxVal)
	return &MDHasher{
		StateStorer: poseidon2.NewMerkleDamgardHasher(),
		state:       fr.Element{},
		buffer:      make([]fr.Element, 0),
	}
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.buffer = d.buffer[:0]
	d.state = fr.Element{}
}

// SetStateFrElement modifies the state (??)
func (d *MDHasher) SetStateFrElement(state fr.Element) {
	d.state.Set(&state)
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts []fr.Element) {
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

func compress(left, right fr.Element) fr.Element {
	var x [2]fr.Element
	x[0].Set(&left)
	x[1].Set(&right)
	res := x[1] // save right to feed forward later
	compressor.Permutation(x[:])
	res.Add(&res, &x[1])
	return res
}

func (d *MDHasher) SumElement() fr.Element {
	for i := 0; i < len(d.buffer); i++ {
		d.state = compress(d.state, d.buffer[i])
	}
	d.buffer = d.buffer[:0]
	return d.state
}

// // WriteElements adds a slice of field elements to the running hash.
// func (d *MDHasher) Write(p []byte) (int, error) {

// 	elemByteSize := field.Bytes // 4 bytes = 1 field element

// 	if len(p)%elemByteSize != 0 {
// 		return 0, fmt.Errorf("input length is not a multiple of 4 byte size")
// 	}
// 	elems := make([]fr.Element, 0, len(p)/elemByteSize)

// 	for start := 0; start < len(p); start += elemByteSize {
// 		chunk := p[start : start+elemByteSize]

// 		var elem fr.Element
// 		elem.SetBytes(chunk)
// 		elems = append(elems, elem)

// 	}
// 	d.WriteElements(elems)
// 	return len(p), nil
// }

// // Sum computes the poseidon2 hash of msg
// func (d *MDHasher) Sum(msg []byte) []byte {
// 	d.Write(msg)
// 	h := d.SumElement()
// 	bytes := types.HashToBytes32(h)
// 	return bytes[:]
// }
// func (d MDHasher) MaxBytes32() types.Bytes32 {
// 	return d.maxValue
// }

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
// func GnarkBlockCompressionMekle(api frontend.API, oldState, block [BlockSize]zk.WrappedVariable) (newState [BlockSize]zk.WrappedVariable) {
// 	panic("unimplemented")
// }
