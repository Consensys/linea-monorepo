package vortex

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// rsEncode encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
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

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
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
