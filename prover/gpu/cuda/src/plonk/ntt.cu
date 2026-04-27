// ═══════════════════════════════════════════════════════════════════════════════
// NTT (Number Theoretic Transform) for BLS12-377 scalar field Fr
//
// The NTT is the finite-field analog of the FFT, computing:
//   Forward:  Ŷ[k] = Σᵢ Y[i] · ωⁱᵏ   (evaluation at roots of unity)
//   Inverse:  Y[i] = (1/n) Σₖ Ŷ[k] · ω⁻ⁱᵏ  (interpolation)
//
// where ω is a primitive n-th root of unity in Fr (BLS12-377 scalar field).
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │  DIF (Decimation-In-Frequency) Butterfly      s = stage (0 = first)   │
// │                                                                        │
// │     a ──────┬────── a' = a + b                                        │
// │             │                                                          │
// │             × ω                                                        │
// │             │                                                          │
// │     b ──────┴────── b' = (a - b) · ω                                  │
// │                                                                        │
// │  Natural input → bit-reversed output                                   │
// │  Stages: s = 0 (largest groups) down to s = log₂(n)-1 (pairs)        │
// ├─────────────────────────────────────────────────────────────────────────┤
// │  DIT (Decimation-In-Time) Butterfly                                    │
// │                                                                        │
// │     a ──────┬────── a' = a + ω·b                                      │
// │             │                                                          │
// │             × ω                                                        │
// │             │                                                          │
// │     b ──────┴────── b' = a - ω·b                                      │
// │                                                                        │
// │  Bit-reversed input → natural output (with 1/n scale)                  │
// │  Stages: s = log₂(n)-1 (pairs) down to s = 0 (largest groups)        │
// └─────────────────────────────────────────────────────────────────────────┘
//
// Kernel dispatch strategy (minimizes kernel launches):
//
//   Forward (DIF):     radix-8 → radix-2 → fused tail
//   Inverse (DIT):     fused tail → radix-8 → radix-2
//
//   Example for n = 2²⁰ (20 stages):
//     DIF: stages 0-2 (r8), 3-5 (r8), 6-8 (r8), 9-19 (fused tail, 11 stages)
//       = 4 kernel launches instead of 20
//
//   Fused tail: last 11-12 stages run entirely in shared memory (one block per
//   chunk of 2¹¹ or 2¹² elements), eliminating global memory round-trips.
//   Adaptive: tail_log=12 on GPUs with ≥128KB shared memory, else tail_log=11.
//
// Data layout: SoA (4 limb arrays of n uint64s each), same as FrVector.
// Twiddle layout: SoA (4 limb arrays), n/2 elements in Montgomery form.
// Twiddle indexing: flat table ω⁰, ω¹, ..., ω^(n/2-1).
//   Stage s butterfly at position j uses twiddle at index j · 2ˢ.
// ═══════════════════════════════════════════════════════════════════════════════

#include "fr_arith.cuh"
#include <cuda_runtime.h>
#include <cstdint>

namespace gnark_gpu {

void launch_transpose_aos_to_soa_fr(uint64_t *limb0, uint64_t *limb1, uint64_t *limb2,
                                    uint64_t *limb3, const uint64_t *aos_data, size_t count,
                                    cudaStream_t stream);

namespace {

constexpr unsigned NTT_THREADS = 256;
constexpr uint32_t NTT_FUSED_TAIL_MIN_N = 1u << 22;

} // namespace

// =============================================================================
// DIF butterfly kernel (one stage of forward NTT)
//
// For stage s (0-indexed from MSB):
//   half = n >> (s+1)
//   For each butterfly tid in [0, n/2):
//     group = tid / half;  j = tid % half
//     idx_a = group * 2*half + j;  idx_b = idx_a + half
//     tw_idx = j * (1 << s)   [into flat twiddle table of n/2 entries]
//     a' = a + b;  b' = (a - b) * w
// =============================================================================

__global__ void ntt_dif_butterfly_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t num_butterflies, uint32_t half, uint32_t half_mask, uint32_t tw_stride)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= num_butterflies) return;

    // half is always a power-of-two: replace costly div/mod with bit ops.
    uint32_t j = tid & half_mask;
    uint32_t group_base = tid & ~half_mask;
    uint32_t idx_a = (group_base << 1) | j;
    uint32_t idx_b = idx_a + half;
    uint32_t tw_idx = j * tw_stride;

    // Load a, b
    uint64_t a[4] = { d0[idx_a], d1[idx_a], d2[idx_a], d3[idx_a] };
    uint64_t b[4] = { d0[idx_b], d1[idx_b], d2[idx_b], d3[idx_b] };

    // Load twiddle through read-only cache path.
    uint64_t w[4] = {
        __ldg(tw0 + tw_idx), __ldg(tw1 + tw_idx),
        __ldg(tw2 + tw_idx), __ldg(tw3 + tw_idx)
    };

    // DIF butterfly: a' = a + b;  b' = (a - b) * w
    uint64_t sum[4], diff[4], prod[4];
    fr_add(sum, a, b);
    fr_sub(diff, a, b);
    fr_mul(prod, diff, w);

    // Store
    d0[idx_a] = sum[0]; d1[idx_a] = sum[1]; d2[idx_a] = sum[2]; d3[idx_a] = sum[3];
    d0[idx_b] = prod[0]; d1[idx_b] = prod[1]; d2[idx_b] = prod[2]; d3[idx_b] = prod[3];
}

// =============================================================================
// Radix-8 DIF butterfly kernel: fuses three adjacent stages (s, s+1, s+2).
//
// Each thread processes 8 elements from one radix-8 group.
// Stage s:   butterflies at distance half_s  = n >> (s+1)
// Stage s+1: butterflies at distance half_s1 = half_s/2
// Stage s+2: butterflies at distance half_s2 = half_s/4
//
// 8 element positions within a group of 2*half_s:
//   p0 = base+j, p1 = p0+half_s2, p2 = p0+half_s1, p3 = p2+half_s2
//   p4 = p0+half_s, p5 = p4+half_s2, p6 = p4+half_s1, p7 = p6+half_s2
//
// Twiddle loads: 4 (stage s) + 2 (stage s+1) + 1 (stage s+2) = 7 total.
// Arithmetic: 12 fr_mul + 12 fr_add + 12 fr_sub per thread.
// =============================================================================

