package symbolic

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// anchoredExpression[T] represents symbolic expression pinned into an overarching
// [ExpressionBoard] expression.
type anchoredExpression[T zk.Element] struct {
	Board  *ExpressionBoard[T]
	ESHash fext.GenericFieldElem
}

// Expression[T] represents a symbolic arithmetic expression. Expression[T] can be
// built using the [Add], [Mul], [Sub] etc... package. However, they can only
// represent polynomial expressions: inversion is not supported.
//
// Expressions is structured like a tree where each node performs an elementary
// operation.
//
// Expression[T] cannot be directly evaluated with an assignment. To do that, the
// expression must be first converted into a [BoardedExpression] using the
// [Expression.Board] method.
type Expression[T zk.Element] struct {
	// ESHash stands for Expression[T] Sensitive Hash and is an identifier for
	// the expression. Specifically, it represents what the expression computes
	// rather than how the expression is structured.
	//
	// For instance, the expressions "a * 2" and "a + a" have the same ESHash,
	// the user should not modify this value directly.
	//
	// It is computed with the following rules:
	//
	//		- ESHash(a + b) = ESHash(a) + ESHash(b)
	// 		- ESHash(a * b) = ESHash(a) * ESHash(b)
	// 		- ESHash(c: Constant) = c
	// 		- ESHash(a: variable) = H(a.String())
	ESHash fext.GenericFieldElem
	// Children stores the list of all the sub-expressions the current
	// Expression[T] uses as operands.
	Children []*Expression[T]
	// Operator stores information relative to operation that the current
	// Expression[T] performs on its inputs.
	Operator Operator[T]
	IsBase   bool
}

// Operator specifies an elementary operation a node of an [Expression] performs
// (Add, Mul, Sub ...)
type Operator[T zk.Element] interface {
	// Evaluate returns an evaluation of the operator from a list of assignments:
	// one for each operand (children) of the expression.
	Evaluate([]sv.SmartVector) sv.SmartVector
	// EvaluateExt returns an evaluation of the operator from a list of assignments:
	// one for each operand (children) of the expression.
	EvaluateExt([]sv.SmartVector) sv.SmartVector
	EvaluateMixed([]sv.SmartVector) sv.SmartVector
	// Validate performs a sanity-check of the expression the Operator belongs
	// to.
	Validate(e *Expression[T]) error
	// Returns the polynomial degree of the expression.
	Degree([]int) int
	// GnarkEval returns an evaluation of the operator in a gnark circuit.
	GnarkEval(frontend.API, []T) T
	GnarkEvalExt(frontend.API, []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T]
}

type OperatorWithResult interface {
	EvaluateExtResult(result sv.SmartVector, inputs []sv.SmartVector)
}

// Board pins down the expression into an ExpressionBoard. This converts the
// Expression[T] into a DAG and runs a topological sorting algorithm over the
// nodes of the expression. This has the effect of removing the duplicates
// nodes and making the expression more efficient to evaluate.
func (f *Expression[T]) Board() ExpressionBoard[T] {
	board := emptyBoard[T]()
	f.anchor(&board)
	return board
}

