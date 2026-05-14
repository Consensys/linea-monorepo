#include "gnark_gpu.h"
#include "field.cuh"

#include <cuda_runtime.h>
#include <chrono>
#include <cstddef>
#include <cstdint>
#include <cstdio>
#include <cstdlib>

namespace {

using Params = gnark_gpu::plonk2::BLS12377FrParams;

constexpr unsigned THREADS = 256;
constexpr unsigned SIS_THREADS = 64;
constexpr unsigned SIS_COLS_PER_BLOCK = 2;
constexpr int MIMC_ROUNDS = 62;
constexpr int SIS_DEGREE = 64;
constexpr int SIS_LIMBS_PER_FIELD = 16;
constexpr int SIS_TWIDDLES_SIZE = 69;

enum RowKind : uint8_t {
	ROW_KIND_REGULAR = 0,
	ROW_KIND_CONSTANT = 1,
};

__host__ __forceinline__ int grid(size_t n) {
	return static_cast<int>((n + THREADS - 1) / THREADS);
}

using Clock = std::chrono::steady_clock;
using TimePoint = std::chrono::time_point<Clock>;

bool timing_enabled() {
	static const bool enabled = [] {
		const char *v = std::getenv("LINEA_PROVER_GPU_PI_VORTEX_TIMINGS");
		return v != nullptr && v[0] != '\0' && !(v[0] == '0' && v[1] == '\0');
	}();
	return enabled;
}

TimePoint now_if(bool enabled) {
	return enabled ? Clock::now() : TimePoint{};
}

double elapsed_ms(TimePoint start, TimePoint stop) {
	return std::chrono::duration<double, std::milli>(stop - start).count();
}

void log_timing(const char *name, size_t rows, size_t cols, size_t elems,
                double malloc_ms, double h2d_ms, double static_h2d_ms,
                double leaf_ms, double tree_ms, double d2h_ms, double total_ms) {
	std::fprintf(stderr,
	             "[gpu-pi-vortex] op=%s rows=%zu cols=%zu elems=%zu "
	             "malloc=%.3fms h2d=%.3fms static_h2d=%.3fms leaf=%.3fms "
	             "tree=%.3fms d2h=%.3fms total=%.3fms\n",
	             name, rows, cols, elems, malloc_ms, h2d_ms, static_h2d_ms,
	             leaf_ms, tree_ms, d2h_ms, total_ms);
}

gnark_gpu_error_t check(cudaError_t err) {
	if(err == cudaSuccess) return GNARK_GPU_SUCCESS;
	if(err == cudaErrorMemoryAllocation) return GNARK_GPU_ERROR_OUT_OF_MEMORY;
	return GNARK_GPU_ERROR_CUDA;
}

__device__ __forceinline__ void load_aos(uint64_t out[Params::LIMBS],
                                         const uint64_t *src, size_t idx) {
	const uint64_t *p = src + idx * Params::LIMBS;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) out[i] = p[i];
}

__device__ __forceinline__ void store_aos(uint64_t *dst, size_t idx,
                                          const uint64_t in[Params::LIMBS]) {
	uint64_t *p = dst + idx * Params::LIMBS;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) p[i] = in[i];
}

__device__ __forceinline__ size_t twiddle_offset(int stage) {
	size_t off = 0;
#pragma unroll
	for(int i = 0; i < 6; i++) {
		if(i >= stage) break;
		off += 1 + (SIS_DEGREE >> (i + 1));
	}
	return off;
}

__device__ __forceinline__ void butterfly(uint64_t a[Params::LIMBS],
                                          uint64_t b[Params::LIMBS]) {
	uint64_t t[Params::LIMBS];
	gnark_gpu::plonk2::set<Params>(t, a);
	gnark_gpu::plonk2::add<Params>(a, a, b);
	gnark_gpu::plonk2::sub<Params>(b, t, b);
}

__device__ __forceinline__ void set_raw_u64(uint64_t out[Params::LIMBS], uint64_t v) {
	out[0] = v;
#pragma unroll
	for(int i = 1; i < Params::LIMBS; i++) out[i] = 0;
}

