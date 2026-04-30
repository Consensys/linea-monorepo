// Package pipeline wires the arithmetization layer to the distributed prover
// using a task-queue model with no global synchronization points.
//
// # Why no sync points
//
// A synchronized pipeline (wait for all GL proofs → compute shared randomness
// → start LPP proofs) has three failure modes the thread identified:
//
//  1. One crashed prover stalls the whole block.
//  2. Memory must be held on one machine until every proof is done.
//  3. The slowest segment sets the wall-clock time for the whole batch
//     (straggler problem).
//
// The task-queue model avoids all three:
//  - Tasks are idempotent: a crashed worker re-enqueues its task; no state is lost.
//  - Tasks are stateless: inputs/outputs live in a shared store (filesystem,
//    object store); no machine needs to hold the full block in memory.
//  - Workers steal from the queue: a fast machine that finishes its segment
//    immediately picks up the next one; no machine idles waiting for a straggler.
//
// # The one unavoidable coordination event
//
// Shared randomness requires all preflight commits (a protocol requirement).
// This is handled as a single lightweight event in the Coordinator:
//
//	all PreflightCommitTasks done → hash commitments → enqueue LPP tasks
//
// The hash takes microseconds. GL tasks are never gated on this event.
//
// # Data flow (no barriers)
//
//	Arithmetization
//	  │
//	  ├─ preflight segment ready → enqueue PreflightCommitTask (per segment)
//	  │                                         │
//	  │                           Worker commits → Coordinator.OnPreflightResult
//	  │                                         │
//	  │                           [all done] → compute shared randomness
//	  │                                         │
//	  │                                         └─ enqueue LPPProveTask (per segment)
//	  │
//	  └─ full witness ready → enqueue GLProveTask (per segment)
//	                                    │
//	                        Worker proves → Coordinator.OnProofResult
//	                                    │
//	                        [any 2 proofs] → enqueue MergeTask (greedy pairing)
//	                                    │
//	                        [root merge] → final proof
package pipeline

import (
	"context"

	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
	"github.com/consensys/linea-monorepo/prover-ray/distributed"
)

// Prover is the top-level entry point. It holds no per-block state; all
// per-block state lives in the Coordinator and the task queue.
type Prover struct {
	kinds       []*distributed.SegmentKind
	arith       arithmetization.Arithmetization
	queue       TaskQueue
	coordinator *Coordinator
}

// NewProver initialises a Prover from N compiled segment kinds.
// Both arithmetization and the prover derived the same column structure by
// compiling the same .bin files; no circuit internals cross the boundary.
func NewProver(kinds []*distributed.SegmentKind, queue TaskQueue) *Prover {
	coord := NewCoordinator(kinds, queue)
	p := &Prover{kinds: kinds, queue: queue, coordinator: coord}
	p.arith.Configure(distributed.ArithmetizationConfig(kinds))
	return p
}

// Prove submits a block for proving. It returns as soon as all tasks have been
// enqueued; the actual proving happens asynchronously on the worker pool.
// The final proof is written to the output store addressed by blockID.
//
// If proving is interrupted (crash, restart), call Prove again with the same
// blockID: the Coordinator will replay results from the durable queue log and
// resume without re-proving already-completed segments.
func (p *Prover) Prove(ctx context.Context, blockID string, req ProveRequest) error {

	// --- Launch arithmetization ------------------------------------------
	// preflightCh: one []PreflightSegment per block, sent early.
	// traceCh:     one TracingResult per block, sent when done.
	preflightCh, traceCh := p.arith.Run(req.TracePath)

	// --- Stream preflight tasks as they arrive ---------------------------
	// We don't wait for the full trace. Each preflight segment is dispatched
	// to the queue as soon as it lands. Workers can start committing
	// immediately, in parallel with arithmetization's full expansion pass.
	go func() {
		select {
		case <-ctx.Done():
			return
		case segs := <-preflightCh:
			// Register how many preflight commits the coordinator should expect.
			// LPP witnesses are not ready yet; they will be attached when the
			// full trace arrives. PSEUDO: this registration should be idempotent
			// for crash-recovery (the coordinator checks if already registered).
			p.coordinator.RegisterBlock(blockID, len(segs), pseudoTotalProofs(segs, p.kinds), nil)

			for _, seg := range segs {
				_ = p.queue.EnqueuePreflight(PreflightCommitTask{
					BlockID:      blockID,
					KindIndex:    seg.ModuleIndex,
					SegmentIndex: seg.SegmentIndex,
					Segment:      seg,
				})
			}
		}
	}()

	// --- Stream GL tasks as full witnesses arrive ------------------------
	// GL tasks do not depend on shared randomness. They are enqueued as soon
	// as each witness is ready — independently of any preflight result.
	go func() {
		select {
		case <-ctx.Done():
			return
		case result := <-traceCh:
			// Attach the LPP witnesses to the coordinator so it can enqueue
			// LPP tasks once shared randomness is available.
			// PSEUDO: update registration rather than overwrite.
			p.coordinator.RegisterBlock(blockID, 0, 0, result.WitnessesLPP)

			for _, w := range result.WitnessesGL {
				outPath := pseudoGLProofPath(blockID, w.ModuleIndex, w.SegmentIndex)
				_ = p.queue.EnqueueGL(GLProveTask{
					BlockID:      blockID,
					KindIndex:    w.ModuleIndex,
					SegmentIndex: w.SegmentIndex,
					Kind:         p.kinds[w.ModuleIndex],
					Witness:      w,
					OutputPath:   outPath,
				})
			}
			// LPP tasks are enqueued by the Coordinator once shared randomness
			// is available — not from here.
		}
	}()

	return nil // task submission complete; proving is async
}

