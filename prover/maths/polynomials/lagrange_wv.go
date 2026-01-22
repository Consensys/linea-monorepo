package polynomials

import (
	"math/big"
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

// GnarkEvaluateLagrangeExtBatch evaluates the same polynomial p at multiple points zs
// more efficiently by sharing the barycentric weight computation.
//
// Uses the barycentric formula: P(z) = (zⁿ - 1) * Σᵢ [ (ωⁱ/n * pᵢ) / (z - ωⁱ) ]
//
// For k evaluation points, this saves approximately (k-1)*n multiplications
// compared to k separate calls to GnarkEvaluateLagrangeExt.
func GnarkEvaluateLagrangeExtBatch(api frontend.API, p []gnarkfext.E4Gen, zs []gnarkfext.E4Gen, gen field.Element, cardinality uint64) []gnarkfext.E4Gen {
	if len(zs) == 0 {
		return nil
	}
	if len(zs) == 1 {
		return []gnarkfext.E4Gen{GnarkEvaluateLagrangeExt(api, p, zs[0], gen, cardinality)}
	}

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	// Precompute barycentric weights: wᵢ = ωⁱ/n
	// and weighted coefficients: wᵢ * pᵢ
	// This is shared across all evaluation points
	weightedP := make([]gnarkfext.E4Gen, cardinality)
	var accOmega field.Element
	accOmega.SetOne()

	// Precompute 1/n as a constant
	var invN field.Element
	invN.SetUint64(cardinality)
	invN.Inverse(&invN)
	bwi := big.NewInt(0)

	for i := uint64(0); i < cardinality; i++ {
		// wᵢ = ωⁱ/n
		var wi field.Element
		wi.Mul(&accOmega, &invN)
		bwi.SetUint64(wi.Uint64())

		// weightedP[i] = wᵢ * pᵢ
		weightedP[i] = *ext4.MulConst(&p[i], bwi)

		accOmega.Mul(&accOmega, &gen)
	}

	// Precompute all ωⁱ values as constants
	omegaPowers := make([]field.Element, cardinality)
	omegaPowers[0].SetOne()
	for i := uint64(1); i < cardinality; i++ {
		omegaPowers[i].Mul(&omegaPowers[i-1], &gen)
	}

	results := make([]gnarkfext.E4Gen, len(zs))
	tb := bits.TrailingZeros(uint(cardinality))

	// Evaluate at each point z
	for j, z := range zs {
		// Compute zⁿ - 1
		zPowN := z
		for i := 0; i < tb; i++ {
			zPowN = *ext4.Square(&zPowN)
		}
		zPowNMinusOne := *ext4.Sub(&zPowN, ext4.One())

		// Compute Σᵢ [ weightedP[i] / (z - ωⁱ) ]
		sum := *ext4.Zero()
		for i := uint64(0); i < cardinality; i++ {
			// z - ωⁱ
			wOmegai := zk.ValueFromKoala(omegaPowers[i])
			omegaiExt := gnarkfext.FromBase(wOmegai)
			zMinusOmegai := *ext4.Sub(&z, &omegaiExt)

			// weightedP[i] / (z - ωⁱ)
			term := *ext4.Div(&weightedP[i], &zMinusOmegai)
			sum = *ext4.Add(&sum, &term)
		}

		// P(z) = (zⁿ - 1) * sum
		results[j] = *ext4.Mul(&zPowNMinusOne, &sum)
	}

	return results
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
	wGen := big.NewInt(0).SetUint64(gen.Uint64())
	var wAccOmega zk.WrappedVariable
	for i := uint64(1); i < cardinality; i++ {
		res[i] = *ext4.MulConst(&res[i-1], wGen)         // res[i] <- ω * res[i-1]
		res[i] = *ext4.Mul(&res[i], &accZetaMinusOmegai) // res[i] <- res[i]*(ζ-ωⁱ⁻¹)
		accOmega.Mul(&accOmega, &gen)                    // accOmega <- accOmega * ω
		wAccOmega = zk.ValueFromKoala(accOmega)
		wAccOmegaExt := gnarkfext.FromBase(wAccOmega)
		accZetaMinusOmegai = *ext4.Sub(&z, &wAccOmegaExt) // accZetaMinusOmegai <- ζ-ωⁱ
		res[i] = *ext4.Div(&res[i], &accZetaMinusOmegai)  // res[i]  <- res[i]/(ζ-ωⁱ)
	}

	return res
}
