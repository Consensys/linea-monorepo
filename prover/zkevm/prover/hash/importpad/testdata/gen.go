package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

const bitsPerLimb = 16
const bytesPerLimb = bitsPerLimb / 8

func main() {

	var (
		limbs     = make([][]field.Element, common.NbLimbU128)
		nBytes    = make([]field.Element, 32)
		toHash    = make([]field.Element, 32)
		index     = make([]field.Element, 32)
		hashNum   = make([]field.Element, 32)
		rng       = rand.New(rand.NewChaCha8([32]byte{}))
		hashNums  = []int{0, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0}
		toHashInt = []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0}
		oF        = files.MustOverwrite("./testdata/input.csv")
	)

	for i := range limbs {
		limbs[i] = make([]field.Element, 32)
	}

	for i := range hashNum {

		if i == 0 {
			index[i] = field.Zero()
		} else if hashNums[i] != hashNums[i-1] {
			index[i] = field.Zero()
		} else if toHashInt[i] == 0 {
			index[i] = index[i-1]
		} else {
			index[i].Add(&index[i-1], new(field.Element).SetOne())
		}

		toHash[i] = field.NewElement(uint64(toHashInt[i]))
		hashNum[i] = field.NewElement(uint64(hashNums[i]))
		numBytesInt, numBytesF := randNBytes(rng)
		nBytes[i] = numBytesF

		limbValues := randLimbs(rng, numBytesInt)
		for j := range limbs {
			limbs[j][i] = limbValues[j]
		}
	}

	header := "TO_HASH,HASH_NUM,INDEX,NBYTES"
	for i := 0; i < common.NbLimbU128; i++ {
		header += fmt.Sprintf(",LIMB_%d", i)
	}
	fmt.Fprintln(oF, header)

	for i := range hashNums {
		row := fmt.Sprintf("%v,%v,%v,%v",
			toHash[i].String(),
			hashNum[i].String(),
			index[i].String(),
			nBytes[i].String(),
		)

		for j := 0; j < common.NbLimbU128; j++ {
			row += fmt.Sprintf(",0x%v", formatFieldAsNBitHex(limbs[j][i], 16))
		}

		fmt.Fprintln(oF, row)
	}

	oF.Close()
}

func randNBytes(rng *rand.Rand) (int, field.Element) {

	// nBytesInt must be in 1..=16
	var (
		nBytesInt = rng.Int32N(16) + 1
		nBytesF   = field.NewElement(uint64(nBytesInt))
	)

	return int(nBytesInt), nBytesF
}

func randLimbs(rng *rand.Rand, nBytes int) []field.Element {
	result := make([]field.Element, common.NbLimbU128)

	resBytes := make([]byte, 16)
	_, _ = utils.ReadPseudoRand(rng, resBytes[:nBytes])

	for i := 0; i < common.NbLimbU128; i++ {
		if i*bytesPerLimb >= nBytes {
			result[i] = field.Zero()
			continue
		}

		bytesToUse := bytesPerLimb
		if (i+1)*bytesPerLimb > nBytes {
			bytesToUse = nBytes - i*bytesPerLimb
		}

		limbBytes := make([]byte, 16)

		copy(limbBytes, resBytes[i*bytesPerLimb:i*bytesPerLimb+bytesToUse])

		if bytesToUse >= 2 {
			limbBytes[1] &= 0xFF
		}

		result[i] = *new(field.Element).SetBytes(limbBytes)
	}

	return result
}

func formatFieldAsNBitHex(n field.Element, nbBits uint) string {
	hexStr := n.Text(16)

	for len(hexStr) < int(nbBits/4) {
		hexStr = "0" + hexStr
	}

	if len(hexStr) > 8 {
		hexStr = hexStr[:int(nbBits/4)]
	}

	return hexStr
}
