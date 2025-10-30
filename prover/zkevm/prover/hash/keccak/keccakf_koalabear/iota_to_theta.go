package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

const (
	BaseChi4   = 14641 // 11^4
	BaseTheta  = 4
	BaseTheta4 = 256 // 4^4
)

var (
	BaseChi4Fr   = field.NewElement(BaseChi4)
	BaseThetaFr  = field.NewElement(BaseTheta)
	BaseTheta4Fr = field.NewElement(BaseTheta4)
)

// convertAndClean module, responsible for converting the state from base dirty BaseChi to BaseTheta.
type convertAndClean struct {
	// state before applying the base conversion step, in base dirty BaseChi.
	stateCurr state
	// state after applying the base conversion step, in base clean BaseTheta.
	StateNext state
	// lookup tables to attest the correctness of base conversion. The first column is in BaseChi and the second in BaseTheta.
	lookupTable [2]ifaces.Column
	// state in the middle of base conversion each lane is divided into two limbs of 4 bits each. This step is to reduce the size of the lookup table.
	stateInternalChi, stateInternalTheta stateIn4Bits
}

// newBaseConversion creates a new base conversion module, declares the columns and constraints and returns its pointer
func NewConvertAndClean(comp *wizard.CompiledIOP, stateCurr [5][5]lane) *convertAndClean {

	var (
		bc = &convertAndClean{
			stateCurr: stateCurr,
		}
		size = stateCurr[0][0][0].Size()
	)
	// declare the lookup table columns
	// bc.lookupTable[0] = comp.InsertCommit(0, ifaces.ColID("BC_LOOKUP_DIRTY_BASECHI"), size)
	// bc.lookupTable[1] = comp.InsertCommit(0, ifaces.ColID("BC_LOOKUP_CLEAN_BASETHETA"), size)

	// declare the columns for the new and internal state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 16; z++ {

				if z < 8 {
					bc.StateNext[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_NEXT_%v_%v_%v", x, y, z), size)
				}

				bc.stateInternalChi[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_INTERNAL_CHI_%v_%v_%v", x, y, z), size)

				bc.stateInternalTheta[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_INTERNAL_THETA_%v_%v_%v", x, y, z), size)
			}
		}
	}

	return bc
}

// assignBaseConversion assigns the values to the columns of base conversion step.
func (bc *convertAndClean) Run(run *wizard.ProverRuntime) convertAndClean {
	// decompose each bytes of the lane into 4 bits (base 12)
	var (
		size = bc.stateCurr[0][0][0].Size()
	)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 8; z++ {
				col := bc.stateCurr[x][y][z].GetColAssignment(run).IntoRegVecSaveAlloc()

				for i := range col {
					if col[i].Uint64() >= BaseChi4*BaseChi4 {
						panic("base conversion: input is out of range")
					}
				}
				// q, r := vector.DivMod(col, BaseChi4Fr) // q = high limb, r = low limb
				v := common.Decompose(col, BaseChi4, 2)
				q := v[1] // high limb
				r := v[0] // low limb

				// set the low limb (4 digits)
				run.AssignColumn(bc.stateInternalChi[x][y][2*z].GetColID(), smartvectors.NewRegular(r))
				// set the high limb (4 digits)
				run.AssignColumn(bc.stateInternalChi[x][y][2*z+1].GetColID(), smartvectors.NewRegular(q))
				// decompose in base BaseChi and convert to bits
				lowLimb := common.BaseRecompose(common.DecomposeAndClean(r, BaseChi, 4), &BaseThetaFr)
				highLimb := common.BaseRecompose(common.DecomposeAndClean(q, BaseChi, 4), &BaseThetaFr)
				// set stateInternalTheta
				run.AssignColumn(bc.stateInternalTheta[x][y][2*z].GetColID(), smartvectors.NewRegular(lowLimb))
				run.AssignColumn(bc.stateInternalTheta[x][y][2*z+1].GetColID(), smartvectors.NewRegular(highLimb))
				// set StateNext (8 digits)
				var recomposed = make([]field.Element, size)
				vector.ScalarMul(recomposed, highLimb, BaseTheta4Fr)
				vector.Add(recomposed, recomposed, lowLimb)

				run.AssignColumn(bc.StateNext[x][y][z].GetColID(), smartvectors.NewRegular(recomposed))
			}
		}
	}
	return convertAndClean{stateCurr: bc.stateCurr}
}
