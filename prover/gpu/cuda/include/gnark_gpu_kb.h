// gnark-gpu KoalaBear + Vortex C API
//
// KoalaBear: P = 2³¹ − 2²⁴ + 1, single uint32 Montgomery elements.
// Vectors are flat uint32_t arrays (no SoA/AoS distinction for 1-limb field).
// E4 elements are 4 consecutive uint32: (b0.a0, b0.a1, b1.a0, b1.a1).

#ifndef GNARK_GPU_KB_H
#define GNARK_GPU_KB_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Reuse context from main gnark_gpu if available, otherwise define here.
#ifndef GNARK_GPU_H
typedef struct GnarkGPUContext *gnark_gpu_context_t;
typedef enum {
    KB_SUCCESS         = 0,
    KB_ERROR_CUDA      = 1,
    KB_ERROR_INVALID   = 2,
    KB_ERROR_OOM       = 3,
    KB_ERROR_SIZE      = 4,
} kb_error_t;
#else
typedef gnark_gpu_error_t kb_error_t;
#define KB_SUCCESS         GNARK_GPU_SUCCESS
#define KB_ERROR_CUDA      GNARK_GPU_ERROR_CUDA
#define KB_ERROR_INVALID   GNARK_GPU_ERROR_INVALID_ARG
#define KB_ERROR_OOM       GNARK_GPU_ERROR_OUT_OF_MEMORY
#define KB_ERROR_SIZE      GNARK_GPU_ERROR_SIZE_MISMATCH
#endif

// ═══════════════════════════════════════════════════════════════════════════
// KoalaBear vector (flat uint32 on GPU)
// ═══════════════════════════════════════════════════════════════════════════

typedef struct KBVec *kb_vec_t;

kb_error_t kb_vec_alloc(gnark_gpu_context_t ctx, size_t n, kb_vec_t *out);
void       kb_vec_free (kb_vec_t v);
size_t     kb_vec_len  (kb_vec_t v);

kb_error_t kb_vec_h2d(gnark_gpu_context_t ctx, kb_vec_t dst,
                      const uint32_t *src, size_t n);
kb_error_t kb_vec_h2d_pinned(gnark_gpu_context_t ctx, kb_vec_t dst,
                              const uint32_t *src, size_t n);

// Page-locked host memory for fast H2D transfers.
kb_error_t kb_pinned_alloc(size_t bytes, uint32_t **out);
void       kb_pinned_free (uint32_t *ptr);
kb_error_t kb_vec_d2h(gnark_gpu_context_t ctx, uint32_t *dst,
                      kb_vec_t src, size_t n);
kb_error_t kb_vec_d2d(gnark_gpu_context_t ctx, kb_vec_t dst, kb_vec_t src);
kb_error_t kb_vec_d2d_offset(gnark_gpu_context_t ctx, uint32_t *dst,
                              const uint32_t *src, size_t n);
kb_error_t kb_vec_d2h_raw(gnark_gpu_context_t ctx, uint32_t *dst,
                           const uint32_t *src, size_t n);
kb_error_t kb_sync(gnark_gpu_context_t ctx);

kb_error_t kb_vec_add(gnark_gpu_context_t ctx, kb_vec_t c, kb_vec_t a, kb_vec_t b);
kb_error_t kb_vec_sub(gnark_gpu_context_t ctx, kb_vec_t c, kb_vec_t a, kb_vec_t b);
kb_error_t kb_vec_mul(gnark_gpu_context_t ctx, kb_vec_t c, kb_vec_t a, kb_vec_t b);
kb_error_t kb_vec_scale(gnark_gpu_context_t ctx, kb_vec_t v, uint32_t scalar);
kb_error_t kb_vec_scale_by_powers(gnark_gpu_context_t ctx, kb_vec_t v, uint32_t g);
kb_error_t kb_vec_batch_invert(gnark_gpu_context_t ctx, kb_vec_t v, kb_vec_t temp);
kb_error_t kb_vec_bitrev(gnark_gpu_context_t ctx, kb_vec_t v);

// ═══════════════════════════════════════════════════════════════════════════
// NTT domain
// ═══════════════════════════════════════════════════════════════════════════

typedef struct KBNtt *kb_ntt_t;

kb_error_t kb_ntt_init(gnark_gpu_context_t ctx, size_t n,
                       const uint32_t *fwd_twiddles,
                       const uint32_t *inv_twiddles,
                       kb_ntt_t *out);
