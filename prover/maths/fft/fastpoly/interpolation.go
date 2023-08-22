package fastpoly

import (
	"math/big"
	"runtime"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

/*
Given the evaluations of a polynomial on a domain (whose
size must be a power of two, or panic), return an evaluation
at a chosen x.

As an input the user can specify that the inputs are given
on a coset.
*/
func Interpolate(poly []field.Element, x fr.Element, oncoset ...bool) field.Element {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(n)
	denominator := make([]field.Element, n)

	one := field.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.Mul(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].Mul(&denominator[i-1], &domain.GeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)
		if denominator[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			return poly[i]
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = field.BatchInvert(denominator)
	res := vector.ScalarProd(poly, denominator)

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	var factor field.Element
	factor.Exp(x, big.NewInt(int64(n)))
	factor.Sub(&factor, &one)
	factor.Mul(&factor, &domain.CardinalityInv)
	res.Mul(&res, &factor)

	return res

}

/*
Given the evaluations of a polynomial on a domain (whose
size must be a power of two, or panic), return an evaluation
at a chosen x.

As an input the user can specify that the inputs are given
on a coset.
*/
func ParInterpolate(poly []field.Element, x fr.Element, numcpus int, oncoset ...bool) field.Element {

	if numcpus == 0 {
		numcpus = runtime.GOMAXPROCS(0)
	}

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(n)
	denominator := make([]field.Element, n)

	one := field.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.Mul(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/

	parallel.Execute(n, func(start, stop int) {
		// denominator start
		denominator[start].Exp(domain.GeneratorInv, big.NewInt(int64(start)))
		denominator[start].Mul(&denominator[start], &x)

		for i := start + 1; i < stop; i++ {
			denominator[i].Mul(&denominator[i-1], &domain.GeneratorInv)
		}

		for i := start; i < stop; i++ {
			denominator[i].Sub(&denominator[i], &one)
		}
	}, numcpus)

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = field.ParBatchInvert(denominator, numcpus)
	res := vector.ParScalarProd(poly, denominator, numcpus)

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	var factor field.Element
	factor.Exp(x, big.NewInt(int64(n)))
	factor.Sub(&factor, &one)
	factor.Mul(&factor, &domain.CardinalityInv)
	res.Mul(&res, &factor)

	return res

}

/*
Given the evaluations of a polynomial on a domain (whose
size must be a power of two, or panic), return an evaluation
at a chosen x.

As an input the user can specify that the inputs are given
on a coset.
*/
func InterpolateWithBuff(poly []field.Element, x fr.Element, denominatorBuf, batchInvertBuff []field.Element, oncoset ...bool) field.Element {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	n := len(poly)

	domain := fft.NewDomain(n)

	if len(denominatorBuf) != n {
		utils.Panic("Passed a non-nil denominator buf but its size (%v) mismatches the polys (%v)", len(denominatorBuf), n)
	}
	denominator := denominatorBuf

	one := field.One()

	if len(oncoset) > 0 && oncoset[0] {
		x.Mul(&x, &domain.FrMultiplicativeGenInv)
	}

	/*
		First, we compute the denominator,

		D_x = \frac{X}{x} - g for x \in H
			where H is the subgroup of the roots of unity (not the coset)
			and g a field element such that gH is the coset
	*/
	denominator[0] = x
	for i := 1; i < n; i++ {
		denominator[i].Mul(&denominator[i-1], &domain.GeneratorInv)
	}

	for i := 0; i < n; i++ {
		denominator[i].Sub(&denominator[i], &one)
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominator = field.BatchInvertWithBuffer(batchInvertBuff, denominator)
	res := vector.ScalarProd(poly, denominator)

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	var factor field.Element
	factor.Exp(x, big.NewInt(int64(n)))
	factor.Sub(&factor, &one)
	factor.Mul(&factor, &domain.CardinalityInv)
	res.Mul(&res, &factor)

	return res

}
