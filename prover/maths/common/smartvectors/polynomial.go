package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
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

// EvaluateLagrange a polynomial in Lagrange basis
func EvaluateLagrange(v SmartVector, x fext.Element, oncoset ...bool) field.Element {
	switch con := v.(type) {
	case *Constant:
		return con.val
	}

	// Maybe there is an optim for windowed here
	res := make([]fext.Element, v.Len())
	v.WriteInSliceExt(res)
	var xExt fext.Element
	// fext.FromBase(&xExt, &x)
	resExt := fastpolyext.EvaluateLagrange(res, xExt, oncoset...)
	return resExt.B0.A0
}

// Batch-evaluate polynomials in Lagrange basis
func BatchInterpolate(vs []SmartVector, x field.Element, oncoset ...bool) []field.Element {

	var (
		polys    = make([][]field.Element, len(vs))
		results  = make([]field.Element, len(vs))
		computed = make([]bool, len(vs))
	)

	// filter out constant vectors
	indexNonConstantVector := -1
	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {
			switch con := vs[i].(type) {
			case *Constant:
				results[i] = con.val
				computed[i] = true
				continue
			}

			// non-constant vectors
			indexNonConstantVector = i
			polys[i], _ = vs[i].IntoRegVecSaveAllocBase()
		}
	})

	// all the vectors are constant, nothing to do more
	if indexNonConstantVector == -1 {
		return results
	}

	// else, we put dummy copy at the constant vector indices, and call BatchEvaluateLagrange
	for i := 0; i < len(polys); i++ {
		if computed[i] {
			polys[i] = polys[indexNonConstantVector]
		}
	}

	// batch evaluate, and replace already computed values from constant vectors
	tmp := fastpoly.BatchEvaluateLagrange(polys, x, oncoset...)
	for i := 0; i < len(polys); i++ {
		if !computed[i] {
			results[i].Set(&tmp[i])
		}
	}

	return results
}

// Evaluate a polynomial in coefficient basis
func EvalCoeff(v SmartVector, x field.Element) field.Element {
	// Maybe there is an optim for windowed here
	res := make([]field.Element, v.Len())
	v.WriteInSlice(res)
	return poly.Eval(res, x)
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
