package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LinCombExt computes a linear combination of the given vectors with integer coefficients.
//   - The function panics if provided SmartVector of different lengths
//   - The function panics if svecs is empty
//   - The function panics if the length of coeffs does not match the length of
//     svecs
func LinCombExt(coeffs []int, svecs []SmartVector, p ...mempool.MemPool) SmartVector {
	// Sanity check : all svec should have the same length
	length := svecs[0].Len()
	for i := 0; i < len(svecs); i++ {
		if svecs[i].Len() != length {
			utils.Panic("bad size %v, expected %v", svecs[i].Len(), length)
		}
	}
	return processOperatorExt(linCombOpExt{}, coeffs, svecs, p...)
}

// ProductExt computes a product of smart-vectors with integer exponents
//   - The function panics if provided SmartVector of different lengths
//   - The function panics if svecs is empty
//   - The function panics if the length of exponents does not match the length of
//     svecs
func ProductExt(exponents []int, svecs []SmartVector, p ...mempool.MemPool) SmartVector {
	return processOperatorExt(productOpExt{}, exponents, svecs, p...)
}

// processOperatorExt computes the result of an [operator] and put the result into res
//   - The function panics if provided SmartVector of different lengths
//   - The function panics if svecs is empty
//   - The function panics if the length of coeffs does not match the length of
//     svecs
func processOperatorExt(op operatorExt, coeffs []int, svecs []SmartVector, p ...mempool.MemPool) SmartVector {

	// There should be as many coeffs than there are vectors
	if len(coeffs) != len(svecs) {
		utils.Panic("there are %v coeffs and %v vectors", len(coeffs), len(svecs))
	}

	// Sanity-check to ensure there is at least one vector to lincombine
	if len(svecs) == 0 {
		utils.Panic("no vector to process")
	}

	// Total number of vector passed as operands*
	totalToMatch := len(svecs)

	// Sanity-check, they should all have the same length
	for i := range svecs {
		assertHasLength(svecs[0].Len(), svecs[i].Len())
	}

	// Sanity-check, length zero or negative should be forbidden
	assertStrictPositiveLen(svecs[0].Len())

	// Accumulate the constant
	constRes, matchedConst := processConstOnlyExt(op, svecs, coeffs)

	// Full-constant operation, return the constant vec
	if matchedConst == totalToMatch {
		return constRes
	}

	// Special-case : if the operation is a product and the constRes is
	// zero, we can early return zero ignoring the rest.
	if _, ok := op.(productOpExt); ok && constRes != nil && constRes.val.IsZero() {
		return constRes
	}

	// Accumulate the windowed smart-vectors
	windowRes, matchedWindow := processWindowedOnlyExt(op, svecs, coeffs)

	// Edge-case : the list of smart-vectors to combine is windowed-only. In
	// this case we can return directly.
	if matchedWindow == totalToMatch {
		return windowRes
	}

	// If we had matches for both constants vectors and the windows, we merge
	// the constant into the window.
	if matchedWindow > 0 && matchedConst > 0 {
		switch w := windowRes.(type) {
		case *PaddedCircularWindowExt:
			op.constTermExtIntoVecExt(w.window, &constRes.val)
			op.constTermExtIntoConstExt(&w.paddingVal, &constRes.val)
		case *RegularExt:
			op.constTermExtIntoVecExt(*w, &constRes.val)
		}
	}

	// Edge-case : all vectors in the list are either window or constants
	if matchedWindow+matchedConst == totalToMatch {
		return windowRes
	}

	// Accumulate the regular part of the vector
	regularRes, matchedRegular := processRegularOnlyExt(op, svecs, coeffs, p...)

	// Sanity-check : all of the vector should fall into only one of the two
	// category.
	if matchedConst+matchedWindow+matchedRegular != totalToMatch {
		utils.Panic("Mismatch between the number of matched vector and the total number of vectors (%v + %v + %v = %v)", matchedConst, matchedWindow, matchedRegular, totalToMatch)
	}

	switch {
	case matchedRegular == totalToMatch:
		return regularRes
	case matchedRegular+matchedConst == totalToMatch:
		// In this case, there are no windowed in the list. This means we only
		// need to merge the const one into the regular one before returning
		op.constTermExtIntoVecExt(regularRes.RegularExt, &constRes.val)
		return regularRes
	default:

		// If windowRes is a regular (can happen if all windows arguments cover the full circle)
		if w, ok := windowRes.(*RegularExt); ok {
			op.vecTermExtIntoVecExt(regularRes.RegularExt, *w)
			return regularRes
		}

		// Overwrite window with its casting into an actual circular windows
		windowRes := windowRes.(*PaddedCircularWindowExt)

		// In this case, the constant is already accumulated into the windowed.
		// Thus, we just have to merge the windowed one into the regular one.
		interval := windowRes.interval()
		regvec := regularRes.RegularExt
		length := len(regvec)

		// The windows rolls over
		if interval.DoesWrapAround() {
			op.vecTermExtIntoVecExt(regvec[:interval.Stop()], windowRes.window[length-interval.Start():])
			op.vecTermExtIntoVecExt(regvec[interval.Start():], windowRes.window[:length-interval.Start()])
			op.constTermExtIntoVecExt(regvec[interval.Stop():interval.Start()], &windowRes.paddingVal)
			return regularRes
		}

		// Else, no roll-over
		op.vecTermExtIntoVecExt(regvec[interval.Start():interval.Stop()], windowRes.window)
		op.constTermExtIntoVecExt(regvec[:interval.Start()], &windowRes.paddingVal)
		op.constTermExtIntoVecExt(regvec[interval.Stop():], &windowRes.paddingVal)
		return regularRes
	}
}

// Returns the result of the linear combination including only the constant. numMatches denotes
// the number of Constant smart-vectors found in the list of arguments.
func processConstOnlyExt(op operatorExt, svecs []SmartVector, coeffs []int) (constRes *ConstantExt, numMatches int) {
	var constVal fext.Element
	for i, svec := range svecs {
		if cnst, ok := svec.(*ConstantExt); ok {
			if numMatches < 1 {
				// First one, no need to add it into constVal since constVal is zero
				op.constExtIntoTermExt(&constVal, &cnst.val, coeffs[i])
				numMatches++
				continue
			}
			op.constExtIntoConstExt(&constVal, &cnst.val, coeffs[i])
			numMatches++
		}
	}

	if numMatches == 0 {
		return nil, 0
	}

	return &ConstantExt{val: constVal, length: svecs[0].Len()}, numMatches
}
