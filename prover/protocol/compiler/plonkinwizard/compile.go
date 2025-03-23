package plonkinwizard

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	plonkinternal "github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// context stores the compilation context for a single PlonkInWizard query
type context struct {
	// Q is the query handled by the current compilation context
	Q *query.PlonkInWizard
	// PlonkCtx is the compilation context relative to the plonk
	// circuit satisfaction.
	PlonkCtx *plonkinternal.CompilationCtx
	// SelOpenings are the local constraints responsible for
	// checking the activators are well-set w.r.t to the circuit mask
	SelOpenings []query.LocalOpening
	// CircuitMask is a precomputed binary column indicating with 1's
	// the rows where [query.PlonkInWizard.Data] corresponds to "potentially" actual public
	// inputs. It is used to ensure that the selector goes to zero on
	// the right position.
	CircuitMask ifaces.Column
	// StackedCircuitData is the column storing the concatenation of all the
	// public inputs.
	StackedCircuitData dedicated.StackedColumn
}

func Compile(comp *wizard.CompiledIOP) {

	qNames := comp.QueriesNoParams.AllKeys()
	for i := range qNames {

		// Skip if it was already compiled
		if comp.QueriesNoParams.IsIgnored(qNames[i]) {
			continue
		}

		q, isType := comp.QueriesNoParams.Data(qNames[i]).(*query.PlonkInWizard)

		// Skip if this is not a PlonkInWizard query
		if !isType {
			continue
		}

		compileQuery(comp, q)

		comp.QueriesNoParams.MarkAsIgnored(q.Name())
	}
}

func compileQuery(comp *wizard.CompiledIOP, q *query.PlonkInWizard) {

	plonkOptions := make([]plonkinternal.Option, len(q.PlonkOptions))
	for i := range plonkOptions {
		// Note: today, there is only one type of PlonkOption but in the
		// the future we might have more.
		plonkOptions[i] = plonkinternal.WithRangecheck(
			q.PlonkOptions[i].RangeCheckNbBits,
			q.PlonkOptions[i].RangeCheckNbLimbs,
			q.PlonkOptions[i].RangeCheckAddGateForRangeCheck,
		)
	}

	var (
		round          = max(q.Data.Round(), q.Selector.Round())
		maxNbInstances = q.GetMaxNbCircuitInstances()
		ctx            = &context{
			Q:        q,
			PlonkCtx: plonkinternal.PlonkCheck(comp, string(q.ID), round, q.Circuit, maxNbInstances, plonkOptions...),
		}
	)

	ctx.StackedCircuitData = dedicated.StackColumn(comp, ctx.PlonkCtx.Columns.TinyPI)

	// Since [StackedCircuitData] already stores the values of tinyPIs and that
	// the column (or rather a commitment to the column is included in the FS)
	// transcript. So not removing it would lead to a costly duplicate.
	for _, pi := range ctx.PlonkCtx.Columns.TinyPI {
		comp.Columns.ExcludeFromProverFS(pi.GetColID())
	}

	checkActivators(comp, ctx)
	checkPublicInputs(comp, ctx)

	comp.RegisterProverAction(round, &circAssignment{context: ctx})
}

// checkPublicInputs adds the constraints ensuring that the public inputs are
// consistent with the one of the PlonkCtx.
func checkPublicInputs(comp *wizard.CompiledIOP, ctx *context) {
	comp.InsertGlobal(
		ctx.Q.GetRound(),
		ifaces.QueryIDf("%v_PUBLIC_INPUTS", ctx.Q.ID),
		sym.Sub(ctx.Q.Data, ctx.StackedCircuitData.Column),
	)
}

// checkActivators adds the constraints checking the activators are well-set w.r.t
// to the circuit mask column. See [compilationCtx.Columns.Activators].
func checkActivators(comp *wizard.CompiledIOP, ctx *context) {

	var (
		openings   = make([]query.LocalOpening, ctx.Q.GetMaxNbCircuitInstances())
		mask       = ctx.Q.Selector
		offset     = utils.NextPowerOfTwo(ctx.Q.GetNbPublicInputs())
		activators = ctx.PlonkCtx.Columns.Activators
		round      = activators[0].Round()
	)

	for i := range openings {
		openings[i] = comp.InsertLocalOpening(
			round,
			ifaces.QueryIDf("%v_ACTIVATOR_LOCAL_OP_%v", ctx.Q.ID, i),
			column.Shift(mask, i*offset),
		)
	}

	comp.RegisterProverAction(ctx.Q.GetRound(), &assignSelOpening{context: ctx})
	comp.RegisterVerifierAction(ctx.Q.GetRound(), &checkActivatorAndMask{context: ctx})
	ctx.SelOpenings = openings
}
