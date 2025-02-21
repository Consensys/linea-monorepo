package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	segcomp "github.com/consensys/linea-monorepo/prover/protocol/distributed/segment_comp.go"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
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
	// number of segments for the module
	NumSegments, SegID int
}

// DistributeLogDerivativeSum distributes a  share from a global [query.LogDerivativeSum] query to the given module.
func DistributeLogDerivativeSum(moduleComp *wizard.CompiledIOP, segIn segcomp.SegmentInputs) {

	var (
		queryID ifaces.QueryID
	)

	for _, qName := range segIn.InitialComp.QueriesParams.AllUnignoredKeys() {

		_, ok := segIn.InitialComp.QueriesParams.Data(qName).(query.LogDerivativeSum)
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
		ModuleComp:  moduleComp,
		InitialComp: segIn.InitialComp,
		Disc:        segIn.Disc.SimpleDiscoverer,
		ModuleName:  segIn.ModuleName,
		Query:       segIn.InitialComp.QueriesParams.Data(queryID).(query.LogDerivativeSum),
		NumSegments: segIn.NumSegmentsInModule,
		SegID:       segIn.SegID,
	})

}

// GetShareOfLogDerivativeSum extracts the share of the given modules from the given LogDerivativeSum query.
// It inserts a new LogDerivativeSum for the extracted share.
func GetShareOfLogDerivativeSum(in DistributionInputs) {

	var (
		moduleComp    = in.ModuleComp
		keyIsInModule bool
		zCatalog      = make(map[int]*query.LogDerivativeSumInput)
		logDeriv      = in.Query
		round         = logDeriv.Round
		// create a translation map from the columns of moduleComp.
		// this does not include verifier columns.
		translationMap = collection.NewMapping[string, *symbolic.Expression]()
	)

	// extract the share of the module from the global sum.
	for size := range logDeriv.Inputs {

		var (
			numerator   []*symbolic.Expression
			denominator []*symbolic.Expression
		)

		for i := range logDeriv.Inputs[size].Numerator {

			// if Denominator is in the module pass the numerator from initialComp to moduleComp
			// Particularly, T might be in the module and needs to take M from initialComp.

			if in.Disc.ExpressionIsInModule(logDeriv.Inputs[size].Denominator[i], in.ModuleName) {

				if !in.Disc.ExpressionIsInModule(logDeriv.Inputs[size].Numerator[i], in.ModuleName) {

					utils.Panic("Denominator is in the module %v but not Numerator", in.ModuleName)
				}

				// update translationMap by adding local coins
				// the previous check guarantees that all the columns
				// from the expression  are in the module
				// Thus we can add the coins locally (i.e., without [distributed.ModuleDiscoverer]).
				distributed.ReplaceExternalCoins(in.InitialComp, moduleComp,
					logDeriv.Inputs[size].Denominator[i], translationMap, in.NumSegments, in.SegID)

				denExpr := distributed.AdjustExpressionForModule(
					in.InitialComp,
					in.ModuleComp,
					logDeriv.Inputs[size].Denominator[i].Replay(translationMap),
					in.NumSegments, in.SegID,
				)

				numExpr := distributed.AdjustExpressionForModule(
					in.InitialComp,
					in.ModuleComp,
					logDeriv.Inputs[size].Numerator[i],
					in.NumSegments, in.SegID,
				)

				denominator = append(denominator, denExpr)

				numerator = append(numerator, numExpr)

				keyIsInModule = true
			}

		}

		// if there in any expression relevant to the current key, add them to zCatalog
		if keyIsInModule {

			board := denominator[0].Board()
			// size of the expressions in the module
			sizeInModule := column.ExprIsOnSameLengthHandles(&board)

			// zCatalog specific to the module
			// due to vertical splitting size in module-segments may be different from size in the initialComp.
			zCatalog[sizeInModule] = &query.LogDerivativeSumInput{
				Size:        sizeInModule,
				Numerator:   numerator,
				Denominator: denominator,
			}
		}

		keyIsInModule = false
	}

	// insert a  LogDerivativeSum specific to the module at round 1 (since initialComp has 2 rounds).
	logDerivQuery := moduleComp.InsertLogDerivativeSum(
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

	// declare [query.LogDerivSumParams] as [wizard.PublicInput]
	moduleComp.PublicInputs = append(moduleComp.PublicInputs,
		wizard.PublicInput{
			Name: constants.LogDerivativeSumPublicInput,
			Acc:  accessors.NewLogDerivSumAccessor(logDerivQuery),
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

func createTranslationMap(comp *wizard.CompiledIOP) collection.Mapping[string, *symbolic.Expression] {

	var (
		exprMap = collection.NewMapping[string, *symbolic.Expression]()
		expr    *symbolic.Expression
	)

	for _, colID := range comp.Columns.AllKeys() {
		expr = ifaces.ColumnAsVariable(comp.Columns.GetHandle(colID))
		exprMap.InsertNew(string(colID), expr)
	}
	for _, coinID := range comp.Coins.AllKeys() {
		expr = symbolic.NewVariable(comp.Coins.Data(coinID))
		exprMap.InsertNew(string(coinID), expr)
	}
	return exprMap
}
