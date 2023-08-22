package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Format for a generic opening the order is important
type LeafOpening struct {
	Prev int64
	Next int64
	HKey Digest //it is mimc hash of the adress
	HVal Digest // is it mimc of account
}

// KVOpeningTuple is simple a tuple type
type KVOpeningTuple[K, V io.WriterTo] struct {
	LeafOpening LeafOpening
	Key         K
	Value       V
}

func (leaf *LeafOpening) WriteTo(w io.Writer) (int64, error) {
	n0, _ := hashtypes.WriteInt64To(w, leaf.Prev)
	n1, _ := hashtypes.WriteInt64To(w, leaf.Next)
	n2, _ := leaf.HKey.WriteTo(w)
	n3, _ := leaf.HVal.WriteTo(w)
	// Sanity-check the written size of the leaf opening
	total := n0 + n1 + n2 + n3
	if total != 128 {
		utils.Panic("bad size")
	}
	return total, nil
}

// Returns a hash of the leaf opening
func (leaf LeafOpening) Hash(conf *smt.Config) Digest {
	return hash(conf, &leaf)
}

// Head returns the "head" of the accumulator set
func Head() LeafOpening {
	return LeafOpening{
		Prev: 0, // Points to itself
		Next: 1,
		HKey: Digest{},
		HVal: Digest{},
	}
}

// Tail returns the "tail" of the accumulator set
func Tail(config *smt.Config) LeafOpening {
	return LeafOpening{
		Prev: 0,
		Next: 1, // Points to itself
		HVal: Digest{},
		HKey: config.HashFunc().MaxDigest(),
	}
}

// Returns true if the leaf opening is either head or tail
func (l *LeafOpening) HeadOrTail(config *smt.Config) bool {
	return l.HKey == config.HashFunc().MaxDigest() || l.HKey == Digest{}
}

// Check the internal consistency of the tuple and
// return the hash of the leaf opening.
func (t KVOpeningTuple[K, V]) CheckAndLeaf(conf *smt.Config) (Digest, error) {

	if t.LeafOpening.HeadOrTail(conf) {
		return t.LeafOpening.Hash(conf), nil
	}

	if t.LeafOpening.HKey != hash(conf, t.Key) {
		return Digest{}, fmt.Errorf("inconsistent key and leaf openings")
	}

	if t.LeafOpening.HVal != hash(conf, t.Value) {
		return Digest{}, fmt.Errorf("inconsistent val and leaf opening")
	}

	return t.LeafOpening.Hash(conf), nil
}

// Copy the leaf opening and set the prev in the copy
func (l LeafOpening) CopyWithPrev(prev int64) LeafOpening {
	l.Prev = prev
	return l
}

// Copy the leaf opening and set the next in the copy
func (l LeafOpening) CopyWithNext(next int64) LeafOpening {
	l.Next = next
	return l
}

// Returns true if the leaf opening opening matches the value
func (l *LeafOpening) MatchValue(conf *smt.Config, v io.WriterTo) bool {
	hval := hash(conf, v)
	return l.HVal == hval
}

// Returns true if the leaf opening opening matches the value
func (l *LeafOpening) MatchKey(conf *smt.Config, k io.WriterTo) bool {
	hkey := hash(conf, k)
	return l.HKey == hkey
}

// Copy the tuple and set the prev in the copy
func (t KVOpeningTuple[K, V]) CopyWithPrev(prev int64) KVOpeningTuple[K, V] {
	t.LeafOpening.Prev = prev
	return t
}

// Copy the tuple and set the next in the copy
func (t KVOpeningTuple[K, V]) CopyWithNext(next int64) KVOpeningTuple[K, V] {
	t.LeafOpening.Next = next
	return t
}

// Copy the tuple and give it a new new value
func (t KVOpeningTuple[K, V]) CopyWithVal(conf *smt.Config, val V) KVOpeningTuple[K, V] {
	t.Value = val
	t.LeafOpening.HVal = hash(conf, val)
	return t
}

// Pretty print a leaf opening
func (l LeafOpening) String() string {
	return fmt.Sprintf(
		"LeafOpening{Prev: %d, Next: %d, HKey: %s, HVal: %s}",
		l.Prev, l.Next, l.HKey.Hex(), l.HVal.Hex(),
	)
}
