#include "field.cuh"

#include <cuda_runtime.h>

namespace gnark_gpu::plonk2 {

namespace {

constexpr unsigned THREADS = 256;
constexpr unsigned NTT_THREADS = 256;
constexpr size_t Z_PREFIX_CHUNK_SIZE = 1024;
constexpr uint32_t NTT_FUSED_TAIL_MIN_N = 1u << 22;

struct ScalarArg {
	uint64_t limbs[MAX_FR_LIMBS];
};

ScalarArg make_scalar_arg(gnark_gpu_plonk2_curve_id_t curve, const uint64_t *limbs) {
	ScalarArg out{};
	int n = curve_limbs(curve);
	for(int i = 0; i < n; i++) out.limbs[i] = limbs[i];
	return out;
}

template <typename Params>
__global__ void copy_aos_to_soa_kernel(FrView dst, const uint64_t *__restrict__ src, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	const uint64_t *in = src + idx * Params::LIMBS;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		dst.limbs[i][idx] = in[i];
	}
}

template <typename Params>
__global__ void copy_soa_to_aos_kernel(uint64_t *__restrict__ dst, ConstFrView src, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t *out = dst + idx * Params::LIMBS;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		out[i] = src.limbs[i][idx];
	}
}

template <typename Params>
__global__ void set_zero_kernel(FrView v, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		v.limbs[i][idx] = 0;
	}
}

template <typename Params>
__global__ void add_kernel(FrView out, ConstFrView a, ConstFrView b, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t av[Params::LIMBS], bv[Params::LIMBS], rv[Params::LIMBS];
	load<Params>(av, a, idx);
	load<Params>(bv, b, idx);
	add<Params>(rv, av, bv);
	store<Params>(out, idx, rv);
}

template <typename Params>
__global__ void sub_kernel(FrView out, ConstFrView a, ConstFrView b, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t av[Params::LIMBS], bv[Params::LIMBS], rv[Params::LIMBS];
	load<Params>(av, a, idx);
	load<Params>(bv, b, idx);
	sub<Params>(rv, av, bv);
	store<Params>(out, idx, rv);
}

template <typename Params>
__global__ void mul_kernel(FrView out, ConstFrView a, ConstFrView b, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t av[Params::LIMBS], bv[Params::LIMBS], rv[Params::LIMBS];
	load<Params>(av, a, idx);
	load<Params>(bv, b, idx);
	mul<Params>(rv, av, bv);
	store<Params>(out, idx, rv);
}

template <typename Params>
__global__ void addmul_kernel(FrView out, ConstFrView a, ConstFrView b, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t ov[Params::LIMBS], av[Params::LIMBS], bv[Params::LIMBS];
	uint64_t prod[Params::LIMBS], rv[Params::LIMBS];
	load<Params>(ov, make_const(out), idx);
	load<Params>(av, a, idx);
	load<Params>(bv, b, idx);
	mul<Params>(prod, av, bv);
	add<Params>(rv, ov, prod);
	store<Params>(out, idx, rv);
}

template <typename Params>
__global__ void scalar_mul_kernel(FrView out, ScalarArg scalar_arg, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t scalar[Params::LIMBS], ov[Params::LIMBS], rv[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) scalar[i] = scalar_arg.limbs[i];
	load<Params>(ov, make_const(out), idx);
	mul<Params>(rv, ov, scalar);
	store<Params>(out, idx, rv);
}

template <typename Params>
__global__ void add_scalar_mul_kernel(
	FrView out, ConstFrView a, ScalarArg scalar_arg, size_t n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t scalar[Params::LIMBS], ov[Params::LIMBS], av[Params::LIMBS];
	uint64_t prod[Params::LIMBS], rv[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) scalar[i] = scalar_arg.limbs[i];
	load<Params>(ov, make_const(out), idx);
	load<Params>(av, a, idx);
	mul<Params>(prod, av, scalar);
	add<Params>(rv, ov, prod);
	store<Params>(out, idx, rv);
}

template <typename Params>
__device__ __forceinline__ bool modulus_minus_two_bit(int bit) {
	uint64_t limb = modulus_limb<Params>(bit / 64);
	if(bit < 64) limb -= 2;
	return ((limb >> (bit & 63)) & 1ULL) != 0;
}

template <typename Params>
__device__ __forceinline__ void inverse_pow(uint64_t out[Params::LIMBS],
                                            const uint64_t in[Params::LIMBS]) {
	uint64_t acc[Params::LIMBS], factor[Params::LIMBS];
	one<Params>(acc);
	set<Params>(factor, in);

	for(int bit = 0; bit < Params::BITS; bit++) {
		if(modulus_minus_two_bit<Params>(bit)) {
			mul<Params>(acc, acc, factor);
		}
		if(bit + 1 < Params::BITS) {
			square<Params>(factor, factor);
		}
	}
	set<Params>(out, acc);
}

template <typename Params>
__global__ void invert_kernel(FrView data, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t value[Params::LIMBS], inv[Params::LIMBS];
	load<Params>(value, make_const(data), idx);
	inverse_pow<Params>(inv, value);
	store<Params>(data, idx, inv);
}

template <typename Params>
__global__ void butterfly4_inverse_kernel(
	FrView b0, FrView b1, FrView b2, FrView b3,
	ScalarArg omega4_inv_arg, ScalarArg quarter_arg, size_t n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t omega4_inv[Params::LIMBS], quarter[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		omega4_inv[i] = omega4_inv_arg.limbs[i];
		quarter[i] = quarter_arg.limbs[i];
	}

	uint64_t v0[Params::LIMBS], v1[Params::LIMBS], v2[Params::LIMBS], v3[Params::LIMBS];
	load<Params>(v0, make_const(b0), idx);
	load<Params>(v1, make_const(b1), idx);
	load<Params>(v2, make_const(b2), idx);
	load<Params>(v3, make_const(b3), idx);

	uint64_t t0[Params::LIMBS], t1[Params::LIMBS], t2[Params::LIMBS], t3[Params::LIMBS];
	uint64_t u0[Params::LIMBS], u1[Params::LIMBS], u2[Params::LIMBS], u3[Params::LIMBS];
	add<Params>(t0, v0, v2);
	sub<Params>(t1, v0, v2);
	add<Params>(t2, v1, v3);
	sub<Params>(t3, v1, v3);
	mul<Params>(t3, t3, omega4_inv);

	add<Params>(u0, t0, t2);
	add<Params>(u1, t1, t3);
	sub<Params>(u2, t0, t2);
	sub<Params>(u3, t1, t3);

	mul<Params>(u0, u0, quarter);
	mul<Params>(u1, u1, quarter);
	mul<Params>(u2, u2, quarter);
	mul<Params>(u3, u3, quarter);

	store<Params>(b0, idx, u0);
	store<Params>(b1, idx, u1);
	store<Params>(b2, idx, u2);
	store<Params>(b3, idx, u3);
}

template <typename Params>
__global__ void reduce_blinded_coset_kernel(
	FrView dst, ConstFrView src, const uint64_t *tail,
	ScalarArg coset_pow_n_arg, size_t n, size_t tail_len) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t value[Params::LIMBS];
	load<Params>(value, src, idx);
	if(idx < tail_len) {
		uint64_t tail_value[Params::LIMBS], coset_pow_n[Params::LIMBS], scaled[Params::LIMBS];
#pragma unroll
		for(int i = 0; i < Params::LIMBS; i++) {
			tail_value[i] = __ldg(tail + idx * Params::LIMBS + i);
			coset_pow_n[i] = coset_pow_n_arg.limbs[i];
		}
		mul<Params>(scaled, tail_value, coset_pow_n);
		add<Params>(value, value, scaled);
	}
	store<Params>(dst, idx, value);
}

template <typename Params>
__global__ void compute_l1_den_kernel(
	FrView out, ConstFrView twiddles, ScalarArg coset_gen_arg, size_t n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t coset_gen[Params::LIMBS], omega_i[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) coset_gen[i] = coset_gen_arg.limbs[i];

	size_t half_n = n >> 1;
	if(idx < half_n) {
		load<Params>(omega_i, twiddles, idx);
	} else {
		uint64_t positive[Params::LIMBS], zero_value[Params::LIMBS];
		load<Params>(positive, twiddles, idx - half_n);
		zero<Params>(zero_value);
		sub<Params>(omega_i, zero_value, positive);
	}

	uint64_t product[Params::LIMBS], one_value[Params::LIMBS], value[Params::LIMBS];
	mul<Params>(product, coset_gen, omega_i);
	one<Params>(one_value);
	sub<Params>(value, product, one_value);
	store<Params>(out, idx, value);
}

