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
}

// GetShareOfLogDerivativeSum extracts the share of the given modules from the given LogDerivativeSum query.
// It inserts a new LogDerivativeSum for the extracted share.
func GetShareOfLogDerivativeSum(in DistributionInputs) {
	var (
		initialComp = in.InitialComp
		moduleComp  = in.ModuleComp
		numerator   []*symbolic.Expression
		denominator []*symbolic.Expression
		zCatalog    = make(map[[2]int]*query.LogDerivativeSumInput)
		lastRound   = in.InitialComp.NumRounds() - 1
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
			if in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Numerator[i], in.ModuleName) {
				if in.Disc.ExpressionIsInModule(logDeriv.Inputs[key].Denominator[i], in.ModuleName) {
					denominator = append(denominator, logDeriv.Inputs[key].Denominator[i])
					numerator = append(numerator, logDeriv.Inputs[key].Numerator[i])
				}
			}
		}

		// if there in any numerator associated with the current key add it to the map.
		if len(numerator) != 0 {
			// zCatalog specific to the module
			zCatalog[key] = &query.LogDerivativeSumInput{
				Round:       key[0],
				Size:        key[1],
				Numerator:   numerator,
				Denominator: denominator,
			}
		}

	}

	// insert a  LogDerivativeSum specific to the module.
	moduleComp.InsertLogDerivativeSum(
		lastRound,
		ifaces.QueryIDf("%v_%v", LogDerivativeSum, in.ModuleName),
		zCatalog,
	)
}

// DistributeLogDerivativeSum extract the LogDerivativeSum query that is subject to the distribution.
// It ignores the inclusion queries in the module compiledIOP and replaces them with its share of LogDerivativeSum.
func DistributeLogDerivativeSum(initialComp, moduleComp *wizard.CompiledIOP, moduleName distributed.ModuleName, disc distributed.ModuleDiscoverer) {

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
		ModuleComp:  moduleComp,
		InitialComp: initialComp,
		Disc:        disc,
		ModuleName:  moduleName,
		QueryID:     queryID,
	}
	GetShareOfLogDerivativeSum(input)

}
