package v1

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/internal/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestFPIConsistency(t *testing.T) {
	fpi := FunctionalPublicInput{
		Y:              [2][]byte{{3}, {4}},
		SnarkHash:      []byte{6},
		Eip4844Enabled: true,
		BatchSums:      [][]byte{{7}, {8}, {9}},
	}
	fpi.X[0], fpi.X[31] = 1<<5, 2
	sum, err := fpi.Sum()
	assert.NoError(t, err)
	sfpi, err := fpi.ToSnarkType()
	assert.NoError(t, err)

	t.Run("3-batches", test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		hsh, err := mimc.NewMiMC(api)
		assert.NoError(t, err)
		return []frontend.Variable{sfpi.Sum(api, &hsh)}
	}, sum))
}
