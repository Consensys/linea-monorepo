//go:build cuda

package plonk

// BLS12-377 Twisted Edwards (TE) curve: -x² + y² = 1 + d·x²·y²
//
// The TE form has parameter a = -1, which enables efficient "a=-1" formulas.
// Extended coordinates (X, Y, T, Z) with x = X/Z, y = Y/Z, T = X·Y/Z.
//
// Compact GPU format G1EdXY: stores only (x_te, y_te) in Fp Montgomery form.
// The GPU accumulate kernel computes T = 2d·x·y on the fly during mixed
// addition (9M formula), trading 2 extra Fp multiplications for 33% less
// memory bandwidth (96 bytes vs 144 bytes per point).
//
// SW ↔ TE isogeny (Short Weierstrass to Twisted Edwards):
//
//	Given SW point (x_sw, y_sw) on y² = x³ + ax + b:
//
//	x_te = (x_sw + 1) / (y_sw · invSqrtMinusA)
//	y_te = (x_sw + 1 - √3) / (x_sw + 1 + √3)
//
//	The conversion uses Montgomery batch inversion to amortize the
//	two per-point divisions across all n points.
//
// TE → SW reverse mapping (from extended projective coordinates X, Y, T, Z):
//
//	y_te_aff = Y/Z
//	n        = (1 + y_te_aff) · Z/(Z - Y) · √3
//	x_sw     = n - 1
//	y_sw     = n · Z / (X · invSqrtMinusA)
//
// The reverse mapping uses batch inversion over {Z, Z-Y, X·invSqrtMinusA} to
// convert a single TE extended point back to Short Weierstrass affine/Jacobian.

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
)

// ---------------------------------------------------------------------------
// G1TEPoint type, constants, and magic header
// ---------------------------------------------------------------------------

// G1TEPoint represents a BLS12-377 G1 point in compact Twisted Edwards format,
// ready for GPU MSM. Layout: [x_te, y_te] each as 6 uint64s in Montgomery form.
// Total size: 96 bytes. This matches the GPU G1EdXY struct layout (AoS).
//
// AoS is optimal here because the GPU accumulate kernel accesses points by random
// index (from radix sort output), so each point load is one contiguous 96-byte fetch.
// SoA would require 2 separate random accesses and double the TLB misses.
type G1TEPoint [12]uint64

// g1TEPointSize is the size of a single G1TEPoint in bytes.
const g1TEPointSize = int(unsafe.Sizeof(G1TEPoint{})) // 96

// g1TEPrecompPoint is the GPU-resident precomputed Twisted Edwards format,
// matching the C struct G1EdYZD layout exactly:
//
//	limbs[0..6]   = (Y_te - X_te) mod p   (Montgomery form, fully reduced)
//	limbs[6..12]  = (Y_te + X_te) mod p
//	limbs[12..18] = (2d · X_te · Y_te) mod p
//
// Trades 50% more memory per point (144 vs 96 bytes) for 2 fewer fp_mul
// in the mixed-addition kernel (9M → 7M). Only used internally — the public
// G1TEPoint type and its serialization formats remain compact.
type g1TEPrecompPoint [18]uint64

// g1TEPrecompPointSize is the size of a g1TEPrecompPoint in bytes (144).
const g1TEPrecompPointSize = int(unsafe.Sizeof(g1TEPrecompPoint{}))

// g1TEMagic identifies the safe serialization format for G1TEPoint data.
var g1TEMagic = [8]byte{'G', 'N', 'R', 'K', 'T', 'E', '0', '2'}

// ---------------------------------------------------------------------------
// TE curve conversion constants
// ---------------------------------------------------------------------------

