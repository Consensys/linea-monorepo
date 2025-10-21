package keccakfkoalabear

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

// theta module, responsible for updating the state in the theta step of keccakf
type theta struct {
	// state before applying the theta step, in base clean 12
	stateCurr state
	// state after applying the theta step, in base dirty 12
	stateNext state
	// intermediate columns for state transition.
	// msb of each byte of the stateNext
	msb [5][5][8]ifaces.Column // MSB of each byte of the state
	// lookup tables to attest the correctness of msb,
	// the first column is the byte, the second column is its msb
	lookupMSB [2]ifaces.Column
}

// newTheta creates a new theta module, declares the columns and constraints and returns its pointer
func (*theta) newTheta(comp *wizard.CompiledIOP,
	numKeccakf int,
	round int,
	stateCurr state) *theta {
	res := &theta{}
	res.stateCurr = stateCurr

	// declare the columns
	res.declareColumnsTheta(comp, numKeccakf, round)

	// declare the constraints
	res.csEqThetaBaseA(comp, round)
	// res.csAThetaDecomposition(comp, round)
	return res
}

// declareColumnsTheta declares the intermediate columns generated during theta step, including the new state.
func (theta *theta) declareColumnsTheta(comp *wizard.CompiledIOP, numKeccakf int, round int) {
	// size of the columns to declare
	colSize := keccakf.NumRows(numKeccakf)
	// declare the new state
	theta.stateNext = state{}
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				theta.stateNext[x][y][z] = comp.InsertCommit(
					round,
					keccakf.DeriveName("A_THETA", x, y, z),
					colSize,
				)
				theta.msb[x][y][z] = comp.InsertCommit(
					round,
					keccakf.DeriveName("MSB_A_THETA", x, y, z),
					colSize,
				)
			}
		}
	}
	// declare the lookup table for msb
}

// csEqThetaBaseA declares the constraints ensuring the correctness of the theta step
func (theta *theta) csEqThetaBaseA(comp *wizard.CompiledIOP, round int) {
	// cc is the bitshifted version of c. Unlike what is specified by in the
	// spec of keccak, the shifting here is not cyclic. Thus, cc uses 65 bits.
	var c, cc [5][8]*symbolic.Expression
	for x := 0; x < 5; x++ {
		for y := 0; y < 8; y++ {
			// c[x][y] = A[x][0][y] + A[x][1][y] + A[x][2][y] + A[x][3][y] + A[x][4][y]
			c[x][y] = symbolic.Add(theta.stateNext[x][0][y],
				theta.stateNext[x][1][y],
				theta.stateNext[x][2][y],
				theta.stateNext[x][3][y],
				theta.stateNext[x][4][y])
			cc[x][y] = symbolic.Mul(c[x][y], keccakf.BaseA)
		}
	}

	// Since cc is not actually a cyclic rotation, the result for eqTheta still
	// requires adding the MSbit to the LSbit to derive the actual result.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				eqTheta := symbolic.Sub(theta.stateNext[x][y][z],
					symbolic.Add(
						theta.stateCurr[x][y][z],
						c[(x-1+5)%5][z],
						cc[(x+1)%5][z]))
				qName := ifaces.QueryIDf("EQ_THETA_%v_%v_%v", x, y, z)
				comp.InsertGlobal(round, qName, eqTheta)
			}
		}
	}
}

// Proves the link between (aThetaBaseA, aThetaBaseAMsb) with the sliced
// decomposition of aThetaBaseA.
func (t *theta) csAThetaDecomposition(comp *wizard.CompiledIOP, round int) {
	// shf64 = BaseA^U64, it is used to left-shift the MSB to lay on the 64 bits
	// and cancel the MSB of aThetaBaseA
	var shf64 big.Int
	shf64.Exp(big.NewInt(keccakf.BaseA), big.NewInt(64), nil)

}

// assignTheta assigns the values to the columns of theta step
func (theta *theta) assignTheta(run *wizard.ProverRuntime, stateCurr state) {
}
