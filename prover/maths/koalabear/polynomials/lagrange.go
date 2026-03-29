package polynomials

import (
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
)

// EvalLagrange evaluates a polynomial given in Lagrange (evaluation) basis at
// point z using the barycentric formula:
//
//	P(z) = (zⁿ - 1) · Σᵢ [ (ωⁱ/n · p[i]) / (z - ωⁱ) ]
//
// p[i] = P(ωⁱ) for i = 0..cardinality-1.
// gen is the generator ω of the multiplicative subgroup (pass d.Generator).
// cardinality must be a power of 2 and z must not be a root of unity.
func EvalLagrange(p field.FieldVec, z field.FieldElem, gen field.Element, cardinality uint64) field.FieldElem {
	if cardinality == 0 {
		return field.ElemZero()
	}
	n := int(cardinality)

	// 1/n in the base field
	var invN field.Element
	invN.SetUint64(cardinality)
	invN.Inverse(&invN)

	// z^n - 1  (n is a power of 2, so compute by repeated squaring)
	tb := bits.TrailingZeros(uint(cardinality))
	zPowN := z
	for i := 0; i < tb; i++ {
		zPowN = zPowN.Square()
	}
	zPowNMinusOne := zPowN.Sub(field.ElemOne())

	// Walk i = 0..n-1, accumulating:
	//   weightedP[i] = (ω^i / n) * p[i]
	//   denomBase[i] = z - ω^i   (as Element, valid when z.IsBase())
	//   denomExt[i]  = z - ω^i   (as Ext,     otherwise)
	var accOmega field.Element
	accOmega.SetOne()

	weightedP := make([]field.FieldElem, n)

	var (
		denomBase []field.Element
		denomExt  []field.Ext
	)
	if z.IsBase() {
		denomBase = make([]field.Element, n)
	} else {
		denomExt = make([]field.Ext, n)
	}

	var pBase []field.Element
	var pExt []field.Ext
	if p.IsBase() {
		pBase = p.AsBase()
	} else {
		pExt = p.AsExt()
	}

	for i := 0; i < n; i++ {
		// wᵢ = ωⁱ / n
		var wi field.Element
		wi.Mul(&accOmega, &invN)

		// weightedP[i] = wᵢ * p[i]
		var pi field.FieldElem
		if p.IsBase() {
			pi = field.ElemFromBase(pBase[i])
		} else {
			pi = field.ElemFromExt(pExt[i])
		}
		weightedP[i] = pi.Mul(field.ElemFromBase(wi))

		// z - ωⁱ
		if z.IsBase() {
			zb := z.AsBase()
			denomBase[i].Sub(&zb, &accOmega)
		} else {
			var zMinusOmega field.Ext
			zMinusOmega = z.AsExt()
			zMinusOmega.B0.A0.Sub(&zMinusOmega.B0.A0, &accOmega)
			denomExt[i] = zMinusOmega
		}

		accOmega.Mul(&accOmega, &gen)
	}

	// Batch invert the denominators
	var invDenom []field.FieldElem
	if z.IsBase() {
		invBase := make([]field.Element, n)
		field.VecBatchInvBase(invBase, denomBase)
		invDenom = make([]field.FieldElem, n)
		for i, e := range invBase {
			invDenom[i] = field.ElemFromBase(e)
		}
	} else {
		invExt := make([]field.Ext, n)
		field.VecBatchInvExt(invExt, denomExt)
		invDenom = make([]field.FieldElem, n)
		for i, e := range invExt {
			invDenom[i] = field.ElemFromExt(e)
		}
	}

	// sum = Σᵢ weightedP[i] * invDenom[i]
	sum := field.ElemZero()
	for i := 0; i < n; i++ {
		sum = sum.Add(weightedP[i].Mul(invDenom[i]))
	}

	return zPowNMinusOne.Mul(sum)
}

