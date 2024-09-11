package internal_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	snarkTestUtils "github.com/consensys/zkevm-monorepo/prover/circuits/internal/test_utils"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestChecksumSlice(t *testing.T) {
	sum := internal.ChecksumSlice([][]byte{{0}, {1}, {2}})
	snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		s := internal.VarSlice{
			Values: []frontend.Variable{0, 1, 2, 3},
			Length: 3,
		}

		if hsh, err := mimc.NewMiMC(api); err != nil {
			panic(err)
		} else {
			return []frontend.Variable{s.Checksum(api, &hsh)}
		}

	}, sum)(t)
}

func TestChecksumSubSlices(t *testing.T) {
	testChecksumSubSlices(t, 2, 1, 1)
	testChecksumSubSlices(t, 2, 1, 2)
	testChecksumSubSlices(t, 2, 2, 1, 1)
	testChecksumSubSlices(t, 8, 6, 1, 2, 3)
}

func testChecksumSubSlices(t *testing.T, bigSliceLength, lengthsSliceLength int, lengths ...int) {
	bigSliceInts := utils.RangeSlice[uint64](bigSliceLength)

	bigSliceBytes := internal.MapSlice(internal.Uint64To32Bytes, bigSliceInts...)
	endPoints := internal.PartialSumsInt(lengths)
	assert.LessOrEqual(t, endPoints[len(endPoints)-1], bigSliceLength)
	assert.LessOrEqual(t, len(lengths), lengthsSliceLength)

	sums := make([]frontend.Variable, len(lengths))
	start := 0
	for i := range sums {
		sums[i] = internal.ChecksumSlice(internal.MapSlice(func(x [32]byte) []byte { return x[:] }, bigSliceBytes[start:endPoints[i]]...))
		start = endPoints[i]
	}

	endPointsSnark := make([]frontend.Variable, lengthsSliceLength)
	for n := utils.Copy(endPointsSnark, internal.PartialSumsInt(lengths)); n < lengthsSliceLength; n++ {
		endPointsSnark[n] = n * 234
	}

	t.Run(fmt.Sprintf("%d,%d,%v", bigSliceLength, lengthsSliceLength, lengths), snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		hsh, err := mimc.NewMiMC(api)
		if err != nil {
			panic(err)
		}
		return internal.ChecksumSubSlices(api, &hsh, utils.ToVariableSlice(bigSliceInts),
			internal.VarSlice{
				Values: endPointsSnark,
				Length: len(lengths),
			})[:len(lengths)]
	}, sums...))
}

func TestConcat(t *testing.T) {
	snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		res := internal.Concat(api, 3, internal.VarSlice{[]frontend.Variable{2}, 1}, internal.VarSlice{[]frontend.Variable{3}, 0})
		return append(res.Values, res.Length)
	}, 2, 0, 0, 1)(t)
}

func TestReduceBytes(t *testing.T) {

	var cases [100][]byte
	b, err := utils.HexDecodeString("0x8de61d4d9c4891236da9646d464bd9b5757991e201678679f3f2abec6df666b8")
	assert.NoError(t, err)
	cases[0] = b

	for i := 1; i < len(cases); i++ {
		var b [32]byte
		_, err = rand.Read(b[:])
		assert.NoError(t, err)
		cases[i] = b[:]
	}

	reduced := make([][]byte, len(cases))
	var x bn254fr.Element
	for i := range cases {
		x.SetBytes(cases[i])
		reducedI := x.Bytes()
		reduced[i] = reducedI[:]
	}

	snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		for i := range cases {
			got := utils.ReduceBytes[emulated.BN254Fr](api, utils.ToVariableSlice(cases[i]))
			internal.AssertSliceEquals(api,
				got,
				utils.ToVariableSlice(reduced[i]),
			)
		}

		return nil
	})(t)
}

