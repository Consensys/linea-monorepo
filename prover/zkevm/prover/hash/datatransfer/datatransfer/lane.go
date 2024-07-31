package datatransfer

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
)

type lane struct {
	// All the lanes of all the blocks
	lane ifaces.Column

	// A lane may be spread over several limbs.
	// Lane is then obtained via an partitioned inner-product
	// as lane = \sum_i coeff[i]* cldSpaghetti[i].
	// Coeff represents the joint coefficients for stitching slices of the lane together.
	coeff ifaces.Column

	// The ipTracker tracks the partitioned inner-product.
	ipTracker ifaces.Column

	// IsLaneActive is 1 of the effective part of lane column.
	// It is of the form (1,...,1,O,..O). Namely starting with elements 1 , followed by zero.
	// The number of ones equals (the number of blocks * 17).
	isLaneActive ifaces.Column

	// It is 1 if the lane is the first lane of the new hash.
	isFirstLaneOfNewHash ifaces.Column

	// shifted version of (effective part of) isLaneComplete
	isLaneCompleteShifted ifaces.Column
}

func (l *lane) newLane(comp *wizard.CompiledIOP,
	round, maxRows, maxNumRowsForLane int,
	sCLD spaghettizedCLD,
) {
	// Declare the columns
	l.insertCommit(comp, round, maxRows, maxNumRowsForLane)

	// Declare the constraints

	// constraints over isLaneActive
	//
	// 1. they are binary
	// 2. isLaneActive starts with ones and ends with zeroes
	l.csIsLaneActive(comp, round, sCLD)

	// Constraints on Coeff, it is a accumulator of cldLenPowersSpaghetti over isLaneComplete
	// coef[0] := 1
	// coeff[i] := coeff[i-1] * cldLenPowersSpaghetti[i-1] * (1-isLaneComplete[i-1]) + isLaneComplete[i-1]
	l.csCoeff(comp, round, sCLD)

	// Constraints on the Recomposition of slices into the lanes
	l.csRecomposeToLanes(comp, round, sCLD)

	// constraints over isFirstLaneOfNewHash
	// Project the isFirstLaneOfNewHash from isFirstByteOfNewHash
	projection.InsertProjection(comp, ifaces.QueryIDf("Project_IsFirstLaneOfHash"),
		[]ifaces.Column{sCLD.isFirstSliceOfNewHash},
		[]ifaces.Column{l.isFirstLaneOfNewHash},
		l.isLaneCompleteShifted, l.isLaneActive)
}

// InsertCommit commits to the columns specific to the submodule.
func (l *lane) insertCommit(comp *wizard.CompiledIOP, round, maxRows, maxNumRowsForLane int) {
	l.lane = comp.InsertCommit(round, ifaces.ColIDf("Lane"), maxNumRowsForLane)
	l.coeff = comp.InsertCommit(round, ifaces.ColIDf("Coefficient"), maxRows)
	l.isLaneActive = comp.InsertCommit(round, ifaces.ColIDf("LaneIsActive"), maxNumRowsForLane)
	l.ipTracker = comp.InsertCommit(round, ifaces.ColIDf("IPTracker_Lane"), maxRows)
	l.isFirstLaneOfNewHash = comp.InsertCommit(round, ifaces.ColIDf("IsFirstLane_Of_NewHash"), maxNumRowsForLane)
	l.isLaneCompleteShifted = comp.InsertCommit(round, ifaces.ColIDf("IsLaneCompleteShifted"), maxRows)
}

// It declares the constraints over isLaneActive
func (l *lane) csIsLaneActive(comp *wizard.CompiledIOP, round int, s spaghettizedCLD) {
	// constraints over the right form of isLaneActive
	//
	// 1. It is binary
	//
	// 2. starts with ones and ends with zeroes.
	//
	// We don't check that it has the same number of ones as s.isLaneComplete,
	// since we later do a projection query that guarantees this fact.
	comp.InsertGlobal(round, ifaces.QueryIDf("IsLaneActive_IsBinary"),
		symbolic.Mul(l.isLaneActive, symbolic.Sub(1, l.isLaneActive)))

	a := symbolic.Sub(l.isLaneActive, column.Shift(l.isLaneActive, 1))
	// a should be binary (for constraint 2)
	comp.InsertGlobal(round, ifaces.QueryIDf("OnesThenZeroes"),
		symbolic.Mul(a, symbolic.Sub(1, a)))

	// constraints over isLaneCompleteShifted
	// isLaneCompleteShifted = shift(isLaneComplete,-1)
	// isLaneCompleteShifted[0] = 1
	comp.InsertGlobal(round, ifaces.QueryIDf("Shift_IsLaneComplete_Glob"),
		symbolic.Mul(symbolic.Sub(l.isLaneCompleteShifted,
			column.Shift(s.isLaneComplete, -1)), s.isActive))

	comp.InsertLocal(round, ifaces.QueryIDf("Shift_IsLaneComplete_Loc"),
		symbolic.Mul(symbolic.Sub(l.isLaneCompleteShifted, 1), s.isActive))
}