__device__ __forceinline__ void from_montgomery(
	uint64_t out[Params::LIMBS],
	const uint64_t in[Params::LIMBS]) {
	uint64_t raw_one[Params::LIMBS] = {1, 0, 0, 0};
	gnark_gpu::plonk2::mul<Params>(out, in, raw_one);
}

__device__ __forceinline__ uint16_t limb16(const uint64_t raw[Params::LIMBS], int limb) {
	return static_cast<uint16_t>((raw[limb >> 2] >> ((limb & 3) * 16)) & 0xffffULL);
}

__device__ __forceinline__ void fft_dif_coset_64_at(
	uint64_t a[SIS_DEGREE][Params::LIMBS],
	const uint64_t *twiddles,
	const uint64_t *coset,
	int tid) {

	uint64_t c[Params::LIMBS];
	load_aos(c, coset, tid);
	gnark_gpu::plonk2::mul<Params>(a[tid], a[tid], c);
	__syncthreads();

#pragma unroll
	for(int stage = 0; stage < 6; stage++) {
		const int half = SIS_DEGREE >> (stage + 1);
		const int segment = half << 1;
		const int local = tid & (segment - 1);
		if(local < half) {
			const int i = tid - local + local;
			const int j = i + half;
			butterfly(a[i], a[j]);
			if(local != 0) {
				uint64_t tw[Params::LIMBS];
				load_aos(tw, twiddles, twiddle_offset(stage) + local);
				gnark_gpu::plonk2::mul<Params>(a[j], a[j], tw);
			}
		}
		__syncthreads();
	}
}

__device__ __forceinline__ void fft_dif_coset_64(
	uint64_t a[SIS_DEGREE][Params::LIMBS],
	const uint64_t *twiddles,
	const uint64_t *coset) {

	fft_dif_coset_64_at(a, twiddles, coset, threadIdx.x);
}

__device__ __forceinline__ void fft_inverse_dit_coset_64_at(
	uint64_t a[SIS_DEGREE][Params::LIMBS],
	const uint64_t *twiddles_inv,
	const uint64_t *coset_inv,
	const uint64_t *cardinality_inv,
	int tid) {

#pragma unroll
	for(int stage = 5; stage >= 0; stage--) {
		const int half = SIS_DEGREE >> (stage + 1);
		const int segment = half << 1;
		const int local = tid & (segment - 1);
		if(local < half) {
			const int i = tid - local + local;
			const int j = i + half;
			if(local != 0) {
				uint64_t tw[Params::LIMBS];
				load_aos(tw, twiddles_inv, twiddle_offset(stage) + local);
				gnark_gpu::plonk2::mul<Params>(a[j], a[j], tw);
			}
			butterfly(a[i], a[j]);
		}
		__syncthreads();
	}

	uint64_t c[Params::LIMBS], n_inv[Params::LIMBS];
	load_aos(c, coset_inv, tid);
	load_aos(n_inv, cardinality_inv, 0);
	gnark_gpu::plonk2::mul<Params>(a[tid], a[tid], c);
	gnark_gpu::plonk2::mul<Params>(a[tid], a[tid], n_inv);
	__syncthreads();
}

__device__ __forceinline__ void fft_inverse_dit_coset_64(
	uint64_t a[SIS_DEGREE][Params::LIMBS],
	const uint64_t *twiddles_inv,
	const uint64_t *coset_inv,
	const uint64_t *cardinality_inv) {

	fft_inverse_dit_coset_64_at(a, twiddles_inv, coset_inv, cardinality_inv, threadIdx.x);
}

__device__ __forceinline__ void mimc_encrypt(
	uint64_t out[Params::LIMBS],
	const uint64_t message[Params::LIMBS],
	const uint64_t key[Params::LIMBS],
	const uint64_t *constants) {

	uint64_t m[Params::LIMBS];
	gnark_gpu::plonk2::set<Params>(m, message);

	for(int r = 0; r < MIMC_ROUNDS; r++) {
		uint64_t c[Params::LIMBS], tmp[Params::LIMBS];
		load_aos(c, constants, static_cast<size_t>(r));
		gnark_gpu::plonk2::add<Params>(tmp, m, key);
		gnark_gpu::plonk2::add<Params>(tmp, tmp, c);

		// tmp^17 = ((((tmp^2)^2)^2)^2) * tmp.
		gnark_gpu::plonk2::square<Params>(m, tmp);
		gnark_gpu::plonk2::square<Params>(m, m);
		gnark_gpu::plonk2::square<Params>(m, m);
		gnark_gpu::plonk2::square<Params>(m, m);
		gnark_gpu::plonk2::mul<Params>(m, m, tmp);
	}

	gnark_gpu::plonk2::add<Params>(out, m, key);
}

