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
