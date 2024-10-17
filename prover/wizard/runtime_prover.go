package wizard

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
)

type RuntimeProver struct {
	runtimeCommon
	topLevelAction func(run *RuntimeProver)
	comp           *CompiledIOP
	originalLogger *slog.Logger
	Logger         *slog.Logger
	scope          string
}

type Proof struct {
	Columns  collection.Mapping[id, smartvectors.SmartVector]
	QueryRes collection.Mapping[id, QueryResult]
}

func (comp *CompiledIOP) NewRuntimeProver(topLevelAction func(run *RuntimeProver)) *RuntimeProver {

	run := &RuntimeProver{
		runtimeCommon: runtimeCommon{
			columns:  collection.NewMapping[id, smartvectors.SmartVector](),
			coins:    collection.NewMapping[id, any](),
			queryRes: collection.NewMapping[id, QueryResult](),
			fs:       fiatshamir.NewMiMCFiatShamir(),
			lock:     &sync.Mutex{},
		},
		topLevelAction: topLevelAction,
		comp:           comp,
	}

	// For safety, we are forced
	run.WithLogger(slog.New(slog.NewTextHandler(
		utils.NoWriter{},
		&slog.HandlerOptions{Level: slog.LevelError},
	)))

	// Pass the precomputed polynomials
	for key, val := range comp.precomputeds.InnerMap() {
		run.columns.InsertNew(key, val)
	}

	return run
}

func (run *RuntimeProver) WithLogger(logger *slog.Logger) *RuntimeProver {
	run.originalLogger = logger.WithGroup("prover-runtime")
	run.Logger = run.originalLogger
	return run
}

func (run *RuntimeProver) Run() *RuntimeProver {

	numRounds := run.comp.NumRounds()

	for currRound := 0; currRound < numRounds; currRound++ {
		run.runFSForRound(run.comp, currRound)

		if currRound == 0 {
			run.topLevelAction(run)
		}

		for _, pa := range run.comp.runtimeProverActions.allAt(currRound) {
			pa.Run(run)
		}
	}

	return run
}

func (run *RuntimeProver) Proof() Proof {

	var (
		proof = Proof{
			Columns:  collection.NewMapping[id, smartvectors.SmartVector](),
			QueryRes: collection.NewMapping[id, QueryResult](),
		}
		proofMsg    = ProofMsg
		allCols     = run.comp.AllMatchingColumns(nil, &proofMsg, nil)
		allQueryRes = run.comp.queries.all()
		err         error
	)

	for _, col := range allCols {
		cola, ok := run.tryGetColumn(&col)
		if !ok {
			err = errors.Join(err, fmt.Errorf("missing column assignment %v. Explainer: \n%v", col.String(), col.Explain()))
		}
		proof.Columns.InsertNew(col.id(), cola)
	}

	for _, q := range allQueryRes {
		qr := run.getOrComputeQueryRes(q)
		proof.QueryRes.InsertNew(q.id(), qr)
	}

	return proof
}

func (run *RuntimeProver) SanityCheck() error {

	var (
		errVAs = run.runAllVerifierCheck(run.comp)
		errQs  error
	)

	for _, q := range run.comp.queries.all() {
		errQ := q.Check(run)
		if errQ != nil {
			errQ = fmt.Errorf("error checking the query %v. Explainer:\n %v", q.String(), q.Explain())
		}
		errQs = errors.Join(errQs, errQ)
	}

	if errQs != nil {
		errQs = fmt.Errorf("errors checking the queries: %w", errQs)
	}

	return errors.Join(errQs, errVAs)
}

func (run *RuntimeProver) ChildScope(subScope string) *RuntimeProver {
	return &RuntimeProver{
		runtimeCommon:  run.runtimeCommon,
		comp:           run.comp,
		scope:          run.scope + "/" + subScope,
		originalLogger: run.originalLogger,
		Logger:         run.originalLogger.With(slog.String("scope", run.scope+"/"+subScope)),
		topLevelAction: run.topLevelAction,
	}
}
