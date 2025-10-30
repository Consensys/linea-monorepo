package poseidon2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

// var rng = rand.New(utils.NewRandSource(0)) // #nosec G404

// // Test different input sizes to ensure consistency with the Merkle-Damgard construction.
// var testCases = []struct {
// 	name        string
// 	numElements int
// }{
// 	{"SmallInput", 10},
// 	{"LargeInput", 512},
// 	{"SingleBlock", 8},
// }

// // This test ensures that the Poseidon2BlockCompression function is correctly implemented and produces the same output as
// // the poseidon2.Poseidon2(), which uses Write and Sum methods to get the final hash output
// //
// // We hash and compress one Octuplet at a time
// func TestPoseidon2BlockCompression(t *testing.T) {

// 	for i := 0; i < 100; i++ {
// 		var state field.Octuplet
// 		var input field.Octuplet

// 		var inputBytes [32]byte
// 		for i := 0; i < 8; i++ {
// 			startIndex := i * 4
// 			input[i] = field.PseudoRand(rng)
// 			valBytes := input[i].Bytes()
// 			copy(inputBytes[startIndex:startIndex+4], valBytes[:])
// 		}

// 		// Compute hash using the Poseidon2BlockCompression.
// 		h := vortex.CompressPoseidon2(state, input)

// 		// Compute hash using the NewMerkleDamgardHasher implementation.
// 		merkleHasher := poseidon2.Poseidon2()
// 		merkleHasher.Reset()
// 		merkleHasher.Write(inputBytes[:]) // write one 32 bytes (equivalent to one Octuplet)
// 		newBytes := merkleHasher.Sum(nil)

// 		var result field.Octuplet
// 		for i := 0; i < 8; i++ {
// 			startIndex := i * 4
// 			segment := newBytes[startIndex : startIndex+4]
// 			var newElement koalabear.Element
// 			newElement.SetBytes(segment)
// 			result[i] = newElement
// 			require.Equal(t, result[i].String(), h[i].String())

// 		}

// 	}
// }

// // This test ensures that the Poseidon2Sponge function is correctly implemented and produces the same output as
// // the poseidon2.Poseidon2(), which uses Write and Sum methods to get the final hash output
// // We write and compress the 'whole slice'
// func TestPoseidon2SpongeConsistency(t *testing.T) {
// 	t.Parallel()

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			for i := 0; i < 10; i++ {
// 				// Generate random input.
// 				input := make([]field.Element, tc.numElements)
// 				inputBytes := make([]byte, 0, tc.numElements*field.Bytes)
// 				for j := 0; j < tc.numElements; j++ {
// 					input[j] = field.PseudoRand(rng)
// 					bytes := input[j].Bytes()
// 					inputBytes = append(inputBytes, bytes[:]...)
// 				}

// 				// Compute hash using the Poseidon2Sponge function.
// 				state := poseidon2.Poseidon2Sponge(input)

// 				// Compute hash using the reference Merkle-Damgard hasher.
// 				merkleHasher := poseidon2.Poseidon2()
// 				merkleHasher.Reset()
// 				merkleHasher.Write(inputBytes[:])
// 				newBytes := merkleHasher.Sum(nil)

// 				var result field.Octuplet
// 				for i := 0; i < 8; i++ {
// 					startIndex := i * 4
// 					segment := newBytes[startIndex : startIndex+4]
// 					var newElement koalabear.Element
// 					newElement.SetBytes(segment)
// 					result[i] = newElement
// 					require.Equal(t, result[i].String(), state[i].String())
// 				}

// 			}
// 		})
// 	}
// }

// func TestFieldHasher(t *testing.T) {
// 	assert := require.New(t)

// 	h1 := poseidon2.Poseidon2()
// 	h2 := poseidon2.Poseidon2()
// 	randInputs := make(field.Vector, 10)
// 	randInputs.MustSetRandom()

// 	// test Write + Sum
// 	for _, elem := range randInputs {
// 		h1.Write(elem.Marshal())
// 	}
// 	dgst1 := h1.Sum(nil)
// 	var dgst1Byte32 types.Bytes32
// 	copy(dgst1Byte32[:], dgst1[:])

// 	// test WriteElement + SumElement
// 	h2.WriteElements(randInputs)
// 	dgst2 := h2.SumElement()
// 	assert.Equal(types.Bytes32ToHash(dgst1Byte32), dgst2, "hashes do not match")

// }

type GnarkHasherCircuit struct {
	Inputs []zk.WrappedVariable
	Ouput  GHash
}

func (ghc *GnarkHasherCircuit) Define(api frontend.API) error {

	h, err := NewGnarkHasher(api)
	if err != nil {
		return err
	}

	// write elmts
	h.Write(ghc.Inputs...)

	// sum
	// res := h.Sum()
	h.Sum()

	// check the result
	// apiGen, err := zk.NewGenericApi(api)
	// if err != nil {
	// 	return err
	// }
	// for i := 0; i < len(res); i++ {
	// 	apiGen.Println(&res[i])
	// }
	// for i := 0; i < 8; i++ {
	// apiGen.AssertIsEqual(&ghc.Ouput[i], &res[i])
	// }

	return nil
}

func getGnarkHasherCircuitWitness() (*GnarkHasherCircuit, *GnarkHasherCircuit) {

	// values to hash
	nbElmts := 2
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		// vals[i].SetRandom()
		vals[i].SetUint64(uint64(10 + i))
	}

	// sum
	phasher := Poseidon2()
	phasher.WriteElements(vals)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness GnarkHasherCircuit
	circuit.Inputs = make([]zk.WrappedVariable, nbElmts)
	witness.Inputs = make([]zk.WrappedVariable, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = zk.ValueOf(vals[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Ouput[i] = zk.ValueOf(res[i].String())
	}

	return &circuit, &witness

}

func TestCircuit(t *testing.T) {

	{
		circuit, witness := getGnarkHasherCircuitWitness()

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getGnarkHasherCircuitWitness()

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
