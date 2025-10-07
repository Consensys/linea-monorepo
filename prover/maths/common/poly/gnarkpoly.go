package poly

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// EvaluateLagrangeAnyDomainGnark mirrors [EvaluateLagrangesAnyDomain] but in
// a gnark circuit. The same usage precautions applies for it.
func EvaluateLagrangeAnyDomainGnark(api frontend.API, domain []zk.WrappedVariable, x zk.WrappedVariable) []zk.WrappedVariable {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	lagrange := make([]zk.WrappedVariable, len(domain))

	for i, hi := range domain {
		lhix := zk.ValueOf(1)
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
			lhix = *apiGen.Mul(&lhix, factor)
			lhix = *apiGen.Mul(&lhix, den)
		}
		lagrange[i] = lhix
	}

	return lagrange

}

// EvaluateUnivariateGnarkMixed evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkMixed(api frontend.API, pol []zk.WrappedVariable, x gnarkfext.E4Gen) gnarkfext.E4Gen {
	var res gnarkfext.E4Gen
	res.SetZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(api, res, x)
		res.AddByBase(api, res, pol[i])
	}
	return res
}

// EvaluateUnivariateGnarkExt evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkExt(api frontend.API, pol []gnarkfext.E4Gen, x gnarkfext.E4Gen) gnarkfext.E4Gen {
	var res gnarkfext.E4Gen
	res.SetZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(api, res, x)
		res.Add(api, res, pol[i])
	}
	return res
}
