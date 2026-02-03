package pi_interconnection

import (
	"encoding/json"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/stretchr/testify/assert"
)

// TODO create test cases with multiple shnarfs
func TestSingleShnarf(t *testing.T) {
	const (
		maxNbShnarf     = 3
		blocksPerShnarf = 2
		maxNbKeccakF    = blocksPerShnarf * maxNbShnarf
	)

	jsons, err := utils.ReadAllJsonFiles("../../../testdata/prover/blob-compression/responses")
	assert.NoError(t, err)

	for filename, jsonString := range jsons {
		var r blobsubmission.Response
		assert.NoError(t, json.Unmarshal(jsonString, &r))
		t.Run(filename, func(t *testing.T) {
			c := testShnarfCircuit{
				Shnarfs:      make([]ShnarfIteration, maxNbShnarf),
				maxNbKeccakF: maxNbKeccakF,
				Results:      make([][32]frontend.Variable, maxNbShnarf),
			}
			a := testShnarfCircuit{
				Shnarfs:   make([]ShnarfIteration, maxNbShnarf),
				NbShnarfs: 1,
				Results:   make([][32]frontend.Variable, maxNbShnarf),
			}

			it := &a.Shnarfs[0]
			copyHexIntoVarArray(t, &a.Prev, r.PrevShnarf)
			copyHexIntoVarArray(t, &a.Final, r.ExpectedShnarf)
			copyHexIntoVarArray(t, &it.EvaluationPointBytes, r.ExpectedX)
			copyHexIntoVarArray(t, &it.EvaluationClaimBytes, r.ExpectedY)
			copyHexIntoVarArray(t, &it.NewStateRootHash, r.FinalStateRootHash)
			copyHexIntoVarArray(t, &it.BlobDataSnarkHash, r.SnarkHash)
			copyHexIntoVarArray(t, &a.Results[0], r.ExpectedShnarf)

			for i := 1; i < len(a.Shnarfs); i++ {
				a.Shnarfs[i].SetZero()
				for j := range a.Results[i] {
					a.Results[i][j] = i*32 + j
				}
			}

			assert.NoError(t, test.IsSolved(&c, &a, ecc.BLS12_377.ScalarField()))

		})
	}
}

type testShnarfCircuit struct {
	Shnarfs      []ShnarfIteration
	Results      [][32]frontend.Variable
	Prev, Final  [32]frontend.Variable
	NbShnarfs    frontend.Variable
	maxNbKeccakF int
}

func (c *testShnarfCircuit) Define(api frontend.API) error {
	hasher := keccak.NewHasher(api, c.maxNbKeccakF)

	shnarfs := ComputeShnarfs(hasher, c.Prev, c.Shnarfs)
	r := internal.NewRange(api, c.NbShnarfs, len(c.Shnarfs))
	final := r.LastArray32(shnarfs)
	r.AssertArrays32Equal(shnarfs, c.Results)
	internal.AssertSliceEquals(api, c.Final[:], final[:])

	return nil
}

func copyHexIntoVarArray(t *testing.T, dst *[32]frontend.Variable, src string) {
	b, err := utils.HexDecodeString(src)
	assert.NoError(t, err)
	assert.Equal(t, len(dst), len(b))
	for i := range b {
		dst[i] = b[i]
	}
}
