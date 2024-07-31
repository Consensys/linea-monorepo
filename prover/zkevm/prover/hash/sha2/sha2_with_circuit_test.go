//go:build !fuzzlight

package sha2

import (
	"strconv"
	"testing"
)

func TestSha2WithCircuit(t *testing.T) {

	var testCases = []testCaseFile{
		{
			InpFile:      "testdata/input.csv",
			ModFile:      "testdata/mod.csv",
			NbBlockLimit: 10,
			WithCircuit:  true,
		},
	}

	for i := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			runTestSha2(t, testCases[i])
		})
	}
}
