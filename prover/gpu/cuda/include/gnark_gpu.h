// gnark-gpu C API for Go bindings
// This header provides extern "C" functions for CGO integration

#ifndef GNARK_GPU_H
#define GNARK_GPU_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// =============================================================================
// Opaque handles
// =============================================================================

typedef struct GnarkGPUContext *gnark_gpu_context_t;
typedef struct GnarkGPUFrVector *gnark_gpu_fr_vector_t;
typedef struct GnarkGPUMSM *gnark_gpu_msm_t;

// =============================================================================
// Error codes
// =============================================================================

typedef enum {
    GNARK_GPU_SUCCESS = 0,
    GNARK_GPU_ERROR_CUDA = 1,
    GNARK_GPU_ERROR_INVALID_ARG = 2,
    GNARK_GPU_ERROR_OUT_OF_MEMORY = 3,
    GNARK_GPU_ERROR_SIZE_MISMATCH = 4,
} gnark_gpu_error_t;

// =============================================================================
// Context lifecycle
// =============================================================================

// Initialize GPU context on specified device
// Returns GNARK_GPU_SUCCESS on success, error code otherwise
gnark_gpu_error_t gnark_gpu_init(int device_id, gnark_gpu_context_t *ctx);

// Destroy GPU context and release resources
void gnark_gpu_destroy(gnark_gpu_context_t ctx);

// =============================================================================
// Fr vector operations
// =============================================================================

// Allocate GPU memory for `count` Fr elements
gnark_gpu_error_t gnark_gpu_fr_vector_alloc(gnark_gpu_context_t ctx, size_t count,
                                            gnark_gpu_fr_vector_t *vec);

// Free GPU memory
void gnark_gpu_fr_vector_free(gnark_gpu_fr_vector_t vec);

// Get the number of elements in the vector
size_t gnark_gpu_fr_vector_len(gnark_gpu_fr_vector_t vec);

// =============================================================================
// Data transfer
// Host data is in AoS format (gnark-crypto layout): [e0.l0, e0.l1, e0.l2, e0.l3, e1.l0, ...]
// GPU storage is SoA format for coalesced memory access
// Transpose happens on GPU during copy operations
// =============================================================================

// Copy from host (AoS) to device (SoA)
gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_device(gnark_gpu_fr_vector_t vec,
                                                     const uint64_t *host_data,
                                                     size_t count);

// Copy from device (SoA) to host (AoS)
gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_host(gnark_gpu_fr_vector_t vec,
                                                   uint64_t *host_data, size_t count);

// =============================================================================
// Arithmetic operations (async - call gnark_gpu_sync to wait)
// All operations are element-wise: result[i] = op(a[i], b[i])
// =============================================================================

// Element-wise Montgomery multiplication: result = a * b (mod p)
gnark_gpu_error_t gnark_gpu_fr_vector_mul(gnark_gpu_context_t ctx,
                                          gnark_gpu_fr_vector_t result,
                                          gnark_gpu_fr_vector_t a,
                                          gnark_gpu_fr_vector_t b);

// Element-wise addition: result = a + b (mod p)
gnark_gpu_error_t gnark_gpu_fr_vector_add(gnark_gpu_context_t ctx,
                                          gnark_gpu_fr_vector_t result,
                                          gnark_gpu_fr_vector_t a,
                                          gnark_gpu_fr_vector_t b);

// Element-wise subtraction: result = a - b (mod p)
gnark_gpu_error_t gnark_gpu_fr_vector_sub(gnark_gpu_context_t ctx,
                                          gnark_gpu_fr_vector_t result,
                                          gnark_gpu_fr_vector_t a,
                                          gnark_gpu_fr_vector_t b);

// Ensure the context's shared staging buffer can hold at least min_count elements.
// Called automatically by copy operations; call explicitly to pre-allocate.
gnark_gpu_error_t gnark_gpu_staging_ensure(gnark_gpu_context_t ctx, size_t min_count);

