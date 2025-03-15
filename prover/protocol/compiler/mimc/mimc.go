package mimc

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// CompileMiMC compiles the MiMC queries by checking each query individually
func CompileMiMC(comp *wizard.CompiledIOP) {
	// Collect all MiMC queries
	mimcQueries := []query.MiMC{}

	for _, id := range comp.QueriesNoParams.AllUnignoredKeys() {
		// Fetch the query
		q := comp.QueriesNoParams.Data(id)
		qMiMC, ok := q.(query.MiMC)
		if !ok {
			// Not a MiMC query, skip it
			continue
		}

		// Mark it as ignored
		comp.QueriesNoParams.MarkAsIgnored(id)
		mimcQueries = append(mimcQueries, qMiMC)
	}

	if len(mimcQueries) == 0 {
		logrus.Debug("MiMC compiler exited: no MiMC queries to compile")
		return
	}

	// Apply manual check to each query individually
	for _, q := range mimcQueries {
		logrus.Debugf("MiMC compiler: checking query %v individually", q.ID)
		manualCheckMiMCBlock(comp, q.Blocks, q.OldState, q.NewState)
	}
}
