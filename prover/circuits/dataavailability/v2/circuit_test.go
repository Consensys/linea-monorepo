package v2

import (
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestFPIConsistency(t *testing.T) {
	fpi := FunctionalPublicInput{
		Y:              [2][16]byte{{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}, {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}},
		SnarkHash:      []byte{6},
		Eip4844Enabled: true,
		BatchSums: []BatchSums{
			{Native: []byte{7}, SmallField: []byte{8}, Total: []byte{9}},
			{Native: []byte{10}, SmallField: []byte{11}, Total: []byte{12}},
			{Native: []byte{13}, SmallField: []byte{14}, Total: []byte{15}},
		},
	}
	fpi.X[0], fpi.X[31] = 1<<5, 2
	sum, err := fpi.Sum()
	assert.NoError(t, err)
	sfpi, err := fpi.ToSnarkType(5)
	assert.NoError(t, err)

	t.Run("3-batches", test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		return []frontend.Variable{sfpi.Sum(api)}
	}, sum))
}