// v[i] *= g^i for i in [0, count). g is 4 uint64s in Montgomery form.
gnark_gpu_error_t gnark_gpu_fr_vector_scale_by_powers(gnark_gpu_context_t ctx,
                                                       gnark_gpu_fr_vector_t v,
                                                       const uint64_t g[4]);

// v[i] *= c for all i. c is 4 uint64s in Montgomery form.
gnark_gpu_error_t gnark_gpu_fr_vector_scalar_mul(gnark_gpu_context_t ctx,
                                                  gnark_gpu_fr_vector_t v,
                                                  const uint64_t c[4]);

// dst[i] = src[i] (device-to-device copy of SoA limbs)
gnark_gpu_error_t gnark_gpu_fr_vector_copy_d2d(gnark_gpu_context_t ctx,
                                                gnark_gpu_fr_vector_t dst,
                                                gnark_gpu_fr_vector_t src);

// Set all elements to zero
gnark_gpu_error_t gnark_gpu_fr_vector_set_zero(gnark_gpu_context_t ctx,
                                                gnark_gpu_fr_vector_t v);

// v[i] += a[i] * b[i] (fused multiply-add)
gnark_gpu_error_t gnark_gpu_fr_vector_addmul(gnark_gpu_context_t ctx,
                                              gnark_gpu_fr_vector_t v,
                                              gnark_gpu_fr_vector_t a,
                                              gnark_gpu_fr_vector_t b);

// v[i] = 1/v[i] using Montgomery batch inversion (parallel two-level scan).
// temp must be a separate FrVector of the same size (used as scratch space).
gnark_gpu_error_t gnark_gpu_fr_vector_batch_invert(gnark_gpu_context_t ctx,
                                                    gnark_gpu_fr_vector_t v,
                                                    gnark_gpu_fr_vector_t temp);

// Size-4 inverse DFT butterfly for decomposed iFFT(4n).
// b0,b1,b2,b3 are 4 FrVectors of size n (modified in-place).
// omega4_inv: inverse of primitive 4th root of unity (4 uint64s, Montgomery form).
// quarter: 1/4 in Montgomery form (4 uint64s).
gnark_gpu_error_t gnark_gpu_fr_vector_butterfly4(gnark_gpu_context_t ctx,
                                                   gnark_gpu_fr_vector_t b0,
                                                   gnark_gpu_fr_vector_t b1,
                                                   gnark_gpu_fr_vector_t b2,
                                                   gnark_gpu_fr_vector_t b3,
                                                   const uint64_t omega4_inv[4],
                                                   const uint64_t quarter[4]);

// =============================================================================
// MSM (Multi-Scalar Multiplication) using Twisted Edwards coordinates
// Points are in compact TE XY format: (x_te, y_te)
// =============================================================================

// Create MSM context for up to max_points compact TE XY points
gnark_gpu_error_t gnark_gpu_msm_create(gnark_gpu_context_t ctx, size_t max_points,
                                       gnark_gpu_msm_t *msm);

// Upload compact TE XY points to GPU (kept resident for reuse)
// points_data layout: [x0.l0..x0.l5, y0.l0..y0.l5, ...]
// Each point is 12 uint64s (96 bytes) in Montgomery form
gnark_gpu_error_t gnark_gpu_msm_load_points(gnark_gpu_msm_t msm,
                                            const uint64_t *points_data,
                                            size_t count);

// Run MSM: result = sum(scalars[i] * points[i]) for i in [0, count)
// scalars: count * 4 uint64s in Montgomery form (kernel converts)
// result: num_windows * 24 uint64s representing per-window TE extended points
gnark_gpu_error_t gnark_gpu_msm_run(gnark_gpu_msm_t msm, uint64_t *result,
                                    const uint64_t *scalars, size_t count);

// Destroy MSM context and free GPU memory
void gnark_gpu_msm_destroy(gnark_gpu_msm_t msm);

