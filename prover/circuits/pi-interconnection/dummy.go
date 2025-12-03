package pi_interconnection

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
)

type DummyCircuit struct {
	AggregationPublicInput   [2]zk.WrappedVariable `gnark:",public"` // the public input of the aggregation circuit; divided big-endian into two 16-byte chunks
	ExecutionPublicInput     []zk.WrappedVariable  `gnark:",public"`
	DecompressionPublicInput []zk.WrappedVariable  `gnark:",public"`

	NbExecution     zk.WrappedVariable
	NbDecompression zk.WrappedVariable

	DecompressionFPI []zk.WrappedVariable
	ExecutionFPI     []zk.WrappedVariable
}

// x -> x^5 to match the dummycircuit package
func sum(api frontend.API, pi zk.WrappedVariable) zk.WrappedVariable {
	p2 := api.Mul(pi, pi)
	return api.Mul(p2, p2, pi)
}

func (c *DummyCircuit) Define(api frontend.API) error {
	api.AssertIsEqual("0x0102030405060708090a0b0c0d0e0f10", c.AggregationPublicInput[0])
	api.AssertIsEqual("0x1112131415161718191a1b1c1d1e1f20", c.AggregationPublicInput[1])

	checkFPI := func(n zk.WrappedVariable, fpi, pi []zk.WrappedVariable) {
		api.AssertIsEqual(len(fpi), len(pi))
		r := internal.NewRange(api, n, len(fpi))
		for i := range fpi {
			r.AssertEqualI(i, fpi[i], sum(api, pi[i]))
		}
	}

	checkFPI(c.NbExecution, c.ExecutionFPI, c.ExecutionPublicInput)
	checkFPI(c.NbDecompression, c.DecompressionFPI, c.DecompressionPublicInput)

	challenge, err := api.(frontend.Committer).Commit(c.AggregationPublicInput[0]) // dummy commitment for aggregation to work
	if err != nil {
		return err
	}
	api.AssertIsDifferent(challenge, 0)

	return nil
}
