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
	Query query.LogDerivativeSum
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

	// get the share of the module from the LogDerivativeSum query
	GetShareOfLogDerivativeSum(DistributionInputs{
		ModuleComp:    moduleComp,
		InitialComp:   initialComp,
		Disc:          disc,
		ModuleName:    moduleName,
		Query:         initialComp.QueriesParams.Data(queryID).(query.LogDerivativeSum),
		InitialProver: initialProver,
	})

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
		zCatalog      = make(map[int]*query.LogDerivativeSumInput)
		logDeriv      = in.Query
		round         = logDeriv.Round
	)

	// extract the share of the module from the global sum.
	for size := range logDeriv.Inputs {

		for i := range logDeriv.Inputs[size].Numerator {

			// if Denominator is in the module pass the numerator from initialComp to moduleComp
			// Particularly, T might be in the module and needs to take M from initialComp.
			if in.Disc.ExpressionIsInModule(logDeriv.Inputs[size].Denominator[i], in.ModuleName) {

				if !in.Disc.ExpressionIsInModule(logDeriv.Inputs[size].Numerator[i], in.ModuleName) {
					distributed.PassColumnToModule(initialComp, moduleComp, initialProver, logDeriv.Inputs[size].Numerator[i])
				}

				denominator = append(denominator, logDeriv.Inputs[size].Denominator[i])
				numerator = append(numerator, logDeriv.Inputs[size].Numerator[i])

				// replaces the external coins with local coins
				// note that they just appear in the denominator.
				distributed.ReplaceExternalCoins(initialComp, moduleComp, logDeriv.Inputs[size].Denominator[i])
				keyIsInModule = true
			}
		}

		// if there in any expression relevant to the current key, add them to zCatalog
		if keyIsInModule {
			// zCatalog specific to the module
			zCatalog[size] = &query.LogDerivativeSumInput{
				Size:        size,
				Numerator:   numerator,
				Denominator: denominator,
			}
		}

		keyIsInModule = false
	}

	// insert a  LogDerivativeSum specific to the module at round 1 (since initialComp has 2 rounds).
	moduleComp.InsertLogDerivativeSum(
		round,
		ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
		zCatalog,
	)

	// prover step to assign the parameters of LogDerivativeSum at the same round.
	moduleComp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		run.AssignLogDerivSum(
			ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
			getLogDerivativeSumResult(zCatalog, run),
		)
	})

}

// getLogDerivativeSumResult is a helper allowing the prover to calculate the result of its associated LogDerivativeSum query.
func getLogDerivativeSumResult(zCatalog map[int]*query.LogDerivativeSumInput, run *wizard.ProverRuntime) field.Element {
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
