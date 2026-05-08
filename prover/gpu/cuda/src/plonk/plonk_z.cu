// ═══════════════════════════════════════════════════════════════════════════════
// GPU Z-polynomial prefix product for PlonK permutation argument
//
// Computes Z[i] = Π_{k=0}^{i-1} ratio[k]  (prefix product of ratio vector)
//
// The Z polynomial encodes the permutation argument in PlonK:
//   ratio[i] = (L[i]+β·ω^i+γ)(R[i]+β·k₁·ω^i+γ)(O[i]+β·k₂·ω^i+γ)
//              ─────────────────────────────────────────────────────
//              (L[i]+β·S₁(ω^i)+γ)(R[i]+β·S₂(ω^i)+γ)(O[i]+β·S₃(ω^i)+γ)
//
// Three-phase parallel scan with GPU/CPU hybrid:
//
//   ratio: [r₀ r₁ r₂ r₃ | r₄ r₅ r₆ r₇ | r₈ r₉ ...]
//           ─── chunk 0 ──  ─── chunk 1 ──  ─ chunk 2 ─
//
//   Phase 1 (GPU): Local prefix product within each chunk of 1024 elements.
//     chunk 0: [r₀, r₀r₁, r₀r₁r₂, r₀r₁r₂r₃]
//     chunk 1: [r₄, r₄r₅, r₄r₅r₆, r₄r₅r₆r₇]
//     → chunk_products: [r₀r₁r₂r₃, r₄r₅r₆r₇, ...]
//
//   Phase 2 (CPU): Sequential scan of ~n/1024 chunk products.
//     scanned_prefix[0] = cp[0]
//     scanned_prefix[1] = cp[0] · cp[1]
//     → At most ~131K elements for n=2²⁷, trivial on CPU.
//
//   Phase 3 (GPU): Multiply each chunk's elements by its global prefix.
//     chunk 1: each elem *= scanned_prefix[0]
//     chunk 2: each elem *= scanned_prefix[1]
//     Then shift right by 1: Z[0] = 1, Z[i] = prefix_product[i-1]
// ═══════════════════════════════════════════════════════════════════════════════

#include "fr_arith.cuh"
#include <cuda_runtime.h>

