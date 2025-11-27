package symbolic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

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
	idToIndex := make(map[NodeID]int32, totalNodes)

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

// --- Binary encoding helpers for PackedExprGraph ---

func writeUvarint(w *bytes.Buffer, v uint64) {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(tmp[:], v)
	w.Write(tmp[:n])
}

func readUvarint(r *bytes.Reader) (uint64, error) {
	return binary.ReadUvarint(r)
}

func encodeZigZag(x int64) uint64 {
	return uint64((x << 1) ^ (x >> 63))
}

func decodeZigZag(u uint64) int64 {
	return int64((u >> 1) ^ uint64((int64(u&1)<<63)>>63))
}

// MetadataFromKey reconstructs Metadata from its String() representation.
// For now it preserves String() exactly via StringVar; you can refine later.
func MetadataFromKey(s string) Metadata {
	return StringVar(s)
}

// MarshalBinary encodes the graph into a compact byte slice.
// Metadata is stored only via Metadata.String().
func (pg *PackedExprGraph) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	// Collect metadata strings for variables.
	metaSet := make(map[string]struct{})
	for _, n := range pg.Nodes {
		if n.Op == OpVar && n.VarMeta != nil {
			metaSet[n.VarMeta.String()] = struct{}{}
		}
	}
	metaList := make([]string, 0, len(metaSet))
	for s := range metaSet {
		metaList = append(metaList, s)
	}
	// Optional: sort.Strings(metaList)

	metaIndex := make(map[string]uint64, len(metaList))
	for i, s := range metaList {
		metaIndex[s] = uint64(i)
	}

	// Write metadata table.
	writeUvarint(&buf, uint64(len(metaList)))
	for _, s := range metaList {
		b := []byte(s)
		writeUvarint(&buf, uint64(len(b)))
		buf.Write(b)
	}

	// Write root index.
	writeUvarint(&buf, uint64(pg.Root))

	// Write node count.
	writeUvarint(&buf, uint64(len(pg.Nodes)))

	// Write nodes.
	for _, n := range pg.Nodes {
		buf.WriteByte(byte(n.Op))

		// children
		writeUvarint(&buf, uint64(len(n.Children)))
		for _, cid := range n.Children {
			writeUvarint(&buf, uint64(cid))
		}

		switch n.Op {
		case OpConst:
			if n.ConstVal == nil {
				writeUvarint(&buf, 0)
				continue
			}
			var bi big.Int
			b := n.ConstVal.BigInt(&bi).Bytes()
			writeUvarint(&buf, uint64(len(b)))
			buf.Write(b)

		case OpVar:
			if n.VarMeta == nil {
				writeUvarint(&buf, ^uint64(0))
				continue
			}
			key := n.VarMeta.String()
			idx, ok := metaIndex[key]
			if !ok {
				return nil, fmt.Errorf("MarshalBinary: unknown metadata key %q", key)
			}
			writeUvarint(&buf, idx)

		case OpLinComb:
			writeUvarint(&buf, uint64(len(n.Coeffs)))
			for _, c := range n.Coeffs {
				writeUvarint(&buf, encodeZigZag(int64(c)))
			}

		case OpProduct:
			writeUvarint(&buf, uint64(len(n.Exponents)))
			for _, ex := range n.Exponents {
				if ex < 0 {
					return nil, fmt.Errorf("MarshalBinary: negative exponent %d", ex)
				}
				writeUvarint(&buf, uint64(ex))
			}

		case OpPolyEval:
			// no extra payload

		default:
			return nil, fmt.Errorf("MarshalBinary: unsupported opcode %d", n.Op)
		}
	}

	return buf.Bytes(), nil
}

