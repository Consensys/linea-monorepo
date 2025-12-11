package modexp2

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/emulated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// base exponent modulus result from arithmetization

const (
	smallModExpSize = 256
	largeModExpSize = 8192

	limbSizeBits             = 128
	nbLargeModexpLimbs       = largeModExpSize / limbSizeBits
	nbSmallModexpLimbs       = smallModExpSize / limbSizeBits
	modexpNumRowsPerInstance = nbLargeModexpLimbs * 4 // 4 operands: base, exponent, modulus, result

	roundNr = 0
)

// input collects references to the columns of the arithmetization containing
// the Modexp statements. These columns are constrained via a projection query
// to describe the same statement as what is being stated in the antichamber
// module. They are also used as a data source to assign the columns of the
// antichamber module.
//
// The columns provided here are columns from the BLK_MDXP module.
type Input struct {
	Settings Settings
	// Binary column indicating if we have base limbs
	IsModExpBase ifaces.Column
	// Binary column indicating if we have exponent limbs
	IsModExpExponent ifaces.Column
	// Binary column indicating if we have modulus limbs
	IsModExpModulus ifaces.Column
	// Binary column indicating if we have result limbs
	IsModExpResult ifaces.Column
	// IsModexp is a constructed column constrained to be equal to the sum of the
	// 4 above columns.
	IsModExp ifaces.Column
	// Multiplexed column containing limbs for base, exponent, modulus, and result
	Limbs ifaces.Column
}

type Settings struct {
	MaxNbInstance256, MaxNbInstanceLarge int
	NbInstancesPerCircuitModexp256       int
	NbInstancesPerCircuitModexpLarge     int
}

func newZkEVMInput(comp *wizard.CompiledIOP, settings Settings) *Input {
	return &Input{
		Settings:         settings,
		IsModExpBase:     comp.Columns.GetHandle("blake2fmodexpdata.IS_MODEXP_BASE"),
		IsModExpExponent: comp.Columns.GetHandle("blake2fmodexpdata.IS_MODEXP_EXPONENT"),
		IsModExpModulus:  comp.Columns.GetHandle("blake2fmodexpdata.IS_MODEXP_MODULUS"),
		IsModExpResult:   comp.Columns.GetHandle("blake2fmodexpdata.IS_MODEXP_RESULT"),
		Limbs:            comp.Columns.GetHandle("blake2fmodexpdata.LIMB"),
	}
}

type Module struct {
	*Input

	IsSmall ifaces.Column
	IsLarge ifaces.Column
	Small   *Modexp
	Large   *Modexp
}

func newModule(comp *wizard.CompiledIOP, input *Input) *Module {
	var (
		settings = input.Settings
		mod      = &Module{
			IsSmall: comp.InsertCommit(0, "MODEXP_IS_SMALL", input.Limbs.Size()),
			IsLarge: comp.InsertCommit(0, "MODEXP_IS_LARGE", input.Limbs.Size()),
			Input:   input,
		}
	)
	input.IsModExp = comp.InsertCommit(0, "MODEXP_INPUT_IS_MODEXP", input.Limbs.Size())
	comp.RegisterProverAction(roundNr, mod)
	// run later to ensure we have the assignments already done
	mod.Small = newModexp(comp, "MODEXP_SMALL", input, mod.IsSmall, smallModExpSize, settings.MaxNbInstance256)
	mod.Large = newModexp(comp, "MODEXP_LARGE", input, mod.IsLarge, largeModExpSize, settings.MaxNbInstanceLarge)

	// pragmas.MarkRightPadded(mod.IsActive)

	mod.csIsModExp(comp)
	// mod.csIsSmallAndLarge(comp)
	// mod.csToCirc(comp)
	// TODO: modexp constraints for projection

	// comp.InsertProjection(
	// 	"MODEXP_BLKMDXP_PROJECTION",
	// 	query.ProjectionInput{ColumnA: []ifaces.Column{mod.Input.Limbs},
	// 		ColumnB: []ifaces.Column{mod.Limbs},
	// 		FilterA: mod..IsModExp,
	// 		FilterB: mod.IsActive})
	return mod
}

func (m *Module) Run(run *wizard.ProverRuntime) {
	m.assignIsActive(run)
	m.assignIsSmallOrLarge(run)
}

// csIsModExp constructs, constraints and set the [isModexpColumn]
func (m *Module) csIsModExp(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"MODEXP_IS_MODEXP_WELL_CONSTRUCTED",
		sym.Sub(
			m.IsModExp,
			m.IsModExpBase,
			m.IsModExpExponent,
			m.IsModExpModulus,
			m.IsModExpResult,
		),
	)
}

