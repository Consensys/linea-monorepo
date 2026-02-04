package logs

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/execution/bridge"
	eth "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
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
