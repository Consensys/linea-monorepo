package mpts

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// MultiPointToSinglePointCompilation holds the compilation context of the
// multipoint to singlepoint step. This step ensures that all the PIOP
// evaluations are done at the same points for all the polynomials. This
// ensures that Vortex can be applied on top.
type MultipointToSinglepointCompilation struct {

	// Queries lists all the compiled evaluation queries that are relevant
	// to the compilation step.
	Queries []query.UnivariateEval

	// Polys lists all the polynomials involved in the compilation. Meaning
	// all the polynomials that are affetcted by one of the queries. The
	// pols appear in the following order: first by round, then by column
	// name (alphabetical order). Finally, each "round" sequence is punctuated
	// by a sequence of "shadow columns", these are columns that are added
	// by the compiler to match a prescribed profile.
	Polys []ifaces.Column

	// EvalPointOfPolys lists, for each entry in [Polys], the entries of
	// [Queries] corresponding to evaluation points affected to the
	// polynomial. The list can be empty for shadow columns.
	EvalPointOfPolys [][]int

	// PolysOfEvalPoint lists, for each entry in [Queries], the entries of
	// [Polys] corresponding to polynomials evaluated at the same point.
	PolysOfEvalPoint [][]int

	// NumColumnProfileOpt is an optional compilation parameter that can be
	// set to control the number of column in every round of the compilation.
	// This can be used to obtain uniform wizards when doing conglomeration.
	NumColumnProfileOpt []int

	// numRow is the number of rows in the columns that are compiled. The
	// value is lazily evaluated and the evaluation procedure sanity-checks
	// that all the columns has the same number of rows. The value of this
	// field should not be accessed directly and the caller should instead
	// use getNumRow().
	numRow_ int

	// NewQuery is the query that is produced by the compilation, also named
	// the "Grail" query in the paper. The evaluation spans overall the
	// polynomials found in polys and the quotient.
	NewQuery query.UnivariateEval

	// Quotient is the column constructed as the quotient of all the
	// polynomials found in polys by their respective evaluation claims.
	Quotient ifaces.Column

	// LinCombCoeffLambda is the linear combination coefficient that is used to
	// accumulate the quotient.
	LinCombCoeffLambda, LinCombCoeffRho coin.Info

	// EvaluationPoint is the random evaluation point that is used to evaluate
	// the quotient and all the evaluation claims.
	EvaluationPoint coin.Info
}

// Compile applies the multipoint to singlepoint compilation pass over `comp`.
func Compile(options ...Option) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		compileMultipointToSinglepoint(comp, options)
	}
}

