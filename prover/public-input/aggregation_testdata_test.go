package public_input

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AggregatedProofJSON matches the JSON structure of aggregatedProof files
type AggregatedProofJSON struct {
	FinalShnarf                             string   `json:"finalShnarf"`
	ParentAggregationFinalShnarf            string   `json:"parentAggregationFinalShnarf"`
	AggregatedProof                         string   `json:"aggregatedProof"`
	AggregatedProverVersion                 string   `json:"aggregatedProverVersion"`
	AggregatedVerifierIndex                 int      `json:"aggregatedVerifierIndex"`
	AggregatedProofPublicInput              string   `json:"aggregatedProofPublicInput"`
	DataHashes                              []string `json:"dataHashes"`
	DataParentHash                          string   `json:"dataParentHash"`
	FinalStateRootHash                      string   `json:"finalStateRootHash"`
	ParentStateRootHash                     string   `json:"parentStateRootHash"`
	ParentAggregationLastBlockTimestamp     uint     `json:"parentAggregationLastBlockTimestamp"`
	LastFinalizedBlockNumber                uint     `json:"lastFinalizedBlockNumber"`
	FinalTimestamp                          uint     `json:"finalTimestamp"`
	FinalBlockNumber                        uint     `json:"finalBlockNumber"`
	LastFinalizedL1RollingHash              string   `json:"lastFinalizedL1RollingHash"`
	L1RollingHash                           string   `json:"l1RollingHash"`
	LastFinalizedL1RollingHashMessageNumber uint     `json:"lastFinalizedL1RollingHashMessageNumber"`
	L1RollingHashMessageNumber              uint     `json:"l1RollingHashMessageNumber"`
	FinalFtxRollingHash                     string   `json:"finalFtxRollingHash"`
	ParentAggregationFtxRollingHash         string   `json:"parentAggregationFtxRollingHash"`
	FinalFtxNumber                          uint     `json:"finalFtxNumber"`
	ParentAggregationFtxNumber              uint     `json:"parentAggregationFtxNumber"`
	L2MerkleRoots                           []string `json:"l2MerkleRoots"`
	L2MerkleTreesDepth                      int      `json:"l2MerkleTreesDepth"`
	L2MessagingBlocksOffsets                string   `json:"l2MessagingBlocksOffsets"`
	FinalBlockHash                          string   `json:"finalBlockHash"`
	ParentAggregationBlockHash              string   `json:"parentAggregationBlockHash"`
	ChainID                                 uint64   `json:"chainID"`
	BaseFee                                 uint64   `json:"baseFee"`
	CoinBase                                string   `json:"coinBase"`
	L2MessageServiceAddr                    string   `json:"l2MessageServiceAddr"`
	IsAllowedCircuitID                      uint64   `json:"isAllowedCircuitID"`
	FilteredAddresses                       []string `json:"filteredAddresses"`
}

