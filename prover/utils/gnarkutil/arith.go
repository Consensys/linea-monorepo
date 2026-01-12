package gnarkutil

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

func RepeatedVariable(x frontend.Variable, n int) []frontend.Variable {
	res := make([]frontend.Variable, n)
	for i := range res {
		res[i] = x
	}
	return res
}
func RepeatedVariableExt(x gnarkfext.Element, n int) []gnarkfext.Element {
	res := make([]gnarkfext.Element, n)
	for i := range res {
		res[i] = x
	}
	return res
}

// Exp in gnark circuit, using the fast exponentiation
func ExpExt(api frontend.API, x gnarkfext.Element, n int) gnarkfext.Element {

	if n < 0 {
		x.Inverse(api, x)
		n = -n
	}

	acc := x
	res := gnarkfext.One()

	// right-to-left
	for n != 0 {
		if n&1 == 1 {
			res.Mul(api, res, acc)
		}
		acc.Mul(api, acc, acc)
		n >>= 1
	}
	return res
}

// Exp in gnark circuit, using the fast exponentiation
func Exp(api frontend.API, x frontend.Variable, n int) frontend.Variable {
	if n < 0 {
		x = api.Inverse(x)
		n = -n
	}

	acc := x
	var res frontend.Variable = 1

	// right-to-left
	for n != 0 {
		if n&1 == 1 {
			res = api.Mul(res, acc)
		}
		acc = api.Mul(acc, acc)
		n >>= 1
	}
	return res
}

// ExpVariableExponent exponentiates x by n in a gnark circuit. Where n is not fixed.
// n is limited to n bits (max)
func ExpVariableExponent(api frontend.API, x frontend.Variable, exp frontend.Variable, expNumBits int) frontend.Variable {
	expBits := api.ToBinary(exp, expNumBits)
	var res frontend.Variable = 1

	for i := len(expBits) - 1; i >= 0; i-- {
		if i != len(expBits)-1 {
			res = api.Mul(res, res)
		}
		tmp := api.Mul(res, x)
		res = api.Select(expBits[i], tmp, res)
	}

	return res
}
