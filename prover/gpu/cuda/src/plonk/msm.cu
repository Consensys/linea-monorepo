// ═══════════════════════════════════════════════════════════════════════════════
// MSM (Multi-Scalar Multiplication) for BLS12-377 G1
//
// Computes:  Q = Σᵢ sᵢ · Pᵢ   for n scalar-point pairs (sᵢ, Pᵢ)
//
// Algorithm: Pippenger's bucket method with signed-digit decomposition
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │                     Pippenger's Method Overview                        │
// │                                                                        │
// │  Each 253-bit scalar sᵢ is decomposed into w windows of c bits:       │
// │    sᵢ = Σⱼ dᵢⱼ · 2^(j·c)   where dᵢⱼ ∈ {-2^(c-1), ..., 2^(c-1)}  │
// │                                                                        │
// │  Signed digits halve the bucket count: 2^(c-1) buckets per window.    │
// │  When dᵢⱼ < 0, we negate the point and use bucket |dᵢⱼ|.            │
// │                                                                        │
// │  For each window j, the bucket sum is:                                 │
// │    Wⱼ = Σ_b  b · (Σ {Pᵢ : |dᵢⱼ| = b})                             │
// │                                                                        │
// │  Final result via Horner's rule:                                       │
// │    Q = (...((W[w-1])·2^c + W[w-2])·2^c + ...)·2^c + W[0]           │
// └─────────────────────────────────────────────────────────────────────────┘
//
// GPU Pipeline: build_pairs → radix sort → boundaries → accumulate → reduce
// Host: Horner combination in TE coordinates, single TE→Jacobian at end.
// ═══════════════════════════════════════════════════════════════════════════════

#include "ec.cuh"
#include <cub/device/device_radix_sort.cuh>
#include <cuda_runtime.h>
#include <cstdlib>

namespace gnark_gpu {

// =============================================================================
// MSM configuration
// =============================================================================

static constexpr int MSM_SCALAR_BITS = 253;
static constexpr int ACCUM_PARALLEL_THREADS = 128;
static constexpr int REDUCE_THREADS_PER_WINDOW = 128;
static constexpr int FINALIZE_THREADS = 32;

// Two-phase bucket accumulation cap.
// Phase 1 (sequential): each thread processes at most CAP entries per bucket.
// Phase 2 (parallel):   128 threads/block handle any remaining entries.
// For uniform scalars (avg bucket ~70), phase 1 handles everything.
// For concentrated scalars (huge buckets), phase 2 distributes the tail work.
static constexpr int ACCUM_SEQ_CAP = 256;

static int forced_window_bits() {
	static int forced_c = -1;
	if(forced_c != -1) return forced_c;

	forced_c = 0;
	const char *env = std::getenv("GNARK_GPU_MSM_FORCE_C");
	if(!env || !*env) return forced_c;

	const int parsed = std::atoi(env);
	// Keep c within a safe range for 32-bit bucket math and 253-bit scalars.
	if(parsed >= 1 && parsed <= 23) forced_c = parsed;
	return forced_c;
}

static size_t host_register_threshold_points() {
	static int threshold = -1;
	if(threshold != -1) return (size_t)threshold;

	threshold = 1 << 20;
	const char *env = std::getenv("GNARK_GPU_MSM_REGISTER_THRESHOLD");
	if(!env || !*env) return (size_t)threshold;

	const int parsed = std::atoi(env);
	if(parsed >= 0) threshold = parsed;
	return (size_t)threshold;
}

static int overflow_compaction_mode() {
	static int mode = -2;
	if(mode != -2) return mode;

	mode = -1;
	const char *env = std::getenv("GNARK_GPU_MSM_COMPACT_OVERFLOWS");
	if(env && *env) mode = std::atoi(env) != 0 ? 1 : 0;
	return mode;
}

static bool compact_overflow_buckets(size_t n) {
	int mode = overflow_compaction_mode();
	if(mode != -1) return mode != 0;
	return n <= (1u << 23);
}

// Window-size schedule for BLS12-377 signed-digit MSM.
// Empirical outcome on real gnark scalar datasets:
//   - c=13 is best for tiny sizes,
//   - c=15 is best for small-mid,
//   - c=17 is consistently best from ~2^19 upward.
//
// We intentionally avoid c=19/c=20 defaults: they can look good on synthetic
// random inputs but regress badly on large real runs due bucket skew and
// higher reduction overhead.
static int compute_optimal_c(size_t n) {
	const int forced_c = forced_window_bits();
	if(forced_c != 0) return forced_c;

	if(n <= (1 << 14)) return 13;
	if(n <= (1 << 18)) return 15;
	return 17;
}

// =============================================================================
// Kernel 1: Build (bucket_id, point_idx) pairs — signed-digit decomposition
//
// For each window, extract c bits + carry, apply signed-digit reduction:
//   digit > 2^(c-1) → negate: digit = 2^c - digit, sign = 1, carry = 1
//   digit ≤ 2^(c-1) → positive: sign = 0, carry = 0
//   digit == 0 → sentinel (skip bucket assignment)
//
// Scalars are decomposed in Montgomery form; host corrects by R^{-1}.
// =============================================================================

__global__ void __launch_bounds__(256) build_pairs_kernel(
	const uint64_t *__restrict__ scalars,
	uint32_t *__restrict__ keys,
	uint32_t *__restrict__ vals,
	size_t n, int c, int num_windows, int num_buckets, int total_buckets,
	size_t point_offset) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t s[4];
	s[0] = scalars[idx * 4 + 0];
	s[1] = scalars[idx * 4 + 1];
	s[2] = scalars[idx * 4 + 2];
	s[3] = scalars[idx * 4 + 3];

