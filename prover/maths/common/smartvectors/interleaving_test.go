//go:build !race

package smartvectors

import (
	"fmt"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/stretchr/testify/require"
)

func TestInterleaving(t *testing.T) {
	for i := 0; i < FUZZ_ITERATION; i++ {
		// We reuse the test case generator for linear combinations. We only
		// care about the first vector.
		builder := NewTestBuilder(i)
		tcase := builder.NewTestCaseForLinComb()

		success := t.Run(
			fmt.Sprintf("fuzzy-interleaving-%v", i),
			func(t *testing.T) {
				svecs := tcase.svecs
				nVecs := utils.NextPowerOfTwo(len(svecs))
				nVecs = utils.Max(2, nVecs)
				vecLen := svecs[0].Len()

				// Pad with constant vectors in order to achieve a power of two
				for i := len(svecs); i < nVecs; i++ {
					svecs = append(svecs, NewConstant(field.Zero(), svecs[0].Len()))
				}

				interleaved := Interleave(svecs...)

				for j := 0; j < nVecs; j++ {
					for k := 0; k < vecLen; k++ {
						actual := interleaved.Get(j + k*nVecs)
						expected := svecs[j].Get(k)
						require.Equal(t, expected.String(), actual.String())
					}
				}
			},
		)

		require.True(t, success)
	}
}
