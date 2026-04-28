// =============================================================================
// gnark-gpu C API bridge (CGO-facing layer)
//
// Purpose:
//   - Keep exported ABI flat and stable (`extern "C"` handles + POD args).
//   - Keep CUDA-heavy logic in dedicated modules (`msm.cu`, `ntt.cu`, etc.).
//   - Keep this file as a thin router + lifecycle owner.
//
// Layering:
//
//   Go wrappers (gpu/*.go)
//           |
//           v
//   C ABI (gnark_gpu.h / this file)
//           |
//           v
//   Internal launchers + contexts (cuda/src/*.cu, *.cuh)
//
// Handle model:
//   GnarkGPUContext  -> owns CUDA stream(s), reusable staging buffers
//   GnarkGPUFrVector -> owns SoA limb allocations
//   GnarkGPUMSM      -> owns persistent point buffers + MSM work buffers
//   GnarkGPUNTTDomain-> owns twiddle tables for one domain size
//
// Design rule:
//   No algorithmic complexity here. This file validates arguments, dispatches
//   to kernels, and translates CUDA/launcher failures to API error codes.
// =============================================================================

#include "gnark_gpu.h"
#include "field.cuh"
#include "../plonk2/field.cuh"
#include <cuda_runtime.h>
#include <vector>

namespace gnark_gpu {

// Forward declarations for kernel launchers (defined in kernels.cu)
void launch_mul_mont_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                        const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                        const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                        const uint64_t *b2, const uint64_t *b3, size_t n,
                        cudaStream_t stream);

void launch_add_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                   const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                   const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                   const uint64_t *b2, const uint64_t *b3, size_t n, cudaStream_t stream);

void launch_sub_fr(uint64_t *c0, uint64_t *c1, uint64_t *c2, uint64_t *c3,
                   const uint64_t *a0, const uint64_t *a1, const uint64_t *a2,
                   const uint64_t *a3, const uint64_t *b0, const uint64_t *b1,
                   const uint64_t *b2, const uint64_t *b3, size_t n, cudaStream_t stream);

void launch_transpose_aos_to_soa_fr(uint64_t *limb0, uint64_t *limb1, uint64_t *limb2,
                                    uint64_t *limb3, const uint64_t *aos_data, size_t count,
                                    cudaStream_t stream);

void launch_transpose_soa_to_aos_fr(uint64_t *aos_data, const uint64_t *limb0,
                                    const uint64_t *limb1, const uint64_t *limb2,
                                    const uint64_t *limb3, size_t count,
                                    cudaStream_t stream);

// Forward declarations for new Fr operations (defined in fr_ops.cu)
void launch_scale_by_powers(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                             const uint64_t g[4], size_t n, cudaStream_t stream);
void launch_scalar_mul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                        const uint64_t c[4], size_t n, cudaStream_t stream);
void launch_addmul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                    const uint64_t *a0, const uint64_t *a1, const uint64_t *a2, const uint64_t *a3,
                    const uint64_t *b0, const uint64_t *b1, const uint64_t *b2, const uint64_t *b3,
                    size_t n, cudaStream_t stream);
cudaError_t launch_batch_invert(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                                 uint64_t *orig0, uint64_t *orig1, uint64_t *orig2, uint64_t *orig3,
                                 size_t n, cudaStream_t stream);
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
    size_t n, cudaStream_t stream);

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
    size_t n, unsigned log2n, cudaStream_t stream);

void launch_compute_l1_den(
    uint64_t *out0, uint64_t *out1, uint64_t *out2, uint64_t *out3,
    const uint64_t *tw0, const uint64_t *tw1, const uint64_t *tw2, const uint64_t *tw3,
    const uint64_t cg[4], size_t n, cudaStream_t stream);

void launch_reduce_blinded_coset(
    uint64_t *dst0, uint64_t *dst1, uint64_t *dst2, uint64_t *dst3,
    const uint64_t *src0, const uint64_t *src1,
    const uint64_t *src2, const uint64_t *src3,
    const uint64_t cpn[4],
    const uint64_t *tail_device,
    uint32_t tail_len, uint32_t n, cudaStream_t stream);

void launch_add_scalar_mul(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                            const uint64_t *a0, const uint64_t *a1, const uint64_t *a2, const uint64_t *a3,
                            const uint64_t scalar[4], size_t n, cudaStream_t stream);

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
    const uint64_t zhKInv[4], size_t n, cudaStream_t stream);

void launch_butterfly4(
    uint64_t *b0_0, uint64_t *b0_1, uint64_t *b0_2, uint64_t *b0_3,
    uint64_t *b1_0, uint64_t *b1_1, uint64_t *b1_2, uint64_t *b1_3,
    uint64_t *b2_0, uint64_t *b2_1, uint64_t *b2_2, uint64_t *b2_3,
    uint64_t *b3_0, uint64_t *b3_1, uint64_t *b3_2, uint64_t *b3_3,
    const uint64_t omega4_inv[4], const uint64_t quarter[4],
    size_t n, cudaStream_t stream);

// Forward declarations for Z prefix product (defined in plonk_z.cu)
cudaError_t launch_z_prefix_phase1(
    uint64_t *z0, uint64_t *z1, uint64_t *z2, uint64_t *z3,
    const uint64_t *r0, const uint64_t *r1, const uint64_t *r2, const uint64_t *r3,
    uint64_t *cp[4],
    size_t n, cudaStream_t stream);
cudaError_t launch_z_prefix_phase3(
    uint64_t *z0, uint64_t *z1, uint64_t *z2, uint64_t *z3,
    uint64_t *temp0, uint64_t *temp1, uint64_t *temp2, uint64_t *temp3,
    uint64_t *sp[4],
    size_t num_chunks, size_t n, cudaStream_t stream);

// Forward declarations for polynomial evaluation (defined in plonk_eval.cu)
void launch_poly_eval_chunks(
    const uint64_t *c0, const uint64_t *c1,
    const uint64_t *c2, const uint64_t *c3,
    const uint64_t z[4],
    uint64_t *out0, uint64_t *out1,
    uint64_t *out2, uint64_t *out3,
    size_t n, size_t *num_chunks_out,
    cudaStream_t stream);

// Forward declarations for MSM functions (defined in msm.cu)
struct MSMContext;
struct G1EdExtended;
MSMContext *msm_create(size_t max_points);
void msm_destroy(MSMContext *ctx);
void msm_load_points(MSMContext *ctx, const void *host_points, size_t count, cudaStream_t stream);
void msm_upload_scalars(MSMContext *ctx, const uint64_t *host_scalars, size_t n, cudaStream_t stream);
void launch_msm(MSMContext *ctx, size_t n, cudaStream_t stream);
void msm_download_results(MSMContext *ctx, G1EdExtended *host_results, cudaStream_t stream);
cudaError_t msm_run_full(MSMContext *ctx, const uint64_t *host_scalars, size_t n,
                         G1EdExtended *host_results, cudaStream_t compute_stream);
void msm_offload_points(MSMContext *ctx);
void msm_unregister_host(MSMContext *ctx);
cudaError_t msm_reload_points(MSMContext *ctx, const void *host_points, size_t count, cudaStream_t stream);
cudaError_t msm_load_points_sw(MSMContext *ctx, const void *host_sw_points, size_t count, cudaStream_t stream);
int msm_get_c(MSMContext *ctx);
int msm_get_num_windows(MSMContext *ctx);
int msm_get_phase_timings(MSMContext *ctx, float *out);
void msm_pin_buffers(MSMContext *ctx);
void msm_release_buffers(MSMContext *ctx);
cudaError_t test_sw_pair_add_run(const uint64_t *p0, const uint64_t *p1, uint64_t *out);
cudaError_t test_sw_to_te_run(const uint64_t *p_sw, uint64_t *out_te);
cudaError_t test_batched_affine_reduce_run(const uint64_t *points_aos, uint64_t *out_aos, int N);

// Forward declarations for NTT functions (defined in ntt.cu)
struct NTTDomain;
NTTDomain *ntt_domain_create(size_t size, const uint64_t *fwd_twiddles_aos,
                              const uint64_t *inv_twiddles_aos, const uint64_t inv_n[4],
                              cudaStream_t stream);
void ntt_domain_destroy(NTTDomain *dom);
void launch_ntt_forward(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                        cudaStream_t stream);
void launch_ntt_forward_coset(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                               const uint64_t g[4], const uint64_t g_half[4],
                               cudaStream_t stream);
void launch_ntt_inverse(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                        cudaStream_t stream);
void launch_ntt_bit_reverse(NTTDomain *dom, uint64_t *d0, uint64_t *d1, uint64_t *d2, uint64_t *d3,
                            cudaStream_t stream);
void ntt_get_fwd_twiddles(const NTTDomain *dom, const uint64_t **out_ptrs);

} // namespace gnark_gpu

namespace gnark_gpu::plonk2 {

void launch_copy_aos_to_soa(gnark_gpu_plonk2_curve_id_t curve, FrView dst,
                            const uint64_t *src, size_t n, cudaStream_t stream);
void launch_copy_soa_to_aos(gnark_gpu_plonk2_curve_id_t curve, uint64_t *dst,
                            ConstFrView src, size_t n, cudaStream_t stream);
void launch_set_zero(gnark_gpu_plonk2_curve_id_t curve, FrView v, size_t n,
                     cudaStream_t stream);
void launch_add(gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a,
                ConstFrView b, size_t n, cudaStream_t stream);
void launch_sub(gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a,
                ConstFrView b, size_t n, cudaStream_t stream);
void launch_mul(gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a,
                ConstFrView b, size_t n, cudaStream_t stream);
void launch_addmul(gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a,
                   ConstFrView b, size_t n, cudaStream_t stream);
void launch_scalar_mul(gnark_gpu_plonk2_curve_id_t curve, FrView out,
                       const uint64_t *scalar, size_t n, cudaStream_t stream);
void launch_add_scalar_mul(gnark_gpu_plonk2_curve_id_t curve, FrView out,
                           ConstFrView a, const uint64_t *scalar, size_t n,
                           cudaStream_t stream);
void launch_batch_invert(gnark_gpu_plonk2_curve_id_t curve, FrView data,
                         size_t n, cudaStream_t stream);
void launch_butterfly4_inverse(gnark_gpu_plonk2_curve_id_t curve,
                               FrView b0, FrView b1, FrView b2, FrView b3,
                               const uint64_t *omega4_inv,
                               const uint64_t *quarter, size_t n,
                               cudaStream_t stream);
void launch_reduce_blinded_coset(gnark_gpu_plonk2_curve_id_t curve,
                                 FrView dst, ConstFrView src,
                                 const uint64_t *tail, size_t tail_len,
                                 const uint64_t *coset_pow_n, size_t n,
                                 cudaStream_t stream);
void launch_compute_l1_den(gnark_gpu_plonk2_curve_id_t curve, FrView out,
                           ConstFrView twiddles, const uint64_t *coset_gen,
                           size_t n, cudaStream_t stream);
void launch_gate_accum(gnark_gpu_plonk2_curve_id_t curve, FrView result,
                       ConstFrView ql, ConstFrView qr, ConstFrView qm,
                       ConstFrView qo, ConstFrView qk, ConstFrView l,
                       ConstFrView r, ConstFrView o,
                       const uint64_t *zh_k_inv, size_t n,
                       cudaStream_t stream);
void launch_perm_boundary(gnark_gpu_plonk2_curve_id_t curve, FrView result,
                          ConstFrView l, ConstFrView r, ConstFrView o,
                          ConstFrView z, ConstFrView s1, ConstFrView s2,
                          ConstFrView s3, ConstFrView l1_den_inv,
                          ConstFrView twiddles, const uint64_t *params,
                          size_t n, cudaStream_t stream);
void launch_z_compute_factors(gnark_gpu_plonk2_curve_id_t curve,
                              FrView l_inout, FrView r_inout,
                              ConstFrView o, const int64_t *perm,
                              ConstFrView twiddles, const uint64_t *params,
                              size_t n, unsigned log2n,
                              cudaStream_t stream);
void launch_z_prefix_phase1(gnark_gpu_plonk2_curve_id_t curve, FrView z,
                            ConstFrView ratio, uint64_t *chunk_products,
                            size_t n, cudaStream_t stream);
void launch_z_prefix_phase3(gnark_gpu_plonk2_curve_id_t curve, FrView z,
                            FrView temp, const uint64_t *scanned_prefixes,
                            size_t num_chunks, size_t n,
                            cudaStream_t stream);
void launch_ntt_forward(gnark_gpu_plonk2_curve_id_t curve, FrView data,
                        ConstFrView twiddles, size_t n, cudaStream_t stream);
void launch_ntt_inverse(gnark_gpu_plonk2_curve_id_t curve, FrView data,
                        ConstFrView twiddles, const uint64_t *inv_n, size_t n,
                        cudaStream_t stream);
void launch_scale_by_powers(gnark_gpu_plonk2_curve_id_t curve, FrView data,
                            const uint64_t *generator, uint64_t *local_powers,
                            size_t n, cudaStream_t stream);
void launch_bit_reverse(gnark_gpu_plonk2_curve_id_t curve, FrView data,
                        size_t n, cudaStream_t stream);
cudaError_t g1_affine_add_run(gnark_gpu_plonk2_curve_id_t curve,
                              const uint64_t *p, const uint64_t *q,
                              uint64_t *out, cudaStream_t stream);
cudaError_t g1_affine_double_run(gnark_gpu_plonk2_curve_id_t curve,
                                 const uint64_t *p, uint64_t *out,
                                 cudaStream_t stream);
cudaError_t msm_naive_run(gnark_gpu_plonk2_curve_id_t curve,
                          const uint64_t *points, const uint64_t *scalars,
                          size_t count, uint64_t *out, cudaStream_t stream);
cudaError_t msm_pippenger_run(gnark_gpu_plonk2_curve_id_t curve,
                              const uint64_t *points, const uint64_t *scalars,
                              size_t count, int window_bits, uint64_t *out,
                              cudaStream_t stream);
cudaError_t msm_pippenger_device_points_run(gnark_gpu_plonk2_curve_id_t curve,
                                            const uint64_t *d_points,
                                            const uint64_t *scalars,
                                            size_t count, int window_bits,
                                            uint64_t *out,
                                            cudaStream_t stream);
cudaError_t msm_pippenger_sort_temp_bytes(gnark_gpu_plonk2_curve_id_t curve,
                                          size_t count, int window_bits,
                                          size_t *temp_bytes);
cudaError_t msm_pippenger_device_points_prealloc_run(
    gnark_gpu_plonk2_curve_id_t curve,
    const uint64_t *d_points,
    const uint64_t *scalars,
    size_t count,
    int window_bits,
    uint64_t *out,
    uint64_t *d_scalars,
    uint64_t *d_out,
    uint32_t *d_keys_in,
    uint32_t *d_keys_out,
    uint32_t *d_vals_in,
    uint32_t *d_vals_out,
    uint32_t *d_bucket_offsets,
    uint32_t *d_bucket_ends,
    void *d_buckets,
    void *d_window_results,
    void *d_partial_totals,
    void *d_partial_sums,
    void *d_sort_temp,
    size_t sort_temp_bytes,
    cudaStream_t stream);

} // namespace gnark_gpu::plonk2

