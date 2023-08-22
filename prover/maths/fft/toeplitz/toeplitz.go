package toeplitz

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/matrix"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft/fastpoly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Represents a precomputed (rectangular) toeplitz matrix
type ToeplitzMatrix struct {
	// Must have numCols + numRows - 1 entries
	precomp          []field.Element
	numCols, numRows int
	domain           *fft.Domain
}

// The coefficients should be given as
// `the first row
// the first column without the first element`
func NewToeplitzMatrix(firstRow, firstColumn []field.Element, domain *fft.Domain) ToeplitzMatrix {

	card := int(domain.Cardinality)
	numCols := len(firstRow)
	numRows := len(firstColumn) + 1

	// The domain should be larger than the coefficient
	if numRows+numCols-1 > int(card) {
		utils.Panic("Coefficients larger than the cardinality %v %v", numRows+numCols-1, card)
	}

	// TODO : When the coeffs are smaller, we could handle this by copying
	// the first numCols entries of `coeff` in the first entries of precomp
	// and the first column in the last entries
	precomp := make([]field.Element, card)
	firstRow_ := vector.DeepCopy(firstRow)
	vector.Reverse(firstRow_)
	copy(precomp[card-numCols-numRows+1:], firstRow_)
	copy(precomp[card-numRows+1:], firstColumn)

	// Pre-apply the FFT on the vector
	domain.FFT(precomp, fft.DIF)

	return ToeplitzMatrix{
		precomp: precomp,
		numCols: numCols,
		numRows: numRows,
		domain:  domain,
	}
}

// Performs a Toeplitz matrix multiplication
func (toep ToeplitzMatrix) Mul(v []field.Element) []field.Element {

	// Sanity-check enforces that the vector has the right size
	if len(v) != toep.numCols {
		utils.Panic("The size of the toeplitz matrix mismatches %v with size of v %v", toep.numCols, len(v))
	}

	// Pad all the inputs with zeroes
	res := make([]field.Element, toep.domain.Cardinality)
	copy(res, v)
	fastpoly.MultModXnMinus1Precomputed(toep.domain, res, res, toep.precomp)

	return res[toep.domain.Cardinality-uint64(toep.numRows):]
}

// Performs a Toeplitz matrix multiplication
func (toep ToeplitzMatrix) BatchMul(v [][]field.Element) [][]field.Element {

	// Sanity-check enforces that the vector has the right size
	if len(v) != toep.numCols {
		utils.Panic("The size of the toeplitz matrix mismatches %v with size of v %v", toep.numCols, len(v))
	}

	// Pad all the inputs with zeroes
	zeroes := matrix.Zeroes(int(toep.domain.Cardinality)-len(v), len(v[0]))
	res := matrix.DeepCopy(v)
	res = append(res, zeroes...)
	fastpoly.BatchMultModXnMinus1Precomputed(toep.domain, res, res, toep.precomp)

	return res[toep.domain.Cardinality-uint64(toep.numRows):]
}