template <typename Params>
__global__ void gate_accum_kernel(
	FrView result,
	ConstFrView ql, ConstFrView qr, ConstFrView qm, ConstFrView qo, ConstFrView qk,
	ConstFrView l, ConstFrView r, ConstFrView o,
	ScalarArg zh_k_inv_arg, size_t n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t acc[Params::LIMBS], zh_k_inv[Params::LIMBS];
	uint64_t l_value[Params::LIMBS], r_value[Params::LIMBS], o_value[Params::LIMBS];
	uint64_t q_value[Params::LIMBS], tmp[Params::LIMBS], lr[Params::LIMBS];

#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		zh_k_inv[i] = zh_k_inv_arg.limbs[i];
	}

	load<Params>(acc, make_const(result), idx);
	load<Params>(l_value, l, idx);
	load<Params>(r_value, r, idx);
	load<Params>(o_value, o, idx);

	load<Params>(q_value, ql, idx);
	mul<Params>(tmp, q_value, l_value);
	add<Params>(acc, acc, tmp);

	load<Params>(q_value, qr, idx);
	mul<Params>(tmp, q_value, r_value);
	add<Params>(acc, acc, tmp);

	load<Params>(q_value, qm, idx);
	mul<Params>(lr, l_value, r_value);
	mul<Params>(tmp, q_value, lr);
	add<Params>(acc, acc, tmp);

	load<Params>(q_value, qo, idx);
	mul<Params>(tmp, q_value, o_value);
	add<Params>(acc, acc, tmp);

	load<Params>(q_value, qk, idx);
	add<Params>(acc, acc, q_value);

	mul<Params>(acc, acc, zh_k_inv);
	store<Params>(result, idx, acc);
}

template <typename Params>
__device__ __forceinline__ void omega_from_twiddles(
	uint64_t out[Params::LIMBS], ConstFrView twiddles, size_t idx, size_t half_n) {

	if(idx < half_n) {
		load<Params>(out, twiddles, idx);
		return;
	}

	uint64_t positive[Params::LIMBS], zero_value[Params::LIMBS];
	load<Params>(positive, twiddles, idx - half_n);
	zero<Params>(zero_value);
	sub<Params>(out, zero_value, positive);
}

template <typename Params>
__global__ void perm_boundary_kernel(
	FrView result,
	ConstFrView l, ConstFrView r, ConstFrView o, ConstFrView z,
	ConstFrView s1, ConstFrView s2, ConstFrView s3, ConstFrView l1_den_inv,
	ConstFrView twiddles,
	ScalarArg alpha_arg, ScalarArg beta_arg, ScalarArg gamma_arg,
	ScalarArg l1_scalar_arg, ScalarArg coset_shift_arg,
	ScalarArg coset_shift_sq_arg, ScalarArg coset_gen_arg,
	size_t n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t alpha[Params::LIMBS], beta[Params::LIMBS], gamma[Params::LIMBS];
	uint64_t l1_scalar[Params::LIMBS], coset_shift[Params::LIMBS];
	uint64_t coset_shift_sq[Params::LIMBS], coset_gen[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		alpha[i] = alpha_arg.limbs[i];
		beta[i] = beta_arg.limbs[i];
		gamma[i] = gamma_arg.limbs[i];
		l1_scalar[i] = l1_scalar_arg.limbs[i];
		coset_shift[i] = coset_shift_arg.limbs[i];
		coset_shift_sq[i] = coset_shift_sq_arg.limbs[i];
		coset_gen[i] = coset_gen_arg.limbs[i];
	}

	uint64_t l_value[Params::LIMBS], r_value[Params::LIMBS], o_value[Params::LIMBS];
	uint64_t z_value[Params::LIMBS], z_next[Params::LIMBS];
	load<Params>(l_value, l, idx);
	load<Params>(r_value, r, idx);
	load<Params>(o_value, o, idx);
	load<Params>(z_value, z, idx);
	load<Params>(z_next, z, idx + 1 < n ? idx + 1 : 0);

	uint64_t omega_i[Params::LIMBS], x_i[Params::LIMBS];
	omega_from_twiddles<Params>(omega_i, twiddles, idx, n >> 1);
	mul<Params>(x_i, coset_gen, omega_i);

	uint64_t id1[Params::LIMBS], id2[Params::LIMBS], id3[Params::LIMBS];
	mul<Params>(id1, beta, x_i);
	mul<Params>(id2, id1, coset_shift);
	mul<Params>(id3, id1, coset_shift_sq);

	uint64_t t1[Params::LIMBS], t2[Params::LIMBS], t3[Params::LIMBS];
	add<Params>(t1, l_value, id1);
	add<Params>(t1, t1, gamma);
	add<Params>(t2, r_value, id2);
	add<Params>(t2, t2, gamma);
	add<Params>(t3, o_value, id3);
	add<Params>(t3, t3, gamma);

	uint64_t num[Params::LIMBS], tmp[Params::LIMBS];
	mul<Params>(num, z_value, t1);
	mul<Params>(tmp, num, t2);
	mul<Params>(num, tmp, t3);

	uint64_t s_value[Params::LIMBS], beta_s[Params::LIMBS];
	load<Params>(s_value, s1, idx);
	mul<Params>(beta_s, beta, s_value);
	add<Params>(t1, l_value, beta_s);
	add<Params>(t1, t1, gamma);

	load<Params>(s_value, s2, idx);
	mul<Params>(beta_s, beta, s_value);
	add<Params>(t2, r_value, beta_s);
	add<Params>(t2, t2, gamma);

	load<Params>(s_value, s3, idx);
	mul<Params>(beta_s, beta, s_value);
	add<Params>(t3, o_value, beta_s);
	add<Params>(t3, t3, gamma);

	uint64_t den[Params::LIMBS];
	mul<Params>(den, z_next, t1);
	mul<Params>(tmp, den, t2);
	mul<Params>(den, tmp, t3);

	uint64_t ordering[Params::LIMBS];
	sub<Params>(ordering, den, num);

	uint64_t l1_den_inv_value[Params::LIMBS], l1_value[Params::LIMBS];
	load<Params>(l1_den_inv_value, l1_den_inv, idx);
	mul<Params>(l1_value, l1_scalar, l1_den_inv_value);

	uint64_t one_value[Params::LIMBS], z_minus_one[Params::LIMBS], local[Params::LIMBS];
	one<Params>(one_value);
	sub<Params>(z_minus_one, z_value, one_value);
	mul<Params>(local, z_minus_one, l1_value);

	uint64_t alpha_local[Params::LIMBS], sum[Params::LIMBS], out[Params::LIMBS];
	mul<Params>(alpha_local, alpha, local);
	add<Params>(sum, ordering, alpha_local);
	mul<Params>(out, alpha, sum);
	store<Params>(result, idx, out);
}

template <typename Params>
__device__ __forceinline__ void perm_identity_eval(
	uint64_t out[Params::LIMBS], int64_t perm_idx, size_t n, unsigned log2n,
	const uint64_t coset_shift[Params::LIMBS],
	const uint64_t coset_shift_sq[Params::LIMBS],
	ConstFrView twiddles) {

	size_t idx = (size_t)perm_idx;
	size_t pos = idx & (n - 1);
	size_t coset = idx >> log2n;

	uint64_t omega_pos[Params::LIMBS];
	omega_from_twiddles<Params>(omega_pos, twiddles, pos, n >> 1);
	if(coset == 0) {
		set<Params>(out, omega_pos);
	} else if(coset == 1) {
		mul<Params>(out, coset_shift, omega_pos);
	} else {
		mul<Params>(out, coset_shift_sq, omega_pos);
	}
}

