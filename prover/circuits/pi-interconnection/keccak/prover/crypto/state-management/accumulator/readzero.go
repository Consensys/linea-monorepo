package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"

	//lint:ignore ST1001 -- the package contains a list of standard types for this repo

	. "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// Trace that allows checking a read zero operation: e.g. proof of non-membership
type ReadZeroTrace[K, V io.WriterTo] struct {
	Type         int         `json:"type"`
	Location     string      `json:"location"`
	Key          K           `json:"key"`
	SubRoot      Bytes32     `json:"subRoot"`
	NextFreeNode int         `json:"nextFreeNode"`
	OpeningMinus LeafOpening `json:"leftLeaf"`
	OpeningPlus  LeafOpening `json:"rightLeaf"`
	ProofMinus   smt.Proof   `json:"leftProof"`
	ProofPlus    smt.Proof   `json:"rightProof"`
}

// ReadZeroAndProve performs a read-zero on the accumulator. Panics if the
// associated key exists in the tree. Returns a ReadZeroTrace object in case of
// success.
func (p *ProverState[K, V]) ReadZeroAndProve(key K) ReadZeroTrace[K, V] {

	// Find the position of the leaf containing our value
	_, found := p.FindKey(key)
	if found {
		utils.Panic("called read-zero, but the key was present")
	}

	iMinus, iPlus := p.findSandwich(key)
	dataMinus := p.Data.MustGet(iMinus)
	dataPlus := p.Data.MustGet(iPlus)

	return ReadZeroTrace[K, V]{
		Location:     p.Location,
		Key:          key,
		SubRoot:      p.SubTreeRoot(),
		ProofMinus:   p.Tree.MustProve(int(iMinus)),
		OpeningMinus: dataMinus.LeafOpening,
		ProofPlus:    p.Tree.MustProve(int(iPlus)),
		OpeningPlus:  dataPlus.LeafOpening,
		NextFreeNode: int(p.NextFreeNode),
	}
}

// ReadZeroVerify verifies a [ReadZeroTrace] and returns an error in case of
// failure.
func (v *VerifierState[K, V]) ReadZeroVerify(trace ReadZeroTrace[K, V]) error {

	// If the location does not match the we return an error
	if v.Location != trace.Location {
		return fmt.Errorf("inconsistent location : %v != %v", v.Location, trace.Location)
	}

	// Check that verifier's root is the same as the one in the traces
	if v.SubTreeRoot != trace.SubRoot {
		return fmt.Errorf("inconsistent root %v != %v", v.SubTreeRoot, trace.SubRoot)
	}

	iMinus, iPlus := int64(trace.ProofMinus.Path), int64(trace.ProofPlus.Path)

	// Check that Plus and Minus point to each other
	// Check that the two sandwich leaf opening point to each other
	if trace.OpeningMinus.Next != iPlus || trace.OpeningPlus.Prev != iMinus {
		return fmt.Errorf(
			"bad sandwich prev and next do not point to each other : prev %++v, next %++v",
			trace.OpeningMinus, trace.OpeningPlus,
		)
	}

	// Check that the opening's hkeys make a correct sandwich
	hkey := hash(v.Config, trace.Key)
	if Bytes32Cmp(hkey, trace.OpeningMinus.HKey) < 1 || Bytes32Cmp(hkey, trace.OpeningPlus.HKey) > -1 {
		return fmt.Errorf(
			"sandwich is incorrect expected %x < %x < %x",
			trace.OpeningMinus.HKey, hkey, trace.OpeningPlus.HKey,
		)
	}

	// Test membership of leaf minus
	leafMinus := hash(v.Config, &trace.OpeningMinus)
	if !trace.ProofMinus.Verify(v.Config, leafMinus, trace.SubRoot) {
		return fmt.Errorf("merkle proof verification failed : minus")
	}

	// Test membership of leaf plus
	leafPlus := hash(v.Config, &trace.OpeningPlus)
	if !trace.ProofPlus.Verify(v.Config, leafPlus, trace.SubRoot) {
		return fmt.Errorf("merkle proof verification failed : plus")
	}

	// Check that NextFreeNode for the Prover and the Verifier are the same
	if v.NextFreeNode != int64(trace.NextFreeNode) {
		return fmt.Errorf("inconsistent NextFreeNode %v != %v", v.NextFreeNode, trace.NextFreeNode)
	}

	return nil
}

// DeferMerkleChecks implements the [Trace] interface.
func (trace ReadZeroTrace[K, V]) DeferMerkleChecks(
	config *smt.Config,
	appendTo []smt.ProvedClaim,
) []smt.ProvedClaim {

	// Test membership of leaf minus
	leafMinus := hash(config, &trace.OpeningMinus)

	// Test membership of leaf plus
	leafPlus := hash(config, &trace.OpeningPlus)

	appendTo = append(appendTo, smt.ProvedClaim{Proof: trace.ProofMinus, Root: trace.SubRoot, Leaf: leafMinus})
	return append(appendTo, smt.ProvedClaim{Proof: trace.ProofPlus, Root: trace.SubRoot, Leaf: leafPlus})
}

func (trace ReadZeroTrace[K, V]) HKey(cfg *smt.Config) Bytes32 {
	return hash(cfg, trace.Key)
}

func (trace ReadZeroTrace[K, V]) RWInt() int {
	return 0
}
