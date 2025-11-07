package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

type (
	state        = [5][5][8]ifaces.Column
	stateIn4Bits = [5][5][16]ifaces.Column
	lane         = [8]ifaces.Column
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
func newConvertAndClean(comp *wizard.CompiledIOP, stateCurr [5][5]lane) *convertAndClean {

	var (
		bc = &convertAndClean{
			stateCurr: stateCurr,
		}
		size = stateCurr[0][0][0].Size()
	)
	// declare the lookup table columns
	dirtyBaseChi, cleanBaseTheta := creatLookupTablesChiToTheta()
	bc.lookupTable[0] = comp.InsertPrecomputed(ifaces.ColID("BC_LOOKUP_DIRTY_BASECHI"), dirtyBaseChi)
	bc.lookupTable[1] = comp.InsertPrecomputed(ifaces.ColID("BC_LOOKUP_CLEAN_BASETHETA"), cleanBaseTheta)

	// declare the columns for the new and internal state
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			for z := 0; z < 16; z++ {

				if z < 8 {
					bc.StateNext[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_NEXT_%v_%v_%v", x, y, z), size)
				}

				bc.stateInternalChi[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_INTERNAL_CHI_%v_%v_%v", x, y, z), size)

				bc.stateInternalTheta[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_INTERNAL_THETA_%v_%v_%v", x, y, z), size)

				// attest the relation between stateInternalChi and stateInternalTheta using the lookup table
				comp.InsertInclusion(0,
					ifaces.QueryIDf("BC_LOOKUP_INCLUSION_%v_%v_%v", x, y, z),
					bc.lookupTable[:], []ifaces.Column{bc.stateInternalChi[x][y][z], bc.stateInternalTheta[x][y][z]})

			}

			for z := 0; z < 8; z++ {
				// asset that stateCurr is decomposed correctly into two slices of stateInternalChi
				exprChi := wizardutils.LinCombExpr(common.BaseChi4, []ifaces.Column{bc.stateInternalChi[x][y][2*z], bc.stateInternalChi[x][y][2*z+1]})
				comp.InsertGlobal(0, ifaces.QueryIDf("BC_LINCOMB_CHI_%v_%v_%v", x, y, z),
					sym.Sub(exprChi, bc.stateCurr[x][y][z]),
				)

				// asset that stateNext is recomposed correctly from two slices of stateInternalTheta
				exprTheta := wizardutils.LinCombExpr(common.BaseTheta4, []ifaces.Column{bc.stateInternalTheta[x][y][2*z], bc.stateInternalTheta[x][y][2*z+1]})
				comp.InsertGlobal(0, ifaces.QueryIDf("BC_LINCOMB_THETA_%v_%v_%v", x, y, z),
					sym.Sub(exprTheta, bc.StateNext[x][y][z]),
				)
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
					if col[i].Uint64() >= common.BaseChi4*common.BaseChi4 {
						panic("base conversion: input is out of range")
					}
				}
				// q, r := vector.DivMod(col, BaseChi4Fr) // q = high limb, r = low limb
				v := common.DecomposeCol(col, common.BaseChi4, 2)
				q := v[1] // high limb
				r := v[0] // low limb

				// set the low limb (4 digits)
				run.AssignColumn(bc.stateInternalChi[x][y][2*z].GetColID(), smartvectors.NewRegular(r))
				// set the high limb (4 digits)
				run.AssignColumn(bc.stateInternalChi[x][y][2*z+1].GetColID(), smartvectors.NewRegular(q))
				// decompose in base BaseChi and convert to bits
				lowLimb := common.RecomposeCols(common.DecomposeAndCleanCol(r, common.BaseChi, 4), &common.BaseThetaFr)
				highLimb := common.RecomposeCols(common.DecomposeAndCleanCol(q, common.BaseChi, 4), &common.BaseThetaFr)
				// set stateInternalTheta
				run.AssignColumn(bc.stateInternalTheta[x][y][2*z].GetColID(), smartvectors.NewRegular(lowLimb))
				run.AssignColumn(bc.stateInternalTheta[x][y][2*z+1].GetColID(), smartvectors.NewRegular(highLimb))
				// set StateNext (8 digits)
				var recomposed = make([]field.Element, size)
				vector.ScalarMul(recomposed, highLimb, common.BaseTheta4Fr)
				vector.Add(recomposed, recomposed, lowLimb)

				run.AssignColumn(bc.StateNext[x][y][z].GetColID(), smartvectors.NewRegular(recomposed))
			}
		}
	}
	return convertAndClean{stateCurr: bc.stateCurr}
}

func creatLookupTablesChiToTheta() (dirtyChi, cleanTheta smartvectors.SmartVector) {
	var (
		lookupDirtyBaseChi   []field.Element
		lookupCleanBaseTheta []field.Element
		cleanValueTheta      field.Element
		targetSize           = utils.NextPowerOfTwo(common.BaseChi4)
	)

	// for each value in base dirty BaseChi (0 to 14640), compute its equivalent in base clean BaseTheta
	for i := 0; i < common.BaseChi4; i++ {
		// decompose in base BaseChi (4 digits) and clean it
		v := common.DecomposeAndCleanFr(field.NewElement(uint64(i)), common.BaseChi, 4)
		// recompose in base BaseTheta
		cleanValueTheta = common.RecomposeRow(v, &common.BaseThetaFr)

		lookupDirtyBaseChi = append(lookupDirtyBaseChi, field.NewElement(uint64(i)))
		lookupCleanBaseTheta = append(lookupCleanBaseTheta, cleanValueTheta)
	}
	dirtyChi = smartvectors.RightPadded(lookupDirtyBaseChi, field.NewElement(common.BaseChi4-1), targetSize)
	cleanTheta = smartvectors.RightPadded(lookupCleanBaseTheta, cleanValueTheta, targetSize)
	return dirtyChi, cleanTheta
}
