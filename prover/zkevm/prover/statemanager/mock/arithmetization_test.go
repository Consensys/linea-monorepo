package mock

import (
	"testing"
)

// CheckDifferentVectorLengths will check that all vectors have the same length
// it outputs true if the vectors are of different lengths
// it outputs false if all vectors are of the same length
func CheckDifferentVectorLengths(v *StateManagerVectors) bool {
	lengthSet := map[int]struct{}{}
	lengthSet[len(v.Address)] = struct{}{}
	lengthSet[len(v.AddressHI)] = struct{}{}
	lengthSet[len(v.AddressLO)] = struct{}{}
	lengthSet[len(v.Nonce)] = struct{}{}
	lengthSet[len(v.NonceNew)] = struct{}{}
	lengthSet[len(v.MimcCodeHash)] = struct{}{}
	lengthSet[len(v.MimcCodeHashNew)] = struct{}{}
	lengthSet[len(v.CodeHashHI)] = struct{}{}
	lengthSet[len(v.CodeHashLO)] = struct{}{}
	lengthSet[len(v.CodeHashHINew)] = struct{}{}
	lengthSet[len(v.CodeHashLONew)] = struct{}{}
	lengthSet[len(v.CodeSizeOld)] = struct{}{}
	lengthSet[len(v.CodeSizeNew)] = struct{}{}
	lengthSet[len(v.BalanceOld)] = struct{}{}
	lengthSet[len(v.BalanceNew)] = struct{}{}
	lengthSet[len(v.KeyHI)] = struct{}{}
	lengthSet[len(v.KeyLO)] = struct{}{}
	lengthSet[len(v.ValueHICurr)] = struct{}{}
	lengthSet[len(v.ValueLOCurr)] = struct{}{}
	lengthSet[len(v.ValueHINext)] = struct{}{}
	lengthSet[len(v.ValueLONext)] = struct{}{}
	lengthSet[len(v.DeploymentNumber)] = struct{}{}
	lengthSet[len(v.DeploymentNumberInf)] = struct{}{}
	lengthSet[len(v.BlockNumber)] = struct{}{}
	lengthSet[len(v.Exists)] = struct{}{}
	lengthSet[len(v.ExistsNew)] = struct{}{}
	lengthSet[len(v.PeekAtAccount)] = struct{}{}
	lengthSet[len(v.PeekAtStorage)] = struct{}{}
	lengthSet[len(v.FirstAOC)] = struct{}{}
	lengthSet[len(v.LastAOC)] = struct{}{}
	lengthSet[len(v.FirstKOC)] = struct{}{}
	lengthSet[len(v.LastKOC)] = struct{}{}

	lengthSet[len(v.FirstAOCBlock)] = struct{}{}
	lengthSet[len(v.LastAOCBlock)] = struct{}{}
	lengthSet[len(v.FirstKOCBlock)] = struct{}{}
	lengthSet[len(v.LastKOCBlock)] = struct{}{}
	lengthSet[len(v.MaxDeploymentBlock)] = struct{}{}
	lengthSet[len(v.MinDeploymentBlock)] = struct{}{}

	return len(lengthSet) > 1
}

// TestArithmetization uses an initial State and a StateLogBuilder to generate frames for different scenarios.
// It then runs the Stitcher on the resulting traces and checks if the Stitcher outputs columns of the same size without failing
func TestArithmetization(t *testing.T) {

	tContext := InitializeContext()
	for i := range tContext.tMessages {
		t.Run(tContext.tMessages[i], func(t *testing.T) {
			frames := tContext.tFunc[i](t, tContext)
			var stitcher Stitcher
			stitcher.Initialize(tContext.blockNumber, tContext.state)
			for index := range frames {
				for _, frame := range frames[index] {
					stitcher.AddFrame(frame)
				}
			}

			acpVectors := stitcher.Finalize(GENERATE_ACP_SAMPLE)
			if CheckDifferentVectorLengths(acpVectors) {
				t.Fatalf("Vectors do not have the same length")
			}

			scpVectors := stitcher.Finalize(GENERATE_SCP_SAMPLE)
			if CheckDifferentVectorLengths(scpVectors) {
				t.Fatalf("Vectors do not have the same length")
			}
		})
	}
}
