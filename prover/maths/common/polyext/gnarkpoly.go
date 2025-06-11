package polyext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

// EvaluateLagrangeAnyDomainGnark mirrors [EvaluateLagrangesAnyDomain] but in
// a gnark circuit. The same usage precautions applies for it.
func EvaluateLagrangeAnyDomainGnark(api frontend.API, domain []gnarkfext.Element, x gnarkfext.Element) []gnarkfext.Element {

	outerAPI := gnarkfext.API{Inner: api}
	lagrange := make([]gnarkfext.Element, len(domain))

	for i, hi := range domain {
		lhix := gnarkfext.Element{field.One(), field.Zero()}
		for j, hj := range domain {
			if i == j {
				// Skip it
				continue
			}
			// more convenient to store -h instead of h
			factor := outerAPI.Sub(x, hj)
			den := outerAPI.Sub(hi, hj) // so x - h
			den = outerAPI.Inverse(den)

			// accumulate the product
			lhix = outerAPI.Mul(lhix, factor, den)
		}
		lagrange[i] = lhix
	}

	return lagrange

}

// EvaluateUnivariateGnark evaluate a univariate polynomial in a gnark circuit.
// It mirrors [EvalUnivariate].
func EvaluateUnivariateGnark(api frontend.API, pol []gnarkfext.Element, x gnarkfext.Element) gnarkfext.Element {
	res := gnarkfext.Element{frontend.Variable(0), frontend.Variable(0)}
	outerAPI := gnarkfext.API{Inner: api}
	for i := len(pol) - 1; i >= 0; i-- {
		res = outerAPI.Mul(res, x)
		res = outerAPI.Add(res, pol[i])
	}
	return res
}
