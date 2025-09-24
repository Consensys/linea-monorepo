package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var _ ifaces.Accessor = &FromExprAccessor{}

// FromExprAccessor symbolizes a value derived from others values via a
// [symbolic.Expression] and implements [ifaces.Accessor].
type FromExprAccessor struct {
	// The expression represented by the accessor
	Expr *symbolic.Expression
	// The boarded expression
	Boarded symbolic.ExpressionBoard
	// An identifier to denote the expression. This will be used to evaluate
	// [ifaces.Accessor.String]
	ExprName string
	// The definition round of the expression
	ExprRound int
}

// NewFromExpression returns an [ifaces.Accessor] symbolizing the evaluation of a
// symbolic expression. The provided expression must be evaluable from verifier
// inputs only meaning in may contain only (accessors or coins) otherwise, calling
// [ifaces.Accessor.GetVal] may panic.
//
// This can be used if we want to use, not the
func NewFromExpression(expr *symbolic.Expression, exprName string) ifaces.Accessor {

	// This sanity-checks the expression to detect if it is valid to be used
	// as an accessor. We need to "board" the expression before we can access
	// its metadata. We also use this to detect the definition round of the
	// expression automatically.
	var (
		board     = expr.Board()
		metadata  = board.ListVariableMetadata()
		exprRound = 0
	)

	for _, m := range metadata {
		switch castedMetadata := m.(type) {
		case variables.X, variables.PeriodicSample:
			// this is not supported
			panic("variables are not supported")
		case ifaces.Accessor:
			// this is always fine because all coins are public
			exprRound = utils.Max(exprRound, castedMetadata.Round())
		case coin.Info:
			// this is always fine because all coins are public
			exprRound = utils.Max(exprRound, castedMetadata.Round)
		default:
			utils.Panic("unsupported type %T", m)
		}
	}

	return &FromExprAccessor{
		Expr:      expr,
		ExprName:  exprName,
		Boarded:   board,
		ExprRound: exprRound,
	}
}

// NewExponent constructs an accessor for the exponentiation of a another
// [ifaces.Accessor] by another constant `n`.
func NewExponent(a symbolic.Metadata, n int) ifaces.Accessor {
	return NewFromExpression(
		symbolic.Pow(a, n),
		fmt.Sprintf("EXP_%v_%v", a.String(), n),
	)
}

// Name implements [ifaces.Accessor]
func (e *FromExprAccessor) Name() string {
	return fmt.Sprintf("EXPR_AS_ACCESSOR_%v", e.ExprName)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (e *FromExprAccessor) String() string {
	return e.Name()
}

// GetVal implements [ifaces.Accessor]
func (e *FromExprAccessor) GetVal(run ifaces.Runtime) field.Element {

	metadata := e.Boarded.ListVariableMetadata()
	inputs := make([]smartvectors.SmartVector, len(metadata))

	for i, m := range metadata {
		switch castedMetadata := m.(type) {
		case ifaces.Accessor:
			x := castedMetadata.GetVal(run)
			inputs[i] = smartvectors.NewConstant(x, 1)
		case coin.Info:
			// this is always fine because all coins are public
			x := run.GetRandomCoinField(castedMetadata.Name)
			inputs[i] = smartvectors.NewConstant(x, 1)
		default:
			utils.Panic("unsupported type %T", m)
		}
	}

	return e.Boarded.Evaluate(inputs).Get(0)
}

func (e *FromExprAccessor) GetValBase(run ifaces.Runtime) (field.Element, error) {
	if e.IsBase() {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]smartvectors.SmartVector, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor:
				x, _ := castedMetadata.GetValBase(run)
				inputs[i] = smartvectors.NewConstant(x, 1)
			case coin.Info:
				// this is always fine because all coins are public
				x := run.GetRandomCoinField(castedMetadata.Name)
				inputs[i] = smartvectors.NewConstant(x, 1)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}

		return e.Boarded.Evaluate(inputs).GetBase(0)

	} else {
		return field.Zero(), fmt.Errorf("requested a base element from an accessor over field extensions")
	}
}

func (e *FromExprAccessor) GetValExt(run ifaces.Runtime) fext.Element {
	if e.IsBase() {
		res, _ := e.GetValBase(run)
		return fext.Lift(res)
	} else {
		// expression is over field extensions
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]smartvectors.SmartVector, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor:
				x := castedMetadata.GetValExt(run)
				inputs[i] = smartvectors.NewConstantExt(x, 1)
			case coin.Info:
				// this is always fine because all coins are public
				x := run.GetRandomCoinFieldExt(castedMetadata.Name)
				inputs[i] = smartvectors.NewConstantExt(x, 1)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}
		return e.Boarded.Evaluate(inputs).GetExt(0)
	}
}

// GetFrontendVariable implements [ifaces.Accessor]
func (e *FromExprAccessor) GetFrontendVariable(api frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {

	metadata := e.Boarded.ListVariableMetadata()
	inputs := make([]frontend.Variable, len(metadata))

	for i, m := range metadata {
		switch castedMetadata := m.(type) {
		case ifaces.Accessor:
			inputs[i] = castedMetadata.GetFrontendVariable(api, circ)
		case coin.Info:
			inputs[i] = circ.GetRandomCoinField(castedMetadata.Name)
		default:
			utils.Panic("unsupported type %T", m)
		}
	}

	return e.Boarded.GnarkEval(api, inputs)
}

func (e *FromExprAccessor) GetFrontendVariableBase(api frontend.API, circ ifaces.GnarkRuntime) (frontend.Variable, error) {
	if e.IsBase() {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]frontend.Variable, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor:
				inputs[i] = castedMetadata.GetFrontendVariable(api, circ)
			case coin.Info:
				inputs[i] = circ.GetRandomCoinField(castedMetadata.Name)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}

		return e.Boarded.GnarkEval(api, inputs), nil
	} else {
		return nil, fmt.Errorf("requested a base element from a col over field extensions")
	}
}

func (e *FromExprAccessor) GetFrontendVariableExt(api frontend.API, circ ifaces.GnarkRuntime) gnarkfext.Element {
	if e.IsBase() {
		baseElem, _ := e.GetFrontendVariableBase(api, circ)
		return gnarkfext.NewFromBase(baseElem)
	} else {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]gnarkfext.Element, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor:
				inputs[i] = castedMetadata.GetFrontendVariableExt(api, circ)
			case coin.Info:
				inputs[i] = circ.GetRandomCoinFieldExt(castedMetadata.Name)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}

		return e.Boarded.GnarkEvalExt(api, inputs)

	}
}

// AsVariable implements the [ifaces.Accessor] interface
func (e *FromExprAccessor) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(e)
}

// Round implements the [ifaces.Accessor] interface
func (e *FromExprAccessor) Round() int {
	return e.ExprRound
}

func (e *FromExprAccessor) IsBase() bool {
	return e.Expr.IsBase
}