// compileMultipointToSinglepoint takes all the uncompiled multipoint to
// singlepoint queries and compile them using a quotient accumulation technique.
func compileMultipointToSinglepoint(comp *wizard.CompiledIOP, options []Option) *MultipointToSinglepointCompilation {

	var (
		ctx = &MultipointToSinglepointCompilation{
			Queries: getAndMarkAsCompiledQueries(comp),
		}
		polysByRound = sortPolynomialsByRoundAndName(ctx.Queries)
	)

	for _, op := range options {
		op(ctx)
	}

	if ctx.NumColumnProfileOpt != nil {
		polysByRound = extendPWithShadowColumns(comp, polysByRound, ctx.NumColumnProfileOpt)
	}

	ctx.Polys = slices.Concat(polysByRound...)

	ctx.LinCombCoeffLambda = comp.InsertCoin(
		ctx.getNumRound(),
		coin.Namef("MPTS_LINCOMB_COEFF_LAMBDA_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.LinCombCoeffRho = comp.InsertCoin(
		ctx.getNumRound(),
		coin.Namef("MPTS_LINCOMB_COEFF_RHO_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.Quotient = comp.InsertCommit(
		ctx.getNumRound(),
		ifaces.ColIDf("MPTS_QUOTIENT_%v", comp.SelfRecursionCount),
		ctx.getNumRow(),
	)

	ctx.EvaluationPoint = comp.InsertCoin(
		ctx.getNumRound()+1,
		coin.Namef("MPTS_EVALUATION_POINT_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.NewQuery = comp.InsertUnivariate(
		ctx.getNumRound()+1,
		ifaces.QueryIDf("MPTS_NEW_QUERY_%v", comp.SelfRecursionCount),
		append(ctx.Polys, ctx.Quotient),
	)

	ctx.EvalPointOfPolys, ctx.PolysOfEvalPoint = indexPolysAndPoints(ctx.Polys, ctx.Queries)

	comp.RegisterProverAction(ctx.getNumRound(), quotientAccumulation{ctx})
	comp.RegisterProverAction(ctx.getNumRound()+1, randomPointEvaluation{ctx})
	comp.RegisterVerifierAction(ctx.getNumRound()+1, verifierAction{ctx})

	return ctx
}

// getNumRow returns the number of rows in the columns that are compiled. The
// function panics if the number of rows is not the same for all the columns
// and sets the result in [numRow_] for memoization.
func (ctx *MultipointToSinglepointCompilation) getNumRow() int {

	if ctx.numRow_ > 0 {
		return ctx.numRow_
	}

	ctx.numRow_ = ctx.Polys[0].Size()

	for i := 1; i < len(ctx.Polys); i++ {
		ctx.numRow_ = max(ctx.numRow_, ctx.Polys[i].Size())
	}

	return ctx.numRow_
}

// getNumRound returns the number of rounds that are compiled. This is also
// the defintion round of the [LinCombCoeff] and [EvaluationPoint] in the
// context.
func (ctx *MultipointToSinglepointCompilation) getNumRound() int {
	// This relies on the fact that the polynomials are sorted by round. So
	// the last one has the highest round number.
	return ctx.Polys[len(ctx.Polys)-1].Round() + 1
}

// getAndMarkAsCompiledQueries returns all the queries that are relevant
// to the multipoint to singlepoint compilation and mark them as ignored.
func getAndMarkAsCompiledQueries(comp *wizard.CompiledIOP) []query.UnivariateEval {

	var (
		selectedQueries = []query.UnivariateEval{}
		allQueries      = comp.QueriesParams.AllUnignoredKeys()
	)

	for _, qName := range allQueries {

		q, isUnivariate := comp.QueriesParams.Data(qName).(query.UnivariateEval)
		if !isUnivariate {
			continue
		}

		comp.QueriesParams.MarkAsIgnored(qName)
		selectedQueries = append(selectedQueries, q)
	}

	return selectedQueries
}

// sortPolynomialsByRoundAndName returns the polynomials sorted by round
// and then by name. The result is a double list of polynomials. Each entry
// corresponds to a round, and the sublists are sorted by name.
func sortPolynomialsByRoundAndName(queries []query.UnivariateEval) [][]ifaces.Column {

	polysByRound := make([][]ifaces.Column, 0)

	for _, q := range queries {
		for _, poly := range q.Pols {
			round := poly.Round()
			polysByRound = utils.GrowSliceSize(polysByRound, round+1)
			polysByRound[round] = append(polysByRound[round], poly)
		}
	}

	// cmpNames is a helper function returning a comparison between two columns
	// by name in alphabetical order.
	cmpNames := func(a, b ifaces.Column) int {
		if a.GetColID() < b.GetColID() {
			return -1
		}
		if a.GetColID() > b.GetColID() {
			return 1
		}
		return 0
	}

	for round := range polysByRound {
		slices.SortFunc(polysByRound[round], cmpNames)
		polysByRound[round] = slices.Compact(polysByRound[round])
		polysByRound[round] = slices.Clip(polysByRound[round])
	}

	return polysByRound
}

// extendPWithShadowColumns adds shadow columns to each sublist of polynomials
// to match a given profile. The profile is a list of integers indicating a
// target number of polynomials to meet for each round. The positions of
// profile are to be understood as "starting from the non-empty" entries of
// p.
//
// For instance if
//
//	p = [[],[],[],[a, b, c],[d, e, f],[d],]
//	profile = [3, 4, 2]
//
// then the result will be
//
//	p = [[],[],[],[a, b, c], [d, e, f, shadow0], [d, e, f, shadow1],]
//
// where shadow0 and shadow1 are the shadow columns.
func extendPWithShadowColumns(comp *wizard.CompiledIOP, p [][]ifaces.Column, profile []int) [][]ifaces.Column {

	var (
		startingRoundMet = false
		startingRound    = 0
		numRow           int
	)

	for round, polysAtRound := range p {

		if len(polysAtRound) == 0 {
			continue
		}

		if !startingRoundMet {
			startingRound = round
			startingRoundMet = true
			numRow = polysAtRound[0].Size()
		}

		var (
			profileRound = profile[round-startingRound]
		)

		for i := len(polysAtRound); i < profileRound; i++ {
			newShadowCol := autoAssignedShadowRow(comp, numRow, round, i-len(polysAtRound))
			// Important: the appending has to be to "p[round]" and not polysAtRound
			// otherwise, the appending will be ineffective.
			p[round] = append(p[round], newShadowCol)
		}
	}

	return p
}

// indexPolysAndPoints indexes the polynomials and points in two lists.
// The first one indicates the queries affecting each polynomial and the
// second one indicates which polynomial are affected by each point.
//
// The returns are two lists of lists of integers referencing the positions
// in the inputs [polys] and [points].
func indexPolysAndPoints(polys []ifaces.Column, points []query.UnivariateEval) (evalPointOfPolys, polysOfEvalPoint [][]int) {

	evalPointOfPolys = make([][]int, len(polys))
	polysOfEvalPoint = make([][]int, len(points))
	polyNameToIndex := make(map[ifaces.ColID]int)

	for i := range polys {
		polyNameToIndex[polys[i].GetColID()] = i
	}

	for queryID, q := range points {
		for _, poly := range q.Pols {
			polyID := polyNameToIndex[poly.GetColID()]
			evalPointOfPolys[polyID] = append(evalPointOfPolys[polyID], queryID)
			polysOfEvalPoint[queryID] = append(polysOfEvalPoint[queryID], polyID)
		}
	}

	// evalPointOfPolys is built by iterating in-order over the queryIDs
	// by appending the queryIDs to each entry of evalPointOfPolys, we
	// can assume that the entries of evalPointOfPolys are sorted in
	// order. But this is not the case for polysOfEvalPoint as nothing
	// indicate they are sorted in [query.Pols] field.
	for i := range polysOfEvalPoint {
		slices.Sort(polysOfEvalPoint[i])
	}

	return evalPointOfPolys, polysOfEvalPoint
}
