package poseidon2_test

import (
	"fmt"
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
		period  int
		numhash int
	}{
		{period: 2, numhash: 5},
		// {period: 3, numhash: 16},
		// {period: 4, numhash: 14},
		// {period: 5, numhash: 17},
	}

	for _, tc := range testcases {

		period := tc.period
		numhash := tc.numhash
		numRowLarge := utils.NextPowerOfTwo(period * numhash)
		numRowSmall := utils.NextPowerOfTwo(numhash)

		define := func(b *wizard.Builder) {
			for i := 0; i < blockSize; i++ {
				tohash[i] = b.RegisterCommit(ifaces.ColIDf("TOHASH_%v", i), numRowLarge)

				expectedhash[i] = b.RegisterCommit(ifaces.ColIDf("HASHED_%v", i), numRowSmall)
			}
			linhash.CheckLinearHash(b.CompiledIOP, "test", tohash, period, numhash, expectedhash)
		}

		prove := func(run *wizard.ProverRuntime) {
			var ex, th [blockSize][]field.Element

			for i := 0; i < blockSize; i++ {
				ex[i] = make([]field.Element, 0, numhash)
				th[i] = make([]field.Element, 0, numhash*period)
			}

			for i := 0; i < numhash; i++ {
				// generate a segment at random, hash it
				// and append the segment to th and the result
				// to ex
				segment := vector.Rand(period * blockSize)
				y := poseidon2.Poseidon2HashVecElement(segment)

				fmt.Printf(" hashes to %v\n", vector.Prettify(y[:]))
				for j := 0; j < blockSize; j++ {
					ex[j] = append(ex[j], y[j])
					for i := 0; i < period; i++ {
						th[j] = append(th[j], segment[i*blockSize+j])
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
