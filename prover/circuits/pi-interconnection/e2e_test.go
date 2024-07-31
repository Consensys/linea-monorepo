//go:build !fuzzlight

package pi_interconnection_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	pi_interconnection "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection"
	pitesting "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/test_utils"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	blobtesting "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

// TODO test with random values instead of small ones

// some of the execution data are faked
func TestSingleBlockBlob(t *testing.T) {
	testPI(t, 103, pitesting.AssignSingleBlockBlob(t))
}

func TestSingleBlobBlobE2E(t *testing.T) {
	req := pitesting.AssignSingleBlockBlob(t)
	config := pi_interconnection.Config{
		MaxNbDecompression:   len(req.Decompressions),
		MaxNbExecution:       len(req.Executions),
		MaxNbKeccakF:         100,
		MaxNbMsgPerExecution: 1,
		L2MsgMerkleDepth:     5,
		L2MessageMaxNbMerkle: 1,
	}
	compiled, err := config.Compile(dummy.Compile)
	assert.NoError(t, err)

	a, err := compiled.Assign(req)
	assert.NoError(t, err)

	for _, gkrMimc := range []struct {
		use  bool
		prep string
	}{{false, "without"}, {true, "with"}} {
		t.Run(gkrMimc.prep+" gkrmimc", func(t *testing.T) {
			c := *compiled.Circuit
			c.UseGkrMimc = gkrMimc.use

			cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &c, frontend.WithCapacity(3_000_000))
			assert.NoError(t, err)

			w, err := frontend.NewWitness(&a, ecc.BLS12_377.ScalarField())
			assert.NoError(t, err)

			assert.NoError(t, cs.IsSolved(w, gkrmimc.SolverOpts(cs)...))
		})
	}
}

// some of the execution data are faked
func TestTinyTwoBatchBlob(t *testing.T) {
	blob := blobtesting.TinyTwoBatchBlob(t)

	execReq := []pi_interconnection.ExecutionRequest{{
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(3)},
		FinalStateRootHash:     internal.Uint64To32Bytes(4),
		FinalBlockNumber:       5,
		FinalBlockTimestamp:    6,
		FinalRollingHash:       internal.Uint64To32Bytes(7),
		FinalRollingHashNumber: 8,
	}, {
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(9)},
		FinalStateRootHash:     internal.Uint64To32Bytes(10),
		FinalBlockNumber:       11,
		FinalBlockTimestamp:    12,
		FinalRollingHash:       internal.Uint64To32Bytes(13),
		FinalRollingHashNumber: 14,
	}}

	blobReq := blobsubmission.Request{
		Eip4844Enabled:      true,
		CompressedData:      base64.StdEncoding.EncodeToString(blob),
		ParentStateRootHash: utils.FmtIntHex32Bytes(1),
		FinalStateRootHash:  utils.HexEncodeToString(execReq[1].FinalStateRootHash[:]),
		PrevShnarf:          utils.FmtIntHex32Bytes(2),
	}

	blobResp, err := blobsubmission.CraftResponse(&blobReq)
	assert.NoError(t, err)

	merkleRoots := aggregation.PackInMiniTrees(test_utils.BlocksToHex(execReq[0].L2MsgHashes, execReq[1].L2MsgHashes))

	req := pi_interconnection.Request{
		DecompDict:     blobtesting.GetDict(t),
		Decompressions: []blobsubmission.Response{*blobResp},
		Executions:     execReq,
		Aggregation: public_input.Aggregation{
			FinalShnarf:                             blobResp.ExpectedShnarf,
			ParentAggregationFinalShnarf:            blobReq.PrevShnarf,
			ParentStateRootHash:                     blobReq.ParentStateRootHash,
			ParentAggregationLastBlockTimestamp:     6,
			FinalTimestamp:                          uint(execReq[1].FinalBlockTimestamp),
			LastFinalizedBlockNumber:                5,
			FinalBlockNumber:                        uint(execReq[1].FinalBlockNumber),
			LastFinalizedL1RollingHash:              utils.FmtIntHex32Bytes(7),
			L1RollingHash:                           utils.HexEncodeToString(execReq[1].FinalRollingHash[:]),
			LastFinalizedL1RollingHashMessageNumber: 8,
			L1RollingHashMessageNumber:              uint(execReq[1].FinalRollingHashNumber),
			L2MsgRootHashes:                         merkleRoots,
			L2MsgMerkleTreeDepth:                    5,
		},
	}

	testPI(t, 100, req)
}

