#include "bench.cuh"
#include "../src/plonk/field.cuh"
#include "../src/plonk/ec.cuh"
#include <cstdio>
#include <cstdlib>
#include <cstring>

using namespace gnark_gpu;

// Forward declarations for MSM functions (defined in msm.cu, linked via libgnark_gpu)
namespace gnark_gpu {
struct MSMContext;
MSMContext *msm_create(size_t max_points);
void msm_destroy(MSMContext *ctx);
void msm_load_points(MSMContext *ctx, const void *host_points, size_t count, cudaStream_t stream);
cudaError_t msm_alloc_work_buffers(MSMContext *ctx);
void msm_upload_scalars(MSMContext *ctx, const uint64_t *host_scalars, size_t n, cudaStream_t stream);
void launch_msm(MSMContext *ctx, size_t n, cudaStream_t stream);
void msm_download_results(MSMContext *ctx, G1EdExtended *host_results, cudaStream_t stream);

// Fr operation launchers (defined in fr_ops.cu)
void launch_scale_by_powers(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                            const uint64_t g[4], size_t n, cudaStream_t stream);
cudaError_t launch_batch_invert(uint64_t *v0, uint64_t *v1, uint64_t *v2, uint64_t *v3,
                                uint64_t *orig0, uint64_t *orig1, uint64_t *orig2, uint64_t *orig3,
                                size_t n, cudaStream_t stream);
}

// =============================================================================
// Kernel launcher helpers for field benchmarks
// =============================================================================

__global__ void mul_mont_fp(const uint64_t *__restrict__ a0, const uint64_t *__restrict__ a1,
							const uint64_t *__restrict__ a2, const uint64_t *__restrict__ a3,
							const uint64_t *__restrict__ a4, const uint64_t *__restrict__ a5,
							const uint64_t *__restrict__ b0, const uint64_t *__restrict__ b1,
							const uint64_t *__restrict__ b2, const uint64_t *__restrict__ b3,
							const uint64_t *__restrict__ b4, const uint64_t *__restrict__ b5, uint64_t *__restrict__ c0,
							uint64_t *__restrict__ c1, uint64_t *__restrict__ c2, uint64_t *__restrict__ c3,
							uint64_t *__restrict__ c4, uint64_t *__restrict__ c5, size_t n) {
	auto idx = blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t a[6] = {a0[idx], a1[idx], a2[idx], a3[idx], a4[idx], a5[idx]};
	uint64_t b[6] = {b0[idx], b1[idx], b2[idx], b3[idx], b4[idx], b5[idx]};
	uint64_t c[6];
	fp_mul(c, a, b);
	c0[idx] = c[0]; c1[idx] = c[1]; c2[idx] = c[2];
	c3[idx] = c[3]; c4[idx] = c[4]; c5[idx] = c[5];
}

void launch_mul_mont(FpVector &c, const FpVector &a, const FpVector &b) {
	auto n = a.size();
	auto threads = 1024u;
	auto blocks = (n + threads - 1) / threads;

	mul_mont_fp<<<blocks, threads>>>(a.limb(0), a.limb(1), a.limb(2), a.limb(3), a.limb(4), a.limb(5), b.limb(0),
									 b.limb(1), b.limb(2), b.limb(3), b.limb(4), b.limb(5), c.limb(0), c.limb(1),
									 c.limb(2), c.limb(3), c.limb(4), c.limb(5), n);
}

// =============================================================================
// Field benchmarks
// =============================================================================

void bench_fp_mul(size_t n) {
	printf("Fp Montgomery multiplication (n=%zu)\n", n);
	printf("-----------------------------------------\n");

	HostFpVector h_a(n), h_b(n);
	for(size_t i = 0; i < n; ++i) {
		for(size_t j = 0; j < Fp_params::LIMBS; ++j) {
			h_a.limb(j)[i] = i + j;
			h_b.limb(j)[i] = i + j + 2;
		}
	}

	FpVector d_a(n), d_b(n), d_c(n);
	d_a.copy_host_to_device(h_a.raw_ptrs());
	d_b.copy_host_to_device(h_b.raw_ptrs());

	auto bytes = d_a.bytes() * 3;
	bench("Fp mul_mont", 100, 10, bytes, [&]() { launch_mul_mont(d_c, d_a, d_b); });
	printf("\n");
}

// =============================================================================
// MSM benchmark
// =============================================================================

void fill_random_te_points(G1EdXY *points, size_t n) {
	uint64_t state = 0xdeadbeefcafe1234ULL;
	for (size_t i = 0; i < n; i++) {
		for (int j = 0; j < 6; j++) {
			state ^= state << 13; state ^= state >> 7; state ^= state << 17;
			points[i].x[j] = state;
		}
		for (int j = 0; j < 6; j++) {
			state ^= state << 13; state ^= state >> 7; state ^= state << 17;
			points[i].y[j] = state;
		}
	}
}

void fill_random_scalars(uint64_t *scalars, size_t n) {
	uint64_t state = 0x123456789abcdef0ULL;
	for (size_t i = 0; i < n * 4; i++) {
		state ^= state << 13; state ^= state >> 7; state ^= state << 17;
		scalars[i] = state;
	}
}

