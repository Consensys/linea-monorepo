package distributedprojection

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/sirupsen/logrus"
)

type distribuedProjectionProverAction struct {
	Name               ifaces.QueryID
	FilterA, FilterB   []*symbolic.Expression
	ColumnA, ColumnB   []*symbolic.Expression
	HornerA, HornerB   []ifaces.Column
	HornerA0, HornerB0 []query.LocalOpening
	EvalCoin           []coin.Info
	IsA, IsB           []bool
}

func (pa *distribuedProjectionProverAction) Run(run *wizard.ProverRuntime) {
	for index := range pa.FilterA {
		if pa.IsA[index] && pa.IsB[index] {
			var (
				colA    = column.EvalExprColumn(run, pa.ColumnA[index].Board()).IntoRegVecSaveAlloc()
				fA      = column.EvalExprColumn(run, pa.FilterA[index].Board()).IntoRegVecSaveAlloc()
				colB    = column.EvalExprColumn(run, pa.ColumnB[index].Board()).IntoRegVecSaveAlloc()
				fB      = column.EvalExprColumn(run, pa.FilterB[index].Board()).IntoRegVecSaveAlloc()
				x       = run.GetRandomCoinField(pa.EvalCoin[index].Name)
				hornerA = poly.CmptHorner(colA, fA, x)
				hornerB = poly.CmptHorner(colB, fB, x)
			)
			run.AssignColumn(pa.HornerA[index].GetColID(), smartvectors.NewRegular(hornerA))
			run.AssignLocalPoint(pa.HornerA0[index].ID, hornerA[0])
			run.AssignColumn(pa.HornerB[index].GetColID(), smartvectors.NewRegular(hornerB))
			run.AssignLocalPoint(pa.HornerB0[index].ID, hornerB[0])
		} else if pa.IsA[index] && !pa.IsB[index] {
			var (
				colA    = column.EvalExprColumn(run, pa.ColumnA[index].Board()).IntoRegVecSaveAlloc()
				fA      = column.EvalExprColumn(run, pa.FilterA[index].Board()).IntoRegVecSaveAlloc()
				x       = run.GetRandomCoinField(pa.EvalCoin[index].Name)
				hornerA = poly.CmptHorner(colA, fA, x)
			)
			run.AssignColumn(pa.HornerA[index].GetColID(), smartvectors.NewRegular(hornerA))
			run.AssignLocalPoint(pa.HornerA0[index].ID, hornerA[0])
		} else if !pa.IsA[index] && pa.IsB[index] {
			var (
				colB    = column.EvalExprColumn(run, pa.ColumnB[index].Board()).IntoRegVecSaveAlloc()
				fB      = column.EvalExprColumn(run, pa.FilterB[index].Board()).IntoRegVecSaveAlloc()
				x       = run.GetRandomCoinField(pa.EvalCoin[index].Name)
				hornerB = poly.CmptHorner(colB, fB, x)
			)
			run.AssignColumn(pa.HornerB[index].GetColID(), smartvectors.NewRegular(hornerB))
			run.AssignLocalPoint(pa.HornerB0[index].ID, hornerB[0])
		} else {
			fmt.Errorf("Invalid prover assignment in distributed projection id: %v", pa.Name)
		}
	}

}

func (pa *distribuedProjectionProverAction) Push(comp *wizard.CompiledIOP, distributedprojection query.DistributedProjection) {
	for index, input := range distributedprojection.Inp {
		if input.IsAInModule && input.IsBInModule {
			pa.FilterA[index] = input.FilterA
			pa.FilterB[index] = input.FilterB
			pa.ColumnA[index] = input.ColumnA
			pa.ColumnB[index] = input.ColumnB
			pa.EvalCoin[index] = comp.Coins.Data(input.EvalCoin)
			pa.IsA[index] = true
			pa.IsB[index] = true

		} else if input.IsAInModule && !input.IsBInModule {
			pa.FilterA[index] = input.FilterA
			pa.ColumnA[index] = input.ColumnA
			pa.EvalCoin[index] = comp.Coins.Data(input.EvalCoin)
			pa.IsA[index] = true
			pa.IsB[index] = false

		} else if !input.IsAInModule && input.IsBInModule {
			pa.FilterB[index] = input.FilterB
			pa.ColumnB[index] = input.ColumnB
			pa.EvalCoin[index] = comp.Coins.Data(input.EvalCoin)
			pa.IsA[index] = false
			pa.IsB[index] = true

		} else {
			logrus.Errorf("Invalid distributed projection query while pushing prover action entries: %v", distributedprojection.ID)
		}
	}

}

