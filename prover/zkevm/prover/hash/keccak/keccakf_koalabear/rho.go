package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// rho module, responsible for updating the state in the rho step of keccakf
type rho struct {
	// state before applying the rho step, in base dirty 12
	stateCurr state
	// state after applying the rho step, in base clean 11.
	stateNext state
	// intermediate columns for state transition.
	decomposedTargetSlice [5][5][4]ifaces.Column // the decomposition of the slice that is cut by the rotation.
	// prover action to fill the decomposedTargetSlice

	// prover action to recompose the state after rotation
}

// newRoha creates a new rhoa module, declares the columns and constraints and returns its pointer
func newRho(comp *wizard.CompiledIOP, numKeccakf int, stateCurr state) *rho {
	// declare the columns
	declareColumnsRho(comp, numKeccakf)
	// base conversion object to convert the state from base dirty 12 to base clean 11
	// at this step state has [stateBaseConversion] representation.
	_ = newBaseConversion(comp, numKeccakf, stateCurr)
	// apply rotation over state

	// return to the [state] representation via recomposition of slices

	return &rho{}
}

// assignRoha assigns the values to the columns of rho step.
func assignRoh(run *wizard.ProverRuntime, stateCurr state) rho {
	return rho{stateCurr: stateCurr,
		stateNext: state{}}
}

// it declares the intermediate columns generated during rho step, including the new state.
func declareColumnsRho(comp *wizard.CompiledIOP, numKeccakf int) {
}
