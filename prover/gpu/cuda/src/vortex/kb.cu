// KoalaBear GPU kernels — vector ops, NTT, Poseidon2, SIS hash, Merkle, linear combination
//
// All operations on KoalaBear field P = 0x7f000001 (31-bit, 1-limb Montgomery).
// Vectors are flat uint32_t arrays; no SoA decomposition needed.
//
// ┌────────────────── Vortex commit pipeline (GPU) ──────────────────┐
// │                                                                    │
// │  stream_xfer           stream_compute                              │
// │  ───────────           ──────────────                              │
// │  For each chunk (~32 rows):                                        │
// │    H2D chunk → d_work ──event──▶ scatter even cols → d_encoded_col │
// │                                  iFFT(DIF) + scale + FFT(DIT)      │
// │                                  scatter odd cols → d_encoded_col  │
// │                                                                    │
// │  After all chunks:                                                 │
// │    kern_sis_hash (1 block/column, shared-mem NTT-512)              │
// │    kern_p2_sponge → d_leaves                                       │
// │    kern_merkle_level (bottom-up) → d_tree                          │
// │    D2H tree → h_tree (host)                                        │
// └────────────────────────────────────────────────────────────────────┘

#include "kb_field.cuh"
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <cuda_runtime.h>

// ═════════════════════════════════════════════════════════════════════════════
// Internal types
// ═════════════════════════════════════════════════════════════════════════════

struct KBVec {
    uint32_t *d_data;
    size_t    n;
};

struct KBNtt {
    uint32_t *d_fwd_tw;   // n/2 forward twiddles
    uint32_t *d_inv_tw;   // n/2 inverse twiddles
    size_t    n;
    int       log_n;
};

struct KBPoseidon2 {
    uint32_t *d_round_keys;  // flat [nb_rounds × key_width]
    uint32_t *d_diag;        // internal MDS diagonal (width elements, Montgomery form)
    int       width;
    int       nb_full_rounds;
    int       nb_partial_rounds;
};

// Ring-SIS hash context: pre-NTT'd keys + NTT domain for degree-d cyclotomic ring
struct KBSis {
    uint32_t *d_ag;            // [n_polys × degree] pre-NTT'd keys (bit-reversed, coset domain)
    uint32_t *d_fwd_tw;       // [degree/2] forward twiddles (ωⁱ)
    uint32_t *d_inv_tw;       // [degree/2] inverse twiddles (ω⁻ⁱ)
    uint32_t *d_coset_table;  // [degree] shift^j (natural order)
    uint32_t *d_coset_inv;    // [degree] shift^{-j} · (1/degree) (natural order)
    int       degree;
    int       log_degree;
    int       n_polys;
    int       log_two_bound;  // limb bit-width (e.g. 16)
};

// Pre-allocated device buffers for the full Vortex commit pipeline.
// Created once in pipeline_init, reused across Commit() calls.
//
// GPU RS encode (rate=2) eliminates CPU RS + halves H2D data:
//
//  h_input (pinned)    d_work           d_encoded_col (col-major)
//  [nR × nC]      ─H2D─▶ [nR × nC] ──transpose──▶ even cols [0,2,4,..]
//  host, row-major       │ (original)
//                        ▼ iFFT(DIF, inv_tw)
//                        │ scale(cosetBR × cardInv)
//                        ▼ FFT(DIT, fwd_tw)
//                        └──transpose──▶ odd cols [1,3,5,..]
//
//  d_encoded_col ──SIS──▶ d_sis ──sponge──▶ d_leaves ──merkle──▶ d_tree
//
//  ┌───────────────────── memory layout ──────────────────────────┐
//  │ d_work          [max_n_rows × n_cols]   H2D + NTT workspace  │
//  │ d_encoded_col   [scw × max_n_rows]      column-major matrix  │
//  │ d_rs_fwd_tw     [n_cols / 2]            RS forward twiddles  │
//  │ d_rs_inv_tw     [n_cols / 2]            RS inverse twiddles  │
//  │ d_scaled_coset  [n_cols]                cosetBR × cardInv    │
//  │ d_sis           [scw × degree]          SIS hash output      │
//  │ d_leaves        [scw × 8]               Poseidon2 digests    │
//  │ d_tree          [2·np × 8]              Merkle heap          │
//  └──────────────────────────────────────────────────────────────┘
struct KBVortexPipeline {
    KBSis       *sis;           // not owned
    KBPoseidon2 *p2_sponge;    // not owned
    KBPoseidon2 *p2_compress;  // not owned
    // RS encode buffers
    uint32_t *d_work;           // [max_n_rows × n_cols], H2D staging + NTT workspace
    uint32_t *d_rs_fwd_tw;     // [n_cols / 2] forward twiddles for RS domain
    uint32_t *d_rs_inv_tw;     // [n_cols / 2] inverse twiddles for RS domain
    uint32_t *d_scaled_coset;  // [n_cols] = CosetTableBitReverse × cardinalityInv (rate=2)
    // Multi-coset RS encode (rate > 2)
    uint32_t *d_coeffs;         // [max_n_rows × n_cols], IFFT partial-state backup
    uint32_t *d_coset_tables;   // [(rate-1) × n_cols], coset scaling tables (bit-reversed)
    // Encoded matrix + downstream
    uint32_t *d_encoded_col;    // [scw × max_n_rows], column-major
    uint32_t *d_sis;            // [scw × degree]
    uint32_t *d_leaves;         // [scw × 8]
    uint32_t *d_tree;           // [2 × tree_np × 8], heap layout
    // Async extraction buffers (overlap D2H with SIS+P2+Merkle compute)
    uint32_t *d_enc_rowmajor;   // [max_n_rows × scw], GPU transpose buffer
    uint32_t *h_enc_pinned;     // [max_n_rows × scw], pinned host for encoded matrix
    uint32_t *h_sis_pinned;     // [scw × degree], pinned host for SIS hashes
    uint32_t *h_leaves_pinned;  // [scw × 8], pinned host for leaf hashes
    cudaEvent_t ev_rs_done;     // RS encoding complete → start D2H encoded
    cudaEvent_t ev_sis_done;    // SIS hashing complete → start D2H SIS
    cudaEvent_t ev_p2_done;     // P2 hashing complete → start D2H leaves
    // Pinned host buffers
    uint32_t *h_input;          // [max_n_rows × n_cols], cudaMallocHost (pinned)
    uint32_t *h_tree;           // [(2·tree_np − 1) × 8], cudaMallocHost (pinned)
    // Streams for pipelined H2D + compute overlap
    cudaStream_t stream_xfer;   // H2D transfers + async D2H extraction
    cudaStream_t stream_compute; // RS encode + transpose
    cudaEvent_t  h2d_event;     // signals H2D chunk completion
    // Dimensions
    size_t max_n_rows;
    size_t n_cols;
    size_t size_codeword;       // n_cols × rate
    size_t tree_np;
    int    rate;
    int    log_n_cols;
    int    degree;              // cached from sis
};

// ═════════════════════════════════════════════════════════════════════════════
// Helpers
// ═════════════════════════════════════════════════════════════════════════════

static constexpr int KB_BLOCK = 256;

static inline int kb_grid(size_t n) {
    return (int)((n + KB_BLOCK - 1) / KB_BLOCK);
}

static inline int ilog2(size_t n) {
    int r = 0;
    while ((1ULL << r) < n) r++;
    return r;
}

static inline size_t next_pow2(size_t n) {
    size_t v = 1;
    while (v < n) v <<= 1;
    return v;
}

#define CUDA_CHECK(call) do {                        \
    cudaError_t err = (call);                        \
    if (err != cudaSuccess) return KB_ERROR_CUDA;    \
} while (0)

// ═════════════════════════════════════════════════════════════════════════════
// Vector element-wise kernels
// ═════════════════════════════════════════════════════════════════════════════

__global__ void kern_kb_add(uint32_t *c, const uint32_t *a, const uint32_t *b, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i < n) c[i] = kb_add(a[i], b[i]);
}

__global__ void kern_kb_sub(uint32_t *c, const uint32_t *a, const uint32_t *b, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i < n) c[i] = kb_sub(a[i], b[i]);
}

__global__ void kern_kb_mul(uint32_t *c, const uint32_t *a, const uint32_t *b, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i < n) c[i] = kb_mul(a[i], b[i]);
}

__global__ void kern_kb_scale(uint32_t *v, uint32_t s, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i < n) v[i] = kb_mul(v[i], s);
}

// v[i] *= gⁱ — each thread computes gⁱ by repeated squaring
__global__ void kern_kb_scale_by_powers(uint32_t *v, uint32_t g, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;
    uint32_t pow = KB_ONE;
    uint32_t base = g;
    size_t exp = i;
    while (exp > 0) {
        if (exp & 1) pow = kb_mul(pow, base);
        base = kb_sqr(base);
        exp >>= 1;
    }
    v[i] = kb_mul(v[i], pow);
}

// Batch version of scale-by-powers:
// for each row r and index j, data[r][j] *= g^j
__global__ void kern_batch_scale_by_powers(uint32_t *data, uint32_t g,
                                            size_t n_rows, size_t row_stride,
                                            size_t n) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t total = n_rows * n;
    if (idx >= total) return;

    size_t row = idx / n;
    size_t j   = idx % n;
    uint32_t *rd = data + row * row_stride;

    uint32_t pow = KB_ONE;
    uint32_t base = g;
    size_t exp = j;
    while (exp > 0) {
        if (exp & 1) pow = kb_mul(pow, base);
        base = kb_sqr(base);
        exp >>= 1;
    }
    rd[j] = kb_mul(rd[j], pow);
}

// ═════════════════════════════════════════════════════════════════════════════
// NTT kernels — DIF (forward) and DIT (inverse)
// ═════════════════════════════════════════════════════════════════════════════
//
// Butterfly:
//   DIF: a' = a + b,  b' = (a − b) · ω
//   DIT: a' = a + ω·b,  b' = a − ω·b
//
// Twiddles: flat array tw[0..n/2), where tw[k] = ωᵏ in Montgomery form.
// Stage s: distance half = n >> (s+1), twiddle index = j << s.

__global__ void kern_ntt_dif(uint32_t *data, const uint32_t *tw,
                              size_t n, int stage) {
    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    size_t half = n >> (stage + 1);
    size_t pairs = n >> 1;
    if (tid >= pairs) return;

    size_t group = tid / half;
    size_t j     = tid % half;
    size_t ia    = group * (2 * half) + j;
    size_t ib    = ia + half;
    size_t tw_idx = j * (1u << stage);

    uint32_t a = data[ia];
    uint32_t b = data[ib];
    uint32_t w = tw[tw_idx];

    data[ia] = kb_add(a, b);
    data[ib] = kb_mul(kb_sub(a, b), w);
}

__global__ void kern_ntt_dit(uint32_t *data, const uint32_t *tw,
                              size_t n, int stage) {
    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    size_t half = n >> (stage + 1);
    size_t pairs = n >> 1;
    if (tid >= pairs) return;

    size_t group = tid / half;
    size_t j     = tid % half;
    size_t ia    = group * (2 * half) + j;
    size_t ib    = ia + half;
    size_t tw_idx = j * (1u << stage);

    uint32_t a  = data[ia];
    uint32_t wb = kb_mul(data[ib], tw[tw_idx]);

    data[ia] = kb_add(a, wb);
    data[ib] = kb_sub(a, wb);
}

// ═════════════════════════════════════════════════════════════════════════════
// Batch NTT — process n_rows independent NTTs of size n in parallel
// ═════════════════════════════════════════════════════════════════════════════
//
// One thread per (row, butterfly-pair). Row data at data[row * row_stride].
// Launch grid: total = n_rows × (n/2), one stage per kernel launch.

__global__ void kern_batch_ntt_dif(uint32_t *data, const uint32_t *tw,
                                    size_t n, int stage,
                                    size_t n_rows, size_t row_stride) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t pairs_per_row = n >> 1;
    size_t total = n_rows * pairs_per_row;
    if (idx >= total) return;

    size_t row  = idx / pairs_per_row;
    size_t pair = idx % pairs_per_row;
    uint32_t *rd = data + row * row_stride;

    size_t half   = n >> (stage + 1);
    size_t group  = pair / half;
    size_t j      = pair % half;
    size_t ia     = group * 2 * half + j;
    size_t ib     = ia + half;
    size_t tw_idx = j * (1u << stage);

    uint32_t a = rd[ia], b = rd[ib], w = tw[tw_idx];
    rd[ia] = kb_add(a, b);
    rd[ib] = kb_mul(kb_sub(a, b), w);
}

__global__ void kern_batch_ntt_dit(uint32_t *data, const uint32_t *tw,
                                    size_t n, int stage,
                                    size_t n_rows, size_t row_stride) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t pairs_per_row = n >> 1;
    size_t total = n_rows * pairs_per_row;
    if (idx >= total) return;

    size_t row  = idx / pairs_per_row;
    size_t pair = idx % pairs_per_row;
    uint32_t *rd = data + row * row_stride;

    size_t half   = n >> (stage + 1);
    size_t group  = pair / half;
    size_t j      = pair % half;
    size_t ia     = group * 2 * half + j;
    size_t ib     = ia + half;
    size_t tw_idx = j * (1u << stage);

    uint32_t a  = rd[ia];
    uint32_t wb = kb_mul(rd[ib], tw[tw_idx]);
    rd[ia] = kb_add(a, wb);
    rd[ib] = kb_sub(a, wb);
}

// Batch element-wise multiply each row by a shared vector: data[row][j] *= vec[j]
__global__ void kern_batch_mul_vec(uint32_t *data, const uint32_t *vec,
                                    size_t n, size_t n_rows, size_t row_stride) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t total = n_rows * n;
    if (idx >= total) return;
    size_t row = idx / n;
    size_t j   = idx % n;
    data[row * row_stride + j] = kb_mul(data[row * row_stride + j], vec[j]);
}

