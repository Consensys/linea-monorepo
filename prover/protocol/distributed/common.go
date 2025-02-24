package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// ReplaceExternalCoins replaces the external coins/verifiercols with local ones, for a given expression.
// It does not check if all the columns from the expression are in the module.
// If this is required should be check before calling ReplaceExternalCoins.
// If the Coin does not exist in the initialComp it panics.
// It adds the local coins/verifiercols to the translationMap.
// note that verifiercols are not captured by the [wizard.CompiledIOP] and we have to handle them on the fly.
func ReplaceExternalCoins(
	initialComp, moduleComp *wizard.CompiledIOP,
	expr *symbolic.Expression,
	translationMap collection.Mapping[string, *symbolic.Expression],
	numSegments, segID int,
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
				// add it to the translation map
				translationMap.InsertNew(v.String(), symbolic.NewVariable(localV))
			}
		default:
			translationMap.InsertNew(v.String(), symbolic.NewVariable(v))
		}
	}
}

// adjust the expression w.r.t the columns in the segment.
func AdjustExpressionForModule(
	initComp, comp *wizard.CompiledIOP,
	expr *symbolic.Expression,
	numSegments, segID int,
) *symbolic.Expression {

	if len(ListColumnsFromExpr(expr, true)) == 0 {
		return expr
	}

	var (
		board          = expr.Board()
		metadatas      = board.ListVariableMetadata()
		translationMap = collection.NewMapping[string, *symbolic.Expression]()
		colTranslation ifaces.Column
		size           = column.ExprIsOnSameLengthHandles(&board)
		segSize        = size / numSegments
	)

	for _, metadata := range metadatas {

		// For each slot, get the expression obtained by replacing the commitment
		// by the appropriated column.

		switch m := metadata.(type) {
		case ifaces.Column:

			colTranslation = ColInModule(initComp, comp, m, segID, segSize, false)
			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(colTranslation))

		case variables.X:
			utils.Panic("unsupported, the value of `x` in the unsplit query and the split would be different")
		case variables.PeriodicSample:
			// Check that the period is not larger than the domain size. If
			// the period is smaller this is a no-op because the period does
			// not change.
			if m.T > segSize {

				panic("unsupported")
			}
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))

		default:
			// Repass the same variable (for coins or other types of single-valued variable)
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}

	}
	return expr.Replay(translationMap)
}

// for the given column it return its counterpart in the module.
// if getNatural is true it returns the natural version of the column , otherwise it applies the proper shifting.
func ColInModule(initComp, moduleComp *wizard.CompiledIOP, col ifaces.Column, segID, segSize int, getNatural bool) ifaces.Column {

	switch v := col.(type) {
	case column.Shifted:

		parent := ColInModule(initComp, moduleComp, v.Parent, segID, segSize, getNatural)
		if getNatural {
			return parent
		}
		return column.Shift(parent, v.Offset)

	case verifiercol.VerifierCol:
		return v.Split(initComp, segID*segSize, (segID+1)*segSize)

	case column.Natural:
		return moduleComp.Columns.GetHandle(col.GetColID())

	default:
		panic("unsupported")
	}

}

// @Azam should be removed later and replaced with GetFreshGLComp() or GetFreshLPPComp()
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

// ListColumnsFromExpr returns the natural version of all the columns in the expression.
// if natural is true, it return the natural version of the columns,
// otherwise it return the original columns.
func ListColumnsFromExpr(expr *symbolic.Expression, natural bool) []ifaces.Column {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		colList  = []ifaces.Column{}
	)

	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:

			if shifted, ok := t.(column.Shifted); ok && natural {
				colList = append(colList, shifted.Parent)
			} else {
				colList = append(colList, t)
			}

		}
	}
	return colList

}

// it checks if the column is a verifier column or a shifted verifier column.
func IsVerifierColumn(col ifaces.Column) bool {

	if _, ok := col.(verifiercol.VerifierCol); ok {
		return true
	}

	if shifted, ok := col.(column.Shifted); ok {
		parent := shifted.Parent
		if _, ok := parent.(verifiercol.VerifierCol); ok {
			return true
		}
	}
	return false
}
