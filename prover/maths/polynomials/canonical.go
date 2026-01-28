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
