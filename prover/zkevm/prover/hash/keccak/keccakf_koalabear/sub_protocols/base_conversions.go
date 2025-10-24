package protocols

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	base4 = 20736 // 12^4
)

var (
	base4Fr = field.NewElement(20736) // 12^4
)

// each lane is 64 bits, represented as 8 bytes.
type lane = [8]ifaces.Column

// keccakf state is a 5x5 matrix of lanes.
type state = [5][5]lane

// state after each base conversion, each lane is decomposed into 16 slices of 4 bits each.
type stateIn4Bits = [5][5][16]ifaces.Column

// state in bits
type stateInBits = [5][5][64]ifaces.Column

// BaseConversion module, responsible for converting the state from base dirty 12 to base 2.
type BaseConversion struct {
	// state before applying the base conversion step, in base dirty 12
	stateCurr state
	// state after applying the base conversion step, in base 2.
	StateNext stateInBits
	// lookup tables to attest the correctness of base conversion,
	// the first column is the 4 digits slice in base 12, the column[1:4] is its decomposition into 4 bits.
	// an even digit is replaced with 0 and an odd digit with 1 in the decomposition columns.
	lookupTable [5]ifaces.Column
	// state in the middle of base conversion, each byte is decomposed into 2 limbs in base dirty 12,
	stateInternal stateIn4Bits
}

// newBaseConversion creates a new base conversion module, declares the columns and constraints and returns its pointer
func NewBaseConversion(comp *wizard.CompiledIOP, numKeccakf int, stateCurr [5][5]lane) *BaseConversion {

	var (
		bc   = &BaseConversion{}
		size = utils.NextPowerOfTwo(numKeccakf * 24)
	)

	// declare the columns for the new and internal state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 64; z++ {
				bc.StateNext[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_NEXT_%v_%v_%v", x, y, z), size)
				if z < 16 {
					bc.stateInternal[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_INTERNAL_%v_%v_%v", x, y, z), size)
				}
			}
		}
	}

	return bc
}

// assignBaseConversion assigns the values to the columns of base conversion step.
func (bc *BaseConversion) Run(run *wizard.ProverRuntime) BaseConversion {
	// decompose each bytes of the lane into 4 bits (base 12)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				col := bc.stateCurr[x][y][z].GetColAssignment(run).IntoRegVecSaveAlloc()
				q, r := vector.DivMod(col, base4Fr)
				// set the low limb (4 digits)
				run.AssignColumn(bc.stateInternal[x][y][2*z].GetColID(), smartvectors.NewRegular(r))
				// set the high limb (4 digits)
				run.AssignColumn(bc.stateInternal[x][y][2*z+1].GetColID(), smartvectors.NewRegular(q))
				// decompose in base 12 and convert to bits
				lowLimb := DecomposeToBits(r, 12, 4)
				highLimb := DecomposeToBits(q, 12, 4)
				// set the bits in the new state
				for b := 0; b < 4; b++ {
					run.AssignColumn(bc.StateNext[x][y][8*z+b].GetColID(), smartvectors.NewRegular(lowLimb[b]))
					run.AssignColumn(bc.StateNext[x][y][8*z+4+b].GetColID(), smartvectors.NewRegular(highLimb[b]))
				}

			}
		}
	}
	return BaseConversion{stateCurr: bc.stateCurr}
}

// it decompose the given field element n into the given base.
func DecomposeUint32(n uint32, base, nb int) []field.Element {
	// It will essentially be used for chunk to slice decomposition
	var (
		res    = make([]field.Element, 0, nb)
		curr   = n
		base32 = uint32(base)
	)
	for curr > 0 {
		limb := field.NewElement(uint64(curr % base32))
		res = append(res, limb)
		curr /= base32
	}

	if len(res) > nb {
		utils.Panic("expected %v limbs, but got %v", nb, len(res))
	}

	// Complete with zeroes
	for len(res) < nb {
		res = append(res, field.Zero())
	}

	return res
}

func Decompose(n []field.Element, base, nb int) [][]field.Element {

	res := make([][]field.Element, len(n))
	for i := range n {
		res[i] = DecomposeUint32(n[i][0], base, nb)
	}
	return res
}

func convertToBits(n []field.Element) []field.Element {
	res := make([]field.Element, len(n))
	for i := range n {
		if n[i][0]%2 == 0 {
			res[i].SetZero()
		} else {
			res[i].SetOne()
		}
	}
	return res
}

func DecomposeToBits(n []field.Element, base, nb int) [][]field.Element {

	limbs := Decompose(n, base, nb)             // in base 'base', nb limbs
	bits := make([][]field.Element, len(limbs)) // bits representation of each limb
	for i := range limbs {
		bits[i] = convertToBits(limbs[i])
	}
	return bits
}
