package jobadapter

import "github.com/LFDT-Lineth/lineth-monorepo/prover-ray/backend"

// aggregationResponse is the getZkRollupAggregationProofV1 success body. Unlike
// the rollup response it has no top-level programVk (the combined programVks set
// lives in publicInputs) and adds l2MessagingBlocksOffsets. Public inputs,
// proof, and revealed arrays are placeholders until the guest lands.
type aggregationResponse struct {
	ProverVersion            string                          `json:"proverVersion"`
	ProofHex                 string                          `json:"proof"`
	StartBlockNumber         uint64                          `json:"startBlockNumber"`
	PublicInputs             aggregationResponsePublicInputs `json:"publicInputs"`
	L2L1Roots                []string                        `json:"l2L1Roots"`
	FilteredAddresses        []string                        `json:"filteredAddresses"`
	L2MessagingBlocksOffsets []uint64                        `json:"l2MessagingBlocksOffsets"`
}

// aggregationResponsePublicInputs is the 20-field aggregation PI tuple (§2.4).
type aggregationResponsePublicInputs struct {
	EndBlockNumber                           uint64   `json:"endBlockNumber"`
	EndBlockTimestamp                        uint64   `json:"endBlockTimestamp"`
	L2L1BridgeTransactionTree                string   `json:"l2L1BridgeTransactionTree"`
	ParentL1L2BridgeRollingHash              string   `json:"parentL1L2BridgeRollingHash"`
	ParentL1L2BridgeRollingHashMessageNumber uint64   `json:"parentL1L2BridgeRollingHashMessageNumber"`
	EndL1L2BridgeRollingHash                 string   `json:"endL1L2BridgeRollingHash"`
	EndL1L2BridgeRollingHashMessageNumber    uint64   `json:"endL1L2BridgeRollingHashMessageNumber"`
	DynamicChainConfigHash                   string   `json:"dynamicChainConfigHash"`
	ParentFtxRollingHash                     string   `json:"parentFtxRollingHash"`
	ParentFtxNumber                          uint64   `json:"parentFtxNumber"`
	EndFtxRollingHash                        string   `json:"endFtxRollingHash"`
	EndProcessedFtxNumber                    uint64   `json:"endProcessedFtxNumber"`
	FilteredAddressesHash                    string   `json:"filteredAddressesHash"`
	ParentDataRollingHash                    string   `json:"parentDataRollingHash"`
	EndDataRollingHash                       string   `json:"endDataRollingHash"`
	ParentBlockHash                          string   `json:"parentBlockHash"`
	EndBlockHash                             string   `json:"endBlockHash"`
	StartOffset                              uint64   `json:"startOffset"`
	EndOffset                                uint64   `json:"endOffset"`
	ProgramVks                               []string `json:"programVks"`
}

func newAggregationResponse(
	result backend.Result, startBlockNumber uint64, proverVersion string,
) aggregationResponse {
	return aggregationResponse{
		ProverVersion:            proverVersion,
		ProofHex:                 hexBytes(result.ProofBytes),
		StartBlockNumber:         startBlockNumber,
		PublicInputs:             placeholderAggregationPublicInputs(),
		L2L1Roots:                []string{},
		FilteredAddresses:        []string{},
		L2MessagingBlocksOffsets: []uint64{},
	}
}

// placeholderAggregationPublicInputs is a zero-valued but schema-valid
// aggregation PI. Populated from backend.Result once extraction exists.
func placeholderAggregationPublicInputs() aggregationResponsePublicInputs {
	zero := hexHash([32]byte{})
	return aggregationResponsePublicInputs{
		L2L1BridgeTransactionTree:   zero,
		ParentL1L2BridgeRollingHash: zero,
		EndL1L2BridgeRollingHash:    zero,
		DynamicChainConfigHash:      zero,
		ParentFtxRollingHash:        zero,
		EndFtxRollingHash:           zero,
		FilteredAddressesHash:       zero,
		ParentDataRollingHash:       zero,
		EndDataRollingHash:          zero,
		ParentBlockHash:             zero,
		EndBlockHash:                zero,
		ProgramVks:                  []string{},
	}
}
