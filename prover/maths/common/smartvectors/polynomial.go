package smartvectors

import (
	"sync/atomic"

	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"

	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Add two vectors representing polynomials in coefficient form.
// a and b may have different sizes
func PolyAdd(a, b SmartVector) SmartVector {

	small, large := a, b
	if a.Len() > b.Len() {
		small, large = large, small
	}

	res := make([]field.Element, large.Len())
	large.WriteInSlice(res)
	for i := 0; i < small.Len(); i++ {
		x := small.Get(i)
		res[i].Add(&res[i], &x)
	}

	return NewRegular(res)
}

func PolySub(a, b SmartVector) SmartVector {

	maxLen := utils.Max(a.Len(), b.Len())
	res := make([]field.Element, maxLen)
	a.WriteInSlice(res[:a.Len()])

	for i := 0; i < b.Len(); i++ {
		bi := b.Get(i)
		res[i].Sub(&res[i], &bi)
	}

	return NewRegular(res)
}

/*
Ruffini division
  - p polynomial in coefficient form
  - q field.Element, caracterizing the divisor X - q
  - quo quotient polynomial in coefficient form, result will be passed
    here. quo is truncated of its first entry in the process
  - expected to be at least as large as `p`

- rem, remainder also equals to P(r)

Supports &p == quo
*/
func RuffiniQuoRem(p SmartVector, q field.Element) (quo SmartVector, rem field.Element) {

	// The case where "p" is zero is assumed to be impossible as every type of
	// smart-vector strongly forbid dealing with zero length smart-vectors.
	if p.Len() == 0 {
		panic("Zero-length smart-vectors are forbidden")
	}

	// If p has length 1, then the general case algorithm does not work
	if p.Len() == 1 {
		quo = NewConstant(field.Zero(), 1)
		rem = p.Get(0)
		return quo, rem
	}

	quo_ := make([]field.Element, p.Len())

	// Pass the last coefficient
	quo_[p.Len()-1] = p.Get(p.Len() - 1)

	for i := p.Len() - 2; i >= 0; i-- {
		var c field.Element
		c.Mul(&quo_[i+1], &q)
		pi := p.Get(i)
		quo_[i].Add(&c, &pi)
	}

	// As we employ custom allocation, we should not pass x[1:]
	rem = quo_[0]
	quo = NewRegular(quo_[1:])

	return quo, rem
}

// EvaluateLagrange a polynomial in Lagrange basis at an field point
func EvaluateLagrange(v SmartVector, x field.Element, oncoset ...bool) field.Element {

	if !IsBase(v) {
		utils.Panic("Provided a non-base smart-vector")
	}

	switch con := v.(type) {
	case *Constant:
		return con.Value
	}

	// Maybe there is an optim for windowed here
	poly := make([]field.Element, v.Len())
	v.WriteInSlice(poly)
	// res := fastpoly.EvaluateLagrangeOnFext(poly, x, oncoset...)
	res := fastpoly.EvaluateLagrange(poly, x, oncoset...)
	return res
}

// EvaluateLagrangeMixed a polynomial in Lagrange basis at an E4 point
func EvaluateLagrangeMixed(v SmartVector, x fext.Element, oncoset ...bool) fext.Element {
	switch con := v.(type) {
	case *Constant:
		var res fext.Element
		fext.SetFromBase(&res, &con.Value)
		return res
	}

	// Maybe there is an optim for windowed here
	poly := make([]field.Element, v.Len())
	v.WriteInSlice(poly)
	res := fastpoly.EvaluateLagrangeMixed(poly, x, oncoset...)

	return res
}

// BatchEvaluateLagrangeMixed polynomials in Lagrange basis at an E4 point
func BatchEvaluateLagrangeMixed(vs []SmartVector, x fext.Element, oncoset ...bool) []fext.Element {

	var (
		polys         = make([][]field.Element, len(vs))
		results       = make([]fext.Element, len(vs))
		computed      = make([]bool, len(vs))
		totalConstant = uint64(0)
	)

	// filter out constant vectors
	indexNonConstantVector := -1
	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {

			if !IsBase(vs[i]) {
				utils.Panic("expected a base-field smart-vector, got %T", vs[i])
			}

			switch con := vs[i].(type) {
			case *Constant:
				fext.SetFromBase(&results[i], &con.Value)
				computed[i] = true
				atomic.AddUint64(&totalConstant, 1)
				continue
			}

			// non-constant vectors
			indexNonConstantVector = i
			polys[i], _ = vs[i].IntoRegVecSaveAllocBase()
		}
	})

	// all the vectors are constant, nothing to do more
	if int(totalConstant) == len(vs) {
		return results
	}

	// else, we put dummy copy at the constant vector indices, and call BatchEvaluateLagrange
	for i := 0; i < len(polys); i++ {
		if computed[i] {
			polys[i] = polys[indexNonConstantVector]
		}
	}

	// batch evaluate, and replace already computed values from constant vectors
	tmp := fastpoly.BatchEvaluateLagrangeMixed(polys, x, oncoset...)
	for i := 0; i < len(polys); i++ {
		if !computed[i] {
			results[i].Set(&tmp[i])
		}
	}

	return results
}

// Evaluate a polynomial in coefficient basis at an E4 point
func EvalCoeff(v SmartVector, x field.Element) field.Element {
	// Maybe there is an optim for windowed here
	res := make([]field.Element, v.Len())
	v.WriteInSlice(res)

	return poly.Eval(res, x)
}

// Evaluate a polynomial in coefficient basis at an E4 point
func EvalCoeffMixed(v SmartVector, x fext.Element) fext.Element {
	// Maybe there is an optim for windowed here
	res := make([]field.Element, v.Len())
	v.WriteInSlice(res)
	return poly.EvalMixed(res, x)
}

func EvalCoeffBivariate(v SmartVector, x field.Element, numCoeffX int, y field.Element) field.Element {

	if v.Len()%numCoeffX != 0 {
		panic("size of v and nb coeff x are inconsistent")
	}

	// naive evaluation : we think it is not performance critical
	slice := make([]field.Element, v.Len())
	v.WriteInSlice(slice)

	foldOnX := make([]field.Element, len(slice)/numCoeffX)
	for i := 0; i < len(slice); i += numCoeffX {
		foldOnX[i/numCoeffX] = poly.Eval(slice[i:i+numCoeffX], x)
	}

	return poly.Eval(foldOnX, y)
}
