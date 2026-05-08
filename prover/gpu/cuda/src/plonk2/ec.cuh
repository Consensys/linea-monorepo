#pragma once

// Curve-generic short-Weierstrass elliptic-curve formulas for gpu/plonk2.
//
// All arithmetic is templated on a base-field Params struct (see field.cuh).
// We use the standard textbook formulas:
//   - Affine point AoS: [X.l0..lN, Y.l0..lN]
//   - Jacobian point:    (X, Y, Z) representing (X/Z^2, Y/Z^3)
//   - Infinity:          Jacobian Z == 0; affine (0, 0)
//
// The MSM uses the mixed Jacobian + affine addition (Z2 = 1 implicit).
// For affine-affine addition (used only by single-point validation kernels)
// we promote first.

#include "field.cuh"

#include <cstdint>

namespace gnark_gpu::plonk2 {

template <typename Fp>
struct AffinePoint {
	uint64_t x[Fp::LIMBS];
	uint64_t y[Fp::LIMBS];
};

template <typename Fp>
struct JacobianPoint {
	uint64_t x[Fp::LIMBS];
	uint64_t y[Fp::LIMBS];
	uint64_t z[Fp::LIMBS];
};

template <typename Fp>
__device__ __forceinline__ bool affine_is_infinity(const AffinePoint<Fp> &p) {
	return is_zero<Fp>(p.x) && is_zero<Fp>(p.y);
}

template <typename Fp>
__device__ __forceinline__ bool jacobian_is_infinity(const JacobianPoint<Fp> &p) {
	return is_zero<Fp>(p.z);
}

template <typename Fp>
__device__ __forceinline__ void jacobian_set_infinity(JacobianPoint<Fp> &p) {
	one<Fp>(p.x);
	one<Fp>(p.y);
	zero<Fp>(p.z);
}

template <typename Fp>
__device__ __forceinline__ void jacobian_from_affine(JacobianPoint<Fp> &out,
                                                     const AffinePoint<Fp> &p) {
	if(affine_is_infinity<Fp>(p)) {
		jacobian_set_infinity<Fp>(out);
		return;
	}
	set<Fp>(out.x, p.x);
	set<Fp>(out.y, p.y);
	one<Fp>(out.z);
}

// out = 2 * a, a in affine. Result in Jacobian.
// Standard "mdbl-2007-bl" formulas.
template <typename Fp>
__device__ __forceinline__ void jacobian_double_mixed(JacobianPoint<Fp> &out,
                                                      const AffinePoint<Fp> &a) {
	if(affine_is_infinity<Fp>(a) || is_zero<Fp>(a.y)) {
		jacobian_set_infinity<Fp>(out);
		return;
	}

	uint64_t xx[Fp::LIMBS], yy[Fp::LIMBS], yyyy[Fp::LIMBS];
	uint64_t s[Fp::LIMBS], m[Fp::LIMBS], t[Fp::LIMBS];
	uint64_t tmp[Fp::LIMBS];

	square<Fp>(xx, a.x);
	square<Fp>(yy, a.y);
	square<Fp>(yyyy, yy);

	add<Fp>(s, a.x, yy);
	square<Fp>(s, s);
	sub<Fp>(s, s, xx);
	sub<Fp>(s, s, yyyy);
	double_element<Fp>(s, s);

	double_element<Fp>(m, xx);
	add<Fp>(m, m, xx);

	square<Fp>(t, m);
	sub<Fp>(t, t, s);
	sub<Fp>(t, t, s);

	set<Fp>(out.x, t);

	sub<Fp>(tmp, s, t);
	mul<Fp>(out.y, tmp, m);
	double_element<Fp>(yyyy, yyyy);
	double_element<Fp>(yyyy, yyyy);
	double_element<Fp>(yyyy, yyyy);
	sub<Fp>(out.y, out.y, yyyy);

	double_element<Fp>(out.z, a.y);
}

// out = a + b, both affine, distinct, neither infinity. Result in Jacobian.
// Standard "mmadd-2007-bl" plus fallbacks for inf and a==b cases.
template <typename Fp>
__device__ __forceinline__ void jacobian_add_affine_affine(JacobianPoint<Fp> &out,
                                                           const AffinePoint<Fp> &a,
                                                           const AffinePoint<Fp> &b) {
	if(affine_is_infinity<Fp>(a)) {
		jacobian_from_affine<Fp>(out, b);
		return;
	}
	if(affine_is_infinity<Fp>(b)) {
		jacobian_from_affine<Fp>(out, a);
		return;
	}
	if(equal<Fp>(a.x, b.x)) {
		if(equal<Fp>(a.y, b.y)) {
			jacobian_double_mixed<Fp>(out, a);
		} else {
			jacobian_set_infinity<Fp>(out);
		}
		return;
	}

	uint64_t h[Fp::LIMBS], hh[Fp::LIMBS], i[Fp::LIMBS];
	uint64_t j[Fp::LIMBS], r[Fp::LIMBS], v[Fp::LIMBS];
	uint64_t tmp[Fp::LIMBS];

	sub<Fp>(h, b.x, a.x);
	square<Fp>(hh, h);
	double_element<Fp>(i, hh);
	double_element<Fp>(i, i);
	mul<Fp>(j, h, i);
	sub<Fp>(r, b.y, a.y);
	double_element<Fp>(r, r);
	mul<Fp>(v, a.x, i);

	square<Fp>(out.x, r);
	sub<Fp>(out.x, out.x, j);
	sub<Fp>(out.x, out.x, v);
	sub<Fp>(out.x, out.x, v);

	sub<Fp>(tmp, v, out.x);
	mul<Fp>(out.y, tmp, r);
	mul<Fp>(j, a.y, j);
	double_element<Fp>(j, j);
	sub<Fp>(out.y, out.y, j);

	double_element<Fp>(out.z, h);
}

// out = a + b, a Jacobian, b affine. Result Jacobian.
// Standard "madd-2007-bl" with fallbacks.
template <typename Fp>
__device__ __forceinline__ void jacobian_add_jacobian_affine(JacobianPoint<Fp> &out,
                                                             const JacobianPoint<Fp> &a,
                                                             const AffinePoint<Fp> &b) {
	if(jacobian_is_infinity<Fp>(a)) {
		jacobian_from_affine<Fp>(out, b);
		return;
	}
	if(affine_is_infinity<Fp>(b)) {
		set<Fp>(out.x, a.x);
		set<Fp>(out.y, a.y);
		set<Fp>(out.z, a.z);
		return;
	}

	uint64_t z1z1[Fp::LIMBS], u2[Fp::LIMBS], s2[Fp::LIMBS];
	uint64_t h[Fp::LIMBS], hh[Fp::LIMBS], i[Fp::LIMBS], j[Fp::LIMBS];
	uint64_t r[Fp::LIMBS], v[Fp::LIMBS], tmp[Fp::LIMBS];

	square<Fp>(z1z1, a.z);
	mul<Fp>(u2, b.x, z1z1);
	mul<Fp>(s2, b.y, a.z);
	mul<Fp>(s2, s2, z1z1);

	if(equal<Fp>(a.x, u2)) {
		if(equal<Fp>(a.y, s2)) {
			jacobian_double_mixed<Fp>(out, b);
			return;
		}
		jacobian_set_infinity<Fp>(out);
		return;
	}

	sub<Fp>(h, u2, a.x);
	square<Fp>(hh, h);
	double_element<Fp>(i, hh);
	double_element<Fp>(i, i);
	mul<Fp>(j, h, i);

	sub<Fp>(r, s2, a.y);
	double_element<Fp>(r, r);
	mul<Fp>(v, a.x, i);

	square<Fp>(out.x, r);
	sub<Fp>(out.x, out.x, j);
	sub<Fp>(out.x, out.x, v);
	sub<Fp>(out.x, out.x, v);

	sub<Fp>(tmp, v, out.x);
	mul<Fp>(out.y, tmp, r);
	mul<Fp>(j, a.y, j);
	double_element<Fp>(j, j);
	sub<Fp>(out.y, out.y, j);

	add<Fp>(out.z, a.z, h);
	square<Fp>(out.z, out.z);
	sub<Fp>(out.z, out.z, z1z1);
	sub<Fp>(out.z, out.z, hh);
}

// out = 2a, a Jacobian.
// Standard "dbl-2009-l" (works for a != b == 0 curves, BN/BLS/BW6 are b != 0
// but the formula does not depend on the curve coefficient since a' == 0
// for these short-Weierstrass curves with Y^2 = X^3 + b).
template <typename Fp>
__device__ __forceinline__ void jacobian_double(JacobianPoint<Fp> &out,
                                                const JacobianPoint<Fp> &a) {
	if(jacobian_is_infinity<Fp>(a)) {
		jacobian_set_infinity<Fp>(out);
		return;
	}

	uint64_t a2[Fp::LIMBS], b2[Fp::LIMBS], c2[Fp::LIMBS];
	uint64_t d2[Fp::LIMBS], e2[Fp::LIMBS], f2[Fp::LIMBS], tmp[Fp::LIMBS];

	square<Fp>(a2, a.x);
	square<Fp>(b2, a.y);
	square<Fp>(c2, b2);

	add<Fp>(d2, a.x, b2);
	square<Fp>(d2, d2);
	sub<Fp>(d2, d2, a2);
	sub<Fp>(d2, d2, c2);
	double_element<Fp>(d2, d2);

	double_element<Fp>(e2, a2);
	add<Fp>(e2, e2, a2);
	square<Fp>(f2, e2);

	sub<Fp>(out.x, f2, d2);
	sub<Fp>(out.x, out.x, d2);

	sub<Fp>(tmp, d2, out.x);
	mul<Fp>(out.y, tmp, e2);
	double_element<Fp>(c2, c2);
	double_element<Fp>(c2, c2);
	double_element<Fp>(c2, c2);
	sub<Fp>(out.y, out.y, c2);

	mul<Fp>(out.z, a.y, a.z);
	double_element<Fp>(out.z, out.z);
}

// out = a + b, both Jacobian.
// Standard "add-2007-bl" with fallbacks.
template <typename Fp>
__device__ __forceinline__ void jacobian_add(JacobianPoint<Fp> &out,
                                             const JacobianPoint<Fp> &a,
                                             const JacobianPoint<Fp> &b) {
	if(jacobian_is_infinity<Fp>(a)) {
		set<Fp>(out.x, b.x);
		set<Fp>(out.y, b.y);
		set<Fp>(out.z, b.z);
		return;
	}
	if(jacobian_is_infinity<Fp>(b)) {
		set<Fp>(out.x, a.x);
		set<Fp>(out.y, a.y);
		set<Fp>(out.z, a.z);
		return;
	}

	uint64_t z1z1[Fp::LIMBS], z2z2[Fp::LIMBS];
	uint64_t u1[Fp::LIMBS], u2[Fp::LIMBS], s1[Fp::LIMBS], s2[Fp::LIMBS];
	uint64_t h[Fp::LIMBS], i[Fp::LIMBS], j[Fp::LIMBS], r[Fp::LIMBS], v[Fp::LIMBS];
	uint64_t tmp[Fp::LIMBS];

	square<Fp>(z1z1, a.z);
	square<Fp>(z2z2, b.z);
	mul<Fp>(u1, a.x, z2z2);
	mul<Fp>(u2, b.x, z1z1);
	mul<Fp>(s1, a.y, b.z);
	mul<Fp>(s1, s1, z2z2);
	mul<Fp>(s2, b.y, a.z);
	mul<Fp>(s2, s2, z1z1);

	if(equal<Fp>(u1, u2)) {
		if(equal<Fp>(s1, s2)) {
			jacobian_double<Fp>(out, a);
			return;
		}
		jacobian_set_infinity<Fp>(out);
		return;
	}

	sub<Fp>(h, u2, u1);
	double_element<Fp>(i, h);
	square<Fp>(i, i);
	mul<Fp>(j, h, i);
	sub<Fp>(r, s2, s1);
	double_element<Fp>(r, r);
	mul<Fp>(v, u1, i);

	square<Fp>(out.x, r);
	sub<Fp>(out.x, out.x, j);
	sub<Fp>(out.x, out.x, v);
	sub<Fp>(out.x, out.x, v);

	sub<Fp>(tmp, v, out.x);
	mul<Fp>(out.y, tmp, r);
	mul<Fp>(j, s1, j);
	double_element<Fp>(j, j);
	sub<Fp>(out.y, out.y, j);

	add<Fp>(out.z, a.z, b.z);
	square<Fp>(out.z, out.z);
	sub<Fp>(out.z, out.z, z1z1);
	sub<Fp>(out.z, out.z, z2z2);
	mul<Fp>(out.z, out.z, h);
}

// out = -a (negate Y).
template <typename Fp>
__device__ __forceinline__ void jacobian_neg(JacobianPoint<Fp> &out,
                                             const JacobianPoint<Fp> &a) {
	set<Fp>(out.x, a.x);
	neg<Fp>(out.y, a.y);
	set<Fp>(out.z, a.z);
}

} // namespace gnark_gpu::plonk2
