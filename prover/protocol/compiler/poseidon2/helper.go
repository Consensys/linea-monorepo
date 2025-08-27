package poseidon2

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	width           = 16
	blockSize       = 8
	fullRounds      = 28
	partialRounds   = 3
	matMulM4TmpSize = 20
	tSize           = 4
)

// poseidon2BlockCompression applies the Poseidon2 block compression function to a given block
// over a given state. This what is run under the hood by the Poseidon2 hash function
func poseidon2BlockCompression(oldState, block [blockSize]field.Element) (newState [blockSize]field.Element) {

	state := make([]field.Element, width)
	copy(state[:8], oldState[:])
	copy(state[8:], block[:])

	// Create a buffer to hold the feed-forward input.
	copy(newState[:], state[8:])

	// Poseidon2 compression
	// The `vortex.CompressPoseidon2` function is the canonical reference.
	// This function implements the Poseidon2 algorithm from scratch for
	// verification and clarity.

	var (
		matMulM4Tmp    = make([][]field.Element, fullRounds) // [fullRounds][matMulM4TmpWidth]field.Element
		matMulM4       = make([][]field.Element, fullRounds) // [fullRounds][width]field.Element
		t              = make([][]field.Element, fullRounds) // [fullRounds][tWidth]field.Element
		matMulExternal = make([][]field.Element, fullRounds) // [fullRounds][width]field.Element
		addRoundKey    = make([][]field.Element, fullRounds) // [fullRounds][width]field.Element
		sBox           = make([][]field.Element, fullRounds) // [fullRounds][width]field.Element
		sBoxSum        = make([]field.Element, fullRounds)
		matMulInternal = make([][]field.Element, fullRounds) // [fullRounds][width]field.Element
	)

	// Initial round
	matMulM4Tmp[0], matMulM4[0], t[0], matMulExternal[0] = matMulExternalInPlace(state)
	state = matMulExternal[0]

	// Rounds 1 - 3
	// External rounds
	for round := 1; round < 1+partialRounds; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, state)
		sBox[round] = make([]field.Element, width)
		for col := 0; col < width; col++ {
			sBox[round][col] = sBoxCompute(col, round, addRoundKey[round])
		}
		matMulM4Tmp[round], matMulM4[round], t[round], matMulExternal[round] = matMulExternalInPlace(sBox[round])
		state = matMulExternal[round]
	}

	// Rounds 4 - 24
	// Internal rounds
	for round := 1 + partialRounds; round < fullRounds-partialRounds; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, state)
		sBox[round] = make([]field.Element, width)
		for col := 0; col < width; col++ {
			sBox[round][col] = sBoxCompute(col, round, addRoundKey[round])
		}
		sBoxSum[round], matMulInternal[round] = matMulInternalInPlace(sBox[round])
		state = matMulInternal[round]
	}

	// Rounds 24 - 27
	// External rounds
	for round := fullRounds - partialRounds; round < fullRounds; round++ {
		addRoundKey[round] = addRoundKeyCompute(round-1, state)
		sBox[round] = make([]field.Element, width)

		for col := 0; col < width; col++ {
			sBox[round][col] = sBoxCompute(col, round, addRoundKey[round])
		}
		matMulM4Tmp[round], matMulM4[round], t[round], matMulExternal[round] = matMulExternalInPlace(sBox[round])
		state = matMulExternal[round]
	}

	// Final round
	// Feed forward
	for i := range newState {
		newState[i].Add(&newState[i], &state[8+i])
	}
	return newState
}

