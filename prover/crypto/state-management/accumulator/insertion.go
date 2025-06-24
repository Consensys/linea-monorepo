package accumulator

import (
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils"

	//lint:ignore ST1001 -- the package contains a list of standard types for this repo

	. "github.com/consensys/linea-monorepo/prover/utils/types"
)

// InsertionTrace gathers all the input needed for a verifier to audit the
// insertion of a key in the map.
type InsertionTrace[K, V io.WriterTo] struct {
	// Identifier for the tree this trace belongs to
	Type     int    `json:"type"`
	Location string `json:"location"`

	NewNextFreeNode int     `json:"newNextFreeNode"`
	OldSubRoot      Bytes32 `json:"oldSubRoot"`
	NewSubRoot      Bytes32 `json:"newSubRoot"`

	// `New` correspond to the inserted leaf
	ProofMinus smt.Proof `json:"leftProof"`
	ProofNew   smt.Proof `json:"newProof"`
	ProofPlus  smt.Proof `json:"rightProof"`
	Key        K         `json:"key"`
	Val        V         `json:"value"`

	// Value of the leaf opening before being modified
	OldOpenMinus LeafOpening `json:"priorLeftLeaf"`
	OldOpenPlus  LeafOpening `json:"priorRightLeaf"`
}

// InsertAndProve inserts in the accumulator and returns a trace. The function
// panics if the key is already in the accumulator or if the tree is corrupted.
func (p *ProverState[K, V]) InsertAndProve(key K, val V) (trace InsertionTrace[K, V]) {

	// Sanity-check : assert that the key is missing in the proof
	_posFound, found := p.FindKey(key)
	if found {
		utils.Panic("called insert, but the key was present : %v as position %v", key, _posFound)
	}

	// Fetch the leaf openings and add them in the trace
	iMinus, iPlus := p.findSandwich(key)
	tupleMinus := p.Data.MustGet(iMinus)
	tuplePlus := p.Data.MustGet(iPlus)

	trace = InsertionTrace[K, V]{
		Location: p.Location,
		Key:      key, Val: val, OldSubRoot: p.SubTreeRoot(),
		OldOpenMinus:    tupleMinus.LeafOpening,
		OldOpenPlus:     tuplePlus.LeafOpening,
		NewNextFreeNode: int(p.NextFreeNode) + 1,
	}

	// Progressively update and prove each leaf opening
	iInserted := p.NextFreeNode

	// 1/ The Prev
	newTupleMinus := tupleMinus.CopyWithNext(int64(iInserted))
	trace.ProofMinus = p.upsertTuple(iMinus, newTupleMinus)

	// 2/ The `New` (i.e) the inserted leaf
	insertedTuple := KVOpeningTuple[K, V]{
		Key: key, Value: val,
		LeafOpening: LeafOpening{
			Prev: int64(iMinus),
			Next: int64(iPlus),
			HKey: Hash(p.Config(), key),
			HVal: Hash(p.Config(), val),
		},
	}
	trace.ProofNew = p.upsertTuple(iInserted, insertedTuple)

	// 3/ The next
	newTuplePlus := tuplePlus.CopyWithPrev(int64(iInserted))
	trace.ProofPlus = p.upsertTuple(iPlus, newTuplePlus)

	// Fetch the root, and we are done
	p.NextFreeNode++ // And increment the next free node counter
	trace.NewSubRoot = p.SubTreeRoot()
	return trace
}

