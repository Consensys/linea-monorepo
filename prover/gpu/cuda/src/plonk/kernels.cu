// =============================================================================
// Core Fr element-wise kernels (SoA backend)
//
// This file provides simple, high-throughput primitives used across the stack:
//   - Montgomery multiply: c[i] = a[i] * b[i] mod r
//   - Add/sub:             c[i] = a[i] ± b[i] mod r
//   - Layout transpose:    AoS <-> SoA
//
// Data layouts:
//
//   AoS (host / gnark-crypto):
//     [e0.l0 e0.l1 e0.l2 e0.l3 | e1.l0 e1.l1 ...]
//
//   SoA (device):
//     limb0: [e0.l0 e1.l0 e2.l0 ...]
//     limb1: [e0.l1 e1.l1 e2.l1 ...]
//     limb2: [e0.l2 e1.l2 e2.l2 ...]
//     limb3: [e0.l3 e1.l3 e2.l3 ...]
//
// Why SoA:
//   Warps operating on consecutive elements read contiguous limb arrays, giving
//   coalesced global memory access.
// =============================================================================

#include "field.cuh"
#include <cuda_runtime.h>

namespace gnark_gpu {

// =============================================================================
// Helper: 64-bit multiply giving 128-bit result
// =============================================================================

__device__ __forceinline__ void mul_wide(uint64_t a, uint64_t b, uint64_t &lo, uint64_t &hi) {
    lo = a * b;
    hi = __umul64hi(a, b);
}

// =============================================================================
// Helper: Add with carry (a + b + carry_in) -> (result, carry_out)
// =============================================================================

__device__ __forceinline__ uint64_t add_with_carry(uint64_t a, uint64_t b, uint64_t carry_in,
                                                   uint64_t &carry_out) {
    uint64_t sum = a + b;
    carry_out = (sum < a) ? 1ULL : 0ULL;
    uint64_t sum2 = sum + carry_in;
    carry_out += (sum2 < sum) ? 1ULL : 0ULL;
    return sum2;
}

// =============================================================================
// CIOS Montgomery Multiplication kernel for Fr (4 limbs)
// Reference: Algorithm 2 from "Montgomery Multiplication on Modern Processors"
// =============================================================================

__global__ void mul_mont_fr_kernel(const uint64_t *__restrict__ a0,
                                   const uint64_t *__restrict__ a1,
                                   const uint64_t *__restrict__ a2,
                                   const uint64_t *__restrict__ a3,
                                   const uint64_t *__restrict__ b0,
                                   const uint64_t *__restrict__ b1,
                                   const uint64_t *__restrict__ b2,
                                   const uint64_t *__restrict__ b3,
                                   uint64_t *__restrict__ c0, uint64_t *__restrict__ c1,
                                   uint64_t *__restrict__ c2, uint64_t *__restrict__ c3,
                                   size_t n) {
    auto idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n)
        return;

    // BLS12-377 Fr modulus
    constexpr uint64_t q[4] = {
        Fr_params::MODULUS[0], Fr_params::MODULUS[1],
        Fr_params::MODULUS[2], Fr_params::MODULUS[3]
    };
    // Montgomery constant: qInvNeg = -q^(-1) mod 2^64
    constexpr uint64_t qInvNeg = Fr_params::INV;

    // Load operands
    uint64_t a[4] = {__ldg(&a0[idx]), __ldg(&a1[idx]), __ldg(&a2[idx]), __ldg(&a3[idx])};
    uint64_t b[4] = {__ldg(&b0[idx]), __ldg(&b1[idx]), __ldg(&b2[idx]), __ldg(&b3[idx])};

    // Working registers T[0..4] - one extra for overflow
    uint64_t t[5] = {0, 0, 0, 0, 0};

    // CIOS: 4 iterations, one per limb of a
    for (int i = 0; i < 4; i++) {
        uint64_t carry = 0;
        uint64_t lo, hi;

        // Step 1: t = t + a[i] * b
        for (int j = 0; j < 4; j++) {
            mul_wide(a[i], b[j], lo, hi);
            // t[j] = t[j] + lo + carry
            uint64_t tmp = t[j] + lo;
            uint64_t c1 = (tmp < t[j]) ? 1ULL : 0ULL;
            uint64_t tmp2 = tmp + carry;
            uint64_t c2 = (tmp2 < tmp) ? 1ULL : 0ULL;
            t[j] = tmp2;
            carry = hi + c1 + c2;
        }
        t[4] += carry;

        // Step 2: m = t[0] * qInvNeg mod 2^64
        uint64_t m = t[0] * qInvNeg;

        // Step 3: t = (t + m * q) / 2^64
        carry = 0;
        for (int j = 0; j < 4; j++) {
            mul_wide(m, q[j], lo, hi);
            uint64_t tmp = t[j] + lo;
            uint64_t c1 = (tmp < t[j]) ? 1ULL : 0ULL;
            uint64_t tmp2 = tmp + carry;
            uint64_t c2 = (tmp2 < tmp) ? 1ULL : 0ULL;
            if (j > 0) {
                t[j - 1] = tmp2;
            }
            carry = hi + c1 + c2;
        }
        t[3] = t[4] + carry;
        t[4] = 0;
    }

