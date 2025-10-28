package poseidon2

import (
	"errors"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// type GnarkFiatShamir struct {

// 	hasher hash.StateStorer
// Write(zkWrappedVariable) Sum() Reset()
// SetState(state zkWrappedVariable), State()

// 	api frontend.API
// }

// Sum() frontend.Variable

// // Write populate the internal state of the hash function with data. The inputs are native field elements.
// Write(data ...frontend.Variable)

// // Reset empty the internal state and put the intermediate state to zero.
// Reset()

// State() []frontend.Variable
// // SetState sets the state of the hash function from a previously stored
// // state retrieved using [StateStorer.State] method. The implementation
// // returns an error if the number of supplied Variable does not match the
// // number of Variable expected.
// SetState(state []frontend.Variable) error

type GHash zk.Octuplet

type GnarkHasher struct {
	apiGen zk.GenericApi

	// Sponge construction state
	state GHash

	// data to hash
	buffer []zk.WrappedVariable
}

// NewGnarkHasher returns a new GHash
func NewGnarkHasher(api frontend.API) (GnarkHasher, error) {
	var res GnarkHasher
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return res, err
	}
	res.apiGen = apiGen
	for i := 0; i < 8; i++ {
		res.state[i] = zk.ValueOf(0)
	}
	return res, nil
}

func (h *GnarkHasher) Reset() {
	for i := 0; i < 8; i++ {
		h.state[i] = zk.ValueOf(0)
	}
}

func (h *GnarkHasher) Write(data ...zk.WrappedVariable) {
	h.buffer = append(h.buffer, data...)
}

func (h *GnarkHasher) Sum() GHash {

	for len(h.buffer) != 0 {
		var buf [blockSize]zk.WrappedVariable
		for i := 0; i < blockSize; i++ {
			buf[i] = zk.ValueOf(0)
		}
		// in this case we left pad by zeroes
		if len(h.buffer) < blockSize {
			copy(buf[blockSize-len(h.buffer):], h.buffer)
			h.buffer = h.buffer[:0]
		} else {
			copy(buf[:], h.buffer)
			h.buffer = h.buffer[blockSize:]
		}

		// h.state = CompressPoseidon2(h.apiGen, h.state, buf)
	}
	return h.state
}

func CompressPoseidon2(apiGen zk.GenericApi, a, b GHash) GHash {
	res := GHash{}

	var x [16]zk.WrappedVariable
	copy(x[:], a[:])
	copy(x[8:], b[:])

	// Create a buffer to hold the feed-forward input.
	copy(res[:], x[8:])
	if err := compressPerm.Permutation(apiGen, x[:]); err != nil {
		// can't error (size is correct)
		panic(err)
	}

	for i := range res {
		res[i] = *apiGen.Add(&res[i], &x[8+i])
	}
	return res
}

//----------------------------------------------------------------------
//------------------------- copy from gnark ----------------
//----------------------------------------------------------------------

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
	compressPerm         permutation
	once                 sync.Once
)

func init() {
	once.Do(func() {
		compressPerm = NewPermutation()
	})
}

func NewPermutation() permutation {
	// same params than gnark-crypto/field/koalabear/vortex/hash.go
	params := poseidon2.NewParameters(16, 6, 21)
	return permutation{params: params}
}

type permutation struct {
	params *poseidon2.Parameters
}

// sBox applies the sBox on buffer[index]
func (h *permutation) sBox(apiGen zk.GenericApi, index int, input []zk.WrappedVariable) {
	// sbox degree is 3
	tmp := input[index]
	input[index] = *apiGen.Mul(&input[index], &input[index])
	input[index] = *apiGen.Mul(&tmp, &input[index])

}

// matMulM4 computes
// s <- M4*s
// where M4=
// (2 3 1 1)
// (1 2 3 1)
// (1 1 2 3)
// (3 1 1 2)
// on chunks of 4 elements on each part of the buffer
// see https://eprint.iacr.org/2023/323.pdf appendix B for the addition chain
func (h *permutation) matMulM4InPlace(apiGen zk.GenericApi, s []zk.WrappedVariable) {

	c := len(s) / 4
	for i := 0; i < c; i++ {
		// var t01, t23, t0123, t01123, t01233 zk.WrappedVariable
		apiGen.Add(&s[4*i], &s[4*i+1])
		// t01 = *apiGen.Add(&s[4*i], &s[4*i+1])
		// t23 = *apiGen.Add(&s[4*i+2], &s[4*i+3])
		// t0123 = *apiGen.Add(&t01, &t23)
		// t01123 = *apiGen.Add(&t0123, &s[4*i+1])
		// t01233 = *apiGen.Add(&t0123, &s[4*i+3])
		// // The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
		// s[4*i+3] = *apiGen.Add(&s[4*i], &s[4*i])
		// s[4*i+3] = *apiGen.Add(&s[4*i+3], &t01233)
		// s[4*i+1] = *apiGen.Add(&s[4*i+2], &s[4*i+2])
		// s[4*i+1] = *apiGen.Add(&s[4*i+1], &t01123)
		// s[4*i] = *apiGen.Add(&t01, &t01123)
		// s[4*i+2] = *apiGen.Add(&t23, &t01233)
	}
}

