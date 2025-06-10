package poly

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func EvalGeneric[T any, fieldPointer field.FieldPointer[T]](pol []T, x T) T {
	var _res fieldPointer
	_res = &pol[len(pol)-1]
	for i := len(pol) - 2; i >= 0; i-- {
		_res.Mul(_res, &x)
		_res.Add(_res, &pol[i])
	}
	return *_res
}

// Eval returns ∑_i pol[i]xⁱ
func Eval(pol []field.Element, x field.Element) field.Element {
	var res field.Element
	for i := len(pol) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &pol[i])
	}
	return res
}

func EvalOnExtField(p []field.Element, x fext.Element) fext.Element {
	var res fext.Element

	randpolyext := make([]fext.Element, len(p))
	for i := 0; i < len(p); i++ {
		fext.FromBase(&randpolyext[i], &p[i])
	}
	for i := len(randpolyext) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &randpolyext[i])
	}

	return res
}

// TODO remove this, only used in /protocol/compiler/univariates/multi_to_single_point.go (otherwise used in tests only)
// Mul returns (∑_i a[i]xⁱ)*(∑_i b[i]xⁱ) using schoolbook algo
func Mul(a, b []field.Element) (res []field.Element) {

	if len(a) == 0 || len(b) == 0 {
		return []field.Element{}
	}

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

// Add adds two polynomials in coefficient form of possibly distinct degree.
// The returned slice has length = max(len(a), len(b)).
// The empty slice is understood as the zero polynomial and if both a and b are
// empty, the function returns the empty slice.
func Add(a, b []field.Element) (res []field.Element) {

	res = make([]field.Element, utils.Max(len(a), len(b)))
	copy(res, a)
	for i := range b {
		res[i].Add(&res[i], &b[i])
	}

	return res
}

// ScalarMul returns x*(∑_i p[i]xⁱ)
// TODO used only in /protocol/compiler/univariates/multi_to_single_point.go
func ScalarMul(p []field.Element, x field.Element) (res []field.Element) {
	res = make([]field.Element, len(p))
	vector.ScalarMul(res, p, x)
	return res
}

// EvaluateLagrangesAnyDomain evaluates all the Lagrange polynomials for a
// custom domain defined as the point point x. The function implements the naive
// schoolbook algorithm and is only relevant for small domains.
//
// The function panics if provided an empty domain.
// TODO used only in /protocol/compiler/univariates/multi_to_single_point.go, could be done using a simple fft
func EvaluateLagrangesAnyDomain(domain []field.Element, x field.Element) []field.Element {

	if len(domain) == 0 {
		utils.Panic("got provided an empty domain")
	}

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
