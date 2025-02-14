package plonkinwizard

import (
	plonk "github.com/consensys/linea-monorepo/prover/proto/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

func Compile(comp *wizard.CompiledIOP) {

	qNames := comp.QueriesNoParams.AllKeys()
	for i := range qNames {

		// Skip if it was already compiled
		if comp.QueriesNoParams.IsIgnored(qNames[i]) {
			continue
		}

		q, isType := comp.QueriesNoParams.Data(qNames[i]).(*query.PlonkInWizard)

		// Skip if this is not a PlonkInWizard query
		if !isType {
			continue
		}

		compileQuery(comp, q)
	}
}

func compileQuery(comp *wizard.CompiledIOP, q *query.PlonkInWizard) {

	var (
		round          = max(q.Data.Round(), q.Selector.Round())
		nbPublic, _, _ = gnarkutil.CountVariables(q.Circuit)
		nbPublicPadded = utils.NextPowerOfTwo(nbPublic)
		maxNbInstances = q.Data.Size() / nbPublicPadded
	)

	ctx := plonk.PlonkCheck(comp, q.Name, round, q.Circuit, maxNbInstances)

}