// =============================================================================
// Internal structures
// =============================================================================

// Scratch buffers for Z prefix product (owned by context, not thread_local).
struct ZPrefixScratch {
    uint64_t *cp[4] = {};   // chunk products (device)
    uint64_t *sp[4] = {};   // scanned prefixes (device)
    size_t capacity = 0;
};

// Scratch buffers for poly eval chunks (owned by context).
struct PolyEvalScratch {
    uint64_t *out[4] = {};  // partial results (device)
    size_t capacity = 0;
};

struct GnarkGPUContext {
    int device_id;
    cudaStream_t stream; // default stream (stream 0), alias for streams[0]
    // Multi-stream support
    cudaStream_t streams[GNARK_GPU_MAX_STREAMS];
    bool stream_created[GNARK_GPU_MAX_STREAMS];
    cudaEvent_t events[GNARK_GPU_MAX_EVENTS];
    bool event_created[GNARK_GPU_MAX_EVENTS];
    // Shared staging buffer for AoS↔SoA transfers (one per context, reused)
    uint64_t *staging_buffer;
    size_t staging_count; // capacity in Fr elements (buffer is 4*staging_count uint64s)
    // Curve-generic staging buffer used by plonk2. Capacity is raw uint64 words.
    uint64_t *plonk2_staging_buffer;
    size_t plonk2_staging_words;
    // Scratch buffers for Z prefix product and poly eval (context-owned)
    ZPrefixScratch z_prefix_scratch;
    PolyEvalScratch poly_eval_scratch;
};

struct GnarkGPUFrVector {
    GnarkGPUContext *ctx;
    size_t count;
    // SoA storage: 4 separate arrays for the 4 limbs
    uint64_t *limbs[4];
};

// =============================================================================
// Helper to convert CUDA errors
// =============================================================================

static gnark_gpu_error_t check_cuda(cudaError_t err) {
    if (err == cudaSuccess) {
        return GNARK_GPU_SUCCESS;
    }
    if (err == cudaErrorMemoryAllocation) {
        return GNARK_GPU_ERROR_OUT_OF_MEMORY;
    }
    return GNARK_GPU_ERROR_CUDA;
}

// Get the CUDA stream for a given stream_id. Returns nullptr on invalid ID.
static cudaStream_t get_stream(GnarkGPUContext *ctx, int stream_id) {
    if (stream_id < 0 || stream_id >= GNARK_GPU_MAX_STREAMS) return nullptr;
    if (!ctx->stream_created[stream_id]) return nullptr;
    return ctx->streams[stream_id];
}

// =============================================================================
// Context lifecycle
// =============================================================================

// Bind the calling OS thread to a CUDA device.
// See header doc — required before any allocation / kernel launch on
// multi-GPU hosts so that the call lands on the right device. Idempotent.
extern "C" gnark_gpu_error_t gnark_gpu_set_device(int device_id) {
    return check_cuda(cudaSetDevice(device_id));
}

extern "C" gnark_gpu_error_t gnark_gpu_init(int device_id, gnark_gpu_context_t *ctx) {
    if (!ctx) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    cudaError_t err = cudaSetDevice(device_id);
    if (err != cudaSuccess) {
        return check_cuda(err);
    }

    GnarkGPUContext *c = new GnarkGPUContext{};
    c->device_id = device_id;
    c->staging_buffer = nullptr;
    c->staging_count = 0;
    c->plonk2_staging_buffer = nullptr;
    c->plonk2_staging_words = 0;

    // Initialize stream/event arrays
    for (int i = 0; i < GNARK_GPU_MAX_STREAMS; i++) {
        c->streams[i] = nullptr;
        c->stream_created[i] = false;
    }
    for (int i = 0; i < GNARK_GPU_MAX_EVENTS; i++) {
        c->events[i] = nullptr;
        c->event_created[i] = false;
    }

    // Create default stream (stream 0)
    err = cudaStreamCreate(&c->streams[0]);
    if (err != cudaSuccess) {
        delete c;
        return check_cuda(err);
    }
    c->stream_created[0] = true;
    c->stream = c->streams[0]; // alias

    *ctx = c;
    return GNARK_GPU_SUCCESS;
}

extern "C" void gnark_gpu_destroy(gnark_gpu_context_t ctx) {
    if (ctx) {
        if (ctx->staging_buffer) {
            cudaFree(ctx->staging_buffer);
        }
        if (ctx->plonk2_staging_buffer) {
            cudaFree(ctx->plonk2_staging_buffer);
        }
        // Free Z prefix scratch
        for (int i = 0; i < 4; i++) {
            if (ctx->z_prefix_scratch.cp[i]) cudaFree(ctx->z_prefix_scratch.cp[i]);
            if (ctx->z_prefix_scratch.sp[i]) cudaFree(ctx->z_prefix_scratch.sp[i]);
        }
        // Free poly eval scratch
        for (int i = 0; i < 4; i++) {
            if (ctx->poly_eval_scratch.out[i]) cudaFree(ctx->poly_eval_scratch.out[i]);
        }
        for (int i = 0; i < GNARK_GPU_MAX_EVENTS; i++) {
            if (ctx->event_created[i]) {
                cudaEventDestroy(ctx->events[i]);
            }
        }
        for (int i = 0; i < GNARK_GPU_MAX_STREAMS; i++) {
            if (ctx->stream_created[i]) {
                cudaStreamDestroy(ctx->streams[i]);
            }
        }
        delete ctx;
    }
}

// =============================================================================
// Fr vector operations
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_alloc(gnark_gpu_context_t ctx, size_t count,
                                                       gnark_gpu_fr_vector_t *vec) {
    if (!ctx || !vec || count == 0) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    GnarkGPUFrVector *v = new GnarkGPUFrVector;
    v->ctx = ctx;
    v->count = count;

    // Allocate SoA limb arrays
    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMalloc(&v->limbs[i], count * sizeof(uint64_t));
        if (err != cudaSuccess) {
            // Cleanup on failure
            for (int j = 0; j < i; j++) {
                cudaFree(v->limbs[j]);
            }
            delete v;
            // Clear sticky error so subsequent CUDA calls aren't poisoned.
            cudaGetLastError();
            return check_cuda(err);
        }
    }

    *vec = v;
    return GNARK_GPU_SUCCESS;
}

extern "C" void gnark_gpu_fr_vector_free(gnark_gpu_fr_vector_t vec) {
    if (vec) {
        for (int i = 0; i < 4; i++) {
            if (vec->limbs[i]) {
                cudaFree(vec->limbs[i]);
            }
        }
        delete vec;
    }
}

extern "C" size_t gnark_gpu_fr_vector_len(gnark_gpu_fr_vector_t vec) {
    if (!vec) {
        return 0;
    }
    return vec->count;
}

// =============================================================================
// Shared staging buffer management
// =============================================================================

// Ensure the context's staging buffer can hold at least min_count Fr elements.
static gnark_gpu_error_t ensure_staging(GnarkGPUContext *ctx, size_t min_count) {
    if (ctx->staging_count >= min_count) {
        return GNARK_GPU_SUCCESS;
    }
    // Free old buffer if any
    if (ctx->staging_buffer) {
        // Must sync before freeing — prior operations may still be using it
        cudaError_t err = cudaStreamSynchronize(ctx->stream);
        if (err != cudaSuccess) return check_cuda(err);
        cudaFree(ctx->staging_buffer);
        ctx->staging_buffer = nullptr;
        ctx->staging_count = 0;
    }
    cudaError_t err = cudaMalloc(&ctx->staging_buffer, min_count * 4 * sizeof(uint64_t));
    if (err != cudaSuccess) {
        return check_cuda(err);
    }
    ctx->staging_count = min_count;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_staging_ensure(gnark_gpu_context_t ctx, size_t min_count) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    return ensure_staging(ctx, min_count);
}

