package poseidon2

import (
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	blockSize = 8
)

// RoundKeys collects the Poseidon2 RoundKeys parsed as field elements
var RoundKeys [][]field.Element = func() [][]field.Element {
	return poseidon2.GetDefaultParameters().RoundKeys
}()

// Poseidon2BlockCompression applies the Poseidon2 block compression function to a given block
// over a given state. This what is run under the hood by the Poseidon2 hash function
func Poseidon2BlockCompression(oldState, block [blockSize]field.Element) (newState [blockSize]field.Element) {
	res := vortex.CompressPoseidon2(oldState, block)
	return res
}

// Poseidon2Sponge returns a Poseidon2 hash of an array of field elements
func Poseidon2Sponge(x []field.Element) (newState [blockSize]field.Element) {
	var state, xBlock [blockSize]field.Element
	for len(x) != 0 {
		if len(x) < blockSize {
			x = cloneLeftPadded(x, blockSize)
		}

		copy(xBlock[:], x[:])
		state = Poseidon2BlockCompression(state, xBlock)
		x = x[blockSize:]
	}
	return state
}

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompressionMekle(api frontend.API, oldState, block [blockSize]zk.WrappedVariable) (newState [blockSize]zk.WrappedVariable) {
	panic("unimplemented")
}

// cloneLeftPadded copies x into a new field element slice of size n.
// If len(x) < n, it will be padded on the left.
// len(x) > n will result in an error.
func cloneLeftPadded(x []field.Element, n int) []field.Element {
	if len(x) > n {
		utils.Panic("state/iv must not exceed the hash block size: %d > %d", len(x), n)
	}
	res := make([]field.Element, n)
	copy(res[n-len(x):], x)
	return res
}
