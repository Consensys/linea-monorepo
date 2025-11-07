package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

// inputs to the keccakf module
type keccakfInputs struct {
	// the state before the keccakf round
	// Note : unlike the original keccakf where the initial State is zero,
	// the initial State here is the first block of the message.
	state common.State
	// the blocks of the message to be absorbed.
	// first blocks of messages are located in positions 0 mod 24 and are represented in base clean 12,
	// other blocks of message are located in positions 23 mod 24 and are represented in base clean 11.
	// otherwise the blocks are zero.
	blocks [common.NumLanesInBlock]common.Lane
	// flag indicating if it is the first block of the message
	isFirstBlock ifaces.Column
	// flag indicating if it is a block of the message
	isBlock ifaces.Column
	// isActive is activation column of the module. namely, it is 1 for rows where the keccakf is active, and 0 otherwise.
	isActive ifaces.Column
}

// Wizard module responsible for proving a sequence of keccakf permutation
type Module struct {
	// maximal number of Keccakf permutation that the module can handle
	maxNumKeccakf int
	// inputs to the keccakf module
	inputs keccakfInputs
	// theta module, responsible for updating the state in the theta step of keccakf
	theta *theta
	// rho pi module, responsible for updating the state in the rho and pi steps of keccakf
	RhoPi *rho
	// chi module, responsible for updating the state in the chi step of keccakf
	ChiIota *chi
	// iota to theta module, responsible for bringing back the state from [ChiIota] to the representation proper for theta step.
}

// NewModule creates a new keccakf module, declares the columns and constraints and returns its pointer
func NewModule(comp *wizard.CompiledIOP, numKeccakf int, inp keccakfInputs) *Module {

	// declare the columns
	declareColumns(comp, numKeccakf)

	// constraints over inputs
	// isFirstBlock is boolean
	commonconstraints.MustBeBinary(comp, inp.isFirstBlock)
	// isBlock is boolean
	commonconstraints.MustBeBinary(comp, inp.isBlock)
	// isActive is activation column of the module
	commonconstraints.MustBeActivationColumns(comp, inp.isActive)
	// must be zero when inactive
	commonconstraints.MustZeroWhenInactive(comp, inp.isActive,
		inp.isBlock,
		inp.isFirstBlock,
	)
	// when isFirstBlock is 1, isBlock must be 1
	comp.InsertGlobal(0, ifaces.QueryID("FIRST_BLOCK_IMPLIES_IS_BLOCK"),
		sym.Mul(inp.isFirstBlock, sym.Sub(1, inp.isBlock)),
	)

	// assign the first blocks of the message to the state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			m := 5*y + x
			if m < common.NumLanesInBlock {
				for z := 0; z < common.NumSlices; z++ {
					// if it is the first block of the message, assign it to the state
					comp.InsertGlobal(0, ifaces.QueryIDf("STATE_IS_SET_TO_FIRST_BLOCK_%v_%v_%v", x, y, z),
						sym.Mul(inp.isFirstBlock,
							sym.Sub(inp.state[x][y][z], inp.blocks[m][z]),
						),
					)
				}
			} else {
				for z := 0; z < common.NumSlices; z++ {
					//  the remaining columns of the state are set to zero
					comp.InsertGlobal(0, ifaces.QueryIDf("STATE_IS_SET_TO_ZERO_%v,%v_%v", x, y, z),
						sym.Mul(inp.isFirstBlock, inp.state[x][y][z]),
					)
				}
			}
		}
	}

	// create the theta module with the state including the message-blocks
	theta := newTheta(comp, numKeccakf, inp.state)
	// create the rho module with the state after theta
	//rho := newRho(comp, numKeccakf)

	return &Module{
		maxNumKeccakf: numKeccakf,
		inputs:        inp,
		theta:         theta,
		//	RhoPi:         rho,
	}
}

// Assign the values to the columns of the keccakf module.
func (m *Module) Assign(run *wizard.ProverRuntime, numKeccakf int, blocks [common.NumLanesInBlock]common.Lane) {
	// initial state is zero
	var state common.State
	// assign the blocks of the message to the state

	// assign the theta module with the state including the message-blocks
	m.theta.assignTheta(run, state)
	// assign the base conversion module with the state after theta
	// m.IotaToTheta.Run(run)
	// assign the rho pi module with the state after theta
	// m.RhoPi.assignRoh(run, m.bc.StateNext)

}

// it declares the columns used in the keccakf module, including the state and the message blocks.
func declareColumns(comp *wizard.CompiledIOP, maxNumKeccakf int) {
}

// Returns the number of rows required to prove `numKeccakf` calls to the
// permutation function. The result is padded to the next power of 2 in order to
// satisfy the requirements of the Wizard to have only powers of 2.
func numRows(numKeccakf int) int {
	return utils.NextPowerOfTwo(numKeccakf * common.NumRounds)
}
