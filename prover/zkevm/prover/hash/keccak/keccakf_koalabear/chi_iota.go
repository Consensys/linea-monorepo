package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

const BaseChi = 11

var baseFr = field.NewElement(BaseChi)

type block [numLanesInBlock][numSlices]ifaces.Column

type chi struct {
	Inputs *chiInputs
	// internal state  recomposing each [numSlices] bits into a base clean BaseChi.
	// it is a placeholder for the linear combination expressions, to facilitate the assignment.
	stateInternal [5][5][numSlices]*sym.Expression
	// state after applying the chi step.
	// It is in the expression form since it will be combined with Iota step
	// to get the standard state later. This avoid declaring extra columns.
	StateNext [5][5][numSlices]ifaces.Column
	// prover actions for linear combinations
	// paLinearCombinations [5][5][numSlices]*protocols.LinearCombination
	// the round constant
	rc [numSlices]*dedicated.RepeatedPattern
}

type chiInputs struct {
	// state before applying the chi step
	stateCurr stateInBits
	// it contains the first block of the message at position 0 mod 24,
	// and any other block at position 0 mod 23.
	blocks block
	// flag for blocks other than the first one
	isBlockOther ifaces.Column
	// max number of keccakf permutations that module can support
	numKeccakf int
}

// newChi define the chi step of the keccak-f permutation.
// It creates the necessary columns and constraints to enforce the chi step.
// the state is updated as follows:
// A[x][y] = A[x][y] + ( (not A[x+1][y]) and A[x+2][y] )  and
// A[0,0] = A[0,0] + RC and then A[x][y] = A[x][y] + block[x+5*y] for all message blocks except the first one.
// the blocks are added to the end of the current Iota step to avoid creating extra columns and facilitating the theta step.
func newChi(comp *wizard.CompiledIOP, in chiInputs) *chi {

	chi := &chi{
		stateInternal: [5][5][numSlices]*sym.Expression{},
		Inputs:        &in,
		StateNext:     [5][5][numSlices]ifaces.Column{},
	}

	var (
		stateNext [5][5][numSlices]*sym.Expression
		size      = numRows(in.numKeccakf)
	)

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < numSlices; z++ {
				// set the internal state column to the result of the linear combination
				chi.stateInternal[x][y][z] = wizardutils.LinCombExpr(BaseChi, in.stateCurr[x][y][z*numSlices:z*numSlices+numSlices])
				// define the state next columns
				chi.StateNext[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("CHI_STATE_NEXT_%v_%v_%v", x, y, z), size)
			}
		}
	}

	// define the round constant columns
	rcCols := ValRCBase2Pattern()
	for i := 0; i < numSlices; i++ {
		chi.rc[i] = dedicated.NewRepeatedPattern(
			comp,
			0,
			rcCols[i],
			verifiercol.NewConstantCol(field.One(), size, "keccak-rc-pattern"),
		)
	}

	// apply complex binary. i.e., A[x][y] = A[x][y] + ( (not A[x+1][y]) and A[x+2][y] ) + block[x+5*Y]  and A[0,0] = A[0,0] + RC
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < numSlices; z++ {
				stateNext[x][y][z] = sym.Add(
					sym.Mul(2, chi.stateInternal[x][y][z]),
					chi.stateInternal[(x+1)%5][y][z],
					sym.Mul(3, chi.stateInternal[(x+2)%5][y][z]),
				)
				if x+5*y < numLanesInBlock {
					// add the message block for all blocks except the first one
					stateNext[x][y][z] = sym.Add(
						stateNext[x][y][z],
						sym.Mul(2, in.blocks[x+5*y][z], in.isBlockOther),
					)
				}

				// add the round constant for position (0,0)
				if x == 0 && y == 0 {
					stateNext[x][y][z] = sym.Add(
						stateNext[x][y][z],
						sym.Mul(2, chi.rc[z].Natural),
					)
				}

				comp.InsertGlobal(0, ifaces.QueryIDf("CHI_STATE_NEXT_%v_%v_%v", x, y, z),
					sym.Sub(stateNext[x][y][z], chi.StateNext[x][y][z]),
				)
			}

		}
	}
	return chi
}

// assignChi assigns the values to the columns of chi step.
func (chi *chi) assignChi(run *wizard.ProverRuntime, stateCurr stateInBits) {
	var (
		u, v          []field.Element
		stateInternal [5][5][numSlices][]field.Element
		size          = stateCurr[0][0][0].Size()
		rcCols        [numSlices][]field.Element
		isBlockOther  = chi.Inputs.isBlockOther.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	// assign the linear combinations for each lane in the state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < numSlices; z++ {
				// get the assignment of the internal state
				stateInternal[x][y][z] = column.EvalExprColumn(run, chi.stateInternal[x][y][z].Board()).IntoRegVecSaveAlloc()

			}
		}
	}

	// assign the state after chi step
	two := field.NewElement(2)
	for i := 0; i < numSlices; i++ {
		chi.rc[i].Assign(run)
		rcCols[i] = chi.rc[i].Natural.GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < numSlices; z++ {
				// A[x][y] = A[x][y] + ( (not A[x+1][y]) and A[x+2][y])
				u = make([]field.Element, size)
				v = make([]field.Element, size)
				vector.ScalarMul(u, stateInternal[x][y][z], two)
				vector.ScalarMul(v, stateInternal[(x+2)%5][y][z], field.NewElement(3))
				vector.Add(u, u, v, stateInternal[(x+1)%5][y][z])

				// add the round constant for position (0,0)
				if x == 0 && y == 0 {
					var tt = make([]field.Element, size)
					vector.ScalarMul(tt, rcCols[z], two)
					vector.Add(u, u, tt)
				}
				// add the message block for all blocks
				if x+5*y < numLanesInBlock {
					blocks := chi.Inputs.blocks[x+5*y][z].GetColAssignment(run).IntoRegVecSaveAlloc()
					var tt = make([]field.Element, size)
					vector.ScalarMul(tt, blocks, two)
					vector.MulElementWise(tt, tt, isBlockOther)
					vector.Add(u, u, tt)
				}
				// assign the result to the state next column
				run.AssignColumn(chi.StateNext[x][y][z].GetColID(), smartvectors.NewRegular(u))

			}
		}
	}

}

// ValRCBase2Pattern returns the round constants in base BaseChi (11) and sliced form.
func ValRCBase2Pattern() [numSlices][]field.Element {

	var (
		res = [numSlices][]field.Element{}
	)

	for j := range keccak.RC {
		var out [numSlices]uint64
		for i := 0; i < numSlices; i++ {
			out[i] = (keccak.RC[j] >> (numSlices * i)) & 0xFF // take each byte, LSB first
			a := keccakf.U64ToBaseX(out[i], &baseFr)
			res[i] = append(res[i], a)
		}

	}
	return res
}
