package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

const (
	thetaBase = 4 // lane is decomposed into 8 slices in base 4.
	// base 4 allows three additions without overflow.
	// it also allows acceptable column size for lookup arguments.
	thetaBase8 = 65536 // thetaBase^8
)

// theta module, responsible for updating the state in the theta step of keccakf
type theta struct {
	// state before applying the theta step, in base clean 4.
	stateCurr state
	// state after applying the theta step, in base clean 4.
	stateInternal state
	// state after applying the theta step, in bits.
	stateNext stateInBits
	// Intermediate columns, after each 3 additions
	cMiddle, cFinal, cMiddleCleanBase, cFinalCleanBase [5][8]ifaces.Column
	// masb, lsb of cFinal used for the rotation and computing cc.
	msb, lsb [5][8]ifaces.Column
}

// newTheta creates a new theta module, declares the columns and constraints and returns its pointer
func newTheta(comp *wizard.CompiledIOP,
	numKeccakf int,
	stateCurr state) *theta {
	theta := &theta{}
	theta.stateCurr = stateCurr

	// declare the new state and intermediate columns
	theta.declareColumnsTheta(comp, numKeccakf)

	/*	var c, cc [5][8]*sym.Expression
		for x := 0; x < 5; x++ {
			for z := 0; z < 8; z++ {
				// c[x][z] = A[x][0][z] + A[x][1][z] + A[x][2][z] + A[x][3][z] + A[x][4][z]
				c[x][z] = sym.Add(
					stateCurr[x][0][z],
					stateCurr[x][1][z],
					stateCurr[x][2][z],
					stateCurr[x][3][z],
					stateCurr[x][4][z])
				// roate each theta.cleanBaseC by 1 position to get cc
				cc[x][z] = sym.Add(
					sym.Mul(theta.cFinal[x][z], thetaBase),
					theta.lsb[x][z],
					sym.Mul(theta.msb[x][z], -1*thetaBase8),
				)
			}
		}
		// Check that the next state of theta is correctly computed
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				for z := 0; z < 8; z++ {
					eqTheta := sym.Sub(theta.stateInternal[x][y][z],
						sym.Add(
							stateCurr[x][y][z],
							c[(x-1+5)%5][z],
							cc[(x+1)%5][z]))
					qName := ifaces.QueryIDf("EQ_THETA_%v_%v_%v", x, y, z)
					comp.InsertGlobal(0, qName, eqTheta)
				}
			}
		}

		// check that  cleanBaseC[x][z] is correctly computed from c[x][z] (via lookup)
	*/

	return theta
}

// declareColumnsTheta declares the intermediate columns generated during theta step, including the new state.
func (theta *theta) declareColumnsTheta(comp *wizard.CompiledIOP, numKeccakf int) {
	// size of the columns to declare
	colSize := keccakf.NumRows(numKeccakf)
	// declare the new state
	theta.stateInternal = state{}
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				theta.stateInternal[x][y][z] = comp.InsertCommit(
					0,
					keccakf.DeriveName("STATE_THETA_%v_%v_%v", x, y, z),
					colSize,
				)

				// declare the new state in bits
				for i := 0; i < 8; i++ {
					theta.stateNext[x][y][z*8+i] = comp.InsertCommit(
						0,
						keccakf.DeriveName("STATE_THETA_BIT_%v_%v_%v", x, y, z*8+i),
						colSize,
					)
				}
			}
		}
	}
	// declare cm ,cf, msb, lsb columns
	/*for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			theta.cFinal[x][z] = comp.InsertCommit(
				0,
				keccakf.DeriveName("CC_BIT_CONVERTED_ROTAYED_%v_%v", x, z),
				colSize,
			)
		}
	}*/
}

