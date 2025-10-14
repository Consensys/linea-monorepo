package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type baseConversion struct {
	// state before applying the base conversion step, in base dirty 12
	stateCurr state
	// state after applying the base conversion step, in base clean 11.
	stateNext stateBaseConversion
	// lookup tables to attest the correctness of base conversion,
	// the first column is the 4 bits slice in base 12, the second column is its representation in base 11.
	lookupTable [2]ifaces.Column
	// prover action to decompose current state into slices of 4 bits.
	paStateDecomposition *wizard.ProverAction
}

func newBaseConversion(comp *wizard.CompiledIOP, numKeccakf int, stateCurr state) *baseConversion {
	// declare the columns
	declareColumnsBaseConv(comp, numKeccakf)

	return &baseConversion{}
}

func assignBaseConversion(run *wizard.ProverRuntime, stateCurr state) baseConversion {
	return baseConversion{stateCurr: stateCurr,
		stateNext: stateBaseConversion{}}
}
func declareColumnsBaseConv(comp *wizard.CompiledIOP, numKeccakf int) {
}
