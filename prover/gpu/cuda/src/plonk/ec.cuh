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

} // namespace gnark_gpu
