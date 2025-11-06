package fetchers_arithmetization

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
	"github.com/stretchr/testify/assert"
)

var (
	testChainIDLimbs = []field.Element{
		field.NewFromString("0xccc0"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
		field.NewFromString("0x0000"),
	}
)

// TestRlpTxnFetcher tests the fetching of the rlp txn data
func TestRlpTxnFetcher(t *testing.T) {

	// initialize sample RlpTxn data from a mock test data CSV file
	ctRlpTxn := util.InitializeCsv("../testdata/rlp_txn_mock.csv", t)
	var (
		rt      *arith.RlpTxn
		fetcher RlpTxnFetcher
	)

	cmp := wizard.Compile(func(b *wizard.Builder) {
		// register sample arithmetization columns
		_, _, rt = arith.DefineTestingArithModules(b, nil, nil, ctRlpTxn)
		fetcher = NewRlpTxnFetcher(b.CompiledIOP, "RLP_TXN_FETCHER_FROM_ARITH", rt)
		// constrain the fetcher
		DefineRlpTxnFetcher(b.CompiledIOP, &fetcher, "RLP_TXN_FETCHER_FROM_ARITH", rt)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		// assign the CSV columns
		arith.AssignTestingArithModules(run, nil, nil, ctRlpTxn)
		AssignRlpTxnFetcher(run, &fetcher, rt)

		for i := range fetcher.Limbs {
			assert.Equal(t, testChainIDLimbs[i], fetcher.ChainID[i].GetColAssignmentAt(run, 0), "ChainID value is incorrect.")
		}
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
