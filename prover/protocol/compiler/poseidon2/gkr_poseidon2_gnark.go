package poseidon2

import (
	"math/big"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// RunGnark implements wizard.VerifierAction.RunGnark for the GKR Poseidon2 verifier.
// It mirrors the native Run method using gnark circuit arithmetic.
func (va *GKRPoseidon2VerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	ctx := va.Ctx
	logN := ctx.LogN
	koalaAPI := koalagnark.NewAPI(api)

	// 1. Extract r0 from FieldExt coins (same as gkrExtractR0FromRuntime but in gnark).
	r0 := make([]koalagnark.Element, logN)
	k := 0
	for _, name := range ctx.R0CoinNames {
		fextVal := run.GetRandomCoinFieldExt(name)
		comps := [4]koalagnark.Element{fextVal.B0.A0, fextVal.B0.A1, fextVal.B1.A0, fextVal.B1.A1}
		for _, c := range comps {
			if k >= logN {
				break
			}
			r0[k] = c
			k++
		}
	}

	// 2. Get InnerProduct params. Ys[w] are koalagnark.Ext; base field component is .B0.A0.
	inputParams := run.GetInnerProductParams(ctx.InputIPID)
	outputParams := run.GetInnerProductParams(ctx.OutputIPID)

	// 3. Read GKR transcript from the Proof column.
	transcriptElems := run.GetColumn(ctx.GKRTranscript.GetColID())

	// 4. Initialize claims from inputParams.
	var claims [width]koalagnark.Element
	for w := 0; w < width; w++ {
		claims[w] = inputParams.Ys[w].B0.A0
	}

	// 5. Apply initial matMulExternal and addRoundKey[0] to get preSBox state for layer 0.
	matMulExternalInPlaceGnark(koalaAPI, &claims)
	applyAddRoundKeyClaimsGnark(0, koalaAPI, &claims)

	// 6. Initialize internal GKR FS seeded with r0 (separate from wizard FS).
	fs := fiatshamir.NewGnarkFSKoalabear(api)
	fs.Update(r0...)

	currentR := r0
	perLayer := gkrTranscriptPerLayer(logN)

	// 7. For each GKR layer, verify the sumcheck transcript.
	for l := 0; l < numGKRLayers; l++ {
		isExt := isGKRExternalLayer(l)
		offset := l * perLayer

		// Extract polys: logN groups of 5 field elements.
		polys := make([][5]koalagnark.Element, logN)
		for kk := 0; kk < logN; kk++ {
			for t := 0; t < 5; t++ {
				polys[kk][t] = transcriptElems[offset+kk*5+t]
			}
		}

		// Extract oracle evals: width elements after the polys.
		var oracleEvals [width]koalagnark.Element
		oracleBase := offset + logN*5
		for w := 0; w < width; w++ {
			oracleEvals[w] = transcriptElems[oracleBase+w]
		}

		// Derive alpha for external rounds.
		var alpha koalagnark.Element
		if isExt {
			alpha = gkrDeriveAlphaGnark(fs)
		}

		// Compute the initial sumcheck claim for this layer.
		var currentSum koalagnark.Element
		if logN > 0 {
			currentSum = koalaAPI.Add(polys[0][0], polys[0][1])
		} else if isExt {
			// logN=0: single instance, sum = Σ_w α^w * claims[w]^3.
			alphaPow := koalaAPI.One()
			for w := 0; w < width; w++ {
				sq := koalaAPI.Mul(claims[w], claims[w])
				cube := koalaAPI.Mul(sq, claims[w])
				term := koalaAPI.Mul(alphaPow, cube)
				currentSum = koalaAPI.Add(currentSum, term)
				alphaPow = koalaAPI.Mul(alphaPow, alpha)
			}
		} else {
			// logN=0 internal: single instance, sum = claims[0]^3.
			sq := koalaAPI.Mul(claims[0], claims[0])
			currentSum = koalaAPI.Mul(sq, claims[0])
		}

		// Verify sumcheck and advance currentR.
		currentR = gkrSumcheckVerifyLayerGnark(koalaAPI, fs, polys, oracleEvals, currentSum, currentR, isExt, alpha)

		// Update claims: oracle evals → apply matMul → apply addRoundKey for next layer.
		claims = oracleEvals
		if isExt {
			matMulExternalInPlaceGnark(koalaAPI, &claims)
		} else {
			matMulInternalInPlaceGnark(koalaAPI, &claims)
		}
		if l < numGKRLayers-1 {
			applyAddRoundKeyClaimsGnark(l+1, koalaAPI, &claims)
		}
	}

	// 8. Final output check: claims[8+j] + outputParams.Ys[8+j].B0.A0 == outputParams.Ys[j].B0.A0
	for j := 0; j < blockSize; j++ {
		expected := koalaAPI.Add(claims[blockSize+j], outputParams.Ys[blockSize+j].B0.A0)
		koalaAPI.AssertIsEqual(expected, outputParams.Ys[j].B0.A0)
	}
}

// gkrDeriveChallengeGnark derives a sumcheck challenge from the gnark FS state.
// Mirrors gkrDeriveChallenge: takes the B0.A0 component of a squeezed fext element.
func gkrDeriveChallengeGnark(fs fiatshamir.GnarkFS) koalagnark.Element {
	return fs.RandomFieldExt().B0.A0
}

// gkrDeriveAlphaGnark derives an alpha batching challenge for external rounds.
func gkrDeriveAlphaGnark(fs fiatshamir.GnarkFS) koalagnark.Element {
	return fs.RandomFieldExt().B0.A0
}

// gkrSumcheckVerifyLayerGnark verifies one sumcheck layer in gnark.
// Mirrors gkrSumcheckVerifyLayer: checks g(0)+g(1)==runningSum each round, derives
// challenges from fs, evaluates the degree-4 polynomial, and does the final oracle check.
func gkrSumcheckVerifyLayerGnark(
	koalaAPI *koalagnark.API,
	fs fiatshamir.GnarkFS,
	polys [][5]koalagnark.Element,
	oracleEvals [width]koalagnark.Element,
	currentSum koalagnark.Element,
	oldR []koalagnark.Element,
	isExternal bool,
	alpha koalagnark.Element,
) []koalagnark.Element {
	logN := len(polys)
	newPoint := make([]koalagnark.Element, logN)
	runningSum := currentSum

	for k := 0; k < logN; k++ {
		// Check g(0) + g(1) == runningSum.
		check := koalaAPI.Add(polys[k][0], polys[k][1])
		koalaAPI.AssertIsEqual(check, runningSum)

		// Absorb polys into FS and derive challenge.
		fs.Update(polys[k][:]...)
		rk := gkrDeriveChallengeGnark(fs)
		newPoint[k] = rk

		// Evaluate degree-4 polynomial at rk.
		runningSum = evalDeg4PolyGnark(koalaAPI, polys[k], rk)
	}

	// Compute f(oracleEvals). Oracle holds post-SBox values (no cube needed).
	var fOracle koalagnark.Element
	if isExternal {
		fOracle = koalaAPI.Zero()
		alphaPow := koalaAPI.One()
		for w := 0; w < width; w++ {
			term := koalaAPI.Mul(alphaPow, oracleEvals[w])
			fOracle = koalaAPI.Add(fOracle, term)
			alphaPow = koalaAPI.Mul(alphaPow, alpha)
		}
	} else {
		fOracle = oracleEvals[0]
	}

	// Compute eq_r(newPoint) and verify final claim.
	// Uses the same reverse-newPoint trick as the native verifier.
	revNewPoint := make([]koalagnark.Element, len(newPoint))
	for k := range newPoint {
		revNewPoint[k] = newPoint[len(newPoint)-1-k]
	}
	eqVal := computeEqAtPointGnark(koalaAPI, oldR, revNewPoint)
	expected := koalaAPI.Mul(eqVal, fOracle)
	koalaAPI.AssertIsEqual(expected, runningSum)

	// Absorb oracle evals into FS.
	fs.Update(oracleEvals[:]...)

	return newPoint
}

// evalDeg4PolyGnark evaluates a degree ≤ 4 polynomial (given values at t=0,1,2,3,4)
// at the gnark variable t using Lagrange interpolation. Mirrors evalDeg4Poly.
func evalDeg4PolyGnark(koalaAPI *koalagnark.API, evals [5]koalagnark.Element, t koalagnark.Element) koalagnark.Element {
	// Precompute (t - m) for m = 0..4.
	diffs := [5]koalagnark.Element{}
	for m := 0; m < 5; m++ {
		mConst := koalaAPI.ConstBig(big.NewInt(int64(m)))
		diffs[m] = koalaAPI.Sub(t, mConst)
	}

	result := koalaAPI.Zero()
	for j := 0; j < 5; j++ {
		// Numerator: product of (t - m) for m ≠ j (4 terms).
		ki := 0
		var num [4]koalagnark.Element
		for m := 0; m < 5; m++ {
			if m != j {
				num[ki] = diffs[m]
				ki++
			}
		}
		numProd := koalaAPI.Mul(koalaAPI.Mul(num[0], num[1]), koalaAPI.Mul(num[2], num[3]))

		// Denominator: Π_{m≠j} (j - m) as a constant field element, then inverted.
		denom := int64(1)
		for m := 0; m < 5; m++ {
			if m != j {
				denom *= int64(j) - int64(m)
			}
		}
		var denomF field.Element
		if denom > 0 {
			denomF.SetUint64(uint64(denom))
		} else {
			denomF.SetUint64(uint64(-denom))
			denomF.Neg(&denomF)
		}
		denomF.Inverse(&denomF)
		denomBig := denomF.BigInt(new(big.Int))

		term := koalaAPI.MulConst(koalaAPI.Mul(evals[j], numProd), denomBig)
		result = koalaAPI.Add(result, term)
	}
	return result
}

// computeEqAtPointGnark computes eq_r(rPrime) = Π_k(r[k]*rPrime[k]+(1-r[k])*(1-rPrime[k])).
// Mirrors computeEqAtPoint.
func computeEqAtPointGnark(koalaAPI *koalagnark.API, r, rPrime []koalagnark.Element) koalagnark.Element {
	result := koalaAPI.One()
	one := koalaAPI.One()
	for k := range r {
		omR := koalaAPI.Sub(one, r[k])
		omRP := koalaAPI.Sub(one, rPrime[k])
		a := koalaAPI.Mul(r[k], rPrime[k])
		b := koalaAPI.Mul(omR, omRP)
		factor := koalaAPI.Add(a, b)
		result = koalaAPI.Mul(result, factor)
	}
	return result
}

// matMulExternalInPlaceGnark applies the external Poseidon2 matrix multiplication in gnark.
// Mirrors matMulExternalInPlace using koalaAPI arithmetic.
func matMulExternalInPlaceGnark(koalaAPI *koalagnark.API, input *[width]koalagnark.Element) {
	var m4 [width]koalagnark.Element

	for i := 0; i < 4; i++ {
		a0 := input[4*i]
		a1 := input[4*i+1]
		a2 := input[4*i+2]
		a3 := input[4*i+3]

		tmp0 := koalaAPI.Add(a0, a1)          // a0+a1
		tmp1 := koalaAPI.Add(a2, a3)          // a2+a3
		tmp2 := koalaAPI.Add(tmp0, tmp1)       // a0+a1+a2+a3
		tmp3 := koalaAPI.Add(tmp2, a1)         // a0+2a1+a2+a3
		tmp4 := koalaAPI.Add(tmp2, a3)         // a0+a1+a2+2a3
		m4[4*i]   = koalaAPI.Add(tmp0, tmp3)   // 2a0+3a1+a2+a3
		m4[4*i+1] = koalaAPI.Add(koalaAPI.Add(a2, a2), tmp3) // 2a2 + (a0+2a1+a2+a3) = a0+2a1+3a2+a3
		m4[4*i+2] = koalaAPI.Add(tmp1, tmp4)   // a0+a1+2a2+3a3
		m4[4*i+3] = koalaAPI.Add(koalaAPI.Add(a0, a0), tmp4) // 2a0 + (a0+a1+a2+2a3) = 3a0+a1+a2+2a3
	}

	zero := koalaAPI.Zero()
	t := [tSize]koalagnark.Element{zero, zero, zero, zero}
	for i := 0; i < 4; i++ {
		t[0] = koalaAPI.Add(t[0], m4[4*i])
		t[1] = koalaAPI.Add(t[1], m4[4*i+1])
		t[2] = koalaAPI.Add(t[2], m4[4*i+2])
		t[3] = koalaAPI.Add(t[3], m4[4*i+3])
	}

	for i := 0; i < 4; i++ {
		input[4*i]   = koalaAPI.Add(m4[4*i],   t[0])
		input[4*i+1] = koalaAPI.Add(m4[4*i+1], t[1])
		input[4*i+2] = koalaAPI.Add(m4[4*i+2], t[2])
		input[4*i+3] = koalaAPI.Add(m4[4*i+3], t[3])
	}
}

// matMulInternalInPlaceGnark applies the internal Poseidon2 matrix multiplication in gnark.
// Mirrors matMulInternalInPlace: computes sum then applies diagonal multipliers.
func matMulInternalInPlaceGnark(koalaAPI *koalagnark.API, input *[width]koalagnark.Element) {
	// Save original values before modification.
	orig := *input

	// Compute sum of all original elements.
	sBoxSum := orig[0]
	for i := 1; i < width; i++ {
		sBoxSum = koalaAPI.Add(sBoxSum, orig[i])
	}

	// Diagonal multipliers from the native implementation:
	// diag = [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
	// Stored as field constants: two=2, half=1065353217, halfExp3=1864368129,
	// halfExp4=1997537281, halfExp8=2122383361, halfExp24=127 (= -1/2^24 mod p).
	two       := big.NewInt(2)
	three     := big.NewInt(3)
	four      := big.NewInt(4)
	half      := big.NewInt(1065353217) // 1/2 mod p
	halfExp3  := big.NewInt(1864368129) // 1/8 mod p
	halfExp4  := big.NewInt(1997537281) // 1/16 mod p
	halfExp8  := big.NewInt(2122383361) // 1/2^8 mod p
	halfExp24 := big.NewInt(127)        // -1/2^24 mod p  (so -diag[11]*x = 127*x)

	input[0]  = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[0],  two))
	input[1]  = koalaAPI.Add(sBoxSum, orig[1])
	input[2]  = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[2],  two))
	input[3]  = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[3],  half))
	input[4]  = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[4],  three))
	input[5]  = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[5],  four))
	input[6]  = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[6],  half))
	input[7]  = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[7],  three))
	input[8]  = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[8],  four))
	input[9]  = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[9],  halfExp8))
	input[10] = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[10], halfExp3))
	input[11] = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[11], halfExp24))
	input[12] = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[12], halfExp8))
	input[13] = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[13], halfExp3))
	input[14] = koalaAPI.Sub(sBoxSum, koalaAPI.MulConst(orig[14], halfExp4))
	input[15] = koalaAPI.Add(sBoxSum, koalaAPI.MulConst(orig[15], halfExp24))
}

// applyAddRoundKeyClaimsGnark adds the round key constants for 'round' (0-indexed) to claims.
// Mirrors applyAddRoundKeyClaims.
func applyAddRoundKeyClaimsGnark(round int, koalaAPI *koalagnark.API, claims *[width]koalagnark.Element) {
	rk := gnarkposeidon2.GetDefaultParameters().RoundKeys[round]
	for j := 0; j < len(rk); j++ {
		rkBig := rk[j].BigInt(new(big.Int))
		rkConst := koalaAPI.ConstBig(rkBig)
		claims[j] = koalaAPI.Add(claims[j], rkConst)
	}
}
