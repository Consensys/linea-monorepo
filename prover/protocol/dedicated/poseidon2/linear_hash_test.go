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

	var tohash ifaces.Column
	var expectedhash [blockSize]ifaces.Column

	testcases := []struct {
		period  int
		numhash int
	}{
		{period: 2, numhash: 4},
		{period: 3, numhash: 16},
		// {period: 4, numhash: 14},
		// {period: 5, numhash: 17},
	}

	for _, tc := range testcases {

		period := tc.period
		numhash := tc.numhash
		numRowLarge := utils.NextPowerOfTwo(blockSize * period * numhash)
		numRowSmall := utils.NextPowerOfTwo(numhash)

		define := func(b *wizard.Builder) {
			tohash = b.RegisterCommit("TOHASH", numRowLarge)
			for i := 0; i < blockSize; i++ {
				expectedhash[i] = b.RegisterCommit(ifaces.ColIDf("HASHED_%v", i), numRowSmall)
			}
			linhash.CheckLinearHash(b.CompiledIOP, "test", tohash, period, numhash, expectedhash)
		}

		prove := func(run *wizard.ProverRuntime) {
			th := make([]field.Element, 0, blockSize*numhash*period)

			var ex [blockSize][]field.Element
			for i := 0; i < blockSize; i++ {
				ex[i] = make([]field.Element, 0, numhash)
			}

			for i := 0; i < numhash; i++ {
				// generate a segment at random, hash it
				// and append the segment to th and the result
				// to ex
				segment := vector.Rand(period * blockSize)
				y := poseidon2.Poseidon2Sponge(segment)

				th = append(th, segment...)
				for j := 0; j < blockSize; j++ {
					ex[j] = append(ex[j], y[j])
				}
			}
			thSV := smartvectors.RightZeroPadded(th, numRowLarge)
			run.AssignColumn(tohash.GetColID(), thSV)

			var exSV [blockSize]smartvectors.SmartVector
			for i := 0; i < blockSize; i++ {
				exSV[i] = smartvectors.RightZeroPadded(ex[i], numRowSmall)

				// And assign them
				run.AssignColumn(expectedhash[i].GetColID(), exSV[i])

			}

		}

		comp := wizard.Compile(define, dummy.Compile)
		proof := wizard.Prove(comp, prove)
		require.NoError(t, wizard.Verify(comp, proof))
	}

}
