package modexp2

import (
	"fmt"
	"math/big"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/emulated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type Modexp struct {
	instanceSize int // smallModExpSize or largeModExpSize
	nbLimbs      int // 256/limbsize for small case, 8096/limbsize for large case

	IsActive ifaces.Column // the active flag for this module
	*Module

	PrevAccumulator emulated.Limbs
	CurrAccumulator emulated.Limbs
	Modulus         emulated.Limbs
	ExponentBits    emulated.Limbs
	Base            emulated.Limbs
	Mone            emulated.Limbs // mod - 1
}

func newModexp(comp *wizard.CompiledIOP, name string, module *Module, isActive ifaces.Column, instanceSize int, nbInstances int) *Modexp {
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
		Module:          module,
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
	// TODO: modexp constraints for projection

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
