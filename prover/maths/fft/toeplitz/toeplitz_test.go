package toeplitz_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/matrix"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft/toeplitz"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"

	"github.com/stretchr/testify/require"
)

func TestToeplitz(t *testing.T) {

	// Create a matrix that returns the first "5" coordinate
	// Only the first coefficient is set to 1, the others are zeroes
	func() {
		numCols := 12
		numRows := 5
		domainSize := numCols + numRows - 1
		domain := fft.NewDomain(domainSize)

		firstColumn := make([]field.Element, numRows-1)
		firstRow := make([]field.Element, numCols)
		firstRow[0].SetOne()

		toep := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, domain)
		vec := vector.Rand(numCols)
		res := toep.Mul(vec)

		require.Equal(t, vec[:numRows], res)
	}()

	// Create a matrix that returns the first coordinates [1:4]
	// Only the first coefficient is set to 1, the others are zeroes
	func() {
		numCols := 6
		numRows := 3
		domainSize := numCols + numRows - 1
		domain := fft.NewDomain(domainSize)

		firstColumn := make([]field.Element, numRows-1)
		firstRow := make([]field.Element, numCols)
		firstRow[1].SetOne()

		toep := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, domain)
		vec := vector.Rand(numCols)
		res := toep.Mul(vec)

		require.Equal(t, vector.Prettify(vec[1:numRows+1]), vector.Prettify(res[:numRows]))
	}()

	// Create a matrix that returns zero appended by the first two coordinates
	func() {
		numCols := 6
		numRows := 3
		domainSize := numCols + numRows - 1
		domain := fft.NewDomain(domainSize)

		firstColumn := make([]field.Element, numRows-1)
		firstRow := make([]field.Element, numCols)
		firstColumn[0].SetOne()

		toep := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, domain)
		vec := vector.Rand(numCols)
		res := toep.Mul(vec)

		require.Equal(t, vector.Prettify(vec[:numRows-1]), vector.Prettify(res[1:]))
		require.Equal(t, res[0], field.NewElement(0))
	}()

	// If we scale it on a different domain, we get the same result
	func() {
		numCols := 6
		numRows := 3
		domainSize := numCols + numRows - 1
		domain := fft.NewDomain(domainSize)
		largerDomain := fft.NewDomain(2 * domainSize)

		firstColumn := make([]field.Element, numRows-1)
		firstRow := make([]field.Element, numCols)
		firstColumn[0].SetOne()

		toep := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, domain)
		toep2 := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, largerDomain)
		vec := vector.Rand(numCols)

		res := toep.Mul(vec)
		res2 := toep2.Mul(vec)
		require.Equal(t, res, res2)
	}()

	// Consistency between `BatchMul` and `Mul`
	func() {
		numCols := 6
		numRows := 3
		nSamples := 10
		domainSize := numCols + numRows - 1
		domain := fft.NewDomain(domainSize)

		firstColumn := vector.Rand(numRows - 1)
		firstRow := vector.Rand(numCols)

		rand := vector.Rand(numCols)
		randMat := matrix.RepeatSubslice(rand, nSamples)
		randMat = matrix.Transpose(randMat)

		toep := toeplitz.NewToeplitzMatrix(firstRow, firstColumn, domain)

		res := toep.Mul(rand)
		resMat := toep.BatchMul(randMat)

		resBis := matrix.Transpose(resMat)[0]

		require.Equal(t, vector.Prettify(res), vector.Prettify(resBis))
	}()
}
