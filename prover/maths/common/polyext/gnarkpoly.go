package polyext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"

	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

// EvaluateLagrangeAnyDomainGnark mirrors [EvaluateLagrangesAnyDomain] but in
// a gnark circuit. The same usage precautions applies for it.
func EvaluateLagrangeAnyDomainGnark(api frontend.API, domain []gnarkfext.Element, x gnarkfext.Element) []gnarkfext.Element {

	lagrange := make([]gnarkfext.Element, len(domain))

	for i, hi := range domain {
		var lhix gnarkfext.Element
		lhix.B0.A0 = field.One()
		for j, hj := range domain {
			if i == j {
				// Skip it
				continue
			}
			var factor, den gnarkfext.Element

			// more convenient to store -h instead of h
			factor.Sub(api, x, hj)
			den.Sub(api, hi, hj) // so x - h
			den.Inverse(api, den)

			// accumulate the product
			lhix.Mul(api, lhix, factor, den)
		}
		lagrange[i] = lhix
	}

	return lagrange

}

// EvaluateUnivariateGnark evaluate a univariate polynomial in a gnark circuit.
// It mirrors [Eval].
func EvaluateUnivariateGnark(api frontend.API, pol []gnarkfext.Element, x gnarkfext.Element) gnarkfext.Element {
	var res gnarkfext.Element
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(api, res, x)
		res.Add(api, res, pol[i])
	}
	return res
}
