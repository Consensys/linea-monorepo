package invalidity

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	smCommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// createMockStateSummary creates a mock statesummary.Module with only the columns
// required by the PublicInputFetcher (IsActive, IsStorage, WorldStateRoot, and AccumulatorStatement)
func createMockStateSummary(comp *wizard.CompiledIOP, size int) *statesummary.Module {
	createCol := func(name string) ifaces.Column {
		return comp.InsertCommit(0, ifaces.ColIDf("MOCK_SS_%s", name), size, true)
	}

	// Create array columns for WorldStateRoot (NbElemPerHash = 8 columns)
	var worldStateRoot [common.NbElemPerHash]ifaces.Column
	for i := range common.NbElemPerHash {
		worldStateRoot[i] = createCol(fmt.Sprintf("WORLD_STATE_ROOT_%d", i))
	}

	// Create array columns for InitialRoot and FinalRoot
	var initialRoot, finalRoot [common.NbElemPerHash]ifaces.Column
	for i := range common.NbElemPerHash {
		initialRoot[i] = createCol(fmt.Sprintf("INITIAL_ROOT_%d", i))
		finalRoot[i] = createCol(fmt.Sprintf("FINAL_ROOT_%d", i))
	}

	return &statesummary.Module{
		IsActive:       createCol("IS_ACTIVE"),
		IsStorage:      createCol("IS_STORAGE"),
		WorldStateRoot: worldStateRoot,
		AccumulatorStatement: statesummary.AccumulatorStatement{
			StateDiff: smCommon.StateDiff{
				InitialRoot: initialRoot,
				FinalRoot:   finalRoot,
			},
		},
	}
}

// assignMockStateSummary assigns test values to the mock state summary columns
func assignMockStateSummary(run *wizard.ProverRuntime, ss *statesummary.Module, size int, initialRoot, finalRoot field.Element) {
	// Create vectors for assignment
	isActiveVec := make([]field.Element, size)
	isStorageVec := make([]field.Element, size)

	// Set first few rows as active (simulate some state operations)
	activeRows := size / 2
	for i := 0; i < activeRows; i++ {
		isActiveVec[i] = field.One()
	}

	run.AssignColumn(ss.IsActive.GetColID(), smartvectors.NewRegular(isActiveVec))
	run.AssignColumn(ss.IsStorage.GetColID(), smartvectors.NewRegular(isStorageVec))

	// Assign WorldStateRoot columns (array of 8 columns)
	for i := range common.NbElemPerHash {
		worldStateRootVec := make([]field.Element, size)
		for j := 0; j < activeRows; j++ {
			if i == 0 {
				worldStateRootVec[j] = initialRoot // Put root value in first limb for simplicity
			}
		}
		run.AssignColumn(ss.WorldStateRoot[i].GetColID(), smartvectors.NewRegular(worldStateRootVec))
	}

	// Assign InitialRoot and FinalRoot columns (arrays of 8 columns each)
	for i := range common.NbElemPerHash {
		initialRootVec := make([]field.Element, size)
		finalRootVec := make([]field.Element, size)
		for j := 0; j < activeRows; j++ {
			if i == 0 {
				initialRootVec[j] = initialRoot // Put root value in first limb for simplicity
				finalRootVec[j] = finalRoot
			}
		}
		run.AssignColumn(ss.AccumulatorStatement.StateDiff.InitialRoot[i].GetColID(), smartvectors.NewRegular(initialRootVec))
		run.AssignColumn(ss.AccumulatorStatement.StateDiff.FinalRoot[i].GetColID(), smartvectors.NewRegular(finalRootVec))
	}
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
				logCols = logs.NewLogColumns(comp, logColSize, "MOCK_LOG_COLUMNS")
				// Create the PublicInputFetcher
				fetcher = NewPublicInputFetcher(comp, ss, logCols)
			}

			prove := func(run *wizard.ProverRuntime) {
				// Assign mock state summary
				assignMockStateSummary(run, ss, ssSize, initialRoot, finalRoot)

				// Assign the mock log columns
				logs.LogColumnsAssign(run, &logCols, testLogs)

				// Use a dummy bridge address for testing
				bridgeAddress := ethcommon.Address{}
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
		logCols = logs.NewLogColumns(comp, logColSize, "MOCK_LOG_COLUMNS")

		// Create the fetcher
		fetcher = NewPublicInputFetcher(comp, ss, logCols)
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
