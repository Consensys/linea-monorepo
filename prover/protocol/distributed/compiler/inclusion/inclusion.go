package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
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

// DistributeLogDerivativeSum distributes a  share from a global [query.LogDerivativeSum] query to the given module.
func DistributeLogDerivativeSum(
	initialComp, moduleComp *wizard.CompiledIOP,
	moduleName distributed.ModuleName,
	disc distributed.ModuleDiscoverer,
	initialProver *wizard.ProverRuntime) {

	var (
		queryID ifaces.QueryID
	)
	for _, qName := range initialComp.QueriesParams.AllUnignoredKeys() {

		_, ok := initialComp.QueriesParams.Data(qName).(query.LogDerivativeSum)
		if !ok {
			continue
		}
		// panic if there is more than a LogDerivativeSum query in the initialComp.
		if string(queryID) != "" {
			utils.Panic("found more than a LogDerivativeSum query in the initialComp")
		}

		queryID = qName
	}
	input := DistributionInputs{
		ModuleComp:    moduleComp,
		InitialComp:   initialComp,
		Disc:          disc,
		ModuleName:    moduleName,
		QueryID:       queryID,
		InitialProver: initialProver,
	}
	// get the share of the module from the LogDerivativeSum query
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

	// @Azam this is because for the moment we dont know how the module-discoverer extracts moduleComp from InitialComp.
	// if we are sure that inclusions are already removed from modComp, we can skip this step here.
	for _, qName := range moduleComp.QueriesNoParams.AllUnignoredKeys() {
		// Filter out non lookup queries
		_, ok := moduleComp.QueriesNoParams.Data(qName).(query.Inclusion)
		if !ok {
			continue
		}
		// ignore the query as it is about to be compiled and replaces with low level queries.
		moduleComp.QueriesNoParams.MarkAsIgnored(qName)
	}

	// extract the share of the module from the global sum.
	for key := range logDeriv.Inputs {
		for i := range logDeriv.Inputs[key].Numerator {
			// if Denominator is in the module pass the numerator from initialComp to moduleComp
			// Particularly, T might be in the module and needs to take M from initialComp.
			if in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Denominator[i], in.ModuleName) {
				if !in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Numerator[i], in.ModuleName) {
					distributed.PassColumnToModule(initialComp, moduleComp, initialProver, logDeriv.Inputs[key].Numerator[i])
				}
				denominator = append(denominator, logDeriv.Inputs[key].Denominator[i])
				numerator = append(numerator, logDeriv.Inputs[key].Numerator[i])
				// replaces the external coins with local coins
				// note that they just appear in the denominator.
				distributed.ReplaceExternalCoins(initialComp, moduleComp, logDeriv.Inputs[key].Denominator[i])
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
	// sanity check; the initialComp has only two rounds
	if initialComp.NumRounds() != 2 {
		utils.Panic("expected initialComp to have 2 rounds but it has %v rounds", initialComp.NumRounds())
	}

	// insert a  LogDerivativeSum specific to the module at round 1 (since initialComp has 2 rounds).
	moduleComp.InsertLogDerivativeSum(
		1,
		ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
		zCatalog,
	)
	// prover step to assign the parameters of LogDerivativeSum at the same round.
	moduleComp.SubProvers.AppendToInner(1, func(run *wizard.ProverRuntime) {
		run.AssignLogDerivSum(
			ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
			GetLogDerivativeSumResult(zCatalog, run),
		)
	})

}

// GetLogDerivativeSumResult is a helper allowing the prover to calculate the result of its associated LogDerivativeSum query.
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
