// =============================================================================
// Extra Fr kernels used by PlonK prover hot paths.
//
// Implemented ops:
//   - ScaleByPowers:    v[i] *= g^i
//   - ScalarMul:        v[i] *= c
//   - AddMul:           v[i] += a[i] * b[i]
//   - AddScalarMul:     v[i] += a[i] * scalar
//   - BatchInvert:      v[i] <- 1/v[i] (parallel prefix-product method)
//
// Batch inversion strategy:
//
//   Given x[0..n-1], compute inv(x[i]) with 1 inversion + O(n) muls.
//
//   prefix[i] = x[0] * ... * x[i]
//   inv_all   = 1 / prefix[n-1]
//   backward pass recovers each inv(x[i]).
//
//   GPU implementation is chunked to keep work parallel and memory-friendly.
// =============================================================================

#include "fr_arith.cuh"
#include <cuda_runtime.h>

namespace gnark_gpu {

// =============================================================================
// ScaleByPowers: v[i] *= g^i
// Each block computes a chunk of consecutive elements.
// Thread 0 computes g^block_start and a small table {g^(2^k)} in shared memory.
// Threads reconstruct g^threadIdx from that table, then process several
// coalesced elements separated by blockDim.x using a g^blockDim stride.
// =============================================================================

__global__ void scale_by_powers_kernel(uint64_t *__restrict__ v0,
                                        uint64_t *__restrict__ v1,
                                        uint64_t *__restrict__ v2,
                                        uint64_t *__restrict__ v3,
                                        const uint64_t g0, const uint64_t g1,
                                        const uint64_t g2, const uint64_t g3,
                                        size_t n) {
    constexpr unsigned ITEMS_PER_THREAD = 4;
    size_t block_start = (size_t)blockIdx.x * blockDim.x * ITEMS_PER_THREAD;
    size_t idx = block_start + threadIdx.x;

    __shared__ uint64_t sh_power[4];       // g^block_start
    __shared__ uint64_t sh_pow2[9][4];     // g^(2^k), k in [0,8]; k=8 is g^256

    if (threadIdx.x == 0) {
        // Precompute g^(2^k) once per block.
        uint64_t pow2[4] = {g0, g1, g2, g3};
        #pragma unroll
        for (int k = 0; k < 9; k++) {
            sh_pow2[k][0] = pow2[0];
            sh_pow2[k][1] = pow2[1];
            sh_pow2[k][2] = pow2[2];
            sh_pow2[k][3] = pow2[3];
            uint64_t sq[4];
            fr_mul(sq, pow2, pow2);
            pow2[0] = sq[0]; pow2[1] = sq[1];
            pow2[2] = sq[2]; pow2[3] = sq[3];
        }

        // Compute g^block_start via repeated squaring.
        uint64_t base[4] = {g0, g1, g2, g3};
        uint64_t result[4] = {
            Fr_params::ONE[0], Fr_params::ONE[1],
            Fr_params::ONE[2], Fr_params::ONE[3]
        };
        size_t exp = block_start;
        while (exp > 0) {
            if (exp & 1) {
                uint64_t tmp[4];
                fr_mul(tmp, result, base);
                result[0] = tmp[0]; result[1] = tmp[1];
                result[2] = tmp[2]; result[3] = tmp[3];
            }
            uint64_t tmp[4];
            fr_mul(tmp, base, base);
            base[0] = tmp[0]; base[1] = tmp[1];
            base[2] = tmp[2]; base[3] = tmp[3];
            exp >>= 1;
        }
        sh_power[0] = result[0]; sh_power[1] = result[1];
        sh_power[2] = result[2]; sh_power[3] = result[3];
    }
    __syncthreads();

    if (idx >= n) return;

    // Reconstruct g^threadIdx from the shared binary-power table.
    uint64_t my_power[4] = {
        Fr_params::ONE[0], Fr_params::ONE[1],
        Fr_params::ONE[2], Fr_params::ONE[3]
    };
    unsigned t = threadIdx.x;
    #pragma unroll
    for (int bit = 0; bit < 8; bit++) {
        if ((t >> bit) & 1u) {
            uint64_t pow2[4] = {
                sh_pow2[bit][0], sh_pow2[bit][1],
                sh_pow2[bit][2], sh_pow2[bit][3],
            };
            uint64_t tmp[4];
            fr_mul(tmp, my_power, pow2);
            my_power[0] = tmp[0]; my_power[1] = tmp[1];
            my_power[2] = tmp[2]; my_power[3] = tmp[3];
        }
    }

    uint64_t power[4];
    uint64_t block_pow[4] = {sh_power[0], sh_power[1], sh_power[2], sh_power[3]};
    fr_mul(power, block_pow, my_power);

    uint64_t stride[4] = {
        sh_pow2[8][0], sh_pow2[8][1], sh_pow2[8][2], sh_pow2[8][3],
    };

    #pragma unroll
    for (unsigned item = 0; item < ITEMS_PER_THREAD; item++) {
        size_t cur = idx + (size_t)item * blockDim.x;
        if (cur < n) {
            uint64_t val[4] = {v0[cur], v1[cur], v2[cur], v3[cur]};
            uint64_t result[4];
            fr_mul(result, val, power);
            v0[cur] = result[0];
            v1[cur] = result[1];
            v2[cur] = result[2];
            v3[cur] = result[3];
        }
        if constexpr (ITEMS_PER_THREAD > 1) {
            if (item + 1 < ITEMS_PER_THREAD) {
                uint64_t next[4];
                fr_mul(next, power, stride);
                power[0] = next[0]; power[1] = next[1];
                power[2] = next[2]; power[3] = next[3];
            }
        }
    }
}

void launch_scale_by_powers(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                             const uint64_t g[4], size_t n, cudaStream_t stream) {
    constexpr unsigned threads = 256;
    constexpr unsigned items_per_thread = 4;
    unsigned blocks = (n + threads * items_per_thread - 1) / (threads * items_per_thread);
    scale_by_powers_kernel<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, g[0], g[1], g[2], g[3], n);
}

// =============================================================================
// ScalarMul: v[i] *= c for all i
// =============================================================================

__global__ void scalar_mul_kernel(uint64_t *__restrict__ v0,
                                   uint64_t *__restrict__ v1,
                                   uint64_t *__restrict__ v2,
                                   uint64_t *__restrict__ v3,
                                   const uint64_t c0, const uint64_t c1,
                                   const uint64_t c2, const uint64_t c3,
                                   size_t n) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n) return;

    uint64_t val[4] = {v0[idx], v1[idx], v2[idx], v3[idx]};
    uint64_t c[4] = {c0, c1, c2, c3};
    uint64_t result[4];
    fr_mul(result, val, c);
    v0[idx] = result[0];
    v1[idx] = result[1];
    v2[idx] = result[2];
    v3[idx] = result[3];
}

void launch_scalar_mul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                        const uint64_t c[4], size_t n, cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    scalar_mul_kernel<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, c[0], c[1], c[2], c[3], n);
}

// =============================================================================
// AddMul: v[i] += a[i] * b[i] (fused multiply-add)
// =============================================================================

__global__ void addmul_kernel(uint64_t *__restrict__ v0,
                               uint64_t *__restrict__ v1,
                               uint64_t *__restrict__ v2,
                               uint64_t *__restrict__ v3,
                               const uint64_t *__restrict__ a0,
                               const uint64_t *__restrict__ a1,
                               const uint64_t *__restrict__ a2,
                               const uint64_t *__restrict__ a3,
                               const uint64_t *__restrict__ b0,
                               const uint64_t *__restrict__ b1,
                               const uint64_t *__restrict__ b2,
                               const uint64_t *__restrict__ b3,
                               size_t n) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n) return;

    uint64_t a[4] = {a0[idx], a1[idx], a2[idx], a3[idx]};
    uint64_t b[4] = {b0[idx], b1[idx], b2[idx], b3[idx]};
    uint64_t prod[4];
    fr_mul(prod, a, b);

    uint64_t v[4] = {v0[idx], v1[idx], v2[idx], v3[idx]};
    uint64_t result[4];
    fr_add(result, v, prod);
    v0[idx] = result[0];
    v1[idx] = result[1];
    v2[idx] = result[2];
    v3[idx] = result[3];
}

