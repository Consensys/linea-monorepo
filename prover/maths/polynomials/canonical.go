package polynomials

import (
	"github.com/consensys/gnark/frontend"
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
func GnarkEvalCanonical(api frontend.API, p []koalagnark.Element, z koalagnark.Ext) koalagnark.Ext {
	f := koalagnark.NewAPI(api)

	res := f.ZeroExt()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = f.MulExt(res, z)
		res = f.AddByBaseExt(res, p[s-1-i])
	}
	return res
}

// GnarkEvalCanonicalExt evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func GnarkEvalCanonicalExt(api frontend.API, p []koalagnark.Ext, z koalagnark.Ext) koalagnark.Ext {
	f := koalagnark.NewAPI(api)

	res := f.ZeroExt()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = f.MulExt(res, z)
		res = f.AddExt(res, p[s-1-i])
	}
	return res
}

// GnarkEvalCanonicalExtBatch evaluates polynomial p at multiple points zs.
// This is more efficient than calling GnarkEvalCanonicalExt multiple times
// when evaluating the same polynomial at different points.
// Uses Horner's method for each point in parallel within the same loop,
// loading each coefficient only once.
func GnarkEvalCanonicalExtBatch(api frontend.API, p []koalagnark.Ext, zs []koalagnark.Element) []koalagnark.Ext {
	if len(zs) == 0 {
		return nil
	}
	if len(zs) == 1 {
		panic("poly should be multiple points")
	}

	f := koalagnark.NewAPI(api)

	// Initialize results for each evaluation point
	results := make([]koalagnark.Ext, len(zs))
	for j := range results {
		results[j] = f.ZeroExt()
	}

	s := len(p)
	for i := 0; i < len(p); i++ {
		coeff := p[s-1-i]
		for j := range zs {
			results[j] = f.MulByFpExt(results[j], zs[j])
			results[j] = f.AddExt(results[j], coeff)
		}
	}

	return results
}
