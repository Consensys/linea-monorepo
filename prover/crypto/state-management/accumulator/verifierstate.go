package accumulator

import (
	"io"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	//lint:ignore ST1001 -- the package contains a list of standard types for this repo
	"github.com/consensys/linea-monorepo/prover/utils/types"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
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
	SubTreeRoot Bytes32
	// Config contains the parameters of the tree
	Config *smt.Config
}

// VerifierState create a verifier state from a prover state. The returned
// verifier state corresponds to a snapshot of the prover's state.
func (p *ProverState[K, V]) VerifierState() VerifierState[K, V] {
	return VerifierState[K, V]{
		Location:     p.Location,
		NextFreeNode: p.NextFreeNode,
		SubTreeRoot:  types.HashToBytes32(p.Tree.Root),
		Config:       p.Tree.Config,
	}
}

// updateCheckRoot audit an atomic update of the merkle tree (e.g. one leaf) and
// returns the new root
func updateCheckRoot(conf *smt.Config, proof smt.Proof, root, old, new Bytes32) (newRoot Bytes32, err error) {

	if ok := proof.Verify(conf, types.Bytes32ToHash(old), types.Bytes32ToHash(root)); !ok {
		return Bytes32{}, errors.New("root update audit failed : could not authenticate the old")
	}

	// Note: all possible errors are already convered by `proof.Verify`
	newRootOct, _ := proof.RecoverRoot(conf, types.Bytes32ToHash(new))
	logrus.Tracef("update check root %v leaf: %x->%x root: %x->%x\n", proof.Path, old, new, root, newRootOct)
	return types.HashToBytes32(newRootOct), nil
}

// TopRoot returns the top-root hash which includes `NextFreeNode` and the
// `SubTreeRoot`
func (v *VerifierState[K, V]) TopRoot() Bytes32 {
	hasher := v.Config.HashFunc()
	WriteInt64On32Bytes(hasher, v.NextFreeNode)
	v.SubTreeRoot.WriteTo(hasher)
	Bytes32 := hasher.Sum(nil)
	return AsBytes32(Bytes32)
}

// deferCheckUpdateRoot appends to `appendTo` the merkle proof claims that are
// audited when checking the updates of a leaf in the tree. The function panics
// if the proof malformed.
func deferCheckUpdateRoot(
	conf *smt.Config,
	proof smt.Proof,
	root, old, new Bytes32,
	appendTo []smt.ProvedClaim,
) (appended []smt.ProvedClaim, newRoot Bytes32) {
	newRootOct, err := proof.RecoverRoot(conf, types.Bytes32ToHash(new))
	if err != nil {
		panic(err)
	}
	newRoot = types.HashToBytes32(newRootOct)
	appended = append(appendTo,
		smt.ProvedClaim{Proof: proof, Leaf: types.Bytes32ToHash(old), Root: types.Bytes32ToHash(root)},
		smt.ProvedClaim{Proof: proof, Leaf: types.Bytes32ToHash(new), Root: types.Bytes32ToHash(newRoot)},
	)

	return appended, newRoot
}