	uint32_t c_mask = (1u << c) - 1;
	uint32_t carry = 0;
	const uint32_t point_base = (uint32_t)(idx + point_offset) & 0x7FFFFFFFu;

	for(int w = 0; w < num_windows; w++) {
		int bit_offset = w * c;
		int limb_idx = bit_offset / 64;
		int bit_shift = bit_offset % 64;

		uint32_t digit;
		if(limb_idx >= 4) {
			digit = 0;
		} else {
			digit = (uint32_t)(s[limb_idx] >> bit_shift);
			if(bit_shift + c > 64 && limb_idx + 1 < 4)
				digit |= (uint32_t)(s[limb_idx + 1] << (64 - bit_shift));
		}
		size_t out_idx = (size_t)idx * num_windows + w;
		digit = (digit & c_mask) + carry;

		carry = (digit > (uint32_t)num_buckets) ? 1u : 0u;
		uint32_t neg_digit = (1u << c) - digit;
		uint32_t use_neg = carry;
		uint32_t bucket = use_neg ? neg_digit : digit;
		uint32_t sign = use_neg;

		// Handle edge case: 2^c - digit == 0 when digit == 2^c (carry overflow)
		uint32_t is_overflow = (bucket == 0 && use_neg) ? 1u : 0u;
		carry |= is_overflow;
		sign &= ~is_overflow;

		keys[out_idx] = (bucket == 0) ? (uint32_t)total_buckets
									   : (uint32_t)(w * num_buckets + (bucket - 1));
		// Store absolute point index (chunk-relative idx + point_offset)
		vals[out_idx] = point_base | (sign << 31);
	}
}

// =============================================================================
// Kernel 2: Detect bucket boundaries in sorted key array
// =============================================================================

__global__ void __launch_bounds__(256) detect_bucket_boundaries_kernel(
	const uint32_t *__restrict__ sorted_keys,
	uint32_t *__restrict__ bucket_offsets,
	uint32_t *__restrict__ bucket_ends,
	size_t assignments, int total_buckets) {

	size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(i >= assignments) return;

	uint32_t key = sorted_keys[i];
	if(key >= (uint32_t)total_buckets) return;

	if(i == 0 || sorted_keys[i - 1] != key) bucket_offsets[key] = (uint32_t)i;
	if(i == assignments - 1 || sorted_keys[i + 1] != key) bucket_ends[key] = (uint32_t)(i + 1);
}

// =============================================================================
// Kernel 3: Accumulate points per bucket (sequential, one thread per bucket)
// =============================================================================

__global__ void __launch_bounds__(256, 2)
	accumulate_buckets_kernel(
		const G1EdXY *__restrict__ points,
		const uint32_t *__restrict__ point_indices,
		const uint32_t *__restrict__ bucket_offsets,
		const uint32_t *__restrict__ bucket_ends,
		G1EdExtended *__restrict__ buckets,
		int total_buckets,
		bool add_to_existing,
		int cap,
		uint32_t *__restrict__ overflow_buckets,
		uint32_t *__restrict__ overflow_count) {

	int bucket_flat = blockIdx.x * blockDim.x + threadIdx.x;
	if(bucket_flat >= total_buckets) return;

	uint32_t start = bucket_offsets[bucket_flat];
	uint32_t full_end = bucket_ends[bucket_flat];
	uint32_t end = full_end;

	// Cap: process at most `cap` entries per bucket (0 = unlimited).
	// Phase 2 (parallel kernel) handles any remainder.
	if(cap > 0 && full_end > start + (uint32_t)cap) {
		end = start + (uint32_t)cap;
		if(overflow_buckets && overflow_count) {
			uint32_t slot = atomicAdd(overflow_count, 1u);
			overflow_buckets[slot] = (uint32_t)bucket_flat;
		}
	}

	G1EdExtended acc;
	if(add_to_existing)
		acc = buckets[bucket_flat];
	else
		ec_te_set_identity(acc);

	for(uint32_t i = start; i < end; i++) {
		uint32_t packed = point_indices[i];
		G1EdXY pt = points[packed & 0x7FFFFFFFu];
		ec_te_cnegate_xy(pt, (bool)(packed >> 31));
		ec_te_unified_mixed_add_xy(acc, pt);
	}

	buckets[bucket_flat] = acc;
}

