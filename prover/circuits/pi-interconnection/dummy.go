package pi_interconnection

import (
	"github.com/consensys/gnark/frontend"
)

type DummyCircuit FunctionalPublicInput

func (c *DummyCircuit) Define(api frontend.API) error {
	toCommit := make([]frontend.Variable, 0, len(c.AggregationPublicInput)+len(c.DecompressionPublicInput)+len(c.ExecutionPublicInput))
	toCommit = append(toCommit, c.AggregationPublicInput[:]...)
	toCommit = append(toCommit, c.DecompressionPublicInput...)
	toCommit = append(toCommit, c.ExecutionPublicInput...)
	commitment, err := api.(frontend.Committer).Commit(toCommit...)
	if err != nil {
		return err
	}
	api.AssertIsDifferent(commitment, 0)
	return nil
}