// Query MSM configuration (c = window bits, num_windows = ceil(253/c))
void gnark_gpu_msm_get_config(gnark_gpu_msm_t msm, int *c, int *num_windows);

// Upload Short-Weierstrass affine points (gnark bls12377.G1Affine layout —
// 12 uint64s per point, Montgomery form) into the optional d_points_sw GPU
// buffer. Used only by the batched-affine accumulate kernel
// (GNARK_GPU_MSM_BATCHED_AFFINE=1). Allocates the buffer on first call.
gnark_gpu_error_t gnark_gpu_msm_load_points_sw(gnark_gpu_msm_t msm,
                                                const uint64_t *points_data,
                                                size_t count);

// Pin work buffers (sort buffers + cudaHostRegister of caller scalars) across
// gnark_gpu_msm_run calls. Without this, msm_run lazily allocates several
// GB of sort buffers and registers caller memory at the start of each call,
// then frees them at the end — costing 5–10 ms of host overhead per call.
// Use this when running back-to-back MSMs (e.g., a wave of PlonK commitments).
// Caller MUST release before any phase that needs the VRAM (e.g., quotient).
gnark_gpu_error_t gnark_gpu_msm_pin_work_buffers(gnark_gpu_msm_t msm);

// Release pinned work buffers immediately (frees VRAM, drops host
// registration). Subsequent gnark_gpu_msm_run calls re-allocate lazily.
gnark_gpu_error_t gnark_gpu_msm_release_work_buffers(gnark_gpu_msm_t msm);

// Test entrypoints for SW affine primitives — used to validate the GPU
// arithmetic against gnark-crypto host reference. Inputs/outputs use
// gnark's bls12377.G1Affine memory layout (12 uint64 limbs, Montgomery form).
gnark_gpu_error_t gnark_gpu_test_sw_pair_add(
    const uint64_t *p0, const uint64_t *p1, uint64_t *out);

// Convert SW affine to TE extended (X, Y, T, Z) — output is 24 uint64s.
gnark_gpu_error_t gnark_gpu_test_sw_to_te(
    const uint64_t *p_sw, uint64_t *out_te);

// Reduce N affine SW points (≤ 256) via batched-affine pairwise reduction
// in shared memory. Output is the SW affine sum (12 uint64s). Used to
// isolate bugs in the multi-wave reduction logic.
gnark_gpu_error_t gnark_gpu_test_batched_affine_reduce(
    const uint64_t *points_aos, uint64_t *out_aos, int N);

// Per-phase timings of the last gnark_gpu_msm_run call. Phase order
// (9 floats, milliseconds):
//   0: H2D (scalar upload)
//   1: build_pairs (signed-digit decomposition)
//   2: sort (CUB radix sort)
//   3: boundaries (memset + detect_bucket_boundaries)
//   4: accumulate_seq (sequential bucket accumulation, with cap)
//   5: accumulate_par (parallel overflow tail; 0 if no overflow buckets)
//   6: reduce_partial (per-window range scan)
//   7: reduce_finalize (combine ranges into per-window result)
//   8: D2H (window results download)
// Returns the number of phases written (9 on success, 0 if msm/out is null).
int gnark_gpu_msm_get_phase_timings(gnark_gpu_msm_t msm, float *out);

// Offload: free d_points from GPU, keep working buffers
gnark_gpu_error_t gnark_gpu_msm_offload_points(gnark_gpu_msm_t msm);

// Reload: re-allocate d_points and upload from (pinned) host memory
gnark_gpu_error_t gnark_gpu_msm_reload_points(gnark_gpu_msm_t msm,
                                               const uint64_t *points_data, size_t count);

// =============================================================================
// NTT (Number Theoretic Transform)
// =============================================================================

typedef struct GnarkGPUNTTDomain *gnark_gpu_ntt_domain_t;