template <typename Params>
__global__ void z_compute_factors_kernel(
	FrView l_inout, FrView r_inout, ConstFrView o,
	const int64_t *perm, ConstFrView twiddles,
	ScalarArg beta_arg, ScalarArg gamma_arg,
	ScalarArg coset_shift_arg, ScalarArg coset_shift_sq_arg,
	size_t n, unsigned log2n) {

	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;

	uint64_t beta[Params::LIMBS], gamma[Params::LIMBS];
	uint64_t coset_shift[Params::LIMBS], coset_shift_sq[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		beta[i] = beta_arg.limbs[i];
		gamma[i] = gamma_arg.limbs[i];
		coset_shift[i] = coset_shift_arg.limbs[i];
		coset_shift_sq[i] = coset_shift_sq_arg.limbs[i];
	}

	uint64_t l_value[Params::LIMBS], r_value[Params::LIMBS], o_value[Params::LIMBS];
	load<Params>(l_value, make_const(l_inout), idx);
	load<Params>(r_value, make_const(r_inout), idx);
	load<Params>(o_value, o, idx);

	uint64_t omega_i[Params::LIMBS], beta_id0[Params::LIMBS];
	omega_from_twiddles<Params>(omega_i, twiddles, idx, n >> 1);
	mul<Params>(beta_id0, beta, omega_i);

	uint64_t beta_id1[Params::LIMBS], beta_id2[Params::LIMBS];
	mul<Params>(beta_id1, coset_shift, beta_id0);
	mul<Params>(beta_id2, coset_shift_sq, beta_id0);

	uint64_t t1[Params::LIMBS], t2[Params::LIMBS], t3[Params::LIMBS];
	add<Params>(t1, l_value, beta_id0);
	add<Params>(t1, t1, gamma);
	add<Params>(t2, r_value, beta_id1);
	add<Params>(t2, t2, gamma);
	add<Params>(t3, o_value, beta_id2);
	add<Params>(t3, t3, gamma);

	uint64_t tmp[Params::LIMBS], num[Params::LIMBS];
	mul<Params>(tmp, t1, t2);
	mul<Params>(num, tmp, t3);

	uint64_t sid0[Params::LIMBS], sid1[Params::LIMBS], sid2[Params::LIMBS];
	perm_identity_eval<Params>(sid0, perm[idx], n, log2n, coset_shift, coset_shift_sq, twiddles);
	perm_identity_eval<Params>(sid1, perm[n + idx], n, log2n, coset_shift, coset_shift_sq, twiddles);
	perm_identity_eval<Params>(sid2, perm[2 * n + idx], n, log2n, coset_shift, coset_shift_sq, twiddles);

	uint64_t beta_sid[Params::LIMBS];
	mul<Params>(beta_sid, beta, sid0);
	add<Params>(t1, l_value, beta_sid);
	add<Params>(t1, t1, gamma);
	mul<Params>(beta_sid, beta, sid1);
	add<Params>(t2, r_value, beta_sid);
	add<Params>(t2, t2, gamma);
	mul<Params>(beta_sid, beta, sid2);
	add<Params>(t3, o_value, beta_sid);
	add<Params>(t3, t3, gamma);

	uint64_t den[Params::LIMBS];
	mul<Params>(tmp, t1, t2);
	mul<Params>(den, tmp, t3);

	store<Params>(l_inout, idx, num);
	store<Params>(r_inout, idx, den);
}

template <typename Params>
__global__ void z_prefix_local_kernel(
	FrView z, ConstFrView ratio, uint64_t *chunk_products, size_t n) {

	size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	size_t num_chunks = (n + Z_PREFIX_CHUNK_SIZE - 1) / Z_PREFIX_CHUNK_SIZE;
	if(chunk_id >= num_chunks) return;

	size_t start = chunk_id * Z_PREFIX_CHUNK_SIZE;
	size_t end = start + Z_PREFIX_CHUNK_SIZE;
	if(end > n) end = n;

	uint64_t acc[Params::LIMBS], elem[Params::LIMBS];
	load<Params>(acc, ratio, start);
	store<Params>(z, start, acc);
	for(size_t i = start + 1; i < end; i++) {
		load<Params>(elem, ratio, i);
		mul<Params>(acc, acc, elem);
		store<Params>(z, i, acc);
	}

#pragma unroll
	for(int limb = 0; limb < Params::LIMBS; limb++) {
		chunk_products[chunk_id * Params::LIMBS + limb] = acc[limb];
	}
}

template <typename Params>
__global__ void z_prefix_fixup_kernel(FrView z, const uint64_t *scanned_prefixes, size_t n) {
	size_t chunk_id = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	size_t num_chunks = (n + Z_PREFIX_CHUNK_SIZE - 1) / Z_PREFIX_CHUNK_SIZE;
	if(chunk_id == 0 || chunk_id >= num_chunks) return;

	size_t start = chunk_id * Z_PREFIX_CHUNK_SIZE;
	size_t end = start + Z_PREFIX_CHUNK_SIZE;
	if(end > n) end = n;

	uint64_t prefix[Params::LIMBS], elem[Params::LIMBS], product[Params::LIMBS];
#pragma unroll
	for(int limb = 0; limb < Params::LIMBS; limb++) {
		prefix[limb] = scanned_prefixes[(chunk_id - 1) * Params::LIMBS + limb];
	}
	for(size_t i = start; i < end; i++) {
		load<Params>(elem, make_const(z), i);
		mul<Params>(product, prefix, elem);
		store<Params>(z, i, product);
	}
}

template <typename Params>
__global__ void z_prefix_shift_right_kernel(FrView z, ConstFrView src, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	if(idx == 0) {
		uint64_t one_value[Params::LIMBS];
		one<Params>(one_value);
		store<Params>(z, 0, one_value);
		return;
	}
	uint64_t prev[Params::LIMBS];
	load<Params>(prev, src, idx - 1);
	store<Params>(z, idx, prev);
}

template <typename Params>
__global__ void ntt_dif_stage_kernel(
	FrView data, ConstFrView twiddles, size_t num_butterflies,
	size_t half, size_t half_mask, size_t tw_stride) {

	size_t tid = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(tid >= num_butterflies) return;

	size_t j = tid & half_mask;
	size_t group_base = tid & ~half_mask;
	size_t idx_a = (group_base << 1) | j;
	size_t idx_b = idx_a + half;
	size_t tw_idx = j * tw_stride;

	uint64_t a[Params::LIMBS], b[Params::LIMBS], w[Params::LIMBS];
	uint64_t sum[Params::LIMBS], diff[Params::LIMBS], prod[Params::LIMBS];
	ConstFrView data_const = make_const(data);
	load<Params>(a, data_const, idx_a);
	load<Params>(b, data_const, idx_b);
	load<Params>(w, twiddles, tw_idx);

	add<Params>(sum, a, b);
	sub<Params>(diff, a, b);
	mul<Params>(prod, diff, w);

	store<Params>(data, idx_a, sum);
	store<Params>(data, idx_b, prod);
}

template <typename Params>
__global__ void ntt_dit_stage_kernel(
	FrView data, ConstFrView twiddles, size_t num_butterflies,
	size_t half, size_t half_mask, size_t tw_stride) {

	size_t tid = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(tid >= num_butterflies) return;

	size_t j = tid & half_mask;
	size_t group_base = tid & ~half_mask;
	size_t idx_a = (group_base << 1) | j;
	size_t idx_b = idx_a + half;
	size_t tw_idx = j * tw_stride;

	uint64_t a[Params::LIMBS], b[Params::LIMBS], w[Params::LIMBS];
	uint64_t wb[Params::LIMBS], sum[Params::LIMBS], diff[Params::LIMBS];
	ConstFrView data_const = make_const(data);
	load<Params>(a, data_const, idx_a);
	load<Params>(b, data_const, idx_b);
	load<Params>(w, twiddles, tw_idx);

	mul<Params>(wb, b, w);
	add<Params>(sum, a, wb);
	sub<Params>(diff, a, wb);

	store<Params>(data, idx_a, sum);
	store<Params>(data, idx_b, diff);
}

template <typename Params>
__global__ void scale_kernel(FrView data, const uint64_t *scalar, size_t n);

