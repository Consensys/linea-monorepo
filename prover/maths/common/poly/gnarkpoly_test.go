package poly

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/stretchr/testify/require"
)

// type Eval struct {
// 	pol []frontend.Variable
// }

// func TestGnarkEval(t *testing.T) {

// }

func TestGnarkEvalAnyDomain(t *testing.T) {

	t.Run("single-variable", func(t *testing.T) {

		def := func(api frontend.API) error {
			var (
				domain                     = vector.IntoGnarkAssignment(vector.ForTest(0))
				x        frontend.Variable = 42
				expected                   = vector.IntoGnarkAssignment(vector.ForTest(1))
				res                        = EvaluateLagrangeAnyDomainGnark(api, domain, x)
			)

			require.Len(t, res, len(expected))
			for i := range expected {
				api.AssertIsEqual(expected[i], res[i])
			}

			return nil
		}

		gnarkutil.AssertCircuitSolved(t, def)
	})

	t.Run("multiple-variable", func(t *testing.T) {

		def := func(api frontend.API) error {
			var (
				domain                     = vector.IntoGnarkAssignment(vector.ForTest(0, 1))
				x        frontend.Variable = 42
				expected                   = vector.IntoGnarkAssignment(vector.ForTest(-41, 42))
				res                        = EvaluateLagrangeAnyDomainGnark(api, domain, x)
			)

			require.Len(t, res, len(expected))
			for i := range expected {
				api.AssertIsEqual(expected[i], res[i])
			}

			return nil
		}

		gnarkutil.AssertCircuitSolved(t, def)
	})

}
