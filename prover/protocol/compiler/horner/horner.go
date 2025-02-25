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
	AccumulatingCols []ifaces.Column

	// Selectors are the list of the selectors for each columns, sorted by
	// size.
	Selectors map[int]*[]ifaces.Column

	// CountingInnerProduct is the inner-product query used to check the
	// counting for each selector.
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

		compileHornerQuery(comp, q)
		comp.QueriesParams.MarkAsIgnored(qNames[i])
	}
}

func compileHornerQuery(comp *wizard.CompiledIOP, q *query.Horner) {

	var (
		round = q.Round
		ctx   = hornerCtx{
			Selectors: make(map[int]*[]ifaces.Column),
			Q:         q,
		}
		iPRound = 0
	)

	for i := range q.Parts {

		col := comp.InsertCommit(
			round,
			ifaces.ColIDf("HORNER_%v_PART_%v_COLUMN", q.ID, i),
			q.Parts[i].Size(),
		)

		loc := comp.InsertLocalOpening(
			round,
			ifaces.QueryIDf("HORNER_%v_PART_%v_LOCAL_OPENING", q.ID, i),
			col,
		)

		comp.InsertGlobal(
			round,
			ifaces.QueryIDf("HORNER_%v_PART_%v_GLOBAL", q.ID, i),
			sym.Sub(
				col,
				sym.Mul(
					sym.Sub(1, q.Parts[i].Selector),
					column.Shift(col, 1),
				),
				sym.Mul(
					q.Parts[i].Selector,
					sym.Add(
						sym.Mul(q.Parts[i].X, column.Shift(col, 1)),
						q.Parts[i].Coefficient),
				),
			),
		)

		// This query takes care of checking the initial value of the column
		comp.InsertLocal(
			round,
			ifaces.QueryIDf("HORNER_%v_PART_%v_LOCAL", q.ID, i),
			sym.Sub(
				column.Shift(col, -1),
				sym.Mul(
					column.Shift(q.Parts[i].Selector, -1),
					column.ShiftExpr(q.Parts[i].Coefficient, -1),
				),
			),
		)

		partSize := q.Parts[i].Size()

		if _, ok := ctx.Selectors[partSize]; !ok {
			ctx.Selectors[partSize] = &[]ifaces.Column{}
		}

		selectorsForSize := ctx.Selectors[partSize]
		*selectorsForSize = append(*selectorsForSize, q.Parts[i].Selector)
		ctx.AccumulatingCols = append(ctx.AccumulatingCols, col)
		ctx.LocOpenings = append(ctx.LocOpenings, loc)
		iPRound = max(iPRound, q.Parts[i].Selector.Round())
	}

	sizes := utils.SortedKeysOf(ctx.Selectors, func(a, b int) bool {
		return a < b
	})

	for _, size := range sizes {

		selectors := ctx.Selectors[size]

		ctx.CountingInnerProducts = append(
			ctx.CountingInnerProducts,
			comp.InsertInnerProduct(
				iPRound,
				ifaces.QueryIDf("HORNER_%v_COUNTING_%v_SRCNT_%v", q.ID, size, comp.SelfRecursionCount),
				verifiercol.NewConstantCol(field.One(), size),
				*selectors,
			),
		)
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

	for i, lo := range a.LocOpenings {

		var (
			col        = make([]field.Element, a.AccumulatingCols[i].Size())
			coeffBoard = a.Q.Parts[i].Coefficient.Board()
			selector   = a.Q.Parts[i].Selector.GetColAssignment(run).IntoRegVecSaveAlloc()
			data       = column.EvalExprColumn(run, coeffBoard).IntoRegVecSaveAlloc()
			x          = a.Q.Parts[i].X.GetVal(run)
			n0         = params.Parts[i].N0
			count      = 0
		)

		col[len(col)-1].Mul(&selector[len(col)-1], &data[len(col)-1])

		for j := len(col) - 2; j >= 0; j-- {

			if selector[j].IsZero() {
				col[j].Set(&col[j+1])
				continue
			}

			if selector[j].IsOne() {
				col[j].Mul(&col[j+1], &x)
				col[j].Add(&col[j], &data[j])
				count++
				continue
			}

			utils.Panic("selector should be a binary column (and this should be enforced by the caller). If this is failing, then the circuit has a soundness error")
		}

		if n0+count != params.Parts[i].N1 {
			utils.Panic("Horner query has %v parts but HornerParams has %v", len(a.Q.Parts), len(params.Parts))
		}

		run.AssignColumn(a.AccumulatingCols[i].GetColID(), smartvectors.NewRegular(col))
		run.AssignLocalPoint(lo.ID, col[0])

		tmp := col[0]

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
		utils.Panic("the horner query is assigned a final result of %v but we recomputed %v\n", res, params.FinalResult)
	}
}

func (a assignHornerIP) Run(run *wizard.ProverRuntime) {

	sizes := utils.SortedKeysOf(a.Selectors, func(a, b int) bool {
		return a < b
	})

	for i, size := range sizes {

		var (
			ip        = a.CountingInnerProducts[i]
			selectors = a.Selectors[size]
			res       = make([]field.Element, len(*selectors))
		)

		for i, selector := range *selectors {
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

	sizes := utils.SortedKeysOf(c.Selectors, func(a, b int) bool {
		return a < b
	})

	// This loop is responsible for checking the consistency of the IP queries
	// with the N1-N0 difference from the Horner params.
	for i := range sizes {

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

				if c.Selector.GetColID() != ipQuery.Bs[k].GetColID() {
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

	sizes := utils.SortedKeysOf(c.Selectors, func(a, b int) bool {
		return a < b
	})

	// This loop is responsible for checking the consistency of the IP queries
	// with the N1-N0 difference from the Horner params.
	for i := range sizes {

		var (
			ipQuery  = c.CountingInnerProducts[i]
			ipParams = run.GetInnerProductParams(c.CountingInnerProducts[i].ID)
			found    = false
		)

		for k := range ipQuery.Bs {

			ipCount := ipParams.Ys[k]

			for j, c := range hornerQuery.Parts {

				if c.Selector.GetColID() != ipQuery.Bs[k].GetColID() {
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
