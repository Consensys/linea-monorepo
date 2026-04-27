#pragma once

// ─────────────────────────────────────────────────────────────────────────────
// Twisted Edwards elliptic curve arithmetic for BLS12-377 G1
//
// BLS12-377 G1 in Short Weierstrass form: y² = x³ + 1
// Maps birationally to Twisted Edwards:   -x² + y² = 1 + d·x²y²  (a = -1)
//
// Why Twisted Edwards for GPU MSM?
//   1. Unified addition formula: works for any two points (no special cases
//      for P+P, P+O, P+(-P)), avoiding warp divergence
//   2. No inversions in projective coordinates
//   3. Compact mixed-add: 9M for accumulator(extended) + point(affine)
//
// Two point representations:
//
//   G1EdExtended (192 bytes, accumulator):
//     (X, Y, T, Z) where x = X/Z, y = Y/Z, T = X·Y/Z
//     Identity: (0, R, 0, R) in Montgomery form
//     Coordinates live in [0, 2p) during computation (lazy reduction).
//     Must call ec_te_reduce() before exporting to host.
//
//   G1EdXY (96 bytes, compact input):
//     (x_te, y_te) affine TE coordinates only
//     The mixed-add formula computes T = 2d·x·y on the fly (2 extra fp_mul)
//     33% less memory than precomputed (y-x, y+x, 2dxy) format (144 bytes)
//     At large sizes (64M+ points), memory bandwidth dominates → 21-23% faster
//     Coordinates in [0, p) (fully reduced, loaded from host).
//
// Lazy reduction strategy:
//   All EC formulas use fp_mul_nr, fp_add_nr, fp_sub_nr internally.
//   Coordinates stay in [0, 2p) across chained additions — see bound table in fp.cuh.
//   This saves 9×12 = 108 instructions per EC add from skipped fp_mul reductions,
//   plus 4×12 = 48 from skipped fp_add reductions.  Total ~156 instr/add saved.
//
// Point addition cost:
//   Mixed add (Extended += XY):    9M  (2M for T_q, 3M for A/B, 1M for C, 3M for X3/Y3/T3/Z3)
//   General add (Extended += Ext): 9M  (1M+1M for C with 2d, 1M for D, rest same)
//   Doubling (host only):          4S + 4M  (dbl-2008-hwcd formula)
// ─────────────────────────────────────────────────────────────────────────────

#include "fp.cuh"