// Conversion constants for BLS12-377 G1 Short Weierstrass ↔ Twisted Edwards mapping.
// From gbotrel/zprize-mobile-harness, in Montgomery form.
var (
	// teSqrtThree is √3 in Montgomery form (from gbotrel/zprize-mobile-harness).
	teSqrtThree = fp.Element{
		0x3fabdfd08894e1e4, 0xcbf921ddcc1f55aa,
		0xd17deff1460edc0c, 0xd394e81e7897028d,
		0xc29c995d0912681a, 0x01515e6caff9d568,
	}

	// teInvSqrtMinusA is 1/√(-a) where a = -2√3 + 3 (from gbotrel/zprize-mobile-harness).
	// This constant appears in both the SW→TE forward mapping (dividing by y_sw · invSqrtMinusA)
	// and the TE→SW reverse mapping (dividing by X · invSqrtMinusA).
	teInvSqrtMinusA = fp.Element{
		0x3b092ce1fd76a6bd, 0x925230d9bba32683,
		0x872d5d2fe991a197, 0x8367c527a82b2ab0,
		0xe285bbb3ef662a15, 0x0160527a9283e729,
	}

	// teDCoeffDouble is 2d for the TE curve, where d = -(2√3 + 3)
	// (from gbotrel/zprize-mobile-harness). Used in the TE addition formula
	// for the C = T1 · T2 · 2d term.
	teDCoeffDouble = fp.Element{
		0xf24b7e8444a706c6, 0xeae0237580faa8fa,
		0x0f4d7cf27ef38fa5, 0x5597097dc5f2bb26,
		0x8bf6c1dd0d95a93e, 0x01784602fbff628a,
	}
)

// ---------------------------------------------------------------------------
// ConvertG1AffineToTE: public entry point
// ---------------------------------------------------------------------------

// ConvertG1AffineToTE converts Short Weierstrass affine points to precomputed
// Twisted Edwards format suitable for GPU MSM. This is the expensive CPU operation
// (batch inversion + field arithmetic). Call once, serialize the result for reuse.
func ConvertG1AffineToTE(points []bls12377.G1Affine) []G1TEPoint {
	return convertToEdMSM(points)
}

// ---------------------------------------------------------------------------
// convertToEdMSM: SW → TE batch conversion
// ---------------------------------------------------------------------------

// convertToEdMSM converts SW affine points to compact TE XY format.
// Uses Montgomery's trick for batch inversion to amortize the two per-point
// divisions across all n points. Follows gbotrel/zprize-mobile-harness
// BatchFromAffineSWC exactly.
//
// SW → TE mapping for each point (x_sw, y_sw):
//
//	Let x1 = x_sw + 1       (translate by SW-to-TE x-offset)
//
//	Denominator 1:  d1 = y_sw · invSqrtMinusA
//	Denominator 2:  d2 = x1 + √3
//
//	x_te = x1 / d1
//	     = (x_sw + 1) / (y_sw · invSqrtMinusA)
//
//	y_te = (x1 - √3) / d2
//	     = (x_sw + 1 - √3) / (x_sw + 1 + √3)
//
// The batch inversion computes all 2n inverses {1/d1[i], 1/d2[i]} with a
// single field inversion plus O(n) multiplications.
//
// Output is in compact (x_te, y_te) format — the GPU computes T = 2d·x·y
// on the fly in the mixed addition kernel.
func convertToEdMSM(points []bls12377.G1Affine) []G1TEPoint {
	n := len(points)
	result := make([]G1TEPoint, n)

	var one fp.Element
	one.SetOne()

	// Compute denominators to batch-invert:
	//   d[i]   = y_sw[i] · invSqrtMinusA    (denominator for x_te)
	//   d[n+i] = x_sw[i] + 1 + √3           (denominator for y_te)
	d := make([]fp.Element, 2*n)
	for i := range points {
		d[i].Mul(&points[i].Y, &teInvSqrtMinusA)
		d[n+i].Add(&points[i].X, &one)
		d[n+i].Add(&d[n+i], &teSqrtThree)
	}

	inv := fp.BatchInvert(d)

	for i := range points {
		// x_te = (x_sw + 1) · inv[i]
		//      = (x_sw + 1) / (y_sw · invSqrtMinusA)
		var xPlusOne fp.Element
		xPlusOne.Add(&points[i].X, &one)

		var xTe fp.Element
		xTe.Mul(&xPlusOne, &inv[i])

		// y_te = (x_sw + 1 - √3) · inv[n+i]
		//      = (x_sw + 1 - √3) / (x_sw + 1 + √3)
		var yTe fp.Element
		yTe.Sub(&xPlusOne, &teSqrtThree)
		yTe.Mul(&yTe, &inv[n+i])

		// Compact format: store (x_te, y_te) directly
		result[i] = packEdXY(xTe, yTe)
	}

	return result
}