func (pa *distribuedProjectionProverAction) RegisterQueries(comp *wizard.CompiledIOP, round int, distributedprojection query.DistributedProjection) {
	for index, input := range distributedprojection.Inp {
		if input.IsAInModule && input.IsBInModule {
			var (
				fA          = pa.FilterA[index]
				fAShifted   = shiftExpression(comp, fA, -1)
				colA        = pa.ColumnA[index]
				colAShifted = shiftExpression(comp, colA, -1)
				fB          = pa.FilterB[index]
				fBShifted   = shiftExpression(comp, fB, -1)
				colB        = pa.ColumnB[index]
				colBShifted = shiftExpression(comp, colB, -1)
			)
			pa.registerForCol(comp, fAShifted, colAShifted, input, "A", round, index)
			pa.registerForCol(comp, fBShifted, colBShifted, input, "B", round, index)
		} else if input.IsAInModule && !input.IsBInModule {
			var (
				fA          = pa.FilterA[index]
				fAShifted   = shiftExpression(comp, fA, -1)
				colA        = pa.ColumnA[index]
				colAShifted = shiftExpression(comp, colA, -1)
			)
			pa.registerForCol(comp, fAShifted, colAShifted, input, "A", round, index)
		} else if !input.IsAInModule && input.IsBInModule {
			var (
				fB          = pa.FilterB[index]
				fBShifted   = shiftExpression(comp, fB, -1)
				colB        = pa.ColumnB[index]
				colBShifted = shiftExpression(comp, colB, -1)
			)
			pa.registerForCol(comp, fBShifted, colBShifted, input, "B", round, index)
		} else {
			fmt.Errorf("Invalid prover action case for the distributed projection query %v", pa.Name)
		}
	}
}

func (pa *distribuedProjectionProverAction) registerForCol(
	comp *wizard.CompiledIOP,
	fShifted, colShifted *sym.Expression,
	input *query.DistributedProjectionInput,
	colName string,
	round int,
	index int,
) {
	switch colName {
	case "A":
		{
			pa.HornerA[index] = comp.InsertCommit(round, ifaces.ColIDf("%v_HORNER_A_%v", pa.Name, index), input.SizeA)
			comp.InsertGlobal(
				round,
				ifaces.QueryIDf("%v_HORNER_A_%v_GLOBAL", pa.Name, index),
				sym.Sub(
					pa.HornerA[index],
					sym.Mul(
						sym.Sub(1, pa.FilterA[index]),
						column.Shift(pa.HornerA[index], 1),
					),
					sym.Mul(
						pa.FilterA[index],
						sym.Add(
							pa.ColumnA[index],
							sym.Mul(
								pa.EvalCoin[index],
								column.Shift(pa.HornerA[index], 1),
							),
						),
					),
				),
			)
			comp.InsertLocal(
				round,
				ifaces.QueryIDf("%v_HORNER_A_LOCAL_END_%v", pa.Name, index),
				sym.Sub(
					column.Shift(pa.HornerA[index], -1),
					sym.Mul(fShifted, colShifted),
				),
			)
			pa.HornerA0[index] = comp.InsertLocalOpening(round, ifaces.QueryIDf("%v_HORNER_A0_%v", pa.Name, index), pa.HornerA[index])
		}
	case "B":
		{
			pa.HornerB[index] = comp.InsertCommit(round, ifaces.ColIDf("%v_HORNER_B_%v", pa.Name, index), input.SizeB)

			comp.InsertGlobal(
				round,
				ifaces.QueryIDf("%v_HORNER_B_%v_GLOBAL", pa.Name, index),
				sym.Sub(
					pa.HornerB[index],
					sym.Mul(
						sym.Sub(1, pa.FilterB[index]),
						column.Shift(pa.HornerB[index], 1),
					),
					sym.Mul(
						pa.FilterB[index],
						sym.Add(pa.ColumnB[index], sym.Mul(pa.EvalCoin[index], column.Shift(pa.HornerB[index], 1))),
					),
				),
			)

			comp.InsertLocal(
				round,
				ifaces.QueryIDf("%v_HORNER_B_LOCAL_END_%v", pa.Name, index),
				sym.Sub(
					column.Shift(pa.HornerB[index], -1),
					sym.Mul(fShifted, colShifted),
				),
			)

			pa.HornerB0[index] = comp.InsertLocalOpening(round, ifaces.QueryIDf("%v_HORNER_B0_%v", pa.Name, index), pa.HornerB[index])
		}
	default:
		fmt.Errorf("Invalid column name %v, should be either A or B", colName)
	}

}

func shiftExpression(comp *wizard.CompiledIOP, expr *symbolic.Expression, nbShift int) *symbolic.Expression {
	var (
		board          = expr.Board()
		metadata       = board.ListVariableMetadata()
		translationMap = collection.NewMapping[string, *sym.Expression]()
	)

	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:
			translationMap.InsertNew(string(t.GetColID()), ifaces.ColumnAsVariable(column.Shift(t, nbShift)))

		case coin.Info:

			if !comp.Coins.Exists(t.Name) {
				utils.Panic("Coin %v does not exist in the InitialComp", t.Name)
			}
			translationMap.InsertNew(t.String(), symbolic.NewVariable(t))
		}
	}
	return expr.Replay(translationMap)
}
