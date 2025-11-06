package common

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"math/big"
)

// BlockCompression mocked version of mimc.BlockCompression that operates on
// several limbs of 16 bits each, instead of a single field element.
func BlockCompression(oldState, block []field.Element) []field.Element {
	var oldStateField, blockField field.Element

	recomposeField(oldState, 16, &oldStateField)
	recomposeField(block, 16, &blockField)

	newStateField := mimc.BlockCompression(oldStateField, blockField)

	newStateLimbs := make([]field.Element, 16)
	decomposeField(newStateField, 16, newStateLimbs)

	return newStateLimbs
}

func decomposeField(input field.Element, nbBits uint, res []field.Element) {
	resBig := make([]*big.Int, len(res))
	for i := range res {
		resBig[i] = res[i].BigInt(new(big.Int))
	}

	base := new(big.Int).Lsh(big.NewInt(1), nbBits)
	tmp := input.BigInt(new(big.Int))
	for i := 0; i < len(res); i++ {
		resBig[len(res)-1-i].Mod(tmp, base)
		tmp.Rsh(tmp, nbBits)
	}

	for i := range res {
		res[i].SetBigInt(resBig[i])
	}
}

func recomposeField(inputs []field.Element, nbBits uint, res *field.Element) {
	resBig := big.NewInt(0)

	buffer := new(big.Int)
	for i := range inputs {
		resBig.Lsh(resBig, nbBits)
		resBig.Add(resBig, inputs[i].BigInt(buffer))
	}

	res.SetBigInt(resBig)
}
