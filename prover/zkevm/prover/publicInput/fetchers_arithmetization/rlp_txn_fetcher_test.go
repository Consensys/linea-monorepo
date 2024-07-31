package fetchers_arithmetization

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// TestRlpTxnFetcher tests the fetching of the rlp txn data
func TestRlpTxnFetcher(t *testing.T) {

	// initialize sample RlpTxn data from a mock test data CSV file
	ctRlpTxn := utilities.InitializeCsv("../testdata/rlp_txn_mock.csv", t)
	var (
		rt      *RlpTxn
		fetcher RlpTxnFetcher
	)

	cmp := wizard.Compile(func(b *wizard.Builder) {
		// register sample arithmetization columns
		rt = &RlpTxn{
			AbsTxNum:       ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM"),
			AbsTxNumMax:    ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM_MAX"),
			ToHashByProver: ctRlpTxn.GetCommit(b, "RL.TO_HASH_BY_PROVER"),
			Limb:           ctRlpTxn.GetCommit(b, "RL.LIMB"),
			NBytes:         ctRlpTxn.GetCommit(b, "RL.NBYTES"),
		}
		fetcher = NewRlpTxnFetcher(b.CompiledIOP, "RLP_TXN_FETCHER_FROM_ARITH", rt)
		// constrain the fetcher
		DefineRlpTxnFetcher(b.CompiledIOP, &fetcher, "RLP_TXN_FETCHER_FROM_ARITH", rt)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		// assign the CSV columns
		ctRlpTxn.Assign(
			run,
			"RT.ABS_TX_NUM",
			"RT.ABS_TX_NUM_MAX",
			"RL.TO_HASH_BY_PROVER",
			"RL.LIMB",
			"RL.NBYTES",
		)
		AssignRlpTxnFetcher(run, &fetcher, rt)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
