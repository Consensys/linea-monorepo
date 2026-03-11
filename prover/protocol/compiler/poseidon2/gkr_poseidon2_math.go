package poseidon2

import (
	"fmt"

	gnarkposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// computeEq computes the equality polynomial eq_r[i] = Π_k(r[k]*bit_k(i)+(1-r[k])*(1-bit_k(i)))
// for i=0..n-1 using tensor-product expansion. r must have length log2(n).
func computeEq(r []field.Element, n int) []field.Element {
	eq := make([]field.Element, n)
	eq[0].SetOne()
	var one field.Element
	one.SetOne()
	for k := 0; k < len(r); k++ {
		rk := r[k]
		var oneMinusRk field.Element
		oneMinusRk.Sub(&one, &rk)
		for i := (1 << k) - 1; i >= 0; i-- {
			eq[2*i+1].Mul(&eq[i], &rk)
			eq[2*i].Mul(&eq[i], &oneMinusRk)
		}
	}
	return eq
}

// evalMLE computes <eq, col> = Σ_i eq[i]*col[i].
func evalMLE(eq, col []field.Element) field.Element {
	var result field.Element
	for i := range eq {
		var tmp field.Element
		tmp.Mul(&eq[i], &col[i])
		result.Add(&result, &tmp)
	}
	return result
}

// evalDeg4Poly evaluates a degree ≤ 4 polynomial given its values at t=0,1,2,3,4
// at the point t via Lagrange interpolation.
func evalDeg4Poly(evals [5]field.Element, t field.Element) field.Element {
	nodes := [5]int64{0, 1, 2, 3, 4}
	var result field.Element
	for j := 0; j < 5; j++ {
		var num field.Element
		num.SetOne()
		for m := 0; m < 5; m++ {
			if m == j {
				continue
			}
			var nm field.Element
			if nodes[m] >= 0 {
				nm.SetUint64(uint64(nodes[m]))
			} else {
				nm.SetUint64(uint64(-nodes[m]))
				nm.Neg(&nm)
			}
			var diff field.Element
			diff.Sub(&t, &nm)
			num.Mul(&num, &diff)
		}
		// denom = Π_{m≠j} (j - m)
		denom := int64(1)
		for m := 0; m < 5; m++ {
			if m == j {
				continue
			}
			denom *= nodes[j] - nodes[m]
		}
		var denomF field.Element
		if denom > 0 {
			denomF.SetUint64(uint64(denom))
		} else {
			denomF.SetUint64(uint64(-denom))
			denomF.Neg(&denomF)
		}
		denomF.Inverse(&denomF)
		num.Mul(&num, &denomF)
		var contrib field.Element
		contrib.Mul(&evals[j], &num)
		result.Add(&result, &contrib)
	}
	return result
}

// linearInterp computes (1-r)*a + r*b.
func linearInterp(a, b, r field.Element) field.Element {
	var diff, res field.Element
	diff.Sub(&b, &a)
	diff.Mul(&diff, &r)
	res.Add(&a, &diff)
	return res
}

// sBoxCubeF computes x^3.
func sBoxCubeF(x field.Element) field.Element {
	var sq field.Element
	sq.Square(&x)
	sq.Mul(&sq, &x)
	return sq
}

// gkrDeriveChallenge derives a scalar sumcheck challenge from the current FS state.
func gkrDeriveChallenge(fs fiatshamir.FS) field.Element {
	return fs.RandomFext().B0.A0
}

// gkrDeriveAlpha derives an alpha batching challenge (for external rounds) from the FS state.
func gkrDeriveAlpha(fs fiatshamir.FS) field.Element {
	return fs.RandomFext().B0.A0
}

// computeEqAtPoint computes eq_r(rPrime) = Π_k(r[k]*rPrime[k]+(1-r[k])*(1-rPrime[k])).
func computeEqAtPoint(r, rPrime []field.Element) field.Element {
	var result field.Element
	result.SetOne()
	var one field.Element
	one.SetOne()
	for k := range r {
		var a, b, omR, omRP field.Element
		omR.Sub(&one, &r[k])
		omRP.Sub(&one, &rPrime[k])
		a.Mul(&r[k], &rPrime[k])
		b.Mul(&omR, &omRP)
		var factor field.Element
		factor.Add(&a, &b)
		result.Mul(&result, &factor)
	}
	return result
}

// applyAddRoundKeyClaims adds the round key for 'round' (0-indexed in gnark-crypto) to claims.
// <eq_r, 1> = 1, so constant round keys add directly.
func applyAddRoundKeyClaims(round int, claims *[width]field.Element) {
	rk := gnarkposeidon2.GetDefaultParameters().RoundKeys[round]
	for j := 0; j < len(rk); j++ {
		claims[j].Add(&claims[j], &rk[j])
	}
}

// applyMatMulExternalClaims applies M_ext to claims in-place.
func applyMatMulExternalClaims(claims *[width]field.Element) {
	matMulExternalInPlace(claims)
}

// applyMatMulInternalClaims applies M_int to claims in-place.
func applyMatMulInternalClaims(claims *[width]field.Element) {
	matMulInternalInPlace(claims)
}

// sBoxExternalClaims returns oracle_j^3 for all j (post-sBox, pre-matMul).
func sBoxExternalClaims(oracleEvals [width]field.Element) [width]field.Element {
	var c [width]field.Element
	for w := 0; w < width; w++ {
		c[w] = sBoxCubeF(oracleEvals[w])
	}
	return c
}

// sBoxInternalClaims returns oracle_0^3 for j=0, oracle_j for j>0 (partial sBox).
func sBoxInternalClaims(oracleEvals [width]field.Element) [width]field.Element {
	var c [width]field.Element
	c[0] = sBoxCubeF(oracleEvals[0])
	for w := 1; w < width; w++ {
		c[w] = oracleEvals[w]
	}
	return c
}

// reverseSlice returns a new slice with elements in reverse order.
func reverseSlice(s []field.Element) []field.Element {
	r := make([]field.Element, len(s))
	for i, v := range s {
		r[len(s)-1-i] = v
	}
	return r
}

// gkrSumcheckProveLayer runs the sumcheck prover for one SBox layer.
// bookEq (size n) and bookState[w] (size n each) are modified in-place by folding.
// bookState already holds post-SBox values: external layers have all 16 cols cubed,
// internal layers have only col 0 cubed. The sumcheck is therefore LINEAR in bookState.
// After the call, bookEq[0] and bookState[w][0] are the final oracle evals.
func gkrSumcheckProveLayer(
	bookEq []field.Element,
	bookState *[width][]field.Element,
	isExternal bool,
	alpha field.Element,
	logN int,
	fs fiatshamir.FS,
) (polys [][5]field.Element, newPoint []field.Element) {

	polys = make([][5]field.Element, logN)
	newPoint = make([]field.Element, logN)
	n := len(bookEq)

	for k := 0; k < logN; k++ {
		halfN := n / 2

		// Compute g_k(t) for t=0..4.
		// bookState already holds post-SBox values, so fVal is a linear combination.
		for tIdx := 0; tIdx < 5; tIdx++ {
			var tF field.Element
			tF.SetUint64(uint64(tIdx))

			var gSum field.Element
			for j := 0; j < halfN; j++ {
				eqAtT := linearInterp(bookEq[2*j], bookEq[2*j+1], tF)

				var fVal field.Element
				if isExternal {
					var alphaPow field.Element
					alphaPow.SetOne()
					for w := 0; w < width; w++ {
						sAtT := linearInterp(bookState[w][2*j], bookState[w][2*j+1], tF)
						var term field.Element
						term.Mul(&alphaPow, &sAtT)
						fVal.Add(&fVal, &term)
						alphaPow.Mul(&alphaPow, &alpha)
					}
				} else {
					fVal = linearInterp(bookState[0][2*j], bookState[0][2*j+1], tF)
				}

				var contrib field.Element
				contrib.Mul(&eqAtT, &fVal)
				gSum.Add(&gSum, &contrib)
			}
			polys[k][tIdx] = gSum
		}

		// Absorb polys, derive challenge, fold
		fs.Update(polys[k][:]...)
		rk := gkrDeriveChallenge(fs)
		newPoint[k] = rk

		for j := 0; j < halfN; j++ {
			bookEq[j] = linearInterp(bookEq[2*j], bookEq[2*j+1], rk)
			for w := 0; w < width; w++ {
				bookState[w][j] = linearInterp(bookState[w][2*j], bookState[w][2*j+1], rk)
			}
		}
		n = halfN
	}

	return polys, newPoint
}

// gkrSumcheckVerifyLayer verifies one sumcheck layer transcript.
// It replays the sumcheck: checks g_k(0)+g_k(1)=runningSum at each round,
// derives challenges from fs, and finally checks that the oracle evals are consistent.
// The final consistency check: runningSum_after_logN == eq_r(newPoint) * f(oracleEvals).
// We compute eq_r(newPoint) from oldR and newPoint, and f(oracleEvals) from oracle evals.
// Returns the new evaluation point and any error.
func gkrSumcheckVerifyLayer(
	polys [][5]field.Element,
	oracleEvals [width]field.Element,
	currentSum field.Element,
	oldR []field.Element,
	isExternal bool,
	alpha field.Element,
	fs fiatshamir.FS,
) (newPoint []field.Element, err error) {

	logN := len(polys)
	newPoint = make([]field.Element, logN)
	runningSum := currentSum

	for k := 0; k < logN; k++ {
		var checkVal field.Element
		checkVal.Add(&polys[k][0], &polys[k][1])
		if checkVal != runningSum {
			err = fmt.Errorf("GKR sumcheck round %d: g(0)+g(1)=%v ≠ claim=%v", k, checkVal, runningSum)
			return
		}
		fs.Update(polys[k][:]...)
		rk := gkrDeriveChallenge(fs)
		newPoint[k] = rk
		runningSum = evalDeg4Poly(polys[k], rk)
	}

	// Compute f(oracleEvals). Oracle holds post-SBox values, so no cube needed.
	var fOracle field.Element
	if isExternal {
		var alphaPow field.Element
		alphaPow.SetOne()
		for w := 0; w < width; w++ {
			var term field.Element
			term.Mul(&alphaPow, &oracleEvals[w])
			fOracle.Add(&fOracle, &term)
			alphaPow.Mul(&alphaPow, &alpha)
		}
	} else {
		fOracle = oracleEvals[0]
	}

	// Compute eq_r(newPoint) and verify final claim.
	// computeEq uses big-endian bit ordering: bit k of index i corresponds to r[logN-1-k].
	// Folding round k fixes bit k (↔ r[logN-1-k]), so after all folds:
	//   bookEq[0] = computeEqAtPoint(r, reverse(newPoint))
	revNewPoint := make([]field.Element, len(newPoint))
	for k := range newPoint {
		revNewPoint[k] = newPoint[len(newPoint)-1-k]
	}
	eqVal := computeEqAtPoint(oldR, revNewPoint)
	var expected field.Element
	expected.Mul(&eqVal, &fOracle)
	if expected != runningSum {
		err = fmt.Errorf("GKR oracle check: eq*f=%v ≠ runningSum=%v", expected, runningSum)
		return
	}

	// Absorb oracle evals into FS
	fs.Update(oracleEvals[:]...)

	return newPoint, nil
}
