package accumulatorsummary

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mock"
)

// TestAccumulatorConnector checks the data consistency between AccumulatorSummary and StateSummary
func TestAccumulatorConnector(t *testing.T) {

	tContext := common.InitializeContext(100)

	for i, tCase := range tContext.TestCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			t.Logf("Test case explainer: %v", tCase.Explainer)

			var (
				ss         statesummary.Module
				mod        accumulator.Module
				accSummary *Module
			)

			define := func(b *wizard.Builder) {
				ss = statesummary.NewModule(b.CompiledIOP, 1<<6)

				mod = accumulator.NewModule(
					b.CompiledIOP,
					accumulator.Settings{
						MaxNumProofs:    1024,
						MerkleTreeDepth: 40,
						Name:            "ACCUMULATOR_TEST",
					})

				accSummary = NewModule(b.CompiledIOP,
					Inputs{
						Name:        "ACCUMULATOR_SUMMARY_TEST",
						Accumulator: mod,
					},
				).ConnectToStateSummary(b.CompiledIOP, &ss)
			}

			prove := func(run *wizard.ProverRuntime) {

				var (
					initState    = tContext.State
					shomeiState  = mock.InitShomeiState(initState)
					stateLogs    = tCase.StateLogsGens(initState)
					shomeiTraces = mock.StateLogsToShomeiTraces(shomeiState, stateLogs)
				)

				ss.Assign(run, shomeiTraces)
				var compressedTraces []statemanager.DecodedTrace //nolint:prealloc // test function
				for _, blockTraces := range shomeiTraces {
					compressedTraces = append(compressedTraces, blockTraces...)
				}

				mod.Assign(run, compressedTraces)
				accSummary.Assign(run)
			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}
		})
	}

}