// EvalLagrangeBatch evaluates the same Lagrange polynomial p at multiple points
// zs, sharing the precomputed barycentric weight vector wᵢ·p[i] and ω^i table.
//
// For k evaluation points, this saves approximately (k-1)·n multiplications
// compared to k individual calls to [EvalLagrange].
func EvalLagrangeBatch(p field.FieldVec, zs []field.FieldElem, gen field.Element, cardinality uint64) []field.FieldElem {
	if len(zs) == 0 {
		return nil
	}
	if len(zs) == 1 {
		return []field.FieldElem{EvalLagrange(p, zs[0], gen, cardinality)}
	}

	n := int(cardinality)

	// 1/n in the base field
	var invN field.Element
	invN.SetUint64(cardinality)
	invN.Inverse(&invN)

	// Precompute omega powers and weighted coefficients once
	omegaPowers := make([]field.Element, n)
	omegaPowers[0].SetOne()
	for i := 1; i < n; i++ {
		omegaPowers[i].Mul(&omegaPowers[i-1], &gen)
	}

	weightedP := make([]field.FieldElem, n)
	{
		var pBase []field.Element
		var pExt []field.Ext
		if p.IsBase() {
			pBase = p.AsBase()
		} else {
			pExt = p.AsExt()
		}
		for i := 0; i < n; i++ {
			var wi field.Element
			wi.Mul(&omegaPowers[i], &invN)
			var pi field.FieldElem
			if p.IsBase() {
				pi = field.ElemFromBase(pBase[i])
			} else {
				pi = field.ElemFromExt(pExt[i])
			}
			weightedP[i] = pi.Mul(field.ElemFromBase(wi))
		}
	}

	tb := bits.TrailingZeros(uint(cardinality))
	results := make([]field.FieldElem, len(zs))

	for j, z := range zs {
		// zⁿ - 1
		zPowN := z
		for i := 0; i < tb; i++ {
			zPowN = zPowN.Square()
		}
		zPowNMinusOne := zPowN.Sub(field.ElemOne())

		// Build denom slice and batch invert
		var invDenom []field.FieldElem
		if z.IsBase() {
			denomBase := make([]field.Element, n)
			zb := z.AsBase()
			for i := 0; i < n; i++ {
				denomBase[i].Sub(&zb, &omegaPowers[i])
			}
			invBase := make([]field.Element, n)
			field.VecBatchInvBase(invBase, denomBase)
			invDenom = make([]field.FieldElem, n)
			for i, e := range invBase {
				invDenom[i] = field.ElemFromBase(e)
			}
		} else {
			denomExt := make([]field.Ext, n)
			for i := 0; i < n; i++ {
				zx := z.AsExt()
				zx.B0.A0.Sub(&zx.B0.A0, &omegaPowers[i])
				denomExt[i] = zx
			}
			invExt := make([]field.Ext, n)
			field.VecBatchInvExt(invExt, denomExt)
			invDenom = make([]field.FieldElem, n)
			for i, e := range invExt {
				invDenom[i] = field.ElemFromExt(e)
			}
		}

		sum := field.ElemZero()
		for i := 0; i < n; i++ {
			sum = sum.Add(weightedP[i].Mul(invDenom[i]))
		}
		results[j] = zPowNMinusOne.Mul(sum)
	}

	return results
}

// ComputeLagrangeAtZ returns the vector [L₀(z), L₁(z), …, L_{n-1}(z)] where
//
//	Lᵢ(z) = ωⁱ/n · (zⁿ - 1) / (z - ωⁱ)
//
// Uses the recurrence relation to avoid n-1 extra inversions:
//
//	L₀(z) = (1/n) · (zⁿ - 1) / (z - 1)
//	Lᵢ(z) = Lᵢ₋₁(z) · ω · (z - ωⁱ⁻¹) / (z - ωⁱ)
//
// gen must be the generator ω of the multiplicative subgroup (pass d.Generator),
// cardinality must be a power of 2, and z must not be a root of unity.
func ComputeLagrangeAtZ(z field.FieldElem, gen field.Element, cardinality uint64) []field.FieldElem {
	n := int(cardinality)
	res := make([]field.FieldElem, n)

	tb := bits.TrailingZeros(uint(cardinality))

	// zⁿ - 1
	zPowN := z
	for i := 0; i < tb; i++ {
		zPowN = zPowN.Square()
	}
	zPowNMinusOne := zPowN.Sub(field.ElemOne())

	// L₀ = (zⁿ - 1) / (z - 1) / n
	zMinusOne := z.Sub(field.ElemOne())
	res[0] = zPowNMinusOne.Div(zMinusOne)
	var invN field.Element
	invN.SetUint64(cardinality)
	invN.Inverse(&invN)
	res[0] = res[0].Mul(field.ElemFromBase(invN))

	// accZMinusOmega tracks (z - ω^{i-1}), initialized to (z - 1)
	accZMinusOmega := zMinusOne
	genElem := field.ElemFromBase(gen)

	var accOmega field.Element
	accOmega.SetOne()

	for i := 1; i < n; i++ {
		res[i] = res[i-1].Mul(genElem)                       // Lᵢ = ω · Lᵢ₋₁
		res[i] = res[i].Mul(accZMinusOmega)                  // Lᵢ *= (z - ω^{i-1})
		accOmega.Mul(&accOmega, &gen)                        // accOmega = ω^i
		accZMinusOmega = z.Sub(field.ElemFromBase(accOmega)) // (z - ω^i)
		res[i] = res[i].Div(accZMinusOmega)                  // Lᵢ /= (z - ω^i)
	}

	return res
}
