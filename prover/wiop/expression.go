package wiop

import (
	"fmt"
	"sync"

	field "github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
)

// Expression is the interface satisfied by all symbolic arithmetic expressions
// in the framework. An expression is a node in an expression AST that
// evaluates to either a single field element or a vector of field elements at
// runtime.
//
// Not all methods are valid on every implementation. Size, IsSized, and
// EvaluateVector may only be called when IsMultiValued() is true.
// EvaluateSingle may only be called when IsMultiValued() is false.
// Implementations signal violations of these preconditions with a panic.
//
// Degree panics on non-polynomial expressions (those containing division or
// inversion).
type Expression interface {
	// IsExtension reports whether this expression involves any column or cell
	// that is evaluated over an extended domain.
	IsExtension() bool
	// IsMultiValued reports whether this expression evaluates to a vector.
	// If false, the expression is scalar.
	IsMultiValued() bool
	// Degree returns the polynomial degree of the expression.
	// Panics if the expression contains a non-polynomial operation (Div,
	// Inverse).
	Degree() int
	// Size returns the length of the vector produced by this expression.
	// Precondition: IsMultiValued() must be true; panics otherwise.
	Size() int
	// IsSized reports whether the vector size of this expression is known.
	// Precondition: IsMultiValued() must be true; panics otherwise.
	IsSized() bool
	// EvaluateVector evaluates this expression against the given runtime and
	// returns the resulting vector.
	// Precondition: IsMultiValued() must be true; panics otherwise.
	EvaluateVector(Runtime) ConcreteVector
	// EvaluateSingle evaluates this expression against the given runtime and
	// returns the resulting scalar.
	// Precondition: IsMultiValued() must be false; panics otherwise.
	EvaluateSingle(Runtime) ConcreteField
	// Module returns the Module whose columns appear in this expression, or
	// nil if the expression contains no column reference. An expression may
	// reference columns from at most one module; mixing columns from different
	// modules is undefined behaviour and callers are expected to validate this
	// before constructing composite expressions.
	Module() *Module
	// Visibility returns the most restrictive [Visibility] level of any leaf
	// in this expression. The ordering from least to most restrictive is:
	// Public < Oracle < Internal. A result of VisibilityInternal means the
	// expression cannot appear in any active query.
	Visibility() Visibility
}

// VectorPromise is the sub-interface of [Expression] satisfied by all
// vector-valued symbolic objects (e.g. [ColumnView]). It carries the same
// method set as [Expression]; the distinct type allows function signatures to
// declare that they require a vector operand, and enables type assertions to
// distinguish vector from scalar expressions.
type VectorPromise interface {
	Expression
}

// FieldPromise is the sub-interface of [Expression] satisfied by all
// scalar-valued symbolic objects (e.g. [Cell]). It carries the same method
// set as [Expression]; the distinct type allows function signatures to declare
// that they require a scalar operand.
type FieldPromise interface {
	Expression
}

// ArithmeticOperator identifies the arithmetic operation performed by an
// [ArithmeticOperation] node.
type ArithmeticOperator int

const (
	ArithmeticOperatorAdd     ArithmeticOperator = iota // a + b  (binary, linear)
	ArithmeticOperatorMul                               // a * b  (binary, product)
	ArithmeticOperatorSub                               // a - b  (binary, linear)
	ArithmeticOperatorDiv                               // a / b  (binary, non-polynomial)
	ArithmeticOperatorDouble                            // 2 * a  (unary, linear)
	ArithmeticOperatorSquare                            // a * a  (unary, product)
	ArithmeticOperatorNegate                            // -a     (unary, linear)
	ArithmeticOperatorInverse                           // 1 / a  (unary, non-polynomial)
)

// arity returns the required number of operands for op. Panics on an unknown
// operator.
func (op ArithmeticOperator) arity() int {
	switch op {
	case ArithmeticOperatorAdd, ArithmeticOperatorMul,
		ArithmeticOperatorSub, ArithmeticOperatorDiv:
		return 2
	case ArithmeticOperatorDouble, ArithmeticOperatorSquare,
		ArithmeticOperatorNegate, ArithmeticOperatorInverse:
		return 1
	default:
		panic(fmt.Sprintf("wiop: unknown ArithmeticOperator %d", int(op)))
	}
}

// combineDegree returns the degree of the expression formed by applying op to
// operands whose degrees are given by operandDegrees.
//
// Panics for non-polynomial operators (Div, Inverse), since degree is
// undefined for non-polynomial expressions.
func (op ArithmeticOperator) combineDegree(operandDegrees []int) int {
	switch op {
	case ArithmeticOperatorAdd, ArithmeticOperatorSub:
		return max(operandDegrees[0], operandDegrees[1])
	case ArithmeticOperatorDouble, ArithmeticOperatorNegate:
		return operandDegrees[0]
	case ArithmeticOperatorMul:
		return operandDegrees[0] + operandDegrees[1]
	case ArithmeticOperatorSquare:
		return 2 * operandDegrees[0]
	case ArithmeticOperatorDiv, ArithmeticOperatorInverse:
		panic(fmt.Sprintf("wiop: Degree() called on non-polynomial expression (%v)", op))
	default:
		panic(fmt.Sprintf("wiop: unknown ArithmeticOperator %d", int(op)))
	}
}

// String implements [fmt.Stringer].
func (op ArithmeticOperator) String() string {
	switch op {
	case ArithmeticOperatorAdd:
		return "Add"
	case ArithmeticOperatorMul:
		return "Mul"
	case ArithmeticOperatorSub:
		return "Sub"
	case ArithmeticOperatorDiv:
		return "Div"
	case ArithmeticOperatorDouble:
		return "Double"
	case ArithmeticOperatorSquare:
		return "Square"
	case ArithmeticOperatorNegate:
		return "Negate"
	case ArithmeticOperatorInverse:
		return "Inverse"
	default:
		return fmt.Sprintf("ArithmeticOperator(%d)", int(op))
	}
}

