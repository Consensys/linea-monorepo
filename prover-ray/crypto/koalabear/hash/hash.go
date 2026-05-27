package hash

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	ext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

const (
	WIDTH              = 16
	SPONGE_WIDTH       = 24
	SPONGE_RATE        = 16
	NB_FULL_ROUND      = 6
	NB_PARTIAL_ROUNDS  = 21
	DIGEST_NB_ELEMENTS = 8
)

// 8 to land in a space big enough to be collision resistant
type Digest [DIGEST_NB_ELEMENTS]koalabear.Element

const StringChunkSize = 3

type FieldHasher interface {
	Reset()
	WriteElements(...koalabear.Element)
	WriteExt(...ext.E4)
	Sum() Digest
}

func NewElement(v uint64) koalabear.Element {
	var e koalabear.Element
	e.SetUint64(v)
	return e
}

func StringToElements(domainTag uint64, s string) []koalabear.Element {
	nbChunks := (len(s) + StringChunkSize - 1) / StringChunkSize
	res := make([]koalabear.Element, 2, 2+nbChunks)
	res[0].SetUint64(domainTag)
	res[1].SetUint64(uint64(len(s)))

	for i := 0; i < len(s); i += StringChunkSize {
		var limb uint64
		for j := 0; j < StringChunkSize && i+j < len(s); j++ {
			limb |= uint64(s[i+j]) << (8 * j)
		}
		var e koalabear.Element
		e.SetUint64(limb)
		res = append(res, e)
	}

	return res
}

func OutputToExt(out Digest) ext.E4 {
	return ElementsToExt(out[0], out[1], out[2], out[3])
}

func ElementsToExt(a0, a1, b0, b1 koalabear.Element) ext.E4 {
	var res ext.E4
	res.B0.A0.Set(&a0)
	res.B0.A1.Set(&a1)
	res.B1.A0.Set(&b0)
	res.B1.A1.Set(&b1)
	return res
}
