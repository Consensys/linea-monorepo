// Curve-generic Pippenger MSM for gpu/plonk2.
//
// This backend deliberately uses gnark-crypto's short-Weierstrass affine input
// layout for every curve. It keeps the PlonK commitment surface independent of
// the BLS12-377 twisted-Edwards specialization in gpu/plonk while retaining the
// same high-level pipeline:
//
//   scalar windows -> CUB radix sort -> bucket boundaries -> bucket sums
//   -> per-window reductions -> Horner combination
//
// Bucket accumulation remains intentionally simple. Window reduction is
// parallelized without exposing extra tuning knobs to the Go API.

#include "ec.cuh"

#include <cub/device/device_radix_sort.cuh>
#include <cuda_runtime.h>

#include <cstdint>
#include <limits>

namespace gnark_gpu::plonk2 {

namespace {

static constexpr int MSM_THREADS = 256;
static constexpr int ACCUM_PARALLEL_THREADS = 128;
static constexpr int ACCUM_SEQ_CAP = 256;
static constexpr int REDUCE_THREADS_PER_WINDOW = 128;
static constexpr int FINALIZE_THREADS = 32;

template <typename Fp>
__device__ __forceinline__ void load_affine(AffinePoint<Fp> &p,
                                            const uint64_t *raw) {
#pragma unroll
	for(int i = 0; i < Fp::LIMBS; i++) {
		p.x[i] = raw[i];
		p.y[i] = raw[Fp::LIMBS + i];
	}
}

template <typename Fp>
__device__ __forceinline__ void load_affine_at(AffinePoint<Fp> &p,
                                               const uint64_t *raw,
                                               size_t idx) {
	load_affine<Fp>(p, raw + idx * (2 * Fp::LIMBS));
}

template <typename Fp>
__device__ __forceinline__ void store_jacobian(const JacobianPoint<Fp> &p,
                                               uint64_t *raw) {
#pragma unroll
	for(int i = 0; i < Fp::LIMBS; i++) {
		raw[i] = p.x[i];
		raw[Fp::LIMBS + i] = p.y[i];
		raw[2 * Fp::LIMBS + i] = p.z[i];
	}
}

template <typename Fr>
__device__ __forceinline__ uint32_t scalar_window(const uint64_t *scalars,
                                                  size_t idx, int bit_offset,
                                                  int window_bits) {
	const uint64_t *scalar = scalars + idx * Fr::LIMBS;
	const int limb_idx = bit_offset >> 6;
	const int bit_shift = bit_offset & 63;
	if(limb_idx >= Fr::LIMBS) return 0;

	uint64_t digit = scalar[limb_idx] >> bit_shift;
	if(bit_shift + window_bits > 64 && limb_idx + 1 < Fr::LIMBS) {
		digit |= scalar[limb_idx + 1] << (64 - bit_shift);
	}
	const uint64_t mask = (uint64_t{1} << window_bits) - 1;
	return static_cast<uint32_t>(digit & mask);
}

template <typename Fr>
__global__ void __launch_bounds__(MSM_THREADS) build_pairs_kernel(
	const uint64_t *__restrict__ scalars,
	uint32_t *__restrict__ keys,
	uint32_t *__restrict__ vals,
	size_t count,
	int window_bits,
	int num_windows,
	int num_buckets,
	int total_buckets) {

	const size_t idx = static_cast<size_t>(blockIdx.x) * blockDim.x + threadIdx.x;
	if(idx >= count) return;

	uint32_t carry = 0;
	const uint32_t point_idx = static_cast<uint32_t>(idx);
	const uint32_t window_mask = (uint32_t{1} << window_bits) - 1;

	for(int w = 0; w < num_windows; w++) {
		uint32_t digit = scalar_window<Fr>(scalars, idx, w * window_bits,
		                                   window_bits);
		digit = (digit & window_mask) + carry;

		carry = digit > static_cast<uint32_t>(num_buckets) ? 1u : 0u;
		const uint32_t neg_digit = (uint32_t{1} << window_bits) - digit;
		uint32_t bucket = carry != 0 ? neg_digit : digit;
		uint32_t sign = carry;

		const uint32_t is_overflow = (bucket == 0 && sign != 0) ? 1u : 0u;
		carry |= is_overflow;
		sign &= ~is_overflow;

		const size_t out_idx = idx * static_cast<size_t>(num_windows) +
		                       static_cast<size_t>(w);
		keys[out_idx] = bucket == 0
			                ? static_cast<uint32_t>(total_buckets)
			                : static_cast<uint32_t>(w * num_buckets + bucket - 1);
		vals[out_idx] = point_idx | (sign << 31);
	}
}

__global__ void __launch_bounds__(MSM_THREADS) detect_bucket_boundaries_kernel(
	const uint32_t *__restrict__ sorted_keys,
	uint32_t *__restrict__ bucket_offsets,
	uint32_t *__restrict__ bucket_ends,
	size_t assignments,
	int total_buckets) {

	const size_t i = static_cast<size_t>(blockIdx.x) * blockDim.x + threadIdx.x;
	if(i >= assignments) return;

	const uint32_t key = sorted_keys[i];
	if(key >= static_cast<uint32_t>(total_buckets)) return;

	if(i == 0 || sorted_keys[i - 1] != key) {
		bucket_offsets[key] = static_cast<uint32_t>(i);
	}
	if(i == assignments - 1 || sorted_keys[i + 1] != key) {
		bucket_ends[key] = static_cast<uint32_t>(i + 1);
	}
}

template <typename Fp>
__global__ void __launch_bounds__(MSM_THREADS, 2) accumulate_buckets_kernel(
	const uint64_t *__restrict__ points,
	const uint32_t *__restrict__ point_indices,
	const uint32_t *__restrict__ bucket_offsets,
	const uint32_t *__restrict__ bucket_ends,
	JacobianPoint<Fp> *__restrict__ buckets,
	int total_buckets,
	bool add_to_existing,
	int cap,
	uint32_t *__restrict__ overflow_buckets,
	uint32_t *__restrict__ overflow_count) {

	const int bucket_flat = blockIdx.x * blockDim.x + threadIdx.x;
	if(bucket_flat >= total_buckets) return;

	JacobianPoint<Fp> acc, tmp;
	if(add_to_existing) {
		acc = buckets[bucket_flat];
	} else {
		jacobian_set_infinity<Fp>(acc);
	}

	const uint32_t start = bucket_offsets[bucket_flat];
	const uint32_t full_end = bucket_ends[bucket_flat];
	uint32_t end = full_end;
	if(cap > 0 && full_end > start + static_cast<uint32_t>(cap)) {
		end = start + static_cast<uint32_t>(cap);
		if(overflow_buckets && overflow_count) {
			const uint32_t slot = atomicAdd(overflow_count, 1u);
			overflow_buckets[slot] = static_cast<uint32_t>(bucket_flat);
		}
	}

	for(uint32_t i = start; i < end; i++) {
		const uint32_t packed = point_indices[i];
		AffinePoint<Fp> p;
		load_affine_at<Fp>(p, points, packed & 0x7fffffffu);
		if((packed >> 31) != 0) {
			neg<Fp>(p.y, p.y);
		}
		jacobian_add_jacobian_affine<Fp>(tmp, acc, p);
		set<Fp>(acc.x, tmp.x);
		set<Fp>(acc.y, tmp.y);
		set<Fp>(acc.z, tmp.z);
	}

	buckets[bucket_flat] = acc;
}

template <typename Fp>
__device__ __noinline__ void jacobian_add_value(JacobianPoint<Fp> &out,
                                                const JacobianPoint<Fp> &a,
                                                const JacobianPoint<Fp> &b) {
	jacobian_add<Fp>(out, a, b);
}

template <typename Fp>
__device__ __noinline__ void jacobian_double_value(JacobianPoint<Fp> &out,
                                                   const JacobianPoint<Fp> &a) {
	jacobian_double<Fp>(out, a);
}

template <typename Fp>
__global__ void __launch_bounds__(ACCUM_PARALLEL_THREADS, 2)
accumulate_buckets_parallel_kernel(
	const uint64_t *__restrict__ points,
	const uint32_t *__restrict__ point_indices,
	const uint32_t *__restrict__ bucket_offsets,
	const uint32_t *__restrict__ bucket_ends,
	const uint32_t *__restrict__ overflow_buckets,
	JacobianPoint<Fp> *__restrict__ buckets,
	bool add_to_existing,
	uint32_t start_offset) {

	const int bucket_flat = overflow_buckets
		                        ? static_cast<int>(overflow_buckets[blockIdx.x])
		                        : static_cast<int>(blockIdx.x);
	const int tid = threadIdx.x;
	const uint32_t start = bucket_offsets[bucket_flat] + start_offset;
	const uint32_t end = bucket_ends[bucket_flat];
	if(start >= end) return;

	JacobianPoint<Fp> acc, tmp;
	jacobian_set_infinity<Fp>(acc);
	for(uint32_t i = start + static_cast<uint32_t>(tid); i < end;
	    i += ACCUM_PARALLEL_THREADS) {
		const uint32_t packed = point_indices[i];
		AffinePoint<Fp> p;
		load_affine_at<Fp>(p, points, packed & 0x7fffffffu);
		if((packed >> 31) != 0) {
			neg<Fp>(p.y, p.y);
		}
		jacobian_add_jacobian_affine<Fp>(tmp, acc, p);
		acc = tmp;
	}

	extern __shared__ unsigned char shared_raw[];
	JacobianPoint<Fp> *shared =
		reinterpret_cast<JacobianPoint<Fp> *>(shared_raw);
	shared[tid] = acc;
	__syncthreads();

	for(int stride = ACCUM_PARALLEL_THREADS / 2; stride > 0; stride >>= 1) {
		if(tid < stride) {
			jacobian_add_value<Fp>(tmp, shared[tid], shared[tid + stride]);
			shared[tid] = tmp;
		}
		__syncthreads();
	}

	if(tid == 0) {
		if(add_to_existing) {
			jacobian_add_value<Fp>(tmp, buckets[bucket_flat], shared[0]);
			buckets[bucket_flat] = tmp;
		} else {
			buckets[bucket_flat] = shared[0];
		}
	}
}

template <typename Fp>
__device__ __forceinline__ void jacobian_mul_small(JacobianPoint<Fp> &out,
                                                   const JacobianPoint<Fp> &in,
                                                   int k) {
	jacobian_set_infinity<Fp>(out);
	if(k <= 0 || jacobian_is_infinity<Fp>(in)) return;

	JacobianPoint<Fp> base = in;
	while(k > 0) {
		if((k & 1) != 0) {
			JacobianPoint<Fp> tmp;
			jacobian_add_value<Fp>(tmp, out, base);
			out = tmp;
		}
		k >>= 1;
		if(k > 0) {
			JacobianPoint<Fp> tmp;
			jacobian_double_value<Fp>(tmp, base);
			base = tmp;
		}
	}
}

template <typename Fp>
__global__ void __launch_bounds__(REDUCE_THREADS_PER_WINDOW, 2)
reduce_windows_partial_kernel(
	const JacobianPoint<Fp> *__restrict__ buckets,
	JacobianPoint<Fp> *__restrict__ partial_totals,
	JacobianPoint<Fp> *__restrict__ partial_sums,
	int num_windows,
	int num_buckets,
	int blocks_per_window) {

	const int block_flat = blockIdx.x;
	const int w = block_flat / blocks_per_window;
	const int part = block_flat % blocks_per_window;
	if(w >= num_windows) return;

	const int tid = threadIdx.x;
	const int range_size = (num_buckets + blocks_per_window - 1) /
	                       blocks_per_window;
	const int high = num_buckets - 1 - part * range_size;
	const int out_idx = w * blocks_per_window + part;

	if(high < 0) {
		if(tid == 0) {
			jacobian_set_infinity<Fp>(partial_totals[out_idx]);
			jacobian_set_infinity<Fp>(partial_sums[out_idx]);
		}
		return;
	}

	int low = high - range_size + 1;
	if(low < 0) low = 0;
	const int range_len = high - low + 1;
	const int chunk_size =
		(range_len + REDUCE_THREADS_PER_WINDOW - 1) / REDUCE_THREADS_PER_WINDOW;
	int chunk_high = high - tid * chunk_size;
	int chunk_low = chunk_high - chunk_size + 1;
	if(chunk_low < low) chunk_low = low;
	if(chunk_high > high) chunk_high = high;
	const bool has_work = chunk_high >= low;

	JacobianPoint<Fp> local_running, local_total, tmp;
	jacobian_set_infinity<Fp>(local_running);
	jacobian_set_infinity<Fp>(local_total);
	int local_len = 0;

	if(has_work) {
		for(int b = chunk_high; b >= chunk_low; b--) {
			jacobian_add_value<Fp>(tmp, local_running,
			                       buckets[w * num_buckets + b]);
			local_running = tmp;

			jacobian_add_value<Fp>(tmp, local_total, local_running);
			local_total = tmp;
			local_len++;
		}
	}

	__shared__ JacobianPoint<Fp> shared[REDUCE_THREADS_PER_WINDOW];
	shared[tid] = local_running;
	__syncthreads();

	for(int d = 1; d < REDUCE_THREADS_PER_WINDOW; d <<= 1) {
		JacobianPoint<Fp> addend;
		const bool do_add = tid >= d;
		if(do_add) addend = shared[tid - d];
		__syncthreads();
		if(do_add) {
			jacobian_add_value<Fp>(tmp, shared[tid], addend);
			shared[tid] = tmp;
		}
		__syncthreads();
	}

	if(tid == 0) {
		partial_sums[out_idx] = shared[REDUCE_THREADS_PER_WINDOW - 1];
	}

	JacobianPoint<Fp> exclusive;
	if(tid == 0) {
		jacobian_set_infinity<Fp>(exclusive);
	} else {
		exclusive = shared[tid - 1];
	}
	__syncthreads();
	shared[tid] = exclusive;
	__syncthreads();

	JacobianPoint<Fp> correction;
	jacobian_mul_small<Fp>(correction, shared[tid], local_len);
	jacobian_add_value<Fp>(tmp, local_total, correction);
	local_total = tmp;

	shared[tid] = local_total;
	__syncthreads();
	for(int stride = REDUCE_THREADS_PER_WINDOW / 2; stride > 0; stride >>= 1) {
		if(tid < stride) {
			jacobian_add_value<Fp>(tmp, shared[tid], shared[tid + stride]);
			shared[tid] = tmp;
		}
		__syncthreads();
	}
	if(tid == 0) partial_totals[out_idx] = shared[0];
}

template <typename Fp>
__global__ void reduce_windows_finalize_kernel(
	const JacobianPoint<Fp> *__restrict__ partial_totals,
	const JacobianPoint<Fp> *__restrict__ partial_sums,
	JacobianPoint<Fp> *__restrict__ window_results,
	int num_windows,
	int num_buckets,
	int blocks_per_window) {

	const int w = blockIdx.x;
	if(w >= num_windows) return;

	const int tid = threadIdx.x;
	const int range_size = (num_buckets + blocks_per_window - 1) /
	                       blocks_per_window;

	extern __shared__ unsigned char raw_shared[];
	JacobianPoint<Fp> *shared =
		reinterpret_cast<JacobianPoint<Fp> *>(raw_shared);

	JacobianPoint<Fp> my_total, my_sum, tmp;
	int my_len = 0;
	if(tid < blocks_per_window) {
		const int high = num_buckets - 1 - tid * range_size;
		if(high >= 0) {
			int low = high - range_size + 1;
			if(low < 0) low = 0;
			my_len = high - low + 1;
			my_total = partial_totals[w * blocks_per_window + tid];
			my_sum = partial_sums[w * blocks_per_window + tid];
		} else {
			jacobian_set_infinity<Fp>(my_total);
			jacobian_set_infinity<Fp>(my_sum);
		}
	} else {
		jacobian_set_infinity<Fp>(my_total);
		jacobian_set_infinity<Fp>(my_sum);
	}

	shared[tid] = my_sum;
	__syncthreads();
	for(int d = 1; d < FINALIZE_THREADS; d <<= 1) {
		JacobianPoint<Fp> addend;
		const bool do_add = tid >= d && tid < blocks_per_window;
		if(do_add) addend = shared[tid - d];
		__syncthreads();
		if(do_add) {
			jacobian_add_value<Fp>(tmp, shared[tid], addend);
			shared[tid] = tmp;
		}
		__syncthreads();
	}

	JacobianPoint<Fp> exclusive;
	if(tid == 0 || tid >= blocks_per_window) {
		jacobian_set_infinity<Fp>(exclusive);
	} else {
		exclusive = shared[tid - 1];
	}
	__syncthreads();
	shared[tid] = exclusive;
	__syncthreads();

	if(tid < blocks_per_window && my_len > 0) {
		JacobianPoint<Fp> correction;
		jacobian_mul_small<Fp>(correction, shared[tid], my_len);
		jacobian_add_value<Fp>(tmp, my_total, correction);
		my_total = tmp;
	}

	shared[tid] = my_total;
	__syncthreads();
	for(int stride = FINALIZE_THREADS / 2; stride > 0; stride >>= 1) {
		if(tid < stride) {
			jacobian_add_value<Fp>(tmp, shared[tid], shared[tid + stride]);
			shared[tid] = tmp;
		}
		__syncthreads();
	}
	if(tid == 0) window_results[w] = shared[0];
}

static int reduce_blocks_per_window(int num_windows, int num_buckets) {
	int max_bpw = num_buckets / REDUCE_THREADS_PER_WINDOW;
	int target_bpw = 752 / num_windows;
	int bpw = max_bpw < target_bpw ? max_bpw : target_bpw;
	if(bpw < 1) bpw = 1;
	if(bpw > FINALIZE_THREADS) bpw = FINALIZE_THREADS;
	return bpw;
}

template <typename Fp>
__global__ void finalize_msm_kernel(
	const JacobianPoint<Fp> *__restrict__ window_results,
	int num_windows,
	int window_bits,
	uint64_t *__restrict__ out_raw) {

	if(blockIdx.x != 0 || threadIdx.x != 0) return;

	JacobianPoint<Fp> acc, tmp;
	jacobian_set_infinity<Fp>(acc);

	for(int w = num_windows - 1; w >= 0; w--) {
		if(w != num_windows - 1) {
			for(int i = 0; i < window_bits; i++) {
				jacobian_double<Fp>(tmp, acc);
				set<Fp>(acc.x, tmp.x);
				set<Fp>(acc.y, tmp.y);
				set<Fp>(acc.z, tmp.z);
			}
		}
		jacobian_add<Fp>(tmp, acc, window_results[w]);
		set<Fp>(acc.x, tmp.x);
		set<Fp>(acc.y, tmp.y);
		set<Fp>(acc.z, tmp.z);
	}

	store_jacobian<Fp>(acc, out_raw);
}

static int signed_window_count(int scalar_bits, int window_bits) {
	return (scalar_bits + 1 + window_bits - 1) / window_bits;
}

static int sort_key_bits(int total_buckets) {
	int bits = 1;
	while((1u << bits) <= static_cast<uint32_t>(total_buckets)) bits++;
	return bits;
}

static int accumulation_seq_cap(size_t assignments, int total_buckets) {
	size_t avg = assignments / static_cast<size_t>(total_buckets);
	size_t cap = 2 * avg + 64;
	if(cap < static_cast<size_t>(ACCUM_SEQ_CAP)) cap = ACCUM_SEQ_CAP;
	if(cap > 4096) cap = 4096;
	return static_cast<int>(cap);
}

template <typename Fp, typename Fr>
cudaError_t run_msm_pippenger_core(
	const uint64_t *points,
	bool points_are_device_resident,
	const uint64_t *scalars,
	size_t count,
	int window_bits,
	uint64_t *out,
	cudaStream_t stream) {

	if(window_bits <= 1 || window_bits > 24) return cudaErrorInvalidValue;
	if(count == 0 || count > static_cast<size_t>(std::numeric_limits<int32_t>::max())) {
		return cudaErrorInvalidValue;
	}

	const int num_windows = signed_window_count(Fr::BITS, window_bits);
	const int num_buckets = 1 << (window_bits - 1);
	const int total_buckets = num_windows * num_buckets;
	const int reduce_bpw = reduce_blocks_per_window(num_windows, num_buckets);
	const int total_partials = num_windows * reduce_bpw;
	const size_t assignments = count * static_cast<size_t>(num_windows);
	if(assignments > static_cast<size_t>(std::numeric_limits<uint32_t>::max())) {
		return cudaErrorInvalidValue;
	}

	const uint64_t *d_points = nullptr;
	uint64_t *owned_d_points = nullptr;
	uint64_t *d_scalars = nullptr;
	uint64_t *d_out = nullptr;
	uint32_t *d_keys_in = nullptr;
	uint32_t *d_keys_out = nullptr;
	uint32_t *d_vals_in = nullptr;
	uint32_t *d_vals_out = nullptr;
	uint32_t *d_bucket_offsets = nullptr;
	uint32_t *d_bucket_ends = nullptr;
	uint32_t *d_overflow_buckets = nullptr;
	uint32_t *d_overflow_count = nullptr;
	JacobianPoint<Fp> *d_buckets = nullptr;
	JacobianPoint<Fp> *d_window_results = nullptr;
	JacobianPoint<Fp> *d_partial_totals = nullptr;
	JacobianPoint<Fp> *d_partial_sums = nullptr;
	void *d_sort_temp = nullptr;
	size_t sort_temp_bytes = 0;

	const size_t point_words = count * 2 * Fp::LIMBS;
	const size_t scalar_words = count * Fr::LIMBS;
	constexpr size_t output_words = 3 * Fp::LIMBS;

	cudaError_t err = cudaSuccess;
	if(points_are_device_resident) {
		d_points = points;
	} else {
		err = cudaMalloc(&owned_d_points, point_words * sizeof(uint64_t));
		if(err != cudaSuccess) goto done;
		d_points = owned_d_points;
	}
	err = cudaMalloc(&d_scalars, scalar_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_out, output_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_keys_in, assignments * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_keys_out, assignments * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_vals_in, assignments * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_vals_out, assignments * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_bucket_offsets, total_buckets * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_bucket_ends, total_buckets * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_overflow_buckets, total_buckets * sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_overflow_count, sizeof(uint32_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_buckets,
	                 total_buckets * sizeof(JacobianPoint<Fp>));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_window_results,
	                 num_windows * sizeof(JacobianPoint<Fp>));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_partial_totals,
	                 total_partials * sizeof(JacobianPoint<Fp>));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_partial_sums,
	                 total_partials * sizeof(JacobianPoint<Fp>));
	if(err != cudaSuccess) goto done;