    // Final reduction: if t >= q, then t = t - q
    uint64_t borrow = 0;
    uint64_t r[4];

    // Subtract q from t
    for (int j = 0; j < 4; j++) {
        uint64_t diff = t[j] - q[j] - borrow;
        borrow = (t[j] < q[j] + borrow) ? 1ULL : ((t[j] == q[j] && borrow) ? 1ULL : 0ULL);
        // More accurate borrow calculation
        if (t[j] < q[j]) {
            borrow = 1;
        } else if (t[j] == q[j]) {
            // borrow stays the same
        } else {
            borrow = 0;
        }
        r[j] = diff;
    }

    // If no borrow, t >= q, use reduced value
    // If borrow, t < q, use original value
    if (borrow) {
        // t < q, use t
        c0[idx] = t[0];
        c1[idx] = t[1];
        c2[idx] = t[2];
        c3[idx] = t[3];
    } else {
        // t >= q, use r = t - q
        c0[idx] = r[0];
        c1[idx] = r[1];
        c2[idx] = r[2];
        c3[idx] = r[3];
    }
}

// =============================================================================
// Addition kernel for Fr (4 limbs) with modular reduction
// result = (a + b) mod p
// =============================================================================

__global__ void add_fr_kernel(const uint64_t *__restrict__ a0,
                              const uint64_t *__restrict__ a1,
                              const uint64_t *__restrict__ a2,
                              const uint64_t *__restrict__ a3,
                              const uint64_t *__restrict__ b0,
                              const uint64_t *__restrict__ b1,
                              const uint64_t *__restrict__ b2,
                              const uint64_t *__restrict__ b3, uint64_t *__restrict__ c0,
                              uint64_t *__restrict__ c1, uint64_t *__restrict__ c2,
                              uint64_t *__restrict__ c3, size_t n) {
    auto idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n)
        return;

    constexpr uint64_t p0 = Fr_params::MODULUS[0], p1 = Fr_params::MODULUS[1];
    constexpr uint64_t p2 = Fr_params::MODULUS[2], p3 = Fr_params::MODULUS[3];

    // Load operands
    uint64_t A0 = __ldg(&a0[idx]), A1 = __ldg(&a1[idx]);
    uint64_t A2 = __ldg(&a2[idx]), A3 = __ldg(&a3[idx]);
    uint64_t B0 = __ldg(&b0[idx]), B1 = __ldg(&b1[idx]);
    uint64_t B2 = __ldg(&b2[idx]), B3 = __ldg(&b3[idx]);

    // Add with carry chain using PTX
    uint64_t r0, r1, r2, r3, carry;

    asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r0) : "l"(A0), "l"(B0));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r1) : "l"(A1), "l"(B1));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r2) : "l"(A2), "l"(B2));
    asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r3) : "l"(A3), "l"(B3));
    asm volatile("addc.u64 %0, 0, 0;" : "=l"(carry));

    // Subtract modulus to check if reduction needed
    uint64_t t0, t1, t2, t3, borrow;

    asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(t0) : "l"(r0), "l"(p0));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t1) : "l"(r1), "l"(p1));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t2) : "l"(r2), "l"(p2));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(t3) : "l"(r3), "l"(p3));
    asm volatile("subc.u64 %0, %1, 0;" : "=l"(borrow) : "l"(carry));

    // If no borrow (borrow == 0), use reduced value; else use original
    // borrow == 0 means r >= p, so we should use t (the reduced value)
    bool use_reduced = (borrow == 0);
    c0[idx] = use_reduced ? t0 : r0;
    c1[idx] = use_reduced ? t1 : r1;
    c2[idx] = use_reduced ? t2 : r2;
    c3[idx] = use_reduced ? t3 : r3;
}

// =============================================================================
// Subtraction kernel for Fr (4 limbs) with modular reduction
// result = (a - b) mod p
// =============================================================================

