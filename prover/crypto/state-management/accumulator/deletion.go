package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Trace the accumulator : deletion
type DeletionTrace[K, V io.WriterTo] struct {
	Location string
	// For consistency, we call it new next free node but
	// the value is not updated during a deletion
	NewNextFreeNode        int
	OldSubRoot, NewSubRoot Digest
	// `New` correspond to the inserted leaf
	ProofMinus, ProofDeleted, ProofPlus smt.Proof
	Key                                 K
	// Value of the leaf opening before being modified
	OldOpenMinus, DeletedOpen, OldOpenPlus LeafOpening
	// The deleted value
	DeletedValue V
}

// Delete an entry in the accumulator and returns a trace
func (p *ProverState[K, V]) DeleteAndProve(key K) (trace DeletionTrace[K, V]) {

	// Sanity-check : assert that the key is missing in the proof
	i, found := p.findKey(key)
	if !found {
		utils.Panic("called read-zero, but the key was present")
	}

	// No need to look for the sandwich, we can find it in the leafopening
	tuple := p.Data.MustGet(i)
	iMinus, iPlus := tuple.LeafOpening.Prev, tuple.LeafOpening.Next
	tupleMinus := p.Data.MustGet(iMinus)
	tuplePlus := p.Data.MustGet(iPlus)

	trace = DeletionTrace[K, V]{
		Location:        p.Location,
		Key:             key,
		OldSubRoot:      p.SubTreeRoot(),
		OldOpenMinus:    tupleMinus.LeafOpening,
		OldOpenPlus:     tuplePlus.LeafOpening,
		DeletedOpen:     tuple.LeafOpening,
		NewNextFreeNode: int(p.NextFreeNode),
	}

	// 1/ The Prev
	newTupleMinus := tupleMinus.CopyWithNext(iPlus)
	trace.ProofMinus = p.upsertTuple(iMinus, newTupleMinus)

	// 2/ Delete the target leaf
	trace.ProofDeleted = p.rmTuple(i)

	// 3/ The next
	newTuplePlus := tuplePlus.CopyWithPrev(iMinus)
	trace.ProofPlus = p.upsertTuple(iPlus, newTuplePlus)

	trace.NewSubRoot = p.SubTreeRoot()
	return trace
}

func (v *VerifierState[K, V]) VerifyDeletion(trace DeletionTrace[K, V]) error {

	// If the location does not match the we return an error
	if v.Location != trace.Location {
		return fmt.Errorf("inconsistent location : %v != %v", v.Location, trace.Location)
	}

	// Check that verifier's root is the same as the one in the traces
	if v.SubTreeRoot != trace.OldSubRoot {
		return fmt.Errorf("inconsistent root %v != %v", v.SubTreeRoot, trace.OldSubRoot)
	}

	iMinus := int64(trace.ProofMinus.Path)
	iDeleted := int64(trace.ProofDeleted.Path)
	iPlus := int64(trace.ProofPlus.Path)

	// Check that minus and the deleted branch point to each other
	if (trace.OldOpenMinus.Next != iDeleted) || (trace.DeletedOpen.Prev != iMinus) {
		return fmt.Errorf(
			"bad sandwich prev and next do not point to each other : prev %++v, next %++v",
			trace.OldOpenMinus, trace.DeletedOpen,
		)
	}

	// Check that the two sandwich leaf opening point to each other
	if (trace.DeletedOpen.Next != iPlus) || (trace.OldOpenPlus.Prev != iDeleted) {
		return fmt.Errorf(
			"bad sandwich prev and next do not point to each other : prev %++v, next %++v",
			trace.DeletedOpen, trace.OldOpenPlus,
		)
	}

	// Check that the deleted entry corresponds to the key we wish to remove
	if !trace.DeletedOpen.MatchKey(v.Config, trace.Key) {
		return fmt.Errorf("deleting the wrong leaf : does not match our key : trace.Key %v - hkey %v", trace.Key, trace.DeletedOpen.HKey)
	}

	currentRoot := trace.OldSubRoot

	// Audit the update of the "minus"
	oldLeafMinus := trace.OldOpenMinus.Hash(v.Config)
	newLeafMinus := trace.OldOpenMinus.CopyWithNext(iPlus).Hash(v.Config)
	currentRoot, err := updateCheckRoot(v.Config, trace.ProofMinus, currentRoot, oldLeafMinus, newLeafMinus)
	if err != nil {
		return fmt.Errorf("audit of the update of old leaf minus failed %v", err)
	}

	// Audit the update of the deleted leaf
	deletedLeaf := hash(v.Config, &trace.DeletedOpen)
	currentRoot, err = updateCheckRoot(v.Config, trace.ProofDeleted, currentRoot, deletedLeaf, smt.EmptyLeaf())
	if err != nil {
		return fmt.Errorf("audit of the update of the middle leaf failed %v", err)
	}

	// Audit the update of the "plus"
	oldLeafPlus := trace.OldOpenPlus.Hash(v.Config)
	newLeafPlus := trace.OldOpenPlus.CopyWithPrev(iMinus).Hash(v.Config)
	currentRoot, err = updateCheckRoot(v.Config, trace.ProofPlus, currentRoot, oldLeafPlus, newLeafPlus)
	if err != nil {
		return fmt.Errorf("audit of the update of old leaf plus failed %v", err)
	}

	// Check that the alleged new root is consistent with the one we reconstructed
	if currentRoot != trace.NewSubRoot {
		return fmt.Errorf("inconsistent root %v != %v", currentRoot, trace.NewSubRoot)
	}

	// Every check passed : update the verifier state
	v.SubTreeRoot = currentRoot
	return nil
}

// DeferMerkleChecks appends all the merkle-proofs checks happening in a trace verification
// into a slice of smt.ProvedClaim.
func (trace DeletionTrace[K, V]) DeferMerkleChecks(
	config *smt.Config,
	appendTo []smt.ProvedClaim,
) []smt.ProvedClaim {
	currentRoot := trace.OldSubRoot
	iMinus := int64(trace.ProofMinus.Path)
	iPlus := int64(trace.ProofPlus.Path)

	// Audit the update of the "minus"
	oldLeafMinus := trace.OldOpenMinus.Hash(config)
	newLeafMinus := trace.OldOpenMinus.CopyWithNext(iPlus).Hash(config)
	appendTo, currentRoot = deferCheckUpdateRoot(config, trace.ProofMinus, currentRoot, oldLeafMinus, newLeafMinus, appendTo)

	// the proof verification for the deleted leaf
	deletedLeaf := hash(config, &trace.DeletedOpen)
	appendTo, currentRoot = deferCheckUpdateRoot(config, trace.ProofDeleted, currentRoot, deletedLeaf, smt.EmptyLeaf(), appendTo)

	// Audit the update of the "plus"
	oldLeafPlus := trace.OldOpenPlus.Hash(config)
	newLeafPlus := trace.OldOpenPlus.CopyWithPrev(iMinus).Hash(config)
	appendTo, _ = deferCheckUpdateRoot(config, trace.ProofPlus, currentRoot, oldLeafPlus, newLeafPlus, appendTo)

	return appendTo
}
