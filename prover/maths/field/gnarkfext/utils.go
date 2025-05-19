package gnarkfext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func FromValue(v fext.Element) E4 {
	return E4{
		B0: E2{A0: v.B0.A0, A1: v.B0.A1},
		B1: E2{A0: v.B1.A0, A1: v.B1.A1},
	}
}

// One returns 1
func One() E4 {
	return E4{
		B0: E2{
			A0: 1,
			A1: 0,
		},
		B1: E2{
			A0: 0,
			A1: 0,
		},
	}
}

// Exp exponentiation in gnark circuit, using the fast exponentiation
func Exp(api frontend.API, x E4, n int) E4 {

	if n < 0 {
		x.Inverse(api, x)
		return Exp(api, x, -n)
	}

	if n == 0 {
		return One()
	}

	if n == 1 {
		return x
	}

	var x2 E4
	x2.Mul(api, x, x)
	if n%2 == 0 {
		return Exp(api, x2, n/2)
	} else {
		res := Exp(api, x2, (n-1)/2)
		return *res.Mul(api, res, x)
	}

}
