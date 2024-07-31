package statesummary

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mock"
)

// TestStateSummaryInternal tests only the StateSummary module internally, without any connectors
func TestStateSummaryInternal(t *testing.T) {

	tContext := common.InitializeContext(100)
	var ss Module

	for i, tCase := range tContext.TestCases {

		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			t.Logf("Test case explainer: %v", tCase.Explainer)

			define := func(b *wizard.Builder) {
				ss = NewModule(b.CompiledIOP, 1<<6)
			}

			prove := func(run *wizard.ProverRuntime) {

				var (
					initState    = tContext.State
					shomeiState  = mock.InitShomeiState(initState)
					stateLogs    = tCase.StateLogsGens(initState)
					shomeiTraces = mock.StateLogsToShomeiTraces(shomeiState, stateLogs)
				)

				ss.Assign(run, shomeiTraces)
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