__global__ void sub_fr_kernel(const uint64_t *__restrict__ a0,
                              const uint64_t *__restrict__ a1,
                              const uint64_t *__restrict__ a2,
                              const uint64_t *__restrict__ a3,
                              const uint64_t *__restrict__ b0,
                              const uint64_t *__restrict__ b1,
                              const uint64_t *__restrict__ b2,
                              const uint64_t *__restrict__ b3, uint64_t *__restrict__ c0,
                              uint64_t *__restrict__ c1, uint64_t *__restrict__ c2,
                              uint64_t *__restrict__ c3, size_t n) {
    auto idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n)
        return;

    constexpr uint64_t p0 = Fr_params::MODULUS[0], p1 = Fr_params::MODULUS[1];
    constexpr uint64_t p2 = Fr_params::MODULUS[2], p3 = Fr_params::MODULUS[3];

    // Load operands
    uint64_t A0 = __ldg(&a0[idx]), A1 = __ldg(&a1[idx]);
    uint64_t A2 = __ldg(&a2[idx]), A3 = __ldg(&a3[idx]);
    uint64_t B0 = __ldg(&b0[idx]), B1 = __ldg(&b1[idx]);
    uint64_t B2 = __ldg(&b2[idx]), B3 = __ldg(&b3[idx]);

    // Subtract with borrow chain using PTX
    uint64_t r0, r1, r2, r3, borrow;

    asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(r0) : "l"(A0), "l"(B0));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(r1) : "l"(A1), "l"(B1));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(r2) : "l"(A2), "l"(B2));
    asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(r3) : "l"(A3), "l"(B3));
    asm volatile("subc.u64 %0, 0, 0;" : "=l"(borrow));

    // If borrow occurred (a < b), add modulus back
    // borrow will be 0xFFFFFFFFFFFFFFFF if underflow occurred
    if (borrow != 0) {
        asm volatile("add.cc.u64 %0, %1, %2;" : "=l"(r0) : "l"(r0), "l"(p0));
        asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r1) : "l"(r1), "l"(p1));
        asm volatile("addc.cc.u64 %0, %1, %2;" : "=l"(r2) : "l"(r2), "l"(p2));
        asm volatile("addc.u64 %0, %1, %2;" : "=l"(r3) : "l"(r3), "l"(p3));
    }

    c0[idx] = r0;
    c1[idx] = r1;
    c2[idx] = r2;
    c3[idx] = r3;
}

// =============================================================================
// AoS → SoA transpose kernel for Fr
// Input:  AoS format [e0.l0, e0.l1, e0.l2, e0.l3, e1.l0, e1.l1, ...]
// Output: SoA format  limb0[e0.l0, e1.l0, ...], limb1[e0.l1, e1.l1, ...], ...
// =============================================================================

__global__ void transpose_aos_to_soa_fr_kernel(uint64_t *__restrict__ limb0,
                                               uint64_t *__restrict__ limb1,
                                               uint64_t *__restrict__ limb2,
                                               uint64_t *__restrict__ limb3,
                                               const uint64_t *__restrict__ aos_data,
                                               size_t count) {
    size_t idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= count)
        return;

    const uint64_t *elem = aos_data + idx * 4;
    limb0[idx] = elem[0];
    limb1[idx] = elem[1];
    limb2[idx] = elem[2];
    limb3[idx] = elem[3];
}

// =============================================================================
// SoA → AoS transpose kernel for Fr
// Input:  SoA format  limb0[e0.l0, e1.l0, ...], limb1[e0.l1, e1.l1, ...], ...
// Output: AoS format [e0.l0, e0.l1, e0.l2, e0.l3, e1.l0, e1.l1, ...]
// =============================================================================

__global__ void transpose_soa_to_aos_fr_kernel(uint64_t *__restrict__ aos_data,
                                               const uint64_t *__restrict__ limb0,
                                               const uint64_t *__restrict__ limb1,
                                               const uint64_t *__restrict__ limb2,
                                               const uint64_t *__restrict__ limb3,
                                               size_t count) {
    size_t idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= count)
        return;

    uint64_t *elem = aos_data + idx * 4;
    elem[0] = limb0[idx];
    elem[1] = limb1[idx];
    elem[2] = limb2[idx];
    elem[3] = limb3[idx];
}

// =============================================================================
// Kernel launchers
// =============================================================================

void launch_mul_mont_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                        const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                        const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                        const uint64_t *b2, const uint64_t *b3, size_t n,
                        cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    mul_mont_fr_kernel<<<blocks, threads, 0, stream>>>(a0, a1, a2, a3, b0, b1, b2, b3, c0,
                                                       c1, c2, c3, n);
}

void launch_add_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                   const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                   const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                   const uint64_t *b2, const uint64_t *b3, size_t n,
                   cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    add_fr_kernel<<<blocks, threads, 0, stream>>>(a0, a1, a2, a3, b0, b1, b2, b3, c0, c1,
                                                  c2, c3, n);
}

void launch_sub_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                   const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                   const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                   const uint64_t *b2, const uint64_t *b3, size_t n,
                   cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    sub_fr_kernel<<<blocks, threads, 0, stream>>>(a0, a1, a2, a3, b0, b1, b2, b3, c0, c1,
                                                  c2, c3, n);
}

void launch_transpose_aos_to_soa_fr(uint64_t *limb0, uint64_t *limb1, uint64_t *limb2,
                                    uint64_t *limb3, const uint64_t *aos_data, size_t count,
                                    cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (count + threads - 1) / threads;
    transpose_aos_to_soa_fr_kernel<<<blocks, threads, 0, stream>>>(limb0, limb1, limb2,
                                                                   limb3, aos_data, count);
}

void launch_transpose_soa_to_aos_fr(uint64_t *aos_data, const uint64_t *limb0,
                                    const uint64_t *limb1, const uint64_t *limb2,
                                    const uint64_t *limb3, size_t count,
                                    cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (count + threads - 1) / threads;
    transpose_soa_to_aos_fr_kernel<<<blocks, threads, 0, stream>>>(aos_data, limb0, limb1,
                                                                   limb2, limb3, count);
}

} // namespace gnark_gpu