template <typename Params>
__global__ void ntt_dit_stage_scale_kernel(
	FrView data, ConstFrView twiddles, const uint64_t *scale,
	size_t num_butterflies, size_t half, size_t half_mask, size_t tw_stride) {

	size_t tid = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(tid >= num_butterflies) return;

	size_t j = tid & half_mask;
	size_t group_base = tid & ~half_mask;
	size_t idx_a = (group_base << 1) | j;
	size_t idx_b = idx_a + half;
	size_t tw_idx = j * tw_stride;

	uint64_t a[Params::LIMBS], b[Params::LIMBS], w[Params::LIMBS];
	uint64_t wb[Params::LIMBS], sum[Params::LIMBS], diff[Params::LIMBS];
	uint64_t scaled[Params::LIMBS], scale_value[Params::LIMBS];
	ConstFrView data_const = make_const(data);
	load<Params>(a, data_const, idx_a);
	load<Params>(b, data_const, idx_b);
	load<Params>(w, twiddles, tw_idx);
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		scale_value[i] = __ldg(scale + i);
	}

	mul<Params>(wb, b, w);
	add<Params>(sum, a, wb);
	sub<Params>(diff, a, wb);
	mul<Params>(scaled, sum, scale_value);
	store<Params>(data, idx_a, scaled);
	mul<Params>(scaled, diff, scale_value);
	store<Params>(data, idx_b, scaled);
}

template <typename Params>
__global__ void ntt_dif_radix8_kernel(
	FrView data, ConstFrView twiddles, uint32_t n, int stage_s) {

	uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
	uint32_t num_r8 = n >> 3;
	if(tid >= num_r8) return;

	uint32_t half_s = n >> (stage_s + 1);
	uint32_t half_s1 = half_s >> 1;
	uint32_t half_s2 = half_s >> 2;

	uint32_t j = tid & (half_s2 - 1);
	uint32_t group = tid >> (__ffs(half_s2) - 1);

	uint32_t base = group * (2 * half_s);
	uint32_t p0 = base + j;
	uint32_t p1 = p0 + half_s2;
	uint32_t p2 = p0 + half_s1;
	uint32_t p3 = p2 + half_s2;
	uint32_t p4 = p0 + half_s;
	uint32_t p5 = p4 + half_s2;
	uint32_t p6 = p4 + half_s1;
	uint32_t p7 = p6 + half_s2;

	uint64_t a0[Params::LIMBS], a1[Params::LIMBS], a2[Params::LIMBS], a3[Params::LIMBS];
	uint64_t a4[Params::LIMBS], a5[Params::LIMBS], a6[Params::LIMBS], a7[Params::LIMBS];
	ConstFrView data_const = make_const(data);
	load<Params>(a0, data_const, p0);
	load<Params>(a1, data_const, p1);
	load<Params>(a2, data_const, p2);
	load<Params>(a3, data_const, p3);
	load<Params>(a4, data_const, p4);
	load<Params>(a5, data_const, p5);
	load<Params>(a6, data_const, p6);
	load<Params>(a7, data_const, p7);

	uint32_t tw_stride_s = 1u << stage_s;
	uint32_t tw_stride_s1 = tw_stride_s << 1;
	uint32_t tw_stride_s2 = tw_stride_s << 2;

	uint64_t w[Params::LIMBS], sum[Params::LIMBS], diff[Params::LIMBS];
	uint32_t twi;

	twi = j * tw_stride_s;
	load<Params>(w, twiddles, twi);
	add<Params>(sum, a0, a4);
	sub<Params>(diff, a0, a4);
	mul<Params>(a4, diff, w);
	set<Params>(a0, sum);

	twi = (j + half_s2) * tw_stride_s;
	load<Params>(w, twiddles, twi);
	add<Params>(sum, a1, a5);
	sub<Params>(diff, a1, a5);
	mul<Params>(a5, diff, w);
	set<Params>(a1, sum);

	twi = (j + half_s1) * tw_stride_s;
	load<Params>(w, twiddles, twi);
	add<Params>(sum, a2, a6);
	sub<Params>(diff, a2, a6);
	mul<Params>(a6, diff, w);
	set<Params>(a2, sum);

	twi = (j + half_s1 + half_s2) * tw_stride_s;
	load<Params>(w, twiddles, twi);
	add<Params>(sum, a3, a7);
	sub<Params>(diff, a3, a7);
	mul<Params>(a7, diff, w);
	set<Params>(a3, sum);

	uint64_t ws1_0[Params::LIMBS], ws1_1[Params::LIMBS];
	twi = j * tw_stride_s1;
	load<Params>(ws1_0, twiddles, twi);
	twi = (j + half_s2) * tw_stride_s1;
	load<Params>(ws1_1, twiddles, twi);

	add<Params>(sum, a0, a2);
	sub<Params>(diff, a0, a2);
	mul<Params>(a2, diff, ws1_0);
	set<Params>(a0, sum);

	add<Params>(sum, a1, a3);
	sub<Params>(diff, a1, a3);
	mul<Params>(a3, diff, ws1_1);
	set<Params>(a1, sum);

	add<Params>(sum, a4, a6);
	sub<Params>(diff, a4, a6);
	mul<Params>(a6, diff, ws1_0);
	set<Params>(a4, sum);

	add<Params>(sum, a5, a7);
	sub<Params>(diff, a5, a7);
	mul<Params>(a7, diff, ws1_1);
	set<Params>(a5, sum);

	twi = j * tw_stride_s2;
	load<Params>(w, twiddles, twi);

	add<Params>(sum, a0, a1);
	sub<Params>(diff, a0, a1);
	mul<Params>(a1, diff, w);
	set<Params>(a0, sum);

	add<Params>(sum, a2, a3);
	sub<Params>(diff, a2, a3);
	mul<Params>(a3, diff, w);
	set<Params>(a2, sum);

	add<Params>(sum, a4, a5);
	sub<Params>(diff, a4, a5);
	mul<Params>(a5, diff, w);
	set<Params>(a4, sum);

	add<Params>(sum, a6, a7);
	sub<Params>(diff, a6, a7);
	mul<Params>(a7, diff, w);
	set<Params>(a6, sum);

	store<Params>(data, p0, a0);
	store<Params>(data, p1, a1);
	store<Params>(data, p2, a2);
	store<Params>(data, p3, a3);
	store<Params>(data, p4, a4);
	store<Params>(data, p5, a5);
	store<Params>(data, p6, a6);
	store<Params>(data, p7, a7);
}

