//go:build cuda

package plonk

// MSM (Multi-Scalar Multiplication) implements GPU-accelerated Pippenger's
// bucket method for computing Q = Σᵢ sᵢ · Pᵢ over BLS12-377 G1.
//
// Architecture:
//   - Public input is compact Twisted Edwards (G1TEPoint, 96 bytes), as
//     produced by ConvertG1AffineToTE or read from disk dumps.
//   - At pin time we expand each point to the precomputed format
//     (Y-X, Y+X, 2d·X·Y) = 144 bytes, matching the GPU G1EdYZD struct.
//     This trades 50% larger pinned/VRAM points for 2 fewer fp_mul per
//     mixed-add (9M → 7M).
//   - G1MSM owns the pinned points and manages GPU uploads internally
//   - GPU pipeline: decompose → sort → accumulate → reduce (see msm.cu)
//   - Host: Horner combination of per-window TE results → single TE→Jacobian
//
// Chunked MSM for large sizes (≥ msmChunkThreshold):
//
//   Points are split into two halves. MultiExp processes all scalar sets
//   against the first half, swaps to the second half, then combines:
//
//     partial₁[i] = Σⱼ₌₀ᵐ⁻¹     s[i][j] · P[j]        (first half)
//     partial₂[i] = Σⱼ₌ₘⁿ⁻¹     s[i][j] · P[j]        (second half)
//     result[i]   = partial₁[i] + partial₂[i]
//
//   This halves the VRAM needed for sort buffers (the dominant allocation).
//   The state machine tracks which half is currently loaded on GPU.

