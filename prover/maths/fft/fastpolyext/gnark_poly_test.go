package fastpolyext_test

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

func TestGnarkInterpolate(t *testing.T) {

	testCases := [][]fext.Element{
		vectorext.ForTest(0, 0, 0, 0),
		vectorext.ForTest(1, 0, 1, 0),
		vectorext.ForTest(1, 1, 1, 1),
	}

	for i := range testCases {

		t.Run(fmt.Sprintf("test-cases-%v", i), func(t *testing.T) {

			def := func(api frontend.API) error {
				x := fext.NewElement(42, 0, 0, 0)
				vec := vectorext.IntoGnarkAssignment(testCases[i])
				expectedY := fastpolyext.Interpolate(testCases[i], x)
				fmt.Print("expectedY=", expectedY)

				computedY := fastpolyext.InterpolateGnark(api, vec, gnarkfext.FromValue(x))
				fmt.Print("computedY=", computedY)
				computedY.AssertIsEqual(api, gnarkfext.FromValue(expectedY))
				return nil
			}

			gnarkutil.AssertCircuitSolved(t, def)
		})
	}

}
