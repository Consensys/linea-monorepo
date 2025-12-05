package serialization

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// RawExprGraph mirrors symbolic.PackedExprGraph but uses serialization-friendly types.
type RawExprGraph struct {
	Root  int32
	Nodes []RawExprNode
}

// RawExprNode mirrors symbolic.PackedExprNode.
// Key difference: ConstVal is *big.Int (for serialization) and Metadata is any (for interface support).
type RawExprNode struct {
	Op        uint8
	Children  []int32
	ConstVal  *big.Int
	Metadata  any // Holds the concrete Metadata object (e.g. column.Natural)
	Coeffs    []int
	Exponents []int
}

func (ser *Serializer) packExprGraph(exprGraph *symbolic.PackedExprGraph) (*RawExprGraph, *serdeError) {
	rawExprGraph := &RawExprGraph{
		Root:  exprGraph.Root,
		Nodes: make([]RawExprNode, len(exprGraph.Nodes)),
	}

	for i := range exprGraph.Nodes {
		rawExprNode, err := ser.packExprNode(&exprGraph.Nodes[i])
		if err != nil {
			return nil, err
		}
		rawExprGraph.Nodes[i] = *rawExprNode
	}

	return rawExprGraph, nil
}

func (de *Deserializer) unpackExprGraph(rawExprGraph *RawExprGraph) (*symbolic.PackedExprGraph, *serdeError) {
	exprGraph := &symbolic.PackedExprGraph{
		Root:  rawExprGraph.Root,
		Nodes: make([]symbolic.PackedExprNode, len(rawExprGraph.Nodes)),
	}

	for i := range rawExprGraph.Nodes {
		exprNode, err := de.unpackExprNode(&rawExprGraph.Nodes[i])
		if err != nil {
			return nil, err
		}
		exprGraph.Nodes[i] = *exprNode
	}

	return exprGraph, nil
}

func (ser *Serializer) packExprNode(exprNode *symbolic.PackedExprNode) (*RawExprNode, *serdeError) {
	rawExprNode := &RawExprNode{
		Op:        uint8(exprNode.Op),
		Children:  exprNode.Children,
		Coeffs:    exprNode.Coeffs,
		Exponents: exprNode.Exponents,
	}

	if exprNode.ConstVal != nil {
		rawExprNode.ConstVal = fieldToSmallBigInt(*exprNode.ConstVal)
	}

	// Simple assignment. The serializer's PackStructObject will see
	// the 'any' field and handle the interface packing automatically.
	if exprNode.VarMeta != nil {
		rawExprNode.Metadata = exprNode.VarMeta
	}

	return rawExprNode, nil
}

func (de *Deserializer) unpackExprNode(rawExprNode *RawExprNode) (*symbolic.PackedExprNode, *serdeError) {
	exprNode := &symbolic.PackedExprNode{
		Op:        symbolic.OpCode(rawExprNode.Op),
		Children:  rawExprNode.Children,
		Coeffs:    rawExprNode.Coeffs,
		Exponents: rawExprNode.Exponents,
	}

	if rawExprNode.ConstVal != nil {
		var fe field.Element
		fe.SetBigInt(rawExprNode.ConstVal)
		exprNode.ConstVal = &fe
	}

	if rawExprNode.Metadata != nil {
		// Because RawExprNode.Metadata is 'any', the serializer's UnpackStructObject
		// will use UnpackInterface (driven by CBOR tags) and return the concrete object.
		// We just need to assert it implements the Metadata interface.
		meta, ok := rawExprNode.Metadata.(symbolic.Metadata)
		if !ok {
			return nil, newSerdeErrorf("unpacked metadata does not implement symbolic.Metadata, got %T", rawExprNode.Metadata)
		}
		exprNode.VarMeta = meta
	}

	return exprNode, nil
}