void       kb_ntt_free(kb_ntt_t d);

kb_error_t kb_ntt_fwd(gnark_gpu_context_t ctx, kb_ntt_t d, kb_vec_t v);
kb_error_t kb_ntt_inv(gnark_gpu_context_t ctx, kb_ntt_t d, kb_vec_t v);
kb_error_t kb_ntt_coset_fwd(gnark_gpu_context_t ctx, kb_ntt_t d,
                            kb_vec_t v, uint32_t g);
kb_error_t kb_ntt_coset_fwd_raw(gnark_gpu_context_t ctx, kb_ntt_t d,
                                 uint32_t *data, uint32_t g);
kb_error_t kb_vec_bitrev_raw(gnark_gpu_context_t ctx, uint32_t *data, size_t n);

// Batch NTT: operates on `batch` vectors packed contiguously in `data`.
// Each vector has `n` uint32 elements. Single call, all kernels queued async.
kb_error_t kb_ntt_batch_coset_fwd_bitrev(gnark_gpu_context_t ctx, kb_ntt_t d,
                                          uint32_t *data, size_t n, size_t batch,
                                          uint32_t g);
kb_error_t kb_ntt_batch_ifft_scale(gnark_gpu_context_t ctx, kb_ntt_t d,
                                    uint32_t *data, size_t n, size_t batch,
                                    uint32_t nInv);

// ═══════════════════════════════════════════════════════════════════════════
// Poseidon2 (batch operations)
// ═══════════════════════════════════════════════════════════════════════════

typedef struct KBPoseidon2 *kb_p2_t;

kb_error_t kb_p2_init(gnark_gpu_context_t ctx, int width,
                      int nb_full_rounds, int nb_partial_rounds,
                      const uint32_t *round_keys,
                      const uint32_t *diag,
                      kb_p2_t *out);
void       kb_p2_free(kb_p2_t p);

kb_error_t kb_p2_compress_batch(gnark_gpu_context_t ctx, kb_p2_t p,
                                const uint32_t *input, uint32_t *output,
                                size_t count);

kb_error_t kb_p2_sponge_batch(gnark_gpu_context_t ctx, kb_p2_t p,
                              const uint32_t *input, size_t input_len,
                              uint32_t *output, size_t count);

// ═══════════════════════════════════════════════════════════════════════════
// Ring-SIS hash
// ═══════════════════════════════════════════════════════════════════════════

typedef struct KBSis *kb_sis_t;

kb_error_t kb_sis_init(gnark_gpu_context_t ctx,
                       int degree, int n_polys, int log_two_bound,
                       const uint32_t *ag,
                       const uint32_t *fwd_tw,
                       const uint32_t *inv_tw,
                       const uint32_t *coset_table,
                       const uint32_t *coset_inv,
                       kb_sis_t *out);
void       kb_sis_free(kb_sis_t s);

// ═══════════════════════════════════════════════════════════════════════════
// Vortex commit pipeline
// ═══════════════════════════════════════════════════════════════════════════

kb_error_t kb_merkle_build(gnark_gpu_context_t ctx, kb_p2_t p,
                           const uint32_t *leaves, size_t n_leaves,
                           uint32_t *tree_buf);

// Pre-allocated pipeline: GPU RS encode + SIS + Merkle in one call.
// RS encode runs on GPU via batch NTT (eliminates CPU RS, halves H2D data).
typedef struct KBVortexPipeline *kb_vortex_pipeline_t;

kb_error_t kb_vortex_pipeline_init(gnark_gpu_context_t ctx,
                                    kb_sis_t sis,
                                    kb_p2_t p2_sponge,
                                    kb_p2_t p2_compress,
                                    size_t max_n_rows,
                                    size_t n_cols,
                                    int rate,
                                    const uint32_t *rs_fwd_tw,
                                    const uint32_t *rs_inv_tw,
                                    const uint32_t *scaled_coset_br,
                                    kb_vortex_pipeline_t *out);
void       kb_vortex_pipeline_free(kb_vortex_pipeline_t p);

// Pinned host buffer accessors (for zero-copy Go slice wrapping).
uint32_t *kb_vortex_pipeline_input_buf(kb_vortex_pipeline_t p);
uint32_t *kb_vortex_pipeline_tree_buf(kb_vortex_pipeline_t p);

