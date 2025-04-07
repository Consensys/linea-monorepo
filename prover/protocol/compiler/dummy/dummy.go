package dummy

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// dummyVerifierAction implements the VerifierAction interface for the dummy compiler.
type dummyVerifierAction struct {
	comp                     *wizard.CompiledIOP
	queriesParamsToCompile   []ifaces.QueryID
	queriesNoParamsToCompile []ifaces.QueryID
}

// Run executes the verifier checks, replicating the original verifier closure logic.
func (d *dummyVerifierAction) Run(run *wizard.VerifierRuntime) error {
	logrus.Infof("started to run the dummy verifier")

	var finalErr error
	lock := sync.Mutex{}

	// Test all queries with parameters
	parallel.Execute(len(d.queriesParamsToCompile), func(start, stop int) {
		for i := start; i < stop; i++ {
			name := d.queriesParamsToCompile[i]
			lock.Lock()
			q := d.comp.QueriesParams.Data(name)
			lock.Unlock()
			if err := q.Check(run); err != nil {
				lock.Lock()
				finalErr = fmt.Errorf("%v\nfailed %v - %v", finalErr, name, err)
				lock.Unlock()
				logrus.Debugf("query %v failed\n", name)
			} else {
				logrus.Debugf("query %v passed\n", name)
			}
		}
	})

	// Test the queries without parameters
	parallel.Execute(len(d.queriesNoParamsToCompile), func(start, stop int) {
		for i := start; i < stop; i++ {
			name := d.queriesNoParamsToCompile[i]
			lock.Lock()
			q := d.comp.QueriesNoParams.Data(name)
			lock.Unlock()
			if err := q.Check(run); err != nil {
				lock.Lock()
				finalErr = fmt.Errorf("%v\nfailed %v - %v", finalErr, name, err)
				lock.Unlock()
			} else {
				logrus.Debugf("query %v passed\n", name)
			}
		}
	})

	// Nil to indicate all checks passed
	return finalErr
}

// RunGnark is a no-op, matching the original no-op gnark verifier.
func (d *dummyVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	// No-op
}

/*
Converts all the oracle commmitments into messages
and ask the verifier to manually verify all queries.

Primary use-case is testing.
*/
// Compile compiles the dummy IOP, marking commitments and queries appropriately.
func Compile(comp *wizard.CompiledIOP) {
	comp.DummyCompiled = true

	/*
		Registers all declared commitments and query parameters
		as messages in the same round. This step is only relevant
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

	for i := 0; i < numRounds; i++ {
		// Mark all the commitments as messages
		coms := comp.Columns.AllKeysAt(i)
		for _, com := range coms {
			// Check the status of the commitment
			status := comp.Columns.Status(com)

			if status == column.Ignored {
				// If the column is ignored, we can just skip it
				continue
			}

			if status.IsPublic() {
				// Nothing specific to do on the prover side
				continue
			}

			// Mark them as "public" to the verifier
			switch status {
			case column.Precomputed:
				// Send it to the verifier directly as part of the verifying key
				comp.Columns.SetStatus(com, column.VerifyingKey)
			case column.Committed:
				// Send it to the verifier directly as part of the proof
				comp.Columns.SetStatus(com, column.Proof)
			default:
				utils.Panic("Unknown status : %v", status.String())
			}
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
		One step to be run at the end, by verifying every constraint
		"a la mano"
	*/
	logrus.Debugf("NB: The gnark circuit does not check the verifier of the dummy reduction\n")
	comp.RegisterVerifierAction(numRounds-1, &dummyVerifierAction{
		comp:                     comp,
		queriesParamsToCompile:   queriesParamsToCompile,
		queriesNoParamsToCompile: queriesNoParamsToCompile,
	})
}
