package packing

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing/dedicated"
)

// laneRepackingInputs collects the inputs of the [newLane] function.
type laneRepackingInputs struct {
	spaghetti spaghettiCtx
	pckInp    PackingInput
}

// laneRepacking stores all the intermediate columns required to constraint the repacking.
type laneRepacking struct {
	Inputs *laneRepackingInputs
	// the set of lanes, they all have the same size
	Lanes ifaces.Column
	// the ProverAction for accumulateUpToMax
	paAccUpToMax wizard.ProverAction
	// the result of ProverAction for accumulateUpToMax
	// It is one if the lane is complete
	isLaneComplete ifaces.Column
	// the coefficient  for computing lanes from decomposedLimbs
	coeff ifaces.Column
	// it is 1 on the effective part of lane module.
	IsLaneActive ifaces.Column
	// It is 1 if the lane is the first lane of the new hash.
	IsFirstLaneOfNewHash ifaces.Column
	// the size of the columns in this submodule.
	Size int
}

// It imposes all the constraints for correct repacking of decompsedLimbs into lanes.
func newLane(comp *wizard.CompiledIOP, spaghetti spaghettiCtx, pckInp PackingInput) laneRepacking {
	var (
		size                  = utils.NextPowerOfTwo(pckInp.PackingParam.NbOfLanesPerBlock() * pckInp.MaxNumBlocks)
		createCol             = common.CreateColFn(comp, LANE+"_"+pckInp.Name, size)
		isFirstSliceOfNewHash = spaghetti.newHashSp
		maxValue              = pckInp.PackingParam.LaneSizeBytes()
		decomposedLenSp       = spaghetti.decLenSp
		pa                    = dedicated.AccumulateUpToMax(comp, maxValue, decomposedLenSp, spaghetti.filterSpaghetti)
		spaghettiSize         = spaghetti.spaghettiSize
	)

	l := laneRepacking{
		Inputs: &laneRepackingInputs{
			pckInp:    pckInp,
			spaghetti: spaghetti,
		},

		Lanes:                createCol("Lane"),
		IsFirstLaneOfNewHash: createCol("IsFirstLaneOfNewHash"),
		IsLaneActive:         createCol("IsLaneActive"),
		coeff:                comp.InsertCommit(0, ifaces.ColIDf("Coefficient_"+pckInp.Name), spaghettiSize),

		paAccUpToMax:   pa,
		isLaneComplete: pa.IsMax,
		Size:           size,
	}

	// Declare the constraints
	commonconstraints.MustBeActivationColumns(comp, l.IsLaneActive)

	// Constraints on Coeff, it is a accumulator of decomposedLenPowers over isLaneComplete.
	// coef[last_active_row] := 1
	// coeff[i] := coeff[i+1] * decomposedLenPowers[i+1] * (1-isLaneComplete[i]) + isLaneComplete[i]
	l.csCoeff(comp, spaghetti)

	// Constraints on the Recomposition of slices into the lanes
	l.csRecomposeToLanes(comp, spaghetti)

	// constraints over isFirstLaneOfNewHash
	// Project the isFirstLaneOfNewHash from isFirstSliceOfNewHash
	projection.InsertProjection(comp, ifaces.QueryIDf("Project_IsFirstLaneOfHash_"+pckInp.Name),
		[]ifaces.Column{isFirstSliceOfNewHash},
		[]ifaces.Column{l.IsFirstLaneOfNewHash},
		l.isLaneComplete, l.IsLaneActive)

	return l
}

