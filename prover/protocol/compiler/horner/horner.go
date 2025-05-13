package horner

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// hornerCtx is a compilation artefact generated during the execution of the
// [CompileHorner] compiler.
type hornerCtx struct {
	// Column is the accumulating column used to check the computation of a
	// horner value for one [HornerPart].
	AccumulatingCols [][]ifaces.Column

	// CountingInnerProduct is the inner-product query used to check the
	// counting for each selector. Each entry of the inner-product query
	// maps to a selector.
	CountingInnerProducts []query.InnerProduct

	// LocOpenings are the local openings used to check the first value of
	// Columns[i]
	LocOpenings []query.LocalOpening

	// Q is the Horner query
	Q *query.Horner
}

// assignHornerCtx is a [wizard.ProverAction] assigning the local openings of
// the Horner accumulating columns and the accumulating columns themselves.
// The function also sanity-checks the parameters assignment.
type assignHornerCtx struct {
	hornerCtx
}

// assignHornerIP is a [wizard.ProverAction] assigning the inner-products of
// the Horner compilation.
type assignHornerIP struct {
	hornerCtx
}

// checkHornerResult is [wizard.VerifierAction] responsible for checking that
// the values of n1 are correct (by checking the consistency with the IP queries)
// and checking that the final result is correctly computed by inspecting the
// local openings.
type checkHornerResult struct {
	hornerCtx
	skipped bool
}

// CompileHoner compiles the [query.Horner] queries one by one by "transforming"
// them into [query.InnerProduct], [query.LocalOpening], [query.Global] and
// [query.Local].
func CompileHorner(comp *wizard.CompiledIOP) {

	qNames := comp.QueriesParams.AllUnignoredKeys()

	for i := range qNames {

		q, ok := comp.QueriesParams.Data(qNames[i]).(*query.Horner)
		if !ok {
			continue
		}

		comp.QueriesParams.MarkAsIgnored(qNames[i])

		compileHornerQuery(comp, q)
	}
}

func compileHornerQuery(comp *wizard.CompiledIOP, q *query.Horner) {

	var (
		round = q.Round
		ctx   = hornerCtx{
			Q: q,
		}
		iPRound = 0
	)

	for i, part := range q.Parts {

		var (
			numList      = len(q.Parts[i].Coefficients)
			accumulators = make([]ifaces.Column, numList)
		)

		for j := 0; j < numList; j++ {

			accumulators[j] = comp.InsertCommit(
				round,
				ifaces.ColIDf("HORNER_%v_PART_%v_COLUMN_%v", q.ID, i, j),
				part.Size(),
			)
		}

		for j := 0; j < numList-1; j++ {

			prevAcc := column.Shift(accumulators[numList-1], 1)
			if j > 0 {
				prevAcc = accumulators[j-1]
			}

			comp.InsertGlobal(
				round,
				ifaces.QueryIDf("HORNER_%v_PART_%v_GLOBAL_%v", q.ID, i, j),
				sym.Sub(
					accumulators[j],
					microAccumulate(
						part.Selectors[j],
						prevAcc,
						part.X,
						part.Coefficients[j],
					),
				),
			)
		}

		loc := comp.InsertLocalOpening(
			round,
			ifaces.QueryIDf("HORNER_%v_PART_%v_LOCAL_OPENING", q.ID, i),
			accumulators[numList-1],
		)

		// This query takes care of checking the initial value of the column
		comp.InsertLocal(
			round,
			ifaces.QueryIDf("HORNER_%v_PART_%v_LOCAL", q.ID, i),
			sym.Sub(
				column.Shift(accumulators[0], -1),
				sym.Mul(
					column.Shift(q.Parts[i].Selectors[0], -1),
					column.ShiftExpr(q.Parts[i].Coefficients[0], -1),
				),
			),
		)

		ctx.AccumulatingCols = append(ctx.AccumulatingCols, accumulators)
		ctx.LocOpenings = append(ctx.LocOpenings, loc)
		iPRound = max(iPRound, q.Parts[i].Selectors[0].Round())
	}

	for i, part := range q.Parts {

		size := part.Size()

		// In theory, it would be more efficient to batch everything in a single
		// inner-product by size. But that would make the code harder to understand
		// and would require backtracking which result of the query corresponds to
		// which part of the Horner query.
		ip := comp.InsertInnerProduct(
			iPRound,
			ifaces.QueryIDf("HORNER_%v_COUNTING_%v_SRCNT_%v_%v", q.ID, size, comp.SelfRecursionCount, i),
			verifiercol.NewConstantCol(field.One(), size),
			ctx.Q.Parts[i].Selectors,
		)

		ctx.CountingInnerProducts = append(ctx.CountingInnerProducts, ip)
	}

	comp.RegisterProverAction(iPRound, assignHornerIP{ctx})
	comp.RegisterProverAction(q.Round, assignHornerCtx{ctx})
	comp.RegisterVerifierAction(q.Round, &checkHornerResult{hornerCtx: ctx})
}

