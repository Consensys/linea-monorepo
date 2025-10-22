package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

const thetaBase = 8

// theta module, responsible for updating the state in the theta step of keccakf
type theta struct {
	// state before applying the theta step, in base clean 12
	stateCurr state
	// state after applying the theta step, in base dirty 12
	stateNext state
	// Intermediate columns
	// cc is the bitshifted version of c used in theta step
	cc [5][8]ifaces.Column
	// bitConvertedCC is the bit converted version of cc used in theta step
	bitConvertedCC [5][8]ifaces.Column
}

// newTheta creates a new theta module, declares the columns and constraints and returns its pointer
func newTheta(comp *wizard.CompiledIOP,
	numKeccakf int,
	stateCurr state,
	l lookupTables) *theta {
	res := &theta{}
	res.stateCurr = stateCurr

	// declare the columns
	res.declareColumnsTheta(comp, numKeccakf)

	// declare the constraints
	res.csEqThetaBaseA(comp)
	res.csLookupCCToBitConvertedCC(comp, l)
	return res
}

// declareColumnsTheta declares the intermediate columns generated during theta step, including the new state.
func (theta *theta) declareColumnsTheta(comp *wizard.CompiledIOP, numKeccakf int) {
	// size of the columns to declare
	colSize := keccakf.NumRows(numKeccakf)
	// declare the new state
	theta.stateNext = state{}
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				theta.stateNext[x][y][z] = comp.InsertCommit(
					0,
					keccakf.DeriveName("A_THETA", x, y, z),
					colSize,
				)
			}
		}
	}
	// declare the bit converted cc columns
	for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			theta.bitConvertedCC[x][z] = comp.InsertCommit(
				0,
				keccakf.DeriveName("CC_BIT_CONVERTED", x, z),
				colSize,
			)
		}
	}
}

// csEqThetaBaseA declares the constraints ensuring the correctness of the theta step
func (theta *theta) csEqThetaBaseA(comp *wizard.CompiledIOP) {
	// cc_computed is the bitshifted version of c. Unlike what is specified by in the
	// spec of keccak, the shifting here is not cyclic. Thus, cc_computed uses 65 bits.
	var c, cc_computed [5][8]*symbolic.Expression
	for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			// c[x][y] = A[x][0][y] + A[x][1][y] + A[x][2][y] + A[x][3][y] + A[x][4][y]
			c[x][z] = symbolic.Add(theta.stateCurr[x][0][z],
				theta.stateCurr[x][1][z],
				theta.stateCurr[x][2][z],
				theta.stateCurr[x][3][z],
				theta.stateCurr[x][4][z])
			cc_computed[x][z] = symbolic.Mul(c[x][z], thetaBase)
		}
	}
	// Check that the next state of theta is correctly computed
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				eqTheta := symbolic.Sub(theta.stateNext[x][y][z],
					symbolic.Add(
						theta.stateCurr[x][y][z],
						c[(x-1+5)%5][z],
						theta.bitConvertedCC[(x+1)%5][z]))
				qName := ifaces.QueryIDf("EQ_THETA_%v_%v_%v", x, y, z)
				comp.InsertGlobal(0, qName, eqTheta)
			}
		}
	}

	// Check that cc computed from c is the same as cc assigned
	for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			eqCC := symbolic.Sub(theta.cc[x][z], cc_computed[x][z])
			qName := ifaces.QueryIDf("EQ_CC_%v_%v", x, z)
			comp.InsertGlobal(0, qName, eqCC)
		}
	}
}

// Lookup between cc and bitConvertedCC, slice by slice
func (theta *theta) csLookupCCToBitConvertedCC(comp *wizard.CompiledIOP,
	l lookupTables) {
	for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			comp.InsertInclusion(
				0,
				ifaces.QueryIDf("LOOKUP_CC_TO_BIT_CONVERTED_%v_%v", x, z),
				[]ifaces.Column{
					l.ccBase8Theta,
					l.ccBitConvertedTheta,
				},
				[]ifaces.Column{
					theta.cc[x][z],
					theta.bitConvertedCC[x][z],
				},
			)
		}
	}
}

// assignTheta assigns the values to the columns of theta step
func (theta *theta) assignTheta(run *wizard.ProverRuntime, stateCurr state) {
}
