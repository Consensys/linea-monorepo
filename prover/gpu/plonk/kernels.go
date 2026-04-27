//go:build cuda

// plonk_kernels.go contains small high-impact GPU kernels used by the prover.
//
// The design here is intentionally simple:
//   - Keep GPU work data-parallel.
//   - Keep tiny global reductions on CPU when cheaper than extra kernels.
//   - Expose plain Go wrappers with deterministic panic-on-misuse semantics.
//
// Implemented primitives:
//
//  1. ZPrefixProduct
//     Computes the permutation product polynomial values:
//     Z[0] = 1
//     Z[i] = Π_{k=0}^{i-1} ratio[k]
//
//     Hybrid scan:
//
//     ratio --> [GPU local chunk scan] --> chunk products
//     |                           |
//     +-----------CPU scan--------+
//     |
//     v
//     [GPU chunk fixup + shift]
//
//     Pseudocode:
//     local_prefix_per_chunk()
//     chunk_scan_on_cpu()
//     apply_chunk_prefixes_and_right_shift()
//
//  2. PolyEvalGPU
//     Evaluates p(z) with chunked Horner:
//
//     p(x) = Σ c_i x^i
//
//     Split coefficients in K=1024 sized chunks.
//     GPU computes each chunk Horner independently.
//     CPU combines chunk partials with z^K.
//
//     partial_j = Horner(c[jK : (j+1)K], z)
//     p(z)      = Σ partial_j * (z^K)^j
//
// This keeps kernels regular and avoids a long serial dependency chain on GPU.
package plonk