// matMulExternalInPlace applies the external matrix multiplication.
// see https://eprint.iacr.org/2023/323.pdf
func matMulExternalInPlace(input []field.Element) (matMulM4Tmp []field.Element, matMulM4 []field.Element, t []field.Element, matMulExternal []field.Element) {

	if len(input) != width {
		utils.Panic("Input slice length must be %v", width)
	}
	matMulM4Tmp = make([]field.Element, matMulM4TmpSize)
	matMulM4 = make([]field.Element, width)
	t = make([]field.Element, tSize)
	matMulExternal = make([]field.Element, width)

	for i := 0; i < 4; i++ {
		matMulM4Tmp[5*i].Add(&input[4*i], &input[4*i+1])
		matMulM4Tmp[5*i+1].Add(&input[4*i+2], &input[4*i+3])
		matMulM4Tmp[5*i+2].Add(&matMulM4Tmp[5*i], &matMulM4Tmp[5*i+1])
		matMulM4Tmp[5*i+3].Add(&matMulM4Tmp[5*i+2], &input[4*i+1])
		matMulM4Tmp[5*i+4].Add(&matMulM4Tmp[5*i+2], &input[4*i+3])

		// The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
		matMulM4[4*i+3].Double(&input[4*i]).Add(&matMulM4[4*i+3], &matMulM4Tmp[5*i+4])
		matMulM4[4*i+1].Double(&input[4*i+2]).Add(&matMulM4[4*i+1], &matMulM4Tmp[5*i+3])
		matMulM4[4*i].Add(&matMulM4Tmp[5*i], &matMulM4Tmp[5*i+3])
		matMulM4[4*i+2].Add(&matMulM4Tmp[5*i+1], &matMulM4Tmp[5*i+4])
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

// matMulInternal applies the internal matrix multiplication.
func matMulInternalInPlace(input []field.Element) (sBoxSum field.Element, matMulInternal []field.Element) {
	if len(input) != width {
		utils.Panic("Input slice length must be %v", width)
	}
	matMulInternal = make([]field.Element, 16)

	sBoxSum.Set(&input[0])
	for i := 1; i < width; i++ {
		sBoxSum.Add(&sBoxSum, &input[i])
	}
	// mul by diag16:
	// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
	two := field.NewElement(2)
	half := field.NewElement(1065353217)
	halfExp3 := field.NewElement(1864368129)
	halfExp4 := field.NewElement(1997537281)
	halfExp8 := field.NewElement(2122383361)
	halfExp24 := field.NewElement(127) // -127

	var temp field.Element
	matMulInternal[0].Sub(&sBoxSum, temp.Double(&input[0]))
	matMulInternal[1].Add(&sBoxSum, &input[1])
	matMulInternal[2].Add(&sBoxSum, temp.Mul(&input[2], &two))
	matMulInternal[3].Add(&sBoxSum, temp.Mul(&input[3], &half))
	matMulInternal[4].Add(&sBoxSum, temp.Double(&input[4]).Add(&temp, &input[4]))
	matMulInternal[5].Add(&sBoxSum, temp.Double(&input[5]).Double(&temp))
	matMulInternal[6].Sub(&sBoxSum, temp.Mul(&input[6], &half))
	matMulInternal[7].Sub(&sBoxSum, temp.Double(&input[7]).Add(&temp, &input[7]))
	matMulInternal[8].Sub(&sBoxSum, temp.Double(&input[8]).Double(&temp))
	matMulInternal[9].Add(&sBoxSum, temp.Mul(&input[9], &halfExp8))
	matMulInternal[10].Add(&sBoxSum, temp.Mul(&input[10], &halfExp3))
	matMulInternal[11].Sub(&sBoxSum, temp.Mul(&input[11], &halfExp24))
	matMulInternal[12].Sub(&sBoxSum, temp.Mul(&input[12], &halfExp8))
	matMulInternal[13].Sub(&sBoxSum, temp.Mul(&input[13], &halfExp3))
	matMulInternal[14].Sub(&sBoxSum, temp.Mul(&input[14], &halfExp4))
	matMulInternal[15].Add(&sBoxSum, temp.Mul(&input[15], &halfExp24))

	return sBoxSum, matMulInternal
}

// addRoundKey adds the round-th key to the input
func addRoundKeyCompute(round int, input []field.Element) (addRoundKey []field.Element) {
	if len(input) != width {
		utils.Panic("Input slice length must be %v", width)
	}
	addRoundKey = make([]field.Element, width)
	for i := 0; i < len(poseidon2.RoundKeys[round]); i++ {
		addRoundKey[i].Add(&input[i], &poseidon2.RoundKeys[round][i])
	}
	for i := len(poseidon2.RoundKeys[round]); i < width; i++ {
		addRoundKey[i] = input[i]
	}
	return addRoundKey
}

// SBoxCompute applies the  s-box on input[index]
func sBoxCompute(index, poseidon2Round int, input []field.Element) (sBox field.Element) {
	if len(input) != width {
		utils.Panic("Input slice length must be %v", width)
	}

	if poseidon2Round < fullRounds-partialRounds && poseidon2Round > partialRounds && index != 0 {
		sBox = input[index]
	} else {
		// sbox degree is 3
		sBox.Square(&input[index]).
			Mul(&sBox, &input[index])
	}

	return sBox
}