// VerifyInsertion audit the insertion of an entry in the accumulator w.r.t to
// the state of the verifier. It returns an error if the verification failed.
func (v *VerifierState[K, V]) VerifyInsertion(trace InsertionTrace[K, V]) error {

	// If the location does not match the we return an error
	if v.Location != trace.Location {
		return fmt.Errorf("inconsistent location : %v != %v", v.Location, trace.Location)
	}

	// Check that verifier's root is the same as the one in the traces
	if v.SubTreeRoot != trace.OldSubRoot {
		return fmt.Errorf("inconsistent root %v != %v", v.SubTreeRoot, trace.OldSubRoot)
	}

	iMinus := int64(trace.ProofMinus.Path)
	iInserted := int64(trace.ProofNew.Path)
	iPlus := int64(trace.ProofPlus.Path)

	// Alledgedly, we write in the leaf #nextFreeNode
	if iInserted != v.NextFreeNode {
		return fmt.Errorf(
			"writing in the wrong place : %v (expected %v)",
			iInserted, v.NextFreeNode,
		)
	}

	// Check that the two sandwich leaf opening point to each other
	if trace.OldOpenMinus.Next != iPlus || trace.OldOpenPlus.Prev != iMinus {
		return fmt.Errorf(
			"bad sandwich prev and next do not point to each other : prev %++v, next %++v",
			trace.OldOpenMinus, trace.OldOpenPlus,
		)
	}

	// Also checks that the their hkey are lower/larger than the inserted one
	hkey := Hash(v.Config, trace.Key)
	if Bytes32Cmp(hkey, trace.OldOpenMinus.HKey) < 1 || Bytes32Cmp(hkey, trace.OldOpenPlus.HKey) > -1 {
		return fmt.Errorf(
			"sandwich is incorrect expected %x < %x < %x",
			trace.OldOpenMinus.HKey, hkey, trace.OldOpenPlus.HKey,
		)
	}

	currentRoot := trace.OldSubRoot

	// Audit the update of the "minus"
	oldLeafMinus := trace.OldOpenMinus.Hash(v.Config)
	newLeafMinus := trace.OldOpenMinus.CopyWithNext(v.NextFreeNode).Hash(v.Config)

	currentRoot, err := updateCheckRoot(v.Config, trace.ProofMinus, currentRoot, oldLeafMinus, newLeafMinus)
	if err != nil {
		return fmt.Errorf("audit of the update of oldLeafMinus failed %v", err)
	}

	// Audit the update of the inserted new leaf
	newLeaf := LeafOpening{
		Prev: int64(trace.ProofMinus.Path),
		Next: int64(trace.ProofPlus.Path),
		HKey: hkey,
		HVal: Hash(v.Config, trace.Val),
	}.Hash(v.Config)

	currentRoot, err = updateCheckRoot(v.Config, trace.ProofNew, currentRoot, smt.EmptyLeaf(), newLeaf)
	if err != nil {
		return fmt.Errorf("audit of the update of the middle leaf failed %v", err)
	}

	// Audit the update of the "plus"
	oldLeafPlus := trace.OldOpenPlus.Hash(v.Config)
	newLeafPlus := trace.OldOpenPlus.CopyWithPrev(iInserted).Hash(v.Config)

	currentRoot, err = updateCheckRoot(v.Config, trace.ProofPlus, currentRoot, oldLeafPlus, newLeafPlus)
	if err != nil {
		return fmt.Errorf("audit of the update of oldLeafPlus failed %v", err)
	}

	// Check that the alleged new root is consistent with the one we reconstructed
	if currentRoot != trace.NewSubRoot {
		return fmt.Errorf("inconsistent root %v != %v", currentRoot, trace.NewSubRoot)
	}

	// Every check passed : update the verifier state
	v.SubTreeRoot = currentRoot
	v.NextFreeNode++
	// Check that the next free node matches with the prover
	if v.NextFreeNode != int64(trace.NewNextFreeNode) {
		return fmt.Errorf("inconsistent next free node %v != %v", v.NextFreeNode, trace.NewNextFreeNode)
	}
	return nil
}

// DeferMerkleChecks implements [Trace]
func (trace InsertionTrace[K, V]) DeferMerkleChecks(
	config *smt.Config,
	appendTo []smt.ProvedClaim,
) []smt.ProvedClaim {

	iInserted := int64(trace.ProofNew.Path)

	// Also checks that the their hkey are lower/larger than the inserted one
	hkey := Hash(config, trace.Key)

	currentRoot := trace.OldSubRoot

	// Audit the update of the "minus"
	oldLeafMinus := trace.OldOpenMinus.Hash(config)
	newLeafMinus := trace.OldOpenMinus.CopyWithNext(int64(trace.ProofNew.Path)).Hash(config)

	appendTo, currentRoot = deferCheckUpdateRoot(config, trace.ProofMinus, currentRoot, oldLeafMinus, newLeafMinus, appendTo)

	// Audit the update of the inserted new leaf
	newLeaf := LeafOpening{
		Prev: int64(trace.ProofMinus.Path),
		Next: int64(trace.ProofPlus.Path),
		HKey: hkey,
		HVal: Hash(config, trace.Val),
	}.Hash(config)

	appendTo, currentRoot = deferCheckUpdateRoot(config, trace.ProofNew, currentRoot, smt.EmptyLeaf(), newLeaf, appendTo)

	// Audit the update of the "plus"
	oldLeafPlus := trace.OldOpenPlus.Hash(config)
	newLeafPlus := trace.OldOpenPlus.CopyWithPrev(iInserted).Hash(config)

	appendTo, _ = deferCheckUpdateRoot(config, trace.ProofPlus, currentRoot, oldLeafPlus, newLeafPlus, appendTo)
	return appendTo
}

func (trace InsertionTrace[K, V]) HKey(cfg *smt.Config) Bytes32 {
	return Hash(cfg, trace.Key)
}

func (trace InsertionTrace[K, V]) RWInt() int {
	return 1
}
