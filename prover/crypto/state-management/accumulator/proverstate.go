package accumulator

import (
	"io"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"

	//lint:ignore ST1001 -- the package contains a list of standard types for this repo
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/sirupsen/logrus"
)

// ProverState holds the state of the accumulator: the tree itself and auxiliary
// structures to help tracking the positions of the tree.
type ProverState[K, V io.WriterTo] struct {
	// Location, identifier for the tree
	Location string
	// Track the index of the next free node
	NextFreeNode int64
	// Internal tree
	Tree *smt.Tree
	// Keys associated to the leaf #i
	Data collection.Mapping[int64, KVOpeningTuple[K, V]]
}

// InitializeProverState returns an initialized empty accumulator state
func InitializeProverState[K, V io.WriterTo](conf *smt.Config, location string) *ProverState[K, V] {
	tree := smt.NewEmptyTree(conf)

	// Insert the head and the tail in the tree
	head, tail := Head(), Tail(conf)
	tree.Update(0, head.Hash(conf))
	tree.Update(1, tail.Hash(conf))

	data := collection.NewMapping[int64, KVOpeningTuple[K, V]]()
	data.InsertNew(0, KVOpeningTuple[K, V]{LeafOpening: head})
	data.InsertNew(1, KVOpeningTuple[K, V]{LeafOpening: tail})

	return &ProverState[K, V]{
		Location:     location,
		NextFreeNode: 2, // because we inserted head and tail
		Tree:         tree,
		Data:         data,
	}
}

// Config returns the configuration of the accumulator.
func (s *ProverState[K, V]) Config() *smt.Config {
	return s.Tree.Config
}

// SubTreeRoot returns the root of the tree
func (s *ProverState[K, V]) SubTreeRoot() Bytes32 {
	return s.Tree.Root
}

// FindKey finds the position of a key in the accumulator. If the key is absent,
// it returns 0, false. The returned position corresponds to the position in the
// tree.
func (s *ProverState[K, V]) FindKey(k K) (int64, bool) {
	// We do so with a linear scan to simplify (since it is only for testing)
	hkey := hash(s.Config(), k)
	for _, i := range s.Data.ListAllKeys() {
		leafOpening := s.Data.MustGet(i).LeafOpening
		if hkey == leafOpening.HKey {
			return i, true
		}
	}
	return 0, false
}

// ListAllKeys is a function used for testing and traces sample generation which never should be called in production code
// as it is extremely inefficient
func (s *ProverState[K, V]) ListAllKeys() []K {
	var containedKeys []K
	// We compute the two keys that are used as bounds, to be able to ignore them later
	lowerBound := Bytes32{}
	upperBound := s.Config().HashFunc().MaxBytes32()
	for _, i := range s.Data.ListAllKeys() {
		tuple := s.Data.MustGet(i)
		if !(Bytes32Cmp(tuple.LeafOpening.HKey, lowerBound) == 0 || Bytes32Cmp(tuple.LeafOpening.HKey, upperBound) == 0) {
			containedKeys = append(containedKeys, tuple.Key)
		}
	}
	return containedKeys
}

// findSandwich finds the position of the two leaves sandwhich the queries leaf.
// It assumes that "k" is not stored in the tree.
func (s *ProverState[K, V]) findSandwich(k K) (int64, int64) {
	// We do so with a linear scanning to simplify (since it is only for testing)
	hkey := hash(s.Config(), k)
	hminus, iminus := Bytes32{}, int64(0)                        // corresponds to head
	hplus, iplus := s.Config().HashFunc().MaxBytes32(), int64(1) // corresponds to tail

	// Technically, we should be able to skip the two first entries
	// The traversal of the data is in non-deterministic order
	for _, i := range s.Data.ListAllKeys() {
		// Hkey of the leaf opening we are visiting
		leafOpening := s.Data.MustGet(i).LeafOpening
		curHkey := leafOpening.HKey

		switch Bytes32Cmp(curHkey, hkey) {
		case 0:
			// We should not have found a match
			utils.Panic("Found a perfect match for %+v", k)
		case 1:
			// curHKey is larger : so we should look for an update of hmax
			if Bytes32Cmp(curHkey, hplus) == -1 {
				hplus, iplus = curHkey, i
			}
		case -1:
			// curHKey is smaller : so we should look for an update of hmin
			if Bytes32Cmp(curHkey, hminus) == 1 {
				hminus, iminus = curHkey, i
			}
		}
	}

	if iminus == iplus {
		utils.Panic("iplus %v, iminus %v, hplus %v, hminus %v, hkey %v", iplus, iminus, hplus, hminus, hkey)
	}

	return iminus, iplus
}

// upsertTuple cleanly upsert a tuple in the accumulator, returns a Merkle proof
func (s *ProverState[K, V]) upsertTuple(i int64, tuple KVOpeningTuple[K, V]) smt.Proof {

	leaf, err := tuple.CheckAndLeaf(s.Config())
	if err != nil {
		utils.Panic("illegal tuple : %v", err)
	}

	if old, found := s.Data.TryGet(i); found {
		// consistency-check of the old tuple
		_, err = old.CheckAndLeaf(s.Config())
		if err != nil {
			utils.Panic("illegal old tuple : %v", err)
		}

		utils.Require(
			old.LeafOpening.HKey == tuple.LeafOpening.HKey,
			"(pos %v) cannot change the key: %++v -> %++v", i, old, tuple)

		// Count the number of changes, exactly one change can happen at a time
		numChanges := 0

		if old.LeafOpening.HVal != tuple.LeafOpening.HVal {
			numChanges++
		}

		if old.LeafOpening.Prev != tuple.LeafOpening.Prev {
			numChanges++
		}

		if old.LeafOpening.Next != tuple.LeafOpening.Next {
			numChanges++
		}

		if numChanges != 1 {
			utils.Panic(
				"illegal change, can only change one field of the leaf opening (but changed %v) \nold: %++v \nnew: %++v)",
				numChanges, old, tuple,
			)
		}
	}

	oldRoot := s.SubTreeRoot()

	// Perform the update
	s.Data.Update(i, tuple)
	s.Tree.Update(int(i), leaf)
	newRoot := s.SubTreeRoot()

	logrus.Tracef("upsert pos %v, leaf %x, root=%x -> %x", i, leaf, oldRoot, newRoot)
	return s.Tree.MustProve(int(i))
}

// rmTuple erases a leaf from the tree.
func (s *ProverState[K, V]) rmTuple(i int64) smt.Proof {
	oldRoot := s.SubTreeRoot()

	// Update the tree with an empty leaf
	s.Data.MustExists(i)
	s.Data.Del(i)
	s.Tree.Update(int(i), smt.EmptyLeaf())
	newRoot := s.SubTreeRoot()

	logrus.Tracef("upsert pos %v, leaf %x, root=%x -> %x", i, smt.EmptyLeaf(), oldRoot, newRoot)
	return s.Tree.MustProve(int(i))
}

// TopRoot returns the top-root hash which includes `NextFreeNode` and the
// `SubTreeRoot`
func (s *ProverState[K, V]) TopRoot() Bytes32 {
	hasher := s.Config().HashFunc()
	WriteInt64On32Bytes(hasher, s.NextFreeNode)
	subTreeRoot := s.SubTreeRoot()
	subTreeRoot.WriteTo(hasher)
	Bytes32 := hasher.Sum(nil)
	return AsBytes32(Bytes32)
}
