package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Compute the FFT of a vector
// Decimation:
//   - Either DIT : input in bit-reverse order - output in normal order
//   - Or DIF : input in normal order - output in bit reversed order
//
// BitReverse:
//   - If set to true, this cancels the decimation and
//     forces : input normal order - output normal order
//
// CosetRatio > CosetID:
//   - Specifies on which coset to perform the operation
//   - 0, 0 to assert that the transformation should not be done over a coset
func FFT(v SmartVector, decimation fft.Decimation, bitReverse bool, cosetRatio int, cosetID int) SmartVector {

	// Sanity-check on the size of the vector v
	assertPowerOfTwoLen(v.Len())

	/*
		Try to capture the special cases
	*/
	switch x := v.(type) {
	case *Constant:
		if x.val.IsZero() {
			// The fft of the zero vec is zero
			return x.DeepCopy()
		}

		if cosetID == 0 && cosetRatio == 0 {
			// The FFT is a (c*N, 0, 0, ...), no matter the bitReverse or decimation
			// It's a multiple of the first Lagrange polynomial.
			constTerm := field.NewElement(uint64(x.length))
			constTerm.Mul(&constTerm, &x.val)
			return NewPaddedCircularWindow([]field.Element{constTerm}, field.Zero(), 0, x.length)
		}
	case *PaddedCircularWindow:
		// The polynomial is the constant polynomial, response does not depends on the decimation
		// or bitReverse
		interval := x.interval()
		if interval.intervalLen == 1 && interval.start() == 0 && x.paddingVal.IsZero() {
			// In this case, the response is a constant vector
			return NewConstant(x.window[0], x.Len())
		}
	}

	// Else : we run the FFT directly
	res := make([]field.Element, v.Len())
	v.WriteInSlice(res)

	domain := fft.NewDomain(v.Len())
	oncoset := false

	if cosetID != 0 || cosetRatio != 0 {
		oncoset = true
		domain = domain.WithCustomCoset(cosetRatio, cosetID)
	}

	if decimation == fft.DIT {
		// Optionally, bitReverse the input
		if bitReverse {
			fft.BitReverse(res)
		}
		domain.FFT(res, fft.DIT, oncoset)
	} else {
		// Likewise, the optionally rearrange the input in correct order
		domain.FFT(res, fft.DIF, oncoset)
		if bitReverse {
			fft.BitReverse(res)
		}
	}
	return NewRegular(res)
}

// Compute the FFT inverse of a vector
// Decimation:
//   - Either DIT : input in bit-reverse order - output in normal order
//   - Or DIF : input in normal order - output in bit reversed order
//
// BitReverse:
//   - If set to true, this cancels the decimation and
//     forces : input normal order - output normal order
//
// CosetRatio > CosetID:
//   - Specifies on which coset to perform the operation
//   - 0, 0 to assert that the transformation should not be done over a coset
func FFTInverse(v SmartVector, decimation fft.Decimation, bitReverse bool, cosetRatio int, cosetID int) SmartVector {

	// Sanity-check on the size of the vector v
	assertPowerOfTwoLen(v.Len())

	/*
		Try to capture the special cases
	*/
	switch x := v.(type) {
	case *Constant:
		if x.val.IsZero() {
			// The fft inverse of the zero vec is zero
			return x.DeepCopy()
		}

		if cosetID == 0 && cosetRatio == 0 {
			// It's the constant polynomial. If it is not on coset then there is a trick
			return NewPaddedCircularWindow([]field.Element{x.val}, field.Zero(), 0, x.length)
		}

	case *PaddedCircularWindow:
		// It's a multiple of the first Lagrange polynomial c * (1 + x + x^2 + x^3 + ...)
		// The response is (c) = (c/N, c/N, c/N, ...)
		interval := x.interval()
		if interval.intervalLen == 1 && interval.start() == 0 && x.paddingVal.IsZero() {
			constTerm := field.NewElement(uint64(x.Len()))
			constTerm.Inverse(&constTerm)
			constTerm.Mul(&constTerm, &x.window[0])
			// In this case, the response is a constant vector
			return NewConstant(constTerm, x.Len())
		}
	}

	// Else : we run the FFT directly
	res := make([]field.Element, v.Len())
	oncoset := false
	v.WriteInSlice(res)

	domain := fft.NewDomain(v.Len())
	if cosetID != 0 || cosetRatio != 0 {
		// Optionally equip the domain with a coset
		oncoset = true
		domain = domain.WithCustomCoset(cosetRatio, cosetID)
	}

	if decimation == fft.DIF {
		// Optionally, bitReverse the output
		domain.FFTInverse(res, fft.DIF, oncoset)
		if bitReverse {
			fft.BitReverse(res)
		}
	} else {
		// Likewise, the optionally rearrange the input in correct order
		if bitReverse {
			fft.BitReverse(res)
		}
		domain.FFTInverse(res, fft.DIT, oncoset)
	}
	return NewRegular(res)
}
