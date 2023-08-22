package poly

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark/frontend"
)

/*
Mirrors EvaluateLagrangesAnyDomain in a gnark circuit
*/
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

// Evaluate a univariate polynomial in a gnark circuit
func EvaluateUnivariateGnark(api frontend.API, pol []frontend.Variable, x frontend.Variable) frontend.Variable {
	res := frontend.Variable(0)
	for i := len(pol) - 1; i >= 0; i-- {
		res = api.Mul(res, x)
		res = api.Add(res, pol[i])
	}
	return res
}

// Evaluate a bivariate polynomial in a gnark circuit
func GnarkEvalCoeffBivariate(api frontend.API, v []frontend.Variable, x frontend.Variable, numCoeffX int, y frontend.Variable) frontend.Variable {

	if len(v)%numCoeffX != 0 {
		panic("size of v and nb coeff x are inconsistent")
	}

	foldOnX := make([]frontend.Variable, len(v)/numCoeffX)
	for i := 0; i < len(v); i += numCoeffX {
		foldOnX[i/numCoeffX] = EvaluateUnivariateGnark(api, v[i:i+numCoeffX], x)
	}

	return EvaluateUnivariateGnark(api, foldOnX, y)
}
