// KoalaBear field arithmetic — GPU device functions
//
// Field: P = 2³¹ − 2²⁴ + 1 = 0x7f000001 (31-bit prime)
// Montgomery form: R = 2³², elements stored as uint32 in [0, P)
//
// Extension tower:
//   E2 = KB[u] / (u² − 3)       →  u² = 3
//   E4 = E2[v] / (v² − u)       →  v² = u
//   E4 element = (a₀, a₁, a₂, a₃) representing a₀ + a₁u + a₂v + a₃uv

#pragma once
#include <cstdint>

// ─── Field constants ────────────────────────────────────────────────────────

static constexpr uint32_t KB_P       = 0x7f000001u;   // prime
static constexpr uint32_t KB_MU      = 0x7effffffu;   // −P⁻¹ mod 2³²
static constexpr uint32_t KB_R2      = 402124772u;     // R² mod P  (for toMont)
static constexpr uint32_t KB_ONE     = 33554430u;      // 1 in Montgomery = 2³² mod P
static constexpr uint32_t KB_THREE_M = 100663290u;     // 3 in Montgomery form
// Note: KB_THREE_M = kb_mul_host(KB_ONE, 3_mont). Verified: 3·R mod P = 100663290.

// ─── Base field: Montgomery u32 ─────────────────────────────────────────────

// Montgomery reduction: x·R⁻¹ mod P
__device__ __forceinline__ uint32_t kb_reduce(uint64_t x) {
    uint32_t q = (uint32_t)x * KB_MU;
    uint32_t r = (uint32_t)((x + (uint64_t)q * KB_P) >> 32);
    return r >= KB_P ? r - KB_P : r;
}

__device__ __forceinline__ uint32_t kb_add(uint32_t a, uint32_t b) {
    uint32_t t = a + b;
    return t >= KB_P ? t - KB_P : t;
}

__device__ __forceinline__ uint32_t kb_sub(uint32_t a, uint32_t b) {
    return a >= b ? a - b : a + KB_P - b;
}

__device__ __forceinline__ uint32_t kb_mul(uint32_t a, uint32_t b) {
    return kb_reduce((uint64_t)a * b);
}

__device__ __forceinline__ uint32_t kb_sqr(uint32_t a) {
    return kb_reduce((uint64_t)a * a);
}

__device__ __forceinline__ uint32_t kb_neg(uint32_t a) {
    return a == 0 ? 0 : KB_P - a;
}

__device__ __forceinline__ uint32_t kb_dbl(uint32_t a) {
    return kb_add(a, a);
}

// v[i] = gⁱ·v[i]  — used for coset shift
__device__ __forceinline__ uint32_t kb_mul3(uint32_t a) {
    return kb_mul(a, KB_THREE_M);
}

// ─── E2 = KB[u] / (u² − 3) ─────────────────────────────────────────────────
//
// Element (a₀, a₁) represents a₀ + a₁·u where u² = 3.
//
// Multiplication: (a₀+a₁u)(b₀+b₁u) = (a₀b₀ + 3·a₁b₁) + (a₀b₁ + a₁b₀)u
//   Karatsuba: k = (a₀+a₁)(b₀+b₁), d₀ = a₀b₀, d₁ = a₁b₁
//   → c₀ = d₀ + 3·d₁,  c₁ = k − d₀ − d₁

struct E2 { uint32_t a0, a1; };

__device__ __forceinline__ E2 e2_add(E2 a, E2 b) {
    return {kb_add(a.a0, b.a0), kb_add(a.a1, b.a1)};
}

__device__ __forceinline__ E2 e2_sub(E2 a, E2 b) {
    return {kb_sub(a.a0, b.a0), kb_sub(a.a1, b.a1)};
}

__device__ __forceinline__ E2 e2_mul(E2 a, E2 b) {
    uint32_t d0 = kb_mul(a.a0, b.a0);
    uint32_t d1 = kb_mul(a.a1, b.a1);
    uint32_t k  = kb_mul(kb_add(a.a0, a.a1), kb_add(b.a0, b.a1));
    return {kb_add(d0, kb_mul3(d1)), kb_sub(k, kb_add(d0, d1))};
}

// Multiply by non-residue u: (a₀+a₁u)·u = 3a₁ + a₀u
__device__ __forceinline__ E2 e2_mul_nr(E2 a) {
    return {kb_mul3(a.a1), a.a0};
}

__device__ __forceinline__ E2 e2_neg(E2 a) {
    return {kb_neg(a.a0), kb_neg(a.a1)};
}

__device__ __forceinline__ E2 e2_sqr(E2 a) {
    return e2_mul(a, a);  // could optimize but clarity > 2 instructions
}