__global__ void ntt_dif_radix8_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t n, int stage_s)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    uint32_t num_r8 = n >> 3;
    if (tid >= num_r8) return;

    uint32_t half_s  = n >> (stage_s + 1);
    uint32_t half_s1 = half_s >> 1;
    uint32_t half_s2 = half_s >> 2;

    uint32_t j     = tid & (half_s2 - 1);
    uint32_t group = tid >> (__ffs(half_s2) - 1);

    uint32_t base = group * (2 * half_s);
    uint32_t p0 = base + j;
    uint32_t p1 = p0 + half_s2;
    uint32_t p2 = p0 + half_s1;
    uint32_t p3 = p2 + half_s2;
    uint32_t p4 = p0 + half_s;
    uint32_t p5 = p4 + half_s2;
    uint32_t p6 = p4 + half_s1;
    uint32_t p7 = p6 + half_s2;

    // Load 8 elements
    uint64_t a0[4] = { d0[p0], d1[p0], d2[p0], d3[p0] };
    uint64_t a1[4] = { d0[p1], d1[p1], d2[p1], d3[p1] };
    uint64_t a2[4] = { d0[p2], d1[p2], d2[p2], d3[p2] };
    uint64_t a3[4] = { d0[p3], d1[p3], d2[p3], d3[p3] };
    uint64_t a4[4] = { d0[p4], d1[p4], d2[p4], d3[p4] };
    uint64_t a5[4] = { d0[p5], d1[p5], d2[p5], d3[p5] };
    uint64_t a6[4] = { d0[p6], d1[p6], d2[p6], d3[p6] };
    uint64_t a7[4] = { d0[p7], d1[p7], d2[p7], d3[p7] };

    uint32_t tw_stride_s  = 1u << stage_s;
    uint32_t tw_stride_s1 = tw_stride_s << 1;
    uint32_t tw_stride_s2 = tw_stride_s << 2;

    uint64_t w[4], sum[4], diff[4];
    uint32_t twi;

    // --- Stage s: 4 DIF butterflies at distance half_s ---
    // Pairs: (a0,a4), (a1,a5), (a2,a6), (a3,a7)
    twi = j * tw_stride_s;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_add(sum, a0, a4); fr_sub(diff, a0, a4); fr_mul(a4, diff, w);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];

    twi = (j + half_s2) * tw_stride_s;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_add(sum, a1, a5); fr_sub(diff, a1, a5); fr_mul(a5, diff, w);
    a1[0]=sum[0]; a1[1]=sum[1]; a1[2]=sum[2]; a1[3]=sum[3];

    twi = (j + half_s1) * tw_stride_s;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_add(sum, a2, a6); fr_sub(diff, a2, a6); fr_mul(a6, diff, w);
    a2[0]=sum[0]; a2[1]=sum[1]; a2[2]=sum[2]; a2[3]=sum[3];

    twi = (j + half_s1 + half_s2) * tw_stride_s;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_add(sum, a3, a7); fr_sub(diff, a3, a7); fr_mul(a7, diff, w);
    a3[0]=sum[0]; a3[1]=sum[1]; a3[2]=sum[2]; a3[3]=sum[3];

    // --- Stage s+1: 4 DIF butterflies at distance half_s1 ---
    // Top: (a0,a2), (a1,a3)  Bottom: (a4,a6), (a5,a7)
    uint64_t ws1_0[4], ws1_1[4];
    twi = j * tw_stride_s1;
    ws1_0[0] = __ldg(tw0+twi); ws1_0[1] = __ldg(tw1+twi);
    ws1_0[2] = __ldg(tw2+twi); ws1_0[3] = __ldg(tw3+twi);
    twi = (j + half_s2) * tw_stride_s1;
    ws1_1[0] = __ldg(tw0+twi); ws1_1[1] = __ldg(tw1+twi);
    ws1_1[2] = __ldg(tw2+twi); ws1_1[3] = __ldg(tw3+twi);

    fr_add(sum, a0, a2); fr_sub(diff, a0, a2); fr_mul(a2, diff, ws1_0);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];

    fr_add(sum, a1, a3); fr_sub(diff, a1, a3); fr_mul(a3, diff, ws1_1);
    a1[0]=sum[0]; a1[1]=sum[1]; a1[2]=sum[2]; a1[3]=sum[3];

    fr_add(sum, a4, a6); fr_sub(diff, a4, a6); fr_mul(a6, diff, ws1_0);
    a4[0]=sum[0]; a4[1]=sum[1]; a4[2]=sum[2]; a4[3]=sum[3];

    fr_add(sum, a5, a7); fr_sub(diff, a5, a7); fr_mul(a7, diff, ws1_1);
    a5[0]=sum[0]; a5[1]=sum[1]; a5[2]=sum[2]; a5[3]=sum[3];

    // --- Stage s+2: 4 DIF butterflies at distance half_s2 ---
    // Pairs: (a0,a1), (a2,a3), (a4,a5), (a6,a7) — all same twiddle
    twi = j * tw_stride_s2;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);

    fr_add(sum, a0, a1); fr_sub(diff, a0, a1); fr_mul(a1, diff, w);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];

    fr_add(sum, a2, a3); fr_sub(diff, a2, a3); fr_mul(a3, diff, w);
    a2[0]=sum[0]; a2[1]=sum[1]; a2[2]=sum[2]; a2[3]=sum[3];

    fr_add(sum, a4, a5); fr_sub(diff, a4, a5); fr_mul(a5, diff, w);
    a4[0]=sum[0]; a4[1]=sum[1]; a4[2]=sum[2]; a4[3]=sum[3];

    fr_add(sum, a6, a7); fr_sub(diff, a6, a7); fr_mul(a7, diff, w);
    a6[0]=sum[0]; a6[1]=sum[1]; a6[2]=sum[2]; a6[3]=sum[3];

    // Store
    d0[p0]=a0[0]; d1[p0]=a0[1]; d2[p0]=a0[2]; d3[p0]=a0[3];
    d0[p1]=a1[0]; d1[p1]=a1[1]; d2[p1]=a1[2]; d3[p1]=a1[3];
    d0[p2]=a2[0]; d1[p2]=a2[1]; d2[p2]=a2[2]; d3[p2]=a2[3];
    d0[p3]=a3[0]; d1[p3]=a3[1]; d2[p3]=a3[2]; d3[p3]=a3[3];
    d0[p4]=a4[0]; d1[p4]=a4[1]; d2[p4]=a4[2]; d3[p4]=a4[3];
    d0[p5]=a5[0]; d1[p5]=a5[1]; d2[p5]=a5[2]; d3[p5]=a5[3];
    d0[p6]=a6[0]; d1[p6]=a6[1]; d2[p6]=a6[2]; d3[p6]=a6[3];
    d0[p7]=a7[0]; d1[p7]=a7[1]; d2[p7]=a7[2]; d3[p7]=a7[3];
}

