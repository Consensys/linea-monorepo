package pi_interconnection

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
)

type ShnarfIteration struct {
	BlobDataSnarkHash                          [32]T
	NewStateRootHash                           [32]T
	EvaluationPointBytes, EvaluationClaimBytes [32]T
}

// ComputeShnarfs DOES NOT check nbShnarfs ≤ len(s.Iterations)
func ComputeShnarfs(h keccak.BlockHasher, parent [32]T, iterations []ShnarfIteration) (result [][32]T) {
	result = make([][32]T, len(iterations))
	prevShnarf := parent

	for i, t := range iterations {
		result[i] = h.Sum(nil, prevShnarf, t.BlobDataSnarkHash, t.NewStateRootHash, t.EvaluationPointBytes, t.EvaluationClaimBytes)
		prevShnarf = result[i]
	}

	return
}

func (i *ShnarfIteration) SetZero() {
	for j := range i.EvaluationClaimBytes {
		i.NewStateRootHash[j] = 0
		i.EvaluationPointBytes[j] = 0
		i.EvaluationClaimBytes[j] = 0
		i.BlobDataSnarkHash[j] = 0
	}
}
