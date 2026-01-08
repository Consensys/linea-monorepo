package v2

import (
	"testing"

	gchash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/config"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFPIConsistency(t *testing.T) {
	fpi := FunctionalPublicInput{
		Y:              [2][16]byte{{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}, {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}},
		SnarkHash:      []byte{6},
		Eip4844Enabled: true,
		BatchSums:      make([]public_input.ExecDataChecksum, 3),
	}

	var (
		data types.Bytes32
		err  error
	)
	hsh := gchash.POSEIDON2_BLS12_377.New()
	for i := range fpi.BatchSums {
		data[len(data)-1] = byte(i)
		fpi.BatchSums[i], err = public_input.NewExecDataChecksum(data[:])
		hsh.Write(fpi.BatchSums[i].Hash[:])
		require.NoError(t, err)
	}
	copy(fpi.AllBatchesSum[:], hsh.Sum(nil))

	const maxNbBatches = 5

	fpi.X[0], fpi.X[31] = 1<<5, 2
	sum, err := fpi.Sum()
	assert.NoError(t, err)
	sfpi, err := fpi.ToSnarkType(maxNbBatches)
	assert.NoError(t, err)

	t.Run("3-batches", test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
		require.NoError(t, sfpi.Check(api, config.CircuitSizes{
			MaxNbBatches: maxNbBatches,
		}))
		return []frontend.Variable{sfpi.Sum(api)}
	}, sum))
}