// =============================================================================
// DIF fused tail kernel: processes the LAST TAIL_LOG stages in shared memory.
//
// For n = 2²⁰ and TAIL_LOG = 11:
//   - Each block handles a chunk of 2¹¹ = 2048 contiguous elements
//   - 512 blocks cover the full array (2²⁰ / 2¹¹)
//   - Stages 9 through 19 (11 stages) execute entirely in shared memory
//   - One global load + one global store replaces 11 × 2 global accesses
//
// Shared memory: 4 limbs × 2^TAIL_LOG × 8 bytes
//   TAIL_LOG=11: 4 × 2048 × 8 = 64 KiB  (fits 99 KB optin on Blackwell)
//   TAIL_LOG=12: 4 × 4096 × 8 = 128 KiB (needs ≥128 KB; A100/H100)
//
// Each thread processes multiple butterflies per stage via strided access.
// __syncthreads() between stages ensures data consistency.
// =============================================================================
template <int TAIL_LOG>
__global__ void __launch_bounds__(1024, 1) ntt_dif_tail_fused_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t n, int stage_start)
{
    constexpr uint32_t span = 1u << TAIL_LOG;
    constexpr uint32_t butterflies_per_chunk = span >> 1;

    uint32_t chunk = (uint32_t)blockIdx.x;
    uint32_t base = chunk * span;
    uint32_t t = threadIdx.x;
    uint32_t P = blockDim.x;

    extern __shared__ uint64_t shmem[];
    uint64_t *s0 = shmem;
    uint64_t *s1 = s0 + span;
    uint64_t *s2 = s1 + span;
    uint64_t *s3 = s2 + span;

    // Load: each thread handles span/P elements
    for (uint32_t i = t; i < span; i += P) {
        uint32_t global_idx = base + i;
        if (global_idx < n) {
            s0[i] = d0[global_idx];
            s1[i] = d1[global_idx];
            s2[i] = d2[global_idx];
            s3[i] = d3[global_idx];
        }
    }
    __syncthreads();

    #pragma unroll
    for (int st = 0; st < TAIL_LOG; st++) {
        int s = stage_start + st;
        uint32_t half = n >> (s + 1);
        uint32_t half_mask = half - 1;
        uint32_t tw_stride = 1u << s;

        for (uint32_t bt = t; bt < butterflies_per_chunk; bt += P) {
            uint32_t j = bt & half_mask;
            uint32_t group_base = bt & ~half_mask;
            uint32_t idx_a = (group_base << 1) | j;
            uint32_t idx_b = idx_a + half;
            uint32_t tw_idx = j * tw_stride;

            uint64_t a[4] = { s0[idx_a], s1[idx_a], s2[idx_a], s3[idx_a] };
            uint64_t b[4] = { s0[idx_b], s1[idx_b], s2[idx_b], s3[idx_b] };
            uint64_t w[4] = {
                __ldg(tw0 + tw_idx), __ldg(tw1 + tw_idx),
                __ldg(tw2 + tw_idx), __ldg(tw3 + tw_idx)
            };

            uint64_t sum[4], diff[4], prod[4];
            fr_add(sum, a, b);
            fr_sub(diff, a, b);
            fr_mul(prod, diff, w);

            s0[idx_a] = sum[0]; s1[idx_a] = sum[1]; s2[idx_a] = sum[2]; s3[idx_a] = sum[3];
            s0[idx_b] = prod[0]; s1[idx_b] = prod[1]; s2[idx_b] = prod[2]; s3[idx_b] = prod[3];
        }
        __syncthreads();
    }

    // Store: each thread handles span/P elements
    for (uint32_t i = t; i < span; i += P) {
        uint32_t global_idx = base + i;
        if (global_idx < n) {
            d0[global_idx] = s0[i];
            d1[global_idx] = s1[i];
            d2[global_idx] = s2[i];
            d3[global_idx] = s3[i];
        }
    }
}

// =============================================================================
// Fused ScaleByPowers + first DIF butterfly stage for CosetFFT.
// Eliminates one full memory round-trip by computing v[i] *= g^i inline
// during the first butterfly stage (s=0) of the DIF NTT.
//
// For s=0: idx_a = tid, idx_b = tid + n/2, tw_idx = tid.
// g^(n/2) is passed as a parameter to avoid recomputation.
// =============================================================================

__global__ void ntt_dif_first_stage_fused_scale_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    const uint64_t g0, const uint64_t g1,
    const uint64_t g2, const uint64_t g3,
    const uint64_t gh0, const uint64_t gh1,  // g^(n/2)
    const uint64_t gh2, const uint64_t gh3,
    uint32_t num_butterflies)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= num_butterflies) return;

    // Stage 0 indexing: idx_a = tid, idx_b = tid + n/2
    uint32_t idx_a = tid;
    uint32_t idx_b = tid + num_butterflies; // n/2 = num_butterflies

    // --- Compute g^idx_a using shared memory power table ---
    __shared__ uint64_t sh_power[4];    // g^block_start
    __shared__ uint64_t sh_pow2[8][4];  // g^(2^k) for k in [0,7]

    if (threadIdx.x == 0) {
        uint64_t pow2[4] = {g0, g1, g2, g3};
        #pragma unroll
        for (int k = 0; k < 8; k++) {
            sh_pow2[k][0] = pow2[0]; sh_pow2[k][1] = pow2[1];
            sh_pow2[k][2] = pow2[2]; sh_pow2[k][3] = pow2[3];
            uint64_t sq[4];
            fr_mul(sq, pow2, pow2);
            pow2[0] = sq[0]; pow2[1] = sq[1];
            pow2[2] = sq[2]; pow2[3] = sq[3];
        }

        // g^block_start via repeated squaring
        uint64_t base[4] = {g0, g1, g2, g3};
        uint64_t result[4] = {
            Fr_params::ONE[0], Fr_params::ONE[1],
            Fr_params::ONE[2], Fr_params::ONE[3]
        };
        size_t exp = (size_t)blockIdx.x * blockDim.x;
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

    // Reconstruct g^threadIdx from binary power table
    uint64_t my_power[4] = {
        Fr_params::ONE[0], Fr_params::ONE[1],
        Fr_params::ONE[2], Fr_params::ONE[3]
    };
    unsigned t = threadIdx.x;
    #pragma unroll
    for (int bit = 0; bit < 8; bit++) {
        if ((t >> bit) & 1u) {
            uint64_t p2[4] = {sh_pow2[bit][0], sh_pow2[bit][1],
                              sh_pow2[bit][2], sh_pow2[bit][3]};
            uint64_t tmp[4];
            fr_mul(tmp, my_power, p2);
            my_power[0] = tmp[0]; my_power[1] = tmp[1];
            my_power[2] = tmp[2]; my_power[3] = tmp[3];
        }
    }

    // g^idx_a = g^block_start * g^threadIdx
    uint64_t g_a[4];
    {
        uint64_t bp[4] = {sh_power[0], sh_power[1], sh_power[2], sh_power[3]};
        fr_mul(g_a, bp, my_power);
    }

    // g^idx_b = g^idx_a * g^(n/2)
    uint64_t g_b[4];
    {
        uint64_t gh[4] = {gh0, gh1, gh2, gh3};
        fr_mul(g_b, g_a, gh);
    }

    // Load a, b and apply scale
    uint64_t a_raw[4] = {d0[idx_a], d1[idx_a], d2[idx_a], d3[idx_a]};
    uint64_t b_raw[4] = {d0[idx_b], d1[idx_b], d2[idx_b], d3[idx_b]};
    uint64_t a[4], b[4];
    fr_mul(a, a_raw, g_a);
    fr_mul(b, b_raw, g_b);

    // Load twiddle (stage 0: tw_idx = tid)
    uint64_t w[4] = {
        __ldg(tw0 + tid), __ldg(tw1 + tid),
        __ldg(tw2 + tid), __ldg(tw3 + tid)
    };

    // DIF butterfly: a' = a + b; b' = (a - b) * w
    uint64_t sum[4], diff[4], prod[4];
    fr_add(sum, a, b);
    fr_sub(diff, a, b);
    fr_mul(prod, diff, w);

    d0[idx_a] = sum[0]; d1[idx_a] = sum[1]; d2[idx_a] = sum[2]; d3[idx_a] = sum[3];
    d0[idx_b] = prod[0]; d1[idx_b] = prod[1]; d2[idx_b] = prod[2]; d3[idx_b] = prod[3];
}

