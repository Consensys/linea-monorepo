package keccakfkoalabear

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	protocols "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/sub_protocols"
)

type chi struct {
	// state before applying the chi step
	stateCurr stateInBits
	// internal state  recomposing each 8 bits into a base clean 11.
	stateInternal state
	// state after applying the chi step. It is in the expression form since it will be combined with Iota step to get the standard state later.
	// this avoid declaring extra columns.
	StateNext [5][5][8]*symbolic.Expression
	// prover actions for linear combinations
	paLinearCombinations [5][5][8]*protocols.LinearCombination
	// state witness after applying the chi step, since the state-witness is needed for the next Iota step.
	stateNextWitness [5][5][8][]field.Element
	// the round constant
	RC ifaces.Column
}

func newChi(comp *wizard.CompiledIOP, numKeccakf int, stateCurr stateInBits) *chi {

	chi := &chi{
		stateCurr:     stateCurr,
		stateInternal: state{},
	}

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				chi.paLinearCombinations[x][y][z] = protocols.NewLinearCombination(comp,
					fmt.Sprintf("CHI_STATE_NEXT_%v_%v_%v", x, y, z),
					stateCurr[x][y][z*8:z*8+8],
					11)
				// set the internal state column to the result of the linear combination
				chi.stateInternal[x][y][z] = chi.paLinearCombinations[x][y][z].CombinationRes
			}
		}
	}

	// apply complex binary. i.e., A[x][y] = A[x][y] + ( (not A[x+1][y]) and A[x+2][y] )  and A[0,0] = A[0,0] + RC
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				chi.StateNext[x][y][z] = sym.Add(
					sym.Mul(2, chi.stateInternal[x][y][z]),
					chi.stateInternal[(x+1)%5][y][z],
					sym.Mul(3, chi.stateInternal[(x+2)%5][y][z]),
				)
			}
			/* if x == 0 && y == 0 {
					chi.StateNext[x][y][0] = sym.Add(
						chi.StateNext[x][y][0],
						chi.RC)
				}
			}*/
		}
	}
	return chi
}

// assignChi assigns the values to the columns of chi step.
func (chi *chi) assignChi(run *wizard.ProverRuntime, stateCurr stateInBits) {
	// assign the linear combinations for each lane in the state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				chi.paLinearCombinations[x][y][z].Run(run)
			}
		}
	}
}
