package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	LogDerivativeSum = "LOGDERIVATIVE_SUM"
)

// DistributionInputs stores the input required for the distribution of a LogDerivativeSum query.
type DistributionInputs struct {
	ModuleComp  *wizard.CompiledIOP
	InitialComp *wizard.CompiledIOP
	// module Discoverer used to detect the relevant part of the query to the module
	Disc distributed.ModuleDiscoverer
	// Name of the module
	ModuleName distributed.ModuleName
	// query is supposed to be the global LogDerivativeSum.
	QueryID ifaces.QueryID
	// it contains the whole witness,
	// and also the witness for the auxiliary columns such as multiplicity column for Inclusion.
	InitialProver *wizard.ProverRuntime
}

// DistributeLogDerivativeSum extract the LogDerivativeSum query that is subject to the distribution.
// It ignores the inclusion queries in the module compiledIOP and replaces them with its share of LogDerivativeSum.
func DistributeLogDerivativeSum(
	initialComp, moduleComp *wizard.CompiledIOP,
	moduleName distributed.ModuleName,
	disc distributed.ModuleDiscoverer,
	initialProver *wizard.ProverRuntime) {

	var queryID ifaces.QueryID
	for _, qName := range initialComp.QueriesParams.AllUnignoredKeys() {
		_, ok := initialComp.QueriesParams.Data(qName).(query.LogDerivativeSum)
		if !ok {
			continue
		}
		queryID = qName
		//@Azam panic if it has more than one.
		// it breaks since we expect only a single query of this type.
		break
	}
	input := DistributionInputs{
		ModuleComp:    moduleComp,
		InitialComp:   initialComp,
		Disc:          disc,
		ModuleName:    moduleName,
		QueryID:       queryID,
		InitialProver: initialProver,
	}
	GetShareOfLogDerivativeSum(input)

}

// GetShareOfLogDerivativeSum extracts the share of the given modules from the given LogDerivativeSum query.
// It inserts a new LogDerivativeSum for the extracted share.
func GetShareOfLogDerivativeSum(in DistributionInputs) {
	var (
		initialComp   = in.InitialComp
		moduleComp    = in.ModuleComp
		initialProver = in.InitialProver
		numerator     []*symbolic.Expression
		denominator   []*symbolic.Expression
		keyIsInModule bool
		zCatalog      = make(map[[2]int]*query.LogDerivativeSumInput)
	)
	// check that the given query is a valid LogDerivateSum query in the CompiledIOP.
	logDeriv, ok := initialComp.QueriesParams.Data(in.QueryID).(query.LogDerivativeSum)
	if !ok {
		panic("the given query is not a valid LogDerivativeSum from the compiledIOP")
	}

	// This ensures that the logDerivative query is not used again in the
	// compilation process for the module.
	/*	_, ok = moduleComp.QueriesParams.Data(in.QueryID).(query.LogDerivativeSum)
		if ok {
			moduleComp.QueriesNoParams.MarkAsIgnored(in.QueryID)
		} */

	// also mark all the inclusion queries in the module as ignored
	// @Azam this is because for the moment we dont know how the module-discoverer extracts moduleComp from InitialComp.
	// if we are sure that inclusions are already removed from modComp, we can skip this step here.
	for _, qName := range moduleComp.QueriesNoParams.AllUnignoredKeys() {
		// Filter out non lookup queries
		_, ok := moduleComp.QueriesNoParams.Data(qName).(query.Inclusion)
		if !ok {
			continue
		}
		moduleComp.QueriesNoParams.MarkAsIgnored(qName)

	}

	// extract the share of the module from the global sum.
	for key := range logDeriv.Inputs {
		for i := range logDeriv.Inputs[key].Numerator {
			// if Denominator is in the module pass the numerator from initialComp to moduleComp
			// Particularly, T might be in the module and needs to take M from initialComp.
			if in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Denominator[i], in.ModuleName) {
				if !in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Numerator[i], in.ModuleName) {
					PassColumnToModule(initialComp, moduleComp, initialProver, logDeriv.Inputs[key].Numerator[i])
				}
				denominator = append(denominator, logDeriv.Inputs[key].Denominator[i])
				numerator = append(numerator, logDeriv.Inputs[key].Numerator[i])
				// replaces the external coins with local coins
				// they just appear in the denominator.
				ReplaceExternalCoins(initialComp, moduleComp, logDeriv.Inputs[key].Denominator[i])
				keyIsInModule = true
			}
		}
		// if there in any expression relevant to the current key, add them to zCatalog
		if keyIsInModule {
			// zCatalog specific to the module
			zCatalog[key] = &query.LogDerivativeSumInput{
				Round:       key[0],
				Size:        key[1],
				Numerator:   numerator,
				Denominator: denominator,
			}
		}
		keyIsInModule = false

	}

	// insert a  LogDerivativeSum specific to the module.
	moduleComp.InsertLogDerivativeSum(
		1,
		ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
		zCatalog,
	)
	// prover step to assign the parameters of LogDerivativeSum
	moduleComp.SubProvers.AppendToInner(1, func(run *wizard.ProverRuntime) {
		run.AssignLogDerivSum(
			ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
			GetLogDerivativeSumResult(zCatalog, run),
		)
	})

}

// GetLogDerivativeSumResult allows the prover to calculate the result of its associated LogDerivativeSum query.
func GetLogDerivativeSumResult(zCatalog map[[2]int]*query.LogDerivativeSumInput, run *wizard.ProverRuntime) field.Element {
	// compute the actual sum from the Numerator and Denominator
	actualSum := field.Zero()
	for key := range zCatalog {
		for i, num := range zCatalog[key].Numerator {

			var (
				numBoard          = num.Board()
				denBoard          = zCatalog[key].Denominator[i].Board()
				numeratorMetadata = numBoard.ListVariableMetadata()
				denominator       = column.EvalExprColumn(run, denBoard).IntoRegVecSaveAlloc()
				numerator         []field.Element
				packedZ           = field.BatchInvert(denominator)
			)

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), zCatalog[key].Size)
			}

			if len(numeratorMetadata) > 0 {
				numerator = column.EvalExprColumn(run, numBoard).IntoRegVecSaveAlloc()
			}

			for k := range packedZ {
				packedZ[k].Mul(&numerator[k], &packedZ[k])
				if k > 0 {
					packedZ[k].Add(&packedZ[k], &packedZ[k-1])
				}
			}
			actualSum.Add(&actualSum, &packedZ[len(packedZ)-1])
		}
	}
	return actualSum
}

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
