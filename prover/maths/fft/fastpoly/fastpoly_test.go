package fastpoly_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/matrix"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft/fastpoly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"

	"github.com/stretchr/testify/require"
)

func TestMulitplication(t *testing.T) {

	n := 8
	// smallN := 6
	domain := fft.NewDomain(n)

	// Unitary X : multiplying by it results in identity
	{
		unit := make([]field.Element, n)
		unit[0].SetOne()

		vec := vector.Rand(n)
		vecBackup := vector.DeepCopy(vec)
		res := make([]field.Element, n)
		fastpoly.MultModXMinus1(domain, res, vec, unit)

		require.Equal(t, vector.Prettify(vecBackup), vector.Prettify(res))
	}

	// Unitary X : multiplying by it results in identity
	// With precomputation
	{
		unit := make([]field.Element, n)
		unit[0].SetOne()
		domain.FFT(unit, fft.DIF)

		vec := vector.Rand(n)
		vecBackup := vector.DeepCopy(vec)
		res := make([]field.Element, n)
		fastpoly.MultModXnMinus1Precomputed(domain, res, vec, unit)

		require.Equal(t, vector.Prettify(vecBackup), vector.Prettify(res))
	}

	// Polynomial X : multiplying by it results in "circular" permutation
	{
		shift := make([]field.Element, n)
		shift[1].SetOne()

		vec := vector.Rand(n)
		vecBackup := vector.DeepCopy(vec)
		res := make([]field.Element, n)
		fastpoly.MultModXMinus1(domain, res, vec, shift)

		require.Equal(t, vector.Prettify(res[1:]), vector.Prettify(vecBackup[:n-1]))
		require.Equal(t, res[0].String(), vecBackup[n-1].String())
	}

	// Polynomial X : multiplication with precomputation
	{
		shift := make([]field.Element, n)
		shift[1].SetOne()
		domain.FFT(shift, fft.DIF)

		vec := vector.Rand(n)
		vecBackup := vector.DeepCopy(vec)
		res := make([]field.Element, n)
		fastpoly.MultModXnMinus1Precomputed(domain, res, vec, shift)

		require.Equal(t, vector.Prettify(res[1:]), vector.Prettify(vecBackup[:n-1]))
		require.Equal(t, res[0].String(), vecBackup[n-1].String())

	}

	// Consistency between batching and non-batching
	{
		precomp := vector.Rand(n)
		rand := vector.Rand(n)
		nSamples := 10

		randMat := matrix.RepeatSubslice(rand, nSamples)
		randMat = matrix.Transpose(randMat)
		resMat := matrix.Zeroes(n, nSamples)
		res := make([]field.Element, n)

		fastpoly.BatchMultModXnMinus1Precomputed(domain, resMat, randMat, precomp)
		fastpoly.MultModXnMinus1Precomputed(domain, res, rand, precomp)

		resBis := matrix.Transpose(resMat)[0]

		require.Equal(t, vector.Prettify(res), vector.Prettify(resBis))

	}
}