// ---------------------------------------------------------------------------
// packEdXY: pack fp.Elements into G1TEPoint
// ---------------------------------------------------------------------------

// packEdXY packs two fp.Element values (x_te, y_te) into a G1TEPoint
// matching the GPU G1EdXY struct layout: 6 limbs of x followed by 6 limbs of y.
func packEdXY(x, y fp.Element) G1TEPoint {
	var out G1TEPoint
	out[0] = x[0]
	out[1] = x[1]
	out[2] = x[2]
	out[3] = x[3]
	out[4] = x[4]
	out[5] = x[5]
	out[6] = y[0]
	out[7] = y[1]
	out[8] = y[2]
	out[9] = y[3]
	out[10] = y[4]
	out[11] = y[5]
	return out
}

// precompFromCompact converts a compact (X, Y) TE point to the precomputed
// (Y-X, Y+X, 2d·X·Y) format used by the GPU MSM accumulator. Inputs and
// outputs are in Montgomery form. Cost per point: 1 fp.Sub + 1 fp.Add + 2 fp.Mul.
func precompFromCompact(c G1TEPoint) g1TEPrecompPoint {
	var x, y, ymx, ypx, twoDxy fp.Element
	x[0] = c[0]
	x[1] = c[1]
	x[2] = c[2]
	x[3] = c[3]
	x[4] = c[4]
	x[5] = c[5]
	y[0] = c[6]
	y[1] = c[7]
	y[2] = c[8]
	y[3] = c[9]
	y[4] = c[10]
	y[5] = c[11]

	ymx.Sub(&y, &x)
	ypx.Add(&y, &x)
	twoDxy.Mul(&x, &y)
	twoDxy.Mul(&twoDxy, &teDCoeffDouble)

	var out g1TEPrecompPoint
	out[0] = ymx[0]
	out[1] = ymx[1]
	out[2] = ymx[2]
	out[3] = ymx[3]
	out[4] = ymx[4]
	out[5] = ymx[5]
	out[6] = ypx[0]
	out[7] = ypx[1]
	out[8] = ypx[2]
	out[9] = ypx[3]
	out[10] = ypx[4]
	out[11] = ypx[5]
	out[12] = twoDxy[0]
	out[13] = twoDxy[1]
	out[14] = twoDxy[2]
	out[15] = twoDxy[3]
	out[16] = twoDxy[4]
	out[17] = twoDxy[5]
	return out
}

// precompBatchFromCompact converts a slice of compact TE points to the
// precomputed format. Allocates a fresh slice; for pinned destinations,
// use writePrecompFromCompact to write directly into a pre-allocated buffer.
func precompBatchFromCompact(src []G1TEPoint) []g1TEPrecompPoint {
	out := make([]g1TEPrecompPoint, len(src))
	for i := range src {
		out[i] = precompFromCompact(src[i])
	}
	return out
}

// writePrecompFromCompact converts compact TE points to the precomputed format
// and writes the results directly into the destination slice (typically
// pinned host memory). Avoids the temporary heap allocation of
// precompBatchFromCompact for large SRS uploads.
func writePrecompFromCompact(dst []g1TEPrecompPoint, src []G1TEPoint) {
	for i := range src {
		dst[i] = precompFromCompact(src[i])
	}
}