func (a assignHornerCtx) Run(run *wizard.ProverRuntime) {

	var (
		params = run.GetHornerParams(a.Q.ID)
		res    = field.Zero()
	)

	for i, part := range a.Q.Parts {

		var (
			arity        = len(part.Selectors)
			datas        = make([]smartvectors.SmartVector, arity)
			selectors    = make([]smartvectors.SmartVector, arity)
			x            = part.X.GetVal(run)
			n0           = params.Parts[i].N0
			count        = 0
			numRow       = part.Size()
			acc          = field.Zero()
			accumulators = make([][]field.Element, arity)
		)

		for k := 0; k < arity; k++ {
			board := part.Coefficients[k].Board()
			datas[k] = column.EvalExprColumn(run, board)
			selectors[k] = part.Selectors[k].GetColAssignment(run)
			accumulators[k] = make([]field.Element, numRow)
		}

		for row := numRow - 1; row >= 0; row-- {
			for k := 0; k < arity; k++ {

				sel := selectors[k].Get(row)
				if sel.IsOne() {
					count++
				}

				acc = computeMicroAccumulate(selectors[k].Get(row), acc, x, datas[k].Get(row))
				accumulators[k][row] = acc
			}
		}

		if n0+count != params.Parts[i].N1 {
			// To update once we merge with the "code 78" branch as it means that a constraint is not satisfied.
			utils.Panic("the counting of the 1s in the filter does not match the one in the local-opening: (%v-%v) != %v", params.Parts[i].N1, n0, count)
		}

		for k := 0; k < arity; k++ {
			run.AssignColumn(a.AccumulatingCols[i][k].GetColID(), smartvectors.NewRegular(accumulators[k]))
		}

		tmp := accumulators[arity-1][0]
		run.AssignLocalPoint(a.LocOpenings[i].ID, tmp)

		if n0 > 0 {
			xN0 := new(field.Element).Exp(x, big.NewInt(int64(n0)))
			tmp.Mul(&tmp, xN0)
		}

		if a.Q.Parts[i].SignNegative {
			tmp.Neg(&tmp)
		}

		res.Add(&res, &tmp)
	}

	if res != params.FinalResult {
		utils.Panic("the horner query %v is assigned a final result of %v but we recomputed %v\n", a.Q.ID, res.String(), params.FinalResult.String())
	}
}

func (a assignHornerIP) Run(run *wizard.ProverRuntime) {

	for i := range a.Q.Parts {

		var (
			ip        = a.CountingInnerProducts[i]
			selectors = a.Q.Parts[i].Selectors
			res       = make([]field.Element, len(selectors))
		)

		for i, selector := range selectors {
			sel := selector.GetColAssignment(run).IntoRegVecSaveAlloc()
			for j := range sel {
				res[i].Add(&res[i], &sel[j])
			}
		}

		run.AssignInnerProduct(ip.ID, res...)
	}

}

