package internal

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMiMCMerkleDamgardConstruction(t *testing.T) {
	// See if H(x,y) = C(C(0,x),y) as expected
	var x, y fr.Element
	_, err := x.SetRandom()
	require.NoError(t, err)
	_, err = y.SetRandom()
	require.NoError(t, err)

	test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		h, err := NewMiMCWithCompressionFunction(api)
		require.NoError(t, err)

		h.Write(x, y)
		standard := h.Sum()

		state := h.Compress(api, 0, x)
		state = h.Compress(api, state, y)

		api.AssertIsEqual(standard, state)

		return nil
	})(t)
}
