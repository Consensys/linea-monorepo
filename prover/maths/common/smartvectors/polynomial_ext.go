package smartvectors

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// AddExt two vectors representing polynomials in coefficient form.
// a and b may have different sizes
func PolyAddExt(a, b SmartVector) SmartVector {

	small, large := a, b
	if a.Len() > b.Len() {
		small, large = large, small
	}

	res := make([]fext.Element, large.Len())
	large.WriteInSliceExt(res)
	for i := 0; i < small.Len(); i++ {
		x := small.GetExt(i)
		res[i].Add(&res[i], &x)
	}

	return NewRegularExt(res)
}

func PolySubExt(a, b SmartVector) SmartVector {

	maxLen := utils.Max(a.Len(), b.Len())
	res := make([]fext.Element, maxLen)
	a.WriteInSliceExt(res[:a.Len()])

	for i := 0; i < b.Len(); i++ {
		bi := b.GetExt(i)
		res[i].Sub(&res[i], &bi)
	}

	return NewRegularExt(res)
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
func RuffiniQuoRemExt(p SmartVector, q fext.Element) (quo SmartVector, rem fext.Element) {

	// The case where "p" is zero is assumed to be impossible as every type of
	// smart-vector strongly forbid dealing with zero length smart-vectors.
	if p.Len() == 0 {
		panic("Zero-length smart-vectors are forbidden")
	}

	// If p has length 1, then the general case algorithm does not work
	if p.Len() == 1 {
		quo = NewConstantExt(fext.Zero(), 1)
		rem = p.GetExt(0)
		return quo, rem
	}

	quo_ := make([]fext.Element, p.Len())

	// Pass the last coefficient
	quo_[p.Len()-1] = p.GetExt(p.Len() - 1)

	for i := p.Len() - 2; i >= 0; i-- {
		var c fext.Element
		c.Mul(&quo_[i+1], &q)
		pi := p.GetExt(i)
		quo_[i].Add(&c, &pi)
	}

	// As we employ custom allocation, we should not pass x[1:]
	rem = quo_[0]
	quo = NewRegularExt(quo_[1:])

	return quo, rem
}

// Evaluate a polynomial in Lagrange basis
func EvaluateLagrangeFullFext(v SmartVector, x fext.Element, oncoset ...bool) fext.Element {
	if con, ok := v.(*ConstantExt); ok {
		return con.Value
	}

	// Maybe there is an optim for windowed here
	res := make([]fext.Element, v.Len())
	v.WriteInSliceExt(res)
	return fastpolyext.EvaluateLagrange(res, x, oncoset...)
}

// Batch-evaluate polynomials in Lagrange basis
func BatchEvaluateLagrangeExt(vs []SmartVector, x fext.Element, oncoset ...bool) []fext.Element {

	var (
		polys         = make([][]fext.Element, len(vs))
		results       = make([]fext.Element, len(vs))
		computed      = make([]bool, len(vs))
		totalConstant = 0
	)

	// smartvector to []fr.element
	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {
			if con, ok := vs[i].(*ConstantExt); ok {
				// constant vectors
				results[i] = con.Value
				computed[i] = true
				totalConstant++
				continue
			}

			// non-constant vectors
			polys[i] = vs[i].IntoRegVecSaveAllocExt()
		}
	})

	if totalConstant == len(vs) {
		return results
	}

	return BatchEvaluateLagrangeSVExt(results, computed, polys, x, oncoset...)
}

// Optimized batch EvaluateLagrange for smart vectors.
// This reduces the number of computation by pre-processing
// constant vectors in advance in BatchEvaluateLagrange()
func BatchEvaluateLagrangeSVExt(results []fext.Element, computed []bool, polys [][]fext.Element, x fext.Element, oncoset ...bool) []fext.Element {

	n := 0
	for i := range polys {
		if len(polys[i]) > 0 {
			n = len(polys[i])
		}
	}

	if n == 0 {
		// that's a possible edge-case and it can happen if all the input polys
		// are constant smart-vectors. This is should be prevented by the the
		// caller.
		return results
	}

	if !utils.IsPowerOfTwo(n) {
		utils.Panic("only support powers of two but poly has length %v", len(polys))
	}

	domain := fft.GetDomainFromCache(uint64(n))
	denominator := make([]fext.Element, n)

	one := fext.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.MulByElement(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].MulByElement(&denominator[i-1], &domain.GeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)

		if denominator[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly

			for k := range polys {
				if computed[k] {
					continue
				}
				results[k] = polys[k][i]
			}

			return results
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = fext.BatchInvert(denominator)

	// Precompute the value of x^n once outside the loop
	xN := new(fext.Element).Exp(x, big.NewInt(int64(n)))

	// Precompute the value of domain.CardinalityInv outside the loop
	cardinalityInv := &domain.CardinalityInv

	// Compute factor as (x^n - 1) * (1 / domain.Cardinality).
	factor := new(fext.Element).Sub(xN, &one)
	factor.MulByElement(factor, cardinalityInv)

	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {

			if computed[k] {
				continue
			}
			// Compute the scalar product.
			res := vectorext.ScalarProd(polys[k], denominator)

			// Multiply res with factor.
			res.Mul(&res, factor)

			// Store the result.
			results[k] = res
		}
	})

	return results
}

// Evaluate a polynomial in coefficient basis
func EvalCoeffExt(v SmartVector, x fext.Element) fext.Element {
	// Maybe there is an optim for windowed here
	res := make([]fext.Element, v.Len())
	v.WriteInSliceExt(res)
	return polyext.Eval(res, x)
}

func EvalCoeffBivariateExt(v SmartVector, x fext.Element, numCoeffX int, y fext.Element) fext.Element {

	if v.Len()%numCoeffX != 0 {
		panic("size of v and nb coeff x are inconsistent")
	}

	// naive evaluation : we think it is not performance critical
	slice := make([]fext.Element, v.Len())
	v.WriteInSliceExt(slice)

	foldOnX := make([]fext.Element, len(slice)/numCoeffX)
	for i := 0; i < len(slice); i += numCoeffX {
		foldOnX[i/numCoeffX] = polyext.Eval(slice[i:i+numCoeffX], x)
	}

	return polyext.Eval(foldOnX, y)
}