// ═════════════════════════════════════════════════════════════════════════════
// Fused batch NTT: DIF tail + scale + DIT head in shared memory
// ═════════════════════════════════════════════════════════════════════════════
//
// After s_cut global DIF stages, data partitions into independent tiles of
// size T = 2^(log_n − s_cut).  This kernel loads one tile, performs:
//
//   ┌─── shared memory (T × 4 bytes) ───────────────────────────┐
//   │  Load from global                                          │
//   │  DIF stages s_cut .. log_n−1          (log_n−s_cut stages) │
//   │  Scale × cosetBR·cardInv              (element-wise)       │
//   │  DIT stages log_n−1 .. s_cut          (log_n−s_cut stages) │
//   │  Store to global                                           │
//   └────────────────────────────────────────────────────────────┘
//
// Replaces 2·(log_n−s_cut)+1 separate kernel launches with one load/store.
// Grid: (tiles_per_row × n_rows) blocks.  Threads: min(T/2, 1024).

__global__ void kern_batch_ntt_fused(
    uint32_t       *data,
    const uint32_t *inv_tw,      // DIF (inverse) twiddles [n/2]
    const uint32_t *fwd_tw,      // DIT (forward) twiddles [n/2]
    const uint32_t *scale_vec,   // cosetBR × cardInv [n]
    int             log_n,
    int             s_cut,       // first local DIF stage
    size_t          n_rows,
    size_t          row_stride)
{
    extern __shared__ uint32_t tile[];

    const int k         = log_n - s_cut;
    const int tile_size = 1 << k;
    const int half_tile = tile_size >> 1;

    int tiles_per_row = 1 << s_cut;
    size_t row  = blockIdx.x / tiles_per_row;
    int    tidx = blockIdx.x % tiles_per_row;
    if (row >= n_rows) return;

    uint32_t *rd = data + row * row_stride + tidx * tile_size;
    const uint32_t *sv = scale_vec + tidx * tile_size;
    int tid = threadIdx.x;

    // Load tile
    for (int i = tid; i < tile_size; i += blockDim.x)
        tile[i] = rd[i];
    __syncthreads();

    // DIF local stages: s = s_cut, s_cut+1, ..., log_n−1
    for (int s = s_cut; s < log_n; s++) {
        int half = tile_size >> (s - s_cut + 1);
        for (int pid = tid; pid < half_tile; pid += blockDim.x) {
            int g  = pid / half;
            int j  = pid % half;
            int ia = g * 2 * half + j;
            int ib = ia + half;
            uint32_t a = tile[ia], b = tile[ib], w = inv_tw[j << s];
            tile[ia] = kb_add(a, b);
            tile[ib] = kb_mul(kb_sub(a, b), w);
        }
        __syncthreads();
    }

    // Scale
    for (int i = tid; i < tile_size; i += blockDim.x)
        tile[i] = kb_mul(tile[i], sv[i]);
    __syncthreads();

    // DIT local stages: s = log_n−1, log_n−2, ..., s_cut
    for (int s = log_n - 1; s >= s_cut; s--) {
        int half = tile_size >> (s - s_cut + 1);
        for (int pid = tid; pid < half_tile; pid += blockDim.x) {
            int g  = pid / half;
            int j  = pid % half;
            int ia = g * 2 * half + j;
            int ib = ia + half;
            uint32_t a  = tile[ia];
            uint32_t wb = kb_mul(tile[ib], fwd_tw[j << s]);
            tile[ia] = kb_add(a, wb);
            tile[ib] = kb_sub(a, wb);
        }
        __syncthreads();
    }

    // Store tile
    for (int i = tid; i < tile_size; i += blockDim.x)
        rd[i] = tile[i];
}

// ═════════════════════════════════════════════════════════════════════════════
// Poseidon2 — device functions
// ═════════════════════════════════════════════════════════════════════════════
//
// Width-16 (Merkle compression) and Width-24 (sponge hash).
//
// Round structure:
//   matMulExternal(state)                    // initial
//   for i in 0..rF/2:   addRC → sBox_full → matMulExternal
//   for i in 0..rP:     addRC[0] → sBox(0) → matMulInternal
//   for i in 0..rF/2:   addRC → sBox_full → matMulExternal
//
// S-box: x³

// M4 = circ(2,3,1,1) via addition chain
__device__ __forceinline__ void p2_matmul_m4(uint32_t &s0, uint32_t &s1,
                                              uint32_t &s2, uint32_t &s3) {
    uint32_t t01 = kb_add(s0, s1);
    uint32_t t23 = kb_add(s2, s3);
    uint32_t t0123 = kb_add(t01, t23);
    uint32_t t01123 = kb_add(t0123, s1);
    uint32_t t01233 = kb_add(t0123, s3);
    s3 = kb_add(kb_dbl(s0), t01233);
    s1 = kb_add(kb_dbl(s2), t01123);
    s0 = kb_add(t01, t01123);
    s2 = kb_add(t23, t01233);
}

// External MDS: circ(2M4, M4, .., M4)
__device__ __forceinline__ void p2_matmul_external(uint32_t *s, int width) {
    for (int i = 0; i < width; i += 4)
        p2_matmul_m4(s[i], s[i+1], s[i+2], s[i+3]);
    uint32_t sum[4] = {0, 0, 0, 0};
    for (int i = 0; i < width; i += 4) {
        sum[0] = kb_add(sum[0], s[i]);
        sum[1] = kb_add(sum[1], s[i+1]);
        sum[2] = kb_add(sum[2], s[i+2]);
        sum[3] = kb_add(sum[3], s[i+3]);
    }
    for (int i = 0; i < width; i += 4) {
        s[i]   = kb_add(s[i],   sum[0]);
        s[i+1] = kb_add(s[i+1], sum[1]);
        s[i+2] = kb_add(s[i+2], sum[2]);
        s[i+3] = kb_add(s[i+3], sum[3]);
    }
}

// Internal MDS: state[i] = sum + dᵢ · state[i]
__device__ __forceinline__ void p2_matmul_internal(uint32_t *s, int width,
                                                    const uint32_t *diag) {
    uint32_t sum = s[0];
    for (int i = 1; i < width; i++) sum = kb_add(sum, s[i]);
    for (int i = 0; i < width; i++) {
        s[i] = kb_add(sum, kb_mul(s[i], diag[i]));
    }
}

// S-box: x ↦ x³
__device__ __forceinline__ uint32_t p2_sbox(uint32_t x) {
    return kb_mul(kb_sqr(x), x);
}

// Full Poseidon2 permutation
__device__ void p2_permutation(uint32_t *state, int width,
                                int rf, int rp,
                                const uint32_t *round_keys,
                                const uint32_t *diag) {
    int half_rf = rf / 2;
    int rk_off = 0;

    p2_matmul_external(state, width);

    for (int r = 0; r < half_rf; r++) {
        for (int j = 0; j < width; j++)
            state[j] = kb_add(state[j], round_keys[rk_off + j]);
        rk_off += width;
        for (int j = 0; j < width; j++)
            state[j] = p2_sbox(state[j]);
        p2_matmul_external(state, width);
    }

    for (int r = 0; r < rp; r++) {
        state[0] = kb_add(state[0], round_keys[rk_off]);
        rk_off += 1;
        state[0] = p2_sbox(state[0]);
        p2_matmul_internal(state, width, diag);
    }

    for (int r = 0; r < half_rf; r++) {
        for (int j = 0; j < width; j++)
            state[j] = kb_add(state[j], round_keys[rk_off + j]);
        rk_off += width;
        for (int j = 0; j < width; j++)
            state[j] = p2_sbox(state[j]);
        p2_matmul_external(state, width);
    }
}

// ═════════════════════════════════════════════════════════════════════════════
// Poseidon2 batch kernels
// ═════════════════════════════════════════════════════════════════════════════

// Batch compress (width=16): one thread per pair → hash
// Feed-forward: hash[j] = permuted_state[8+j] + right[j]
//
// Shared-mem diag: each block loads the 16-element diag vector into
// shared memory once, so the 21 × 2 = 42 partial-round reads per
// permutation hit shared memory (~zero latency) instead of L2/global.
__global__ void __launch_bounds__(KB_BLOCK, 4)
kern_p2_compress(const uint32_t *input, uint32_t *output,
                  const uint32_t *round_keys,
                  const uint32_t *diag,
                  size_t count) {
    __shared__ uint32_t s_diag[16];
    if (threadIdx.x < 16) s_diag[threadIdx.x] = diag[threadIdx.x];
    __syncthreads();

    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= count) return;

    uint32_t state[16];
    const uint32_t *pair = input + tid * 16;
    for (int j = 0; j < 16; j++) state[j] = pair[j];

    uint32_t ff[8];
    for (int j = 0; j < 8; j++) ff[j] = state[8 + j];

    p2_permutation(state, 16, 6, 21, round_keys, s_diag);

    uint32_t *out = output + tid * 8;
    for (int j = 0; j < 8; j++)
        out[j] = kb_add(state[8 + j], ff[j]);
}

// Batch sponge (width=24): one thread per input → 8-element digest
// Absorb: overwrite state[8..23] with input block, permute. Squeeze: state[0..7].
__global__ void __launch_bounds__(KB_BLOCK, 4)
kern_p2_sponge(const uint32_t *input, size_t input_len,
                uint32_t *output,
                const uint32_t *round_keys,
                const uint32_t *diag,
                size_t count) {
    // width=24 sponge — diag has 24 elements; cache in shared mem.
    __shared__ uint32_t s_diag[24];
    if (threadIdx.x < 24) s_diag[threadIdx.x] = diag[threadIdx.x];
    __syncthreads();

    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= count) return;

    const int rate = 16;
    uint32_t state[24];
    for (int j = 0; j < 24; j++) state[j] = 0;

    const uint32_t *inp = input + tid * input_len;

    for (size_t off = 0; off < input_len; off += rate) {
        size_t chunk = (off + rate <= input_len) ? rate : input_len - off;
        for (size_t j = 0; j < chunk; j++)
            state[8 + j] = inp[off + j];
        p2_permutation(state, 24, 6, 21, round_keys, s_diag);
    }

    uint32_t *out = output + tid * 8;
    for (int j = 0; j < 8; j++) out[j] = state[j];
}

// Batch Merkle-Damgard hash (width=16, rate=8): one thread per column → 8-element digest.
// Matches CPU CompressPoseidon2x16: iterative Davies-Meyer with width-16 Poseidon2.
//   state[0..7]  = running hash (capacity, initially zero)
//   state[8..15] = message block (8 input elements per step)
//   After permutation: state[j] = P(state)[8+j] + input[j] for j=0..7
// Davies-Meyer SIS leaf-hash. Each thread processes one column's full
// SIS digest of length input_len = degree (typically 512). With 21
// partial rounds × ⌈input_len/8⌉ permutations per thread × 16 diag
// reads per partial round, caching diag in shared mem saves a lot of
// L2 traffic.
__global__ void __launch_bounds__(KB_BLOCK, 4)
kern_p2_md_hash(const uint32_t *input, size_t input_len,
                 uint32_t *output,
                 const uint32_t *round_keys,
                 const uint32_t *diag,
                 size_t count) {
    __shared__ uint32_t s_diag[16];
    if (threadIdx.x < 16) s_diag[threadIdx.x] = diag[threadIdx.x];
    __syncthreads();

    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= count) return;

    const int rate = 8;
    uint32_t state[16];
    for (int j = 0; j < 16; j++) state[j] = 0;

    const uint32_t *inp = input + tid * input_len;

    for (size_t off = 0; off < input_len; off += rate) {
        // Load message block into state[8..15]
        for (int j = 0; j < rate; j++)
            state[8 + j] = (off + j < input_len) ? inp[off + j] : 0;

        p2_permutation(state, 16, 6, 21, round_keys, s_diag);

        // Davies-Meyer feed-forward: state[j] = P(state)[8+j] + input[j]
        for (int j = 0; j < 8; j++) {
            uint32_t m = (off + j < input_len) ? inp[off + j] : 0;
            state[j] = kb_add(state[8 + j], m);
        }
    }

    uint32_t *out = output + tid * 8;
    for (int j = 0; j < 8; j++) out[j] = state[j];
}

// ═════════════════════════════════════════════════════════════════════════════
// SIS hash kernel
// ═════════════════════════════════════════════════════════════════════════════
//
// Ring-SIS: H(column) = IFFT_coset( Σᵢ FFT_coset(limbs_i) ⊙ Ag[i] )
//
// One thread-block per column. 256 threads cooperate on NTT-512 in shared memory.
// Shared memory: 2 × degree uint32 (work buffer + accumulator).
//
// Limb decomposition (logTwoBound=16):
//   element (Montgomery) → canonical (kb_from_mont) → [lo16, hi16] (raw, not Montgomery)

