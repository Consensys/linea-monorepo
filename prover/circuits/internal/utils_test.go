package internal_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/emulated"
	test_vector_utils "github.com/consensys/gnark/std/utils/test_vectors_utils"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

func TestChecksumSlice(t *testing.T) {
	sum := internal.ChecksumSlice([][]byte{{0}, {1}, {2}})
	test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
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
	bigSliceInts := test_utils.Range[uint64](bigSliceLength)

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
	for n := internal.Copy(endPointsSnark, internal.PartialSumsInt(lengths)); n < lengthsSliceLength; n++ {
		endPointsSnark[n] = n * 234
	}

	t.Run(fmt.Sprintf("%d,%d,%v", bigSliceLength, lengthsSliceLength, lengths), test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		hsh, err := mimc.NewMiMC(api)
		if err != nil {
			panic(err)
		}
		return internal.ChecksumSubSlices(api, &hsh, test_vector_utils.ToVariableSlice(bigSliceInts),
			internal.VarSlice{
				Values: endPointsSnark,
				Length: len(lengths),
			})[:len(lengths)]
	}, sums...))
}

func TestConcat(t *testing.T) {
	test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
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

	test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		for i := range cases {
			got := internal.ReduceBytes[emulated.BN254Fr](api, test_vector_utils.ToVariableSlice(cases[i]))
			internal.AssertSliceEquals(api,
				got,
				test_vector_utils.ToVariableSlice(reduced[i]),
			)
		}

		return nil
	})(t)
}
