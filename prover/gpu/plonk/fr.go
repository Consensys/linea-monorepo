//go:build cuda

package plonk

/*
#include "gnark_gpu.h"
*/
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// FrVector represents a vector of BLS12-377 scalar field (Fr) elements stored
// on the GPU in Structure-of-Arrays (SoA) layout for coalesced memory access.
//
// Memory layout on GPU (SoA, 4 limbs × n elements):
//
//	limb0: [a₀[0], a₁[0], a₂[0], ..., aₙ₋₁[0]]   ← contiguous
//	limb1: [a₀[1], a₁[1], a₂[1], ..., aₙ₋₁[1]]   ← contiguous
//	limb2: [a₀[2], a₁[2], a₂[2], ..., aₙ₋₁[2]]   ← contiguous
//	limb3: [a₀[3], a₁[3], a₂[3], ..., aₙ₋₁[3]]   ← contiguous
//
// All elements are in Montgomery form. The SoA layout enables coalesced
// memory access when a GPU warp processes consecutive elements.
//
// All operations accept an optional gpu.StreamID parameter. When omitted,
// operations run on the default compute stream (stream 0). Pass a gpu.StreamID
// to run on a specific CUDA stream for pipeline parallelism.
type FrVector struct {
	handle C.gnark_gpu_fr_vector_t
	dev    *gpu.Device
	n      int
}

// NewFrVector allocates GPU memory for n Fr elements.
// Free is still recommended for deterministic VRAM release in hot paths.
// A finalizer is installed as a safety net.
func NewFrVector(dev *gpu.Device, n int) (*FrVector, error) {
	if dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if n <= 0 {
		return nil, &gpu.Error{Code: -1, Message: "count must be positive"}
	}

	var handle C.gnark_gpu_fr_vector_t
	if err := toError(C.gnark_gpu_fr_vector_alloc(devCtx(dev), C.size_t(n), &handle)); err != nil {
		return nil, err
	}

	v := &FrVector{handle: handle, dev: dev, n: n}
	runtime.SetFinalizer(v, (*FrVector).Free)
	return v, nil
}

// Free releases GPU memory associated with this vector.
// It is safe to call Free multiple times.
func (v *FrVector) Free() {
	if v.handle != nil {
		C.gnark_gpu_fr_vector_free(v.handle)
		v.handle = nil
		runtime.SetFinalizer(v, nil)
	}
}

// Len returns the number of elements in the vector.
func (v *FrVector) Len() int {
	return v.n
}

// ─────────────────────────────────────────────────────────────────────────────
// Host ↔ Device transfers
// ─────────────────────────────────────────────────────────────────────────────

// CopyFromHost copies data from a gnark-crypto fr.Vector (host) to GPU.
// The host data is transposed from AoS to SoA on the fly.
// Panics if len(src) != v.Len() (programmer error, matches gnark-crypto style).
func (v *FrVector) CopyFromHost(src fr.Vector, stream ...gpu.StreamID) {
	if len(src) != v.n {
		panic("gpu: CopyFrom size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_copy_to_device_stream(
			v.handle,
			(*C.uint64_t)(unsafe.Pointer(&src[0])),
			C.size_t(v.n),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_fr_vector_copy_to_device(
			v.handle,
			(*C.uint64_t)(unsafe.Pointer(&src[0])),
			C.size_t(v.n),
		))
	}
	if err != nil {
		panic("gpu: CopyFrom failed: " + err.Error())
	}
}

// CopyToHost copies data from GPU back to a gnark-crypto fr.Vector (host).
// The GPU data is transposed from SoA back to AoS on the fly.
// Panics if len(dst) != v.Len() (programmer error, matches gnark-crypto style).
func (v *FrVector) CopyToHost(dst fr.Vector, stream ...gpu.StreamID) {
	if len(dst) != v.n {
		panic("gpu: CopyTo size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_copy_to_host_stream(
			v.handle,
			(*C.uint64_t)(unsafe.Pointer(&dst[0])),
			C.size_t(v.n),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_fr_vector_copy_to_host(
			v.handle,
			(*C.uint64_t)(unsafe.Pointer(&dst[0])),
			C.size_t(v.n),
		))
	}
	if err != nil {
		panic("gpu: CopyTo failed: " + err.Error())
	}
}

// CopyFromDevice copies src to v (GPU-to-GPU, no host roundtrip).
// Panics on size mismatch.
func (v *FrVector) CopyFromDevice(src *FrVector, stream ...gpu.StreamID) {
	if v.n != src.n {
		panic("gpu: CopyFromDevice size mismatch")
	}
	if v.dev != src.dev {
		panic("gpu: CopyFromDevice device mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_copy_d2d_stream(
			devCtx(v.dev), v.handle, src.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_copy_d2d(devCtx(v.dev), v.handle, src.handle))
	}
	if err != nil {
		panic("gpu: CopyFromDevice failed: " + err.Error())
	}
}

// SetZero sets all elements to zero.
// The operation is asynchronous; call dev.Sync() to wait for completion.
func (v *FrVector) SetZero(stream ...gpu.StreamID) {
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_set_zero_stream(
			devCtx(v.dev), v.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_set_zero(devCtx(v.dev), v.handle))
	}
	if err != nil {
		panic("gpu: SetZero failed: " + err.Error())
	}
}

