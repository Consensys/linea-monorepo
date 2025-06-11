package fastpoly

import (
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// EvaluateLagrangeGnark a polynomial in lagrange basis on a gnark circuit
func EvaluateLagrangeGnark(api frontend.API, poly []frontend.Variable, x frontend.Variable) frontend.Variable {

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
	dens := make([]frontend.Variable, size) // [x-1, x-ω, x-ω², ...]
	for i := 0; i < size; i++ {
		dens[i] = api.Sub(x, accw)
		accw.Mul(&accw, &omega)
	}
	invdens := make([]frontend.Variable, size) // [1/x-1, 1/x-ω, 1/x-ω², ...]
	for i := 0; i < size; i++ {
		invdens[i] = api.Inverse(dens[i])
	}

	var tmp frontend.Variable
	tmp = gnarkutil.Exp(api, x, size)
	tmp = api.Sub(tmp, one) // xⁿ-1
	li := api.Inverse(size)
	li = api.Mul(tmp, li) // 1/n * (xⁿ-1)

	var res frontend.Variable
	res = 0
	for i := 0; i < size; i++ {
		li = api.Mul(li, invdens[i])
		tmp = api.Mul(li, poly[i]) // pᵢ *  ωⁱ/n * ( xⁿ-1)/(x-ωⁱ)
		res = api.Add(res, tmp)
		li = api.Mul(li, dens[i])
		li = api.Mul(li, omega)
	}

	return res

}
