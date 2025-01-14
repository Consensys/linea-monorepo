//go:build !fuzzlight

package ecpair

import (
	"testing"
)

func TestPairingDataCircuit(t *testing.T) {
	for _, tc := range pairingDataTestCases {
		testModule(t, tc, true, false, true, false)
	}
}

func TestMembershipCircuit(t *testing.T) {
	for _, tc := range membershipTestCases {
		testModule(t, tc, false, true, false, true)
	}
}

func TestGeneratedData(t *testing.T) {
	var generatedData []pairingDataTestCase
	filepath.Walk("testdata/generated", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "input.csv") {
			tc := pairingDataTestCase{
				InputFName:       path,
				NbMillerLoops:    5,
				NbFinalExps:      1,
				NbSubgroupChecks: 5,
			}
			generatedData = append(generatedData, tc)
		}
		return nil
	})
	for _, tc := range generatedData {
		testModule(t, tc, true, true, true, true)
	}
}
