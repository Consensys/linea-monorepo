package public_input

import (
	"fmt"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"math/bits"
	"slices"
)

func interpolateLagrangeBls12381(field *emulated.Field[emulated.BLS12381Fr], unitCircleEvaluations []*emulated.Element[emulated.BLS12381Fr], evaluationPoint *emulated.Element[emulated.BLS12381Fr]) (evaluation *emulated.Element[emulated.BLS12381Fr], err error) {
	logN := bits.Len(uint(len(unitCircleEvaluations))) - 1
	if len(unitCircleEvaluations) != 1<<logN {
		return nil, fmt.Errorf("expected power of 2 number of evaluations, got %d", len(unitCircleEvaluations))
	}

	x, nInv := genInterpolateLagrangeParams(len(unitCircleEvaluations))

	summands := make([]*emulated.Element[emulated.BLS12381Fr], len(unitCircleEvaluations))
	for i := range summands {
		xI := field.NewElement(x[i])
		summands[i] = field.Div(
			field.Mul(unitCircleEvaluations[i], xI),
			field.Sub(evaluationPoint, xI),
		)
	}

	minPolyAtEp := evaluationPoint
	for i := 0; i < logN; i++ {
		minPolyAtEp = field.Mul(minPolyAtEp, minPolyAtEp)
	}
	minPolyAtEp = field.Sub(minPolyAtEp, field.NewElement(1))

	coeff := field.Mul(field.NewElement(nInv), minPolyAtEp)

	return field.Mul(coeff, sum(field, summands...)), nil
}

func sum[T emulated.FieldParams](field *emulated.Field[T], elems ...*emulated.Element[T]) *emulated.Element[T] {
	switch len(elems) {
	case 0:
		return field.NewElement(0)
	case 1:
		return elems[0]
	}
	res := field.Add(elems[0], elems[1])
	for i := 2; i < len(elems); i++ {
		res = field.Add(res, elems[i])
	}
	return res
}

// compile time parameter generation
// returns the nth roots of unity and 1/n
func genInterpolateLagrangeParams(n int) (x []fr381.Element, nInv fr381.Element) {
	x = make([]fr381.Element, n)
	x1, err := fr381.Generator(uint64(n)) // x1 is the canonical nth root of unity
	if err != nil {
		panic(err)
	}
	x[0].SetOne()
	x[1] = x1
	for i := 2; i < len(x); i++ {
		x[i].Mul(&x[i-1], &x1)
	}

	nInv.SetUint64(uint64(n)).Inverse(&nInv)

	return
}

// VerifyBlobConsistency opens the "commitment" to the blob at evaluationChallenge; if bypassEip4844 is set, it does so in a KZG-like manner using a Lagrange basis on the unit circle. if not, a Reed-Solomon type method is used.
// TODO consider using the batch hashes as "snarkHash" instead of hashing the data here to save on constraints
func VerifyBlobConsistency(api frontend.API, blobBits []frontend.Variable, evaluationChallenge [2]frontend.Variable, eip4844Enabled frontend.Variable) (evaluation [2]frontend.Variable, err error) {
	snarkFieldLen := api.Compiler().Field().BitLen()
	if snarkFieldLen >= fr381.Bits {
		err = fmt.Errorf("large field moduli ( %d≥%d ) not yet supported", snarkFieldLen, fr381.Bits)
		return
	}

	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		return
	}

	blobEmulated := packBitsEmulated(api, blobBits) // perf-TODO use the original blob bytes
	evaluationChallengeEmulated := newElementFromVars(api, evaluationChallenge)

	blobEmulatedBitReversed := make([]*emulated.Element[emulated.BLS12381Fr], len(blobEmulated))
	copy(blobEmulatedBitReversed, blobEmulated)
	bitReverseSlice(blobEmulatedBitReversed)
	lagrangeEval, err := interpolateLagrangeBls12381(field, blobEmulatedBitReversed, evaluationChallengeEmulated)
	if err != nil {
		return
	}

	polyEval, err := evalPolyBls12381(field, blobEmulated, evaluationChallengeEmulated)
	if err != nil {
		return
	}
	l := bls12381ScalarToBls12377Scalars(api, lagrangeEval)
	p := bls12381ScalarToBls12377Scalars(api, polyEval)

	api.AssertIsBoolean(eip4844Enabled)
	evaluation[0] = api.Select(eip4844Enabled, l[0], p[0])
	evaluation[1] = api.Select(eip4844Enabled, l[1], p[1])

	return
}