__global__ void kern_sis_hash(
    const uint32_t *d_encoded,    // [sizeCodeWord × nRows], column-major, Montgomery
    int n_rows,
    int size_codeword,            // unused (kept for ABI compat)
    const uint32_t *d_ag,         // [nPolys × degree], pre-NTT'd keys (bit-reversed)
    int n_polys,
    int degree,
    int log_degree,
    const uint32_t *d_fwd_tw,    // [degree/2] forward twiddles
    const uint32_t *d_inv_tw,    // [degree/2] inverse twiddles
    const uint32_t *d_coset_table, // [degree] shift^j, natural order
    const uint32_t *d_coset_inv,   // [degree] shift^{-j} · (1/degree), natural order
    uint32_t *d_sis_out)           // [sizeCodeWord × degree]
{
    int col = blockIdx.x;
    int tid = threadIdx.x;
    int bdim = blockDim.x;

    extern __shared__ uint32_t shared[];
    uint32_t *work = shared;               // [degree]
    uint32_t *acc  = shared + degree;      // [degree]

    // Zero accumulator
    for (int j = tid; j < degree; j += bdim) acc[j] = 0;
    __syncthreads();

    for (int poly = 0; poly < n_polys; poly++) {
        // ── 1. Extract limbs → work[0..degree-1] ────────────────────────
        // Each element gives 2 limbs (16-bit). limb_idx = poly*degree + j.
        int limb_base = poly * degree;
        for (int j = tid; j < degree; j += bdim) {
            int limb_idx = limb_base + j;
            int elem_idx = limb_idx >> 1;
            int limb_half = limb_idx & 1;

            uint32_t val = 0;
            if (elem_idx < n_rows) {
                // Column-major: d_encoded[col * n_rows + row]. Coalesced when
                // consecutive threads read consecutive rows (elem_idx).
                // gnark-crypto extracts LE uint16 limbs from canonical form.
                uint32_t canonical = kb_from_mont(d_encoded[(size_t)col * n_rows + elem_idx]);
                val = limb_half == 0 ? (canonical & 0xFFFFu) : (canonical >> 16);
            }
            work[j] = val;
        }
        __syncthreads();

        // ── 2. Coset shift: work[j] *= shift^j ──────────────────────────
        for (int j = tid; j < degree; j += bdim)
            work[j] = kb_mul(work[j], d_coset_table[j]);
        __syncthreads();

        // ── 3. Forward DIF NTT in shared memory ─────────────────────────
        for (int s = 0; s < log_degree; s++) {
            int half = degree >> (s + 1);
            int pairs = degree >> 1;
            for (int t = tid; t < pairs; t += bdim) {
                int group = t / half;
                int jj    = t % half;
                int ia    = group * 2 * half + jj;
                int ib    = ia + half;
                int tw_idx = jj * (1 << s);

                uint32_t a = work[ia], b = work[ib];
                uint32_t w = d_fwd_tw[tw_idx];
                work[ia] = kb_add(a, b);
                work[ib] = kb_mul(kb_sub(a, b), w);
            }
            __syncthreads();
        }

        // ── 4. Pointwise mul by Ag[poly] + accumulate ───────────────────
        const uint32_t *ag_poly = d_ag + poly * degree;
        for (int j = tid; j < degree; j += bdim)
            acc[j] = kb_add(acc[j], kb_mul(work[j], ag_poly[j]));
        __syncthreads();
    }

    // ── 5. Inverse DIT NTT on accumulator ────────────────────────────────
    for (int j = tid; j < degree; j += bdim) work[j] = acc[j];
    __syncthreads();

    for (int s = log_degree - 1; s >= 0; s--) {
        int half = degree >> (s + 1);
        int pairs = degree >> 1;
        for (int t = tid; t < pairs; t += bdim) {
            int group = t / half;
            int jj    = t % half;
            int ia    = group * 2 * half + jj;
            int ib    = ia + half;
            int tw_idx = jj * (1 << s);

            uint32_t a  = work[ia];
            uint32_t wb = kb_mul(work[ib], d_inv_tw[tw_idx]);
            work[ia] = kb_add(a, wb);
            work[ib] = kb_sub(a, wb);
        }
        __syncthreads();
    }

    // ── 6. Inverse coset shift + scale by 1/n ───────────────────────────
    for (int j = tid; j < degree; j += bdim)
        d_sis_out[col * degree + j] = kb_mul(work[j], d_coset_inv[j]);
}

// ═════════════════════════════════════════════════════════════════════════════
// Scatter-transpose kernel — row-major → column-major with optional stride
// ═════════════════════════════════════════════════════════════════════════════
//
// Transposes rows into column-major output.  Supports column stride/offset
// for rate-2 interleaving: col_stride=2, col_offset=0 writes even columns,
// col_stride=2, col_offset=1 writes odd columns.
//
// Uses shared-memory tiled transpose (32×32, +1 pad) for coalesced R+W.
//
//   dst[(src_col * col_stride + col_offset) * total_rows + row] = src[row * n_src_cols + src_col]

__global__ void kern_scatter_transpose(const uint32_t * __restrict__ src,
                                        uint32_t * __restrict__ dst,
                                        size_t n_rows, size_t n_src_cols,
                                        size_t total_rows,
                                        int col_stride, int col_offset) {
    __shared__ uint32_t tile[32][33];
    unsigned bx = blockIdx.x * 32;   // source-column tile
    unsigned by = blockIdx.y * 32;   // row tile

    // Load tile: coalesced read from row-major src
    for (int i = 0; i < 32; i += 8) {
        unsigned r = by + threadIdx.y + i;
        unsigned c = bx + threadIdx.x;
        if (r < n_rows && c < n_src_cols)
            tile[threadIdx.y + i][threadIdx.x] = src[(size_t)r * n_src_cols + c];
    }
    __syncthreads();

    // Store tile: coalesced write to column-major dst
    for (int i = 0; i < 32; i += 8) {
        unsigned local_row = by + threadIdx.x;
        unsigned src_col   = bx + threadIdx.y + i;
        if (local_row < n_rows && src_col < n_src_cols) {
            size_t dst_col = (size_t)src_col * col_stride + col_offset;
            dst[dst_col * total_rows + local_row] =
                tile[threadIdx.x][threadIdx.y + i];
        }
    }
}

// ═════════════════════════════════════════════════════════════════════════════
// Merkle tree kernel
// ═════════════════════════════════════════════════════════════════════════════
//
// Bottom-up: hash pairs of sibling hashes using Poseidon2 compression (width=16).
// Heap layout: tree[1]=root, tree[2i]=left child, tree[2i+1]=right child.

// Merkle tree node hash: hashLR(left, right) using Merkle-Damgard Poseidon2.
//
// Matches the CPU poseidon2_koalabear.MDHasher which processes 16 input
// elements in TWO 8-element blocks with zero initial state:
//   Block 1 (left):  state = CompressPoseidon2([0,...,0], left)
//   Block 2 (right): state = CompressPoseidon2(state,    right)
// Merkle tree compression: hash(left, right) using Poseidon2 MD hash (width=16).
// Matches smt_koalabear.hashLR which calls MDHasher on left[8]||right[8]:
//   h1 = CompressPoseidon2(zero, left)
//   h2 = CompressPoseidon2(h1, right)
//   output = h2
// Merkle tree level compress: each thread compresses one (left, right)
// pair into one parent hash via two Poseidon2 permutations + Davies-
// Meyer feed-forward.
//
// Shared-mem diag avoids 2 × 21 × 16 = 672 L2 reads per thread
// (every partial round's matmul_internal reads the full diag vector).
__global__ void __launch_bounds__(KB_BLOCK, 4)
kern_merkle_level(const uint32_t *children, uint32_t *parents,
                   const uint32_t *round_keys,
                   const uint32_t *diag,
                   size_t n_pairs) {
    __shared__ uint32_t s_diag[16];
    if (threadIdx.x < 16) s_diag[threadIdx.x] = diag[threadIdx.x];
    __syncthreads();

    size_t tid = blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= n_pairs) return;

    uint32_t state[16];
    // Initialize state to zero (MD IV)
    for (int j = 0; j < 16; j++) state[j] = 0;

    // ── Block 1: left child ──────────────────────────────────────────────
    for (int j = 0; j < 8; j++)
        state[8 + j] = children[(2 * tid) * 8 + j];

    uint32_t ff[8];
    for (int j = 0; j < 8; j++) ff[j] = state[8 + j];

    p2_permutation(state, 16, 6, 21, round_keys, s_diag);

    // Feed-forward → new capacity (state[0:8])
    for (int j = 0; j < 8; j++)
        state[j] = kb_add(state[8 + j], ff[j]);

    // ── Block 2: right child ─────────────────────────────────────────────
    for (int j = 0; j < 8; j++)
        state[8 + j] = children[(2 * tid + 1) * 8 + j];

    for (int j = 0; j < 8; j++) ff[j] = state[8 + j];

    p2_permutation(state, 16, 6, 21, round_keys, s_diag);

    // Feed-forward → output hash
    uint32_t *out = parents + tid * 8;
    for (int j = 0; j < 8; j++)
        out[j] = kb_add(state[8 + j], ff[j]);
}

// ═════════════════════════════════════════════════════════════════════════════
// Linear combination: UAlpha[j] = Σᵢ αⁱ · row[i][j]  (result ∈ E4)
// ═════════════════════════════════════════════════════════════════════════════

__global__ void kern_lincomb_e4(const uint32_t * const *rows,
                                 size_t n_rows, size_t n_cols,
                                 E4 alpha,
                                 uint32_t *result) {
    size_t j = blockIdx.x * blockDim.x + threadIdx.x;
    if (j >= n_cols) return;

    E4 acc = e4_zero();
    E4 alpha_pow = {{KB_ONE, 0}, {0, 0}};

    for (size_t i = 0; i < n_rows; i++) {
        uint32_t val = rows[i][j];
        e4_mulacc(acc, val, alpha_pow);
        alpha_pow = e4_mul(alpha_pow, alpha);
    }

    result[j * 4 + 0] = acc.b0.a0;
    result[j * 4 + 1] = acc.b0.a1;
    result[j * 4 + 2] = acc.b1.a0;
    result[j * 4 + 3] = acc.b1.a1;
}

// Linear combination on column-major encoded matrix (for GPU Prove).
// UAlpha[j] = Σᵢ αⁱ · d_encoded_col[j * n_rows + i], result ∈ E4^scw
__global__ void kern_lincomb_e4_colmajor(const uint32_t *d_encoded_col,
                                          size_t n_rows, size_t scw,
                                          E4 alpha, uint32_t *result) {
    size_t j = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (j >= scw) return;

    const uint32_t *col = d_encoded_col + j * n_rows;
    E4 acc = e4_zero();
    E4 alpha_pow = {{KB_ONE, 0}, {0, 0}};

    for (size_t i = 0; i < n_rows; i++) {
        e4_mulacc(acc, col[i], alpha_pow);
        alpha_pow = e4_mul(alpha_pow, alpha);
    }

    result[j * 4 + 0] = acc.b0.a0;
    result[j * 4 + 1] = acc.b0.a1;
    result[j * 4 + 2] = acc.b1.a0;
    result[j * 4 + 3] = acc.b1.a1;
}

// ═════════════════════════════════════════════════════════════════════════════
// C ABI implementations
// ═════════════════════════════════════════════════════════════════════════════

#include "gnark_gpu_kb.h"

// ── Vector lifecycle ────────────────────────────────────────────────────────

extern "C" kb_error_t kb_vec_alloc(gnark_gpu_context_t, size_t n, kb_vec_t *out) {
    auto *v = new(std::nothrow) KBVec;
    if (!v) return KB_ERROR_OOM;
    cudaError_t err = cudaMalloc(&v->d_data, n * sizeof(uint32_t));
    if (err != cudaSuccess) { delete v; return KB_ERROR_CUDA; }
    v->n = n;
    *out = v;
    return KB_SUCCESS;
}

extern "C" void kb_vec_free(kb_vec_t v) {
    if (v) { cudaFree(v->d_data); delete v; }
}

extern "C" size_t kb_vec_len(kb_vec_t v) {
    return v ? v->n : 0;
}