// Create NTT domain with precomputed twiddle factors.
// size: must be a power of 2
// fwd_twiddles_aos: n/2 elements in AoS format (4 uint64s each), Montgomery form
//   These are w^0, w^1, ..., w^(n/2-1) where w is the n-th root of unity.
// inv_twiddles_aos: n/2 elements in AoS format, twiddles for inverse NTT
//   These are w_inv^0, w_inv^1, ..., w_inv^(n/2-1) where w_inv = w^{-1}.
// inv_n: 1/n in Montgomery form (4 uint64s)
gnark_gpu_error_t gnark_gpu_ntt_domain_create(gnark_gpu_context_t ctx, size_t size,
                                               const uint64_t *fwd_twiddles_aos,
                                               const uint64_t *inv_twiddles_aos,
                                               const uint64_t *inv_n,
                                               gnark_gpu_ntt_domain_t *domain);

// Destroy NTT domain and free GPU twiddle memory.
void gnark_gpu_ntt_domain_destroy(gnark_gpu_ntt_domain_t domain);

// Forward NTT (DIF): natural-order input -> bit-reversed output.
// data must be an allocated FrVector of domain size.
gnark_gpu_error_t gnark_gpu_ntt_forward(gnark_gpu_ntt_domain_t domain,
                                         gnark_gpu_fr_vector_t data);

// Inverse NTT (DIT): bit-reversed input -> natural-order output, scaled by 1/n.
// data must be an allocated FrVector of domain size.
gnark_gpu_error_t gnark_gpu_ntt_inverse(gnark_gpu_ntt_domain_t domain,
                                         gnark_gpu_fr_vector_t data);

// Bit-reversal permutation on an FrVector.
gnark_gpu_error_t gnark_gpu_ntt_bit_reverse(gnark_gpu_ntt_domain_t domain,
                                             gnark_gpu_fr_vector_t data);

// Fused CosetFFT forward: ScaleByPowers + DIF NTT + BitReverse in one call.
// Eliminates one memory round-trip vs separate ScaleByPowers + FFT.
// g: coset generator (4 uint64s, Montgomery form)
// g_half: g^(n/2) precomputed by caller (4 uint64s, Montgomery form)
gnark_gpu_error_t gnark_gpu_ntt_forward_coset(gnark_gpu_ntt_domain_t domain,
                                                gnark_gpu_fr_vector_t data,
                                                const uint64_t g[4],
                                                const uint64_t g_half[4]);

gnark_gpu_error_t gnark_gpu_ntt_forward_coset_stream(gnark_gpu_ntt_domain_t domain,
                                                       gnark_gpu_fr_vector_t data,
                                                       const uint64_t g[4],
                                                       const uint64_t g_half[4],
                                                       int stream_id);

// Fused permutation + boundary constraint kernel for PlonK.
// Computes: result[i] = alpha * (ordering_constraint + alpha * boundary_constraint)
// Inputs: L,R,O,Z (wire evals), S1,S2,S3 (perm evals), L1_denInv (batch-inverted denoms).
// Scalars packed into params array (7 * 4 uint64s):
//   [alpha, beta, gamma, l1_scalar, coset_shift, coset_shift_sq, coset_gen]
// twiddles: forward twiddle array from NTT domain (n/2 elements, SoA layout).
gnark_gpu_error_t gnark_gpu_plonk_perm_boundary(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    gnark_gpu_fr_vector_t Z,
    gnark_gpu_fr_vector_t S1, gnark_gpu_fr_vector_t S2, gnark_gpu_fr_vector_t S3,
    gnark_gpu_fr_vector_t L1_denInv,
    const uint64_t params[28],
    gnark_gpu_ntt_domain_t domain);

// =============================================================================
// Device memory helpers (for permutation table etc.)
// =============================================================================

// Allocate device memory and copy int64 data from host.
// Caller must free with gnark_gpu_device_free_ptr.
gnark_gpu_error_t gnark_gpu_device_alloc_copy_int64(gnark_gpu_context_t ctx,
                                                      const int64_t *host_data, size_t count,
                                                      void **d_ptr);