__device__ __forceinline__ void mimc_absorb(
	uint64_t state[Params::LIMBS],
	const uint64_t message[Params::LIMBS],
	const uint64_t *constants) {

	uint64_t encrypted[Params::LIMBS], next[Params::LIMBS];
	mimc_encrypt(encrypted, message, state, constants);
	gnark_gpu::plonk2::add<Params>(next, encrypted, state);
	gnark_gpu::plonk2::add<Params>(state, next, message);
}

__global__ void sis_leaf_kernel(
	const uint64_t *__restrict__ col_hashes,
	size_t chunk_size,
	const uint64_t *__restrict__ constants,
	uint64_t *__restrict__ nodes,
	size_t num_leaves) {

	size_t leaf = static_cast<size_t>(blockIdx.x) * blockDim.x + threadIdx.x;
	if(leaf >= num_leaves) return;

	uint64_t state[Params::LIMBS];
	gnark_gpu::plonk2::zero<Params>(state);

	size_t base = leaf * chunk_size;
	for(size_t j = 0; j < chunk_size; j++) {
		uint64_t msg[Params::LIMBS];
		load_aos(msg, col_hashes, base + j);
		mimc_absorb(state, msg, constants);
	}

	store_aos(nodes, leaf, state);
}

__global__ void parent_kernel(
	const uint64_t *__restrict__ nodes,
	size_t prev_offset,
	size_t next_offset,
	const uint64_t *__restrict__ constants,
	uint64_t *__restrict__ out_nodes,
	size_t num_parents) {

	size_t parent = static_cast<size_t>(blockIdx.x) * blockDim.x + threadIdx.x;
	if(parent >= num_parents) return;

	uint64_t state[Params::LIMBS], left[Params::LIMBS], right[Params::LIMBS];
	gnark_gpu::plonk2::zero<Params>(state);
	load_aos(left, nodes, prev_offset + 2 * parent);
	load_aos(right, nodes, prev_offset + 2 * parent + 1);
	mimc_absorb(state, left, constants);
	mimc_absorb(state, right, constants);
	store_aos(out_nodes, next_offset + parent, state);
}

__global__ void sis_mimc_leaf_kernel(
	const uint64_t *__restrict__ rows,
	const uint8_t *__restrict__ row_kinds,
	const uint64_t *__restrict__ row_constants,
	size_t num_rows,
	size_t num_cols,
	const uint64_t *__restrict__ ag,
	size_t num_polys,
	const uint64_t *__restrict__ twiddles,
	const uint64_t *__restrict__ twiddles_inv,
	const uint64_t *__restrict__ coset,
	const uint64_t *__restrict__ coset_inv,
	const uint64_t *__restrict__ cardinality_inv,
	const uint64_t *__restrict__ mimc_constants,
	uint64_t *__restrict__ out_col_hashes,
	uint64_t *__restrict__ out_nodes) {

	const size_t col = static_cast<size_t>(blockIdx.x);
	if(col >= num_cols || threadIdx.x >= SIS_THREADS) return;

	const int tid = threadIdx.x;
	__shared__ uint64_t k[SIS_DEGREE][Params::LIMBS];
	__shared__ uint64_t res[SIS_DEGREE][Params::LIMBS];
	__shared__ uint64_t raw_rows[4][Params::LIMBS];

	gnark_gpu::plonk2::zero<Params>(res[tid]);

	for(size_t poly = 0; poly < num_polys; poly++) {
		const int local_row = tid / SIS_LIMBS_PER_FIELD;
		const int local_limb = tid - local_row * SIS_LIMBS_PER_FIELD;
		const size_t row = poly * 4 + static_cast<size_t>(local_row);

		if(local_limb == 0) {
			if(row < num_rows) {
				uint64_t mont[Params::LIMBS];
				if(row_kinds[row] == ROW_KIND_CONSTANT) {
					load_aos(mont, row_constants, row);
				} else {
					load_aos(mont, rows, row * num_cols + col);
				}
				from_montgomery(raw_rows[local_row], mont);
			} else {
				gnark_gpu::plonk2::zero<Params>(raw_rows[local_row]);
			}
		}
		__syncthreads();

		const uint16_t l = limb16(raw_rows[local_row], local_limb);
		set_raw_u64(k[tid], static_cast<uint64_t>(l));
		__syncthreads();

		fft_dif_coset_64(k, twiddles, coset);

		uint64_t a[Params::LIMBS], prod[Params::LIMBS];
		load_aos(a, ag, poly * SIS_DEGREE + static_cast<size_t>(tid));
		gnark_gpu::plonk2::mul<Params>(prod, k[tid], a);
		gnark_gpu::plonk2::add<Params>(res[tid], res[tid], prod);
		__syncthreads();
	}

	fft_inverse_dit_coset_64(res, twiddles_inv, coset_inv, cardinality_inv);

	store_aos(out_col_hashes, col * SIS_DEGREE + static_cast<size_t>(tid), res[tid]);
	__syncthreads();

	if(tid == 0) {
		uint64_t state[Params::LIMBS];
		gnark_gpu::plonk2::zero<Params>(state);
		for(int j = 0; j < SIS_DEGREE; j++) {
			mimc_absorb(state, res[j], mimc_constants);
		}
		store_aos(out_nodes, col, state);
	}
}