func TestPartitionSliceEmulated(t *testing.T) {
	selectors := []int{1, 0, 2, 2, 1}

	s := make([]fr381.Element, len(selectors))
	for i := range s {
		_, err := s[i].SetRandom()
		assert.NoError(t, err)
	}

	subs := make([][]fr381.Element, 3)
	for i := range subs {
		subs[i] = make([]fr381.Element, 0, len(selectors)-1)
	}

	for i := range s {
		subs[selectors[i]] = append(subs[selectors[i]], s[i])
	}

	snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {

		field, err := emulated.NewField[emulated.BLS12381Fr](api)
		assert.NoError(t, err)

		// convert randomized elements to emulated
		sEmulated := elementsToEmulated(field, s)

		subsEmulatedExpected := internal.MapSlice(func(s []fr381.Element) []emulated.Element[emulated.BLS12381Fr] {
			return elementsToEmulated(field, append(s, make([]fr381.Element, cap(s)-len(s))...)) // pad with zeros to see if padding is done correctly
		}, subs...)

		subsEmulated := internal.PartitionSliceEmulated(api, sEmulated, utils.ToVariableSlice(selectors), internal.MapSlice(func(s []fr381.Element) int { return cap(s) }, subs...)...)

		assert.Equal(t, len(subsEmulatedExpected), len(subsEmulated))
		for i := range subsEmulated {
			assert.Equal(t, len(subsEmulatedExpected[i]), len(subsEmulated[i]))
			for j := range subsEmulated[i] {
				field.AssertIsEqual(&subsEmulated[i][j], &subsEmulatedExpected[i][j])
			}
		}

		return nil
	})(t)
}

func elementsToEmulated(field *emulated.Field[emulated.BLS12381Fr], s []fr381.Element) []emulated.Element[emulated.BLS12381Fr] {
	return internal.MapSlice(func(element fr381.Element) emulated.Element[emulated.BLS12381Fr] {
		return *field.NewElement(internal.MapSlice(func(x uint64) frontend.Variable { return x }, element[:]...))
	}, s...)
}

func TestPartitionSlice(t *testing.T) {
	const (
		nbSubs   = 3
		sliceLen = 10
	)

	test := func(slice []frontend.Variable, selectors []int, subsSlack []int) func(*testing.T) {
		assert.Equal(t, len(selectors), len(slice))
		assert.Equal(t, len(subsSlack), nbSubs)

		subs := make([][]frontend.Variable, nbSubs)
		for j := range subs {
			subs[j] = make([]frontend.Variable, 0, sliceLen)
		}

		for j := range slice {
			subs[selectors[j]] = append(subs[selectors[j]], slice[j])
		}

		for j := range subs {
			subs[j] = append(subs[j], utils.ToVariableSlice(make([]int, subsSlack[j]))...) // add some padding
		}

		return snarkTestUtils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {

			slice := utils.ToVariableSlice(slice)

			subsEncountered := internal.MapSlice(func(s []frontend.Variable) []frontend.Variable { return make([]frontend.Variable, len(s)) }, subs...)
			internal.PartitionSlice(api, slice, utils.ToVariableSlice(selectors), subsEncountered...)

			assert.Equal(t, len(subs), len(subsEncountered))
			for j := range subsEncountered {
				internal.AssertSliceEquals(api, subsEncountered[j], subs[j])
			}

			return nil
		})
	}

	test([]frontend.Variable{5}, []int{2}, []int{1, 0, 0})(t)
	test([]frontend.Variable{1, 2, 3}, []int{0, 1, 2}, []int{0, 0, 0})
	test(utils.ToVariableSlice(utils.RangeSlice[int](10)), []int{0, 1, 2, 0, 0, 0, 1, 1, 1, 2}, []int{0, 0, 0})

	for i := 0; i < 200; i++ {

		slice := make([]frontend.Variable, sliceLen)
		for j := range slice {
			var x fr377.Element
			_, err := x.SetRandom()
			slice[j] = &x
			assert.NoError(t, err)
		}

		selectors := test_utils.RandIntSliceN(sliceLen, nbSubs)
		subsSlack := test_utils.RandIntSliceN(nbSubs, 2)

		t.Run(fmt.Sprintf("selectors=%v,slack=%v", selectors, subsSlack), test(slice, selectors, subsSlack))
	}
}