// ---------------------------------------------------------------------------
// teExtended: TE extended coordinates and point arithmetic
// ---------------------------------------------------------------------------

// teExtended represents a point on the BLS12-377 Twisted Edwards curve
// in extended coordinates (X, Y, T, Z), where:
//
//	x_affine = X / Z
//	y_affine = Y / Z
//	T        = X · Y / Z    (auxiliary coordinate for faster addition)
//
// The curve equation in extended coordinates is:
//
//	-X² + Y² = Z² + d · T²
//
// Extended coordinates avoid inversions during point addition and doubling,
// requiring only field multiplications.
type teExtended struct {
	x, y, t, z fp.Element
}

// teDouble doubles a TE extended point in place: p = 2·p.
//
// Uses the dedicated doubling formula "dbl-2008-hwcd" for twisted Edwards
// curves with a = -1. This is more efficient than the general unified addition
// formula: 4 squarings + 4 multiplications vs 9 multiplications.
//
// Reference: Hisil-Wong-Carter-Dawson 2008, Section 3.3.
// https://eprint.iacr.org/2008/522
//
// Formula (dbl-2008-hwcd, a = -1):
//
//	A  = X1²              (squaring)
//	B  = Y1²              (squaring)
//	C  = 2 · Z1²          (squaring + double)
//	D  = -A               (negate, since a = -1; for general a: D = a · A)
//	E  = (X1 + Y1)² - A - B   (squaring + 2 sub; this equals 2·X1·Y1)
//	G  = D + B            (= -X1² + Y1², the "curve" term)
//	F  = G - C            (= -X1² + Y1² - 2Z1²)
//	H  = D - B            (= -X1² - Y1²)
//
//	X3 = E · F
//	Y3 = G · H
//	T3 = E · H
//	Z3 = F · G
//
// Cost: 4S + 4M (4 squarings counted as muls, 4 multiplications for outputs).
func teDouble(p *teExtended) {
	var a0, b0, c0, d0, e0, f0, g0, h0 fp.Element

	a0.Mul(&p.x, &p.x) // A = X1²
	b0.Mul(&p.y, &p.y) // B = Y1²
	c0.Mul(&p.z, &p.z) // C = Z1²
	c0.Add(&c0, &c0)   // C = 2·Z1²
	d0.Neg(&a0)        // D = -A  (since a_coeff = -1)

	e0.Add(&p.x, &p.y) // E = X1 + Y1
	e0.Mul(&e0, &e0)   // E = (X1 + Y1)²
	e0.Sub(&e0, &a0)   // E = (X1 + Y1)² - A
	e0.Sub(&e0, &b0)   // E = (X1 + Y1)² - A - B  (= 2·X1·Y1)

	g0.Add(&d0, &b0) // G = D + B  (= -X1² + Y1²)
	f0.Sub(&g0, &c0) // F = G - C  (= -X1² + Y1² - 2Z1²)
	h0.Sub(&d0, &b0) // H = D - B  (= -X1² - Y1²)

	p.x.Mul(&e0, &f0) // X3 = E · F
	p.y.Mul(&g0, &h0) // Y3 = G · H
	p.t.Mul(&e0, &h0) // T3 = E · H
	p.z.Mul(&f0, &g0) // Z3 = F · G
}