template <typename Params, bool FUSE_SCALE>
__global__ void ntt_dit_radix8_kernel(
	FrView data, ConstFrView twiddles, const uint64_t *scale, uint32_t n, int stage_s) {

	uint32_t tid = (uint32_t)blockIdx.x * blockDim.x + threadIdx.x;
	uint32_t num_r8 = n >> 3;
	if(tid >= num_r8) return;

	uint32_t half_s = n >> (stage_s + 1);
	uint32_t half_s1 = half_s << 1;
	uint32_t half_s2 = half_s << 2;

	uint32_t j = tid & (half_s - 1);
	uint32_t group = tid >> (__ffs(half_s) - 1);

	uint32_t base = group * (8 * half_s);
	uint32_t p0 = base + j;
	uint32_t p1 = p0 + half_s;
	uint32_t p2 = p0 + half_s1;
	uint32_t p3 = p1 + half_s1;
	uint32_t p4 = p0 + half_s2;
	uint32_t p5 = p1 + half_s2;
	uint32_t p6 = p2 + half_s2;
	uint32_t p7 = p3 + half_s2;

	uint64_t a0[Params::LIMBS], a1[Params::LIMBS], a2[Params::LIMBS], a3[Params::LIMBS];
	uint64_t a4[Params::LIMBS], a5[Params::LIMBS], a6[Params::LIMBS], a7[Params::LIMBS];
	ConstFrView data_const = make_const(data);
	load<Params>(a0, data_const, p0);
	load<Params>(a1, data_const, p1);
	load<Params>(a2, data_const, p2);
	load<Params>(a3, data_const, p3);
	load<Params>(a4, data_const, p4);
	load<Params>(a5, data_const, p5);
	load<Params>(a6, data_const, p6);
	load<Params>(a7, data_const, p7);

	uint32_t tw_stride_s = 1u << stage_s;
	uint32_t tw_stride_s1 = tw_stride_s >> 1;
	uint32_t tw_stride_s2 = tw_stride_s >> 2;

	uint64_t w[Params::LIMBS], t[Params::LIMBS], sum[Params::LIMBS], diff[Params::LIMBS];
	uint32_t twi;

	twi = j * tw_stride_s;
	load<Params>(w, twiddles, twi);
	mul<Params>(t, a1, w);
	add<Params>(sum, a0, t);
	sub<Params>(diff, a0, t);
	set<Params>(a0, sum);
	set<Params>(a1, diff);

	mul<Params>(t, a3, w);
	add<Params>(sum, a2, t);
	sub<Params>(diff, a2, t);
	set<Params>(a2, sum);
	set<Params>(a3, diff);

	mul<Params>(t, a5, w);
	add<Params>(sum, a4, t);
	sub<Params>(diff, a4, t);
	set<Params>(a4, sum);
	set<Params>(a5, diff);

	mul<Params>(t, a7, w);
	add<Params>(sum, a6, t);
	sub<Params>(diff, a6, t);
	set<Params>(a6, sum);
	set<Params>(a7, diff);

	uint64_t ws1_a[Params::LIMBS], ws1_b[Params::LIMBS];
	twi = j * tw_stride_s1;
	load<Params>(ws1_a, twiddles, twi);
	twi = (j + half_s) * tw_stride_s1;
	load<Params>(ws1_b, twiddles, twi);

	mul<Params>(t, a2, ws1_a);
	add<Params>(sum, a0, t);
	sub<Params>(diff, a0, t);
	set<Params>(a0, sum);
	set<Params>(a2, diff);

	mul<Params>(t, a3, ws1_b);
	add<Params>(sum, a1, t);
	sub<Params>(diff, a1, t);
	set<Params>(a1, sum);
	set<Params>(a3, diff);

	mul<Params>(t, a6, ws1_a);
	add<Params>(sum, a4, t);
	sub<Params>(diff, a4, t);
	set<Params>(a4, sum);
	set<Params>(a6, diff);

	mul<Params>(t, a7, ws1_b);
	add<Params>(sum, a5, t);
	sub<Params>(diff, a5, t);
	set<Params>(a5, sum);
	set<Params>(a7, diff);

	twi = j * tw_stride_s2;
	load<Params>(w, twiddles, twi);
	mul<Params>(t, a4, w);
	add<Params>(sum, a0, t);
	sub<Params>(diff, a0, t);
	set<Params>(a0, sum);
	set<Params>(a4, diff);

	twi = (j + half_s) * tw_stride_s2;
	load<Params>(w, twiddles, twi);
	mul<Params>(t, a5, w);
	add<Params>(sum, a1, t);
	sub<Params>(diff, a1, t);
	set<Params>(a1, sum);
	set<Params>(a5, diff);

	twi = (j + half_s1) * tw_stride_s2;
	load<Params>(w, twiddles, twi);
	mul<Params>(t, a6, w);
	add<Params>(sum, a2, t);
	sub<Params>(diff, a2, t);
	set<Params>(a2, sum);
	set<Params>(a6, diff);

	twi = (j + half_s1 + half_s) * tw_stride_s2;
	load<Params>(w, twiddles, twi);
	mul<Params>(t, a7, w);
	add<Params>(sum, a3, t);
	sub<Params>(diff, a3, t);
	set<Params>(a3, sum);
	set<Params>(a7, diff);

	if constexpr(FUSE_SCALE) {
		uint64_t scale_value[Params::LIMBS], scaled[Params::LIMBS];
#pragma unroll
		for(int i = 0; i < Params::LIMBS; i++) {
			scale_value[i] = __ldg(scale + i);
		}
		mul<Params>(scaled, a0, scale_value); set<Params>(a0, scaled);
		mul<Params>(scaled, a1, scale_value); set<Params>(a1, scaled);
		mul<Params>(scaled, a2, scale_value); set<Params>(a2, scaled);
		mul<Params>(scaled, a3, scale_value); set<Params>(a3, scaled);
		mul<Params>(scaled, a4, scale_value); set<Params>(a4, scaled);
		mul<Params>(scaled, a5, scale_value); set<Params>(a5, scaled);
		mul<Params>(scaled, a6, scale_value); set<Params>(a6, scaled);
		mul<Params>(scaled, a7, scale_value); set<Params>(a7, scaled);
	}

	store<Params>(data, p0, a0);
	store<Params>(data, p1, a1);
	store<Params>(data, p2, a2);
	store<Params>(data, p3, a3);
	store<Params>(data, p4, a4);
	store<Params>(data, p5, a5);
	store<Params>(data, p6, a6);
	store<Params>(data, p7, a7);
}

template <typename Params, int TAIL_LOG>
__global__ void __launch_bounds__(1024, 1) ntt_dif_tail_fused_kernel(
	FrView data, ConstFrView twiddles, uint32_t n, int stage_start) {

	constexpr uint32_t span = 1u << TAIL_LOG;
	constexpr uint32_t butterflies_per_chunk = span >> 1;

	uint32_t chunk = (uint32_t)blockIdx.x;
	uint32_t base = chunk * span;
	uint32_t t = threadIdx.x;
	uint32_t p = blockDim.x;

	extern __shared__ uint64_t shmem[];
	uint64_t *s[MAX_FR_LIMBS];
	s[0] = shmem;
#pragma unroll
	for(int limb = 1; limb < Params::LIMBS; limb++) {
		s[limb] = s[limb - 1] + span;
	}

	for(uint32_t i = t; i < span; i += p) {
		uint32_t global_idx = base + i;
		if(global_idx < n) {
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				s[limb][i] = data.limbs[limb][global_idx];
			}
		}
	}
	__syncthreads();

#pragma unroll
	for(int st = 0; st < TAIL_LOG; st++) {
		int stage = stage_start + st;
		uint32_t half = n >> (stage + 1);
		uint32_t half_mask = half - 1;
		uint32_t tw_stride = 1u << stage;

		for(uint32_t bt = t; bt < butterflies_per_chunk; bt += p) {
			uint32_t j = bt & half_mask;
			uint32_t group_base = bt & ~half_mask;
			uint32_t idx_a = (group_base << 1) | j;
			uint32_t idx_b = idx_a + half;
			uint32_t tw_idx = j * tw_stride;

			uint64_t a[Params::LIMBS], b[Params::LIMBS], w[Params::LIMBS];
			uint64_t sum[Params::LIMBS], diff[Params::LIMBS], prod[Params::LIMBS];
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				a[limb] = s[limb][idx_a];
				b[limb] = s[limb][idx_b];
			}
			load<Params>(w, twiddles, tw_idx);

			add<Params>(sum, a, b);
			sub<Params>(diff, a, b);
			mul<Params>(prod, diff, w);

#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				s[limb][idx_a] = sum[limb];
				s[limb][idx_b] = prod[limb];
			}
		}
		__syncthreads();
	}

	for(uint32_t i = t; i < span; i += p) {
		uint32_t global_idx = base + i;
		if(global_idx < n) {
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				data.limbs[limb][global_idx] = s[limb][i];
			}
		}
	}
}

template <typename Params, int TAIL_LOG>
__global__ void __launch_bounds__(1024, 1) ntt_dit_tail_fused_kernel(
	FrView data, ConstFrView twiddles, uint32_t n, int stage_start) {

	constexpr uint32_t span = 1u << TAIL_LOG;
	constexpr uint32_t butterflies_per_chunk = span >> 1;

	uint32_t chunk = (uint32_t)blockIdx.x;
	uint32_t base = chunk * span;
	uint32_t t = threadIdx.x;
	uint32_t p = blockDim.x;

	extern __shared__ uint64_t shmem[];
	uint64_t *s[MAX_FR_LIMBS];
	s[0] = shmem;
#pragma unroll
	for(int limb = 1; limb < Params::LIMBS; limb++) {
		s[limb] = s[limb - 1] + span;
	}

	for(uint32_t i = t; i < span; i += p) {
		uint32_t global_idx = base + i;
		if(global_idx < n) {
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				s[limb][i] = data.limbs[limb][global_idx];
			}
		}
	}
	__syncthreads();

