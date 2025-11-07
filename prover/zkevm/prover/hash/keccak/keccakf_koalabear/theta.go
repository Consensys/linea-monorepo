package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

// theta module, responsible for updating the state in the theta step of keccakf
type theta struct {
	// state before applying the theta step, in base clean 12
	stateCurr common.State
	// state after applying the theta step, in base dirty 12
	stateNext common.State
	// intermediate columns for state transition.
	msb [5][5][8]ifaces.Column // MSB of each byte of the state
	// lookup tables to attest the correctness of msb,
	// the first column is the byte, the second column is its msb
	lookupMSB [2]ifaces.Column
}

// newTheta creates a new theta module, declares the columns and constraints and returns its pointer
func newTheta(comp *wizard.CompiledIOP, maxNumKeccakf int, stateCurr common.State) *theta {
	// declare the columns
	declareColumnsTheta(comp, maxNumKeccakf)
	return &theta{}
}

// assignTheta assigns the values to the columns of theta step
func (theta *theta) assignTheta(run *wizard.ProverRuntime, stateCurr common.State) {
}

// it declares the intermediate columns generated during theta step, including the new state.
func declareColumnsTheta(comp *wizard.CompiledIOP, numKeccakf int) {
}