// teAdd computes p += q in TE extended coordinates.
//
// Uses the unified addition formula "add-2008-hwcd" for twisted Edwards curves
// with a = -1. This formula works for any two points (including doubling, though
// the dedicated teDouble is more efficient for that case).
//
// Reference: Hisil-Wong-Carter-Dawson 2008, Section 3.1.
// https://eprint.iacr.org/2008/522
//
// Formula (add-2008-hwcd, a = -1, cost 9M):
//
//	A = (Y1 - X1) · (Y2 - X2)    (cross-difference product)
//	B = (Y1 + X1) · (Y2 + X2)    (cross-sum product)
//	C = T1 · T2 · 2d             (twist term; 2d is precomputed as teDCoeffDouble)
//	D = 2 · Z1 · Z2              (projective scaling)
//
//	E = B - A     (encodes 2·(X1·Y2 + X2·Y1) via expansion)
//	F = D - C     (denominator-like term for x)
//	G = D + C     (denominator-like term for y)
//	H = B + A     (encodes 2·(X1·X2 + Y1·Y2) via expansion, sign depends on a=-1)
//
//	X3 = E · F
//	Y3 = G · H
//	T3 = E · H
//	Z3 = F · G
//
// Note: The GPU kernel uses a "mixed addition" variant (9M) where q is affine
// (Z2 = 1, T2 = 2d·x·y precomputed or computed on-the-fly), saving one
// multiplication in the D term.
func teAdd(p *teExtended, q *teExtended) {
	var a0, b0, c0, d0, e0, f0, g0, h0, t1, t2 fp.Element

	// A = (Y1 - X1) · (Y2 - X2)
	t1.Sub(&p.y, &p.x)
	t2.Sub(&q.y, &q.x)
	a0.Mul(&t1, &t2)

	// B = (Y1 + X1) · (Y2 + X2)
	t1.Add(&p.y, &p.x)
	t2.Add(&q.y, &q.x)
	b0.Mul(&t1, &t2)

	// C = T1 · T2 · 2d
	c0.Mul(&p.t, &q.t)
	c0.Mul(&c0, &teDCoeffDouble)

	// D = 2 · Z1 · Z2
	d0.Mul(&p.z, &q.z)
	d0.Add(&d0, &d0)

	// E = B - A,  H = B + A
	e0.Sub(&b0, &a0)
	h0.Add(&b0, &a0)

	// F = D - C,  G = D + C
	f0.Sub(&d0, &c0)
	g0.Add(&d0, &c0)

	// X3 = E · F,  Y3 = G · H,  T3 = E · H,  Z3 = F · G
	p.x.Mul(&e0, &f0)
	p.y.Mul(&g0, &h0)
	p.t.Mul(&e0, &h0)
	p.z.Mul(&f0, &g0)
}

// unpackTEExtended unpacks 24 uint64s (as returned by the GPU MSM kernel for
// each window result) into a teExtended struct. The layout is:
//
//	w[0..5]   = X  (6 limbs of Fp in Montgomery form)
//	w[6..11]  = Y
//	w[12..17] = T
//	w[18..23] = Z
func unpackTEExtended(w [24]uint64) teExtended {
	var p teExtended
	p.x = fp.Element{w[0], w[1], w[2], w[3], w[4], w[5]}
	p.y = fp.Element{w[6], w[7], w[8], w[9], w[10], w[11]}
	p.t = fp.Element{w[12], w[13], w[14], w[15], w[16], w[17]}
	p.z = fp.Element{w[18], w[19], w[20], w[21], w[22], w[23]}
	return p
}

// teExtended2jac converts a teExtended point to gnark-crypto G1Jac by
// packing into 24 uint64s and delegating to te2jac.
func teExtended2jac(p teExtended) bls12377.G1Jac {
	var w [24]uint64
	w[0] = p.x[0]
	w[1] = p.x[1]
	w[2] = p.x[2]
	w[3] = p.x[3]
	w[4] = p.x[4]
	w[5] = p.x[5]
	w[6] = p.y[0]
	w[7] = p.y[1]
	w[8] = p.y[2]
	w[9] = p.y[3]
	w[10] = p.y[4]
	w[11] = p.y[5]
	w[12] = p.t[0]
	w[13] = p.t[1]
	w[14] = p.t[2]
	w[15] = p.t[3]
	w[16] = p.t[4]
	w[17] = p.t[5]
	w[18] = p.z[0]
	w[19] = p.z[1]
	w[20] = p.z[2]
	w[21] = p.z[3]
	w[22] = p.z[4]
	w[23] = p.z[5]
	return te2jac(w)
}