// =============================================================================
// Kernel 3b: Parallel bucket accumulation (one block per bucket, tree reduce)
// =============================================================================

__global__ void __launch_bounds__(128, 4)
	accumulate_buckets_parallel_kernel(
		const G1EdXY *__restrict__ points,
		const uint32_t *__restrict__ point_indices,
		const uint32_t *__restrict__ bucket_offsets,
		const uint32_t *__restrict__ bucket_ends,
		const uint32_t *__restrict__ overflow_buckets,
		G1EdExtended *__restrict__ buckets,
		bool add_to_existing,
		uint32_t start_offset) {

	int bucket_flat = overflow_buckets ? overflow_buckets[blockIdx.x] : blockIdx.x;
	int tid = threadIdx.x;
	uint32_t start = bucket_offsets[bucket_flat] + start_offset;
	uint32_t end = bucket_ends[bucket_flat];

	if(start >= end) {
		return;
	}

	G1EdExtended acc;
	ec_te_set_identity(acc);
	for(uint32_t i = start + tid; i < end; i += ACCUM_PARALLEL_THREADS) {
		uint32_t packed = point_indices[i];
		G1EdXY pt = points[packed & 0x7FFFFFFFu];
		ec_te_cnegate_xy(pt, (bool)(packed >> 31));
		ec_te_unified_mixed_add_xy(acc, pt);
	}

	extern __shared__ G1EdExtended shared[];
	shared[tid] = acc;
	__syncthreads();
	for(int stride = ACCUM_PARALLEL_THREADS / 2; stride > 0; stride >>= 1) {
		if(tid < stride) ec_te_unified_add(shared[tid], shared[tid + stride]);
		__syncthreads();
	}
	if(tid == 0) {
		if(add_to_existing)
			ec_te_unified_add(buckets[bucket_flat], shared[0]);
		else
			buckets[bucket_flat] = shared[0];
	}
}

// =============================================================================
// Small TE scalar multiply (double-and-add, for reduce corrections)
// =============================================================================

__device__ __forceinline__ void ec_te_mul_small(G1EdExtended &out, const G1EdExtended &in, int k) {
	ec_te_set_identity(out);
	if(k <= 0) return;
	G1EdExtended base = in;
	while(k > 0) {
		if(k & 1) ec_te_unified_add(out, base);
		k >>= 1;
		if(k > 0) ec_te_unified_add(base, base);
	}
}

// =============================================================================
// Kernel 4a: Partial reduce — running sum trick per block-range
//
//   For b = high down to low:  S += B[b]; Total += S
//   Result: Total = Σ (b+1)·B[b]  (weights = bucket digit = b+1)
// =============================================================================

__global__ void __launch_bounds__(256, 2) reduce_buckets_partial_kernel(
	const G1EdExtended *__restrict__ buckets,
	G1EdExtended *__restrict__ partial_totals,
	G1EdExtended *__restrict__ partial_sums,
	int num_windows, int num_buckets, int blocks_per_window) {

	int block_flat = blockIdx.x;
	int w = block_flat / blocks_per_window;
	int part = block_flat % blocks_per_window;
	if(w >= num_windows) return;

	int tid = threadIdx.x;
	int P = blockDim.x;

	int range_size = (num_buckets + blocks_per_window - 1) / blocks_per_window;
	int high = num_buckets - 1 - part * range_size;
	if(high < 0) {
		if(tid == 0) {
			int out_idx = w * blocks_per_window + part;
			ec_te_set_identity(partial_totals[out_idx]);
			ec_te_set_identity(partial_sums[out_idx]);
		}
		return;
	}
	int low = high - range_size + 1;
	if(low < 0) low = 0;
	int range_len = high - low + 1;

	int chunk_size = (range_len + P - 1) / P;
	int chunk_high = high - tid * chunk_size;
	int chunk_low = chunk_high - chunk_size + 1;
	if(chunk_low < low) chunk_low = low;
	if(chunk_high > high) chunk_high = high;
	bool has_work = (chunk_high >= low);

	G1EdExtended local_running, local_total;
	ec_te_set_identity(local_running);
	ec_te_set_identity(local_total);
	int local_len = 0;

	if(has_work) {
		for(int b = chunk_high; b >= chunk_low; b--) {
			ec_te_unified_add(local_running, buckets[w * num_buckets + b]);
			ec_te_unified_add(local_total, local_running);
			local_len++;
		}
	}

	// Hillis-Steele inclusive prefix scan of running sums
	__shared__ G1EdExtended shared_prefix[REDUCE_THREADS_PER_WINDOW];
	shared_prefix[tid] = local_running;
	__syncthreads();
	for(int d = 1; d < P; d <<= 1) {
		G1EdExtended tmp;
		bool do_add = (tid >= d);
		if(do_add) tmp = shared_prefix[tid - d];
		__syncthreads();
		if(do_add) ec_te_unified_add(shared_prefix[tid], tmp);
		__syncthreads();
	}

	if(tid == 0)
		partial_sums[w * blocks_per_window + part] = shared_prefix[P - 1];

	// Convert to exclusive prefix
	G1EdExtended my_exclusive;
	if(tid == 0) ec_te_set_identity(my_exclusive);
	else         my_exclusive = shared_prefix[tid - 1];
	__syncthreads();
	shared_prefix[tid] = my_exclusive;
	__syncthreads();

	G1EdExtended correction;
	ec_te_mul_small(correction, shared_prefix[tid], local_len);
	ec_te_unified_add(local_total, correction);

	// Tree reduction of corrected totals
	shared_prefix[tid] = local_total;
	__syncthreads();
	for(int stride = P / 2; stride > 0; stride >>= 1) {
		if(tid < stride) ec_te_unified_add(shared_prefix[tid], shared_prefix[tid + stride]);
		__syncthreads();
	}
	if(tid == 0)
		partial_totals[w * blocks_per_window + part] = shared_prefix[0];
}

