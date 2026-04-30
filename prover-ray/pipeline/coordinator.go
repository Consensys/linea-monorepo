package pipeline

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
	"github.com/consensys/linea-monorepo/prover-ray/distributed"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// Coordinator is the thin orchestration layer. It does two things:
//
//  1. Tracks incoming PreflightCommitTask results. When all N expected
//     preflight commits for a block have arrived, it computes the shared
//     randomness (a single hash operation) and fires a SharedRandomnessEvent.
//     This is the only coordination point in the entire pipeline.
//
//  2. Tracks incoming GL/LPP proof paths. When any two proofs are available,
//     it enqueues a MergeTask. The merge tree grows incrementally; no barrier
//     waits for all proofs before the first merge can start.
//
// The Coordinator holds no proving state and does no heavy computation.
// It can run on the smallest machine in the cluster.
//
// Fault model: if the Coordinator restarts, it replays WorkerResult messages
// from the queue's durable log and reconstructs its in-memory state.
// PSEUDO: the real implementation would persist state to a durable store.
type Coordinator struct {
	mu sync.Mutex

	// preflightState tracks, per block, how many preflight commits we are
	// waiting for and the commitments received so far.
	preflightState map[string]*preflightTracker

	// mergeState tracks available proofs per block for merge pairing.
	mergeState map[string]*mergeTracker

	// queue is used to enqueue LPP and Merge tasks once conditions are met.
	queue TaskQueue

	// kinds is the compiled set of segment kinds (one per .bin file).
	kinds []*distributed.SegmentKind
}

// preflightTracker holds per-block preflight state.
type preflightTracker struct {
	expected     int // total number of preflight commits to wait for
	received     int
	commitments  []distributed.LPPCommitment
	lppWitnesses []*arithmetization.ModuleWitnessLPP // held until shared randomness is ready
}

// mergeTracker holds available proof paths waiting to be paired for merging.
type mergeTracker struct {
	available []string // proof paths not yet paired
	total     int      // total proofs expected (GL + LPP)
	merged    int      // proofs consumed by merge tasks
}

// NewCoordinator creates a Coordinator for the given segment kinds and queue.
func NewCoordinator(kinds []*distributed.SegmentKind, queue TaskQueue) *Coordinator {
	return &Coordinator{
		preflightState: make(map[string]*preflightTracker),
		mergeState:     make(map[string]*mergeTracker),
		queue:          queue,
		kinds:          kinds,
	}
}

// RegisterBlock tells the Coordinator how many preflight commits and total
// proofs to expect for a given block. Called by the producer (Prove) before
// any tasks are enqueued.
func (c *Coordinator) RegisterBlock(blockID string, totalSegments, totalProofs int, lppWitnesses []*arithmetization.ModuleWitnessLPP) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.preflightState[blockID] = &preflightTracker{
		expected:     totalSegments,
		commitments:  make([]distributed.LPPCommitment, 0, totalSegments),
		lppWitnesses: lppWitnesses,
	}
	c.mergeState[blockID] = &mergeTracker{
		total: totalProofs,
	}
}

// OnPreflightResult is called when a PreflightCommitTask completes.
// If all preflight commits for the block have now arrived, it fires the
// shared-randomness event: injects InitialFiatShamirState into every LPP
// witness and enqueues all LPP tasks. This is the only coordination point.
func (c *Coordinator) OnPreflightResult(result WorkerResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tracker := c.preflightState[result.BlockID]
	if result.Err != nil {
		// PSEUDO: re-enqueue the failed preflight task.
		// The retry logic lives in the queue implementation.
		return
	}

	tracker.commitments = append(tracker.commitments, *result.Commitment)
	tracker.received++

	if tracker.received < tracker.expected {
		return // still waiting for more preflight commits
	}

	// All preflight commits are in — compute shared randomness.
	// This is the only coordination point and is a single cheap hash.
	sharedRandomness := distributed.GetSharedRandomness(tracker.commitments)

	// Inject shared randomness into every LPP witness and enqueue LPP tasks.
	// From this point forward LPP tasks are fully independent of each other
	// and of all GL tasks.
	for i, w := range tracker.lppWitnesses {
		w.InitialFiatShamirState = sharedRandomness
		// PSEUDO: determine output path for this LPP proof.
		outPath := pseudoLPPProofPath(result.BlockID, w.ModuleIndex, w.SegmentIndex)
		_ = c.queue.EnqueueLPP(LPPProveTask{
			BlockID:      result.BlockID,
			KindIndex:    w.ModuleIndex,
			SegmentIndex: w.SegmentIndex,
			Kind:         c.kinds[w.ModuleIndex],
			Witness:      tracker.lppWitnesses[i],
			OutputPath:   outPath,
		})
	}

	delete(c.preflightState, result.BlockID) // free memory
}

// OnProofResult is called when a GLProveTask or LPPProveTask completes.
// It pairs available proofs and enqueues MergeTasks greedily — the merge
// tree grows as proofs trickle in, without waiting for all of them.
func (c *Coordinator) OnProofResult(result WorkerResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	mt := c.mergeState[result.BlockID]
	if result.Err != nil {
		// PSEUDO: re-enqueue the failed proof task.
		return
	}

	mt.available = append(mt.available, result.ProofPath)

	// Greedily pair any two available proofs into a merge task.
	// This means merging starts as soon as the first two proofs are ready,
	// not after all proofs finish. The binary tree grows bottom-up.
	for len(mt.available) >= 2 {
		left := mt.available[0]
		right := mt.available[1]
		mt.available = mt.available[2:]
		mt.merged += 2

		isFinal := (mt.merged == mt.total) && len(mt.available) == 0
		outPath := pseudoMergeProofPath(result.BlockID, mt.merged)
		_ = c.queue.EnqueueMerge(MergeTask{
			BlockID:     result.BlockID,
			LeftPath:    left,
			RightPath:   right,
			OutputPath:  outPath,
			Kind:        c.kinds[0], // PSEUDO: real merge kind depends on proof types
			IsFinalStep: isFinal,
		})
	}
}

// OnMergeResult is called when a MergeTask completes. The merged proof is
// treated as a new available proof and fed back into OnProofResult for further
// pairing — the same greedy logic applies at every level of the tree.
func (c *Coordinator) OnMergeResult(result WorkerResult) {
	c.OnProofResult(result)
}

// ---------------------------------------------------------------------------
// Pseudo stubs
// ---------------------------------------------------------------------------

func pseudoLPPProofPath(blockID string, kindIdx, segIdx int) string {
	return blockID + "/lpp/" + itoa(kindIdx) + "/" + itoa(segIdx) + ".proof"
}

func pseudoMergeProofPath(blockID string, mergedCount int) string {
	return blockID + "/merge/" + itoa(mergedCount) + ".proof"
}

func itoa(n int) string {
	// PSEUDO: use strconv.Itoa in real code
	_ = n
	return "N"
}

func pseudoLPPWitnessFromEvent(_ SharedRandomnessEvent, _ field.Octuplet) []*arithmetization.ModuleWitnessLPP {
	panic("pseudo")
}