extern "C" kb_error_t kb_vec_h2d(gnark_gpu_context_t, kb_vec_t dst,
                                  const uint32_t *src, size_t n) {
    if (!dst || n > dst->n) return KB_ERROR_INVALID;
    CUDA_CHECK(cudaMemcpy(dst->d_data, src, n * sizeof(uint32_t), cudaMemcpyHostToDevice));
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_d2h(gnark_gpu_context_t, uint32_t *dst,
                                  kb_vec_t src, size_t n) {
    if (!src || n > src->n) return KB_ERROR_INVALID;
    CUDA_CHECK(cudaMemcpy(dst, src->d_data, n * sizeof(uint32_t), cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_d2d(gnark_gpu_context_t, kb_vec_t dst, kb_vec_t src) {
    if (!dst || !src || dst->n != src->n) return KB_ERROR_INVALID;
    CUDA_CHECK(cudaMemcpyAsync(dst->d_data, src->d_data, src->n * sizeof(uint32_t), cudaMemcpyDeviceToDevice, 0));
    return KB_SUCCESS;
}

// D2D copy with raw pointers (async on default stream).
extern "C" kb_error_t kb_vec_d2d_offset(gnark_gpu_context_t, uint32_t *dst,
                                         const uint32_t *src, size_t n) {
    CUDA_CHECK(cudaMemcpyAsync(dst, src, n * sizeof(uint32_t), cudaMemcpyDeviceToDevice, 0));
    return KB_SUCCESS;
}

// D2H with raw pointers.
extern "C" kb_error_t kb_vec_d2h_raw(gnark_gpu_context_t, uint32_t *dst,
                                      const uint32_t *src, size_t n) {
    CUDA_CHECK(cudaMemcpy(dst, src, n * sizeof(uint32_t), cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

// Synchronize the default CUDA stream (wait for all queued ops to complete).
extern "C" kb_error_t kb_sync(gnark_gpu_context_t) {
    CUDA_CHECK(cudaStreamSynchronize(0));
    return KB_SUCCESS;
}

// Bulk H2D from pre-pinned host memory (cudaMallocHost).
// src must point into a buffer allocated by kb_pinned_alloc.
extern "C" kb_error_t kb_vec_h2d_pinned(gnark_gpu_context_t, kb_vec_t dst,
                                         const uint32_t *src, size_t n) {
    if (!dst || n > dst->n) return KB_ERROR_INVALID;
    size_t bytes = n * sizeof(uint32_t);
    CUDA_CHECK(cudaMemcpyAsync(dst->d_data, src, bytes, cudaMemcpyHostToDevice, 0));
    CUDA_CHECK(cudaStreamSynchronize(0));
    return KB_SUCCESS;
}

// Allocate page-locked host memory for fast H2D.
extern "C" kb_error_t kb_pinned_alloc(size_t bytes, uint32_t **out) {
    CUDA_CHECK(cudaMallocHost(out, bytes));
    return KB_SUCCESS;
}

extern "C" void kb_pinned_free(uint32_t *ptr) {
    if (ptr) cudaFreeHost(ptr);
}

// ── Vector arithmetic ───────────────────────────────────────────────────────

extern "C" kb_error_t kb_vec_add(gnark_gpu_context_t, kb_vec_t c, kb_vec_t a, kb_vec_t b) {
    if (!c || !a || !b) return KB_ERROR_INVALID;
    size_t n = c->n;
    kern_kb_add<<<kb_grid(n), KB_BLOCK>>>(c->d_data, a->d_data, b->d_data, n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_sub(gnark_gpu_context_t, kb_vec_t c, kb_vec_t a, kb_vec_t b) {
    if (!c || !a || !b) return KB_ERROR_INVALID;
    size_t n = c->n;
    kern_kb_sub<<<kb_grid(n), KB_BLOCK>>>(c->d_data, a->d_data, b->d_data, n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_mul(gnark_gpu_context_t, kb_vec_t c, kb_vec_t a, kb_vec_t b) {
    if (!c || !a || !b) return KB_ERROR_INVALID;
    size_t n = c->n;
    kern_kb_mul<<<kb_grid(n), KB_BLOCK>>>(c->d_data, a->d_data, b->d_data, n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_scale(gnark_gpu_context_t, kb_vec_t v, uint32_t scalar) {
    if (!v) return KB_ERROR_INVALID;
    kern_kb_scale<<<kb_grid(v->n), KB_BLOCK>>>(v->d_data, scalar, v->n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_scale_by_powers(gnark_gpu_context_t, kb_vec_t v, uint32_t g) {
    if (!v) return KB_ERROR_INVALID;
    kern_kb_scale_by_powers<<<kb_grid(v->n), KB_BLOCK>>>(v->d_data, g, v->n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_batch_invert(gnark_gpu_context_t, kb_vec_t, kb_vec_t) {
    return KB_ERROR_INVALID; // TODO
}

// ── NTT ─────────────────────────────────────────────────────────────────────

extern "C" kb_error_t kb_ntt_init(gnark_gpu_context_t, size_t n,
                                   const uint32_t *fwd_tw,
                                   const uint32_t *inv_tw,
                                   kb_ntt_t *out) {
    if (!fwd_tw || !inv_tw || n == 0 || (n & (n-1)) != 0) return KB_ERROR_INVALID;
    auto *d = new(std::nothrow) KBNtt;
    if (!d) return KB_ERROR_OOM;

    size_t half = n / 2;
    size_t bytes = half * sizeof(uint32_t);
    if (cudaMalloc(&d->d_fwd_tw, bytes) != cudaSuccess ||
        cudaMalloc(&d->d_inv_tw, bytes) != cudaSuccess) {
        delete d;
        return KB_ERROR_CUDA;
    }
    cudaMemcpy(d->d_fwd_tw, fwd_tw, bytes, cudaMemcpyHostToDevice);
    cudaMemcpy(d->d_inv_tw, inv_tw, bytes, cudaMemcpyHostToDevice);
    d->n = n;
    d->log_n = ilog2(n);
    *out = d;
    return KB_SUCCESS;
}

extern "C" void kb_ntt_free(kb_ntt_t d) {
    if (d) {
        cudaFree(d->d_fwd_tw);
        cudaFree(d->d_inv_tw);
        delete d;
    }
}

// Bit-reversal permutation: data[i] ↔ data[bitrev(i, log_n)].
// In-place, handles n elements. Only swaps when bitrev(i) > i to avoid double-swap.
__global__ void kern_bitrev(uint32_t *data, int log_n, size_t n) {
    size_t i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;
    size_t j = __brev((unsigned int)i) >> (32 - log_n);
    if (j > i) {
        uint32_t tmp = data[i];
        data[i] = data[j];
        data[j] = tmp;
    }
}

// Bit-reversal for a batch of row-major vectors.
__global__ void kern_batch_bitrev(uint32_t *data, int log_n,
                                   size_t n_rows, size_t row_stride, size_t n) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t total = n_rows * n;
    if (idx >= total) return;

    size_t row = idx / n;
    size_t i   = idx % n;
    size_t j = __brev((unsigned int)i) >> (32 - log_n);
    if (j > i) {
        uint32_t *rd = data + row * row_stride;
        uint32_t tmp = rd[i];
        rd[i] = rd[j];
        rd[j] = tmp;
    }
}

extern "C" kb_error_t kb_vec_bitrev(gnark_gpu_context_t, kb_vec_t v) {
    if (!v) return KB_ERROR_INVALID;
    int log_n = __builtin_ctz((unsigned int)v->n);
    kern_bitrev<<<kb_grid(v->n), KB_BLOCK>>>(v->d_data, log_n, v->n);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_ntt_fwd(gnark_gpu_context_t, kb_ntt_t d, kb_vec_t v) {
    if (!d || !v || v->n != d->n) return KB_ERROR_INVALID;
    size_t pairs = d->n >> 1;
    for (int s = 0; s < d->log_n; s++)
        kern_ntt_dif<<<kb_grid(pairs), KB_BLOCK>>>(v->d_data, d->d_fwd_tw, d->n, s);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_ntt_inv(gnark_gpu_context_t, kb_ntt_t d, kb_vec_t v) {
    if (!d || !v || v->n != d->n) return KB_ERROR_INVALID;
    size_t pairs = d->n >> 1;
    for (int s = d->log_n - 1; s >= 0; s--)
        kern_ntt_dit<<<kb_grid(pairs), KB_BLOCK>>>(v->d_data, d->d_inv_tw, d->n, s);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_ntt_coset_fwd(gnark_gpu_context_t ctx, kb_ntt_t d,
                                        kb_vec_t v, uint32_t g) {
    if (!d || !v || v->n != d->n) return KB_ERROR_INVALID;
    kern_kb_scale_by_powers<<<kb_grid(v->n), KB_BLOCK>>>(v->d_data, g, v->n);
    return kb_ntt_fwd(ctx, d, v);
}

// ── Batch NTT (operates on packed contiguous vectors) ────────────────────────

// Batch coset forward NTT: for each of `batch` vectors of size `n`,
// apply scale-by-powers(g) then DIF NTT then bit-reversal.
// data layout: [vec0[0..n-1], vec1[0..n-1], ..., vec_{batch-1}[0..n-1]]
// Single C call for all vectors → avoids per-vector CGO overhead.
extern "C" kb_error_t kb_ntt_batch_coset_fwd_bitrev(
    gnark_gpu_context_t ctx, kb_ntt_t d,
    uint32_t *data, size_t n, size_t batch, uint32_t g)
{
    if (!d || !data || n != d->n || batch == 0) return KB_ERROR_INVALID;
    int log_n = d->log_n;
    size_t pairs = n >> 1;
    size_t total_pairs = batch * pairs;

    // Scale all vectors by the same coset power table in one launch.
    kern_batch_scale_by_powers<<<kb_grid(batch * n), KB_BLOCK>>>(data, g, batch, n, n);

    // DIF stages over all rows.
    for (int s = 0; s < log_n; s++)
        kern_batch_ntt_dif<<<kb_grid(total_pairs), KB_BLOCK>>>(data, d->d_fwd_tw, n, s, batch, n);

    // Natural-order output.
    kern_batch_bitrev<<<kb_grid(batch * n), KB_BLOCK>>>(data, log_n, batch, n, n);
    return KB_SUCCESS;
}

// Batch IFFT + scale(1/n): for each of `batch` vectors,
// apply bit-reversal then DIT inverse NTT then scale by nInv.
extern "C" kb_error_t kb_ntt_batch_ifft_scale(
    gnark_gpu_context_t ctx, kb_ntt_t d,
    uint32_t *data, size_t n, size_t batch, uint32_t nInv)
{
    if (!d || !data || n != d->n || batch == 0) return KB_ERROR_INVALID;
    int log_n = d->log_n;
    size_t pairs = n >> 1;
    size_t total_pairs = batch * pairs;
    size_t total = batch * n;

    // Natural-order evaluations -> bit-reversed order for DIT inverse path.
    kern_batch_bitrev<<<kb_grid(total), KB_BLOCK>>>(data, log_n, batch, n, n);

    // DIT inverse stages over all rows.
    for (int s = log_n - 1; s >= 0; s--)
        kern_batch_ntt_dit<<<kb_grid(total_pairs), KB_BLOCK>>>(data, d->d_inv_tw, n, s, batch, n);

    // Global scaling by 1/n.
    kern_kb_scale<<<kb_grid(total), KB_BLOCK>>>(data, nInv, total);
    return KB_SUCCESS;
}

// ── Raw pointer variants (for selective per-root operations) ─────────────────

extern "C" kb_error_t kb_ntt_coset_fwd_raw(gnark_gpu_context_t ctx, kb_ntt_t d,
                                             uint32_t *data, uint32_t g) {
    if (!d) return KB_ERROR_INVALID;
    size_t n = d->n;
    kern_kb_scale_by_powers<<<kb_grid(n), KB_BLOCK>>>(data, g, n);
    size_t pairs = n >> 1;
    for (int s = 0; s < d->log_n; s++)
        kern_ntt_dif<<<kb_grid(pairs), KB_BLOCK>>>(data, d->d_fwd_tw, n, s);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_vec_bitrev_raw(gnark_gpu_context_t, uint32_t *data, size_t n) {
    int log_n = __builtin_ctz((unsigned int)n);
    kern_bitrev<<<kb_grid(n), KB_BLOCK>>>(data, log_n, n);
    return KB_SUCCESS;
}

// ── Poseidon2 ───────────────────────────────────────────────────────────────

extern "C" kb_error_t kb_p2_init(gnark_gpu_context_t, int width,
                                  int nb_full_rounds, int nb_partial_rounds,
                                  const uint32_t *round_keys,
                                  const uint32_t *diag,
                                  kb_p2_t *out) {
    if ((width != 16 && width != 24) || !round_keys || !diag) return KB_ERROR_INVALID;
    auto *p = new(std::nothrow) KBPoseidon2;
    if (!p) return KB_ERROR_OOM;

    // Round keys
    int half_rf = nb_full_rounds / 2;
    size_t nkeys = (size_t)half_rf * width + nb_partial_rounds + (size_t)half_rf * width;
    size_t rk_bytes = nkeys * sizeof(uint32_t);

    if (cudaMalloc(&p->d_round_keys, rk_bytes) != cudaSuccess) {
        delete p;
        return KB_ERROR_CUDA;
    }
    cudaMemcpy(p->d_round_keys, round_keys, rk_bytes, cudaMemcpyHostToDevice);

    // Diagonal
    size_t diag_bytes = width * sizeof(uint32_t);
    if (cudaMalloc(&p->d_diag, diag_bytes) != cudaSuccess) {
        cudaFree(p->d_round_keys);
        delete p;
        return KB_ERROR_CUDA;
    }
    cudaMemcpy(p->d_diag, diag, diag_bytes, cudaMemcpyHostToDevice);

    p->width = width;
    p->nb_full_rounds = nb_full_rounds;
    p->nb_partial_rounds = nb_partial_rounds;
    *out = p;
    return KB_SUCCESS;
}

extern "C" void kb_p2_free(kb_p2_t p) {
    if (p) {
        cudaFree(p->d_round_keys);
        cudaFree(p->d_diag);
        delete p;
    }
}

extern "C" kb_error_t kb_p2_compress_batch(gnark_gpu_context_t, kb_p2_t p,
                                            const uint32_t *input, uint32_t *output,
                                            size_t count) {
    if (!p || p->width != 16) return KB_ERROR_INVALID;

    uint32_t *d_in, *d_out;
    CUDA_CHECK(cudaMalloc(&d_in,  count * 16 * sizeof(uint32_t)));
    CUDA_CHECK(cudaMalloc(&d_out, count *  8 * sizeof(uint32_t)));
    CUDA_CHECK(cudaMemcpy(d_in, input, count * 16 * sizeof(uint32_t), cudaMemcpyHostToDevice));

    kern_p2_compress<<<kb_grid(count), KB_BLOCK>>>(d_in, d_out, p->d_round_keys, p->d_diag, count);

    CUDA_CHECK(cudaMemcpy(output, d_out, count * 8 * sizeof(uint32_t), cudaMemcpyDeviceToHost));
    cudaFree(d_in);
    cudaFree(d_out);
    return KB_SUCCESS;
}

extern "C" kb_error_t kb_p2_sponge_batch(gnark_gpu_context_t, kb_p2_t p,
                                           const uint32_t *input, size_t input_len,
                                           uint32_t *output, size_t count) {
    if (!p || p->width != 24) return KB_ERROR_INVALID;

    uint32_t *d_in, *d_out;
    CUDA_CHECK(cudaMalloc(&d_in,  count * input_len * sizeof(uint32_t)));
    CUDA_CHECK(cudaMalloc(&d_out, count * 8 * sizeof(uint32_t)));
    CUDA_CHECK(cudaMemcpy(d_in, input, count * input_len * sizeof(uint32_t), cudaMemcpyHostToDevice));

    kern_p2_sponge<<<kb_grid(count), KB_BLOCK>>>(d_in, input_len, d_out,
                                                   p->d_round_keys, p->d_diag, count);

    CUDA_CHECK(cudaMemcpy(output, d_out, count * 8 * sizeof(uint32_t), cudaMemcpyDeviceToHost));
    cudaFree(d_in);
    cudaFree(d_out);
    return KB_SUCCESS;
}

// ── SIS ─────────────────────────────────────────────────────────────────────

extern "C" kb_error_t kb_sis_init(gnark_gpu_context_t,
                                   int degree, int n_polys, int log_two_bound,
                                   const uint32_t *ag,
                                   const uint32_t *fwd_tw,
                                   const uint32_t *inv_tw,
                                   const uint32_t *coset_table,
                                   const uint32_t *coset_inv,
                                   kb_sis_t *out) {
    if (degree <= 0 || (degree & (degree-1)) != 0) return KB_ERROR_INVALID;
    if (!ag || !fwd_tw || !inv_tw || !coset_table || !coset_inv) return KB_ERROR_INVALID;

    auto *s = new(std::nothrow) KBSis;
    if (!s) return KB_ERROR_OOM;

    size_t deg_bytes = degree * sizeof(uint32_t);
    size_t half_bytes = (degree / 2) * sizeof(uint32_t);
    size_t ag_bytes = (size_t)n_polys * deg_bytes;

    #define SIS_ALLOC(ptr, sz) if (cudaMalloc(&ptr, sz) != cudaSuccess) { delete s; return KB_ERROR_CUDA; }
    SIS_ALLOC(s->d_ag, ag_bytes);
    SIS_ALLOC(s->d_fwd_tw, half_bytes);
    SIS_ALLOC(s->d_inv_tw, half_bytes);
    SIS_ALLOC(s->d_coset_table, deg_bytes);
    SIS_ALLOC(s->d_coset_inv, deg_bytes);
    #undef SIS_ALLOC

    cudaMemcpy(s->d_ag, ag, ag_bytes, cudaMemcpyHostToDevice);
    cudaMemcpy(s->d_fwd_tw, fwd_tw, half_bytes, cudaMemcpyHostToDevice);
    cudaMemcpy(s->d_inv_tw, inv_tw, half_bytes, cudaMemcpyHostToDevice);
    cudaMemcpy(s->d_coset_table, coset_table, deg_bytes, cudaMemcpyHostToDevice);
    cudaMemcpy(s->d_coset_inv, coset_inv, deg_bytes, cudaMemcpyHostToDevice);

    s->degree = degree;
    s->log_degree = ilog2(degree);
    s->n_polys = n_polys;
    s->log_two_bound = log_two_bound;
    *out = s;
    return KB_SUCCESS;
}

extern "C" void kb_sis_free(kb_sis_t s) {
    if (s) {
        cudaFree(s->d_ag);
        cudaFree(s->d_fwd_tw);
        cudaFree(s->d_inv_tw);
        cudaFree(s->d_coset_table);
        cudaFree(s->d_coset_inv);
        delete s;
    }
}

// ── Vortex pipeline (pre-allocated device buffers) ───────────────────────────

static bool kb_timing_enabled() {
    static int val = -1;
    if (val < 0) val = getenv("KB_VORTEX_TIMING") ? 1 : 0;
    return val == 1;
}

extern "C" void kb_vortex_pipeline_free(kb_vortex_pipeline_t p) {
    if (!p) return;
    cudaFree(p->d_work);
    cudaFree(p->d_rs_fwd_tw);
    cudaFree(p->d_rs_inv_tw);
    cudaFree(p->d_scaled_coset);
    cudaFree(p->d_coeffs);
    cudaFree(p->d_coset_tables);
    cudaFree(p->d_encoded_col);
    cudaFree(p->d_sis);
    cudaFree(p->d_leaves);
    cudaFree(p->d_tree);
    // Async extraction buffers
    cudaFree(p->d_enc_rowmajor);
    cudaFreeHost(p->h_enc_pinned);
    cudaFreeHost(p->h_sis_pinned);
    cudaFreeHost(p->h_leaves_pinned);
    if (p->ev_rs_done)  cudaEventDestroy(p->ev_rs_done);
    if (p->ev_sis_done) cudaEventDestroy(p->ev_sis_done);
    if (p->ev_p2_done)  cudaEventDestroy(p->ev_p2_done);
    // Original buffers
    cudaFreeHost(p->h_input);
    cudaFreeHost(p->h_tree);
    if (p->stream_xfer)   cudaStreamDestroy(p->stream_xfer);
    if (p->stream_compute) cudaStreamDestroy(p->stream_compute);
    if (p->h2d_event)     cudaEventDestroy(p->h2d_event);
    delete p;
}

extern "C" kb_error_t kb_vortex_pipeline_init(gnark_gpu_context_t,
                                               kb_sis_t sis,
                                               kb_p2_t p2_sponge,
                                               kb_p2_t p2_compress,
                                               size_t max_n_rows,
                                               size_t n_cols,
                                               int rate,
                                               const uint32_t *rs_fwd_tw,
                                               const uint32_t *rs_inv_tw,
                                               const uint32_t *scaled_coset_br,
                                               kb_vortex_pipeline_t *out) {
    if (!sis || !p2_sponge || !p2_compress || !max_n_rows || !n_cols)
        return KB_ERROR_INVALID;
    if (!rs_fwd_tw || !rs_inv_tw || !scaled_coset_br)
        return KB_ERROR_INVALID;

    auto *p = new(std::nothrow) KBVortexPipeline{};
    if (!p) return KB_ERROR_OOM;

    p->sis = sis;
    p->p2_sponge = p2_sponge;
    p->p2_compress = p2_compress;
    p->max_n_rows = max_n_rows;
    p->n_cols = n_cols;
    p->rate = rate;
    p->size_codeword = n_cols * rate;
    p->tree_np = next_pow2(p->size_codeword);
    p->log_n_cols = ilog2(n_cols);
    p->degree = sis->degree;

    size_t scw = p->size_codeword;
    int degree = p->degree;
    size_t np = p->tree_np;
    size_t half_n = n_cols / 2;

    // Device buffers
    #define PIPE_ALLOC(ptr, nbytes) \
        if (cudaMalloc(&(ptr), (nbytes)) != cudaSuccess) { \
            kb_vortex_pipeline_free(p); return KB_ERROR_CUDA; }
    PIPE_ALLOC(p->d_work,          max_n_rows * n_cols * sizeof(uint32_t));
    PIPE_ALLOC(p->d_rs_fwd_tw,    half_n * sizeof(uint32_t));
    PIPE_ALLOC(p->d_rs_inv_tw,    half_n * sizeof(uint32_t));
    PIPE_ALLOC(p->d_scaled_coset, n_cols * sizeof(uint32_t));
    PIPE_ALLOC(p->d_encoded_col,  scw * max_n_rows * sizeof(uint32_t));
    PIPE_ALLOC(p->d_sis,          scw * degree * sizeof(uint32_t));
    PIPE_ALLOC(p->d_leaves,       scw * 8 * sizeof(uint32_t));
    PIPE_ALLOC(p->d_tree,         2 * np * 8 * sizeof(uint32_t));
    // d_enc_rowmajor: lazy-allocated on first use (commit_and_extract / extract_all_rowmajor).
    // Not needed by the CommitDirect + SnapshotEncoded path, saving ~scw*maxRows*4 bytes.
    p->d_enc_rowmajor = nullptr;
    #undef PIPE_ALLOC

    // Upload RS domain data (one-time)
    cudaMemcpy(p->d_rs_fwd_tw,    rs_fwd_tw,        half_n * sizeof(uint32_t), cudaMemcpyHostToDevice);
    cudaMemcpy(p->d_rs_inv_tw,    rs_inv_tw,        half_n * sizeof(uint32_t), cudaMemcpyHostToDevice);
    cudaMemcpy(p->d_scaled_coset, scaled_coset_br, n_cols  * sizeof(uint32_t), cudaMemcpyHostToDevice);

    // Pinned host buffers
    #define PIN_ALLOC(ptr, nbytes) \
        if (cudaMallocHost(&(ptr), (nbytes)) != cudaSuccess) { \
            kb_vortex_pipeline_free(p); return KB_ERROR_CUDA; }
    PIN_ALLOC(p->h_input, max_n_rows * n_cols * sizeof(uint32_t));
    PIN_ALLOC(p->h_tree,  (2 * np - 1) * 8 * sizeof(uint32_t));
    // h_sis_pinned: always allocated — used by commit's overlapped SIS D2H.
    PIN_ALLOC(p->h_sis_pinned, scw * degree * sizeof(uint32_t));
    // h_enc_pinned, h_leaves_pinned: lazy (only for commit_and_extract path).
    p->h_enc_pinned = nullptr;
    p->h_leaves_pinned = nullptr;
    #undef PIN_ALLOC

    // Streams for pipelined H2D + compute
    CUDA_CHECK(cudaStreamCreate(&p->stream_xfer));
    CUDA_CHECK(cudaStreamCreate(&p->stream_compute));
    CUDA_CHECK(cudaEventCreateWithFlags(&p->h2d_event, cudaEventDisableTiming));
    // Events for async extraction overlap
    CUDA_CHECK(cudaEventCreateWithFlags(&p->ev_rs_done, cudaEventDisableTiming));
    CUDA_CHECK(cudaEventCreateWithFlags(&p->ev_sis_done, cudaEventDisableTiming));
    CUDA_CHECK(cudaEventCreateWithFlags(&p->ev_p2_done, cudaEventDisableTiming));

    *out = p;
    return KB_SUCCESS;
}

// Accessors for pinned host buffers (Go wraps these as zero-copy slices)
extern "C" uint32_t *kb_vortex_pipeline_input_buf(kb_vortex_pipeline_t p) { return p ? p->h_input : nullptr; }
extern "C" uint32_t *kb_vortex_pipeline_tree_buf(kb_vortex_pipeline_t p)  { return p ? p->h_tree  : nullptr; }

// Set coset scaling tables for rate > 2 RS encoding (multi-coset NTT).
// coset_tables: [(rate-1) × n_cols] flat array, table k at offset k*n_cols.
// Each table: coset_k_br[j] = (Ω^k)^{bitrev(j)} / n  (bit-reversed, normalized).
extern "C" kb_error_t kb_vortex_pipeline_set_coset_tables(
    kb_vortex_pipeline_t p,
    const uint32_t *coset_tables,
    size_t n_tables) {
    if (!p || !coset_tables) return KB_ERROR_INVALID;
    if ((int)n_tables != p->rate - 1) return KB_ERROR_SIZE;

    size_t nc = p->n_cols;
    size_t table_bytes = n_tables * nc * sizeof(uint32_t);
    size_t coeffs_bytes = p->max_n_rows * nc * sizeof(uint32_t);

    // Allocate coset tables on device
    if (!p->d_coset_tables) {
        if (cudaMalloc(&p->d_coset_tables, table_bytes) != cudaSuccess)
            return KB_ERROR_CUDA;
    }
    cudaMemcpy(p->d_coset_tables, coset_tables, table_bytes, cudaMemcpyHostToDevice);

    // Allocate coefficients backup buffer
    if (!p->d_coeffs) {
        if (cudaMalloc(&p->d_coeffs, coeffs_bytes) != cudaSuccess)
            return KB_ERROR_CUDA;
    }

    return KB_SUCCESS;
}

// ── Merkle tree ─────────────────────────────────────────────────────────────

extern "C" kb_error_t kb_merkle_build(gnark_gpu_context_t, kb_p2_t p,
                                       const uint32_t *leaves, size_t n_leaves,
                                       uint32_t *tree_buf) {
    if (!p || p->width != 16) return KB_ERROR_INVALID;

    size_t np = next_pow2(n_leaves);
    size_t total_nodes = 2 * np;
    size_t hash_bytes = 8 * sizeof(uint32_t);

    uint32_t *d_tree;
    CUDA_CHECK(cudaMalloc(&d_tree, total_nodes * hash_bytes));
    CUDA_CHECK(cudaMemset(d_tree, 0, total_nodes * hash_bytes));

    // Copy leaves into bottom level (indices np .. np+n_leaves-1)
    CUDA_CHECK(cudaMemcpy(d_tree + np * 8, leaves, n_leaves * hash_bytes, cudaMemcpyHostToDevice));

    // Build bottom-up
    for (size_t level_size = np; level_size > 1; level_size >>= 1) {
        size_t parent_start = level_size / 2;
        size_t n_pairs = level_size / 2;
        kern_merkle_level<<<kb_grid(n_pairs), KB_BLOCK>>>(
            d_tree + level_size * 8,
            d_tree + parent_start * 8,
            p->d_round_keys, p->d_diag,
            n_pairs);
    }

    // Copy tree back (skip index 0; root at index 1)
    CUDA_CHECK(cudaMemcpy(tree_buf, d_tree + 8, (total_nodes - 1) * hash_bytes, cudaMemcpyDeviceToHost));
    cudaFree(d_tree);
    return KB_SUCCESS;
}

// ── Linear combination ──────────────────────────────────────────────────────

extern "C" kb_error_t kb_lincomb_e4(gnark_gpu_context_t,
                                     kb_vec_t *rows, size_t n_rows, size_t n_cols,
                                     const uint32_t alpha_raw[4], uint32_t *result) {
    if (!rows || n_rows == 0 || n_cols == 0) return KB_ERROR_INVALID;

    const uint32_t **h_ptrs = new const uint32_t*[n_rows];
    for (size_t i = 0; i < n_rows; i++) {
        if (!rows[i]) { delete[] h_ptrs; return KB_ERROR_INVALID; }
        h_ptrs[i] = rows[i]->d_data;
    }

    const uint32_t **d_ptrs;
    CUDA_CHECK(cudaMalloc(&d_ptrs, n_rows * sizeof(uint32_t*)));
    CUDA_CHECK(cudaMemcpy((void*)d_ptrs, h_ptrs, n_rows * sizeof(uint32_t*), cudaMemcpyHostToDevice));
    delete[] h_ptrs;

    uint32_t *d_result;
    CUDA_CHECK(cudaMalloc(&d_result, n_cols * 4 * sizeof(uint32_t)));

    E4 alpha = {{alpha_raw[0], alpha_raw[1]}, {alpha_raw[2], alpha_raw[3]}};
    kern_lincomb_e4<<<kb_grid(n_cols), KB_BLOCK>>>(d_ptrs, n_rows, n_cols, alpha, d_result);

    CUDA_CHECK(cudaMemcpy(result, d_result, n_cols * 4 * sizeof(uint32_t), cudaMemcpyDeviceToHost));
    cudaFree((void*)d_ptrs);
    cudaFree(d_result);
    return KB_SUCCESS;
}

// ═════════════════════════════════════════════════════════════════════════════
// Vortex commit pipeline — GPU RS encode + SIS + Merkle
// ═════════════════════════════════════════════════════════════════════════════
//
// RS encode (rate=2) runs entirely on GPU via batch NTT, eliminating CPU RS
// and halving H2D data volume (only raw rows, not encoded codewords).
//
//  raw_rows [nR × nC]                    d_encoded_col [scw × nR]
//  host, pinned       ──H2D──▶ d_work
//                               │ scatter_transpose → even cols
//                               │ iFFT_DIF(inv_tw) + scale(cosetBR·cardInv) + FFT_DIT(fwd_tw)
//                               └─scatter_transpose → odd cols
//
//  d_encoded_col ──SIS──▶ d_sis ──sponge──▶ d_leaves ──merkle──▶ d_tree
//
// Set KB_VORTEX_TIMING=1 for per-phase CUDA event timing on stderr.

// Helper: RS encode chunk_rows rows starting at d_chunk in d_work,
// scatter even/odd columns into d_encoded_col offset by row_off.
static void rs_encode_chunk(KBVortexPipeline *p, uint32_t *d_chunk,
                             size_t chunk_rows, size_t total_rows,
                             size_t row_off, cudaStream_t s) {
    size_t nc = p->n_cols;
    int log_nc = p->log_n_cols;

    // Transpose original → even columns (d_encoded_col + row_off)
    {
        dim3 blk(32, 8);
        dim3 grd(((unsigned)nc + 31) / 32, ((unsigned)chunk_rows + 31) / 32);
        kern_scatter_transpose<<<grd, blk, 0, s>>>(
            d_chunk, p->d_encoded_col + row_off,
            chunk_rows, nc, total_rows, 2, 0);
    }

    // RS encode (fused NTT: global DIF → fused tile → global DIT)
    {
        size_t pairs_per_row = nc >> 1;
        size_t total_pairs = chunk_rows * pairs_per_row;
        int grid1d = kb_grid(total_pairs);

        int tile_log = log_nc < 13 ? log_nc : 13;
        int s_cut    = log_nc - tile_log;

        for (int st = 0; st < s_cut; st++)
            kern_batch_ntt_dif<<<grid1d, KB_BLOCK, 0, s>>>(
                d_chunk, p->d_rs_inv_tw, nc, st, chunk_rows, nc);

        {
            int tile_size = 1 << tile_log;
            int tiles_per_row = (int)(nc >> tile_log);
            int n_blocks = (int)((size_t)tiles_per_row * chunk_rows);
            int threads = tile_size >> 1;
            if (threads > 1024) threads = 1024;
            size_t fused_smem = tile_size * sizeof(uint32_t);

            kern_batch_ntt_fused<<<n_blocks, threads, fused_smem, s>>>(
                d_chunk, p->d_rs_inv_tw, p->d_rs_fwd_tw, p->d_scaled_coset,
                log_nc, s_cut, chunk_rows, nc);
        }

        for (int st = s_cut - 1; st >= 0; st--)
            kern_batch_ntt_dit<<<grid1d, KB_BLOCK, 0, s>>>(
                d_chunk, p->d_rs_fwd_tw, nc, st, chunk_rows, nc);
    }

    // Transpose FFT result → odd columns
    {
        dim3 blk(32, 8);
        dim3 grd(((unsigned)nc + 31) / 32, ((unsigned)chunk_rows + 31) / 32);
        kern_scatter_transpose<<<grd, blk, 0, s>>>(
            d_chunk, p->d_encoded_col + row_off,
            chunk_rows, nc, total_rows, 2, 1);
    }
}

// RS encode for rate > 2 via multi-coset NTT.
//
// For rate ρ, the codeword evaluates f on ρ cosets of the small domain:
//   coset k = Ω^k · {ω⁰, ω¹, ..., ω^{n-1}},  k = 0..ρ-1
//
// Algorithm per chunk:
//   1. Scatter original values → d_encoded_col at stride=ρ, offset=0
//   2. Partial IFFT (s_cut global DIF stages) on d_chunk
//   3. Save partial state → d_coeffs
//   4. For k = 1..ρ-1:
//      a. Restore d_coeffs → d_chunk
//      b. Fused IFFT tail + scale(coset_k_br) + FFT head
//      c. Global DIT stages
//      d. Scatter → d_encoded_col at stride=ρ, offset=k
static void rs_encode_chunk_general(KBVortexPipeline *p, uint32_t *d_chunk,
                                     size_t chunk_rows, size_t total_rows,
                                     size_t row_off, cudaStream_t s) {
    size_t nc = p->n_cols;
    int log_nc = p->log_n_cols;
    int rate = p->rate;

    // 1. Scatter original values → coset 0 (stride=ρ, offset=0)
    {
        dim3 blk(32, 8);
        dim3 grd(((unsigned)nc + 31) / 32, ((unsigned)chunk_rows + 31) / 32);
        kern_scatter_transpose<<<grd, blk, 0, s>>>(
            d_chunk, p->d_encoded_col + row_off,
            chunk_rows, nc, total_rows, rate, 0);
    }

    // 2. Partial IFFT: global DIF stages [0..s_cut)
    size_t pairs_per_row = nc >> 1;
    size_t total_pairs = chunk_rows * pairs_per_row;
    int grid1d = kb_grid(total_pairs);

    int tile_log = log_nc < 13 ? log_nc : 13;
    int s_cut    = log_nc - tile_log;

    for (int st = 0; st < s_cut; st++)
        kern_batch_ntt_dif<<<grid1d, KB_BLOCK, 0, s>>>(
            d_chunk, p->d_rs_inv_tw, nc, st, chunk_rows, nc);

    // 3. Save partial IFFT state
    cudaMemcpyAsync(p->d_coeffs, d_chunk,
                    chunk_rows * nc * sizeof(uint32_t),
                    cudaMemcpyDeviceToDevice, s);

    // 4. For each coset k = 1..rate-1
    int tile_size = 1 << tile_log;
    int tiles_per_row = (int)(nc >> tile_log);
    int n_blocks = (int)((size_t)tiles_per_row * chunk_rows);
    int threads = tile_size >> 1;
    if (threads > 1024) threads = 1024;
    size_t fused_smem = tile_size * sizeof(uint32_t);

    for (int k = 1; k < rate; k++) {
        // 4a. Restore partial state
        cudaMemcpyAsync(d_chunk, p->d_coeffs,
                        chunk_rows * nc * sizeof(uint32_t),
                        cudaMemcpyDeviceToDevice, s);

        // 4b. Fused: complete IFFT + scale by coset_k + start FFT
        const uint32_t *coset_k = p->d_coset_tables + (size_t)(k - 1) * nc;
        kern_batch_ntt_fused<<<n_blocks, threads, fused_smem, s>>>(
            d_chunk, p->d_rs_inv_tw, p->d_rs_fwd_tw, coset_k,
            log_nc, s_cut, chunk_rows, nc);

        // 4c. Complete FFT: global DIT stages [s_cut-1..0]
        for (int st = s_cut - 1; st >= 0; st--)
            kern_batch_ntt_dit<<<grid1d, KB_BLOCK, 0, s>>>(
                d_chunk, p->d_rs_fwd_tw, nc, st, chunk_rows, nc);

        // 4d. Scatter → d_encoded_col at stride=ρ, offset=k
        dim3 blk(32, 8);
        dim3 grd(((unsigned)nc + 31) / 32, ((unsigned)chunk_rows + 31) / 32);
        kern_scatter_transpose<<<grd, blk, 0, s>>>(
            d_chunk, p->d_encoded_col + row_off,
            chunk_rows, nc, total_rows, rate, k);
    }
}

// Forward declaration for transpose kernel (defined later with kb_vortex_extract_all_rowmajor).
__global__ void kern_transpose_col_to_row(const uint32_t *__restrict__ col_major,
                                           uint32_t *__restrict__ row_major,
                                           size_t n_rows, size_t scw);

extern "C" kb_error_t kb_vortex_commit(kb_vortex_pipeline_t p,
                                        const uint32_t *raw_rows,
                                        size_t n_rows) {
    if (!p || !raw_rows) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;

    size_t nc  = p->n_cols;
    size_t scw = p->size_codeword;
    int degree = p->degree;
    size_t np  = p->tree_np;

    cudaStream_t sx = p->stream_xfer;
    cudaStream_t sc = p->stream_compute;

    // ── Optional CUDA event timing ───────────────────────────────────────
    cudaEvent_t t[6];
    bool timing = kb_timing_enabled();
    if (timing) {
        for (int i = 0; i < 6; i++) cudaEventCreate(&t[i]);
        cudaEventRecord(t[0], sc);
    }

    // ── 1-4. Pipelined H2D + RS encode (2-chunk overlap) ────────────────
    //
    //  stream_xfer:   [H2D chunk0]──event──[H2D chunk1]──event──
    //  stream_compute:              [RS encode chunk0]   [RS encode chunk1]
    //
    //  Chunk 0 occupies d_work[0 .. c0*nc), chunk 1 occupies d_work[c0*nc ..].
    //  Row offset into d_encoded_col ensures correct global row positions.
    {
        // Chunk rows for H2D overlap + L2 cache locality.
        // Sweet spot: ~32 rows/chunk at nc=2^19 (64 MB, fits 96 MB L2).
        int N_CHUNKS = (int)n_rows / 32;
        if (N_CHUNKS < 1) N_CHUNKS = 1;
        size_t chunk_size = n_rows / N_CHUNKS;
        for (int k = 0; k < N_CHUNKS; k++) {
            size_t c_rows = (k < N_CHUNKS - 1) ? chunk_size : n_rows - k * chunk_size;
            size_t row_off = k * chunk_size;
            uint32_t *d_chunk = p->d_work + row_off * nc;

            // Async H2D on transfer stream
            CUDA_CHECK(cudaMemcpyAsync(d_chunk, raw_rows + row_off * nc,
                                       c_rows * nc * sizeof(uint32_t),
                                       cudaMemcpyHostToDevice, sx));
            cudaEventRecord(p->h2d_event, sx);

            // Compute stream waits for this chunk's H2D
            cudaStreamWaitEvent(sc, p->h2d_event);

            // RS encode this chunk on compute stream
            if (p->rate == 2)
                rs_encode_chunk(p, d_chunk, c_rows, n_rows, row_off, sc);
            else
                rs_encode_chunk_general(p, d_chunk, c_rows, n_rows, row_off, sc);
        }
    }
    if (timing) cudaEventRecord(t[1], sc);

    // ── 5. SIS hash (all columns, needs full d_encoded_col) ──────────────
    size_t smem = 2 * degree * sizeof(uint32_t);
    kern_sis_hash<<<(int)scw, KB_BLOCK, smem, sc>>>(
        p->d_encoded_col, (int)n_rows, (int)scw,
        p->sis->d_ag, p->sis->n_polys, degree, p->sis->log_degree,
        p->sis->d_fwd_tw, p->sis->d_inv_tw,
        p->sis->d_coset_table, p->sis->d_coset_inv,
        p->d_sis);
    if (timing) cudaEventRecord(t[2], sc);

    // ── 6. Poseidon2 MD hash (width=16): SIS hashes → leaf hashes ─────
    // Matches CPU CompressPoseidon2x16: iterative Davies-Meyer, width=16.
    kern_p2_md_hash<<<kb_grid(scw), KB_BLOCK, 0, sc>>>(
        p->d_sis, (size_t)degree, p->d_leaves,
        p->p2_compress->d_round_keys, p->p2_compress->d_diag, scw);
    if (timing) cudaEventRecord(t[3], sc);

    // ── 7. Overlap: SIS D2H on transfer stream while Merkle builds on compute ─
    //  After P2 hash, d_sis is no longer read — safe to D2H on sx.
    cudaEventRecord(p->h2d_event, sc);  // mark P2 hash completion
    cudaStreamWaitEvent(sx, p->h2d_event);
    CUDA_CHECK(cudaMemcpyAsync(p->h_sis_pinned, p->d_sis,
                               scw * degree * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sx));

    // ── 8. Merkle tree (bottom-up Poseidon2 compression) on compute stream ─
    CUDA_CHECK(cudaMemsetAsync(p->d_tree, 0, 2 * np * 8 * sizeof(uint32_t), sc));
    CUDA_CHECK(cudaMemcpyAsync(p->d_tree + np * 8, p->d_leaves,
                               scw * 8 * sizeof(uint32_t),
                               cudaMemcpyDeviceToDevice, sc));
    for (size_t level = np; level > 1; level >>= 1) {
        size_t n_pairs = level / 2;
        kern_merkle_level<<<kb_grid(n_pairs), KB_BLOCK, 0, sc>>>(
            p->d_tree + level * 8,
            p->d_tree + (level / 2) * 8,
            p->p2_compress->d_round_keys, p->p2_compress->d_diag, n_pairs);
    }
    if (timing) cudaEventRecord(t[4], sc);

    // ── 9. D2H tree → pinned host buffer ────────────────────────────────
    CUDA_CHECK(cudaMemcpyAsync(p->h_tree, p->d_tree + 8,
                               (2 * np - 1) * 8 * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sc));
    CUDA_CHECK(cudaStreamSynchronize(sc));
    CUDA_CHECK(cudaStreamSynchronize(sx));  // ensure SIS D2H is done too
    if (timing) cudaEventRecord(t[5], sc);

    // ── Print timing ─────────────────────────────────────────────────────
    if (timing) {
        cudaDeviceSynchronize();
        static const char *labels[] = {
            "H2D+RS encode", "SIS hash", "sponge", "merkle", "D2H tree"
        };
        fprintf(stderr, "vortex_commit (n_rows=%zu, nc=%zu, scw=%zu):\n",
                n_rows, nc, scw);
        for (int i = 0; i < 5; i++) {
            float ms;
            cudaEventElapsedTime(&ms, t[i], t[i + 1]);
            fprintf(stderr, "  %-16s %8.2f ms\n", labels[i], ms);
        }
        float total;
        cudaEventElapsedTime(&total, t[0], t[5]);
        fprintf(stderr, "  %-16s %8.2f ms\n", "TOTAL", total);
        for (int i = 0; i < 6; i++) cudaEventDestroy(t[i]);
    }

    return KB_SUCCESS;
}

// ═════════════════════════════════════════════════════════════════════════════
// Async commit + extract: overlaps D2H with SIS/P2/Merkle compute
// ═════════════════════════════════════════════════════════════════════════════
//
// Timeline:
//   stream_compute:  [H2D+RS encode] → ev_rs_done → [SIS] → ev_sis_done → [P2] → ev_p2_done → [Merkle] → [D2H tree]
//   stream_xfer:                       wait(ev_rs) → [transpose] → [D2H enc]
//                                                     wait(ev_sis) → [D2H sis]
//                                                                     wait(ev_p2) → [D2H leaves]
//
// After both streams sync, pinned host buffers contain:
//   h_enc_pinned:    row-major encoded matrix [n_rows × scw]
//   h_sis_pinned:    flat SIS hashes [scw × degree]
//   h_leaves_pinned: leaf hashes [scw × 8]
//   h_tree:          Merkle tree (heap layout)

extern "C" kb_error_t kb_vortex_commit_and_extract(kb_vortex_pipeline_t p,
                                                     const uint32_t *raw_rows,
                                                     size_t n_rows) {
    if (!p || !raw_rows) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;

    size_t nc  = p->n_cols;
    size_t scw = p->size_codeword;
    int degree = p->degree;
    size_t np  = p->tree_np;

    // Lazy-allocate buffers only needed by commit_and_extract path
    if (!p->d_enc_rowmajor) {
        if (cudaMalloc(&p->d_enc_rowmajor, scw * p->max_n_rows * sizeof(uint32_t)) != cudaSuccess)
            return KB_ERROR_CUDA;
    }
    if (!p->h_enc_pinned) {
        if (cudaMallocHost(&p->h_enc_pinned, scw * p->max_n_rows * sizeof(uint32_t)) != cudaSuccess)
            return KB_ERROR_CUDA;
    }
    if (!p->h_sis_pinned) {
        if (cudaMallocHost(&p->h_sis_pinned, scw * degree * sizeof(uint32_t)) != cudaSuccess)
            return KB_ERROR_CUDA;
    }
    if (!p->h_leaves_pinned) {
        if (cudaMallocHost(&p->h_leaves_pinned, scw * 8 * sizeof(uint32_t)) != cudaSuccess)
            return KB_ERROR_CUDA;
    }

    cudaStream_t sx = p->stream_xfer;
    cudaStream_t sc = p->stream_compute;

    bool timing = (getenv("KB_VORTEX_TIMING") && atoi(getenv("KB_VORTEX_TIMING")));
    cudaEvent_t t0, t1, t2, t3, t4, t5;
    if (timing) {
        cudaEventCreate(&t0); cudaEventCreate(&t1); cudaEventCreate(&t2);
        cudaEventCreate(&t3); cudaEventCreate(&t4); cudaEventCreate(&t5);
        CUDA_CHECK(cudaDeviceSynchronize());
        cudaEventRecord(t0, sc);
    }

    // ── 1-4. Pipelined H2D + RS encode (same as kb_vortex_commit) ────────
    {
        int N_CHUNKS = (int)n_rows / 32;
        if (N_CHUNKS < 1) N_CHUNKS = 1;
        size_t chunk_size = n_rows / N_CHUNKS;
        for (int k = 0; k < N_CHUNKS; k++) {
            size_t c_rows = (k < N_CHUNKS - 1) ? chunk_size : n_rows - k * chunk_size;
            size_t row_off = k * chunk_size;
            uint32_t *d_chunk = p->d_work + row_off * nc;

            CUDA_CHECK(cudaMemcpyAsync(d_chunk, raw_rows + row_off * nc,
                                       c_rows * nc * sizeof(uint32_t),
                                       cudaMemcpyHostToDevice, sx));
            cudaEventRecord(p->h2d_event, sx);
            cudaStreamWaitEvent(sc, p->h2d_event);

            if (p->rate == 2)
                rs_encode_chunk(p, d_chunk, c_rows, n_rows, row_off, sc);
            else
                rs_encode_chunk_general(p, d_chunk, c_rows, n_rows, row_off, sc);
        }
    }

    // ── RS encoding complete: signal xfer stream to start D2H ────────────
    if (timing) cudaEventRecord(t1, sc);
    cudaEventRecord(p->ev_rs_done, sc);

    // ── stream_xfer: transpose + async D2H of encoded matrix ─────────────
    cudaStreamWaitEvent(sx, p->ev_rs_done);
    size_t enc_total = scw * n_rows;
    kern_transpose_col_to_row<<<kb_grid(enc_total), KB_BLOCK, 0, sx>>>(
        p->d_encoded_col, p->d_enc_rowmajor, n_rows, scw);
    CUDA_CHECK(cudaMemcpyAsync(p->h_enc_pinned, p->d_enc_rowmajor,
                               enc_total * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sx));

    // ── stream_compute: SIS hash ─────────────────────────────────────────
    if (timing) cudaEventRecord(t2, sc);
    size_t smem = 2 * degree * sizeof(uint32_t);
    kern_sis_hash<<<(int)scw, KB_BLOCK, smem, sc>>>(
        p->d_encoded_col, (int)n_rows, (int)scw,
        p->sis->d_ag, p->sis->n_polys, degree, p->sis->log_degree,
        p->sis->d_fwd_tw, p->sis->d_inv_tw,
        p->sis->d_coset_table, p->sis->d_coset_inv,
        p->d_sis);
    cudaEventRecord(p->ev_sis_done, sc);

    // ── stream_xfer: async D2H of SIS hashes (after SIS compute done) ───
    cudaStreamWaitEvent(sx, p->ev_sis_done);
    CUDA_CHECK(cudaMemcpyAsync(p->h_sis_pinned, p->d_sis,
                               scw * degree * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sx));

    // ── stream_compute: Poseidon2 MD hash (width=16) ───────────────────
    if (timing) cudaEventRecord(t3, sc);
    kern_p2_md_hash<<<kb_grid(scw), KB_BLOCK, 0, sc>>>(
        p->d_sis, (size_t)degree, p->d_leaves,
        p->p2_compress->d_round_keys, p->p2_compress->d_diag, scw);
    cudaEventRecord(p->ev_p2_done, sc);

    // ── stream_xfer: async D2H of leaves (after P2 compute done) ────────
    cudaStreamWaitEvent(sx, p->ev_p2_done);
    CUDA_CHECK(cudaMemcpyAsync(p->h_leaves_pinned, p->d_leaves,
                               scw * 8 * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sx));

    // ── stream_compute: Merkle tree + D2H tree ──────────────────────────
    if (timing) cudaEventRecord(t4, sc);
    CUDA_CHECK(cudaMemsetAsync(p->d_tree, 0, 2 * np * 8 * sizeof(uint32_t), sc));
    CUDA_CHECK(cudaMemcpyAsync(p->d_tree + np * 8, p->d_leaves,
                               scw * 8 * sizeof(uint32_t),
                               cudaMemcpyDeviceToDevice, sc));
    for (size_t level = np; level > 1; level >>= 1) {
        size_t n_pairs = level / 2;
        kern_merkle_level<<<kb_grid(n_pairs), KB_BLOCK, 0, sc>>>(
            p->d_tree + level * 8,
            p->d_tree + (level / 2) * 8,
            p->p2_compress->d_round_keys, p->p2_compress->d_diag, n_pairs);
    }
    CUDA_CHECK(cudaMemcpyAsync(p->h_tree, p->d_tree + 8,
                               (2 * np - 1) * 8 * sizeof(uint32_t),
                               cudaMemcpyDeviceToHost, sc));

    // ── Sync both streams ────────────────────────────────────────────────
    CUDA_CHECK(cudaStreamSynchronize(sc));
    CUDA_CHECK(cudaStreamSynchronize(sx));

    if (timing) {
        cudaEventRecord(t5, sc);
        cudaEventSynchronize(t5);
        float ms01, ms12, ms23, ms34, ms45, ms05;
        cudaEventElapsedTime(&ms01, t0, t1);
        cudaEventElapsedTime(&ms12, t1, t2);
        cudaEventElapsedTime(&ms23, t2, t3);
        cudaEventElapsedTime(&ms34, t3, t4);
        cudaEventElapsedTime(&ms45, t4, t5);
        cudaEventElapsedTime(&ms05, t0, t5);
        fprintf(stderr, "vortex_commit_and_extract (n_rows=%zu, nc=%zu, scw=%zu):\n"
                        "  H2D+RS encode     %8.2f ms\n"
                        "  xfer setup        %8.2f ms\n"
                        "  SIS hash          %8.2f ms\n"
                        "  P2 MD hash        %8.2f ms\n"
                        "  Merkle+D2H+sync   %8.2f ms\n"
                        "  TOTAL             %8.2f ms\n",
                n_rows, nc, scw, ms01, ms12, ms23, ms34, ms45, ms05);
        cudaEventDestroy(t0); cudaEventDestroy(t1); cudaEventDestroy(t2);
        cudaEventDestroy(t3); cudaEventDestroy(t4); cudaEventDestroy(t5);
    }

    return KB_SUCCESS;
}

// Accessors for async extraction pinned host buffers.
extern "C" uint32_t *kb_vortex_h_enc_pinned(kb_vortex_pipeline_t p)    { return p ? p->h_enc_pinned    : nullptr; }
extern "C" uint32_t *kb_vortex_h_sis_pinned(kb_vortex_pipeline_t p)    { return p ? p->h_sis_pinned    : nullptr; }
extern "C" uint32_t *kb_vortex_h_leaves_pinned(kb_vortex_pipeline_t p) { return p ? p->h_leaves_pinned : nullptr; }

// ═════════════════════════════════════════════════════════════════════════════
// Symbolic expression evaluator — GPU bytecode VM
// ═════════════════════════════════════════════════════════════════════════════
//
// One thread per element, every thread executes the same bytecode → zero warp
// divergence.  Slots live in per-thread local memory (L1-cached).
//
//  thread i:
//    E4 slots[num_slots]
//    for pc in program:
//      OP_CONST   → slots[dst] = consts[ci]
//      OP_INPUT   → slots[dst] = read_input(inputs[id], i, n)
//      OP_MUL     → slots[dst] = Π slots[sₖ]^eₖ
//      OP_LINCOMB → slots[dst] = Σ cₖ · slots[sₖ]
//      OP_POLYEVAL→ slots[dst] = Horner(x, c₀..cₘ)
//    out[i] = slots[result_slot]

struct KBSymProgram {
    uint32_t *d_program;
    uint32_t *d_consts;
    uint32_t  pgm_len;
    uint32_t  num_consts;
    uint32_t  num_slots;
    uint32_t  result_slot;

    // Per-program reusable buffers, sized once on first call and reused on
    // every subsequent call. Eliminates two cudaMalloc + cudaFree pairs
    // per kb_sym_eval, which used to be a measurable fraction of the
    // quotient hot path (especially at small n where the launch + alloc
    // costs dominate the actual symbolic eval kernel).
    SymInputDesc *d_inputs_pool;
    size_t        d_inputs_capacity; // in elements
    uint32_t     *d_out_pool;
    size_t        d_out_capacity;    // in uint32_t (== 4 × n_elements_E4)
};

__device__ E4 sym_read_input(const SymInputDesc *desc, uint32_t i, uint32_t n) {
    switch (desc->tag) {
    case 0:  return e4_from_kb(desc->d_ptr[i]);                        // KB column
    case 1:  return E4{{desc->val[0], desc->val[1]},                   // E4 constant
                       {desc->val[2], desc->val[3]}};
    case 2: {uint32_t j = (i + desc->offset) % n;                     // rotated KB
             return e4_from_kb(desc->d_ptr[j]);}
    case 3: {const uint32_t *p = &desc->d_ptr[i * 4];                 // E4 vector
             return E4{{p[0], p[1]}, {p[2], p[3]}};}
    case 4: {const uint32_t *p = desc->d_ptr;                         // E4 vector (SoA)
             return E4{{p[i], p[n + i]}, {p[2 * n + i], p[3 * n + i]}};}
    case 5: {uint32_t j = (i + desc->offset) % n;                     // rotated E4 vector (SoA)
             const uint32_t *p = desc->d_ptr;
             return E4{{p[j], p[n + j]}, {p[2 * n + j], p[3 * n + j]}};}
    case 6: {uint32_t j = (i + desc->offset) % n;                     // rotated E4 vector (AoS)
             const uint32_t *p = &desc->d_ptr[j * 4];
             return E4{{p[0], p[1]}, {p[2], p[3]}};}
    default: return e4_zero();
    }
}

__global__ void kern_symbolic_eval(
    const uint32_t    *program,   uint32_t pgm_len,
    const uint32_t    *consts,    // E4 constants, 4 words each
    const SymInputDesc *inputs,
    uint32_t n,
    uint32_t result_slot,
    uint32_t *d_out)
{
    uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
    if (tid >= n) return;

    E4 slots[SYM_MAX_SLOTS];
    uint32_t pc = 0;

    while (pc < pgm_len) {
        uint32_t op  = program[pc++];
        uint32_t dst = program[pc++];

        switch (op) {
        case 0: { // OP_CONST
            uint32_t ci = program[pc++];
            const uint32_t *c = consts + ci * 4;
            slots[dst] = E4{{c[0], c[1]}, {c[2], c[3]}};
            break;
        }
        case 1: { // OP_INPUT
            uint32_t id = program[pc++];
            slots[dst] = sym_read_input(&inputs[id], tid, n);
            break;
        }
        case 2: { // OP_MUL: dst = Π slots[sₖ]^eₖ
            uint32_t nc = program[pc++];
            uint32_t s0 = program[pc++];
            uint32_t e0 = program[pc++];
            E4 acc = e4_pow(slots[s0], e0);
            for (uint32_t k = 1; k < nc; k++) {
                uint32_t s = program[pc++];
                uint32_t e = program[pc++];
                acc = e4_mul(acc, e4_pow(slots[s], e));
            }
            slots[dst] = acc;
            break;
        }
        case 3: { // OP_LINCOMB: dst = Σ cₖ · slots[sₖ]
            uint32_t nc = program[pc++];
            uint32_t s0 = program[pc++];
            int32_t  c0 = (int32_t)program[pc++];
            E4 acc = e4_scale_signed(slots[s0], c0);
            for (uint32_t k = 1; k < nc; k++) {
                uint32_t s = program[pc++];
                int32_t  c = (int32_t)program[pc++];
                acc = e4_add(acc, e4_scale_signed(slots[s], c));
            }
            slots[dst] = acc;
            break;
        }
        case 4: { // OP_POLYEVAL: Horner  children=[x, c₀, c₁, ..., cₘ]
            //   P(x) = c₀ + c₁x + c₂x² + ... + cₘxᵐ
            //   Horner: acc = cₘ; for k = m-1..0: acc = acc·x + cₖ
            uint32_t nc = program[pc++];
            E4 x   = slots[program[pc]];
            E4 acc  = slots[program[pc + nc - 1]];
            for (int k = (int)nc - 2; k >= 1; k--)
                acc = e4_add(e4_mul(acc, x), slots[program[pc + k]]);
            pc += nc;
            slots[dst] = acc;
            break;
        }
        } // switch
    } // while

    // Write result as flat E4
    d_out[tid * 4 + 0] = slots[result_slot].b0.a0;
    d_out[tid * 4 + 1] = slots[result_slot].b0.a1;
    d_out[tid * 4 + 2] = slots[result_slot].b1.a0;
    d_out[tid * 4 + 3] = slots[result_slot].b1.a1;
}

// ── Symbolic C ABI ──────────────────────────────────────────────────────────

extern "C" kb_error_t kb_sym_compile(gnark_gpu_context_t,
                                      const uint32_t *bytecode, uint32_t pgm_len,
                                      const uint32_t *constants, uint32_t num_consts,
                                      uint32_t num_slots,
                                      uint32_t result_slot,
                                      kb_sym_program_t *out) {
    if ((!bytecode && pgm_len) || num_slots > SYM_MAX_SLOTS)
        return KB_ERROR_INVALID;

    auto *p = new(std::nothrow) KBSymProgram{};
    if (!p) return KB_ERROR_OOM;

    p->pgm_len     = pgm_len;
    p->num_consts  = num_consts;
    p->num_slots   = num_slots;
    p->result_slot = result_slot;

    if (pgm_len > 0) {
        CUDA_CHECK(cudaMalloc(&p->d_program, pgm_len * sizeof(uint32_t)));
        CUDA_CHECK(cudaMemcpy(p->d_program, bytecode,
                              pgm_len * sizeof(uint32_t), cudaMemcpyHostToDevice));
    }
    if (num_consts > 0) {
        CUDA_CHECK(cudaMalloc(&p->d_consts, num_consts * 4 * sizeof(uint32_t)));
        CUDA_CHECK(cudaMemcpy(p->d_consts, constants,
                              num_consts * 4 * sizeof(uint32_t), cudaMemcpyHostToDevice));
    }
    *out = p;
    return KB_SUCCESS;
}

extern "C" void kb_sym_free(kb_sym_program_t p) {
    if (!p) return;
    cudaFree(p->d_program);
    cudaFree(p->d_consts);
    if (p->d_inputs_pool) cudaFree(p->d_inputs_pool);
    if (p->d_out_pool)    cudaFree(p->d_out_pool);
    delete p;
}

extern "C" kb_error_t kb_sym_eval(gnark_gpu_context_t,
                                   kb_sym_program_t program,
                                   const SymInputDesc *h_inputs, uint32_t num_inputs,
                                   uint32_t n,
                                   uint32_t *h_out) {
    if (!program || !h_out || n == 0) return KB_ERROR_INVALID;

    // Reuse the per-program input descriptor buffer; grow on demand.
    // Was: cudaMalloc + cudaMemcpy(sync) + cudaFree on every call.
    // Now: zero alloc on the steady state; D2H stays sync because the
    //      caller (gpu/quotient.RunGPU) needs the result before the
    //      host-side annulator scaling pass.
    if (num_inputs > 0) {
        if (program->d_inputs_capacity < num_inputs) {
            if (program->d_inputs_pool) cudaFree(program->d_inputs_pool);
            CUDA_CHECK(cudaMalloc(&program->d_inputs_pool,
                                  num_inputs * sizeof(SymInputDesc)));
            program->d_inputs_capacity = num_inputs;
        }
        CUDA_CHECK(cudaMemcpy(program->d_inputs_pool, h_inputs,
                              num_inputs * sizeof(SymInputDesc),
                              cudaMemcpyHostToDevice));
    }

    // Reuse the per-program output buffer; grow on demand.
    size_t out_count = (size_t)n * 4;
    if (program->d_out_capacity < out_count) {
        if (program->d_out_pool) cudaFree(program->d_out_pool);
        CUDA_CHECK(cudaMalloc(&program->d_out_pool,
                              out_count * sizeof(uint32_t)));
        program->d_out_capacity = out_count;
    }

    kern_symbolic_eval<<<kb_grid(n), KB_BLOCK>>>(
        program->d_program, program->pgm_len,
        program->d_consts,
        program->d_inputs_pool,
        n,
        program->result_slot,
        program->d_out_pool);

    // D2H of the eval result. Caller materializes it as []fext.E4 in Go
    // and runs the per-element annulator scaling on host. A future
    // optimization moves that scaling onto the GPU and D2Hs the smaller
    // base-field result instead — see worklog Step 4 outstanding items.
    CUDA_CHECK(cudaMemcpy(h_out, program->d_out_pool,
                          out_count * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));

    return KB_SUCCESS;
}

extern "C" uint32_t *kb_vec_device_ptr(kb_vec_t v) {
    return v ? v->d_data : nullptr;
}

// ── GPU Prove helpers ────────────────────────────────────────────────────────

// Linear combination on column-major encoded matrix:
//   UAlpha[j] = Σᵢ αⁱ · d_encoded_col[j × n_rows + i]
// Result is E4 vector of length scw, written to host.
extern "C" kb_error_t kb_vortex_lincomb(kb_vortex_pipeline_t p,
                                         size_t n_rows,
                                         const uint32_t alpha_raw[4],
                                         uint32_t *result) {
    if (!p || !alpha_raw || !result) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;

    size_t scw = p->size_codeword;
    E4 alpha = {{alpha_raw[0], alpha_raw[1]}, {alpha_raw[2], alpha_raw[3]}};

    uint32_t *d_result;
    CUDA_CHECK(cudaMalloc(&d_result, scw * 4 * sizeof(uint32_t)));

    kern_lincomb_e4_colmajor<<<kb_grid(scw), KB_BLOCK>>>(
        p->d_encoded_col, n_rows, scw, alpha, d_result);

    CUDA_CHECK(cudaMemcpy(result, d_result,
                          scw * 4 * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    cudaFree(d_result);
    return KB_SUCCESS;
}

// Extract a single column from the column-major encoded matrix to host.
//   out[i] = d_encoded_col[col_idx × n_rows + i],  i ∈ [0, n_rows)
extern "C" kb_error_t kb_vortex_extract_col(kb_vortex_pipeline_t p,
                                             size_t n_rows, int col_idx,
                                             uint32_t *out) {
    if (!p || !out) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;
    if (col_idx < 0 || (size_t)col_idx >= p->size_codeword) return KB_ERROR_INVALID;

    CUDA_CHECK(cudaMemcpy(out,
                          p->d_encoded_col + (size_t)col_idx * n_rows,
                          n_rows * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

// Extract full encoded matrix from GPU to host in column-major layout.
// out: [scw × n_rows] uint32, column-major: out[col * n_rows + row].
extern "C" kb_error_t kb_vortex_extract_all(kb_vortex_pipeline_t p,
                                             size_t n_rows, uint32_t *out) {
    if (!p || !out) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;

    size_t scw = p->size_codeword;
    CUDA_CHECK(cudaMemcpy(out, p->d_encoded_col,
                          scw * n_rows * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

// Transpose kernel: column-major [scw × n_rows] → row-major [n_rows × scw].
// One thread per element.
__global__ void kern_transpose_col_to_row(const uint32_t *__restrict__ col_major,
                                           uint32_t *__restrict__ row_major,
                                           size_t n_rows, size_t scw) {
    size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
    size_t total = n_rows * scw;
    if (idx >= total) return;
    size_t row = idx / scw;
    size_t col = idx % scw;
    row_major[idx] = col_major[col * n_rows + row];
}

// Extract full encoded matrix from GPU to host in row-major layout.
// out: [n_rows × scw] uint32, row-major: out[row * scw + col].
// Transposes on GPU before D2H to avoid costly CPU transposition.
extern "C" kb_error_t kb_vortex_extract_all_rowmajor(kb_vortex_pipeline_t p,
                                                      size_t n_rows,
                                                      uint32_t *out) {
    if (!p || !out) return KB_ERROR_INVALID;
    if (n_rows > p->max_n_rows) return KB_ERROR_SIZE;

    size_t scw = p->size_codeword;
    size_t total = scw * n_rows;

    // Allocate temp buffer on GPU for row-major result
    uint32_t *d_rowmajor = nullptr;
    CUDA_CHECK(cudaMalloc(&d_rowmajor, total * sizeof(uint32_t)));

    // Transpose on GPU
    kern_transpose_col_to_row<<<kb_grid(total), KB_BLOCK>>>(
        p->d_encoded_col, d_rowmajor, n_rows, scw);
    CUDA_CHECK(cudaGetLastError());

    // D2H the row-major buffer
    CUDA_CHECK(cudaMemcpy(out, d_rowmajor, total * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));

    cudaFree(d_rowmajor);
    return KB_SUCCESS;
}

// Get raw device pointer to column-major encoded matrix.
extern "C" uint32_t *kb_vortex_encoded_device_ptr(kb_vortex_pipeline_t p) {
    return p ? p->d_encoded_col : nullptr;
}

// Lincomb from a standalone column-major device buffer (not pipeline-bound).
//   result[j] = Σᵢ αⁱ · d_encoded[j * n_rows + i] ∈ E4,  j ∈ [0, scw)
extern "C" kb_error_t kb_lincomb_e4_colmajor(gnark_gpu_context_t ctx,
                                              const uint32_t *d_encoded_col,
                                              size_t n_rows, size_t scw,
                                              const uint32_t alpha_raw[4],
                                              uint32_t *result) {
    (void)ctx;
    if (!d_encoded_col || !alpha_raw || !result) return KB_ERROR_INVALID;
    if (n_rows == 0 || scw == 0) return KB_ERROR_SIZE;

    E4 alpha = {{alpha_raw[0], alpha_raw[1]}, {alpha_raw[2], alpha_raw[3]}};

    uint32_t *d_result;
    CUDA_CHECK(cudaMalloc(&d_result, scw * 4 * sizeof(uint32_t)));

    kern_lincomb_e4_colmajor<<<kb_grid(scw), KB_BLOCK>>>(
        d_encoded_col, n_rows, scw, alpha, d_result);

    CUDA_CHECK(cudaMemcpy(result, d_result,
                          scw * 4 * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    cudaFree(d_result);
    return KB_SUCCESS;
}

// Extract SIS column hashes from GPU to host.
//   out: flat [scw × degree] uint32, same layout as d_sis.
extern "C" kb_error_t kb_vortex_extract_sis(kb_vortex_pipeline_t p,
                                             size_t n_rows, uint32_t *out) {
    if (!p || !out) return KB_ERROR_INVALID;
    (void)n_rows;  // SIS hashes are per-column, not per-row
    size_t scw = p->size_codeword;
    size_t deg = (size_t)p->degree;
    CUDA_CHECK(cudaMemcpy(out, p->d_sis,
                          scw * deg * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

// Extract leaf hashes (Poseidon2 digests) from GPU to host.
//   out: flat [scw × 8] uint32, same layout as d_leaves.
extern "C" kb_error_t kb_vortex_extract_leaves(kb_vortex_pipeline_t p,
                                                uint32_t *out) {
    if (!p || !out) return KB_ERROR_INVALID;
    size_t scw = p->size_codeword;
    CUDA_CHECK(cudaMemcpy(out, p->d_leaves,
                          scw * 8 * sizeof(uint32_t),
                          cudaMemcpyDeviceToHost));
    return KB_SUCCESS;
}

// Return sizeCodeWord for the pipeline.
extern "C" size_t kb_vortex_scw(kb_vortex_pipeline_t p) {
    return p ? p->size_codeword : 0;
}

// Return degree (SIS polynomial degree) for the pipeline.
extern "C" int kb_vortex_degree(kb_vortex_pipeline_t p) {
    return p ? p->degree : 0;
}
