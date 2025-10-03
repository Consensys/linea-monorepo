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
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ProjectionContext is a compilation artefact generated during the execution of
// the [InsertProjection] and which is used to instantiate the Horner query.
type ProjectionContext[T zk.Element] struct {
	// Query is the Horner query generated during the compilation of the projection
	// queries.
	Query query.Horner[T]
}

// AssignHornerQuery is a [wizard.ProverAction] that assigns the Horner query from
// the [projectionContext] to the [wizard.ProverRuntime[T]]. The final value is zero
// and the N0 values are zero. The function additionally sanity-checks the values
// of the Horner query.
type AssignHornerQuery[T zk.Element] struct {
	ProjectionContext[T]
}

// CheckHornerQuery result is a [wizard.VerifierAction] that can be used to check
// the value of a Horner query.
type CheckHornerQuery[T zk.Element] struct {
	ProjectionContext[T]
	skipped bool `serde:"omit"`
}

// ProjectionToHorner is a compilation step that compiles [query.Projection] queries
// into [query.Horner] queries.
func ProjectionToHorner[T zk.Element](comp *wizard.CompiledIOP[T]) {

	round := 0
	parts := []query.HornerPart[T]{}
	ctx := ProjectionContext[T]{}

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non projection queries
		projection, ok := comp.QueriesNoParams.Data(qName).(query.Projection[T])
		if !ok {
			continue
		}

		comp.QueriesNoParams.MarkAsIgnored(qName)

		var (
			qRound     = comp.QueriesNoParams.Round(qName)
			widthA     = len(projection.Inp.ColumnsA)
			widthB     = len(projection.Inp.ColumnsB)
			numCols    = len(projection.Inp.ColumnsA[0])
			as         = make([]*sym.Expression[T], widthA)
			bs         = make([]*sym.Expression[T], widthB)
			selectorsA = make([]ifaces.Column[T], widthA)
			selectorsB = make([]ifaces.Column[T], widthB)
			gamma      coin.Info[T]
			alpha      = comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_ALPHA"), coin.FieldExt)
		)

		round = max(round, qRound+1)

		if numCols > 1 {
			gamma = comp.InsertCoin(qRound+1, coin.Name(qName+"_COIN_GAMMA"), coin.FieldExt)
		}

		for i := 0; i < widthA; i++ {

			as[widthA-i-1] = sym.NewVariable[T](projection.Inp.ColumnsA[i][0])
			if numCols > 1 {
				as[widthA-i-1] = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnsA[i])
			}

			// The reversal in the assignment is required due to the order
			// in which the [Horner] query iterates over the coefficient in
			// the multi-ary settings.
			selectorsA[widthA-i-1] = projection.Inp.FiltersA[i]
		}

		for i := 0; i < widthB; i++ {

			bs[widthB-i-1] = sym.NewVariable[T](projection.Inp.ColumnsB[i][0])
			if numCols > 1 {
				bs[widthB-i-1] = wizardutils.RandLinCombColSymbolic(gamma, projection.Inp.ColumnsB[i])
			}

			// The reversal in the assignment is required due to the order
			// in which the [Horner] query iterates over the coefficient in
			// the multi-ary settings.
			selectorsB[widthB-i-1] = projection.Inp.FiltersB[i]
		}

		parts = append(
			parts,
			query.HornerPart[T]{
				Name:         string(qName) + "_A",
				Coefficients: as,
				Selectors:    selectorsA,
				X:            accessors.NewFromCoin(alpha),
			},
			query.HornerPart[T]{
				Name:         string(qName) + "_B",
				SignNegative: true,
				Coefficients: bs,
				Selectors:    selectorsB,
				X:            accessors.NewFromCoin(alpha),
			},
		)
	}

	if len(parts) == 0 {
		return
	}

	ctx.Query = comp.InsertHornerQuery(round, ifaces.QueryIDf("PROJECTION_TO_HORNER_%v", comp.SelfRecursionCount), parts)
	comp.RegisterProverAction(round, AssignHornerQuery[T]{ctx})
	comp.RegisterVerifierAction(round, &CheckHornerQuery[T]{ProjectionContext: ctx})
}

func (a AssignHornerQuery[T]) Run(run *wizard.ProverRuntime[T]) {

	params := query.HornerParams[T]{}

	for range a.Query.Parts {
		params.Parts = append(params.Parts, query.HornerParamsPart[T]{
			N0: 0,
		})
	}

	params.SetResult(run, a.Query)

	if !params.FinalResult.IsZero() {
		utils.Panic("expected final result to be zero, but computed %v", params.FinalResult.String())
	}

	run.AssignHornerParams(a.Query.ID, params)
}

func (c *CheckHornerQuery[T]) Run(run wizard.Runtime[T]) error {

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

func (c *CheckHornerQuery[T]) RunGnark(api frontend.API, run wizard.GnarkRuntime[T]) {
	params := run.GetHornerParams(c.Query.ID)
	api.AssertIsEqual(params.FinalResult, 0)

	for _, p := range params.Parts {
		api.AssertIsEqual(p.N0, 0)
	}
}

func (c *CheckHornerQuery[T]) Skip() {
	c.skipped = true
}

func (c *CheckHornerQuery[T]) IsSkipped() bool {
	return c.skipped
}