__global__ void sis_mimc_leaf_kernel_tiled2(
	const uint64_t *__restrict__ rows,
	const uint8_t *__restrict__ row_kinds,
	const uint64_t *__restrict__ row_constants,
	size_t num_rows,
	size_t num_cols,
	const uint64_t *__restrict__ ag,
	size_t num_polys,
	const uint64_t *__restrict__ twiddles,
	const uint64_t *__restrict__ twiddles_inv,
	const uint64_t *__restrict__ coset,
	const uint64_t *__restrict__ coset_inv,
	const uint64_t *__restrict__ cardinality_inv,
	const uint64_t *__restrict__ mimc_constants,
	uint64_t *__restrict__ out_col_hashes,
	uint64_t *__restrict__ out_nodes) {

	const int tile = threadIdx.x / SIS_THREADS;
	const int tid = threadIdx.x - tile * SIS_THREADS;
	const size_t col = static_cast<size_t>(blockIdx.x) * SIS_COLS_PER_BLOCK +
	                   static_cast<size_t>(tile);
	if(tile >= SIS_COLS_PER_BLOCK || col >= num_cols) return;

	__shared__ uint64_t k[SIS_COLS_PER_BLOCK][SIS_DEGREE][Params::LIMBS];
	__shared__ uint64_t res[SIS_COLS_PER_BLOCK][SIS_DEGREE][Params::LIMBS];
	__shared__ uint64_t raw_rows[SIS_COLS_PER_BLOCK][4][Params::LIMBS];

	gnark_gpu::plonk2::zero<Params>(res[tile][tid]);

	for(size_t poly = 0; poly < num_polys; poly++) {
		const int local_row = tid / SIS_LIMBS_PER_FIELD;
		const int local_limb = tid - local_row * SIS_LIMBS_PER_FIELD;
		const size_t row = poly * 4 + static_cast<size_t>(local_row);

		if(local_limb == 0) {
			if(row < num_rows) {
				uint64_t mont[Params::LIMBS];
				if(row_kinds[row] == ROW_KIND_CONSTANT) {
					load_aos(mont, row_constants, row);
				} else {
					load_aos(mont, rows, row * num_cols + col);
				}
				from_montgomery(raw_rows[tile][local_row], mont);
			} else {
				gnark_gpu::plonk2::zero<Params>(raw_rows[tile][local_row]);
			}
		}
		__syncthreads();

		const uint16_t l = limb16(raw_rows[tile][local_row], local_limb);
		set_raw_u64(k[tile][tid], static_cast<uint64_t>(l));
		__syncthreads();

		fft_dif_coset_64_at(k[tile], twiddles, coset, tid);

		uint64_t a[Params::LIMBS], prod[Params::LIMBS];
		load_aos(a, ag, poly * SIS_DEGREE + static_cast<size_t>(tid));
		gnark_gpu::plonk2::mul<Params>(prod, k[tile][tid], a);
		gnark_gpu::plonk2::add<Params>(res[tile][tid], res[tile][tid], prod);
		__syncthreads();
	}

	fft_inverse_dit_coset_64_at(res[tile], twiddles_inv, coset_inv, cardinality_inv, tid);

	store_aos(out_col_hashes, col * SIS_DEGREE + static_cast<size_t>(tid), res[tile][tid]);
	__syncthreads();

	if(tid == 0) {
		uint64_t state[Params::LIMBS];
		gnark_gpu::plonk2::zero<Params>(state);
		for(int j = 0; j < SIS_DEGREE; j++) {
			mimc_absorb(state, res[tile][j], mimc_constants);
		}
		store_aos(out_nodes, col, state);
	}
}

