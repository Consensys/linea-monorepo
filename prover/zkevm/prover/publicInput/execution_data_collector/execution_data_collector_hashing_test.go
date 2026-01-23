package execution_data_collector

import (
	"fmt"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/types"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"

	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	fetch "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// fixedTestDataRlpLimbs stores the RLP limbs corresponding to each transaction.
// This data is consistent to the CSV testdata. The input is the absolute transaction ID.
var fixedTestDataRlpLimbs = [][]string{
	{"f9", "ccc0"},
	{"123456", "f0"},
	{"47d5f9a0"},
	{"477aaaa0", "aaaa", "cc"},
	{"bb"},
	{"aaaaaaaaaa"},
	{"ff"},
	{"aaa0"},
	{"aaa0"},
	{"eee0", "1a2a3a4a5a", "b1b2b3", "10"},
}

// ComputeMiMCHashFixedTestData computes the execution data hash that corresponds to the test
// data in our CSV files. It concatenates all the bytes in the same order as the ExecutionDataCollector
// into a large string that contains all the bytes in hexadecimal.
// It then pads the output with 0s until it is a multiple of 62 (this corresponds to 31 bytes, which is the
// maximum number of bytes that the packing module fits into a single field element).
// We then split the large string into subslices of length 62, and use them to instantiate new field elements.
// Finally, the field elements are hashed using MiMC.
func ComputeMiMCHashFixedTestData() field.Element {
	// 4 byte timestamp values for each block
	var timestamps = [...]string{"0000000a", "000000ab", "000000bc", "000000cd"}
	// the maximal number of transaction for each block is stored in 2 bytes
	var noTxString = [...]string{"0003", "0004", "0002", "0001"}
	// here are maximal number of transaction for each block, stored as int
	var noTx = [...]int{3, 4, 2, 1}
	// 4 bytes for the Hi part of the sender address of each transaction
	var senderAddrHi = [...]string{"aaaaaaaa", "bbbbbbbb", "cccccccc", "dddddddd", "eeeeeeee", "ffffffff", "aaaaaaaa", "bbbbbbbb", "cccccccc", "dddddddd"}
	// 16 bytes for the Lo part of the sender address of each transaction
	var senderAddrLo = [...]string{"ffffffffffffffffffffffffffffffff", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccc", "dddddddddddddddddddddddddddddddd", "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "ffffffffffffffffffffffffffffffff", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccc"}
	// prepare a strings Builder to concatenate the test data in the proper manner
	var fullString strings.Builder

	absTxId := 1 // a counter that will iterate over the absolute transaction ID.
	// iterate over the blocks
	for blockId := 0; blockId < 4; blockId++ {
		// concatenate the bytes for the number of transactions, the block timestamp and the blockhash
		fullString.WriteString(noTxString[blockId])
		fullString.WriteString(timestamps[blockId])                // 6 bytes, timestamps
		fullString.WriteString("00000000000000000000000000000000") // 16 bytes, BlockHash HI
		fullString.WriteString("00000000000000000000000000000000") // 16 bytes, BlockHash LO
		// absTxIdLimit is the absolute transaction ID of the first transaction in the next block
		absTxIdLimit := absTxId + noTx[blockId]
		// iterate over the id of the transactions in the current blocks
		for absTxId < absTxIdLimit {
			// concatenate the sender address of the transaction with absTxId
			fullString.WriteString(senderAddrHi[absTxId-1])
			fullString.WriteString(senderAddrLo[absTxId-1])
			// get the RLP limbs for the transaction and concatenate them
			rlpLimbs := fixedTestDataRlpLimbs[absTxId-1]
			for _, rlpLimb := range rlpLimbs {
				fullString.WriteString(rlpLimb)
			}
			// done with the data for this transaction (sender address + RLP limbs), move on to the next
			absTxId++
		}
		absTxId = absTxIdLimit
	}

	// padding with 0s until the length of the large concatenated string is a multiple of 62
	// (corresponding to 31 bytes)
	for fullString.Len()%62 > 0 {
		fullString.WriteString("0") // pad with 0s
	}

	// pack the bytes in the large string into a vector vectFieldElem of field elements
	vectFieldElem := make([]field.Element, fullString.Len()/62)
	finalString := fullString.String()
	for i := 0; i < fullString.Len()/62; i++ {
		// compute stringSlice, a 62 characters/31-bytes subslice of the large string
		stringSlice := fmt.Sprintf("0x%s", finalString[i*62:(i+1)*62])
		// instantiate a field element from the subslice and add it to the vector
		elem := field.NewFromString(stringSlice)
		vectFieldElem[i].Set(&elem)
	}
	// compute and return the MiMC hash of the packed field elements
	finalHash := hasher_factory.HashVec(vectFieldElem)
	return finalHash
}

// TestExecutionDataCollectAndHash instantiates mock arithmetization modules
// BlockData, TxnData and RlpTxn from CSV test files. It then defines and assigns
// an ExecutionDataCollector, then pads, packs and MiMC-hashes its limbs.
func TestExecutionDataCollectorAndHash(t *testing.T) {
	ctBlockData := util.InitializeCsv("../testdata/blockdata_mock.csv", t)
	ctTxnData := util.InitializeCsv("../testdata/txndata_mock.csv", t)
	ctRlpTxn := util.InitializeCsv("../testdata/rlp_txn_mock.csv", t)
	blockHashList := [1 << 10]types.FullBytes32{}

	var (
		execDataCollector   *ExecutionDataCollector
		blockTxnMeta        fetch.BlockTxnMetadata
		blockDataFetcher    *fetch.BlockDataFetcher
		txnDataFetcher      fetch.TxnDataFetcher
		rlpTxnFetcher       fetch.RlpTxnFetcher
		txnDataCols         *arith.TxnData
		blockDataCols       *arith.BlockDataCols
		rlpTxn              *arith.RlpTxn
		ppp                 PoseidonPadderPacker
		mimcHasher          PoseidonHasher
		genericPadderPacker GenericPadderPacker
		chainIDFetcher      fetch.ChainIDFetcher
	)

	define := func(b *wizard.Builder) {
		// define the arith test modules
		blockDataCols, txnDataCols, rlpTxn = arith.DefineTestingArithModules(b, ctBlockData, ctTxnData, ctRlpTxn)
		// create and define a metadata fetcher
		blockTxnMeta = fetch.NewBlockTxnMetadata(b.CompiledIOP, "BLOCK_TX_METADATA", txnDataCols)
		fetch.DefineBlockTxnMetaData(b.CompiledIOP, &blockTxnMeta, "BLOCK_TX_METADATA", txnDataCols)
		// create a new timestamp fetcher
		blockDataFetcher = fetch.NewBlockDataFetcher(b.CompiledIOP, "TIMESTAMP_FETCHER_FROM_ARITH", blockDataCols)
		// constrain the timestamp fetcher
		fetch.DefineBlockDataFetcher(b.CompiledIOP, blockDataFetcher, "TIMESTAMP_FETCHER_FROM_ARITH", blockDataCols)
		txnDataFetcher = fetch.NewTxnDataFetcher(b.CompiledIOP, "TXN_DATA_FETCHER_FROM_ARITH", txnDataCols)
		fetch.DefineTxnDataFetcher(b.CompiledIOP, &txnDataFetcher, "TXN_DATA_FETCHER_FROM_ARITH", txnDataCols)

		rlpTxnFetcher = fetch.NewRlpTxnFetcher(b.CompiledIOP, "RLP_TXN_FETCHER_FROM_ARITH", rlpTxn)
		// constrain the fetcher
		fetch.DefineRlpTxnFetcher(b.CompiledIOP, &rlpTxnFetcher, "RLP_TXN_FETCHER_FROM_ARITH", rlpTxn)
		chainIDFetcher = fetch.NewChainIDFetcher(b.CompiledIOP, "PUBLIC_INPUT_CHAIN_ID_FETCHER", blockDataCols)
		fetch.DefineChainIDFetcher(b.CompiledIOP, &chainIDFetcher, "PUBLIC_INPUT_CHAIN_ID_FETCHER", blockDataCols)

		limbColSize := GetSummarySize(txnDataCols, rlpTxn)
		// we need to artificially blow up the column size by 2, or padding will fail
		limbColSize = 2 * limbColSize
		// create a new ExecutionDataCollector
		execDataCollector = NewExecutionDataCollector(b.CompiledIOP, "EXECUTION_DATA_COLLECTOR", limbColSize)
		// define the ExecutionDataCollector
		DefineExecutionDataCollector(b.CompiledIOP, execDataCollector, "EXECUTION_DATA_COLLECTOR", blockDataFetcher, blockTxnMeta, txnDataFetcher, rlpTxnFetcher)

		genericPadderPacker = NewGenericPadderPacker(b.CompiledIOP, execDataCollector.Limbs, execDataCollector.NoBytes, execDataCollector.IsActive, "GENERIC_PADDER_PACKER_FOR_EXECUTION_DATA_COLLECTOR")
		ppp = NewPoseidonPadderPacker(b.CompiledIOP, genericPadderPacker.OutputData, genericPadderPacker.OutputIsActive, "POSEIDON_PADDER_PACKER_FOR_EXECUTION_DATA_COLLECTOR")
		DefinePoseidonPadderPacker(b.CompiledIOP, ppp, "POSEIDON_PADDER_PACKER_FOR_EXECUTION_DATA_COLLECTOR")
		// create a MiMC hasher
		mimcHasher = NewPoseidonHasher(b.CompiledIOP, ppp.OutputData, ppp.OutputIsActive[0], "MIMC_HASHER")
		// define the hasher
		DefinePoseidonHasher(b.CompiledIOP, mimcHasher, "EXECUTION_DATA_COLLECTOR_MIMC_HASHER")
	}

	prove := func(run *wizard.ProverRuntime) {
		// assign the CSV data for the mock BlockData, TxnData and RlpTxn arithmetization modules
		arith.AssignTestingArithModules(run, ctBlockData, ctTxnData, ctRlpTxn, blockDataCols, txnDataCols, rlpTxn)
		// assign the fetchers
		fetch.AssignBlockDataFetcher(run, blockDataFetcher, blockDataCols)
		fetch.AssignBlockTxnMetadata(run, blockTxnMeta, txnDataCols)
		fetch.AssignTxnDataFetcher(run, txnDataFetcher, txnDataCols)
		fetch.AssignRlpTxnFetcher(run, &rlpTxnFetcher, rlpTxn)
		fetch.AssignChainIDFetcher(run, &chainIDFetcher, blockDataCols)
		// assign the ExecutionDataCollector
		AssignExecutionDataCollector(run, execDataCollector, blockDataFetcher, blockTxnMeta, txnDataFetcher, rlpTxnFetcher, blockHashList[:])

		AssignGenericPadderPacker(run, genericPadderPacker)
		// assign the repacker for Poseidon hashing
		AssignPoseidonPadderPacker(run, ppp)
		// assign the hasher
		AssignPoseidonHasher(run, mimcHasher, ppp.OutputData, ppp.OutputIsActive[0])
		for i := range mimcHasher.HashFinal {
			fmt.Println("Computed Execution Data Hash:", mimcHasher.HashFinal[i].GetColAssignment(run).Pretty())
		}

		// compute the MiMC hash of the fixed TestData
		//fixedHash := ComputeMiMCHashFixedTestData()
		// assert that we are computing the hash correctly
		//assert.Equal(t, fixedHash, mimcHasher.HashFinal.GetColAssignmentAt(run, 0), "Final Hash Value is Incorrect")
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
}
