package vortex

// Per-round commit handle stored under VortexProverStateName(round) in the
// Wizard prover runtime State map. A round's committed matrix lives in
// exactly one of two places:
//
//   *committedHandle.host  — vortex_koalabear.EncodedMatrix (= []SmartVector)
//                            BLS rounds, NoSIS rounds, precomputeds, and
//                            SIS rounds when the GPU path is disabled.
//
//   *committedHandle.gpu   — *gpuvortex.CommitState
//                            SIS-applied Koala rounds when
//                            LIMITLESS_GPU_VORTEX=1 and a GPU is bound.
//
// Why this exists
// ───────────────
// Before this type, the value stored in run.State was the raw EncodedMatrix.
// The GPU "drop-in" CommitMerkleWithSIS therefore had to D2H the entire
// encoded matrix (8 GiB at 2^27 segment size) and reconstruct it as a
// []SmartVector in Go-managed memory before the prover state could hold
// it. That reconstruction cost ~1.4 s of pure host-side work — enough to
// turn a 4× GPU win into a 0.65× *regression*.
//
// With this handle, the GPU SIS path skips the full D2H. The encoded
// matrix stays on device; downstream actions (LinComb, Open) call
// device-resident methods that produce only the small outputs they need
// (UAlpha vector ~16 MiB, selected columns ~few MiB).
//
// Lifecycle
// ─────────
// 1. ColumnAssignmentProverAction.Run inserts a handle.
// 2. LinearCombinationComputationProverAction.Run reads (no mutation).
// 3. OpenSelectedColumnsProverAction.Run reads, then for GPU handles
//    calls FreeGPU() once columns are extracted.

import (
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	gpuvortex "github.com/consensys/linea-monorepo/prover/gpu/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

type committedHandle struct {
	gpu  *gpuvortex.CommitState         // non-nil when device-resident
	host vortex_koalabear.EncodedMatrix // non-nil when host-resident
}

func newHostHandle(m vortex_koalabear.EncodedMatrix) *committedHandle {
	return &committedHandle{host: m}
}

func newGPUHandle(cs *gpuvortex.CommitState) *committedHandle {
	return &committedHandle{gpu: cs}
}

// isGPU reports whether the matrix is device-resident.
func (h *committedHandle) isGPU() bool { return h.gpu != nil }

// numRows is the row count without forcing a host materialization.
func (h *committedHandle) numRows() int {
	if h.gpu != nil {
		return h.gpu.NRows()
	}
	return len(h.host)
}

// hostMatrix returns the matrix as []SmartVector. For GPU handles this
// triggers a full D2H — used only by callers that genuinely need the
// host-side encoded matrix (self-recursion / debug). The two hot actions
// (LinComb and Open) avoid this and call into gpu/vortex directly.
func (h *committedHandle) hostMatrix() vortex_koalabear.EncodedMatrix {
	if h.gpu != nil {
		return h.gpu.GetEncodedMatrix()
	}
	return h.host
}

// extractColumns returns columns[entryIdx][rowIdx] for each entry in
// entries. For GPU handles this issues a small D2H of only the selected
// columns (typically O(few MiB) regardless of matrix size). For host
// handles it gathers from the SmartVector slice.
//
// Used by OpenSelectedColumnsProverAction to fill proof.Columns without
// materializing the full encoded matrix.
func (h *committedHandle) extractColumns(entries []int) [][]field.Element {
	if h.gpu != nil {
		// gpu.ExtractColumns may fail if the device pipeline was already
		// freed; the OpenSelectedColumns action is the only caller and
		// runs before free(), so we surface the error via panic to keep
		// the call sites simple.
		cols, err := h.gpu.ExtractColumns(entries)
		if err != nil {
			panic("vortex: GPU ExtractColumns: " + err.Error())
		}
		return cols
	}
	out := make([][]field.Element, len(entries))
	for i, c := range entries {
		col := make([]field.Element, len(h.host))
		for r, row := range h.host {
			col[r] = row.Get(c)
		}
		out[i] = col
	}
	return out
}

// free releases device buffers. Idempotent. No-op for host handles.
func (h *committedHandle) free() {
	if h.gpu != nil {
		h.gpu.FreeGPU()
	}
}

// asHandle promotes a value read from run.State into a *committedHandle.
// Accepts either:
//   - the new *committedHandle wrapper (preferred path)
//   - a raw vortex_koalabear.EncodedMatrix (legacy callers / non-SIS rounds
//     that haven't been migrated yet)
//
// Returns nil if v is neither.
func asHandle(v any) *committedHandle {
	switch x := v.(type) {
	case *committedHandle:
		return x
	case vortex_koalabear.EncodedMatrix:
		return newHostHandle(x)
	}
	return nil
}
