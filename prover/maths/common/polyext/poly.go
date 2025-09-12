package polyext

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func EvalUnivariateMixed(pol []fext.GenericFieldElem, x fext.GenericFieldElem) fext.GenericFieldElem {
	res := fext.GenericFieldZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(&x)
		res.Add(&pol[i])
	}
	return res
}