// Free device memory allocated by gnark_gpu_device_alloc_copy_int64.
void gnark_gpu_device_free_ptr(void *d_ptr);

// =============================================================================
// PlonK Z-polynomial ratio computation
// =============================================================================

// Compute per-element Z ratio factors on GPU.
// For each i: num[i] and den[i] are computed from wire evaluations L, R, O,
// the permutation table, and identity polynomial evaluations.
//
// On exit: L_inout contains numerators, R_inout contains denominators.
// O is read-only (not modified). The caller should then:
//   1. BatchInvert R (denominators → 1/den)
//   2. Mul(L, L, R) to get ratios = num / den
//   3. Download and do CPU prefix product to build Z
//
// params layout: [beta[4], gamma[4], g_mul[4], g_sq[4]] (16 uint64s, Montgomery form)
// d_perm: device pointer to permutation table (3n int64s), from gnark_gpu_device_alloc_copy_int64
// log2n: log2 of domain size n
gnark_gpu_error_t gnark_gpu_plonk_z_compute_factors(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t L_inout,
    gnark_gpu_fr_vector_t R_inout,
    gnark_gpu_fr_vector_t O_in,
    const void *d_perm,
    const uint64_t params[16],
    unsigned log2n,
    gnark_gpu_ntt_domain_t domain);

// =============================================================================
// Pinned memory management
// =============================================================================

// Allocate pinned (page-locked) host memory for fast DMA transfers
gnark_gpu_error_t gnark_gpu_alloc_pinned(void **ptr, size_t bytes);

// Free pinned host memory
void gnark_gpu_free_pinned(void *ptr);

// =============================================================================
// GPU L1 denominator computation
// =============================================================================

// Compute out[i] = cosetGen * omega^i - 1 for i in [0, n).
// Uses forward twiddle factors from the NTT domain.
// The caller should BatchInvert the result to get L1DenInv.
gnark_gpu_error_t gnark_gpu_compute_l1_den(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t out,
    const uint64_t coset_gen[4],
    gnark_gpu_ntt_domain_t domain);

// =============================================================================
// Patch elements
// =============================================================================

// Write `count` AoS elements from host into the SoA GPU vector starting at `offset`.
// host_data_aos layout: [e0.l0, e0.l1, e0.l2, e0.l3, e1.l0, ...]
// Useful for patching a few blinding correction elements without a full H2D transfer.
gnark_gpu_error_t gnark_gpu_fr_vector_patch(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t vec,
    size_t offset,
    const uint64_t *host_data_aos,
    size_t count);

// =============================================================================
// Synchronization
// =============================================================================

// Wait for all GPU operations on the context to complete
gnark_gpu_error_t gnark_gpu_sync(gnark_gpu_context_t ctx);

// =============================================================================
// Multi-stream support
// =============================================================================

#define GNARK_GPU_MAX_STREAMS 4
#define GNARK_GPU_MAX_EVENTS  16

// Create a CUDA stream at the given index. Stream 0 is created automatically.
// stream_id must be in [1, GNARK_GPU_MAX_STREAMS).
gnark_gpu_error_t gnark_gpu_create_stream(gnark_gpu_context_t ctx, int stream_id);

// Record an event on a stream. The event can later be waited on by another stream.
// event_id must be in [0, GNARK_GPU_MAX_EVENTS).
gnark_gpu_error_t gnark_gpu_record_event(gnark_gpu_context_t ctx, int stream_id, int event_id);

// Make a stream wait for an event recorded on another stream.
gnark_gpu_error_t gnark_gpu_wait_event(gnark_gpu_context_t ctx, int stream_id, int event_id);

// Synchronize a specific stream (wait for all operations on it to complete).
gnark_gpu_error_t gnark_gpu_sync_stream(gnark_gpu_context_t ctx, int stream_id);

// =============================================================================
// Stream-aware data transfer
// =============================================================================