/*
#include "gnark_gpu.h"
#include <stdlib.h>
*/
import "C"
import (
	"math/big"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// ZPrefixProduct computes Z[i] = product(ratio[0..i-1]) entirely on GPU
// with a small CPU scan of chunk products.
//
// ratioVec: input vector of per-element ratios (n elements, on GPU)
// zVec: output vector for the Z polynomial (n elements, on GPU)
// tempVec: scratch vector (n elements, on GPU)
//
// After completion, zVec contains the Z polynomial in Lagrange form:
// Z[0] = 1, Z[i] = Z[i-1] * ratio[i-1].
func ZPrefixProduct(dev *gpu.Device, zVec, ratioVec, tempVec *FrVector) {
	if zVec.n != ratioVec.n || zVec.n != tempVec.n {
		panic("gpu: ZPrefixProduct size mismatch")
	}
	n := ratioVec.n

	// Phase 1: GPU local prefix products + extract chunk products
	maxChunks := (n + 1023) / 1024 // Z_CHUNK_SIZE = 1024
	cpHost := make([]uint64, maxChunks*4)
	var numChunks C.size_t

	if err := toError(C.gnark_gpu_z_prefix_phase1(
		devCtx(dev),
		zVec.handle,
		ratioVec.handle,
		(*C.uint64_t)(unsafe.Pointer(&cpHost[0])),
		&numChunks,
	)); err != nil {
		panic("gpu: ZPrefixProduct phase1 failed: " + err.Error())
	}

	nc := int(numChunks)
	if nc <= 1 {
		// Only one chunk — just need to shift. Phase3 handles it.
		spHost := make([]uint64, 4)
		// scanned prefix for chunk 0 is just the chunk product itself
		copy(spHost, cpHost[:4])
		if err := toError(C.gnark_gpu_z_prefix_phase3(
			devCtx(dev), zVec.handle, tempVec.handle,
			(*C.uint64_t)(unsafe.Pointer(&spHost[0])),
			C.size_t(nc),
		)); err != nil {
			panic("gpu: ZPrefixProduct phase3 failed: " + err.Error())
		}
		return
	}

	// Phase 2: CPU sequential inclusive prefix product of chunk products.
	// cpHost is [c0.l0, c0.l1, c0.l2, c0.l3, c1.l0, ...] in AoS layout.
	// scanned[i] = product(cpHost[0..i]).
	spHost := make([]uint64, nc*4)
	copy(spHost[:4], cpHost[:4]) // scanned[0] = cp[0]

	for i := 1; i < nc; i++ {
		var prev, cur, prod fr.Element
		prev[0] = spHost[(i-1)*4+0]
		prev[1] = spHost[(i-1)*4+1]
		prev[2] = spHost[(i-1)*4+2]
		prev[3] = spHost[(i-1)*4+3]
		cur[0] = cpHost[i*4+0]
		cur[1] = cpHost[i*4+1]
		cur[2] = cpHost[i*4+2]
		cur[3] = cpHost[i*4+3]
		prod.Mul(&prev, &cur)
		spHost[i*4+0] = prod[0]
		spHost[i*4+1] = prod[1]
		spHost[i*4+2] = prod[2]
		spHost[i*4+3] = prod[3]
	}

	// Phase 3: GPU fixup + shift
	if err := toError(C.gnark_gpu_z_prefix_phase3(
		devCtx(dev), zVec.handle, tempVec.handle,
		(*C.uint64_t)(unsafe.Pointer(&spHost[0])),
		C.size_t(nc),
	)); err != nil {
		panic("gpu: ZPrefixProduct phase3 failed: " + err.Error())
	}
}

// PolyEvalGPU evaluates a polynomial (already on GPU in SoA format) at a single
// point z using chunked Horner on GPU + CPU combine. Returns the result.
func PolyEvalGPU(dev *gpu.Device, vec *FrVector, z fr.Element) fr.Element {
	n := vec.n
	if n == 0 {
		return fr.Element{}
	}

	maxChunks := (n + 1023) / 1024
	partialsHost := make([]uint64, maxChunks*4)
	var numChunks C.size_t

	zArr := [4]C.uint64_t{C.uint64_t(z[0]), C.uint64_t(z[1]), C.uint64_t(z[2]), C.uint64_t(z[3])}

	if err := toError(C.gnark_gpu_poly_eval_chunks(
		devCtx(dev), vec.handle,
		&zArr[0],
		(*C.uint64_t)(unsafe.Pointer(&partialsHost[0])),
		&numChunks,
	)); err != nil {
		panic("gpu: PolyEvalGPU failed: " + err.Error())
	}

	nc := int(numChunks)
	return combinePolyEvalPartials(partialsHost, nc, z)
}

// PolyEvalFromHost evaluates a host polynomial on GPU. Uploads coefficients to a
// temporary GPU vector, evaluates, and returns the result.
func PolyEvalFromHost(dev *gpu.Device, coeffs fr.Vector, z fr.Element) fr.Element {
	n := len(coeffs)
	if n == 0 {
		return fr.Element{}
	}

	vec, err := NewFrVector(dev, n)
	if err != nil {
		// Fallback to CPU Horner
		return evalHorner(coeffs, z)
	}
	defer vec.Free()
	vec.CopyFromHost(coeffs)

	return PolyEvalGPU(dev, vec, z)
}

// combinePolyEvalPartials combines GPU chunked Horner partial results on CPU.
// partials[j] is the Horner result for chunk j.
// Full polynomial = Σ_j partials[j] * z^(j*K) where K=1024.
func combinePolyEvalPartials(partialsHost []uint64, nc int, z fr.Element) fr.Element {
	if nc == 0 {
		return fr.Element{}
	}
	if nc == 1 {
		var r fr.Element
		r[0] = partialsHost[0]
		r[1] = partialsHost[1]
		r[2] = partialsHost[2]
		r[3] = partialsHost[3]
		return r
	}

	// Compute z^K where K = 1024
	var zK fr.Element
	zK.Set(&z).Exp(zK, big.NewInt(1024))

	// Horner with z^K: result = partials[nc-1]; for j=nc-2 downto 0: result = result * zK + partials[j]
	var result fr.Element
	result[0] = partialsHost[(nc-1)*4+0]
	result[1] = partialsHost[(nc-1)*4+1]
	result[2] = partialsHost[(nc-1)*4+2]
	result[3] = partialsHost[(nc-1)*4+3]

	for j := nc - 2; j >= 0; j-- {
		result.Mul(&result, &zK)
		var p fr.Element
		p[0] = partialsHost[j*4+0]
		p[1] = partialsHost[j*4+1]
		p[2] = partialsHost[j*4+2]
		p[3] = partialsHost[j*4+3]
		result.Add(&result, &p)
	}

	return result
}

// evalHorner is the CPU fallback for polynomial evaluation.
func evalHorner(coeffs fr.Vector, z fr.Element) fr.Element {
	var r fr.Element
	for i := len(coeffs) - 1; i >= 0; i-- {
		r.Mul(&r, &z).Add(&r, &coeffs[i])
	}
	return r
}

// ReduceBlindedCoset reduces a blinded polynomial for coset evaluation on GPU.
//
// Given blinded poly coefficients [c₀..c_{n-1}] in src (SoA on GPU) and a
// tiny tail [c_n, c_{n+1}, ...] (AoS on host), computes:
//
//	dst[i] = src[i] + tail[i] · cosetPowN   for i < tail_len
//	dst[i] = src[i]                           for i ≥ tail_len
//
// tail_len is typically 2-3 (blinding order). The tiny tail is uploaded
// inline via stream-ordered alloc — no persistent device buffer needed.
func ReduceBlindedCoset(dst, src *FrVector, tail []fr.Element, cosetPowN fr.Element) {
	if dst.n != src.n {
		panic("gpu: ReduceBlindedCoset size mismatch")
	}
	if dst.dev != src.dev {
		panic("gpu: ReduceBlindedCoset device mismatch")
	}

	var tailPtr *C.uint64_t
	tailLen := len(tail)
	if tailLen > 0 {
		tailPtr = (*C.uint64_t)(unsafe.Pointer(&tail[0]))
	}

	if err := toError(C.gnark_gpu_reduce_blinded_coset(
		devCtx(dst.dev),
		dst.handle,
		src.handle,
		tailPtr,
		C.size_t(tailLen),
		(*C.uint64_t)(unsafe.Pointer(&cosetPowN)),
	)); err != nil {
		panic("gpu: ReduceBlindedCoset failed: " + err.Error())
	}
}

// PlonkZComputeFactors computes per-element Z polynomial ratio factors on GPU.
//
// For each i in [0, n):
//
//	num[i] = (L[i]+β·ω^i+γ) · (R[i]+β·g·ω^i+γ) · (O[i]+β·g²·ω^i+γ)
//	den[i] = (L[i]+β·id[S[i]]+γ) · (R[i]+β·id[S[n+i]]+γ) · (O[i]+β·id[S[2n+i]]+γ)
//
// On exit: L contains numerators, R contains denominators. O is unchanged.
// The caller should BatchInvert(R), then Mul(L, L, R) to get ratios.
func PlonkZComputeFactors(
	L, R, O *FrVector,
	dPerm unsafe.Pointer,
	beta, gamma, gMul, gSq fr.Element,
	log2n uint,
	domain *GPUFFTDomain,
) {
	n := L.n
	if R.n != n || O.n != n {
		panic("gpu: PlonkZComputeFactors size mismatch")
	}
	if domain.size != n {
		panic("gpu: PlonkZComputeFactors domain size mismatch")
	}
	dev := L.dev
	if R.dev != dev || O.dev != dev {
		panic("gpu: PlonkZComputeFactors device mismatch")
	}

	var params [16]uint64
	copy(params[0:4], beta[:])
	copy(params[4:8], gamma[:])
	copy(params[8:12], gMul[:])
	copy(params[12:16], gSq[:])

	if err := toError(C.gnark_gpu_plonk_z_compute_factors(
		devCtx(dev),
		L.handle, R.handle, O.handle,
		dPerm,
		(*C.uint64_t)(unsafe.Pointer(&params[0])),
		C.uint(log2n),
		domain.handle,
	)); err != nil {
		panic("gpu: PlonkZComputeFactors failed: " + err.Error())
	}
}

// ComputeL1Den computes out[i] = cosetGen · ω^i - 1 for all i in [0, n).
// Uses twiddle factors from the NTT domain (already on GPU).
// The caller should BatchInvert the result to get L1DenInv.
func ComputeL1Den(out *FrVector, cosetGen fr.Element, domain *GPUFFTDomain, stream ...gpu.StreamID) {
	if domain.size != out.n {
		panic("gpu: ComputeL1Den domain size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_compute_l1_den_stream(
			devCtx(out.dev), out.handle,
			(*C.uint64_t)(unsafe.Pointer(&cosetGen)),
			domain.handle, C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_compute_l1_den(
			devCtx(out.dev), out.handle,
			(*C.uint64_t)(unsafe.Pointer(&cosetGen)),
			domain.handle,
		))
	}
	if err != nil {
		panic("gpu: ComputeL1Den failed: " + err.Error())
	}
}

// PatchElements writes a small number of AoS host elements into the SoA GPU vector
// starting at the given offset. Useful for blinding correction patches.
func (v *FrVector) PatchElements(offset int, elems []fr.Element) {
	if offset+len(elems) > v.n {
		panic("gpu: PatchElements out of bounds")
	}
	if len(elems) == 0 {
		return
	}
	if err := toError(C.gnark_gpu_fr_vector_patch(
		devCtx(v.dev),
		v.handle,
		C.size_t(offset),
		(*C.uint64_t)(unsafe.Pointer(&elems[0])),
		C.size_t(len(elems)),
	)); err != nil {
		panic("gpu: PatchElements failed: " + err.Error())
	}
}

// DeviceAllocCopyInt64 uploads an int64 slice to GPU device memory.
// Returns a device pointer that must be freed with DeviceFreePtr.
func DeviceAllocCopyInt64(dev *gpu.Device, data []int64) (unsafe.Pointer, error) {
	var dPtr unsafe.Pointer
	if err := toError(C.gnark_gpu_device_alloc_copy_int64(
		devCtx(dev),
		(*C.int64_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&dPtr,
	)); err != nil {
		return nil, err
	}
	return dPtr, nil
}

// DeviceFreePtr frees device memory allocated by DeviceAllocCopyInt64.
func DeviceFreePtr(ptr unsafe.Pointer) {
	if ptr != nil {
		C.gnark_gpu_device_free_ptr(ptr)
	}
}

// PlonkGateAccum computes the fused gate constraint accumulation for PlonK quotient.
//
//	result[i] = (result[i] + Ql[i]·L[i] + Qr[i]·R[i] + Qm[i]·L[i]·R[i]
//	            - Qo[i]·O[i] + Qk[i]) · zhKInv
//
// result already contains permutation+boundary contributions.
func PlonkGateAccum(
	result, Ql, Qr, Qm, Qo, Qk, L, R, O *FrVector,
	zhKInv fr.Element,
) {
	n := result.n
	if Ql.n != n || Qr.n != n || Qm.n != n || Qo.n != n || Qk.n != n ||
		L.n != n || R.n != n || O.n != n {
		panic("gpu: PlonkGateAccum size mismatch")
	}
	dev := result.dev
	if Ql.dev != dev || Qr.dev != dev || Qm.dev != dev || Qo.dev != dev || Qk.dev != dev ||
		L.dev != dev || R.dev != dev || O.dev != dev {
		panic("gpu: PlonkGateAccum device mismatch")
	}

	if err := toError(C.gnark_gpu_plonk_gate_accum(
		devCtx(dev),
		result.handle,
		Ql.handle, Qr.handle, Qm.handle, Qo.handle, Qk.handle,
		L.handle, R.handle, O.handle,
		(*C.uint64_t)(unsafe.Pointer(&zhKInv)),
	)); err != nil {
		panic("gpu: PlonkGateAccum failed: " + err.Error())
	}
}

// PlonkPermBoundary computes the fused permutation + boundary constraint for PlonK.
//
// For each i in [0, n):
//
//	x_i = cosetGen · ω^i
//	num = Z[i]·(L[i]+β·x_i+γ)·(R[i]+β·x_i·k1+γ)·(O[i]+β·x_i·k2+γ)
//	den = Z[(i+1)%n]·(L[i]+β·S1[i]+γ)·(R[i]+β·S2[i]+γ)·(O[i]+β·S3[i]+γ)
//	L1_i = l1Scalar · L1DenInv[i]
//	result[i] = α · ((den-num) + α · (Z[i]-1) · L1_i)
func PlonkPermBoundary(
	result, L, R, O, Z, S1, S2, S3, L1DenInv *FrVector,
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen fr.Element,
	domain *GPUFFTDomain,
	stream ...gpu.StreamID,
) {
	n := result.n
	if L.n != n || R.n != n || O.n != n || Z.n != n ||
		S1.n != n || S2.n != n || S3.n != n || L1DenInv.n != n {
		panic("gpu: PlonkPermBoundary size mismatch")
	}
	if domain.size != n {
		panic("gpu: PlonkPermBoundary domain size mismatch")
	}
	dev := result.dev
	if L.dev != dev || R.dev != dev || O.dev != dev || Z.dev != dev ||
		S1.dev != dev || S2.dev != dev || S3.dev != dev || L1DenInv.dev != dev {
		panic("gpu: PlonkPermBoundary device mismatch")
	}

	// Pack scalar params: 7 × 4 uint64s = 28 uint64s
	var params [28]uint64
	copy(params[0:4], alpha[:])
	copy(params[4:8], beta[:])
	copy(params[8:12], gamma[:])
	copy(params[12:16], l1Scalar[:])
	copy(params[16:20], cosetShift[:])
	copy(params[20:24], cosetShiftSq[:])
	copy(params[24:28], cosetGen[:])

	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_plonk_perm_boundary_stream(
			devCtx(dev),
			result.handle,
			L.handle, R.handle, O.handle,
			Z.handle,
			S1.handle, S2.handle, S3.handle,
			L1DenInv.handle,
			(*C.uint64_t)(unsafe.Pointer(&params[0])),
			domain.handle, C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_plonk_perm_boundary(
			devCtx(dev),
			result.handle,
			L.handle, R.handle, O.handle,
			Z.handle,
			S1.handle, S2.handle, S3.handle,
			L1DenInv.handle,
			(*C.uint64_t)(unsafe.Pointer(&params[0])),
			domain.handle,
		))
	}
	if err != nil {
		panic("gpu: PlonkPermBoundary failed: " + err.Error())
	}
}
