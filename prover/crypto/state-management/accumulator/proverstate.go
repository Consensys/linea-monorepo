package accumulator

import (
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/sirupsen/logrus"
)

// Holds the state of the accumulator
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

// Returns an initialized empty accumulator state
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

// FindHKey returns the position of a leaf opening whose HKey is
func (s *ProverState[K, V]) Config() *smt.Config {
	return s.Tree.Config
}

// Returns the root of the tree
func (s *ProverState[K, V]) SubTreeRoot() Digest {
	return s.Tree.Root
}

// Find the position of a key in the accumulator. If the key is
// absent, it returns 0, false. The returned position corresponds
// to the position in the tree
func (s *ProverState[K, V]) findKey(k K) (int64, bool) {
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

// Find the position of the two leaves sandwhich the queries leaf.
// It assumes that "k" is not stored in the tree.
func (s *ProverState[K, V]) findSandwich(k K) (int64, int64) {
	// We do so with a linear scanning to simplify (since it is only for testing)
	hkey := hash(s.Config(), k)
	hminus, iminus := Digest{}, int64(0)                        // corresponds to head
	hplus, iplus := s.Config().HashFunc().MaxDigest(), int64(1) // corresponds to tail

	// Technically, we should be able to skip the two first entries
	// The traversal of the data is in non-deterministic order
	for _, i := range s.Data.ListAllKeys() {
		// Hkey of the leaf opening we are visiting
		leafOpening := s.Data.MustGet(i).LeafOpening
		curHkey := leafOpening.HKey

		switch hashtypes.Cmp(curHkey, hkey) {
		case 0:
			// We should not have found a match
			utils.Panic("Found a perfect match for %+v", k)
		case 1:
			// curHKey is larger : so we should look for an update of hmax
			if hashtypes.Cmp(curHkey, hplus) == -1 {
				hplus, iplus = curHkey, i
			}
		case -1:
			// curHKey is smaller : so we should look for an update of hmin
			if hashtypes.Cmp(curHkey, hminus) == 1 {
				hminus, iminus = curHkey, i
			}
		}
	}

	if iminus == iplus {
		utils.Panic("iplus %v, iminus %v, hplus %v, hminus %v, hkey %v", iplus, iminus, hplus, hminus, hkey)
	}

	return iminus, iplus
}

// Cleanly upsert a tuple in the accumulator, returns a Merkle proof
func (p *ProverState[K, V]) upsertTuple(i int64, tuple KVOpeningTuple[K, V]) smt.Proof {

	leaf, err := tuple.CheckAndLeaf(p.Config())
	if err != nil {
		utils.Panic("illegal tuple : %v", err)
	}

	if old, found := p.Data.TryGet(i); found {
		// consistency-check of the old tuple
		_, err = old.CheckAndLeaf(p.Config())
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

	oldRoot := p.SubTreeRoot()

	// Perform the update
	p.Data.Update(i, tuple)
	p.Tree.Update(int(i), leaf)
	newRoot := p.SubTreeRoot()

	logrus.Tracef("upsert pos %v, leaf %x, root=%x -> %x", i, leaf, oldRoot, newRoot)
	return p.Tree.Prove(int(i))
}

func (p *ProverState[K, V]) rmTuple(i int64) smt.Proof {
	oldRoot := p.SubTreeRoot()

	// Update the tree with an empty leaf
	p.Data.MustExists(i)
	p.Data.Del(i)
	p.Tree.Update(int(i), smt.EmptyLeaf())
	newRoot := p.SubTreeRoot()

	logrus.Tracef("upsert pos %v, leaf %x, root=%x -> %x", i, smt.EmptyLeaf(), oldRoot, newRoot)
	return p.Tree.Prove(int(i))
}

// Returns the top-root hash which includes `NextFreeNode` and the `SubTreeRoot`
func (p *ProverState[K, V]) TopRoot() Digest {
	hasher := p.Config().HashFunc()
	hashtypes.WriteInt64To(hasher, p.NextFreeNode)
	subTreeRoot := p.SubTreeRoot()
	subTreeRoot.WriteTo(hasher)
	digest := hasher.Sum(nil)
	return hashtypes.BytesToDigest(digest)
}