func (theta *theta) assignTheta(run *wizard.ProverRuntime, stateCurr state) {

	var (
		stateCurrWit [5][5][8][]field.Element
		cm, cf       [5][8][]field.Element
		cc           [5][8][]field.Element
		// cmClean, cfClean, msb,lsb [5][8]*common.VectorBuilder
		col           []field.Element
		stateInternal [5][5][8]*common.VectorBuilder
		stateBinary   [5][5][64]*common.VectorBuilder
		lsb           [5][8][]field.Element
	)
	// get the current state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				stateCurrWit[x][y][z] = stateCurr[x][y][z].GetColAssignment(run).IntoRegVecSaveAlloc()
			}
		}
	}
	size := len(stateCurrWit[0][0][0])
	// compute c and cc, cleanBaseC, msb, lsb
	for x := 0; x < 5; x++ {
		for z := 0; z < 8; z++ {
			cm[x][z] = make([]field.Element, size)
			cf[x][z] = make([]field.Element, size)

			// cmClean[x][z] = common.NewVectorBuilder(theta.cMiddleCleanBase[x][z])
			// cfClean[x][z] = common.NewVectorBuilder(theta.cFinalCleanBase[x][z])

			// msb[x][z] = common.NewVectorBuilder(theta.msb[x][z])
			// lsb[x][z] = common.NewVectorBuilder(theta.lsb[x][z])
			// c[x][z] = A[x][0][z] + A[x][1][z] + A[x][2][z] + A[x][3][z] + A[x][4][z]
			vector.Add(cm[x][z],
				stateCurrWit[x][0][z],
				stateCurrWit[x][1][z],
				stateCurrWit[x][2][z],
			)

			vector.Add(cf[x][z],
				cm[x][z],
				stateCurrWit[x][3][z],
				stateCurrWit[x][4][z],
			)

			// get lsbPrev for computing cc
			// lsbPrev := theta.lsb[x][(z-1)%8].GetColAssignment(run).IntoRegVecSaveAlloc()
			for i := 0; i < len(cf[x][z]); i++ {
				// decompose c[x][z]  and clean it
				// resm := clean(Decompose(cm[x][z][i].Uint64(), thetaBase, 8))
				resf := clean(Decompose(cf[x][z][i].Uint64(), thetaBase, 8))
				lsb[x][z] = append(lsb[x][z], field.NewElement(uint64(resf[0])))

				// recompse to get cleanBaseC
				// cmCleaned := 0
				cfCleaned := 0
				for i := len(resf) - 1; i >= 0; i-- {
					//	cmCleaned = cmCleaned*thetaBase + resm[i]
					cfCleaned = cfCleaned*thetaBase + resf[i]
				}

				// cmClean[x][z].PushInt(cmCleaned)
				// cfClean[x][z].PushInt(cfCleaned)
				// msb[x][z].PushInt(resf[len(resf)-1])
				// lsb[x][z].PushInt(resf[0])
				a := cfCleaned*thetaBase - resf[len(resf)-1]*thetaBase8 + int(lsb[x][(z-1)%8][i].Uint64())
				cc[x][z] = append(cc[x][z], field.NewElement(uint64(a)))
			}
			// assign cleanBaseC, msb, lsb columns
			/*cmClean[x][z].PadAndAssign(run)
			cfClean[x][z].PadAndAssign(run)
			msb[x][z].PadAndAssign(run)
			lsb[x][z].PadAndAssign(run)
			*/

		}
	}

	// assign internal and final state.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {

				stateInternal[x][y][z] = common.NewVectorBuilder(theta.stateInternal[x][y][z])
				// assign the binary state
				for i := 0; i < 8; i++ {
					stateBinary[x][y][z*8+i] = common.NewVectorBuilder(theta.stateNext[x][y][z*8+i])
				}
				// A'[x][y][z] = A[x][y][z] + c[(x-1+5)%5][z] + cc[(x+1)%5][z]
				vector.Add(col, stateCurrWit[x][y][z], cf[(x-1+5)%5][z], cc[(x+1)%5][z])
				// clean the state
				for i := 0; i < len(col); i++ {
					// decompose col and clean it, buggy
					res := clean(Decompose(col[i].Uint64(), thetaBase, 8))
					// recompse to get clean state
					stateCleaned := 0
					for i := len(res) - 1; i >= 0; i-- {
						stateCleaned = stateCleaned*thetaBase + res[i]
					}
					stateInternal[x][y][z].PushInt(stateCleaned)
					for i := 0; i < 8; i++ {
						stateBinary[x][y][z*8+i].PushInt(res[i])
					}
				}
				// assign the internal state
				stateInternal[x][y][z].PadAndAssign(run)
				// assign the binary state
				for i := 0; i < 8; i++ {
					stateBinary[x][y][z*8+i].PadAndAssign(run)
				}
			}
		}
	}

}

// clean converts a slice of uint64 values into a new slice of ints of the
// same length where each element is 1 if the corresponding input value is
// odd and 0 if it is even. The input slice is not modified.
func clean(in []uint64) (out []int) {
	out = make([]int, len(in))
	for i, element := range in {
		if element%2 == 0 {
			out[i] = 0
		} else {
			out[i] = 1
		}
	}
	return out
}
