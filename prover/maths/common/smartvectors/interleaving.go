package smartvectors

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

func Interleave(vecs ...SmartVector) SmartVector {

	numvecs := len(vecs)

	if numvecs < 2 {
		panic("interleaving less than two vectors")
	}

	if !utils.IsPowerOfTwo(numvecs) {
		panic("expected a power of two number of polys")
	}

	res := make([]field.Element, numvecs*vecs[0].Len())

	for i := range vecs {
		for j := 0; j < vecs[0].Len(); j++ {
			res[j*numvecs+i] = vecs[i].Get(j)
		}
	}

	return NewRegular(res)
}