// =============================================================================
// Kernel 4b: Finalize — combine partial ranges into one result per window
// =============================================================================

__global__ void reduce_buckets_finalize_kernel(
	const G1EdExtended *__restrict__ partial_totals,
	const G1EdExtended *__restrict__ partial_sums,
	G1EdExtended *__restrict__ window_results,
	int num_windows, int num_buckets, int blocks_per_window) {

	int w = blockIdx.x;
	if(w >= num_windows) return;
	int tid = threadIdx.x;
	extern __shared__ G1EdExtended smem[];

	int range_size = (num_buckets + blocks_per_window - 1) / blocks_per_window;

	G1EdExtended my_total, my_sum;
	int my_len = 0;
	if(tid < blocks_per_window) {
		int high = num_buckets - 1 - tid * range_size;
		if(high >= 0) {
			int low = high - range_size + 1;
			if(low < 0) low = 0;
			my_len = high - low + 1;
			my_total = partial_totals[w * blocks_per_window + tid];
			my_sum = partial_sums[w * blocks_per_window + tid];
		} else { ec_te_set_identity(my_total); ec_te_set_identity(my_sum); }
	} else { ec_te_set_identity(my_total); ec_te_set_identity(my_sum); }

	// Exclusive prefix scan of partial_sums
	smem[tid] = my_sum;
	__syncthreads();
	for(int d = 1; d < FINALIZE_THREADS; d <<= 1) {
		G1EdExtended tmp;
		bool do_add = (tid >= d && tid < blocks_per_window);
		if(do_add) tmp = smem[tid - d];
		__syncthreads();
		if(do_add) ec_te_unified_add(smem[tid], tmp);
		__syncthreads();
	}
	G1EdExtended my_exclusive;
	if(tid == 0)                      ec_te_set_identity(my_exclusive);
	else if(tid < blocks_per_window)  my_exclusive = smem[tid - 1];
	else                              ec_te_set_identity(my_exclusive);
	__syncthreads();
	smem[tid] = my_exclusive;
	__syncthreads();

	if(tid < blocks_per_window && my_len > 0) {
		G1EdExtended correction;
		ec_te_mul_small(correction, smem[tid], my_len);
		ec_te_unified_add(my_total, correction);
	}

	smem[tid] = my_total;
	__syncthreads();
	for(int stride = FINALIZE_THREADS / 2; stride > 0; stride >>= 1) {
		if(tid < stride) ec_te_unified_add(smem[tid], smem[tid + stride]);
		__syncthreads();
	}
	if(tid == 0) {
		G1EdExtended result = smem[0];
		ec_te_reduce(result);
		window_results[w] = result;
	}
}

// =============================================================================
// MSM context
// =============================================================================

struct MSMContext {
	size_t max_points;
	int c, num_windows, num_buckets, sort_key_bits, reduce_blocks_per_window;

	G1EdXY *d_points;
	uint64_t *d_scalars;
	uint32_t *d_bucket_offsets, *d_bucket_ends, *d_point_indices;
	G1EdExtended *d_buckets, *d_window_results, *d_window_accum;
	G1EdExtended *d_window_partial_totals, *d_window_partial_sums;
	uint32_t *d_keys_in, *d_keys_out, *d_vals_in;
	uint32_t *d_overflow_buckets, *d_overflow_count;
	void *d_sort_temp;
	size_t sort_temp_bytes;

	// Double-buffered pinned staging for overlapped CPU memcpy + GPU DMA
	void *h_scalar_staging;    // pinned buffer A
	void *h_scalar_staging_b;  // pinned buffer B
	size_t staging_buf_bytes;  // per-buffer size in bytes (0 if alloc failed)

