package accumulator

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
)

func accumulatorTestingModule(maxNumProofs int) (
	define wizard.DefineFunc,
	prover func(traces []statemanager.DecodedTrace) wizard.MainProverStep,
) {

	var mod Module

	// The testing wizard uniquely calls the accumulator module
	define = func(b *wizard.Builder) {
		mod = NewModule(b.CompiledIOP, Settings{
			Name:            "ACCUMULATOR_TEST",
			MaxNumProofs:    maxNumProofs,
			MerkleTreeDepth: 40,
		})
	}

	// And the prover (instanciated for traces) is called
	prover = func(
		traces []statemanager.DecodedTrace,
	) wizard.MainProverStep {
		return func(run *wizard.ProverRuntime) {
			// Assigns the module
			mod.Assign(run, traces)
		}
	}

	return define, prover
}

func TestShomei(t *testing.T) {
	// Generate trace file
	var decodedTrace []statemanager.DecodedTrace
	acc := statemanager.NewStorageTrie(statemanager.POSEIDON2_CONFIG, types.EthAddress{})
	traceInsert := acc.InsertAndProve(types.FullBytes32FromHex("0x32"), types.FullBytes32FromHex("0x12"))
	insertDecodedTrace := statemanager.DecodedTrace{
		Location: "this is the first trace",
		Type: 2,
		Underlying: traceInsert,
		IsSkipped: true,
	}
	decodedTrace = append(decodedTrace, insertDecodedTrace)
	definer, prover := accumulatorTestingModule(1024)
	comp := wizard.Compile(definer, dummy.Compile)
	proof := wizard.Prove(comp, prover(decodedTrace))
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid accumulator proof")
}
