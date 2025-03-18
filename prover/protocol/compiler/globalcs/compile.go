package globalcs

import (
	"fmt"
	"time"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/sirupsen/logrus"
)

const (
	GLOBAL_REDUCTION                string = "GLOBAL_REDUCTION"
	OFFSET_RANDOMNESS               string = "OFFSET_RANDOMNESS"
	DEGREE_RANDOMNESS               string = "DEGREE_RANDOMNESS"
	QUOTIENT_POLY_TMPL              string = "QUOTIENT_DEG_%v_SHARE_%v"
	EVALUATION_RANDOMESS            string = "EVALUATION_RANDOMNESS"
	UNIVARIATE_EVAL_ALL_HANDLES     string = "UNIV_EVAL_ALL_HANDLES"
	UNIVARIATE_EVAL_QUOTIENT_SHARES string = "UNIV_EVAL_QUOTIENT_%v_OVER_%v"
)

// Compile takes all the uncompiled global constraint found in comp and compile
// them using Plonk's quotient technique. The compiler also applies symbolic
// expression optimization and runtime memory optimizations for the prover.
func Compile(comp *wizard.CompiledIOP) {

	BenchmarkGlobalConstraint(comp)

	logrus.Trace("started global constraint compiler")
	defer logrus.Trace("finished global constraint compiler")

	merging, anyCs := accumulateConstraints(comp)
	if !anyCs {
		return
	}

	var (
		aggregateExprs  = merging.aggregateConstraints(comp)
		factoredExprs   = factorExpressionList(comp, aggregateExprs)
		quotientCtx     = createQuotientCtx(comp, merging.Ratios, factoredExprs)
		evaluationCtx   = declareUnivariateQueries(comp, quotientCtx)
		quotientRound   = quotientCtx.QuotientShares[0][0].Round()
		evaluationRound = quotientRound + 1
	)

	comp.RegisterProverAction(quotientRound, &quotientCtx)
	comp.RegisterProverAction(evaluationRound, evaluationProver(evaluationCtx))
	comp.RegisterVerifierAction(evaluationRound, &evaluationVerifier{evaluationCtx: evaluationCtx})

}

func deriveName(comp *wizard.CompiledIOP, s string, args ...any) string {
	fmts := fmt.Sprintf(s, args...)
	return fmt.Sprintf("%v_%v_%v", GLOBAL_REDUCTION, comp.SelfRecursionCount, fmts)
}

// BenchmarkGlobalConstraint adds a prover step where the prover checks every
// constraints one-by-one.
func BenchmarkGlobalConstraint(comp *wizard.CompiledIOP) {

	var (
		queries  = []*symbolic.Expression{}
		ratios   = []int{}
		ids      = []ifaces.QueryID{}
		allQs    = comp.QueriesNoParams.AllKeys()
		maxRound = 0
	)

	for _, qname := range allQs {

		if comp.QueriesNoParams.IsIgnored(qname) {
			continue
		}

		q := comp.QueriesNoParams.Data(qname)

		globcs, ok := q.(query.GlobalConstraint)
		if !ok {
			continue
		}

		maxRound = max(maxRound, comp.QueriesNoParams.Round(qname))
		bndCancelledExpr := getBoundCancelledExpression(globcs)
		ratio := getExprRatio(bndCancelledExpr)
		queries = append(queries, bndCancelledExpr)
		ratios = append(ratios, ratio)
		ids = append(ids, globcs.ID)
	}

	comp.SubProvers.AppendToInner(maxRound, func(run *wizard.ProverRuntime) {

		for i := range queries {
			board := queries[i].Board()
			t0 := time.Now()
			_ = column.EvalExprColumn(run, board)
			fmt.Printf("[benchmark.global] time=%v ratio=%v qname=%v\n", time.Since(t0), ratios[i], ids[i])
		}

	})

}
