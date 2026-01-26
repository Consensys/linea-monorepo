package globalcs

import (
	"reflect"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// DegreeReductionStep is the first step of the global constraint compilation
// process. It creates intermediate columns aimed at reducing the degree of the
// global constraints. The structure additionally implements the
// [wizard.ProverAction] package.
type DegreeReductionStep struct {
	// NewColumns are the intermediate columns returned by the step
	NewColumns []column.Natural
	// NewColumnsExpressions are the symbolic expressions corresponding to the
	// intermediate columns.
	NewColumnsExpressions []*symbolic.ExpressionBoard
	// DegreeReducedExpression is the list of expressions after degree reduction
	DegreeReducedExpression []*symbolic.Expression
	// DomainSize is the domain size of the global constraints
	DomainSize int
	// MaxRound is the maximum round of the global constraints
	MaxRound int
}

// DegreeReduce reduces the degree of the global constraints.
func DegreeReduce(comp *wizard.CompiledIOP, degreeBound int) *DegreeReductionStep {

	var (
		constraints, domainSize, maxRound = listAllGlobalConstraints(comp)
		exprs                             = make([]*symbolic.Expression, len(constraints))
	)

	for i, cs := range constraints {
		exprs[i] = getBoundCancelledExpression(cs)
	}

	if len(exprs) == 0 {
		return nil
	}

	var (
		degreeReducedExpr, elimExpr, newVariables = symbolic.ReduceDegreeOfExpressions(
			exprs,
			degreeBound*domainSize,
			GetDegree(domainSize),
			// ignore degrees smaller than 100, which will correspond to product
			// of the form (X-a)(X-b)(..)
			symbolic.DegreeReductionConfig{
				MinDegreeForCandidate: 3 * domainSize / 2,
				MinWeightForTerm:      100,
				NLast:                 15,
				MaxCandidatePerRound:  1 << 13,
			},
		)

		degRedStep = &DegreeReductionStep{
			NewColumnsExpressions:   make([]*symbolic.ExpressionBoard, len(newVariables)),
			NewColumns:              make([]column.Natural, len(newVariables)),
			DegreeReducedExpression: degreeReducedExpr,
			DomainSize:              domainSize,
			MaxRound:                maxRound,
		}
	)

	// Create the columns for the intermediate columns
	for i, v := range newVariables {

		var (
			newColumn = comp.InsertCommit(
				maxRound,
				ifaces.ColID(v.String()),
				domainSize,
				elimExpr[i].IsBase,
			).(column.Natural)

			board = elimExpr[i].Board()
		)

		board.Compile()
		degRedStep.NewColumns[i] = newColumn
		degRedStep.NewColumnsExpressions[i] = &board

		// substitute "v" for the newly created proper column meant to store
		// its values.
		for k := range degRedStep.DegreeReducedExpression {
			degRedStep.DegreeReducedExpression[k] = degRedStep.DegreeReducedExpression[k].ReconstructBottomUpSingleThreaded(
				func(e *symbolic.Expression, children []*symbolic.Expression) (new *symbolic.Expression) {
					if op, isVar := e.Operator.(symbolic.Variable); isVar && op.Metadata.String() == v.String() {
						return symbolic.NewVariable(newColumn)
					}
					return e.SameWithNewChildren(children)
				})
		}

		degRedStep.DegreeReducedExpression = append(
			degRedStep.DegreeReducedExpression,
			symbolic.Sub(newColumn, elimExpr[i]),
		)
	}

	return degRedStep
}

// Run implements the [wizard.ProverAction] interface. Namely, it assigns the
// NewColumns by evaluating the NewColumnsExpressions.
func (d *DegreeReductionStep) Run(run *wizard.ProverRuntime) {

	for i, expr := range d.NewColumnsExpressions {

		var (
			metadatas  = expr.ListVariableMetadata()
			evalInputs = make([]sv.SmartVector, len(metadatas))
		)

		for k, metadata := range metadatas {

			switch metadata := metadata.(type) {
			case ifaces.Column:
				evalInputs[k] = metadata.GetColAssignment(run)
			case coin.Info:
				if metadata.IsBase() {
					utils.Panic("unsupported, coins are always over field extensions")

				} else {
					evalInputs[k] = sv.NewConstantExt(run.GetRandomCoinFieldExt(metadata.Name), d.DomainSize)
				}
			case variables.X:
				evalInputs[k] = metadata.EvalCoset(d.DomainSize, 0, 1, false)
			case variables.PeriodicSample:
				evalInputs[k] = metadata.EvalCoset(d.DomainSize, 0, 1, false)
			case ifaces.Accessor:
				if metadata.IsBase() {
					evalInputs[k] = sv.NewConstant(metadata.GetVal(run), d.DomainSize)
				} else {
					evalInputs[k] = sv.NewConstantExt(metadata.GetValExt(run), d.DomainSize)
				}
			default:
				utils.Panic("Not a variable type %v", reflect.TypeOf(metadata))
			}
		}

		run.AssignColumn(d.NewColumns[i].GetColID(), expr.Evaluate(evalInputs))
	}

}

// listAllGlobalConstraints lists all the global constraints in the
// [wizard.CompiledIOP] that can be compiled by the current compilation pass.
//
// The function also initializes the value of ctx.DomainSize.
// And it marks all the returned global constraints as ignored to signify they
// are being compiled (and won't need to be compiled again).
func listAllGlobalConstraints(
	comp *wizard.CompiledIOP,
) (
	constraints []query.GlobalConstraint,
	domainSize int,
	maxRound int,
) {

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter only the global constraints, excluding all other type of queries
		cs, ok := comp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			// Not a global constraint
			continue
		}

		// For the first iteration, the domain size is unset so we need to initialize
		// it. This works because the domain size of a constraint cannot legally
		// be 0.
		if domainSize == 0 {
			domainSize = cs.DomainSize
		}

		// This enforces the precondition that all the global constraint must
		// share the same domain.
		if cs.DomainSize != domainSize {
			utils.Panic("At this point in the compilation process, we expect all constraints to have the same domain")
		}

		// Mark the constraint as ignored, so that it does not get compiled a
		// second time by a sub-sequent round of compilation.
		comp.QueriesNoParams.MarkAsIgnored(qName)
		constraints = append(constraints, cs)
		maxRound = max(maxRound, comp.QueriesNoParams.Round(qName))
	}

	return constraints, domainSize, maxRound
}