	// Optional persistent registration of caller scalar memory.
	const void *registered_host_ptr;
	size_t registered_host_bytes;
	bool host_registered;
};

// ── Lazy work buffer management ──
//
// Sort buffers (d_keys_in/out, d_vals_in, d_point_indices, d_sort_temp,
// d_scalars) dominate MSM VRAM at large n (~49 GiB at n=2^27). They are
// only needed during msm_run_full, so we allocate them lazily before each
// run and free them immediately after. This allows the quotient phase to
// reclaim all that VRAM for working vectors + selector uploads.
//
// At n=2^27 with c=17, lazy alloc/free adds ~5-10ms per MSM call (negligible
// vs 200-1700ms compute). The permanent allocations (points, buckets, window
// results) stay resident.

cudaError_t msm_alloc_work_buffers(MSMContext *ctx) {
	if(ctx->d_keys_in) return cudaSuccess; // already allocated
	size_t max_assignments = ctx->max_points * (size_t)ctx->num_windows;
	cudaGetLastError(); // clear any prior error
	cudaMalloc(&ctx->d_scalars, ctx->max_points * 4 * sizeof(uint64_t));
	cudaMalloc(&ctx->d_keys_in, max_assignments * sizeof(uint32_t));
	cudaMalloc(&ctx->d_keys_out, max_assignments * sizeof(uint32_t));
	cudaMalloc(&ctx->d_vals_in, max_assignments * sizeof(uint32_t));
	cudaMalloc(&ctx->d_point_indices, max_assignments * sizeof(uint32_t));
	cudaMalloc(&ctx->d_sort_temp, ctx->sort_temp_bytes);
	cudaError_t err = cudaGetLastError();
	if(err != cudaSuccess) {
		// Allocation failed — free whatever was partially allocated and clear error.
		auto safe_free = [](auto &p) { if(p) { cudaFree(p); p = nullptr; } };
		safe_free(ctx->d_scalars);
		safe_free(ctx->d_keys_in); safe_free(ctx->d_keys_out);
		safe_free(ctx->d_vals_in); safe_free(ctx->d_point_indices);
		safe_free(ctx->d_sort_temp);
		cudaGetLastError();
	}
	return err;
}

void msm_free_work_buffers(MSMContext *ctx) {
	auto free = [](auto &p) { if(p) { cudaFree(p); p = nullptr; } };
	free(ctx->d_scalars);
	free(ctx->d_keys_in); free(ctx->d_keys_out);
	free(ctx->d_vals_in); free(ctx->d_point_indices);
	free(ctx->d_sort_temp);
}

