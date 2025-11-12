package poseidon2_bls12377

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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
	state [2]fr.Element

	// data to hash
	buffer []fr.Element

	permutation poseidon2.Permutation
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.buffer = d.buffer[:0]
	d.state[0].SetZero()
	d.state[1].SetZero()
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) writeElements(elmts []fr.Element) {
	quo := (len(d.buffer) + len(elmts)) / maxSizeBuf
	rem := (len(d.buffer) + len(elmts)) % maxSizeBuf
	off := len(d.buffer)
	for i := 0; i < quo; i++ {
		d.buffer = append(d.buffer, elmts[:maxSizeBuf-off]...)
		_ = d.sumElement()
		d.buffer = d.buffer[:0] // flush the buffer once maxSizeBuf is reached
		off = len(d.buffer)
	}
	d.buffer = append(d.buffer, elmts[:rem-off]...)
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts []field.Element) {
	// convert koalabear to bls12377
	_elmts := ToBls12377(elmts)
	d.writeElements(_elmts)
}

// // WriteElements adds a slice of field elements to the running hash.
// func (d *MDHasher) Write(p []byte) (int, error) {

// 	elemByteSize := field.Bytes // 4 bytes = 1 field element

// 	if len(p)%elemByteSize != 0 {
// 		return 0, fmt.Errorf("input length is not a multiple of 4 byte size")
// 	}
// 	elems := make([]field.Element, 0, len(p)/elemByteSize)

// 	for start := 0; start < len(p); start += elemByteSize {
// 		chunk := p[start : start+elemByteSize]

// 		var elem field.Element
// 		elem.SetBytes(chunk)
// 		elems = append(elems, elem)

// 	}
// 	d.WriteElements(elems)
// 	return len(p), nil
// }

// sumElement follows MD scheme:
// b = [state || buf]
// permutation(b)
// b = [new_state || *** ], return new_state
func (d *MDHasher) sumElement() fr.Element {
	for len(d.buffer) != 0 {
		var buf [BlockSize]fr.Element
		// in this case we left pad by zeroes
		if len(d.buffer) < BlockSize {
			copy(buf[BlockSize-len(d.buffer):], d.buffer)
			d.buffer = d.buffer[:0]
		} else {
			copy(buf[:], d.buffer)
			d.buffer = d.buffer[BlockSize:]
		}
		// TODO @thomas fixme handle err
		copy(d.state[BlockSize:], buf[:])
		err := d.permutation.Permutation(d.state[:])
		if err != nil {
			panic(err)
		}
	}
	return d.state[0]
}

func (d *MDHasher) SumElement() field.Octuplet {
	t := d.sumElement()
	r := ToKoalabear([]fr.Element{t})
	return r[0]
}

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

// Constructor for Poseidon2MDHasher
func NewMDHasher() *MDHasher {
	// default parameters
	// see https://github.com/Consensys/gnark-crypto/blob/master/ecc/bls12-377/fr/poseidon2/poseidon2.go
	permutation := poseidon2.NewPermutation(2, 6, 26)
	return &MDHasher{
		StateStorer: poseidon2.NewMerkleDamgardHasher(),
		permutation: *permutation,
	}
}

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompressionMekle(api frontend.API, oldState, block [BlockSize]zk.WrappedVariable) (newState [BlockSize]zk.WrappedVariable) {
	panic("unimplemented")
}

// example:
// v = [a, b, c, d, e, f, g, h, i]
// returns
// [ (a<<31+b)<<192 + (c<<31+d)<<128 + (e<<31+f)<<64 + (g<<31+h), (i<<31)<<192]
func ToBls12377(v []field.Element) []fr.Element {
	offset := 0
	res := make([]fr.Element, (len(v)+7)/8)
	for i := 0; i < len(v); i += 8 {
		for j := 0; j < 8 || i+j < len(v); j++ {
			if j%2 == 0 {
				offset = 0
			} else {
				offset = 31
			}
			res[i/8][j] += uint64(v[i+j][0]) << offset
		}
	}
	return res
}

// example:
// v = [a]
// returns
// [a & mask, (a & mask<<31)>>31, ...]
// where mask = 1<<31-1
func ToKoalabear(v []fr.Element) []field.Octuplet {
	mask := uint32((1 << 31) - 1)
	res := make([]field.Octuplet, 8*len(v))
	for i := 0; i < len(v); i++ {
		for j := 0; j < 4; j++ {
			res[i][2*j][0] = uint32(v[i][j]) & mask
			res[i][2*j+1][0] = (uint32(v[i][j]) & (mask << 31)) >> 31
		}
	}
	return res
}