namespace gnark_gpu {

constexpr size_t Z_CHUNK_SIZE = 1024;

// Phase 1: Each thread processes one chunk of Z_CHUNK_SIZE ratios.
// Computes prefix products within the chunk in-place.
// z_out[chunk_start] = ratio[chunk_start]
// z_out[chunk_start+1] = ratio[chunk_start] * ratio[chunk_start+1]
// ...
// chunk_products[chunk_id] = product of all ratios in this chunk
__global__ void z_prefix_local_kernel(
    uint64_t *__restrict__ z0, uint64_t *__restrict__ z1,
    uint64_t *__restrict__ z2, uint64_t *__restrict__ z3,
    const uint64_t *__restrict__ r0, const uint64_t *__restrict__ r1,
    const uint64_t *__restrict__ r2, const uint64_t *__restrict__ r3,
    uint64_t *__restrict__ cp0, uint64_t *__restrict__ cp1,
    uint64_t *__restrict__ cp2, uint64_t *__restrict__ cp3,
    size_t n)
{
    size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t num_chunks = (n + Z_CHUNK_SIZE - 1) / Z_CHUNK_SIZE;
    if (chunk_id >= num_chunks) return;

    size_t start = chunk_id * Z_CHUNK_SIZE;
    size_t end = start + Z_CHUNK_SIZE;
    if (end > n) end = n;

    // Prefix product: z[start] = r[start], z[i] = z[i-1] * r[i]
    uint64_t acc[4] = {r0[start], r1[start], r2[start], r3[start]};
    z0[start] = acc[0]; z1[start] = acc[1]; z2[start] = acc[2]; z3[start] = acc[3];

    for (size_t i = start + 1; i < end; i++) {
        uint64_t elem[4] = {r0[i], r1[i], r2[i], r3[i]};
        uint64_t prod[4];
        fr_mul(prod, acc, elem);
        acc[0] = prod[0]; acc[1] = prod[1];
        acc[2] = prod[2]; acc[3] = prod[3];
        z0[i] = acc[0]; z1[i] = acc[1]; z2[i] = acc[2]; z3[i] = acc[3];
    }

    // Store chunk product
    cp0[chunk_id] = acc[0]; cp1[chunk_id] = acc[1];
    cp2[chunk_id] = acc[2]; cp3[chunk_id] = acc[3];
}

// Phase 3: Apply global prefix fixup.
// For chunk k > 0: z[i] *= scanned_prefix[k-1]
__global__ void z_prefix_fixup_kernel(
    uint64_t *__restrict__ z0, uint64_t *__restrict__ z1,
    uint64_t *__restrict__ z2, uint64_t *__restrict__ z3,
    const uint64_t *__restrict__ sp0, const uint64_t *__restrict__ sp1,
    const uint64_t *__restrict__ sp2, const uint64_t *__restrict__ sp3,
    size_t n)
{
    size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t num_chunks = (n + Z_CHUNK_SIZE - 1) / Z_CHUNK_SIZE;
    if (chunk_id == 0 || chunk_id >= num_chunks) return;

    size_t start = chunk_id * Z_CHUNK_SIZE;
    size_t end = start + Z_CHUNK_SIZE;
    if (end > n) end = n;

    uint64_t prefix[4] = {sp0[chunk_id-1], sp1[chunk_id-1],
                           sp2[chunk_id-1], sp3[chunk_id-1]};

    for (size_t i = start; i < end; i++) {
        uint64_t elem[4] = {z0[i], z1[i], z2[i], z3[i]};
        uint64_t prod[4];
        fr_mul(prod, prefix, elem);
        z0[i] = prod[0]; z1[i] = prod[1]; z2[i] = prod[2]; z3[i] = prod[3];
    }
}

// Shift right by 1: z[i] = z[i-1] for i > 0, z[0] = Montgomery 1.
// After this, z[i] = product(ratio[0..i-1]) which is the Z polynomial.
__global__ void z_shift_right_kernel(
    uint64_t *__restrict__ z0, uint64_t *__restrict__ z1,
    uint64_t *__restrict__ z2, uint64_t *__restrict__ z3,
    const uint64_t *__restrict__ src0, const uint64_t *__restrict__ src1,
    const uint64_t *__restrict__ src2, const uint64_t *__restrict__ src3,
    size_t n)
{
    size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    if (i == 0) {
        z0[0] = Fr_params::ONE[0]; z1[0] = Fr_params::ONE[1];
        z2[0] = Fr_params::ONE[2]; z3[0] = Fr_params::ONE[3];
    } else {
        z0[i] = src0[i-1]; z1[i] = src1[i-1];
        z2[i] = src2[i-1]; z3[i] = src3[i-1];
    }
}

// Phase 1 launch: requires caller to provide cp[4] device arrays.
cudaError_t launch_z_prefix_phase1(
    uint64_t *z0, uint64_t *z1, uint64_t *z2, uint64_t *z3,
    const uint64_t *r0, const uint64_t *r1, const uint64_t *r2, const uint64_t *r3,
    uint64_t *cp[4],
    size_t n, cudaStream_t stream)
{
    if (n == 0) return cudaSuccess;

    size_t num_chunks = (n + Z_CHUNK_SIZE - 1) / Z_CHUNK_SIZE;
    constexpr unsigned threads = 256;
    unsigned blocks = (num_chunks + threads - 1) / threads;

    z_prefix_local_kernel<<<blocks, threads, 0, stream>>>(
        z0, z1, z2, z3, r0, r1, r2, r3,
        cp[0], cp[1], cp[2], cp[3], n);

    return cudaSuccess;
}

// Phase 3 launch: requires caller to provide sp[4] device arrays (already uploaded).
cudaError_t launch_z_prefix_phase3(
    uint64_t *z0, uint64_t *z1, uint64_t *z2, uint64_t *z3,
    uint64_t *temp0, uint64_t *temp1, uint64_t *temp2, uint64_t *temp3,
    uint64_t *sp[4],
    size_t num_chunks, size_t n, cudaStream_t stream)
{
    if (n == 0) return cudaSuccess;

    constexpr unsigned threads = 256;
    unsigned blocks_chunks = (num_chunks + threads - 1) / threads;
    unsigned blocks_n = (n + threads - 1) / threads;

    // Apply global fixup
    z_prefix_fixup_kernel<<<blocks_chunks, threads, 0, stream>>>(
        z0, z1, z2, z3, sp[0], sp[1], sp[2], sp[3], n);

    // Shift right by 1: z[0]=1, z[i]=z_old[i-1]
    // Copy z→temp, then shift temp→z
    cudaError_t err;
    err = cudaMemcpyAsync(temp0, z0, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(temp1, z1, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(temp2, z2, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;
    err = cudaMemcpyAsync(temp3, z3, n * sizeof(uint64_t), cudaMemcpyDeviceToDevice, stream);
    if (err != cudaSuccess) return err;

    z_shift_right_kernel<<<blocks_n, threads, 0, stream>>>(
        z0, z1, z2, z3, temp0, temp1, temp2, temp3, n);

    return cudaSuccess;
}

} // namespace gnark_gpu