// it declares the constraints over coefficients.
func (l *laneRepacking) csCoeff(comp *wizard.CompiledIOP, s spaghettiCtx) {
	var (
		partialCoeff = s.decLenPowerSp // decomposedLenPowers in spaghetti form.
		isActive     = s.filterSpaghetti
	)

	// coeff[last-active-row] = 1
	comp.InsertGlobal(
		0, ifaces.QueryIDf("%v_Coeff_In_Last_Active_Row", l.Inputs.pckInp.Name),
		sym.Mul(isActive,
			sym.Sub(column.Shift(isActive, 1), 1),
			sym.Sub(l.coeff,
				1)),
	)

	// coeff[last] = 1 // to cover the case where; last-active-row ==  last-row
	comp.InsertLocal(
		0, ifaces.QueryIDf("%v_Coeff-In_Last_Row", l.Inputs.pckInp.Name),
		sym.Sub(column.Shift(l.coeff, -1), 1),
	)

	// coeff[i] := coeff[i+1] * partialCoeff[i+1] * (1-isLaneComplete[i+1]) + isLaneComplete[i+1]
	res := sym.Mul(
		column.Shift(l.coeff, 1),
		column.Shift(partialCoeff, 1),
		sym.Sub(1, column.Shift(l.isLaneComplete, 1)),
	)
	res = sym.Add(res, column.Shift(l.isLaneComplete, 1))
	expr := sym.Mul(sym.Sub(l.coeff, res), column.Shift(isActive, 1))
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_Coefficient_Glob", l.Inputs.pckInp.Name), expr)
}

// It declares the constraints over the lanes
// Lanes are the recomposition of slices.
func (l *laneRepacking) csRecomposeToLanes(comp *wizard.CompiledIOP, s spaghettiCtx) {
	// compute the partitioned inner product
	//ipTaker[i] = (decomposedLimbs[i] * coeff[i]) + ipTracker[i+1]* (1- isLaneComplete[i+1])
	// Constraints on the Partitioned Inner-Products
	ipTracker := dedicated.InsertPartitionedIP(comp, l.Inputs.pckInp.Name+"_PIP_For_LaneRePacking",
		s.decLimbSp,
		l.coeff,
		l.isLaneComplete,
	)

	// Project the lanes from ipTracker over the Lane column.
	projection.InsertProjection(comp, ifaces.QueryIDf("%v_ProjectOverLanes", l.Inputs.pckInp.Name),
		[]ifaces.Column{ipTracker},
		[]ifaces.Column{l.Lanes},
		l.isLaneComplete, l.IsLaneActive,
	)

}

// It assigns the columns specific to the submodule
func (l *laneRepacking) Assign(run *wizard.ProverRuntime) {
	// assign the spaghetti forms
	l.Inputs.spaghetti.pa.Run(run)
	// assign the IsMax column from  the ProverAction (AccumulateUpToMax).
	l.paAccUpToMax.Run(run)
	// assign coeff
	l.assignCoeff(run)
	// assign the Lanes, isFirstLaneofNewHash , IsLaneActive
	l.assignLane(run)
}

