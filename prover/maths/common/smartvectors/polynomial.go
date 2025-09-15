package smartvectors

import (
	"sync/atomic"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

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

// EvaluateBasePolyLagrange a polynomial in Lagrange basis at an E4 point
func EvaluateBasePolyLagrange(v SmartVector, x fext.Element, oncoset ...bool) fext.Element {
	if con, ok := v.(*Constant); ok {
		var res fext.Element
		fext.SetFromBase(&res, &con.Value)
		return res
	}

	// Maybe there is an optim for windowed here
	poly := make([]field.Element, v.Len())
	v.WriteInSlice(poly)

	if len(oncoset) > 0 && oncoset[0] {
		genFr := fft.GeneratorFullMultiplicativeGroup()
		genFr.Inverse(&genFr)
		x.MulByElement(&x, &genFr)
	}

	res, _ := vortex.EvalBasePolyLagrange(poly, x)

	return res
}

// BatchEvaluateBasePolyLagrange polynomials in Lagrange basis at an E4 point
func BatchEvaluateBasePolyLagrange(vs []SmartVector, x fext.Element, oncoset ...bool) []fext.Element {
	results := make([]fext.Element, len(vs))

	if len(vs) == 0 {
		return results
	}

	// Separate constants from polynomials
	type workItem struct {
		index      int
		poly       []field.Element
		isConstant bool
		value      fext.Element
	}

	workItems := make([]workItem, len(vs))
	var totalConstant uint64

	// Process in parallel to identify constants and extract polynomials
	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {
			if !IsBase(vs[i]) {
				utils.Panic("expected a base-field smart-vector, got %T", vs[i])
			}

			if con, ok := vs[i].(*Constant); ok {
				var result fext.Element
				fext.SetFromBase(&result, &con.Value)
				workItems[i] = workItem{
					index:      i,
					isConstant: true,
					value:      result,
				}
				atomic.AddUint64(&totalConstant, 1)
			} else {
				poly, _ := vs[i].IntoRegVecSaveAllocBase()
				workItems[i] = workItem{
					index:      i,
					poly:       poly,
					isConstant: false,
				}
			}
		}
	})

	// Early return if all constants
	if int(totalConstant) == len(vs) {
		for _, item := range workItems {
			results[item.index] = item.value
		}
		return results
	}

	// Collect only non-constant polynomials
	nonConstantPolys := make([][]field.Element, 0, len(vs)-int(totalConstant))
	nonConstantIndices := make([]int, 0, len(vs)-int(totalConstant))

	for _, item := range workItems {
		if item.isConstant {
			results[item.index] = item.value
		} else {
			nonConstantPolys = append(nonConstantPolys, item.poly)
			nonConstantIndices = append(nonConstantIndices, item.index)
		}
	}

	// Batch evaluate only non-constant polynomials
	if len(nonConstantPolys) > 0 {
		polyResults, err := vortex.BatchEvalBasePolyLagrange(nonConstantPolys, x, oncoset...)
		if err != nil {
			panic(err)
		}

		// Map results back to original positions
		for i, result := range polyResults {
			results[nonConstantIndices[i]] = result
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
func EvalBasePolyHorner(v SmartVector, x fext.Element) fext.Element {
	// Maybe there is an optim for windowed here
	res := make([]field.Element, v.Len())
	v.WriteInSlice(res)
	return vortex.EvalBasePolyHorner(res, x)
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
