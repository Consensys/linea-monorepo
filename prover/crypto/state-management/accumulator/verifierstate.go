package accumulator

import (
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Track the informations that are maintained by the verifier
type VerifierState[K, V io.WriterTo] struct {
	// Location, identifier for the tree
	Location string
	// Track the index of the next free node
	NextFreeNode int64
	// Internal tree
	SubTreeRoot Digest
	// Config contains the parameters of the tree
	Config *smt.Config
}

// Create a verifier state from a prover state. The returned verifier
// state corresponds to a snapshot of the prover's state.
func (p *ProverState[K, V]) VerifierState() VerifierState[K, V] {
	return VerifierState[K, V]{
		Location:     p.Location,
		NextFreeNode: p.NextFreeNode,
		SubTreeRoot:  p.Tree.Root,
		Config:       p.Tree.Config,
	}
}

// Audit an atomic update of the merkle tree and returns the new root
func updateCheckRoot(conf *smt.Config, proof smt.Proof, root, old, new Digest) (newRoot Digest, err error) {

	if ok := proof.Verify(conf, old, root); !ok {
		return Digest{}, errors.New("root update audit failed : could not authenticate the old")
	}
	newRoot = proof.RecoverRoot(conf, new)
	logrus.Tracef("update check root %v leaf: %x->%x root: %x->%x\n", proof.Path, old, new, root, newRoot)
	return newRoot, nil
}

// Returns the top-root hash which includes `NextFreeNode` and the `SubTreeRoot`
func (v *VerifierState[K, V]) TopRoot() Digest {
	hasher := v.Config.HashFunc()
	hashtypes.WriteInt64To(hasher, v.NextFreeNode)
	v.SubTreeRoot.WriteTo(hasher)
	digest := hasher.Sum(nil)
	return hashtypes.BytesToDigest(digest)
}

// Registers the merkle proof claims that are audited when checking the updates of
// a trace.
func deferCheckUpdateRoot(
	conf *smt.Config,
	proof smt.Proof,
	root, old, new Digest,
	appendTo []smt.ProvedClaim,
) (appended []smt.ProvedClaim, newRoot Digest) {
	newRoot = proof.RecoverRoot(conf, new)
	appended = append(appendTo,
		smt.ProvedClaim{Proof: proof, Leaf: old, Root: root},
		smt.ProvedClaim{Proof: proof, Leaf: new, Root: newRoot},
	)
	return appended, newRoot
}
