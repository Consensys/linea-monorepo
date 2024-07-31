package publicInput

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	fetch "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
	"testing"
)

func TestExecutionDataCollector(t *testing.T) {
	ctBlockData := utilities.InitializeCsv("testdata/blockdata_mock.csv", t)
	ctTxnData := utilities.InitializeCsv("testdata/txndata_mock.csv", t)
	ctRlpTxn := utilities.InitializeCsv("testdata/rlp_txn_mock.csv", t)

	var (
		edc              ExecutionDataCollector
		btm              fetch.BlockTxnMetadata
		timestampFetcher fetch.TimestampFetcher
		txnDataFetcher   fetch.TxnDataFetcher
		rlpTxnFetcher    fetch.RlpTxnFetcher
		txd              *fetch.TxnData
		bdc              *fetch.BlockDataCols
		rt               *fetch.RlpTxn
	)

	define := func(b *wizard.Builder) {
		bdc = &fetch.BlockDataCols{
			RelBlock: ctBlockData.GetCommit(b, "REL_BLOCK"),
			Inst:     ctBlockData.GetCommit(b, "INST"),
			Ct:       ctBlockData.GetCommit(b, "CT"),
			DataHi:   ctBlockData.GetCommit(b, "DATA_HI"),
			DataLo:   ctBlockData.GetCommit(b, "DATA_LO"),
		}
		txd = &fetch.TxnData{
			AbsTxNum:        ctTxnData.GetCommit(b, "TD.ABS_TX_NUM"),
			AbsTxNumMax:     ctTxnData.GetCommit(b, "TD.ABS_TX_NUM_MAX"),
			RelTxNum:        ctTxnData.GetCommit(b, "TD.REL_TX_NUM"),
			RelTxNumMax:     ctTxnData.GetCommit(b, "TD.REL_TX_NUM_MAX"),
			Ct:              ctTxnData.GetCommit(b, "TD.CT"),
			FromHi:          ctTxnData.GetCommit(b, "TD.FROM_HI"),
			FromLo:          ctTxnData.GetCommit(b, "TD.FROM_LO"),
			IsLastTxOfBlock: ctTxnData.GetCommit(b, "TD.IS_LAST_TX_OF_BLOCK"),
			RelBlock:        ctTxnData.GetCommit(b, "TD.REL_BLOCK"),
		}
		rt = &fetch.RlpTxn{
			AbsTxNum:       ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM"),
			AbsTxNumMax:    ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM_MAX"),
			ToHashByProver: ctRlpTxn.GetCommit(b, "RL.TO_HASH_BY_PROVER"),
			Limb:           ctRlpTxn.GetCommit(b, "RL.LIMB"),
			NBytes:         ctRlpTxn.GetCommit(b, "RL.NBYTES"),
		}

		btm = fetch.NewBlockTxnMetadata(b.CompiledIOP, "BLOCK_TX_METADATA", txd)
		fetch.DefineBlockTxnMetaData(b.CompiledIOP, &btm, "BLOCK_TX_METADATA", txd)
		// create a new timestamp fetcher
		timestampFetcher = fetch.NewTimestampFetcher(b.CompiledIOP, "TIMESTAMP_FETCHER_FROM_ARITH", bdc)
		// constrain the timestamp fetcher
		fetch.DefineTimestampFetcher(b.CompiledIOP, &timestampFetcher, "TIMESTAMP_FETCHER_FROM_ARITH", bdc)
		txnDataFetcher = fetch.NewTxnDataFetcher(b.CompiledIOP, "TXN_DATA_FETCHER_FROM_ARITH", txd)
		fetch.DefineTxnDataFetcher(b.CompiledIOP, &txnDataFetcher, "TXN_DATA_FETCHER_FROM_ARITH", txd)

		rlpTxnFetcher = fetch.NewRlpTxnFetcher(b.CompiledIOP, "RLP_TXN_FETCHER_FROM_ARITH", rt)
		// constrain the fetcher
		fetch.DefineRlpTxnFetcher(b.CompiledIOP, &rlpTxnFetcher, "RLP_TXN_FETCHER_FROM_ARITH", rt)

		limbColSize := GetSummarySize(&btm, bdc, txd, rt)
		edc = NewLimbSummary(b.CompiledIOP, "LIMB_SUMMARY", limbColSize)
		DefineLimbSummary(b.CompiledIOP, &edc, "LIMB_SUMMARY", timestampFetcher, btm, txnDataFetcher, rlpTxnFetcher)
	}

	prove := func(run *wizard.ProverRuntime) {
		ctBlockData.Assign(
			run,
			"REL_BLOCK",
			"INST",
			"CT",
			"DATA_HI",
			"DATA_LO",
		)
		ctTxnData.Assign(
			run,
			"TD.ABS_TX_NUM",
			"TD.ABS_TX_NUM_MAX",
			"TD.REL_TX_NUM",
			"TD.REL_TX_NUM_MAX",
			"TD.CT",
			"TD.FROM_HI",
			"TD.FROM_LO",
			"TD.IS_LAST_TX_OF_BLOCK",
			"TD.REL_BLOCK",
		)
		ctRlpTxn.Assign(
			run,
			"RT.ABS_TX_NUM",
			"RT.ABS_TX_NUM_MAX",
			"RL.TO_HASH_BY_PROVER",
			"RL.LIMB",
			"RL.NBYTES",
		)
		fetch.AssignTimestampFetcher(run, timestampFetcher, bdc)
		fetch.AssignBlockTxnMetadata(run, btm, txd)
		fetch.AssignTxnDataFetcher(run, txnDataFetcher, txd)
		fetch.AssignRlpTxnFetcher(run, &rlpTxnFetcher, rt)
		AssignLimbSummary(run, edc, timestampFetcher, btm, txnDataFetcher, rlpTxnFetcher)
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
}
