package polynomials

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// gnarkEvalCanonical evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func gnarkEvalCanonical(api frontend.API, p []zk.WrappedVariable, z gnarkfext.E4Gen) gnarkfext.E4Gen {

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
