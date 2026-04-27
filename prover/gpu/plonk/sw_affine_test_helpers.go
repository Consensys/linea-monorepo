//go:build cuda

package plonk

// Test helpers exposing the GPU SW affine validation entrypoints.
// These wrap the C functions registered in api.cu / msm.cu and are intended
// for use only by external tests in package plonk_test (which cannot directly
// invoke cgo). They are not part of the public API.

/*
#include "gnark_gpu.h"
*/
import "C"
import "unsafe"

// TestSWPairAddGPU runs a single SW affine pair-add on the GPU.
// Inputs/outputs use gnark's bls12377.G1Affine memory layout (12 uint64 limbs
// in Montgomery form: x[0..6] then y[0..6]).
func TestSWPairAddGPU(p0, p1, out *[12]uint64) error {
	return toError(C.gnark_gpu_test_sw_pair_add(
		(*C.uint64_t)(unsafe.Pointer(&p0[0])),
		(*C.uint64_t)(unsafe.Pointer(&p1[0])),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	))
}

// TestSWToTEGPU runs the SW affine → TE extended conversion on the GPU.
// Output is 24 uint64 limbs in Montgomery form (X[0..6], Y[0..6], T[0..6], Z[0..6]).
func TestSWToTEGPU(pSW *[12]uint64, outTE *[24]uint64) error {
	return toError(C.gnark_gpu_test_sw_to_te(
		(*C.uint64_t)(unsafe.Pointer(&pSW[0])),
		(*C.uint64_t)(unsafe.Pointer(&outTE[0])),
	))
}

// TestBatchedAffineReduceGPU reduces N (≤ 256) affine SW points via the same
// pairwise-reduce pipeline used in the bucket accumulator. Returns the SW
// affine sum.
func TestBatchedAffineReduceGPU(pointsAoS []uint64, out *[12]uint64, N int) error {
	if len(pointsAoS) < N*12 {
		panic("pointsAoS smaller than N*12")
	}
	return toError(C.gnark_gpu_test_batched_affine_reduce(
		(*C.uint64_t)(unsafe.Pointer(&pointsAoS[0])),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
		C.int(N),
	))
}
