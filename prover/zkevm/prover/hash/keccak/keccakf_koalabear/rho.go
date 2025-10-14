package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type roha struct {
	// state before applying the rho step, in base dirty 12
	stateCurr state
	// state after applying the rho step, in base clean 11.
	stateNext state
	// intermediate columns for state transition.
	decomposedTargetSlice [5][5][4]ifaces.Column // the decomposition of the slice that is cut by the rotation.
	// prover action to fill the decomposedTargetSlice

	// prover action to recompose the state after rotation
}

func newRoha(comp *wizard.CompiledIOP, numKeccakf int, stateCurr state) *roha {
	// declare the columns
	declareColumnsRho(comp, numKeccakf)
	// base conversion object to convert the state from base dirty 12 to base clean 11
	// at this step state has [stateBaseConversion] representation.
	_ = newBaseConversion(comp, numKeccakf, stateCurr)
	// apply rotation over state

	// return to the [state] representation via recomposition of slices

	return &roha{}
}
func assignRoha(run *wizard.ProverRuntime, stateCurr state) roha {
	return roha{stateCurr: stateCurr,
		stateNext: state{}}
}
func declareColumnsRho(comp *wizard.CompiledIOP, numKeccakf int) {
}
