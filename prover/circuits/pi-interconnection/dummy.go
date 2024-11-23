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
	// defining constraints to make sure none of the Plonk selector columns are zero
	// this is needed for the incomplete arithmetic formulas in the emulated Plonk verifier to work
	x := api.Add(toCommit[0], commitment, 1) // Ql, Qr, Qo, Qc ≠ 0
	api.AssertIsDifferent(x, 0)              // Qm ≠ 0
	return nil
}