// assignIsActive evaluates and assigns the IsActive column
func (m *Module) assignIsActive(run *wizard.ProverRuntime) {

	var (
		isBase     = m.IsModExpBase.GetColAssignment(run)
		isExponent = m.IsModExpExponent.GetColAssignment(run)
		isModulus  = m.IsModExpModulus.GetColAssignment(run)
		isResult   = m.IsModExpResult.GetColAssignment(run)
		isActive   = smartvectors.Add(isBase, isExponent, isModulus, isResult)
	)

	run.AssignColumn(m.IsModExp.GetColID(), isActive)
}

func (m *Module) assignIsSmallOrLarge(run *wizard.ProverRuntime) {
	var (
		srcIsModExp = m.IsModExp.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcLimbs    = m.Limbs.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	var (
		dstIsSmall = common.NewVectorBuilder(m.IsSmall)
		dstIsLarge = common.NewVectorBuilder(m.IsLarge)
	)
	checkSmall := func(ptr int) bool {
		for k := 0; k < nbLargeModexpLimbs-nbSmallModexpLimbs; k++ {
			for j := 0; j < 4; j++ {
				if !srcLimbs[ptr+k+j*nbLargeModexpLimbs].IsZero() {
					return false
				}
			}
		}
		return true
	}

	for ptr := 0; ptr < len(srcLimbs); {
		if srcIsModExp[ptr].IsZero() {
			dstIsSmall.PushZero()
			dstIsLarge.PushZero()
			ptr++
			continue
		}
		// found a modexp instance. We read the modexp inputs from here.
		isSmall := checkSmall(ptr)
		for range modexpNumRowsPerInstance {
			if isSmall {
				dstIsSmall.PushOne()
				dstIsLarge.PushZero()
			} else {
				dstIsSmall.PushZero()
				dstIsLarge.PushOne()
			}
		}
		ptr += modexpNumRowsPerInstance
	}
	dstIsLarge.PadAndAssign(run)
	dstIsSmall.PadAndAssign(run)
}

func NewModuleZkEvm(comp *wizard.CompiledIOP, settings Settings) *Module {
	return newModule(comp, newZkEVMInput(comp, settings))
}

type Modexp struct {
	instanceSize int // smallModExpSize or largeModExpSize
	nbLimbs      int // 256/limbsize for small case, 8096/limbsize for large case

	IsActive ifaces.Column // the active flag for this module
	*Input

	PrevAccumulator emulated.Limbs
	CurrAccumulator emulated.Limbs
	Modulus         emulated.Limbs
	ExponentBits    emulated.Limbs
	Base            emulated.Limbs
	Mone            emulated.Limbs // mod - 1
}

func newModexp(comp *wizard.CompiledIOP, name string, input *Input, isActive ifaces.Column, instanceSize int, nbInstances int) *Modexp {
	// we omit the Modexp module when there is no instance to process
	if nbInstances == 0 {
		return nil
	}
	var nbLimbs, nbRows int
	switch instanceSize {
	case smallModExpSize:
		nbLimbs = nbSmallModexpLimbs
		nbRows = nbInstances * smallModExpSize
	case largeModExpSize:
		nbLimbs = nbLargeModexpLimbs
		nbRows = nbInstances * largeModExpSize
	default:
		utils.Panic("unsupported modexp instance size: %d", instanceSize)
	}
	// we assume that the limb widths are aligned between arithmetization and emulated field ops
	prevAcc := emulated.NewLimbs(comp, roundNr, name+"_PREV_ACC", nbLimbs, nbRows)
	currAcc := emulated.NewLimbs(comp, roundNr, name+"_CURR_ACC", nbLimbs, nbRows)
	modulus := emulated.NewLimbs(comp, roundNr, name+"_MODULUS", nbLimbs, nbRows)
	exponentBits := emulated.NewLimbs(comp, roundNr, name+"_EXPONENT_BITS", nbLimbs, nbRows)
	base := emulated.NewLimbs(comp, roundNr, name+"_BASE", nbLimbs, nbRows)
	mone := emulated.NewLimbs(comp, roundNr, name+"_MONE", nbLimbs, nbRows)

	me := &Modexp{
		instanceSize:    instanceSize,
		nbLimbs:         nbLimbs,
		PrevAccumulator: prevAcc,
		CurrAccumulator: currAcc,
		Modulus:         modulus,
		ExponentBits:    exponentBits,
		Base:            base,
		Mone:            mone,
		IsActive:        isActive,
		Input:           input,
	}
	// register prover action before emulated evaluation so that the limb assignements are available
	comp.RegisterProverAction(roundNr, me)
	emulated.EmulatedEvaluation(comp, name+"_EVAL", limbSizeBits, modulus, [][]emulated.Limbs{
		// R_i = R_{i-1}^2 + R_{i-1}^2*e_i*base - R_{i-1}^2 * e_i
		{prevAcc, prevAcc}, {prevAcc, prevAcc, exponentBits, base}, {mone, prevAcc, prevAcc, exponentBits}, {mone, currAcc},
	})

	// TODO: add constraints that accumulator is copied correctly
	// TODO: add constraint that the base is correctly assigned
	// TODO: add constraint that the modulus is correctly assigned over all rows
	// TODO: add constraint for exponent bits assignment

	return me
}

func (m *Modexp) Run(run *wizard.ProverRuntime) {
	m.assignLimbs(run)
}

func (m *Modexp) assignLimbs(run *wizard.ProverRuntime) {
	var (
		srcLimbs    = m.Input.Limbs.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsModExp = m.IsModExp.GetColAssignment(run).IntoRegVecSaveAlloc()
		nbRows      = m.IsModExp.Size()
	)
	var (
		dstPrevAccLimbs  = make([]*common.VectorBuilder, len(m.PrevAccumulator.Columns))
		dstCurrAccLimbs  = make([]*common.VectorBuilder, len(m.CurrAccumulator.Columns))
		dstModulusLimbs  = make([]*common.VectorBuilder, len(m.Modulus.Columns))
		dstExponentLimbs = make([]*common.VectorBuilder, len(m.ExponentBits.Columns))
		dstBaseLimbs     = make([]*common.VectorBuilder, len(m.Base.Columns))
		dstMoneLimbs     = make([]*common.VectorBuilder, len(m.Mone.Columns))
	)
	for i := range dstPrevAccLimbs {
		dstPrevAccLimbs[i] = common.NewVectorBuilder(m.PrevAccumulator.Columns[i])
	}
	for i := range dstCurrAccLimbs {
		dstCurrAccLimbs[i] = common.NewVectorBuilder(m.CurrAccumulator.Columns[i])
	}
	for i := range dstModulusLimbs {
		dstModulusLimbs[i] = common.NewVectorBuilder(m.Modulus.Columns[i])
	}
	for i := range dstExponentLimbs {
		dstExponentLimbs[i] = common.NewVectorBuilder(m.ExponentBits.Columns[i])
	}
	for i := range dstBaseLimbs {
		dstBaseLimbs[i] = common.NewVectorBuilder(m.Base.Columns[i])
	}
	for i := range dstMoneLimbs {
		dstMoneLimbs[i] = common.NewVectorBuilder(m.Mone.Columns[i])
	}
	buf := make([]*big.Int, m.nbLimbs)
	for i := range buf {
		buf[i] = new(big.Int)
	}
	baseBi := new(big.Int)
	exponentBi := new(big.Int)
	modulusBi := new(big.Int)
	moneBi := new(big.Int)
	expectedBi := new(big.Int)
	exponentBits := make([]uint, m.instanceSize) // in reversed order MSB first

	prevAccumulatorBi := new(big.Int)
	currAccumulatorBi := new(big.Int)

	instMod := make([]field.Element, m.nbLimbs)
	instMone := make([]field.Element, m.nbLimbs)
	instBase := make([]field.Element, m.nbLimbs)

	expectedZeroPadding := nbLargeModexpLimbs - m.nbLimbs

	// scan through all the rows to find the modexp instances
	for ptr := 0; ptr < nbRows; {
		// we didn't find anything, move on
		if srcIsModExp[ptr].IsZero() {
			ptr++
			continue
		}
		// found a modexp instance. We read the modexp inputs from here.
		// first we do a sanity-check to see that we have expected number of inputs everywhere:
		if len(srcLimbs[ptr:]) < modexpNumRowsPerInstance {
			utils.Panic("A new modexp is starting but there is not enough rows (ptr=%v len(srcLimbs)=%v)", ptr, len(srcLimbs))
		}
		// we also sanity check that the inputs are consequtively marked as modexp
		for k := range modexpNumRowsPerInstance {
			if srcIsModExp[ptr+k].IsZero() {
				utils.Panic("A modexp instance is missing the modexp selector at row %v", ptr+k)
			}
		}
		// regardless if we are on small or large modexp, the arithmetization
		// sends us inputs on nbLargeModexpLimbs rows. So for small modexp we
		// only read the first nbSmallModexpLimbs rows and ignore the rest.
		base := srcLimbs[ptr+expectedZeroPadding : ptr+nbLargeModexpLimbs]
		exponent := srcLimbs[ptr+nbLargeModexpLimbs+expectedZeroPadding : ptr+2*nbLargeModexpLimbs]
		modulus := srcLimbs[ptr+2*nbLargeModexpLimbs+expectedZeroPadding : ptr+3*nbLargeModexpLimbs]
		expected := srcLimbs[ptr+3*nbLargeModexpLimbs+expectedZeroPadding : ptr+4*nbLargeModexpLimbs]

		// assign the big-int values for intermediate computation
		for i := range base {
			base[i].BigInt(buf[i])
			instBase[i].SetBigInt(buf[i])
		}
		if err := emulated.IntLimbRecompose(buf, limbSizeBits, baseBi); err != nil {
			utils.Panic("could not convert base limbs to big.Int: %v", err)
		}
		for i := range exponent {
			exponent[i].BigInt(buf[i])
		}
		if err := emulated.IntLimbRecompose(buf, limbSizeBits, exponentBi); err != nil {
			utils.Panic("could not convert exponent limbs to big.Int: %v", err)
		}
		for i := range modulus {
			modulus[i].BigInt(buf[i])
			instMod[i].SetBigInt(buf[i])
		}
		if err := emulated.IntLimbRecompose(buf, limbSizeBits, modulusBi); err != nil {
			utils.Panic("could not convert modulus limbs to big.Int: %v", err)
		}
		for i := range expected {
			expected[i].BigInt(buf[i])
		}
		if err := emulated.IntLimbRecompose(buf, limbSizeBits, expectedBi); err != nil {
			utils.Panic("could not convert result limbs to big.Int: %v", err)
		}
		// compute mod - 1
		moneBi.Sub(modulusBi, big.NewInt(1))
		if err := emulated.IntLimbDecompose(moneBi, limbSizeBits, buf); err != nil {
			utils.Panic("could not decompose mod-1 into limbs: %v", err)
		}
		for j := range instMone {
			instMone[j].SetBigInt(buf[j])
		}
		// extract exponent bits
		for i := 0; i < m.instanceSize; i++ {
			exponentBits[m.instanceSize-1-i] = uint(exponentBi.Bit(i))
		}
		// initialize all intermediate values and assign them
		prevAccumulatorBi.SetInt64(1)
		for i := range exponentBits {
			currAccumulatorBi.Mul(prevAccumulatorBi, prevAccumulatorBi)
			if exponentBits[i] == 1 {
				currAccumulatorBi.Mul(currAccumulatorBi, baseBi)
			}
			currAccumulatorBi.Mod(currAccumulatorBi, modulusBi)
			if err := emulated.IntLimbDecompose(prevAccumulatorBi, limbSizeBits, buf); err != nil {
				utils.Panic("could not decompose prevAccumulatorBi into limbs: %v", err)
			}
			for j := range m.PrevAccumulator.Columns {
				var f field.Element
				f.SetBigInt(buf[j])
				dstPrevAccLimbs[j].PushField(f)
			}
			if err := emulated.IntLimbDecompose(currAccumulatorBi, limbSizeBits, buf); err != nil {
				utils.Panic("could not decompose currAccumulatorBi into limbs: %v", err)
			}
			for j := range m.CurrAccumulator.Columns {
				var f field.Element
				f.SetBigInt(buf[j])
				dstCurrAccLimbs[j].PushField(f)
			}
			// swap
			prevAccumulatorBi.Set(currAccumulatorBi)

			// set the exponent bit
			dstExponentLimbs[0].PushInt(int(exponentBits[i]))
			for j := range m.nbLimbs - 1 {
				dstExponentLimbs[j+1].PushInt(0)
			}
			// and also set the constant per MODEXP-instance values
			for j := range instBase {
				dstBaseLimbs[j].PushField(instBase[j])
			}
			for j := range instMod {
				dstModulusLimbs[j].PushField(instMod[j])
			}
			for j := range instMone {
				dstMoneLimbs[j].PushField(instMone[j])
			}
		}
		ptr += modexpNumRowsPerInstance
	}
	// commit all built vectors
	for i := range dstPrevAccLimbs {
		dstPrevAccLimbs[i].PadAndAssign(run)
	}
	for i := range dstCurrAccLimbs {
		dstCurrAccLimbs[i].PadAndAssign(run)
	}
	for i := range dstModulusLimbs {
		dstModulusLimbs[i].PadAndAssign(run)
	}
	for i := range dstExponentLimbs {
		dstExponentLimbs[i].PadAndAssign(run)
	}
	for i := range dstBaseLimbs {
		dstBaseLimbs[i].PadAndAssign(run)
	}
	for i := range dstMoneLimbs {
		dstMoneLimbs[i].PadAndAssign(run)
	}
}
