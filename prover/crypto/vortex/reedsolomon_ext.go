package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// rsEncode encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
func (p *Params) rsEncodeExt(v smartvectors.SmartVector) smartvectors.SmartVector {

	// Short path, v is a constant vector. It's encoding is also a constant vector
	// with the same value.
	if cons, ok := v.(*smartvectors.Constant); ok {
		return smartvectors.NewConstant(cons.Val(), p.NumEncodedCols())
	}

	expandedCoeffs := make([]fext.Element, p.NumEncodedCols())

	// copy the input
	v.WriteInSliceExt(expandedCoeffs[:v.Len()])

	const rho = 2
	if rho != p.BlowUpFactor {
		smallDomain := p.Domains[0]
		largeDomain := p.Domains[1]

		smallDomain.FFTInverseExt(expandedCoeffs[:v.Len()], fft.DIF, fft.WithNbTasks(2))
		utils.BitReverse(expandedCoeffs[:v.Len()])

		largeDomain.FFTExt(expandedCoeffs, fft.DIF, fft.WithNbTasks(2))
		utils.BitReverse(expandedCoeffs)

		return smartvectors.NewRegularExt(expandedCoeffs)
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.
	inputCoeffs := extensions.Vector(expandedCoeffs[:p.NbColumns])

	p.Domains[0].FFTInverseExt(inputCoeffs, fft.DIF, fft.WithNbTasks(2))
	inputCoeffs.MulByElement(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFTExt(inputCoeffs, fft.DIT, fft.WithNbTasks(2))
	for j := p.NbColumns - 1; j >= 0; j-- {
		expandedCoeffs[rho*j+1] = expandedCoeffs[j]
		expandedCoeffs[rho*j] = v.GetExt(j)
	}

	return smartvectors.NewRegularExt(expandedCoeffs)
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) isCodewordExt(v smartvectors.SmartVector) error {
	coeffs := smartvectors.FFTInverseExt(v, fft.DIT, true, 0, 0)
	for i := p.NbColumns; i < p.NumEncodedCols(); i++ {
		c := coeffs.GetExt(i)
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
