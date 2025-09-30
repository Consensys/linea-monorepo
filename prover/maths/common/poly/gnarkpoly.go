package poly

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// EvaluateLagrangeAnyDomainGnark mirrors [EvaluateLagrangesAnyDomain] but in
// a gnark circuit. The same usage precautions applies for it.
func EvaluateLagrangeAnyDomainGnark[T zk.Element](api frontend.API, domain []T, x T) []T {

	lagrange := make([]T, len(domain))

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	for i, hi := range domain {
		lhix := zk.ValueOf[T](1)
		for j, hj := range domain {
			if i == j {
				// Skip it
				continue
			}
			// more convenient to store -h instead of h
			factor := apiGen.Sub(&x, &hj)
			den := apiGen.Sub(&hi, &hj) // so x - h
			den = apiGen.Inverse(den)

			// accumulate the product
			lhix = apiGen.Mul(lhix, factor)
			lhix = apiGen.Mul(lhix, den)
		}
		lagrange[i] = *lhix
	}

	return lagrange

}

// EvaluateUnivariateGnarkMixed evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkMixed[T zk.Element](api frontend.API, pol []T, x gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {
	// var res gnarkfext.E4Gen[T]
	var z fext.Element
	res := gnarkfext.NewE4Gen[T](z)

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	for i := len(pol) - 1; i >= 0; i-- {
		res = *e4Api.Mul(&res, &x)
		res = *e4Api.AddByBase(&res, &pol[i])
	}
	return res
}

// EvaluateUnivariateGnarkExt evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkExt[T zk.Element](api frontend.API, pol []gnarkfext.E4Gen[T], x gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {
	var z fext.Element
	res := gnarkfext.NewE4Gen[T](z)

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	for i := len(pol) - 1; i >= 0; i-- {
		res = *e4Api.Mul(&res, &x)
		res = *e4Api.Add(&res, &pol[i])
	}
	return res
}