void launch_addmul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                    const uint64_t *a0, const uint64_t *a1, const uint64_t *a2, const uint64_t *a3,
                    const uint64_t *b0, const uint64_t *b1, const uint64_t *b2, const uint64_t *b3,
                    size_t n, cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    addmul_kernel<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, a0, a1, a2, a3, b0, b1, b2, b3, n);
}

// =============================================================================
// AddScalarMul: v[i] += a[i] * scalar (broadcast scalar multiply-add)
// =============================================================================

__global__ void add_scalar_mul_kernel(uint64_t *__restrict__ v0,
                                       uint64_t *__restrict__ v1,
                                       uint64_t *__restrict__ v2,
                                       uint64_t *__restrict__ v3,
                                       const uint64_t *__restrict__ a0,
                                       const uint64_t *__restrict__ a1,
                                       const uint64_t *__restrict__ a2,
                                       const uint64_t *__restrict__ a3,
                                       const uint64_t s0, const uint64_t s1,
                                       const uint64_t s2, const uint64_t s3,
                                       size_t n) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (idx >= n) return;

    uint64_t a[4] = {a0[idx], a1[idx], a2[idx], a3[idx]};
    uint64_t s[4] = {s0, s1, s2, s3};
    uint64_t prod[4];
    fr_mul(prod, a, s);

    uint64_t v[4] = {v0[idx], v1[idx], v2[idx], v3[idx]};
    uint64_t result[4];
    fr_add(result, v, prod);
    v0[idx] = result[0];
    v1[idx] = result[1];
    v2[idx] = result[2];
    v3[idx] = result[3];
}

void launch_add_scalar_mul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                            const uint64_t *a0, const uint64_t *a1, const uint64_t *a2, const uint64_t *a3,
                            const uint64_t scalar[4], size_t n, cudaStream_t stream) {
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    add_scalar_mul_kernel<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, a0, a1, a2, a3,
        scalar[0], scalar[1], scalar[2], scalar[3], n);
}

// =============================================================================
// Fused gate constraint accumulation for PlonK quotient computation.
// result[i] = (result[i] + Ql[i]*L[i] + Qr[i]*R[i] + Qm[i]*L[i]*R[i]
//              + Qo[i]*O[i] + Qk[i]) * zhKInv
// Single pass replaces 6 separate kernel launches per coset.
// =============================================================================

__global__ void plonk_gate_accum_kernel(
    uint64_t *__restrict__ res0, uint64_t *__restrict__ res1,
    uint64_t *__restrict__ res2, uint64_t *__restrict__ res3,
    const uint64_t *__restrict__ Ql0, const uint64_t *__restrict__ Ql1,
    const uint64_t *__restrict__ Ql2, const uint64_t *__restrict__ Ql3,
    const uint64_t *__restrict__ Qr0, const uint64_t *__restrict__ Qr1,
    const uint64_t *__restrict__ Qr2, const uint64_t *__restrict__ Qr3,
    const uint64_t *__restrict__ Qm0, const uint64_t *__restrict__ Qm1,
    const uint64_t *__restrict__ Qm2, const uint64_t *__restrict__ Qm3,
    const uint64_t *__restrict__ Qo0, const uint64_t *__restrict__ Qo1,
    const uint64_t *__restrict__ Qo2, const uint64_t *__restrict__ Qo3,
    const uint64_t *__restrict__ Qk0, const uint64_t *__restrict__ Qk1,
    const uint64_t *__restrict__ Qk2, const uint64_t *__restrict__ Qk3,
    const uint64_t *__restrict__ L0, const uint64_t *__restrict__ L1_,
    const uint64_t *__restrict__ L2, const uint64_t *__restrict__ L3,
    const uint64_t *__restrict__ R0, const uint64_t *__restrict__ R1,
    const uint64_t *__restrict__ R2, const uint64_t *__restrict__ R3,
    const uint64_t *__restrict__ O0, const uint64_t *__restrict__ O1,
    const uint64_t *__restrict__ O2, const uint64_t *__restrict__ O3,
    const uint64_t zh0, const uint64_t zh1,
    const uint64_t zh2, const uint64_t zh3,
    size_t n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t zhKInv[4] = {zh0, zh1, zh2, zh3};

    // Load result (perm+boundary already accumulated)
    uint64_t r[4] = {res0[i], res1[i], res2[i], res3[i]};

    // Load wires
    uint64_t l[4] = {L0[i], L1_[i], L2[i], L3[i]};
    uint64_t rv[4] = {R0[i], R1[i], R2[i], R3[i]};
    uint64_t o[4] = {O0[i], O1[i], O2[i], O3[i]};

    // r += Ql * L
    uint64_t ql[4] = {Ql0[i], Ql1[i], Ql2[i], Ql3[i]};
    uint64_t tmp[4];
    fr_mul(tmp, ql, l);
    fr_add(r, r, tmp);

    // r += Qr * R
    uint64_t qr[4] = {Qr0[i], Qr1[i], Qr2[i], Qr3[i]};
    fr_mul(tmp, qr, rv);
    fr_add(r, r, tmp);

    // r += Qm * L * R
    uint64_t qm[4] = {Qm0[i], Qm1[i], Qm2[i], Qm3[i]};
    uint64_t lr[4];
    fr_mul(lr, l, rv);
    fr_mul(tmp, qm, lr);
    fr_add(r, r, tmp);

    // r += Qo * O
    uint64_t qo[4] = {Qo0[i], Qo1[i], Qo2[i], Qo3[i]};
    fr_mul(tmp, qo, o);
    fr_add(r, r, tmp);

    // r += Qk
    uint64_t qk[4] = {Qk0[i], Qk1[i], Qk2[i], Qk3[i]};
    fr_add(r, r, qk);

    // r *= zhKInv
    uint64_t out[4];
    fr_mul(out, r, zhKInv);

    res0[i] = out[0]; res1[i] = out[1]; res2[i] = out[2]; res3[i] = out[3];
}

void launch_plonk_gate_accum(
    uint64_t *res0, uint64_t *res1, uint64_t *res2, uint64_t *res3,
    const uint64_t *Ql0, const uint64_t *Ql1, const uint64_t *Ql2, const uint64_t *Ql3,
    const uint64_t *Qr0, const uint64_t *Qr1, const uint64_t *Qr2, const uint64_t *Qr3,
    const uint64_t *Qm0, const uint64_t *Qm1, const uint64_t *Qm2, const uint64_t *Qm3,
    const uint64_t *Qo0, const uint64_t *Qo1, const uint64_t *Qo2, const uint64_t *Qo3,
    const uint64_t *Qk0, const uint64_t *Qk1, const uint64_t *Qk2, const uint64_t *Qk3,
    const uint64_t *L0, const uint64_t *L1, const uint64_t *L2, const uint64_t *L3,
    const uint64_t *R0, const uint64_t *R1, const uint64_t *R2, const uint64_t *R3,
    const uint64_t *O0, const uint64_t *O1, const uint64_t *O2, const uint64_t *O3,
    const uint64_t zhKInv[4], size_t n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    plonk_gate_accum_kernel<<<blocks, threads, 0, stream>>>(
        res0, res1, res2, res3,
        Ql0, Ql1, Ql2, Ql3,
        Qr0, Qr1, Qr2, Qr3,
        Qm0, Qm1, Qm2, Qm3,
        Qo0, Qo1, Qo2, Qo3,
        Qk0, Qk1, Qk2, Qk3,
        L0, L1, L2, L3,
        R0, R1, R2, R3,
        O0, O1, O2, O3,
        zhKInv[0], zhKInv[1], zhKInv[2], zhKInv[3],
        n);
}

// =============================================================================
// BatchInvert: v[i] = 1/v[i] using Montgomery batch inversion
// Two-level parallel prefix scan:
//   1. Forward: per-chunk prefix products (parallel), then fixup
//   2. Invert total product via Fermat's little theorem
//   3. Backward: per-chunk sweep to recover individual inverses (parallel)
// =============================================================================

