package invalidity

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	smCommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// createMockLogColumns creates mock columns that simulate the arithmetization log columns
// with the names expected by NewPublicInputFetcher
func createMockLogColumns(comp *wizard.CompiledIOP, size int) logs.LogColumns {
	createCol := func(name string) ifaces.Column {
		return comp.InsertCommit(0, ifaces.ColID(name), size)
	}

	return logs.LogColumns{
		IsLog0:       createCol("loginfo.IS_LOG_X_0"),
		IsLog1:       createCol("loginfo.IS_LOG_X_1"),
		IsLog2:       createCol("loginfo.IS_LOG_X_2"),
		IsLog3:       createCol("loginfo.IS_LOG_X_3"),
		IsLog4:       createCol("loginfo.IS_LOG_X_4"),
		AbsLogNum:    createCol("loginfo.ABS_LOG_NUM"),
		AbsLogNumMax: createCol("loginfo.ABS_LOG_NUM_MAX"),
		Ct:           createCol("loginfo.CT"),
		DataHi:       createCol("loginfo.DATA_HI"),
		DataLo:       createCol("loginfo.DATA_LO"),
		TxEmitsLogs:  createCol("loginfo.TXN_EMITS_LOGS"),
	}
}

// createMockStateSummary creates a mock statesummary.Module with only the columns
// required by the PublicInputFetcher (IsActive, IsStorage, WorldStateRoot, and AccumulatorStatement)
func createMockStateSummary(comp *wizard.CompiledIOP, size int) *statesummary.Module {
	createCol := func(name string) ifaces.Column {
		return comp.InsertCommit(0, ifaces.ColIDf("MOCK_SS_%s", name), size)
	}

	return &statesummary.Module{
		IsActive:       createCol("IS_ACTIVE"),
		IsStorage:      createCol("IS_STORAGE"),
		WorldStateRoot: createCol("WORLD_STATE_ROOT"),
		AccumulatorStatement: statesummary.AccumulatorStatement{
			StateDiff: smCommon.StateDiff{
				InitialRoot: createCol("INITIAL_ROOT"),
				FinalRoot:   createCol("FINAL_ROOT"),
			},
		},
	}
}

// assignMockStateSummary assigns test values to the mock state summary columns
func assignMockStateSummary(run *wizard.ProverRuntime, ss *statesummary.Module, size int, initialRoot, finalRoot field.Element) {
	// Create vectors for assignment
	isActiveVec := make([]field.Element, size)
	isStorageVec := make([]field.Element, size)
	worldStateRootVec := make([]field.Element, size)
	initialRootVec := make([]field.Element, size)
	finalRootVec := make([]field.Element, size)

	// Set first few rows as active (simulate some state operations)
	activeRows := size / 2
	for i := 0; i < activeRows; i++ {
		isActiveVec[i] = field.One()
		initialRootVec[i] = initialRoot
		finalRootVec[i] = finalRoot
		worldStateRootVec[i] = initialRoot
	}

	run.AssignColumn(ss.IsActive.GetColID(), smartvectors.NewRegular(isActiveVec))
	run.AssignColumn(ss.IsStorage.GetColID(), smartvectors.NewRegular(isStorageVec))
	run.AssignColumn(ss.WorldStateRoot.GetColID(), smartvectors.NewRegular(worldStateRootVec))
	run.AssignColumn(ss.AccumulatorStatement.StateDiff.InitialRoot.GetColID(), smartvectors.NewRegular(initialRootVec))
	run.AssignColumn(ss.AccumulatorStatement.StateDiff.FinalRoot.GetColID(), smartvectors.NewRegular(finalRootVec))
}

// fetcherTestCase defines a test case for the PublicInputFetcher
type fetcherTestCase struct {
	name      string
	numL2Logs int
}

// TestPublicInputFetcher tests that the PublicInputFetcher is properly created and assigns values correctly
func TestPublicInputFetcher(t *testing.T) {
	testCases := []fetcherTestCase{
		{
			name:      "no_logs",
			numL2Logs: 0,
		},
		{
			name:      "with_l2l1_logs",
			numL2Logs: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ss      *statesummary.Module
				fetcher PublicInputFetcher
				logCols logs.LogColumns
			)

			// Generate test logs
			var testLogs []logs.LogInfo
			if tc.numL2Logs > 0 {
				testLogs, _, _ = logs.GenerateSimpleL2L1Test()
			} else {
				testLogs, _, _ = logs.GenerateTestWithoutRelevantLogs()
			}

			// Compute appropriate column size for logs
			fullSize := logs.ComputeSize(testLogs)
			logColSize := utils.NextPowerOfTwo(fullSize)
			ssSize := 1 << 6 // 64

			// Test root hash values
			initialRoot := field.NewElement(0x1234567890abcdef)
			finalRoot := field.NewElement(0xfedcba0987654321)

			define := func(b *wizard.Builder) {
				comp := b.CompiledIOP

				// Create mock state summary with only required columns
				ss = createMockStateSummary(comp, ssSize)

				// Create mock log columns with names expected by NewPublicInputFetcher
				logCols = createMockLogColumns(comp, logColSize)

				// Create the PublicInputFetcher
				fetcher = NewPublicInputFetcher(comp, ss)
			}

			prove := func(run *wizard.ProverRuntime) {
				// Assign mock state summary
				assignMockStateSummary(run, ss, ssSize, initialRoot, finalRoot)

				// Assign the mock log columns
				logs.LogColumnsAssign(run, &logCols, testLogs)

				// Use a dummy bridge address for testing
				bridgeAddress := common.Address{}
				copy(bridgeAddress[:], []byte{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02})

				// Assign the fetcher
				fetcher.Assign(run, [20]byte(bridgeAddress))
			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}

			// Verify that the fetcher components were created properly
			if fetcher.RootHashFetcher == nil {
				t.Error("RootHashFetcher should not be nil")
			}
			if fetcher.StateSummary == nil {
				t.Error("StateSummary should not be nil")
			}
		})
	}
}

// TestPublicInputFetcherWithL2L1Logs tests the fetcher specifically with L2L1 logs
func TestPublicInputFetcherWithL2L1Logs(t *testing.T) {
	// Generate test logs that include L2L1 logs
	testLogs, bridgeAddr, testName := logs.GenerateLargeTest()
	t.Logf("Using test logs: %s", testName)

	fullSize := logs.ComputeSize(testLogs)
	logColSize := utils.NextPowerOfTwo(fullSize)
	ssSize := 1 << 6

	initialRoot := field.NewElement(0xAABBCCDDEEFF0011)
	finalRoot := field.NewElement(0x1100FFEEDDCCBBAA)

	var (
		ss      *statesummary.Module
		fetcher PublicInputFetcher
		logCols logs.LogColumns
	)

	define := func(b *wizard.Builder) {
		comp := b.CompiledIOP

		// Create mock state summary
		ss = createMockStateSummary(comp, ssSize)

		// Create mock log columns
		logCols = createMockLogColumns(comp, logColSize)

		// Create the fetcher
		fetcher = NewPublicInputFetcher(comp, ss)
	}

	prove := func(run *wizard.ProverRuntime) {
		// Assign mock state summary
		assignMockStateSummary(run, ss, ssSize, initialRoot, finalRoot)

		// Assign log columns
		logs.LogColumnsAssign(run, &logCols, testLogs)

		// Assign the fetcher with the bridge address from the test
		var bridgeAddress [20]byte
		copy(bridgeAddress[:], bridgeAddr[:])
		fetcher.Assign(run, bridgeAddress)
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	err := wizard.Verify(comp, proof)

	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
}
