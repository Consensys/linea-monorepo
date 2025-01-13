package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ReplaceExternalCoins replaces the external coins with local coins, for a given expression.
// It does not check if all the columns from the expression are in the module.
// If this is required should be check before calling ReplaceExternalCoins.
// If the Coin does not exist in the initialComp it panics.
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
				moduleComp.InsertCoin(v.Round, v.Name, coin.Field)
			}
		}
	}
}

// GetFreshModuleComp returns a [wizard.DefineFunc] that creates
// a [wizard.CompiledIOP] object including only the columns
// relevant to the module. It also contains the prover steps
// for assigning the module column
func GetFreshModuleComp(
	initialComp *wizard.CompiledIOP,
	disc ModuleDiscoverer,
	moduleName ModuleName,
) *wizard.CompiledIOP {

	var (
		// initialize the moduleComp
		moduleComp = wizard.NewCompiledIOP()
	)

	for round := 0; round < initialComp.NumRounds(); round++ {
		var columnsInRound []ifaces.Column
		// get the columns per round
		for _, colName := range initialComp.Columns.AllKeysAt(round) {

			col := initialComp.Columns.GetHandle(colName)
			if !disc.ColumnIsInModule(col, moduleName) {
				continue
			}

			moduleComp.InsertCommit(col.Round(), col.GetColID(), col.Size())
			columnsInRound = append(columnsInRound, col)
		}

		// create a new  moduleProver
		moduleProver := moduleProver{
			cols:  columnsInRound,
			round: round,
		}

		// register Prover action for the module to assign columns per round
		moduleComp.RegisterProverAction(round, moduleProver)
	}

	return moduleComp
}

// it stores the input for the module prover
type moduleProver struct {
	round int
	// columns for a specific round
	cols []ifaces.Column
}

// It implements [wizard.ProverAction] for the module prover.
func (p moduleProver) Run(run *wizard.ProverRuntime) {

	if run.ParentRuntime == nil {
		utils.Panic("invalid call: the runtime does not have a [ParentRuntime]")
	}

	for _, col := range p.cols {
		// get the witness from the initialProver
		colWitness := run.ParentRuntime.GetColumn(col.GetColID())
		// assign it in the module in the round col was declared
		run.AssignColumn(col.GetColID(), colWitness, col.Round())
	}
}
