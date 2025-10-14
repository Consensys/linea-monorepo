package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// theta module, responsible for updating the state in the theta step of keccakf
type theta struct {
	// state before applying the theta step, in base clean 12
	stateCurr state
	// state after applying the theta step, in base dirty 12
	stateNext state
	// intermediate columns for state transition.
	msb [5][5][8]ifaces.Column // MSB of each byte of the state
	// lookup tables to attest the correctness of msb,
	// the first column is the byte, the second column is its msb
	lookupMSB [2]ifaces.Column
}

func newTheta(comp *wizard.CompiledIOP, maxNumKeccakf int, stateCurr state) *theta {
	// declare the columns
	declareColumnsTheta(comp, maxNumKeccakf)
	return &theta{}
}
func assignTheta(run *wizard.ProverRuntime, stateCurr state) theta {
	return theta{stateCurr: stateCurr,
		stateNext: state{}}
}

// it declares the intermediate columns generated during theta step
func declareColumnsTheta(comp *wizard.CompiledIOP, numKeccakf int) {
}
