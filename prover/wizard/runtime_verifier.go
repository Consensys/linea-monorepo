package wizard

import (
	"sync"

	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
)

type RuntimeVerifier struct {
	runtimeCommon
}

func (comp *CompiledIOP) Verify(proof Proof) error {

	run := &RuntimeVerifier{
		runtimeCommon: runtimeCommon{
			columns:  proof.Columns,
			queryRes: proof.QueryRes,
			coins:    collection.NewMapping[id, any](),
			fs:       fiatshamir.NewMiMCFiatShamir(),
			lock:     &sync.Mutex{},
		},
	}

	for i := 0; i < comp.NumRounds(); i++ {
		run.runFSForRound(comp, i)
	}

	return run.runAllVerifierCheck(comp)
}