// Full GPU vortex commit: RS encode + SIS hash + Merkle tree.
//   raw_rows: [n_rows × n_cols], host pinned (from input_buf), Montgomery.
//   SIS hashes stay on device (sponge reads d_sis).  Tree → pinned h_tree.
//   Encoded matrix retained on GPU for Prove (lincomb, extract_col).
// Set KB_VORTEX_TIMING=1 for per-phase timing on stderr.
kb_error_t kb_vortex_commit(kb_vortex_pipeline_t pipeline,
                            const uint32_t *raw_rows,
                            size_t n_rows);

// GPU Prove: linear combination on encoded matrix (kept on device after commit).
//   result[j] = Σᵢ αⁱ · encoded[j][i] ∈ E4,  j ∈ [0, scw)
kb_error_t kb_vortex_lincomb(kb_vortex_pipeline_t pipeline,
                              size_t n_rows,
                              const uint32_t alpha[4],
                              uint32_t *result);

// GPU Prove: extract single column from encoded matrix to host.
kb_error_t kb_vortex_extract_col(kb_vortex_pipeline_t pipeline,
                                  size_t n_rows, int col_idx,
                                  uint32_t *out);

// Extract full encoded matrix from GPU to host in column-major layout.
//   out: [scw × n_rows] uint32, column-major: out[col * n_rows + row].
kb_error_t kb_vortex_extract_all(kb_vortex_pipeline_t pipeline,
                                  size_t n_rows, uint32_t *out);

// Extract full encoded matrix from GPU to host in row-major layout.
// Transposes on GPU before D2H to avoid costly CPU transposition.
//   out: [n_rows × scw] uint32, row-major: out[row * scw + col].
kb_error_t kb_vortex_extract_all_rowmajor(kb_vortex_pipeline_t pipeline,
                                           size_t n_rows, uint32_t *out);

// Return sizeCodeWord for the pipeline.
size_t     kb_vortex_scw(kb_vortex_pipeline_t pipeline);

// Set multi-coset scaling tables for rate > 2 RS encoding.
// coset_tables: [(rate-1) × n_cols] flat array, table k at offset k*n_cols.
// Each table k: coset_k_br[j] = (Ω^{k+1})^{bitrev(j)} / n  (bit-reversed, normalized).
// Must be called before kb_vortex_commit when rate > 2.
kb_error_t kb_vortex_pipeline_set_coset_tables(kb_vortex_pipeline_t p,
                                                const uint32_t *coset_tables,
                                                size_t n_tables);

// Commit + async extract: overlaps D2H of encoded/SIS/leaves with compute.
// After return, pinned host buffers contain the results.
// Use kb_vortex_h_*_pinned() to get pointers to the pinned buffers.
kb_error_t kb_vortex_commit_and_extract(kb_vortex_pipeline_t pipeline,
                                         const uint32_t *raw_rows,
                                         size_t n_rows);

// Accessors for pinned host buffers (valid after kb_vortex_commit_and_extract).
uint32_t *kb_vortex_h_enc_pinned(kb_vortex_pipeline_t pipeline);
uint32_t *kb_vortex_h_sis_pinned(kb_vortex_pipeline_t pipeline);
uint32_t *kb_vortex_h_leaves_pinned(kb_vortex_pipeline_t pipeline);

// Extract SIS column hashes from GPU to host.
//   out: flat [scw × degree] uint32, same layout as d_sis.
kb_error_t kb_vortex_extract_sis(kb_vortex_pipeline_t pipeline,
                                  size_t n_rows, uint32_t *out);

// Extract leaf hashes (Poseidon2 digests) from GPU to host.
//   out: flat [scw × 8] uint32.
kb_error_t kb_vortex_extract_leaves(kb_vortex_pipeline_t pipeline,
                                     uint32_t *out);

// Return degree (SIS polynomial degree) for the pipeline.
int        kb_vortex_degree(kb_vortex_pipeline_t pipeline);

// Get raw device pointer to pipeline's column-major encoded matrix.
// Layout: d_encoded[col * n_rows + row], col ∈ [0, scw), row ∈ [0, n_rows).
uint32_t  *kb_vortex_encoded_device_ptr(kb_vortex_pipeline_t pipeline);

// Lincomb from a standalone column-major device buffer (not pipeline-bound).
//   result[j] = Σᵢ αⁱ · d_encoded[j * n_rows + i] ∈ E4,  j ∈ [0, scw)
kb_error_t kb_lincomb_e4_colmajor(gnark_gpu_context_t ctx,
                                   const uint32_t *d_encoded_col,
                                   size_t n_rows, size_t scw,
                                   const uint32_t alpha[4],
                                   uint32_t *result);

