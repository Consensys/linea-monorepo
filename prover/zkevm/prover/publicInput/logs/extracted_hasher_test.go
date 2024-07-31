package logs

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/execution/bridge"
	eth "github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

// FirstTopicL2l1 is a helper function that outputs the Hi/Lo parts of the expected first topic of an L2L1 log
func FirstTopicL2l1() (field.Element, field.Element) {
	var firstTopicHi, firstTopicLo field.Element
	firstTopicBytes := bridge.L2L1Topic0() // fixed expected value for the topic on the first topic row
	firstTopicHi.SetBytes(firstTopicBytes[:16])
	firstTopicLo.SetBytes(firstTopicBytes[16:])
	return firstTopicHi, firstTopicLo
}

// FirstTopicRolling is a helper function that outputs the Hi/Lo parts of the expected first topic of a RollingHash log
func FirstTopicRolling() (field.Element, field.Element) {
	var firstTopicRollingHi, firstTopicRollingLo field.Element
	firstTopicRollingBytes := bridge.GetRollingHashUpdateTopic0() // fixed expected value for the topic on the first topic row
	firstTopicRollingHi.SetBytes(firstTopicRollingBytes[:16])
	firstTopicRollingLo.SetBytes(firstTopicRollingBytes[16:])
	return firstTopicRollingHi, firstTopicRollingLo
}

// GenerateTopicsAndAddresses is a common test function that outputs sample topics and addresses
func GenerateTopicsAndAddresses() ([]field.Element, []field.Element, []field.Element, []field.Element) {
	var (
		topicsHi = []field.Element{
			field.NewElement(11),
			field.NewElement(21),
			field.NewElement(31),
			field.NewElement(41),
			field.NewElement(51),
			field.NewElement(61),
		}
		topicsLo = []field.Element{
			field.NewElement(91),
			field.NewElement(101),
			field.NewElement(111),
			field.NewElement(121),
			field.NewElement(131),
			field.NewElement(141),
		}
		address = []eth.Address{
			types.DummyAddress(54),
			types.DummyAddress(64),
			types.DummyAddress(74),
			types.DummyAddress(84),
			types.DummyAddress(94),
		}
	)

	// convert the test addresses into Hi and Lo parts
	addressHi := make([]field.Element, len(address))
	addressLo := make([]field.Element, len(address))
	for i := 0; i < len(address); i++ {
		hi, lo := ConvertAddress(address[i])
		addressHi[i] = hi
		addressLo[i] = lo
	}

	return topicsHi, topicsLo, addressHi, addressLo
}

// GenerateSimpleL2L1Test generates test log info for testing the fetcher of message data from L2L1 logs
func GenerateSimpleL2L1Test() ([]LogInfo, types.EthAddress, string) {
	// get sample topics and addresses
	topicsHi, topicsLo, addressHi, addressLo := GenerateTopicsAndAddresses()

	// fixed expected value for the topic on the first topic row
	firstTopicHi, firstTopicLo := FirstTopicL2l1()

	// compute a dummy bridge address for testing
	bridgeAddress := types.DummyAddress(2)
	bridgeAddrHi, bridgeAddrLo := ConvertAddress(bridgeAddress)
	var testLogs = [...]LogInfo{
		{
			LogType:   2, // log with 2 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{2},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{topicsHi[0], topicsHi[1]},
			TopicsLo:  []field.Element{topicsLo[0], topicsLo[1]},
		},
		{
			LogType:   4, // an L2L1 log of the type we are interested in
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicHi, topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicLo, topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   MISSING_LOG, // missing log
			DataSize:  field.Element{},
			noTopics:  field.Element{},
			AddressHi: field.Element{},
			AddressLo: field.Element{},
			TopicsHi:  nil,
			TopicsLo:  nil,
		},
	}
	return testLogs[:], bridgeAddress, "GenerateSimpleL2L1Test"
}

// GenerateSimpleRollingTest generates test log info for testing the fetcher of message data from RollingHash logs
func GenerateSimpleRollingTest() ([]LogInfo, types.EthAddress, string) {
	// get sample topics and addresses
	topicsHi, topicsLo, addressHi, addressLo := GenerateTopicsAndAddresses()

	// fixed expected value for the topic on the first topic row
	firstTopicRollingHi, firstTopicRollingLo := FirstTopicRolling()

	// compute a dummy bridge address for testing
	bridgeAddress := types.DummyAddress(2)
	bridgeAddrHi, bridgeAddrLo := ConvertAddress(bridgeAddress)
	var testLogs = [...]LogInfo{
		{
			LogType:   3, // RollingHash log with 3 topics
			DataSize:  field.Element{3},
			noTopics:  field.Element{3},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicRollingHi, topicsHi[2], topicsHi[3]},
			TopicsLo:  []field.Element{firstTopicRollingLo, topicsLo[2], topicsLo[3]},
		},
		{
			LogType:   2, // log with 2 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{2},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{topicsHi[0], topicsHi[1]},
			TopicsLo:  []field.Element{topicsLo[0], topicsLo[1]},
		},
	}
	return testLogs[:], bridgeAddress, "GenerateSimpleRollingTest"
}