MSMContext *msm_create(size_t max_points) {
	cudaGetLastError();

	MSMContext *ctx = new MSMContext;
	memset(ctx, 0, sizeof(MSMContext));
	ctx->max_points = max_points;
	ctx->c = compute_optimal_c(max_points);
	ctx->num_windows = (MSM_SCALAR_BITS + ctx->c - 1) / ctx->c;
	ctx->num_buckets = 1 << (ctx->c - 1);   // signed digits halve bucket count

	int total_buckets = ctx->num_windows * ctx->num_buckets;
	size_t max_assignments = max_points * (size_t)ctx->num_windows;

	int key_val = total_buckets;
	ctx->sort_key_bits = 1;
	while((1 << ctx->sort_key_bits) <= key_val) ctx->sort_key_bits++;

	{
		int max_bpw = ctx->num_buckets / REDUCE_THREADS_PER_WINDOW;
		int target_bpw = 752 / ctx->num_windows;
		ctx->reduce_blocks_per_window = max_bpw < target_bpw ? max_bpw : target_bpw;
		if(ctx->reduce_blocks_per_window < 1) ctx->reduce_blocks_per_window = 1;
		if(ctx->reduce_blocks_per_window > FINALIZE_THREADS) ctx->reduce_blocks_per_window = FINALIZE_THREADS;
	}

	int total_partials = ctx->num_windows * ctx->reduce_blocks_per_window;

	// Permanent small allocations (points, buckets, window results).
	// Sort buffers are allocated lazily in msm_run_full.
	cudaMalloc(&ctx->d_points, max_points * sizeof(G1EdXY));
	cudaMalloc(&ctx->d_bucket_offsets, total_buckets * sizeof(uint32_t));
	cudaMalloc(&ctx->d_bucket_ends, total_buckets * sizeof(uint32_t));
	cudaMalloc(&ctx->d_buckets, total_buckets * sizeof(G1EdExtended));
	cudaMalloc(&ctx->d_window_results, ctx->num_windows * sizeof(G1EdExtended));
	cudaMalloc(&ctx->d_window_accum, ctx->num_windows * sizeof(G1EdExtended));
	cudaMalloc(&ctx->d_window_partial_totals, total_partials * sizeof(G1EdExtended));
	cudaMalloc(&ctx->d_window_partial_sums, total_partials * sizeof(G1EdExtended));
	cudaMalloc(&ctx->d_overflow_buckets, total_buckets * sizeof(uint32_t));
	cudaMalloc(&ctx->d_overflow_count, sizeof(uint32_t));

	// Query CUB sort temp size (no allocation, just the size query).
	ctx->sort_temp_bytes = 0;
	cub::DeviceRadixSort::SortPairs(nullptr, ctx->sort_temp_bytes,
		(uint32_t *)nullptr, (uint32_t *)nullptr, (uint32_t *)nullptr,
		(uint32_t *)nullptr, max_assignments, 0, ctx->sort_key_bits);

	// Also check chunk-sized sort temp (CUB may need more for smaller inputs).
	static constexpr size_t STAGING_CAP = 256ULL * 1024 * 1024;
	size_t total_scalar_bytes = max_points * 4 * sizeof(uint64_t);
	size_t per_buf = total_scalar_bytes / 2;
	if(per_buf > STAGING_CAP) per_buf = STAGING_CAP;
	if(per_buf < 32) per_buf = 32;
	ctx->staging_buf_bytes = per_buf;

	size_t chunk_size = per_buf / (4 * sizeof(uint64_t));
	if(chunk_size > 0 && chunk_size < max_points) {
		size_t chunk_assignments = chunk_size * (size_t)ctx->num_windows;
		size_t chunk_sort_temp = 0;
		cub::DeviceRadixSort::SortPairs(nullptr, chunk_sort_temp,
			(uint32_t *)nullptr, (uint32_t *)nullptr, (uint32_t *)nullptr,
			(uint32_t *)nullptr, chunk_assignments, 0, ctx->sort_key_bits);
		if(chunk_sort_temp > ctx->sort_temp_bytes)
			ctx->sort_temp_bytes = chunk_sort_temp;
	}

	// Double-buffered pinned staging for scalar upload.
	cudaError_t err_a = cudaHostAlloc(&ctx->h_scalar_staging, per_buf, cudaHostAllocDefault);
	cudaError_t err_b = cudaHostAlloc(&ctx->h_scalar_staging_b, per_buf, cudaHostAllocDefault);
	if(err_a != cudaSuccess || err_b != cudaSuccess) {
		if(ctx->h_scalar_staging) { cudaFreeHost(ctx->h_scalar_staging); ctx->h_scalar_staging = nullptr; }
		if(ctx->h_scalar_staging_b) { cudaFreeHost(ctx->h_scalar_staging_b); ctx->h_scalar_staging_b = nullptr; }
		ctx->staging_buf_bytes = 0;
	}

	return ctx;
}

void msm_destroy(MSMContext *ctx) {
	if(!ctx) return;
	cudaFree(ctx->d_points); cudaFree(ctx->d_scalars);
	cudaFree(ctx->d_bucket_offsets); cudaFree(ctx->d_bucket_ends);
	cudaFree(ctx->d_point_indices); cudaFree(ctx->d_buckets);
	cudaFree(ctx->d_window_results); cudaFree(ctx->d_window_accum);
	cudaFree(ctx->d_window_partial_totals); cudaFree(ctx->d_window_partial_sums);
	cudaFree(ctx->d_keys_in); cudaFree(ctx->d_keys_out); cudaFree(ctx->d_vals_in);
	cudaFree(ctx->d_overflow_buckets); cudaFree(ctx->d_overflow_count);
	cudaFree(ctx->d_sort_temp);
	if(ctx->h_scalar_staging) cudaFreeHost(ctx->h_scalar_staging);
	if(ctx->h_scalar_staging_b) cudaFreeHost(ctx->h_scalar_staging_b);
	if(ctx->host_registered && ctx->registered_host_ptr) cudaHostUnregister((void *)ctx->registered_host_ptr);
	delete ctx;
}

void msm_load_points(MSMContext *ctx, const void *host_points, size_t count, cudaStream_t stream) {
	cudaMemcpyAsync(ctx->d_points, host_points, count * sizeof(G1EdXY), cudaMemcpyHostToDevice, stream);
}
void msm_offload_points(MSMContext *ctx) {
	if(ctx->d_points) { cudaFree(ctx->d_points); ctx->d_points = nullptr; }
}
void msm_unregister_host(MSMContext *ctx) {
	if(ctx->host_registered && ctx->registered_host_ptr) {
		cudaHostUnregister((void *)ctx->registered_host_ptr);
		cudaGetLastError(); // clear non-sticky error from failed unregister
		ctx->host_registered = false;
		ctx->registered_host_ptr = nullptr;
		ctx->registered_host_bytes = 0;
	}
}
cudaError_t msm_reload_points(MSMContext *ctx, const void *host_points, size_t count, cudaStream_t stream) {
	cudaError_t err = cudaMalloc(&ctx->d_points, count * sizeof(G1EdXY));
	if(err != cudaSuccess) return err;
	cudaMemcpyAsync(ctx->d_points, host_points, count * sizeof(G1EdXY), cudaMemcpyHostToDevice, stream);
	return cudaSuccess;
}
void msm_upload_scalars(MSMContext *ctx, const uint64_t *host_scalars, size_t n, cudaStream_t stream) {
	size_t bytes = n * 4 * sizeof(uint64_t);
	if(ctx->h_scalar_staging && ctx->h_scalar_staging_b && bytes <= 2 * ctx->staging_buf_bytes) {
		// Double-buffered: overlap CPU memcpy with GPU DMA.
		size_t half = bytes / 2;
		const char *src = (const char *)host_scalars;

		memcpy(ctx->h_scalar_staging, src, half);
		cudaMemcpyAsync(ctx->d_scalars, ctx->h_scalar_staging, half, cudaMemcpyHostToDevice, stream);
		memcpy(ctx->h_scalar_staging_b, src + half, bytes - half);
		cudaMemcpyAsync((char *)ctx->d_scalars + half, ctx->h_scalar_staging_b, bytes - half, cudaMemcpyHostToDevice, stream);
	} else {
		cudaMemcpyAsync(ctx->d_scalars, host_scalars, bytes, cudaMemcpyHostToDevice, stream);
	}
}

