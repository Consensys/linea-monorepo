package protocols

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// each lane is 64 bits, represented as 8 bytes.
type lane = [8]ifaces.Column

// keccakf state is a 5x5 matrix of lanes.
type state = [5][5]lane

// state after each base conversion, each lane is decomposed into 16 slices of 4 bits each.
type stateIn4Bits = [5][5][16]ifaces.Column

// baseConversion module, responsible for converting the state from base dirty 12 to base 2.
type baseConversion struct {
	// state before applying the base conversion step, in base dirty 12
	stateCurr state
	// state after applying the base conversion step, in base 2.
	StateNext stateIn4Bits
	// lookup tables to attest the correctness of base conversion,
	// the first column is the 4 bits slice in base 12, the second column is its representation in base 2.
	lookupTable [2]ifaces.Column
	// prover action to decompose current state into slices of 4 bits.
	paStateDecomposition *wizard.ProverAction
}

// newBaseConversion creates a new base conversion module, declares the columns and constraints and returns its pointer
func NewBaseConversion(comp *wizard.CompiledIOP, numKeccakf int, stateCurr [5][5]lane) *baseConversion {
	// declare the columns
	declareColumnsBaseConv(comp, numKeccakf)

	return &baseConversion{}
}

// assignBaseConversion assigns the values to the columns of base conversion step.
func (bc *baseConversion) Run(run *wizard.ProverRuntime) baseConversion {
	return baseConversion{stateCurr: bc.stateCurr}
}

// it declares the intermediate columns generated during base conversion step, including the new state.
func declareColumnsBaseConv(comp *wizard.CompiledIOP, numKeccakf int) {
}
