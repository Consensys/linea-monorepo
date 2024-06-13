package execution

import "github.com/consensys/zkevm-monorepo/prover/utils"

func (j *Response) ComputeProofInput() {

	debugData := &j.DebugData

	// 1 - For each block
	for i := range j.BlocksData {
		blockInputs := &j.BlocksData[i]
		blockDebug := &PerBlockDebugData{}

		// - Hash the rlp transactions one by one
		txHashes := make([]string, len(blockInputs.RlpEncodedTransactions))
		for i := range txHashes {
			txHashes[i] = utils.HexHashHex(blockInputs.RlpEncodedTransactions[i])
		}
		blockDebug.TxHashes = txHashes
		// - Hash of the transaction hashes concatenated altogether
		blockDebug.HashOfTxHashes = utils.HexHashHex(blockDebug.TxHashes...)
		// - Hash of the log hashes concatenated altogether
		blockDebug.HashOfLogHashes = utils.HexHashHex(blockInputs.L2ToL1MsgHashes...)
		// - Hash of the positions (encoded as uint256)
		{
			batchReceptionIndices := blockInputs.BatchReceptionIndices
			// encode each of them as a uint256
			casted := []string{}
			for i := range batchReceptionIndices {
				casted = append(casted, utils.FmtIntHex32Bytes(int(batchReceptionIndices[i])))
			}
			blockDebug.HashOfPositions = utils.HexHashHex(casted...)
		}

		// Hash the from addresses : concatenated the addresses and hash the result.
		blockDebug.HashOfFromAddresses = utils.HexHashHex(blockInputs.FromAddresses)

		// f - Get the final hashes
		blockDebug.HashForBlock = utils.HexHashHex(
			blockDebug.HashOfTxHashes,
			blockDebug.HashOfLogHashes,
			blockDebug.HashOfPositions,
			blockDebug.HashOfFromAddresses,
		)

		debugData.Blocks = append(debugData.Blocks, *blockDebug)
	}

	// 2 - Hash for all the blocks altogether
	debugData.HashForAllBlocks = j.HashForAllBlocks()

	// 3 - Hash of the time stamps
	debugData.TimeStampsHash = j.TimeStampHashes()

	// 4 - Hash of the log hashes
	debugData.HashOfRootHashes = j.HashOfRootHashes()

	// 5 - Finally accumulate the first and last block number
	debugData.FinalHash = utils.HexHashHex(
		debugData.HashForAllBlocks,
		utils.FmtIntHex32Bytes(j.FirstBlockNumber),
		debugData.TimeStampsHash,
		debugData.HashOfRootHashes,
	)

	// and apply the modulus
	debugData.FinalHash = utils.ApplyModulusBls12377(debugData.FinalHash)
}

func (j *Response) TimeStampHashes() string {
	// Collect the timestamps
	timestamps := []uint64{}
	for _, b := range j.BlocksData {
		timestamps = append(timestamps, b.TimeStamp)
	}
	// Then, returns the hash
	return utils.HexHashUint64(timestamps...)
}

func (j *Response) HashOfRootHashes() string {
	// Collect the root hashes
	rootHashes := []string{j.ParentStateRootHash}
	for _, b := range j.BlocksData {
		rootHashes = append(rootHashes, b.RootHash)
	}
	// Then, returns the hash
	return utils.HexHashHex(rootHashes...)
}

func (j *Response) HashForAllBlocks() string {
	hashesForBlocks := []string{}
	// Concatenation of the hashes for each block
	for _, b := range j.DebugData.Blocks {
		hashesForBlocks = append(hashesForBlocks, b.HashForBlock)
	}
	// Then, return the hash
	return utils.HexHashHex(hashesForBlocks...)
}