// VRAM guardrail: each BatchInvert scratch arena stores bp/sa/tmp for four
// limbs (12 uint64 arrays). The large-vector path keeps two arenas, so retained
// scratch is roughly:
//
//     2 * 12 * 8 * ceil(n / BATCH_INV_CHUNK) bytes
//
// At n=2^27: chunk=256 => 96 MiB, 128 => 192 MiB, 64 => 384 MiB.
// Chunk 64 was faster in isolation but made repeated 2^27 PlonK proofs OOM.
// Do not lower this without adding an explicit scratch release/shrink path and
// validating repeated BenchmarkPlonkECMul750 runs.
constexpr size_t BATCH_INV_CHUNK = 256;

// Device function: field inversion via Fermat's little theorem (a^(q-2) mod q)
__device__ void fr_invert(uint64_t result[4], const uint64_t a[4]) {
    // q-2 for BLS12-377 Fr field
    static constexpr uint64_t EXP[4] = {
        0x0a117fffffffffffULL,
        0x59aa76fed0000001ULL,
        0x60b44d1e5c37b001ULL,
        0x12ab655e9a2ca556ULL,
    };

    uint64_t base[4] = {a[0], a[1], a[2], a[3]};
    result[0] = Fr_params::ONE[0]; result[1] = Fr_params::ONE[1];
    result[2] = Fr_params::ONE[2]; result[3] = Fr_params::ONE[3];

    for (int word = 0; word < 4; word++) {
        uint64_t bits = EXP[word];
        for (int bit = 0; bit < 64; bit++) {
            if (bits & 1) {
                uint64_t tmp[4];
                fr_mul(tmp, result, base);
                result[0] = tmp[0]; result[1] = tmp[1];
                result[2] = tmp[2]; result[3] = tmp[3];
            }
            uint64_t tmp[4];
            fr_mul(tmp, base, base);
            base[0] = tmp[0]; base[1] = tmp[1];
            base[2] = tmp[2]; base[3] = tmp[3];
            bits >>= 1;
        }
    }
}