// Worker is a stateless prover process. It pulls tasks from the queue in a
// loop and reports results to the Coordinator. Multiple workers can run on
// multiple machines simultaneously; they share no memory.
//
// A crashed worker simply stops pulling tasks. The queue re-delivers any
// in-flight tasks (at-least-once semantics). Tasks are idempotent so
// re-execution is safe.
type Worker struct {
	queue       TaskQueue
	coordinator *Coordinator
}

// Run is the worker main loop. It blocks until ctx is cancelled.
// PSEUDO: real implementation pulls tasks from the queue backend.
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// PSEUDO: pull one task from the queue (blocking dequeue with timeout).
			task := pseudoDequeue(w.queue)
			result := w.execute(ctx, task)
			w.report(result)
		}
	}
}

// execute runs a single task and returns a WorkerResult.
// All task types are handled here; the worker doesn't know or care what block
// is being proven or what machine the coordinator runs on.
func (w *Worker) execute(ctx context.Context, task any) WorkerResult {
	_ = ctx
	switch t := task.(type) {

	case PreflightCommitTask:
		commitment := distributed.CommitLPPColumns(t.Segment)
		return WorkerResult{
			BlockID:      t.BlockID,
			KindIndex:    t.KindIndex,
			SegmentIndex: t.SegmentIndex,
			Commitment:   &commitment,
		}

	case GLProveTask:
		// PSEUDO: run the GL circuit prover for t.Kind using t.Witness.
		// Write the proof atomically to t.OutputPath.
		pseudoProveGL(t)
		return WorkerResult{
			BlockID:      t.BlockID,
			KindIndex:    t.KindIndex,
			SegmentIndex: t.SegmentIndex,
			ProofPath:    t.OutputPath,
		}

	case LPPProveTask:
		// PSEUDO: run the LPP circuit prover for t.Kind using t.Witness.
		// InitialFiatShamirState is already set in t.Witness by the coordinator.
		pseudoProveLPP(t)
		return WorkerResult{
			BlockID:      t.BlockID,
			KindIndex:    t.KindIndex,
			SegmentIndex: t.SegmentIndex,
			ProofPath:    t.OutputPath,
		}

	case MergeTask:
		// PSEUDO: load left and right proofs from store, run one conglomeration
		// step, write merged proof to t.OutputPath.
		pseudoMerge(t)
		return WorkerResult{
			BlockID:   t.BlockID,
			ProofPath: t.OutputPath,
		}

	default:
		panic("unknown task type")
	}
}

// report routes the result to the appropriate Coordinator handler.
func (w *Worker) report(result WorkerResult) {
	if result.Commitment != nil {
		w.coordinator.OnPreflightResult(result)
	} else if result.ProofPath != "" {
		// PSEUDO: distinguish merge from segment proofs by path prefix.
		if pseudoIsMerge(result.ProofPath) {
			w.coordinator.OnMergeResult(result)
		} else {
			w.coordinator.OnProofResult(result)
		}
	}
}

// ---------------------------------------------------------------------------
// Pseudo stubs
// ---------------------------------------------------------------------------

func pseudoTotalProofs(_ []arithmetization.PreflightSegment, _ []*distributed.SegmentKind) int {
	panic("pseudo: total = nbGLSegments + nbLPPSegments")
}
func pseudoGLProofPath(blockID string, kindIdx, segIdx int) string {
	return blockID + "/gl/" + itoa(kindIdx) + "/" + itoa(segIdx) + ".proof"
}
func pseudoDequeue(_ TaskQueue) any          { panic("pseudo: blocking dequeue from queue backend") }
func pseudoProveGL(_ GLProveTask)            { panic("pseudo: run GL circuit prover") }
func pseudoProveLPP(_ LPPProveTask)          { panic("pseudo: run LPP circuit prover") }
func pseudoMerge(_ MergeTask)               { panic("pseudo: load + merge two proofs") }
func pseudoIsMerge(_ string) bool           { panic("pseudo: check path prefix") }