// =============================================================================
// DIT butterfly kernel (one stage of inverse NTT)
//
// For stage s (0-indexed from LSB):
//   half = n >> (s+1)
//   Same indexing as DIF.
//   t = b * w;  a' = a + t;  b' = a - t
//
// Template: FUSE_SCALE multiplies outputs by scale before storing (for inv NTT).
// =============================================================================

template <bool FUSE_SCALE>
__global__ void ntt_dit_butterfly_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t num_butterflies, uint32_t half, uint32_t half_mask, uint32_t tw_stride,
    uint64_t scale0, uint64_t scale1, uint64_t scale2, uint64_t scale3)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= num_butterflies) return;

    uint32_t j = tid & half_mask;
    uint32_t group_base = tid & ~half_mask;
    uint32_t idx_a = (group_base << 1) | j;
    uint32_t idx_b = idx_a + half;
    uint32_t tw_idx = j * tw_stride;

    // Load a, b
    uint64_t a[4] = { d0[idx_a], d1[idx_a], d2[idx_a], d3[idx_a] };
    uint64_t b[4] = { d0[idx_b], d1[idx_b], d2[idx_b], d3[idx_b] };

    // Load twiddle through read-only cache path.
    uint64_t w[4] = {
        __ldg(tw0 + tw_idx), __ldg(tw1 + tw_idx),
        __ldg(tw2 + tw_idx), __ldg(tw3 + tw_idx)
    };

    // DIT butterfly: t = b * w;  a' = a + t;  b' = a - t
    uint64_t t[4], sum[4], diff[4];
    fr_mul(t, b, w);
    fr_add(sum, a, t);
    fr_sub(diff, a, t);

    if constexpr (FUSE_SCALE) {
        uint64_t sc[4] = {scale0, scale1, scale2, scale3};
        uint64_t r[4];
        fr_mul(r, sum, sc); sum[0]=r[0]; sum[1]=r[1]; sum[2]=r[2]; sum[3]=r[3];
        fr_mul(r, diff, sc); diff[0]=r[0]; diff[1]=r[1]; diff[2]=r[2]; diff[3]=r[3];
    }

    // Store
    d0[idx_a] = sum[0]; d1[idx_a] = sum[1]; d2[idx_a] = sum[2]; d3[idx_a] = sum[3];
    d0[idx_b] = diff[0]; d1[idx_b] = diff[1]; d2[idx_b] = diff[2]; d3[idx_b] = diff[3];
}

// =============================================================================
// Radix-8 DIT butterfly kernel: fuses three adjacent stages (s, s-1, s-2).
//
// Each thread processes 8 elements from one radix-8 group.
// Stage s (innermost, first):   half_s = n >> (s+1)
// Stage s-1:                    half_s1 = 2*half_s
// Stage s-2 (outermost, last):  half_s2 = 4*half_s
//
// 8 element positions in a group of 8*half_s:
//   p0 = base+j, p1 = p0+half_s, p2 = p0+2*half_s, p3 = p1+2*half_s
//   p4 = p0+4*half_s, p5 = p1+4*half_s, p6 = p2+4*half_s, p7 = p3+4*half_s
//
// Twiddle loads: 1 (stage s) + 2 (stage s-1) + 4 (stage s-2) = 7 total.
// Template: FUSE_SCALE multiplies outputs by scale before storing.
// =============================================================================

