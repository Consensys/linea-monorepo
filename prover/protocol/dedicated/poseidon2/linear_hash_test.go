package poseidon2_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	linhash "github.com/consensys/linea-monorepo/prover/protocol/dedicated/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

const blockSize = 8

func TestLinearHash(t *testing.T) {

	var tohash [blockSize]ifaces.Column
	var expectedhash [blockSize]ifaces.Column

	testcases := []struct {
		colSize int
		numhash int
	}{
		{colSize: 16, numhash: 4},  // Full Chunks
		{colSize: 20, numhash: 16}, // Partial last chunk
		{colSize: 31, numhash: 8},  // Partial last chunk
	}

	for _, tc := range testcases {

		colChunks := (tc.colSize + blockSize - 1) / blockSize
		numhash := tc.numhash
		colSize := tc.colSize
		numRowLarge := utils.NextPowerOfTwo(colChunks * numhash)
		numRowSmall := utils.NextPowerOfTwo(numhash)

		define := func(b *wizard.Builder) {
			for i := 0; i < blockSize; i++ {
				tohash[i] = b.RegisterCommit(ifaces.ColIDf("TOHASH_%v", i), numRowLarge)

				expectedhash[i] = b.RegisterCommit(ifaces.ColIDf("HASHED_%v", i), numRowSmall)
			}
			linhash.CheckLinearHash(b.CompiledIOP, "test", tohash, colChunks, numhash, expectedhash)
		}

		prove := func(run *wizard.ProverRuntime) {
			var ex, th [blockSize][]field.Element

			for i := 0; i < blockSize; i++ {
				ex[i] = make([]field.Element, 0, numhash)
				th[i] = make([]field.Element, 0, colChunks*numhash)
			}

			for i := 0; i < numhash; i++ {
				// generate a segment at random, hash it
				// and append the segment to th and the result
				// to ex
				segment := vector.Rand(colSize)
				y := poseidon2.Poseidon2Sponge(segment)

				for j := 0; j < blockSize; j++ {
					ex[j] = append(ex[j], y[j])

					// Allocate segments to TOHASH columns
					completeChunks := colSize / blockSize
					for k := 0; k < completeChunks; k++ {
						th[j] = append(th[j], segment[k*blockSize+j])
					}

					lastChunkElements := colSize % blockSize
					lastChunkPadding := 0
					if lastChunkElements > 0 {
						lastChunkPadding = blockSize - lastChunkElements
					}
					if lastChunkElements > 0 {
						k := completeChunks
						if j < lastChunkPadding {
							// Left padding
							th[j] = append(th[j], field.Zero())
						} else {
							// Actual data
							actualIdx := k*blockSize + (j - lastChunkPadding)
							th[j] = append(th[j], segment[actualIdx])
						}
					}

				}
			}

			var exSV, thSV [blockSize]smartvectors.SmartVector
			for i := 0; i < blockSize; i++ {
				thSV[i] = smartvectors.RightZeroPadded(th[i], numRowLarge)
				exSV[i] = smartvectors.RightZeroPadded(ex[i], numRowSmall)

				// And assign them
				run.AssignColumn(tohash[i].GetColID(), thSV[i])
				run.AssignColumn(expectedhash[i].GetColID(), exSV[i])

			}

		}

		comp := wizard.Compile(define, dummy.Compile)
		proof := wizard.Prove(comp, prove)
		require.NoError(t, wizard.Verify(comp, proof))
	}

}
func getSegmentElement(segment []field.Element, k, j, blockSize, colSize int) field.Element {
	idx := k*blockSize + j
	if idx < colSize {
		return segment[idx]
	}
	return field.Zero()
}
