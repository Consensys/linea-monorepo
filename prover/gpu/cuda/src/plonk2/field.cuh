#pragma once

// Curve-generic scalar-field primitives for gpu/plonk2.
//
// This layer favors one small, auditable implementation over per-curve copied
// arithmetic. It is intentionally separate from src/plonk/fr_arith.cuh, which
// remains the optimized BLS12-377 path used by the existing prover.

#include "gnark_gpu.h"

#include <cstddef>
#include <cstdint>

#ifdef __CUDACC__
#include <cuda_runtime.h>
#else
#define __host__
#define __device__
#define __forceinline__
#endif

namespace gnark_gpu::plonk2 {

static constexpr int MAX_FIELD_LIMBS = 12;
static constexpr int MAX_FR_LIMBS = MAX_FIELD_LIMBS;

struct FrView {
	uint64_t *limbs[MAX_FR_LIMBS];
};

struct ConstFrView {
	const uint64_t *limbs[MAX_FR_LIMBS];
};

__host__ __device__ __forceinline__ ConstFrView make_const(FrView v) {
	ConstFrView out{};
#pragma unroll
	for(int i = 0; i < MAX_FR_LIMBS; i++) out.limbs[i] = v.limbs[i];
	return out;
}

struct BN254FrParams {
	static constexpr int LIMBS = 4;
	static constexpr int BITS = 254;
	static constexpr gnark_gpu_plonk2_curve_id_t CURVE = GNARK_GPU_PLONK2_CURVE_BN254;
	static constexpr uint64_t INV = 0xc2e1f593efffffffULL;
	static constexpr uint64_t MODULUS[MAX_FR_LIMBS] = {
		0x43e1f593f0000001ULL,
		0x2833e84879b97091ULL,
		0xb85045b68181585dULL,
		0x30644e72e131a029ULL,
		0x0000000000000000ULL,
		0x0000000000000000ULL,
	};
};

struct BLS12377FrParams {
	static constexpr int LIMBS = 4;
	static constexpr int BITS = 253;
	static constexpr gnark_gpu_plonk2_curve_id_t CURVE = GNARK_GPU_PLONK2_CURVE_BLS12_377;
	static constexpr uint64_t INV = 0x0a117fffffffffffULL;
	static constexpr uint64_t MODULUS[MAX_FR_LIMBS] = {
		0x0a11800000000001ULL,
		0x59aa76fed0000001ULL,
		0x60b44d1e5c37b001ULL,
		0x12ab655e9a2ca556ULL,
		0x0000000000000000ULL,
		0x0000000000000000ULL,
	};
};

struct BW6761FrParams {
	static constexpr int LIMBS = 6;
	static constexpr int BITS = 377;
	static constexpr gnark_gpu_plonk2_curve_id_t CURVE = GNARK_GPU_PLONK2_CURVE_BW6_761;
	static constexpr uint64_t INV = 0x8508bfffffffffffULL;
	static constexpr uint64_t MODULUS[MAX_FR_LIMBS] = {
		0x8508c00000000001ULL,
		0x170b5d4430000000ULL,
		0x1ef3622fba094800ULL,
		0x1a22d9f300f5138fULL,
		0xc63b05c06ca1493bULL,
		0x01ae3a4617c510eaULL,
	};
};

struct BN254FpParams {
	static constexpr int LIMBS = 4;
	static constexpr int BITS = 254;
	static constexpr uint64_t INV = 0x87d20782e4866389ULL;
};

struct BLS12377FpParams {
	static constexpr int LIMBS = 6;
	static constexpr int BITS = 377;
	static constexpr uint64_t INV = 0x8508bfffffffffffULL;
};

struct BW6761FpParams {
	static constexpr int LIMBS = 12;
	static constexpr int BITS = 761;
	static constexpr uint64_t INV = 0x0a5593568fa798ddULL;
};

template <typename Params>
__device__ __forceinline__ uint64_t modulus_limb(int i);

template <>
__device__ __forceinline__ uint64_t modulus_limb<BN254FrParams>(int i) {
	switch(i) {
	case 0:
		return 0x43e1f593f0000001ULL;
	case 1:
		return 0x2833e84879b97091ULL;
	case 2:
		return 0xb85045b68181585dULL;
	case 3:
		return 0x30644e72e131a029ULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t modulus_limb<BLS12377FrParams>(int i) {
	switch(i) {
	case 0:
		return 0x0a11800000000001ULL;
	case 1:
		return 0x59aa76fed0000001ULL;
	case 2:
		return 0x60b44d1e5c37b001ULL;
	case 3:
		return 0x12ab655e9a2ca556ULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t modulus_limb<BW6761FrParams>(int i) {
	switch(i) {
	case 0:
		return 0x8508c00000000001ULL;
	case 1:
		return 0x170b5d4430000000ULL;
	case 2:
		return 0x1ef3622fba094800ULL;
	case 3:
		return 0x1a22d9f300f5138fULL;
	case 4:
		return 0xc63b05c06ca1493bULL;
	case 5:
		return 0x01ae3a4617c510eaULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t modulus_limb<BN254FpParams>(int i) {
	switch(i) {
	case 0:
		return 0x3c208c16d87cfd47ULL;
	case 1:
		return 0x97816a916871ca8dULL;
	case 2:
		return 0xb85045b68181585dULL;
	case 3:
		return 0x30644e72e131a029ULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t modulus_limb<BLS12377FpParams>(int i) {
	switch(i) {
	case 0:
		return 0x8508c00000000001ULL;
	case 1:
		return 0x170b5d4430000000ULL;
	case 2:
		return 0x1ef3622fba094800ULL;
	case 3:
		return 0x1a22d9f300f5138fULL;
	case 4:
		return 0xc63b05c06ca1493bULL;
	case 5:
		return 0x01ae3a4617c510eaULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t modulus_limb<BW6761FpParams>(int i) {
	switch(i) {
	case 0:
		return 0xf49d00000000008bULL;
	case 1:
		return 0xe6913e6870000082ULL;
	case 2:
		return 0x160cf8aeeaf0a437ULL;
	case 3:
		return 0x98a116c25667a8f8ULL;
	case 4:
		return 0x71dcd3dc73ebff2eULL;
	case 5:
		return 0x8689c8ed12f9fd90ULL;
	case 6:
		return 0x03cebaff25b42304ULL;
	case 7:
		return 0x707ba638e584e919ULL;
	case 8:
		return 0x528275ef8087be41ULL;
	case 9:
		return 0xb926186a81d14688ULL;
	case 10:
		return 0xd187c94004faff3eULL;
	case 11:
		return 0x0122e824fb83ce0aULL;
	default:
		return 0;
	}
}

template <typename Params>
__device__ __forceinline__ uint64_t one_limb(int i);

template <>
__device__ __forceinline__ uint64_t one_limb<BN254FrParams>(int i) {
	switch(i) {
	case 0:
		return 0xac96341c4ffffffbULL;
	case 1:
		return 0x36fc76959f60cd29ULL;
	case 2:
		return 0x666ea36f7879462eULL;
	case 3:
		return 0x0e0a77c19a07df2fULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t one_limb<BLS12377FrParams>(int i) {
	switch(i) {
	case 0:
		return 0x7d1c7ffffffffff3ULL;
	case 1:
		return 0x7257f50f6ffffff2ULL;
	case 2:
		return 0x16d81575512c0feeULL;
	case 3:
		return 0x0d4bda322bbb9a9dULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t one_limb<BW6761FrParams>(int i) {
	switch(i) {
	case 0:
		return 0x02cdffffffffff68ULL;
	case 1:
		return 0x51409f837fffffb1ULL;
	case 2:
		return 0x9f7db3a98a7d3ff2ULL;
	case 3:
		return 0x7b4e97b76e7c6305ULL;
	case 4:
		return 0x4cf495bf803c84e8ULL;
	case 5:
		return 0x008d6661e2fdf49aULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t one_limb<BN254FpParams>(int i) {
	switch(i) {
	case 0:
		return 0xd35d438dc58f0d9dULL;
	case 1:
		return 0x0a78eb28f5c70b3dULL;
	case 2:
		return 0x666ea36f7879462cULL;
	case 3:
		return 0x0e0a77c19a07df2fULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t one_limb<BLS12377FpParams>(int i) {
	switch(i) {
	case 0:
		return 0x02cdffffffffff68ULL;
	case 1:
		return 0x51409f837fffffb1ULL;
	case 2:
		return 0x9f7db3a98a7d3ff2ULL;
	case 3:
		return 0x7b4e97b76e7c6305ULL;
	case 4:
		return 0x4cf495bf803c84e8ULL;
	case 5:
		return 0x008d6661e2fdf49aULL;
	default:
		return 0;
	}
}

template <>
__device__ __forceinline__ uint64_t one_limb<BW6761FpParams>(int i) {
	switch(i) {
	case 0:
		return 0x0202ffffffff85d5ULL;
	case 1:
		return 0x5a5826358fff8ce7ULL;
	case 2:
		return 0x9e996e43827faadeULL;
	case 3:
		return 0xda6aff320ee47df4ULL;
	case 4:
		return 0xece9cb3e1d94b80bULL;
	case 5:
		return 0xc0e667a25248240bULL;
	case 6:
		return 0xa74da5bfdcad3905ULL;
	case 7:
		return 0x2352e7fe462f2103ULL;
	case 8:
		return 0x7b56588008b1c87cULL;
	case 9:
		return 0x45848a63e711022fULL;
	case 10:
		return 0xd7a81ebb9f65a9dfULL;
	case 11:
		return 0x0051f77ef127e87dULL;
	default:
		return 0;
	}
}

__host__ __device__ __forceinline__ int curve_base_limbs(gnark_gpu_plonk2_curve_id_t curve) {
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return BN254FpParams::LIMBS;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return BLS12377FpParams::LIMBS;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return BW6761FpParams::LIMBS;
	default:
		return 0;
	}
}

__host__ __device__ __forceinline__ int curve_limbs(gnark_gpu_plonk2_curve_id_t curve) {
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return 4;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return 6;
	default:
		return 0;
	}
}

__host__ __device__ __forceinline__ int curve_bits(gnark_gpu_plonk2_curve_id_t curve) {
	switch(curve) {
	case GNARK_GPU_PLONK2_CURVE_BN254:
		return BN254FrParams::BITS;
	case GNARK_GPU_PLONK2_CURVE_BLS12_377:
		return BLS12377FrParams::BITS;
	case GNARK_GPU_PLONK2_CURVE_BW6_761:
		return BW6761FrParams::BITS;
	default:
		return 0;
	}
}

__device__ __forceinline__ uint64_t add_carry(uint64_t a, uint64_t b, uint64_t &carry) {
	uint64_t s = a + b;
	uint64_t c = s < a;
	uint64_t r = s + carry;
	c += r < s;
	carry = c;
	return r;
}

__device__ __forceinline__ uint64_t sub_borrow(uint64_t a, uint64_t b, uint64_t &borrow) {
	uint64_t bb = b + borrow;
	uint64_t bcarry = bb < b;
	uint64_t r = a - bb;
	borrow = (a < bb) || bcarry;
	return r;
}

template <typename Params>
__device__ __forceinline__ void load(uint64_t out[Params::LIMBS], ConstFrView v, size_t idx) {
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		out[i] = __ldg(v.limbs[i] + idx);
	}
}

template <typename Params>
__device__ __forceinline__ void store(FrView v, size_t idx, const uint64_t in[Params::LIMBS]) {
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		v.limbs[i][idx] = in[i];
	}
}

template <typename Params>
__device__ __forceinline__ uint64_t subtract_modulus(
	uint64_t out[Params::LIMBS], const uint64_t in[Params::LIMBS]) {
	uint64_t borrow = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		out[i] = sub_borrow(in[i], modulus_limb<Params>(i), borrow);
	}
	return borrow;
}

template <typename Params>
__device__ __forceinline__ void add(uint64_t r[Params::LIMBS],
                                    const uint64_t a[Params::LIMBS],
                                    const uint64_t b[Params::LIMBS]) {
	uint64_t sum[Params::LIMBS];
	uint64_t carry = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		sum[i] = add_carry(a[i], b[i], carry);
	}

	uint64_t reduced[Params::LIMBS];
	uint64_t borrow = subtract_modulus<Params>(reduced, sum);
	bool use_reduced = carry != 0 || borrow == 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		r[i] = use_reduced ? reduced[i] : sum[i];
	}
}

template <typename Params>
__device__ __forceinline__ void sub(uint64_t r[Params::LIMBS],
                                    const uint64_t a[Params::LIMBS],
                                    const uint64_t b[Params::LIMBS]) {
	uint64_t diff[Params::LIMBS];
	uint64_t borrow = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		diff[i] = sub_borrow(a[i], b[i], borrow);
	}

	if(borrow == 0) {
#pragma unroll
		for(int i = 0; i < Params::LIMBS; i++) r[i] = diff[i];
		return;
	}

	uint64_t carry = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		r[i] = add_carry(diff[i], modulus_limb<Params>(i), carry);
	}
}

template <typename Params>
__device__ __forceinline__ void mul(uint64_t r[Params::LIMBS],
                                    const uint64_t a[Params::LIMBS],
                                    const uint64_t b[Params::LIMBS]) {
	uint64_t t[Params::LIMBS + 1];
#pragma unroll
	for(int i = 0; i <= Params::LIMBS; i++) t[i] = 0;

#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		unsigned __int128 carry = 0;
#pragma unroll
		for(int j = 0; j < Params::LIMBS; j++) {
			unsigned __int128 uv =
				(unsigned __int128)t[j] +
				(unsigned __int128)a[j] * (unsigned __int128)b[i] +
				carry;
			t[j] = (uint64_t)uv;
			carry = uv >> 64;
		}
		unsigned __int128 top = (unsigned __int128)t[Params::LIMBS] + carry;
		t[Params::LIMBS] = (uint64_t)top;

		uint64_t m = t[0] * Params::INV;
		carry = 0;
#pragma unroll
		for(int j = 0; j < Params::LIMBS; j++) {
			unsigned __int128 uv =
				(unsigned __int128)t[j] +
				(unsigned __int128)m * (unsigned __int128)modulus_limb<Params>(j) +
				carry;
			uint64_t word = (uint64_t)uv;
			carry = uv >> 64;
			if(j > 0) t[j - 1] = word;
		}
		top = (unsigned __int128)t[Params::LIMBS] + carry;
		t[Params::LIMBS - 1] = (uint64_t)top;
		t[Params::LIMBS] = (uint64_t)(top >> 64);
	}

	uint64_t candidate[Params::LIMBS];
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) candidate[i] = t[i];

	uint64_t reduced[Params::LIMBS];
	uint64_t borrow = subtract_modulus<Params>(reduced, candidate);
	bool use_reduced = t[Params::LIMBS] != 0 || borrow == 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		r[i] = use_reduced ? reduced[i] : candidate[i];
	}
}

template <typename Params>
__device__ __forceinline__ void zero(uint64_t r[Params::LIMBS]) {
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) r[i] = 0;
}

template <typename Params>
__device__ __forceinline__ void one(uint64_t r[Params::LIMBS]) {
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) r[i] = one_limb<Params>(i);
}

template <typename Params>
__device__ __forceinline__ void set(uint64_t r[Params::LIMBS], const uint64_t a[Params::LIMBS]) {
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) r[i] = a[i];
}

template <typename Params>
__device__ __forceinline__ bool is_zero(const uint64_t a[Params::LIMBS]) {
	uint64_t acc = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) acc |= a[i];
	return acc == 0;
}

template <typename Params>
__device__ __forceinline__ bool equal(const uint64_t a[Params::LIMBS], const uint64_t b[Params::LIMBS]) {
	uint64_t acc = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) acc |= a[i] ^ b[i];
	return acc == 0;
}

template <typename Params>
__device__ __forceinline__ void double_element(uint64_t r[Params::LIMBS],
                                               const uint64_t a[Params::LIMBS]) {
	add<Params>(r, a, a);
}

template <typename Params>
__device__ __forceinline__ void square(uint64_t r[Params::LIMBS],
                                       const uint64_t a[Params::LIMBS]) {
	mul<Params>(r, a, a);
}

template <typename Params>
__device__ __forceinline__ void neg(uint64_t r[Params::LIMBS],
                                    const uint64_t a[Params::LIMBS]) {
	if(is_zero<Params>(a)) {
		zero<Params>(r);
		return;
	}

	uint64_t borrow = 0;
#pragma unroll
	for(int i = 0; i < Params::LIMBS; i++) {
		r[i] = sub_borrow(modulus_limb<Params>(i), a[i], borrow);
	}
}

} // namespace gnark_gpu::plonk2