template <bool FUSE_SCALE>
__global__ void ntt_dit_radix8_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t n, int stage_s,
    uint64_t scale0, uint64_t scale1, uint64_t scale2, uint64_t scale3)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    uint32_t num_r8 = n >> 3;
    if (tid >= num_r8) return;

    uint32_t half_s  = n >> (stage_s + 1);
    uint32_t half_s1 = half_s << 1;  // 2*half_s = n >> s
    uint32_t half_s2 = half_s << 2;  // 4*half_s = n >> (s-1)

    uint32_t j     = tid & (half_s - 1);
    uint32_t group = tid >> (__ffs(half_s) - 1);

    uint32_t base = group * (8 * half_s);
    uint32_t p0 = base + j;
    uint32_t p1 = p0 + half_s;
    uint32_t p2 = p0 + half_s1;
    uint32_t p3 = p1 + half_s1;
    uint32_t p4 = p0 + half_s2;
    uint32_t p5 = p1 + half_s2;
    uint32_t p6 = p2 + half_s2;
    uint32_t p7 = p3 + half_s2;

    // Load 8 elements
    uint64_t a0[4] = { d0[p0], d1[p0], d2[p0], d3[p0] };
    uint64_t a1[4] = { d0[p1], d1[p1], d2[p1], d3[p1] };
    uint64_t a2[4] = { d0[p2], d1[p2], d2[p2], d3[p2] };
    uint64_t a3[4] = { d0[p3], d1[p3], d2[p3], d3[p3] };
    uint64_t a4[4] = { d0[p4], d1[p4], d2[p4], d3[p4] };
    uint64_t a5[4] = { d0[p5], d1[p5], d2[p5], d3[p5] };
    uint64_t a6[4] = { d0[p6], d1[p6], d2[p6], d3[p6] };
    uint64_t a7[4] = { d0[p7], d1[p7], d2[p7], d3[p7] };

    uint32_t tw_stride_s  = 1u << stage_s;
    uint32_t tw_stride_s1 = tw_stride_s >> 1;  // 1 << (s-1)
    uint32_t tw_stride_s2 = tw_stride_s >> 2;  // 1 << (s-2)

    uint64_t w[4], t[4], sum[4], diff[4];
    uint32_t twi;

    // --- Stage s: 4 DIT butterflies at distance half_s ---
    // Pairs: (a0,a1), (a2,a3), (a4,a5), (a6,a7) — all same twiddle
    twi = j * tw_stride_s;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);

    fr_mul(t, a1, w); fr_add(sum, a0, t); fr_sub(diff, a0, t);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];
    a1[0]=diff[0]; a1[1]=diff[1]; a1[2]=diff[2]; a1[3]=diff[3];

    fr_mul(t, a3, w); fr_add(sum, a2, t); fr_sub(diff, a2, t);
    a2[0]=sum[0]; a2[1]=sum[1]; a2[2]=sum[2]; a2[3]=sum[3];
    a3[0]=diff[0]; a3[1]=diff[1]; a3[2]=diff[2]; a3[3]=diff[3];

    fr_mul(t, a5, w); fr_add(sum, a4, t); fr_sub(diff, a4, t);
    a4[0]=sum[0]; a4[1]=sum[1]; a4[2]=sum[2]; a4[3]=sum[3];
    a5[0]=diff[0]; a5[1]=diff[1]; a5[2]=diff[2]; a5[3]=diff[3];

    fr_mul(t, a7, w); fr_add(sum, a6, t); fr_sub(diff, a6, t);
    a6[0]=sum[0]; a6[1]=sum[1]; a6[2]=sum[2]; a6[3]=sum[3];
    a7[0]=diff[0]; a7[1]=diff[1]; a7[2]=diff[2]; a7[3]=diff[3];

    // --- Stage s-1: 4 DIT butterflies at distance 2*half_s ---
    // Pairs: (a0,a2), (a1,a3), (a4,a6), (a5,a7)
    uint64_t ws1_a[4], ws1_b[4];
    twi = j * tw_stride_s1;
    ws1_a[0] = __ldg(tw0+twi); ws1_a[1] = __ldg(tw1+twi);
    ws1_a[2] = __ldg(tw2+twi); ws1_a[3] = __ldg(tw3+twi);
    twi = (j + half_s) * tw_stride_s1;
    ws1_b[0] = __ldg(tw0+twi); ws1_b[1] = __ldg(tw1+twi);
    ws1_b[2] = __ldg(tw2+twi); ws1_b[3] = __ldg(tw3+twi);

    fr_mul(t, a2, ws1_a); fr_add(sum, a0, t); fr_sub(diff, a0, t);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];
    a2[0]=diff[0]; a2[1]=diff[1]; a2[2]=diff[2]; a2[3]=diff[3];

    fr_mul(t, a3, ws1_b); fr_add(sum, a1, t); fr_sub(diff, a1, t);
    a1[0]=sum[0]; a1[1]=sum[1]; a1[2]=sum[2]; a1[3]=sum[3];
    a3[0]=diff[0]; a3[1]=diff[1]; a3[2]=diff[2]; a3[3]=diff[3];

    fr_mul(t, a6, ws1_a); fr_add(sum, a4, t); fr_sub(diff, a4, t);
    a4[0]=sum[0]; a4[1]=sum[1]; a4[2]=sum[2]; a4[3]=sum[3];
    a6[0]=diff[0]; a6[1]=diff[1]; a6[2]=diff[2]; a6[3]=diff[3];

    fr_mul(t, a7, ws1_b); fr_add(sum, a5, t); fr_sub(diff, a5, t);
    a5[0]=sum[0]; a5[1]=sum[1]; a5[2]=sum[2]; a5[3]=sum[3];
    a7[0]=diff[0]; a7[1]=diff[1]; a7[2]=diff[2]; a7[3]=diff[3];

    // --- Stage s-2: 4 DIT butterflies at distance 4*half_s ---
    // Pairs: (a0,a4), (a1,a5), (a2,a6), (a3,a7) — 4 different twiddles
    twi = j * tw_stride_s2;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_mul(t, a4, w); fr_add(sum, a0, t); fr_sub(diff, a0, t);
    a0[0]=sum[0]; a0[1]=sum[1]; a0[2]=sum[2]; a0[3]=sum[3];
    a4[0]=diff[0]; a4[1]=diff[1]; a4[2]=diff[2]; a4[3]=diff[3];

    twi = (j + half_s) * tw_stride_s2;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_mul(t, a5, w); fr_add(sum, a1, t); fr_sub(diff, a1, t);
    a1[0]=sum[0]; a1[1]=sum[1]; a1[2]=sum[2]; a1[3]=sum[3];
    a5[0]=diff[0]; a5[1]=diff[1]; a5[2]=diff[2]; a5[3]=diff[3];

    twi = (j + half_s1) * tw_stride_s2;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_mul(t, a6, w); fr_add(sum, a2, t); fr_sub(diff, a2, t);
    a2[0]=sum[0]; a2[1]=sum[1]; a2[2]=sum[2]; a2[3]=sum[3];
    a6[0]=diff[0]; a6[1]=diff[1]; a6[2]=diff[2]; a6[3]=diff[3];

    twi = (j + half_s1 + half_s) * tw_stride_s2;
    w[0] = __ldg(tw0+twi); w[1] = __ldg(tw1+twi);
    w[2] = __ldg(tw2+twi); w[3] = __ldg(tw3+twi);
    fr_mul(t, a7, w); fr_add(sum, a3, t); fr_sub(diff, a3, t);
    a3[0]=sum[0]; a3[1]=sum[1]; a3[2]=sum[2]; a3[3]=sum[3];
    a7[0]=diff[0]; a7[1]=diff[1]; a7[2]=diff[2]; a7[3]=diff[3];

    // Optional fused scale (1/n for inverse NTT)
    if constexpr (FUSE_SCALE) {
        uint64_t sc[4] = {scale0, scale1, scale2, scale3};
        uint64_t r[4];
        fr_mul(r, a0, sc); a0[0]=r[0]; a0[1]=r[1]; a0[2]=r[2]; a0[3]=r[3];
        fr_mul(r, a1, sc); a1[0]=r[0]; a1[1]=r[1]; a1[2]=r[2]; a1[3]=r[3];
        fr_mul(r, a2, sc); a2[0]=r[0]; a2[1]=r[1]; a2[2]=r[2]; a2[3]=r[3];
        fr_mul(r, a3, sc); a3[0]=r[0]; a3[1]=r[1]; a3[2]=r[2]; a3[3]=r[3];
        fr_mul(r, a4, sc); a4[0]=r[0]; a4[1]=r[1]; a4[2]=r[2]; a4[3]=r[3];
        fr_mul(r, a5, sc); a5[0]=r[0]; a5[1]=r[1]; a5[2]=r[2]; a5[3]=r[3];
        fr_mul(r, a6, sc); a6[0]=r[0]; a6[1]=r[1]; a6[2]=r[2]; a6[3]=r[3];
        fr_mul(r, a7, sc); a7[0]=r[0]; a7[1]=r[1]; a7[2]=r[2]; a7[3]=r[3];
    }

    // Store
    d0[p0]=a0[0]; d1[p0]=a0[1]; d2[p0]=a0[2]; d3[p0]=a0[3];
    d0[p1]=a1[0]; d1[p1]=a1[1]; d2[p1]=a1[2]; d3[p1]=a1[3];
    d0[p2]=a2[0]; d1[p2]=a2[1]; d2[p2]=a2[2]; d3[p2]=a2[3];
    d0[p3]=a3[0]; d1[p3]=a3[1]; d2[p3]=a3[2]; d3[p3]=a3[3];
    d0[p4]=a4[0]; d1[p4]=a4[1]; d2[p4]=a4[2]; d3[p4]=a4[3];
    d0[p5]=a5[0]; d1[p5]=a5[1]; d2[p5]=a5[2]; d3[p5]=a5[3];
    d0[p6]=a6[0]; d1[p6]=a6[1]; d2[p6]=a6[2]; d3[p6]=a6[3];
    d0[p7]=a7[0]; d1[p7]=a7[1]; d2[p7]=a7[2]; d3[p7]=a7[3];
}

