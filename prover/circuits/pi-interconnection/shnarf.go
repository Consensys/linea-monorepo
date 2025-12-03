package pi_interconnection

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
)

type ShnarfIteration struct {
	BlobDataSnarkHash                          [32]zk.WrappedVariable
	NewStateRootHash                           [32]zk.WrappedVariable
	EvaluationPointBytes, EvaluationClaimBytes [32]zk.WrappedVariable
}

// ComputeShnarfs DOES NOT check nbShnarfs ≤ len(s.Iterations)
func ComputeShnarfs(h keccak.BlockHasher, parent [32]zk.WrappedVariable, iterations []ShnarfIteration) (result [][32]zk.WrappedVariable) {
	result = make([][32]zk.WrappedVariable, len(iterations))
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