func TestTwoTwoBatchBlobs(t *testing.T) {
	blobs := blobtesting.ConsecutiveBlobs(t, 2, 2)

	execReq := []pi_interconnection.ExecutionRequest{{
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(3)},
		FinalStateRootHash:     internal.Uint64To32Bytes(4),
		FinalBlockNumber:       5,
		FinalBlockTimestamp:    6,
		FinalRollingHash:       internal.Uint64To32Bytes(7),
		FinalRollingHashNumber: 8,
	}, {
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(9)},
		FinalStateRootHash:     internal.Uint64To32Bytes(10),
		FinalBlockNumber:       11,
		FinalBlockTimestamp:    12,
		FinalRollingHash:       internal.Uint64To32Bytes(13),
		FinalRollingHashNumber: 14,
	}, {
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(15)},
		FinalStateRootHash:     internal.Uint64To32Bytes(16),
		FinalBlockNumber:       17,
		FinalBlockTimestamp:    18,
		FinalRollingHash:       internal.Uint64To32Bytes(19),
		FinalRollingHashNumber: 20,
	}, {
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(21)},
		FinalStateRootHash:     internal.Uint64To32Bytes(22),
		FinalBlockNumber:       23,
		FinalBlockTimestamp:    24,
		FinalRollingHash:       internal.Uint64To32Bytes(25),
		FinalRollingHashNumber: 26,
	}}

	blobReq0 := blobsubmission.Request{
		Eip4844Enabled:      true,
		CompressedData:      base64.StdEncoding.EncodeToString(blobs[0]),
		ParentStateRootHash: utils.FmtIntHex32Bytes(1),
		FinalStateRootHash:  utils.HexEncodeToString(execReq[1].FinalStateRootHash[:]),
		PrevShnarf:          utils.FmtIntHex32Bytes(2),
	}

	blobResp0, err := blobsubmission.CraftResponse(&blobReq0)
	assert.NoError(t, err)

	blobReq1 := blobsubmission.Request{
		Eip4844Enabled:      true,
		CompressedData:      base64.StdEncoding.EncodeToString(blobs[1]),
		ParentStateRootHash: blobReq0.FinalStateRootHash,
		FinalStateRootHash:  utils.HexEncodeToString(execReq[3].FinalStateRootHash[:]),
		PrevShnarf:          blobResp0.ExpectedShnarf,
	}

	blobResp1, err := blobsubmission.CraftResponse(&blobReq1)
	assert.NoError(t, err)

	merkleRoots := aggregation.PackInMiniTrees(test_utils.BlocksToHex(execReq[0].L2MsgHashes, execReq[1].L2MsgHashes, execReq[2].L2MsgHashes, execReq[3].L2MsgHashes))

	req := pi_interconnection.Request{
		DecompDict:     blobtesting.GetDict(t),
		Decompressions: []blobsubmission.Response{*blobResp0, *blobResp1},
		Executions:     execReq,
		Aggregation: public_input.Aggregation{
			FinalShnarf:                             blobResp1.ExpectedShnarf,
			ParentAggregationFinalShnarf:            blobReq0.PrevShnarf,
			ParentStateRootHash:                     blobReq0.ParentStateRootHash,
			ParentAggregationLastBlockTimestamp:     6,
			FinalTimestamp:                          uint(execReq[3].FinalBlockTimestamp),
			LastFinalizedBlockNumber:                5,
			FinalBlockNumber:                        uint(execReq[3].FinalBlockNumber),
			LastFinalizedL1RollingHash:              utils.FmtIntHex32Bytes(7),
			L1RollingHash:                           utils.HexEncodeToString(execReq[3].FinalRollingHash[:]),
			LastFinalizedL1RollingHashMessageNumber: 8,
			L1RollingHashMessageNumber:              uint(execReq[3].FinalRollingHashNumber),
			L2MsgRootHashes:                         merkleRoots,
			L2MsgMerkleTreeDepth:                    5,
		},
	}

	testPI(t, 101, req)
}

func testPI(t *testing.T, maxNbKeccakF int, req pi_interconnection.Request) {
	var slack [4]int
	for i := 0; i < 81; i++ {

		decomposeLittleEndian(t, slack[:], i, 3)

		config := pi_interconnection.Config{
			MaxNbDecompression:   len(req.Decompressions) + slack[0],
			MaxNbExecution:       len(req.Executions) + slack[1],
			MaxNbKeccakF:         maxNbKeccakF,
			MaxNbMsgPerExecution: 1 + slack[2],
			L2MsgMerkleDepth:     5,
			L2MessageMaxNbMerkle: 1 + slack[3],
		}

		t.Run(fmt.Sprintf("slack profile %v", slack), func(t *testing.T) {
			compiled, err := config.Compile(dummy.Compile)
			assert.NoError(t, err)

			a, err := compiled.Assign(req)
			assert.NoError(t, err)

			assert.NoError(t, test.IsSolved(compiled.Circuit, &a, ecc.BLS12_377.ScalarField()))
		})
	}
}

func decomposeLittleEndian(t *testing.T, digits []int, n, base int) {
	for i := range digits {
		digits[i] = n % base
		n /= base
	}
	assert.Zero(t, n)
}
