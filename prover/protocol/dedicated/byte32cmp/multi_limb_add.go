package byte32cmp

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
)

// MultiLimbAddIn is the input structure for the MultiLimbAdd operation.
type MultiLimbAddIn struct {
	// Name is a unique prefix for the operation.
	Name string
	// ALimbs is the LimbColumns object representing the "a" operand.
	//
	// Note: The number of limbs must be at least as many as in the "a" operand.
	ALimbs LimbColumns
	// BLimbs is the LimbColumns object representing the "b" operand.
	BLimbs LimbColumns
	// Result is the LimbColumns object that will store the result of the addition.
	// It can be omited, then the result will be computed and returned as a brand
	// new column.
	Result LimbColumns
	// Mask is an expression to use to mask the rows to be processed. Binary check
	// is performed inside for this value.
	Mask *sym.Expression
	// See [wizard.CompiledIOP.InsertGlobal] for more details.
	NoBoundCancel bool
}

// MultiLimbAdd is a module that constraints the addition of a column to a
// LimbColumns. It takes a LimbColumns object representing the "a" operand, a column
// "b" to be added, and produces a new LimbColumns object representing the result of
// the addition. It also computes the carry bits for each limb, which are stored in
// a separate LimbColumns object. The addition is performed in a big-endian manner,
// meaning the most significant limb is at the end of the list.
//
//   - a     := (a0, a1, a2, a3) - the limbs of the first operand
//   - b     := (b0, b1, b2, b3) - the limbs of the second operand
//   - res   := (res0, res1, res2, res3) - the result of the addition
//   - carry := (carry0, carry1, carry2, carry3) - the carry bits of the addition
//
// base = 2^limbBitSize
//
//	res3 + (carry2 * base) = a3 + b3
//	res2 + (carry1 * base) = a2 + b2 + carry2
//	res1 + (carry0 * base) = a1 + b1 + carry1
//	res0                   = a0 + b0 + carry0
//
// res_i, res_j, b in [0, base)
type MultiLimbAdd struct {
	// name stores a unique prefix for the operation.
	name string
	// aLimbs stores the list of the columns, each one storing a part of the "a" operand.
	aLimbs LimbColumns
	// bLimbs stores the list of the columns, each one storing a part of the "b" operand.
	bLimbs LimbColumns
	// result stores the list of the columns that represent the result of the addition.
	result LimbColumns
	// withResult indicates whether the result should be computed and stored in a brand
	// new column. If true, the result column was provided as the part of the input.
	withResult bool
	// carry stores the carry bits of addition for each limb.
	carry LimbColumns
	// mask is an expression to use to mask the rows to be processed. It is a binary
	// expression, i.e. 0 or 1.
	mask *sym.Expression
	// See [wizard.CompiledIOP.InsertGlobal] for more details.
	noBoundCancel bool
	// isAddition defines whether we need to apply an addition or subtraction.
	isAddition bool
}

