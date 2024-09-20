package dedicated

import (
	"strconv"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// AccumulatorInputs stores the inputs for  [AccumulateUpToMax] function.
type AccumulatorInputs struct {
	// ColA is subject to the acumulation.
	ColA ifaces.Column
	// accumulating should not goes beyond the given maxValue.
	MaxValue int
}

/*
accumulateUpToMax stores the intermediate columns for [AccumulateUpToMax] function.
AccumulateUpToMax accumulates the values of column ColA (upward) up to the given max,

after reaching the max, it restarts the accumulation.

To prevent the accumulation going beyond the maximum,
we force the accumulator to reach the max at the first row.

The column IsMax indicate where the accumulator reaches the max,

	it is a fully-constrained binary column.
*/
type accumulateUpToMax struct {
	Inputs AccumulatorInputs
	// It is 1 when the accumulator reaches the max value.
	IsMax    ifaces.Column
	IsActive ifaces.Column
	// the  ProverAction for IsZero()
	pa wizard.ProverAction
	// It accumulate the elements from ColA.
	Accumulator ifaces.Column
	// size of the accumulator
	Size int
}

func AccumulateUpToMax(comp *wizard.CompiledIOP, maxValue int, colA, isActive ifaces.Column) *accumulateUpToMax {
	var (
		uniqueID  = strconv.Itoa(len(comp.ListCommitments()))
		size      = colA.Size()
		createCol = common.CreateColFn(comp, "ACCUMULATE_UP_TO_MAX_"+uniqueID, size)
	)

	acc := &accumulateUpToMax{
		Inputs: AccumulatorInputs{MaxValue: maxValue,
			ColA: colA},
		Accumulator: createCol("Accumulator"),
		IsActive:    isActive,
		Size:        size,
	}

	acc.IsMax, acc.pa = dedicated.IsZero(comp, sym.Sub(maxValue, acc.Accumulator))

	// Constraints over the accumulator
	// Accumulator[last] =ColA[last]
	comp.InsertLocal(0, ifaces.QueryIDf("AccCLDLenSpaghetti_Loc_"+uniqueID),
		sym.Sub(
			column.Shift(acc.Accumulator, -1), column.Shift(acc.Inputs.ColA, -1),
		),
	)

	// Accumulator[i] = Accumulator[i+1]*(1-acc.IsMax[i+1]) +ColA[i]; i standing for row-index.
	res := sym.Sub(1, column.Shift(acc.IsMax, 1)) // 1-acc.IsMax[i+1]

	comp.InsertGlobal(0, ifaces.QueryIDf("AccCLDLenSpaghetti_Glob_"+uniqueID),
		sym.Sub(
			sym.Add(
				sym.Mul(
					column.Shift(acc.Accumulator, 1), res),
				acc.Inputs.ColA),
			acc.Accumulator,
		),
	)

	// IsMax[0] = 1
	comp.InsertLocal(0, ifaces.QueryID("IS_1_AT_POS_0_"+uniqueID),
		sym.Sub(acc.IsMax, acc.IsActive),
	)

	return acc

}

func (la *accumulateUpToMax) Run(run *wizard.ProverRuntime) {

	var (
		column    = la.Inputs.ColA.GetColAssignment(run).IntoRegVecSaveAlloc()
		targetVal = la.Inputs.MaxValue
		acc       = make([]field.Element, len(column))
	)

	// populate Accumulator
	s := uint64(0)
	for j := len(column) - 1; j >= 0; j-- {

		s = s + column[j].Uint64()
		if s > uint64(targetVal) {
			utils.Panic("Should not reach a value larger than target value")
		}
		if s == uint64(targetVal) {
			acc[j] = field.NewElement(s)
			s = 0
		} else {
			acc[j] = field.NewElement(s)
		}

	}

	run.AssignColumn(la.Accumulator.GetColID(), smartvectors.RightZeroPadded(acc, la.Size))
	// assign IsMax
	la.pa.Run(run)
}
