package reedsolomon

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover-v2/maths/koalabear/field"
)

type RsParams struct {

	// Domain[0] for FFT^-1, Domain[1] for FFT
	Domains [2]*fft.Domain

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
func (p *RsParams) RsEncodeBase(v []field.Element) []field.Element {

	expandedCoeffs := make([]field.Element, p.NbEncodedColumns())
	copy(expandedCoeffs, v)
	n := len(v)

	const rho = 2
	if rho != p.Rate {

		smallDomain := p.Domains[0]
		largeDomain := p.Domains[1]
		smallDomain.FFTInverse(expandedCoeffs[:n], fft.DIF, fft.WithNbTasks(1))

		// @thomas this seems to work... bitreverse commutes with scaling ?
		for j := n - 1; j > 0; j-- {
			expandedCoeffs[p.Rate*j] = expandedCoeffs[j]
			expandedCoeffs[j] = field.Element{}
		}

		largeDomain.FFT(expandedCoeffs, fft.DIT, fft.WithNbTasks(1))
		return expandedCoeffs
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.

	inputCoeffs := koalabear.Vector(expandedCoeffs[:p.NbColumns()])
	p.Domains[0].FFTInverse(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.Mul(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFT(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns() - 1; j >= 0; j-- {
		expandedCoeffs[rho*j+1] = expandedCoeffs[j]
		expandedCoeffs[rho*j] = v[j]
	}

	return expandedCoeffs
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise. The function panics if the provided v does
// not have the expected length for a codeword.
func (p *RsParams) IsCodeword(v []field.Element) error {

	if len(v) != p.NbEncodedColumns() {
		return fmt.Errorf("invalid length for a codeword, expected %v got %v", p.NbEncodedColumns(), len(v))
	}

	coeffs := make([]field.Element, p.NbEncodedColumns())
	copy(coeffs, v)

	p.Domains[1].FFTInverse(coeffs, fft.DIF, fft.WithNbTasks(1))
	utils.BitReverse(coeffs)
	for i := p.NbColumns(); i < p.NbEncodedColumns(); i++ {
		c := coeffs[i]
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}

// rsEncode encodes a vector `v` and returns the corresponding the Reed-Solomon
// codeword. The input vector is interpreted as a polynomial in Lagrange basis
// over a domain of n-roots of unity Omega_n and returns its representation in
// the Lagrange basis Omega_{n * blow-up} where blow-up corresponds to the
// inverse-rate of the code. The code is systematic as the original vector is
// interleaved within the encoded vector.
func (p *RsParams) rsEncodeExt(v []field.Ext) []field.Ext {

	expandedCoeffs := make([]field.Ext, p.NbEncodedColumns())
	copy(expandedCoeffs, v)
	n := len(v)

	const rho = 2
	if rho != p.Rate {

		smallDomain := p.Domains[0]
		largeDomain := p.Domains[1]
		smallDomain.FFTInverseExt(expandedCoeffs[:n], fft.DIF, fft.WithNbTasks(1))

		// this loop dispatches the values that are all located at the beginning
		// of the domain to the entire domain by homothety
		for j := n - 1; j > 0; j-- {
			expandedCoeffs[p.Rate*j] = expandedCoeffs[j]
			expandedCoeffs[j] = field.Ext{}
		}

		largeDomain.FFTExt(expandedCoeffs, fft.DIT, fft.WithNbTasks(1))
		return expandedCoeffs
	}

	// fast path; we avoid the bit reverse operations and work on the smaller domain.
	inputCoeffs := extensions.Vector(expandedCoeffs[:p.NbColumns()])
	p.Domains[0].FFTInverseExt(inputCoeffs, fft.DIF, fft.WithNbTasks(1))
	inputCoeffs.MulByElement(inputCoeffs, p.CosetTableBitReverse)

	p.Domains[0].FFTExt(inputCoeffs, fft.DIT, fft.WithNbTasks(1))
	for j := p.NbColumns() - 1; j >= 0; j-- {
		expandedCoeffs[rho*j+1] = expandedCoeffs[j]
		expandedCoeffs[rho*j] = v[j]
	}

	return expandedCoeffs
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *RsParams) IsCodewordExt(v []field.Ext) error {

	if len(v) != p.NbEncodedColumns() {
		return fmt.Errorf("invalid length for a codeword")
	}

	coeffs := make([]field.Ext, p.NbEncodedColumns())
	copy(coeffs, v)

	p.Domains[1].FFTInverseExt(coeffs, fft.DIF, fft.WithNbTasks(1))
	utils.BitReverse(coeffs)
	for i := p.NbColumns(); i < p.NbEncodedColumns(); i++ {
		c := coeffs[i]
		if !c.IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