void launch_msm(MSMContext *ctx, size_t n, cudaStream_t stream) {
	constexpr unsigned threads = 256;
	int total_buckets = ctx->num_windows * ctx->num_buckets;
	size_t assignments = n * (size_t)ctx->num_windows;

	// Build pairs
	{
		unsigned blocks = ((unsigned)n + threads - 1) / threads;
		build_pairs_kernel<<<blocks, threads, 0, stream>>>(
			ctx->d_scalars, ctx->d_keys_in, ctx->d_vals_in, n,
			ctx->c, ctx->num_windows, ctx->num_buckets, total_buckets, 0);
	}

	// Radix sort
	size_t sort_bytes = ctx->sort_temp_bytes;
	cub::DeviceRadixSort::SortPairs(ctx->d_sort_temp, sort_bytes,
		ctx->d_keys_in, ctx->d_keys_out, ctx->d_vals_in, ctx->d_point_indices,
		assignments, 0, ctx->sort_key_bits, stream);

	cudaMemsetAsync(ctx->d_bucket_offsets, 0, total_buckets * sizeof(uint32_t), stream);
	cudaMemsetAsync(ctx->d_bucket_ends, 0, total_buckets * sizeof(uint32_t), stream);

	// Detect boundaries
	{
		unsigned blocks = (unsigned)((assignments + threads - 1) / threads);
		detect_bucket_boundaries_kernel<<<blocks, threads, 0, stream>>>(
			ctx->d_keys_out, ctx->d_bucket_offsets, ctx->d_bucket_ends,
			assignments, total_buckets);
	}

	// Accumulate (two-phase: sequential with cap, then parallel for overflow)
	//
	// Dynamic cap: max(ACCUM_SEQ_CAP, 2·avg + 64). For uniform scalars, the
	// cap exceeds the max bucket size (Poisson tail), so phase 1 handles
	// everything. For concentrated scalars (bucket size >> cap), phase 2
	// distributes the tail across 128 threads.
	{
		size_t avg = assignments / (size_t)total_buckets;
		int cap = (int)(2 * avg + 64);
		if(cap < ACCUM_SEQ_CAP) cap = ACCUM_SEQ_CAP;
		if(cap > 4096) cap = 4096;

		unsigned seq_blocks = ((unsigned)total_buckets + threads - 1) / threads;

		if(compact_overflow_buckets(n)) {
			// Phase 1: Sequential — each thread handles min(bucket_size, cap).
			cudaMemsetAsync(ctx->d_overflow_count, 0, sizeof(uint32_t), stream);
			accumulate_buckets_kernel<<<seq_blocks, threads, 0, stream>>>(
				ctx->d_points, ctx->d_point_indices,
				ctx->d_bucket_offsets, ctx->d_bucket_ends, ctx->d_buckets,
				total_buckets, false, cap, ctx->d_overflow_buckets, ctx->d_overflow_count);

			// Phase 2: Parallel tree-reduce only buckets that exceeded the cap.
			// Random proving scalars normally produce no overflow buckets; compacting
			// avoids launching one empty 128-thread block for every bucket.
			uint32_t overflow_count = 0;
			cudaMemcpyAsync(&overflow_count, ctx->d_overflow_count, sizeof(uint32_t),
			                cudaMemcpyDeviceToHost, stream);
			cudaStreamSynchronize(stream);
			if(overflow_count > 0) {
				size_t smem = ACCUM_PARALLEL_THREADS * sizeof(G1EdExtended);
				accumulate_buckets_parallel_kernel<<<overflow_count, ACCUM_PARALLEL_THREADS, smem, stream>>>(
					ctx->d_points, ctx->d_point_indices,
					ctx->d_bucket_offsets, ctx->d_bucket_ends, ctx->d_overflow_buckets,
					ctx->d_buckets, true, (uint32_t)cap);
			}
		} else {
			accumulate_buckets_kernel<<<seq_blocks, threads, 0, stream>>>(
				ctx->d_points, ctx->d_point_indices,
				ctx->d_bucket_offsets, ctx->d_bucket_ends, ctx->d_buckets,
				total_buckets, false, cap, nullptr, nullptr);

			size_t smem = ACCUM_PARALLEL_THREADS * sizeof(G1EdExtended);
			accumulate_buckets_parallel_kernel<<<total_buckets, ACCUM_PARALLEL_THREADS, smem, stream>>>(
				ctx->d_points, ctx->d_point_indices,
				ctx->d_bucket_offsets, ctx->d_bucket_ends, nullptr,
				ctx->d_buckets, true, (uint32_t)cap);
		}
	}

	// Reduce
	{
		int bpw = ctx->reduce_blocks_per_window;
		reduce_buckets_partial_kernel<<<ctx->num_windows * bpw, REDUCE_THREADS_PER_WINDOW, 0, stream>>>(
			ctx->d_buckets, ctx->d_window_partial_totals, ctx->d_window_partial_sums,
			ctx->num_windows, ctx->num_buckets, bpw);

		size_t smem = FINALIZE_THREADS * sizeof(G1EdExtended);
		reduce_buckets_finalize_kernel<<<ctx->num_windows, FINALIZE_THREADS, smem, stream>>>(
			ctx->d_window_partial_totals, ctx->d_window_partial_sums, ctx->d_window_results,
			ctx->num_windows, ctx->num_buckets, bpw);
	}
}

