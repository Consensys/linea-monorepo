package poseidon2

import (
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/gnark/frontend"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const (
	blockSize = 8
)

// NewMiMC wraps [mimc.NewMiMC], this is used to limit the number of gnark-crypto imports.
func NewMiMC() hash.Hash {
	return mimc.NewMiMC()
}

// NewPoseidon2 wraps [poseidon2.NewMerkleDamgardHasher], this is used to limit the number of gnark-crypto imports.
func NewPoseidon2() hash.Hash {
	return poseidon2.NewMerkleDamgardHasher()
}

// Constants collects the MiMC constants parsed as field elements
var Constants []field.Element = func() []field.Element {
	bigConsts := mimc.GetConstants()
	res := make([]field.Element, len(bigConsts))
	for i := range res {
		res[i].SetBigInt(&bigConsts[i])
	}
	return res
}()

// TODO@yao Sponge and Merkle

// BlockCompression applies the MiMC block compression function to a given block
// over a given state. This what is run under the hood by the MiMC hash function
// in Miyaguchi-Preneel mode.
func BlockCompressionMekle(oldState, block [blockSize]field.Element) (newState [blockSize]field.Element) {
	res := vortex.Hash{}
	var x [2 * blockSize]field.Element
	copy(x[:], oldState[:])
	copy(x[8:], block[:])

	// Create a buffer to hold the feed-forward input.
	copy(res[:], x[8:])
	if err := poseidon2.NewPermutation(16, 6, 21).Permutation(x[:]); err != nil {
		panic(err)
	}

	for i := range res {
		res[i].Add(&res[i], &x[8+i])
	}
	return res
}

// BlockCompression applies the MiMC block compression function to a given block
// over a given state. This what is run under the hood by the MiMC hash function
// in Miyaguchi-Preneel mode.
func BlockCompression(oldState, block field.Element) (newState field.Element) {
	res := block
	var tmp field.Element

	// s <- (s + old + c)^17
	for i := range Constants {
		// We don't use the loop value of Constant to explictly
		// show the linter that we are not mutating the loop value.
		c := Constants[i]
		res.Add(&res, &c)
		res.Add(&res, &oldState)
		tmp.Square(&res)
		tmp.Square(&tmp)
		tmp.Square(&tmp)
		tmp.Square(&tmp)
		res.Mul(&tmp, &res)
	}

	res.Add(&res, &oldState).Add(&res, &oldState).Add(&res, &block)
	return res
}

// GnarkBlockCompression applies the MiMC permutation to a given block within
// a gnark circuit and mirrors exactly [BlockCompression].
func GnarkBlockCompression(api frontend.API, oldState, block frontend.Variable) (newState frontend.Variable) {

	res := block
	var tmp frontend.Variable

	// s <- (s + old + c)^17
	for _, c := range Constants {
		res = api.Add(res, c)
		res = api.Add(res, oldState)
		tmp = api.Mul(res, res)
		tmp = api.Mul(tmp, tmp)
		tmp = api.Mul(tmp, tmp)
		tmp = api.Mul(tmp, tmp)
		res = api.Mul(tmp, res)
	}

	res = api.Add(res, oldState)
	res = api.Add(res, oldState)
	res = api.Add(res, block)

	return res
}

// Poseidon2HashVec hashes a vector of field elements to a leaf
func Poseidon2HashVec(v []field.Element) (h [8]field.Element) {
	state := poseidon2.NewMerkleDamgardHasher()
	for i := range v {
		vBytes := v[i].Bytes()
		state.Write(vBytes[:])
	}
	h = types.Bytes32ToHash(types.Bytes32(state.Sum(nil)))
	return h
}
