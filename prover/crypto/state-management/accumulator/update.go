package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/pkg/errors"
)

// Trace the accumulator : update
type UpdateTrace[K, V io.WriterTo] struct {
	Location           string
	Key                K
	OldValue, NewValue V
	// We call it new next free node, but the value is not updated
	// during the update.
	NewNextFreeNode        int
	OldSubRoot, NewSubRoot Digest
	OldOpening             LeafOpening
	Proof                  smt.Proof
}

// Perform a read on the accumulator. Panics if the
// the associated value is zero. Returns a trace object
// containing the
func (p *ProverState[K, V]) UpdateAndProve(key K, newVal V) UpdateTrace[K, V] {

	// Find the position of the leaf containing our value
	i, found := p.findKey(key)
	if !found {
		utils.Panic("called read-non-zero, but the key was not present")
	}

	tuple := p.Data.MustGet(i)

	if hash(p.Config(), key) != hash(p.Config(), tuple.Key) {
		utils.Panic("sanity-check : the key mismatched")
	}

	oldRoot := p.SubTreeRoot()
	oldOpening := tuple.LeafOpening
	oldValue := tuple.Value

	// Compute the new value and update the tree
	tuple.Value = newVal
	tuple.LeafOpening.HVal = hash(p.Config(), tuple.Value)
	p.Data.Update(i, tuple)

	newLeaf := tuple.LeafOpening.Hash(p.Config())
	p.Tree.Update(int(i), newLeaf)

	return UpdateTrace[K, V]{
		Location:        p.Location,
		Key:             tuple.Key,
		OldValue:        oldValue,
		NewValue:        newVal,
		OldOpening:      oldOpening,
		OldSubRoot:      oldRoot,
		NewSubRoot:      p.SubTreeRoot(),
		NewNextFreeNode: int(p.NextFreeNode),
		Proof:           p.Tree.Prove(int(i)),
	}
}

// Verify a read on the accumulator. Panics if the associated
// value is non-zero.
func (v *VerifierState[K, V]) UpdateVerify(trace UpdateTrace[K, V]) error {

	// If the location does not match the we return an error
	if v.Location != trace.Location {
		return fmt.Errorf("inconsistent location : %v != %v", v.Location, trace.Location)
	}

	// Check that verifier's root is the same as the one in the traces
	if v.SubTreeRoot != trace.OldSubRoot {
		return fmt.Errorf("inconsistent root %v != %v", v.SubTreeRoot, trace.OldSubRoot)
	}

	tuple := KVOpeningTuple[K, V]{
		Key:         trace.Key,
		Value:       trace.OldValue,
		LeafOpening: trace.OldOpening,
	}

	leaf, err := tuple.CheckAndLeaf(v.Config)
	if err != nil {
		return errors.WithMessage(err, "read update verifier failed")
	}

	if !trace.Proof.Verify(v.Config, leaf, trace.OldSubRoot) {
		return fmt.Errorf("merkle proof verification failed")
	}

	newTuple := tuple
	newTuple.Value = trace.NewValue
	newTuple.LeafOpening.HVal = hash(v.Config, trace.NewValue)

	// We panic because if the consistency check passed
	newLeaf := hash(v.Config, &newTuple.LeafOpening)

	newRoot, err := updateCheckRoot(v.Config, trace.Proof, trace.OldSubRoot, leaf, newLeaf)
	if err != nil {
		return errors.Wrap(err, "update check failed : invalid proof")
	}

	if newRoot != trace.NewSubRoot {
		return errors.New("new root does not match")
	}

	// Update the verifier's view
	v.SubTreeRoot = trace.NewSubRoot
	return nil
}

// DeferMerkleChecks appends all the merkle-proofs checks happening in a trace verification
// into a slice of smt.ProvedClaim.
func (trace UpdateTrace[K, V]) DeferMerkleChecks(
	config *smt.Config,
	appendTo []smt.ProvedClaim,
) []smt.ProvedClaim {

	tuple := KVOpeningTuple[K, V]{
		Key:         trace.Key,
		Value:       trace.OldValue,
		LeafOpening: trace.OldOpening,
	}

	leaf, _ := tuple.CheckAndLeaf(config)

	newTuple := tuple
	newTuple.Value = trace.NewValue
	newTuple.LeafOpening.HVal = hash(config, trace.NewValue)

	// We panic because if the consistency check passed
	newLeaf := hash(config, &newTuple.LeafOpening)
	appendTo, _ = deferCheckUpdateRoot(config, trace.Proof, trace.OldSubRoot, leaf, newLeaf, appendTo)
	return appendTo
}