// =============================================================================
// Data transfer with AoS↔SoA transpose (using shared staging buffer)
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_device(gnark_gpu_fr_vector_t vec,
                                                                const uint64_t *host_data,
                                                                size_t count) {
    if (!vec || !host_data) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (count != vec->count) {
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    }

    GnarkGPUContext *ctx = vec->ctx;
    gnark_gpu_error_t gerr = ensure_staging(ctx, count);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    cudaStream_t stream = ctx->stream;

    // Copy AoS data from host to shared staging buffer
    cudaError_t err = cudaMemcpyAsync(ctx->staging_buffer, host_data,
                                      count * 4 * sizeof(uint64_t),
                                      cudaMemcpyHostToDevice, stream);
    if (err != cudaSuccess) {
        return check_cuda(err);
    }

    // Transpose from AoS to SoA on GPU
    gnark_gpu::launch_transpose_aos_to_soa_fr(vec->limbs[0], vec->limbs[1], vec->limbs[2],
                                              vec->limbs[3], ctx->staging_buffer, count, stream);

    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_host(gnark_gpu_fr_vector_t vec,
                                                              uint64_t *host_data,
                                                              size_t count) {
    if (!vec || !host_data) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (count != vec->count) {
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    }

    GnarkGPUContext *ctx = vec->ctx;
    gnark_gpu_error_t gerr = ensure_staging(ctx, count);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    cudaStream_t stream = ctx->stream;

    // Transpose from SoA to AoS on GPU into shared staging buffer
    gnark_gpu::launch_transpose_soa_to_aos_fr(ctx->staging_buffer, vec->limbs[0], vec->limbs[1],
                                              vec->limbs[2], vec->limbs[3], count, stream);

    // Copy AoS data from staging buffer to host
    cudaError_t err = cudaMemcpyAsync(host_data, ctx->staging_buffer,
                                      count * 4 * sizeof(uint64_t),
                                      cudaMemcpyDeviceToHost, stream);
    if (err != cudaSuccess) {
        return check_cuda(err);
    }

    // Must sync to ensure data is available on host
    err = cudaStreamSynchronize(stream);
    return check_cuda(err);
}

// =============================================================================
// Arithmetic operations
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_mul(gnark_gpu_context_t ctx,
                                                     gnark_gpu_fr_vector_t result,
                                                     gnark_gpu_fr_vector_t a,
                                                     gnark_gpu_fr_vector_t b) {
    if (!ctx || !result || !a || !b) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (result->count != a->count || a->count != b->count) {
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    }

    gnark_gpu::launch_mul_mont_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                                  result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                                  a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                                  b->limbs[3], a->count, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_add(gnark_gpu_context_t ctx,
                                                     gnark_gpu_fr_vector_t result,
                                                     gnark_gpu_fr_vector_t a,
                                                     gnark_gpu_fr_vector_t b) {
    if (!ctx || !result || !a || !b) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (result->count != a->count || a->count != b->count) {
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    }

    gnark_gpu::launch_add_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                             result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                             a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                             b->limbs[3], a->count, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_sub(gnark_gpu_context_t ctx,
                                                     gnark_gpu_fr_vector_t result,
                                                     gnark_gpu_fr_vector_t a,
                                                     gnark_gpu_fr_vector_t b) {
    if (!ctx || !result || !a || !b) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (result->count != a->count || a->count != b->count) {
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    }

    gnark_gpu::launch_sub_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                             result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                             a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                             b->limbs[3], a->count, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// MSM operations (Twisted Edwards)
// =============================================================================

struct GnarkGPUMSM {
    GnarkGPUContext *ctx;
    gnark_gpu::MSMContext *msm_ctx;
    size_t max_points;
    size_t loaded_points;
};

extern "C" gnark_gpu_error_t gnark_gpu_msm_create(gnark_gpu_context_t ctx, size_t max_points,
                                                   gnark_gpu_msm_t *msm) {
    if (!ctx || !msm || max_points == 0) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    cudaError_t err = cudaSetDevice(ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::MSMContext *msm_ctx = gnark_gpu::msm_create(max_points);
    if (!msm_ctx) return GNARK_GPU_ERROR_OUT_OF_MEMORY;

    GnarkGPUMSM *m = new GnarkGPUMSM;
    m->ctx = ctx;
    m->msm_ctx = msm_ctx;
    m->max_points = max_points;
    m->loaded_points = 0;

    *msm = m;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_load_points(gnark_gpu_msm_t msm,
                                                        const uint64_t *points_data,
                                                        size_t count) {
    if (!msm || !points_data || count == 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count > msm->max_points) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::msm_load_points(msm->msm_ctx, points_data, count, msm->ctx->stream);

    cudaError_t err = cudaStreamSynchronize(msm->ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    msm->loaded_points = count;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_run(gnark_gpu_msm_t msm, uint64_t *result,
                                                const uint64_t *scalars, size_t count) {
    if (!msm || !result || !scalars || count == 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count > msm->loaded_points) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    // msm_run_full handles: lazy alloc sort buffers → upload → compute →
    // sync → unregister host → free sort buffers.
    cudaError_t err = gnark_gpu::msm_run_full(
        msm->msm_ctx, scalars, count,
        reinterpret_cast<gnark_gpu::G1EdExtended *>(result), msm->ctx->stream);

    return check_cuda(err);
}

extern "C" void gnark_gpu_msm_destroy(gnark_gpu_msm_t msm) {
    if (msm) {
        gnark_gpu::msm_destroy(msm->msm_ctx);
        delete msm;
    }
}

extern "C" void gnark_gpu_msm_get_config(gnark_gpu_msm_t msm, int *c, int *num_windows) {
    if (msm && msm->msm_ctx) {
        if (c) *c = gnark_gpu::msm_get_c(msm->msm_ctx);
        if (num_windows) *num_windows = gnark_gpu::msm_get_num_windows(msm->msm_ctx);
    }
}

extern "C" int gnark_gpu_msm_get_phase_timings(gnark_gpu_msm_t msm, float *out) {
    if (!msm || !msm->msm_ctx || !out) return 0;
    return gnark_gpu::msm_get_phase_timings(msm->msm_ctx, out);
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_pin_work_buffers(gnark_gpu_msm_t msm) {
    if (!msm || !msm->msm_ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    gnark_gpu::msm_pin_buffers(msm->msm_ctx);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_release_work_buffers(gnark_gpu_msm_t msm) {
    if (!msm || !msm->msm_ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    cudaError_t err = cudaSetDevice(msm->ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);
    gnark_gpu::msm_release_buffers(msm->msm_ctx);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_load_points_sw(gnark_gpu_msm_t msm,
                                                           const uint64_t *points_data,
                                                           size_t count) {
    if (!msm || !points_data || count == 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count > msm->max_points) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaError_t err = cudaSetDevice(msm->ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    cudaError_t cuda_err = gnark_gpu::msm_load_points_sw(
        msm->msm_ctx, points_data, count, msm->ctx->stream);
    if (cuda_err != cudaSuccess) return check_cuda(cuda_err);

    err = cudaStreamSynchronize(msm->ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_test_sw_pair_add(
    const uint64_t *p0, const uint64_t *p1, uint64_t *out) {
    if (!p0 || !p1 || !out) return GNARK_GPU_ERROR_INVALID_ARG;
    return check_cuda(gnark_gpu::test_sw_pair_add_run(p0, p1, out));
}

extern "C" gnark_gpu_error_t gnark_gpu_test_sw_to_te(
    const uint64_t *p_sw, uint64_t *out_te) {
    if (!p_sw || !out_te) return GNARK_GPU_ERROR_INVALID_ARG;
    return check_cuda(gnark_gpu::test_sw_to_te_run(p_sw, out_te));
}

extern "C" gnark_gpu_error_t gnark_gpu_test_batched_affine_reduce(
    const uint64_t *points_aos, uint64_t *out_aos, int N) {
    if (!points_aos || !out_aos || N <= 0) return GNARK_GPU_ERROR_INVALID_ARG;
    return check_cuda(gnark_gpu::test_batched_affine_reduce_run(points_aos, out_aos, N));
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_offload_points(gnark_gpu_msm_t msm) {
    if (!msm) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaError_t err = cudaSetDevice(msm->ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::msm_offload_points(msm->msm_ctx);
    msm->loaded_points = 0;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_msm_reload_points(gnark_gpu_msm_t msm,
                                                          const uint64_t *points_data,
                                                          size_t count) {
    if (!msm || !points_data || count == 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count > msm->max_points) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaError_t err = cudaSetDevice(msm->ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    cudaError_t cuda_err = gnark_gpu::msm_reload_points(
        msm->msm_ctx, points_data, count, msm->ctx->stream);
    if (cuda_err != cudaSuccess) return check_cuda(cuda_err);

    err = cudaStreamSynchronize(msm->ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    msm->loaded_points = count;
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// NTT operations
// =============================================================================

struct GnarkGPUNTTDomain {
    GnarkGPUContext *ctx;
    gnark_gpu::NTTDomain *ntt_dom;
    size_t size;
};

extern "C" gnark_gpu_error_t gnark_gpu_ntt_domain_create(gnark_gpu_context_t ctx, size_t size,
                                                          const uint64_t *fwd_twiddles_aos,
                                                          const uint64_t *inv_twiddles_aos,
                                                          const uint64_t *inv_n,
                                                          gnark_gpu_ntt_domain_t *domain) {
    if (!ctx || !fwd_twiddles_aos || !inv_twiddles_aos || !inv_n || !domain || size == 0) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    // Verify power of 2
    if ((size & (size - 1)) != 0) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    cudaError_t err = cudaSetDevice(ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::NTTDomain *ntt_dom = gnark_gpu::ntt_domain_create(
        size, fwd_twiddles_aos, inv_twiddles_aos, inv_n, ctx->stream);
    if (!ntt_dom) return GNARK_GPU_ERROR_OUT_OF_MEMORY;

    GnarkGPUNTTDomain *d = new GnarkGPUNTTDomain;
    d->ctx = ctx;
    d->ntt_dom = ntt_dom;
    d->size = size;

    *domain = d;
    return GNARK_GPU_SUCCESS;
}

extern "C" void gnark_gpu_ntt_domain_destroy(gnark_gpu_ntt_domain_t domain) {
    if (domain) {
        gnark_gpu::ntt_domain_destroy(domain->ntt_dom);
        delete domain;
    }
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_forward(gnark_gpu_ntt_domain_t domain,
                                                    gnark_gpu_fr_vector_t data) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_ntt_forward(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3],
        domain->ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// Fused CosetFFT: ScaleByPowers + forward NTT + BitReverse
static gnark_gpu_error_t ntt_forward_coset_impl(gnark_gpu_ntt_domain_t domain,
                                                  gnark_gpu_fr_vector_t data,
                                                  const uint64_t g[4],
                                                  const uint64_t g_half[4],
                                                  cudaStream_t stream) {
    if (!domain || !data || !g || !g_half) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_ntt_forward_coset(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3],
        g, g_half, stream);
    gnark_gpu::launch_ntt_bit_reverse(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3],
        stream);

    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_forward_coset(gnark_gpu_ntt_domain_t domain,
                                                           gnark_gpu_fr_vector_t data,
                                                           const uint64_t g[4],
                                                           const uint64_t g_half[4]) {
    if (!domain) return GNARK_GPU_ERROR_INVALID_ARG;
    return ntt_forward_coset_impl(domain, data, g, g_half, domain->ctx->stream);
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_forward_coset_stream(gnark_gpu_ntt_domain_t domain,
                                                                  gnark_gpu_fr_vector_t data,
                                                                  const uint64_t g[4],
                                                                  const uint64_t g_half[4],
                                                                  int stream_id) {
    if (!domain) return GNARK_GPU_ERROR_INVALID_ARG;
    cudaStream_t stream = get_stream(domain->ctx, stream_id);
    return ntt_forward_coset_impl(domain, data, g, g_half, stream);
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_inverse(gnark_gpu_ntt_domain_t domain,
                                                    gnark_gpu_fr_vector_t data) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_ntt_inverse(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3],
        domain->ctx->stream);

    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_bit_reverse(gnark_gpu_ntt_domain_t domain,
                                                        gnark_gpu_fr_vector_t data) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_ntt_bit_reverse(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3],
        domain->ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// PlonK2 curve-generic Fr and NTT operations
// =============================================================================

struct GnarkGPUPlonk2FrVector {
    GnarkGPUContext *ctx;
    gnark_gpu_plonk2_curve_id_t curve;
    int limbs;
    size_t count;
    uint64_t *limbs_dev[gnark_gpu::plonk2::MAX_FR_LIMBS];
};

struct GnarkGPUPlonk2NTTDomain {
    GnarkGPUContext *ctx;
    gnark_gpu_plonk2_curve_id_t curve;
    int limbs;
    size_t size;
    uint64_t *fwd[gnark_gpu::plonk2::MAX_FR_LIMBS];
    uint64_t *inv[gnark_gpu::plonk2::MAX_FR_LIMBS];
    uint64_t *inv_n;
};

struct GnarkGPUPlonk2MSM {
    GnarkGPUContext *ctx;
    gnark_gpu_plonk2_curve_id_t curve;
    int base_limbs;
    int scalar_limbs;
    size_t point_count;
    int window_bits;
    int num_windows;
    int num_buckets;
    int total_buckets;
    int reduce_bpw;
    size_t assignments_capacity;
    size_t sort_temp_bytes;
    uint64_t *d_points;
    uint64_t *d_scalars;
    uint64_t *d_out;
    uint32_t *d_keys_in;
    uint32_t *d_keys_out;
    uint32_t *d_vals_in;
    uint32_t *d_vals_out;
    uint32_t *d_bucket_offsets;
    uint32_t *d_bucket_ends;
    void *d_buckets;
    void *d_window_results;
    void *d_partial_totals;
    void *d_partial_sums;
    void *d_sort_temp;
};

static int plonk2_limbs(gnark_gpu_plonk2_curve_id_t curve) {
    return gnark_gpu::plonk2::curve_limbs(curve);
}

static int plonk2_base_limbs(gnark_gpu_plonk2_curve_id_t curve) {
    return gnark_gpu::plonk2::curve_base_limbs(curve);
}

static bool plonk2_valid_curve(gnark_gpu_plonk2_curve_id_t curve) {
    return plonk2_limbs(curve) > 0;
}

static bool plonk2_mul_overflows(size_t a, size_t b) {
    return a != 0 && b > SIZE_MAX / a;
}

static int plonk2_scalar_bits(gnark_gpu_plonk2_curve_id_t curve) {
    switch (curve) {
    case GNARK_GPU_PLONK2_CURVE_BN254:
        return 254;
    case GNARK_GPU_PLONK2_CURVE_BLS12_377:
        return 253;
    case GNARK_GPU_PLONK2_CURVE_BW6_761:
        return 377;
    default:
        return 0;
    }
}

static int plonk2_signed_window_count(int scalar_bits, int window_bits) {
    return (scalar_bits + 1 + window_bits - 1) / window_bits;
}

static int plonk2_reduce_blocks_per_window(int num_windows, int num_buckets) {
    constexpr int reduce_threads_per_window = 128;
    constexpr int finalize_threads = 32;
    int max_bpw = num_buckets / reduce_threads_per_window;
    int target_bpw = 752 / num_windows;
    int bpw = max_bpw < target_bpw ? max_bpw : target_bpw;
    if (bpw < 1) bpw = 1;
    if (bpw > finalize_threads) bpw = finalize_threads;
    return bpw;
}

static bool plonk2_msm_has_work_buffers(const GnarkGPUPlonk2MSM *msm) {
    return msm && msm->d_scalars && msm->d_out && msm->d_keys_in &&
           msm->d_keys_out && msm->d_vals_in && msm->d_vals_out &&
           msm->d_bucket_offsets && msm->d_bucket_ends && msm->d_buckets &&
           msm->d_window_results && msm->d_partial_totals &&
           msm->d_partial_sums && msm->d_sort_temp;
}

static void plonk2_msm_free_work_buffers(GnarkGPUPlonk2MSM *msm) {
    if (!msm) return;
    if (msm->d_scalars) cudaFree(msm->d_scalars);
    if (msm->d_out) cudaFree(msm->d_out);
    if (msm->d_keys_in) cudaFree(msm->d_keys_in);
    if (msm->d_keys_out) cudaFree(msm->d_keys_out);
    if (msm->d_vals_in) cudaFree(msm->d_vals_in);
    if (msm->d_vals_out) cudaFree(msm->d_vals_out);
    if (msm->d_bucket_offsets) cudaFree(msm->d_bucket_offsets);
    if (msm->d_bucket_ends) cudaFree(msm->d_bucket_ends);
    if (msm->d_buckets) cudaFree(msm->d_buckets);
    if (msm->d_window_results) cudaFree(msm->d_window_results);
    if (msm->d_partial_totals) cudaFree(msm->d_partial_totals);
    if (msm->d_partial_sums) cudaFree(msm->d_partial_sums);
    if (msm->d_sort_temp) cudaFree(msm->d_sort_temp);
    msm->d_scalars = nullptr;
    msm->d_out = nullptr;
    msm->d_keys_in = nullptr;
    msm->d_keys_out = nullptr;
    msm->d_vals_in = nullptr;
    msm->d_vals_out = nullptr;
    msm->d_bucket_offsets = nullptr;
    msm->d_bucket_ends = nullptr;
    msm->d_buckets = nullptr;
    msm->d_window_results = nullptr;
    msm->d_partial_totals = nullptr;
    msm->d_partial_sums = nullptr;
    msm->d_sort_temp = nullptr;
}

static void plonk2_msm_free_all(GnarkGPUPlonk2MSM *msm) {
    if (!msm) return;
    plonk2_msm_free_work_buffers(msm);
    if (msm->d_points) cudaFree(msm->d_points);
    msm->d_points = nullptr;
}

static cudaError_t plonk2_msm_alloc_work_buffers(GnarkGPUPlonk2MSM *m) {
    if (!m) return cudaErrorInvalidValue;
    if (plonk2_msm_has_work_buffers(m)) return cudaSuccess;
    plonk2_msm_free_work_buffers(m);

    size_t scalar_words = m->point_count * (size_t)m->scalar_limbs;
    size_t output_words = (size_t)3 * (size_t)m->base_limbs;
    size_t bucket_words =
        (size_t)m->total_buckets * (size_t)3 * (size_t)m->base_limbs;
    size_t window_words =
        (size_t)m->num_windows * (size_t)3 * (size_t)m->base_limbs;
    size_t partial_words =
        (size_t)m->num_windows * (size_t)m->reduce_bpw *
        (size_t)3 * (size_t)m->base_limbs;

    cudaError_t err = cudaMalloc(&m->d_scalars, scalar_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_out, output_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_keys_in, m->assignments_capacity * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_keys_out, m->assignments_capacity * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_vals_in, m->assignments_capacity * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_vals_out, m->assignments_capacity * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_bucket_offsets,
                     (size_t)m->total_buckets * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_bucket_ends,
                     (size_t)m->total_buckets * sizeof(uint32_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_buckets, bucket_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_window_results, window_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_partial_totals, partial_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_partial_sums, partial_words * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;
    err = cudaMalloc(&m->d_sort_temp, m->sort_temp_bytes);
    if (err != cudaSuccess) goto fail;
    return cudaSuccess;

fail:
    plonk2_msm_free_work_buffers(m);
    return err;
}

static bool plonk2_power_of_two(size_t n) {
    return n != 0 && (n & (n - 1)) == 0;
}

static gnark_gpu::plonk2::FrView plonk2_view(uint64_t *const limbs[]) {
    gnark_gpu::plonk2::FrView out{};
    for (int i = 0; i < gnark_gpu::plonk2::MAX_FR_LIMBS; i++) out.limbs[i] = limbs[i];
    return out;
}

static gnark_gpu::plonk2::ConstFrView plonk2_const_view(uint64_t *const limbs[]) {
    gnark_gpu::plonk2::ConstFrView out{};
    for (int i = 0; i < gnark_gpu::plonk2::MAX_FR_LIMBS; i++) out.limbs[i] = limbs[i];
    return out;
}

static gnark_gpu::plonk2::FrView plonk2_view(gnark_gpu_plonk2_fr_vector_t v) {
    return plonk2_view(v->limbs_dev);
}

static gnark_gpu::plonk2::ConstFrView plonk2_const_view(gnark_gpu_plonk2_fr_vector_t v) {
    return plonk2_const_view(v->limbs_dev);
}

static bool plonk2_same_shape(gnark_gpu_plonk2_fr_vector_t result,
                              gnark_gpu_plonk2_fr_vector_t a,
                              gnark_gpu_plonk2_fr_vector_t b) {
    return result->ctx == a->ctx && a->ctx == b->ctx &&
           result->curve == a->curve && a->curve == b->curve &&
           result->count == a->count && a->count == b->count;
}

static bool plonk2_same_vector_shape(gnark_gpu_plonk2_fr_vector_t a,
                                     gnark_gpu_plonk2_fr_vector_t b) {
    return a->ctx == b->ctx && a->curve == b->curve && a->count == b->count;
}

static gnark_gpu_error_t ensure_plonk2_staging_words(GnarkGPUContext *ctx, size_t words) {
    if (ctx->plonk2_staging_words >= words) return GNARK_GPU_SUCCESS;

    if (ctx->plonk2_staging_buffer) {
        cudaError_t err = cudaStreamSynchronize(ctx->stream);
        if (err != cudaSuccess) return check_cuda(err);
        cudaFree(ctx->plonk2_staging_buffer);
        ctx->plonk2_staging_buffer = nullptr;
        ctx->plonk2_staging_words = 0;
    }

    cudaError_t err = cudaMalloc(&ctx->plonk2_staging_buffer, words * sizeof(uint64_t));
    if (err != cudaSuccess) return check_cuda(err);
    ctx->plonk2_staging_words = words;
    return GNARK_GPU_SUCCESS;
}

static gnark_gpu_error_t plonk2_upload_aos(GnarkGPUContext *ctx,
                                           gnark_gpu_plonk2_curve_id_t curve,
                                           uint64_t *dst_limbs[],
                                           const uint64_t *host_data,
                                           size_t count) {
    int limbs = plonk2_limbs(curve);
    if (limbs == 0 || plonk2_mul_overflows(count, (size_t)limbs)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    size_t words = count * (size_t)limbs;
    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(ctx, words);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    cudaError_t err = cudaMemcpyAsync(ctx->plonk2_staging_buffer, host_data,
                                      words * sizeof(uint64_t),
                                      cudaMemcpyHostToDevice, ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::plonk2::launch_copy_aos_to_soa(curve, plonk2_view(dst_limbs),
                                              ctx->plonk2_staging_buffer,
                                              count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_alloc(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    size_t count,
    gnark_gpu_plonk2_fr_vector_t *vec) {
    if (!ctx || !vec || count == 0 || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    GnarkGPUPlonk2FrVector *v = new GnarkGPUPlonk2FrVector{};
    v->ctx = ctx;
    v->curve = curve_id;
    v->limbs = plonk2_limbs(curve_id);
    v->count = count;

    for (int i = 0; i < v->limbs; i++) {
        cudaError_t err = cudaMalloc(&v->limbs_dev[i], count * sizeof(uint64_t));
        if (err != cudaSuccess) {
            for (int j = 0; j < i; j++) cudaFree(v->limbs_dev[j]);
            delete v;
            cudaGetLastError();
            return check_cuda(err);
        }
    }

    *vec = v;
    return GNARK_GPU_SUCCESS;
}

extern "C" void gnark_gpu_plonk2_fr_vector_free(gnark_gpu_plonk2_fr_vector_t vec) {
    if (!vec) return;
    for (int i = 0; i < vec->limbs; i++) {
        if (vec->limbs_dev[i]) cudaFree(vec->limbs_dev[i]);
    }
    delete vec;
}

extern "C" size_t gnark_gpu_plonk2_fr_vector_len(gnark_gpu_plonk2_fr_vector_t vec) {
    return vec ? vec->count : 0;
}

extern "C" int gnark_gpu_plonk2_fr_vector_limbs(gnark_gpu_plonk2_fr_vector_t vec) {
    return vec ? vec->limbs : 0;
}

extern "C" gnark_gpu_plonk2_curve_id_t gnark_gpu_plonk2_fr_vector_curve(
    gnark_gpu_plonk2_fr_vector_t vec) {
    return vec ? vec->curve : (gnark_gpu_plonk2_curve_id_t)0;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_copy_to_device(
    gnark_gpu_plonk2_fr_vector_t vec,
    const uint64_t *host_data,
    size_t count) {
    if (!vec || !host_data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count != vec->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;
    return plonk2_upload_aos(vec->ctx, vec->curve, vec->limbs_dev, host_data, count);
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_copy_to_host(
    gnark_gpu_plonk2_fr_vector_t vec,
    uint64_t *host_data,
    size_t count) {
    if (!vec || !host_data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count != vec->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;
    if (plonk2_mul_overflows(count, (size_t)vec->limbs)) return GNARK_GPU_ERROR_INVALID_ARG;

    size_t words = count * (size_t)vec->limbs;
    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(vec->ctx, words);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    gnark_gpu::plonk2::launch_copy_soa_to_aos(vec->curve, vec->ctx->plonk2_staging_buffer,
                                              plonk2_const_view(vec), count, vec->ctx->stream);
    cudaError_t err = cudaGetLastError();
    if (err != cudaSuccess) return check_cuda(err);

    err = cudaMemcpyAsync(host_data, vec->ctx->plonk2_staging_buffer,
                          words * sizeof(uint64_t),
                          cudaMemcpyDeviceToHost, vec->ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);
    return check_cuda(cudaStreamSynchronize(vec->ctx->stream));
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_copy_d2d(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t dst,
    gnark_gpu_plonk2_fr_vector_t src) {
    if (!ctx || !dst || !src || dst->ctx != ctx || !plonk2_same_vector_shape(dst, src)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    for (int i = 0; i < dst->limbs; i++) {
        cudaError_t err = cudaMemcpyAsync(dst->limbs_dev[i], src->limbs_dev[i],
                                          dst->count * sizeof(uint64_t),
                                          cudaMemcpyDeviceToDevice, ctx->stream);
        if (err != cudaSuccess) return check_cuda(err);
    }
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_set_zero(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec) {
    if (!ctx || !vec || vec->ctx != ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    gnark_gpu::plonk2::launch_set_zero(vec->curve, plonk2_view(vec), vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_add(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t result,
    gnark_gpu_plonk2_fr_vector_t a,
    gnark_gpu_plonk2_fr_vector_t b) {
    if (!ctx || !result || !a || !b || result->ctx != ctx || !plonk2_same_shape(result, a, b)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_add(result->curve, plonk2_view(result),
                                  plonk2_const_view(a), plonk2_const_view(b),
                                  result->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_sub(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t result,
    gnark_gpu_plonk2_fr_vector_t a,
    gnark_gpu_plonk2_fr_vector_t b) {
    if (!ctx || !result || !a || !b || result->ctx != ctx || !plonk2_same_shape(result, a, b)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_sub(result->curve, plonk2_view(result),
                                  plonk2_const_view(a), plonk2_const_view(b),
                                  result->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_mul(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t result,
    gnark_gpu_plonk2_fr_vector_t a,
    gnark_gpu_plonk2_fr_vector_t b) {
    if (!ctx || !result || !a || !b || result->ctx != ctx || !plonk2_same_shape(result, a, b)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_mul(result->curve, plonk2_view(result),
                                  plonk2_const_view(a), plonk2_const_view(b),
                                  result->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_addmul(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec,
    gnark_gpu_plonk2_fr_vector_t a,
    gnark_gpu_plonk2_fr_vector_t b) {
    if (!ctx || !vec || !a || !b || vec->ctx != ctx || !plonk2_same_shape(vec, a, b)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_addmul(vec->curve, plonk2_view(vec),
                                     plonk2_const_view(a), plonk2_const_view(b),
                                     vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_scalar_mul(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec,
    const uint64_t *scalar) {
    if (!ctx || !vec || !scalar || vec->ctx != ctx) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_scalar_mul(vec->curve, plonk2_view(vec),
                                         scalar, vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_add_scalar_mul(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec,
    gnark_gpu_plonk2_fr_vector_t a,
    const uint64_t *scalar) {
    if (!ctx || !vec || !a || !scalar || vec->ctx != ctx ||
        !plonk2_same_vector_shape(vec, a)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_add_scalar_mul(vec->curve, plonk2_view(vec),
                                             plonk2_const_view(a), scalar,
                                             vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_batch_invert(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec,
    gnark_gpu_plonk2_fr_vector_t temp) {
    if (!ctx || !vec || !temp || vec->ctx != ctx || !plonk2_same_vector_shape(vec, temp)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_batch_invert(vec->curve, plonk2_view(vec),
                                           vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_butterfly4_inverse(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t b0,
    gnark_gpu_plonk2_fr_vector_t b1,
    gnark_gpu_plonk2_fr_vector_t b2,
    gnark_gpu_plonk2_fr_vector_t b3,
    const uint64_t *omega4_inv,
    const uint64_t *quarter) {
    if (!ctx || !b0 || !b1 || !b2 || !b3 || !omega4_inv || !quarter || b0->ctx != ctx ||
        !plonk2_same_vector_shape(b0, b1) || !plonk2_same_vector_shape(b0, b2) ||
        !plonk2_same_vector_shape(b0, b3)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    gnark_gpu::plonk2::launch_butterfly4_inverse(
        b0->curve, plonk2_view(b0), plonk2_view(b1), plonk2_view(b2), plonk2_view(b3),
        omega4_inv, quarter, b0->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_reduce_blinded_coset(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t dst,
    gnark_gpu_plonk2_fr_vector_t src,
    const uint64_t *tail,
    size_t tail_len,
    const uint64_t *coset_pow_n) {
    if (!ctx || !dst || !src || !tail || !coset_pow_n || dst->ctx != ctx ||
        !plonk2_same_vector_shape(dst, src)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (tail_len > dst->count) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    size_t tail_words = tail_len * (size_t)dst->limbs;
    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(ctx, tail_words);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;
    cudaError_t err = cudaMemcpyAsync(ctx->plonk2_staging_buffer, tail,
                                      tail_words * sizeof(uint64_t),
                                      cudaMemcpyHostToDevice, ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);
    gnark_gpu::plonk2::launch_reduce_blinded_coset(
        dst->curve, plonk2_view(dst), plonk2_const_view(src),
        ctx->plonk2_staging_buffer, tail_len, coset_pow_n, dst->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_compute_l1_den(
    gnark_gpu_plonk2_ntt_domain_t domain,
    gnark_gpu_plonk2_fr_vector_t out,
    const uint64_t *coset_gen) {
    if (!domain || !out || !coset_gen || domain->ctx != out->ctx ||
        domain->curve != out->curve) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (domain->size != out->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::plonk2::launch_compute_l1_den(
        domain->curve, plonk2_view(out), plonk2_const_view(domain->fwd),
        coset_gen, domain->size, domain->ctx->stream);
    return check_cuda(cudaGetLastError());
}

static bool plonk2_same_context_curve_size(gnark_gpu_plonk2_fr_vector_t a,
                                           gnark_gpu_plonk2_fr_vector_t b) {
    return a && b && a->ctx == b->ctx && a->curve == b->curve && a->count == b->count;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_gate_accum(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t result,
    gnark_gpu_plonk2_fr_vector_t ql,
    gnark_gpu_plonk2_fr_vector_t qr,
    gnark_gpu_plonk2_fr_vector_t qm,
    gnark_gpu_plonk2_fr_vector_t qo,
    gnark_gpu_plonk2_fr_vector_t qk,
    gnark_gpu_plonk2_fr_vector_t l,
    gnark_gpu_plonk2_fr_vector_t r,
    gnark_gpu_plonk2_fr_vector_t o,
    const uint64_t *zh_k_inv) {
    if (!ctx || !result || !ql || !qr || !qm || !qo || !qk || !l || !r || !o ||
        !zh_k_inv || result->ctx != ctx ||
        !plonk2_same_context_curve_size(result, ql) ||
        !plonk2_same_context_curve_size(result, qr) ||
        !plonk2_same_context_curve_size(result, qm) ||
        !plonk2_same_context_curve_size(result, qo) ||
        !plonk2_same_context_curve_size(result, qk) ||
        !plonk2_same_context_curve_size(result, l) ||
        !plonk2_same_context_curve_size(result, r) ||
        !plonk2_same_context_curve_size(result, o)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    gnark_gpu::plonk2::launch_gate_accum(
        result->curve, plonk2_view(result),
        plonk2_const_view(ql), plonk2_const_view(qr), plonk2_const_view(qm),
        plonk2_const_view(qo), plonk2_const_view(qk),
        plonk2_const_view(l), plonk2_const_view(r), plonk2_const_view(o),
        zh_k_inv, result->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_perm_boundary(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t result,
    gnark_gpu_plonk2_fr_vector_t l,
    gnark_gpu_plonk2_fr_vector_t r,
    gnark_gpu_plonk2_fr_vector_t o,
    gnark_gpu_plonk2_fr_vector_t z,
    gnark_gpu_plonk2_fr_vector_t s1,
    gnark_gpu_plonk2_fr_vector_t s2,
    gnark_gpu_plonk2_fr_vector_t s3,
    gnark_gpu_plonk2_fr_vector_t l1_den_inv,
    const uint64_t *params,
    gnark_gpu_plonk2_ntt_domain_t domain) {
    if (!ctx || !result || !l || !r || !o || !z || !s1 || !s2 || !s3 ||
        !l1_den_inv || !params || !domain || result->ctx != ctx ||
        domain->ctx != ctx || domain->curve != result->curve ||
        domain->size != result->count ||
        !plonk2_same_context_curve_size(result, l) ||
        !plonk2_same_context_curve_size(result, r) ||
        !plonk2_same_context_curve_size(result, o) ||
        !plonk2_same_context_curve_size(result, z) ||
        !plonk2_same_context_curve_size(result, s1) ||
        !plonk2_same_context_curve_size(result, s2) ||
        !plonk2_same_context_curve_size(result, s3) ||
        !plonk2_same_context_curve_size(result, l1_den_inv)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    gnark_gpu::plonk2::launch_perm_boundary(
        result->curve, plonk2_view(result),
        plonk2_const_view(l), plonk2_const_view(r), plonk2_const_view(o),
        plonk2_const_view(z), plonk2_const_view(s1), plonk2_const_view(s2),
        plonk2_const_view(s3), plonk2_const_view(l1_den_inv),
        plonk2_const_view(domain->fwd), params, result->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_z_compute_factors(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t l_inout,
    gnark_gpu_plonk2_fr_vector_t r_inout,
    gnark_gpu_plonk2_fr_vector_t o_in,
    const void *d_perm,
    const uint64_t *params,
    unsigned log2n,
    gnark_gpu_plonk2_ntt_domain_t domain) {
    if (!ctx || !l_inout || !r_inout || !o_in || !d_perm || !params || !domain ||
        l_inout->ctx != ctx || domain->ctx != ctx || domain->curve != l_inout->curve ||
        domain->size != l_inout->count ||
        !plonk2_same_context_curve_size(l_inout, r_inout) ||
        !plonk2_same_context_curve_size(l_inout, o_in)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    gnark_gpu::plonk2::launch_z_compute_factors(
        l_inout->curve, plonk2_view(l_inout), plonk2_view(r_inout),
        plonk2_const_view(o_in), static_cast<const int64_t *>(d_perm),
        plonk2_const_view(domain->fwd), params, l_inout->count, log2n,
        ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_z_prefix_phase1(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t z_vec,
    gnark_gpu_plonk2_fr_vector_t ratio_vec,
    uint64_t *chunk_products_host,
    size_t *num_chunks_out) {
    if (!ctx || !z_vec || !ratio_vec || !chunk_products_host || !num_chunks_out ||
        z_vec->ctx != ctx || !plonk2_same_vector_shape(z_vec, ratio_vec)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    size_t n = ratio_vec->count;
    size_t num_chunks = (n + 1023) / 1024;
    size_t words = num_chunks * (size_t)ratio_vec->limbs;
    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(ctx, words);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    gnark_gpu::plonk2::launch_z_prefix_phase1(
        z_vec->curve, plonk2_view(z_vec), plonk2_const_view(ratio_vec),
        ctx->plonk2_staging_buffer, n, ctx->stream);
    cudaError_t err = cudaGetLastError();
    if (err != cudaSuccess) return check_cuda(err);
    err = cudaStreamSynchronize(ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);
    err = cudaMemcpy(chunk_products_host, ctx->plonk2_staging_buffer,
                     words * sizeof(uint64_t), cudaMemcpyDeviceToHost);
    if (err != cudaSuccess) return check_cuda(err);
    *num_chunks_out = num_chunks;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_z_prefix_phase3(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t z_vec,
    gnark_gpu_plonk2_fr_vector_t temp_vec,
    const uint64_t *scanned_prefixes_host,
    size_t num_chunks) {
    if (!ctx || !z_vec || !temp_vec || !scanned_prefixes_host ||
        z_vec->ctx != ctx || !plonk2_same_vector_shape(z_vec, temp_vec)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    size_t words = num_chunks * (size_t)z_vec->limbs;
    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(ctx, words);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;
    cudaError_t err = cudaMemcpyAsync(ctx->plonk2_staging_buffer,
                                      scanned_prefixes_host,
                                      words * sizeof(uint64_t),
                                      cudaMemcpyHostToDevice, ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::plonk2::launch_z_prefix_phase3(
        z_vec->curve, plonk2_view(z_vec), plonk2_view(temp_vec),
        ctx->plonk2_staging_buffer, num_chunks, z_vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_fr_vector_scale_by_powers(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_fr_vector_t vec,
    const uint64_t *generator) {
    if (!ctx || !vec || !generator || vec->ctx != ctx) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    gnark_gpu_error_t gerr = ensure_plonk2_staging_words(ctx, 256ULL * (size_t)vec->limbs);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    gnark_gpu::plonk2::launch_scale_by_powers(
        vec->curve, plonk2_view(vec), generator, ctx->plonk2_staging_buffer,
        vec->count, ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_ntt_domain_create(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    size_t size,
    const uint64_t *fwd_twiddles_aos,
    const uint64_t *inv_twiddles_aos,
    const uint64_t *inv_n,
    gnark_gpu_plonk2_ntt_domain_t *domain) {
    if (!ctx || !domain || !inv_n || !plonk2_valid_curve(curve_id) || !plonk2_power_of_two(size)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (size > 1 && (!fwd_twiddles_aos || !inv_twiddles_aos)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    cudaError_t err = cudaSetDevice(ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    GnarkGPUPlonk2NTTDomain *d = new GnarkGPUPlonk2NTTDomain{};
    d->ctx = ctx;
    d->curve = curve_id;
    d->limbs = plonk2_limbs(curve_id);
    d->size = size;

    size_t twiddle_count = size / 2;
    for (int i = 0; i < d->limbs; i++) {
        if (twiddle_count > 0) {
            err = cudaMalloc(&d->fwd[i], twiddle_count * sizeof(uint64_t));
            if (err != cudaSuccess) goto fail;
            err = cudaMalloc(&d->inv[i], twiddle_count * sizeof(uint64_t));
            if (err != cudaSuccess) goto fail;
        }
    }

    err = cudaMalloc(&d->inv_n, (size_t)d->limbs * sizeof(uint64_t));
    if (err != cudaSuccess) goto fail;

    err = cudaMemcpyAsync(d->inv_n, inv_n, (size_t)d->limbs * sizeof(uint64_t),
                          cudaMemcpyHostToDevice, ctx->stream);
    if (err != cudaSuccess) goto fail;

    if (twiddle_count > 0) {
        gnark_gpu_error_t gerr = plonk2_upload_aos(ctx, curve_id, d->fwd,
                                                   fwd_twiddles_aos, twiddle_count);
        if (gerr != GNARK_GPU_SUCCESS) {
            gnark_gpu_plonk2_ntt_domain_destroy(d);
            return gerr;
        }
        gerr = plonk2_upload_aos(ctx, curve_id, d->inv,
                                 inv_twiddles_aos, twiddle_count);
        if (gerr != GNARK_GPU_SUCCESS) {
            gnark_gpu_plonk2_ntt_domain_destroy(d);
            return gerr;
        }
    }

    *domain = d;
    return GNARK_GPU_SUCCESS;

fail:
    gnark_gpu_plonk2_ntt_domain_destroy(d);
    cudaGetLastError();
    return check_cuda(err);
}

extern "C" void gnark_gpu_plonk2_ntt_domain_destroy(gnark_gpu_plonk2_ntt_domain_t domain) {
    if (!domain) return;
    for (int i = 0; i < domain->limbs; i++) {
        if (domain->fwd[i]) cudaFree(domain->fwd[i]);
        if (domain->inv[i]) cudaFree(domain->inv[i]);
    }
    if (domain->inv_n) cudaFree(domain->inv_n);
    delete domain;
}

extern "C" size_t gnark_gpu_plonk2_ntt_domain_size(gnark_gpu_plonk2_ntt_domain_t domain) {
    return domain ? domain->size : 0;
}

extern "C" gnark_gpu_plonk2_curve_id_t gnark_gpu_plonk2_ntt_domain_curve(
    gnark_gpu_plonk2_ntt_domain_t domain) {
    return domain ? domain->curve : (gnark_gpu_plonk2_curve_id_t)0;
}

static gnark_gpu_error_t plonk2_check_domain_vector(
    gnark_gpu_plonk2_ntt_domain_t domain,
    gnark_gpu_plonk2_fr_vector_t data) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (domain->ctx != data->ctx || domain->curve != data->curve) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (domain->size != data->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_ntt_forward(
    gnark_gpu_plonk2_ntt_domain_t domain,
    gnark_gpu_plonk2_fr_vector_t data) {
    gnark_gpu_error_t gerr = plonk2_check_domain_vector(domain, data);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;
    gnark_gpu::plonk2::launch_ntt_forward(domain->curve, plonk2_view(data),
                                          plonk2_const_view(domain->fwd),
                                          domain->size, domain->ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_ntt_inverse(
    gnark_gpu_plonk2_ntt_domain_t domain,
    gnark_gpu_plonk2_fr_vector_t data) {
    gnark_gpu_error_t gerr = plonk2_check_domain_vector(domain, data);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;
    gnark_gpu::plonk2::launch_ntt_inverse(domain->curve, plonk2_view(data),
                                          plonk2_const_view(domain->inv),
                                          domain->inv_n, domain->size,
                                          domain->ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_ntt_bit_reverse(
    gnark_gpu_plonk2_ntt_domain_t domain,
    gnark_gpu_plonk2_fr_vector_t data) {
    gnark_gpu_error_t gerr = plonk2_check_domain_vector(domain, data);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;
    gnark_gpu::plonk2::launch_bit_reverse(domain->curve, plonk2_view(data),
                                          domain->size, domain->ctx->stream);
    return check_cuda(cudaGetLastError());
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_test_g1_affine_add(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    const uint64_t *p,
    const uint64_t *q,
    uint64_t *out) {
    if (!ctx || !p || !q || !out || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    cudaError_t err = gnark_gpu::plonk2::g1_affine_add_run(
        curve_id, p, q, out, ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_test_g1_affine_double(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    const uint64_t *p,
    uint64_t *out) {
    if (!ctx || !p || !out || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    cudaError_t err = gnark_gpu::plonk2::g1_affine_double_run(
        curve_id, p, out, ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_test_msm_naive(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    const uint64_t *points,
    const uint64_t *scalars,
    size_t count,
    uint64_t *out) {
    if (!ctx || !points || !scalars || !out || count == 0 || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    cudaError_t err = gnark_gpu::plonk2::msm_naive_run(
        curve_id, points, scalars, count, out, ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_msm_pippenger(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    const uint64_t *points,
    const uint64_t *scalars,
    size_t count,
    int window_bits,
    uint64_t *out) {
    if (!ctx || !points || !scalars || !out || count == 0 || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (window_bits <= 1 || window_bits > 24) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    cudaError_t err = gnark_gpu::plonk2::msm_pippenger_run(
        curve_id, points, scalars, count, window_bits, out, ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_msm_create(
    gnark_gpu_context_t ctx,
    gnark_gpu_plonk2_curve_id_t curve_id,
    const uint64_t *points,
    size_t point_count,
    int window_bits,
    gnark_gpu_plonk2_msm_t *msm) {
    if (!ctx || !points || !msm || point_count == 0 || !plonk2_valid_curve(curve_id)) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (window_bits <= 1 || window_bits > 24) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    int base_limbs = plonk2_base_limbs(curve_id);
    int scalar_limbs = plonk2_limbs(curve_id);
    int scalar_bits = plonk2_scalar_bits(curve_id);
    if (base_limbs == 0 || scalar_limbs == 0 ||
        scalar_bits == 0 ||
        plonk2_mul_overflows(point_count, (size_t)(2 * base_limbs))) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    int num_windows = plonk2_signed_window_count(scalar_bits, window_bits);
    int num_buckets = 1 << (window_bits - 1);
    if (num_windows <= 0 || num_buckets <= 0 ||
        num_windows > INT32_MAX / num_buckets) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    int total_buckets = num_windows * num_buckets;
    if (plonk2_mul_overflows(point_count, (size_t)num_windows) ||
        point_count * (size_t)num_windows > UINT32_MAX) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    size_t assignments_capacity = point_count * (size_t)num_windows;

    GnarkGPUPlonk2MSM *m = new GnarkGPUPlonk2MSM{};
    m->ctx = ctx;
    m->curve = curve_id;
    m->base_limbs = base_limbs;
    m->scalar_limbs = scalar_limbs;
    m->point_count = point_count;
    m->window_bits = window_bits;
    m->num_windows = num_windows;
    m->num_buckets = num_buckets;
    m->total_buckets = total_buckets;
    m->reduce_bpw = plonk2_reduce_blocks_per_window(num_windows, num_buckets);
    m->assignments_capacity = assignments_capacity;

    size_t point_words = point_count * (size_t)(2 * base_limbs);
    cudaError_t err = gnark_gpu::plonk2::msm_pippenger_sort_temp_bytes(
        curve_id, point_count, window_bits, &m->sort_temp_bytes);
    if (err != cudaSuccess) {
        delete m;
        return check_cuda(err);
    }

    err = cudaMalloc(&m->d_points, point_words * sizeof(uint64_t));
    if (err != cudaSuccess) {
        delete m;
        cudaGetLastError();
        return check_cuda(err);
    }
    err = plonk2_msm_alloc_work_buffers(m);
    if (err != cudaSuccess) goto fail;

    err = cudaMemcpyAsync(m->d_points, points, point_words * sizeof(uint64_t),
                          cudaMemcpyHostToDevice, ctx->stream);
    if (err == cudaSuccess) {
        err = cudaStreamSynchronize(ctx->stream);
    }
    if (err != cudaSuccess) goto fail;

    *msm = m;
    return GNARK_GPU_SUCCESS;

fail:
    plonk2_msm_free_all(m);
    delete m;
    return check_cuda(err);
}

extern "C" void gnark_gpu_plonk2_msm_destroy(gnark_gpu_plonk2_msm_t msm) {
    if (!msm) return;
    plonk2_msm_free_all(msm);
    delete msm;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_msm_pin_work_buffers(
    gnark_gpu_plonk2_msm_t msm) {
    if (!msm) return GNARK_GPU_ERROR_INVALID_ARG;
    return check_cuda(plonk2_msm_alloc_work_buffers(msm));
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_msm_release_work_buffers(
    gnark_gpu_plonk2_msm_t msm) {
    if (!msm) return GNARK_GPU_ERROR_INVALID_ARG;
    cudaError_t err = cudaStreamSynchronize(msm->ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);
    plonk2_msm_free_work_buffers(msm);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk2_msm_run(
    gnark_gpu_plonk2_msm_t msm,
    const uint64_t *scalars,
    size_t count,
    uint64_t *out) {
    if (!msm || !scalars || !out || count == 0 || count > msm->point_count) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }
    if (!plonk2_msm_has_work_buffers(msm)) {
        cudaError_t alloc_err = plonk2_msm_alloc_work_buffers(msm);
        if (alloc_err != cudaSuccess) return check_cuda(alloc_err);
    }
    cudaError_t err = gnark_gpu::plonk2::msm_pippenger_device_points_prealloc_run(
        msm->curve, msm->d_points, scalars, count, msm->window_bits, out,
        msm->d_scalars, msm->d_out, msm->d_keys_in, msm->d_keys_out,
        msm->d_vals_in, msm->d_vals_out, msm->d_bucket_offsets,
        msm->d_bucket_ends, msm->d_buckets, msm->d_window_results,
        msm->d_partial_totals, msm->d_partial_sums, msm->d_sort_temp,
        msm->sort_temp_bytes, msm->ctx->stream);
    return check_cuda(err);
}

// =============================================================================
// New Fr vector operations (ScaleByPowers, ScalarMul, D2D copy, SetZero, AddMul)
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_scale_by_powers(gnark_gpu_context_t ctx,
                                                                   gnark_gpu_fr_vector_t v,
                                                                   const uint64_t g[4]) {
    if (!ctx || !v || !g) return GNARK_GPU_ERROR_INVALID_ARG;
    gnark_gpu::launch_scale_by_powers(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                                       g, v->count, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_scalar_mul(gnark_gpu_context_t ctx,
                                                              gnark_gpu_fr_vector_t v,
                                                              const uint64_t c[4]) {
    if (!ctx || !v || !c) return GNARK_GPU_ERROR_INVALID_ARG;
    gnark_gpu::launch_scalar_mul(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                                  c, v->count, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_d2d(gnark_gpu_context_t ctx,
                                                            gnark_gpu_fr_vector_t dst,
                                                            gnark_gpu_fr_vector_t src) {
    if (!ctx || !dst || !src) return GNARK_GPU_ERROR_INVALID_ARG;
    if (dst->count != src->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = ctx->stream;
    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMemcpyAsync(dst->limbs[i], src->limbs[i],
                                          dst->count * sizeof(uint64_t),
                                          cudaMemcpyDeviceToDevice, stream);
        if (err != cudaSuccess) return check_cuda(err);
    }
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_set_zero(gnark_gpu_context_t ctx,
                                                            gnark_gpu_fr_vector_t v) {
    if (!ctx || !v) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = ctx->stream;
    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMemsetAsync(v->limbs[i], 0,
                                          v->count * sizeof(uint64_t), stream);
        if (err != cudaSuccess) return check_cuda(err);
    }
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_addmul(gnark_gpu_context_t ctx,
                                                          gnark_gpu_fr_vector_t v,
                                                          gnark_gpu_fr_vector_t a,
                                                          gnark_gpu_fr_vector_t b) {
    if (!ctx || !v || !a || !b) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != a->count || a->count != b->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_addmul(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                              a->limbs[0], a->limbs[1], a->limbs[2], a->limbs[3],
                              b->limbs[0], b->limbs[1], b->limbs[2], b->limbs[3],
                              a->count, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_batch_invert(gnark_gpu_context_t ctx,
                                                                gnark_gpu_fr_vector_t v,
                                                                gnark_gpu_fr_vector_t temp) {
    if (!ctx || !v || !temp) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != temp->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaError_t err = gnark_gpu::launch_batch_invert(
        v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
        temp->limbs[0], temp->limbs[1], temp->limbs[2], temp->limbs[3],
        v->count, ctx->stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_butterfly4(gnark_gpu_context_t ctx,
                                                               gnark_gpu_fr_vector_t b0,
                                                               gnark_gpu_fr_vector_t b1,
                                                               gnark_gpu_fr_vector_t b2,
                                                               gnark_gpu_fr_vector_t b3,
                                                               const uint64_t omega4_inv[4],
                                                               const uint64_t quarter[4]) {
    if (!ctx || !b0 || !b1 || !b2 || !b3 || !omega4_inv || !quarter)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (b0->count != b1->count || b1->count != b2->count || b2->count != b3->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_butterfly4(
        b0->limbs[0], b0->limbs[1], b0->limbs[2], b0->limbs[3],
        b1->limbs[0], b1->limbs[1], b1->limbs[2], b1->limbs[3],
        b2->limbs[0], b2->limbs[1], b2->limbs[2], b2->limbs[3],
        b3->limbs[0], b3->limbs[1], b3->limbs[2], b3->limbs[3],
        omega4_inv, quarter, b0->count, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// PlonK fused constraint kernel
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_plonk_perm_boundary(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    gnark_gpu_fr_vector_t Z,
    gnark_gpu_fr_vector_t S1, gnark_gpu_fr_vector_t S2, gnark_gpu_fr_vector_t S3,
    gnark_gpu_fr_vector_t L1_denInv,
    const uint64_t params[28],
    gnark_gpu_ntt_domain_t domain) {
    if (!ctx || !result || !L || !R || !O || !Z || !S1 || !S2 || !S3 ||
        !L1_denInv || !params || !domain)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = result->count;
    if (L->count != n || R->count != n || O->count != n || Z->count != n ||
        S1->count != n || S2->count != n || S3->count != n || L1_denInv->count != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    if (domain->size != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    // Pack scalar params into struct. Layout: alpha, beta, gamma, l1_scalar,
    // coset_shift, coset_shift_sq, coset_gen — each 4 uint64s.
    gnark_gpu::PlonkPermBoundaryParams p;
    for (int j = 0; j < 4; j++) {
        p.alpha[j]          = params[0*4 + j];
        p.beta[j]           = params[1*4 + j];
        p.gamma[j]          = params[2*4 + j];
        p.l1_scalar[j]      = params[3*4 + j];
        p.coset_shift[j]    = params[4*4 + j];
        p.coset_shift_sq[j] = params[5*4 + j];
        p.coset_gen[j]      = params[6*4 + j];
    }

    // Get forward twiddle pointers from NTT domain via accessor
    const uint64_t *tw[4];
    gnark_gpu::ntt_get_fwd_twiddles(domain->ntt_dom, tw);

    gnark_gpu::launch_plonk_perm_boundary(
        result->limbs[0], result->limbs[1], result->limbs[2], result->limbs[3],
        L->limbs[0], L->limbs[1], L->limbs[2], L->limbs[3],
        R->limbs[0], R->limbs[1], R->limbs[2], R->limbs[3],
        O->limbs[0], O->limbs[1], O->limbs[2], O->limbs[3],
        Z->limbs[0], Z->limbs[1], Z->limbs[2], Z->limbs[3],
        S1->limbs[0], S1->limbs[1], S1->limbs[2], S1->limbs[3],
        S2->limbs[0], S2->limbs[1], S2->limbs[2], S2->limbs[3],
        S3->limbs[0], S3->limbs[1], S3->limbs[2], S3->limbs[3],
        L1_denInv->limbs[0], L1_denInv->limbs[1], L1_denInv->limbs[2], L1_denInv->limbs[3],
        p, tw[0], tw[1], tw[2], tw[3],
        n, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Device memory helpers
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_device_alloc_copy_int64(gnark_gpu_context_t ctx,
                                                                 const int64_t *host_data, size_t count,
                                                                 void **d_ptr) {
    if (!ctx || !host_data || count == 0 || !d_ptr) return GNARK_GPU_ERROR_INVALID_ARG;

    int64_t *dev_buf = nullptr;
    cudaError_t err = cudaMalloc(&dev_buf, count * sizeof(int64_t));
    if (err != cudaSuccess) return check_cuda(err);

    err = cudaMemcpyAsync(dev_buf, host_data, count * sizeof(int64_t),
                          cudaMemcpyHostToDevice, ctx->stream);
    if (err != cudaSuccess) {
        cudaFree(dev_buf);
        return check_cuda(err);
    }

    err = cudaStreamSynchronize(ctx->stream);
    if (err != cudaSuccess) {
        cudaFree(dev_buf);
        return check_cuda(err);
    }

    *d_ptr = dev_buf;
    return GNARK_GPU_SUCCESS;
}

extern "C" void gnark_gpu_device_free_ptr(void *d_ptr) {
    if (d_ptr) {
        cudaFree(d_ptr);
    }
}

// =============================================================================
// PlonK Z-polynomial ratio computation
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_plonk_z_compute_factors(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t L_inout,
    gnark_gpu_fr_vector_t R_inout,
    gnark_gpu_fr_vector_t O_in,
    const void *d_perm,
    const uint64_t params[16],
    unsigned log2n,
    gnark_gpu_ntt_domain_t domain) {
    if (!ctx || !L_inout || !R_inout || !O_in || !d_perm || !params || !domain)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = L_inout->count;
    if (R_inout->count != n || O_in->count != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    if (domain->size != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::PlonkZRatioParams p;
    for (int j = 0; j < 4; j++) {
        p.beta[j]  = params[0*4 + j];
        p.gamma[j] = params[1*4 + j];
        p.g_mul[j] = params[2*4 + j];
        p.g_sq[j]  = params[3*4 + j];
    }

    const uint64_t *tw[4];
    gnark_gpu::ntt_get_fwd_twiddles(domain->ntt_dom, tw);

    gnark_gpu::launch_plonk_z_ratio(
        L_inout->limbs[0], L_inout->limbs[1], L_inout->limbs[2], L_inout->limbs[3],
        R_inout->limbs[0], R_inout->limbs[1], R_inout->limbs[2], R_inout->limbs[3],
        O_in->limbs[0], O_in->limbs[1], O_in->limbs[2], O_in->limbs[3],
        static_cast<const int64_t *>(d_perm),
        p, tw[0], tw[1], tw[2], tw[3],
        n, log2n, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Pinned memory management
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_alloc_pinned(void **ptr, size_t bytes) {
    if (!ptr || bytes == 0) return GNARK_GPU_ERROR_INVALID_ARG;
    return check_cuda(cudaHostAlloc(ptr, bytes, cudaHostAllocDefault));
}

extern "C" void gnark_gpu_free_pinned(void *ptr) {
    if (ptr) {
        cudaFreeHost(ptr);
    }
}

// =============================================================================
// GPU L1 denominator computation
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_compute_l1_den(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t out,
    const uint64_t coset_gen[4],
    gnark_gpu_ntt_domain_t domain) {
    if (!ctx || !out || !coset_gen || !domain)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = out->count;
    if (domain->size != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    const uint64_t *tw[4];
    gnark_gpu::ntt_get_fwd_twiddles(domain->ntt_dom, tw);

    gnark_gpu::launch_compute_l1_den(
        out->limbs[0], out->limbs[1], out->limbs[2], out->limbs[3],
        tw[0], tw[1], tw[2], tw[3],
        coset_gen, n, ctx->stream);

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Patch elements (write a few AoS host elements into SoA GPU vector)
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_patch(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t vec,
    size_t offset,
    const uint64_t *host_data_aos,
    size_t count) {
    if (!ctx || !vec || !host_data_aos || count == 0)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (offset + count > vec->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = ctx->stream;

    // For each element, copy its 4 limbs to the correct SoA positions.
    // Each element in AoS is [limb0, limb1, limb2, limb3].
    for (size_t i = 0; i < count; i++) {
        for (int limb = 0; limb < 4; limb++) {
            cudaError_t err = cudaMemcpyAsync(
                vec->limbs[limb] + offset + i,
                host_data_aos + i * 4 + limb,
                sizeof(uint64_t),
                cudaMemcpyHostToDevice, stream);
            if (err != cudaSuccess) return check_cuda(err);
        }
    }

    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Synchronization
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_sync(gnark_gpu_context_t ctx) {
    if (!ctx) {
        return GNARK_GPU_ERROR_INVALID_ARG;
    }

    cudaError_t err = cudaStreamSynchronize(ctx->stream);
    return check_cuda(err);
}

// =============================================================================
// Multi-stream API
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_create_stream(gnark_gpu_context_t ctx, int stream_id) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id < 0 || stream_id >= GNARK_GPU_MAX_STREAMS) return GNARK_GPU_ERROR_INVALID_ARG;
    if (ctx->stream_created[stream_id]) return GNARK_GPU_SUCCESS; // already created

    cudaError_t err = cudaSetDevice(ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    err = cudaStreamCreate(&ctx->streams[stream_id]);
    if (err != cudaSuccess) return check_cuda(err);

    ctx->stream_created[stream_id] = true;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_record_event(gnark_gpu_context_t ctx,
                                                      int stream_id, int event_id) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    if (event_id < 0 || event_id >= GNARK_GPU_MAX_EVENTS) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    // Lazily create the event
    if (!ctx->event_created[event_id]) {
        cudaError_t err = cudaEventCreateWithFlags(&ctx->events[event_id], cudaEventDisableTiming);
        if (err != cudaSuccess) return check_cuda(err);
        ctx->event_created[event_id] = true;
    }

    cudaError_t err = cudaEventRecord(ctx->events[event_id], stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_wait_event(gnark_gpu_context_t ctx,
                                                    int stream_id, int event_id) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    if (event_id < 0 || event_id >= GNARK_GPU_MAX_EVENTS) return GNARK_GPU_ERROR_INVALID_ARG;
    if (!ctx->event_created[event_id]) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    cudaError_t err = cudaStreamWaitEvent(stream, ctx->events[event_id], 0);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_sync_stream(gnark_gpu_context_t ctx, int stream_id) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    cudaError_t err = cudaStreamSynchronize(stream);
    return check_cuda(err);
}

// =============================================================================
// Stream-aware data transfer
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_device_stream(
    gnark_gpu_fr_vector_t vec, const uint64_t *host_data,
    size_t count, int stream_id) {
    if (!vec || !host_data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count != vec->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    GnarkGPUContext *ctx = vec->ctx;
    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu_error_t gerr = ensure_staging(ctx, count);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    cudaError_t err = cudaMemcpyAsync(ctx->staging_buffer, host_data,
                                      count * 4 * sizeof(uint64_t),
                                      cudaMemcpyHostToDevice, stream);
    if (err != cudaSuccess) return check_cuda(err);

    gnark_gpu::launch_transpose_aos_to_soa_fr(vec->limbs[0], vec->limbs[1], vec->limbs[2],
                                              vec->limbs[3], ctx->staging_buffer, count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_to_host_stream(
    gnark_gpu_fr_vector_t vec, uint64_t *host_data,
    size_t count, int stream_id) {
    if (!vec || !host_data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (count != vec->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    GnarkGPUContext *ctx = vec->ctx;
    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu_error_t gerr = ensure_staging(ctx, count);
    if (gerr != GNARK_GPU_SUCCESS) return gerr;

    gnark_gpu::launch_transpose_soa_to_aos_fr(ctx->staging_buffer, vec->limbs[0], vec->limbs[1],
                                              vec->limbs[2], vec->limbs[3], count, stream);

    cudaError_t err = cudaMemcpyAsync(host_data, ctx->staging_buffer,
                                      count * 4 * sizeof(uint64_t),
                                      cudaMemcpyDeviceToHost, stream);
    if (err != cudaSuccess) return check_cuda(err);

    err = cudaStreamSynchronize(stream);
    return check_cuda(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_copy_d2d_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src, int stream_id) {
    if (!ctx || !dst || !src) return GNARK_GPU_ERROR_INVALID_ARG;
    if (dst->count != src->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMemcpyAsync(dst->limbs[i], src->limbs[i],
                                          dst->count * sizeof(uint64_t),
                                          cudaMemcpyDeviceToDevice, stream);
        if (err != cudaSuccess) return check_cuda(err);
    }
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Stream-aware NTT operations
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_ntt_forward_stream(gnark_gpu_ntt_domain_t domain,
                                                            gnark_gpu_fr_vector_t data,
                                                            int stream_id) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(domain->ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = domain->ctx->streams[0];

    gnark_gpu::launch_ntt_forward(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3], stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_inverse_stream(gnark_gpu_ntt_domain_t domain,
                                                            gnark_gpu_fr_vector_t data,
                                                            int stream_id) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(domain->ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = domain->ctx->streams[0];

    gnark_gpu::launch_ntt_inverse(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3], stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_ntt_bit_reverse_stream(gnark_gpu_ntt_domain_t domain,
                                                                gnark_gpu_fr_vector_t data,
                                                                int stream_id) {
    if (!domain || !data) return GNARK_GPU_ERROR_INVALID_ARG;
    if (data->count != domain->size) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(domain->ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = domain->ctx->streams[0];

    gnark_gpu::launch_ntt_bit_reverse(domain->ntt_dom,
        data->limbs[0], data->limbs[1], data->limbs[2], data->limbs[3], stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Stream-aware arithmetic operations
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_scale_by_powers_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    const uint64_t g[4], int stream_id) {
    if (!ctx || !v || !g) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_scale_by_powers(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                                       g, v->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_scalar_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    const uint64_t c[4], int stream_id) {
    if (!ctx || !v || !c) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_scalar_mul(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                                  c, v->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_set_zero_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v, int stream_id) {
    if (!ctx || !v) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMemsetAsync(v->limbs[i], 0,
                                          v->count * sizeof(uint64_t), stream);
        if (err != cudaSuccess) return check_cuda(err);
    }
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_add_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id) {
    if (!ctx || !result || !a || !b) return GNARK_GPU_ERROR_INVALID_ARG;
    if (result->count != a->count || a->count != b->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_add_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                             result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                             a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                             b->limbs[3], a->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_sub_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id) {
    if (!ctx || !result || !a || !b) return GNARK_GPU_ERROR_INVALID_ARG;
    if (result->count != a->count || a->count != b->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_sub_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                             result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                             a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                             b->limbs[3], a->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id) {
    if (!ctx || !result || !a || !b) return GNARK_GPU_ERROR_INVALID_ARG;
    if (result->count != a->count || a->count != b->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_mul_mont_fr(result->limbs[0], result->limbs[1], result->limbs[2],
                                  result->limbs[3], a->limbs[0], a->limbs[1], a->limbs[2],
                                  a->limbs[3], b->limbs[0], b->limbs[1], b->limbs[2],
                                  b->limbs[3], a->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_addmul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, gnark_gpu_fr_vector_t b, int stream_id) {
    if (!ctx || !v || !a || !b) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != a->count || a->count != b->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_addmul(v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
                              a->limbs[0], a->limbs[1], a->limbs[2], a->limbs[3],
                              b->limbs[0], b->limbs[1], b->limbs[2], b->limbs[3],
                              a->count, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_batch_invert_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t temp, int stream_id) {
    if (!ctx || !v || !temp) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != temp->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    cudaError_t err = gnark_gpu::launch_batch_invert(
        v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
        temp->limbs[0], temp->limbs[1], temp->limbs[2], temp->limbs[3],
        v->count, stream);
    return check_cuda(err);
}

// =============================================================================
// AddScalarMul: v[i] += a[i] * scalar
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_add_scalar_mul(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, const uint64_t scalar[4]) {
    if (!ctx || !v || !a || !scalar) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != a->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_add_scalar_mul(
        v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
        a->limbs[0], a->limbs[1], a->limbs[2], a->limbs[3],
        scalar, a->count, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_fr_vector_add_scalar_mul_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t v,
    gnark_gpu_fr_vector_t a, const uint64_t scalar[4], int stream_id) {
    if (!ctx || !v || !a || !scalar) return GNARK_GPU_ERROR_INVALID_ARG;
    if (v->count != a->count) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::launch_add_scalar_mul(
        v->limbs[0], v->limbs[1], v->limbs[2], v->limbs[3],
        a->limbs[0], a->limbs[1], a->limbs[2], a->limbs[3],
        scalar, a->count, stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Stream-aware PlonK operations
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_compute_l1_den_stream(
    gnark_gpu_context_t ctx, gnark_gpu_fr_vector_t out,
    const uint64_t coset_gen[4], gnark_gpu_ntt_domain_t domain,
    int stream_id) {
    if (!ctx || !out || !coset_gen || !domain) return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = out->count;
    if (domain->size != n) return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    const uint64_t *tw[4];
    gnark_gpu::ntt_get_fwd_twiddles(domain->ntt_dom, tw);

    gnark_gpu::launch_compute_l1_den(
        out->limbs[0], out->limbs[1], out->limbs[2], out->limbs[3],
        tw[0], tw[1], tw[2], tw[3],
        coset_gen, n, stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_plonk_perm_boundary_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    gnark_gpu_fr_vector_t Z,
    gnark_gpu_fr_vector_t S1, gnark_gpu_fr_vector_t S2, gnark_gpu_fr_vector_t S3,
    gnark_gpu_fr_vector_t L1_denInv,
    const uint64_t params[28],
    gnark_gpu_ntt_domain_t domain, int stream_id) {
    if (!ctx || !result || !L || !R || !O || !Z || !S1 || !S2 || !S3 ||
        !L1_denInv || !params || !domain)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = result->count;
    if (L->count != n || R->count != n || O->count != n || Z->count != n ||
        S1->count != n || S2->count != n || S3->count != n || L1_denInv->count != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    if (domain->size != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    cudaStream_t stream = get_stream(ctx, stream_id);
    if (!stream && stream_id != 0) return GNARK_GPU_ERROR_INVALID_ARG;
    if (stream_id == 0) stream = ctx->streams[0];

    gnark_gpu::PlonkPermBoundaryParams p;
    for (int j = 0; j < 4; j++) {
        p.alpha[j]          = params[0*4 + j];
        p.beta[j]           = params[1*4 + j];
        p.gamma[j]          = params[2*4 + j];
        p.l1_scalar[j]      = params[3*4 + j];
        p.coset_shift[j]    = params[4*4 + j];
        p.coset_shift_sq[j] = params[5*4 + j];
        p.coset_gen[j]      = params[6*4 + j];
    }

    const uint64_t *tw[4];
    gnark_gpu::ntt_get_fwd_twiddles(domain->ntt_dom, tw);

    gnark_gpu::launch_plonk_perm_boundary(
        result->limbs[0], result->limbs[1], result->limbs[2], result->limbs[3],
        L->limbs[0], L->limbs[1], L->limbs[2], L->limbs[3],
        R->limbs[0], R->limbs[1], R->limbs[2], R->limbs[3],
        O->limbs[0], O->limbs[1], O->limbs[2], O->limbs[3],
        Z->limbs[0], Z->limbs[1], Z->limbs[2], Z->limbs[3],
        S1->limbs[0], S1->limbs[1], S1->limbs[2], S1->limbs[3],
        S2->limbs[0], S2->limbs[1], S2->limbs[2], S2->limbs[3],
        S3->limbs[0], S3->limbs[1], S3->limbs[2], S3->limbs[3],
        L1_denInv->limbs[0], L1_denInv->limbs[1], L1_denInv->limbs[2], L1_denInv->limbs[3],
        p, tw[0], tw[1], tw[2], tw[3],
        n, stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// GPU Z prefix product
// =============================================================================

// Helper: ensure Z prefix scratch buffers are large enough.
static cudaError_t z_prefix_scratch_ensure(ZPrefixScratch &s, size_t num_chunks) {
    if (num_chunks <= s.capacity) return cudaSuccess;

    for (int i = 0; i < 4; i++) {
        if (s.cp[i]) { cudaFree(s.cp[i]); s.cp[i] = nullptr; }
        if (s.sp[i]) { cudaFree(s.sp[i]); s.sp[i] = nullptr; }
    }
    s.capacity = 0;

    size_t alloc = num_chunks < 64 ? 64 : num_chunks;
    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMalloc(&s.cp[i], alloc * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
        err = cudaMalloc(&s.sp[i], alloc * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
    }
    s.capacity = alloc;
    return cudaSuccess;
}

extern "C" gnark_gpu_error_t gnark_gpu_z_prefix_phase1(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t z_vec,
    gnark_gpu_fr_vector_t ratio_vec,
    uint64_t *chunk_products_host,
    size_t *num_chunks_out) {
    if (!ctx || !z_vec || !ratio_vec || !chunk_products_host || !num_chunks_out)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (z_vec->count != ratio_vec->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    size_t n = ratio_vec->count;
    size_t num_chunks = (n + 1023) / 1024;

    // Ensure context-owned scratch buffers are large enough.
    cudaError_t err = z_prefix_scratch_ensure(ctx->z_prefix_scratch, num_chunks);
    if (err != cudaSuccess) return check_cuda(err);

    err = gnark_gpu::launch_z_prefix_phase1(
        z_vec->limbs[0], z_vec->limbs[1], z_vec->limbs[2], z_vec->limbs[3],
        ratio_vec->limbs[0], ratio_vec->limbs[1], ratio_vec->limbs[2], ratio_vec->limbs[3],
        ctx->z_prefix_scratch.cp, n, ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    // Sync to ensure kernel is done before downloading chunk products.
    err = cudaStreamSynchronize(ctx->stream);
    if (err != cudaSuccess) return check_cuda(err);

    // Bulk download: 4 per-limb cudaMemcpy + host SoA→AoS transpose.
    // Reuse a temporary host buffer for per-limb contiguous data.
    std::vector<uint64_t> limb_buf(num_chunks);
    for (int limb = 0; limb < 4; limb++) {
        err = cudaMemcpy(limb_buf.data(), ctx->z_prefix_scratch.cp[limb],
                         num_chunks * sizeof(uint64_t), cudaMemcpyDeviceToHost);
        if (err != cudaSuccess) return check_cuda(err);
        // Scatter into AoS host layout: cpHost[c*4 + limb] = limb_buf[c]
        for (size_t c = 0; c < num_chunks; c++) {
            chunk_products_host[c * 4 + limb] = limb_buf[c];
        }
    }

    *num_chunks_out = num_chunks;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_z_prefix_phase3(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t z_vec,
    gnark_gpu_fr_vector_t temp_vec,
    const uint64_t *scanned_prefixes_host,
    size_t num_chunks) {
    if (!ctx || !z_vec || !temp_vec || !scanned_prefixes_host)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (z_vec->count != temp_vec->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    size_t n = z_vec->count;

    // Ensure scratch is available (should already be from phase1, but be safe).
    cudaError_t err = z_prefix_scratch_ensure(ctx->z_prefix_scratch, num_chunks);
    if (err != cudaSuccess) return check_cuda(err);

    // Bulk upload: host AoS → gather per-limb → 4 cudaMemcpy.
    std::vector<uint64_t> limb_buf(num_chunks);
    for (int limb = 0; limb < 4; limb++) {
        for (size_t c = 0; c < num_chunks; c++) {
            limb_buf[c] = scanned_prefixes_host[c * 4 + limb];
        }
        err = cudaMemcpy(ctx->z_prefix_scratch.sp[limb], limb_buf.data(),
                         num_chunks * sizeof(uint64_t), cudaMemcpyHostToDevice);
        if (err != cudaSuccess) return check_cuda(err);
    }

    err = gnark_gpu::launch_z_prefix_phase3(
        z_vec->limbs[0], z_vec->limbs[1], z_vec->limbs[2], z_vec->limbs[3],
        temp_vec->limbs[0], temp_vec->limbs[1], temp_vec->limbs[2], temp_vec->limbs[3],
        ctx->z_prefix_scratch.sp, num_chunks, n, ctx->stream);
    return check_cuda(err);
}

// =============================================================================
// GPU polynomial evaluation (chunked Horner)
// =============================================================================

// Helper: ensure poly eval scratch buffers are large enough.
static cudaError_t poly_eval_scratch_ensure(PolyEvalScratch &s, size_t num_chunks) {
    if (num_chunks <= s.capacity) return cudaSuccess;

    for (int i = 0; i < 4; i++) {
        if (s.out[i]) { cudaFree(s.out[i]); s.out[i] = nullptr; }
    }
    s.capacity = 0;

    size_t alloc = num_chunks < 64 ? 64 : num_chunks;
    for (int i = 0; i < 4; i++) {
        cudaError_t err = cudaMalloc(&s.out[i], alloc * sizeof(uint64_t));
        if (err != cudaSuccess) return err;
    }
    s.capacity = alloc;
    return cudaSuccess;
}

static gnark_gpu_error_t poly_eval_chunks_impl(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t coeffs,
    const uint64_t z[4],
    uint64_t *partials_host,
    size_t *num_chunks_out,
    cudaStream_t stream) {
    if (!ctx || !coeffs || !z || !partials_host || !num_chunks_out)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = coeffs->count;
    if (n == 0) {
        *num_chunks_out = 0;
        return GNARK_GPU_SUCCESS;
    }

    size_t nc = (n + 1023) / 1024;

    // Ensure context-owned scratch buffers are large enough.
    cudaError_t err = poly_eval_scratch_ensure(ctx->poly_eval_scratch, nc);
    if (err != cudaSuccess) return check_cuda(err);

    uint64_t *d_out[4] = {ctx->poly_eval_scratch.out[0], ctx->poly_eval_scratch.out[1],
                           ctx->poly_eval_scratch.out[2], ctx->poly_eval_scratch.out[3]};

    size_t nc_out;
    gnark_gpu::launch_poly_eval_chunks(
        coeffs->limbs[0], coeffs->limbs[1], coeffs->limbs[2], coeffs->limbs[3],
        z, d_out[0], d_out[1], d_out[2], d_out[3],
        n, &nc_out, stream);

    // Synchronize to ensure kernel is done before downloading
    err = cudaStreamSynchronize(stream);
    if (err != cudaSuccess) return check_cuda(err);

    // Bulk download: 4 per-limb cudaMemcpy + host SoA→AoS transpose.
    std::vector<uint64_t> limb_buf(nc_out);
    for (int limb = 0; limb < 4; limb++) {
        err = cudaMemcpy(limb_buf.data(), d_out[limb],
                         nc_out * sizeof(uint64_t), cudaMemcpyDeviceToHost);
        if (err != cudaSuccess) return check_cuda(err);
        for (size_t c = 0; c < nc_out; c++) {
            partials_host[c * 4 + limb] = limb_buf[c];
        }
    }

    *num_chunks_out = nc_out;
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_poly_eval_chunks(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t coeffs,
    const uint64_t z[4],
    uint64_t *partials_host,
    size_t *num_chunks_out) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    return poly_eval_chunks_impl(ctx, coeffs, z, partials_host, num_chunks_out, ctx->stream);
}

extern "C" gnark_gpu_error_t gnark_gpu_poly_eval_chunks_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t coeffs,
    const uint64_t z[4],
    uint64_t *partials_host,
    size_t *num_chunks_out,
    int stream_id) {
    if (!ctx) return GNARK_GPU_ERROR_INVALID_ARG;
    cudaStream_t stream = get_stream(ctx, stream_id);
    return poly_eval_chunks_impl(ctx, coeffs, z, partials_host, num_chunks_out, stream);
}

// =============================================================================
// Fused gate constraint accumulation for PlonK quotient
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_plonk_gate_accum(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t result,
    gnark_gpu_fr_vector_t Ql, gnark_gpu_fr_vector_t Qr,
    gnark_gpu_fr_vector_t Qm, gnark_gpu_fr_vector_t Qo,
    gnark_gpu_fr_vector_t Qk,
    gnark_gpu_fr_vector_t L, gnark_gpu_fr_vector_t R, gnark_gpu_fr_vector_t O,
    const uint64_t zhKInv[4]) {
    if (!ctx || !result || !Ql || !Qr || !Qm || !Qo || !Qk ||
        !L || !R || !O || !zhKInv)
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = result->count;
    if (Ql->count != n || Qr->count != n || Qm->count != n || Qo->count != n ||
        Qk->count != n || L->count != n || R->count != n || O->count != n)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    gnark_gpu::launch_plonk_gate_accum(
        result->limbs[0], result->limbs[1], result->limbs[2], result->limbs[3],
        Ql->limbs[0], Ql->limbs[1], Ql->limbs[2], Ql->limbs[3],
        Qr->limbs[0], Qr->limbs[1], Qr->limbs[2], Qr->limbs[3],
        Qm->limbs[0], Qm->limbs[1], Qm->limbs[2], Qm->limbs[3],
        Qo->limbs[0], Qo->limbs[1], Qo->limbs[2], Qo->limbs[3],
        Qk->limbs[0], Qk->limbs[1], Qk->limbs[2], Qk->limbs[3],
        L->limbs[0], L->limbs[1], L->limbs[2], L->limbs[3],
        R->limbs[0], R->limbs[1], R->limbs[2], R->limbs[3],
        O->limbs[0], O->limbs[1], O->limbs[2], O->limbs[3],
        zhKInv, n, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// Reduce blinded polynomial for coset evaluation
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_reduce_blinded_coset(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src,
    const uint64_t *blinding_tail_host,
    size_t tail_len,
    const uint64_t cosetPowN[4]) {
    if (!ctx || !dst || !src || !cosetPowN)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (dst->count != src->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;

    size_t n = src->count;

    // Upload tiny tail to device (typically 2-3 elements = 64-96 bytes)
    uint64_t *d_tail = nullptr;
    if (tail_len > 0 && blinding_tail_host) {
        size_t tail_bytes = tail_len * 4 * sizeof(uint64_t);
        auto err = cudaMallocAsync(&d_tail, tail_bytes, ctx->stream);
        if (err != cudaSuccess) return check_cuda(err);
        err = cudaMemcpyAsync(d_tail, blinding_tail_host, tail_bytes,
                               cudaMemcpyHostToDevice, ctx->stream);
        if (err != cudaSuccess) { cudaFreeAsync(d_tail, ctx->stream); return check_cuda(err); }
    }

    gnark_gpu::launch_reduce_blinded_coset(
        dst->limbs[0], dst->limbs[1], dst->limbs[2], dst->limbs[3],
        src->limbs[0], src->limbs[1], src->limbs[2], src->limbs[3],
        cosetPowN, d_tail, (uint32_t)tail_len, (uint32_t)n, ctx->stream);

    if (d_tail) cudaFreeAsync(d_tail, ctx->stream);
    return GNARK_GPU_SUCCESS;
}

extern "C" gnark_gpu_error_t gnark_gpu_reduce_blinded_coset_stream(
    gnark_gpu_context_t ctx,
    gnark_gpu_fr_vector_t dst,
    gnark_gpu_fr_vector_t src,
    const uint64_t *blinding_tail_host,
    size_t tail_len,
    const uint64_t cosetPowN[4],
    int stream_id) {
    if (!ctx || !dst || !src || !cosetPowN)
        return GNARK_GPU_ERROR_INVALID_ARG;
    if (dst->count != src->count)
        return GNARK_GPU_ERROR_SIZE_MISMATCH;
    if (stream_id < 0 || stream_id >= GNARK_GPU_MAX_STREAMS || !ctx->stream_created[stream_id])
        return GNARK_GPU_ERROR_INVALID_ARG;

    size_t n = src->count;
    cudaStream_t stream = ctx->streams[stream_id];

    uint64_t *d_tail = nullptr;
    if (tail_len > 0 && blinding_tail_host) {
        size_t tail_bytes = tail_len * 4 * sizeof(uint64_t);
        auto err = cudaMallocAsync(&d_tail, tail_bytes, stream);
        if (err != cudaSuccess) return check_cuda(err);
        err = cudaMemcpyAsync(d_tail, blinding_tail_host, tail_bytes,
                               cudaMemcpyHostToDevice, stream);
        if (err != cudaSuccess) { cudaFreeAsync(d_tail, stream); return check_cuda(err); }
    }

    gnark_gpu::launch_reduce_blinded_coset(
        dst->limbs[0], dst->limbs[1], dst->limbs[2], dst->limbs[3],
        src->limbs[0], src->limbs[1], src->limbs[2], src->limbs[3],
        cosetPowN, d_tail, (uint32_t)tail_len, (uint32_t)n, stream);

    if (d_tail) cudaFreeAsync(d_tail, stream);
    return GNARK_GPU_SUCCESS;
}

// =============================================================================
// GPU memory info
// =============================================================================

extern "C" gnark_gpu_error_t gnark_gpu_mem_get_info(gnark_gpu_context_t ctx,
                                                      size_t *free_bytes, size_t *total_bytes) {
    if (!ctx || !free_bytes || !total_bytes) return GNARK_GPU_ERROR_INVALID_ARG;

    cudaError_t err = cudaSetDevice(ctx->device_id);
    if (err != cudaSuccess) return check_cuda(err);

    err = cudaMemGetInfo(free_bytes, total_bytes);
    return check_cuda(err);
}