func (pg *PackedExprGraph) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)

	// metadata table
	metaCount, err := readUvarint(r)
	if err != nil {
		return fmt.Errorf("UnmarshalBinary: read meta count: %w", err)
	}
	metas := make([]string, metaCount)
	for i := uint64(0); i < metaCount; i++ {
		l, err := readUvarint(r)
		if err != nil {
			return fmt.Errorf("UnmarshalBinary: read meta len: %w", err)
		}
		if l > uint64(r.Len()) {
			return fmt.Errorf("UnmarshalBinary: meta len %d > remaining %d", l, r.Len())
		}
		buf := make([]byte, l)
		if _, err := r.Read(buf); err != nil {
			return fmt.Errorf("UnmarshalBinary: read meta bytes: %w", err)
		}
		metas[i] = string(buf)
	}

	// root index
	rootIdx, err := readUvarint(r)
	if err != nil {
		return fmt.Errorf("UnmarshalBinary: read root index: %w", err)
	}

	// node count
	nodeCount, err := readUvarint(r)
	if err != nil {
		return fmt.Errorf("UnmarshalBinary: read node count: %w", err)
	}
	if nodeCount == 0 {
		pg.Nodes = nil
		pg.Root = -1
		return nil
	}

	nodes := make([]PackedExprNode, int(nodeCount))

	for i := 0; i < int(nodeCount); i++ {
		opb, err := r.ReadByte()
		if err != nil {
			return fmt.Errorf("UnmarshalBinary: read op: %w", err)
		}
		n := PackedExprNode{Op: OpCode(opb)}

		cc, err := readUvarint(r)
		if err != nil {
			return fmt.Errorf("UnmarshalBinary: read child count: %w", err)
		}
		n.Children = make([]int32, cc)
		for j := uint64(0); j < cc; j++ {
			ci, err := readUvarint(r)
			if err != nil {
				return fmt.Errorf("UnmarshalBinary: read child idx: %w", err)
			}
			n.Children[j] = int32(ci)
		}

		switch n.Op {
		case OpConst:
			lb, err := readUvarint(r)
			if err != nil {
				return fmt.Errorf("UnmarshalBinary: read const len: %w", err)
			}
			var fe field.Element
			if lb > 0 {
				if lb > uint64(r.Len()) {
					return fmt.Errorf("UnmarshalBinary: const len %d > remaining %d", lb, r.Len())
				}
				bb := make([]byte, lb)
				if _, err := r.Read(bb); err != nil {
					return fmt.Errorf("UnmarshalBinary: read const bytes: %w", err)
				}
				var bi big.Int
				bi.SetBytes(bb)
				fe.SetBigInt(&bi)
			} else {
				fe.SetZero()
			}
			n.ConstVal = &fe

		case OpVar:
			mi, err := readUvarint(r)
			if err != nil {
				return fmt.Errorf("UnmarshalBinary: read meta idx: %w", err)
			}
			if mi == ^uint64(0) {
				n.VarMeta = nil
				break
			}
			if mi >= uint64(len(metas)) {
				return fmt.Errorf("UnmarshalBinary: meta idx %d out of range", mi)
			}
			n.VarMeta = MetadataFromKey(metas[mi])

		case OpLinComb:
			cnt, err := readUvarint(r)
			if err != nil {
				return fmt.Errorf("UnmarshalBinary: read lin coeff cnt: %w", err)
			}
			n.Coeffs = make([]int, cnt)
			for k := uint64(0); k < cnt; k++ {
				uv, err := readUvarint(r)
				if err != nil {
					return fmt.Errorf("UnmarshalBinary: read lin coeff: %w", err)
				}
				n.Coeffs[k] = int(decodeZigZag(uv))
			}

		case OpProduct:
			cnt, err := readUvarint(r)
			if err != nil {
				return fmt.Errorf("UnmarshalBinary: read prod exp cnt: %w", err)
			}
			n.Exponents = make([]int, cnt)
			for k := uint64(0); k < cnt; k++ {
				ev, err := readUvarint(r)
				if err != nil {
					return fmt.Errorf("UnmarshalBinary: read prod exp: %w", err)
				}
				n.Exponents[k] = int(ev)
			}

		case OpPolyEval:
			// nothing

		default:
			return fmt.Errorf("UnmarshalBinary: unknown opcode %d", n.Op)
		}

		nodes[i] = n
	}

	pg.Root = int32(rootIdx)
	pg.Nodes = nodes
	return nil
}
