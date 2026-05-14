// ═══════════════════════════════════════════════════════════════════════════════
// GPU chunked Horner polynomial evaluation at a single point z
//
// Evaluates p(z) = c₀ + c₁z + c₂z² + ... + cₙ₋₁zⁿ⁻¹
//
// Strategy: Divide n coefficients into K chunks of 1024, each thread evaluates
// its chunk independently via Horner's method, then CPU combines.
//
//   Chunk j evaluates: partial[j] = c[jK] + z·(c[jK+1] + z·(... + z·c[(j+1)K-1]))
//
//   Full result: p(z) = partial[0] + z^K · partial[1] + z^(2K) · partial[2] + ...
//                      = Σⱼ partial[j] · z^(jK)
//
// The CPU computes z^K once, then combines via Horner on the partial results:
//   p(z) = partial[0] + z^K · (partial[1] + z^K · (partial[2] + ...))
//
// For n = 2²⁷: 131072 chunks × 1024 coeffs/chunk × 1023 fr_mul/chunk
// Each thread does 1023 sequential fr_muls — heavily compute-bound, ideal for GPU.
// ═══════════════════════════════════════════════════════════════════════════════

#include "fr_arith.cuh"
#include <cuda_runtime.h>

namespace gnark_gpu {

constexpr int EVAL_CHUNK_SIZE = 1024;

// Each thread evaluates one chunk of coefficients via Horner's method.
// chunk j computes: partial[j] = c[j*K] + z*(c[j*K+1] + z*(...+ z*c[(j+1)*K-1]))
// where K = EVAL_CHUNK_SIZE.
// The full polynomial is recovered as: p(z) = Σ_j partial[j] * z^(j*K).
__global__ void poly_eval_chunks_kernel(
    const uint64_t *__restrict__ c0,
    const uint64_t *__restrict__ c1,
    const uint64_t *__restrict__ c2,
    const uint64_t *__restrict__ c3,
    const uint64_t z0, const uint64_t z1,
    const uint64_t z2, const uint64_t z3,
    uint64_t *__restrict__ out0,
    uint64_t *__restrict__ out1,
    uint64_t *__restrict__ out2,
    uint64_t *__restrict__ out3,
    size_t n)
{
    size_t chunk = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t start = chunk * EVAL_CHUNK_SIZE;
    if (start >= n) return;

    size_t end = start + EVAL_CHUNK_SIZE;
    if (end > n) end = n;

    uint64_t z[4] = {z0, z1, z2, z3};

    // Horner: result = c[end-1]; for i = end-2 downto start: result = result*z + c[i]
    size_t last = end - 1;
    uint64_t r[4] = {c0[last], c1[last], c2[last], c3[last]};

    for (size_t i = last; i > start; ) {
        --i;
        uint64_t t[4];
        fr_mul(t, r, z);
        uint64_t ci[4] = {c0[i], c1[i], c2[i], c3[i]};
        fr_add(r, t, ci);
    }

    out0[chunk] = r[0];
    out1[chunk] = r[1];
    out2[chunk] = r[2];
    out3[chunk] = r[3];
}

// Launch the chunked Horner evaluation kernel.
// out_partials must be pre-allocated device memory for (num_chunks) elements in SoA.
// Returns the number of chunks in *num_chunks_out.
void launch_poly_eval_chunks(
    const uint64_t *c0, const uint64_t *c1,
    const uint64_t *c2, const uint64_t *c3,
    const uint64_t z[4],
    uint64_t *out0, uint64_t *out1,
    uint64_t *out2, uint64_t *out3,
    size_t n, size_t *num_chunks_out,
    cudaStream_t stream)
{
    size_t nc = (n + EVAL_CHUNK_SIZE - 1) / EVAL_CHUNK_SIZE;
    *num_chunks_out = nc;

    constexpr unsigned threads = 256;
    unsigned blocks = (nc + threads - 1) / threads;
    poly_eval_chunks_kernel<<<blocks, threads, 0, stream>>>(
        c0, c1, c2, c3,
        z[0], z[1], z[2], z[3],
        out0, out1, out2, out3, n);
}

} // namespace gnark_gpu
