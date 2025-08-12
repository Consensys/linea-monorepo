package poly

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

// EvaluateLagrangeAnyDomainGnark mirrors [EvaluateLagrangesAnyDomain] but in
// a gnark circuit. The same usage precautions applies for it.
func EvaluateLagrangeAnyDomainGnark(api frontend.API, domain []frontend.Variable, x frontend.Variable) []frontend.Variable {

	lagrange := make([]frontend.Variable, len(domain))

	for i, hi := range domain {
		lhix := frontend.Variable(field.One())
		for j, hj := range domain {
			if i == j {
				// Skip it
				continue
			}
			// more convenient to store -h instead of h
			factor := api.Sub(x, hj)
			den := api.Sub(hi, hj) // so x - h
			den = api.Inverse(den)

			// accumulate the product
			lhix = api.Mul(lhix, factor, den)
		}
		lagrange[i] = lhix
	}

	return lagrange

}

// EvaluateUnivariateGnarkMixed evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkMixed(api frontend.API, pol []frontend.Variable, x gnarkfext.Element) gnarkfext.Element {
	var res gnarkfext.Element
	res.SetZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(api, res, x)
		res.AddByBase(api, res, pol[i])
	}
	return res
}

// EvaluateUnivariateGnarkExt evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnarkExt(api frontend.API, pol []gnarkfext.Element, x gnarkfext.Element) gnarkfext.Element {
	var res gnarkfext.Element
	res.SetZero()
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(api, res, x)
		res.Add(api, res, pol[i])
	}
	return res
}
