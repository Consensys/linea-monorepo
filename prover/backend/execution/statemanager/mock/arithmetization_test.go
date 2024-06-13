package mock

import (
	"testing"
)

// CheckDifferentVectorLengths will check that all vectors have the same length
// it outputs true if the vectors are of different lengths
// it outputs false if all vectors are of the same length
func CheckDifferentVectorLengths(v *StateManagerVectors) bool {
	length := len(v.address)
	if length != len(v.addressHI) || length != len(v.addressLO) || length != len(v.nonce) || length != len(v.nonceNew) {
		return true
	}
	if length != len(v.mimcCodeHash) || length != len(v.mimcCodeHashNew) {
		return true
	}
	if length != len(v.codeHashHI) || length != len(v.codeHashLO) || length != len(v.codeHashHINew) || length != len(v.codeHashLONew) {
		return true
	}
	if length != len(v.codeSizeOld) || length != len(v.codeSizeNew) || length != len(v.balanceOld) || length != len(v.balanceNew) {
		return true
	}
	if length != len(v.keyHI) || length != len(v.keyLO) {
		return true
	}
	if length != len(v.valueHICurr) || length != len(v.valueLOCurr) || length != len(v.valueHINext) || length != len(v.valueLONext) {
		return true
	}
	if length != len(v.deploymentNumber) || length != len(v.deploymentNumberInf) || length != len(v.blockNumber) {
		return true
	}
	if length != len(v.exists) || length != len(v.existsNew) || length != len(v.peekAtAccount) || length != len(v.peekAtStorage) {
		return true
	}
	if length != len(v.firstAOC) || length != len(v.lastAOC) || length != len(v.firstKOC) || length != len(v.lastKOC) {
		return true
	}
	return false // false when all vectors have the same length
}

// TestArithmetization uses an inital State and a StateLogBuilder to generate frames for different scenarios.
// It then runs the Stitcher on the resulting traces and checks if the Stitcher outputs columns of the same size without failing
func TestArithmetization(t *testing.T) {

	tContext := InitializeContext()
	for i := range tContext.tMessages {
		t.Run(tContext.tMessages[i], func(t *testing.T) {
			frames := tContext.tFunc[i](t, tContext)
			var stitcher Stitcher
			stitcher.InitializeFromState(tContext.blockNumber, tContext.state)
			for index := range frames {
				for _, frame := range frames[index] {
					stitcher.AddFrame(frame)
				}
			}

			columns := stitcher.Finalize()
			if CheckDifferentVectorLengths(columns) {
				t.Fatalf("Vectors do not have the same length")
			}
		})
	}
}