// anchor pins down the Expression[T] onto an ExpressionBoard.
func (f *Expression[T]) anchor(b *ExpressionBoard[T]) anchoredExpression[T] {

	/*
		Check if the expression is a duplicate of another expression bearing
		the same GenericFieldElem and
	*/
	if _, ok := b.ESHashesToPos[f.ESHash]; ok {
		return anchoredExpression[T]{Board: b, ESHash: f.ESHash}
	}

	/*
		Recurse the call in all the children to ensure all
		subexpressions are anchored. And get their levels
	*/
	maxChildrenLevel := 0
	childrenIDs := []nodeID{}
	for _, child := range f.Children {
		_ = child.anchor(b)
		childID, ok := b.ESHashesToPos[child.ESHash]
		if !ok {
			utils.Panic("Children not found in expr")
		}
		maxChildrenLevel = utils.Max(maxChildrenLevel, childID.level())
		childrenIDs = append(childrenIDs, childID)
	}

	newLevel := maxChildrenLevel + 1
	if len(f.Children) == 0 {
		// Then it is a leaf node and the level should be 0
		newLevel = 0
	}

	/*
		Computes the new ID of the node by adding a new one into the list.
		We extend the outer-list of Nodes if necessary.
	*/
	if len(b.Nodes) <= newLevel {
		b.Nodes = append(b.Nodes, []Node[T]{})
	}
	NewNodeID := newNodeID(newLevel, len(b.Nodes[newLevel]))

	newNode := Node[T]{
		ESHash:   f.ESHash,
		Parents:  []nodeID{},
		Children: childrenIDs,
		Operator: f.Operator,
	}

	/*
		Registers this NewNodeID in the children as a parents and in the
		`ESHashToPosition` registry and in the children as a new `Parent`.
		Also, add it to the list of nodes at the position corresponding to
		its new NodeID
	*/
	b.ESHashesToPos[f.ESHash] = NewNodeID
	for _, childID := range childrenIDs {
		b.getNode(childID).addParent(NewNodeID)
	}
	b.Nodes[NewNodeID.level()] = append(b.Nodes[NewNodeID.level()], newNode)

	/*
		And returns the new Anchored expression
	*/
	return anchoredExpression[T]{
		ESHash: f.ESHash,
		Board:  b,
	}
}

// Validate operates a list of sanity-checks over the current expression to
// assess its well-formedness. It returns an error if the check fails.
func (e *Expression[T]) Validate() error {
	if e.IsBase {
		return e.ValidateBase()
	} else {
		return e.ValidateExt()
	}
}

