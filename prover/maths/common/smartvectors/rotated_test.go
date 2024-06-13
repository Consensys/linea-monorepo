package smartvectors_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func TestRotatedSubVector(t *testing.T) {

	size := 16
	original := vector.Rand(size)
	myVec := smartvectors.NewRegular(original)

	for offset := 0; offset < size; offset++ {
		rotated := smartvectors.NewRotated(*myVec, offset)
		for start := 0; start < size; start++ {
			for stop := start + 1; stop <= size; stop++ {

				sub := rotated.SubVector(start, stop)
				recovered := sub.Get(0)
				require.Equal(t, recovered.String(), original[utils.PositiveMod(start+offset, size)].String())

			}
		}
	}

}

func TestRotatedWriteInSlice(t *testing.T) {

	size := 16
	original := vector.Rand(size)
	myVec := smartvectors.NewRegular(original)

	for offset := 0; offset < size; offset++ {
		rotated := smartvectors.NewRotated(*myVec, offset)
		written := make([]field.Element, size)
		rotated.WriteInSlice(written)

		for i := range written {
			require.Equal(t, written[i].String(), original[utils.PositiveMod(i+offset, size)].String())
		}
	}
}