// Copy from host (AoS) to device (SoA) on a specific stream.
// For truly async transfers, host_data should be pinned memory.
gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_device_stream(
    gnark_gpu_fr_vector_t vec, const uint64_t *host_data,
    size_t count, int stream_id);

// Copy from device (SoA) to host (AoS) on a specific stream.
// Synchronizes the stream before returning to ensure data is available.
gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_host_stream(
    gnark_gpu_fr_vector_t vec, uint64_t *host_data,
    size_t count, int stream_id);

// Device-to-device copy on a specific stream.
gnark_gpu_error_t gnark_gpu_fr_vector_copy_d2d_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src, int stream_id);

// =============================================================================
// Stream-aware NTT operations
// =============================================================================

gnark_gpu_error_t gnark_gpu_ntt_forward_stream(gnark_gpu_ntt_domain_t domain,
                                                gnark_gpu_fr_vector_t data,
                                                int stream_id);

gnark_gpu_error_t gnark_gpu_ntt_inverse_stream(gnark_gpu_ntt_domain_t domain,
                                                gnark_gpu_fr_vector_t data,
                                                int stream_id);

gnark_gpu_error_t gnark_gpu_ntt_bit_reverse_stream(gnark_gpu_ntt_domain_t domain,
                                                    gnark_gpu_fr_vector_t data,
                                                    int stream_id);

// =============================================================================
// Stream-aware arithmetic operations
// =============================================================================

gnark_gpu_error_t gnark_gpu_fr_vector_scale_by_powers_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    const uint64_t g[4], int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_scalar_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    const uint64_t c[4], int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_set_zero_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v, int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_add_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_sub_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_addmul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id);

gnark_gpu_error_t gnark_gpu_fr_vector_batch_invert_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t temp, int stream_id);

// =============================================================================
// AddScalarMul: v[i] += a[i] * scalar (broadcast scalar multiply-add)
// =============================================================================

gnark_gpu_error_t gnark_gpu_fr_vector_add_scalar_mul(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, const uint64_t scalar[4]);

gnark_gpu_error_t gnark_gpu_fr_vector_add_scalar_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, const uint64_t scalar[4], int stream_id);

// =============================================================================
// Stream-aware PlonK operations
// =============================================================================

gnark_gpu_error_t gnark_gpu_compute_l1_den_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t out,
    const uint64_t coset_gen[4], gnark_gpu_ntt_domain_t domain,
    int stream_id);

gnark_gpu_error_t gnark_gpu_plonk_perm_boundary_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    gnark_gpu_fr_vector_t Z,
    gnark_gpu_fr_vector_t S1, gnark_gpu_fr_vector_t S2, gnark_gpu_fr_vector_t S3,
    gnark_gpu_fr_vector_t L1_denInv,
    const uint64_t params[28],
    gnark_gpu_ntt_domain_t domain, int stream_id);

// =============================================================================
// GPU Z prefix product (two-level parallel scan)
// =============================================================================

// Phase 1: Compute local prefix products and extract chunk products.
// z_vec: output vector (n elements), receives local prefix products
// ratio_vec: input vector (n elements), the per-element ratios
// chunk_products_host: output (num_chunks * 4 uint64s, AoS), chunk products downloaded to host
// num_chunks_out: output, number of chunks
// After this call, the host must compute a sequential prefix scan of chunk_products_host,
// then call gnark_gpu_z_prefix_phase3 with the scanned prefixes.
gnark_gpu_error_t gnark_gpu_z_prefix_phase1(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t z_vec,
    gnark_gpu_fr_vector_t ratio_vec,
    uint64_t *chunk_products_host,
    size_t *num_chunks_out);

// Phase 3: Upload scanned chunk prefixes and apply fixup + shift.
// z_vec: the vector from phase1 (modified in-place)
// temp_vec: scratch vector (same size as z_vec)
// scanned_prefixes_host: the CPU-scanned prefix products (num_chunks * 4 uint64s, AoS)
// num_chunks: from phase1
gnark_gpu_error_t gnark_gpu_z_prefix_phase3(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t z_vec,
    gnark_gpu_fr_vector_t temp_vec,
    const uint64_t *scanned_prefixes_host,
    size_t num_chunks);

