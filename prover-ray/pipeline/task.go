package pipeline

import (
	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
	"github.com/consensys/linea-monorepo/prover-ray/distributed"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// Tasks are the atomic units of work in the proving pipeline. Every task is:
//   - Stateless: all inputs are carried in the task struct (or referenced by path).
//   - Idempotent: re-running a task produces the same output; safe to retry.
//   - Independently schedulable: any worker on any machine can execute any task.
//
// There is no shared memory between tasks. Workers communicate only through the
// task queue and a lightweight output store (e.g. a shared filesystem or object
// store addressed by block/kind/segment).
//
// The only coordination event in the entire pipeline is:
//
//	"all PreflightCommitTasks for a given block are done
//	 → compute shared randomness
//	 → enqueue LPPProveTasks"
//
// This is a single cheap hash operation on the coordinator, not a proving sync.

// PreflightCommitTask commits to the FSSchedule[0] columns of one segment.
// It is enqueued by arithmetization as soon as those columns are ready,
// independently of every other segment.
//
// Output: an LPPCommitment (stored by the coordinator; never written to disk).
type PreflightCommitTask struct {
	BlockID      string
	KindIndex    int
	SegmentIndex int
	// Segment carries the FSSchedule[0] column data in-memory (small; only the
	// looked-up columns). For very large deployments this could be a path
	// reference instead.
	Segment arithmetization.PreflightSegment
}

// GLProveTask proves one GL segment witness.
// It is enqueued as soon as the segment witness is available, regardless of
// shared randomness or any other segment's status.
//
// Output path: OutputPath (written atomically; safe to retry).
type GLProveTask struct {
	BlockID      string
	KindIndex    int
	SegmentIndex int
	Kind         *distributed.SegmentKind
	Witness      *arithmetization.ModuleWitnessGL
	OutputPath   string // where to write the resulting SegmentProof
}

// LPPProveTask proves one LPP segment witness.
// It is enqueued only after shared randomness is available (the single
// coordination event). Once enqueued it is fully independent of all other
// LPP tasks and all GL tasks.
//
// Output path: OutputPath (written atomically; safe to retry).
type LPPProveTask struct {
	BlockID        string
	KindIndex      int
	SegmentIndex   int
	Kind           *distributed.SegmentKind
	Witness        *arithmetization.ModuleWitnessLPP // InitialFiatShamirState already set
	OutputPath     string
}

// MergeTask merges two segment proofs into one via one step of hierarchical
// conglomeration. It can be enqueued as soon as any two proofs are available —
// the merge tree grows incrementally without waiting for all leaves.
//
// Output path: OutputPath.
type MergeTask struct {
	BlockID     string
	LeftPath    string // path to first input proof
	RightPath   string // path to second input proof
	OutputPath  string // path for merged output proof
	Kind        *distributed.SegmentKind
	IsFinalStep bool // true for the root merge → outer proof extraction follows
}

// TaskQueue is the interface to whatever queuing backend is used (in-process
// channel, Redis, SQS, …). Workers call Dequeue in a loop; producers call
// Enqueue. The queue guarantees at-least-once delivery.
//
// PSEUDO: real implementation would use a persistent queue so that tasks
// survive worker crashes and machine restarts.
type TaskQueue interface {
	EnqueuePreflight(t PreflightCommitTask) error
	EnqueueGL(t GLProveTask) error
	EnqueueLPP(t LPPProveTask) error
	EnqueueMerge(t MergeTask) error
}

// WorkerResult is what a worker emits after completing any task.
// The coordinator uses these to decide when to fire the shared-randomness
// event and when to enqueue merge tasks.
type WorkerResult struct {
	BlockID      string
	KindIndex    int
	SegmentIndex int
	// Commitment is populated for completed PreflightCommitTasks.
	Commitment *distributed.LPPCommitment
	// ProofPath is populated for completed GL/LPP/Merge tasks.
	ProofPath string
	// Err is non-nil if the task failed. The coordinator re-enqueues the task.
	Err error
}

// SharedRandomnessEvent is fired by the coordinator exactly once per block,
// as soon as all PreflightCommitTasks for that block have succeeded.
// The event carries the shared randomness and the full LPP witness set so
// the coordinator can immediately enqueue all LPP tasks.
type SharedRandomnessEvent struct {
	BlockID         string
	SharedRandomness field.Octuplet
	LPPWitnesses    []*arithmetization.ModuleWitnessLPP
}
