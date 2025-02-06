package distributedprojection

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/sirupsen/logrus"
)

type distribuedProjectionProverAction struct {
	Name               ifaces.QueryID
	FilterA, FilterB   []*sym.Expression
	ColumnA, ColumnB   []*sym.Expression
	HornerA, HornerB   []ifaces.Column
	HornerA0, HornerB0 []query.LocalOpening
	EvalCoin           []coin.Info
	IsA, IsB           []bool
}

// Run executes the distributed projection prover action.
// It iterates over the input filters, columns, and evaluation coins.
// Depending on the values of IsA and IsB, it computes the Horner traces for columns A and B,
// and assigns them to the corresponding columns and local points in the prover runtime.
// If both IsA and IsB are true, it computes the Horner traces for both columns A and B.
// If IsA is true and IsB is false, it computes the Horner trace for column A only.
// If IsA is false and IsB is true, it computes the Horner trace for column B only.
// If neither IsA nor IsB is true, it panics with an error message indicating an invalid prover assignment.
func (pa *distribuedProjectionProverAction) Run(run *wizard.ProverRuntime) {
    for index := range pa.FilterA {
        if pa.IsA[index] && pa.IsB[index] {
            var (
                colA    = column.EvalExprColumn(run, pa.ColumnA[index].Board()).IntoRegVecSaveAlloc()
                fA      = column.EvalExprColumn(run, pa.FilterA[index].Board()).IntoRegVecSaveAlloc()
                colB    = column.EvalExprColumn(run, pa.ColumnB[index].Board()).IntoRegVecSaveAlloc()
                fB      = column.EvalExprColumn(run, pa.FilterB[index].Board()).IntoRegVecSaveAlloc()
                x       = run.GetRandomCoinField(pa.EvalCoin[index].Name)
                hornerA = poly.GetHornerTrace(colA, fA, x)
                hornerB = poly.GetHornerTrace(colB, fB, x)
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
                hornerA = poly.GetHornerTrace(colA, fA, x)
            )
            run.AssignColumn(pa.HornerA[index].GetColID(), smartvectors.NewRegular(hornerA))
            run.AssignLocalPoint(pa.HornerA0[index].ID, hornerA[0])
        } else if !pa.IsA[index] && pa.IsB[index] {
            var (
                colB    = column.EvalExprColumn(run, pa.ColumnB[index].Board()).IntoRegVecSaveAlloc()
                fB      = column.EvalExprColumn(run, pa.FilterB[index].Board()).IntoRegVecSaveAlloc()
                x       = run.GetRandomCoinField(pa.EvalCoin[index].Name)
                hornerB = poly.GetHornerTrace(colB, fB, x)
            )
            run.AssignColumn(pa.HornerB[index].GetColID(), smartvectors.NewRegular(hornerB))
            run.AssignLocalPoint(pa.HornerB0[index].ID, hornerB[0])
        } else {
            utils.Panic("Invalid prover assignment in distributed projection id: %v", pa.Name)
        }
    }
}

// Push populates the distribuedProjectionProverAction with data from the provided DistributedProjection query.
// It processes each input in the query and assigns the corresponding values to the prover action's fields
// based on whether the input is in module A, module B, or both.
//
// Parameters:
//   - comp: A pointer to the CompiledIOP, used to access coin data.
//   - distributedprojection: The DistributedProjection query containing the inputs to be processed.
//
// The function does not return any value, but updates the fields of the distribuedProjectionProverAction in-place.
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

// RegisterQueries registers the necessary queries for the distributed projection prover action.
// It processes each input in the distributed projection, shifting expressions and registering
// queries for columns A and B based on their presence in the respective modules.
//
// Parameters:
//   - comp: A pointer to the CompiledIOP, used for inserting commits and queries.
//   - round: An integer representing the current round of the protocol.
//   - distributedprojection: A DistributedProjection query containing the inputs to be processed.
//
// The function does not return any value, but updates the internal state of the prover action
// by registering the required queries for each input.
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
            utils.Panic("Invalid prover action case for the distributed projection query %v", pa.Name)
        }
    }
}

// registerForCol registers queries for a specific column (A or B) in the distributed projection prover action.
// It inserts commits, global queries, local queries, and local openings for the Horner polynomial evaluation.
//
// Parameters:
//   - comp: A pointer to the CompiledIOP, used for inserting commits and queries.
//   - fShifted: A shifted filter expression.
//   - colShifted: A shifted column expression.
//   - input: A pointer to the DistributedProjectionInput containing size information.
//   - colName: A string indicating which column to register ("A" or "B").
//   - round: An integer representing the current round of the protocol.
//   - index: An integer used to uniquely identify the registered queries.
//
// The function doesn't return any value but updates the internal state of the prover action
// by registering the required queries and commits for the specified column.
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
        utils.Panic("Invalid column name %v, should be either A or B", colName)
    }

}

// shiftExpression shifts a column which is a symbolic expression by a specified number of positions.
// It creates a new expression with the shifted column while maintaining the structure of the original expression.
//
// Parameters:
//   - comp: A pointer to the CompiledIOP, used to check for the existence of coins.
//   - expr: The original symbolic expression to be shifted.
//   - nbShift: The number of positions to shift the column. Positive values shift forward, negative values shift backward.
//
// Returns:
//   A new *sym.Expression with the column shifted according to the specified nbShift.
func shiftExpression(comp *wizard.CompiledIOP, expr *sym.Expression, nbShift int) *sym.Expression {
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
            translationMap.InsertNew(t.String(), sym.NewVariable(t))
        default:
            utils.Panic("Unsupported type for shift expression operation")
        }
    }
    return expr.Replay(translationMap)
}
