package polynomials

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

func GnarkEvaluateLagrangeExt(api frontend.API, p []gnarkfext.E4Gen, z gnarkfext.E4Gen, gen field.Element, cardinality uint64) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}
	res := *ext4.Zero()
	lagranges := gnarkComputeLagrangeAtZ(api, z, gen, cardinality)

	for i := uint64(0); i < cardinality; i++ {
		tmp := ext4.Mul(&lagranges[i], &p[i])
		res = *ext4.Add(&res, tmp)
	}

	return res
}

// computeLagrange returns Lᵢ(ζ) for i=1..n
// with lᵢ(ζ) = ωⁱ/n*(ζⁿ-1)/(ζ - ωⁱ)
func gnarkComputeLagrangeAtZ(api frontend.API, z gnarkfext.E4Gen, gen field.Element, cardinality uint64) []gnarkfext.E4Gen {

	res := make([]gnarkfext.E4Gen, cardinality)
	tb := bits.TrailingZeros(uint(cardinality))

	ext4, err := gnarkfext.NewExt4(api)

	if err != nil {
		panic(err)
	}

	// ζⁿ-1
	res[0] = z
	for i := 0; i < tb; i++ {
		res[0] = *ext4.Square(&res[0])
	}
	wOne := ext4.One()
	res[0] = *ext4.Sub(&res[0], wOne)

	// ζ-1
	accZetaMinusOmegai := *ext4.Sub(&z, wOne)

	// (ζⁿ-1)/(ζ-1)
	res[0] = *ext4.Div(&res[0], &accZetaMinusOmegai)

	// 1/n*(ζⁿ-1)/(ζ-1)
	wCardinality := zk.ValueOf(cardinality)
	res[0] = *ext4.DivByBase(&res[0], wCardinality)

	// res[i] <- res[i-1] * (ζ-ωⁱ⁻¹)/(ζ-ωⁱ) * ω
	var accOmega field.Element
	accOmega.SetOne()
	wGen := zk.ValueFromKoala(gen)
	var wAccOmega zk.WrappedVariable
	for i := uint64(1); i < cardinality; i++ {
		res[i] = *ext4.MulByFp(&res[i-1], wGen)          // res[i] <- ω * res[i-1]
		res[i] = *ext4.Mul(&res[i], &accZetaMinusOmegai) // res[i] <- res[i]*(ζ-ωⁱ⁻¹)
		accOmega.Mul(&accOmega, &gen)                    // accOmega <- accOmega * ω
		wAccOmega = zk.ValueFromKoala(accOmega)
		wAccOmegaExt := gnarkfext.FromBase(wAccOmega)
		accZetaMinusOmegai = *ext4.Sub(&z, &wAccOmegaExt) // accZetaMinusOmegai <- ζ-ωⁱ
		res[i] = *ext4.Div(&res[i], &accZetaMinusOmegai)  // res[i]  <- res[i]/(ζ-ωⁱ)
	}

	return res
}
