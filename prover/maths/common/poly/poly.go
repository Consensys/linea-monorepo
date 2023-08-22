package poly

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Evaluate the polynomial, given a vector of coefficients
// Coefficients are for increasing degree monomials
// The computation is done using the Horner method
func EvalUnivariate(pol []field.Element, x field.Element) field.Element {
	var res field.Element
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &pol[i])
	}
	return res
}

// Multiply two polynomials expressed by their coefficients using the naive method
// And write the result in res
func Mul(a, b []field.Element) (res []field.Element) {
	res = make([]field.Element, len(a)+len(b)-1)

	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			var tmp field.Element
			tmp.Mul(&a[i], &b[j])
			res[i+j].Add(&res[i+j], &tmp)
		}
	}

	return res
}

/*
Polynomial addition,
*/
func Add(a, b []field.Element) (res []field.Element) {

	res = make([]field.Element, utils.Max(len(a), len(b)))
	copy(res, a)
	for i := range b {
		res[i].Add(&res[i], &b[i])
	}

	return res
}

/*
Polynomial substraction,
*/
func Sub(a, b []field.Element) (res []field.Element) {
	res = make([]field.Element, utils.Max(len(a), len(b)))
	copy(res, a)
	for i := range b {
		res[i].Sub(&res[i], &b[i])
	}

	return res
}

/*
Multiplication of a polynomial by a scalar;
(not in-place)
*/
func ScalarMul(p []field.Element, x field.Element) (res []field.Element) {
	res = make([]field.Element, len(p))
	vector.ScalarMul(res, p, x)
	return res
}

// Evaluates the lagrange polynomials for a given domain in r
// The result maps the position of h \in domain, to L_domain(x)
func EvaluateLagrangesAnyDomain(domain []field.Element, x field.Element) []field.Element {

	lagrange := make([]field.Element, len(domain))

	for i := range domain {
		// allocate outside of the loop to avoid memory aliasing in for loop
		// (gosec G601)
		hi := domain[i]

		lhix := field.One()
		for j := range domain {
			// allocate outside of the loop to avoid memory aliasing in for loop
			// (gosec G601)
			hj := domain[j]

			if i == j {
				// Skip it
				continue
			}

			// Otherwise, it would divide by zeri
			if hi == hj {
				utils.Panic("the domain contained a duplicate %v (at %v and %v)", hi.String(), i, j)
			}

			// more convenient to store -h instead of h
			hj.Neg(&hj)
			factor := x
			factor.Add(&factor, &hj)
			hj.Add(&hi, &hj) // so x - h
			hj.Inverse(&hj)
			factor.Mul(&factor, &hj)

			lhix.Mul(&lhix, &factor)
		}
		lagrange[i] = lhix
	}

	return lagrange
}