func evalPolyBls12381(field *emulated.Field[emulated.BLS12381Fr], coeffs []*emulated.Element[emulated.BLS12381Fr], evaluationPoint *emulated.Element[emulated.BLS12381Fr]) (evaluation *emulated.Element[emulated.BLS12381Fr], err error) {
	// lower degree coeff first
	evaluation = coeffs[len(coeffs)-1]
	for i := len(coeffs) - 2; i >= 0; i-- {
		evaluation = field.Mul(evaluation, evaluationPoint)
		evaluation = field.Add(evaluation, coeffs[i])
	}
	return
}

func mapSlice[X, Y any](slice []X, f func(X) Y) []Y {
	res := make([]Y, len(slice))
	for i, x := range slice {
		res[i] = f(x)
	}
	return res
}

func newElementFromVars(api frontend.API, x [2]frontend.Variable) *emulated.Element[emulated.BLS12381Fr] {
	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		panic(err)
	}
	snarkFieldLen := api.Compiler().Field().BitLen()
	lSize := snarkFieldLen - 1
	lBin := api.ToBinary(x[1], lSize)
	hSize := fr381.Bits - lSize
	hBin := api.ToBinary(x[0], hSize)
	return field.FromBits(append(lBin, hBin...)...)
}

func bitReverse(n, logN int) int {
	return int(bits.Reverse64(uint64(n)) >> (64 - logN))
}

func bitReverseSlice[K interface{}](list []K) {
	n := uint64(len(list))

	// The standard library's bits.Reverse64 inverts its input as a 64-bit unsigned integer.
	// However, we need to invert it as a log2(len(list))-bit integer, so we need to correct this by
	// shifting appropriately.
	shiftCorrection := uint64(64 - bits.TrailingZeros64(n))

	for i := uint64(0); i < n; i++ {
		// Find index irev, such that i and irev get swapped
		irev := bits.Reverse64(i) >> shiftCorrection
		if irev > i {
			list[i], list[irev] = list[irev], list[i]
		}
	}
}

func packBitsEmulated(api frontend.API, bits []frontend.Variable) []*emulated.Element[emulated.BLS12381Fr] {
	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		panic(err)
	}
	const bitsPerElem = fr381.Bits - 1
	res := make([]*emulated.Element[emulated.BLS12381Fr], (len(bits)+bitsPerElem-1)/bitsPerElem)
	if len(bits) != len(res)*bitsPerElem {
		tmp := bits
		bits = make([]frontend.Variable, len(res)*bitsPerElem)
		copy(bits, tmp)
		for i := len(tmp); i < len(bits); i++ {
			bits[i] = 0
		}
	}
	for i := range res {
		currBits := bits[i*bitsPerElem : (i+1)*bitsPerElem]
		slices.Reverse(currBits)
		res[i] = field.FromBits(currBits...)
		slices.Reverse(currBits) // don't modify the input
	}
	return res
}

// bls12377ScalarToBls12381Scalar converts a scalar in the BLS12-377 field to a scalar in the BLS12-381 field. It assumes the input is only 252 bits long to accommodate arbitrary data
func bls12377ScalarToBls12381Scalar(api frontend.API, v frontend.Variable) *emulated.Element[emulated.BLS12381Fr] {
	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		panic(err)
	}
	return field.FromBits(api.ToBinary(v, api.Compiler().FieldBitLen()-1)...)
}

func bls12381ScalarToBls12377Scalars(api frontend.API, e *emulated.Element[emulated.BLS12381Fr]) (r [2]frontend.Variable) {
	field, err := emulated.NewField[emulated.BLS12381Fr](api)
	if err != nil {
		panic(err)
	}
	snarkFieldLen := api.Compiler().FieldBitLen()
	e = field.Reduce(e)
	bts := field.ToBits(e)
	r[1] = api.FromBinary(bts[:snarkFieldLen-1]...)
	r[0] = api.FromBinary(bts[snarkFieldLen-1:]...)
	return
}
