package symbolic

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// OpCode identifies the concrete operator type.
type OpCode uint8

const (
	OpConst OpCode = iota
	OpVar
	OpLinComb
	OpProduct
	OpPolyEval
)

// PackedExprGraph is a DAG encoding of an Expression in topo order.
type PackedExprGraph struct {
	// Root is the index of the root node in Nodes.
	Root  int32
	Nodes []PackedExprNode
}

type PackedExprNode struct {
	Op        OpCode
	Children  []int32        // indices into Nodes
	ConstVal  *field.Element // OpConst
	VarMeta   Metadata       // OpVar
	Coeffs    []int          // OpLinComb
	Exponents []int          // OpProduct
}

// ToPackedExprFromBoard converts a boarded expression into a PackedExprGraph.
// It assumes `board` comes from e.Board() for some Expression.
func ToPackedExprFromBoard(board ExpressionBoard) PackedExprGraph {
	// 1) Flatten nodes level-by-level and assign dense indices.
	totalNodes := 0
	for lvl := range board.Nodes {
		totalNodes += len(board.Nodes[lvl])
	}
	nodes := make([]PackedExprNode, totalNodes)

	// Map nodeID -> flat index.
	idToIndex := make(map[nodeID]int32, totalNodes)

	flatIdx := 0
	for lvl := range board.Nodes {
		for pos := range board.Nodes[lvl] {
			id := newNodeID(lvl, pos)
			idToIndex[id] = int32(flatIdx)
			flatIdx++
		}
	}

	// 2) Fill PackedExprNode entries in the same flat order.
	flatIdx = 0
	for lvl := range board.Nodes {
		for pos := range board.Nodes[lvl] {
			n := board.Nodes[lvl][pos]

			// Children indices.
			children := make([]int32, len(n.Children))
			for i, cid := range n.Children {
				children[i] = idToIndex[cid]
			}

			pn := PackedExprNode{Children: children}
			switch op := n.Operator.(type) {
			case Constant:
				v := op.Val
				pn.Op = OpConst
				pn.ConstVal = &v
			case Variable:
				pn.Op = OpVar
				pn.VarMeta = op.Metadata
			case LinComb:
				pn.Op = OpLinComb
				pn.Coeffs = append([]int(nil), op.Coeffs...)
			case Product:
				pn.Op = OpProduct
				pn.Exponents = append([]int(nil), op.Exponents...)
			case PolyEval:
				pn.Op = OpPolyEval
			default:
				panic(fmt.Sprintf("ToPackedExprFromBoard: unsupported operator %T", op))
			}

			nodes[flatIdx] = pn
			flatIdx++
		}
	}

	// Root is last node in board-level order.
	root := int32(totalNodes - 1)
	return PackedExprGraph{Root: root, Nodes: nodes}
}

// FromPackedExpr reconstructs an Expression from a PackedExprGraph.
func FromPackedExpr(pg PackedExprGraph) *Expression {
	if len(pg.Nodes) == 0 {
		return nil
	}
	exprs := make([]*Expression, len(pg.Nodes))

	var build func(int32) *Expression
	build = func(i int32) *Expression {
		if exprs[i] != nil {
			return exprs[i]
		}
		pn := pg.Nodes[i]

		children := make([]*Expression, len(pn.Children))
		for j, cid := range pn.Children {
			children[j] = build(cid)
		}

		var e *Expression
		switch pn.Op {
		case OpConst:
			if pn.ConstVal == nil {
				panic("FromPackedExpr: OpConst without value")
			}
			e = NewConstant(*pn.ConstVal)
		case OpVar:
			if pn.VarMeta == nil {
				panic("FromPackedExpr: OpVar without metadata")
			}
			e = NewVariable(pn.VarMeta)
		case OpLinComb:
			e = NewLinComb(children, pn.Coeffs)
		case OpProduct:
			e = NewProduct(children, pn.Exponents)
		case OpPolyEval:
			if len(children) < 2 {
				panic("FromPackedExpr: OpPolyEval with fewer than 2 children")
			}
			x := children[0]
			coeffs := children[1:]
			e = NewPolyEval(x, coeffs)
		default:
			panic(fmt.Sprintf("FromPackedExpr: unsupported opcode %d", pn.Op))
		}

		exprs[i] = e
		return e
	}

	if pg.Root < 0 || int(pg.Root) >= len(pg.Nodes) {
		panic(fmt.Sprintf("FromPackedExpr: invalid root index %d", pg.Root))
	}
	return build(pg.Root)
}
