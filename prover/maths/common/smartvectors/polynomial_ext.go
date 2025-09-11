package smartvectors

import (
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/polyext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
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
		results       = make([]fext.Element, len(vs))
		totalConstant = 0
	)

	// First pass: handle constants and collect non-constant polynomials
	var nonConstantPolys [][]fext.Element
	var nonConstantIndices []int

	// Process smartvectors to separate constants from polynomials
	for i := 0; i < len(vs); i++ {
		switch con := vs[i].(type) {
		case *ConstantExt:
			// constant vectors - store result directly
			results[i] = con.Value
			totalConstant++
		default:
			// non-constant vectors - collect for batch processing
			poly := vs[i].IntoRegVecSaveAllocExt()
			nonConstantPolys = append(nonConstantPolys, poly)
			nonConstantIndices = append(nonConstantIndices, i)
		}
	}

	// If all vectors are constant, return early
	if totalConstant == len(vs) {
		return results
	}

	// Batch evaluate non-constant polynomials
	if len(nonConstantPolys) > 0 {
		nonConstantResults, _ := vortex.BatchEvalFextPolyLagrange(nonConstantPolys, x, oncoset...)

		// Map results back to original positions
		for j, result := range nonConstantResults {
			originalIndex := nonConstantIndices[j]
			results[originalIndex] = result
		}
	}

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
