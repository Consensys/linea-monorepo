package symbolic

import (
	"fmt"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
)

/*
An anchored expression is an expression whose all nodes have been
anchored inside a board.
*/
type anchoredExpression struct {
	Board  *ExpressionBoard
	ESHash field.Element
}

/*
A floating expression is an expression whose definition is irrespective
of any-board. Typically, when we construct an expression we do it in a
floating expression and we "anchor" it.
*/
type Expression struct {
	ESHash   field.Element
	Children []*Expression
	Operator Operator
}

/*
Converts the expression into a board
*/
func (f *Expression) Board() ExpressionBoard {
	board := EmptyBoard()
	f.anchor(&board)
	return board
}

/*
Anchores an expression into the board
*/
func (f *Expression) anchor(b *ExpressionBoard) anchoredExpression {

	/*
		Check if the expression has not been already anchored
	*/
	if _, ok := b.ESHashesToPos[f.ESHash]; ok {
		return anchoredExpression{Board: b, ESHash: f.ESHash}
	}

	/*
		Recurse the call in all the children to ensure all
		subexpressions are anchored. And get their levels
	*/
	maxChildrenLevel := 0
	childrenIDs := []NodeID{}
	for _, child := range f.Children {
		_ = child.anchor(b)
		childID, ok := b.ESHashesToPos[child.ESHash]
		if !ok {
			utils.Panic("Children not found in expr")
		}
		maxChildrenLevel = utils.Max(maxChildrenLevel, childID.Level())
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
		b.Nodes = append(b.Nodes, []Node{})
	}
	NewNodeID := NewNodeID(newLevel, len(b.Nodes[newLevel]))

	newNode := Node{
		ESHash:   f.ESHash,
		Parents:  []NodeID{},
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
		b.GetNode(childID).addParent(NewNodeID)
	}
	b.Nodes[NewNodeID.Level()] = append(b.Nodes[NewNodeID.Level()], newNode)

	/*
		And returns the new Anchored expression
	*/
	return anchoredExpression{
		ESHash: f.ESHash,
		Board:  b,
	}
}

func (e *Expression) Validate() error {

	eshashes := make([]sv.SmartVector, len(e.Children))
	for i := range e.Children {
		eshashes[i] = sv.NewConstant(e.Children[i].ESHash, 1)
	}

	if len(e.Children) > 0 {
		// The cast back to sv.Constant is not functionally important but is an easy
		// sanity check.
		expectedESH := e.Operator.Evaluate(eshashes).(*sv.Constant).Get(0)
		if expectedESH != e.ESHash {
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

func (e *Expression) AssertValid() {
	if err := e.Validate(); err != nil {
		panic(err)
	}
}

// Reapply the expression by replacing the variables
func (e *Expression) Replay(translationMap collection.Mapping[string, *Expression]) (res *Expression) {

	switch op := e.Operator.(type) {
	// Constant
	case Constant:
		return NewConstant(op.Val)
	// Variable
	case Variable:
		res := translationMap.MustGet(op.Metadata.String())
		return res
	// LinComb
	case LinComb:
		children := make([]*Expression, len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewLinComb(children, op.Coeffs)
		return res
	// Product
	case Product:
		children := make([]*Expression, len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewProduct(children, op.Exponents)
		return res
	// PolyEval
	case PolyEval:
		children := make([]*Expression, len(e.Children))
		for i, c := range e.Children {
			children[i] = c.Replay(translationMap)
		}
		res := NewPolyEval(children[0], children[1:])
		return res
	}
	panic("unreachable")
}
