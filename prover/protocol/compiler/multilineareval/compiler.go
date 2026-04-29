package multilineareval

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// polRef identifies one polynomial within a MultilinearEval query.
type polRef struct {
	QueryName ifaces.QueryID
	PolIdx    int
	Col       ifaces.Column
}

// context holds the compiled artifacts for one batch of polynomials sharing
// the same (round, numVars). All polRefs have NumVars == NumVars.
// The lambda coin, round-poly proof column, and residual query live at Round+1.
type context struct {
	PolRefs    []polRef              // flat list of (queryID, polIdx) pairs
	NumVars    int                   // shared n for all polys in this batch
	Round      int
	LambdaCoin coin.Info
	RoundPolys ifaces.Column
	Residual   query.MultilinearEval // all polys at the shared sumcheck point c
}

// Compile returns a wizard compilation pass that replaces all MultilinearEval
// queries with a batched sumcheck reduction. Polynomials are grouped by
// (round, numVars) — each poly in each query is placed in the matching group
// independently. The residual MultilinearEval for each group is at the single
// shared sumcheck point.
//
// After compilation the input queries are marked as ignored, and a residual
// MultilinearEval query is inserted at round+1. A subsequent compiler pass
// (e.g. Vortex) is responsible for handling the residual.
func Compile(comp *wizard.CompiledIOP) {
	type groupKey struct{ round, numVars int }

	groups := map[groupKey][]polRef{}
	var orderedKeys []groupKey

	for _, name := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		comp.QueriesParams.MarkAsIgnored(name)
		r := comp.QueriesParams.Round(name)

		for j, col := range q.Pols {
			n := q.NumVars[j]
			k := groupKey{r, n}
			if _, seen := groups[k]; !seen {
				orderedKeys = append(orderedKeys, k)
			}
			groups[k] = append(groups[k], polRef{QueryName: q.Name(), PolIdx: j, Col: col})
		}
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

// CompileIgnored processes all currently-IGNORED MultilinearEval queries in comp,
// compiling them exactly as Compile does for unignored ones. Call this after
// DistributeWizard so that the ML columns are only created on the Bootstrapper
// after FilterCompiledIOP has already run.
func CompileIgnored(comp *wizard.CompiledIOP) {
	type groupKey struct{ round, numVars int }

	groups := map[groupKey][]polRef{}
	var orderedKeys []groupKey

	for _, name := range comp.QueriesParams.AllKeys() {
		if !comp.QueriesParams.IsIgnored(name) {
			continue
		}
		q, ok := comp.QueriesParams.Data(name).(query.MultilinearEval)
		if !ok {
			continue
		}
		r := comp.QueriesParams.Round(name)

		for j, col := range q.Pols {
			n := q.NumVars[j]
			k := groupKey{r, n}
			if _, seen := groups[k]; !seen {
				orderedKeys = append(orderedKeys, k)
			}
			groups[k] = append(groups[k], polRef{QueryName: q.Name(), PolIdx: j, Col: col})
		}
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
func buildContext(comp *wizard.CompiledIOP, round, numVars int, refs []polRef) *context {
	cols := make([]ifaces.Column, len(refs))
	for i, ref := range refs {
		cols[i] = ref.Col
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
		cols,
	)

	return &context{
		PolRefs:    refs,
		NumVars:    numVars,
		Round:      round,
		LambdaCoin: lambdaCoin,
		RoundPolys: roundPolysCol,
		Residual:   residual,
	}
}