bool is_power_of_two(size_t n) {
	return n != 0 && (n & (n - 1)) == 0;
}

} // namespace

extern "C" gnark_gpu_error_t gnark_gpu_bls12377_mimc_sis_tree(
	gnark_gpu_context_t ctx,
	const uint64_t *col_hashes,
	size_t num_leaves,
	size_t chunk_size,
	const uint64_t *constants,
	uint64_t *out_nodes) {

	if(ctx == nullptr || col_hashes == nullptr || constants == nullptr || out_nodes == nullptr) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}
	if(!is_power_of_two(num_leaves) || chunk_size == 0) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}

	const size_t input_elems = num_leaves * chunk_size;
	if(input_elems / chunk_size != num_leaves) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}
	const size_t total_nodes = 2 * num_leaves - 1;
	if(total_nodes < num_leaves) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}

	uint64_t *d_input = nullptr;
	uint64_t *d_constants = nullptr;
	uint64_t *d_nodes = nullptr;
	const bool timed = timing_enabled();
	const auto t_total_start = now_if(timed);
	auto t_phase = t_total_start;
	double malloc_ms = 0;
	double h2d_ms = 0;
	double static_h2d_ms = 0;
	double leaf_ms = 0;
	double tree_ms = 0;
	double d2h_ms = 0;

	auto cleanup = [&]() {
		if(d_input != nullptr) cudaFree(d_input);
		if(d_constants != nullptr) cudaFree(d_constants);
		if(d_nodes != nullptr) cudaFree(d_nodes);
	};

	cudaError_t err = cudaMalloc(&d_input, input_elems * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_constants, MIMC_ROUNDS * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_nodes, total_nodes * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		const auto t_now = Clock::now();
		malloc_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	err = cudaMemcpy(d_input, col_hashes, input_elems * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		const auto t_now = Clock::now();
		h2d_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}
	err = cudaMemcpy(d_constants, constants, MIMC_ROUNDS * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		const auto t_now = Clock::now();
		static_h2d_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	sis_leaf_kernel<<<grid(num_leaves), THREADS>>>(d_input, chunk_size, d_constants, d_nodes, num_leaves);
	err = cudaGetLastError();
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		err = cudaDeviceSynchronize();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		const auto t_now = Clock::now();
		leaf_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	size_t prev_offset = 0;
	size_t next_offset = num_leaves;
	size_t level_size = num_leaves;
	while(level_size > 1) {
		size_t parents = level_size / 2;
		parent_kernel<<<grid(parents), THREADS>>>(
			d_nodes, prev_offset, next_offset, d_constants, d_nodes, parents);
		err = cudaGetLastError();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		prev_offset = next_offset;
		next_offset += parents;
		level_size = parents;
	}
	if(timed) {
		err = cudaDeviceSynchronize();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		const auto t_now = Clock::now();
		tree_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	err = cudaMemcpy(out_nodes, d_nodes, total_nodes * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyDeviceToHost);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaDeviceSynchronize();
	if(timed) {
		const auto t_now = Clock::now();
		d2h_ms = elapsed_ms(t_phase, t_now);
		log_timing("mimc_tree", chunk_size, num_leaves, input_elems,
		           malloc_ms, h2d_ms, static_h2d_ms, leaf_ms, tree_ms, d2h_ms,
		           elapsed_ms(t_total_start, t_now));
	}
	cleanup();
	return check(err);
}

extern "C" gnark_gpu_error_t gnark_gpu_bls12377_sis_mimc_tree(
	gnark_gpu_context_t ctx,
	const uintptr_t *row_ptrs,
	const uint8_t *row_kinds,
	const uint64_t *row_constants,
	size_t num_rows,
	size_t num_cols,
	const uint64_t *ag,
	size_t num_polys,
	const uint64_t *twiddles,
	const uint64_t *twiddles_inv,
	const uint64_t *coset,
	const uint64_t *coset_inv,
	const uint64_t *cardinality_inv,
	const uint64_t *mimc_constants,
	uint64_t *out_col_hashes,
	uint64_t *out_nodes) {

	if(ctx == nullptr || row_ptrs == nullptr || row_kinds == nullptr ||
	   row_constants == nullptr || ag == nullptr || twiddles == nullptr ||
	   twiddles_inv == nullptr || coset == nullptr || coset_inv == nullptr ||
	   cardinality_inv == nullptr || mimc_constants == nullptr ||
	   out_col_hashes == nullptr || out_nodes == nullptr) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}
	if(num_rows == 0 || num_cols == 0 || !is_power_of_two(num_cols) || num_polys == 0) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}
	if(num_polys != (num_rows * SIS_LIMBS_PER_FIELD + SIS_DEGREE - 1) / SIS_DEGREE) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}

	const size_t row_elems = num_rows * num_cols;
	const size_t col_hash_elems = num_cols * SIS_DEGREE;
	const size_t total_nodes = 2 * num_cols - 1;
	if(row_elems / num_cols != num_rows || col_hash_elems / SIS_DEGREE != num_cols ||
	   total_nodes < num_cols) {
		return GNARK_GPU_ERROR_INVALID_ARG;
	}

	uint64_t *d_rows = nullptr;
	uint8_t *d_row_kinds = nullptr;
	uint64_t *d_row_constants = nullptr;
	uint64_t *d_ag = nullptr;
	uint64_t *d_twiddles = nullptr;
	uint64_t *d_twiddles_inv = nullptr;
	uint64_t *d_coset = nullptr;
	uint64_t *d_coset_inv = nullptr;
	uint64_t *d_cardinality_inv = nullptr;
	uint64_t *d_mimc_constants = nullptr;
	uint64_t *d_col_hashes = nullptr;
	uint64_t *d_nodes = nullptr;
	const bool timed = timing_enabled();
	const auto t_total_start = now_if(timed);
	auto t_phase = t_total_start;
	double malloc_ms = 0;
	double h2d_ms = 0;
	double static_h2d_ms = 0;
	double leaf_ms = 0;
	double tree_ms = 0;
	double d2h_ms = 0;

	auto cleanup = [&]() {
		if(d_rows != nullptr) cudaFree(d_rows);
		if(d_row_kinds != nullptr) cudaFree(d_row_kinds);
		if(d_row_constants != nullptr) cudaFree(d_row_constants);
		if(d_ag != nullptr) cudaFree(d_ag);
		if(d_twiddles != nullptr) cudaFree(d_twiddles);
		if(d_twiddles_inv != nullptr) cudaFree(d_twiddles_inv);
		if(d_coset != nullptr) cudaFree(d_coset);
		if(d_coset_inv != nullptr) cudaFree(d_coset_inv);
		if(d_cardinality_inv != nullptr) cudaFree(d_cardinality_inv);
		if(d_mimc_constants != nullptr) cudaFree(d_mimc_constants);
		if(d_col_hashes != nullptr) cudaFree(d_col_hashes);
		if(d_nodes != nullptr) cudaFree(d_nodes);
	};

	cudaError_t err = cudaMalloc(&d_rows, row_elems * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_row_kinds, num_rows * sizeof(uint8_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_row_constants, num_rows * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_ag, num_polys * SIS_DEGREE * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_twiddles, SIS_TWIDDLES_SIZE * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_twiddles_inv, SIS_TWIDDLES_SIZE * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_coset, SIS_DEGREE * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_coset_inv, SIS_DEGREE * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_cardinality_inv, Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_mimc_constants, MIMC_ROUNDS * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_col_hashes, col_hash_elems * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMalloc(&d_nodes, total_nodes * Params::LIMBS * sizeof(uint64_t));
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		const auto t_now = Clock::now();
		malloc_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	for(size_t row = 0; row < num_rows; row++) {
		if(row_kinds[row] != ROW_KIND_REGULAR && row_kinds[row] != ROW_KIND_CONSTANT) {
			cleanup();
			return GNARK_GPU_ERROR_INVALID_ARG;
		}
		if(row_kinds[row] != ROW_KIND_REGULAR) {
			continue;
		}
		const auto *src = reinterpret_cast<const uint64_t *>(row_ptrs[row]);
		if(src == nullptr) {
			cleanup();
			return GNARK_GPU_ERROR_INVALID_ARG;
		}
		err = cudaMemcpy(d_rows + row * num_cols * Params::LIMBS, src,
		                 num_cols * Params::LIMBS * sizeof(uint64_t),
		                 cudaMemcpyHostToDevice);
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
	}
	if(timed) {
		const auto t_now = Clock::now();
		h2d_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	err = cudaMemcpy(d_row_kinds, row_kinds, num_rows * sizeof(uint8_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_row_constants, row_constants, num_rows * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_ag, ag, num_polys * SIS_DEGREE * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_twiddles, twiddles, SIS_TWIDDLES_SIZE * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_twiddles_inv, twiddles_inv,
	                 SIS_TWIDDLES_SIZE * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_coset, coset, SIS_DEGREE * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_coset_inv, coset_inv, SIS_DEGREE * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_cardinality_inv, cardinality_inv, Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(d_mimc_constants, mimc_constants, MIMC_ROUNDS * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyHostToDevice);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		const auto t_now = Clock::now();
		static_h2d_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	if(num_cols % SIS_COLS_PER_BLOCK == 0) {
		sis_mimc_leaf_kernel_tiled2<<<static_cast<unsigned>(num_cols / SIS_COLS_PER_BLOCK),
		                               SIS_COLS_PER_BLOCK * SIS_THREADS>>>(
			d_rows, d_row_kinds, d_row_constants, num_rows, num_cols, d_ag, num_polys,
			d_twiddles, d_twiddles_inv, d_coset, d_coset_inv, d_cardinality_inv,
			d_mimc_constants, d_col_hashes, d_nodes);
	} else {
		sis_mimc_leaf_kernel<<<static_cast<unsigned>(num_cols), SIS_THREADS>>>(
			d_rows, d_row_kinds, d_row_constants, num_rows, num_cols, d_ag, num_polys,
			d_twiddles, d_twiddles_inv, d_coset, d_coset_inv, d_cardinality_inv,
			d_mimc_constants, d_col_hashes, d_nodes);
	}
	err = cudaGetLastError();
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	if(timed) {
		err = cudaDeviceSynchronize();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		const auto t_now = Clock::now();
		leaf_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	size_t prev_offset = 0;
	size_t next_offset = num_cols;
	size_t level_size = num_cols;
	while(level_size > 1) {
		size_t parents = level_size / 2;
		parent_kernel<<<grid(parents), THREADS>>>(
			d_nodes, prev_offset, next_offset, d_mimc_constants, d_nodes, parents);
		err = cudaGetLastError();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		prev_offset = next_offset;
		next_offset += parents;
		level_size = parents;
	}
	if(timed) {
		err = cudaDeviceSynchronize();
		if(err != cudaSuccess) {
			cleanup();
			return check(err);
		}
		const auto t_now = Clock::now();
		tree_ms = elapsed_ms(t_phase, t_now);
		t_phase = t_now;
	}

	err = cudaMemcpy(out_col_hashes, d_col_hashes,
	                 col_hash_elems * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyDeviceToHost);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaMemcpy(out_nodes, d_nodes, total_nodes * Params::LIMBS * sizeof(uint64_t),
	                 cudaMemcpyDeviceToHost);
	if(err != cudaSuccess) {
		cleanup();
		return check(err);
	}
	err = cudaDeviceSynchronize();
	if(timed) {
		const auto t_now = Clock::now();
		d2h_ms = elapsed_ms(t_phase, t_now);
		log_timing("sis_mimc_tree", num_rows, num_cols, row_elems,
		           malloc_ms, h2d_ms, static_h2d_ms, leaf_ms, tree_ms, d2h_ms,
		           elapsed_ms(t_total_start, t_now));
	}
	cleanup();
	return check(err);
}