#pragma unroll
	for(int st = 0; st < TAIL_LOG; st++) {
		int stage = stage_start - st;
		uint32_t half = n >> (stage + 1);
		uint32_t half_mask = half - 1;
		uint32_t tw_stride = 1u << stage;

		for(uint32_t bt = t; bt < butterflies_per_chunk; bt += p) {
			uint32_t j = bt & half_mask;
			uint32_t group_base = bt & ~half_mask;
			uint32_t idx_a = (group_base << 1) | j;
			uint32_t idx_b = idx_a + half;
			uint32_t tw_idx = j * tw_stride;

			uint64_t a[Params::LIMBS], b[Params::LIMBS], w[Params::LIMBS];
			uint64_t wb[Params::LIMBS], sum[Params::LIMBS], diff[Params::LIMBS];
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				a[limb] = s[limb][idx_a];
				b[limb] = s[limb][idx_b];
			}
			load<Params>(w, twiddles, tw_idx);

			mul<Params>(wb, b, w);
			add<Params>(sum, a, wb);
			sub<Params>(diff, a, wb);

#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				s[limb][idx_a] = sum[limb];
				s[limb][idx_b] = diff[limb];
			}
		}
		__syncthreads();
	}

	for(uint32_t i = t; i < span; i += p) {
		uint32_t global_idx = base + i;
		if(global_idx < n) {
#pragma unroll
			for(int limb = 0; limb < Params::LIMBS; limb++) {
				data.limbs[limb][global_idx] = s[limb][i];
			}
		}
	}
}

template <typename Params>
__global__ void scale_kernel(FrView data, const uint64_t *scalar, size_t n) {
	size_t idx = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(idx >= n) return;
	uint64_t a[Params::LIMBS], out[Params::LIMBS];
	load<Params>(a, make_const(data), idx);
	mul<Params>(out, a, scalar);
	store<Params>(data, idx, out);
}

template <typename Params>
__device__ __forceinline__ void pow_uint64(
	uint64_t out[Params::LIMBS], const uint64_t base[Params::LIMBS], uint64_t exp) {

	uint64_t acc[Params::LIMBS], factor[Params::LIMBS];
	one<Params>(acc);
	set<Params>(factor, base);

	while(exp != 0) {
		if((exp & 1ULL) != 0) {
			mul<Params>(acc, acc, factor);
		}
		exp >>= 1;
		if(exp != 0) {
			square<Params>(factor, factor);
		}
	}

	set<Params>(out, acc);
}

template <typename Params>
__global__ void scale_by_powers_kernel(
	FrView data, ScalarArg generator_arg, const uint64_t *local_powers, size_t n) {

	__shared__ uint64_t block_base[Params::LIMBS];
	uint64_t generator[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) generator[i] = generator_arg.limbs[i];

	size_t block_start = (size_t)blockIdx.x * blockDim.x;

	if(threadIdx.x == 0) {
		pow_uint64<Params>(block_base, generator, (uint64_t)block_start);
	}
	__syncthreads();

	size_t idx = block_start + threadIdx.x;
	if(idx >= n) return;

	uint64_t local_power[Params::LIMBS], power[Params::LIMBS];
	uint64_t value[Params::LIMBS], out[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		local_power[i] = __ldg(local_powers + (size_t)threadIdx.x * Params::LIMBS + i);
	}
	mul<Params>(power, block_base, local_power);
	load<Params>(value, make_const(data), idx);
	mul<Params>(out, value, power);
	store<Params>(data, idx, out);
}

template <typename Params>
__global__ void local_power_table_kernel(ScalarArg generator_arg, uint64_t *local_powers) {
	uint64_t generator[Params::LIMBS], power[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) generator[i] = generator_arg.limbs[i];

	one<Params>(power);
	for(unsigned i = 0; i < THREADS; i++) {
#pragma unroll
		for(int limb = 0; limb < Params::LIMBS; limb++) {
			local_powers[(size_t)i * Params::LIMBS + limb] = power[limb];
		}
		mul<Params>(power, power, generator);
	}
}

__device__ __forceinline__ size_t bit_reverse(size_t x, int log_n) {
	uint32_t y = __brev((uint32_t)x);
	return (size_t)(y >> (32 - log_n));
}

template <typename Params>
__global__ void bit_reverse_kernel(FrView data, size_t n, int log_n) {
	size_t i = (size_t)blockIdx.x * blockDim.x + threadIdx.x;
	if(i >= n) return;

	size_t j = bit_reverse(i, log_n);
	if(j <= i) return;

#pragma unroll
	for(int limb = 0; limb < Params::LIMBS; limb++) {
		uint64_t a = data.limbs[limb][i];
		uint64_t b = data.limbs[limb][j];
		data.limbs[limb][i] = b;
		data.limbs[limb][j] = a;
	}
}

int log2_exact(size_t n) {
	int log = 0;
	while(n > 1) {
		n >>= 1;
		log++;
	}
	return log;
}

template <typename Params>
int select_tail_log(int log_n) {
	if(log_n < 10) return 0;

	int max_shmem = 0;
	cudaDeviceGetAttribute(&max_shmem, cudaDevAttrMaxSharedMemoryPerBlockOptin, 0);
	for(int candidate = 12; candidate >= 10; candidate--) {
		size_t required = (size_t)Params::LIMBS * ((size_t)1 << candidate) * sizeof(uint64_t);
		if(log_n > candidate && required <= (size_t)max_shmem) {
			return candidate;
		}
	}
	return 0;
}

template <typename Params>
uint32_t radix8_min_n() {
	return 1u << 18;
}

template <>
uint32_t radix8_min_n<BW6761FrParams>() {
	return 1u << 21;
}

template <typename Kernel>
void set_dynamic_shared_memory(Kernel kernel, size_t shmem_bytes) {
	cudaFuncSetAttribute(kernel, cudaFuncAttributeMaxDynamicSharedMemorySize, (int)shmem_bytes);
}

template <typename Params, int TAIL_LOG>
void launch_dif_tail_fixed(FrView data, ConstFrView twiddles, uint32_t n,
                           int stage_start, cudaStream_t stream) {
	constexpr uint32_t span = 1u << TAIL_LOG;
	unsigned threads = span > 1024 ? 1024 : span;
	unsigned blocks = (n + span - 1) / span;
	size_t shmem_bytes = (size_t)Params::LIMBS * span * sizeof(uint64_t);
	set_dynamic_shared_memory(ntt_dif_tail_fused_kernel<Params, TAIL_LOG>, shmem_bytes);
	ntt_dif_tail_fused_kernel<Params, TAIL_LOG><<<blocks, threads, shmem_bytes, stream>>>(
		data, twiddles, n, stage_start);
}

template <typename Params>
void launch_dif_tail(FrView data, ConstFrView twiddles, uint32_t n,
                     int stage_start, int tail_log, cudaStream_t stream) {
	switch(tail_log) {
	case 12:
		launch_dif_tail_fixed<Params, 12>(data, twiddles, n, stage_start, stream);
		break;
	case 11:
		launch_dif_tail_fixed<Params, 11>(data, twiddles, n, stage_start, stream);
		break;
	case 10:
		launch_dif_tail_fixed<Params, 10>(data, twiddles, n, stage_start, stream);
		break;
	default:
		break;
	}
}

template <typename Params, int TAIL_LOG>
void launch_dit_tail_fixed(FrView data, ConstFrView twiddles, uint32_t n,
                           int stage_start, cudaStream_t stream) {
	constexpr uint32_t span = 1u << TAIL_LOG;
	unsigned threads = span > 1024 ? 1024 : span;
	unsigned blocks = (n + span - 1) / span;
	size_t shmem_bytes = (size_t)Params::LIMBS * span * sizeof(uint64_t);
	set_dynamic_shared_memory(ntt_dit_tail_fused_kernel<Params, TAIL_LOG>, shmem_bytes);
	ntt_dit_tail_fused_kernel<Params, TAIL_LOG><<<blocks, threads, shmem_bytes, stream>>>(
		data, twiddles, n, stage_start);
}

template <typename Params>
void launch_dit_tail(FrView data, ConstFrView twiddles, uint32_t n,
                     int stage_start, int tail_log, cudaStream_t stream) {
	switch(tail_log) {
	case 12:
		launch_dit_tail_fixed<Params, 12>(data, twiddles, n, stage_start, stream);
		break;
	case 11:
		launch_dit_tail_fixed<Params, 11>(data, twiddles, n, stage_start, stream);
		break;
	case 10:
		launch_dit_tail_fixed<Params, 10>(data, twiddles, n, stage_start, stream);
		break;
	default:
		break;
	}
}

