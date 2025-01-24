package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// ReplaceExternalCoinsVerifCols replaces the external coins/verifiercols with local ones, for a given expression.
// It does not check if all the columns from the expression are in the module.
// If this is required should be check before calling ReplaceExternalCoins.
// If the Coin does not exist in the initialComp it panics.
// It adds the local coins/verifiercols to the translationMap.
// note that verifiercols are not captured by the [wizard.CompiledIOP] and we have to handle them on the fly.
func ReplaceExternalCoinsVerifCols(
	initialComp, moduleComp *wizard.CompiledIOP,
	expr *symbolic.Expression,
	translationMap collection.Mapping[string, *symbolic.Expression],
	numSegments int,
) {
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

				if !initialComp.Coins.Exists("SEED") {
					utils.Panic("Expect to find a seed in the initialComp")
				}
				// register a local coin of type FieldFromSeed.
				name := coin.Namef("%v_%v", v.Name, "FieldFromSeed")
				localV := moduleComp.InsertCoin(v.Round, name, coin.FieldFromSeed)
				translationMap.InsertNew(v.String(), symbolic.NewVariable(localV))
			}
		case ifaces.Column:
			// create the local verfiercols and add them to the translationMap.
			if vCol, ok := v.(verifiercol.VerifierCol); ok {
				if constCol, ok := vCol.(verifiercol.ConstCol); ok {
					verifcol := verifiercol.NewConstantCol(constCol.F,
						constCol.Size_/numSegments)
					translationMap.InsertNew(v.String(), ifaces.ColumnAsVariable(verifcol))
				} else {
					panic("this case is not supported for now")
				}
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
