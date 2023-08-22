package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/pkg/errors"
)

// Trace the accumulator : read non-zero
type ReadNonZeroTrace[K, V io.WriterTo] struct {
	Location     string
	NextFreeNode int
	Key          K
	Value        V
	SubRoot      Digest
	LeafOpening  LeafOpening
	Proof        smt.Proof
}

// Perform a read on the accumulator. Panics if the
// the associated value is zero. Returns a trace object
// containing the
func (p *ProverState[K, V]) ReadNonZeroAndProve(key K) ReadNonZeroTrace[K, V] {

	// Find the position of the leaf containing our value
	i, found := p.findKey(key)
	if !found {
		utils.Panic("called read-non-zero, but the key was not present")
	}

	tuple := p.Data.MustGet(i)

	if hash(p.Config(), key) != hash(p.Config(), tuple.Key) {
		utils.Panic("sanity-check : the key mismatched")
	}

	return ReadNonZeroTrace[K, V]{
		Location:     p.Location,
		Key:          tuple.Key,
		Value:        tuple.Value,
		LeafOpening:  tuple.LeafOpening,
		SubRoot:      p.SubTreeRoot(),
		Proof:        p.Tree.Prove(int(i)),
		NextFreeNode: int(p.NextFreeNode),
	}
}

// Verify a read on the accumulator. Panics if the associated
// value is non-zero.
func (v *VerifierState[K, V]) ReadNonZeroVerify(trace ReadNonZeroTrace[K, V]) error {

	// If the location does not match the we return an error
	if v.Location != trace.Location {
		return fmt.Errorf("inconsistent location : %v != %v", v.Location, trace.Location)
	}

	// Check that verifier's root is the same as the one in the traces
	if v.SubTreeRoot != trace.SubRoot {
		return fmt.Errorf("inconsistent root %v != %v", v.SubTreeRoot, trace.SubRoot)
	}

	tuple := KVOpeningTuple[K, V]{
		Key:         trace.Key,
		Value:       trace.Value,
		LeafOpening: trace.LeafOpening,
	}

	leaf, err := tuple.CheckAndLeaf(v.Config)
	if err != nil {
		return errors.WithMessage(err, "read non zero verifier failed")
	}

	if !trace.Proof.Verify(v.Config, leaf, trace.SubRoot) {
		return fmt.Errorf("merkle proof verification failed")
	}

	return nil
}

// DeferMerkleChecks appends all the merkle-proofs checks happening in a trace verification
// into a slice of smt.ProvedClaim.
func (trace ReadNonZeroTrace[K, V]) DeferMerkleChecks(
	config *smt.Config,
	appendTo []smt.ProvedClaim,
) []smt.ProvedClaim {
	tuple := KVOpeningTuple[K, V]{
		Key:         trace.Key,
		Value:       trace.Value,
		LeafOpening: trace.LeafOpening,
	}

	leaf, _ := tuple.CheckAndLeaf(config)
	appendTo = append(appendTo, smt.ProvedClaim{Proof: trace.Proof, Root: trace.SubRoot, Leaf: leaf})
	return appendTo
}