// NewMultiLimbAdd creates a new MultiLimbAdd module. It return the LimbColumns
// representing the result of the addition and a wizard.ProverAction that should be run.
//
// If the result columns are provided in input, then the same columns are returned
// and no new are created.
func NewMultiLimbAdd(comp *wizard.CompiledIOP, inp *MultiLimbAddIn, isAdd bool) (LimbColumns, wizard.ProverAction) {
	if !inp.ALimbs.IsBigEndian {
		utils.Panic("MultiLimbAdd only supports big-endian limbs")
	}

	if len(inp.ALimbs.Limbs) < len(inp.BLimbs.Limbs) {
		utils.Panic("MultiLimbAdd: aLimbs must have at least as many limbs as bLimbs")
	}

	if len(inp.ALimbs.Limbs) == 0 {
		utils.Panic("MultiLimbAdd: aLimbs must have at least one limb")
	}

	numRows := ifaces.AssertSameLength(append(inp.ALimbs.Limbs, inp.BLimbs.Limbs...)...)

	result := inp.Result
	if result.Limbs == nil {
		result.Limbs = make([]ifaces.Column, len(inp.ALimbs.Limbs))
		result.LimbBitSize = inp.ALimbs.LimbBitSize
		result.IsBigEndian = inp.ALimbs.IsBigEndian

		for i := range result.Limbs {
			result.Limbs[i] = comp.InsertCommit(0,
				ifaces.ColIDf("%v_ADD_COL_TO_LIMBS_RESULT_%v", inp.Name, i), numRows, true)
		}
	}

	res := &MultiLimbAdd{
		name:       inp.Name,
		aLimbs:     inp.ALimbs,
		bLimbs:     inp.BLimbs,
		mask:       inp.Mask,
		result:     result,
		withResult: inp.Result.Limbs != nil,
		carry: LimbColumns{
			Limbs: make([]ifaces.Column, len(inp.ALimbs.Limbs)-1),
		},
		noBoundCancel: inp.NoBoundCancel,
		isAddition:    isAdd,
	}

	for i := range res.carry.Limbs {
		res.carry.Limbs[i] = comp.InsertCommit(0,
			ifaces.ColIDf("%v_ADD_COL_TO_LIMBS_CARRY_%v", inp.Name, i), numRows, true)
	}

	if isAdd {
		res.csAddition(comp, res.aLimbs, res.bLimbs, res.result)
	} else {
		res.csAddition(comp, res.result, res.bLimbs, res.aLimbs)
	}

	res.csRangeChecks(comp)

	return result, res
}

func (m *MultiLimbAdd) csRangeChecks(comp *wizard.CompiledIOP) {
	for i := range m.carry.Limbs {
		commonconstraints.MustBeBinary(comp, m.carry.Limbs[i])
	}

	limbMax := 1 << m.aLimbs.LimbBitSize

	for i := range m.bLimbs.Limbs {
		comp.InsertRange(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_B_RANGE_%d", m.name, i),
			m.bLimbs.Limbs[i], limbMax,
		)
	}

	for i := range m.aLimbs.Limbs {
		comp.InsertRange(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_A_RANGE_%v", m.name, i),
			m.aLimbs.Limbs[i], limbMax,
		)

		comp.InsertRange(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_RESULT_RANGE_%v", m.name, i),
			m.result.Limbs[i], limbMax,
		)
	}
}

func (m *MultiLimbAdd) csAddition(comp *wizard.CompiledIOP, aLimbs, bLimbs, resultLimbs LimbColumns) {
	limbMax := field.NewElement(uint64(1) << aLimbs.LimbBitSize)
	lastLimbIdx := len(aLimbs.Limbs) - 1

	// Mask binary check
	// mask * (1 - mask)
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_MASK", m.name),
		sym.Mul(m.mask, sym.Sub(1, m.mask)),
	)

	// Constraint for a single limb
	// result[last] = a[last] + b[last]
	if lastLimbIdx == 0 {
		comp.InsertGlobal(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_CONSTRAINT_LSB", m.name),
			sym.Mul(
				m.mask,
				sym.Sub(
					resultLimbs.Limbs[lastLimbIdx],
					sym.Add(aLimbs.Limbs[lastLimbIdx], bLimbs.Limbs[lastLimbIdx]),
				),
			),
			m.noBoundCancel,
		)

		return
	}

	abLenOffset := len(aLimbs.Limbs) - len(bLimbs.Limbs)

	// Constraint for the least significant limb
	// result[last] + carry[last-1] * 2^limbBitSize = a[last] + b[last]
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_CONSTRAINT_LSB", m.name),
		sym.Mul(
			m.mask,
			sym.Sub(
				sym.Add(resultLimbs.Limbs[lastLimbIdx], sym.Mul(limbMax, m.carry.Limbs[lastLimbIdx-1])),
				sym.Add(aLimbs.Limbs[lastLimbIdx], bLimbs.Limbs[lastLimbIdx-abLenOffset]),
			),
		),
		m.noBoundCancel,
	)

	// Constraints for all limbs except the most significant one
	// result[i] + carry[i-1] * 2^limbBitSize = a[i] + b[i] + carry[i]
	for i := lastLimbIdx - 1; i > 0; i-- {
		// The number of limbs in bLimbs may be less than in aLimbs
		scndOp := sym.Add(aLimbs.Limbs[i], m.carry.Limbs[i])
		if lastLimbIdx-i > abLenOffset {
			scndOp = sym.Add(scndOp, bLimbs.Limbs[i-abLenOffset])
		}

		comp.InsertGlobal(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_CONSTRAINT_%v", m.name, i),
			sym.Mul(
				m.mask,
				sym.Sub(
					sym.Add(resultLimbs.Limbs[i], sym.Mul(limbMax, m.carry.Limbs[i-1])),
					scndOp,
				),
			),
			m.noBoundCancel,
		)
	}

	// The number of limbs in bLimbs may be less than in aLimbs
	scndOp := sym.Add(aLimbs.Limbs[0], m.carry.Limbs[0])
	if len(aLimbs.Limbs) == len(bLimbs.Limbs) {
		scndOp = sym.Add(scndOp, bLimbs.Limbs[0])
	}

	// Constraint for the most significant limb (no carry out)
	// result[0] = a[0] + b[0] + carry[0]
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_ADD_COL_TO_LIMBS_CONSTRAINT_MSB", m.name),
		sym.Mul(
			m.mask,
			sym.Sub(
				resultLimbs.Limbs[0],
				scndOp,
			),
		),
		m.noBoundCancel,
	)
}