/*
#include "gnark_gpu.h"
*/
import "C"
import (
	"fmt"
	"io"
	"math/big"
	"runtime"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// frRInv is R^{-1} mod q where R = 2^256 mod q (the Fr Montgomery constant).
// Used to correct MSM results when skipping Montgomery reduction on GPU.
// Internal Montgomery representation is {1, 0, 0, 0} since R^{-1} * R = 1.
var frRInv big.Int

func init() {
	var rInv fr.Element
	rInv[0] = 1 // Montgomery representation of R^{-1}: R^{-1} * R mod q = 1
	rInv.BigInt(&frRInv)
}

// Maximum MSM window parameters. Actual values are queried from GPU context.
// c=17 → 15 windows. Buffer sized for up to 20 windows to support future c changes.
const msmMaxWindows = 20

// msmChunkThreshold: split base points in 2 for sizes >= this value.
// This halves sort buffer VRAM at the cost of 2 GPU passes per MultiExp.
// Sizes 2^23..2^26 fit sort buffers without chunking; 2^27+ needs chunking.
const msmChunkThreshold = 1 << 27

// ─────────────────────────────────────────────────────────────────────────────
// G1MSMPoints — pinned host memory for pre-converted TE points
// ─────────────────────────────────────────────────────────────────────────────

// G1MSMPoints holds pre-converted TE points in pinned host memory in the
// GPU's precomputed format (144 bytes per point: Y-X, Y+X, 2d·X·Y).
// Reusable across MSM contexts and reload cycles.
type G1MSMPoints struct {
	pinned unsafe.Pointer // pinned host memory (18 uint64s per point, 144 bytes)
	N      int
}

// ConvertG1Points converts SW affine points to TE precomputed format and stores
// the result in pinned (page-locked) host memory for fast DMA transfers.
// This is the expensive CPU operation — call once, reuse for all reloads.
func ConvertG1Points(points []bls12377.G1Affine) (*G1MSMPoints, error) {
	n := len(points)
	if n == 0 {
		return nil, &gpu.Error{Code: -1, Message: "points must not be empty"}
	}

	// SW → compact TE (with batch inversion)
	tePoints := convertToEdMSM(points)

	// Allocate pinned host memory at the precomputed (144-byte) layout, then
	// expand each compact point in place so we never materialize a heap-allocated
	// []g1TEPrecompPoint at full SRS size.
	nbytes := C.size_t(n) * C.size_t(g1TEPrecompPointSize)
	var pinned unsafe.Pointer
	if err := toError(C.gnark_gpu_alloc_pinned(&pinned, nbytes)); err != nil {
		return nil, err
	}

	dst := unsafe.Slice((*g1TEPrecompPoint)(pinned), n)
	writePrecompFromCompact(dst, tePoints)

	pts := &G1MSMPoints{pinned: pinned, N: n}
	runtime.SetFinalizer(pts, (*G1MSMPoints).Free)
	return pts, nil
}

// Free releases pinned host memory.
func (p *G1MSMPoints) Free() {
	if p.pinned != nil {
		C.gnark_gpu_free_pinned(p.pinned)
		p.pinned = nil
		runtime.SetFinalizer(p, nil)
	}
}

// Len returns the number of points.
func (p *G1MSMPoints) Len() int {
	return p.N
}

// PinG1TEPoints copies compact TE points into pinned (page-locked) host
// memory and expands them to the GPU's precomputed (144-byte) format.
func PinG1TEPoints(points []G1TEPoint) (*G1MSMPoints, error) {
	n := len(points)
	if n == 0 {
		return nil, &gpu.Error{Code: -1, Message: "points must not be empty"}
	}
	nbytes := C.size_t(n) * C.size_t(g1TEPrecompPointSize)
	var pinned unsafe.Pointer
	if err := toError(C.gnark_gpu_alloc_pinned(&pinned, nbytes)); err != nil {
		return nil, err
	}
	dst := unsafe.Slice((*g1TEPrecompPoint)(pinned), n)
	writePrecompFromCompact(dst, points)
	pts := &G1MSMPoints{pinned: pinned, N: n}
	runtime.SetFinalizer(pts, (*G1MSMPoints).Free)
	return pts, nil
}

// ReadG1TEPointsPinned reads n compact TE points from a raw memory dump
// (96 bytes/point on disk) and stores them in pinned host memory expanded
// to the GPU's precomputed (144-byte) format.
//
// To avoid allocating ~12 GB of heap for a 128M-point SRS, we read in
// fixed-size batches into a small reusable buffer and expand each batch
// directly into pinned memory.
func ReadG1TEPointsPinned(r io.Reader, n int) (*G1MSMPoints, error) {
	if n <= 0 {
		return nil, &gpu.Error{Code: -1, Message: "count must be positive"}
	}
	nbytes := C.size_t(n) * C.size_t(g1TEPrecompPointSize)
	var pinned unsafe.Pointer
	if err := toError(C.gnark_gpu_alloc_pinned(&pinned, nbytes)); err != nil {
		return nil, err
	}
	dst := unsafe.Slice((*g1TEPrecompPoint)(pinned), n)

	const batchPoints = 1 << 16 // 64K points per batch ≈ 6 MB compact buffer
	batch := make([]G1TEPoint, batchPoints)
	read := 0
	for read < n {
		take := batchPoints
		if remaining := n - read; remaining < take {
			take = remaining
		}
		buf := unsafe.Slice((*byte)(unsafe.Pointer(&batch[0])), take*g1TEPointSize)
		if _, err := io.ReadFull(r, buf); err != nil {
			C.gnark_gpu_free_pinned(pinned)
			return nil, err
		}
		writePrecompFromCompact(dst[read:read+take], batch[:take])
		read += take
	}

	pts := &G1MSMPoints{pinned: pinned, N: n}
	runtime.SetFinalizer(pts, (*G1MSMPoints).Free)
	return pts, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// G1MSM — GPU MSM context with owned points and automatic chunking
// ─────────────────────────────────────────────────────────────────────────────

// G1MSM holds GPU resources for multi-scalar multiplication.
//
// Owns the base points (G1MSMPoints, pinned host memory). For large point sets
// (≥ msmChunkThreshold), splits into two halves to halve sort buffer VRAM.
//
// Montgomery reduction is always skipped in the GPU kernel (no per-scalar
// fr_from_mont). Instead, a single host-side EC scalar mul by R^{-1} corrects
// the result. This saves n field multiplications per MSM call.
//
// Points loaded state machine (chunked mode):
//
//	┌─────────┐  loadPart(1)  ┌─────────┐
//	│  first  │◄─────────────►│  second  │
//	│  half   │  loadPart(2)  │  half    │
//	└─────────┘               └─────────┘
type G1MSM struct {
	handle        C.gnark_gpu_msm_t
	dev           *gpu.Device
	pts           *G1MSMPoints // owned pinned points (freed on Close)
	n             int          // total usable points
	chunked       bool         // true when n >= msmChunkThreshold
	half          int          // GPU context capacity: ceil(n/2) when chunked, n otherwise
	loaded        int          // 0=none, 1=first half, 2=second half (chunked only)
	c             int          // window size in bits (queried from GPU context)
	numWindows    int          // ceil(253 / c) (queried from GPU context)
	windowResults [msmMaxWindows][24]uint64
}

// NewG1MSM creates an MSM context that owns the given pinned points.
// All points are usable. Call Close() to release GPU + pinned memory.
func NewG1MSM(dev *gpu.Device, pts *G1MSMPoints) (*G1MSM, error) {
	return NewG1MSMN(dev, pts, pts.N)
}

// NewG1MSMN creates an MSM context using the first maxN points.
// Useful when the pinned buffer contains more points than needed
// (e.g., canonical SRS sized for the largest circuit).
func NewG1MSMN(dev *gpu.Device, pts *G1MSMPoints, maxN int) (*G1MSM, error) {
	if devCtx(dev) == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if pts == nil || pts.pinned == nil || pts.N == 0 {
		return nil, &gpu.Error{Code: -1, Message: "points must not be empty"}
	}
	if maxN <= 0 || maxN > pts.N {
		maxN = pts.N
	}

	n := maxN
	chunked := n >= msmChunkThreshold
	half := n
	if chunked {
		half = (n + 1) / 2
	}

	// Create MSM context sized for half (or full if not chunked).
	// Sort buffers scale with this capacity, so chunking halves VRAM.
	var handle C.gnark_gpu_msm_t
	if err := toError(C.gnark_gpu_msm_create(devCtx(dev), C.size_t(half), &handle)); err != nil {
		return nil, err
	}

	// Initial point upload: first half (or all if not chunked).
	loadN := half
	if !chunked {
		loadN = n
	}
	if err := toError(C.gnark_gpu_msm_load_points(
		handle,
		(*C.uint64_t)(pts.pinned),
		C.size_t(loadN),
	)); err != nil {
		C.gnark_gpu_msm_destroy(handle)
		return nil, err
	}

	loaded := 0
	if chunked {
		loaded = 1 // first half loaded
	}

	msm := &G1MSM{
		handle:  handle,
		dev:     dev,
		pts:     pts,
		n:       n,
		chunked: chunked,
		half:    half,
		loaded:  loaded,
	}
	msm.queryConfig()
	runtime.SetFinalizer(msm, (*G1MSM).Close)
	return msm, nil
}

// queryConfig retrieves the MSM window configuration (c and numWindows) from the GPU context.
func (m *G1MSM) queryConfig() {
	var cVal, nw C.int
	C.gnark_gpu_msm_get_config(m.handle, &cVal, &nw)
	m.c = int(cVal)
	m.numWindows = int(nw)
}

// MSMPhase identifies a phase of the MSM pipeline. Order matches the
// gnark_gpu_msm_get_phase_timings return values.
type MSMPhase int

const (
	MSMPhaseH2D MSMPhase = iota
	MSMPhaseBuildPairs
	MSMPhaseSort
	MSMPhaseBoundaries
	MSMPhaseAccumSeq
	MSMPhaseAccumPar
	MSMPhaseReducePartial
	MSMPhaseReduceFinalize
	MSMPhaseD2H
	MSMPhaseCount
)

// String returns a short label for an MSM phase.
func (p MSMPhase) String() string {
	switch p {
	case MSMPhaseH2D:
		return "h2d"
	case MSMPhaseBuildPairs:
		return "build_pairs"
	case MSMPhaseSort:
		return "sort"
	case MSMPhaseBoundaries:
		return "boundaries"
	case MSMPhaseAccumSeq:
		return "accum_seq"
	case MSMPhaseAccumPar:
		return "accum_par"
	case MSMPhaseReducePartial:
		return "reduce_partial"
	case MSMPhaseReduceFinalize:
		return "reduce_finalize"
	case MSMPhaseD2H:
		return "d2h"
	}
	return "unknown"
}

// LastPhaseTimings returns per-phase timings (in milliseconds) of the most
// recent kernel run on this MSM context. Index by MSMPhase. Phases that did
// not run (e.g. accum_par when there were no overflow buckets) report 0.
func (m *G1MSM) LastPhaseTimings() [MSMPhaseCount]float32 {
	var out [MSMPhaseCount]C.float
	C.gnark_gpu_msm_get_phase_timings(m.handle, (*C.float)(unsafe.Pointer(&out[0])))
	var result [MSMPhaseCount]float32
	for i := 0; i < int(MSMPhaseCount); i++ {
		result[i] = float32(out[i])
	}
	return result
}

// PinWorkBuffers keeps the MSM's sort buffers and host registration alive
// across MultiExp calls. Use this around a wave of back-to-back MSMs to
// amortize ~5–10 ms of cudaMalloc/Free + cudaHostRegister/Unregister
// overhead per call.
//
// MUST be paired with ReleaseWorkBuffers before any phase that needs the
// reclaimed VRAM (e.g., the quotient computation).
//
// At n=2²⁷, sort buffers are ~50 GiB. Verify VRAM headroom before pinning.
func (m *G1MSM) PinWorkBuffers() error {
	return toError(C.gnark_gpu_msm_pin_work_buffers(m.handle))
}

// ReleaseWorkBuffers releases pinned work buffers immediately. Subsequent
// MultiExp calls re-allocate lazily.
func (m *G1MSM) ReleaseWorkBuffers() error {
	return toError(C.gnark_gpu_msm_release_work_buffers(m.handle))
}

// ─────────────────────────────────────────────────────────────────────────────
// MultiExp — the core MSM API
// ─────────────────────────────────────────────────────────────────────────────

// MultiExp computes Q[i] = Σⱼ scalars[i][j] · P[j] for each scalar set.
//
// Accepts a variadic list of scalar vectors. Each vector's length must be ≤ n.
// Returns a slice of G1Jac results with the same dimension as the input.
//
// For chunked mode (n ≥ msmChunkThreshold), all scalar sets are processed
// against the first half of points, then the second half, then combined.
// This batches the point-swap cost across multiple MSMs.
func (m *G1MSM) MultiExp(scalars ...[]fr.Element) []bls12377.G1Jac {
	k := len(scalars)
	if k == 0 {
		return nil
	}
	for i, s := range scalars {
		if len(s) > m.n {
			panic(fmt.Sprintf("gpu: MSM scalar set %d has %d elements, exceeds %d points", i, len(s), m.n))
		}
		if len(s) == 0 {
			panic(fmt.Sprintf("gpu: MSM scalar set %d is empty", i))
		}
	}

	if !m.chunked {
		results := make([]bls12377.G1Jac, k)
		for i, s := range scalars {
			results[i] = m.runSingle(s)
		}
		return results
	}

	// ── Chunked mode ──
	// Phase 1: compute partial results for first half of points
	m.loadPart(1)
	partials := make([]bls12377.G1Jac, k)
	for i, s := range scalars {
		firstN := len(s)
		if firstN > m.half {
			firstN = m.half
		}
		partials[i] = m.runSingle(s[:firstN])
	}

	// Phase 2: compute partial results for second half, combine
	needSecond := false
	for _, s := range scalars {
		if len(s) > m.half {
			needSecond = true
			break
		}
	}

	if !needSecond {
		return partials
	}

	m.loadPart(2)
	results := make([]bls12377.G1Jac, k)
	for i, s := range scalars {
		if len(s) <= m.half {
			results[i] = partials[i]
			continue
		}
		partial2 := m.runSingle(s[m.half:])
		results[i] = partials[i]
		results[i].AddAssign(&partial2)
	}
	return results
}

// runSingle runs a single MSM on the currently loaded points.
// Handles GPU execution, Horner combination, and Montgomery correction.
func (m *G1MSM) runSingle(scalars []fr.Element) bls12377.G1Jac {
	if len(scalars) == 0 {
		var zero bls12377.G1Jac
		return zero
	}

	if err := toError(C.gnark_gpu_msm_run(
		m.handle,
		(*C.uint64_t)(unsafe.Pointer(&m.windowResults[0])),
		(*C.uint64_t)(unsafe.Pointer(&scalars[0])),
		C.size_t(len(scalars)),
	)); err != nil {
		panic(fmt.Sprintf("gpu: MSM run failed: %v", err))
	}

	// Horner combination of per-window results in TE extended coordinates.
	acc := unpackTEExtended(m.windowResults[m.numWindows-1])
	for j := m.numWindows - 2; j >= 0; j-- {
		for range m.c {
			teDouble(&acc)
		}
		wj := unpackTEExtended(m.windowResults[j])
		teAdd(&acc, &wj)
	}
	result := teExtended2jac(acc)

	// Montgomery correction: GPU decomposes Montgomery-form scalars directly
	// without fr_from_mont, so result = R * correct_result.
	result.ScalarMultiplication(&result, &frRInv)
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal: point loading state machine (chunked mode)
// ─────────────────────────────────────────────────────────────────────────────

// loadPart swaps which half of points is resident on GPU.
// part=1: first half [0, half), part=2: second half [half, n).
func (m *G1MSM) loadPart(part int) {
	if m.loaded == part {
		return
	}

	// Offload current points
	if m.loaded != 0 {
		if err := toError(C.gnark_gpu_msm_offload_points(m.handle)); err != nil {
			panic("gpu: MSM offload failed: " + err.Error())
		}
	}

	switch part {
	case 1:
		count := m.half
		if count > m.n {
			count = m.n
		}
		if err := toError(C.gnark_gpu_msm_reload_points(
			m.handle,
			(*C.uint64_t)(m.pts.pinned),
			C.size_t(count),
		)); err != nil {
			panic("gpu: MSM reload first half failed: " + err.Error())
		}
	case 2:
		offset := uintptr(m.half) * uintptr(g1TEPrecompPointSize)
		ptr := unsafe.Add(m.pts.pinned, offset)
		count := m.n - m.half
		if err := toError(C.gnark_gpu_msm_reload_points(
			m.handle,
			(*C.uint64_t)(ptr),
			C.size_t(count),
		)); err != nil {
			panic("gpu: MSM reload second half failed: " + err.Error())
		}
	default:
		panic("gpu: invalid loadPart")
	}
	m.loaded = part
}

// ─────────────────────────────────────────────────────────────────────────────
// Point offloading for memory management
//
// During non-MSM phases (e.g., quotient computation), MSM points consume GPU
// memory unnecessarily. OffloadPoints frees the GPU-side point storage;
// ReloadPoints restores it from pinned host memory before the next MSM call.
//
// At n=2^27 this saves ~6 GiB of VRAM during the quotient phase.
// ─────────────────────────────────────────────────────────────────────────────

// OffloadPoints frees GPU-resident base points to reclaim VRAM.
// Call ReloadPoints before the next MultiExp.
func (m *G1MSM) OffloadPoints() {
	if m.loaded == 0 && !m.chunked {
		// Non-chunked: points are always loaded. Offload them.
		if err := toError(C.gnark_gpu_msm_offload_points(m.handle)); err != nil {
			panic("gpu: MSM offload failed: " + err.Error())
		}
	} else if m.loaded != 0 {
		if err := toError(C.gnark_gpu_msm_offload_points(m.handle)); err != nil {
			panic("gpu: MSM offload failed: " + err.Error())
		}
	}
	m.loaded = -1 // sentinel: points offloaded
}

// ReloadPoints restores GPU-resident base points after OffloadPoints.
func (m *G1MSM) ReloadPoints() {
	if m.loaded != -1 {
		return // not offloaded
	}
	if m.chunked {
		m.loaded = 0 // reset; loadPart(1) will reload on next MultiExp
	} else {
		// Non-chunked: reload all points
		if err := toError(C.gnark_gpu_msm_reload_points(
			m.handle,
			(*C.uint64_t)(m.pts.pinned),
			C.size_t(m.n),
		)); err != nil {
			panic("gpu: MSM reload failed: " + err.Error())
		}
		m.loaded = 0
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Lifecycle
// ─────────────────────────────────────────────────────────────────────────────

// Close releases GPU resources and pinned host memory.
func (m *G1MSM) Close() {
	if m.handle != nil {
		C.gnark_gpu_msm_destroy(m.handle)
		m.handle = nil
	}
	if m.pts != nil {
		m.pts.Free()
		m.pts = nil
	}
	runtime.SetFinalizer(m, nil)
}
