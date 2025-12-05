package symbolic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	Root      int32
	Nodes     []PackedExprNode
	Metadatas []Metadata // Kept separate for the Serializer to handle
}

type PackedExprNode struct {
	Op        OpCode
	Children  []int32
	ConstVal  *field.Element
	MetaIdx   int // Index into PackedExprGraph.Metadatas
	Coeffs    []int
	Exponents []int
}

// ToPackedExprFromBoard converts a boarded expression into a PackedExprGraph.
// Uses optimized O(1) indexing and metadata deduplication.
func ToPackedExprFromBoard(board ExpressionBoard) PackedExprGraph {
	totalNodes := 0
	levelOffsets := make([]int32, len(board.Nodes))
	for lvl, nodes := range board.Nodes {
		levelOffsets[lvl] = int32(totalNodes)
		totalNodes += len(nodes)
	}

	nodes := make([]PackedExprNode, totalNodes)

	// Use string keys for stable deduplication without "unhashable type" panic
	metaMap := make(map[string]int)
	metadatas := make([]Metadata, 0)

	flatIdx := 0
	for _, levelNodes := range board.Nodes {
		for _, n := range levelNodes {
			children := make([]int32, len(n.Children))
			for i, cid := range n.Children {
				cLvl := cid.level()
				cPos := cid.posInLevel()
				children[i] = levelOffsets[cLvl] + int32(cPos)
			}

			pn := PackedExprNode{Children: children, MetaIdx: -1}

			switch op := n.Operator.(type) {
			case Constant:
				v := op.Val
				pn.Op = OpConst
				pn.ConstVal = &v
			case Variable:
				pn.Op = OpVar
				key := op.Metadata.String()
				if idx, ok := metaMap[key]; ok {
					pn.MetaIdx = idx
				} else {
					idx = len(metadatas)
					metadatas = append(metadatas, op.Metadata)
					metaMap[key] = idx
					pn.MetaIdx = idx
				}
			case LinComb:
				pn.Op = OpLinComb
				if len(op.Coeffs) > 0 {
					pn.Coeffs = make([]int, len(op.Coeffs))
					copy(pn.Coeffs, op.Coeffs)
				}
			case Product:
				pn.Op = OpProduct
				if len(op.Exponents) > 0 {
					pn.Exponents = make([]int, len(op.Exponents))
					copy(pn.Exponents, op.Exponents)
				}
			case PolyEval:
				pn.Op = OpPolyEval
			default:
				panic(fmt.Sprintf("ToPackedExprFromBoard: unsupported operator %T", op))
			}
			nodes[flatIdx] = pn
			flatIdx++
		}
	}

	return PackedExprGraph{
		Root:      int32(totalNodes - 1),
		Nodes:     nodes,
		Metadatas: metadatas,
	}
}

// FromPackedExpr reconstructs an Expression using the graph + external metadata.
func FromPackedExpr(pg PackedExprGraph) *Expression {
	if len(pg.Nodes) == 0 {
		return nil
	}
	exprs := make([]*Expression, len(pg.Nodes))

	for i := range pg.Nodes {
		pn := &pg.Nodes[i]

		children := make([]*Expression, len(pn.Children))
		for k, cid := range pn.Children {
			children[k] = exprs[cid]
		}

		var e *Expression
		switch pn.Op {
		case OpConst:
			e = NewConstant(*pn.ConstVal)
		case OpVar:
			// Uses the restored rich Metadata object
			e = NewVariable(pg.Metadatas[pn.MetaIdx])
		case OpLinComb:
			e = NewLinComb(children, pn.Coeffs)
		case OpProduct:
			e = NewProduct(children, pn.Exponents)
		case OpPolyEval:
			if len(children) < 2 {
				if len(children) == 1 {
					e = children[0]
				} else {
					panic("OpPolyEval < 2 children")
				}
			} else {
				e = NewPolyEval(children[0], children[1:])
			}
		}
		exprs[i] = e
	}
	return exprs[pg.Root]
}

// MarshalTopology encodes ONLY the graph structure, no metadata values.
func (pg *PackedExprGraph) MarshalTopology() ([]byte, error) {
	estSize := 16 + len(pg.Nodes)*20
	buf := bytes.NewBuffer(make([]byte, 0, estSize))

	writeUvarint(buf, uint64(pg.Root))
	writeUvarint(buf, uint64(len(pg.Nodes)))

	for i := range pg.Nodes {
		n := &pg.Nodes[i]
		buf.WriteByte(byte(n.Op))

		writeUvarint(buf, uint64(len(n.Children)))
		for _, cid := range n.Children {
			writeUvarint(buf, uint64(cid))
		}

		switch n.Op {
		case OpConst:
			var bi big.Int
			b := n.ConstVal.BigInt(&bi).Bytes()
			writeUvarint(buf, uint64(len(b)))
			buf.Write(b)
		case OpVar:
			// Only store the index. The Serializer stores the actual object.
			writeUvarint(buf, uint64(n.MetaIdx))
		case OpLinComb:
			writeUvarint(buf, uint64(len(n.Coeffs)))
			for _, c := range n.Coeffs {
				writeUvarint(buf, encodeZigZag(int64(c)))
			}
		case OpProduct:
			writeUvarint(buf, uint64(len(n.Exponents)))
			for _, ex := range n.Exponents {
				writeUvarint(buf, uint64(ex))
			}
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalTopology decodes the graph structure.
func (pg *PackedExprGraph) UnmarshalTopology(data []byte) error {
	r := bytes.NewReader(data)

	rootIdx, err := readUvarint(r)
	if err != nil {
		return err
	}
	pg.Root = int32(rootIdx)

	nodeCount, err := readUvarint(r)
	if err != nil {
		return err
	}

	pg.Nodes = make([]PackedExprNode, nodeCount)

	for i := 0; i < int(nodeCount); i++ {
		opb, _ := r.ReadByte()
		n := PackedExprNode{Op: OpCode(opb)}

		cc, _ := readUvarint(r)
		if cc > 0 {
			n.Children = make([]int32, cc)
			for j := uint64(0); j < cc; j++ {
				ci, _ := readUvarint(r)
				n.Children[j] = int32(ci)
			}
		}

		switch n.Op {
		case OpConst:
			lb, _ := readUvarint(r)
			bb := make([]byte, lb)
			io.ReadFull(r, bb)
			var bi big.Int
			bi.SetBytes(bb)
			var fe field.Element
			fe.SetBigInt(&bi)
			n.ConstVal = &fe
		case OpVar:
			idx, _ := readUvarint(r)
			n.MetaIdx = int(idx)
		case OpLinComb:
			cnt, _ := readUvarint(r)
			if cnt > 0 {
				n.Coeffs = make([]int, cnt)
				for k := uint64(0); k < cnt; k++ {
					uv, _ := readUvarint(r)
					n.Coeffs[k] = int(decodeZigZag(uv))
				}
			}
		case OpProduct:
			cnt, _ := readUvarint(r)
			if cnt > 0 {
				n.Exponents = make([]int, cnt)
				for k := uint64(0); k < cnt; k++ {
					ev, _ := readUvarint(r)
					n.Exponents[k] = int(ev)
				}
			}
		}
		pg.Nodes[i] = n
	}
	return nil
}

// Helpers
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
