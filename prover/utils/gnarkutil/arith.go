package gnarkutil

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

func RepeatedVariable(x zk.WrappedVariable, n int) []zk.WrappedVariable {
	res := make([]zk.WrappedVariable, n)
	for i := range res {
		res[i] = x
	}
	return res
}

// Exp in gnark circuit, using the fast exponentiation
func ExpExt(api frontend.API, x gnarkfext.E4Gen, n int) gnarkfext.E4Gen {

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	if n < 0 {
		x = *e4Api.Inverse(&x)
		n = -n
	}

	acc := x
	res := gnarkfext.NewE4GenFromBase(1)

	// right-to-left
	for n != 0 {
		if n&1 == 1 {
			res = *e4Api.Mul(&res, &acc)
		}
		acc = *e4Api.Mul(&acc, &acc)
		n >>= 1
	}
	return res
}

// Exp in gnark circuit, using the fast exponentiation
func Exp(api frontend.API, x zk.WrappedVariable, n int) zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	if n < 0 {
		x = apiGen.Inverse(x)
		n = -n
	}

	acc := x
	res := zk.ValueOf(1)

	// right-to-left
	for n != 0 {
		if n&1 == 1 {
			res = apiGen.Mul(res, acc)
		}
		acc = apiGen.Mul(acc, acc)
		n >>= 1
	}
	return res
}

// ExpVariableExponent exponentiates x by n in a gnark circuit. Where n is not fixed.
// n is limited to n bits (max)
func ExpVariableExponent(api frontend.API, x zk.WrappedVariable, exp zk.WrappedVariable, expNumBits int) zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	expBits := api.ToBinary(exp, expNumBits)
	res := zk.ValueOf(1)

	for i := len(expBits) - 1; i >= 0; i-- {
		if i != len(expBits)-1 {
			res = apiGen.Mul(res, res)
		}
		tmp := apiGen.Mul(res, x)
		res = apiGen.Select(expBits[i], tmp, res)
	}

	return res
}
