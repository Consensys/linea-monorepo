package vortex2

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Reevaluates a smart-vector into a coset. rho represents the
// blowup factor. The vector is viewed as a vector of evaluations
func (p *Params) rsEncode(v smartvectors.SmartVector) smartvectors.SmartVector {

	// Short path, v is a constant vector. It's encoding is also a constant vector
	// with the same value.
	if cons, ok := v.(*smartvectors.Constant); ok {
		return smartvectors.NewConstant(cons.Val(), p.NumEncodedCols())
	}

	// Interpret the smart-vectors as a polynomial in Lagrange form
	// and returns a vector of coefficients.
	asCoeffs := smartvectors.FFTInverse(v, fft.DIT, true, 0, 0)

	// Pad the coefficients
	expandedCoeffs := make([]field.Element, p.NumEncodedCols())
	asCoeffs.WriteInSlice(expandedCoeffs[:asCoeffs.Len()])

	return smartvectors.FFT(smartvectors.NewRegular(expandedCoeffs), fft.DIT, true, 0, 0)
}

// IsCodeword returns true iff the argument is a correct codeword
func (p *Params) isCodeword(v smartvectors.SmartVector) error {
	coeffs := smartvectors.FFTInverse(v, fft.DIT, true, 0, 0)
	for i := p.NbColumns; i < p.NumEncodedCols(); i++ {
		c := coeffs.Get(i)
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
