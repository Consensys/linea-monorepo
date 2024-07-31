package dedicated

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

/*
	InsertIsTargetValue is a query to indicate where a target value in the column appears.

It receives three main inputs;

- the TargetValue

- columnA that is subject to the search

- columnB, a binary column that is 1 where columnA equals TargetValue, and is zero everywhere else

The query IsTargetValue asserts that columnB indeed has the right form, namely;
for the row-index i, column B[i] =1 iff column A[i] =TargetValue.

	Note : The query does not check that colB is binary
*/
func InsertIsTargetValue(
	comp *wizard.CompiledIOP,
	round int,
	queryName ifaces.QueryID,
	targetVal field.Element,
	colA ifaces.Column,
	colB any,
) {
	// declare the new column colC
	colC := comp.InsertCommit(round, ifaces.ColIDf(string(queryName)), colA.Size())
	// to have colB[i] = 1 iff colA[i]=targetValue
	// impose three following constrains (where t = targetValue -colA)
	//
	// 1. t * (t * colC -1) =0
	//
	// 2. t * colB = 0
	//
	// 3. (t * colC + colB - 1) = 0
	//
	// t :=  targetVal - colA
	t := symbolic.Sub(targetVal, colA)

	// if t[i] !=0 ---> colC[i] = t^(-1)
	// i.e., t * (t * colC -1) =0
	expr := symbolic.Mul(symbolic.Sub(symbolic.Mul(t, colC), 1), t)
	comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v", string(queryName), 1), expr)

	// t * colB = 0
	expr = symbolic.Mul(t, colB)
	comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v", string(queryName), 2), expr)

	// (t * colC + colB - 1) = 0
	expr = symbolic.Sub(symbolic.Add(symbolic.Mul(t, colC), colB), 1)
	comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v", string(queryName), 3), expr)
	comp.SubProvers.AppendToInner(round,
		func(run *wizard.ProverRuntime) {
			assignInverse(run, targetVal, colA, colC)
		},
	)
}
func assignInverse(run *wizard.ProverRuntime, targetVal field.Element, colA, colC ifaces.Column) {
	cola := colA.GetColAssignment(run).IntoRegVecSaveAlloc()
	colCWit := make([]field.Element, len(cola))
	var res field.Element
	for i := range cola {
		res.Sub(&targetVal, &cola[i])
		colCWit[i].Inverse(&res)
	}
	run.AssignColumn(colC.GetColID(), smartvectors.NewRegular(colCWit))

}
