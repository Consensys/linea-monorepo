package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	LogDerivativeSum = "LOGDERIVATIVE_SUM"
)

// distributionInputs stores the input required for the distribution of a LogDerivativeSum query.
type distributionInputs struct {
	comp *wizard.CompiledIOP
	// module Discoverer used to detect the relevant part of the query to the module
	disc distributed.ModuleDiscoverer
	// Name of the module
	moduleName string
	// query is supposed to be the global LogDerivativeSum.
	queryID ifaces.QueryID
}

// DistributeLogDerivativeSum distribute the LogDerivativeSum among the modules.
// It detect the relevant share of the module from the global LogDerivativeSum.
// It generates a new LogDerivateSum query relevant to the module.
func DistributeLogDerivativeSum(in distributionInputs) {
	var (
		comp        = in.comp
		numerator   []*symbolic.Expression
		denominator []*symbolic.Expression
		zCatalog    map[[2]int]*query.LogDerivativeSumInput
		lastRound   = in.comp.NumRounds() - 1
	)
	// check that the given query is a valid LogDerivateSum query in the CompiledIOP.
	logDeriv, ok := comp.QueriesParams.Data(in.queryID).(query.LogDerivativeSum)
	if !ok {
		panic("the given query is not a valid LogDerivativeSum from the compiledIOP")
	}

	// This ensures that the logDerivative query is not used again in the
	// compilation process.
	comp.QueriesNoParams.MarkAsIgnored(in.queryID)

	// extract the share of the module from the global sum.
	for key := range logDeriv.Inputs {
		for i := range logDeriv.Inputs[key].Numerator {
			if in.disc.ExpressionIsInModule(logDeriv.Inputs[key].Numerator[i], in.moduleName) {
				numerator = append(numerator, logDeriv.Inputs[key].Numerator[i])
			}
			if in.disc.ExpressionIsInModule(logDeriv.Inputs[key].Denominator[i], in.moduleName) {
				denominator = append(denominator, logDeriv.Inputs[key].Denominator[i])
			}
		}

		// zCatalog specific to the module
		zCatalog[key] = &query.LogDerivativeSumInput{
			Round:       key[0],
			Size:        key[1],
			Numerator:   numerator,
			Denominator: denominator,
		}

	}

	// insert a  LogDerivativeSum specific to the module.
	comp.InsertLogDerivativeSum(
		lastRound,
		ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.moduleName),
		zCatalog,
	)
}