// it declares the constraints over coeff
func (l *lane) csCoeff(
	comp *wizard.CompiledIOP,
	round int,
	s spaghettizedCLD,
) {
	//local; coeff[0]=1
	comp.InsertLocal(round, ifaces.QueryIDf("Coeffcient_Loc"), symbolic.Mul(symbolic.Sub(l.coeff, 1), s.isActive))

	// coeff[i] := coeff[i-1] * partialCoeff[i-1] * (1-isLaneComplete[i-1]) + isLaneComplete[i-1]
	res := symbolic.Mul(column.Shift(l.coeff, -1), column.Shift(s.cldLenPowersSpaghetti, -1))
	res = symbolic.Mul(res, symbolic.Sub(1, column.Shift(s.isLaneComplete, -1)))
	res = symbolic.Add(res, column.Shift(s.isLaneComplete, -1))
	expr := symbolic.Mul(symbolic.Sub(l.coeff, res), s.isActive)
	comp.InsertGlobal(round, ifaces.QueryIDf("Coefficient_Glob"), expr)
}

// It declares the constraints over the lanes
// Lanes are the recomposition of slices in SpaghettizedCLD.
func (l *lane) csRecomposeToLanes(
	comp *wizard.CompiledIOP,
	round int,
	s spaghettizedCLD,
) {
	// compute the partitioned inner product
	//ipTaker[i] = (cldSpaghetti[i] * coeff[i]) + ipTracker[i-1]* isLaneComplete[i]
	// Constraints on the Partitioned Inner-Products
	dedicated.InsertPartitionedIP(comp, round, s.cldSpaghetti,
		l.coeff,
		column.Shift(s.isLaneComplete, -1),
		l.ipTracker)

	// Project the lane from ipTracker over the lane
	projection.InsertProjection(comp, ifaces.QueryIDf("ProjectOverLanes"),
		[]ifaces.Column{l.ipTracker},
		[]ifaces.Column{l.lane},
		s.isLaneComplete, l.isLaneActive)
}

// It assigns the columns specific to the submodule
func (l *lane) assignLane(
	run *wizard.ProverRuntime,
	iPadd importAndPadd,
	sCLD spaghettizedCLD,
	permTrace keccak.PermTraces,
	maxRows, maxNumRowsForLane int) {
	// assign l.isLaneActive
	run.AssignColumn(l.isLaneActive.GetColID(),
		smartvectors.RightZeroPadded(vector.Repeat(field.One(), len(permTrace.Blocks)*numLanesInBlock), maxNumRowsForLane))

	// assign coeff
	l.assignCoeff(run, sCLD, maxRows)

	// assign the lanes
	l.assignLaneColumn(run, maxNumRowsForLane, permTrace, iPadd)

	// isLanecompleteShifted
	witSize := smartvectors.Density(sCLD.isLaneComplete.GetColAssignment(run))
	isLaneComplete := sCLD.isLaneComplete.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	var shifted []field.Element
	if witSize != 0 {
		shifted = append(shifted, field.One())
		shifted = append(shifted, isLaneComplete[:witSize-1]...)
	}
	run.AssignColumn(l.isLaneCompleteShifted.GetColID(), smartvectors.RightZeroPadded(shifted, sCLD.isLaneComplete.Size()))
}

// It assigns column coeff
func (l *lane) assignCoeff(
	run *wizard.ProverRuntime,
	s spaghettizedCLD,
	maxNumRows int) {

	one := field.One()
	witSize := smartvectors.Density(s.isLaneComplete.GetColAssignment(run))
	isLaneComplete := s.isLaneComplete.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	//partialCoeff := cldLenPowersSpaghetti
	// Coeff[0] = 1
	// Coeff[i] := (Coeff[i-1] * partialCoeff[i-1] * (1-isLaneComplete[i-1])) + isLaneComplete[i-1]
	coeff := make([]field.Element, witSize)
	partialCoeff := s.cldLenPowersSpaghetti.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	if witSize != 0 {
		coeff[0] = field.One()
	}
	var res, notComplete field.Element
	for i := 1; i < witSize; i++ {
		notComplete.Sub(&one, &isLaneComplete[i-1])
		res.Mul(&notComplete, &partialCoeff[i-1])
		res.Mul(&res, &coeff[i-1])
		coeff[i].Add(&res, &isLaneComplete[i-1])
	}

	// assign the columns
	run.AssignColumn(l.coeff.GetColID(), smartvectors.RightZeroPadded(coeff, maxNumRows))
}

// It assigns the lanes
func (l *lane) assignLaneColumn(
	run *wizard.ProverRuntime,
	maxNumRows int,
	trace keccak.PermTraces,
	iPadd importAndPadd,
) {
	// Instead of building lane from the same formula defined in newLane(),
	// we assign it via trace of the permutation that is already tested.
	blocks := trace.Blocks
	var lane []field.Element
	for j := range blocks {
		for i := range blocks[0] {
			lane = append(lane, field.NewElement(blocks[j][i]))
		}
	}
	run.AssignColumn(l.lane.GetColID(), smartvectors.RightZeroPadded(lane, maxNumRows))

	// populate isFirstLaneOfNewHash
	witSize := smartvectors.Density(iPadd.isNewHash.GetColAssignment(run))
	isNewHash := iPadd.isNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	nByte := iPadd.nByte.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	sum := 0
	var t []int
	for j := range nByte {
		if isNewHash[j] == field.One() {
			// the length of the stream when we reach a newHash
			t = append(t, sum)
		}
		sum = sum + int(nByte[j].Uint64())
	}

	ctr := 0
	var col []field.Element
	for j := range lane {
		if ctr < len(t) && t[ctr] == 8*j {
			col = append(col, field.One())
			ctr++
		} else {
			col = append(col, field.Zero())
		}
	}

	//assign the columns
	run.AssignColumn(l.isFirstLaneOfNewHash.GetColID(), smartvectors.RightZeroPadded(col, maxNumRows))

}
