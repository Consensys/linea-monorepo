package smartvectors

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

func Add(vecs ...SmartVector) SmartVector {

	coeffs := make([]int, len(vecs))
	for i := range coeffs {
		coeffs[i] = 1
	}

	return LinComb(coeffs, vecs)
}

func Sub(vecs ...SmartVector) SmartVector {
	coeffs := make([]int, len(vecs))
	for i := range coeffs {
		if i == 0 {
			coeffs[i] = 1
		} else {
			coeffs[i] = -1
		}
	}

	return LinComb(coeffs, vecs)
}

func Mul(vecs ...SmartVector) SmartVector {
	coeffs := make([]int, len(vecs))
	for i := range coeffs {
		coeffs[i] = 1
	}

	return Product(coeffs, vecs)
}

func ScalarMul(vec SmartVector, x field.Element) SmartVector {
	xVec := NewConstant(x, vec.Len())
	return Mul(vec, xVec)
}

func InnerProduct(a, b SmartVector) field.Element {
	if a.Len() != b.Len() {
		panic("length mismatch")
	}

	var res field.Element

	for i := 0; i < a.Len(); i++ {
		var tmp field.Element
		a_, b_ := a.Get(i), b.Get(i)
		tmp.Mul(&a_, &b_)
		res.Add(&res, &tmp)
	}

	return res
}

func PolyEval(vecs []SmartVector, x field.Element) SmartVector {

	if len(vecs) == 0 {
		panic("no input vectors")
	}

	length := vecs[0].Len()

	// Preallocate the intermediate values
	resReg := make([]field.Element, length)
	tmpVec := make([]field.Element, length)
	var tmpF, resCon field.Element
	var anyReg, anyCon bool
	xPow := field.One()

	accumulateReg := func(acc, v []field.Element, x field.Element) {
		for i := 0; i < length; i++ {
			tmpF.Mul(&v[i], &x)
			acc[i].Add(&acc[i], &tmpF)
		}
	}

	// Computes the polynomial operation separately on the const,
	// windows and regular and the aggregate the results at the end.
	// The computation is done following horner's method.
	for _, v := range vecs {

		switch casted := v.(type) {
		case *Constant:
			anyCon = true
			tmpF.Mul(&casted.val, &xPow)
			resCon.Add(&resCon, &tmpF)
		case *Regular:
			anyReg = true
			v := *casted
			accumulateReg(resReg, v, xPow)
		case *PaddedCircularWindow:
			// treat it as a regular, reusing the buffer
			anyReg = true
			casted.WriteInSlice(tmpVec)
			accumulateReg(resReg, tmpVec, xPow)
		}

		xPow.Mul(&x, &xPow)
	}

	switch {
	case anyCon && anyReg:
		for i := range resReg {
			resReg[i].Add(&resReg[i], &resCon)
		}
		return NewRegular(resReg)
	case anyCon && !anyReg:
		return NewConstant(resCon, length)
	case !anyCon && anyReg:
		return NewRegular(resReg)
	}

	// can only happen if no vectors are found or if an unknow type is found
	panic("unreachable")
}

// Naive implementation of the butterfly
func Butterfly(a, b SmartVector) (SmartVector, SmartVector) {
	return Add(a, b), Sub(a, b)
}