template <typename Params>
void launch_ntt_forward_typed(FrView data, ConstFrView twiddles, size_t n,
                              cudaStream_t stream) {
	const int log_n = log2_exact(n);
	const size_t butterflies = n >> 1;
	unsigned blocks_r2 = (unsigned)((butterflies + NTT_THREADS - 1) / NTT_THREADS);
	uint32_t n32 = (uint32_t)n;
	uint32_t radix8_count = n32 >> 3;
	unsigned blocks_r8 = (radix8_count + NTT_THREADS - 1) / NTT_THREADS;

	int tail_log = select_tail_log<Params>(log_n);
	bool use_tail = tail_log > 0 && n >= NTT_FUSED_TAIL_MIN_N;
	int regular_stages = use_tail ? log_n - tail_log : log_n;

	int stage = 0;
	if(n >= radix8_min_n<Params>()) {
		for(; stage + 2 < regular_stages; stage += 3) {
			ntt_dif_radix8_kernel<Params><<<blocks_r8, NTT_THREADS, 0, stream>>>(
				data, twiddles, n32, stage);
		}
	}
	for(; stage < regular_stages; stage++) {
		size_t half = n >> (stage + 1);
		size_t tw_stride = (size_t)1 << stage;
		ntt_dif_stage_kernel<Params><<<blocks_r2, NTT_THREADS, 0, stream>>>(
			data, twiddles, butterflies, half, half - 1, tw_stride);
	}

	if(use_tail) {
		launch_dif_tail<Params>(data, twiddles, n32, regular_stages, tail_log, stream);
	}
}

template <typename Params>
void launch_ntt_inverse_typed(FrView data, ConstFrView twiddles, const uint64_t *inv_n,
                              size_t n, cudaStream_t stream) {
	const int log_n = log2_exact(n);
	const size_t butterflies = n >> 1;
	unsigned blocks_r2 = (unsigned)((butterflies + NTT_THREADS - 1) / NTT_THREADS);
	uint32_t n32 = (uint32_t)n;
	uint32_t radix8_count = n32 >> 3;
	unsigned blocks_r8 = (radix8_count + NTT_THREADS - 1) / NTT_THREADS;

	int tail_log = select_tail_log<Params>(log_n);
	bool use_tail = tail_log > 0 && n >= NTT_FUSED_TAIL_MIN_N;
	int stage = log_n - 1;
	if(use_tail) {
		launch_dit_tail<Params>(data, twiddles, n32, stage, tail_log, stream);
		stage -= tail_log;
	}

	bool scaled = false;
	if(n >= radix8_min_n<Params>()) {
		for(; stage - 2 >= 0; stage -= 3) {
			if(stage < 3) {
				ntt_dit_radix8_kernel<Params, true><<<blocks_r8, NTT_THREADS, 0, stream>>>(
					data, twiddles, inv_n, n32, stage);
				scaled = true;
			} else {
				ntt_dit_radix8_kernel<Params, false><<<blocks_r8, NTT_THREADS, 0, stream>>>(
					data, twiddles, inv_n, n32, stage);
			}
		}
	}
	for(; stage >= 0; stage--) {
		size_t half = n >> (stage + 1);
		size_t tw_stride = (size_t)1 << stage;
		if(stage == 0) {
			ntt_dit_stage_scale_kernel<Params><<<blocks_r2, NTT_THREADS, 0, stream>>>(
				data, twiddles, inv_n, butterflies, half, half - 1, tw_stride);
			scaled = true;
		} else {
			ntt_dit_stage_kernel<Params><<<blocks_r2, NTT_THREADS, 0, stream>>>(
				data, twiddles, butterflies, half, half - 1, tw_stride);
		}
	}

	if(!scaled) {
		unsigned scale_blocks = (unsigned)((n + NTT_THREADS - 1) / NTT_THREADS);
		scale_kernel<Params><<<scale_blocks, NTT_THREADS, 0, stream>>>(data, inv_n, n);
	}
}

} // namespace

void launch_copy_aos_to_soa(
	gnark_gpu_plonk2_curve_id_t curve, FrView dst, const uint64_t *src,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		copy_aos_to_soa_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		copy_aos_to_soa_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		copy_aos_to_soa_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	default:
		break;
	}
}

void launch_copy_soa_to_aos(
	gnark_gpu_plonk2_curve_id_t curve, uint64_t *dst, ConstFrView src,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		copy_soa_to_aos_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		copy_soa_to_aos_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		copy_soa_to_aos_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(dst, src, n);
		break;
	default:
		break;
	}
}

void launch_set_zero(
	gnark_gpu_plonk2_curve_id_t curve, FrView v, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		set_zero_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(v, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		set_zero_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(v, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		set_zero_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(v, n);
		break;
	default:
		break;
	}
}

void launch_add(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a, ConstFrView b,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		add_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		add_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		add_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	default:
		break;
	}
}

void launch_sub(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a, ConstFrView b,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		sub_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		sub_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		sub_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	default:
		break;
	}
}

void launch_mul(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a, ConstFrView b,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		mul_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		mul_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		mul_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	default:
		break;
	}
}

void launch_addmul(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a, ConstFrView b,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		addmul_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		addmul_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		addmul_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(out, a, b, n);
		break;
	default:
		break;
	}
}

void launch_scalar_mul(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, const uint64_t *scalar,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg scalar_arg = make_scalar_arg(curve, scalar);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		scalar_mul_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			out, scalar_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		scalar_mul_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			out, scalar_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		scalar_mul_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			out, scalar_arg, n);
		break;
	default:
		break;
	}
}

void launch_add_scalar_mul(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView a,
	const uint64_t *scalar, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg scalar_arg = make_scalar_arg(curve, scalar);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		add_scalar_mul_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			out, a, scalar_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		add_scalar_mul_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			out, a, scalar_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		add_scalar_mul_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			out, a, scalar_arg, n);
		break;
	default:
		break;
	}
}

void launch_batch_invert(
	gnark_gpu_plonk2_curve_id_t curve, FrView data, size_t n,
	cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		invert_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(data, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		invert_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(data, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		invert_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(data, n);
		break;
	default:
		break;
	}
}

void launch_butterfly4_inverse(
	gnark_gpu_plonk2_curve_id_t curve, FrView b0, FrView b1, FrView b2, FrView b3,
	const uint64_t *omega4_inv, const uint64_t *quarter, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg omega4_inv_arg = make_scalar_arg(curve, omega4_inv);
	ScalarArg quarter_arg = make_scalar_arg(curve, quarter);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		butterfly4_inverse_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			b0, b1, b2, b3, omega4_inv_arg, quarter_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		butterfly4_inverse_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			b0, b1, b2, b3, omega4_inv_arg, quarter_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		butterfly4_inverse_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			b0, b1, b2, b3, omega4_inv_arg, quarter_arg, n);
		break;
	default:
		break;
	}
}

void launch_reduce_blinded_coset(
	gnark_gpu_plonk2_curve_id_t curve, FrView dst, ConstFrView src,
	const uint64_t *tail, size_t tail_len, const uint64_t *coset_pow_n,
	size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg coset_pow_n_arg = make_scalar_arg(curve, coset_pow_n);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		reduce_blinded_coset_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			dst, src, tail, coset_pow_n_arg, n, tail_len);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		reduce_blinded_coset_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			dst, src, tail, coset_pow_n_arg, n, tail_len);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		reduce_blinded_coset_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			dst, src, tail, coset_pow_n_arg, n, tail_len);
		break;
	default:
		break;
	}
}

void launch_compute_l1_den(
	gnark_gpu_plonk2_curve_id_t curve, FrView out, ConstFrView twiddles,
	const uint64_t *coset_gen, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg coset_gen_arg = make_scalar_arg(curve, coset_gen);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		compute_l1_den_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			out, twiddles, coset_gen_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		compute_l1_den_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			out, twiddles, coset_gen_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		compute_l1_den_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			out, twiddles, coset_gen_arg, n);
		break;
	default:
		break;
	}
}

void launch_gate_accum(
	gnark_gpu_plonk2_curve_id_t curve, FrView result,
	ConstFrView ql, ConstFrView qr, ConstFrView qm, ConstFrView qo, ConstFrView qk,
	ConstFrView l, ConstFrView r, ConstFrView o,
	const uint64_t *zh_k_inv, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg zh_k_inv_arg = make_scalar_arg(curve, zh_k_inv);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		gate_accum_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			result, ql, qr, qm, qo, qk, l, r, o, zh_k_inv_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		gate_accum_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			result, ql, qr, qm, qo, qk, l, r, o, zh_k_inv_arg, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		gate_accum_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			result, ql, qr, qm, qo, qk, l, r, o, zh_k_inv_arg, n);
		break;
	default:
		break;
	}
}

