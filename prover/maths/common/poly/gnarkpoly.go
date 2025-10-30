package poly

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// EvaluateUnivariateGnarkMixed evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkMixed(api frontend.API, pol []zk.WrappedVariable, x gnarkfext.E4Gen) gnarkfext.E4Gen {

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := gnarkfext.NewE4GenFromBase(0)
	for i := len(pol) - 1; i >= 0; i-- {
		res = *e4Api.Mul(&res, &x)
		res = *e4Api.AddByBase(&res, pol[i])
	}
	return res
}

// EvaluateUnivariateGnarkExt evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkExt(api frontend.API, pol []gnarkfext.E4Gen, x gnarkfext.E4Gen) gnarkfext.E4Gen {

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := gnarkfext.NewE4GenFromBase(0)
	for i := len(pol) - 1; i >= 0; i-- {
		res = *e4Api.Mul(&res, &x)
		res = *e4Api.Add(&res, &pol[i])
	}
	return res
}