// ToAggregation converts JSON data to Aggregation struct for public input computation
func (j *AggregatedProofJSON) ToAggregation() Aggregation {
	var coinBase types.EthAddress
	if coinBaseBytes, err := utils.HexDecodeString(j.CoinBase); err == nil {
		copy(coinBase[:], coinBaseBytes)
	}

	var l2MessageServiceAddr types.EthAddress
	if l2MsgSvcBytes, err := utils.HexDecodeString(j.L2MessageServiceAddr); err == nil {
		copy(l2MessageServiceAddr[:], l2MsgSvcBytes)
	}

	var filteredAddresses = make([]types.EthAddress, len(j.FilteredAddresses))
	for i, addr := range j.FilteredAddresses {
		var filteredAddr types.EthAddress
		if addrBytes, err := utils.HexDecodeString(addr); err == nil {
			copy(filteredAddr[:], addrBytes)
		}
		filteredAddresses[i] = filteredAddr
	}
	return Aggregation{
		FinalShnarf:                             j.FinalShnarf,
		ParentAggregationFinalShnarf:            j.ParentAggregationFinalShnarf,
		ParentStateRootHash:                     j.ParentStateRootHash,
		ParentAggregationLastBlockTimestamp:     j.ParentAggregationLastBlockTimestamp,
		FinalTimestamp:                          j.FinalTimestamp,
		LastFinalizedBlockNumber:                j.LastFinalizedBlockNumber,
		FinalBlockNumber:                        j.FinalBlockNumber,
		LastFinalizedL1RollingHash:              j.LastFinalizedL1RollingHash,
		L1RollingHash:                           j.L1RollingHash,
		LastFinalizedL1RollingHashMessageNumber: j.LastFinalizedL1RollingHashMessageNumber,
		L1RollingHashMessageNumber:              j.L1RollingHashMessageNumber,
		LastFinalizedFtxRollingHash:             j.ParentAggregationFtxRollingHash,
		FinalFtxRollingHash:                     j.FinalFtxRollingHash,
		LastFinalizedFtxNumber:                  j.ParentAggregationFtxNumber,
		FinalFtxNumber:                          j.FinalFtxNumber,
		L2MsgRootHashes:                         j.L2MerkleRoots,
		L2MsgMerkleTreeDepth:                    j.L2MerkleTreesDepth,
		FinalBlockHash:                          j.FinalBlockHash,
		ParentAggregationBlockHash:              j.ParentAggregationBlockHash,
		ChainID:                                 j.ChainID,
		BaseFee:                                 j.BaseFee,
		CoinBase:                                coinBase,
		L2MessageServiceAddr:                    l2MessageServiceAddr,
		FilteredAddresses:                       filteredAddresses,
	}

}

func TestAggregationPublicInputFromTestData(t *testing.T) {
	// Get the workspace root by finding go.mod
	// The test data is in contracts/test/hardhat/_testData relative to workspace root
	workspaceRoot := findWorkspaceRoot(t)
	testDataBase := filepath.Join(workspaceRoot, "contracts/test/hardhat/_testData")

	testCases := []struct {
		name     string
		jsonPath string
	}{
		{
			name:     "EIP4844/multipleProofs/aggregatedProof-1-81",
			jsonPath: "compressedDataEip4844/multipleProofs/aggregatedProof-1-81.json",
		},
		{
			name:     "EIP4844/multipleProofs/aggregatedProof-82-139",
			jsonPath: "compressedDataEip4844/multipleProofs/aggregatedProof-82-139.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullPath := filepath.Join(testDataBase, tc.jsonPath)

			// Read the JSON file
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Skipf("Skipping %s: file not found at %s", tc.name, fullPath)
				return
			}

			// Parse the JSON
			var proofData AggregatedProofJSON
			require.NoError(t, json.Unmarshal(data, &proofData), "Failed to parse JSON")

			// Convert to Aggregation struct
			agg := proofData.ToAggregation()

			// Compute the public input
			computedPI := agg.GetPublicInputHex()
			// from the json file, generated by the test generator.
			expectedPI := proofData.AggregatedProofPublicInput

			// Log detailed info
			t.Logf("File: %s", tc.jsonPath)
			t.Logf("FinalFtxRollingHash: %s", proofData.FinalFtxRollingHash)
			t.Logf("ParentAggregationFtxRollingHash: %s", proofData.ParentAggregationFtxRollingHash)
			t.Logf("FinalFtxNumber: %d", proofData.FinalFtxNumber)
			t.Logf("ParentAggregationFtxNumber: %d", proofData.ParentAggregationFtxNumber)
			t.Logf("Expected PI: %s", expectedPI)
			t.Logf("Computed PI: %s", computedPI)

			// Assert they match
			assert.Equal(t, expectedPI, computedPI,
				"Public input mismatch for %s\nExpected: %s\nComputed: %s",
				tc.name, expectedPI, computedPI)
		})
	}
}

// findWorkspaceRoot walks up directories to find the workspace root (contains go.mod for prover)
func findWorkspaceRoot(t *testing.T) string {
	// Start from current working directory
	dir, err := os.Getwd()
	require.NoError(t, err)

	// Walk up to find the linea-monorepo root (has contracts/ directory)
	for {
		contractsPath := filepath.Join(dir, "contracts")
		if _, err := os.Stat(contractsPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find workspace root")
		}
		dir = parent
	}
}