func (e *Expression[T]) ValidateExt() error {

	eshashes := make([]sv.SmartVector, len(e.Children))
	for i := range e.Children {
		eshashes[i] = sv.NewConstantExt(e.Children[i].ESHash.GetExt(), 1)
	}

	if len(e.Children) > 0 {
		// The cast back to sv.Constant is not functionally important but is an
		// easy sanity check.
		expectedESH := e.Operator.EvaluateExt(eshashes).(*sv.ConstantExt).GetExt(0)

		if !e.ESHash.IsEqualExt(&expectedESH) {
			return fmt.Errorf("esh mismatch %v %v", expectedESH.String(), e.ESHash.String())
		}
	}

	// Operator specific validation
	if err := e.Operator.Validate(e); err != nil {
		return err
	}

	// Validate the children recursively
	for i := range e.Children {
		if err := e.Children[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Expression[T]) ValidateBase() error {

	eshashes := make([]sv.SmartVector, len(e.Children))
	for i := range e.Children {
		baseESHash, _ := e.Children[i].ESHash.GetBase()
		eshashes[i] = sv.NewConstant(baseESHash, 1)
	}

	if len(e.Children) > 0 {
		// The cast back to sv.Constant is not functionally important but is an
		// easy sanity check.
		expectedESH, _ := e.Operator.Evaluate(eshashes).(*sv.Constant).GetBase(0)
		expressionBaseESHash, _ := e.ESHash.GetBase()
		if expectedESH != expressionBaseESHash {
			return fmt.Errorf("esh mismatch %v %v", expectedESH.String(), e.ESHash.String())
		}
	}

	// Operator specific validation
	if err := e.Operator.Validate(e); err != nil {
		return err
	}

	// Validate the children recursively
	for i := range e.Children {
		if err := e.Children[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Same as [Expression.Validate] but panics on error.
func (e *Expression[T]) AssertValid() {
	if err := e.Validate(); err != nil {
		panic(err)
	}
}

// Replay constructs an analogous expression. Replacing all the [Variable] by
// those given in the translation map.
func (e *Expression[T]) Replay(translationMap collection.Mapping[string, *Expression[T]]) (res *Expression[T]) {

	switch op := e.Operator.(type) {
	// Constant
	case Constant[T]:
		return NewConstant[T](op.Val)
	// Variable
	case Variable[T]:
		res := translationMap.MustGet(op.Metadata.String())
		return res
	// LinCombExt
	case LinComb[T]:
		children := make([]*Expression[T], len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewLinComb(children, op.Coeffs)
		return res
	// ProductExt
	case Product[T]:
		children := make([]*Expression[T], len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewProduct(children, op.Exponents)
		return res
	// LinearCombinationExt
	case PolyEval[T]:
		children := make([]*Expression[T], len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewPolyEval(children[0], children[1:])
		return res
	}
	panic("unreachable")
}

// ReconstructBottomUp applies the constructor function from the bottom-up of
// the current expression. It can be used to construct a new expression from
// `expr`. This is useful for symbolic expression simplication routines etc...
// The constructor function must not modify the value of the input expression
// nor it children.
func (e *Expression[T]) ReconstructBottomUp(
	constructor func(e *Expression[T], children []*Expression[T]) (new *Expression[T]),
) *Expression[T] {

	switch e.Operator.(type) {
	// Constant, indicating we reached the bottom of the expression. Thus it
	// applies the mutator and returns.
	case Constant[T], Variable[T]:
		return constructor(e, []*Expression[T]{})
	// LinCombExt or ProductExt or LinearCombinationExt. This is an intermediate expression.
	case LinComb[T], Product[T], PolyEval[T]:
		children := make([]*Expression[T], len(e.Children))
		// var wg sync.WaitGroup
		// wg.Add(len(e.Children))

		for i, c := range e.Children {
			// TODO @gbotrel next fix that --> too many go routines.
			// go func(i int, c *Expression[T]){ {
			// 	defer wg.Done()
			children[i] = c.ReconstructBottomUp(constructor)
			// }(i, c)
		}

		// wg.Wait()
		return constructor(e, children)
	}

	panic("unreachable")
}

// ReconstructBottomUpSingleThreaded is the same as [Expression.ReconstructBottomUp]
// but it is single threaded.
func (e *Expression[T]) ReconstructBottomUpSingleThreaded(
	constructor func(e *Expression[T], children []*Expression[T]) (new *Expression[T]),
) *Expression[T] {

	switch e.Operator.(type) {
	// Constant, indicating we reached the bottom of the expression. Thus it
	// applies the mutator and returns.
	case Constant[T], Variable[T]:
		x := constructor(e, []*Expression[T]{})
		if x == nil {
			panic(x)
		}
		return x
	// LinComb or Product or PolyEval. This is an intermediate expression.
	case LinComb[T], Product[T], PolyEval[T]:
		children := make([]*Expression[T], len(e.Children))
		for i, c := range e.Children {
			children[i] = c.ReconstructBottomUp(constructor)
			if children[i] == nil {
				panic(children[i])
			}
		}
		x := constructor(e, children)
		if x == nil {
			panic(x)
		}
		return x
	}

	panic("unreachable")
}

// SameWithNewChildren constructs a new expression that is a copy-cat of the
// receiver expression but swapping the children with the new ones instead. It
// is common for rebuilding expressions. If the expression is a variable or a
// constant, it returns itself.
func (e *Expression[T]) SameWithNewChildren(newChildren []*Expression[T]) *Expression[T] {

	switch op := e.Operator.(type) {
	// Constant
	case Constant[T], Variable[T]:
		return e
	// LinCombExt
	case LinComb[T]:
		return NewLinComb(newChildren, op.Coeffs)
	// ProductExt
	case Product[T]:
		return NewProduct(newChildren, op.Exponents)
	// LinearCombinationExt
	case PolyEval[T]:
		return NewPolyEval(newChildren[0], newChildren[1:])
	default:
		panic("unexpected type: " + reflect.TypeOf(op).String())
	}
}

// MarshalJSONString returns a JSON string returns a JSON string representation
// of the expression.
func (e *Expression[T]) MarshalJSONString() string {
	js, jsErr := json.MarshalIndent(e, "", "  ")
	if jsErr != nil {
		utils.Panic("failed to marshal expression: %v", jsErr)
	}
	return string(js)
}

// computeIsBaseFromChildren determines if an expression is a base expression
// based on its children.
func computeIsBaseFromChildren[T zk.Element](children []*Expression[T]) bool {
	for _, child := range children {
		if !child.IsBase {
			// at least one child is not a base expression, therefore the expression is itself not a base expression
			return false
		}
	}
	// all children are base expressions, therefore the expression is itself a base expression
	return true
}
