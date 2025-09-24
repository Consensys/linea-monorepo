package polyext

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func EvalUnivariateMixed(pol []fext.GenericFieldElem, x fext.GenericFieldElem) fext.GenericFieldElem {
	res := fext.GenericFieldZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(&x)
		res.Add(&pol[i])
	}
	return res
}

func MulByElement(a []fext.Element, b []field.Element) (res []fext.Element) {

	if len(a) == 0 || len(b) == 0 {
		return []fext.Element{}
	}

	res = make([]fext.Element, len(a)+len(b)-1)

	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			var tmp fext.Element
			tmp.MulByElement(&a[i], &b[j])
			res[i+j].Add(&res[i+j], &tmp)
		}
	}

	return res
}

func Add(a, b []fext.Element) (res []fext.Element) {

	res = make([]fext.Element, utils.Max(len(a), len(b)))
	copy(res, a)
	for i := range b {
		res[i].Add(&res[i], &b[i])
	}

	return res
}
