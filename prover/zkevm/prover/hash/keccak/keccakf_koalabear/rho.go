package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// rho module, responsible for updating the state in the rho step of keccakf
type rho struct {
	// state before applying the rho step
	stateCurr stateInBits
	// state after bit rotation, in bits
	stateNext *stateInBits
	// prover actions for recomposition to base clean 11
	// paRecomposition [5][5][8]wizard.ProverAction
}

// newRho creates a new rho module, declares the columns and constraints and returns its pointer
func newRho(comp *wizard.CompiledIOP, numKeccakf int, stateCurr stateInBits) *rho {

	rho := &rho{
		stateCurr: stateCurr,
		stateNext: &stateInBits{},
	}
	// declare the columns for the new state after rho
	// size := numRows(numKeccakf * 24)

	/*for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 64; z++ {
				rho.stateNext[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("RHO_STATE_NEXT_%v_%v_%v", x, y, z), size)
			}
		}
	}*/

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 64; z++ {
				// find the new position of each bit after rotation
				newPos := (z + keccak.LR[x][y]) % 64
				// assign the bit to the new position
				rho.stateNext[y][(2*x+3*y)%5][newPos] = stateCurr[x][y][z]
			}

			/*
				// return to the [state] representation in base clean 11 via recomposition of slices.
				for z := 0; z < 64; z = +8 {
					name := fmt.Sprintf("%v_%v_%v", x, y, z/8)
					br := protocols.LinearCombination(comp, name, rho.stateBitRotation[x][y][z:z+8], 11)
					rho.paRecomposition[x][y][z/8] = br
					// set the corresponding column of rho.stateNext
					rho.stateNext[x][y][z/8] = br.CombinationRes

				}*/

		}
	}
	return rho
}

// assignRho assigns the values to the columns of rho step.
func (rho *rho) assignRoh(run *wizard.ProverRuntime, stateCurr stateInBits) {
}

// rotation performs the bit rotation in the rho step of keccakf
func rotation(stateBitRotation *stateInBits, x, y int, bitDec [64]ifaces.Column) {

}
