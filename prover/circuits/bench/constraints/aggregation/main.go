package main

import (
	"context"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/zkevm-monorepo/prover/circuits"
	"github.com/consensys/zkevm-monorepo/prover/circuits/aggregation"
	"github.com/consensys/zkevm-monorepo/prover/circuits/dummy"
	pi_interconnection "github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/zkevm-monorepo/prover/utils/test_utils"
	"github.com/stretchr/testify/assert"
)

func main() {

	const nbCircuits = 400

	piCircuit := pi_interconnection.DummyCircuit{
		ExecutionPublicInput:     make([]frontend.Variable, nbCircuits),
		DecompressionPublicInput: make([]frontend.Variable, nbCircuits),
		DecompressionFPI:         make([]frontend.Variable, nbCircuits),
		ExecutionFPI:             make([]frontend.Variable, nbCircuits),
	}

	var t test_utils.FakeTestingT

	innerField := ecc.BLS12_377.ScalarField()
	piCs, err := frontend.Compile(innerField, scs.NewBuilder, &piCircuit)
	assert.NoError(t, err)

	srsProvider := circuits.NewUnsafeSRSProvider()

	piSetup, err := circuits.MakeSetup(context.TODO(), "public-input-interconnection", piCs, srsProvider, nil)
	assert.NoError(t, err)

	// Having more than 2 types of inner circuit to reflect various versions
	innerVks := make([]plonk.VerifyingKey, 10)
	for i := range innerVks {
		setup, err := dummy.MakeUnsafeSetup(srsProvider, 0, innerField)
		assert.NoError(t, err)
		innerVks[i] = setup.VerifyingKey
	}

	p := profile.Start(profile.WithPath("aggregation.pprof"))
	defer p.Stop()
	_, err = aggregation.MakeCS(nbCircuits, piSetup, innerVks)
	assert.NoError(t, err)
}