namespace gnark_gpu {

// =============================================================================
// Twisted Edwards types and constants
// =============================================================================

// 2d coefficient for the Twisted Edwards curve (Montgomery form, from gbotrel/zprize-mobile-harness)
__device__ __constant__ const uint64_t TE_D_COEFF_DOUBLE[6] = {
	0xf24b7e8444a706c6ULL, 0xeae0237580faa8faULL, 0x0f4d7cf27ef38fa5ULL,
	0x5597097dc5f2bb26ULL, 0x8bf6c1dd0d95a93eULL, 0x01784602fbff628aULL,
};

// Extended Twisted Edwards point: (X, Y, T, Z) — 192 bytes, accumulator
// Represents affine (X/Z, Y/Z) with T = X*Y/Z
// Identity: X=0, Y=R, T=0, Z=R (Montgomery one)
struct G1EdExtended {
	uint64_t x[6];
	uint64_t y[6];
	uint64_t t[6];
	uint64_t z[6];
};

// Compact Twisted Edwards MSM point format: 96 bytes per point
// Stores only (x_te, y_te). The mixed add formula computes T = 2d*x*y on the fly.
// 33% smaller than the precomputed (y-x, y+x, 2dxy) format at the cost of 2 extra fp_mul.
struct G1EdXY {
	uint64_t x[6]; // x_te (affine TE x-coordinate)
	uint64_t y[6]; // y_te (affine TE y-coordinate)
};

// Set TE extended point to identity: (0, R, 0, R)
__device__ __forceinline__ void ec_te_set_identity(G1EdExtended &p) {
	fp_set_zero(p.x);
	fp_set_one(p.y);
	fp_set_zero(p.t);
	fp_set_one(p.z);
}

// Branchless conditional negate of TE XY point x-coordinate.
// If negate==true: p.x = p - p.x.  If negate==false: p.x unchanged.
// Avoids warp divergence in MSM accumulate (signs are random ~50/50).
// Input p.x must be in [0, p) (loaded from device memory, fully reduced).
__device__ __forceinline__ void ec_te_cnegate_xy(G1EdXY &p, bool negate) {
	const uint64_t *mod = FP_MODULUS;
	uint64_t neg[6];
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(neg[0]) : "l"(mod[0]), "l"(p.x[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[1]) : "l"(mod[1]), "l"(p.x[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[2]) : "l"(mod[2]), "l"(p.x[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[3]) : "l"(mod[3]), "l"(p.x[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[4]) : "l"(mod[4]), "l"(p.x[4]));
	asm volatile("subc.u64 %0, %1, %2;" : "=l"(neg[5]) : "l"(mod[5]), "l"(p.x[5]));
	fp_ccopy(p.x, neg, negate);
}

// Reduce all coordinates of a TE extended point from [0, 2p) to [0, p).
// Call before exporting to host (e.g., MSM window results for Horner combination).
__device__ __forceinline__ void ec_te_reduce(G1EdExtended &p) {
	fp_reduce(p.x);
	fp_reduce(p.y);
	fp_reduce(p.t);
	fp_reduce(p.z);
}

// =============================================================================
// Unified mixed addition: G1EdExtended += G1EdXY (9M, strongly unified)
//
// Uses lazy reduction: all fp_mul_nr, fp_add_nr, fp_sub_nr.
// Accumulator coordinates in [0, 2p), point coordinates in [0, p).
// Output coordinates in [0, 2p) — invariant maintained across chained adds.
//
// q is in compact XY form: (x_te, y_te). We compute T = 2d*xq*yq on the fly.
//
// Formula (EFD madd-2008-hwcd-2, adapted for a=-1, on-the-fly T):
//   T_q = 2d * q.x * q.y       [1M+2M]
//   A = (Y1-X1) * (Y_q - X_q)  [3M]       A ∈ [0, 2p)
//   B = (Y1+X1) * (Y_q + X_q)  [4M]       B ∈ [0, 2p)
//   C = T1 * T_q                [5M]       C ∈ [0, 2p)
//   D = 2 * Z1                             D ∈ [0, 4p)
//   E = B - A                              E ∈ [0, 4p)
//   H = B + A                              H ∈ [0, 4p)
//   F = D - C                              F ∈ [0, 6p)
//   G = D + C                              G ∈ [0, 6p)
//   X3 = E * F  → [0, 2p)      [6M]
//   Y3 = G * H  → [0, 2p)      [7M]
//   T3 = E * H  → [0, 2p)      [8M]
//   Z3 = F * G  → [0, 2p)      [9M]
// =============================================================================

__device__ __forceinline__ void ec_te_unified_mixed_add_xy(G1EdExtended &p, const G1EdXY &q) {
	// T_q = 2d * xq * yq  (on the fly, q coords in [0, p))
	uint64_t T_q[6];
	fp_mul_nr(T_q, q.x, q.y);               // [0, 2p)
	fp_mul_nr(T_q, T_q, TE_D_COEFF_DOUBLE); // [0, 2p)

	// A = (Y1-X1) * (Yq - Xq)
	uint64_t A[6], t1[6];
	fp_sub_nr(A, p.y, p.x);                 // p coords [0, 2p) → A ∈ [0, 4p)
	fp_sub(t1, q.y, q.x);                   // q coords [0, p)  → t1 ∈ [0, p)
	fp_mul_nr(A, A, t1);                     // [0, 2p)

	// B = (Y1+X1) * (Yq + Xq)
	uint64_t B[6];
	fp_add_nr(B, p.y, p.x);                 // [0, 4p)
	fp_add_nr(t1, q.y, q.x);               // [0, 2p)
	fp_mul_nr(B, B, t1);                     // [0, 2p)

	// C = T1 * T_q
	uint64_t C[6];
	fp_mul_nr(C, p.t, T_q);                 // [0, 2p)

	// D = 2 * Z1
	uint64_t D[6];
	fp_add_nr(D, p.z, p.z);                 // [0, 4p)

	// E = B - A,  H = B + A     (both [0, 2p) inputs)
	uint64_t E[6], H[6];
	fp_sub_nr(E, B, A);                     // [0, 4p)
	fp_add_nr(H, B, A);                     // [0, 4p)

	// F = D - C,  G = D + C     (D [0, 4p), C [0, 2p))
	uint64_t F[6], G[6];
	fp_sub_nr(F, D, C);                     // [0, 6p)
	fp_add_nr(G, D, C);                     // [0, 6p)

	// Final products: all outputs ∈ [0, 2p) since max input 6p < R
	fp_mul_nr(p.x, E, F);                   // [0, 2p)
	fp_mul_nr(p.y, G, H);                   // [0, 2p)
	fp_mul_nr(p.t, E, H);                   // [0, 2p)
	fp_mul_nr(p.z, F, G);                   // [0, 2p)
}

// =============================================================================
// Unified general addition: G1EdExtended += G1EdExtended (9M, strongly unified)
//
// Both operands have coordinates in [0, 2p). Same lazy reduction strategy.
//
// Formula (EFD add-2008-hwcd, for a=-1):
//   A = (Y1-X1) * (Y2-X2)      [1M]       A ∈ [0, 2p)
//   B = (Y1+X1) * (Y2+X2)      [2M]       B ∈ [0, 2p)
//   C = T1 * 2d * T2            [3M+4M]    C ∈ [0, 2p)
//   D = 2 * Z1 * Z2             [5M]       D ∈ [0, 4p)
//   E = B - A                              E ∈ [0, 4p)
//   H = B + A                              H ∈ [0, 4p)
//   F = D - C                              F ∈ [0, 6p)
//   G = D + C                              G ∈ [0, 6p)
//   X3 = E * F  → [0, 2p)      [6M]
//   Y3 = G * H  → [0, 2p)      [7M]
//   T3 = E * H  → [0, 2p)      [8M]
//   Z3 = F * G  → [0, 2p)      [9M]
// =============================================================================

__device__ __forceinline__ void ec_te_unified_add(G1EdExtended &p, const G1EdExtended &q) {
	// A = (Y1-X1) * (Y2-X2)
	uint64_t A[6], t1[6];
	fp_sub_nr(A, p.y, p.x);                 // [0, 4p)
	fp_sub_nr(t1, q.y, q.x);               // [0, 4p)
	fp_mul_nr(A, A, t1);                     // [0, 2p)

	// B = (Y1+X1) * (Y2+X2)
	uint64_t B[6];
	fp_add_nr(B, p.y, p.x);                 // [0, 4p)
	fp_add_nr(t1, q.y, q.x);               // [0, 4p)
	fp_mul_nr(B, B, t1);                     // [0, 2p)

	// C = T1 * T2 * 2d
	uint64_t C[6];
	fp_mul_nr(C, p.t, q.t);                 // [0, 2p)
	fp_mul_nr(C, C, TE_D_COEFF_DOUBLE);     // [0, 2p)

	// D = 2 * Z1 * Z2
	uint64_t D[6];
	fp_mul_nr(D, p.z, q.z);                 // [0, 2p)
	fp_add_nr(D, D, D);                     // [0, 4p)

	// E = B - A, H = B + A
	uint64_t E[6], H[6];
	fp_sub_nr(E, B, A);                     // [0, 4p)
	fp_add_nr(H, B, A);                     // [0, 4p)

	// F = D - C, G = D + C
	uint64_t F[6], G[6];
	fp_sub_nr(F, D, C);                     // [0, 6p)
	fp_add_nr(G, D, C);                     // [0, 6p)

	// Final products
	fp_mul_nr(p.x, E, F);                   // [0, 2p)
	fp_mul_nr(p.y, G, H);                   // [0, 2p)
	fp_mul_nr(p.t, E, H);                   // [0, 2p)
	fp_mul_nr(p.z, F, G);                   // [0, 2p)
}

// =============================================================================
// Precomputed Twisted Edwards mixed-add input format (G1EdYZD)
//
// Stores per-point precomputed (Y-X, Y+X, 2d·X·Y) — three Fp coords (144 B
// total, vs 96 B for compact G1EdXY). The mixed-add formula then drops the
// on-the-fly T_q = 2d·X·Y computation, saving 2 fp_mul per add (9M → 7M).
//
// Tradeoff: 50% larger point memory; but for compute-bound accumulate phases
// at moderate n (≲ 2²⁵ on Blackwell), the saved muls dominate the extra
// bandwidth. See WORKLOG.md for measurements.
//
// Coordinates are loaded from device memory in [0, p) (fully reduced by the
// host conversion). Output of mixed-add stays in [0, 2p) — same lazy
// reduction discipline as G1EdXY.
// =============================================================================

struct G1EdYZD {
	uint64_t y_minus_x[6];  // (Y_te - X_te) mod p
	uint64_t y_plus_x[6];   // (Y_te + X_te) mod p
	uint64_t two_d_xy[6];   // (2d * X_te * Y_te) mod p
};

// Branchless conditional negate of a precomputed point.
//
// Negating a TE point (X, Y) → (-X, Y) corresponds in precomputed format to:
//   y_minus_x ↔ y_plus_x  (swap), two_d_xy → -two_d_xy (mod p).
//
// We swap by computing both candidate values (no-op or negate) and picking
// branchlessly with fp_ccopy. Saves warp divergence in MSM accumulate.
__device__ __forceinline__ void ec_te_cnegate_yzd(G1EdYZD &p, bool negate) {
	// Snapshot y_minus_x and y_plus_x before swap.
	uint64_t orig_minus[6];
#pragma unroll
	for(int i = 0; i < 6; i++) orig_minus[i] = p.y_minus_x[i];

	// If negate: y_minus_x = y_plus_x_old; y_plus_x = y_minus_x_old.
	fp_ccopy(p.y_minus_x, p.y_plus_x, negate);
	fp_ccopy(p.y_plus_x, orig_minus, negate);

	// Negate two_d_xy: candidate = p - two_d_xy (mod p).
	const uint64_t *mod = FP_MODULUS;
	uint64_t neg[6];
	asm volatile("sub.cc.u64 %0, %1, %2;" : "=l"(neg[0]) : "l"(mod[0]), "l"(p.two_d_xy[0]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[1]) : "l"(mod[1]), "l"(p.two_d_xy[1]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[2]) : "l"(mod[2]), "l"(p.two_d_xy[2]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[3]) : "l"(mod[3]), "l"(p.two_d_xy[3]));
	asm volatile("subc.cc.u64 %0, %1, %2;" : "=l"(neg[4]) : "l"(mod[4]), "l"(p.two_d_xy[4]));
	asm volatile("subc.u64 %0, %1, %2;" : "=l"(neg[5]) : "l"(mod[5]), "l"(p.two_d_xy[5]));
	fp_ccopy(p.two_d_xy, neg, negate);
}

// =============================================================================
// Unified mixed addition with precomputed input: G1EdExtended += G1EdYZD (7M)
//
// Same lazy-reduction strategy as ec_te_unified_mixed_add_xy, two muls
// cheaper because T_q is precomputed.
//
// Formula (madd-2008-hwcd, a=-1, with precomputed q):
//   A = (Y1-X1) * y_minus_x   [1M]   A ∈ [0, 2p)
//   B = (Y1+X1) * y_plus_x    [2M]   B ∈ [0, 2p)
//   C = T1 * two_d_xy          [3M]   C ∈ [0, 2p)
//   D = 2 * Z1                         D ∈ [0, 4p)
//   E = B - A;  H = B + A
//   F = D - C;  G = D + C
//   X3 = E * F  → [0, 2p)      [4M]
//   Y3 = G * H  → [0, 2p)      [5M]
//   T3 = E * H  → [0, 2p)      [6M]
//   Z3 = F * G  → [0, 2p)      [7M]
// =============================================================================
__device__ __forceinline__ void ec_te_unified_mixed_add_yzd(G1EdExtended &p, const G1EdYZD &q) {
	// A = (Y1 - X1) * (Y_q - X_q)
	uint64_t A[6];
	fp_sub_nr(A, p.y, p.x);                  // [0, 4p)
	fp_mul_nr(A, A, q.y_minus_x);            // [0, 2p)

	// B = (Y1 + X1) * (Y_q + X_q)
	uint64_t B[6];
	fp_add_nr(B, p.y, p.x);                  // [0, 4p)
	fp_mul_nr(B, B, q.y_plus_x);             // [0, 2p)

	// C = T1 * (2d * X_q * Y_q)
	uint64_t C[6];
	fp_mul_nr(C, p.t, q.two_d_xy);           // [0, 2p)

	// D = 2 * Z1
	uint64_t D[6];
	fp_add_nr(D, p.z, p.z);                  // [0, 4p)

	// E = B - A,  H = B + A    (both [0, 2p) inputs)
	uint64_t E[6], H[6];
	fp_sub_nr(E, B, A);                      // [0, 4p)
	fp_add_nr(H, B, A);                      // [0, 4p)

	// F = D - C,  G = D + C    (D [0, 4p), C [0, 2p))
	uint64_t F[6], G[6];
	fp_sub_nr(F, D, C);                      // [0, 6p)
	fp_add_nr(G, D, C);                      // [0, 6p)

	// Final products: outputs ∈ [0, 2p) since max input 6p < R.
	fp_mul_nr(p.x, E, F);                    // [0, 2p)
	fp_mul_nr(p.y, G, H);                    // [0, 2p)
	fp_mul_nr(p.t, E, H);                    // [0, 2p)
	fp_mul_nr(p.z, F, G);                    // [0, 2p)
}

// =============================================================================
// Short Weierstrass G1 affine arithmetic for batched-affine MSM.
//
// BLS12-377 G1 in SW form: y² = x³ + 1   (a=0, b=1).
//
// We use SW affine for the batched bucket-accumulation phase (see msm.cu).
// Each pair-add (P0 + P1) costs 1S + 3M *given* a precomputed 1/(x1-x0).
// Across N pairs in a batch, Montgomery's trick amortizes a single inversion
// with 3N field multiplications, so per-pair effective cost is 1S + 6M.
//
// The compact format matches gnark-crypto's bls12377.G1Affine memory layout
// (12 limbs in Montgomery form). Identity is encoded as (0, 0) — distinct
// from any on-curve point since 0² = 0 ≠ 1 = 0³ + 1.
//
// Layout invariants:
//   p.x, p.y are fully reduced (in [0, p)). Outputs of g1sw_pair_add likewise.
// =============================================================================

struct G1AffineSW {
	uint64_t x[6];
	uint64_t y[6];
};

__device__ __forceinline__ void g1sw_set_identity(G1AffineSW &p) {
	fp_set_zero(p.x);
	fp_set_zero(p.y);
}

__device__ __forceinline__ bool g1sw_is_identity(const G1AffineSW &p) {
	return fp_is_zero(p.x) && fp_is_zero(p.y);
}

__device__ __forceinline__ void g1sw_neg(G1AffineSW &out, const G1AffineSW &p) {
	fp_copy(out.x, p.x);
	fp_negate(out.y, p.y);
}

__device__ __forceinline__ void g1sw_cnegate(G1AffineSW &p, bool negate) {
	uint64_t neg_y[6];
	fp_negate(neg_y, p.y);
	fp_ccopy(p.y, neg_y, negate);
}

// =============================================================================
// SW affine point doubling (a=0, b=1).
//
//   λ = (3·x²) / (2·y)
//   x3 = λ² - 2x
//   y3 = λ(x - x3) - y
//
// Caller passes the precomputed inv2y = 1/(2y). Cost: 1S + 3M (after λ).
// Used rarely in batched-affine since random-scalar MSM almost never sees
// repeated points; included for completeness.
// =============================================================================
__device__ __forceinline__ void g1sw_double_with_inv2y(
	G1AffineSW &out, const G1AffineSW &p, const uint64_t inv2y[6]) {

	uint64_t three_x_sq[6], x_sq[6], two_x_sq[6];
	fp_sqr(x_sq, p.x);
	fp_add(two_x_sq, x_sq, x_sq);
	fp_add(three_x_sq, two_x_sq, x_sq);

	uint64_t lambda[6];
	fp_mul(lambda, three_x_sq, inv2y);

	uint64_t lam_sq[6], two_x[6];
	fp_sqr(lam_sq, lambda);
	fp_add(two_x, p.x, p.x);

	uint64_t x3[6];
	fp_sub(x3, lam_sq, two_x);

	uint64_t x_minus_x3[6], lam_dx[6], y3[6];
	fp_sub(x_minus_x3, p.x, x3);
	fp_mul(lam_dx, lambda, x_minus_x3);
	fp_sub(y3, lam_dx, p.y);

	fp_copy(out.x, x3);
	fp_copy(out.y, y3);
}

// =============================================================================
// SW affine pair add given precomputed 1/(x1-x0).
//
//   λ = (y1 - y0) · inv_dx
//   x3 = λ² - x0 - x1
//   y3 = λ(x0 - x3) - y0
//
// Cost: 1S + 3M (the λ multiply is one of the 3M).
//
// Special cases:
//   - Either operand is identity: result = the non-identity operand.
//   - x0 == x1, y0 == y1 (P + P): would need doubling; the precomputed
//     inv_dx is undefined. Caller must detect ahead and dispatch to double.
//   - x0 == x1, y0 != y1 (P + (-P)): result = identity.
// For random-scalar MSM with sorted-by-bucket pairs, neither degenerate case
// occurs (each pair is two distinct point indices contributing to the same
// bucket — almost surely different x's).
// =============================================================================
__device__ __forceinline__ void g1sw_pair_add_with_inv_dx(
	G1AffineSW &out, const G1AffineSW &p0, const G1AffineSW &p1,
	const uint64_t inv_dx[6]) {

	// λ = (y1 - y0) · inv_dx
	uint64_t dy[6];
	fp_sub(dy, p1.y, p0.y);
	uint64_t lambda[6];
	fp_mul(lambda, dy, inv_dx);

	// x3 = λ² - x0 - x1
	uint64_t lam_sq[6], x3[6];
	fp_sqr(lam_sq, lambda);
	uint64_t sum_x[6];
	fp_add(sum_x, p0.x, p1.x);
	fp_sub(x3, lam_sq, sum_x);

	// y3 = λ(x0 - x3) - y0
	uint64_t x0_minus_x3[6], lam_dx[6], y3[6];
	fp_sub(x0_minus_x3, p0.x, x3);
	fp_mul(lam_dx, lambda, x0_minus_x3);
	fp_sub(y3, lam_dx, p0.y);

	fp_copy(out.x, x3);
	fp_copy(out.y, y3);
}

// =============================================================================
// SW affine → TE extended conversion.
//
// Used at the boundary between the new batched-affine accumulator (output:
// per-bucket SW affine point) and the existing reduce phase (input: per-bucket
// G1EdExtended). Mirrors `convertToEdMSM` from g1_te.go but for a single point
// at a time (no batched inversion — the single fp_inv is amortized over the
// many adds whose sum produced this point).
//
// Mapping (gbotrel/zprize-mobile-harness):
//   x_te = (x_sw + 1) / (y_sw · invSqrtMinusA)
//   y_te = (x_sw + 1 - √3) / (x_sw + 1 + √3)
//
// Identity SW (0, 0) → identity TE extended (0, 1, 0, 1).
// =============================================================================

// 1/√(-a) where a = -2√3 + 3, in Montgomery form (matches teInvSqrtMinusA in g1_te.go).
__device__ __constant__ const uint64_t TE_INV_SQRT_MINUS_A[6] = {
	0x3b092ce1fd76a6bdULL, 0x925230d9bba32683ULL,
	0x872d5d2fe991a197ULL, 0x8367c527a82b2ab0ULL,
	0xe285bbb3ef662a15ULL, 0x0160527a9283e729ULL,
};

// √3 in Montgomery form (matches teSqrtThree in g1_te.go).
__device__ __constant__ const uint64_t TE_SQRT_THREE[6] = {
	0x3fabdfd08894e1e4ULL, 0xcbf921ddcc1f55aaULL,
	0xd17deff1460edc0cULL, 0xd394e81e7897028dULL,
	0xc29c995d0912681aULL, 0x01515e6caff9d568ULL,
};

__device__ __forceinline__ void g1sw_to_te_extended(
	G1EdExtended &out, const G1AffineSW &p) {

	if(g1sw_is_identity(p)) {
		ec_te_set_identity(out);
		return;
	}

	uint64_t one[6];
	fp_set_one(one);

	uint64_t x_plus_one[6];
	fp_add(x_plus_one, p.x, one);

	// Denominator 1: y_sw · invSqrtMinusA
	// Denominator 2: x_sw + 1 + √3
	uint64_t d1[6], d2[6];
	fp_mul(d1, p.y, TE_INV_SQRT_MINUS_A);
	fp_add(d2, x_plus_one, TE_SQRT_THREE);

	// Single batched inverse via the well-known: invert(d1*d2) once, recover
	// inv_d1 = d2 * inv(d1*d2), inv_d2 = d1 * inv(d1*d2).
	uint64_t prod[6], inv_prod[6];
	fp_mul(prod, d1, d2);
	fp_inv(inv_prod, prod);

	uint64_t inv_d1[6], inv_d2[6];
	fp_mul(inv_d1, d2, inv_prod);
	fp_mul(inv_d2, d1, inv_prod);

	// x_te = x_plus_one * inv_d1 ;  y_te = (x_plus_one - √3) * inv_d2
	uint64_t x_te[6], y_te[6], x_minus_sqrt3[6];
	fp_mul(x_te, x_plus_one, inv_d1);
	fp_sub(x_minus_sqrt3, x_plus_one, TE_SQRT_THREE);
	fp_mul(y_te, x_minus_sqrt3, inv_d2);

	// Pack into extended TE: (X=x_te, Y=y_te, T=x_te*y_te, Z=1).
	fp_copy(out.x, x_te);
	fp_copy(out.y, y_te);
	fp_mul(out.t, x_te, y_te);
	fp_set_one(out.z);
}

} // namespace gnark_gpu
