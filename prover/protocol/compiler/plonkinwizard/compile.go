package plonkinwizard

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	plonk "github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard/internal/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
)

// context stores the compilation context for a single PlonkInWizard query
type context struct {
	// Q is the query handled by the current compilation context
	Q *query.PlonkInWizard
	// PlonkCtx is the compilation context relative to the plonk
	// circuit satisfaction.
	PlonkCtx *plonk.CompilationCtx
	// SelOpenings are the local constraints responsible for
	// checking the activators are well-set w.r.t to the circuit mask
	SelOpenings []query.LocalOpening
	// CircuitMask is a precomputed binary column indicating with 1's
	// the rows where [Data] corresponds to "potentially" actual public
	// inputs. It is used to ensure that the selector goes to zero on
	// the right position.
	CircuitMask ifaces.Column
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
	}
}

func compileQuery(comp *wizard.CompiledIOP, q *query.PlonkInWizard) {

	var (
		round          = max(q.Data.Round(), q.Selector.Round())
		maxNbInstances = q.GetMaxNbCircuitInstances()
		ctx            = &context{
			Q:        q,
			PlonkCtx: plonk.PlonkCheck(comp, string(q.ID), round, q.Circuit, maxNbInstances),
		}
	)

	checkSelectorAndData(comp, ctx)
	checkActivators(comp, ctx)
	comp.RegisterProverAction(round, &circAssignment{context: *ctx})
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

	comp.RegisterProverAction(ctx.Q.GetRound(), &assignSelOpening{context: *ctx})
	comp.RegisterVerifierAction(ctx.Q.GetRound(), &checkActivatorAndMask{context: *ctx})
	ctx.SelOpenings = openings
}

// checkSelectorAndData adds the constraints ensuring that the selector and the data
// column are well-set w.r.t. to the circuit mask column.
func checkSelectorAndData(comp *wizard.CompiledIOP, ctx *context) {

	var (
		round       = ctx.Q.GetRound()
		nbPub       = ctx.Q.GetNbPublicInputs()
		nbPubPadded = utils.NextPowerOfTwo(nbPub)
		fullSize    = ctx.Q.Data.Size()
		circMaskVal = make([]field.Element, ctx.Q.Data.Size())
	)

	for i := 0; i < fullSize; i += nbPubPadded {
		for k := 0; k < nbPub; k++ {
			circMaskVal[i+k] = field.One()
		}
	}

	ctx.CircuitMask = comp.InsertPrecomputed(
		ifaces.ColIDf("%v_CIRCMASK", ctx.Q.ID),
		smartvectors.NewRegular(circMaskVal),
	)

	commonconstraints.MustBeActivationColumns(comp, ctx.Q.Selector)

	// This query ensures that mask[i]=0 => data[i]=0
	comp.InsertGlobal(
		round,
		ifaces.QueryIDf("%v_DATA_IS_ZERO_WHEN_MASK_IS_ZERO", ctx.Q.ID),
		sym.Sub(
			ctx.Q.Data,
			sym.Mul(ctx.Q.Data, ctx.CircuitMask),
		),
	)

	// This query ensures that sel[i] - sel[i+1] == 1 => mask[i] - mask[i+1] == 1
	// Note, that since sel is constrained to be an activation column the difference
	// is already constrained to never be "-1", this allows simplifying the constraint
	// as follows. The constraint only works if the number of public inputs is not
	// exactly a power of two but in this case, the implemented approach does not
	// work.
	comp.InsertGlobal(
		round,
		ifaces.QueryIDf("%v_MASK_DEC_WHEN_SEL_DEC", ctx.Q.ID),
		sym.Mul(
			sym.Sub(ctx.Q.Selector, column.Shift(ctx.Q.Selector, 1)),
			sym.Sub(ctx.CircuitMask, column.Shift(ctx.CircuitMask, 1), 1),
		),
	)

}