void bench_msm(size_t n) {
	printf("MSM (Pippenger/TE, n=%zu)\n", n);
	printf("-----------------------------------------\n");

	G1EdXY *h_points = (G1EdXY *)malloc(n * sizeof(G1EdXY));
	uint64_t *h_scalars = (uint64_t *)malloc(n * 4 * sizeof(uint64_t));
	fill_random_te_points(h_points, n);
	fill_random_scalars(h_scalars, n);

	cudaStream_t stream;
	cudaStreamCreate(&stream);

	MSMContext *ctx = msm_create(n);
	msm_load_points(ctx, h_points, n, stream);
	msm_alloc_work_buffers(ctx);
	cudaStreamSynchronize(stream);

	G1EdExtended h_results[24]; // must be >= max num_windows (c=13 → 20)

	msm_upload_scalars(ctx, h_scalars, n, stream);
	cudaStreamSynchronize(stream);

	// Warmup
	for (int i = 0; i < 3; i++) {
		launch_msm(ctx, n, stream);
		cudaStreamSynchronize(stream);
	}

	// Timed kernel breakdown
	launch_msm(ctx, n, stream);

	// Benchmark full MSM pipeline
	bench("MSM total", 3, 5, 0, [&]() {
		msm_upload_scalars(ctx, h_scalars, n, stream);
		launch_msm(ctx, n, stream);
		msm_download_results(ctx, h_results, stream);
		cudaStreamSynchronize(stream);
	});

	// Benchmark just the GPU kernels (no upload/download)
	bench("MSM kernels only", 3, 5, 0, [&]() {
		launch_msm(ctx, n, stream);
		cudaStreamSynchronize(stream);
	});

	msm_destroy(ctx);
	cudaStreamDestroy(stream);
	free(h_points);
	free(h_scalars);

	printf("\n");
}

// =============================================================================
// Fr operation benchmarks (PLONK-relevant kernels)
// =============================================================================

void fill_random_fr_limbs(HostFrVector &h) {
	uint64_t state = 0x8f3a4d0b7c12e9f1ULL;
	for(size_t i = 0; i < h.size(); i++) {
		for(size_t limb = 0; limb < Fr_params::LIMBS; limb++) {
			state ^= state << 13;
			state ^= state >> 7;
			state ^= state << 17;
			h.limb(limb)[i] = state;
		}
		// Avoid zero for batch inversion.
		h.limb(0)[i] |= 1ULL;
	}
}

void bench_fr_ops(size_t n) {
	printf("Fr Operation Benchmarks (n=%zu)\n", n);
	printf("-----------------------------------------\n");

	HostFrVector h_v(n);
	fill_random_fr_limbs(h_v);

	FrVector d_v(n), d_tmp(n);
	d_v.copy_host_to_device(h_v.raw_ptrs());

	// Use a non-trivial generator in Montgomery form.
	const uint64_t g[4] = {
		0x2f3f7d4c1a9b5e81ULL,
		0x1c0de5f1b39a77d2ULL,
		0x0f6b8a4d239dbe13ULL,
		0x048e72ab91c3d5f7ULL,
	};

	// Approximate bytes moved by one kernel invocation.
	size_t bytes_scale = n * 8 * sizeof(uint64_t); // 4 limbs read + 4 limbs write
	bench("Fr scale_by_powers", 20, 20, bytes_scale, [&]() {
		launch_scale_by_powers(
			d_v.limb(0), d_v.limb(1), d_v.limb(2), d_v.limb(3),
			g, n, 0);
	});

	// BatchInvert mutates input. Run twice per iteration so input returns to original.
	// This avoids per-iteration H2D/D2D reset overhead in the benchmark loop.
	size_t bytes_batch_inv = n * 8 * sizeof(uint64_t);
	bench("Fr batch_invert x2", 5, 10, bytes_batch_inv, [&]() {
		launch_batch_invert(
			d_v.limb(0), d_v.limb(1), d_v.limb(2), d_v.limb(3),
			d_tmp.limb(0), d_tmp.limb(1), d_tmp.limb(2), d_tmp.limb(3),
			n, 0);
		launch_batch_invert(
			d_v.limb(0), d_v.limb(1), d_v.limb(2), d_v.limb(3),
			d_tmp.limb(0), d_tmp.limb(1), d_tmp.limb(2), d_tmp.limb(3),
			n, 0);
	});

	printf("\n");
}

// =============================================================================
// Main
// =============================================================================

int main(int argc, char **argv) {
	printf("gnark-gpu BLS12-377 Profiling Sandbox\n");
	printf("=====================================\n\n");

	bool run_fp = true;
	bool run_msm = true;
	bool run_fr = false;
	size_t msm_n = 1 << 20; // 1M points default
	size_t fr_n = 1 << 23;  // ~8M elements default

	for (int i = 1; i < argc; i++) {
		if (strcmp(argv[i], "--msm-only") == 0) { run_fp = false; }
		if (strcmp(argv[i], "--fp-only") == 0) { run_msm = false; }
		if (strcmp(argv[i], "--fr-only") == 0) { run_fp = false; run_msm = false; run_fr = true; }
		if (strcmp(argv[i], "--fr") == 0) { run_fr = true; }
		if (strncmp(argv[i], "--n=", 4) == 0) { msm_n = atol(argv[i] + 4); }
		if (strncmp(argv[i], "--fr-n=", 7) == 0) { fr_n = atol(argv[i] + 7); }
	}

	if (run_fp) {
		printf("Field Benchmarks:\n");
		printf("=================\n\n");
		bench_fp_mul(10'000'000);
	}

	if (run_msm) {
		printf("MSM Benchmarks:\n");
		printf("===============\n\n");
		bench_msm(msm_n);
	}

	if (run_fr) {
		printf("Fr Benchmarks:\n");
		printf("==============\n\n");
		bench_fr_ops(fr_n);
	}

	return 0;
}
