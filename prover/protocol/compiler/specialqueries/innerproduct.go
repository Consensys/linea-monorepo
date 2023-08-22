package specialqueries

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

const (
	InnerIPCoin     coin.Name      = "INNER_PRODUCT_INNER_COIN"
	OuterIPCoin     coin.Name      = "INNER_PRODUCT_OUTER_COIN"
	LocalStart      ifaces.QueryID = "INNER_PRODUCT_LOCAL_START"
	LocalEnd        ifaces.QueryID = "INNER_PRODUCT_LOCAL_OPENING_END"
	SummationGlobal ifaces.QueryID = "INNER_PRODUCT_SUMMATION_GLOBAL"
	SumHandle       ifaces.ColID   = "INNER_PRODUCT_SUM_HANDLE"
)

type ipCtx struct {
	queryMap             map[int][]query.InnerProduct
	sizeIndexes          []int
	startRound           int
	innerCoin, outerCoin coin.Info
	// Contains linears combinations of the the product (a*b) grouped by sizes
	ExprMap []ifaces.Column
	// Accumulate the sum of the above query
	SumsHandle []ifaces.Column

	// Queries created by the compilation step
	SummationGlobal     []query.GlobalConstraint
	SummationLocalStart []query.LocalConstraint
	SummationLocalEnd   []query.LocalOpening
}

func newIpCtx() ipCtx {
	return ipCtx{
		queryMap:    map[int][]query.InnerProduct{},
		sizeIndexes: []int{},
		ExprMap:     []ifaces.Column{},
	}
}

func (ctx *ipCtx) groupIPQueries(comp *wizard.CompiledIOP) {
	// We sort the query by sizes
	queryMap := map[int][]query.InnerProduct{}
	sizeIndexes := []int{} // The index list is used to allow iterating in a deterministic order
	startRound := 0

	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {
		q, ok := comp.QueriesParams.Data(qName).(query.InnerProduct)
		if !ok {
			// not an inner-product query, ignore
			continue
		}

		comp.QueriesParams.MarkAsIgnored(qName)
		startRound = utils.Max(startRound, comp.QueriesParams.Round(qName))
		size := q.A.Size()

		if _, ok := queryMap[size]; !ok {
			// add the size in the index
			sizeIndexes = append(sizeIndexes, size)
		}

		queryMap[size] = append(queryMap[size], q)
	}

	ctx.startRound = startRound
	ctx.queryMap = queryMap
	ctx.sizeIndexes = sizeIndexes
}

func (ctx *ipCtx) aggregateQueries(comp *wizard.CompiledIOP) {
	// Compute the linear combination expression for each

	ctx.ExprMap = make([]ifaces.Column, len(ctx.sizeIndexes))
	ctx.SumsHandle = make([]ifaces.Column, len(ctx.sizeIndexes))

	ctx.SummationGlobal = make([]query.GlobalConstraint, len(ctx.sizeIndexes))
	ctx.SummationLocalStart = make([]query.LocalConstraint, len(ctx.sizeIndexes))
	ctx.SummationLocalEnd = make([]query.LocalOpening, len(ctx.sizeIndexes))

	for k, size := range ctx.sizeIndexes {
		queries := ctx.queryMap[size]
		queriesExpr := make([]*symbolic.Expression, len(queries))
		for i, q := range queries {
			bsVar := make([]*symbolic.Expression, len(q.Bs))
			for j, h := range q.Bs {
				bsVar[j] = ifaces.ColumnAsVariable(h)
			}
			queriesExpr[i] = symbolic.NewPolyEval(ctx.innerCoin.AsVariable(), bsVar).
				Mul(ifaces.ColumnAsVariable(q.A))
		}
		aggExpr := symbolic.NewPolyEval(ctx.outerCoin.AsVariable(), queriesExpr)
		ctx.ExprMap[k] = expr_handle.ExprHandle(comp, aggExpr)

		ctx.SumsHandle[k] = comp.InsertCommit(
			ctx.startRound+1,
			deriveNameWithCnt[ifaces.ColID](comp, SumHandle, size),
			size,
		)

		// s[+1] = s + h
		ctx.SummationGlobal[k] = comp.InsertGlobal(ctx.startRound+1,
			deriveNameWithCnt[ifaces.QueryID](comp, SummationGlobal, size),
			ifaces.ColumnAsVariable(ctx.SumsHandle[k]).
				Add(ifaces.ColumnAsVariable(column.Shift(ctx.ExprMap[k], 1))).
				Sub(ifaces.ColumnAsVariable(column.Shift(ctx.SumsHandle[k], 1))))

		ctx.SummationLocalStart[k] = comp.InsertLocal(ctx.startRound+1,
			deriveNameWithCnt[ifaces.QueryID](comp, LocalStart, size),
			ifaces.ColumnAsVariable(ctx.SumsHandle[k]).Sub(ifaces.ColumnAsVariable(ctx.ExprMap[k])))

		ctx.SummationLocalEnd[k] = comp.InsertLocalOpening(ctx.startRound+1,
			deriveNameWithCnt[ifaces.QueryID](comp, LocalEnd, size), column.Shift(ctx.SumsHandle[k], -1))
	}

}

