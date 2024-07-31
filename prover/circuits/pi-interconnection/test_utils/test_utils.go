package test_utils

import (
	"encoding/base64"
	"github.com/consensys/zkevm-monorepo/prover/backend/aggregation"
	"github.com/consensys/zkevm-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal/test_utils"
	pi_interconnection "github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection"
	blobtesting "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1/test_utils"

	"github.com/consensys/zkevm-monorepo/prover/public-input"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssignSingleBlockBlob(t require.TestingT) pi_interconnection.Request {
	blob := blobtesting.SingleBlockBlob(t)

	finalStateRootHash := internal.Uint64To32Bytes(2)

	blobReq := blobsubmission.Request{
		Eip4844Enabled:      true,
		CompressedData:      base64.StdEncoding.EncodeToString(blob),
		ParentStateRootHash: utils.FmtIntHex32Bytes(1),
		FinalStateRootHash:  utils.HexEncodeToString(finalStateRootHash[:]),
		PrevShnarf:          utils.FmtIntHex32Bytes(3),
	}

	blobResp, err := blobsubmission.CraftResponse(&blobReq)
	assert.NoError(t, err)

	execReq := pi_interconnection.ExecutionRequest{
		L2MsgHashes:            [][32]byte{internal.Uint64To32Bytes(4)},
		FinalStateRootHash:     finalStateRootHash,
		FinalBlockNumber:       9,
		FinalBlockTimestamp:    10,
		FinalRollingHash:       internal.Uint64To32Bytes(11),
		FinalRollingHashNumber: 12,
	}

	merkleRoots := aggregation.PackInMiniTrees(test_utils.BlocksToHex(execReq.L2MsgHashes))

	return pi_interconnection.Request{
		DecompDict:     blobtesting.GetDict(t),
		Decompressions: []blobsubmission.Response{*blobResp},
		Executions:     []pi_interconnection.ExecutionRequest{execReq},
		Aggregation: public_input.Aggregation{
			FinalShnarf:                             blobResp.ExpectedShnarf,
			ParentAggregationFinalShnarf:            blobReq.PrevShnarf,
			ParentStateRootHash:                     blobReq.ParentStateRootHash,
			ParentAggregationLastBlockTimestamp:     6,
			FinalTimestamp:                          uint(execReq.FinalBlockTimestamp),
			LastFinalizedBlockNumber:                5,
			FinalBlockNumber:                        uint(execReq.FinalBlockNumber),
			LastFinalizedL1RollingHash:              utils.FmtIntHex32Bytes(7),
			L1RollingHash:                           utils.HexEncodeToString(execReq.FinalRollingHash[:]),
			LastFinalizedL1RollingHashMessageNumber: 8,
			L1RollingHashMessageNumber:              uint(execReq.FinalRollingHashNumber),
			L2MsgRootHashes:                         merkleRoots,
			L2MsgMerkleTreeDepth:                    5,
		},
	}
}