// te2jac converts a GPU TE extended point (24 uint64s: X[6], Y[6], T[6], Z[6])
// back to gnark-crypto G1Jac in Short Weierstrass coordinates.
// Follows gbotrel/zprize-mobile-harness G1Affine.FromExtendedEd exactly.
//
// TE → SW mapping (from extended projective coordinates X, Y, T, Z):
//
// Step 1: Compute affine TE coordinates from projective.
//
//	x_te_aff = X / Z
//	y_te_aff = Y / Z
//
// Step 2: Map TE affine to SW affine. The isogeny inverts the SW→TE map:
//
//	Given SW→TE:
//	  x_te = (x_sw + 1) / (y_sw · invSqrtMinusA)
//	  y_te = (x_sw + 1 - √3) / (x_sw + 1 + √3)
//
//	Solving for x_sw, y_sw:
//	  Let n = (1 + y_te) / (1 - y_te) · √3
//	        = (Z + Y) / (Z - Y) · √3       (using projective form)
//	  x_sw = n - 1
//	  y_sw = n · Z / (X · invSqrtMinusA)    (= n / x_te_aff / invSqrtMinusA)
//
// Step 3: Convert SW affine to Jacobian via gnark-crypto FromAffine.
//
// Implementation uses batch inversion over {Z, Z-Y, X·invSqrtMinusA} to
// compute all three needed inverses with a single field inversion.
//
// Identity handling: TE identity is (0, 1, 0, 1) in extended coordinates.
// Detected by X == 0 && Y == Z, mapped to the Jacobian point at infinity.
func te2jac(w [24]uint64) bls12377.G1Jac {
	var xTE, yTE, tTE, zTE fp.Element
	xTE[0] = w[0]
	xTE[1] = w[1]
	xTE[2] = w[2]
	xTE[3] = w[3]
	xTE[4] = w[4]
	xTE[5] = w[5]
	yTE[0] = w[6]
	yTE[1] = w[7]
	yTE[2] = w[8]
	yTE[3] = w[9]
	yTE[4] = w[10]
	yTE[5] = w[11]
	tTE[0] = w[12]
	tTE[1] = w[13]
	tTE[2] = w[14]
	tTE[3] = w[15]
	tTE[4] = w[16]
	tTE[5] = w[17]
	zTE[0] = w[18]
	zTE[1] = w[19]
	zTE[2] = w[20]
	zTE[3] = w[21]
	zTE[4] = w[22]
	zTE[5] = w[23]
	_ = tTE // T not needed for TE→SW conversion

	// Check for identity: X == 0 && Y == Z (TE identity is (0, 1, 0, 1))
	if xTE.IsZero() && yTE.Equal(&zTE) {
		var jac bls12377.G1Jac
		jac.X.SetOne()
		jac.Y.SetOne()
		jac.Z.SetZero()
		return jac
	}

	// Batch-invert {Z, Z-Y, X·invSqrtMinusA} for the three divisions needed:
	//   inv[0] = 1/Z             → used for y_te_aff = Y/Z
	//   inv[1] = 1/(Z-Y)        → used for n = (Z+Y)/(Z-Y) · √3
	//   inv[2] = 1/(X·invSqrtMinusA) → used for y_sw = n·Z/(X·invSqrtMinusA)
	var d1, d2, d3 fp.Element
	d1.Set(&zTE)
	d2.Sub(&zTE, &yTE)
	d3.Mul(&xTE, &teInvSqrtMinusA)

	inv := fp.BatchInvert([]fp.Element{d1, d2, d3})

	// Multiply inv[1] and inv[2] by Z to convert from 1/(Z-Y) → Z/(Z-Y)
	// and 1/(X·invSqrtMinusA) → Z/(X·invSqrtMinusA), as in the reference.
	inv[1].Mul(&inv[1], &zTE) // = Z/(Z-Y)
	inv[2].Mul(&inv[2], &zTE) // = Z/(X·invSqrtMinusA)

	// y_te_aff = Y/Z = Y · inv[0]
	var yAff fp.Element
	yAff.Mul(&yTE, &inv[0])

	// Check for identity in affine: x_te = 0, y_te = 1
	var one fp.Element
	one.SetOne()
	var xAff fp.Element
	xAff.Sub(&zTE, &yTE)
	xAff.Mul(&xAff, &inv[0])
	if xAff.IsZero() && yAff.Equal(&one) {
		var jac bls12377.G1Jac
		jac.X.SetOne()
		jac.Y.SetOne()
		jac.Z.SetZero()
		return jac
	}

	// n = (1 + y_te_aff) · Z/(Z-Y) · √3
	//   = (1 + Y/Z) · Z/(Z-Y) · √3
	//   = (Z + Y)/Z · Z/(Z-Y) · √3
	//   = (Z + Y)/(Z - Y) · √3
	var n fp.Element
	n.Add(&one, &yAff)
	n.Mul(&n, &inv[1])
	n.Mul(&n, &teSqrtThree)

	// x_sw = n - 1
	var xSW fp.Element
	xSW.Sub(&n, &one)

	// y_sw = n · Z/(X · invSqrtMinusA) = n · inv[2]
	var ySW fp.Element
	ySW.Mul(&n, &inv[2])

	// Convert SW affine → G1Jac
	var aff bls12377.G1Affine
	aff.X = xSW
	aff.Y = ySW

	var jac bls12377.G1Jac
	jac.FromAffine(&aff)
	return jac
}