void msm_download_results(MSMContext *ctx, G1EdExtended *host_results, cudaStream_t stream) {
	cudaMemcpyAsync(host_results, ctx->d_window_results,
		ctx->num_windows * sizeof(G1EdExtended), cudaMemcpyDeviceToHost, stream);
}

// =============================================================================
// Full MSM pipeline: fast scalar upload + single-pass compute
//
// For large n, uses cudaHostRegister to pin the caller's (Go heap) memory
// in-place, enabling full-bandwidth DMA without CPU-side memcpy through
// staging buffers. This avoids CUDA's internal pageable→pinned staging
// which is the main transfer bottleneck.
//
// Fallback: staging buffers for small n, or if registration fails.
// =============================================================================

cudaError_t msm_run_full(MSMContext *ctx, const uint64_t *host_scalars, size_t n,
                         G1EdExtended *host_results, cudaStream_t compute_stream) {

	// Lazy alloc sort buffers (d_scalars, d_keys, d_sort_temp, etc.)
	cudaError_t alloc_err = msm_alloc_work_buffers(ctx);
	if(alloc_err != cudaSuccess) return alloc_err;

	size_t total_bytes = n * 4 * sizeof(uint64_t);
	const size_t register_threshold = host_register_threshold_points();

	// For large n, try to pin the caller's memory for fast DMA
	bool registered = false;
	if(register_threshold > 0 && n >= register_threshold) {
		if(ctx->host_registered) {
			const bool pointer_changed = (ctx->registered_host_ptr != host_scalars);
			const bool need_larger_span = !pointer_changed &&
			                              (ctx->registered_host_bytes < total_bytes);
			if(pointer_changed || need_larger_span) {
				cudaHostUnregister((void *)ctx->registered_host_ptr);
				ctx->host_registered = false;
				ctx->registered_host_ptr = nullptr;
				ctx->registered_host_bytes = 0;
			}
		}
		if(!ctx->host_registered) {
			cudaError_t reg_err = cudaHostRegister(
				(void *)host_scalars, total_bytes, cudaHostRegisterDefault);
			if(reg_err == cudaSuccess) {
				ctx->host_registered = true;
				ctx->registered_host_ptr = host_scalars;
				ctx->registered_host_bytes = total_bytes;
			} else {
				cudaGetLastError();
			}
		}
		registered = ctx->host_registered;
	}

	if(registered) {
		cudaMemcpyAsync(ctx->d_scalars, host_scalars, total_bytes,
		                cudaMemcpyHostToDevice, compute_stream);
	} else {
		msm_upload_scalars(ctx, host_scalars, n, compute_stream);
	}
	launch_msm(ctx, n, compute_stream);
	msm_download_results(ctx, host_results, compute_stream);

	// Sync before freeing work buffers (kernels must finish using them).
	cudaError_t sync_err = cudaStreamSynchronize(compute_stream);

	// Unregister host memory before freeing sort buffers.
	msm_unregister_host(ctx);

	// Free sort buffers to reclaim VRAM for other phases.
	msm_free_work_buffers(ctx);

	return sync_err;
}

int msm_get_c(MSMContext *ctx) { return ctx->c; }
int msm_get_num_windows(MSMContext *ctx) { return ctx->num_windows; }

} // namespace gnark_gpu