// mustSameDeviceAndSize panics if vectors don't share the same device and size.
func mustSameDeviceAndSize(v, a, b *FrVector) {
	if v.n != a.n || a.n != b.n {
		panic("gpu: vector size mismatch")
	}
	if v.dev != a.dev || a.dev != b.dev {
		panic("gpu: vectors from different devices")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Element-wise arithmetic
//
// All operations are asynchronous; call dev.Sync() to wait for completion.
// All operations panic on size or device mismatch (programmer error).
// All operations accept an optional gpu.StreamID for multi-stream pipelining.
// ─────────────────────────────────────────────────────────────────────────────

// Mul performs element-wise Montgomery multiplication: v[i] = a[i] · b[i] (mod r).
func (v *FrVector) Mul(a, b *FrVector, stream ...gpu.StreamID) {
	mustSameDeviceAndSize(v, a, b)
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_mul_stream(
			devCtx(v.dev), v.handle, a.handle, b.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_mul(devCtx(v.dev), v.handle, a.handle, b.handle))
	}
	if err != nil {
		panic("gpu: Mul failed: " + err.Error())
	}
}

// Add performs element-wise addition: v[i] = a[i] + b[i] (mod r).
func (v *FrVector) Add(a, b *FrVector, stream ...gpu.StreamID) {
	mustSameDeviceAndSize(v, a, b)
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_add_stream(
			devCtx(v.dev), v.handle, a.handle, b.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_add(devCtx(v.dev), v.handle, a.handle, b.handle))
	}
	if err != nil {
		panic("gpu: Add failed: " + err.Error())
	}
}

// Sub performs element-wise subtraction: v[i] = a[i] - b[i] (mod r).
func (v *FrVector) Sub(a, b *FrVector, stream ...gpu.StreamID) {
	mustSameDeviceAndSize(v, a, b)
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_sub_stream(
			devCtx(v.dev), v.handle, a.handle, b.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_sub(devCtx(v.dev), v.handle, a.handle, b.handle))
	}
	if err != nil {
		panic("gpu: Sub failed: " + err.Error())
	}
}

// AddMul performs fused multiply-add: v[i] += a[i] · b[i] (mod r).
func (v *FrVector) AddMul(a, b *FrVector, stream ...gpu.StreamID) {
	mustSameDeviceAndSize(v, a, b)
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_addmul_stream(
			devCtx(v.dev), v.handle, a.handle, b.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_addmul(devCtx(v.dev), v.handle, a.handle, b.handle))
	}
	if err != nil {
		panic("gpu: AddMul failed: " + err.Error())
	}
}

// AddScalarMul performs broadcast scalar multiply-add: v[i] += a[i] · c (mod r).
// The scalar c is broadcast to all elements.
func (v *FrVector) AddScalarMul(a *FrVector, scalar fr.Element, stream ...gpu.StreamID) {
	if v.n != a.n {
		panic("gpu: AddScalarMul size mismatch")
	}
	if v.dev != a.dev {
		panic("gpu: AddScalarMul device mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_add_scalar_mul_stream(
			devCtx(v.dev), v.handle, a.handle,
			(*C.uint64_t)(unsafe.Pointer(&scalar)),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_fr_vector_add_scalar_mul(
			devCtx(v.dev), v.handle, a.handle,
			(*C.uint64_t)(unsafe.Pointer(&scalar)),
		))
	}
	if err != nil {
		panic("gpu: AddScalarMul failed: " + err.Error())
	}
}

// ScalarMul multiplies every element by a constant: v[i] *= c (mod r).
func (v *FrVector) ScalarMul(c fr.Element, stream ...gpu.StreamID) {
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_scalar_mul_stream(
			devCtx(v.dev), v.handle,
			(*C.uint64_t)(unsafe.Pointer(&c)),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_fr_vector_scalar_mul(
			devCtx(v.dev), v.handle,
			(*C.uint64_t)(unsafe.Pointer(&c)),
		))
	}
	if err != nil {
		panic("gpu: ScalarMul failed: " + err.Error())
	}
}

// ScaleByPowers computes v[i] *= gⁱ for i ∈ [0, n).
//
// Used for coset FFT shifting: given polynomial p(X) with coefficients v[i],
// ScaleByPowers(g) computes p(gX), enabling evaluation on the coset g·H
// where H is the FFT domain.
func (v *FrVector) ScaleByPowers(g fr.Element, stream ...gpu.StreamID) {
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_scale_by_powers_stream(
			devCtx(v.dev), v.handle,
			(*C.uint64_t)(unsafe.Pointer(&g)),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_fr_vector_scale_by_powers(
			devCtx(v.dev), v.handle,
			(*C.uint64_t)(unsafe.Pointer(&g)),
		))
	}
	if err != nil {
		panic("gpu: ScaleByPowers failed: " + err.Error())
	}
}

// BatchInvert computes v[i] = 1/v[i] using Montgomery's batch inversion trick.
//
// Montgomery's batch inversion for n elements a₀, a₁, …, aₙ₋₁:
//
//	Step 1 — forward scan (prefix products):
//	  p₀ = a₀,  pᵢ = pᵢ₋₁ · aᵢ  for i = 1…n−1
//
//	Step 2 — single inversion:
//	  inv = pₙ₋₁⁻¹
//
//	Step 3 — backward scan (extract individual inverses):
//	  for i = n−1 down to 1:
//	    aᵢ⁻¹ = inv · pᵢ₋₁
//	    inv   = inv · aᵢ      (restore running product)
//	  a₀⁻¹ = inv
//
// Cost: 3(n−1) multiplications + 1 inversion, vs n inversions naively.
//
// temp must be a separate FrVector of the same size (used as scratch space).
func (v *FrVector) BatchInvert(temp *FrVector, stream ...gpu.StreamID) {
	if v.n != temp.n {
		panic("gpu: BatchInvert size mismatch")
	}
	if v.dev != temp.dev {
		panic("gpu: BatchInvert device mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_fr_vector_batch_invert_stream(
			devCtx(v.dev), v.handle, temp.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_fr_vector_batch_invert(devCtx(v.dev), v.handle, temp.handle))
	}
	if err != nil {
		panic("gpu: BatchInvert failed: " + err.Error())
	}
}