// Linear combination: result[j] = Σᵢ αⁱ · rows[i][j]
kb_error_t kb_lincomb_e4(gnark_gpu_context_t ctx,
                         kb_vec_t *rows, size_t n_rows, size_t n_cols,
                         const uint32_t alpha[4], uint32_t *result);

// ═══════════════════════════════════════════════════════════════════════════
// Symbolic expression evaluator (GPU bytecode VM)
// ═══════════════════════════════════════════════════════════════════════════
//
// Evaluates a compiled arithmetic DAG over n E4 elements in parallel.
// One GPU thread per element, zero warp divergence.
//
//  Go compiler                        CUDA kernel
//  ───────────                        ───────────
//  ExpressionBoard.Nodes[]            kern_symbolic_eval
//       │  liveness + regalloc             │
//       ▼                                  │
//  GPUProgram {bytecode, consts}    ──H2D──▶ thread i:
//                                          │   E4 slots[S]
//  SymInput[] (device ptrs)         ──H2D──▶   for pc in pgm: exec(i)
//                                          │   out[i] = slots[R]
//
// Opcodes (same layout as CPU compiler):
//   0 OP_CONST:    [0, dst, const_idx]              →  slots[dst] = consts[ci]
//   1 OP_INPUT:    [1, dst, input_id]               →  slots[dst] = read(inputs[id], i)
//   2 OP_MUL:      [2, dst, n, s₀,e₀, ..., sₙ,eₙ] →  slots[dst] = Π slots[sₖ]^eₖ
//   3 OP_LINCOMB:  [3, dst, n, s₀,c₀, ..., sₙ,cₙ] →  slots[dst] = Σ cₖ·slots[sₖ]
//   4 OP_POLYEVAL: [4, dst, n, x, c₀, ..., cₘ]     →  Horner(x, c₀..cₘ)

// Self-recursion boards can require up to ~5000 slots.
// Each slot is 16 bytes (E4) in per-thread local memory.
// 8192 × 16 = 128 KB/thread, well within CUDA's 512 KB limit.
#define SYM_MAX_SLOTS 8192

// Input descriptor — tells the kernel how to read element [i] for one variable.
//   tag=0 (KB):        d_ptr[i]                          → embed as (val,0,0,0)
//   tag=1 (CONST_E4):  broadcast val[4]                  → same E4 for all threads
//   tag=2 (ROT_KB):    d_ptr[(i+offset)%n]               → rotated base column
//   tag=3 (E4_VEC_AOS): d_ptr[i*4..i*4+3]                → E4 AoS vector
//   tag=4 (E4_VEC_SOA): d_ptr[c*n + i], c∈{0,1,2,3}      → E4 SoA vector
//   tag=5 (ROT_E4_SOA): d_ptr[c*n + ((i+offset)%n)]      → rotated E4 SoA vector
//   tag=6 (ROT_E4_AOS): d_ptr[((i+offset)%n)*4..+3]       → rotated E4 AoS vector
typedef struct {
    uint32_t *d_ptr;     // device pointer to KB elements (NULL for CONST)
    uint32_t  val[4];    // E4 constant value (tag=CONST only)
    uint32_t  tag;       // 0=KB, 1=CONST_E4, 2=ROTATED_KB, 3=E4_VEC
    uint32_t  offset;    // rotation offset (tag=ROTATED only)
} SymInputDesc;

// Compiled GPU program handle (device-resident bytecode + constants).
typedef struct KBSymProgram *kb_sym_program_t;

kb_error_t kb_sym_compile(gnark_gpu_context_t ctx,
                           const uint32_t *bytecode,  uint32_t pgm_len,
                           const uint32_t *constants, uint32_t num_consts,
                           uint32_t num_slots,
                           uint32_t result_slot,
                           kb_sym_program_t *out);
void       kb_sym_free(kb_sym_program_t p);

// Evaluate: n elements, result written to h_out (host buffer, n × 4 uint32).
kb_error_t kb_sym_eval(gnark_gpu_context_t ctx,
                        kb_sym_program_t program,
                        const SymInputDesc *h_inputs, uint32_t num_inputs,
                        uint32_t n,
                        uint32_t *h_out);

// Get raw device pointer from a KBVector (for constructing SymInputDesc).
uint32_t *kb_vec_device_ptr(kb_vec_t v);

#ifdef __cplusplus
}
#endif

#endif // GNARK_GPU_KB_H
