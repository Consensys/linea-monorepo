package symbolic

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/arena"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

type GetDegree = func(m interface{}) int

// ListVariableMetadata return the list of the metadata of the variables
// into the board. Importantly, the order in which the metadata is returned
// matches the order in which the assignments to the boarded expression must be
// provided.
func (b *ExpressionBoard) ListVariableMetadata() []Metadata {
	res := []Metadata{}
	for i := range b.Nodes {
		if vari, ok := b.Nodes[i].Operator.(Variable); ok {
			res = append(res, vari.Metadata)
		}
	}
	return res
}

// Degree returns the overall degree of the expression board. It admits a custom
// function `getDeg` which is used to assign a degree to the [Variable] leaves
// of the ExpressionBoard.
func (b *ExpressionBoard) Degree(getdeg GetDegree) int {
	if len(b.Nodes) == 0 {
		return 0
	}
	degrees := make([]int, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch v := node.Operator.(type) {
		case Constant:
			degrees[i] = 0
		case Variable:
			degrees[i] = getdeg(v.Metadata)
			inputCursor++
		default:
			childrenDeg := make([]int, len(node.Children))
			for k, childID := range node.Children {
				childrenDeg[k] = degrees[childID]
			}
			degrees[i] = node.Operator.Degree(childrenDeg)
		}
	}
	return degrees[len(b.Nodes)-1]
}

/*
GnarkEval evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEval(api frontend.API, inputs []koalagnark.Element) koalagnark.Element {
	if len(b.Nodes) == 0 {
		panic("empty board")
	}
	results := make([]koalagnark.Element, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch op := node.Operator.(type) {
		case Constant:
			tmp := op.Val.GetExt()
			results[i] = koalagnark.NewElementFromKoala(tmp.B0.A0) // @thomas ext or base ?
		case Variable:
			results[i] = inputs[inputCursor]
			inputCursor++
		default:
			nodeInputs := make([]koalagnark.Element, len(node.Children))
			for k, childID := range node.Children {
				nodeInputs[k] = results[childID]
			}
			results[i] = node.Operator.GnarkEval(api, nodeInputs)
		}
	}
	return results[len(b.Nodes)-1]
}

/*
GnarkEvalExt evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEvalExt(api frontend.API, inputs []koalagnark.Ext) koalagnark.Ext {
	if len(b.Nodes) == 0 {
		panic("empty board")
	}
	results := make([]koalagnark.Ext, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch op := node.Operator.(type) {
		case Constant:
			results[i] = koalagnark.NewExt(op.Val.GetExt())
		case Variable:
			results[i] = inputs[inputCursor]
			inputCursor++
		default:
			nodeInputs := make([]koalagnark.Ext, len(node.Children))
			for k, childID := range node.Children {
				nodeInputs[k] = results[childID]
			}
			results[i] = node.Operator.GnarkEvalExt(api, nodeInputs)
		}
	}
	return results[len(b.Nodes)-1]
}

// DumpToString is a debug utility which print out the expression in a readable
// format.
func (b *ExpressionBoard) DumpToString() string {
	res := ""
	for i, node := range b.Nodes {
		res += fmt.Sprintf("%d: (%T) %++v\n", i, node.Operator, node)
	}
	return res
}

// CountNodes returns the node count of the expression
func (b *ExpressionBoard) CountNodes() int {
	return len(b.Nodes)
}
