package smartvectors

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// LinComb computes a linear combination of the given vectors with integer coefficients
func LinComb(coeffs []int, svecs []SmartVector) SmartVector {
	// Sanity check : all svec should have the same length
	length := svecs[0].Len()
	for i := 0; i < len(svecs); i++ {
		if svecs[i].Len() != length {
			utils.Panic("bad size %v, expected %v", svecs[i].Len(), length)
		}
	}
	return processOperator(linCombOp{}, coeffs, svecs)
}

// Product computes a product of smart-vectors with integer exponents
func Product(exponents []int, svecs []SmartVector) SmartVector {
	return processOperator(productOp{}, exponents, svecs)
}

// Computes the result of the linear combination and put the result into res
func processOperator(op operator, coeffs []int, svecs []SmartVector) SmartVector {

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
	constRes, matchedConst := processConstOnly(op, svecs, coeffs)

	// Full-constant operation, return the constant vec
	if matchedConst == totalToMatch {
		return constRes
	}

	// Special-case : if the operation is a product and the constRes is
	// zero, we can early return zero ignoring the rest.
	if _, ok := op.(productOp); ok && constRes != nil && constRes.val.IsZero() {
		return constRes
	}

	// Accumulate the windowed smart-vectors
	windowRes, matchedWindow := processWindowedOnly(op, svecs, coeffs)

	// Edge-case : the list of smart-vectors to combine is windowed-only
	if matchedWindow == totalToMatch {
		return windowRes
	}

	// If we had both matches for constants vectors, we merge the constant
	// into the window
	if matchedWindow > 0 && matchedConst > 0 {
		switch w := windowRes.(type) {
		case *PaddedCircularWindow:
			op.ConstTermIntoVec(w.window, &constRes.val)
			op.ConstTermIntoConst(&w.paddingVal, &constRes.val)
		case *Regular:
			op.ConstTermIntoVec(*w, &constRes.val)
		}
	}

	// Edge-case : all vectors in the list are either window or constants
	if matchedWindow+matchedConst == totalToMatch {
		return windowRes
	}

	// Accumulate the regular part of the vector
	regularRes, matchedRegular := processRegularOnly(op, svecs, coeffs)

	// Sanity-check : all of the vector should fall into only one of the two
	// category.
	if matchedConst+matchedWindow+matchedRegular != totalToMatch {
		utils.Panic("Mismatch between the number of matched vector and the total number of vectors (%v + %v + %v = %v)", matchedConst, matchedWindow, matchedRegular, totalToMatch)
	}

	switch {
	case matchedRegular == totalToMatch:
		// Do nuffing
		return regularRes
	case matchedRegular+matchedConst == totalToMatch:
		// In this case, there are no windowed in the list. This means we only
		// need to merge the const one into the regular one before returning
		op.ConstTermIntoVec(*regularRes, &constRes.val)
		return regularRes
	default:

		// If windowRes is a regular (can happen if all windows arguments cover the full circle)
		if w, ok := windowRes.(*Regular); ok {
			op.VecTermIntoVec(*regularRes, *w)
			return regularRes
		}

		// Overwrite window with its casting into an actual circular windows
		windowRes := windowRes.(*PaddedCircularWindow)

		// In this case, the constant is already accumulated into the windowed.
		// Thus, we just have to merge the windowed one into the regular one.
		interval := windowRes.interval()
		regvec := *regularRes
		length := len(regvec)

		// The windows rolls over
		if interval.doesWrapAround() {
			op.VecTermIntoVec(regvec[:interval.stop()], windowRes.window[length-interval.start():])
			op.VecTermIntoVec(regvec[interval.start():], windowRes.window[:length-interval.start()])
			op.ConstTermIntoVec(regvec[interval.stop():interval.start()], &windowRes.paddingVal)
			return regularRes
		}

		// Else, no roll-over
		op.VecTermIntoVec(regvec[interval.start():interval.stop()], windowRes.window)
		op.ConstTermIntoVec(regvec[:interval.start()], &windowRes.paddingVal)
		op.ConstTermIntoVec(regvec[interval.stop():], &windowRes.paddingVal)
		return regularRes
	}
}

// Returns the result of the linear combination including only the constant. numMatches denotes
// the number of Constant smart-vectors found in the list of arguments.
func processConstOnly(op operator, svecs []SmartVector, coeffs []int) (constRes *Constant, numMatches int) {
	var constVal field.Element
	for i, svec := range svecs {
		if cnst, ok := svec.(*Constant); ok {
			if numMatches < 1 {
				// First one, no need to add it into constVal since constVal is zero
				op.ConstIntoTerm(&constVal, &cnst.val, coeffs[i])
				numMatches++
				continue
			}
			op.ConstIntoConst(&constVal, &cnst.val, coeffs[i])
			numMatches++
		}
	}

	if numMatches == 0 {
		return nil, 0
	}

	return &Constant{val: constVal, length: svecs[0].Len()}, numMatches
}
