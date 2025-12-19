//go:build !fuzzlight

package poseidon2_koalabear

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func TestPoseidon2Hash(t *testing.T) {
	tests := []struct {
		name        string
		inputChunks [][]uint64 // Allows testing multiple Write calls
		expectedOut [8]uint64
	}{
		{
			name:        "Empty hash (default state)",
			inputChunks: [][]uint64{},
			expectedOut: [8]uint64{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "Single chunk input",
			inputChunks: [][]uint64{
				{0, 1, 2, 3, 4, 5, 6, 7},
			},
			expectedOut: [8]uint64{1402915949, 2083603674, 612464613, 903296317, 1327519169, 1646568100, 1786812904, 1917132066},
		},
		{
			name: "Two chunks input",
			inputChunks: [][]uint64{
				{20, 30, 40, 50, 60, 70, 80, 90},
				{120, 130, 140, 150, 160, 170, 180, 190},
			},
			expectedOut: [8]uint64{1735559108, 1813232198, 1226565528, 1795826654, 232717972, 522260852, 1109013755, 1232040797},
		},
		{
			name: "Three chunks input",
			inputChunks: [][]uint64{
				{20, 30, 40, 50, 60, 70, 80, 90},
				{120, 130, 140, 150, 160, 170, 180, 190},
				{320, 330, 340, 350, 360, 370, 380, 390},
			},
			expectedOut: [8]uint64{2089473715, 1269546008, 32120401, 1497392503, 1194364877, 1059604979, 1458772903, 705747974},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash := NewMDHasher()

			// Write input chunks sequentially
			for _, chunk := range tc.inputChunks {
				elems := uint64ArrayToElements(chunk)
				hash.WriteElements(elems[:]...)
			}

			hashValue := hash.SumElement()
			expectedElements := uint64ArrayToElements(tc.expectedOut[:])

			require.Equal(t, expectedElements, hashValue)
		})
	}
}

// Helper to convert uint64 slice to field.Octuplet
func uint64ArrayToElements(in []uint64) field.Octuplet {
	var out field.Octuplet
	for i := 0; i < BlockSize && i < len(in); i++ {
		out[i] = field.NewElement(in[i])
	}
	return out
}
