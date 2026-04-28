// Single-point validation kernels for gpu/plonk2.
//
// These are not used by the MSM. They exist only so the curve-generic G1
// formulas in ec.cuh can be tested against gnark-crypto from Go before the
// MSM batches them up.

#include "ec.cuh"

#include <cuda_runtime.h>

namespace gnark_gpu::plonk2 {

namespace {

template <typename Fp>
__device__ __forceinline__ void load_affine(AffinePoint<Fp> &p, const uint64_t *raw) {
#pragma unroll
	for(int i = 0; i < Fp::LIMBS; i++) {
		p.x[i] = raw[i];
		p.y[i] = raw[Fp::LIMBS + i];
	}
}

template <typename Fp>
__device__ __forceinline__ void store_jacobian(const JacobianPoint<Fp> &p, uint64_t *raw) {
#pragma unroll
	for(int i = 0; i < Fp::LIMBS; i++) {
		raw[i] = p.x[i];
		raw[Fp::LIMBS + i] = p.y[i];
		raw[2 * Fp::LIMBS + i] = p.z[i];
	}
}

template <typename Fp>
__device__ __forceinline__ void load_affine_at(AffinePoint<Fp> &p, const uint64_t *raw, size_t idx) {
	const uint64_t *point = raw + idx * (2 * Fp::LIMBS);
	load_affine<Fp>(p, point);
}

template <typename Fr>
__device__ __forceinline__ bool scalar_bit(const uint64_t *scalars, size_t idx, int bit) {
	const uint64_t *scalar = scalars + idx * Fr::LIMBS;
	int limb = bit / 64;
	if(limb >= Fr::LIMBS) return false;
	return ((scalar[limb] >> (bit & 63)) & 1ULL) != 0;
}

template <typename Fp, typename Fr>
__device__ __forceinline__ void scalar_mul_affine(
	JacobianPoint<Fp> &out, const AffinePoint<Fp> &point, const uint64_t *scalars, size_t idx) {

	JacobianPoint<Fp> acc, base, tmp;
	jacobian_set_infinity<Fp>(acc);
	jacobian_from_affine<Fp>(base, point);

	for(int bit = 0; bit < Fr::BITS; bit++) {
		if(scalar_bit<Fr>(scalars, idx, bit)) {
			jacobian_add<Fp>(tmp, acc, base);
			set<Fp>(acc.x, tmp.x);
			set<Fp>(acc.y, tmp.y);
			set<Fp>(acc.z, tmp.z);
		}
		jacobian_double<Fp>(tmp, base);
		set<Fp>(base.x, tmp.x);
		set<Fp>(base.y, tmp.y);
		set<Fp>(base.z, tmp.z);
	}

	set<Fp>(out.x, acc.x);
	set<Fp>(out.y, acc.y);
	set<Fp>(out.z, acc.z);
}

template <typename Fp>
__global__ void g1_affine_add_kernel(const uint64_t *p_raw, const uint64_t *q_raw, uint64_t *out_raw) {
	AffinePoint<Fp> p, q;
	JacobianPoint<Fp> out;
	load_affine<Fp>(p, p_raw);
	load_affine<Fp>(q, q_raw);
	jacobian_add_affine_affine<Fp>(out, p, q);
	store_jacobian<Fp>(out, out_raw);
}

template <typename Fp>
__global__ void g1_affine_double_kernel(const uint64_t *p_raw, uint64_t *out_raw) {
	AffinePoint<Fp> p;
	JacobianPoint<Fp> out;
	load_affine<Fp>(p, p_raw);
	jacobian_double_mixed<Fp>(out, p);
	store_jacobian<Fp>(out, out_raw);
}

template <typename Fp, typename Fr>
__global__ void msm_naive_kernel(
	const uint64_t *points, const uint64_t *scalars, size_t count, uint64_t *out_raw) {

	JacobianPoint<Fp> acc, term, tmp;
	jacobian_set_infinity<Fp>(acc);

	for(size_t i = 0; i < count; i++) {
		AffinePoint<Fp> p;
		load_affine_at<Fp>(p, points, i);
		scalar_mul_affine<Fp, Fr>(term, p, scalars, i);
		jacobian_add<Fp>(tmp, acc, term);
		set<Fp>(acc.x, tmp.x);
		set<Fp>(acc.y, tmp.y);
		set<Fp>(acc.z, tmp.z);
	}

	store_jacobian<Fp>(acc, out_raw);
}

template <typename Fp>
cudaError_t run_g1_add(const uint64_t *p, const uint64_t *q, uint64_t *out, cudaStream_t stream) {
	uint64_t *d_p = nullptr, *d_q = nullptr, *d_out = nullptr;
	constexpr size_t input_words = 2 * Fp::LIMBS;
	constexpr size_t output_words = 3 * Fp::LIMBS;

	cudaError_t err = cudaMalloc(&d_p, input_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_q, input_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_out, output_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(d_p, p, input_words * sizeof(uint64_t), cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;
	err = cudaMemcpyAsync(d_q, q, input_words * sizeof(uint64_t), cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;

	g1_affine_add_kernel<Fp><<<1, 1, 0, stream>>>(d_p, d_q, d_out);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(out, d_out, output_words * sizeof(uint64_t), cudaMemcpyDeviceToHost, stream);
	if(err != cudaSuccess) goto done;
	err = cudaStreamSynchronize(stream);

done:
	if(d_p) cudaFree(d_p);
	if(d_q) cudaFree(d_q);
	if(d_out) cudaFree(d_out);
	return err;
}

template <typename Fp>
cudaError_t run_g1_double(const uint64_t *p, uint64_t *out, cudaStream_t stream) {
	uint64_t *d_p = nullptr, *d_out = nullptr;
	constexpr size_t input_words = 2 * Fp::LIMBS;
	constexpr size_t output_words = 3 * Fp::LIMBS;

	cudaError_t err = cudaMalloc(&d_p, input_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_out, output_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(d_p, p, input_words * sizeof(uint64_t), cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;

	g1_affine_double_kernel<Fp><<<1, 1, 0, stream>>>(d_p, d_out);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(out, d_out, output_words * sizeof(uint64_t), cudaMemcpyDeviceToHost, stream);
	if(err != cudaSuccess) goto done;
	err = cudaStreamSynchronize(stream);

done:
	if(d_p) cudaFree(d_p);
	if(d_out) cudaFree(d_out);
	return err;
}

} // namespace

cudaError_t g1_affine_add_run(
	gnark_gpu_plonk2_curve_id_t curve, const uint64_t *p,
	const uint64_t *q, uint64_t *out, cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_g1_add<BN254FpParams>(p, q, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_g1_add<BLS12377FpParams>(p, q, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_g1_add<BW6761FpParams>(p, q, out, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

cudaError_t g1_affine_double_run(
	gnark_gpu_plonk2_curve_id_t curve, const uint64_t *p,
	uint64_t *out, cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_g1_double<BN254FpParams>(p, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_g1_double<BLS12377FpParams>(p, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_g1_double<BW6761FpParams>(p, out, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

template <typename Fp, typename Fr>
cudaError_t run_msm_naive(
	const uint64_t *points, const uint64_t *scalars, size_t count,
	uint64_t *out, cudaStream_t stream) {

	uint64_t *d_points = nullptr, *d_scalars = nullptr, *d_out = nullptr;
	const size_t point_words = count * 2 * Fp::LIMBS;
	const size_t scalar_words = count * Fr::LIMBS;
	constexpr size_t output_words = 3 * Fp::LIMBS;

	cudaError_t err = cudaMalloc(&d_points, point_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_scalars, scalar_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;
	err = cudaMalloc(&d_out, output_words * sizeof(uint64_t));
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(d_points, points, point_words * sizeof(uint64_t),
	                      cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;
	err = cudaMemcpyAsync(d_scalars, scalars, scalar_words * sizeof(uint64_t),
	                      cudaMemcpyHostToDevice, stream);
	if(err != cudaSuccess) goto done;

	msm_naive_kernel<Fp, Fr><<<1, 1, 0, stream>>>(d_points, d_scalars, count, d_out);
	err = cudaGetLastError();
	if(err != cudaSuccess) goto done;

	err = cudaMemcpyAsync(out, d_out, output_words * sizeof(uint64_t),
	                      cudaMemcpyDeviceToHost, stream);
	if(err != cudaSuccess) goto done;
	err = cudaStreamSynchronize(stream);

done:
	if(d_points) cudaFree(d_points);
	if(d_scalars) cudaFree(d_scalars);
	if(d_out) cudaFree(d_out);
	return err;
}

cudaError_t msm_naive_run(
	gnark_gpu_plonk2_curve_id_t curve, const uint64_t *points,
	const uint64_t *scalars, size_t count, uint64_t *out,
	cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return run_msm_naive<BN254FpParams, BN254FrParams>(
			points, scalars, count, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return run_msm_naive<BLS12377FpParams, BLS12377FrParams>(
			points, scalars, count, out, stream);
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return run_msm_naive<BW6761FpParams, BW6761FrParams>(
			points, scalars, count, out, stream);
	default:
		return cudaErrorInvalidValue;
	}
}

} // namespace gnark_gpu::plonk2
