package reedsolomon

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

type RsParams struct {

	// Domain[0] for FFT^-1, Domain[1] for FFT
	Domains [2]*fft.Domain

	// PolyDomain is the standard T-point NTT domain (no coset shift).
	// It is used to convert T Lagrange evaluations of the original polynomial
	// directly into T monomial coefficients, avoiding any N-length work.
	PolyDomain *fft.Domain

	// Coset table Domain[0], bit reversed
	CosetTableBitReverse field.Vector

	// code rate
	Rate int
}

func NewRsParams(size, rate int) *RsParams {

	var res RsParams
	res.Rate = rate

	// TODO @thomas fixme handle error
	shift, err := koalabear.Generator(uint64(size * rate))
	if err != nil {
		panic(err)
	}

	res.Domains[0] = fft.NewDomain(uint64(size), fft.WithShift(shift))
	res.Domains[1] = fft.NewDomain(uint64(rate * size))
	res.PolyDomain = fft.NewDomain(uint64(size))

	cosetTable, err := res.Domains[0].CosetTable()
	// TODO @thomas fixme handle error
	if err != nil {
		panic(err)
	}
	cosetTableBitReverse := make(field.Vector, len(cosetTable))
	copy(cosetTableBitReverse, cosetTable)
	utils.BitReverse(cosetTableBitReverse)

	res.CosetTableBitReverse = cosetTableBitReverse

	return &res
}

// TODO @thomas rename that
func (r *RsParams) NbEncodedColumns() int {
	return int(r.Domains[1].Cardinality)
}

// TODO @thomas rename that
func (r *RsParams) NbColumns() int {
	return int(r.Domains[0].Cardinality)
}

// RsEncodeBase encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
func (p *RsParams) RsEncodeBase(v smartvectors.SmartVector) smartvectors.SmartVector {

	// Short path, v is a constant vector. It's encoding is also a constant vector
	// with the same value.
	if cons, ok := v.(*smartvectors.Constant); ok {
		return smartvectors.NewConstant(cons.Val(), p.NbEncodedColumns())
	}

	expandedCoeffs := make([]field.Element, p.NbEncodedColumns())

	// copy the input
	v.WriteInSlice(expandedCoeffs[:v.Len()])

	const rho = 2
	if rho != p.Rate {

		smallDomain := p.Domains[0]
		largeDomain := p.Domains[1]

		smallDomain.FFTInverse(expandedCoeffs[:v.Len()], fft.DIF, fft.WithNbTasks(1))

		n := v.Len()

		// @thomas this seems to work... bitreverse commutes with scaling ?
		for j := n - 1; j > 0; j-- {
			expandedCoeffs[p.Rate*j] = expandedCoeffs[j]
			expandedCoeffs[j] = field.Element{}
		}

		largeDomain.FFT(expandedCoeffs, fft.DIT, fft.WithNbTasks(1))

		return smartvectors.NewRegular(expandedCoeffs)
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.
	inputCoeffs := koalabear.Vector(expandedCoeffs[:p.NbColumns()])

	p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns() - 1; j >= 0; j-- {
		expandedCoeffs[rho*j+1] = expandedCoeffs[j]
		expandedCoeffs[rho*j], _ = v.GetBase(j)
	}

	return smartvectors.NewRegular(expandedCoeffs)
}

// ExtEvalToCoefficients converts a Reed-Solomon codeword (N = T×RS E4 evaluations)
// to its T polynomial coefficients. The returned slice has length NbColumns().
// Panics if the input is not an RS codeword.
func (p *RsParams) ExtEvalToCoefficients(v smartvectors.SmartVector) smartvectors.SmartVector {
	n := p.NbEncodedColumns()
	t := p.NbColumns()

	coeffs := make([]fext.Element, n)
	v.WriteInSliceExt(coeffs)
	p.Domains[1].FFTInverseExt(coeffs, fft.DIF, fft.WithNbTasks(1))
	utils.BitReverse(coeffs)
	return smartvectors.NewRegularExt(coeffs[:t])
}

// ExtCoefficientsToAllEvaluations evaluates the degree-T polynomial given by
// T monomial coefficients (E4) at all N = NbEncodedColumns() points of the RS
// domain (ω_N^0, ω_N^1, ..., ω_N^{N-1}).  The returned slice has length N.
//
// It is cheaper than calling ExtCoefficientsEvalAt K times when K > blowup×log₂(N).
func (p *RsParams) ExtCoefficientsToAllEvaluations(coefficients []fext.Element) []fext.Element {
	n := p.NbEncodedColumns()

	buf := make([]fext.Element, n)
	copy(buf, coefficients)
	// DIT FFT expects bit-reversed input; natural-order evaluations come out.
	utils.BitReverse(buf)
	p.Domains[1].FFTExt(buf, fft.DIT, fft.WithNbTasks(1))
	return buf
}
