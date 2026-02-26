package test_utils

import (
	"encoding/base64"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	pi_interconnection "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	blobtesting "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"

	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
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

	execReq := public_input.Execution{
		L2MessageHashes:              [][32]byte{internal.Uint64To32Bytes(4)},
		InitialBlockTimestamp:        7,
		FinalStateRootHash:           finalStateRootHash,
		FinalBlockNumber:             9,
		FinalBlockTimestamp:          10,
		LastRollingHashUpdate:        internal.Uint64To32Bytes(11),
		LastRollingHashUpdateNumber:  9,
		FirstRollingHashUpdateNumber: 9,
		InitialBlockNumber:           6,
		InitialStateRootHash:         internal.Uint64To32Bytes(1),
	}

	prevFtxRollingHash := types.Bytes32FromHex("0x0123")
	txHash := types.FullBytes32FromHex("0x0ab0")

	ftxRollingHashBytes := ComputeFtxRollingHash(
		prevFtxRollingHash,
		common.Hash(txHash),
		9,
		types.DummyAddress(32),
	)
	ftxRollingHash := types.Bytes32(ftxRollingHashBytes)

	invalReq := public_input.Invalidity{
		TxHash:              common.Hash(txHash),
		TxNumber:            4,
		StateRootHash:       execReq.InitialStateRootHash,
		ExpectedBlockHeight: 9,
		FromAddress:         types.DummyAddress(32),
		FtxRollingHash:      ftxRollingHash,
	}

	merkleRoots := aggregation.PackInMiniTrees(test_utils.BlocksToHex(execReq.L2MessageHashes))

	agg := public_input.Aggregation{
		FinalShnarf:                             blobResp.ExpectedShnarf,
		ParentAggregationFinalShnarf:            blobReq.PrevShnarf,
		ParentStateRootHash:                     blobReq.ParentStateRootHash,
		ParentAggregationLastBlockTimestamp:     6,
		FinalTimestamp:                          uint(execReq.FinalBlockTimestamp),
		LastFinalizedBlockNumber:                5,
		FinalBlockNumber:                        uint(execReq.FinalBlockNumber),
		LastFinalizedL1RollingHash:              utils.FmtIntHex32Bytes(7),
		L1RollingHash:                           utils.HexEncodeToString(execReq.LastRollingHashUpdate[:]),
		LastFinalizedL1RollingHashMessageNumber: 8,
		L1RollingHashMessageNumber:              uint(execReq.LastRollingHashUpdateNumber),
		L2MsgRootHashes:                         merkleRoots,
		L2MsgMerkleTreeDepth:                    5,
		LastFinalizedFtxNumber:                  3,
		FinalFtxNumber:                          4,
		LastFinalizedFtxRollingHash:             utils.HexEncodeToString(prevFtxRollingHash[:]),
		FinalFtxRollingHash:                     utils.HexEncodeToString(ftxRollingHash[:]),
		// filtered addresses
		FilteredAddresses: make([]types.EthAddress, 0),
	}

	return pi_interconnection.Request{
		Decompressions: []blobsubmission.Response{*blobResp},
		Executions:     []public_input.Execution{execReq},
		Invalidity:     []public_input.Invalidity{invalReq},
		Aggregation:    agg,
	}
}
func ComputeFtxRollingHash(prevFtxRollingHash types.Bytes32, txHash common.Hash, expectedBlockHeight uint64, fromAddress types.EthAddress) []byte {
	mimc := mimc.NewMiMC()
	mimc.Write(prevFtxRollingHash[:])
	mimc.Write(txHash[:16])
	mimc.Write(txHash[16:])
	types.WriteInt64On32Bytes(mimc, int64(expectedBlockHeight))
	mimc.Write(fromAddress[:])
	return mimc.Sum(nil)
}
