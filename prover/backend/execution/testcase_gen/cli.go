package testcase_gen

const (
	// seed for the RNG
	SEED_FLAG  string = "seed"
	SEED_USAGE string = "required, integer, seed to use for the RNG"

	// chain-id
	CHAIN_ID_FLAG  string = "chainid"
	CHAIN_ID_USAGE string = "required, integer, the chainID of L2"

	// l2 bridge address
	L2_MESSAGE_SERVICE_CONTRACT_FLAG  string = "l2-bridge-address"
	L2_MESSAGE_SERVICE_CONTRACT_USAGE string = "l2-bridge-address: required, the 0x prefixed address of the bridge"

	// time stamp
	START_FROM_TIMESTAMP_FLAG  string = "start-from-timestamp"
	START_FROM_TIMESTAMP_USAGE string = "required, UNIX-timestamp, timestamp starting which we generate the blocks"

	// initial root hash
	START_FROM_ROOT_HASH_FLAG  string = "start-from-root-hash"
	START_FROM_ROOT_HASH_USAGE string = "root upon which we generate a proof"

	// final root hash
	END_WITH_ROOT_HASH_FLAG  string = "end-with-root-hash"
	END_WITH_ROOT_HASH_USAGE string = "optional, root hash at which the sequence of blocks should end"

	// output file
	OUTPUT_FILE_FLAG  string = "ofile"
	OUTPUT_FILE_USAGE string = "required, path, file where to write the testcase file"

	// max number of block
	NUM_BLOCKS_FLAG  string = "num-blocks"
	NUM_BLOCKS_USAGE string = "optional, integer, maximal number of block in the conflated batch"

	// bound on the number of tx per block
	MAX_TX_PER_BLOCK_FLAG  string = "max-tx-per-block"
	MAX_TX_PER_BLOCK_USAGE string = "optional, integer, maximal number of tx per block"

	// bound on the number of l2 msg logs per block
	// #nosec G101 -- Not a credential
	MAX_L2L1_LOGS_PER_BLOCKS_FLAG  string = "max-l2-l1-logs-per-block"
	MAX_L2L1_LOGS_PER_BLOCKS_USAGE string = "optional, int, maximal number of l2 msg emitted per block"

	// bound on the number of msg receipt confirmation
	MAX_L1L2_MSG_RECEIPT_PER_BLOCK_FLAG  string = "max-l1-l2-msg-receipt-per-block"
	MAX_L1L2_MSG_RECEIPT_PER_BLOCK_USAGE string = "optional, int, maximal number of l1 msg received per block (at most one batch per block)."

	// min tx size
	MIN_TX_BYTE_SIZE_FLAG  string = "min-tx-byte-size"
	MIN_TX_BYTE_SIZE_USAGE string = "optional, int, minimal size of an rlp encoded tx"

	// max-tx-size
	MAX_TX_BYTE_SIZE_FLAG  string = "max-tx-byte-size"
	MAX_TX_BYTE_SIZE_USAGE string = "optional, int, maximal size of an rlp encoded tx"

	// calldata-with-zero
	ZEROES_IN_CALLDATA_FLAG  string = "zeroes-in-calldata"
	ZEROES_IN_CALLDATA_USAGE string = "optional, bool, default to true. If set to true 75pc of the bytes of the calldata of the generated txs is set to zero"

	// start from block number
	START_FROM_BLOCK_NUMBER_FLAG  string = "start-from-block"
	START_FROM_BLOCK_NUMBER_USAGE string = "required, integer"
)

// placeholder variables for the flag values
var (
	seed, chainid                                        *int64
	l2BridgeAddress, startFromRootHash, oFile            *string
	endWithRootHash                                      *string
	startFromTimestamp                                   *uint64
	numBlock, maxTxPerBlock, maxL2L1LogsPerBlock         *int
	maxL1L2ReceiptPerBlock, minTxByteSize, maxTxByteSize *int
	startFromBlock                                       *int
	zeroesInCalldata                                     *bool
)