// =============================================================================
// DIT fused tail kernel: processes the FIRST TAIL_LOG stages in shared memory.
//
// DIT runs stages from highest (s = log_n-1) down to lowest (s = 0).
// The fused tail handles the highest stages (largest s, smallest butterfly span)
// which fit in shared memory. This is the OPPOSITE of DIF tail which handles
// the lowest stages.
//
// For n = 2²⁰ and TAIL_LOG = 11:
//   - Stages 19 down to 9 (11 stages) run first in shared memory
//   - Then radix-8/4/2 kernels handle stages 8 down to 0
//
// The stage iteration goes: stage_start, stage_start-1, ..., stage_start-TAIL_LOG+1
// (highest s first = smallest half = smallest butterfly span = fits in shared memory)
// =============================================================================

template <int TAIL_LOG>
__global__ void __launch_bounds__(1024, 1) ntt_dit_tail_fused_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t *__restrict__ tw0, const uint64_t *__restrict__ tw1,
    const uint64_t *__restrict__ tw2, const uint64_t *__restrict__ tw3,
    uint32_t n, int stage_start)  // stage_start = TAIL_LOG-1 (highest of the tail stages)
{
    constexpr uint32_t span = 1u << TAIL_LOG;
    constexpr uint32_t butterflies_per_chunk = span >> 1;

    uint32_t chunk = (uint32_t)blockIdx.x;
    uint32_t base = chunk * span;
    uint32_t t = threadIdx.x;
    uint32_t P = blockDim.x;

    extern __shared__ uint64_t shmem[];
    uint64_t *s0 = shmem;
    uint64_t *s1 = s0 + span;
    uint64_t *s2 = s1 + span;
    uint64_t *s3 = s2 + span;

    // Load: each thread handles span/P elements
    for (uint32_t i = t; i < span; i += P) {
        uint32_t global_idx = base + i;
        if (global_idx < n) {
            s0[i] = d0[global_idx];
            s1[i] = d1[global_idx];
            s2[i] = d2[global_idx];
            s3[i] = d3[global_idx];
        }
    }
    __syncthreads();

    // DIT: process stages from stage_start down to 0
    #pragma unroll
    for (int st = 0; st < TAIL_LOG; st++) {
        int s = stage_start - st;  // stages: stage_start, stage_start-1, ..., 0
        uint32_t half = n >> (s + 1);
        uint32_t half_mask = half - 1;
        uint32_t tw_stride = 1u << s;

        for (uint32_t bt = t; bt < butterflies_per_chunk; bt += P) {
            uint32_t j = bt & half_mask;
            uint32_t group_base = bt & ~half_mask;
            uint32_t idx_a = (group_base << 1) | j;
            uint32_t idx_b = idx_a + half;
            uint32_t tw_idx = j * tw_stride;

            uint64_t a[4] = { s0[idx_a], s1[idx_a], s2[idx_a], s3[idx_a] };
            uint64_t b[4] = { s0[idx_b], s1[idx_b], s2[idx_b], s3[idx_b] };
            uint64_t w[4] = {
                __ldg(tw0 + tw_idx), __ldg(tw1 + tw_idx),
                __ldg(tw2 + tw_idx), __ldg(tw3 + tw_idx)
            };

            // DIT butterfly: t = b * w; a' = a + t; b' = a - t
            uint64_t tw_b[4], sum[4], diff[4];
            fr_mul(tw_b, b, w);
            fr_add(sum, a, tw_b);
            fr_sub(diff, a, tw_b);

            s0[idx_a] = sum[0]; s1[idx_a] = sum[1]; s2[idx_a] = sum[2]; s3[idx_a] = sum[3];
            s0[idx_b] = diff[0]; s1[idx_b] = diff[1]; s2[idx_b] = diff[2]; s3[idx_b] = diff[3];
        }
        __syncthreads();
    }

    // Store: each thread handles span/P elements
    for (uint32_t i = t; i < span; i += P) {
        uint32_t global_idx = base + i;
        if (global_idx < n) {
            d0[global_idx] = s0[i];
            d1[global_idx] = s1[i];
            d2[global_idx] = s2[i];
            d3[global_idx] = s3[i];
        }
    }
}

// =============================================================================
// Bit-reversal permutation kernel (in-place)
// Each thread handles one swap where i < bit_reverse(i).
// =============================================================================

__device__ __forceinline__ uint32_t bit_reverse(uint32_t x, int log_n) {
    x = __brev(x);
    x >>= (32 - log_n);
    return x;
}

__global__ void ntt_bit_reverse_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    size_t n, int log_n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    size_t j = bit_reverse((uint32_t)i, log_n);
    if (i >= j) return; // only swap once per pair

    // Swap elements i and j
    uint64_t tmp;

    tmp = d0[i]; d0[i] = d0[j]; d0[j] = tmp;
    tmp = d1[i]; d1[i] = d1[j]; d1[j] = tmp;
    tmp = d2[i]; d2[i] = d2[j]; d2[j] = tmp;
    tmp = d3[i]; d3[i] = d3[j]; d3[j] = tmp;
}

// =============================================================================
// Scale kernel: multiply all elements by a constant (1/n for inverse NTT)
// Kept as fallback for edge cases where fused scale cannot be applied.
// =============================================================================

__global__ void ntt_scale_kernel(
    uint64_t *__restrict__ d0, uint64_t *__restrict__ d1,
    uint64_t *__restrict__ d2, uint64_t *__restrict__ d3,
    const uint64_t inv_n0, const uint64_t inv_n1,
    const uint64_t inv_n2, const uint64_t inv_n3,
    size_t n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    uint64_t a[4] = { d0[i], d1[i], d2[i], d3[i] };
    uint64_t c[4] = { inv_n0, inv_n1, inv_n2, inv_n3 };
    uint64_t r[4];
    fr_mul(r, a, c);

    d0[i] = r[0]; d1[i] = r[1]; d2[i] = r[2]; d3[i] = r[3];
}

// =============================================================================
// NTT Domain: holds twiddle factors on GPU
// =============================================================================

