#pragma once

// ─────────────────────────────────────────────────────────────────────────────
// BLS12-377 scalar field Fr arithmetic (253 bits, 4 × 64-bit limbs)
//
// All elements are in Montgomery form: ā = a · R mod q, where R = 2²⁵⁶.
//
// q = 0x12ab655e9a2ca556_60b44d1e5c37b001_59aa76fed0000001_0a11800000000001
//   = 8444461749428370424248824938781546531375899335154063827935233455917409239041
//
// All functions are __forceinline__: NTT butterfly kernels use only ~20 regs/thread
// (1 fr_mul + 2 fr_add/sub per butterfly), unlike MSM's EC ops which inline 7-9
// fp_mul calls and blow up to 166+ regs. No register pressure concern here.
//
// Operations:
//   fr_add: (a + b) mod q   PTX carry chain + conditional subtract   (8 asm)
//   fr_sub: (a - b) mod q   PTX borrow chain + conditional add-back  (8 asm)
//   fr_mul: a · b · R⁻¹ mod q  CIOS Montgomery multiply              (32 mul + 32 add)
// ─────────────────────────────────────────────────────────────────────────────

#include "field.cuh"
#include <cuda_runtime.h>

namespace gnark_gpu {

// Fr modulus as device constant (mirrors Fr_params::MODULUS)
__device__ __constant__ const uint64_t FR_MODULUS[4] = {
	0x0a11800000000001ULL,
	0x59aa76fed0000001ULL,
	0x60b44d1e5c37b001ULL,
	0x12ab655e9a2ca556ULL,
};

// =============================================================================
// Fr modular addition: r = (a + b) mod q
// =============================================================================

__device__ __forceinline__ void fr_add(uint64_t r[4], const uint64_t a[4], const uint64_t b[4]) {
    constexpr uint64_t q0 = Fr_params::MODULUS[0], q1 = Fr_params::MODULUS[1];
    constexpr uint64_t q2 = Fr_params::MODULUS[2], q3 = Fr_params::MODULUS[3];

    uint64_t s0, s1, s2, s3, carry;
    asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(s0) : "l"(a[0]), "l"(b[0]));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s1) : "l"(a[1]), "l"(b[1]));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s2) : "l"(a[2]), "l"(b[2]));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s3) : "l"(a[3]), "l"(b[3]));
    asm volatile("addc.u64 %0, 0, 0;" : "=l"(carry));

    uint64_t t0, t1, t2, t3, borrow;
    asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(t0) : "l"(s0), "l"(q0));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t1) : "l"(s1), "l"(q1));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t2) : "l"(s2), "l"(q2));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t3) : "l"(s3), "l"(q3));
    asm volatile("subc.u64 %0, %1, 0;" : "=l"(borrow) : "l"(carry));

    bool use_reduced = (borrow == 0);
    r[0] = use_reduced ? t0 : s0;
    r[1] = use_reduced ? t1 : s1;
    r[2] = use_reduced ? t2 : s2;
    r[3] = use_reduced ? t3 : s3;
}

// =============================================================================
// Fr modular subtraction: r = (a - b) mod q
// =============================================================================

__device__ __forceinline__ void fr_sub(uint64_t r[4], const uint64_t a[4], const uint64_t b[4]) {
    constexpr uint64_t q0 = Fr_params::MODULUS[0], q1 = Fr_params::MODULUS[1];
    constexpr uint64_t q2 = Fr_params::MODULUS[2], q3 = Fr_params::MODULUS[3];

    uint64_t s0, s1, s2, s3, borrow;
    asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s0) : "l"(a[0]), "l"(b[0]));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s1) : "l"(a[1]), "l"(b[1]));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s2) : "l"(a[2]), "l"(b[2]));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s3) : "l"(a[3]), "l"(b[3]));
    asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

    // Branchless correction: add q if borrow (same pattern as fp_sub)
    uint64_t mask = -(borrow != 0);
    uint64_t c0 = q0 & mask, c1 = q1 & mask, c2 = q2 & mask, c3 = q3 & mask;

    asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r[0]) : "l"(s0), "l"(c0));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[1]) : "l"(s1), "l"(c1));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[2]) : "l"(s2), "l"(c2));
    asm volatile("addc.u64 %0, %1, %2;" : "=l"(r[3]) : "l"(s3), "l"(c3));
}

// =============================================================================
// Fr CIOS Montgomery multiplication: r = a · b · R⁻¹ mod q
//
// Even/odd split CIOS with 8×32-bit limbs (ARITH23 technique, from sppark/yrrid).
// Splits accumulator into even[0,2,4,6] and odd[1,3,5,7] for unbroken PTX
// carry chains. Each chain is 8 instructions (4 limb-pairs).
//
// q mod 2^32 = 1, so -q^{-1} mod 2^32 = 0xFFFFFFFF.
// =============================================================================

// ── 32-bit PTX intrinsics for Fr ─────────────────────────────────────────────

