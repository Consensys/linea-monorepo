package ringsis

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/sis"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimbDecompose64(t *testing.T) {
	a_ := []int{42, 27, 78}
	b_ := []int{15, 41, 23}
	a := field.NewElement(uint64(a_[0] + a_[1]*256 + a_[2]*1<<16))
	b := field.NewElement(uint64(b_[0] + b_[1]*256 + b_[2]*1<<16))

	result := make([]field.Element, 64)
	limbDecompose64(result, a, b)

	for i := range result {
		result[i] = field.MulR(result[i])
	}

	assert.Equal(t, uint64(a_[0]), result[0].Uint64())
	assert.Equal(t, uint64(a_[1]), result[1].Uint64())
	assert.Equal(t, uint64(a_[2]), result[2].Uint64())
	assert.Equal(t, uint64(b_[0]), result[32+0].Uint64())
	assert.Equal(t, uint64(b_[1]), result[32+1].Uint64())
	assert.Equal(t, uint64(b_[2]), result[32+2].Uint64())
}

func TestSmartTransversalSisHash(t *testing.T) {

	numCols := 32

	randCon := func() smartvectors.SmartVector {
		var x field.Element
		x.SetRandom()
		return smartvectors.NewConstant(x, numCols)
	}

	randReg := func() smartvectors.SmartVector {
		return smartvectors.Rand(numCols)
	}

	list := []smartvectors.SmartVector{
		randCon(),
		randReg(),
		randReg(),
		randReg(),
		randReg(),
		randCon(),
		randCon(),
		randCon(),
		randCon(),
		randReg(),
		randReg(),
		randReg(),
		randReg(),
		randCon(),
		randCon(),
		randCon(),
	}

	params := Params{LogTwoBound: 8, LogTwoDegree: 6}
	key := GenerateKey(params, len(list))

	result := smartTransversalHash_SIS_8_64(
		key.gnarkInternal.Ag,
		list,
		key.twiddleCosets,
		key.gnarkInternal.Domain,
	)

	expected := key.TransversalHash(list)

	for i := range result {
		require.Equalf(t, expected[i].String(), result[i].String(), "at position %v", i)
	}

}

func BenchmarkFFT64(b *testing.B) {

	vec := vector.Rand(64)
	params := Params{LogTwoBound: 8, LogTwoDegree: 6}
	key := GenerateKey(params, 32)
	twids := key.twiddleCosets

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sis.FFT64(vec, twids)
	}

}

func BenchmarkLimbDecompose(b *testing.B) {
	var x, y field.Element
	x.SetRandom()
	y.SetRandom()
	limbs := make([]field.Element, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limbDecompose64(limbs, x, y)
	}
}

func BenchmarkMulModAcc(b *testing.B) {
	x, y, res := vector.Rand(64), vector.Rand(64), vector.Rand(64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mulModAcc(res, x, y)
	}
}