// =============================================================================
// GPU polynomial evaluation (chunked Horner)
// =============================================================================

// Evaluate a polynomial at a single point using chunked Horner on GPU.
// coeffs: FrVector of n coefficients (on GPU, SoA format).
// z: evaluation point (4 uint64s, Montgomery form).
// partials_host: output buffer for partial chunk results (num_chunks * 4 uint64s, AoS).
//   Caller must pre-allocate at least ceil(n/1024) * 4 uint64s.
// num_chunks_out: output, number of chunks.
// After this call, the caller combines partials on CPU:
//   zK = z^1024
//   result = partials[C-1]
//   for j = C-2 downto 0: result = result * zK + partials[j]
gnark_gpu_error_t gnark_gpu_poly_eval_chunks(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t coeffs,
    const uint64_t z[4],
    uint64_t *partials_host,
    size_t *num_chunks_out);

gnark_gpu_error_t gnark_gpu_poly_eval_chunks_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t coeffs,
    const uint64_t z[4],
    uint64_t *partials_host,
    size_t *num_chunks_out,
    int stream_id);

// =============================================================================
// Fused gate constraint accumulation for PlonK quotient
// =============================================================================

// Compute result[i] = (result[i] + Ql[i]*L[i] + Qr[i]*R[i] + Qm[i]*L[i]*R[i]
//                      + Qo[i]*O[i] + Qk[i]) * zhKInv
// in a single pass. result already contains permutation+boundary contributions.
// zhKInv is the inverse of Z_H(coset_gen^n - 1), 4 uint64s in Montgomery form.
gnark_gpu_error_t gnark_gpu_plonk_gate_accum(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t Ql, gnark_gpu_fr_vector_t Qr,
    gnark_gpu_fr_vector_t Qm, gnark_gpu_fr_vector_t Qo,
    gnark_gpu_fr_vector_t Qk,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    const uint64_t zhKInv[4]);

// =============================================================================
// Reduce blinded polynomial for coset evaluation
//
// Computes: dst[i] = src[i] + src[n+j] * cosetPowN  for j in [0, tail_len)
//           dst[i] = src[i]                           for i in [tail_len, n)
//
// src: GPU FrVector of length n (first n coefficients of blinded poly)
// blinding_tail_host: pointer to tail coefficients in AoS layout (host memory)
// tail_len: number of tail coefficients (typically 2 or 3)
// cosetPowN: coset generator raised to power n (4 uint64s, Montgomery form)
// dst: output FrVector of length n (receives reduced coefficients)
// =============================================================================

gnark_gpu_error_t gnark_gpu_reduce_blinded_coset(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src,
    const uint64_t *blinding_tail_host,
    size_t tail_len,
    const uint64_t cosetPowN[4]);

gnark_gpu_error_t gnark_gpu_reduce_blinded_coset_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src,
    const uint64_t *blinding_tail_host,
    size_t tail_len,
    const uint64_t cosetPowN[4],
    int stream_id);

// =============================================================================
// GPU Horner quotient: h(X) = (p(X) - p(z)) / (X - z) in-place
//
// Computes the quotient polynomial on GPU. The input FrVector is modified
// in-place: after completion, poly[0] = p(z) (evaluation), poly[1:] = quotient.
// =============================================================================

gnark_gpu_error_t gnark_gpu_horner_quotient(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t poly,
    gnark_gpu_fr_vector_t temp,
    const uint64_t z[4]);

// =============================================================================
// GPU memory info
// =============================================================================

// Query free and total GPU memory in bytes.
gnark_gpu_error_t gnark_gpu_mem_get_info(gnark_gpu_context_t ctx,
                                          size_t *free_bytes, size_t *total_bytes);

#ifdef __cplusplus
}
#endif

#endif // GNARK_GPU_H
