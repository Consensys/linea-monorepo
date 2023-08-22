package coordinator

import (
	"encoding/json"

	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
)

// JSON schema of the message to return as an output of the prover
// Notations:
//   - N denotes the number of conflated blocks.
//   - Ti, denotes the number of transactions in the block i
//   - T denotes the total number of transactions (nimaltogether
//   - L denotes the total number of L2 to L1 logs in the conflated
type ProverOutput struct {

	// Proof in 0x prefixed hexstring format
	Proof string `json:"proof"`
	// Prover mode "Light|Medium|Full"
	ProverMode    string `json:"proverMode"`
	VerifierIndex int    `json:"verifierIndex"`
	// Block dependant inputs for the proof
	BlocksData []BlockData `json:"blocksData"`
	// Initial root hash before executing the conflated block
	ParentStateRootHash string `json:"parentStateRootHash"`
	// Version : "0.0.1"
	Version string `json:"proverVersion"`

	// First block number
	FirstBlockNumber int `json:"firstBlockNumber"`

	// Debug data, helps tracking issues for deserializing the hashes
	DebugData struct {
		Blocks []PerBlockDebugData `json:"blocks"`
		// Hash for all the blocks
		HashForAllBlocks string `json:"hashForAllBlocks"`
		// Hasf of the n+1 root hashes concatenated altogether
		HashOfRootHashes string `json:"hashOfRootHashes"`
		// Hash of the time stamps
		TimeStampsHash string `json:"timestampHashes"`
		// Final hash, after applying the modulus
		FinalHash string `json:"finalHash"`
	}
}

type BlockData struct {
	// T Transaction in 0x-prefixed hex format
	RlpEncodedTransactions []string `json:"rlpEncodedTransactions"`
	// L2 to L1 message hashes
	L2ToL1MsgHashes []string `json:"l2ToL1MsgHashes"`
	// List of the N timestamps for each blocks. To optimize
	// for space we put the timestamps in uint64 form
	TimeStamp uint64 `json:"timestamp"`
	// List of the n+1 root hashes in chronological order
	// The first root hash is the initial root hash before
	// execution of the first block in the conflated batch
	// and the last one is the final root hash of the state
	// after execution of the last block in the conflated batch.
	RootHash string `json:"rootHash"`
	// The from addresses of the transactions in the block all concatenated
	// in a single hex string.
	FromAddresses string `json:"fromAddresses"`
	// Not part of the inputs to hash. Flag indicating whether the block
	// contains a BatchL1MsgReceiptConfirmation
	BatchReceptionIndices []uint16 `json:"batchReceptionIndices"`
}

type PerBlockDebugData struct {
	// TxHashes in 0x prefixed format
	TxHashes []string `json:"txHashes"`
	// Hash of the txHashes
	HashOfTxHashes string `json:"hashOfTxHashes"`
	// Hash of the log hashes
	HashOfLogHashes string `json:"hashOfLogHashes"`
	// Hash of the L1 reception txs positions
	HashOfPositions string `json:"hashOfPositions"`
	// Hash of the from addresses
	HashOfFromAddresses string `json:"hashOfFromAddresses"`
	// Final resulting hash obtained after all inputs have been hashed
	HashForBlock string `json:"HashForBlock"`
}

/*
Writes the proof and the public inputs into the given filepath
*/
func (j *ProverOutput) WriteInFile(p string) {

	// This overwrites the file if it already exists
	f := files.MustOverwrite(p)

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(j); err != nil {
		panic(err)
	}
}
