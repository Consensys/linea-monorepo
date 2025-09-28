package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// FromExprAccessor[T] symbolizes a value derived from others values via a
// [symbolic.Expression[T]] and implements [ifaces.Accessor[T]].
type FromExprAccessor[T zk.Element] struct {
	// The expression represented by the accessor
	Expr *symbolic.Expression[T]
	// The boarded expression
	Boarded symbolic.ExpressionBoard[T]
	// An identifier to denote the expression. This will be used to evaluate
	// [ifaces.Accessor[T].String]
	ExprName string
	// The definition round of the expression
	ExprRound int
}

// NewFromExpression[T] returns an [ifaces.Accessor[T]] symbolizing the evaluation of a
// symbolic expression. The provided expression must be evaluable from verifier
// inputs only meaning in may contain only (accessors or coins) otherwise, calling
// [ifaces.Accessor[T].GetVal] may panic.
//
// This can be used if we want to use, not the
func NewFromExpression[T zk.Element](expr *symbolic.Expression[T], exprName string) ifaces.Accessor[T] {

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
		case variables.X[T], variables.PeriodicSample[T]:
			// this is not supported
			panic("variables are not supported")
		case ifaces.Accessor[T]:
			// this is always fine because all coins are public
			exprRound = utils.Max(exprRound, castedMetadata.Round())
		case coin.Info:
			// this is always fine because all coins are public
			exprRound = utils.Max(exprRound, castedMetadata.Round)
		default:
			utils.Panic("unsupported type %T", m)
		}
	}

	return &FromExprAccessor[T]{
		Expr:      expr,
		ExprName:  exprName,
		Boarded:   board,
		ExprRound: exprRound,
	}
}

// NewExponent constructs an accessor for the exponentiation of a another
// [ifaces.Accessor[T]] by another constant `n`.
func NewExponent[T zk.Element](a symbolic.Metadata, n int) ifaces.Accessor[T] {
	return NewFromExpression[T](
		symbolic.Pow[T](a, n),
		fmt.Sprintf("EXP_%v_%v", a.String(), n),
	)
}

// Name implements [ifaces.Accessor[T]]
func (e *FromExprAccessor[T]) Name() string {
	return fmt.Sprintf("EXPR_AS_ACCESSOR_%v", e.ExprName)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (e *FromExprAccessor[T]) String() string {
	return e.Name()
}

// GetVal implements [ifaces.Accessor[T]]
func (e *FromExprAccessor[T]) GetVal(run ifaces.Runtime) field.Element {

	metadata := e.Boarded.ListVariableMetadata()
	inputs := make([]smartvectors.SmartVector, len(metadata))

	for i, m := range metadata {
		switch castedMetadata := m.(type) {
		case ifaces.Accessor[T]:
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

func (e *FromExprAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	if e.IsBase() {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]smartvectors.SmartVector, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor[T]:
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

func (e *FromExprAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	if e.IsBase() {
		res, _ := e.GetValBase(run)
		return fext.Lift(res)
	} else {
		// expression is over field extensions
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]smartvectors.SmartVector, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor[T]:
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

// GetFrontendVariable implements [ifaces.Accessor[T]]
func (e *FromExprAccessor[T]) GetFrontendVariable(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {

	metadata := e.Boarded.ListVariableMetadata()
	inputs := make([]T, len(metadata))

	for i, m := range metadata {
		switch castedMetadata := m.(type) {
		case ifaces.Accessor[T]:
			inputs[i] = castedMetadata.GetFrontendVariable(api, circ)
		case coin.Info:
			inputs[i] = circ.GetRandomCoinField(castedMetadata.Name)
		default:
			utils.Panic("unsupported type %T", m)
		}
	}

	return e.Boarded.GnarkEval(api.GnarkAPI(), inputs)
}

func (e *FromExprAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) (T, error) {
	if e.IsBase() {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]T, len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor[T]:
				inputs[i] = castedMetadata.GetFrontendVariable(api, circ)
			case coin.Info:
				inputs[i] = circ.GetRandomCoinField(castedMetadata.Name)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}

		return e.Boarded.GnarkEval(api.GnarkAPI(), inputs), nil
	} else {
		var a T
		return a, fmt.Errorf("requested a base element from a col over field extensions")
	}
}

func (e *FromExprAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {

	e4Api, err := gnarkfext.NewExt4[T](api.GnarkAPI())
	if err != nil {
		panic(err)
	}

	if e.IsBase() {
		baseElem, _ := e.GetFrontendVariableBase(api, circ)
		return *e4Api.NewFromBase(baseElem)
	} else {
		metadata := e.Boarded.ListVariableMetadata()
		inputs := make([]gnarkfext.E4Gen[T], len(metadata))

		for i, m := range metadata {
			switch castedMetadata := m.(type) {
			case ifaces.Accessor[T]:
				inputs[i] = castedMetadata.GetFrontendVariableExt(api, circ)
			case coin.Info:
				inputs[i] = circ.GetRandomCoinFieldExt(castedMetadata.Name)
			default:
				utils.Panic("unsupported type %T", m)
			}
		}

		return e.Boarded.GnarkEvalExt(api.GnarkAPI(), inputs)

	}
}

// AsVariable implements the [ifaces.Accessor[T]] interface
func (e *FromExprAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](e)
}

// Round implements the [ifaces.Accessor[T]] interface
func (e *FromExprAccessor[T]) Round() int {
	return e.ExprRound
}

func (e *FromExprAccessor[T]) IsBase() bool {
	return e.IsBase()
}