// Multiply E2 by base field scalar
__device__ __forceinline__ E2 e2_scale(E2 a, uint32_t s) {
    return {kb_mul(a.a0, s), kb_mul(a.a1, s)};
}

// ─── E4 = E2[v] / (v² − u) ─────────────────────────────────────────────────
//
// Element (b₀, b₁) represents b₀ + b₁·v where v² = u, and bᵢ ∈ E2.
// Flat layout: (a₀, a₁, a₂, a₃) = (b₀.a0, b₀.a1, b₁.a0, b₁.a1)
//
// Multiplication: (b₀+b₁v)(c₀+c₁v) = (b₀c₀ + b₁c₁·u) + (b₀c₁+b₁c₀)v
//   Karatsuba: k = (b₀+b₁)(c₀+c₁), d₀ = b₀c₀, d₁ = b₁c₁
//   → r₀ = d₀ + mulNR(d₁),  r₁ = k − d₀ − d₁

struct E4 { E2 b0, b1; };

__device__ __forceinline__ E4 e4_add(E4 a, E4 b) {
    return {e2_add(a.b0, b.b0), e2_add(a.b1, b.b1)};
}

__device__ __forceinline__ E4 e4_sub(E4 a, E4 b) {
    return {e2_sub(a.b0, b.b0), e2_sub(a.b1, b.b1)};
}

__device__ __forceinline__ E4 e4_mul(E4 a, E4 b) {
    E2 d0 = e2_mul(a.b0, b.b0);
    E2 d1 = e2_mul(a.b1, b.b1);
    E2 k  = e2_mul(e2_add(a.b0, a.b1), e2_add(b.b0, b.b1));
    return {e2_add(d0, e2_mul_nr(d1)), e2_sub(k, e2_add(d0, d1))};
}

// Multiply E4 by base field scalar:  s · (a₀+a₁u+a₂v+a₃uv) = (sa₀+sa₁u+sa₂v+sa₃uv)
__device__ __forceinline__ E4 e4_scale(E4 a, uint32_t s) {
    return {e2_scale(a.b0, s), e2_scale(a.b1, s)};
}

// Accumulate: dst += scalar · e4  (used in linear combination)
__device__ __forceinline__ void e4_mulacc(E4& acc, uint32_t s, E4 alpha_pow) {
    // acc += s * alpha_pow  (scalar from base field, alpha_pow ∈ E4)
    acc = e4_add(acc, e4_scale(alpha_pow, s));
}

static __device__ __forceinline__ E4 e4_zero() {
    return {{0, 0}, {0, 0}};
}

// ─── Montgomery conversion ──────────────────────────────────────────────────
// from_mont: stored (a·R mod P) → canonical a
// to_mont:   canonical a → stored (a·R mod P)

__device__ __forceinline__ uint32_t kb_from_mont(uint32_t a) {
    return kb_reduce((uint64_t)a);
}

__device__ __forceinline__ uint32_t kb_to_mont(uint32_t a) {
    return kb_reduce((uint64_t)a * KB_R2);
}

// ─── E4 extended helpers ────────────────────────────────────────────────────

__device__ __forceinline__ E4 e4_one() {
    return {{KB_ONE, 0}, {0, 0}};
}

__device__ __forceinline__ E4 e4_neg(E4 a) {
    return {e2_neg(a.b0), e2_neg(a.b1)};
}

__device__ __forceinline__ E4 e4_sqr(E4 a) {
    return e4_mul(a, a);
}

// Square-and-multiply: base^exp.  Fast paths for exp ∈ {0,1,2}.
__device__ __forceinline__ E4 e4_pow(E4 base, uint32_t exp) {
    if (exp == 0) return e4_one();
    if (exp == 1) return base;
    if (exp == 2) return e4_sqr(base);
    E4 r = e4_one();
    while (exp > 0) {
        if (exp & 1) r = e4_mul(r, base);
        base = e4_sqr(base);
        exp >>= 1;
    }
    return r;
}

// Scale E4 by small signed integer coefficient.
// Fast paths: c ∈ {0, ±1, 2}; general case via Montgomery.
__device__ __forceinline__ E4 e4_scale_signed(E4 a, int32_t c) {
    if (c ==  1) return a;
    if (c == -1) return e4_neg(a);
    if (c ==  0) return e4_zero();
    if (c ==  2) return e4_add(a, a);
    uint32_t abs_c = kb_to_mont((uint32_t)(c < 0 ? -c : c));
    E4 r = e4_scale(a, abs_c);
    return c < 0 ? e4_neg(r) : r;
}

// Embed KB base field scalar into E4: val ↦ (val, 0, 0, 0)
__device__ __forceinline__ E4 e4_from_kb(uint32_t val) {
    return {{val, 0}, {0, 0}};
}
