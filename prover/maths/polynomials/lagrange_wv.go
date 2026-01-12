package polynomials

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

func GnarkEvaluateLagrangeExt(api frontend.API, p []gnarkfext.Element, z gnarkfext.Element, gen field.Element, cardinality uint64) gnarkfext.Element {

	res := gnarkfext.Zero()
	lagranges := gnarkComputeLagrangeAtZ(api, z, gen, cardinality)

	var tmp gnarkfext.Element
	for i := uint64(0); i < cardinality; i++ {
		tmp.Mul(api, lagranges[i], p[i])
		res.Add(api, res, tmp)
	}

	return res
}

// computeLagrange returns Lᵢ(ζ) for i=1..n
// with lᵢ(ζ) = ωⁱ/n*(ζⁿ-1)/(ζ - ωⁱ)
func gnarkComputeLagrangeAtZ(api frontend.API, z gnarkfext.Element, gen field.Element, cardinality uint64) []gnarkfext.Element {

	res := make([]gnarkfext.Element, cardinality)
	tb := bits.TrailingZeros(uint(cardinality))

	// ζⁿ-1
	res[0] = z
	for i := 0; i < tb; i++ {
		res[0].Mul(api, res[0], res[0])
	}
	wOne := gnarkfext.One()
	res[0].Sub(api, res[0], wOne)

	// ζ-1
	var accZetaMinusOmegai gnarkfext.Element
	accZetaMinusOmegai.Sub(api, z, gnarkfext.One())

	// (ζⁿ-1)/(ζ-1)
	res[0].Div(api, res[0], accZetaMinusOmegai)

	// 1/n*(ζⁿ-1)/(ζ-1)
	res[0].DivByBase(api, res[0], cardinality)

	// res[i] <- res[i-1] * (ζ-ωⁱ⁻¹)/(ζ-ωⁱ) * ω
	var accOmega field.Element
	accOmega.SetOne()
	wGen := field.NewFromKoala(gen)
	for i := uint64(1); i < cardinality; i++ {
		res[i].MulByFp(api, res[i-1], wGen)         // res[i] <- ω * res[i-1]
		res[i].Mul(api, res[i], accZetaMinusOmegai) // res[i] <- res[i]*(ζ-ωⁱ⁻¹)
		accOmega.Mul(&accOmega, &gen)               // accOmega <- accOmega * ω
		wAccOmega := field.NewFromKoala(accOmega)
		wAccOmegaExt := gnarkfext.NewFromBase(wAccOmega)
		accZetaMinusOmegai.Sub(api, z, wAccOmegaExt) // accZetaMinusOmegai <- ζ-ωⁱ
		res[i].Div(api, res[i], accZetaMinusOmegai)  // res[i]  <- res[i]/(ζ-ωⁱ)
	}

	return res
}
