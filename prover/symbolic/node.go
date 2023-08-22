package symbolic

import (
	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

/*
NodeID represents the position of a Node in the DAG. It consists of
two subIDs:
  - the 32 MSB represents the level of the node in the DAG
  - the 32 LSB represents the index of the node in its level

The two together gives the position of the node = expression.Nodes[level][pos]
*/
type NodeID uint64

/*
Interface for generic functions representing an operation :
`add, sub, mul, neg` etc...
*/
type Operator interface {
	Evaluate([]sv.SmartVector) sv.SmartVector
	Validate(e *Expression) error
	Degree([]int) int
	GnarkEval(frontend.API, []frontend.Variable) frontend.Variable
}

/*
A nodes consists of an operator and a list of operands. Additionally; it
gather "DAG-wise" data such as the parents of the node etc...
*/
type Node struct {
	// Indicates the parents ID of the current node
	Parents  []NodeID
	Children []NodeID
	/*
		ESHash (Expression Sensitive Hash) is a field.Element representing an expression.
		Two expression with the same ESHash are considered as equals. This helps expression
		pruning.
	*/
	ESHash field.Element
	// Operator contains the logic to evaluate an expression
	Operator Operator
}

// Add a parent to the node
func (n *Node) addParent(p NodeID) {
	n.Parents = append(n.Parents, p)
}

// Returns the position in the level from a NodeID
func (i NodeID) PosInLevel() int {
	res := i & ((1 << 32) - 1)
	return int(res)
}

// Returns the Level from a NodeID
func (i NodeID) Level() int {
	return int(i >> 32)
}

// Returns the node id given its level and its position in a level
func NewNodeID(level, posInLevel int) NodeID {
	if level >= 1<<32 {
		utils.Panic("level is too large %v", level)
	}
	if posInLevel >= 1<<32 {
		utils.Panic("pos in level is too large %v", posInLevel)
	}
	return NodeID(uint64(level)<<32 | uint64(posInLevel))
}