	if(!points_are_device_resident) {
		err = cudaMemcpyAsync(owned_d_points, points,
		                      point_words * sizeof(uint64_t),
		                      cudaMemcpyHostToDevice, stream);
		if(err != cudaSuccess) goto done;
	}
	err = cudaMemcpyAsync(d_scalars, scalars, scalar_words * sizeof(uint64_t),
	                      cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;

	{
		const unsigned blocks =
			static_cast<unsigned>((count + MSM_THREADS - 1) / MSM_THREADS);
		build_pairs_kernel<Fr><<<blocks, MSM_THREADS, 0, stream>>>(
			d_scalars, d_keys_in, d_vals_in, count, window_bits, num_windows,
			num_buckets, total_buckets);
		err = cudaGetLastError();
		if(err != cudaSuccess) goto done;
	}

	err = cub::DeviceRadixSort::SortPairs(
		nullptr, sort_temp_bytes, d_keys_in, d_keys_out, d_vals_in, d_vals_out,
		assignments, 0, sort_key_bits(total_buckets), stream);
	if(err != cudaSuccess) goto done;

	err = cudaMalloc(&d_sort_temp, sort_temp_bytes);
	if(err != cudaSuccess) goto done;

	err = cub::DeviceRadixSort::SortPairs(
		d_sort_temp, sort_temp_bytes, d_keys_in, d_keys_out, d_vals_in,
		d_vals_out, assignments, 0, sort_key_bits(total_buckets), stream);
	if(err != cudaSuccess) goto done;

	err = cudaMemsetAsync(d_bucket_offsets, 0,
	                      total_buckets * sizeof(uint32_t), stream);
	if(err != cudaSuccess) goto done;
	err = cudaMemsetAsync(d_bucket_ends, 0,
	                      total_buckets * sizeof(uint32_t), stream);
	if(err != cudaSuccess) goto done;

	{
		const unsigned blocks = static_cast<unsigned>(
			(assignments + MSM_THREADS - 1) / MSM_THREADS);
		detect_bucket_boundaries_kernel<<<blocks, MSM_THREADS, 0, stream>>>(
			d_keys_out, d_bucket_offsets, d_bucket_ends, assignments,
			total_buckets);
		err = cudaGetLastError();
		if(err != cudaSuccess) goto done;
	}

	{
		const int cap = accumulation_seq_cap(assignments, total_buckets);
		const unsigned blocks =
			static_cast<unsigned>((total_buckets + MSM_THREADS - 1) /
			                      MSM_THREADS);
		err = cudaMemsetAsync(d_overflow_count, 0, sizeof(uint32_t), stream);
		if(err != cudaSuccess) goto done;
		accumulate_buckets_kernel<Fp><<<blocks, MSM_THREADS, 0, stream>>>(
			d_points, d_vals_out, d_bucket_offsets, d_bucket_ends, d_buckets,
			total_buckets, false, cap, d_overflow_buckets, d_overflow_count);
		err = cudaGetLastError();
		if(err != cudaSuccess) goto done;

		uint32_t overflow_count = 0;
		err = cudaMemcpyAsync(&overflow_count, d_overflow_count,
		                      sizeof(uint32_t), cudaMemcpyDeviceToHost,
		                      stream);
		if(err != cudaSuccess) goto done;
		err = cudaStreamSynchronize(stream);
		if(err != cudaSuccess) goto done;
		if(overflow_count > 0) {
			const size_t smem =
				ACCUM_PARALLEL_THREADS * sizeof(JacobianPoint<Fp>);
			accumulate_buckets_parallel_kernel<Fp>
				<<<overflow_count, ACCUM_PARALLEL_THREADS, smem, stream>>>(
					d_points, d_vals_out, d_bucket_offsets, d_bucket_ends,
					d_overflow_buckets, d_buckets, true,
					static_cast<uint32_t>(cap));
			err = cudaGetLastError();
			if(err != cudaSuccess) goto done;
		}
	}

	reduce_windows_partial_kernel<Fp>
		<<<total_partials, REDUCE_THREADS_PER_WINDOW, 0, stream>>>(
			d_buckets, d_partial_totals, d_partial_sums, num_windows,
			num_buckets, reduce_bpw);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	reduce_windows_finalize_kernel<Fp>
		<<<num_windows, FINALIZE_THREADS,
		   FINALIZE_THREADS * sizeof(JacobianPoint<Fp>), stream>>>(
			d_partial_totals, d_partial_sums, d_window_results, num_windows,
			num_buckets, reduce_bpw);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	finalize_msm_kernel<Fp><<<1, 1, 0, stream>>>(
		d_window_results, num_windows, window_bits, d_out);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(out, d_out, output_words * sizeof(uint64_t),
	                      cudaMemcpyDeviceToHost, stream);
	if(err != cudaSuccess) goto done;
	err = cudaStreamSynchronize(stream);

done:
	if(owned_d_points) cudaFree(owned_d_points);
	if(d_scalars) cudaFree(d_scalars);
	if(d_out) cudaFree(d_out);
	if(d_keys_in) cudaFree(d_keys_in);
	if(d_keys_out) cudaFree(d_keys_out);
	if(d_vals_in) cudaFree(d_vals_in);
	if(d_vals_out) cudaFree(d_vals_out);
	if(d_bucket_offsets) cudaFree(d_bucket_offsets);
	if(d_bucket_ends) cudaFree(d_bucket_ends);
	if(d_overflow_buckets) cudaFree(d_overflow_buckets);
	if(d_overflow_count) cudaFree(d_overflow_count);
	if(d_buckets) cudaFree(d_buckets);
	if(d_window_results) cudaFree(d_window_results);
	if(d_partial_totals) cudaFree(d_partial_totals);
	if(d_partial_sums) cudaFree(d_partial_sums);
	if(d_sort_temp) cudaFree(d_sort_temp);
	return err;
}

template <typename Fp, typename Fr>
cudaError_t run_msm_pippenger(
	const uint64_t *points,
	const uint64_t *scalars,
	size_t count,
	int window_bits,
	uint64_t *out,
	cudaStream_t stream) {

	return run_msm_pippenger_core<Fp, Fr>(
		points, false, scalars, count, window_bits, out, stream);
}

template <typename Fp, typename Fr>
cudaError_t run_msm_pippenger_device_points(
	const uint64_t *d_points,
	const uint64_t *scalars,
	size_t count,
	int window_bits,
	uint64_t *out,
	cudaStream_t stream) {

	return run_msm_pippenger_core<Fp, Fr>(
		d_points, true, scalars, count, window_bits, out, stream);
}

template <typename Fr>
cudaError_t sort_temp_bytes_for(size_t count, int window_bits,
                                size_t *temp_bytes) {
	if(!temp_bytes || window_bits <= 1 || window_bits > 24 || count == 0) {
		return cudaErrorInvalidValue;
	}
	const int num_windows = signed_window_count(Fr::BITS, window_bits);
	const int num_buckets = 1 << (window_bits - 1);
	const int total_buckets = num_windows * num_buckets;
	const size_t assignments = count * static_cast<size_t>(num_windows);
	if(assignments > static_cast<size_t>(std::numeric_limits<uint32_t>::max())) {
		return cudaErrorInvalidValue;
	}

	uint32_t *keys_in = nullptr;
	uint32_t *keys_out = nullptr;
	uint32_t *vals_in = nullptr;
	uint32_t *vals_out = nullptr;
	return cub::DeviceRadixSort::SortPairs(
		nullptr, *temp_bytes, keys_in, keys_out, vals_in, vals_out,
		assignments, 0, sort_key_bits(total_buckets), nullptr);
}

template <typename Fp, typename Fr>
cudaError_t run_msm_pippenger_prealloc_core(
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
	uint32_t *d_overflow_buckets,
	uint32_t *d_overflow_count,
	void *d_buckets_raw,
	void *d_window_results_raw,
	void *d_partial_totals_raw,
	void *d_partial_sums_raw,
	void *d_sort_temp,
	size_t sort_temp_bytes,
	cudaEvent_t *phase_events,
	float *phase_timings_ms,
	cudaStream_t stream) {

	if(window_bits <= 1 || window_bits > 24) return cudaErrorInvalidValue;
	if(count == 0 || count > static_cast<size_t>(std::numeric_limits<int32_t>::max())) {
		return cudaErrorInvalidValue;
	}

	const int num_windows = signed_window_count(Fr::BITS, window_bits);
	const int num_buckets = 1 << (window_bits - 1);
	const int total_buckets = num_windows * num_buckets;
	const int reduce_bpw = reduce_blocks_per_window(num_windows, num_buckets);
	const int total_partials = num_windows * reduce_bpw;
	const size_t assignments = count * static_cast<size_t>(num_windows);
	if(assignments > static_cast<size_t>(std::numeric_limits<uint32_t>::max())) {
		return cudaErrorInvalidValue;
	}

	if(!d_points || !scalars || !out || !d_scalars || !d_out || !d_keys_in ||
	   !d_keys_out || !d_vals_in || !d_vals_out || !d_bucket_offsets ||
	   !d_bucket_ends || !d_overflow_buckets || !d_overflow_count ||
	   !d_buckets_raw || !d_window_results_raw || !d_partial_totals_raw ||
	   !d_partial_sums_raw || !d_sort_temp) {
		return cudaErrorInvalidValue;
	}
	if(sort_temp_bytes == 0) return cudaErrorInvalidValue;

	JacobianPoint<Fp> *d_buckets =
		reinterpret_cast<JacobianPoint<Fp> *>(d_buckets_raw);
	JacobianPoint<Fp> *d_window_results =
		reinterpret_cast<JacobianPoint<Fp> *>(d_window_results_raw);
	JacobianPoint<Fp> *d_partial_totals =
		reinterpret_cast<JacobianPoint<Fp> *>(d_partial_totals_raw);
	JacobianPoint<Fp> *d_partial_sums =
		reinterpret_cast<JacobianPoint<Fp> *>(d_partial_sums_raw);

	const size_t scalar_words = count * Fr::LIMBS;
	constexpr size_t output_words = 3 * Fp::LIMBS;

	auto record_phase = [&](int idx) {
		if(phase_events) cudaEventRecord(phase_events[idx], stream);
	};
	record_phase(0);

	cudaError_t err = cudaMemcpyAsync(
		d_scalars, scalars, scalar_words * sizeof(uint64_t),
		cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) return err;
	record_phase(1);

	{
		const unsigned blocks =
			static_cast<unsigned>((count + MSM_THREADS - 1) / MSM_THREADS);
		build_pairs_kernel<Fr><<<blocks, MSM_THREADS, 0, stream>>>(
			d_scalars, d_keys_in, d_vals_in, count, window_bits, num_windows,
			num_buckets, total_buckets);
		err = cudaGetLastError();
		if(err != cudaSuccess) return err;
	}
	record_phase(2);

	err = cub::DeviceRadixSort::SortPairs(
		d_sort_temp, sort_temp_bytes, d_keys_in, d_keys_out, d_vals_in,
		d_vals_out, assignments, 0, sort_key_bits(total_buckets), stream);
	if(err != cudaSuccess) return err;
	record_phase(3);

	err = cudaMemsetAsync(d_bucket_offsets, 0,
	                      total_buckets * sizeof(uint32_t), stream);
	if(err != cudaSuccess) return err;
	err = cudaMemsetAsync(d_bucket_ends, 0,
	                      total_buckets * sizeof(uint32_t), stream);
	if(err != cudaSuccess) return err;

	{
		const unsigned blocks = static_cast<unsigned>(
			(assignments + MSM_THREADS - 1) / MSM_THREADS);
		detect_bucket_boundaries_kernel<<<blocks, MSM_THREADS, 0, stream>>>(
			d_keys_out, d_bucket_offsets, d_bucket_ends, assignments,
			total_buckets);
		err = cudaGetLastError();
		if(err != cudaSuccess) return err;
	}
	record_phase(4);

	{
		const int cap = accumulation_seq_cap(assignments, total_buckets);
		const unsigned blocks =
			static_cast<unsigned>((total_buckets + MSM_THREADS - 1) /
			                      MSM_THREADS);
		err = cudaMemsetAsync(d_overflow_count, 0, sizeof(uint32_t), stream);
		if(err != cudaSuccess) return err;
		accumulate_buckets_kernel<Fp><<<blocks, MSM_THREADS, 0, stream>>>(
			d_points, d_vals_out, d_bucket_offsets, d_bucket_ends, d_buckets,
			total_buckets, false, cap, d_overflow_buckets, d_overflow_count);
		err = cudaGetLastError();
		if(err != cudaSuccess) return err;
	}
	record_phase(5);

	{
		uint32_t overflow_count = 0;
		err = cudaMemcpyAsync(&overflow_count, d_overflow_count,
		                      sizeof(uint32_t), cudaMemcpyDeviceToHost,
		                      stream);
		if(err != cudaSuccess) return err;
		err = cudaStreamSynchronize(stream);
		if(err != cudaSuccess) return err;
		if(overflow_count > 0) {
			const int cap = accumulation_seq_cap(assignments, total_buckets);
			const size_t smem =
				ACCUM_PARALLEL_THREADS * sizeof(JacobianPoint<Fp>);
			accumulate_buckets_parallel_kernel<Fp>
				<<<overflow_count, ACCUM_PARALLEL_THREADS, smem, stream>>>(
					d_points, d_vals_out, d_bucket_offsets, d_bucket_ends,
					d_overflow_buckets, d_buckets, true,
					static_cast<uint32_t>(cap));
			err = cudaGetLastError();
			if(err != cudaSuccess) return err;
		}
	}
	record_phase(6);

	reduce_windows_partial_kernel<Fp>
		<<<total_partials, REDUCE_THREADS_PER_WINDOW, 0, stream>>>(
			d_buckets, d_partial_totals, d_partial_sums, num_windows,
			num_buckets, reduce_bpw);
	err = cudaGetLastError();
	if(err != cudaSuccess) return err;
	record_phase(7);

	reduce_windows_finalize_kernel<Fp>
		<<<num_windows, FINALIZE_THREADS,
		   FINALIZE_THREADS * sizeof(JacobianPoint<Fp>), stream>>>(
			d_partial_totals, d_partial_sums, d_window_results, num_windows,
			num_buckets, reduce_bpw);
	err = cudaGetLastError();
	if(err != cudaSuccess) return err;
	record_phase(8);

	finalize_msm_kernel<Fp><<<1, 1, 0, stream>>>(
		d_window_results, num_windows, window_bits, d_out);
	err = cudaGetLastError();
	if(err != cudaSuccess) return err;

	err = cudaMemcpyAsync(out, d_out, output_words * sizeof(uint64_t),
	                      cudaMemcpyDeviceToHost, stream);
	if(err != cudaSuccess) return err;
	record_phase(9);

	err = cudaStreamSynchronize(stream);
	if(err != cudaSuccess) return err;

	if(phase_events && phase_timings_ms) {
		auto elapsed = [&](int from, int to) -> float {
			float ms = 0.0f;
			if(cudaEventElapsedTime(&ms, phase_events[from], phase_events[to])
			   != cudaSuccess) {
				cudaGetLastError();
				ms = 0.0f;
			}
			return ms;
		};
		phase_timings_ms[0] = elapsed(0, 1);
		phase_timings_ms[1] = elapsed(1, 2);
		phase_timings_ms[2] = elapsed(2, 3);
		phase_timings_ms[3] = elapsed(3, 4);
		phase_timings_ms[4] = elapsed(4, 5);
		phase_timings_ms[5] = elapsed(5, 6);
		phase_timings_ms[6] = elapsed(6, 7);
		phase_timings_ms[7] = elapsed(7, 8);
		phase_timings_ms[8] = elapsed(8, 9);
	}
	return cudaSuccess;
}

} // namespace

