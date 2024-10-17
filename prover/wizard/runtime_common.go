package wizard

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
)

type runtimeCommon struct {
	columns  collection.Mapping[id, smartvectors.SmartVector]
	queryRes collection.Mapping[id, QueryResult]
	coins    collection.Mapping[id, any]
	fs       *fiatshamir.State
	lock     *sync.Mutex
}

// runFSForRound updates the Fiat-Shamir state to prepare entering at round `n`
// if n == 0: the state is updated with a description of the protocol and the
// value of the precomputed polynomials.
func (run *runtimeCommon) runFSForRound(comp *CompiledIOP, n int) {

	var (
		err error
	)

	if n == 0 {
		run.fs.Update(comp.protocolHash)
	}

	if n > 0 {

		var (
			// Sanity-check : Make sure all issued random coin have been "consumed"
			// by all the verifiers steps, in the round we are "closing"
			allColumnsPrevRound = comp.columns.allAt(n - 1)
			allQueriesPrevRound = comp.queries.allAt(n - 1)
			allCoins            = comp.coins.allAt(n)
		)

		for _, col := range allColumnsPrevRound {
			if !col.visibility.IsPublic() {
				continue
			}

			sv, found := run.columns.TryGet(col.id())
			if !found {
				err = errors.Join(err, fmt.Errorf("missing column assignment: %v | Explainer: %v", col.String(), col.Explain()))
			}

			run.fs.UpdateSV(sv)
		}

		if err != nil {
			utils.Panic("got one of more error while generating FS for round %v: %v", n, err.Error())
		}

		for _, q := range allQueriesPrevRound {
			if q.IsDeferredToVerifier() {
				continue
			}
			run.getOrComputeQueryRes(q).UpdateFS(run.fs)
		}

		for _, c := range allCoins {
			v := c.sample(run.fs)
			run.coins.InsertNew(c.id(), v)
		}

	}
}

func (run *runtimeCommon) tryGetColumn(col *ColNatural) (smartvectors.SmartVector, bool) {
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.columns.TryGet(col.id())
}

func (run *runtimeCommon) tryGetQueryRes(q Query) (QueryResult, bool) {
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.queryRes.TryGet(q.id())
}

func (run *runtimeCommon) getOrComputeQueryRes(q Query) QueryResult {
	if v, ok := run.tryGetQueryRes(q); ok {
		return v
	}

	v := q.computeResult(run)
	run.queryRes.InsertNew(q.id(), v)
	return v
}

func (run *runtimeCommon) tryGetCoin(c Coin) (any, bool) {
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.coins.TryGet(c.id())
}

func (run *runtimeCommon) runAllVerifierCheck(comp *CompiledIOP) error {

	var err error = nil
	for _, va := range comp.runtimeVerifierActions.all() {
		errVA := va.Run(run)
		if errVA != nil {
			errVA = fmt.Errorf("verifier check %v. Explainer:\n%v", va.String(), va.Explain())
		}
		err = errors.Join(err, errVA)
	}

	if err == nil {
		return nil
	}

	return fmt.Errorf("errors while running the verifier checks: %w", err)
}
