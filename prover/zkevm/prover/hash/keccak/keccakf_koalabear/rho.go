package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bits"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
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
	// prover action for recomposition of state in base clean 11.
	paStateRecomposition wizard.ProverAction
}

// newRoha creates a new rho module, declares the columns and constraints and returns its pointer
func newRho(comp *wizard.CompiledIOP, numKeccakf int, stateCurr state) *rho {

	var (
		bitDecompose     *bits.BitDecomposed
		stateBitRotation stateInBits // state after bit rotation, in bits
	)
	rho := &rho{
		stateCurr: stateCurr,
	}
	// declare the columns
	declareColumnsRho(comp, numKeccakf)
	// base conversion object to convert the state from base dirty 12 to base 2.
	// at this step state has [stateBaseConversion] representation.
	bc := protocols.NewBaseConversion(comp, numKeccakf, stateCurr)
	// decompose the state into  bits
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 16; z++ {
				// decompose each 4-bit slice into bits
				bitDecompose = bits.BitDecompose(comp, bc.StateNext[x][y][z], 4)
				rho.paBitDecomposition[x][y][z] = bitDecompose
				// perform the rotation and assign to the new state
				rotation(&stateBitRotation, x, y, z, bitDecompose)
			}
			// return to the [state] representation in base clean 11 via recomposition of slices.
			for z := 0; z < 64; z = +8 {
				stateNext := BaseRecompose(stateBitRotation[x][y][z:z+8], 11)
				// check  that rho state is set to stateNext
				comp.InsertGlobal(0, ifaces.QueryIDf("RHO_NEW_STATE_IS-CORRECT_%d_%d_%d", x, y, z/8),
					symbolic.Sub(rho.stateNext[x][y][z/8], stateNext))

			}

		}
	}
	return rho
}

// assignRho assigns the values to the columns of rho step.
func (rho *rho) assignRoh(run *wizard.ProverRuntime, stateCurr state) {
	rho.paBaseConversion.Run(run)
}

// it declares the intermediate columns generated during rho step, including the new state.
func declareColumnsRho(comp *wizard.CompiledIOP, numKeccakf int) {
}

func rotation(stateBitRotation *stateInBits, x, y, z int, bitDec *bits.BitDecomposed) {
	var (
		bits = bitDec.Bits
		LR   = keccak.LR[x][y]
	)
	// find the new position of each bit after rotation
	for i := 0; i < 4; i++ {
		newPos := (4*z + i + LR) % 64
		bit := bits[i]
		// assign the bit to the new position
		stateBitRotation[x][y][newPos] = bit
	}
}
