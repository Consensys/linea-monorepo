package accumulator

import (
	"io"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"

	//lint:ignore ST1001 -- the package contains a list of standard types for this repo
	. "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// LeafOpening represents the opening of a leaf in the accumulator's tree.
//
// The `json` format and the order of the fields is important
type LeafOpening struct {
	Prev int64   `json:"prevLeaf"`
	Next int64   `json:"nextLeaf"`
	HKey Bytes32 `json:"hkey"` //it is mimc hash of the adress
	HVal Bytes32 `json:"hval"` // is it mimc of account
}

// KVOpeningTuple is simple a tuple type of (key, value) adding the
// corresponding leaf opening
type KVOpeningTuple[K, V io.WriterTo] struct {
	LeafOpening LeafOpening
	Key         K
	Value       V
}

// WriteTo implements the [io.WriterTo] interface and is used to hash the leaf
// opening into the leaves that we store in the tree.
func (leaf *LeafOpening) WriteTo(w io.Writer) (int64, error) {
	n0, _ := WriteInt64On32Bytes(w, leaf.Prev)
	n1, _ := WriteInt64On32Bytes(w, leaf.Next)
	n2, _ := leaf.HKey.WriteTo(w)
	n3, _ := leaf.HVal.WriteTo(w)
	// Sanity-check the written size of the leaf opening
	total := n0 + n1 + n2 + n3
	if total != 128 {
		utils.Panic("bad size")
	}
	return total, nil
}