// when t=2,3 the buffer is multiplied by circ(2,1) and circ(2,1,1)
// see https://eprint.iacr.org/2023/323.pdf page 15, case t=2,3
//
// when t=0[4], the buffer is multiplied by circ(2M4,M4,..,M4)
// see https://eprint.iacr.org/2023/323.pdf
func (h *permutation) matMulExternalInPlace(apiGen zk.GenericApi, input []zk.WrappedVariable) {
	if h.params.Width%4 != 0 {
		panic("only Width = 0 mod 4 are supported")
	}
	// at this stage t is supposed to be a multiple of 4
	// the MDS matrix is circ(2M4,M4,..,M4)
	h.matMulM4InPlace(apiGen, input)
	// tmp := make([]zk.WrappedVariable, 4)
	// for i := 0; i < h.params.Width/4; i++ {
	// 	tmp[0] = *apiGen.Add(&tmp[0], &input[4*i])
	// 	tmp[1] = *apiGen.Add(&tmp[1], &input[4*i+1])
	// 	tmp[2] = *apiGen.Add(&tmp[2], &input[4*i+2])
	// 	tmp[3] = *apiGen.Add(&tmp[3], &input[4*i+3])
	// }
	// for i := 0; i < h.params.Width/4; i++ {
	// 	input[4*i] = *apiGen.Add(&input[4*i], &tmp[0])
	// 	input[4*i+1] = *apiGen.Add(&input[4*i+1], &tmp[1])
	// 	input[4*i+2] = *apiGen.Add(&input[4*i+2], &tmp[2])
	// 	input[4*i+3] = *apiGen.Add(&input[4*i+3], &tmp[3])
	// }
}

// when t=2,3 the matrix are respectively [[2,1][1,3]] and [[2,1,1][1,2,1][1,1,3]]
// otherwise the matrix is filled with ones except on the diagonal,
func (h *permutation) matMulInternalInPlace(apiGen zk.GenericApi, input []zk.WrappedVariable) {
	for i := 0; i < len(input); i++ {
		apiGen.Println(&input[i])
	}
	// two := zk.ValueOf(2)
	// if h.params.Width == 2 {
	// 	sum := *apiGen.Add(&input[0], &input[1])
	// 	input[0] = *apiGen.Add(&input[0], &sum)
	// 	input[1] = *apiGen.Mul(&two, &input[1])
	// 	input[1] = *apiGen.Add(&input[1], &sum)
	// } else if h.params.Width == 3 {
	// 	sum := *apiGen.Add(&input[0], &input[1])
	// 	sum = *apiGen.Add(&sum, &input[2])
	// 	input[0] = *apiGen.Add(&input[0], &sum)
	// 	input[1] = *apiGen.Add(&input[1], &sum)
	// 	input[2] = *apiGen.Mul(&input[2], &two)
	// 	input[2] = *apiGen.Add(&input[2], &sum)
	// } else {
	// 	// TODO: we don't have general case implemented in gnark-crypto side.
	// 	// Currently we only have the hardcoded matrices for t=2,3. If we would
	// 	// use `h.params.diagInternalMatrices` we would need to set it, but
	// 	// currently they are nil.

	// 	// var sum frontend.Variable
	// 	// sum = input[0]
	// 	// for i := 1; i < h.params.width; i++ {
	// 	// 	sum = api.Add(sum, input[i])
	// 	// }
	// 	// for i := 0; i < h.params.width; i++ {
	// 	// 	input[i] = api.Mul(input[i], h.params.diagInternalMatrices[i])
	// 	// 	input[i] = api.Add(input[i], sum)
	// 	// }
	// 	panic("only T=2,3 is supported")
	// }
}

// addRoundKeyInPlace adds the round-th key to the buffer
func (h *permutation) addRoundKeyInPlace(apiGen zk.GenericApi, round int, input []zk.WrappedVariable) {
	var rk zk.WrappedVariable
	for i := 0; i < len(h.params.RoundKeys[round]); i++ {
		rk = zk.ValueOf(h.params.RoundKeys[round][i])
		input[i] = *apiGen.Add(&input[i], &rk)
	}
}

// permutation applies the permutation on input, and stores the result in input.
func (h *permutation) Permutation(apiGen zk.GenericApi, input []zk.WrappedVariable) error {
	if len(input) != h.params.Width {
		return ErrInvalidSizebuffer
	}

	// external matrix multiplication, cf https://eprint.iacr.org/2023/323.pdf page 14 (part 6)
	h.matMulExternalInPlace(apiGen, input)

	// rf := h.params.NbFullRounds / 2
	// for i := 0; i < rf; i++ {
	// 	// one round = matMulExternal(sBox_Full(addRoundKey))
	// 	h.addRoundKeyInPlace(apiGen, i, input)
	// 	for j := 0; j < h.params.Width; j++ {
	// 		h.sBox(apiGen, j, input)
	// 	}
	// 	h.matMulExternalInPlace(apiGen, input)
	// }

	// for i := rf; i < rf+h.params.NbPartialRounds; i++ {
	// 	// one round = matMulInternal(sBox_sparse(addRoundKey))
	// 	h.addRoundKeyInPlace(apiGen, i, input)
	// 	h.sBox(apiGen, 0, input)
	// 	h.matMulInternalInPlace(apiGen, input)
	// }
	// for i := rf + h.params.NbPartialRounds; i < h.params.NbFullRounds+h.params.NbPartialRounds; i++ {
	// 	// one round = matMulExternal(sBox_Full(addRoundKey))
	// 	h.addRoundKeyInPlace(apiGen, i, input)
	// 	for j := 0; j < h.params.Width; j++ {
	// 		h.sBox(apiGen, j, input)
	// 	}
	// 	h.matMulExternalInPlace(apiGen, input)
	// }

	return nil
}