void launch_perm_boundary(
	gnark_gpu_plonk2_curve_id_t curve, FrView result,
	ConstFrView l, ConstFrView r, ConstFrView o, ConstFrView z,
	ConstFrView s1, ConstFrView s2, ConstFrView s3, ConstFrView l1_den_inv,
	ConstFrView twiddles, const uint64_t *params, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	int limbs = curve_limbs(curve);
	ScalarArg alpha = make_scalar_arg(curve, params);
	ScalarArg beta = make_scalar_arg(curve, params + limbs);
	ScalarArg gamma = make_scalar_arg(curve, params + 2 * limbs);
	ScalarArg l1_scalar = make_scalar_arg(curve, params + 3 * limbs);
	ScalarArg coset_shift = make_scalar_arg(curve, params + 4 * limbs);
	ScalarArg coset_shift_sq = make_scalar_arg(curve, params + 5 * limbs);
	ScalarArg coset_gen = make_scalar_arg(curve, params + 6 * limbs);

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		perm_boundary_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			result, l, r, o, z, s1, s2, s3, l1_den_inv, twiddles,
			alpha, beta, gamma, l1_scalar, coset_shift, coset_shift_sq, coset_gen, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		perm_boundary_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			result, l, r, o, z, s1, s2, s3, l1_den_inv, twiddles,
			alpha, beta, gamma, l1_scalar, coset_shift, coset_shift_sq, coset_gen, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		perm_boundary_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			result, l, r, o, z, s1, s2, s3, l1_den_inv, twiddles,
			alpha, beta, gamma, l1_scalar, coset_shift, coset_shift_sq, coset_gen, n);
		break;
	default:
		break;
	}
}

void launch_z_compute_factors(
	gnark_gpu_plonk2_curve_id_t curve, FrView l_inout, FrView r_inout,
	ConstFrView o, const int64_t *perm, ConstFrView twiddles,
	const uint64_t *params, size_t n, unsigned log2n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	int limbs = curve_limbs(curve);
	ScalarArg beta = make_scalar_arg(curve, params);
	ScalarArg gamma = make_scalar_arg(curve, params + limbs);
	ScalarArg coset_shift = make_scalar_arg(curve, params + 2 * limbs);
	ScalarArg coset_shift_sq = make_scalar_arg(curve, params + 3 * limbs);

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		z_compute_factors_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			l_inout, r_inout, o, perm, twiddles, beta, gamma,
			coset_shift, coset_shift_sq, n, log2n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		z_compute_factors_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			l_inout, r_inout, o, perm, twiddles, beta, gamma,
			coset_shift, coset_shift_sq, n, log2n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		z_compute_factors_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			l_inout, r_inout, o, perm, twiddles, beta, gamma,
			coset_shift, coset_shift_sq, n, log2n);
		break;
	default:
		break;
	}
}

void launch_z_prefix_phase1(
	gnark_gpu_plonk2_curve_id_t curve, FrView z, ConstFrView ratio,
	uint64_t *chunk_products, size_t n, cudaStream_t stream) {

	size_t num_chunks = (n + Z_PREFIX_CHUNK_SIZE - 1) / Z_PREFIX_CHUNK_SIZE;
	unsigned blocks = (unsigned)((num_chunks + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		z_prefix_local_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			z, ratio, chunk_products, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		z_prefix_local_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			z, ratio, chunk_products, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		z_prefix_local_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			z, ratio, chunk_products, n);
		break;
	default:
		break;
	}
}

void launch_z_prefix_phase3(
	gnark_gpu_plonk2_curve_id_t curve, FrView z, FrView temp,
	const uint64_t *scanned_prefixes, size_t num_chunks, size_t n,
	cudaStream_t stream) {

	unsigned chunk_blocks = (unsigned)((num_chunks + THREADS - 1) / THREADS);
	unsigned n_blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		z_prefix_fixup_kernel<BN254FrParams><<<chunk_blocks, THREADS, 0, stream>>>(
			z, scanned_prefixes, n);
		for(int i = 0; i < BN254FrParams::LIMBS; i++) {
			cudaMemcpyAsync(temp.limbs[i], z.limbs[i], n * sizeof(uint64_t),
			                cudaMemcpyDeviceToDevice, stream);
		}
		z_prefix_shift_right_kernel<BN254FrParams><<<n_blocks, THREADS, 0, stream>>>(
			z, make_const(temp), n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		z_prefix_fixup_kernel<BLS12377FrParams><<<chunk_blocks, THREADS, 0, stream>>>(
			z, scanned_prefixes, n);
		for(int i = 0; i < BLS12377FrParams::LIMBS; i++) {
			cudaMemcpyAsync(temp.limbs[i], z.limbs[i], n * sizeof(uint64_t),
			                cudaMemcpyDeviceToDevice, stream);
		}
		z_prefix_shift_right_kernel<BLS12377FrParams><<<n_blocks, THREADS, 0, stream>>>(
			z, make_const(temp), n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		z_prefix_fixup_kernel<BW6761FrParams><<<chunk_blocks, THREADS, 0, stream>>>(
			z, scanned_prefixes, n);
		for(int i = 0; i < BW6761FrParams::LIMBS; i++) {
			cudaMemcpyAsync(temp.limbs[i], z.limbs[i], n * sizeof(uint64_t),
			                cudaMemcpyDeviceToDevice, stream);
		}
		z_prefix_shift_right_kernel<BW6761FrParams><<<n_blocks, THREADS, 0, stream>>>(
			z, make_const(temp), n);
		break;
	default:
		break;
	}
}

void launch_ntt_forward(
	gnark_gpu_plonk2_curve_id_t curve, FrView data, ConstFrView twiddles,
	size_t n, cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		launch_ntt_forward_typed<BN254FrParams>(data, twiddles, n, stream);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		launch_ntt_forward_typed<BLS12377FrParams>(data, twiddles, n, stream);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		launch_ntt_forward_typed<BW6761FrParams>(data, twiddles, n, stream);
		break;
	default:
		break;
	}
}

void launch_ntt_inverse(
	gnark_gpu_plonk2_curve_id_t curve, FrView data, ConstFrView twiddles,
	const uint64_t *inv_n, size_t n, cudaStream_t stream) {

	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		launch_ntt_inverse_typed<BN254FrParams>(data, twiddles, inv_n, n, stream);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		launch_ntt_inverse_typed<BLS12377FrParams>(data, twiddles, inv_n, n, stream);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		launch_ntt_inverse_typed<BW6761FrParams>(data, twiddles, inv_n, n, stream);
		break;
	default:
		break;
	}
}

void launch_scale_by_powers(
	gnark_gpu_plonk2_curve_id_t curve, FrView data, const uint64_t *generator,
	uint64_t *local_powers, size_t n, cudaStream_t stream) {

	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	ScalarArg generator_arg = make_scalar_arg(curve, generator);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		local_power_table_kernel<BN254FrParams><<<1, 1, 0, stream>>>(
			generator_arg, local_powers);
		scale_by_powers_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(
			data, generator_arg, local_powers, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		local_power_table_kernel<BLS12377FrParams><<<1, 1, 0, stream>>>(
			generator_arg, local_powers);
		scale_by_powers_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(
			data, generator_arg, local_powers, n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		local_power_table_kernel<BW6761FrParams><<<1, 1, 0, stream>>>(
			generator_arg, local_powers);
		scale_by_powers_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(
			data, generator_arg, local_powers, n);
		break;
	default:
		break;
	}
}

void launch_bit_reverse(
	gnark_gpu_plonk2_curve_id_t curve, FrView data, size_t n, cudaStream_t stream) {

	int log_n = log2_exact(n);
	unsigned blocks = (unsigned)((n + THREADS - 1) / THREADS);
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		bit_reverse_kernel<BN254FrParams><<<blocks, THREADS, 0, stream>>>(data, n, log_n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		bit_reverse_kernel<BLS12377FrParams><<<blocks, THREADS, 0, stream>>>(data, n, log_n);
		break;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		bit_reverse_kernel<BW6761FrParams><<<blocks, THREADS, 0, stream>>>(data, n, log_n);
		break;
	default:
		break;
	}
}

} // namespace gnark_gpu::plonk2
