package keccakfkoalabear

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
	protocols "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/sub_protocols"
)

type block [17][8]ifaces.Column

type chi struct {
	// state before applying the chi step
	stateCurr stateInBits
	// internal state  recomposing each 8 bits into a base clean 11.
	stateInternal state
	// state after applying the chi step.
	// It is in the expression form since it will be combined with Iota step
	// to get the standard state later. This avoid declaring extra columns.
	StateNext [5][5][8]*symbolic.Expression
	// prover actions for linear combinations
	paLinearCombinations [5][5][8]*protocols.LinearCombination
	// state witness after applying the chi step, since the state-witness is needed for the Iota step.
	stateNextWitness [5][5][8][]field.Element
	// the round constant
	RC [8]*dedicated.RepeatedPattern
	// it contains the first block of the message at position 0 mod 24,
	// and any other block at position 0 mod 23.
	BlocksOther block
	// flag for blocks other than the first one
	IsBlockOther ifaces.Column
}

func newChi(comp *wizard.CompiledIOP, numKeccakf int, stateCurr stateInBits, blocks block, isBlockOther ifaces.Column) *chi {

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

	rcCols := ValRCBase2Pattern()

	// define the round constant columns
	for i := 0; i < 8; i++ {
		chi.RC[i] = dedicated.NewRepeatedPattern(
			comp,
			0,
			rcCols[i],
			verifiercol.NewConstantCol(field.One(), numRows(numKeccakf), "keccak-rc-pattern"),
		)
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
				if x+5*y < 17 {
					// add the message block for all blocks except the first one
					chi.StateNext[x][y][z] = sym.Add(
						chi.StateNext[x][y][z],
						sym.Mul(2, blocks[x+5*y][z], isBlockOther),
					)
				}

				if x == 0 && y == 0 {
					chi.StateNext[x][y][0] = sym.Add(
						chi.StateNext[x][y][0],
						// sym.Mul(2, chi.RC.Natural),
					)
				}
			}

		}
	}
	return chi
}

// assignChi assigns the values to the columns of chi step.
func (chi *chi) assignChi(run *wizard.ProverRuntime, stateCurr stateInBits) {
	var (
		u, v          []field.Element
		stateInternal [5][5][8][]field.Element
		size          = stateCurr[0][0][0].Size()
		rcCols        [8][]field.Element
	)
	// assign the linear combinations for each lane in the state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				chi.paLinearCombinations[x][y][z].Run(run)
				stateInternal[x][y][z] = chi.stateInternal[x][y][z].GetColAssignment(run).IntoRegVecSaveAlloc()
			}
		}
	}

	// assign the state after chi step
	// eleven := field.NewElement(11)
	two := field.NewElement(2)
	for i := 0; i < 8; i++ {
		chi.RC[i].Assign(run)
		rcCols[i] = chi.RC[i].Natural.GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				// A[x][y] = A[x][y] + ( (not A[x+1][y]) and A[x+2][y])
				u = make([]field.Element, size)
				v = make([]field.Element, size)
				vector.ScalarMul(u, stateInternal[x][y][z], field.NewElement(2))
				vector.ScalarMul(v, stateInternal[(x+2)%5][y][z], field.NewElement(3))
				vector.Add(u, u, v, stateInternal[(x+1)%5][y][z])
				// var k field.Element
				// If it is the first lane, then add the round constant
				/*if x == 0 && y == 0 && z == 0 {
					for i := 0; i < size; i++ {
						a := keccakf.U64ToBaseX(keccak.RC[i%keccak.NumRound], &eleven)
						u[i].Add(&u[i], k.Mul(&two, &a))
					}
				}*/

				if x == 0 && y == 0 {
					var tt = make([]field.Element, size)
					vector.ScalarMul(tt, rcCols[z], two)
					vector.Add(u, u, tt)
				}
				chi.stateNextWitness[x][y][z] = u

			}
		}
	}

}

// to be removed later
func Decompose(r uint64, base int, nb int, pos ...int) (res []uint64) {
	// It will essentially be used for chunk to slice decomposition
	res = make([]uint64, 0, nb)
	base64 := uint64(base)
	curr := r
	for curr > 0 {
		limb := curr % base64
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > nb {
		utils.Panic("expected %v limbs, but got %v, overflow at pos %v", nb, len(res), pos)
	}

	// Complete with zeroes
	for len(res) < nb {
		res = append(res, 0)
	}
	return res
}

func cleanBase(in []uint64) (out []uint64) {
	out = make([]uint64, len(in))
	for i := 0; i < len(in); i++ {
		// take the second bit
		out[i] = in[i] >> 1 & 1
	}
	return out
}

func ValRCBase2Pattern() [8][]field.Element {

	var (
		res    = [8][]field.Element{}
		baseBF = field.NewElement(11)
	)

	for j := range keccak.RC {
		var out [8]uint64
		for i := 0; i < 8; i++ {
			out[i] = (keccak.RC[j] >> (8 * i)) & 0xFF // take each byte, LSB first
			a := keccakf.U64ToBaseX(out[i], &baseBF)
			res[i] = append(res[i], a)
		}

	}
	return res
}
