package vortex

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

func TestLinearCombinationStreamingMatchesBaseline(t *testing.T) {
	const (
		numCols    = 512
		rate       = 2
		numEncoded = numCols * rate
	)
	rsParams := reedsolomon.NewRsParams(numCols, rate)

	testCases := []struct {
		name            string
		numEncodedRows  int // rows already encoded (NoSIS)
		numOriginalRows int // rows to re-encode on the fly (SIS L2)
	}{
		{"all_encoded", 20, 0},
		{"all_original", 0, 20},
		{"mixed_5_15", 5, 15},
		{"mixed_10_10", 10, 10},
		{"single_original", 0, 1},
		{"single_encoded", 1, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			totalRows := tc.numEncodedRows + tc.numOriginalRows

			// Create original rows
			originalRows := make([]smartvectors.SmartVector, totalRows)
			for i := 0; i < totalRows; i++ {
				vec := make([]field.Element, numCols)
				for j := range vec {
					vec[j] = field.NewElement(uint64((i+1)*(j+1) + 7))
				}
				originalRows[i] = smartvectors.NewRegular(vec)
			}

			// RS-encode all rows for the reference
			allEncoded := make([]smartvectors.SmartVector, totalRows)
			for i := range originalRows {
				allEncoded[i] = rsParams.RsEncodeBase(originalRows[i])
			}

			randomCoin := fext.NewFromUint(42, 17, 3, 99)

			// Reference: LinearCombination with all encoded rows
			refProof := &OpeningProof{}
			LinearCombination(refProof, allEncoded, randomCoin)

			// Split: first numEncodedRows are pre-encoded, rest are original
			encodedPart := allEncoded[:tc.numEncodedRows]
			originalPart := originalRows[tc.numEncodedRows:]

			// Streaming: encoded + original
			streamProof := &OpeningProof{}
			LinearCombinationStreaming(streamProof, encodedPart, originalPart, rsParams, randomCoin)

			// Compare
			require.Equal(t, refProof.LinearCombination.Len(), streamProof.LinearCombination.Len(),
				"output length mismatch")
			for i := 0; i < refProof.LinearCombination.Len(); i++ {
				refVal := refProof.LinearCombination.GetExt(i)
				streamVal := streamProof.LinearCombination.GetExt(i)
				require.Equal(t, refVal, streamVal,
					"mismatch at column %d", i)
			}
		})
	}
}