// It assigns column coeff
func (l *laneRepacking) assignCoeff(
	run *wizard.ProverRuntime) {

	var (
		isLaneComplete       = l.isLaneComplete.GetColAssignment(run).IntoRegVecSaveAlloc()
		size                 = len(isLaneComplete)
		decomposedLenPowerSp = l.Inputs.spaghetti.decLenPowerSp
		partialCoeff         = decomposedLenPowerSp.GetColAssignment(run).IntoRegVecSaveAlloc()
		one                  = field.One()
		isActive             = l.Inputs.spaghetti.filterSpaghetti.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	//partialCoeff := decomposedLenPowers
	// partialCoeff = 0 over s.isActive = 0
	// coeff[i] := coeff[i+1] * partialCoeff[i+1] * (1-isLaneComplete[i+1]) + isLaneComplete[i+1]
	coeff := make([]field.Element, size)

	coeff[size-1] = field.One()

	var res, notComplete field.Element
	for i := size - 2; i >= 0; i-- {
		notComplete.Sub(&one, &isLaneComplete[i+1])
		res.Mul(&notComplete, &partialCoeff[i+1])
		res.Mul(&res, &coeff[i+1])
		coeff[i].Add(&res, &isLaneComplete[i+1])

		if isActive[i].IsOne() && isActive[i+1].IsZero() {
			coeff[i] = field.One()
		}
	}

	// assign the columns
	run.AssignColumn(l.coeff.GetColID(), smartvectors.RightZeroPadded(coeff, size))
}

// it assigns the lanes
func (l *laneRepacking) assignLane(run *wizard.ProverRuntime) {
	var (
		lane                 = common.NewVectorBuilder(l.Lanes)
		param                = l.Inputs.pckInp.PackingParam
		isFirstLaneofNewHash = common.NewVectorBuilder(l.IsFirstLaneOfNewHash)
		isActive             = common.NewVectorBuilder(l.IsLaneActive)
		laneBytes            = param.LaneSizeBytes()
		blocks, flag         = l.getBlocks(run, l.Inputs.pckInp)
	)
	var f field.Element

	if len(flag) != len(blocks) {
		utils.Panic("should have one flag per block numFlags=%v numBlocks=%v", len(flag), len(blocks))
	}

	for k, block := range blocks {
		if len(block) != param.BlockSizeBytes() {
			utils.Panic("blocks should be of size %v, but it is of size %v", param.BlockSizeBytes(), len(block))
		}
		for j := 0; j < param.NbOfLanesPerBlock(); j++ {
			laneBytes := block[j*laneBytes : j*laneBytes+laneBytes]
			f.SetBytes(laneBytes)
			lane.PushField(f)
			if flag[k] == 1 && j == 0 {
				isFirstLaneofNewHash.PushInt(1)
			} else {
				isFirstLaneofNewHash.PushInt(0)
			}
			isActive.PushInt(1)
		}

	}
	lane.PadAndAssign(run, field.Zero())
	isFirstLaneofNewHash.PadAndAssign(run, field.Zero())
	isActive.PadAndAssign(run, field.Zero())
}

//	It outputs the expected blocks.
//
// The expected blocks are used for the assignments of lanes.
func (l *laneRepacking) getBlocks(run *wizard.ProverRuntime, inp PackingInput) ([][]byte, []int) {

	var (
		s = 0
		// counter for the number of blocks
		ctr       = 0
		blockSize = inp.PackingParam.BlockSizeBytes()
		imported  = inp.Imported
		limbs     = smartvectors.Window(imported.Limb.GetColAssignment(run))
		nBytes    = smartvectors.Window(imported.NByte.GetColAssignment(run))
		isNewHash = smartvectors.Window(imported.IsNewHash.GetColAssignment(run))
	)

	limbSerialized := [32]byte{}
	var stream []byte
	var block [][]byte
	var isFirstBlockOfHash []int
	isFirstBlockOfHash = append(isFirstBlockOfHash, 1)
	for pos := 0; pos < len(limbs); pos++ {
		nbyte := field.ToInt(&nBytes[pos])
		s = s + nbyte

		// Extract the limb, which is left aligned to the 16-th byte
		limbSerialized = limbs[pos].Bytes()

		usefulByte := limbSerialized[LEFT_ALIGNMENT : LEFT_ALIGNMENT+nbyte]
		if s > blockSize || s == blockSize {
			// extra part that should be moved to the next block
			s = s - blockSize
			res := usefulByte[:(nbyte - s)]
			newBlock := append(stream, res...)
			if len(newBlock) != blockSize {
				utils.Panic("could not extract the new Block")
			}
			block = append(block, newBlock)
			if pos+1 != len(limbs) && isNewHash[pos+1].IsOne() {
				// the next block is the first block of hash
				isFirstBlockOfHash = append(isFirstBlockOfHash, 1)
			} else if pos+1 != len(limbs) {
				isFirstBlockOfHash = append(isFirstBlockOfHash, 0)
			}
			stream = make([]byte, s)
			copy(stream, usefulByte[(nbyte-s):])
			ctr++
		} else {
			stream = append(stream, usefulByte[:]...)
		}
		// If we are on the last limb or if it is a new hash
		// the steam should be empty,
		// unless \sum_i NByte_i does not divide the blockSize (for i from the same hash)
		if pos+1 == len(limbs) || isNewHash[pos+1].Uint64() == 1 {
			if len(stream) != 0 {
				utils.Panic("The stream-length should be zero before launching a new hash/batch len(stream) = %v", len(stream))
			}
		}
		if ctr > inp.MaxNumBlocks {
			utils.Panic("the number of the blocks %v passes the limit %v", ctr, inp.MaxNumBlocks)
		}
	}

	// This corresponds to the edge-case were no blocks are being processed. In
	// that situation we can simply return empty lists. This addressment is
	// necessary because by default, [isFirstBlockOfHash] is initialized with
	// one value while the blocks are not.
	if len(block) == 0 {
		return [][]byte{}, []int{}
	}

	return block, isFirstBlockOfHash
}
