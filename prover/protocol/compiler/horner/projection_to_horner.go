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
	"github.com/consensys/linea-monorepo/prover/symbolic"
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

		qRound := comp.QueriesNoParams.Round(qName)
		round = max(round, qRound+1)

		var (
			alpha = comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_ALPHA"), coin.Field)
			a     = symbolic.NewVariable(projection.Inp.ColumnA[0])
			b     = symbolic.NewVariable(projection.Inp.ColumnB[0])
		)

		if len(projection.Inp.ColumnA) > 1 {
			gamma := comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_GAMMA"), coin.Field)
			a = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnA)
			b = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnB)
		}

		parts = append(
			parts,
			query.HornerPart{
				Coefficient: a,
				Selector:    projection.Inp.FilterA,
				X:           accessors.NewFromCoin(alpha),
			},
			query.HornerPart{
				SignNegative: true,
				Coefficient:  b,
				Selector:     projection.Inp.FilterB,
				X:            accessors.NewFromCoin(alpha),
			},
		)

		ctx.Xs = append(ctx.Xs, alpha, alpha)
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