// Run executes the addition of a column to the limbs, assigning the
// results to the result and carry columns.
func (m *MultiLimbAdd) Run(run *wizard.ProverRuntime) {
	if m.isAddition {
		m.runAddition(run)
	} else {
		m.runSubtraction(run)
	}
}

func (m *MultiLimbAdd) runAddition(run *wizard.ProverRuntime) {
	aLimbs := make([][]field.Element, len(m.aLimbs.Limbs))
	for i := range m.aLimbs.Limbs {
		aLimbs[i] = m.aLimbs.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	bLimbs := make([][]field.Element, len(m.bLimbs.Limbs))
	for i := range m.bLimbs.Limbs {
		bLimbs[i] = m.bLimbs.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	var res []*common.VectorBuilder
	if !m.withResult {
		res = make([]*common.VectorBuilder, len(m.result.Limbs))
		for i := range m.result.Limbs {
			res[i] = common.NewVectorBuilder(m.result.Limbs[i])
		}
	}

	carry := make([]*common.VectorBuilder, len(m.carry.Limbs))
	for i := range m.carry.Limbs {
		carry[i] = common.NewVectorBuilder(m.carry.Limbs[i])
	}

	limbMax := uint64(1) << m.aLimbs.LimbBitSize
	lastLimbIdx := len(m.aLimbs.Limbs) - 1
	lastCarryIdx := len(m.carry.Limbs) - 1

	nbRows := m.bLimbs.Limbs[0].Size()
	for row := 0; row < nbRows; row++ {
		carryVals := make([]uint64, len(m.carry.Limbs))

		sum := aLimbs[lastLimbIdx][row].Uint64()

		if lastLimbIdx < len(bLimbs) {
			sum += bLimbs[lastLimbIdx][row].Uint64()
		}

		if res != nil {
			res[lastLimbIdx].PushField(field.NewElement(sum % limbMax))
		}

		if len(m.aLimbs.Limbs) > 1 {
			carryVals[lastCarryIdx] = sum / limbMax
			carry[lastCarryIdx].PushField(field.NewElement(carryVals[lastCarryIdx]))
		}

		// Process intermediate limbs
		for i := lastLimbIdx - 1; i > 0; i-- {
			sum = aLimbs[i][row].Uint64() + carryVals[i]

			// The number of limbs in bLimbs may be less than in aLimbs
			if i < len(bLimbs) {
				sum += bLimbs[i][row].Uint64()
			}

			if res != nil {
				res[i].PushField(field.NewElement(sum % limbMax))
			}

			carryVals[i-1] = sum / limbMax
			carry[i-1].PushField(field.NewElement(carryVals[i-1]))
		}

		// Process the most significant limb
		if len(m.aLimbs.Limbs) > 1 && res != nil {
			sum = aLimbs[0][row].Uint64() + bLimbs[0][row].Uint64() + carryVals[0]
			res[0].PushField(field.NewElement(sum))
		}
	}

	for i := range res {
		res[i].PadAndAssign(run, field.Zero())
	}

	for i := range carry {
		carry[i].PadAndAssign(run, field.Zero())
	}
}

func (m *MultiLimbAdd) runSubtraction(run *wizard.ProverRuntime) {
	aLimbs := make([][]field.Element, len(m.aLimbs.Limbs))
	for i := range m.aLimbs.Limbs {
		aLimbs[i] = m.aLimbs.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	bLimbs := make([][]field.Element, len(m.bLimbs.Limbs))
	for i := range m.bLimbs.Limbs {
		bLimbs[i] = m.bLimbs.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	var res []*common.VectorBuilder
	if !m.withResult {
		res = make([]*common.VectorBuilder, len(m.result.Limbs))
		for i := range m.result.Limbs {
			res[i] = common.NewVectorBuilder(m.result.Limbs[i])
		}
	}

	carry := make([]*common.VectorBuilder, len(m.carry.Limbs))
	for i := range m.carry.Limbs {
		carry[i] = common.NewVectorBuilder(m.carry.Limbs[i])
	}

	limbMax := uint64(1) << m.aLimbs.LimbBitSize
	lastLimbIdx := len(m.aLimbs.Limbs) - 1
	lastCarryIdx := len(m.carry.Limbs) - 1

	nbRows := m.bLimbs.Limbs[0].Size()
	for row := 0; row < nbRows; row++ {
		carryVals := make([]uint64, len(m.carry.Limbs))

		aLimbUint64 := aLimbs[lastLimbIdx][row].Uint64()
		bLimbUint64 := uint64(0)
		if lastLimbIdx < len(bLimbs) {
			bLimbUint64 = bLimbs[lastLimbIdx][row].Uint64()
		}

		carryValue := uint64(0)
		if len(m.aLimbs.Limbs) > 1 {
			carryVals[lastCarryIdx] = 0
			if bLimbUint64 > aLimbUint64 {
				carryValue = limbMax
				carryVals[lastCarryIdx] = 1
			}

			carry[lastCarryIdx].PushField(field.NewElement(carryVals[lastCarryIdx]))
		}

		if res != nil {
			sub := carryValue + aLimbUint64 - bLimbUint64
			res[lastLimbIdx].PushField(field.NewElement(sub))
		}

		// Process intermediate limbs
		for i := lastLimbIdx - 1; i > 0; i-- {
			aLimbUint64 = aLimbs[i][row].Uint64()
			bLimbUint64 = uint64(0)
			if lastLimbIdx < len(bLimbs) {
				bLimbUint64 = bLimbs[i][row].Uint64()
			}

			carryValue = 0
			if len(m.aLimbs.Limbs) > 1 {
				carryVals[i-1] = 0
				if bLimbUint64+carryVals[i] > aLimbUint64 {
					carryValue = limbMax
					carryVals[i-1] = 1
				}

				carry[i-1].PushField(field.NewElement(carryVals[i-1]))
			}

			if res != nil {
				sub := carryValue + aLimbUint64 - bLimbUint64 - carryVals[i]
				res[i].PushField(field.NewElement(sub))
			}
		}

		// Process the most significant limb
		if len(m.aLimbs.Limbs) > 1 && res != nil {
			sub := aLimbs[0][row].Uint64() - bLimbs[0][row].Uint64() - carryVals[0]
			res[0].PushField(field.NewElement(sub % limbMax))
		}
	}

	for i := range res {
		res[i].PadAndAssign(run, field.Zero())
	}

	for i := range carry {
		carry[i].PadAndAssign(run, field.Zero())
	}
}
