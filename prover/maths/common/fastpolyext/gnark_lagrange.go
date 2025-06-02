package fastpolyext

import (
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext" // Assuming gnarkfext has the Exp function
	"github.com/consensys/linea-monorepo/prover/utils"
)

// EvaluateLagrangeGnark evaluates a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark(api frontend.API, poly []gnarkfext.Element, x gnarkfext.Element) gnarkfext.Element {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	size := len(poly)
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		// TODO handle that properly
		panic(err)
	}

	var accw, one field.Element
	one.SetOne()
	accw.SetOne()
	dens := make([]gnarkfext.Element, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i] = x
		dens[i].B0.A0 = api.Sub(x.B0.A0, accw) // accw lives in the base field
		accw.Mul(&accw, &omega)
	}
	invdens := make([]gnarkfext.Element, size) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	for i := 0; i < size; i++ {
		invdens[i].Inverse(api, dens[i])
	}

	var tmp gnarkfext.Element
	tmp = gnarkfext.Exp(api, x, size)
	tmp.B0.A0 = api.Sub(tmp.B0.A0, one) // xⁿ-1
	li := api.Inverse(size)
	liExt := gnarkfext.FromBaseField(li)
	liExt.Mul(api, tmp, liExt) // 1/n * (xⁿ-1)

	res := gnarkfext.FromBaseField(0)
	for i := 0; i < size; i++ {
		liExt.Mul(api, liExt, invdens[i])
		tmp.Mul(api, liExt, poly[i]) // pᵢ *  ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		res.Add(api, res, tmp)
		liExt.Mul(api, liExt, dens[i])
		liExt.MulByFp(api, liExt, omega)
	}

	return res

}