static __device__ __forceinline__ uint32_t fr_ptx_mul_lo(uint32_t x, uint32_t y) {
	uint32_t r; asm("mul.lo.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_mul_hi(uint32_t x, uint32_t y) {
	uint32_t r; asm("mul.hi.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_mad_lo_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("mad.lo.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_madc_lo_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.lo.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_madc_hi_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.hi.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_madc_hi(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.hi.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_add_cc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("add.cc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_addc_cc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("addc.cc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
static __device__ __forceinline__ uint32_t fr_ptx_addc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("addc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}

// ── Fr multiply helpers (8×32-bit) ───────────────────────────────────────────

// Initial multiply: acc[i]=lo(a[i]*bi), acc[i+1]=hi(a[i]*bi) for even i
static __device__ __forceinline__ void fr32_mul_n(uint32_t *acc, const uint32_t *a, uint32_t bi) {
	#pragma unroll
	for (int i = 0; i < 8; i += 2) {
		acc[i]     = fr_ptx_mul_lo(a[i], bi);
		acc[i + 1] = fr_ptx_mul_hi(a[i], bi);
	}
}

// Chained multiply-accumulate with unbroken carry chain (8 instructions)
static __device__ __forceinline__ void fr32_cmad_n(uint32_t *acc, const uint32_t *a, uint32_t bi) {
	acc[0] = fr_ptx_mad_lo_cc(a[0], bi, acc[0]);
	acc[1] = fr_ptx_madc_hi_cc(a[0], bi, acc[1]);
	acc[2] = fr_ptx_madc_lo_cc(a[2], bi, acc[2]);
	acc[3] = fr_ptx_madc_hi_cc(a[2], bi, acc[3]);
	acc[4] = fr_ptx_madc_lo_cc(a[4], bi, acc[4]);
	acc[5] = fr_ptx_madc_hi_cc(a[4], bi, acc[5]);
	acc[6] = fr_ptx_madc_lo_cc(a[6], bi, acc[6]);
	acc[7] = fr_ptx_madc_hi_cc(a[6], bi, acc[7]);
}

// Right-shifted multiply-accumulate (consumes carry from previous op)
static __device__ __forceinline__ void fr32_madc_n_rshift(uint32_t *odd, const uint32_t *a, uint32_t bi) {
	odd[0] = fr_ptx_madc_lo_cc(a[0], bi, odd[2]);
	odd[1] = fr_ptx_madc_hi_cc(a[0], bi, odd[3]);
	odd[2] = fr_ptx_madc_lo_cc(a[2], bi, odd[4]);
	odd[3] = fr_ptx_madc_hi_cc(a[2], bi, odd[5]);
	odd[4] = fr_ptx_madc_lo_cc(a[4], bi, odd[6]);
	odd[5] = fr_ptx_madc_hi_cc(a[4], bi, odd[7]);
	odd[6] = fr_ptx_madc_lo_cc(a[6], bi, 0);
	odd[7] = fr_ptx_madc_hi(a[6], bi, 0);
}

// One fused multiply + Montgomery reduction step
static __device__ __forceinline__ void fr32_mad_n_redc(
	uint32_t *even, uint32_t *odd, const uint32_t *a, uint32_t bi,
	const uint32_t *MOD, bool first) {
	if (first) {
		fr32_mul_n(odd, a + 1, bi);
		fr32_mul_n(even, a, bi);
	} else {
		even[0] = fr_ptx_add_cc(even[0], odd[1]);
		fr32_madc_n_rshift(odd, a + 1, bi);
		fr32_cmad_n(even, a, bi);
		odd[7] = fr_ptx_addc(odd[7], 0);
	}
	uint32_t mi = even[0] * 0xFFFFFFFFu; // -q⁻¹ mod 2³² (q ≡ 1 mod 2³²)
	fr32_cmad_n(odd, MOD + 1, mi);
	fr32_cmad_n(even, MOD, mi);
	odd[7] = fr_ptx_addc(odd[7], 0);
}

// Fr Montgomery multiplication using 8×32-bit even/odd split CIOS
__device__ __forceinline__ void fr_mul(uint64_t r[4], const uint64_t a[4], const uint64_t b[4]) {
	const uint32_t *a32 = (const uint32_t *)a;
	const uint32_t *b32 = (const uint32_t *)b;
	const uint32_t *MOD = (const uint32_t *)FR_MODULUS;

	__align__(8) uint32_t even[8];
	__align__(8) uint32_t odd[8];

	#pragma unroll
	for (int i = 0; i < 8; i += 2) {
		fr32_mad_n_redc(even, odd, a32, b32[i], MOD, i == 0);
		fr32_mad_n_redc(odd, even, a32, b32[i + 1], MOD, false);
	}

	// Merge even and odd arrays
	even[0] = fr_ptx_add_cc(even[0], odd[1]);
	even[1] = fr_ptx_addc_cc(even[1], odd[2]);
	even[2] = fr_ptx_addc_cc(even[2], odd[3]);
	even[3] = fr_ptx_addc_cc(even[3], odd[4]);
	even[4] = fr_ptx_addc_cc(even[4], odd[5]);
	even[5] = fr_ptx_addc_cc(even[5], odd[6]);
	even[6] = fr_ptx_addc_cc(even[6], odd[7]);
	even[7] = fr_ptx_addc(even[7], 0);

	// Final reduction: branchless conditional subtract q
	const uint64_t *e64 = (const uint64_t *)even;
	const uint64_t *q = FR_MODULUS;
	uint64_t s[4], borrow;

	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(e64[0]), "l"(q[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(e64[1]), "l"(q[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(e64[2]), "l"(q[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(e64[3]), "l"(q[3]));
	asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

	uint64_t mask = -(borrow != 0);
	r[0] = (e64[0] & mask) | (s[0] & ~mask);
	r[1] = (e64[1] & mask) | (s[1] & ~mask);
	r[2] = (e64[2] & mask) | (s[2] & ~mask);
	r[3] = (e64[3] & mask) | (s[3] & ~mask);
}

} // namespace gnark_gpu
