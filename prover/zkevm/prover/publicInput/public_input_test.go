package publicInput

import (
	"fmt"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	arith "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/logs"
	util "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
	stmCommon "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager/common"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager/mock"
	stmgr "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

// Test Defining and Assigning all modules using test data, and then generating
// a PublicInput, along with a FunctionalInputExtractor
func TestPublicInputDefineAndAssign(t *testing.T) {
	ctBlockData := util.InitializeCsv("testdata/blockdata_mock.csv", t)
	ctTxnData := util.InitializeCsv("testdata/txndata_mock.csv", t)
	ctRlpTxn := util.InitializeCsv("testdata/rlp_txn_mock.csv", t)

	var (
		pub       *PublicInput
		inp       InputModules
		extractor *FunctionalInputExtractor
	)
	stateSummaryContext := stmCommon.InitializeContext(100)
	testLogs, bridgeAddress, _ := logs.GenerateLargeTest()
	logFullSize := logs.ComputeSize(testLogs[:])
	logColSize := utils.NextPowerOfTwo(logFullSize)

	define := func(b *wizard.Builder) {
		// Define BlockData, TxnData and RlpTxn
		inp.BlockData, inp.TxnData, inp.RlpTxn = arith.DefineTestingArithModules(b, ctBlockData, ctTxnData, ctRlpTxn)
		// Define State Summary
		ss := stmgr.NewModule(b.CompiledIOP, 1<<6)
		inp.StateSummary = &ss
		// Define the Logs
		inp.LogCols = logs.NewLogColumns(b.CompiledIOP, logColSize, "MOCK")
		pub = newPublicInput(b.CompiledIOP, &inp, Settings{
			BridgeAddress: bridgeAddress,
			ChainID:       field.NewFromString("0xccc00000000000000000000000000000"),
			Name:          "TESTING",
		})
		// Compute an extractor
		extractor = pub.GetExtractor()
		fmt.Println(extractor) // do something else with it
	}

	prove := func(run *wizard.ProverRuntime) {
		// Assign BlockData, TxnData and RlpTxn
		arith.AssignTestingArithModules(run, ctBlockData, ctTxnData, ctRlpTxn)
		var (
			initState    = stateSummaryContext.State
			shomeiState  = mock.InitShomeiState(initState)
			stateLogs    = stateSummaryContext.TestCases[0].StateLogsGens(initState)
			shomeiTraces = mock.StateLogsToShomeiTraces(shomeiState, stateLogs)
		)
		// Assign the StateSummary
		inp.StateSummary.Assign(run, shomeiTraces)
		// Assign the Logs
		logs.LogColumnsAssign(run, &inp.LogCols, testLogs[:])
		pub.Assign(run)

	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

}
