package horner

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// projectionContext is a compilation artefact generated during the execution of
// the [InsertProjection] and which is used to instantiate the Horner query.
type projectionContext struct {
	// Xs are the coins used as X values in the Horner query that compiles the
	// projection queries.
	Xs []coin.Info
	// Query is the Horner query generated during the compilation of the projection
	// queries.
	Query query.Horner
}

// assignHornerQuery is a [wizard.ProverAction] that assigns the Horner query from
// the [projectionContext] to the [wizard.ProverRuntime]. The final value is zero
// and the N0 values are zero. The function additionally sanity-checks the values
// of the Horner query.
type assignHornerQuery struct {
	projectionContext
}

// checkHornerQuery result is a [wizard.VerifierAction] that can be used to check
// the value of a Horner query.
type checkHornerQuery struct {
	projectionContext
	skipped bool
}

// ProjectionToHorner is a compilation step that compiles [query.Projection] queries
// into [query.Horner] queries.
func ProjectionToHorner(comp *wizard.CompiledIOP) {

	round := 0
	parts := []query.HornerPart{}
	ctx := projectionContext{}

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non projection queries
		projection, ok := comp.QueriesNoParams.Data(qName).(query.Projection)
		if !ok {
			continue
		}

		comp.QueriesNoParams.MarkAsIgnored(qName)

		var (
			qRound     = comp.QueriesNoParams.Round(qName)
			widthA     = len(projection.Inp.ColumnsA)
			widthB     = len(projection.Inp.ColumnsB)
			numCols    = len(projection.Inp.ColumnsA[0])
			as         = make([]*sym.Expression, widthA)
			bs         = make([]*sym.Expression, widthB)
			selectorsA = make([]ifaces.Column, widthA)
			selectorsB = make([]ifaces.Column, widthB)
			gamma      coin.Info
			alpha      = comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_ALPHA"), coin.Field)
		)

		round = max(round, qRound+1)

		if numCols > 1 {
			gamma = comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_GAMMA"), coin.Field)
		}

		for i := 0; i < widthA; i++ {

			as[i] = sym.NewVariable(projection.Inp.ColumnsA[i][0])
			if numCols > 1 {
				as[i] = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnsA[i])
			}

			selectorsA[i] = projection.Inp.FiltersA[i]
		}

		for i := 0; i < widthB; i++ {

			bs[i] = sym.NewVariable(projection.Inp.ColumnsB[i][0])
			if numCols > 1 {
				bs[i] = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnsB[i])
			}

			selectorsB[i] = projection.Inp.FiltersB[i]
		}

		parts = append(
			parts,
			query.HornerPart{
				Coefficients: as,
				Selectors:    selectorsA,
				X:            accessors.NewFromCoin(alpha),
			},
			query.HornerPart{
				SignNegative: true,
				Coefficients: bs,
				Selectors:    selectorsB,
				X:            accessors.NewFromCoin(alpha),
			},
		)

		ctx.Xs = append(ctx.Xs, alpha, alpha)
	}

	if len(parts) == 0 {
		return
	}

	ctx.Query = comp.InsertHornerQuery(round, ifaces.QueryIDf("PROJECTION_TO_HORNER_%v", comp.SelfRecursionCount), parts)
	comp.RegisterProverAction(round, assignHornerQuery{ctx})
	comp.RegisterVerifierAction(round, &checkHornerQuery{projectionContext: ctx})
}

func (a assignHornerQuery) Run(run *wizard.ProverRuntime) {

	params := query.HornerParams{}

	for range a.Query.Parts {
		params.Parts = append(params.Parts, query.HornerParamsPart{
			N0: 0,
		})
	}

	params.SetResult(run, a.Query)

	if !params.FinalResult.IsZero() {
		utils.Panic("expected final result to be zero, but computed %v", params.FinalResult.String())
	}

	run.AssignHornerParams(a.Query.ID, params)
}

func (c *checkHornerQuery) Run(run wizard.Runtime) error {

	params := run.GetHornerParams(c.Query.ID)

	if !params.FinalResult.IsZero() {
		return fmt.Errorf("expected final result to be zero")
	}

	for _, p := range params.Parts {
		if p.N0 != 0 {
			return fmt.Errorf("expected N0 to be zero")
		}
	}

	return nil
}

func (c *checkHornerQuery) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	params := run.GetHornerParams(c.Query.ID)
	api.AssertIsEqual(params.FinalResult, 0)

	for _, p := range params.Parts {
		api.AssertIsEqual(p.N0, 0)
	}
}

func (c *checkHornerQuery) Skip() {
	c.skipped = true
}

func (c *checkHornerQuery) IsSkipped() bool {
	return c.skipped
}
