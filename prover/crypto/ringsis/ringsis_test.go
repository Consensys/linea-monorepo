package ringsis_test

import (
	"fmt"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashModXnMinusOne(t *testing.T) {

	key := ringsis.StdParams.GenerateKey(2)

	limbs := vector.Repeat(field.One(), key.Degree*len(key.A))

	dualHash := key.HashModXnMinus1(limbs)
	laidoutKey := key.LaidOutKey() // accounts for the Montgommery skip
	recomputedHash := []field.Element{}

	for i := range key.A {
		ai := laidoutKey[i*key.Degree : (i+1)*key.Degree]
		si := limbs[i*key.Degree : (i+1)*key.Degree]
		tmp := poly.Mul(ai, si)
		recomputedHash = poly.Add(recomputedHash, tmp)
	}

	for i := key.Degree; i < len(recomputedHash); i++ {
		recomputedHash[i-key.Degree].Add(&recomputedHash[i-key.Degree], &recomputedHash[i])
	}

	recomputedHash = recomputedHash[:key.Degree]
	require.Equal(t, dualHash, recomputedHash)
}

func TestHashModXnMinusOne_LengthDoesNotDivide(t *testing.T) {

	numInput := 5
	params := ringsis.Params{LogTwoBound_: 8, LogTwoDegree: 6}
	key := params.GenerateKey(numInput)
	limbs := vector.Repeat(field.One(), key.NumLimbs()*numInput)
	dualHash := key.HashModXnMinus1(limbs)
	laidoutKey := key.LaidOutKey() // accounts for the Montgommery skip
	recomputedHash := []field.Element{}

	for i := range key.A {
		ai := laidoutKey[i*key.Degree : (i+1)*key.Degree]
		si := make([]field.Element, key.Degree)
		copy(si, limbs[i*key.Degree:])
		tmp := poly.Mul(ai, si)
		recomputedHash = poly.Add(recomputedHash, tmp)
	}

	for i := key.Degree; i < len(recomputedHash); i++ {
		recomputedHash[i-key.Degree].Add(&recomputedHash[i-key.Degree], &recomputedHash[i])
	}

	recomputedHash = recomputedHash[:key.Degree]
	require.Equal(t, dualHash, recomputedHash)
}

func TestLimbSplit(t *testing.T) {

	sisKeys := []ringsis.Key{
		(&ringsis.Params{LogTwoBound_: 4}).GenerateKey(4),
		ringsis.StdParams.GenerateKey(4),
		(&ringsis.Params{LogTwoBound_: 8}).GenerateKey(4),
	}

	randX := mimc.BlockCompression(field.Zero(), field.Zero())
	randX.Inverse(&randX)

	arrays := [][]field.Element{
		{field.One()},
		{field.One(), field.One()},
		{field.NewElement(0), field.NewElement(1), field.NewElement(2), field.NewElement(3)},
		{field.NewElement(0xffffffffffffffff), field.NewFromString("-1"), field.NewElement(8)},
		{randX, field.One(), field.NewFromString("-1")},
	}

	coreTest := func(subT *testing.T, v []field.Element, key ringsis.Key) {
		bound := field.NewElement(1 << key.LogTwoBound)
		limbs := key.LimbSplit(v)
		for i := range v {
			subLimbs := limbs[i*key.NumLimbs() : (i+1)*key.NumLimbs()]
			recomposed := poly.EvalUnivariate(subLimbs, bound)
			assert.Equal(subT, v[i].String(), recomposed.String())
		}
	}

	for arrID, arr := range arrays {
		for keyID, key := range sisKeys {
			t.Run(
				fmt.Sprintf("Array #%v - Key #%v", arrID, keyID),
				func(t *testing.T) {
					coreTest(t, arr, key)
				},
			)
		}
	}
}

func TestHashFromLimbs(t *testing.T) {

	coreTest := func(v []field.Element) {
		key := ringsis.StdParams.GenerateKey(len(v))
		hashed := key.Hash(v)
		limbs := key.LimbSplit(v)
		hashedFromLimbs := key.HashFromLimbs(limbs)
		assert.Equal(t, hashed, hashedFromLimbs)
	}

	testcases := [][]field.Element{
		{field.One()},
		{field.One(), field.One()},
		{field.NewElement(0), field.NewElement(1), field.NewElement(2), field.NewElement(3)},
		{field.NewElement(0xffffffffffffffff), field.NewFromString("-1"), field.NewElement(8)},
		vector.Rand(4),
		vector.Rand(16),
	}

	for _, testcase := range testcases {
		coreTest(testcase)
	}
}

func TestHashFromLimbs_LengthDoesNotDivideDegree(t *testing.T) {

	coreTest := func(v []field.Element) {
		params := ringsis.Params{LogTwoBound_: 8, LogTwoDegree: 6}
		key := params.GenerateKey(len(v))
		hashed := key.Hash(v)
		limbs := key.LimbSplit(v)
		hashedFromLimbs := key.HashFromLimbs(limbs)
		assert.Equal(t, vector.Prettify(hashed), vector.Prettify(hashedFromLimbs))
	}

	testcases := [][]field.Element{
		{field.NewElement(0), field.NewElement(1), field.NewElement(2), field.NewElement(3), field.NewElement(4)},
	}

	for _, testcase := range testcases {
		coreTest(testcase)
	}
}

func TestTransveralHashFromLimbs(t *testing.T) {

	key := ringsis.StdParams.GenerateKey(4)

	inputs := make([]smartvectors.SmartVector, 4)
	for i := range inputs {
		inputs[i] = smartvectors.Rand(16)
	}

	transposed := make([][]field.Element, 16)
	for i := range transposed {
		transposed[i] = make([]fr.Element, 4)
		for j := range transposed[i] {
			transposed[i][j] = inputs[j].Get(i)
		}
	}

	res := key.TransversalHash(inputs)
	for i := range transposed {
		baseline := key.Hash(transposed[i])

		assert.Equal(t, baseline, res[i*key.OutputSize():(i+1)*key.OutputSize()])
	}
}

func TestHashLimbsFromSlice(t *testing.T) {

	numInputs := 4
	params := ringsis.Params{LogTwoBound_: 8, LogTwoDegree: 2}
	key := params.GenerateKey(numInputs)

	inputs := vector.Rand(numInputs)
	keyVec := key.LaidOutKey()
	limbs := key.LimbSplit(inputs)

	expectedHash := key.HashFromLimbs(limbs)
	hashToTest := ringsis.HashLimbsWithSlice(keyVec, limbs, key.Domain, key.Degree)

	require.Equal(t, vector.Prettify(expectedHash), vector.Prettify(hashToTest))
}

func TestHashLimbsFromSlice_LengthDoesNotDivide(t *testing.T) {

	numInputs := 5
	params := ringsis.Params{LogTwoBound_: 8, LogTwoDegree: 6}
	key := params.GenerateKey(numInputs)

	inputs := vector.Rand(numInputs)
	keyVec := key.LaidOutKey()
	limbs := key.LimbSplit(inputs)

	expectedHash := key.HashFromLimbs(limbs)
	hashToTest := ringsis.HashLimbsWithSlice(keyVec, limbs, key.Domain, key.Degree)

	require.Equal(t, vector.Prettify(expectedHash), vector.Prettify(hashToTest))
}
