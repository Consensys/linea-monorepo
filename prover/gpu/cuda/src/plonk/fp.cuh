#pragma once

// ─────────────────────────────────────────────────────────────────────────────
// BLS12-377 base field Fp arithmetic (377 bits, 6 × 64-bit limbs)
//
// All elements are in Montgomery form: ā = a · R mod p, where R = 2³⁸⁴.
// This avoids a modular reduction after every multiply: instead of computing
// (a · b) mod p directly, we compute ā · b̄ · R⁻¹ mod p which equals (a·b)·R mod p.
//
// Operations are branchless (no warp divergence on GPU):
//   fp_add:  PTX carry chain + conditional select       (12 asm instructions)
//   fp_sub:  PTX borrow chain + masked correction       (12 asm instructions)
//   fp_mul:  CIOS Montgomery multiply (__noinline__)    (72 mul + 72 add)
//   fp_sqr:  Delegates to fp_mul(a, a)                  (same cost)
//
// p = 0x01ae3a4617c510ea_c63b05c06ca1493b_1a22d9f300f5138f_
//       1ef3622fba094800_170b5d4430000000_8508c00000000001
//   = 258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177
// ─────────────────────────────────────────────────────────────────────────────

#include "field.cuh"

namespace gnark_gpu {

// =============================================================================
// Fp Montgomery constants (BLS12-377 base field, 377 bits, 6 limbs)
// =============================================================================

// Fp modulus as device constant (mirrors Fp_params::MODULUS)
__device__ __constant__ const uint64_t FP_MODULUS[6] = {
	0x8508c00000000001ULL, 0x170b5d4430000000ULL, 0x1ef3622fba094800ULL,
	0x1a22d9f300f5138fULL, 0xc63b05c06ca1493bULL, 0x01ae3a4617c510eaULL,
};

// 2p — used by fp_sub_nr for lazy-reduction subtraction correction.
// Since p ≈ 2^377 and container is 384 bits, 2p ≈ 2^378 fits comfortably.
__device__ __constant__ const uint64_t FP_MODULUS_2X[6] = {
	0x0a11800000000002ULL, 0x2e16ba8860000001ULL, 0x3de6c45f74129000ULL,
	0x3445b3e601ea271eULL, 0x8c760b80d9429276ULL, 0x035c748c2f8a21d5ULL,
};

// -p^{-1} mod 2^64
__device__ __constant__ const uint64_t FP_INV = 0x8508bfffffffffffULL;

// R = 2^384 mod p (Montgomery one)
__device__ __constant__ const uint64_t FP_R[6] = {
	0x02cdffffffffff68ULL, 0x51409f837fffffb1ULL, 0x9f7db3a98a7d3ff2ULL,
	0x7b4e97b76e7c6305ULL, 0x4cf495bf803c84e8ULL, 0x008d6661e2fdf49aULL,
};

// R^2 = 2^768 mod p (for converting to Montgomery form)
__device__ __constant__ const uint64_t FP_R_SQUARED[6] = {
	0xb786686c9400cd22ULL, 0x0329fcaab00431b1ULL, 0x22a5f11162d6b46dULL,
	0xbfdf7d03827dc3acULL, 0x837e92f041790bf9ULL, 0x006dfccb1e914b88ULL,
};

// =============================================================================
// Fp arithmetic: add, sub, mul, sqr for 6-limb Montgomery form
// All operations are branchless to avoid warp divergence on GPU.
// =============================================================================

// fp_add: r = (a + b) mod p — branchless using PTX carry chains + conditional select
__device__ __forceinline__ void fp_add(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	const uint64_t *p = FP_MODULUS;

	uint64_t s[6], carry;

	// Add with carry chain
	asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(a[0]), "l"(b[0]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(a[1]), "l"(b[1]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(a[2]), "l"(b[2]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(a[3]), "l"(b[3]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s[4]) : "l"(a[4]), "l"(b[4]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(s[5]) : "l"(a[5]), "l"(b[5]));
	asm volatile("addc.u64 %0, 0, 0;" : "=l"(carry));

	// Subtract modulus
	uint64_t t[6], borrow;
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(t[0]) : "l"(s[0]), "l"(p[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[1]) : "l"(s[1]), "l"(p[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[2]) : "l"(s[2]), "l"(p[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[3]) : "l"(s[3]), "l"(p[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[4]) : "l"(s[4]), "l"(p[4]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[5]) : "l"(s[5]), "l"(p[5]));
	asm volatile("subc.u64 %0, %1, 0;" : "=l"(borrow) : "l"(carry));

	// Branchless select: if borrow, use unreduced sum; else use reduced
	// borrow == 0 means s >= p, use t (reduced)
	// borrow != 0 means s < p, use s (unreduced)
	uint64_t mask = -(borrow != 0); // 0xFFF...F if borrow, 0 if no borrow
	r[0] = (s[0] & mask) | (t[0] & ~mask);
	r[1] = (s[1] & mask) | (t[1] & ~mask);
	r[2] = (s[2] & mask) | (t[2] & ~mask);
	r[3] = (s[3] & mask) | (t[3] & ~mask);
	r[4] = (s[4] & mask) | (t[4] & ~mask);
	r[5] = (s[5] & mask) | (t[5] & ~mask);
}

// fp_sub: r = (a - b) mod p — branchless
__device__ __forceinline__ void fp_sub(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	const uint64_t *p = FP_MODULUS;

	uint64_t s[6], borrow;

	// Subtract with borrow chain
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(a[0]), "l"(b[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(a[1]), "l"(b[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(a[2]), "l"(b[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(a[3]), "l"(b[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[4]) : "l"(a[4]), "l"(b[4]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[5]) : "l"(a[5]), "l"(b[5]));
	asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

	// Branchless correction: mask = all-ones if borrow, zero if no borrow
	// If borrow: add p back (s + p). If no borrow: keep s.
	uint64_t mask = -(borrow != 0); // 0xFFF...F if underflow
	uint64_t corr[6];
	corr[0] = p[0] & mask;
	corr[1] = p[1] & mask;
	corr[2] = p[2] & mask;
	corr[3] = p[3] & mask;
	corr[4] = p[4] & mask;
	corr[5] = p[5] & mask;

	asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r[0]) : "l"(s[0]), "l"(corr[0]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[1]) : "l"(s[1]), "l"(corr[1]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[2]) : "l"(s[2]), "l"(corr[2]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[3]) : "l"(s[3]), "l"(corr[3]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[4]) : "l"(s[4]), "l"(corr[4]));
	asm volatile("addc.u64 %0, %1, %2;" : "=l"(r[5]) : "l"(s[5]), "l"(corr[5]));
}

// ═════════════════════════════════════════════════════════════════════════════
// Lazy-reduction ("_nr" = no reduce) variants for EC formulas
//
// These skip the final conditional subtraction, keeping results in [0, 2p)
// instead of [0, p). This is safe because:
//
//   1. p ≈ 2^377 fits in 384-bit container with 7 bits headroom
//   2. 2p, 4p, 6p all fit in 6 × 64-bit limbs (max intermediate ≈ 2^380)
//   3. CIOS with inputs < R = 2^384 produces output < 2p (since 4p < R)
//
// Bound tracking through TE mixed-add formula (worst-case per variable):
//
//   ┌──────────────────┬─────────────────────────────┬──────────────┐
//   │  Variable        │  Expression                 │  Bound       │
//   ├──────────────────┼─────────────────────────────┼──────────────┤
//   │  T_q             │  mul_nr(mul_nr(x,y), 2d)    │  [0, 2p)     │
//   │  A = (Y-X)(Yq-Xq)│ mul_nr(sub_nr, sub)         │  [0, 2p)     │
//   │  B = (Y+X)(Yq+Xq)│ mul_nr(add_nr, add_nr)      │  [0, 2p)     │
//   │  C = T1 · T_q    │  mul_nr(2p, 2p)             │  [0, 2p)     │
//   │  D = 2·Z1        │  add_nr(2p, 2p)             │  [0, 4p)     │
//   │  E = B - A       │  sub_nr(2p, 2p)             │  [0, 4p)     │
//   │  H = B + A       │  add_nr(2p, 2p)             │  [0, 4p)     │
//   │  F = D - C       │  sub_nr(4p, 2p)             │  [0, 6p)     │
//   │  G = D + C       │  add_nr(4p, 2p)             │  [0, 6p)     │
//   │  X3= E·F         │  mul_nr(4p, 6p)  → < 2p    │  [0, 2p)     │
//   │  Y3= G·H         │  mul_nr(6p, 4p)  → < 2p    │  [0, 2p)     │
//   │  T3= E·H         │  mul_nr(4p, 4p)  → < 2p    │  [0, 2p)     │
//   │  Z3= F·G         │  mul_nr(6p, 6p)  → < 2p    │  [0, 2p)     │
//   └──────────────────┴─────────────────────────────┴──────────────┘
//
//   All outputs ∈ [0, 2p): invariant is maintained across chained EC adds.
//
//   Max mul input 6p ≈ 2^380 < R = 2^384 ✓
//   CIOS output bound: (6p·6p)/R + p = 36p²/R + p < 2p ✓  (since 36p < R)
// ═════════════════════════════════════════════════════════════════════════════

// fp_reduce: conditional subtract p. Brings [0, 2p) → [0, p). Branchless.
__device__ __forceinline__ void fp_reduce(uint64_t r[6]) {
	const uint64_t *p = FP_MODULUS;
	uint64_t s[6], borrow;
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(r[0]), "l"(p[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(r[1]), "l"(p[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(r[2]), "l"(p[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(r[3]), "l"(p[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[4]) : "l"(r[4]), "l"(p[4]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[5]) : "l"(r[5]), "l"(p[5]));
	asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

	uint64_t mask = -(borrow != 0);
	r[0] = (r[0] & mask) | (s[0] & ~mask);
	r[1] = (r[1] & mask) | (s[1] & ~mask);
	r[2] = (r[2] & mask) | (s[2] & ~mask);
	r[3] = (r[3] & mask) | (s[3] & ~mask);
	r[4] = (r[4] & mask) | (s[4] & ~mask);
	r[5] = (r[5] & mask) | (s[5] & ~mask);
}

// fp_add_nr: r = a + b (no modular reduction)
// Inputs must satisfy a + b < 2^384 (guaranteed when a,b < 6p ≈ 2^380).
// Saves 12 instructions vs fp_add by skipping the conditional subtract.
__device__ __forceinline__ void fp_add_nr(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r[0]) : "l"(a[0]), "l"(b[0]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[1]) : "l"(a[1]), "l"(b[1]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[2]) : "l"(a[2]), "l"(b[2]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[3]) : "l"(a[3]), "l"(b[3]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[4]) : "l"(a[4]), "l"(b[4]));
	asm volatile("addc.u64 %0, %1, %2;" : "=l"(r[5]) : "l"(a[5]), "l"(b[5]));
}

// fp_sub_nr: r = (a - b), add 2p on borrow. Branchless.
// For unreduced inputs in [0, 2p): result in [0, 4p).
// For mixed inputs (one [0, 4p), one [0, 2p)): result in [0, 6p).
// Same instruction count as fp_sub, but uses 2p correction for wider inputs.
__device__ __forceinline__ void fp_sub_nr(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	const uint64_t *p2 = FP_MODULUS_2X;

	uint64_t s[6], borrow;
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(a[0]), "l"(b[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(a[1]), "l"(b[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(a[2]), "l"(b[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(a[3]), "l"(b[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[4]) : "l"(a[4]), "l"(b[4]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[5]) : "l"(a[5]), "l"(b[5]));
	asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

	uint64_t mask = -(borrow != 0);
	uint64_t corr[6];
	corr[0] = p2[0] & mask;
	corr[1] = p2[1] & mask;
	corr[2] = p2[2] & mask;
	corr[3] = p2[3] & mask;
	corr[4] = p2[4] & mask;
	corr[5] = p2[5] & mask;

	asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r[0]) : "l"(s[0]), "l"(corr[0]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[1]) : "l"(s[1]), "l"(corr[1]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[2]) : "l"(s[2]), "l"(corr[2]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[3]) : "l"(s[3]), "l"(corr[3]));
	asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r[4]) : "l"(s[4]), "l"(corr[4]));
	asm volatile("addc.u64 %0, %1, %2;" : "=l"(r[5]) : "l"(s[5]), "l"(corr[5]));
}

// ─────────────────────────────────────────────────────────────────────────────
// fp_mul: r = a · b mod p   (even/odd split CIOS, 12×32-bit limbs)
//
// Uses the ARITH23 even/odd split technique (adapted from sppark/yrrid-msm):
//   http://www.acsel-lab.com/arithmetic/arith23/data/1616a047.pdf
//
// Key idea: split the 12-limb accumulator into even[0,2,4,...] and odd[1,3,5,...]
// positions. Products a[even]*bi flow into the even array, a[odd]*bi into odd.
// Within each array, carries chain via PTX mad.lo.cc → madc.hi.cc (unbroken).
// After 12 iterations, merge even+odd for the final result.
//
// Each mad/madc pair is 2 native 32-bit instructions (1 cycle each, 128 ops/SM).
// vs old 64-bit CIOS: each mul64 decomposes to 4× mul32 + carry management.
//
// Register pressure: __noinline__ is CRITICAL for MSM performance.
// EC add functions must stay __forceinline__ (see ec.cuh header comment).
// ─────────────────────────────────────────────────────────────────────────────

// ── 32-bit PTX intrinsics ───────────────────────────────────────────────────

__device__ __forceinline__ uint32_t ptx_mul_lo(uint32_t x, uint32_t y) {
	uint32_t r; asm("mul.lo.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
__device__ __forceinline__ uint32_t ptx_mul_hi(uint32_t x, uint32_t y) {
	uint32_t r; asm("mul.hi.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
__device__ __forceinline__ uint32_t ptx_mad_lo_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("mad.lo.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
__device__ __forceinline__ uint32_t ptx_madc_lo_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.lo.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
__device__ __forceinline__ uint32_t ptx_madc_hi_cc(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.hi.cc.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
__device__ __forceinline__ uint32_t ptx_madc_hi(uint32_t x, uint32_t y, uint32_t z) {
	uint32_t r; asm volatile("madc.hi.u32 %0, %1, %2, %3;" : "=r"(r) : "r"(x), "r"(y), "r"(z)); return r;
}
__device__ __forceinline__ uint32_t ptx_add_cc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("add.cc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
__device__ __forceinline__ uint32_t ptx_addc_cc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("addc.cc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}
__device__ __forceinline__ uint32_t ptx_addc(uint32_t x, uint32_t y) {
	uint32_t r; asm volatile("addc.u32 %0, %1, %2;" : "=r"(r) : "r"(x), "r"(y)); return r;
}

// ── Multiply helpers ────────────────────────────────────────────────────────

// Initial multiply: acc[i]=lo(a[i]*bi), acc[i+1]=hi(a[i]*bi) for even i
static __device__ __forceinline__ void fp32_mul_n(uint32_t *acc, const uint32_t *a, uint32_t bi) {
	#pragma unroll
	for (int i = 0; i < 12; i += 2) {
		acc[i]     = ptx_mul_lo(a[i], bi);
		acc[i + 1] = ptx_mul_hi(a[i], bi);
	}
}

// Chained multiply-accumulate with unbroken carry chain (12 instructions)
static __device__ __forceinline__ void fp32_cmad_n(uint32_t *acc, const uint32_t *a, uint32_t bi) {
	acc[0]  = ptx_mad_lo_cc(a[0], bi, acc[0]);
	acc[1]  = ptx_madc_hi_cc(a[0], bi, acc[1]);
	acc[2]  = ptx_madc_lo_cc(a[2], bi, acc[2]);
	acc[3]  = ptx_madc_hi_cc(a[2], bi, acc[3]);
	acc[4]  = ptx_madc_lo_cc(a[4], bi, acc[4]);
	acc[5]  = ptx_madc_hi_cc(a[4], bi, acc[5]);
	acc[6]  = ptx_madc_lo_cc(a[6], bi, acc[6]);
	acc[7]  = ptx_madc_hi_cc(a[6], bi, acc[7]);
	acc[8]  = ptx_madc_lo_cc(a[8], bi, acc[8]);
	acc[9]  = ptx_madc_hi_cc(a[8], bi, acc[9]);
	acc[10] = ptx_madc_lo_cc(a[10], bi, acc[10]);
	acc[11] = ptx_madc_hi_cc(a[10], bi, acc[11]);
}

// Right-shifted multiply-accumulate (consumes carry from previous op)
static __device__ __forceinline__ void fp32_madc_n_rshift(uint32_t *odd, const uint32_t *a, uint32_t bi) {
	odd[0]  = ptx_madc_lo_cc(a[0], bi, odd[2]);
	odd[1]  = ptx_madc_hi_cc(a[0], bi, odd[3]);
	odd[2]  = ptx_madc_lo_cc(a[2], bi, odd[4]);
	odd[3]  = ptx_madc_hi_cc(a[2], bi, odd[5]);
	odd[4]  = ptx_madc_lo_cc(a[4], bi, odd[6]);
	odd[5]  = ptx_madc_hi_cc(a[4], bi, odd[7]);
	odd[6]  = ptx_madc_lo_cc(a[6], bi, odd[8]);
	odd[7]  = ptx_madc_hi_cc(a[6], bi, odd[9]);
	odd[8]  = ptx_madc_lo_cc(a[8], bi, odd[10]);
	odd[9]  = ptx_madc_hi_cc(a[8], bi, odd[11]);
	odd[10] = ptx_madc_lo_cc(a[10], bi, 0);
	odd[11] = ptx_madc_hi(a[10], bi, 0);
}

// One fused multiply + Montgomery reduction step
static __device__ __forceinline__ void fp32_mad_n_redc(
	uint32_t *even, uint32_t *odd, const uint32_t *a, uint32_t bi,
	const uint32_t *MOD, bool first) {
	if (first) {
		fp32_mul_n(odd, a + 1, bi);
		fp32_mul_n(even, a, bi);
	} else {
		even[0] = ptx_add_cc(even[0], odd[1]);
		fp32_madc_n_rshift(odd, a + 1, bi);
		fp32_cmad_n(even, a, bi);
		odd[11] = ptx_addc(odd[11], 0);
	}
	uint32_t mi = even[0] * 0xFFFFFFFFu; // -p⁻¹ mod 2³² (p ≡ 1 mod 2³²)
	fp32_cmad_n(odd, MOD + 1, mi);
	fp32_cmad_n(even, MOD, mi);
	odd[11] = ptx_addc(odd[11], 0);
}

// ── fp_mul and fp_mul_nr ────────────────────────────────────────────────────

static __device__ __noinline__ void fp_mul(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	const uint32_t *a32 = (const uint32_t *)a;
	const uint32_t *b32 = (const uint32_t *)b;
	const uint32_t *MOD = (const uint32_t *)FP_MODULUS;

	__align__(8) uint32_t even[12];
	__align__(8) uint32_t odd[12];

	#pragma unroll
	for (int i = 0; i < 12; i += 2) {
		fp32_mad_n_redc(even, odd, a32, b32[i], MOD, i == 0);
		fp32_mad_n_redc(odd, even, a32, b32[i + 1], MOD, false);
	}

	// Merge even and odd arrays
	even[0]  = ptx_add_cc(even[0], odd[1]);
	even[1]  = ptx_addc_cc(even[1], odd[2]);
	even[2]  = ptx_addc_cc(even[2], odd[3]);
	even[3]  = ptx_addc_cc(even[3], odd[4]);
	even[4]  = ptx_addc_cc(even[4], odd[5]);
	even[5]  = ptx_addc_cc(even[5], odd[6]);
	even[6]  = ptx_addc_cc(even[6], odd[7]);
	even[7]  = ptx_addc_cc(even[7], odd[8]);
	even[8]  = ptx_addc_cc(even[8], odd[9]);
	even[9]  = ptx_addc_cc(even[9], odd[10]);
	even[10] = ptx_addc_cc(even[10], odd[11]);
	even[11] = ptx_addc(even[11], 0);

	// Final reduction: branchless conditional subtract p
	const uint64_t *e64 = (const uint64_t *)even;
	const uint64_t *q = FP_MODULUS;
	uint64_t s[6], borrow;

	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(s[0]) : "l"(e64[0]), "l"(q[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[1]) : "l"(e64[1]), "l"(q[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[2]) : "l"(e64[2]), "l"(q[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[3]) : "l"(e64[3]), "l"(q[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[4]) : "l"(e64[4]), "l"(q[4]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(s[5]) : "l"(e64[5]), "l"(q[5]));
	asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

	uint64_t mask = -(borrow != 0);
	r[0] = (e64[0] & mask) | (s[0] & ~mask);
	r[1] = (e64[1] & mask) | (s[1] & ~mask);
	r[2] = (e64[2] & mask) | (s[2] & ~mask);
	r[3] = (e64[3] & mask) | (s[3] & ~mask);
	r[4] = (e64[4] & mask) | (s[4] & ~mask);
	r[5] = (e64[5] & mask) | (s[5] & ~mask);
}

// fp_mul_nr: result in [0, 2p), no final reduction
static __device__ __noinline__ void fp_mul_nr(uint64_t r[6], const uint64_t a[6], const uint64_t b[6]) {
	const uint32_t *a32 = (const uint32_t *)a;
	const uint32_t *b32 = (const uint32_t *)b;
	const uint32_t *MOD = (const uint32_t *)FP_MODULUS;

	__align__(8) uint32_t even[12];
	__align__(8) uint32_t odd[12];

	#pragma unroll
	for (int i = 0; i < 12; i += 2) {
		fp32_mad_n_redc(even, odd, a32, b32[i], MOD, i == 0);
		fp32_mad_n_redc(odd, even, a32, b32[i + 1], MOD, false);
	}

	// Merge even and odd arrays
	even[0]  = ptx_add_cc(even[0], odd[1]);
	even[1]  = ptx_addc_cc(even[1], odd[2]);
	even[2]  = ptx_addc_cc(even[2], odd[3]);
	even[3]  = ptx_addc_cc(even[3], odd[4]);
	even[4]  = ptx_addc_cc(even[4], odd[5]);
	even[5]  = ptx_addc_cc(even[5], odd[6]);
	even[6]  = ptx_addc_cc(even[6], odd[7]);
	even[7]  = ptx_addc_cc(even[7], odd[8]);
	even[8]  = ptx_addc_cc(even[8], odd[9]);
	even[9]  = ptx_addc_cc(even[9], odd[10]);
	even[10] = ptx_addc_cc(even[10], odd[11]);
	even[11] = ptx_addc(even[11], 0);

	// No final reduction — copy to output
	uint32_t *r32 = (uint32_t *)r;
	#pragma unroll
	for (int i = 0; i < 12; i++) r32[i] = even[i];
}

// fp_sqr: r = a^2 mod p (uses mul for simplicity; same correctness)
__device__ __forceinline__ void fp_sqr(uint64_t r[6], const uint64_t a[6]) {
	fp_mul(r, a, a);
}

// fp_is_zero: check if all limbs are zero
__device__ __forceinline__ bool fp_is_zero(const uint64_t a[6]) {
	return (a[0] | a[1] | a[2] | a[3] | a[4] | a[5]) == 0;
}

// fp_eq: check if a == b
__device__ __forceinline__ bool fp_eq(const uint64_t a[6], const uint64_t b[6]) {
	return ((a[0] ^ b[0]) | (a[1] ^ b[1]) | (a[2] ^ b[2]) |
	        (a[3] ^ b[3]) | (a[4] ^ b[4]) | (a[5] ^ b[5])) == 0;
}

// fp_copy: r = a
__device__ __forceinline__ void fp_copy(uint64_t r[6], const uint64_t a[6]) {
	r[0] = a[0]; r[1] = a[1]; r[2] = a[2];
	r[3] = a[3]; r[4] = a[4]; r[5] = a[5];
}

// fp_set_zero: r = 0
__device__ __forceinline__ void fp_set_zero(uint64_t r[6]) {
	r[0] = 0; r[1] = 0; r[2] = 0;
	r[3] = 0; r[4] = 0; r[5] = 0;
}

// fp_set_one: r = R (Montgomery form of 1)
__device__ __forceinline__ void fp_set_one(uint64_t r[6]) {
	r[0] = FP_R[0]; r[1] = FP_R[1]; r[2] = FP_R[2];
	r[3] = FP_R[3]; r[4] = FP_R[4]; r[5] = FP_R[5];
}

// fp_conditional_copy: r = condition ? src : r (branchless)
__device__ __forceinline__ void fp_ccopy(uint64_t r[6], const uint64_t src[6], bool condition) {
	uint64_t mask = -(uint64_t)condition; // all 1s if true, all 0s if false
	r[0] = (src[0] & mask) | (r[0] & ~mask);
	r[1] = (src[1] & mask) | (r[1] & ~mask);
	r[2] = (src[2] & mask) | (r[2] & ~mask);
	r[3] = (src[3] & mask) | (r[3] & ~mask);
	r[4] = (src[4] & mask) | (r[4] & ~mask);
	r[5] = (src[5] & mask) | (r[5] & ~mask);
}

// fp_inv: r = a^(p-2) mod p (Fermat's little theorem inversion).
//
// Uses square-and-multiply over the 377-bit exponent p-2. Cost on the order of
// 377 fp_sqr + popcount(p-2) fp_mul ≈ 565 fp_mul calls (one inversion takes
// roughly the cost of 565 multiplications).
//
// Use sparingly: in batched-inversion contexts call this only on the single
// global product. For block-local batched invert we still pay this cost once
// per block per wave; the fp_mul calls in the prefix-product/back-scan
// dominate over many waves.
__device__ __noinline__ void fp_inv(uint64_t r[6], const uint64_t a[6]) {
	// p - 2, little-endian limbs (BLS12-377 base field).
	// p[0] = 0x8508c00000000001 → p[0]-2 = 0x8508bfffffffffff (no borrow propagation).
	static constexpr uint64_t P_MINUS_2[6] = {
		0x8508bfffffffffffULL, 0x170b5d4430000000ULL,
		0x1ef3622fba094800ULL, 0x1a22d9f300f5138fULL,
		0xc63b05c06ca1493bULL, 0x01ae3a4617c510eaULL,
	};

	uint64_t result[6];
	fp_set_one(result);

	// Square-and-multiply, MSB first. The leading zeros above bit 376 cost a few
	// no-op squarings of 1 (negligible).
	#pragma unroll 1
	for(int limb = 5; limb >= 0; limb--) {
		uint64_t e = P_MINUS_2[limb];
		#pragma unroll 1
		for(int bit = 63; bit >= 0; bit--) {
			uint64_t sq[6];
			fp_sqr(sq, result);
			fp_copy(result, sq);
			if((e >> bit) & 1ULL) {
				uint64_t prod[6];
				fp_mul(prod, result, a);
				fp_copy(result, prod);
			}
		}
	}
	fp_copy(r, result);
}

// fp_negate: r = -a mod p (branchless: 0 stays 0)
__device__ __forceinline__ void fp_negate(uint64_t r[6], const uint64_t a[6]) {
	const uint64_t *p = FP_MODULUS;
	bool is_zero = fp_is_zero(a);

	// Compute p - a
	uint64_t t[6];
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(t[0]) : "l"(p[0]), "l"(a[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[1]) : "l"(p[1]), "l"(a[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[2]) : "l"(p[2]), "l"(a[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[3]) : "l"(p[3]), "l"(a[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t[4]) : "l"(p[4]), "l"(a[4]));
	asm volatile("subc.u64 %0, %1, %2;" : "=l"(t[5]) : "l"(p[5]), "l"(a[5]));

	// If a was zero, result should be zero (not p)
	uint64_t zero_mask = -(uint64_t)is_zero;
	r[0] = t[0] & ~zero_mask;
	r[1] = t[1] & ~zero_mask;
	r[2] = t[2] & ~zero_mask;
	r[3] = t[3] & ~zero_mask;
	r[4] = t[4] & ~zero_mask;
	r[5] = t[5] & ~zero_mask;
}

} // namespace gnark_gpu
