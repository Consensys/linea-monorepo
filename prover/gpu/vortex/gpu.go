// GPU wrappers for KoalaBear vector operations, NTT, Poseidon2, SIS, and Vortex commit.
//
// Build constraint: requires CGO + CUDA library.
// Without CGO, the package falls back to pure Go (gnark-crypto CPU) via vortex.go.

//go:build cuda

package vortex

/*
#cgo LDFLAGS: -L${SRCDIR}/../cuda/build -lgnark_gpu -L/usr/local/cuda/lib64 -lcudart -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/../cuda/include

#include "gnark_gpu.h"
#include "gnark_gpu_kb.h"
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>

// Allocate pre-faulted memory with huge page hint. Avoids 262K page faults
// (~2s) for GB-scale allocations by pre-populating pages in-kernel.
static void *alloc_prefaulted(size_t nbytes) {
    void *p = mmap(NULL, nbytes, PROT_READ|PROT_WRITE,
                   MAP_PRIVATE|MAP_ANONYMOUS|MAP_POPULATE, -1, 0);
    if (p == MAP_FAILED) return NULL;
    madvise(p, nbytes, MADV_HUGEPAGE);
    return p;
}
static void free_prefaulted(void *p, size_t nbytes) {
    munmap(p, nbytes);
}
*/
import "C"
import (
	"fmt"
	"math/bits"
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	refvortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// devCtx casts the common gpu.Device handle back to the C type for CGO calls.
func devCtx(d *gpu.Device) C.gnark_gpu_context_t {
	return C.gnark_gpu_context_t(d.Handle())
}

// ─────────────────────────────────────────────────────────────────────────────
// KBVector — KoalaBear vector on GPU
// ─────────────────────────────────────────────────────────────────────────────

type KBVector struct {
	dev    *gpu.Device
	handle C.kb_vec_t
	n      int
}

func NewKBVector(d *gpu.Device, n int) (*KBVector, error) {
	var h C.kb_vec_t
	if err := kbError(C.kb_vec_alloc(devCtx(d), C.size_t(n), &h)); err != nil {
		return nil, err
	}
	v := &KBVector{dev: d, handle: h, n: n}
	runtime.SetFinalizer(v, (*KBVector).Free)
	return v, nil
}

func (v *KBVector) Free() {
	if v.handle != nil {
		C.kb_vec_free(v.handle)
		v.handle = nil
	}
}

func (v *KBVector) Len() int { return v.n }

func (v *KBVector) CopyFromHost(src []koalabear.Element) {
	if len(src) != v.n {
		panic(fmt.Sprintf("vortex: CopyFromHost size mismatch: got %d, want %d", len(src), v.n))
	}
	ptr := (*C.uint32_t)(unsafe.Pointer(&src[0]))
	if err := kbError(C.kb_vec_h2d(devCtx(v.dev), v.handle, ptr, C.size_t(v.n))); err != nil {
		panic("vortex: CopyFromHost: " + err.Error())
	}
}

// CopyFromHostPinned copies from a pre-pinned host buffer (allocated by AllocPinned).
// Much faster than CopyFromHost for large buffers (DMA without staging).
func (v *KBVector) CopyFromHostPinned(src []koalabear.Element) {
	if len(src) != v.n {
		panic(fmt.Sprintf("vortex: CopyFromHostPinned size mismatch: got %d, want %d", len(src), v.n))
	}
	ptr := (*C.uint32_t)(unsafe.Pointer(&src[0]))
	if err := kbError(C.kb_vec_h2d_pinned(devCtx(v.dev), v.handle, ptr, C.size_t(v.n))); err != nil {
		panic("vortex: CopyFromHostPinned: " + err.Error())
	}
}

// AllocPinned allocates page-locked host memory for fast H2D.
// Returns a Go slice backed by CUDA pinned memory. Free with FreePinned.
func AllocPinned(n int) []koalabear.Element {
	var ptr *C.uint32_t
	if err := kbError(C.kb_pinned_alloc(C.size_t(n*4), &ptr)); err != nil { // 4 bytes per element
		panic("vortex: AllocPinned: " + err.Error())
	}
	return unsafe.Slice((*koalabear.Element)(unsafe.Pointer(ptr)), n)
}

// FreePinned frees memory allocated by AllocPinned.
func FreePinned(buf []koalabear.Element) {
	if len(buf) > 0 {
		C.kb_pinned_free((*C.uint32_t)(unsafe.Pointer(&buf[0])))
	}
}

func (v *KBVector) CopyToHost(dst []koalabear.Element) {
	if len(dst) != v.n {
		panic(fmt.Sprintf("vortex: CopyToHost size mismatch: got %d, want %d", len(dst), v.n))
	}
	ptr := (*C.uint32_t)(unsafe.Pointer(&dst[0]))
	if err := kbError(C.kb_vec_d2h(devCtx(v.dev), ptr, v.handle, C.size_t(v.n))); err != nil {
		panic("vortex: CopyToHost: " + err.Error())
	}
}

func (v *KBVector) Add(a, b *KBVector) {
	must(C.kb_vec_add(devCtx(v.dev), v.handle, a.handle, b.handle))
}
func (v *KBVector) Sub(a, b *KBVector) {
	must(C.kb_vec_sub(devCtx(v.dev), v.handle, a.handle, b.handle))
}
func (v *KBVector) Mul(a, b *KBVector) {
	must(C.kb_vec_mul(devCtx(v.dev), v.handle, a.handle, b.handle))
}

func (v *KBVector) Scale(scalar koalabear.Element) {
	must(C.kb_vec_scale(devCtx(v.dev), v.handle, C.uint32_t(scalar[0])))
}

func (v *KBVector) ScaleByPowers(g koalabear.Element) {
	must(C.kb_vec_scale_by_powers(devCtx(v.dev), v.handle, C.uint32_t(g[0])))
}

// BitReverse applies the bit-reversal permutation in-place.
// Required because GPU NTT uses DIF/DIT without internal bit-reversal:
//
//	IFFT: bitrev(input) → kb_ntt_inv → natural-order coefficients
//	FFT:  kb_ntt_fwd → bitrev(output) → natural-order evaluations
func (v *KBVector) BitReverse() {
	must(C.kb_vec_bitrev(devCtx(v.dev), v.handle))
}

// Sync waits for all queued GPU operations on the default stream to complete.
func Sync(d *gpu.Device) {
	must(C.kb_sync(devCtx(d)))
}

// D2DRaw copies n uint32 from src to dst device pointers (async, no sync).
func D2DRaw(d *gpu.Device, dst, src unsafe.Pointer, n int) {
	must(C.kb_vec_d2d_offset(devCtx(d), (*C.uint32_t)(dst), (*C.uint32_t)(src), C.size_t(n)))
}

// CosetFFTRaw applies coset forward NTT on raw device pointer (async, no sync).
func (f *GPUFFTDomain) CosetFFTRaw(data unsafe.Pointer, g koalabear.Element) {
	must(C.kb_ntt_coset_fwd_raw(devCtx(f.dev), f.handle, (*C.uint32_t)(data), C.uint32_t(g[0])))
}

// BitRevRaw applies bit-reversal on raw device pointer of n elements (async, no sync).
func BitRevRaw(d *gpu.Device, data unsafe.Pointer, n int) {
	must(C.kb_vec_bitrev_raw(devCtx(d), (*C.uint32_t)(data), C.size_t(n)))
}

// D2HRaw copies n uint32 from device src to host dst (synchronous).
func D2HRaw(d *gpu.Device, dst []koalabear.Element, src unsafe.Pointer, n int) {
	must(C.kb_vec_d2h_raw(devCtx(d), (*C.uint32_t)(unsafe.Pointer(&dst[0])), (*C.uint32_t)(src), C.size_t(n)))
}

// CopyFromDevice2 copies n elements starting at offset srcOff from src to this vector.
func (v *KBVector) CopyFromDevice2(src *KBVector, srcOff int) {
	if srcOff+v.n > src.n {
		panic(fmt.Sprintf("vortex: CopyFromDevice2 bounds: srcOff=%d n=%d src.n=%d", srcOff, v.n, src.n))
	}
	// Use raw device pointer arithmetic
	srcDevPtr := unsafe.Add(src.DevicePtr(), srcOff*4) // 4 bytes per uint32
	dstDevPtr := v.DevicePtr()
	must(C.kb_vec_d2d_offset(devCtx(v.dev),
		(*C.uint32_t)(dstDevPtr),
		(*C.uint32_t)(srcDevPtr),
		C.size_t(v.n)))
}

// CopyFromDevice copies data from another device-resident KBVector (GPU→GPU, same size).
func (v *KBVector) CopyFromDevice(src *KBVector) {
	if v.n != src.n {
		panic(fmt.Sprintf("vortex: CopyFromDevice size mismatch: got %d, want %d", src.n, v.n))
	}
	if err := kbError(C.kb_vec_d2d(devCtx(v.dev), v.handle, src.handle)); err != nil {
		panic("vortex: CopyFromDevice: " + err.Error())
	}
}

// DevicePtr returns the raw device pointer for cross-package access (e.g. symbolic eval).
func (v *KBVector) DevicePtr() unsafe.Pointer {
	return unsafe.Pointer(C.kb_vec_device_ptr(v.handle))
}

// ─────────────────────────────────────────────────────────────────────────────
// FFTDomain — NTT twiddles on GPU
// ─────────────────────────────────────────────────────────────────────────────

type GPUFFTDomain struct {
	dev    *gpu.Device
	handle C.kb_ntt_t
	n      int
}

func NewGPUFFTDomain(d *gpu.Device, size int) (*GPUFFTDomain, error) {
	domain := fft.NewDomain(uint64(size))
	halfN := size / 2

	fwdTw := make([]koalabear.Element, halfN)
	invTw := make([]koalabear.Element, halfN)
	fwdTw[0].SetOne()
	invTw[0].SetOne()
	gen := domain.Generator
	genInv := domain.GeneratorInv
	for i := 1; i < halfN; i++ {
		fwdTw[i].Mul(&fwdTw[i-1], &gen)
		invTw[i].Mul(&invTw[i-1], &genInv)
	}

	var h C.kb_ntt_t
	fptr := (*C.uint32_t)(unsafe.Pointer(&fwdTw[0]))
	iptr := (*C.uint32_t)(unsafe.Pointer(&invTw[0]))
	if err := kbError(C.kb_ntt_init(devCtx(d), C.size_t(size), fptr, iptr, &h)); err != nil {
		return nil, err
	}
	return &GPUFFTDomain{dev: d, handle: h, n: size}, nil
}

func (f *GPUFFTDomain) Free() {
	if f.handle != nil {
		C.kb_ntt_free(f.handle)
		f.handle = nil
	}
}

func (f *GPUFFTDomain) FFT(v *KBVector) { must(C.kb_ntt_fwd(devCtx(f.dev), f.handle, v.handle)) }

// BatchCosetFFTBitRev applies coset forward NTT + bit-reversal to `batch` packed vectors.
// `data` must contain batch*n elements packed contiguously. Single CGO call.
func (f *GPUFFTDomain) BatchCosetFFTBitRev(data *KBVector, batch int, g koalabear.Element) {
	must(C.kb_ntt_batch_coset_fwd_bitrev(devCtx(f.dev), f.handle,
		(*C.uint32_t)(data.DevicePtr()), C.size_t(f.n), C.size_t(batch), C.uint32_t(g[0])))
}

// BatchIFFTScale applies bit-reversal + inverse NTT + scale(nInv) to `batch` packed vectors.
func (f *GPUFFTDomain) BatchIFFTScale(data *KBVector, batch int, nInv koalabear.Element) {
	must(C.kb_ntt_batch_ifft_scale(devCtx(f.dev), f.handle,
		(*C.uint32_t)(data.DevicePtr()), C.size_t(f.n), C.size_t(batch), C.uint32_t(nInv[0])))
}

func (f *GPUFFTDomain) FFTInverse(v *KBVector) { must(C.kb_ntt_inv(devCtx(f.dev), f.handle, v.handle)) }
func (f *GPUFFTDomain) CosetFFT(v *KBVector, g koalabear.Element) {
	must(C.kb_ntt_coset_fwd(devCtx(f.dev), f.handle, v.handle, C.uint32_t(g[0])))
}

// ─────────────────────────────────────────────────────────────────────────────
// E4 NTT — FFT on KoalaBear extension field (degree-4)
//
// E4 elements are (B0.A0, B0.A1, B1.A0, B1.A1) — 4 base-field components.
// An E4 NTT decomposes into 4 independent base-field NTTs (one per component).
//
// Data layout on GPU (SoA): for n E4 elements, stored as 4*n base-field
// elements in 4 contiguous blocks: [A0(0..n), A1(0..n), A2(0..n), A3(0..n)].
//
// The GPU pipeline: AoS→SoA transpose → 4× batch NTT → SoA→AoS transpose.
// ─────────────────────────────────────────────────────────────────────────────

// FFTE4 performs forward NTT on n E4 elements.
// Input: a in natural order (AoS). Output: a in bit-reversed order (AoS).
// Decomposes into 4 independent base-field NTTs (one per E4 component).
func (f *GPUFFTDomain) FFTE4(a []fext.E4) {
	n := f.n
	if len(a) != n {
		panic(fmt.Sprintf("vortex: FFTE4 size mismatch: got %d, want %d", len(a), n))
	}
	soa := e4AoSToSoA(a)
	vecs := allocE4Components(f.dev, n)
	defer freeE4Components(vecs)
	copyE4SoAToGPU(vecs, soa, n)
	for c := 0; c < 4; c++ {
		f.FFT(vecs[c])
	}
	Sync(f.dev)
	copyE4GPUToSoA(vecs, soa, n)
	e4SoAToAoS(soa, a)
}

// FFTInverseE4 performs inverse NTT on n E4 elements.
// Input: a in bit-reversed order (AoS). Output: a in natural order (AoS).
// Note: like the base-field GPU IFFT, this does NOT include 1/n scaling.
func (f *GPUFFTDomain) FFTInverseE4(a []fext.E4) {
	n := f.n
	if len(a) != n {
		panic(fmt.Sprintf("vortex: FFTInverseE4 size mismatch: got %d, want %d", len(a), n))
	}
	soa := e4AoSToSoA(a)
	vecs := allocE4Components(f.dev, n)
	defer freeE4Components(vecs)
	copyE4SoAToGPU(vecs, soa, n)
	for c := 0; c < 4; c++ {
		f.FFTInverse(vecs[c])
	}
	Sync(f.dev)
	copyE4GPUToSoA(vecs, soa, n)
	e4SoAToAoS(soa, a)
}

// CosetFFTE4 performs coset forward NTT on n E4 elements.
// Input: a holds coefficients in natural order (AoS).
// Output: a holds evaluations on coset g·H in natural order (AoS).
// Internally: ScaleByPowers(g) + forward NTT + bit-reversal, per component.
func (f *GPUFFTDomain) CosetFFTE4(a []fext.E4, g koalabear.Element) {
	n := f.n
	if len(a) != n {
		panic(fmt.Sprintf("vortex: CosetFFTE4 size mismatch: got %d, want %d", len(a), n))
	}
	soa := e4AoSToSoA(a)
	vecs := allocE4Components(f.dev, n)
	defer freeE4Components(vecs)
	copyE4SoAToGPU(vecs, soa, n)
	for c := 0; c < 4; c++ {
		f.CosetFFT(vecs[c], g)
	}
	Sync(f.dev)
	copyE4GPUToSoA(vecs, soa, n)
	e4SoAToAoS(soa, a)
}

// BatchCosetFFTE4BitRev performs coset FFT + bit-reversal on E4 data in SoA layout.
// data must contain nE4*4 base-field elements (4 component vectors of length nE4).
// This is the zero-copy variant for use in pipelines that manage their own GPU buffers.
func (f *GPUFFTDomain) BatchCosetFFTE4BitRev(data *KBVector, nE4 int, g koalabear.Element) {
	if data.Len() != nE4*4 {
		panic(fmt.Sprintf("vortex: BatchCosetFFTE4BitRev size mismatch: got %d, want %d", data.Len(), nE4*4))
	}
	must(C.kb_ntt_batch_coset_fwd_bitrev(devCtx(f.dev), f.handle,
		(*C.uint32_t)(data.DevicePtr()), C.size_t(nE4), C.size_t(4), C.uint32_t(g[0])))
}

// BatchIFFTScaleE4 performs batch IFFT + scale on E4 data in SoA layout.
// data must contain nE4*4 base-field elements.
func (f *GPUFFTDomain) BatchIFFTScaleE4(data *KBVector, nE4 int, nInv koalabear.Element) {
	if data.Len() != nE4*4 {
		panic(fmt.Sprintf("vortex: BatchIFFTScaleE4 size mismatch: got %d, want %d", data.Len(), nE4*4))
	}
	must(C.kb_ntt_batch_ifft_scale(devCtx(f.dev), f.handle,
		(*C.uint32_t)(data.DevicePtr()), C.size_t(nE4), C.size_t(4), C.uint32_t(nInv[0])))
}

// ── E4 NTT helpers ──────────────────────────────────────────────────────────

// e4AoSToSoA transposes n E4 elements from AoS to SoA layout.
// AoS: [e0.B0.A0, e0.B0.A1, e0.B1.A0, e0.B1.A1, e1.B0.A0, ...]
// SoA: [all B0.A0s | all B0.A1s | all B1.A0s | all B1.A1s]
func e4AoSToSoA(a []fext.E4) []koalabear.Element {
	n := len(a)
	soa := make([]koalabear.Element, n*4)
	d0, d1, d2, d3 := soa[:n], soa[n:2*n], soa[2*n:3*n], soa[3*n:]
	for i := range a {
		d0[i] = a[i].B0.A0
		d1[i] = a[i].B0.A1
		d2[i] = a[i].B1.A0
		d3[i] = a[i].B1.A1
	}
	return soa
}

// e4SoAToAoS transposes SoA back to AoS, writing into the provided slice.
func e4SoAToAoS(soa []koalabear.Element, a []fext.E4) {
	n := len(a)
	d0, d1, d2, d3 := soa[:n], soa[n:2*n], soa[2*n:3*n], soa[3*n:]
	for i := range a {
		a[i].B0.A0 = d0[i]
		a[i].B0.A1 = d1[i]
		a[i].B1.A0 = d2[i]
		a[i].B1.A1 = d3[i]
	}
}

// allocE4Components allocates 4 KBVectors (one per E4 component).
func allocE4Components(dev *gpu.Device, n int) [4]*KBVector {
	var vecs [4]*KBVector
	for c := 0; c < 4; c++ {
		v, err := NewKBVector(dev, n)
		if err != nil {
			for j := 0; j < c; j++ {
				vecs[j].Free()
			}
			panic("vortex: allocE4Components: " + err.Error())
		}
		vecs[c] = v
	}
	return vecs
}

func freeE4Components(vecs [4]*KBVector) {
	for _, v := range vecs {
		v.Free()
	}
}

func copyE4SoAToGPU(vecs [4]*KBVector, soa []koalabear.Element, n int) {
	for c := 0; c < 4; c++ {
		vecs[c].CopyFromHost(soa[c*n : (c+1)*n])
	}
}

func copyE4GPUToSoA(vecs [4]*KBVector, soa []koalabear.Element, n int) {
	for c := 0; c < 4; c++ {
		vecs[c].CopyToHost(soa[c*n : (c+1)*n])
	}
	Sync(vecs[0].dev)
}

// ─────────────────────────────────────────────────────────────────────────────
// Poseidon2 — GPU batch hashing
// ─────────────────────────────────────────────────────────────────────────────

type GPUPoseidon2 struct {
	dev    *gpu.Device
	handle C.kb_p2_t
	width  int
}

// NewGPUPoseidon2 creates a Poseidon2 instance with standard parameters (rf=6, rp=21).
func NewGPUPoseidon2(d *gpu.Device, width int) (*GPUPoseidon2, error) {
	const (
		rf = 6
		rp = 21
	)
	params := poseidon2.NewParameters(width, rf, rp)

	// Flatten round keys
	var flat []koalabear.Element
	for _, rk := range params.RoundKeys {
		flat = append(flat, rk...)
	}

	// Get internal MDS diagonal
	diag := poseidon2Diag(width)

	var h C.kb_p2_t
	rkPtr := (*C.uint32_t)(unsafe.Pointer(&flat[0]))
	dPtr := (*C.uint32_t)(unsafe.Pointer(&diag[0]))
	if err := kbError(C.kb_p2_init(devCtx(d), C.int(width), C.int(rf), C.int(rp), rkPtr, dPtr, &h)); err != nil {
		return nil, err
	}
	return &GPUPoseidon2{dev: d, handle: h, width: width}, nil
}

func (p *GPUPoseidon2) Free() {
	if p.handle != nil {
		C.kb_p2_free(p.handle)
		p.handle = nil
	}
}

func (p *GPUPoseidon2) CompressBatch(input []koalabear.Element, count int) []Hash {
	if p.width != 16 {
		panic("vortex: CompressBatch requires width=16 Poseidon2")
	}
	output := make([]Hash, count)
	iptr := (*C.uint32_t)(unsafe.Pointer(&input[0]))
	optr := (*C.uint32_t)(unsafe.Pointer(&output[0]))
	must(C.kb_p2_compress_batch(devCtx(p.dev), p.handle, iptr, optr, C.size_t(count)))
	return output
}

// ─────────────────────────────────────────────────────────────────────────────
// GPU linear combination
// ─────────────────────────────────────────────────────────────────────────────

func GPULinCombE4(dev *gpu.Device, rows []*KBVector, alpha fext.E4, nCols int) []fext.E4 {
	nRows := len(rows)
	handles := make([]C.kb_vec_t, nRows)
	for i, r := range rows {
		handles[i] = r.handle
	}
	alphaRaw := [4]C.uint32_t{
		C.uint32_t(alpha.B0.A0[0]),
		C.uint32_t(alpha.B0.A1[0]),
		C.uint32_t(alpha.B1.A0[0]),
		C.uint32_t(alpha.B1.A1[0]),
	}
	result := make([]fext.E4, nCols)
	rptr := (*C.uint32_t)(unsafe.Pointer(&result[0]))
	must(C.kb_lincomb_e4(devCtx(dev), &handles[0], C.size_t(nRows), C.size_t(nCols), &alphaRaw[0], rptr))
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// GPUVortex — GPU-accelerated Vortex commit
// ─────────────────────────────────────────────────────────────────────────────
//
// Full GPU pipeline: RS encode (batch NTT) + SIS hash + Poseidon2 + Merkle.
// Raw rows are uploaded to GPU; RS encoding, hashing, and tree construction
// all run on device.  Encoded matrix stays on GPU for Prove (lincomb + column
// extraction), eliminating the 8 GiB D2H bottleneck.
//
//   gv, _ := NewGPUVortex(dev, params, nRows)
//   defer gv.Free()
//   cs, root, _ := gv.Commit(rows)
//   proof, _ := cs.Prove(alpha, selectedCols)

type GPUVortex struct {
	dev      *gpu.Device
	sis      C.kb_sis_t
	p2s      C.kb_p2_t // width=24 sponge (SIS→leaf)
	p2c      C.kb_p2_t // width=16 compress (Merkle tree)
	pipeline C.kb_vortex_pipeline_t
	params   *Params

	// Pinned host buffers (zero-copy Go slices over cudaMallocHost memory).
	inputBuf []koalabear.Element // [maxNRows × nCols], raw input rows
	treeBuf  []Hash              // [2·np − 1]
}

// NewGPUVortex initializes GPU resources for Vortex commit.
// maxNRows is the maximum number of rows that will be passed to Commit().
// Pre-allocates all device buffers + RS domain data for zero-allocation commits.
func NewGPUVortex(dev *gpu.Device, params *Params, maxNRows int) (*GPUVortex, error) {
	// Pin this thread to the chosen device before any allocation. CUDA's
	// "current device" is per-OS-thread state; without this the pipeline's
	// device buffers (multi-GB) silently land on device 0 even when
	// `dev.DeviceID() == 1`, defeating multi-GPU.
	if err := dev.Bind(); err != nil {
		return nil, fmt.Errorf("vortex: Bind device %d: %w", dev.DeviceID(), err)
	}
	inner := params.inner
	sisKey := inner.Key
	degree := sisKey.Degree
	halfDeg := degree / 2
	nCols := inner.NbColumns
	rate := inner.ReedSolomonInvRate

	// ── Build SIS domain twiddle arrays ──────────────────────────────────
	sisDom := sisKey.Domain
	sisFwd := make([]koalabear.Element, halfDeg)
	sisInv := make([]koalabear.Element, halfDeg)
	sisFwd[0].SetOne()
	sisInv[0].SetOne()
	gen := sisDom.Generator
	genInv := sisDom.GeneratorInv
	for i := 1; i < halfDeg; i++ {
		sisFwd[i].Mul(&sisFwd[i-1], &gen)
		sisInv[i].Mul(&sisInv[i-1], &genInv)
	}

	// ── Coset tables (SIS domain) ────────────────────────────────────────
	cosetTable, err := sisDom.CosetTable()
	if err != nil {
		return nil, fmt.Errorf("vortex: CosetTable: %w", err)
	}
	cosetTableInv, err := sisDom.CosetTableInv()
	if err != nil {
		return nil, fmt.Errorf("vortex: CosetTableInv: %w", err)
	}
	cardInvSIS := sisDom.CardinalityInv
	cosetInv := make([]koalabear.Element, degree)
	for j := 0; j < degree; j++ {
		cosetInv[j].Mul(&cosetTableInv[j], &cardInvSIS)
	}

	// ── Flatten SIS Ag keys ──────────────────────────────────────────────
	nPolys := len(sisKey.Ag)
	agFlat := make([]koalabear.Element, nPolys*degree)
	for i := 0; i < nPolys; i++ {
		copy(agFlat[i*degree:(i+1)*degree], sisKey.Ag[i])
	}

	// ── Init SIS on GPU ──────────────────────────────────────────────────
	var sisHandle C.kb_sis_t
	if err := kbError(C.kb_sis_init(devCtx(dev),
		C.int(degree), C.int(nPolys), C.int(sisKey.LogTwoBound),
		(*C.uint32_t)(unsafe.Pointer(&agFlat[0])),
		(*C.uint32_t)(unsafe.Pointer(&sisFwd[0])),
		(*C.uint32_t)(unsafe.Pointer(&sisInv[0])),
		(*C.uint32_t)(unsafe.Pointer(&cosetTable[0])),
		(*C.uint32_t)(unsafe.Pointer(&cosetInv[0])),
		&sisHandle)); err != nil {
		return nil, fmt.Errorf("vortex: kb_sis_init: %w", err)
	}

	// ── Init Poseidon2 (sponge width=24 + compress width=16) ─────────────
	const (
		rf = 6
		rp = 21
	)
	initP2 := func(width int) (C.kb_p2_t, error) {
		p := poseidon2.NewParameters(width, rf, rp)
		var flat []koalabear.Element
		for _, rk := range p.RoundKeys {
			flat = append(flat, rk...)
		}
		diag := poseidon2Diag(width)
		var h C.kb_p2_t
		if err := kbError(C.kb_p2_init(devCtx(dev), C.int(width), C.int(rf), C.int(rp),
			(*C.uint32_t)(unsafe.Pointer(&flat[0])),
			(*C.uint32_t)(unsafe.Pointer(&diag[0])),
			&h)); err != nil {
			return nil, err
		}
		return h, nil
	}

	p2s, err := initP2(24)
	if err != nil {
		C.kb_sis_free(sisHandle)
		return nil, fmt.Errorf("vortex: p2_sponge init: %w", err)
	}
	p2c, err := initP2(16)
	if err != nil {
		C.kb_sis_free(sisHandle)
		C.kb_p2_free(p2s)
		return nil, fmt.Errorf("vortex: p2_compress init: %w", err)
	}

	// ── Build RS domain twiddle arrays ───────────────────────────────────
	rsDom := inner.Domains[0]
	halfNC := nCols / 2
	rsFwd := make([]koalabear.Element, halfNC)
	rsInv := make([]koalabear.Element, halfNC)
	rsFwd[0].SetOne()
	rsInv[0].SetOne()
	rsGen := rsDom.Generator
	rsGenInv := rsDom.GeneratorInv
	for i := 1; i < halfNC; i++ {
		rsFwd[i].Mul(&rsFwd[i-1], &rsGen)
		rsInv[i].Mul(&rsInv[i-1], &rsGenInv)
	}

	// ── Build scaled coset table: CosetTableBitReverse × cardinalityInv ─
	cosetBR := inner.CosetTableBitReverse
	cardInvRS := rsDom.CardinalityInv
	scaledCoset := make([]koalabear.Element, nCols)
	for i := 0; i < nCols; i++ {
		scaledCoset[i].Mul(&cosetBR[i], &cardInvRS)
	}

	// ── Init pipeline ────────────────────────────────────────────────────
	sizeCodeWord := inner.SizeCodeWord()
	treeNP := nextPow2u(sizeCodeWord)
	treeNodes := 2*treeNP - 1

	var pipeHandle C.kb_vortex_pipeline_t
	if err := kbError(C.kb_vortex_pipeline_init(devCtx(dev),
		sisHandle, p2s, p2c,
		C.size_t(maxNRows), C.size_t(nCols), C.int(rate),
		(*C.uint32_t)(unsafe.Pointer(&rsFwd[0])),
		(*C.uint32_t)(unsafe.Pointer(&rsInv[0])),
		(*C.uint32_t)(unsafe.Pointer(&scaledCoset[0])),
		&pipeHandle)); err != nil {
		C.kb_sis_free(sisHandle)
		C.kb_p2_free(p2s)
		C.kb_p2_free(p2c)
		return nil, fmt.Errorf("vortex: pipeline init: %w", err)
	}

	// ── Upload multi-coset scaling tables for rate > 2 ──────────────
	// coset_k_br[j] = CosetTableBitReverse[j]^k × cardinalityInv
	// Derived iteratively: table_k = table_{k-1} ⊙ CosetTableBitReverse
	if rate > 2 {
		cosetBR := inner.CosetTableBitReverse
		nTables := rate - 1
		flat := make([]koalabear.Element, nTables*nCols)
		// Table 1: cosetBR[j]^1 × cardInv
		for j := 0; j < nCols; j++ {
			flat[j].Mul(&cosetBR[j], &cardInvRS)
		}
		// Tables 2..rate-1: table_k[j] = table_{k-1}[j] × cosetBR[j]
		for k := 2; k < rate; k++ {
			prev := (k - 2) * nCols
			cur := (k - 1) * nCols
			for j := 0; j < nCols; j++ {
				flat[cur+j].Mul(&flat[prev+j], &cosetBR[j])
			}
		}
		if err := kbError(C.kb_vortex_pipeline_set_coset_tables(
			pipeHandle,
			(*C.uint32_t)(unsafe.Pointer(&flat[0])),
			C.size_t(nTables))); err != nil {
			C.kb_vortex_pipeline_free(pipeHandle)
			C.kb_sis_free(sisHandle)
			C.kb_p2_free(p2s)
			C.kb_p2_free(p2c)
			return nil, fmt.Errorf("vortex: set_coset_tables: %w", err)
		}
	}

	// Wrap pinned host buffers as Go slices (zero-copy, page-locked DMA)
	inputBuf := C.kb_vortex_pipeline_input_buf(pipeHandle)
	treePtr := C.kb_vortex_pipeline_tree_buf(pipeHandle)

	gv := &GPUVortex{
		dev: dev, sis: sisHandle, p2s: p2s, p2c: p2c,
		pipeline: pipeHandle, params: params,
		inputBuf: unsafe.Slice((*koalabear.Element)(unsafe.Pointer(inputBuf)), maxNRows*nCols),
		treeBuf:  unsafe.Slice((*Hash)(unsafe.Pointer(treePtr)), treeNodes),
	}
	runtime.SetFinalizer(gv, (*GPUVortex).Free)
	return gv, nil
}

func (gv *GPUVortex) Free() {
	if gv.pipeline != nil {
		C.kb_vortex_pipeline_free(gv.pipeline)
		gv.pipeline = nil
	}
	if gv.sis != nil {
		C.kb_sis_free(gv.sis)
		gv.sis = nil
	}
	if gv.p2s != nil {
		C.kb_p2_free(gv.p2s)
		gv.p2s = nil
	}
	if gv.p2c != nil {
		C.kb_p2_free(gv.p2c)
		gv.p2c = nil
	}
}

// Commit performs GPU-accelerated Vortex commit.
//
// Raw rows are copied to pinned memory and uploaded to GPU. RS encoding
// (batch NTT), SIS hashing, Poseidon2, and Merkle tree all run on device.
// The encoded matrix stays on GPU for Prove().
//
// A new Commit() call invalidates any previously returned CommitState.
func (gv *GPUVortex) Commit(rows [][]koalabear.Element) (*CommitState, Hash, error) {
	inner := gv.params.inner
	nCols := inner.NbColumns
	sizeCodeWord := inner.SizeCodeWord()
	nRows := len(rows)
	treeNP := nextPow2u(sizeCodeWord)

	// ── 1. Copy raw rows to pinned host buffer (parallel) ───────────────
	inputBuf := gv.inputBuf[:nRows*nCols]
	parallel.Execute(nRows, func(start, stop int) {
		for i := start; i < stop; i++ {
			copy(inputBuf[i*nCols:(i+1)*nCols], rows[i])
		}
	})

	// ── 2. GPU pipeline: RS encode + SIS + sponge + Merkle ──────────────
	inPtr := (*C.uint32_t)(unsafe.Pointer(&inputBuf[0]))
	if err := kbError(C.kb_vortex_commit(gv.pipeline, inPtr, C.size_t(nRows))); err != nil {
		return nil, Hash{}, fmt.Errorf("vortex: GPU commit: %w", err)
	}

	// Tree is in pinned host buffer (SIS hashes stay on device only)
	treeBuf := gv.treeBuf

	// ── 3. Reconstruct MerkleTree from flat heap buffer ─────────────────
	depth := bits.Len(uint(treeNP)) - 1
	levels := make([][]Hash, depth+1)
	for d := 0; d <= depth; d++ {
		start := (1 << d) - 1
		end := (1 << (d + 1)) - 1
		levels[d] = treeBuf[start:end]
	}

	root := levels[0][0]

	cs := &CommitState{
		pipeline:     gv.pipeline,
		params:       inner,
		nRows:        nRows,
		merkle:       &refvortex.MerkleTree{Levels: levels},
		sizeCodeWord: sizeCodeWord,
	}

	return cs, root, nil
}

// CommitDirect writes rows directly to pinned memory via the loadRow callback,
// avoiding intermediate Go heap allocations. Each call to loadRow(i, dst)
// must fill dst[:nCols] with the i-th row's data.
func (gv *GPUVortex) CommitDirect(nRows int, loadRow func(i int, dst []koalabear.Element)) (*CommitState, Hash, error) {
	inner := gv.params.inner
	nCols := inner.NbColumns
	sizeCodeWord := inner.SizeCodeWord()
	treeNP := nextPow2u(sizeCodeWord)

	// ── 1. Write rows directly to pinned host buffer (parallel) ──────
	inputBuf := gv.inputBuf[:nRows*nCols]
	parallel.Execute(nRows, func(start, stop int) {
		for i := start; i < stop; i++ {
			loadRow(i, inputBuf[i*nCols:(i+1)*nCols])
		}
	})

	// ── 2. GPU pipeline: RS encode + SIS + sponge + Merkle ──────────────
	inPtr := (*C.uint32_t)(unsafe.Pointer(&inputBuf[0]))
	if err := kbError(C.kb_vortex_commit(gv.pipeline, inPtr, C.size_t(nRows))); err != nil {
		return nil, Hash{}, fmt.Errorf("vortex: GPU commit: %w", err)
	}

	// Tree is in pinned host buffer
	treeBuf := gv.treeBuf

	// ── 3. Reconstruct MerkleTree from flat heap buffer ─────────────────
	depth := bits.Len(uint(treeNP)) - 1
	levels := make([][]Hash, depth+1)
	for d := 0; d <= depth; d++ {
		start := (1 << d) - 1
		end := (1 << (d + 1)) - 1
		levels[d] = treeBuf[start:end]
	}

	root := levels[0][0]

	cs := &CommitState{
		pipeline:     gv.pipeline,
		params:       inner,
		nRows:        nRows,
		merkle:       &refvortex.MerkleTree{Levels: levels},
		sizeCodeWord: sizeCodeWord,
	}

	return cs, root, nil
}

// CommitAndExtract performs GPU-accelerated Vortex commit with overlapped D2H.
//
// Overlaps D2H transfer of the encoded matrix, SIS hashes, and leaf hashes
// with SIS/Poseidon2/Merkle computation on GPU (uses two CUDA streams).
// Returns all results needed for the drop-in replacement in a single call,
// avoiding sequential Extract calls after Commit.
//
// A new CommitAndExtract() call invalidates previously returned pinned buffers.
func (gv *GPUVortex) CommitAndExtract(rows [][]koalabear.Element) (
	encodedRows [][]koalabear.Element,
	sisHashes []koalabear.Element,
	leaves []Hash,
	root Hash,
	tree *refvortex.MerkleTree,
	err error,
) {
	inner := gv.params.inner
	nCols := inner.NbColumns
	sizeCodeWord := inner.SizeCodeWord()
	nRows := len(rows)
	treeNP := nextPow2u(sizeCodeWord)
	degree := inner.Key.Degree

	// ── 1. Copy raw rows to pinned host buffer (parallel) ───────────────
	inputBuf := gv.inputBuf[:nRows*nCols]
	parallel.Execute(nRows, func(start, stop int) {
		for i := start; i < stop; i++ {
			copy(inputBuf[i*nCols:(i+1)*nCols], rows[i])
		}
	})

	// ── 2. GPU pipeline: commit + overlapped D2H ─────────────────────────
	inPtr := (*C.uint32_t)(unsafe.Pointer(&inputBuf[0]))
	if err = kbError(C.kb_vortex_commit_and_extract(gv.pipeline, inPtr, C.size_t(nRows))); err != nil {
		err = fmt.Errorf("vortex: GPU commit_and_extract: %w", err)
		return
	}

	// ── 3. Copy from pinned host buffers to Go-managed memory ────────────
	// Pinned buffers are reused on next Commit, so we must copy out.
	// Encoded matrix uses parallel copy to saturate memory bandwidth.

	encPtr := C.kb_vortex_h_enc_pinned(gv.pipeline)
	pinnedEnc := unsafe.Slice((*koalabear.Element)(unsafe.Pointer(encPtr)), nRows*sizeCodeWord)
	encBacking := make([]koalabear.Element, nRows*sizeCodeWord)
	{
		total := nRows * sizeCodeWord
		const chunk = 256 * 1024 // 256K elements = 1 MB per goroutine
		numChunks := (total + chunk - 1) / chunk
		parallel.Execute(numChunks, func(start, stop int) {
			for c := start; c < stop; c++ {
				off := c * chunk
				end := off + chunk
				if end > total {
					end = total
				}
				copy(encBacking[off:end], pinnedEnc[off:end])
			}
		})
	}
	encodedRows = make([][]koalabear.Element, nRows)
	for r := range encodedRows {
		encodedRows[r] = encBacking[r*sizeCodeWord : (r+1)*sizeCodeWord]
	}

	// SIS hashes: [scw × degree] — small, single copy.
	sisPtr := C.kb_vortex_h_sis_pinned(gv.pipeline)
	pinnedSIS := unsafe.Slice((*koalabear.Element)(unsafe.Pointer(sisPtr)), sizeCodeWord*degree)
	sisHashes = make([]koalabear.Element, sizeCodeWord*degree)
	copy(sisHashes, pinnedSIS)

	// Leaves: [scw] Hash — tiny, single copy.
	leavesPtr := C.kb_vortex_h_leaves_pinned(gv.pipeline)
	pinnedLeaves := unsafe.Slice((*Hash)(unsafe.Pointer(leavesPtr)), sizeCodeWord)
	leaves = make([]Hash, sizeCodeWord)
	copy(leaves, pinnedLeaves)

	// ── 4. Reconstruct MerkleTree from flat heap buffer ──────────────────
	treeBuf := gv.treeBuf
	depth := bits.Len(uint(treeNP)) - 1
	levels := make([][]Hash, depth+1)
	for d := 0; d <= depth; d++ {
		start := (1 << d) - 1
		end := (1 << (d + 1)) - 1
		levels[d] = treeBuf[start:end]
	}
	root = levels[0][0]
	tree = &refvortex.MerkleTree{Levels: levels}

	return
}

// ─────────────────────────────────────────────────────────────────────────────
// CommitState + Prove — GPU lincomb + column extraction
// ─────────────────────────────────────────────────────────────────────────────

// CommitState holds prover state after commit.
// For GPU commits, the encoded matrix stays on device; Prove extracts via D2H.
// For CPU commits (benchmark baseline), delegates to gnark-crypto's ProverState.
//
// Three GPU storage modes:
//  1. pipeline set, encodedGPU nil  — device-resident in shared pipeline (single-use only)
//  2. pipeline nil, encodedGPU set  — device-resident in per-round snapshot (safe across rounds)
//  3. pipeline nil, encodedMatrix set — host-resident CPU fallback
type CommitState struct {
	pipeline      C.kb_vortex_pipeline_t // nil for CPU commits and snapshots
	params        *refvortex.Params
	nRows         int
	merkle        *refvortex.MerkleTree
	sizeCodeWord  int
	cpuState      *refvortex.ProverState     // non-nil for CPU Commit() baseline
	encodedMatrix []smartvectors.SmartVector // non-nil for CPU CommitSIS fallback
	encodedGPU    *KBVector                  // per-round snapshot: [scw × nRows] column-major
	dev           *gpu.Device                // device handle for snapshot operations
}

// NRows returns the number of rows in this commit.
func (cs *CommitState) NRows() int { return cs.nRows }

// SnapshotEncoded copies the pipeline's device-resident encoded matrix to a
// per-round GPU buffer (D2D copy). This decouples this round's data from the
// shared pipeline, which will be overwritten by subsequent CommitDirect calls.
//
// After snapshot, LinComb and ExtractColumns use the per-round buffer.
// The pipeline reference is cleared to prevent accidental stale access.
func (cs *CommitState) SnapshotEncoded(dev *gpu.Device) error {
	if cs.pipeline == nil {
		return fmt.Errorf("vortex: SnapshotEncoded: no pipeline")
	}
	scw := cs.sizeCodeWord
	nRows := cs.nRows
	total := scw * nRows

	// Get raw device pointer to pipeline's column-major encoded matrix
	srcPtr := C.kb_vortex_encoded_device_ptr(cs.pipeline)
	if srcPtr == nil {
		return fmt.Errorf("vortex: SnapshotEncoded: null device pointer")
	}

	// Allocate per-round device buffer and D2D copy
	buf, err := NewKBVector(dev, total)
	if err != nil {
		return fmt.Errorf("vortex: SnapshotEncoded: alloc: %w", err)
	}
	D2DRaw(dev, buf.DevicePtr(), unsafe.Pointer(srcPtr), total)
	Sync(dev)

	cs.encodedGPU = buf
	cs.dev = dev
	cs.pipeline = nil // detach from shared pipeline
	return nil
}

// FreeGPU releases GPU-resident memory immediately rather than waiting for GC.
// After this call, the CommitState falls back to CPU for any remaining operations.
func (cs *CommitState) FreeGPU() {
	if cs.encodedGPU != nil {
		cs.encodedGPU.Free()
		cs.encodedGPU = nil
	}
	cs.pipeline = nil
}

// IsDeviceResident reports whether this state is backed by GPU-resident data.
func (cs *CommitState) IsDeviceResident() bool {
	return cs.encodedGPU != nil || (cs.pipeline != nil && cs.encodedMatrix == nil)
}

// GetEncodedMatrix returns the host-side encoded matrix as SmartVectors.
// For GPU commits, extracts from device (full D2H). For CPU fallbacks,
// returns the stored host matrix directly.
func (cs *CommitState) GetEncodedMatrix() []smartvectors.SmartVector {
	if cs.encodedMatrix != nil {
		return cs.encodedMatrix
	}
	if cs.encodedGPU != nil {
		return cs.snapshotToEncodedMatrix()
	}
	if cs.pipeline == nil {
		return nil
	}
	rows, err := cs.ExtractAllRows()
	if err != nil {
		panic("vortex: GetEncodedMatrix: " + err.Error())
	}
	em := make([]smartvectors.SmartVector, len(rows))
	for i, row := range rows {
		em[i] = smartvectors.NewRegular(row)
	}
	return em
}

// snapshotToEncodedMatrix downloads a column-major GPU snapshot to host and
// converts to row-major SmartVectors. Used by GetEncodedMatrix for recursion.
func (cs *CommitState) snapshotToEncodedMatrix() []smartvectors.SmartVector {
	scw := cs.sizeCodeWord
	nRows := cs.nRows
	colMajor := make([]koalabear.Element, scw*nRows)
	cs.encodedGPU.CopyToHost(colMajor)
	Sync(cs.dev)
	em := make([]smartvectors.SmartVector, nRows)
	for i := range em {
		row := make([]koalabear.Element, scw)
		for j := 0; j < scw; j++ {
			row[j] = colMajor[j*nRows+i]
		}
		em[i] = smartvectors.NewRegular(row)
	}
	return em
}

// Prove generates a Vortex opening proof.
//
// Linear combination (UAlpha) is computed on GPU using the device-resident
// encoded matrix. Opened columns are extracted via small D2H transfers.
// Merkle proofs are computed on the host tree buffer.
func (cs *CommitState) Prove(alpha fext.E4, selectedCols []int) (*Proof, error) {
	// CPU fallback path (for Params.Commit baseline)
	if cs.cpuState != nil {
		cs.cpuState.OpenLinComb(alpha)
		vp, err := cs.cpuState.OpenColumns(selectedCols)
		if err != nil {
			return nil, err
		}
		return &Proof{
			UAlpha:       vp.UAlpha,
			Columns:      vp.OpenedColumns,
			MerkleProofs: vp.MerkleProofOpenedColumns,
		}, nil
	}

	scw := cs.sizeCodeWord
	nRows := cs.nRows

	// ── 1. GPU linear combination: UAlpha[j] = Σᵢ αⁱ · encoded[j][i] ──
	alphaRaw := [4]C.uint32_t{
		C.uint32_t(alpha.B0.A0[0]),
		C.uint32_t(alpha.B0.A1[0]),
		C.uint32_t(alpha.B1.A0[0]),
		C.uint32_t(alpha.B1.A1[0]),
	}
	uAlpha := make([]fext.E4, scw)
	if err := kbError(C.kb_vortex_lincomb(cs.pipeline,
		C.size_t(nRows), &alphaRaw[0],
		(*C.uint32_t)(unsafe.Pointer(&uAlpha[0])))); err != nil {
		return nil, fmt.Errorf("vortex: GPU lincomb: %w", err)
	}

	// ── 2. Extract opened columns from GPU ──────────────────────────────
	columns := make([][]koalabear.Element, len(selectedCols))
	for i, c := range selectedCols {
		col := make([]koalabear.Element, nRows)
		if err := kbError(C.kb_vortex_extract_col(cs.pipeline,
			C.size_t(nRows), C.int(c),
			(*C.uint32_t)(unsafe.Pointer(&col[0])))); err != nil {
			return nil, fmt.Errorf("vortex: extract col %d: %w", c, err)
		}
		columns[i] = col
	}

	// ── 3. Merkle proofs from host tree buffer ──────────────────────────
	merkleProofs := make([]MerkleProof, len(selectedCols))
	for i, c := range selectedCols {
		merkleProofs[i] = merkleProve(cs.merkle, c)
	}

	return &Proof{
		UAlpha:       uAlpha,
		Columns:      columns,
		MerkleProofs: merkleProofs,
	}, nil
}

// LinComb computes UAlpha[j] = Σᵢ αⁱ · encoded[i][j].
//
// GPU path: single kernel call on device-resident column-major matrix.
// CPU fallback: iterates encodedMatrix SmartVectors on host.
func (cs *CommitState) LinComb(alpha fext.E4) ([]fext.E4, error) {
	if cs.encodedGPU != nil {
		// Per-round GPU snapshot: use standalone lincomb kernel
		alphaRaw := [4]C.uint32_t{
			C.uint32_t(alpha.B0.A0[0]),
			C.uint32_t(alpha.B0.A1[0]),
			C.uint32_t(alpha.B1.A0[0]),
			C.uint32_t(alpha.B1.A1[0]),
		}
		uAlpha := make([]fext.E4, cs.sizeCodeWord)
		if err := kbError(C.kb_lincomb_e4_colmajor(devCtx(cs.dev),
			(*C.uint32_t)(cs.encodedGPU.DevicePtr()),
			C.size_t(cs.nRows), C.size_t(cs.sizeCodeWord),
			&alphaRaw[0],
			(*C.uint32_t)(unsafe.Pointer(&uAlpha[0])))); err != nil {
			return nil, fmt.Errorf("vortex: GPU snapshot lincomb: %w", err)
		}
		return uAlpha, nil
	}
	if cs.pipeline != nil {
		alphaRaw := [4]C.uint32_t{
			C.uint32_t(alpha.B0.A0[0]),
			C.uint32_t(alpha.B0.A1[0]),
			C.uint32_t(alpha.B1.A0[0]),
			C.uint32_t(alpha.B1.A1[0]),
		}
		uAlpha := make([]fext.E4, cs.sizeCodeWord)
		if err := kbError(C.kb_vortex_lincomb(cs.pipeline,
			C.size_t(cs.nRows), &alphaRaw[0],
			(*C.uint32_t)(unsafe.Pointer(&uAlpha[0])))); err != nil {
			return nil, fmt.Errorf("vortex: GPU lincomb: %w", err)
		}
		return uAlpha, nil
	}
	if cs.encodedMatrix != nil {
		return linCombCPU(cs.encodedMatrix, alpha), nil
	}
	return nil, fmt.Errorf("vortex: CommitState has neither GPU pipeline nor encodedMatrix")
}

// linCombCPU computes UAlpha[j] = Σᵢ αⁱ · rows[i].Get(j) on CPU.
func linCombCPU(rows []smartvectors.SmartVector, alpha fext.E4) []fext.E4 {
	n := rows[0].Len()
	result := make([]fext.E4, n)
	var pow fext.E4
	pow.SetOne()
	for _, row := range rows {
		for j := 0; j < n; j++ {
			v := row.Get(j)
			var term fext.E4
			term.B0.A0 = v
			term.Mul(&term, &pow)
			result[j].Add(&result[j], &term)
		}
		pow.Mul(&pow, &alpha)
	}
	return result
}

// ExtractColumns extracts selected columns from the encoded matrix.
//
// GPU path: small D2H per column from device-resident column-major matrix.
// CPU fallback: gathers from encodedMatrix SmartVectors on host.
// Returns columns[i][row] for each selectedCols[i], row 0..nRows-1.
func (cs *CommitState) ExtractColumns(selectedCols []int) ([][]koalabear.Element, error) {
	if cs.encodedGPU != nil {
		// Per-round snapshot: D2H from column-major buffer at offset col*nRows
		columns := make([][]koalabear.Element, len(selectedCols))
		for i, c := range selectedCols {
			col := make([]koalabear.Element, cs.nRows)
			srcOff := unsafe.Add(cs.encodedGPU.DevicePtr(), c*cs.nRows*4) // 4 bytes per uint32
			D2HRaw(cs.dev, col, srcOff, cs.nRows)
			columns[i] = col
		}
		return columns, nil
	}
	if cs.pipeline != nil {
		columns := make([][]koalabear.Element, len(selectedCols))
		for i, c := range selectedCols {
			col := make([]koalabear.Element, cs.nRows)
			if err := kbError(C.kb_vortex_extract_col(cs.pipeline,
				C.size_t(cs.nRows), C.int(c),
				(*C.uint32_t)(unsafe.Pointer(&col[0])))); err != nil {
				return nil, fmt.Errorf("vortex: extract col %d: %w", c, err)
			}
			columns[i] = col
		}
		return columns, nil
	}
	if cs.encodedMatrix != nil {
		return extractColumnsCPU(cs.encodedMatrix, selectedCols), nil
	}
	return nil, fmt.Errorf("vortex: CommitState has neither GPU pipeline nor encodedMatrix")
}

// extractColumnsCPU gathers selected columns from host-side SmartVectors.
func extractColumnsCPU(rows []smartvectors.SmartVector, selectedCols []int) [][]koalabear.Element {
	columns := make([][]koalabear.Element, len(selectedCols))
	for i, c := range selectedCols {
		col := make([]koalabear.Element, len(rows))
		for r, row := range rows {
			col[r] = row.Get(c)
		}
		columns[i] = col
	}
	return columns
}

// ExtractAllRows downloads the full GPU encoded matrix and returns it as
// row-major [][]koalabear.Element (one slice per row, length sizeCodeWord).
// The GPU stores column-major, so this transposes during extraction.
func (cs *CommitState) ExtractAllRows() ([][]koalabear.Element, error) {
	if cs.pipeline == nil {
		return nil, fmt.Errorf("vortex: no GPU pipeline in CommitState")
	}
	scw := cs.sizeCodeWord
	nRows := cs.nRows

	// D2H: row-major flat buffer [nRows × scw], transposed on GPU.
	// Single contiguous allocation reduces GC pressure.
	rowMajor := make([]koalabear.Element, nRows*scw)
	if err := kbError(C.kb_vortex_extract_all_rowmajor(cs.pipeline,
		C.size_t(nRows),
		(*C.uint32_t)(unsafe.Pointer(&rowMajor[0])))); err != nil {
		return nil, fmt.Errorf("vortex: extract all rowmajor: %w", err)
	}

	// Slice the contiguous buffer into per-row slices (no copy).
	rows := make([][]koalabear.Element, nRows)
	for r := range rows {
		rows[r] = rowMajor[r*scw : (r+1)*scw]
	}
	return rows, nil
}

// MerkleTree returns the reconstructed Merkle tree from the GPU commit.
func (cs *CommitState) MerkleTree() *refvortex.MerkleTree {
	return cs.merkle
}

// ExtractSISHashes returns the SIS column hashes already transferred to
// the pipeline's pinned host buffer during kb_vortex_commit (overlapped with
// Merkle tree construction).
func (cs *CommitState) ExtractSISHashes() ([]koalabear.Element, error) {
	if cs.pipeline == nil {
		return nil, fmt.Errorf("vortex: no GPU pipeline in CommitState")
	}
	scw := cs.sizeCodeWord
	degree := int(C.kb_vortex_degree(cs.pipeline))
	n := scw * degree

	// SIS hashes were already D2H'd to h_sis_pinned during commit.
	sisPtr := C.kb_vortex_h_sis_pinned(cs.pipeline)
	if sisPtr == nil {
		return nil, fmt.Errorf("vortex: h_sis_pinned is nil")
	}

	// Allocate pre-faulted memory with huge pages, then memcpy from pinned.
	// Go's make() for ~1 GB triggers 262K page faults (~2s). MAP_POPULATE
	// pre-faults in-kernel, and MADV_HUGEPAGE uses 2 MB pages (512 faults).
	nbytes := C.size_t(n) * 4
	cPtr := C.alloc_prefaulted(nbytes)
	if cPtr == nil {
		return nil, fmt.Errorf("vortex: alloc_prefaulted(%d) failed", nbytes)
	}
	C.memcpy(cPtr, unsafe.Pointer(sisPtr), nbytes)
	out := unsafe.Slice((*koalabear.Element)(cPtr), n)
	return out, nil
}

// ExtractLeaves extracts the Poseidon2 leaf hashes from GPU to host.
// Returns []Hash (field.Octuplet) of length sizeCodeWord.
func (cs *CommitState) ExtractLeaves() ([]Hash, error) {
	if cs.pipeline == nil {
		return nil, fmt.Errorf("vortex: no GPU pipeline in CommitState")
	}
	scw := cs.sizeCodeWord
	out := make([]Hash, scw)
	if err := kbError(C.kb_vortex_extract_leaves(cs.pipeline,
		(*C.uint32_t)(unsafe.Pointer(&out[0])))); err != nil {
		return nil, fmt.Errorf("vortex: extract leaves: %w", err)
	}
	return out, nil
}

// merkleProve computes a Merkle inclusion proof for column colIdx.
// The leaf is: Poseidon2_sponge(SIS_hash[colIdx]).
// Tree layout: heap array, root at level 0 (single entry).
func merkleProve(tree *refvortex.MerkleTree, colIdx int) MerkleProof {
	// The tree has depth levels: levels[0]=[root], levels[d]=[2^d hashes].
	// Leaf index = colIdx at the bottom level. Walk up collecting siblings.
	depth := len(tree.Levels) - 1
	proof := make(MerkleProof, depth)
	idx := colIdx
	for d := depth; d > 0; d-- {
		// Sibling index at this level
		sibling := idx ^ 1
		if sibling < len(tree.Levels[d]) {
			proof[depth-d] = tree.Levels[d][sibling]
		}
		idx >>= 1
	}
	return proof
}

// ─────────────────────────────────────────────────────────────────────────────
// CPU Commit (used as benchmark baseline and for tests without GPU)
// ─────────────────────────────────────────────────────────────────────────────

func (p *Params) Commit(rows [][]koalabear.Element) (*CommitState, Hash, error) {
	ps, err := refvortex.Commit(p.inner, rows)
	if err != nil {
		return nil, Hash{}, err
	}
	root := ps.GetCommitment()
	// Wrap in GPU CommitState — lincomb/column extraction will fall back to CPU
	return &CommitState{
		params:       p.inner,
		nRows:        len(rows),
		merkle:       ps.MerkleTree,
		sizeCodeWord: p.inner.SizeCodeWord(),
		cpuState:     ps,
	}, root, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// poseidon2Diag returns the internal MDS diagonal for Poseidon2 (width 16 or 24).
// Values match gnark-crypto's poseidon2/hash.go init() — SetUint64 converts to Montgomery.
func poseidon2Diag(width int) []koalabear.Element {
	var vals []uint64
	switch width {
	case 16:
		vals = []uint64{
			2130706431, 1, 2, 1065353217, 3, 4, 1065353216, 2130706430,
			2130706429, 2122383361, 1864368129, 2130706306,
			8323072, 266338304, 133169152, 127,
		}
	case 24:
		vals = []uint64{
			2130706431, 1, 2, 1065353217, 3, 4, 1065353216, 2130706430,
			2130706429, 2122383361, 1598029825, 1864368129,
			1997537281, 2064121857, 2097414145, 2130706306,
			8323072, 266338304, 133169152, 66584576,
			33292288, 16646144, 4161536, 127,
		}
	default:
		panic(fmt.Sprintf("vortex: unsupported Poseidon2 width %d", width))
	}
	diag := make([]koalabear.Element, len(vals))
	for i, v := range vals {
		diag[i].SetUint64(v)
	}
	return diag
}

func nextPow2u(n int) int {
	v := 1
	for v < n {
		v <<= 1
	}
	return v
}

func kbError(code C.kb_error_t) error {
	switch code {
	case C.KB_SUCCESS:
		return nil
	case C.KB_ERROR_CUDA:
		return fmt.Errorf("vortex: CUDA error")
	case C.KB_ERROR_INVALID:
		return fmt.Errorf("vortex: invalid argument")
	case C.KB_ERROR_OOM:
		return fmt.Errorf("vortex: out of GPU memory")
	case C.KB_ERROR_SIZE:
		return fmt.Errorf("vortex: size mismatch")
	default:
		return fmt.Errorf("vortex: unknown error %d", int(code))
	}
}

func must(code C.kb_error_t) {
	if err := kbError(code); err != nil {
		panic(err)
	}
}
