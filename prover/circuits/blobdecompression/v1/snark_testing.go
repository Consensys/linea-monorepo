package v1

import (
	"crypto/rand"
	"encoding/binary"
	"os"

	"github.com/consensys/gnark/frontend"
	test_vector_utils "github.com/consensys/gnark/std/utils/test_vectors_utils"
	"github.com/stretchr/testify/require"
)

// TODO Delete most of the following

// Larger tests will need to be run as independent executables. So we need to make some testing utils publicly available.

const testDictPath = "../../../lib/compressor/compressor_dict.bin"

func assertSliceEquals(api frontend.API, a, b []frontend.Variable) {
	api.AssertIsEqual(len(a), len(b)) // TODO checked in compile time?
	for i := range a {
		api.AssertIsEqual(a[i], b[i])
	}
}

func padBytes(b []byte, targetLen int) []frontend.Variable {
	padded := make([]byte, targetLen)
	copy(padded, b)
	if _, err := rand.Read(padded[len(b):]); err != nil {
		panic(err)
	}
	return test_vector_utils.ToVariableSlice(padded)
}

func getDict(t require.TestingT) []byte {
	dict, err := os.ReadFile(testDictPath)
	require.NoError(t, err)
	return dict
}

func randIntn(n int) int {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}