// ---------------------------------------------------------------------------
// Serialization: WriteG1TEPoints, ReadG1TEPoints, raw variants
// ---------------------------------------------------------------------------

// WriteG1TEPoints writes TE points in a safe format with magic header and count.
// Format: [8-byte magic "GNRKTE02"][8-byte uint64 LE count][count × 96 bytes data]
func WriteG1TEPoints(w io.Writer, points []G1TEPoint) error {
	if _, err := w.Write(g1TEMagic[:]); err != nil {
		return err
	}
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(len(points)))
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	return WriteG1TEPointsRaw(w, points)
}

// ReadG1TEPoints reads TE points from the safe format, validating the magic header.
func ReadG1TEPoints(r io.Reader) ([]G1TEPoint, error) {
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, fmt.Errorf("gpu: reading G1TEPoint header: %w", err)
	}
	if magic != g1TEMagic {
		return nil, fmt.Errorf("gpu: invalid G1TEPoint file magic")
	}
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("gpu: reading G1TEPoint count: %w", err)
	}
	n := int(binary.LittleEndian.Uint64(buf[:]))
	return ReadG1TEPointsRaw(r, n)
}

// WriteG1TEPointsRaw writes TE points as a raw memory dump with no header.
// The caller is responsible for tracking the count. This is the fastest path
// for serialization and produces files suitable for ReadG1TEPointsPinned.
func WriteG1TEPointsRaw(w io.Writer, points []G1TEPoint) error {
	if len(points) == 0 {
		return nil
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&points[0])), len(points)*g1TEPointSize)
	_, err := w.Write(buf)
	return err
}

// ReadG1TEPointsRaw reads n TE points from a raw memory dump (no header).
// Returns a Go-allocated slice. For zero-copy pinned memory, use ReadG1TEPointsPinned.
func ReadG1TEPointsRaw(r io.Reader, n int) ([]G1TEPoint, error) {
	if n <= 0 {
		return nil, nil
	}
	points := make([]G1TEPoint, n)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&points[0])), n*g1TEPointSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return points, nil
}