/*
func (ser *Serializer) packExprNode(exprNode *symbolic.PackedExprNode) (*RawExprNode, *serdeError) {

	rawExprNode := &RawExprNode{
		Op:        uint8(exprNode.Op),
		Children:  exprNode.Children,
		Coeffs:    exprNode.Coeffs,
		Exponents: exprNode.Exponents,
	}

	if exprNode.ConstVal != nil {
		rawExprNode.ConstVal = fieldToSmallBigInt(*exprNode.ConstVal)
	}

	if exprNode.VarMeta != nil {
		_valAny, err := ser.PackInterface(reflect.ValueOf(exprNode.VarMeta))
		if err != nil {
			return nil, err
		}
		rawExprNode.Metadata = _valAny
	}

	return rawExprNode, nil
}

func (de *Deserializer) unpackExprNode(rawExprNode *RawExprNode) (*symbolic.PackedExprNode, *serdeError) {

	exprNode := &symbolic.PackedExprNode{
		Op:        symbolic.OpCode(rawExprNode.Op),
		Children:  rawExprNode.Children,
		Coeffs:    rawExprNode.Coeffs,
		Exponents: rawExprNode.Exponents,
	}

	if rawExprNode.ConstVal != nil {
		var fe *field.Element
		fe.SetBigInt(rawExprNode.ConstVal)
		exprNode.ConstVal = fe
	}

	if rawExprNode.Metadata != nil {
		v_ := make(map[any]any)
		val, err := de.UnpackInterface(v_, reflect.TypeOf(exprNode.VarMeta))
		if err != nil {
			return nil, err
		}

		exprNode.VarMeta = val.Interface().(symbolic.Metadata)
	}

	return exprNode, nil
}  */

/*

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

// marshalBinary encodes the graph into a compact byte slice.
// Metadata is stored only via Metadata.String().
func marshalBinary(pg *symbolic.PackedExprGraph) ([]byte, error) {
	var buf bytes.Buffer

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
		case symbolic.OpConst:
			if n.ConstVal == nil {
				writeUvarint(&buf, 0)
				continue
			}
			var bi big.Int
			b := n.ConstVal.BigInt(&bi).Bytes()
			writeUvarint(&buf, uint64(len(b)))
			buf.Write(b)

		case symbolic.OpVar:
			if n.VarMeta == nil {
				writeUvarint(&buf, ^uint64(0))
				continue
			}
			// key := n.VarMeta.String()
			// idx, ok := metaIndex[key]
			// if !ok {
			// 	return nil, fmt.Errorf("MarshalBinary: unknown metadata key %q", key)
			// }


			writeUvarint(&buf, idx)

		case symbolic.OpLinComb:
			writeUvarint(&buf, uint64(len(n.Coeffs)))
			for _, c := range n.Coeffs {
				writeUvarint(&buf, encodeZigZag(int64(c)))
			}

		case symbolic.OpProduct:
			writeUvarint(&buf, uint64(len(n.Exponents)))
			for _, ex := range n.Exponents {
				if ex < 0 {
					return nil, fmt.Errorf("MarshalBinary: negative exponent %d", ex)
				}
				writeUvarint(&buf, uint64(ex))
			}

		case symbolic.OpPolyEval:
			// no extra payload

		default:
			return nil, fmt.Errorf("MarshalBinary: unsupported opcode %d", n.Op)
		}
	}

	return buf.Bytes(), nil
}

func unmarshalBinary(pg *symbolic.PackedExprGraph, data []byte) error {
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

	nodes := make([]symbolic.PackedExprNode, int(nodeCount))

	for i := 0; i < int(nodeCount); i++ {
		opb, err := r.ReadByte()
		if err != nil {
			return fmt.Errorf("UnmarshalBinary: read op: %w", err)
		}
		n := symbolic.PackedExprNode{Op: symbolic.OpCode(opb)}

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
		case symbolic.OpConst:
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

		case symbolic.OpVar:
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

		case symbolic.OpLinComb:
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

		case symbolic.OpProduct:
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

		case symbolic.OpPolyEval:
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

*/
