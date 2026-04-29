package sumcheck

import (
	"runtime"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// evalSubChunkSize is the number of boolean-hypercube points processed in one
// inner-loop batch. Matches the legacy tuning constant (128).
const evalSubChunkSize = 128

// eqChunkSize is the number of eq-table entries computed per parallel task
// during eq-table construction.
const eqChunkSize = 256

// ThreadScratch holds the per-thread working buffers used inside
// [ProverState.ComputeRoundPoly]. All slices are pre-allocated in
// [NewProverConfig] and reused across rounds.
type ThreadScratch struct {
	// GateOut holds gate outputs for one sub-chunk.
	GateOut []field.Ext
	// Eqs holds the incremented eq values during the t≥2 loop.
	Eqs []field.Ext
	// DEqs holds the per-entry eq deltas (eq_top − eq_bottom).
	DEqs []field.Ext
	// Xs holds the gate inputs (all columns packed flat: column-major order,
	// stride = evalSubChunkSize). At t=1 they are set to table[mid:]; for
	// t≥2 each entry is advanced by DXs.
	Xs []field.Ext
	// DXs holds the per-entry column deltas (top − bottom).
	DXs []field.Ext
	// EvalBuf is a slice of sub-slices into Xs, one per gate input column,
	// passed to Gate.EvalBatch.
	EvalBuf [][]field.Ext
	// Accum accumulates the partial round polynomial for tasks owned by this
	// thread: Accum[0] = ΣP(0), Accum[1] = ΣP(2), …, Accum[d−1] = ΣP(d).
	Accum []field.Ext
}

// ProverConfig holds all pre-allocated scratch memory for a sumcheck prover.
// A single ProverConfig can be reused across multiple sequential proofs
// (of the same or smaller logN / nInputs).
// It is not safe for concurrent use from multiple provers simultaneously.
type ProverConfig struct {
	// EqScratch is the full 2^logN eq table for the current proof.
	EqScratch []field.Ext
	// EqTmp is used when accumulating additional eq tables in multi-sumcheck.
	EqTmp []field.Ext
	// Tables holds the ext-field copies of the prover's input tables,
	// one slice per column. Tables[k] is used for round [0, logN).
	Tables [][]field.Ext
	// PerThread is indexed by goroutine ID.
	PerThread []ThreadScratch
	// NumCPU is the parallelism cap passed to ExecuteThreadAware.
	NumCPU int
}

// NewProverConfig allocates all buffers needed for proofs with up to logN
// variables and at most maxInputs gate-input columns.
// numCPU = 0 means runtime.GOMAXPROCS(0).
func NewProverConfig(logN, maxInputs, numCPU int) *ProverConfig {
	if numCPU <= 0 {
		numCPU = runtime.GOMAXPROCS(0)
	}
	n := 1 << logN

	tables := make([][]field.Ext, maxInputs)
	for i := range tables {
		tables[i] = make([]field.Ext, n)
	}

	perThread := make([]ThreadScratch, numCPU)
	for i := range perThread {
		perThread[i] = ThreadScratch{
			GateOut: make([]field.Ext, evalSubChunkSize),
			Eqs:     make([]field.Ext, evalSubChunkSize),
			DEqs:    make([]field.Ext, evalSubChunkSize),
			Xs:      make([]field.Ext, evalSubChunkSize*maxInputs),
			DXs:     make([]field.Ext, evalSubChunkSize*maxInputs),
			EvalBuf: make([][]field.Ext, maxInputs),
		}
	}

	return &ProverConfig{
		EqScratch: make([]field.Ext, n),
		EqTmp:     make([]field.Ext, n),
		Tables:    tables,
		PerThread: perThread,
		NumCPU:    numCPU,
	}
}
