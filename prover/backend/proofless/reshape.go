package proofless

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type ProverOutput struct {
	BlocksData          []RawBlockData `json:"blocksData"`
	ProverMode          int            `json:"proverMode"`
	ParentStateRootHash string         `json:"parentStateRootHash"`
	Version             string         `json:"version"`
	FirstBlockNumber    int            `json:"firstBlockNumber"`
	Proof               string         `json:"proof"`
	DebugData           DebugData      `json:"debugData"`
}

type RawBlockData struct {
	RootHash               string   `json:"rootHash"`
	RlpEncodedTransactions []string `json:"rlpEncodedTransactions"`
	Timestamp              int      `json:"timestamp"`
	L2ToL1MsgHashes        []string `json:"l2ToL1MsgHashes"`
	FromAddresses          string   `json:"fromAddresses"`
	BatchReceptionIndices  []int    `json:"batchReceptionIndices"`
}

type FormattedBlockData struct {
	BlockRootHash         string   `json:"blockRootHash"`
	Transactions          []string `json:"transactions"`
	L2BlockTimestamp      int      `json:"l2BlockTimestamp"`
	L2ToL1MsgHashes       []string `json:"l2ToL1MsgHashes"`
	FromAddresses         string   `json:"fromAddresses"`
	BatchReceptionIndices []int    `json:"batchReceptionIndices"`
}

type DebugBlock struct {
	TxHashes        []string `json:"txHashes"`
	HashOfTxHashes  string   `json:"hashOfTxHashes"`
	LogHashes       []string `json:"logHashes"`
	HashOfLogHashes string   `json:"hashOfLogHashes"`
	HashOfPositions string   `json:"hashOfPositions"`
	HashForBlock    string   `json:"HashForBlock"`
}

type DebugData struct {
	Blocks           []DebugBlock `json:"blocks"`
	HashForAllBlocks string       `json:"hashForAllBlocks"`
	HashOfRootHashes string       `json:"hashOfRootHashes"`
	TimestampHashes  string       `json:"timestampHashes"`
	FinalHash        string       `json:"finalHash"`
}

func ReshapeProverOutput(proverOutputFile string) {

	po, err := getProverOutputData(proverOutputFile)
	if err != nil {
		utils.Panic("Could not get prover output file : %s", err)
	}

	// generate contract input data
	var inputData [][]interface{}
	for _, b := range po["blocks"].([]FormattedBlockData) {
		inputData = append(inputData, []interface{}{
			b.BlockRootHash,
			b.L2BlockTimestamp,
			b.Transactions,
			b.L2ToL1MsgHashes,
			b.FromAddresses,
			b.BatchReceptionIndices,
		})
	}

	inputFilePath := filepath.Join(proverOutputFile + ".inputs")

	inputDataJSON, err := json.Marshal(inputData)
	if err != nil {
		utils.Panic("Could not marshal reshaped prover output data : %s", err)
	}

	err = os.WriteFile(inputFilePath, inputDataJSON, os.ModePerm)
	if err != nil {
		utils.Panic("Could not write reshaped prover output file : %s", err)
	}
}

func getProverOutputData(filename string) (map[string]interface{}, error) {
	testFilePath := filepath.Join(filename)

	proverOutputData, err := os.ReadFile(testFilePath)
	if err != nil {
		utils.Panic("Could not read prover output file : %s", err)
	}

	var proverOutput ProverOutput
	err = json.Unmarshal(proverOutputData, &proverOutput)
	if err != nil {
		utils.Panic("Could not unmarshal prover output file: %s", err)
	}

	formattedBlocks := make([]FormattedBlockData, len(proverOutput.BlocksData))
	for i, block := range proverOutput.BlocksData {
		formattedBlocks[i] = FormattedBlockData{
			BlockRootHash:         block.RootHash,
			Transactions:          block.RlpEncodedTransactions,
			L2BlockTimestamp:      block.Timestamp,
			L2ToL1MsgHashes:       block.L2ToL1MsgHashes,
			FromAddresses:         block.FromAddresses,
			BatchReceptionIndices: block.BatchReceptionIndices,
		}
	}

	result := map[string]interface{}{
		"blocks":              formattedBlocks,
		"proverMode":          proverOutput.ProverMode,
		"parentStateRootHash": proverOutput.ParentStateRootHash,
		"version":             proverOutput.Version,
		"firstBlockNumber":    proverOutput.FirstBlockNumber,
		"proof":               proverOutput.Proof,
		"debugData":           proverOutput.DebugData,
	}

	return result, nil
}
