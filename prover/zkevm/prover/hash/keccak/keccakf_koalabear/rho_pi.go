package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// rho module, responsible for updating the state in the rho step of keccakf
type rho struct {
	// state before applying the rho step
	stateCurr stateInBits
	// state after bit rotation, in bits
	stateNext *stateInBits
}

// newRho creates a new rho module, declares the columns and constraints and returns its pointer
func newRho(comp *wizard.CompiledIOP, numKeccakf int, stateCurr stateInBits) *rho {

	rho := &rho{
		stateCurr: stateCurr,
		stateNext: &stateInBits{},
	}

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 64; z++ {
				// find the new position of each bit after rotation
				newPos := (z + keccak.LR[x][y]) % 64
				// assign the bit to the new position
				rho.stateNext[y][(2*x+3*y)%5][newPos] = stateCurr[x][y][z]
			}
		}
	}
	return rho
}

// assignRho assigns the values to the columns of rho step.
func (rho *rho) assignRoh(run *wizard.ProverRuntime, stateCurr stateInBits) {
	// it does noting as it is just rotation and shuffleing of columns and does not creat any new columns.
}
