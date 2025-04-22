package mpts

import (
	"fmt"
	"slices"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
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
	//
	// The first round is a little special as the precomputed polynomials are
	// considered in their own round, placed at the beginning.
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

	// NumColumnProfilePrecomputed is as [NumColumnProfileOpt] but for
	// precomputed polynomials. The value of this field is only read if the
	// value of [NumColumnProfileOpt] is not nil. It serves the same purpose
	// as [NumColumnProfileOpt] as well.
	NumColumnProfilePrecomputed int

	// numRow is the number of rows in the columns that are compiled. The
	// value is lazily evaluated and the evaluation procedure sanity-checks
	// that all the columns has the same number of rows. The value of this
	// field should not be accessed directly and the caller should instead
	// use getNumRow().
	numRow int

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
		polysByRound, polyPrecomputed = sortPolynomialsByRoundAndName(comp, ctx.Queries)
	)

	for _, op := range options {
		op(ctx)
	}

	ctx.setMaxNumberOfRowsOf(slices.Concat(
		append(polysByRound, polyPrecomputed)...),
	)

	if ctx.NumColumnProfileOpt != nil {

		// startingRound is the first round that is not empty in polyRound
		// ignoring precomputed columns.
		startingRound := getStartingRound(comp, polysByRound)

		for round := startingRound; round < len(polysByRound); round++ {
			polysByRound[round] = extendPWithShadowColumns(comp, round,
				ctx.numRow, polysByRound[round], ctx.NumColumnProfileOpt[round-startingRound], false)
		}

		polyPrecomputed = extendPWithShadowColumns(comp, 0,
			ctx.numRow, polyPrecomputed, ctx.NumColumnProfilePrecomputed, true)
	}

	ctx.Polys = slices.Concat(append([][]ifaces.Column{polyPrecomputed}, polysByRound...)...)

	ctx.LinCombCoeffLambda = comp.InsertCoin(
		ctx.getNumRound(comp),
		coin.Namef("MPTS_LINCOMB_COEFF_LAMBDA_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.LinCombCoeffRho = comp.InsertCoin(
		ctx.getNumRound(comp),
		coin.Namef("MPTS_LINCOMB_COEFF_RHO_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.Quotient = comp.InsertCommit(
		ctx.getNumRound(comp),
		ifaces.ColIDf("MPTS_QUOTIENT_%v", comp.SelfRecursionCount),
		ctx.numRow,
	)

	ctx.EvaluationPoint = comp.InsertCoin(
		ctx.getNumRound(comp)+1,
		coin.Namef("MPTS_EVALUATION_POINT_%v", comp.SelfRecursionCount),
		coin.Field,
	)

	ctx.NewQuery = comp.InsertUnivariate(
		ctx.getNumRound(comp)+1,
		ifaces.QueryIDf("MPTS_NEW_QUERY_%v", comp.SelfRecursionCount),
		append(ctx.Polys, ctx.Quotient),
	)

	ctx.EvalPointOfPolys, ctx.PolysOfEvalPoint = indexPolysAndPoints(ctx.Polys, ctx.Queries)

	comp.RegisterProverAction(ctx.getNumRound(comp), quotientAccumulation{ctx})
	comp.RegisterProverAction(ctx.getNumRound(comp)+1, randomPointEvaluation{ctx})
	comp.RegisterVerifierAction(ctx.getNumRound(comp)+1, verifierAction{ctx})

	return ctx
}

// setMaxNumberOfRowsOf scans the list of columns and returns the largest size.
// and set it in the context.
func (ctx *MultipointToSinglepointCompilation) setMaxNumberOfRowsOf(columns []ifaces.Column) int {
	numRow := columns[0].Size()
	for i := 1; i < len(columns); i++ {
		numRow = max(numRow, columns[i].Size())
	}
	ctx.numRow = numRow
	return numRow
}

// getNumRow returns the number of rows and panics if the field is not set
// in the context.
func (ctx *MultipointToSinglepointCompilation) getNumRow() int {
	if ctx.numRow == 0 {
		utils.Panic("the number of rows is not set")
	}
	return ctx.numRow
}

// getNumRound returns the number of rounds that are compiled. This is also
// the defintion round of the [LinCombCoeff] and [EvaluationPoint] in the
// context.
//
// The function assumes that the number of queries to compile is not 0.
func (ctx *MultipointToSinglepointCompilation) getNumRound(comp *wizard.CompiledIOP) int {

	maxRound := -1
	for i := range ctx.Queries {
		round := comp.QueriesParams.Round(ctx.Queries[i].Name())
		if round > maxRound {
			maxRound = round
		}
	}

	if maxRound < 0 {
		panic("no queries")
	}

	// This relies on the fact that the polynomials are sorted by round. So
	// the last one has the highest round number.
	return maxRound + 1
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
//
// Precomputed polynomials can be considered at round -1 and are placed at
// the beginning.
func sortPolynomialsByRoundAndName(comp *wizard.CompiledIOP, queries []query.UnivariateEval) ([][]ifaces.Column, []ifaces.Column) {

	var (
		polysByRound    = make([][]ifaces.Column, 0)
		polyPrecomputed = make([]ifaces.Column, 0)
	)

	for _, q := range queries {
		for _, poly := range q.Pols {

			if _, isShf := poly.(column.Shifted); isShf {
				utils.Panic("shifted polys are not supported. Please, run the naturalization pass prior to calling the MPTS pass")
			}

			// This works assuming the input polys are [colum.Natural] columns.
			// Otherwise, the check would fail even if the column were a shifted
			// version of a precomputed column.
			if comp.Precomputed.Exists(poly.GetColID()) {
				polyPrecomputed = append(polyPrecomputed, poly)
				continue
			}

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

	// cleanSubList sorts and removes duplicates and empty entries from its
	// input list. The input slice is mutated and should not be used anymore.
	cleanSubList := func(s []ifaces.Column) []ifaces.Column {
		slices.SortFunc(s, cmpNames)
		s = slices.CompactFunc(s, func(a, b ifaces.Column) bool { return a.GetColID() == b.GetColID() })
		s = slices.Clip(s)
		return s
	}

	polyPrecomputed = cleanSubList(polyPrecomputed)
	for round := range polysByRound {
		polysByRound[round] = cleanSubList(polysByRound[round])
	}

	return polysByRound, polyPrecomputed
}

// extendPWithShadowColumns adds shadow columns to the given list of polynomials
// to match a given profile. The profile corresponds to a target number of columns
// to meet in "p". The function will ignore the verifiercol from the count.
func extendPWithShadowColumns(comp *wizard.CompiledIOP, round int, numRow int, p []ifaces.Column, profile int, precomputed bool) []ifaces.Column {

	if len(p) > profile {
		utils.Panic("the profile is too small for the given polynomials list")
	}

	numP := len(p)

	// This loop effective remove the verifiercol from consideration when evaluating
	// how many shadow columns are needed.
	for i := range p {
		_, isVcol := p[i].(verifiercol.VerifierCol)
		if isVcol {
			numP--
		}
	}

	for i := numP; i < profile; i++ {

		var newShadowCol ifaces.Column
		if precomputed {
			newShadowCol = precomputedShadowRow(comp, numRow, i)
		} else {
			newShadowCol = autoAssignedShadowRow(comp, numRow, round, i-numP)
		}

		p = append(p, newShadowCol)
	}

	return p
}

// getStartingRound returns the first position in s with a non-empty
// sublist. The function panics if all the sublists are empty of if s is an
// empty of nil list itself.
//
// The function will ignore the first round if it only contains precomputed
// columns. The function also ignores [VerifierCol].
func getStartingRound(comp *wizard.CompiledIOP, s [][]ifaces.Column) int {

	if len(s) == 0 {
		panic("empty list")
	}

	for j := range s[0] {

		_, isVcol := s[0][j].(verifiercol.VerifierCol)
		if isVcol {
			continue
		}

		if !comp.Precomputed.Exists(s[0][j].GetColID()) {
			fmt.Printf("found non precomputed column %s\n", s[0][j].GetColID())
			return 0
		}
	}

	for i := 1; i < len(s); i++ {
		if len(s[i]) > 0 {
			return i
		}
	}

	panic("all sublists are empty")
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