struct NTTDomain {
    size_t size;
    int log_size;
    int tail_log;  // adaptive: 11 or 12 based on GPU shared memory capacity
    // SoA twiddle arrays, each n/2 elements
    uint64_t *d_twiddles_fwd[4];
    uint64_t *d_twiddles_inv[4];
    // 1/n in Montgomery form
    uint64_t inv_n[4];
};

// =============================================================================
// Host-side functions
// =============================================================================

NTTDomain *ntt_domain_create(size_t size, const uint64_t *fwd_twiddles_aos,
                              const uint64_t *inv_twiddles_aos, const uint64_t inv_n[4],
                              cudaStream_t stream) {
    NTTDomain *dom = new NTTDomain;
    dom->size = size;

    // Compute log2(size)
    int log_n = 0;
    { size_t tmp = size; while (tmp > 1) { tmp >>= 1; log_n++; } }
    dom->log_size = log_n;

    // Query max shared memory per block for adaptive tail sizing
    int max_shmem = 0;
    cudaDeviceGetAttribute(&max_shmem, cudaDevAttrMaxSharedMemoryPerBlockOptin, 0);
    // tail_log=12 needs 4 * 4096 * 8 = 131072 bytes of shared memory
    // tail_log=11 needs 4 * 2048 * 8 = 65536 bytes (fits in 99KB optin on Blackwell)
    dom->tail_log = (max_shmem >= 131072 && log_n >= 12) ? 12 : 11;

    // Copy inv_n
    for (int i = 0; i < 4; i++) dom->inv_n[i] = inv_n[i];

    size_t half_n = size / 2;

    // Allocate device twiddle arrays (SoA)
    for (int i = 0; i < 4; i++) {
        cudaMalloc(&dom->d_twiddles_fwd[i], half_n * sizeof(uint64_t));
        cudaMalloc(&dom->d_twiddles_inv[i], half_n * sizeof(uint64_t));
    }

    if (half_n > 0) {
        uint64_t *tw_aos = nullptr;
        cudaMalloc(&tw_aos, half_n * 4 * sizeof(uint64_t));

        // Forward twiddles: copy AoS once, transpose on GPU.
        cudaMemcpyAsync(tw_aos, fwd_twiddles_aos, half_n * 4 * sizeof(uint64_t),
                        cudaMemcpyHostToDevice, stream);
        launch_transpose_aos_to_soa_fr(dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
                                       dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
                                       tw_aos, half_n, stream);
        // Reuse tw_aos for inverse twiddles only after forward transpose is done.
        cudaStreamSynchronize(stream);

        // Inverse twiddles: copy AoS once, transpose on GPU.
        cudaMemcpyAsync(tw_aos, inv_twiddles_aos, half_n * 4 * sizeof(uint64_t),
                        cudaMemcpyHostToDevice, stream);
        launch_transpose_aos_to_soa_fr(dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
                                       dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
                                       tw_aos, half_n, stream);

        cudaStreamSynchronize(stream);
        cudaFree(tw_aos);
    }

    return dom;
}

void ntt_domain_destroy(NTTDomain *dom) {
    if (!dom) return;
    for (int i = 0; i < 4; i++) {
        cudaFree(dom->d_twiddles_fwd[i]);
        cudaFree(dom->d_twiddles_inv[i]);
    }
    delete dom;
}

// =============================================================================
// Helper: launch DIF fused tail (dispatches tail_log=11 or 12)
// =============================================================================

static void launch_dif_tail(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                            uint32_t n, int stage_start, cudaStream_t stream) {
    int tail_log = dom->tail_log;
    uint32_t span = 1u << tail_log;
    unsigned tail_threads = (span > 1024) ? 1024 : span;
    unsigned tail_blocks = (n + span - 1) / span;
    size_t shmem_bytes = 4ull * span * sizeof(uint64_t);
    if (tail_log == 12) {
        cudaFuncSetAttribute(ntt_dif_tail_fused_kernel<12>,
                             cudaFuncAttributeMaxDynamicSharedMemorySize, shmem_bytes);
        ntt_dif_tail_fused_kernel<12><<<tail_blocks, tail_threads, shmem_bytes, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            n, stage_start);
    } else {
        cudaFuncSetAttribute(ntt_dif_tail_fused_kernel<11>,
                             cudaFuncAttributeMaxDynamicSharedMemorySize, shmem_bytes);
        ntt_dif_tail_fused_kernel<11><<<tail_blocks, tail_threads, shmem_bytes, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            n, stage_start);
    }
}

static void launch_dit_tail(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                            uint32_t n, int stage_start, cudaStream_t stream) {
    int tail_log = dom->tail_log;
    uint32_t span = 1u << tail_log;
    unsigned tail_threads = (span > 1024) ? 1024 : span;
    unsigned tail_blocks = (n + span - 1) / span;
    size_t shmem_bytes = 4ull * span * sizeof(uint64_t);

    if (tail_log == 12) {
        cudaFuncSetAttribute(ntt_dit_tail_fused_kernel<12>,
                             cudaFuncAttributeMaxDynamicSharedMemorySize, shmem_bytes);
        ntt_dit_tail_fused_kernel<12><<<tail_blocks, tail_threads, shmem_bytes, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
            dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
            n, stage_start);
    } else {
        cudaFuncSetAttribute(ntt_dit_tail_fused_kernel<11>,
                             cudaFuncAttributeMaxDynamicSharedMemorySize, shmem_bytes);
        ntt_dit_tail_fused_kernel<11><<<tail_blocks, tail_threads, shmem_bytes, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
            dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
            n, stage_start);
    }
}

static bool should_use_fused_tail(const NTTDomain *dom, uint32_t n) {
    return (dom->log_size > dom->tail_log) && (n >= NTT_FUSED_TAIL_MIN_N);
}

// =============================================================================
// Forward NTT (DIF): radix-8 → radix-4 → radix-2 → fused tail
// =============================================================================

void launch_ntt_forward(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                        cudaStream_t stream) {
    uint32_t n = (uint32_t)dom->size;
    uint32_t num_butterflies = n >> 1;
    unsigned blocks_r2 = (num_butterflies + NTT_THREADS - 1) / NTT_THREADS;
    uint32_t num_r8 = n >> 3;
    unsigned blocks_r8 = (num_r8 + NTT_THREADS - 1) / NTT_THREADS;

    bool use_fused_tail = should_use_fused_tail(dom, n);
    int regular_stages = dom->log_size;
    if (use_fused_tail) {
        regular_stages = dom->log_size - dom->tail_log;
    }

    // DIF: radix-8 for 3 stages at a time, radix-2 for remainder
    int s = 0;
    for (; s + 2 < regular_stages; s += 3) {
        ntt_dif_radix8_kernel<<<blocks_r8, NTT_THREADS, 0, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            n, s);
    }
    for (; s < regular_stages; s++) {
        uint32_t half = n >> (s + 1);
        uint32_t half_mask = half - 1;
        uint32_t tw_stride = 1u << s;
        ntt_dif_butterfly_kernel<<<blocks_r2, NTT_THREADS, 0, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            num_butterflies, half, half_mask, tw_stride);
    }

    if (use_fused_tail) {
        launch_dif_tail(dom, d0, d1, d2, d3, n, regular_stages, stream);
    }
}

