package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// _rsEncodeBase encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
func (p *Params) _rsEncodeBase(v smartvectors.SmartVector) smartvectors.SmartVector {

	// Short path, v is a constant vector. It's encoding is also a constant vector
	// with the same value.
	if cons, ok := v.(*smartvectors.Constant); ok {
		return smartvectors.NewConstant(cons.Val(), p.NumEncodedCols())
	}

	expandedCoeffs := make([]field.Element, p.NumEncodedCols())

	// copy the input
	v.WriteInSlice(expandedCoeffs[:v.Len()])

	const rho = 2
	if rho != p.BlowUpFactor {
		smallDomain := p.Domains[0]
		largeDomain := p.Domains[1]

		smallDomain.FFTInverse(expandedCoeffs[:v.Len()], fft.DIF, fft.WithNbTasks(1))
		utils.BitReverse(expandedCoeffs[:v.Len()])

		largeDomain.FFT(expandedCoeffs, fft.DIF, fft.WithNbTasks(1))
		utils.BitReverse(expandedCoeffs)

		return smartvectors.NewRegular(expandedCoeffs)
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.
	inputCoeffs := koalabear.Vector(expandedCoeffs[:p.NbColumns])

	p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns - 1; j >= 0; j-- {
		expandedCoeffs[rho*j+1] = expandedCoeffs[j]
		expandedCoeffs[rho*j] = v.Get(j)
	}

	return smartvectors.NewRegular(expandedCoeffs)
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) isCodeword(v smartvectors.SmartVector) error {
	if v.Len() != p.NumEncodedCols() {
		return fmt.Errorf("invalid length for a codeword")
	}
	coeffs := make([]field.Element, p.NumEncodedCols())
	v.WriteInSlice(coeffs)
	utils.BitReverse(coeffs)
	p.Domains[1].FFTInverse(coeffs, fft.DIT, fft.WithNbTasks(1))
	for i := p.NbColumns; i < p.NumEncodedCols(); i++ {
		c := coeffs[i]
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