func (c *checkHornerResult) Run(run wizard.Runtime) error {

	var (
		hornerQuery  = c.Q
		hornerParams = run.GetHornerParams(hornerQuery.ID)
		res          = field.Zero()
	)

	for i := range c.Q.Parts {

		var (
			ipQuery  = c.CountingInnerProducts[i]
			ipParams = run.GetInnerProductParams(c.CountingInnerProducts[i].ID)
			found    = false
		)

		for k := range ipQuery.Bs {

			if !ipParams.Ys[k].IsUint64() {
				return errors.New("ip result does not fit on a uint64")
			}

			ipCount := int(ipParams.Ys[k].Uint64())

			for j, c := range hornerQuery.Parts {

				if c.Selectors[0].GetColID() != ipQuery.Bs[k].GetColID() {
					continue
				}

				found = true
				params := hornerParams.Parts[j]
				n0, n1 := params.N0, params.N1

				if n1-n0 != ipCount {
					return fmt.Errorf("inner-product and horner params do not match: %v - %v (%v) != %v", n1, n0, n1-n0, ipCount)
				}

				break
			}

			if !found {
				utils.Panic("could not find selector %v from the Horner query", ipQuery.Bs[k].String())
			}
		}
	}

	// This loop is responsible for checking that the final result is correctly
	// computed by inspecting the local openings.
	for i, lo := range c.LocOpenings {

		var (
			tmp = run.GetLocalPointEvalParams(lo.ID).Y
			n0  = hornerParams.Parts[i].N0
			x   = hornerQuery.Parts[i].X.GetVal(run)
		)

		if n0 > 0 {
			xN0 := new(field.Element).Exp(x, big.NewInt(int64(n0)))
			tmp.Mul(&tmp, xN0)
		}

		if hornerQuery.Parts[i].SignNegative {
			tmp.Neg(&tmp)
		}

		res.Add(&res, &tmp)
	}

	if res != hornerParams.FinalResult {
		return fmt.Errorf("horner query has finalResult %v but we recovered %v", res.String(), hornerParams.FinalResult.String())
	}

	return nil
}

func (c *checkHornerResult) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		hornerQuery  = c.Q
		hornerParams = run.GetHornerParams(hornerQuery.ID)
		res          = frontend.Variable(0)
	)

	for i := range c.Q.Parts {

		var (
			ipQuery  = c.CountingInnerProducts[i]
			ipParams = run.GetInnerProductParams(c.CountingInnerProducts[i].ID)
			found    = false
		)

		for k := range ipQuery.Bs {

			ipCount := ipParams.Ys[k]

			for j, c := range hornerQuery.Parts {

				if c.Selectors[0].GetColID() != ipQuery.Bs[k].GetColID() {
					continue
				}

				found = true
				params := hornerParams.Parts[j]
				n0, n1 := params.N0, params.N1

				api.AssertIsEqual(n1, api.Add(n0, ipCount))
				break
			}

			if !found {
				utils.Panic("could not find selector %v from the Horner query", ipQuery.Bs[k].String())
			}
		}
	}

	// This loop is responsible for checking that the final result is correctly
	// computed by inspecting the local openings.
	for i, lo := range c.LocOpenings {

		var (
			tmp = run.GetLocalPointEvalParams(lo.ID).Y
			n0  = hornerParams.Parts[i].N0
			x   = hornerQuery.Parts[i].X.GetFrontendVariable(api, run)
		)

		xN0 := gnarkutil.ExpVariableExponent(api, x, n0, 64)
		tmp = api.Mul(tmp, xN0)

		if hornerQuery.Parts[i].SignNegative {
			tmp = api.Neg(tmp)
		}

		res = api.Add(res, tmp)
	}

	api.AssertIsEqual(res, hornerParams.FinalResult)
}

func (c *checkHornerResult) Skip() {
	c.skipped = true
}

func (c *checkHornerResult) IsSkipped() bool {
	return c.skipped
}

// microAccumulate returns an atomic accumulator update expression.
//
// the returned expression evaluates to:
//
// ```
//
//	sel == 1 => acc.X + p
//	sel == 0 => acc
//
// ```
func microAccumulate(sel, acc, x, p any) *sym.Expression {
	return sym.Add(
		sym.Mul(
			sel,
			sym.Add(p, sym.Mul(x, acc)),
		),
		sym.Mul(
			sym.Sub(1, sel),
			acc,
		),
	)
}

func computeMicroAccumulate(sel, acc, x, p field.Element) field.Element {

	if sel.IsZero() {
		return acc
	}

	if sel.IsOne() {
		var tmp field.Element
		tmp.Mul(&x, &acc)
		tmp.Add(&tmp, &p)
		return tmp
	}

	panic("selector is non-binary")
}
