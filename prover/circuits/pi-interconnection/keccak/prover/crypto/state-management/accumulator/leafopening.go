package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
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

// Hash returns a hash of the leaf opening
func (leaf LeafOpening) Hash(conf *smt.Config) Bytes32 {
	return hash(conf, &leaf)
}

// Head returns the "head" of the accumulator set
func Head() LeafOpening {
	return LeafOpening{
		Prev: 0, // Points to itself
		Next: 1,
		HKey: Bytes32{},
		HVal: Bytes32{},
	}
}

// Tail returns the "tail" of the accumulator set
func Tail(config *smt.Config) LeafOpening {
	return LeafOpening{
		Prev: 0,
		Next: 1, // Points to itself
		HVal: Bytes32{},
		HKey: config.HashFunc().MaxBytes32(),
	}
}

// HeadOrTail returns true if the leaf opening is either head or tail
func (leaf *LeafOpening) HeadOrTail(config *smt.Config) bool {
	return leaf.HKey == config.HashFunc().MaxBytes32() || leaf.HKey == Bytes32{}
}

// CheckAndLeaf check the internal consistency of the tuple and returns the hash
// of the leaf opening (corresponding to a leaf).
func (t KVOpeningTuple[K, V]) CheckAndLeaf(conf *smt.Config) (Bytes32, error) {

	if t.LeafOpening.HeadOrTail(conf) {
		return t.LeafOpening.Hash(conf), nil
	}

	if t.LeafOpening.HKey != hash(conf, t.Key) {
		return Bytes32{}, fmt.Errorf("inconsistent key and leaf openings")
	}

	if t.LeafOpening.HVal != hash(conf, t.Value) {
		return Bytes32{}, fmt.Errorf("inconsistent val and leaf opening")
	}

	return t.LeafOpening.Hash(conf), nil
}

// CopyWithPrev copies the leaf opening and set the prev in the copy
func (leaf LeafOpening) CopyWithPrev(prev int64) LeafOpening {
	leaf.Prev = prev
	return leaf
}

// CopyWithNext copies the leaf opening and set the next in the copy
func (leaf LeafOpening) CopyWithNext(next int64) LeafOpening {
	leaf.Next = next
	return leaf
}

// MatchValue returns true if the leaf opening opening matches the value
func (l *LeafOpening) MatchValue(conf *smt.Config, v io.WriterTo) bool {
	hval := hash(conf, v)
	return l.HVal == hval
}

// MatchKey returns true if the leaf opening opening matches the value
func (l *LeafOpening) MatchKey(conf *smt.Config, k io.WriterTo) bool {
	hkey := hash(conf, k)
	return l.HKey == hkey
}

// CopyWithPrev copies the tuple and set the prev in the copy
func (t KVOpeningTuple[K, V]) CopyWithPrev(prev int64) KVOpeningTuple[K, V] {
	t.LeafOpening.Prev = prev
	return t
}

// CopyWithNext copies the tuple and set the next in the copy
func (t KVOpeningTuple[K, V]) CopyWithNext(next int64) KVOpeningTuple[K, V] {
	t.LeafOpening.Next = next
	return t
}

// CopyWithVal copies the tuple and give it a new new value
func (t KVOpeningTuple[K, V]) CopyWithVal(conf *smt.Config, val V) KVOpeningTuple[K, V] {
	t.Value = val
	t.LeafOpening.HVal = hash(conf, val)
	return t
}

// String pretty prints a leaf opening
func (l LeafOpening) String() string {
	return fmt.Sprintf(
		"LeafOpening{Prev: %d, Next: %d, HKey: %s, HVal: %s}",
		l.Prev, l.Next, l.HKey.Hex(), l.HVal.Hex(),
	)
}
