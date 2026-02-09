package polynomials

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

func eval(poly []fext.Element, x fext.Element) fext.Element {
	var res fext.Element
	s := len(poly)
	for i := 0; i < len(poly); i++ {
		res.Mul(&res, &x)
		res.Add(&res, &poly[s-1-i])
	}
	return res
}

// GnarkEvalCanonical evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func GnarkEvalCanonical(koalaAPI *koalagnark.API, p []koalagnark.Element, z koalagnark.Ext) koalagnark.Ext {

	res := koalaAPI.ZeroExt()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = koalaAPI.MulExt(res, z)
		res = koalaAPI.AddByBaseExt(res, p[s-1-i])
	}
	return res
}

// GnarkEvalCanonicalExt evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func GnarkEvalCanonicalExt(koalaAPI *koalagnark.API, p []koalagnark.Ext, z koalagnark.Ext) koalagnark.Ext {

	res := koalaAPI.ZeroExt()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = koalaAPI.MulExt(res, z)
		res = koalaAPI.AddExt(res, p[s-1-i])
	}
	return res
}

// GnarkEvalCanonicalBatch evaluates multiple different polynomials (polys) at the same point z.
// This is more efficient than calling GnarkEvalCanonical multiple times when evaluating
// different polynomials at the same point.
//
// Uses a shared computation of powers of z to avoid redundant multiplications.
// For k polynomials each of degree d, this saves approximately (k-1)*d MulExt operations
// by computing z, z², z³, ... only once and reusing them.
//
// Algorithm:
//  1. Find the maximum degree among all polynomials
//  2. Precompute powers: z⁰, z¹, z², ..., z^maxDegree (computed once)
//  3. For each polynomial, use the precomputed powers to evaluate via direct summation
//
// Cost comparison (for k polynomials of degree d):
//   - Individual calls: k * d MulExt operations (Horner's method)
//   - Batch: d MulExt (powers) ≈ MUCH LESS MulExt count
func GnarkEvalCanonicalBatch(koalaAPI *koalagnark.API, polys [][]koalagnark.Element, z koalagnark.Ext) []koalagnark.Ext {
	if len(polys) == 0 {
		return nil
	}
	if len(polys) == 1 {
		return []koalagnark.Ext{GnarkEvalCanonical(koalaAPI, polys[0], z)}
	}

	// Find maximum polynomial degree to determine how many powers we need
	maxDegree := 0
	for _, poly := range polys {
		if len(poly) > maxDegree {
			maxDegree = len(poly)
		}
	}

	if maxDegree == 0 {
		results := make([]koalagnark.Ext, len(polys))
		zero := koalaAPI.ZeroExt()
		for i := range results {
			results[i] = zero
		}
		return results
	}

	// Precompute powers of z: [z⁰, z¹, z², ..., z^(maxDegree-1)]
	// This is computed once and shared across all polynomial evaluations
	powers := make([]koalagnark.Ext, maxDegree)
	powers[0] = koalaAPI.OneExt()
	for i := 1; i < maxDegree; i++ {
		powers[i] = koalaAPI.MulExt(powers[i-1], z)
		if i%5 == 0 {
			// Reduce every 4 multiplications to keep size in check
			powers[i] = koalaAPI.ModReduceExt(powers[i])
		}
	}

	// Evaluate each polynomial using the precomputed powers
	// For polynomial p(x) = ∑ᵢ p[i]·xⁱ, compute ∑ᵢ p[i]·powers[i]
	results := make([]koalagnark.Ext, len(polys))

	for j, poly := range polys {
		if len(poly) == 0 {
			results[j] = koalaAPI.ZeroExt()
			continue
		}
		if len(poly) == 1 {
			results[j] = koalaAPI.MulByFpExt(powers[0], poly[0])
			continue
		}

		terms := make([]koalagnark.Ext, len(poly))
		for i := range poly {
			terms[i] = koalaAPI.MulByFpExt(powers[i], poly[i])
		}
		result := koalaAPI.SumExt(terms...)
		results[j] = result
	}

	return results
}
