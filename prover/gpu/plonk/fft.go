//go:build cuda

package plonk

/*
#include "gnark_gpu.h"
*/
import "C"
import (
	"math/big"
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// GPUFFTDomain holds GPU-resident twiddle factors for Number Theoretic Transform
// (NTT) operations over the BLS12-377 scalar field.
//
// The NTT is a finite-field analog of the FFT, operating modulo a prime r.
// Given domain size n (power of 2) and primitive n-th root of unity ω:
//
//	Forward NTT (DIF):  ŷ[k] = Σᵢ y[i] · ωⁱᵏ     (evaluation form)
//	Inverse NTT (DIT):  y[i] = (1/n) Σₖ ŷ[k] · ω⁻ⁱᵏ  (coefficient form)
//
// The Cooley-Tukey butterfly structure:
//
//	DIF (Decimation-In-Frequency):
//	  ┌───┐         ┌───┐
//	  │ a │────+────►│ a'│   a' = a + b
//	  └───┘    │     └───┘
//	           ×w
//	  ┌───┐    │     ┌───┐
//	  │ b │────+────►│ b'│   b' = (a - b) · ω
//	  └───┘          └───┘
//
//	Natural input → bit-reversed output.
//
//	DIT (Decimation-In-Time):
//	  ┌───┐         ┌───┐
//	  │ a │────+────►│ a'│   a' = a + ω·b
//	  └───┘    │     └───┘
//	           ×w
//	  ┌───┐    │     ┌───┐
//	  │ b │────+────►│ b'│   b' = a - ω·b
//	  └───┘          └───┘
//
//	Bit-reversed input → natural output (scaled by 1/n).
//
// Implementation uses radix-4 + radix-2 kernels with a fused tail
// (last 9 stages in shared memory, 512 threads per block).
//
// All NTT operations accept an optional StreamID for multi-stream pipelining.
//
// Create with Device.NewFFTDomain and release with Close.
type GPUFFTDomain struct {
	handle C.gnark_gpu_ntt_domain_t
	dev    *gpu.Device
	size   int
}

// NewFFTDomain creates a GPU NTT domain for the given size (must be power of 2).
//
// Twiddle factors are computed on CPU using gnark-crypto's fft.Domain,
// then uploaded to GPU in SoA format. This is a one-time cost per domain size.
//
// Twiddle table layout (n/2 elements each):
//
//	Forward: ω⁰, ω¹, ω², …, ωⁿ/²⁻¹   where ω = primitive n-th root of unity
//	Inverse: ω̄⁰, ω̄¹, ω̄², …, ω̄ⁿ/²⁻¹   where ω̄ = ω⁻¹
//
// Memory: 32n bytes total (forward + inverse twiddles).
func NewFFTDomain(dev *gpu.Device, size int) (*GPUFFTDomain, error) {
	if devCtx(dev) == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if size <= 0 || (size&(size-1)) != 0 {
		return nil, &gpu.Error{Code: -1, Message: "size must be a positive power of 2"}
	}

	// Use gnark-crypto to get the primitive root of unity
	domain := fft.NewDomain(uint64(size))

	halfN := size / 2

	// Build flat twiddle tables: w^0, w^1, ..., w^(n/2-1)
	// Forward twiddles use domain.Generator (omega)
	// Inverse twiddles use domain.GeneratorInv (omega^{-1})
	fwdTwiddles := make([]fr.Element, halfN)
	invTwiddles := make([]fr.Element, halfN)

	if halfN > 0 {
		fwdTwiddles[0].SetOne()
		invTwiddles[0].SetOne()
		for i := 1; i < halfN; i++ {
			fwdTwiddles[i].Mul(&fwdTwiddles[i-1], &domain.Generator)
			invTwiddles[i].Mul(&invTwiddles[i-1], &domain.GeneratorInv)
		}
	}

	// CardinalityInv = 1/n in Montgomery form
	invN := domain.CardinalityInv

	var handle C.gnark_gpu_ntt_domain_t
	var fwdPtr, invPtr *C.uint64_t
	if halfN > 0 {
		fwdPtr = (*C.uint64_t)(unsafe.Pointer(&fwdTwiddles[0]))
		invPtr = (*C.uint64_t)(unsafe.Pointer(&invTwiddles[0]))
	}
	invNPtr := (*C.uint64_t)(unsafe.Pointer(&invN))

	if err := toError(C.gnark_gpu_ntt_domain_create(
		devCtx(dev),
		C.size_t(size),
		fwdPtr,
		invPtr,
		invNPtr,
		&handle,
	)); err != nil {
		return nil, err
	}

	dom := &GPUFFTDomain{handle: handle, dev: dev, size: size}
	runtime.SetFinalizer(dom, (*GPUFFTDomain).Close)
	return dom, nil
}

// Size returns the domain size.
func (f *GPUFFTDomain) Size() int {
	return f.size
}

// Close releases GPU resources associated with this FFT domain.
// It is safe to call Close multiple times.
func (f *GPUFFTDomain) Close() {
	if f.handle != nil {
		C.gnark_gpu_ntt_domain_destroy(f.handle)
		f.handle = nil
		runtime.SetFinalizer(f, nil)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Forward / Inverse FFT
// ─────────────────────────────────────────────────────────────────────────────

// FFT performs a forward NTT (DIF): natural-order input → bit-reversed output.
//
//	For log₂(n) stages s = log₂(n)−1 down to 0:
//	  half = 2ˢ
//	  For each group j and element k:
//	    a, b = v[j+k], v[j+k+half]
//	    v[j+k]      = a + b
//	    v[j+k+half] = (a - b) · ω[twiddle_idx]
//
// The data vector must have exactly Size() elements.
// The operation is asynchronous; call dev.Sync() to wait for completion.
func (f *GPUFFTDomain) FFT(v *FrVector, stream ...gpu.StreamID) {
	if v.n != f.size {
		panic("gpu: FFT size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_ntt_forward_stream(f.handle, v.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_ntt_forward(f.handle, v.handle))
	}
	if err != nil {
		panic("gpu: FFT failed: " + err.Error())
	}
}

// FFTInverse performs an inverse NTT (DIT): bit-reversed input → natural-order output.
// The result is scaled by 1/n. The data vector must have exactly Size() elements.
// The operation is asynchronous; call dev.Sync() to wait for completion.
func (f *GPUFFTDomain) FFTInverse(v *FrVector, stream ...gpu.StreamID) {
	if v.n != f.size {
		panic("gpu: FFTInverse size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_ntt_inverse_stream(f.handle, v.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_ntt_inverse(f.handle, v.handle))
	}
	if err != nil {
		panic("gpu: FFTInverse failed: " + err.Error())
	}
}

// BitReverse performs a bit-reversal permutation on the data vector.
//
//	For each index i with bit-reversed counterpart j = bitrev(i):
//	  swap v[i], v[j]
//
// The data vector must have exactly Size() elements.
func (f *GPUFFTDomain) BitReverse(v *FrVector, stream ...gpu.StreamID) {
	if v.n != f.size {
		panic("gpu: BitReverse size mismatch")
	}
	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_ntt_bit_reverse_stream(f.handle, v.handle, C.int(stream[0])))
	} else {
		err = toError(C.gnark_gpu_ntt_bit_reverse(f.handle, v.handle))
	}
	if err != nil {
		panic("gpu: BitReverse failed: " + err.Error())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Coset FFT
//
// A coset FFT evaluates a polynomial p(X) on the coset g·H = {g·ωⁱ : i=0..n-1}
// instead of the subgroup H = {ωⁱ}. This is equivalent to:
//
//   1. Scale coefficients: p̃[i] = p[i] · gⁱ     (ScaleByPowers)
//   2. Forward FFT:        ŷ = FFT(p̃)
//
// The fused CosetFFT kernel combines steps 1+2 into a single kernel launch,
// saving one memory round-trip over the two-step approach.
// ─────────────────────────────────────────────────────────────────────────────

// CosetFFT evaluates a polynomial in canonical form on coset g·H.
// Input: v holds canonical coefficients in natural order.
// Output: v holds p(g·ω⁰), p(g·ω¹), …, p(g·ωⁿ⁻¹) in natural order.
// Uses fused ScaleByPowers+DIF+BitReverse kernel (saves one memory round-trip).
func (f *GPUFFTDomain) CosetFFT(v *FrVector, g fr.Element, stream ...gpu.StreamID) {
	if v.n != f.size {
		panic("gpu: CosetFFT size mismatch")
	}
	// Precompute g^(n/2) for the fused kernel
	var gHalf fr.Element
	gHalf.Set(&g).Exp(gHalf, big.NewInt(int64(f.size/2)))

	var err error
	if len(stream) > 0 {
		err = toError(C.gnark_gpu_ntt_forward_coset_stream(
			f.handle, v.handle,
			(*C.uint64_t)(unsafe.Pointer(&g)),
			(*C.uint64_t)(unsafe.Pointer(&gHalf)),
			C.int(stream[0]),
		))
	} else {
		err = toError(C.gnark_gpu_ntt_forward_coset(
			f.handle, v.handle,
			(*C.uint64_t)(unsafe.Pointer(&g)),
			(*C.uint64_t)(unsafe.Pointer(&gHalf)),
		))
	}
	if err != nil {
		panic("gpu: CosetFFT failed: " + err.Error())
	}
}

// CosetFFTInverse recovers canonical coefficients from coset evaluations.
// Input: v holds evaluations in natural order on coset g·H.
// Output: v holds canonical coefficients in natural order.
// gInv must be the inverse of the coset generator g.
//
// Decomposition: BitReverse → InverseFFT → ScaleByPowers(g⁻¹)
func (f *GPUFFTDomain) CosetFFTInverse(v *FrVector, gInv fr.Element, stream ...gpu.StreamID) {
	if v.n != f.size {
		panic("gpu: CosetFFTInverse size mismatch")
	}
	f.BitReverse(v, stream...)
	f.FFTInverse(v, stream...)
	v.ScaleByPowers(gInv, stream...)
}

// ─────────────────────────────────────────────────────────────────────────────
// Decomposed iFFT(4n) via Butterfly4
//
// For the PlonK quotient polynomial h(X) of degree < 4n, the iFFT(4n) can be
// decomposed into 4 independent iFFT(n) operations plus a size-4 butterfly:
//
//   Given evaluations ŷ[0..4n-1] on coset u·H₄ₙ, split into 4 sub-cosets:
//     block_k[i] = ŷ[4i+k]   for k=0,1,2,3 and i=0..n-1
//
//   Each block_k lives on sub-coset u·g₁ᵏ·H_n where g₁ = primitive 4n-th root.
//
//   Step 1: CosetFFTInverse each block_k independently (4 × iFFT(n))
//   Step 2: Butterfly4Inverse combines the 4 results
//   Step 3: Scale block_l by u⁻ˡⁿ for l=1,2,3
//
// This halves memory usage compared to a monolithic iFFT(4n) by working
// with n-sized vectors instead of 4n.
// ─────────────────────────────────────────────────────────────────────────────

// Butterfly4Inverse applies a size-4 inverse DFT butterfly across 4 FrVectors.
//
// Computes (element-wise for each index i):
//
//	t0 = b0[i] + b2[i]          t1 = b0[i] - b2[i]
//	t2 = b1[i] + b3[i]          t3 = (b1[i] - b3[i]) · ω₄⁻¹
//
//	b0[i] = (t0 + t2) / 4       b1[i] = (t1 + t3) / 4
//	b2[i] = (t0 - t2) / 4       b3[i] = (t1 - t3) / 4
//
// omega4Inv: inverse of primitive 4th root of unity.
// quarter: 1/4 in Montgomery form.
func Butterfly4Inverse(b0, b1, b2, b3 *FrVector, omega4Inv, quarter fr.Element) {
	if b0.n != b1.n || b1.n != b2.n || b2.n != b3.n {
		panic("gpu: Butterfly4Inverse size mismatch")
	}
	if b0.dev != b1.dev || b1.dev != b2.dev || b2.dev != b3.dev {
		panic("gpu: Butterfly4Inverse device mismatch")
	}
	if err := toError(C.gnark_gpu_fr_vector_butterfly4(
		devCtx(b0.dev),
		b0.handle, b1.handle, b2.handle, b3.handle,
		(*C.uint64_t)(unsafe.Pointer(&omega4Inv)),
		(*C.uint64_t)(unsafe.Pointer(&quarter)),
	)); err != nil {
		panic("gpu: Butterfly4Inverse failed: " + err.Error())
	}
}
