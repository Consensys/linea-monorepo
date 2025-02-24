package plonkinwizard

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// DistributePlonkInWizard scans initComp for non-ignored PlonkInWizard queries and
// identifies the ones corresponding to moduleName. For each of those queries, the
// function inserts a corresponding query into the module. The function will panic
// if:
//
//   - the query is cross-module
//
// The new query is analog to the one in initComp but and shares the same name.
func DistributePlonkInWizard(
	moduleName distributed.ModuleName,
	initComp, module *wizard.CompiledIOP,
	disc distributed.ModuleDiscoverer,
) {

	qNames := initComp.QueriesNoParams.AllKeys()

	for i := range qNames {

		// Skip if it was already compiled
		if initComp.QueriesNoParams.IsIgnored(qNames[i]) {
			continue
		}

		// Skip if this is not a PlonkInWizard query
		q, isPiw := initComp.QueriesNoParams.Data(qNames[i]).(*query.PlonkInWizard)
		if !isPiw {
			continue
		}

		var (
			dataInModule = disc.ColumnIsInModule(q.Data, moduleName)
			selInModule  = disc.ColumnIsInModule(q.Selector, moduleName)
			hasMask      = q.CircuitMask != nil
			maskInModule = hasMask && disc.ColumnIsInModule(q.CircuitMask, moduleName)
		)

		if !dataInModule {
			continue
		}

		if dataInModule != selInModule {
			utils.Panic("data and selector must be in the same module")
		}

		if hasMask && dataInModule != maskInModule {
			utils.Panic("data and mask must be in the same module")
		}

		distributedQuery := &query.PlonkInWizard{
			ID:           q.ID,
			Data:         module.Columns.GetHandle(q.Data.GetColID()),
			Selector:     module.Columns.GetHandle(q.Selector.GetColID()),
			CircuitMask:  module.Columns.GetHandle(q.CircuitMask.GetColID()),
			Circuit:      q.Circuit,
			PlonkOptions: q.PlonkOptions,
		}

		module.InsertPlonkInWizard(distributedQuery)
	}
}
