package symbolic

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type boardAssignment [][]nodeAssignment

type nodeAssignment struct {
	Addr            [2]int
	Node            *Node
	Value           sv.SmartVector
	NumKnownParents int
}

func (b *ExpressionBoard) prepareNodeAssignments(inputs []sv.SmartVector) boardAssignment {

	var (
		nodeAssignments = make(boardAssignment, len(b.Nodes))
		length          = inputs[0].Len()
		inputCursor     = 0
	)

	// This loops pre-allocate all the inner-slices of nodeAssignment
	for lvl := range nodeAssignments {
		nodeAssignments[lvl] = make([]nodeAssignment, len(b.Nodes[lvl]))
	}

	// This loop stores the values of the leaves of the expression (e.g. the
	// inputs of the circuit and the constants of the circuit)
	for i := range b.Nodes[0] {

		nodeAssignments[0][i] = nodeAssignment{
			Node: &b.Nodes[0][i],
		}

		switch op := b.Nodes[0][i].Operator.(type) {
		case Constant:
			// The constants are identified to constant vectors
			nodeAssignments[0][i].Value = sv.NewConstant(op.Val, length)
		case Variable:
			// Sanity-check the input should have the correct length
			if inputs[inputCursor].Len() != length {
				utils.Panic("Subvector failed, subvector should have size %v but size is %v", length, inputs[inputCursor].Len())
			}
			nodeAssignments[0][i].Value = inputs[inputCursor]
			inputCursor++
		}
	}

	// This loop pre-assigns the wires that are constants
	for lvl := 1; lvl < len(nodeAssignments); lvl++ {
		for pil := range nodeAssignments[lvl] {

			var (
				node = nodeAssignment{
					Node: &b.Nodes[lvl][pil],
					Addr: [2]int{lvl, pil},
				}
				inputs  = nodeAssignments.inputOf(&node)
				success = node.tryGuessEval(inputs)
			)

			if success {
				for i := range inputs {
					nodeAssignments.incParentKnownCountOf(&inputs[i], nil, true)
				}
			}

			nodeAssignments[lvl][pil] = node
		}
	}

	return nodeAssignments
}

func (b boardAssignment) eval(na *nodeAssignment, pool *mempool.Pool) {

	if (na.allParentsKnown() && na.hasParents()) || na.hasAValue() {
		return
	}

	var (
		val = b.inputOf(na)
		smv = make([]sv.SmartVector, len(val))
	)

	for i, v := range val {
		if v.Value == nil {
			panic("found a nil")
		}

		smv[i] = v.Value
	}

	na.Value = na.Node.Operator.Evaluate(smv, pool)

	for i := range val {
		b.incParentKnownCountOf(&val[i], pool, false)
	}
}

func (na *nodeAssignment) tryGuessEval(val []nodeAssignment) bool {

	if na.hasAValue() {
		return true
	}

	var (
		anyIsZero  bool
		allAreCnst bool = true
		input           = make([]sv.SmartVector, len(val))
		length          = 0
	)

	for i, v := range val {
		var (
			c, isC = v.constValue()
			isZero = isC && (c.Val() == field.Element{})
		)

		allAreCnst = allAreCnst && isC
		anyIsZero = anyIsZero && isZero
		input[i] = c

		if isC {
			length = c.Len()
		}
	}

	switch na.Node.Operator.(type) {

	case LinComb, PolyEval:
		if allAreCnst {
			na.Value = na.Node.Operator.Evaluate(input, nil)
			return true
		}
		return false

	case Product:
		if anyIsZero {
			na.Value = sv.NewConstant(field.Element{}, length)
			return true
		}
		return false
	default:
		panic("unexpected type")
	}
}

func (na *nodeAssignment) hasAValue() bool {
	return na.Value != nil
}

func (na *nodeAssignment) allParentsKnown() bool {
	return na.NumKnownParents == len(na.Node.Parents)
}

func (na *nodeAssignment) hasParents() bool {
	return len(na.Node.Parents) > 0
}

func (na *nodeAssignment) constValue() (*sv.Constant, bool) {

	if na.Value == nil {
		return nil, false
	}

	if c, ok := na.Value.(*sv.Constant); ok {
		return c, true
	}

	return nil, false
}

func (b boardAssignment) inputOf(na *nodeAssignment) []nodeAssignment {

	if na.Node == nil {
		panic("na has a nil node")
	}

	nodeInputs := make([]nodeAssignment, len(na.Node.Children))

	for i, childID := range na.Node.Children {
		var (
			lvl = childID.level()
			pil = childID.posInLevel()
		)

		nodeInputs[i] = b[lvl][pil]
	}
	return nodeInputs
}

func (b boardAssignment) incParentKnownCountOf(na *nodeAssignment, pool *mempool.Pool, recursive bool) (wasDeleted bool) {
	na.NumKnownParents++

	// Sanity-checking that this function is not called too many time
	if na.NumKnownParents > len(na.Node.Parents) {
		panic("invalid count: overflowing the total number of parent")
	}

	if na.allParentsKnown() {

		if recursive {
			children := b.inputOf(na)
			for i := range children {
				b.incParentKnownCountOf(&children[i], pool, recursive)
			}
		}

		return na.tryFree(pool)
	}

	return false
}

func (na *nodeAssignment) tryFree(pool *mempool.Pool) bool {
	if pool == nil {
		return false
	}

	if na.Value == nil {
		return false
	}

	switch na.Node.Operator.(type) {
	case Constant, Variable:
		return false
	}

	if !na.allParentsKnown() {
		return false
	}

	if reg, ok := na.Value.(*sv.Regular); ok {
		na.Value = nil
		v := []field.Element(*reg)
		pool.Free(&v)
		return true
	}

	return false
}
