package poseidon2

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

const blockSize = 8

// Poseidon2BlockCompression applies the Poseidon2 block compression function to a given block
// over a given state. This what is run under the hood by the Poseidon2 hash function
func Poseidon2BlockCompression(oldState, block [blockSize]field.Element) (newState [16]field.Element) {
	res := vortex.Hash{}
	var input [2 * blockSize]field.Element
	copy(input[:], oldState[:])
	copy(input[8:], block[:])

	// Create a buffer to hold the feed-forward input.
	copy(res[:], input[8:])

	var (
		matMulM4Tmp = make([][20]field.Element, 28)
		matMulM4    = make([][16]field.Element, 28)

		t              = make([][4]field.Element, 28)
		matMulExternal = make([][16]field.Element, 28)

		addRoundKey = make([][16]field.Element, 28)

		sBox = make([][16]field.Element, 28)

		sBoxSum = make([]field.Element, 28)

		matMulInternal = make([][16]field.Element, 28)
	)
	matMulM4Tmp[0], matMulM4[0], t[0], matMulExternal[0] = matMulExternalInPlace(input)
	input = matMulExternal[0]

	// Rounds 1 - 3
	for round := 1; round < 4; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, input)
		for col := 0; col < 16; col++ {
			sBox[round][col] = sBoxCompute(col, addRoundKey[round])
		}
		matMulM4Tmp[round], matMulM4[round], t[round], matMulExternal[round] = matMulExternalInPlace(sBox[round])
		input = matMulExternal[round]
	}

	// Rounds 4 - 24
	for round := 4; round < 25; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, input)
		sBox[round] = addRoundKey[round]
		sBox[round][0] = sBoxCompute(0, addRoundKey[round])
		sBoxSum[round], matMulInternal[round] = matMulInternalInPlace(sBox[round])
		input = matMulInternal[round]
	}

	// Rounds 24 - 27
	for round := 25; round < 28; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, input)
		for col := 0; col < 16; col++ {
			sBox[round][col] = sBoxCompute(col, addRoundKey[round])
		}
		matMulM4Tmp[round], matMulM4[round], t[round], matMulExternal[round] = matMulExternalInPlace(sBox[round])
		input = matMulExternal[round]
	}
	fmt.Printf("init matMulExternal= %v\n", input)

	// for i := range res {
	// 	res[i].Add(&res[i], &input[8+i])
	// }
	// return res
	return input
}

// when Width = 0 mod 4, the buffer is multiplied by circ(2M4,M4,..,M4)
// see https://eprint.iacr.org/2023/323.pdf
func matMulExternalInPlace(input [16]field.Element) (matMulM4Tmp [20]field.Element, matMulM4 [16]field.Element, t [4]field.Element, matMulExternal [16]field.Element) {

	if len(input) != 16 {
		panic("Input slice length must be 16")
	}
	for i := 0; i < 4; i++ {
		matMulM4Tmp[5*i].Add(&input[4*i], &input[4*i+1])
		matMulM4Tmp[5*i+1].Add(&input[4*i+2], &input[4*i+3])
		matMulM4Tmp[5*i+2].Add(&matMulM4Tmp[5*i], &matMulM4Tmp[5*i+1])
		matMulM4Tmp[5*i+3].Add(&matMulM4Tmp[5*i+2], &input[4*i+1])
		matMulM4Tmp[5*i+4].Add(&matMulM4Tmp[5*i+2], &input[4*i+3])

		// The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
		matMulM4[4*i].Add(&matMulM4Tmp[5*i], &matMulM4Tmp[5*i+3])
		matMulM4[4*i+1].Double(&input[4*i+2]).Add(&matMulM4[4*i+1], &matMulM4Tmp[5*i+3])
		matMulM4[4*i+2].Add(&matMulM4Tmp[5*i+1], &matMulM4Tmp[5*i+4])
		matMulM4[4*i+3].Double(&matMulM4[4*i]).Add(&matMulM4[4*i+3], &matMulM4Tmp[5*i+4])
	}

	for i := 0; i < 4; i++ {
		t[0].Add(&t[0], &matMulM4[4*i])
		t[1].Add(&t[1], &matMulM4[4*i+1])
		t[2].Add(&t[2], &matMulM4[4*i+2])
		t[3].Add(&t[3], &matMulM4[4*i+3])
	}
	for i := 0; i < 4; i++ {
		matMulExternal[4*i].Add(&matMulM4[4*i], &t[0])
		matMulExternal[4*i+1].Add(&matMulM4[4*i+1], &t[1])
		matMulExternal[4*i+2].Add(&matMulM4[4*i+2], &t[2])
		matMulExternal[4*i+3].Add(&matMulM4[4*i+3], &t[3])
	}
	return matMulM4Tmp, matMulM4, t, matMulExternal
}

// when Width = 0 mod 4 the matrix is filled with ones except on the diagonal
func matMulInternalInPlace(input [16]field.Element) (sBoxSum field.Element, matMulInternal [16]field.Element) {

	sBoxSum.Set(&input[0])
	for i := 1; i < 16; i++ {
		sBoxSum.Add(&sBoxSum, &input[i])
	}
	// mul by diag16:
	// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
	var temp field.Element
	matMulInternal[0].Sub(&sBoxSum, temp.Double(&input[0]))
	matMulInternal[1].Add(&sBoxSum, &input[1])
	matMulInternal[2].Add(&sBoxSum, temp.Double(&input[2]))
	temp.Set(&input[3]).Halve()
	matMulInternal[3].Add(&sBoxSum, &temp)
	matMulInternal[4].Add(&sBoxSum, temp.Double(&input[4]).Add(&temp, &input[4]))
	matMulInternal[5].Add(&sBoxSum, temp.Double(&input[5]).Double(&temp))
	temp.Set(&input[6]).Halve()
	matMulInternal[6].Sub(&sBoxSum, &temp)
	matMulInternal[7].Sub(&sBoxSum, temp.Double(&input[7]).Add(&temp, &input[7]))
	matMulInternal[8].Sub(&sBoxSum, temp.Double(&input[8]).Double(&temp))
	matMulInternal[9].Add(&sBoxSum, temp.Mul2ExpNegN(&input[9], 8))
	matMulInternal[10].Add(&sBoxSum, temp.Mul2ExpNegN(&input[10], 3))
	matMulInternal[11].Add(&sBoxSum, temp.Mul2ExpNegN(&input[11], 24))
	matMulInternal[12].Sub(&sBoxSum, temp.Mul2ExpNegN(&input[12], 8))
	matMulInternal[13].Sub(&sBoxSum, temp.Mul2ExpNegN(&input[13], 3))
	matMulInternal[14].Sub(&sBoxSum, temp.Mul2ExpNegN(&input[14], 4))
	matMulInternal[15].Sub(&sBoxSum, temp.Mul2ExpNegN(&input[15], 24))

	return sBoxSum, matMulInternal
}

// addRoundKey adds the round-th key to the buffer
func addRoundKeyCompute(round int, input [16]field.Element) (addRoundKey [16]field.Element) {
	for i := 0; i < len(poseidon2.RoundKeys[round]); i++ {
		addRoundKey[i].Add(&input[i], &poseidon2.RoundKeys[round][i])
	}
	for i := len(poseidon2.RoundKeys[round]); i < 16; i++ {
		addRoundKey[i] = input[i]
	}
	return addRoundKey
}

// SBoxCompute applies the SBoxCompute on buffer[index]
func sBoxCompute(index int, input [16]field.Element) (sBox field.Element) {
	// sbox degree is 3
	sBox.Square(&input[index]).
		Mul(&sBox, &input[index])
	return sBox
}