// GenerateTestWithoutRelevantLogs generates tests that contain no relevant L2L1/RollingHash logs
func GenerateTestWithoutRelevantLogs() ([]LogInfo, types.EthAddress, string) {
	// get sample topics and addresses
	topicsHi, topicsLo, addressHi, addressLo := GenerateTopicsAndAddresses()

	// fixed expected value for the topic on the first topic row
	firstTopicHi, firstTopicLo := FirstTopicL2l1()

	// compute a dummy bridge address for testing
	bridgeAddress := types.DummyAddress(2)
	bridgeAddrHi, bridgeAddrLo := ConvertAddress(bridgeAddress)
	var testLogs = [...]LogInfo{
		{
			LogType:   2, // log with 2 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{2},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{topicsHi[0], topicsHi[1]},
			TopicsLo:  []field.Element{topicsLo[0], topicsLo[1]},
		},
		{
			LogType:   0, // log with 0 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{21},
			AddressHi: field.Element{121},
			AddressLo: field.Element{333},
			TopicsHi:  nil,
			TopicsLo:  nil,
		},
		{
			LogType:   3, // log with 3 topics
			DataSize:  field.Element{3},
			noTopics:  field.Element{3},
			AddressHi: addressHi[1],
			AddressLo: addressLo[1],
			TopicsHi:  []field.Element{topicsHi[1], topicsHi[2], topicsHi[3]},
			TopicsLo:  []field.Element{topicsLo[1], topicsLo[2], topicsLo[3]},
		},
		{
			LogType:   4, // log of type 4 which has a wrong bridge data and will be skipped
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{firstTopicHi, topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicLo, topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   3, // log with 3 topics, but not a RollingHash log (wrong address)
			DataSize:  field.Element{191},
			noTopics:  field.Element{191},
			AddressHi: addressHi[3],
			AddressLo: addressLo[3],
			TopicsHi:  []field.Element{topicsHi[1], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{topicsLo[1], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   4, // log of type 4 which has a wrong first topic in TopicsHi/TopicsLo
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{topicsHi[3], topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{topicsHi[3], topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   MISSING_LOG, // missing log
			DataSize:  field.Element{},
			noTopics:  field.Element{},
			AddressHi: field.Element{},
			AddressLo: field.Element{},
			TopicsHi:  nil,
			TopicsLo:  nil,
		},
	}
	return testLogs[:], bridgeAddress, "GenerateTestWithoutRelevantLogs"
}

// GenerateTestData generates test log info for testing the fetcher of message data from L2L1/RollingHash logs
func GenerateLargeTest() ([]LogInfo, types.EthAddress, string) {
	// get sample topics and addresses
	topicsHi, topicsLo, addressHi, addressLo := GenerateTopicsAndAddresses()

	// fixed expected value for the topic on the first topic row for L2L1/RollingHash logs
	firstTopicL2L1Hi, firstTopicL2L1Lo := FirstTopicL2l1()
	firstTopicRollingHi, firstTopicRollingLo := FirstTopicRolling()

	// compute a dummy bridge address for testing
	bridgeAddress := types.DummyAddress(2)
	bridgeAddrHi, bridgeAddrLo := ConvertAddress(bridgeAddress)
	var testLogs = [...]LogInfo{
		{
			LogType:   2, // log with 2 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{2},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{topicsHi[0], topicsHi[1]},
			TopicsLo:  []field.Element{topicsLo[0], topicsLo[1]},
		},
		{
			LogType:   0, // log with 0 topics
			DataSize:  field.Element{21},
			noTopics:  field.Element{21},
			AddressHi: field.Element{121},
			AddressLo: field.Element{333},
			TopicsHi:  nil,
			TopicsLo:  nil,
		},
		{
			LogType:   3, // RollingHash log with 3 topics
			DataSize:  field.Element{3},
			noTopics:  field.Element{3},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicRollingHi, topicsHi[2], topicsHi[3]},
			TopicsLo:  []field.Element{firstTopicRollingLo, topicsLo[2], topicsLo[3]},
		},
		{
			LogType:   4, // an L2L1 log of the type we are interested in
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicL2L1Hi, topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicL2L1Lo, topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   4, // log of type 4 which has a wrong bridge data and will be skipped
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: addressHi[0],
			AddressLo: addressLo[0],
			TopicsHi:  []field.Element{firstTopicL2L1Hi, topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicL2L1Lo, topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   3, // log with 3 topics, but not a RollingHash log
			DataSize:  field.Element{191},
			noTopics:  field.Element{191},
			AddressHi: addressHi[3],
			AddressLo: addressLo[3],
			TopicsHi:  []field.Element{topicsHi[1], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{topicsLo[1], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   4, // an L2L1 log of the type we are interested in
			DataSize:  field.Element{34443},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicL2L1Hi, topicsHi[2], topicsHi[3], topicsHi[4]},
			TopicsLo:  []field.Element{firstTopicL2L1Lo, topicsLo[2], topicsLo[3], topicsLo[4]},
		},
		{
			LogType:   4, // log of type 4 which has a wrong first topic in TopicsHi/TopicsLo
			DataSize:  field.Element{213},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{topicsHi[3], topicsHi[3], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{topicsHi[3], topicsLo[3], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   4, // an L2L1 log of the type we are interested in
			DataSize:  field.Element{100},
			noTopics:  field.Element{4},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicL2L1Hi, topicsHi[0], topicsHi[4], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicL2L1Lo, topicsLo[0], topicsLo[4], topicsLo[5]},
		},
		{
			LogType:   MISSING_LOG, // missing log
			DataSize:  field.Element{},
			noTopics:  field.Element{},
			AddressHi: field.Element{},
			AddressLo: field.Element{},
			TopicsHi:  nil,
			TopicsLo:  nil,
		},
		{
			LogType:   3, // RollingHash log with 3 topics
			DataSize:  field.Element{312},
			noTopics:  field.Element{3},
			AddressHi: bridgeAddrHi,
			AddressLo: bridgeAddrLo,
			TopicsHi:  []field.Element{firstTopicRollingHi, topicsHi[0], topicsHi[5]},
			TopicsLo:  []field.Element{firstTopicRollingLo, topicsLo[0], topicsLo[5]},
		},
	}
	return testLogs[:], bridgeAddress, "GenerateLargeTest"
}

// test the fetcher of message data from L2L1/RollingHash logs
func TestLogsDataFetcher(t *testing.T) {

	var testGenVect = [...]func() ([]LogInfo, types.EthAddress, string){
		GenerateSimpleL2L1Test,
		GenerateSimpleRollingTest,
		GenerateTestWithoutRelevantLogs,
		GenerateLargeTest,
	}

	for i, GenerateTest := range testGenVect {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {
			testLogs, bridgeAddress, testName := GenerateTest()
			t.Logf("Test case explainer: %v", testName)
			// test log parsing
			for _, logInfo := range testLogs {
				if logInfo.LogType == LOG4 {
					log := logInfo.ConvertToL2L1Log()
					bridge.ParseL2L1Log(log)
				}
			}

			// compute the cummulative size of the LogColumns we will compute next
			fullSize := ComputeSize(testLogs[:])
			colSize := utils.NextPowerOfTwo(fullSize)

			// LogColumns and selectors on which we will run our tests
			var (
				logCols                                            LogColumns
				selectors                                          Selectors
				fetchedL2L1, fetchedRollingMsg, fetchedRollingHash ExtractedData
				hasherL2l1                                         LogHasher
				rollingSelector                                    RollingSelector
			)

			define := func(b *wizard.Builder) {
				// define mock log columns
				logCols = NewLogColumns(b.CompiledIOP, colSize, "MOCK")
				// initialize extracted data to be fetched from logs
				fetchedL2L1 = NewExtractedData(b.CompiledIOP, colSize, "L2L1LOGS")
				fetchedRollingMsg = NewExtractedData(b.CompiledIOP, colSize, "ROLLING_MSG")
				fetchedRollingHash = NewExtractedData(b.CompiledIOP, colSize, "ROLLING_HASH")
				// initialize selectors
				selectors = NewSelectorColumns(b.CompiledIOP, logCols, common.Address(bridgeAddress))
				// initialize hashers
				hasherL2l1 = NewLogHasher(b.CompiledIOP, colSize, "L2L1LOGS")
				// initialize rolling selector
				rollingSelector = NewRollingSelector(b.CompiledIOP, "ROLLING_SEL")
				// define extracted data from logs and associated filters
				DefineExtractedData(b.CompiledIOP, logCols, selectors, fetchedL2L1, L2L1)
				DefineExtractedData(b.CompiledIOP, logCols, selectors, fetchedRollingMsg, RollingMsgNo)
				DefineExtractedData(b.CompiledIOP, logCols, selectors, fetchedRollingHash, RollingHash)
				// define the L2L1 hasher
				DefineHasher(b.CompiledIOP, hasherL2l1, "L2L1LOGS", fetchedL2L1)
				// define the Rolling selector
				DefineRollingSelector(b.CompiledIOP, rollingSelector, "ROLLING_SEL", fetchedRollingHash, fetchedRollingMsg)
			}

			prove := func(run *wizard.ProverRuntime) {
				// assign the mock log columns
				LogColumnsAssign(run, &logCols, testLogs[:])
				// assign the selectors
				selectors.Assign(run)
				// assign the data extracted from logs
				AssignExtractedData(run, logCols, selectors, fetchedL2L1, L2L1)
				AssignExtractedData(run, logCols, selectors, fetchedRollingMsg, RollingMsgNo)
				AssignExtractedData(run, logCols, selectors, fetchedRollingHash, RollingHash)
				// assign the L2L1 hasher
				AssignHasher(run, hasherL2l1, fetchedL2L1)
				// assign the Rolling selector
				AssignRollingSelector(run, rollingSelector, fetchedRollingHash, fetchedRollingMsg)
			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}
		})
	}

}