// ArithmeticOperation is an [Expression] node that applies an
// [ArithmeticOperator] to one or two sub-expressions.
//
// Size, IsSized, and EvaluateVector may only be called when IsMultiValued()
// is true. EvaluateSingle may only be called when IsMultiValued() is false.
type ArithmeticOperation struct {
	Operator ArithmeticOperator
	Operands []Expression
	// isMultiValuedCached caches the result of the IsMultiValued traversal.
	// A nil pointer means the value has not been computed yet.
	isMultiValuedCached *bool
	// once ensures the expression is compiled exactly once, even under
	// concurrent calls to EvaluateVector.
	once sync.Once
	// prog is the compiled bytecode representation of this subtree. It is
	// nil until the first call to EvaluateVector.
	prog *compiledProgram
}

// NewArithmeticOperation constructs an ArithmeticOperation, enforcing the
// arity contract of the given operator. Panics if the operand count is wrong
// or any operand is nil.
func NewArithmeticOperation(op ArithmeticOperator, operands ...Expression) *ArithmeticOperation {
	want := op.arity()
	if len(operands) != want {
		panic(fmt.Sprintf("wiop: %v requires %d operand(s), got %d", op, want, len(operands)))
	}
	for i, o := range operands {
		if o == nil {
			panic(fmt.Sprintf("wiop: operand %d of %v is nil", i, op))
		}
	}
	return &ArithmeticOperation{Operator: op, Operands: operands}
}

// IsExtension implements [Expression]. Returns true if any operand involves
// an extended-domain column or cell.
func (a *ArithmeticOperation) IsExtension() bool {
	for _, o := range a.Operands {
		if o.IsExtension() {
			return true
		}
	}
	return false
}

// IsMultiValued implements [Expression]. Returns true if any operand is
// vector-valued. The result is computed once and cached.
func (a *ArithmeticOperation) IsMultiValued() bool {
	if a.isMultiValuedCached != nil {
		return *a.isMultiValuedCached
	}
	result := false
	for _, o := range a.Operands {
		if o.IsMultiValued() {
			result = true
			break
		}
	}
	a.isMultiValuedCached = &result
	return result
}

// Degree implements [Expression]. Combines the degrees of the operands using
// the operator's own degree-combination rule. Panics for non-polynomial
// operators (Div, Inverse).
func (a *ArithmeticOperation) Degree() int {
	degrees := make([]int, len(a.Operands))
	for i, o := range a.Operands {
		degrees[i] = o.Degree()
	}
	return a.Operator.combineDegree(degrees)
}

// Size implements [Expression]. Returns the size of the first vector-valued
// operand. Panics if IsMultiValued() is false.
func (a *ArithmeticOperation) Size() int {
	if !a.IsMultiValued() {
		panic("wiop: Size() called on a scalar ArithmeticOperation; check IsMultiValued() first")
	}
	for _, o := range a.Operands {
		if o.IsMultiValued() {
			return o.Size()
		}
	}
	panic("unreachable")
}

// IsSized implements [Expression]. Returns true if all vector-valued operands
// are sized. Panics if IsMultiValued() is false.
func (a *ArithmeticOperation) IsSized() bool {
	if !a.IsMultiValued() {
		panic("wiop: IsSized() called on a scalar ArithmeticOperation; check IsMultiValued() first")
	}
	for _, o := range a.Operands {
		if o.IsMultiValued() && !o.IsSized() {
			return false
		}
	}
	return true
}

// EvaluateVector implements [Expression].
// Panics if IsMultiValued() is false.
//
// On the first call the expression subtree is compiled into a [compiledProgram]
// and cached. Subsequent calls reuse the compiled program directly.
func (a *ArithmeticOperation) EvaluateVector(rt Runtime) ConcreteVector {
	if !a.IsMultiValued() {
		panic("wiop: EvaluateVector() called on a scalar ArithmeticOperation; check IsMultiValued() first")
	}
	a.once.Do(func() { a.prog = compileExpr(a) })
	result := a.prog.evaluateVector(rt)
	return ConcreteVector{Plain: []field.FieldVec{result}, promise: a}
}

// EvaluateSingle implements [Expression].
// Panics if IsMultiValued() is true.
//
// TODO: Implement once Runtime is defined.
func (a *ArithmeticOperation) EvaluateSingle(_ Runtime) ConcreteField {
	if a.IsMultiValued() {
		panic("wiop: EvaluateSingle() called on a vector ArithmeticOperation; check IsMultiValued() first")
	}
	panic("wiop: ArithmeticOperation.EvaluateSingle not yet implemented")
}

// Module implements [Expression]. Returns the module of the first
// vector-valued operand, or nil if all operands are scalar. All vector-valued
// operands are expected to share the same module; this invariant is the
// caller's responsibility when constructing the expression.
func (a *ArithmeticOperation) Module() *Module {
	for _, o := range a.Operands {
		if o.IsMultiValued() {
			return o.Module()
		}
	}
	return nil
}

// Visibility implements [Expression]. Returns the most restrictive visibility
// level among all operands: the minimum of their Visibility values (Internal
// < Oracle < Public). This reflects that an expression involving any internal
// leaf cannot appear in an active query.
func (a *ArithmeticOperation) Visibility() Visibility {
	best := VisibilityPublic
	for _, o := range a.Operands {
		if v := o.Visibility(); v < best {
			best = v
		}
	}
	return best
}
