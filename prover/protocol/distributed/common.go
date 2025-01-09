package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// For an expression that has all its columns in the module, it replaces the external coins with local coins
func ReplaceExternalCoins(initialComp, moduleComp *wizard.CompiledIOP, expr *symbolic.Expression) {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
	)
	for _, m := range metadata {
		switch v := m.(type) {
		case coin.Info:

			if !initialComp.Coins.Exists(v.Name) {
				utils.Panic("Coin %v does not exist in the InitialComp", v.Name)
			}
			if v.Round != 1 {
				utils.Panic("Coin %v is declared in round %v != 1", v.Name, v.Round)
			}
			if !moduleComp.Coins.Exists(v.Name) {
				moduleComp.InsertCoin(1, v.Name, coin.Field)
			}
		}
	}
}

// PassColumnToModule passes the column, underlying the expression, from initialComp to moduleComp.
// It also handles the prover steps to assign the passed column in the moduleColumn.
func PassColumnToModule(
	initComp, moduleComp *wizard.CompiledIOP,
	initialProver *wizard.ProverRuntime,
	expr *symbolic.Expression) {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
	)
	// check that the expression is a single column
	if len(metadata) != 1 {
		utils.Panic("expected a single metadata")
	}
	for _, m := range metadata {
		switch v := m.(type) {
		case ifaces.Column:
			// check that the column is in the initComp
			if !initComp.Columns.Exists(v.GetColID()) {
				utils.Panic("Expected to find column %v in the initialComp", v.GetColID())
			}

			// commit to the column in the moduleComp
			moduleComp.InsertCommit(0, v.GetColID(), v.Size())

			// assign the column in the moduleComp
			moduleComp.SubProvers.AppendToInner(v.Round(), func(run *wizard.ProverRuntime) {
				run.AssignColumn(v.GetColID(), initialProver.GetColumn(v.GetColID()))

			})

		default:
			utils.Panic("expected only column type in the expression")
		}

	}
}
