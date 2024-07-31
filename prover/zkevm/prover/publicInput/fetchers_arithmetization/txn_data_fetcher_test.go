package fetchers_arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
	"testing"
)

// TestTxnDataFetcher tests the fetching of the sender address data
func TestTxnDataFetcher(t *testing.T) {

	// initialize sample TxnData from a mock test data CSV file
	ctTxnData := utilities.InitializeCsv("../testdata/txndata_mock.csv", t)
	var (
		txd     *TxnData
		fetcher TxnDataFetcher
	)

	cmp := wizard.Compile(func(b *wizard.Builder) {
		// register sample arithmetization columns
		txd = &TxnData{
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
		// create a new txn data fetcher
		fetcher = NewTxnDataFetcher(b.CompiledIOP, "TIMESTAMP_FETCHER_FROM_ARITH", txd)
		// constrain the fetcher
		DefineTxnDataFetcher(b.CompiledIOP, &fetcher, "TIMESTAMP_FETCHER_FROM_ARITH", txd)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		// assign the CSV columns
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
		// assign the txn data fetcher
		AssignTxnDataFetcher(run, fetcher, txd)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
