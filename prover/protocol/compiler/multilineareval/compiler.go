package multilineareval

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// context holds the compiled artifacts for one batch of MultilinearEval queries
// that share the same numVars. All input queries are at wizard round [Round];
// the lambda coin, round-poly proof column, and residual query live at [Round+1].
type context struct {
	// InputQueries are the MultilinearEval queries being compiled away.
	InputQueries []query.MultilinearEval
	// NumVars is the shared n (= log2 of column size) for this batch.
	NumVars int
	// Round is the wizard round at which the input queries were registered.
	Round int
	// LambdaCoin is the batching coin sampled after the prover commits to the
	// input evaluations. It seeds the sumcheck transcript.
	LambdaCoin coin.Info
	// RoundPolys is a proof column of size NumVars*3 holding the three
	// evaluation points of each per-round sumcheck polynomial (fext).
	RoundPolys ifaces.Column
	// Residual is the output MultilinearEval query at the single shared
	// sumcheck point, combining all input columns.
	Residual query.MultilinearEval
}

// Compile returns a wizard compilation pass that replaces all MultilinearEval
// queries with a batched sumcheck reduction. Each group of queries sharing the
// same (round, numVars) is compiled independently.
//
// After compilation the input queries are marked as ignored, and a residual
// MultilinearEval query is inserted at round+1. A subsequent compiler pass
// (e.g. Vortex) is responsible for handling the residual.
func Compile(comp *wizard.CompiledIOP) {
	type groupKey struct{ round, numVars int }

	groups := map[groupKey][]query.MultilinearEval{}
	var orderedKeys []groupKey

	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)
		k := groupKey{r, q.NumVars}
		if _, seen := groups[k]; !seen {
			orderedKeys = append(orderedKeys, k)
		}
		groups[k] = append(groups[k], q)
	}

	if len(orderedKeys) == 0 {
		return
	}

	for _, k := range orderedKeys {
		ctx := buildContext(comp, k.round, k.numVars, groups[k])
		comp.RegisterProverAction(k.round+1, &proverAction{ctx: ctx})
		comp.RegisterVerifierAction(comp.NumRounds()-1, &verifierAction{ctx: ctx})
	}
}

func nextPow2(x int) int {
	if x <= 1 {
		return 1
	}
	p := 1
	for p < x {
		p <<= 1
	}
	return p
}

// buildContext inserts the protocol items for one group and returns the context.
func buildContext(comp *wizard.CompiledIOP, round, numVars int, queries []query.MultilinearEval) *context {
	var allCols []ifaces.Column
	for _, q := range queries {
		allCols = append(allCols, q.Pols...)
	}

	suffix := fmt.Sprintf("n%d_r%d_%d", numVars, round, comp.SelfRecursionCount)

	lambdaCoin := comp.InsertCoin(
		round+1,
		coin.Name(fmt.Sprintf("MLEVAL_LAMBDA_%s", suffix)),
		coin.FieldExt,
	)

	roundPolysCol := comp.InsertProof(
		round+1,
		ifaces.ColID(fmt.Sprintf("MLEVAL_ROUND_POLYS_%s", suffix)),
		nextPow2(numVars*3),
		false,
	)

	residual := comp.InsertMultilinear(
		round+1,
		ifaces.QueryID(fmt.Sprintf("MLEVAL_RESIDUAL_%s", suffix)),
		numVars,
		allCols,
	)

	return &context{
		InputQueries: queries,
		NumVars:      numVars,
		Round:        round,
		LambdaCoin:   lambdaCoin,
		RoundPolys:   roundPolysCol,
		Residual:     residual,
	}
}