func (ctx *ipCtx) prover(run *wizard.ProverRuntime) {
	// The alleged evaluation to obtain after applying the scalar-product query
	// They are obtained by taking the same random linear combination but over the
	// alleged values this time.
	stopTimer := profiling.LogTimer("inner-product - proving step")
	for k := range ctx.sizeIndexes {
		witnessAgg := ctx.ExprMap[k].GetColAssignment(run)

		// Compute the sumhandle witness
		sumWitness := make([]field.Element, witnessAgg.Len())
		sumWitness[0] = witnessAgg.Get(0)
		for i := 1; i < len(sumWitness); i++ {
			v := witnessAgg.Get(i)
			sumWitness[i].Add(&sumWitness[i-1], &v)
		}

		run.AssignColumn(ctx.SumsHandle[k].GetColID(), smartvectors.NewRegular(sumWitness))
		run.AssignLocalPoint(ctx.SummationLocalEnd[k].ID, sumWitness[len(sumWitness)-1])
	}
	stopTimer()
}

func (ctx *ipCtx) verifier(assi *wizard.VerifierRuntime) error {
	// The alleged evaluation to obtain after applying the scalar-product query
	// They are obtained by taking the same random linear combination but over the
	// alleged values this time.

	for k, size := range ctx.sizeIndexes {
		queries := ctx.queryMap[size]
		innerRLC := make([]field.Element, len(queries))
		for i, q := range queries {
			ipys := assi.GetInnerProductParams(q.ID)
			innerRLC[i] = poly.EvalUnivariate(ipys.Ys, assi.GetRandomCoinField(ctx.innerCoin.Name))
		}

		expected := poly.EvalUnivariate(innerRLC, assi.GetRandomCoinField(ctx.outerCoin.Name))
		actual := assi.GetLocalPointEvalParams(ctx.SummationLocalEnd[k].ID).Y

		if actual != expected {
			return fmt.Errorf("inner-product verification failed %v != %v", actual.String(), expected.String())
		}
	}

	return nil
}

func (ctx *ipCtx) gnarkVerifier(api frontend.API, c *wizard.WizardVerifierCircuit) {
	// The alleged evaluation to obtain after applying the scalar-product query
	// They are obtained by taking the same random linear combination but over the
	// alleged values this time.

	for k, size := range ctx.sizeIndexes {
		queries := ctx.queryMap[size]
		innerRLC := []frontend.Variable{}
		for _, q := range queries {
			ipys := c.GetInnerProductParams(q.ID).Ys
			v := poly.EvaluateUnivariateGnark(api, ipys, c.GetRandomCoinField(ctx.innerCoin.Name))
			innerRLC = append(innerRLC, v)
		}

		actual := c.GetLocalPointEvalParams(ctx.SummationLocalEnd[k].ID).Y
		expected := poly.EvaluateUnivariateGnark(api, innerRLC, c.GetRandomCoinField(ctx.outerCoin.Name))
		api.AssertIsEqual(expected, actual)
	}
}

func CompileInnerProduct(comp *wizard.CompiledIOP) {

	ctx := newIpCtx()
	ctx.groupIPQueries(comp)

	if len(ctx.sizeIndexes) == 0 {
		logrus.Infof("No inner-product query found, so skipping the inner-product compilation step")
		return
	}

	// Sample two coins, one for a linear combination of the Bs within an IP query, and the outer
	// to aggregate the inner RLC into a single one for each size of commitment.
	ctx.innerCoin = comp.InsertCoin(ctx.startRound+1, coin.Namef("%v_%v", InnerIPCoin, comp.SelfRecursionCount), coin.Field)
	ctx.outerCoin = comp.InsertCoin(ctx.startRound+1, coin.Namef("%v_%v", OuterIPCoin, comp.SelfRecursionCount), coin.Field)

	ctx.aggregateQueries(comp)

	comp.SubProvers.AppendToInner(ctx.startRound+1, ctx.prover)
	comp.InsertVerifier(ctx.startRound+1, ctx.verifier, ctx.gnarkVerifier)
}
