package gnarkutil

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

func RepeatedVariable(x T, n int) []T {
	res := make([]T, n)
	for i := range res {
		res[i] = x
	}
	return res
}

func RepeatedVariableGen[T zk.Element](x T, n int) []T {
	res := make([]T, n)
	for i := range res {
		res[i] = x
	}
	return res
}

/*
exponentiation in gnark circuit, using the fast exponentiation
*/
func Exp(api frontend.API, x T, n int) T {

	if n < 0 {
		x = api.Inverse(x)
		return Exp(api, x, -n)
	}

	if n == 0 {
		return T(1)
	}

	if n == 1 {
		return x
	}

	if n%2 == 0 {
		x2 := api.Mul(x, x)
		return Exp(api, x2, n/2)
	}

	if n%2 == 1 {
		x2 := api.Mul(x, x)
		res := Exp(api, x2, (n-1)/2)
		return api.Mul(res, x)
	}

	panic("unreachable")
}

// ExpVariableExponent exponentiates x by n in a gnark circuit. Where n is not fixed.
// n is limited to n bits (max)
func ExpVariableExponent(api frontend.API, x T, exp T, expNumBits int) T {

	expBits := api.ToBinary(exp, expNumBits)
	res := T(1)

	for i := len(expBits) - 1; i >= 0; i-- {

		if i != len(expBits)-1 {
			res = api.Mul(res, res)
		}

		res = api.Select(expBits[i], api.Mul(res, x), res)
	}

	return res
}