// =============================================================================
// Fused CosetFFT forward: ScaleByPowers + DIF NTT in one pass.
// Stage 0 uses fused scale kernel, then radix-8/4/2 for remaining stages.
// =============================================================================

void launch_ntt_forward_coset(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                               const uint64_t g[4], const uint64_t g_half[4],
                               cudaStream_t stream) {
    uint32_t n = (uint32_t)dom->size;
    uint32_t num_butterflies = n >> 1;
    unsigned blocks_r2 = (num_butterflies + NTT_THREADS - 1) / NTT_THREADS;
    uint32_t num_r8 = n >> 3;
    unsigned blocks_r8 = (num_r8 + NTT_THREADS - 1) / NTT_THREADS;

    bool use_fused_tail = should_use_fused_tail(dom, n);
    int regular_stages = dom->log_size;
    if (use_fused_tail) {
        regular_stages = dom->log_size - dom->tail_log;
    }

    // Stage 0: fused ScaleByPowers + DIF butterfly
    ntt_dif_first_stage_fused_scale_kernel<<<blocks_r2, NTT_THREADS, 0, stream>>>(
        d0, d1, d2, d3,
        dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
        dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
        g[0], g[1], g[2], g[3],
        g_half[0], g_half[1], g_half[2], g_half[3],
        num_butterflies);

    // Stages 1+: radix-8 for 3 stages at a time, radix-2 for remainder
    int s = 1;
    for (; s + 2 < regular_stages; s += 3) {
        ntt_dif_radix8_kernel<<<blocks_r8, NTT_THREADS, 0, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            n, s);
    }
    for (; s < regular_stages; s++) {
        uint32_t half = n >> (s + 1);
        uint32_t half_mask = half - 1;
        uint32_t tw_stride = 1u << s;
        ntt_dif_butterfly_kernel<<<blocks_r2, NTT_THREADS, 0, stream>>>(
            d0, d1, d2, d3,
            dom->d_twiddles_fwd[0], dom->d_twiddles_fwd[1],
            dom->d_twiddles_fwd[2], dom->d_twiddles_fwd[3],
            num_butterflies, half, half_mask, tw_stride);
    }

    if (use_fused_tail) {
        launch_dif_tail(dom, d0, d1, d2, d3, n, regular_stages, stream);
    }
}

// =============================================================================
// Inverse NTT (DIT): fused tail → radix-8 → radix-4 → radix-2
// Scale by 1/n is fused into the last kernel launch.
// =============================================================================

void launch_ntt_inverse(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                        cudaStream_t stream) {
    uint32_t n = (uint32_t)dom->size;
    uint32_t num_butterflies = n >> 1;
    unsigned blocks_r2 = (num_butterflies + NTT_THREADS - 1) / NTT_THREADS;
    uint32_t num_r8 = n >> 3;
    unsigned blocks_r8 = (num_r8 + NTT_THREADS - 1) / NTT_THREADS;

    bool use_fused_tail = should_use_fused_tail(dom, n);

    // DIT fused tail FIRST: stages log_n-1 down to log_n-tail_log
    int first_regular;
    if (use_fused_tail) {
        launch_dit_tail(dom, d0, d1, d2, d3, n, dom->log_size - 1, stream);
        first_regular = dom->log_size - dom->tail_log - 1;
    } else {
        first_regular = dom->log_size - 1;
    }

    // Regular stages: from first_regular down to 0.
    // Radix-8 for 3 stages at a time, radix-2 for remainder.
    // The last kernel fuses the 1/n scale.
    bool scaled = false;
    int s = first_regular;

    // DIT radix-8: fuses stages (s, s-1, s-2)
    for (; s - 2 >= 0; s -= 3) {
        if (s < 3) {
            // This is the last kernel — fuse scale
            ntt_dit_radix8_kernel<true><<<blocks_r8, NTT_THREADS, 0, stream>>>(
                d0, d1, d2, d3,
                dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
                dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
                n, s,
                dom->inv_n[0], dom->inv_n[1], dom->inv_n[2], dom->inv_n[3]);
            scaled = true;
        } else {
            ntt_dit_radix8_kernel<false><<<blocks_r8, NTT_THREADS, 0, stream>>>(
                d0, d1, d2, d3,
                dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
                dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
                n, s,
                0, 0, 0, 0);
        }
    }

    // DIT radix-2 for remaining stages (0, 1, or 2 stages)
    for (; s >= 0; s--) {
        uint32_t half = n >> (s + 1);
        uint32_t half_mask = half - 1;
        uint32_t tw_stride = 1u << s;
        if (s == 0) {
            // Last stage — fuse 1/n scale
            ntt_dit_butterfly_kernel<true><<<blocks_r2, NTT_THREADS, 0, stream>>>(
                d0, d1, d2, d3,
                dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
                dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
                num_butterflies, half, half_mask, tw_stride,
                dom->inv_n[0], dom->inv_n[1], dom->inv_n[2], dom->inv_n[3]);
            scaled = true;
        } else {
            ntt_dit_butterfly_kernel<false><<<blocks_r2, NTT_THREADS, 0, stream>>>(
                d0, d1, d2, d3,
                dom->d_twiddles_inv[0], dom->d_twiddles_inv[1],
                dom->d_twiddles_inv[2], dom->d_twiddles_inv[3],
                num_butterflies, half, half_mask, tw_stride,
                0, 0, 0, 0);
        }
    }

    // Fallback: separate scale kernel (only for edge cases like n=1)
    if (!scaled) {
        unsigned blocks_n = (n + NTT_THREADS - 1) / NTT_THREADS;
        ntt_scale_kernel<<<blocks_n, NTT_THREADS, 0, stream>>>(
            d0, d1, d2, d3,
            dom->inv_n[0], dom->inv_n[1], dom->inv_n[2], dom->inv_n[3],
            n);
    }
}

void launch_ntt_bit_reverse(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                            cudaStream_t stream) {
    size_t n = dom->size;
    constexpr unsigned threads = 256;
    unsigned blocks = (n + threads - 1) / threads;
    ntt_bit_reverse_kernel<<<blocks, threads, 0, stream>>>(d0, d1, d2, d3, n, dom->log_size);
}

// Accessor: get forward twiddle pointers (used by PlonK constraint kernel)
void ntt_get_fwd_twiddles(const NTTDomain *dom, const uint64_t **out_ptrs) {
    for (int i = 0; i < 4; i++) {
        out_ptrs[i] = dom->d_twiddles_fwd[i];
    }
}

} // namespace gnark_gpu
