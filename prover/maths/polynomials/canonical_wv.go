package polynomials

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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
func GnarkEvalCanonical(api frontend.API, p []zk.WrappedVariable, z gnarkfext.E4Gen) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := ext4.Zero()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = ext4.Mul(res, &z)
		res = ext4.AddByBase(res, p[s-1-i])
	}
	return *res
}

// GnarkEvalCanonicalExt evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func GnarkEvalCanonicalExt(api frontend.API, p []gnarkfext.E4Gen, z gnarkfext.E4Gen) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := ext4.Zero()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = ext4.Mul(res, &z)
		res = ext4.Add(res, &p[s-1-i])
	}
	return *res
}
