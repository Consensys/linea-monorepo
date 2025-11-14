package encoding

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

func Encode8KoalabearToFrElement(elements [8]field.Element) fr.Element {
	var res fr.Element
	var bres, part big.Int
	for i := 0; i < 8; i++ {
		part.SetInt64(int64(elements[7-i].Bits()[0]))
		shift := uint(31 * i)  // Shift based on little-endian order
		part.Lsh(&part, shift) // Shift left by the appropriate position for little-endian
		bres.Add(&bres, &part) // Bitwise OR to combine
	}
	res.SetBigInt(&bres)
	return res
}

func EncodeKoalabearsToFrElement(elements []field.Element) []fr.Element {
	var res []fr.Element
	for len(elements) != 0 {
		var buf [8]field.Element
		// in this case we left pad by zeroes
		if len(elements) < 8 {
			copy(buf[8-len(elements):], elements[:])
			elements = elements[:0]
		} else {
			copy(buf[:], elements[:8])
			elements = elements[8:]
		}
		res = append(res, Encode8KoalabearToFrElement(buf))
	}
	return res
}
