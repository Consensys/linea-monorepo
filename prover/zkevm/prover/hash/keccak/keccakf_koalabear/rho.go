package keccakfkoalabear

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bits"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	protocols "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/sub_protocols"
)

// rho module, responsible for updating the state in the rho step of keccakf
type rho struct {
	// state before applying the rho step, in base dirty 12
	stateCurr state
	// state after applying the rho step, in base clean 11.
	stateNext state
	// prover action for base conversion to binary
	paBaseConversion wizard.ProverAction
	// prover actions for bit decomposition of binary state
	paBitDecomposition [5][5][16]wizard.ProverAction
	// state after bit rotation, in bits
	stateBitRotation *stateInBits
	// prover actions for recomposition to base clean 11
	paRecomposition [5][5][8]wizard.ProverAction
}

// newRoha creates a new rho module, declares the columns and constraints and returns its pointer
func newRho(comp *wizard.CompiledIOP, numKeccakf int, stateCurr state) *rho {

	var (
		bitDecompose [5][5][16]*bits.BitDecomposed
	)
	rho := &rho{
		stateCurr: stateCurr,
	}
	// base conversion object to convert the state from base dirty 12 to base 2.
	// at this step state has [stateBaseConversion] representation.
	bc := protocols.NewBaseConversion(comp, numKeccakf, stateCurr)
	// decompose the state into  bits
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 16; z++ {
				// decompose each 4-bit slice into bits
				bitDecompose[x][y][z] = bits.BitDecompose(comp, bc.StateNext[x][y][z], 4)
				rho.paBitDecomposition[x][y][z] = bitDecompose[x][y][z]
			}
			// perform the rotation on the current lane and update the new state
			rotation(rho.stateBitRotation, x, y, bitDecompose[x][y])

			// return to the [state] representation in base clean 11 via recomposition of slices.
			for z := 0; z < 64; z = +8 {
				name := fmt.Sprintf("%v_%v_%v", x, y, z/8)
				br := protocols.LinearCombination(comp, name, rho.stateBitRotation[x][y][z:z+8], 11)
				rho.paRecomposition[x][y][z/8] = br
				// set the corresponding column of rho.stateNext
				rho.stateNext[x][y][z/8] = br.CombinationRes

			}

		}
	}
	return rho
}

// assignRho assigns the values to the columns of rho step.
func (rho *rho) assignRoh(run *wizard.ProverRuntime, stateCurr state) {
	// take the current state, cut each lane into 4-bit slices and convert from base dirty 12 to base 2.
	rho.paBaseConversion.Run(run)
	// decompose each 4-bit slice into bits
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 16; z++ {
				rho.paBitDecomposition[x][y][z].Run(run)
			}
		}
	}
	// recomposition: after rotation of bits,
	// each 8 bits are recomposed into a base clean 11 slice.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 64; z = +8 {
				rho.paRecomposition[x][y][z/8].Run(run)
			}
		}
	}
}

// rotation performs the bit rotation in the rho step of keccakf
func rotation(stateBitRotation *stateInBits, x, y int, bitDec [16]*bits.BitDecomposed) {
	for z := 0; z < 16; z++ {
		bits := bitDec[z].Bits
		// find the new position of each bit after rotation
		for i := 0; i < 4; i++ {
			newPos := (4*z + i + keccak.LR[x][y]) % 64
			// assign the bit to the new position
			stateBitRotation[x][y][newPos] = bits[i]
		}
	}
}
