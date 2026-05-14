// Adapter: converts symbolic.ExpressionBoard → []NodeOp for GPU compilation.
//
// The ExpressionBoard uses Go interface types (Variable, Constant, LinComb,
// Product, PolyEval) while the GPU compiler needs flat []NodeOp. This thin
// glue converts between the two representations.
package symbolic

import (
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// BoardToNodeOps converts an ExpressionBoard's topologically-sorted nodes
// into the GPU-portable NodeOp representation.
//
//	board.Nodes[i].Operator  →  NodeOp.Kind
//	  Variable               →  OpInput  (leaf, references next input variable)
//	  Constant               →  OpConst  (leaf, E4 value in Montgomery form)
//	  LinComb{Coeffs}        →  OpLinComb{Children, Coeffs}
//	  Product{Exponents}     →  OpProduct{Children, Coeffs=exponents}
//	  PolyEval{}             →  OpPolyEval{Children}
func BoardToNodeOps(board *symbolic.ExpressionBoard) []NodeOp {
	nodes := board.Nodes
	ops := make([]NodeOp, len(nodes))

	for i, node := range nodes {
		children := make([]int, len(node.Children))
		for j, c := range node.Children {
			children[j] = int(c)
		}

		switch op := node.Operator.(type) {
		case symbolic.Variable:
			ops[i] = NodeOp{Kind: OpInput}

		case symbolic.Constant:
			// Extract E4 value (always available via GetExt, in Montgomery form)
			val := op.Val.GetExt()
			ops[i] = NodeOp{
				Kind: OpConst,
				ConstVal: [4]uint32{
					uint32(val.B0.A0[0]), uint32(val.B0.A1[0]),
					uint32(val.B1.A0[0]), uint32(val.B1.A1[0]),
				},
			}

		case symbolic.LinComb:
			ops[i] = NodeOp{
				Kind:     OpLinComb,
				Children: children,
				Coeffs:   append([]int(nil), op.Coeffs...),
			}

		case symbolic.Product:
			ops[i] = NodeOp{
				Kind:     OpProduct,
				Children: children,
				Coeffs:   append([]int(nil), op.Exponents...),
			}

		case symbolic.PolyEval:
			ops[i] = NodeOp{
				Kind:     OpPolyEval,
				Children: children,
			}

		default:
			panic("gpu/symbolic: BoardToNodeOps: unknown operator type")
		}
	}
	return ops
}
