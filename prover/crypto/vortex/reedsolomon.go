package vortex

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

// rsEncode encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
func (p *Params) rsEncode(v smartvectors.SmartVector, pool mempool.MemPool) smartvectors.SmartVector {

	// Short path, v is a constant vector. It's encoding is also a constant vector
	// with the same value.
	if cons, ok := v.(*smartvectors.Constant); ok {
		return smartvectors.NewConstant(cons.Val(), p.NumEncodedCols())
	}

	// Interpret the smart-vectors as a polynomial in Lagrange form
	// and returns a vector of coefficients.
	asCoeffs := smartvectors.FFTInverse(v, fft.DIT, true, 0, 0, pool)
	if pool != nil {
		defer func() {
			if pooled, ok := asCoeffs.(*smartvectors.Pooled); ok {
				pooled.Free(pool)
			}
		}()
	}

	// Pad the coefficients
	expandedCoeffs := make([]field.Element, p.NumEncodedCols())
	asCoeffs.WriteInSlice(expandedCoeffs[:asCoeffs.Len()])

	// This is not memory that will be recycled easily
	return smartvectors.FFT(smartvectors.NewRegular(expandedCoeffs), fft.DIT, true, 0, 0, nil)
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) isCodeword(v smartvectors.SmartVector) error {
	coeffs := smartvectors.FFTInverse(v, fft.DIT, true, 0, 0, nil)
	for i := p.NbColumns; i < p.NumEncodedCols(); i++ {
		c := coeffs.Get(i)
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
