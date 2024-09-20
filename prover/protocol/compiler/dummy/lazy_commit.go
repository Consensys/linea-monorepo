package dummy

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	LAZY_COMMIT_SIZE   int    = 2
	LAZY_COMMIT_PREFIX string = "LAZY_COMMITMENT"
)

/*
Converts all the oracle commmitments into messages
and ask the verifier to manually verify all queries.

Primary use-case is testing.
*/
func LazyCommit(comp *wizard.CompiledIOP) {

	/*
		Registers all declared commitments and query parameters
		as messages in the same round. This steps is only relevant
		for the compiledIOP. We elaborate on how to update the provers
		and verifiers to account for this. Additionally, we take the queries
		as we compile them from the `CompiledIOP`.
	*/
	numRounds := comp.NumRounds()

	/*
		The filter returns true, as long as the query has not been marked as
		already compiled. This is to avoid them being compiled a second time.
	*/
	queriesParamsToCompile := comp.QueriesParams.AllUnignoredKeys()
	queriesNoParamsToCompile := comp.QueriesNoParams.AllUnignoredKeys()
	comsAllRounds := [][]ifaces.ColID{}

	for roundID := 0; roundID < numRounds; roundID++ {
		/*
			The commitments
		*/
		comsAllRounds = append(comsAllRounds, comp.Columns.AllKeysCommittedAt(roundID))
		for _, com := range comsAllRounds[roundID] {
			comp.InsertProof(roundID, renameComToLazyMsg(com), LAZY_COMMIT_SIZE)
		}
	}

	/*
		And mark the queries as already compiled
	*/
	for _, q := range queriesNoParamsToCompile {
		comp.QueriesNoParams.MarkAsIgnored(q)
	}

	for _, q := range queriesParamsToCompile {
		comp.QueriesParams.MarkAsIgnored(q)
	}

	/*
		Make the prover send all commitment he made as message.
		Is done round by round.
	*/
	for roundID_ := 0; roundID_ < numRounds; roundID_++ {

		/*
			Capture the roundID in a separate variable, to capture it in the
			closure. If we don't do that, all steps are going to run the last round
			because the `roundID_` is captured but not its value.
		*/
		roundID := roundID_
		prover := func(run *wizard.ProverRuntime) {

			/*
				The commitments
			*/
			for _, com := range comsAllRounds[roundID] {
				run.AssignColumn(renameComToLazyMsg(com), smartvectors.AllocateRegular(LAZY_COMMIT_SIZE))
			}
		}

		/*
			And append each to the result. This functions will later be
			called during the `Prover`
		*/
		comp.SubProvers.AppendToInner(roundID, prover)
	}

}

func renameComToLazyMsg(name ifaces.ColID) ifaces.ColID {
	msgName := fmt.Sprintf("%v_%v", LAZY_COMMIT_PREFIX, name)
	return ifaces.ColID(msgName)
}