cudaError_t msm_pippenger_sort_temp_bytes(
	gnark_gpu_plonk2_curve_id_t curve,
	size_t count,
	int window_bits,
	size_t *temp_bytes) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return sort_temp_bytes_for<BN254FrParams>(
			count, window_bits, temp_bytes);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return sort_temp_bytes_for<BLS12377FrParams>(
			count, window_bits, temp_bytes);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return sort_temp_bytes_for<BW6761FrParams>(
			count, window_bits, temp_bytes);
	default:
		return cudaErrorInvalidValue;
	}
}

cudaError_t msm_pippenger_run(
	gnark_gpu_plonk2_curve_id_t curve,
	const uint64_t *points,
	const uint64_t *scalars,
	size_t count,
	int window_bits,
	uint64_t *out,
	cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_msm_pippenger<BN254FpParams, BN254FrParams>(
			points, scalars, count, window_bits, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_msm_pippenger<BLS12377FpParams, BLS12377FrParams>(
			points, scalars, count, window_bits, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_msm_pippenger<BW6761FpParams, BW6761FrParams>(
			points, scalars, count, window_bits, out, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

cudaError_t msm_pippenger_device_points_run(
	gnark_gpu_plonk2_curve_id_t curve,
	const uint64_t *d_points,
	const uint64_t *scalars,
	size_t count,
	int window_bits,
	uint64_t *out,
	cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_msm_pippenger_device_points<BN254FpParams, BN254FrParams>(
			d_points, scalars, count, window_bits, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_msm_pippenger_device_points<BLS12377FpParams, BLS12377FrParams>(
			d_points, scalars, count, window_bits, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_msm_pippenger_device_points<BW6761FpParams, BW6761FrParams>(
			d_points, scalars, count, window_bits, out, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

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
	uint32_t *d_overflow_buckets,
	uint32_t *d_overflow_count,
	void *d_buckets,
	void *d_window_results,
	void *d_partial_totals,
	void *d_partial_sums,
	void *d_sort_temp,
	size_t sort_temp_bytes,
	cudaEvent_t *phase_events,
	float *phase_timings_ms,
	cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_msm_pippenger_prealloc_core<BN254FpParams, BN254FrParams>(
			d_points, scalars, count, window_bits, out, d_scalars, d_out,
			d_keys_in, d_keys_out, d_vals_in, d_vals_out, d_bucket_offsets,
			d_bucket_ends, d_overflow_buckets, d_overflow_count, d_buckets,
			d_window_results, d_partial_totals, d_partial_sums, d_sort_temp,
			sort_temp_bytes, phase_events, phase_timings_ms, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_msm_pippenger_prealloc_core<BLS12377FpParams, BLS12377FrParams>(
			d_points, scalars, count, window_bits, out, d_scalars, d_out,
			d_keys_in, d_keys_out, d_vals_in, d_vals_out, d_bucket_offsets,
			d_bucket_ends, d_overflow_buckets, d_overflow_count, d_buckets,
			d_window_results, d_partial_totals, d_partial_sums, d_sort_temp,
			sort_temp_bytes, phase_events, phase_timings_ms, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_msm_pippenger_prealloc_core<BW6761FpParams, BW6761FrParams>(
			d_points, scalars, count, window_bits, out, d_scalars, d_out,
			d_keys_in, d_keys_out, d_vals_in, d_vals_out, d_bucket_offsets,
			d_bucket_ends, d_overflow_buckets, d_overflow_count, d_buckets,
			d_window_results, d_partial_totals, d_partial_sums, d_sort_temp,
			sort_temp_bytes, phase_events, phase_timings_ms, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

} // namespace gnark_gpu::plonk2