// Kernel 1: Local prefix products per chunk + store chunk product.
// Each thread handles one chunk of BATCH_INV_CHUNK elements sequentially.
__global__ void batch_invert_prefix_local(
    uint64_t *__restrict__ v0, uint64_t *__restrict__ v1,
    uint64_t *__restrict__ v2, uint64_t *__restrict__ v3,
    uint64_t *__restrict__ bp0, uint64_t *__restrict__ bp1,
    uint64_t *__restrict__ bp2, uint64_t *__restrict__ bp3,
    size_t n)
{
    size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t num_chunks = (n + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;
    if (chunk_id >= num_chunks) return;

    size_t start = chunk_id * BATCH_INV_CHUNK;
    size_t end = start + BATCH_INV_CHUNK;
    if (end > n) end = n;

    uint64_t acc[4] = {v0[start], v1[start], v2[start], v3[start]};
    for (size_t i = start + 1; i < end; i++) {
        uint64_t elem[4] = {v0[i], v1[i], v2[i], v3[i]};
        uint64_t prod[4];
        fr_mul(prod, acc, elem);
        acc[0] = prod[0]; acc[1] = prod[1];
        acc[2] = prod[2]; acc[3] = prod[3];
        v0[i] = acc[0]; v1[i] = acc[1];
        v2[i] = acc[2]; v3[i] = acc[3];
    }

    bp0[chunk_id] = acc[0]; bp1[chunk_id] = acc[1];
    bp2[chunk_id] = acc[2]; bp3[chunk_id] = acc[3];
}

// Kernel 2: Serial fixup (single thread).
// - Prefix product of block products
// - Invert total
// - Compute backward starting accs (suffix product * inv_total)
__global__ void batch_invert_serial_fixup(
    uint64_t *__restrict__ bp0, uint64_t *__restrict__ bp1,
    uint64_t *__restrict__ bp2, uint64_t *__restrict__ bp3,
    uint64_t *__restrict__ sa0, uint64_t *__restrict__ sa1,
    uint64_t *__restrict__ sa2, uint64_t *__restrict__ sa3,
    size_t num_chunks)
{
    if (threadIdx.x != 0 || blockIdx.x != 0) return;

    // Save original block products to sa[] temporarily
    for (size_t i = 0; i < num_chunks; i++) {
        sa0[i] = bp0[i]; sa1[i] = bp1[i];
        sa2[i] = bp2[i]; sa3[i] = bp3[i];
    }

    // Forward prefix product of block products (in-place in bp[])
    for (size_t i = 1; i < num_chunks; i++) {
        uint64_t prev[4] = {bp0[i-1], bp1[i-1], bp2[i-1], bp3[i-1]};
        uint64_t curr[4] = {bp0[i], bp1[i], bp2[i], bp3[i]};
        uint64_t prod[4];
        fr_mul(prod, prev, curr);
        bp0[i] = prod[0]; bp1[i] = prod[1];
        bp2[i] = prod[2]; bp3[i] = prod[3];
    }

    // Invert total product
    uint64_t total[4] = {
        bp0[num_chunks-1], bp1[num_chunks-1],
        bp2[num_chunks-1], bp3[num_chunks-1]
    };
    uint64_t inv[4];
    fr_invert(inv, total);

    // Compute backward starting accs: start_acc[k] = inv * prod(orig_bp[k+1..last])
    // Process right-to-left, saving next_orig before overwriting.
    uint64_t acc[4] = {inv[0], inv[1], inv[2], inv[3]};
    uint64_t next_orig[4] = {
        sa0[num_chunks-1], sa1[num_chunks-1],
        sa2[num_chunks-1], sa3[num_chunks-1]
    };
    sa0[num_chunks-1] = acc[0]; sa1[num_chunks-1] = acc[1];
    sa2[num_chunks-1] = acc[2]; sa3[num_chunks-1] = acc[3];

    for (int k = (int)num_chunks - 2; k >= 0; k--) {
        uint64_t tmp[4];
        fr_mul(tmp, acc, next_orig);
        acc[0] = tmp[0]; acc[1] = tmp[1];
        acc[2] = tmp[2]; acc[3] = tmp[3];
        next_orig[0] = sa0[k]; next_orig[1] = sa1[k];
        next_orig[2] = sa2[k]; next_orig[3] = sa3[k];
        sa0[k] = acc[0]; sa1[k] = acc[1];
        sa2[k] = acc[2]; sa3[k] = acc[3];
    }
}

// Prefix-only serial fixup (single thread).
// Computes inclusive prefix products of bp[] in-place.
// Used as the second level of a hierarchical scan to avoid long serial loops.
__global__ void batch_prefix_serial_fixup(
    uint64_t *__restrict__ bp0, uint64_t *__restrict__ bp1,
    uint64_t *__restrict__ bp2, uint64_t *__restrict__ bp3,
    size_t num_chunks)
{
    if (threadIdx.x != 0 || blockIdx.x != 0) return;
    for (size_t i = 1; i < num_chunks; i++) {
        uint64_t prev[4] = {bp0[i-1], bp1[i-1], bp2[i-1], bp3[i-1]};
        uint64_t curr[4] = {bp0[i], bp1[i], bp2[i], bp3[i]};
        uint64_t prod[4];
        fr_mul(prod, prev, curr);
        bp0[i] = prod[0]; bp1[i] = prod[1];
        bp2[i] = prod[2]; bp3[i] = prod[3];
    }
}

// Kernel 3: Apply forward fixup — multiply each element in chunk k (for k>=1)
// by bp[k-1] (prefix of all prior block products).
__global__ void batch_invert_apply_fixup(
    uint64_t *__restrict__ v0, uint64_t *__restrict__ v1,
    uint64_t *__restrict__ v2, uint64_t *__restrict__ v3,
    const uint64_t *__restrict__ bp0, const uint64_t *__restrict__ bp1,
    const uint64_t *__restrict__ bp2, const uint64_t *__restrict__ bp3,
    size_t n)
{
    size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t num_chunks = (n + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;
    if (chunk_id == 0 || chunk_id >= num_chunks) return;

    size_t start = chunk_id * BATCH_INV_CHUNK;
    size_t end = start + BATCH_INV_CHUNK;
    if (end > n) end = n;

    uint64_t prefix[4] = {bp0[chunk_id-1], bp1[chunk_id-1],
                           bp2[chunk_id-1], bp3[chunk_id-1]};
    for (size_t i = start; i < end; i++) {
        uint64_t elem[4] = {v0[i], v1[i], v2[i], v3[i]};
        uint64_t prod[4];
        fr_mul(prod, prefix, elem);
        v0[i] = prod[0]; v1[i] = prod[1];
        v2[i] = prod[2]; v3[i] = prod[3];
    }
}

// Kernel 4: Backward sweep — compute inverses in-place in v[].
// v[] holds corrected prefix products (read within own chunk + bp for boundary).
// orig[] holds original values (read-only). bp[] holds prefixed block products.
// Results are written back to v[] (safe: only write to own chunk, and within a
// chunk we process backward so v[i-1] is read before v[i] is overwritten).
// Cross-chunk boundary: v[start-1] is read via bp[chunk_id-1] to avoid races.
__global__ void batch_invert_backward(
    uint64_t *__restrict__ v0, uint64_t *__restrict__ v1,
    uint64_t *__restrict__ v2, uint64_t *__restrict__ v3,
    const uint64_t *__restrict__ orig0, const uint64_t *__restrict__ orig1,
    const uint64_t *__restrict__ orig2, const uint64_t *__restrict__ orig3,
    const uint64_t *__restrict__ bp0, const uint64_t *__restrict__ bp1,
    const uint64_t *__restrict__ bp2, const uint64_t *__restrict__ bp3,
    const uint64_t *__restrict__ sa0, const uint64_t *__restrict__ sa1,
    const uint64_t *__restrict__ sa2, const uint64_t *__restrict__ sa3,
    size_t n)
{
    size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t num_chunks = (n + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;
    if (chunk_id >= num_chunks) return;

    size_t start = chunk_id * BATCH_INV_CHUNK;
    size_t end = start + BATCH_INV_CHUNK;
    if (end > n) end = n;

    uint64_t acc[4] = {sa0[chunk_id], sa1[chunk_id],
                        sa2[chunk_id], sa3[chunk_id]};

    for (size_t i = end; i > start; ) {
        i--;
        uint64_t orig_val[4] = {orig0[i], orig1[i], orig2[i], orig3[i]};
        if (i > 0) {
            uint64_t prefix_prev[4];
            if (i == start && chunk_id > 0) {
                // Cross-chunk boundary: use bp[chunk_id-1] instead of v[start-1]
                // bp[k] = prefix product of elements 0..(k+1)*CHUNK-1 after fixup
                prefix_prev[0] = bp0[chunk_id-1]; prefix_prev[1] = bp1[chunk_id-1];
                prefix_prev[2] = bp2[chunk_id-1]; prefix_prev[3] = bp3[chunk_id-1];
            } else {
                // Within-chunk: safe to read v[i-1] (not yet overwritten)
                prefix_prev[0] = v0[i-1]; prefix_prev[1] = v1[i-1];
                prefix_prev[2] = v2[i-1]; prefix_prev[3] = v3[i-1];
            }
            uint64_t result[4];
            fr_mul(result, acc, prefix_prev);
            v0[i] = result[0]; v1[i] = result[1];
            v2[i] = result[2]; v3[i] = result[3];
        } else {
            // i == 0: result = acc (no prefix before element 0)
            v0[0] = acc[0]; v1[0] = acc[1];
            v2[0] = acc[2]; v3[0] = acc[3];
        }
        uint64_t tmp[4];
        fr_mul(tmp, acc, orig_val);
        acc[0] = tmp[0]; acc[1] = tmp[1];
        acc[2] = tmp[2]; acc[3] = tmp[3];
    }
}

// BatchInvert scratch memory for block-level arrays (bp and sa).
// Pre-allocated to avoid per-call cudaMalloc/cudaFree overhead.
struct BatchInvertScratch {
    uint64_t *bp[4] = {};  // block products (num_chunks per limb)
    uint64_t *sa[4] = {};  // starting accs (num_chunks per limb)
    uint64_t *tmp[4] = {}; // extra workspace (num_chunks per limb)
    size_t capacity = 0;   // number of chunks allocated
};

// Ensure scratch has capacity for at least num_chunks.
static cudaError_t batch_invert_scratch_ensure(BatchInvertScratch &s, size_t num_chunks) {
    if (num_chunks <= s.capacity) return cudaSuccess;

    // Free old
    for (int i = 0; i < 4; i++) {
        if (s.bp[i]) { cudaFree(s.bp[i]); s.bp[i] = nullptr; }
        if (s.sa[i]) { cudaFree(s.sa[i]); s.sa[i] = nullptr; }
        if (s.tmp[i]) { cudaFree(s.tmp[i]); s.tmp[i] = nullptr; }
    }
    s.capacity = 0;

    // Allocate new (with 2x growth factor)
    size_t alloc_chunks = num_chunks < 64 ? 64 : num_chunks;
    cudaError_t err;
    for (int i = 0; i < 4; i++) {
        err = cudaMalloc(&s.bp[i], alloc_chunks * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
        err = cudaMalloc(&s.sa[i], alloc_chunks * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
        err = cudaMalloc(&s.tmp[i], alloc_chunks * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
    }
    s.capacity = alloc_chunks;
    return cudaSuccess;
}

// Per-context scratch (thread-local would be needed for multi-context; for now, stored in context).
// The context struct will hold this.
static thread_local BatchInvertScratch g_batch_inv_scratch;
static thread_local BatchInvertScratch g_batch_inv_aux_scratch;

// Baseline two-level batch inversion implementation using a single scratch arena.
// Good for small/medium vectors where num_chunks is not huge.
static cudaError_t launch_batch_invert_baseline(
    uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
    uint64_t *orig0, uint64_t *orig1, uint64_t *orig2, uint64_t *orig3,
    size_t n, cudaStream_t stream,
    BatchInvertScratch &scratch)
{
    if (n == 0) return cudaSuccess;

    size_t num_chunks = (n + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;

    cudaError_t err = batch_invert_scratch_ensure(scratch, num_chunks);
    if (err != cudaSuccess) return err;

    auto &s = scratch;

    err = cudaMemcpyAsync(orig0, v0, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig1, v1, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig2, v2, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig3, v3, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;

    constexpr unsigned threads = 256;
    unsigned blocks = (num_chunks + threads - 1) / threads;

    batch_invert_prefix_local<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, s.bp[0], s.bp[1], s.bp[2], s.bp[3], n);

    batch_invert_serial_fixup<<<1, 1, 0, stream>>>(
        s.bp[0], s.bp[1], s.bp[2], s.bp[3],
        s.sa[0], s.sa[1], s.sa[2], s.sa[3], num_chunks);

    batch_invert_apply_fixup<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, s.bp[0], s.bp[1], s.bp[2], s.bp[3], n);

    batch_invert_backward<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, orig0, orig1, orig2, orig3,
        s.bp[0], s.bp[1], s.bp[2], s.bp[3],
        s.sa[0], s.sa[1], s.sa[2], s.sa[3], n);

    return cudaSuccess;
}

// Launch function: orchestrates the full batch inversion pipeline.
// v is modified in-place to contain 1/v[i].
// orig[] is a temp buffer (same size as v) for original values.
// Fully async — no internal synchronization.
cudaError_t launch_batch_invert(
    uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
    uint64_t *orig0, uint64_t *orig1, uint64_t *orig2, uint64_t *orig3,
    size_t n, cudaStream_t stream)
{
    if (n == 0) return cudaSuccess;

    size_t num_chunks = (n + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;
    if (num_chunks <= BATCH_INV_CHUNK) {
        return launch_batch_invert_baseline(
            v0, v1, v2, v3,
            orig0, orig1, orig2, orig3,
            n, stream, g_batch_inv_scratch);
    }

    // Hierarchical path for large vectors:
    // 1) parallel local scan on v
    // 2) hierarchical scan of chunk products (avoids long serial prefix loops)
    // 3) build chunk start accumulators from scanned chunk prefixes + original chunk products
    // 4) backward sweep on v
    cudaError_t err = batch_invert_scratch_ensure(g_batch_inv_scratch, num_chunks);
    if (err != cudaSuccess) return err;
    err = batch_invert_scratch_ensure(g_batch_inv_aux_scratch, num_chunks);
    if (err != cudaSuccess) return err;

    auto &s = g_batch_inv_scratch;
    auto &aux = g_batch_inv_aux_scratch;

    err = cudaMemcpyAsync(orig0, v0, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig1, v1, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig2, v2, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(orig3, v3, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;

    constexpr unsigned threads = 256;
    unsigned blocks = (num_chunks + threads - 1) / threads;

    // Step 1: local scan within each primary chunk + primary chunk products in s.bp.
    batch_invert_prefix_local<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, s.bp[0], s.bp[1], s.bp[2], s.bp[3], n);

    // Step 2: hierarchical scan of primary chunk products.
    size_t super_chunks = (num_chunks + BATCH_INV_CHUNK - 1) / BATCH_INV_CHUNK;
    unsigned super_blocks = (super_chunks + threads - 1) / threads;

    batch_invert_prefix_local<<<super_blocks, threads, 0, stream>>>(
        s.bp[0], s.bp[1], s.bp[2], s.bp[3],
        aux.bp[0], aux.bp[1], aux.bp[2], aux.bp[3],
        num_chunks);

    batch_prefix_serial_fixup<<<1, 1, 0, stream>>>(
        aux.bp[0], aux.bp[1], aux.bp[2], aux.bp[3], super_chunks);

    batch_invert_apply_fixup<<<super_blocks, threads, 0, stream>>>(
        s.bp[0], s.bp[1], s.bp[2], s.bp[3],
        aux.bp[0], aux.bp[1], aux.bp[2], aux.bp[3],
        num_chunks);

    // Apply global chunk-prefix fixup to the original vector.
    batch_invert_apply_fixup<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, s.bp[0], s.bp[1], s.bp[2], s.bp[3], n);

    // Step 3: start_acc[k] = 1/prefix[k]. Keep scanned prefixes in s.bp for
    // boundary reads in backward sweep, and invert a copy in aux.tmp.
    for (int i = 0; i < 4; i++) {
        err = cudaMemcpyAsync(
            aux.tmp[i], s.bp[i], num_chunks * sizeof(uint64_t),
            cudaMemcpyDeviceToDevice, stream);
        if (err != cudaSuccess) return err;
    }

    err = launch_batch_invert_baseline(
        aux.tmp[0], aux.tmp[1], aux.tmp[2], aux.tmp[3],
        s.tmp[0], s.tmp[1], s.tmp[2], s.tmp[3],
        num_chunks, stream, aux);
    if (err != cudaSuccess) return err;

    // Step 4: backward sweep with hierarchical start accumulators.
    batch_invert_backward<<<blocks, threads, 0, stream>>>(
        v0, v1, v2, v3, orig0, orig1, orig2, orig3,
        s.bp[0], s.bp[1], s.bp[2], s.bp[3],
        aux.tmp[0], aux.tmp[1], aux.tmp[2], aux.tmp[3], n);

    return cudaSuccess;
}

// =============================================================================
// Butterfly4: Size-4 inverse DFT butterfly for decomposed iFFT(4n).
// For each j in [0, n):
//   (a0,a1,a2,a3) = (b0[j], b1[j], b2[j], b3[j])
//   b0[j] = (a0 + a1 + a2 + a3) * quarter
//   b1[j] = ((a0 - a2) + omega4_inv*(a1 - a3)) * quarter
//   b2[j] = (a0 - a1 + a2 - a3) * quarter
//   b3[j] = ((a0 - a2) - omega4_inv*(a1 - a3)) * quarter
// =============================================================================

__global__ void butterfly4_kernel(
    uint64_t *__restrict__ b0_0, uint64_t *__restrict__ b0_1,
    uint64_t *__restrict__ b0_2, uint64_t *__restrict__ b0_3,
    uint64_t *__restrict__ b1_0, uint64_t *__restrict__ b1_1,
    uint64_t *__restrict__ b1_2, uint64_t *__restrict__ b1_3,
    uint64_t *__restrict__ b2_0, uint64_t *__restrict__ b2_1,
    uint64_t *__restrict__ b2_2, uint64_t *__restrict__ b2_3,
    uint64_t *__restrict__ b3_0, uint64_t *__restrict__ b3_1,
    uint64_t *__restrict__ b3_2, uint64_t *__restrict__ b3_3,
    const uint64_t w0, const uint64_t w1,
    const uint64_t w2, const uint64_t w3,
    const uint64_t q0, const uint64_t q1,
    const uint64_t q2, const uint64_t q3,
    size_t n)
{
    size_t j = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (j >= n) return;

    uint64_t a0[4] = {b0_0[j], b0_1[j], b0_2[j], b0_3[j]};
    uint64_t a1[4] = {b1_0[j], b1_1[j], b1_2[j], b1_3[j]};
    uint64_t a2[4] = {b2_0[j], b2_1[j], b2_2[j], b2_3[j]};
    uint64_t a3[4] = {b3_0[j], b3_1[j], b3_2[j], b3_3[j]};

    uint64_t omega4_inv[4] = {w0, w1, w2, w3};
    uint64_t quarter[4] = {q0, q1, q2, q3};

    // t0 = a0 + a2, t1 = a0 - a2
    uint64_t t0[4], t1[4];
    fr_add(t0, a0, a2);
    fr_sub(t1, a0, a2);

    // t2 = a1 + a3, t3 = omega4_inv * (a1 - a3)
    uint64_t t2[4], t3[4], diff13[4];
    fr_add(t2, a1, a3);
    fr_sub(diff13, a1, a3);
    fr_mul(t3, omega4_inv, diff13);

    // r0 = (t0 + t2) * quarter = (a0+a1+a2+a3)/4
    uint64_t sum02[4], r0[4];
    fr_add(sum02, t0, t2);
    fr_mul(r0, sum02, quarter);

    // r1 = (t1 + t3) * quarter = ((a0-a2) + w*(a1-a3))/4
    uint64_t sum13[4], r1[4];
    fr_add(sum13, t1, t3);
    fr_mul(r1, sum13, quarter);

    // r2 = (t0 - t2) * quarter = (a0-a1+a2-a3)/4
    uint64_t sub02[4], r2[4];
    fr_sub(sub02, t0, t2);
    fr_mul(r2, sub02, quarter);

    // r3 = (t1 - t3) * quarter = ((a0-a2) - w*(a1-a3))/4
    uint64_t sub13[4], r3[4];
    fr_sub(sub13, t1, t3);
    fr_mul(r3, sub13, quarter);

    b0_0[j] = r0[0]; b0_1[j] = r0[1]; b0_2[j] = r0[2]; b0_3[j] = r0[3];
    b1_0[j] = r1[0]; b1_1[j] = r1[1]; b1_2[j] = r1[2]; b1_3[j] = r1[3];
    b2_0[j] = r2[0]; b2_1[j] = r2[1]; b2_2[j] = r2[2]; b2_3[j] = r2[3];
    b3_0[j] = r3[0]; b3_1[j] = r3[1]; b3_2[j] = r3[2]; b3_3[j] = r3[3];
}

void launch_butterfly4(
    uint64_t *b0_0, uint64_t *b0_1, uint64_t *b0_2, uint64_t *b0_3,
    uint64_t *b1_0, uint64_t *b1_1, uint64_t *b1_2, uint64_t *b1_3,
    uint64_t *b2_0, uint64_t *b2_1, uint64_t *b2_2, uint64_t *b2_3,
    uint64_t *b3_0, uint64_t *b3_1, uint64_t *b3_2, uint64_t *b3_3,
    const uint64_t omega4_inv[4], const uint64_t quarter[4],
    size_t n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    butterfly4_kernel<<<blocks, threads, 0, stream>>>(
        b0_0, b0_1, b0_2, b0_3,
        b1_0, b1_1, b1_2, b1_3,
        b2_0, b2_1, b2_2, b2_3,
        b3_0, b3_1, b3_2, b3_3,
        omega4_inv[0], omega4_inv[1], omega4_inv[2], omega4_inv[3],
        quarter[0], quarter[1], quarter[2], quarter[3],
        n);
}

// =============================================================================
// Fused permutation + boundary constraint kernel for PlonK.
// For each thread i:
//   x_i = coset_gen * omega_powers[i]  (identity point on coset)
//   id1 = beta * x_i
//   id2 = id1 * u     (u = coset_shift)
//   id3 = id1 * u_sq  (u² = coset_shift²)
//   num = Z[i] * (L[i]+id1+gamma) * (R[i]+id2+gamma) * (O[i]+id3+gamma)
//   ZS  = Z[(i+1) % n]
//   den = ZS * (L[i]+beta*S1[i]+gamma) * (R[i]+beta*S2[i]+gamma) * (O[i]+beta*S3[i]+gamma)
//   L1  = L1_scalar * L1_denInv[i]
//   loc = (Z[i] - 1) * L1
//   result[i] = alpha * ((den - num) + alpha * loc)
// =============================================================================

__global__ void plonk_perm_boundary_kernel(
    // Output
    uint64_t *__restrict__ res0, uint64_t *__restrict__ res1,
    uint64_t *__restrict__ res2, uint64_t *__restrict__ res3,
    // Wire evaluations (SoA, natural order)
    const uint64_t *__restrict__ L0, const uint64_t *__restrict__ L1_,
    const uint64_t *__restrict__ L2, const uint64_t *__restrict__ L3,
    const uint64_t *__restrict__ R0, const uint64_t *__restrict__ R1,
    const uint64_t *__restrict__ R2, const uint64_t *__restrict__ R3,
    const uint64_t *__restrict__ O0, const uint64_t *__restrict__ O1,
    const uint64_t *__restrict__ O2, const uint64_t *__restrict__ O3,
    // Z polynomial (SoA, natural order)
    const uint64_t *__restrict__ Z0, const uint64_t *__restrict__ Z1_,
    const uint64_t *__restrict__ Z2, const uint64_t *__restrict__ Z3,
    // Permutation polynomials (SoA, natural order)
    const uint64_t *__restrict__ S1_0, const uint64_t *__restrict__ S1_1,
    const uint64_t *__restrict__ S1_2, const uint64_t *__restrict__ S1_3,
    const uint64_t *__restrict__ S2_0, const uint64_t *__restrict__ S2_1,
    const uint64_t *__restrict__ S2_2, const uint64_t *__restrict__ S2_3,
    const uint64_t *__restrict__ S3_0, const uint64_t *__restrict__ S3_1,
    const uint64_t *__restrict__ S3_2, const uint64_t *__restrict__ S3_3,
    // Batch-inverted L1 denominators
    const uint64_t *__restrict__ dinv0, const uint64_t *__restrict__ dinv1,
    const uint64_t *__restrict__ dinv2, const uint64_t *__restrict__ dinv3,
    // Scalar parameters (passed by value as 4 uint64 each)
    const uint64_t al0, const uint64_t al1, const uint64_t al2, const uint64_t al3,
    const uint64_t be0, const uint64_t be1, const uint64_t be2, const uint64_t be3,
    const uint64_t ga0, const uint64_t ga1, const uint64_t ga2, const uint64_t ga3,
    const uint64_t ls0, const uint64_t ls1, const uint64_t ls2, const uint64_t ls3,
    const uint64_t u0, const uint64_t u1, const uint64_t u2, const uint64_t u3,
    const uint64_t usq0, const uint64_t usq1, const uint64_t usq2, const uint64_t usq3,
    const uint64_t cg0, const uint64_t cg1, const uint64_t cg2, const uint64_t cg3,
    // Twiddle factors: omega^0..omega^(n/2-1) in SoA (half-size from NTT domain)
    // For i >= n/2: omega^i = -omega^(i-n/2) since omega^(n/2) = -1
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    size_t n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t alpha[4]     = {al0, al1, al2, al3};
    uint64_t beta[4]      = {be0, be1, be2, be3};
    uint64_t gamma[4]     = {ga0, ga1, ga2, ga3};
    uint64_t l1_scalar[4] = {ls0, ls1, ls2, ls3};
    uint64_t cshift[4]    = {u0, u1, u2, u3};
    uint64_t cshift_sq[4] = {usq0, usq1, usq2, usq3};
    uint64_t coset_gen[4] = {cg0, cg1, cg2, cg3};

    // Load wires
    uint64_t li[4] = {L0[i], L1_[i], L2[i], L3[i]};
    uint64_t ri[4] = {R0[i], R1[i], R2[i], R3[i]};
    uint64_t oi[4] = {O0[i], O1[i], O2[i], O3[i]};

    // Load Z[i] and Z[(i+1)%n]
    uint64_t zi[4] = {Z0[i], Z1_[i], Z2[i], Z3[i]};
    size_t i_next = (i + 1 < n) ? (i + 1) : 0;
    uint64_t zs[4] = {Z0[i_next], Z1_[i_next], Z2[i_next], Z3[i_next]};

    // Identity permutation: x_i = coset_gen * omega^i
    // Twiddles are half-size (n/2). For i >= n/2: omega^i = -omega^(i-n/2)
    size_t half_n = n >> 1;
    uint64_t tw[4];
    if (i < half_n) {
        tw[0] = tw0[i]; tw[1] = tw1[i]; tw[2] = tw2[i]; tw[3] = tw3[i];
    } else {
        // omega^i = -omega^(i-n/2)
        size_t j = i - half_n;
        uint64_t pos[4] = {tw0[j], tw1[j], tw2[j], tw3[j]};
        uint64_t zero[4] = {0, 0, 0, 0};
        fr_sub(tw, zero, pos); // tw = -pos
    }
    uint64_t x_i[4];
    fr_mul(x_i, coset_gen, tw);

    // id1 = beta * x_i,  id2 = id1 * u,  id3 = id1 * u²
    uint64_t id1[4], id2[4], id3[4];
    fr_mul(id1, beta, x_i);
    fr_mul(id2, id1, cshift);
    fr_mul(id3, id1, cshift_sq);

    // num = Z[i] * (L+id1+gamma) * (R+id2+gamma) * (O+id3+gamma)
    uint64_t t1[4], t2[4], t3[4];
    fr_add(t1, li, id1);
    fr_add(t1, t1, gamma);
    fr_add(t2, ri, id2);
    fr_add(t2, t2, gamma);
    fr_add(t3, oi, id3);
    fr_add(t3, t3, gamma);

    uint64_t num[4], tmp[4];
    fr_mul(num, zi, t1);
    fr_mul(tmp, num, t2);
    fr_mul(num, tmp, t3);

    // den = ZS * (L+beta*S1+gamma) * (R+beta*S2+gamma) * (O+beta*S3+gamma)
    uint64_t s1[4] = {S1_0[i], S1_1[i], S1_2[i], S1_3[i]};
    uint64_t s2[4] = {S2_0[i], S2_1[i], S2_2[i], S2_3[i]};
    uint64_t s3[4] = {S3_0[i], S3_1[i], S3_2[i], S3_3[i]};

    uint64_t bs1[4], bs2[4], bs3[4];
    fr_mul(bs1, beta, s1);
    fr_mul(bs2, beta, s2);
    fr_mul(bs3, beta, s3);

    fr_add(t1, li, bs1);
    fr_add(t1, t1, gamma);
    fr_add(t2, ri, bs2);
    fr_add(t2, t2, gamma);
    fr_add(t3, oi, bs3);
    fr_add(t3, t3, gamma);

    uint64_t den[4];
    fr_mul(den, zs, t1);
    fr_mul(tmp, den, t2);
    fr_mul(den, tmp, t3);

    // ordering = den - num (gnark convention: ZS*prod_sigma - Z*prod_id)
    uint64_t ordering[4];
    fr_sub(ordering, den, num);

    // L1_i = l1_scalar * L1_denInv[i]
    uint64_t dinv[4] = {dinv0[i], dinv1[i], dinv2[i], dinv3[i]};
    uint64_t l1_val[4];
    fr_mul(l1_val, l1_scalar, dinv);

    // local = (Z[i] - 1) * L1_i
    uint64_t one[4] = {Fr_params::ONE[0], Fr_params::ONE[1],
                        Fr_params::ONE[2], Fr_params::ONE[3]};
    uint64_t zm1[4];
    fr_sub(zm1, zi, one);
    uint64_t local_val[4];
    fr_mul(local_val, zm1, l1_val);

    // result[i] = alpha * (ordering + alpha * local)
    uint64_t al_local[4];
    fr_mul(al_local, alpha, local_val);
    uint64_t sum[4];
    fr_add(sum, ordering, al_local);
    uint64_t result[4];
    fr_mul(result, alpha, sum);

    res0[i] = result[0];
    res1[i] = result[1];
    res2[i] = result[2];
    res3[i] = result[3];
}

// Struct to pass all scalar parameters to the launch function
struct PlonkPermBoundaryParams {
    uint64_t alpha[4];
    uint64_t beta[4];
    uint64_t gamma[4];
    uint64_t l1_scalar[4];
    uint64_t coset_shift[4];
    uint64_t coset_shift_sq[4];
    uint64_t coset_gen[4];
};

void launch_plonk_perm_boundary(
    uint64_t *res0, uint64_t *res1, uint64_t *res2, uint64_t *res3,
    const uint64_t *L0, const uint64_t *L1, const uint64_t *L2, const uint64_t *L3,
    const uint64_t *R0, const uint64_t *R1, const uint64_t *R2, const uint64_t *R3,
    const uint64_t *O0, const uint64_t *O1, const uint64_t *O2, const uint64_t *O3,
    const uint64_t *Z0, const uint64_t *Z1, const uint64_t *Z2, const uint64_t *Z3,
    const uint64_t *S1_0, const uint64_t *S1_1, const uint64_t *S1_2, const uint64_t *S1_3,
    const uint64_t *S2_0, const uint64_t *S2_1, const uint64_t *S2_2, const uint64_t *S2_3,
    const uint64_t *S3_0, const uint64_t *S3_1, const uint64_t *S3_2, const uint64_t *S3_3,
    const uint64_t *dinv0, const uint64_t *dinv1, const uint64_t *dinv2, const uint64_t *dinv3,
    const PlonkPermBoundaryParams &params,
    const uint64_t *tw0, const uint64_t *tw1, const uint64_t *tw2, const uint64_t *tw3,
    size_t n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    plonk_perm_boundary_kernel<<<blocks, threads, 0, stream>>>(
        res0, res1, res2, res3,
        L0, L1, L2, L3,
        R0, R1, R2, R3,
        O0, O1, O2, O3,
        Z0, Z1, Z2, Z3,
        S1_0, S1_1, S1_2, S1_3,
        S2_0, S2_1, S2_2, S2_3,
        S3_0, S3_1, S3_2, S3_3,
        dinv0, dinv1, dinv2, dinv3,
        params.alpha[0], params.alpha[1], params.alpha[2], params.alpha[3],
        params.beta[0], params.beta[1], params.beta[2], params.beta[3],
        params.gamma[0], params.gamma[1], params.gamma[2], params.gamma[3],
        params.l1_scalar[0], params.l1_scalar[1], params.l1_scalar[2], params.l1_scalar[3],
        params.coset_shift[0], params.coset_shift[1], params.coset_shift[2], params.coset_shift[3],
        params.coset_shift_sq[0], params.coset_shift_sq[1], params.coset_shift_sq[2], params.coset_shift_sq[3],
        params.coset_gen[0], params.coset_gen[1], params.coset_gen[2], params.coset_gen[3],
        tw0, tw1, tw2, tw3,
        n);
}

// =============================================================================
// PlonK Z-polynomial per-element ratio computation.
// For each i in [0, n):
//   num[i] = (L[i]+β*ω^i+γ) * (R[i]+β*g*ω^i+γ) * (O[i]+β*g²*ω^i+γ)
//   den[i] = (L[i]+β*id[S[i]]+γ) * (R[i]+β*id[S[n+i]]+γ) * (O[i]+β*id[S[2n+i]]+γ)
// where id[j] = g^(j>>log2n) * ω^(j&(n-1))
//
// In-place: L is overwritten with num, R with den. O is read-only.
// =============================================================================

// Helper: compute omega^pos from twiddle table.
// Twiddles store ω^0..ω^(n/2-1). For pos >= n/2: ω^pos = -ω^(pos-n/2).
static __device__ __forceinline__ void get_omega(
    uint64_t result[4], size_t pos, size_t half_n,
    const uint64_t *tw0, const uint64_t *tw1,
    const uint64_t *tw2, const uint64_t *tw3)
{
    if (pos < half_n) {
        result[0] = tw0[pos]; result[1] = tw1[pos];
        result[2] = tw2[pos]; result[3] = tw3[pos];
    } else {
        size_t j = pos - half_n;
        uint64_t p[4] = {tw0[j], tw1[j], tw2[j], tw3[j]};
        uint64_t z[4] = {0, 0, 0, 0};
        fr_sub(result, z, p);
    }
}

// Helper: compute identity permutation evaluation at index perm_idx.
// id[perm_idx] = g^coset * ω^pos where coset = perm_idx >> log2n, pos = perm_idx & (n-1).
static __device__ __forceinline__ void get_perm_id_eval(
    uint64_t result[4], int64_t perm_idx,
    size_t n, size_t half_n, unsigned log2n,
    const uint64_t g_mul[4], const uint64_t g_sq[4],
    const uint64_t *tw0, const uint64_t *tw1,
    const uint64_t *tw2, const uint64_t *tw3)
{
    unsigned coset = (unsigned)((size_t)perm_idx >> log2n);
    size_t pos = (size_t)perm_idx & (n - 1);

    uint64_t omega_pos[4];
    get_omega(omega_pos, pos, half_n, tw0, tw1, tw2, tw3);

    if (coset == 0) {
        result[0] = omega_pos[0]; result[1] = omega_pos[1];
        result[2] = omega_pos[2]; result[3] = omega_pos[3];
    } else if (coset == 1) {
        fr_mul(result, g_mul, omega_pos);
    } else {
        fr_mul(result, g_sq, omega_pos);
    }
}

__global__ void plonk_z_ratio_kernel(
    // L/num (read L, write num in-place)
    uint64_t *__restrict__ LN0, uint64_t *__restrict__ LN1,
    uint64_t *__restrict__ LN2, uint64_t *__restrict__ LN3,
    // R/den (read R, write den in-place)
    uint64_t *__restrict__ RD0, uint64_t *__restrict__ RD1,
    uint64_t *__restrict__ RD2, uint64_t *__restrict__ RD3,
    // O (read-only)
    const uint64_t *__restrict__ O0, const uint64_t *__restrict__ O1,
    const uint64_t *__restrict__ O2, const uint64_t *__restrict__ O3,
    // Permutation table (3n int64s, device memory)
    const int64_t *__restrict__ perm,
    // Scalar parameters (passed by value)
    const uint64_t be0, const uint64_t be1, const uint64_t be2, const uint64_t be3,
    const uint64_t ga0, const uint64_t ga1, const uint64_t ga2, const uint64_t ga3,
    const uint64_t gm0, const uint64_t gm1, const uint64_t gm2, const uint64_t gm3,
    const uint64_t gs0, const uint64_t gs1, const uint64_t gs2, const uint64_t gs3,
    // Twiddle factors (omega^0..omega^(n/2-1), SoA)
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    size_t n, unsigned log2n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t beta[4]  = {be0, be1, be2, be3};
    uint64_t gamma[4] = {ga0, ga1, ga2, ga3};
    uint64_t g_mul[4] = {gm0, gm1, gm2, gm3};
    uint64_t g_sq[4]  = {gs0, gs1, gs2, gs3};

    size_t half_n = n >> 1;

    // Read L[i], R[i], O[i] into registers before overwriting
    uint64_t li[4] = {LN0[i], LN1[i], LN2[i], LN3[i]};
    uint64_t ri[4] = {RD0[i], RD1[i], RD2[i], RD3[i]};
    uint64_t oi[4] = {O0[i], O1[i], O2[i], O3[i]};

    // Compute omega^i from twiddle table
    uint64_t omega_i[4];
    get_omega(omega_i, i, half_n, tw0, tw1, tw2, tw3);

    // β * identity evaluations: β*ω^i, β*g*ω^i, β*g²*ω^i
    // Optimize: compute β*ω^i once, then multiply by g and g²
    uint64_t bid0[4];
    fr_mul(bid0, beta, omega_i);      // β * ω^i
    uint64_t bid1[4];
    fr_mul(bid1, g_mul, bid0);        // g * β * ω^i = β * g * ω^i
    uint64_t bid2[4];
    fr_mul(bid2, g_sq, bid0);         // g² * β * ω^i = β * g² * ω^i

    // Numerator: (L + β*id0 + γ) * (R + β*id1 + γ) * (O + β*id2 + γ)
    uint64_t t1[4], t2[4], t3[4], tmp[4];
    fr_add(t1, li, bid0);
    fr_add(t1, t1, gamma);
    fr_add(t2, ri, bid1);
    fr_add(t2, t2, gamma);
    fr_add(t3, oi, bid2);
    fr_add(t3, t3, gamma);

    uint64_t num_val[4];
    fr_mul(tmp, t1, t2);
    fr_mul(num_val, tmp, t3);

    // Denominator: look up permutation for each wire and compute identity evaluation
    uint64_t sid0[4], sid1[4], sid2[4];
    get_perm_id_eval(sid0, perm[i],       n, half_n, log2n, g_mul, g_sq, tw0, tw1, tw2, tw3);
    get_perm_id_eval(sid1, perm[n + i],   n, half_n, log2n, g_mul, g_sq, tw0, tw1, tw2, tw3);
    get_perm_id_eval(sid2, perm[2*n + i], n, half_n, log2n, g_mul, g_sq, tw0, tw1, tw2, tw3);

    uint64_t bs0[4], bs1[4], bs2[4];
    fr_mul(bs0, beta, sid0);
    fr_mul(bs1, beta, sid1);
    fr_mul(bs2, beta, sid2);

    fr_add(t1, li, bs0);
    fr_add(t1, t1, gamma);
    fr_add(t2, ri, bs1);
    fr_add(t2, t2, gamma);
    fr_add(t3, oi, bs2);
    fr_add(t3, t3, gamma);

    uint64_t den_val[4];
    fr_mul(tmp, t1, t2);
    fr_mul(den_val, tmp, t3);

    // Write num to L, den to R (in-place)
    LN0[i] = num_val[0]; LN1[i] = num_val[1]; LN2[i] = num_val[2]; LN3[i] = num_val[3];
    RD0[i] = den_val[0]; RD1[i] = den_val[1]; RD2[i] = den_val[2]; RD3[i] = den_val[3];
}

struct PlonkZRatioParams {
    uint64_t beta[4];
    uint64_t gamma[4];
    uint64_t g_mul[4];
    uint64_t g_sq[4];
};

void launch_plonk_z_ratio(
    uint64_t *LN0, uint64_t *LN1, uint64_t *LN2, uint64_t *LN3,
    uint64_t *RD0, uint64_t *RD1, uint64_t *RD2, uint64_t *RD3,
    const uint64_t *O0, const uint64_t *O1, const uint64_t *O2, const uint64_t *O3,
    const int64_t *d_perm,
    const PlonkZRatioParams &params,
    const uint64_t *tw0, const uint64_t *tw1, const uint64_t *tw2, const uint64_t *tw3,
    size_t n, unsigned log2n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    plonk_z_ratio_kernel<<<blocks, threads, 0, stream>>>(
        LN0, LN1, LN2, LN3,
        RD0, RD1, RD2, RD3,
        O0, O1, O2, O3,
        d_perm,
        params.beta[0], params.beta[1], params.beta[2], params.beta[3],
        params.gamma[0], params.gamma[1], params.gamma[2], params.gamma[3],
        params.g_mul[0], params.g_mul[1], params.g_mul[2], params.g_mul[3],
        params.g_sq[0], params.g_sq[1], params.g_sq[2], params.g_sq[3],
        tw0, tw1, tw2, tw3,
        n, log2n);
}

// =============================================================================
// ComputeL1Den: out[i] = cosetGen * omega^i - 1
// Uses the same twiddle access pattern as plonk_perm_boundary_kernel.
// The caller should BatchInvert the result to get L1DenInv.
// =============================================================================

__global__ void compute_l1_den_kernel(
    uint64_t *__restrict__ out0, uint64_t *__restrict__ out1,
    uint64_t *__restrict__ out2, uint64_t *__restrict__ out3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint64_t cg0, uint64_t cg1, uint64_t cg2, uint64_t cg3, size_t n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t coset_gen[4] = {cg0, cg1, cg2, cg3};

    // Get omega^i from twiddle table (half-size, same pattern as perm kernel)
    size_t half_n = n >> 1;
    uint64_t tw[4];
    if (i < half_n) {
        tw[0] = tw0[i]; tw[1] = tw1[i]; tw[2] = tw2[i]; tw[3] = tw3[i];
    } else {
        size_t j = i - half_n;
        uint64_t pos[4] = {tw0[j], tw1[j], tw2[j], tw3[j]};
        uint64_t zero[4] = {0, 0, 0, 0};
        fr_sub(tw, zero, pos);
    }

    // result = cosetGen * omega^i - 1
    uint64_t prod[4];
    fr_mul(prod, coset_gen, tw);

    uint64_t one[4] = {Fr_params::ONE[0], Fr_params::ONE[1],
                        Fr_params::ONE[2], Fr_params::ONE[3]};
    uint64_t result[4];
    fr_sub(result, prod, one);

    out0[i] = result[0];
    out1[i] = result[1];
    out2[i] = result[2];
    out3[i] = result[3];
}

void launch_compute_l1_den(
    uint64_t *out0, uint64_t *out1, uint64_t *out2, uint64_t *out3,
    const uint64_t *tw0, const uint64_t *tw1, const uint64_t *tw2, const uint64_t *tw3,
    const uint64_t cg[4], size_t n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    compute_l1_den_kernel<<<blocks, threads, 0, stream>>>(
        out0, out1, out2, out3,
        tw0, tw1, tw2, tw3,
        cg[0], cg[1], cg[2], cg[3], n);
}

// =============================================================================
// Reduce blinded polynomial for coset evaluation
//
// dst[i] = src[i] + tail[j] * cosetPowN   for j in [0, tail_len), i = j
// dst[i] = src[i]                           for i in [tail_len, n)
//
// tail coefficients are loaded from shared memory (uploaded from host AoS).
// tail_len is tiny (2-3), so we broadcast via shared memory.
// =============================================================================

__global__ void reduce_blinded_coset_kernel(
    uint64_t *__restrict__ dst0, uint64_t *__restrict__ dst1,
    uint64_t *__restrict__ dst2, uint64_t *__restrict__ dst3,
    const uint64_t *__restrict__ src0, const uint64_t *__restrict__ src1,
    const uint64_t *__restrict__ src2, const uint64_t *__restrict__ src3,
    const uint64_t cpn0, const uint64_t cpn1,
    const uint64_t cpn2, const uint64_t cpn3,
    const uint64_t *__restrict__ tail_aos,  // AoS: [t0_l0..t0_l3, t1_l0..t1_l3, ...]
    uint32_t tail_len, uint32_t n)
{
    uint32_t i = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t v[4] = { src0[i], src1[i], src2[i], src3[i] };

    if (i < tail_len) {
        // v[i] += tail[i] * cosetPowN
        uint64_t t[4] = { tail_aos[i*4], tail_aos[i*4+1], tail_aos[i*4+2], tail_aos[i*4+3] };
        uint64_t cpn[4] = { cpn0, cpn1, cpn2, cpn3 };
        uint64_t prod[4];
        fr_mul(prod, t, cpn);
        fr_add(v, v, prod);
    }

    dst0[i] = v[0]; dst1[i] = v[1]; dst2[i] = v[2]; dst3[i] = v[3];
}

void launch_reduce_blinded_coset(
    uint64_t *dst0, uint64_t *dst1, uint64_t *dst2, uint64_t *dst3,
    const uint64_t *src0, const uint64_t *src1,
    const uint64_t *src2, const uint64_t *src3,
    const uint64_t cpn[4],
    const uint64_t *tail_device,
    uint32_t tail_len, uint32_t n, cudaStream_t stream)
{
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    reduce_blinded_coset_kernel<<<blocks, threads, 0, stream>>>(
        dst0, dst1, dst2, dst3,
        src0, src1, src2, src3,
        cpn[0], cpn[1], cpn[2], cpn[3],
        tail_device, tail_len, n);
}

// =============================================================================
} // namespace gnark_gpu
