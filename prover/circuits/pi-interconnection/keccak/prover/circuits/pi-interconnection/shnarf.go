package pi_interconnection

import (
	"github.com/consensys/gnark/frontend"
)

type ShnarfIteration struct {
	BlobDataSnarkHash                          [32]frontend.Variable
	NewStateRootHash                           [32]frontend.Variable
	EvaluationPointBytes, EvaluationClaimBytes [32]frontend.Variable
}

func (i *ShnarfIteration) SetZero() {
	for j := range i.EvaluationClaimBytes {
		i.NewStateRootHash[j] = 0
		i.EvaluationPointBytes[j] = 0
		i.EvaluationClaimBytes[j] = 0
		i.BlobDataSnarkHash[j] = 0
	}
}
