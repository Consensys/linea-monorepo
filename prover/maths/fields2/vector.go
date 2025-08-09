package field

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type anyF interface {
	Fr | Ext
}

// DeepCopy deep-copies the input vector
func DeepCopy[F anyF](pol []F) []F {
	return append([]F{}, pol...)
}

// VecScalarMul multiplies a vector by a scalar - in place. The result should
// be preallocated or it is going to panic. The res = vec is a valid parameter
// assignment. The argument can be either of the following:
//
// - [res=[]Fr, vec=[]Fr, scalar=Fr]
// - [res=[]Ext, vec=[]Fr, scalar=Fr]
// - [res=[]Ext, vec=[]Ext, scalar=Fr]
// - [res=[]Ext, vec=[]Fr, scalar=Ext]
// - [res=[]Ext, vec=[]Ext, scalar=Ext]
//
// The case res=[]Ext, vec=[]Fr, scalar=Fr is implemented but not optimized as
// it is unlikely to appear IRL.
func VecScalarMul(res, vec, scalar any) {

	var (
		resFr, isResFr   = res.([]Fr)
		vecFr, isVecFr   = vec.([]Fr)
		resExt, isResExt = res.([]Ext)
		vecExt, isVecExt = vec.([]Ext)
		sclFr, isSclFr   = scalar.(Fr)
		sclExt, isSclExt = scalar.(Ext)
	)

	if len(resFr)+len(resExt) != len(vecFr)+len(vecExt) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vecFr)+len(vecExt),
		)
	}

	switch {
	// Here we can use the SIMD optimized version of gnark
	case isResFr && isVecFr && isSclFr:
		{
			res := *unsafeCast[[]Fr, koalabear.Vector](&resFr)
			vec := *unsafeCast[[]Fr, koalabear.Vector](&vecFr)
			scalar := unsafeCast[Fr, koalabear.Element](&sclFr)
			res.ScalarMul(vec, scalar)
		}

	case isResExt && isVecExt && isSclFr:
		for i := range resExt {
			resExt[i].MulByElement(&vecExt[i], &sclFr)
		}

	case isResExt && isVecFr && isSclExt:
		for i := range resExt {
			resExt[i].MulByElement(&sclExt, &vecFr[i])
		}

	case isResExt && isVecExt && isSclExt:
		for i := range resExt {
			resExt[i].Mul(&sclExt, &vecExt[i])
		}

	case isResExt && isVecFr && isSclFr:
		for i := range resExt {
			var tmp Fr
			tmp.Mul(&sclFr, &vecFr[i])
			resExt[i].SetFr(&tmp)
		}

	default:
		utils.Panic("Invalid combination of types for VecScalarMul: res=%T vec=%T scalar=%T", res, vec, scalar)
	}
}

// VecAdd adds two vectors - into res. The result should be preallocated and it
// accepts the following combination of types.
//
// - [res=[]Fr, vec1=[]Fr, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Fr, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Ext, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Fr, vec2=[]Ext]
// - [res=[]Ext, vec1=[]Ext, vec2=[]Ext]
func VecAdd(res, vec1, vec2 any) {

	var (
		resFr, isResFr     = res.([]Fr)
		vec1Fr, isVec1Fr   = vec1.([]Fr)
		vec2Fr, isVec2Fr   = vec2.([]Fr)
		resExt, isResExt   = res.([]Ext)
		vec1Ext, isVec1Ext = vec1.([]Ext)
		vec2Ext, isVec2Ext = vec2.([]Ext)
	)

	if len(resFr)+len(resExt) != len(vec1Fr)+len(vec1Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec1Fr)+len(vec1Ext),
		)
	}

	if len(resFr)+len(resExt) != len(vec2Fr)+len(vec2Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec2Fr)+len(vec2Ext),
		)
	}

	switch {

	// This uses the SIMD implementation of gnark-crypto
	case isResFr && isVec1Fr && isVec2Fr:
		{
			res := *unsafeCast[[]Fr, koalabear.Vector](&resFr)
			vec1 := *unsafeCast[[]Fr, koalabear.Vector](&vec1Fr)
			vec2 := *unsafeCast[[]Fr, koalabear.Vector](&vec2Fr)
			res.Add(vec1, vec2)
		}

	case isResExt && isVec1Ext && isVec2Ext:
		for i := range resExt {
			resExt[i].Add(&vec1Ext[i], &vec2Ext[i])
		}

	case isResExt && isVec1Fr && isVec2Ext:
		for i := range resExt {
			resExt[i].AddMixed(&vec2Ext[i], &vec1Fr[i])
		}

	case isResExt && isVec1Ext && isVec2Fr:
		for i := range resExt {
			resExt[i].AddMixed(&vec1Ext[i], &vec2Fr[i])
		}

	case isResExt && isVec1Fr && isVec2Fr:
		for i := range resExt {
			var tmp Fr
			tmp.Add(&vec1Fr[i], &vec2Fr[i])
			resExt[i].SetFr(&tmp)
		}

	default:
		utils.Panic("Invalid combination of types for VecAdd: res=%T vec1=%T vec2=%T", res, vec1, vec2)
	}
}

// VecSub adds two vectors - into res. The result should be preallocated and it
// accepts the following combination of types.
//
// - [res=[]Fr, vec1=[]Fr, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Fr, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Ext, vec2=[]Fr]
// - [res=[]Ext, vec1=[]Fr, vec2=[]Ext]
// - [res=[]Ext, vec1=[]Ext, vec2=[]Ext]
func VecSub(res, vec1, vec2 any) {

	var (
		resFr, isResFr     = res.([]Fr)
		vec1Fr, isVec1Fr   = vec1.([]Fr)
		vec2Fr, isVec2Fr   = vec2.([]Fr)
		resExt, isResExt   = res.([]Ext)
		vec1Ext, isVec1Ext = vec1.([]Ext)
		vec2Ext, isVec2Ext = vec2.([]Ext)
	)

	if len(resFr)+len(resExt) != len(vec1Fr)+len(vec1Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec1Fr)+len(vec1Ext),
		)
	}

	if len(resFr)+len(resExt) != len(vec2Fr)+len(vec2Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec2Fr)+len(vec2Ext),
		)
	}

	switch {

	// This uses the SIMD implementation of gnark-crypto
	case isResFr && isVec1Fr && isVec2Fr:
		{
			res := *unsafeCast[[]Fr, koalabear.Vector](&resFr)
			vec1 := *unsafeCast[[]Fr, koalabear.Vector](&vec1Fr)
			vec2 := *unsafeCast[[]Fr, koalabear.Vector](&vec2Fr)
			res.Sub(vec1, vec2)
		}

	case isResExt && isVec1Ext && isVec2Ext:
		for i := range resExt {
			resExt[i].Sub(&vec1Ext[i], &vec2Ext[i])
		}

	case isResExt && isVec1Fr && isVec2Ext:
		for i := range resExt {
			resExt[i].SubFrFromBase(&vec2Ext[i], &vec1Fr[i])
		}

	case isResExt && isVec1Ext && isVec2Fr:
		for i := range resExt {
			resExt[i].SubExtFromFr(&vec1Fr[i], &vec2Ext[i])
		}

	case isResExt && isVec1Fr && isVec2Fr:
		for i := range resExt {
			var tmp Fr
			tmp.Sub(&vec1Fr[i], &vec2Fr[i])
			resExt[i].SetFr(&tmp)
		}

	default:
		utils.Panic("Invalid combination of types for VecSub: res=%T vec1=%T vec2=%T", res, vec1, vec2)
	}
}

// VecMul is as VecAdd but multiplies instead of adding
func VecMul(res, vec1, vec2 any) {

	var (
		resFr, isResFr     = res.([]Fr)
		vec1Fr, isVec1Fr   = vec1.([]Fr)
		vec2Fr, isVec2Fr   = vec2.([]Fr)
		resExt, isResExt   = res.([]Ext)
		vec1Ext, isVec1Ext = vec1.([]Ext)
		vec2Ext, isVec2Ext = vec2.([]Ext)
	)

	if len(resFr)+len(resExt) != len(vec1Fr)+len(vec1Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec1Fr)+len(vec1Ext),
		)
	}

	if len(resFr)+len(resExt) != len(vec2Fr)+len(vec2Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(resFr)+len(resExt), len(vec2Fr)+len(vec2Ext),
		)
	}

	switch {

	// This uses the SIMD implementation of gnark-crypto
	case isResFr && isVec1Fr && isVec2Fr:
		{
			res := *unsafeCast[[]Fr, koalabear.Vector](&resFr)
			vec1 := *unsafeCast[[]Fr, koalabear.Vector](&vec1Fr)
			vec2 := *unsafeCast[[]Fr, koalabear.Vector](&vec2Fr)
			res.Mul(vec1, vec2)
		}

	case isResExt && isVec1Ext && isVec2Ext:
		for i := range resExt {
			resExt[i].Mul(&vec1Ext[i], &vec2Ext[i])
		}

	case isResExt && isVec1Fr && isVec2Ext:
		for i := range resExt {
			resExt[i].MulByElement(&vec2Ext[i], &vec1Fr[i])
		}

	case isResExt && isVec1Ext && isVec2Fr:
		for i := range resExt {
			resExt[i].MulByElement(&vec1Ext[i], &vec2Fr[i])
		}

	case isResExt && isVec1Fr && isVec2Fr:
		for i := range resExt {
			var tmp Fr
			tmp.Mul(&vec1Fr[i], &vec2Fr[i])
			resExt[i].SetFr(&tmp)
		}

	default:
		utils.Panic("Invalid combination of types for VecAdd: res=%T vec1=%T vec2=%T", res, vec1, vec2)
	}
}

// InnerProduct computes the inner-product of a vector and returns the result
func InnerProduct(vec1, vec2 any) any {

	var (
		vec1Fr, isVec1Fr   = vec1.([]Fr)
		vec1Ext, isVec1Ext = vec1.([]Ext)
		vec2Fr, isVec2Fr   = vec2.([]Fr)
		vec2Ext, isVec2Ext = vec2.([]Ext)
	)

	if len(vec2Fr)+len(vec2Ext) != len(vec1Fr)+len(vec1Ext) {
		utils.Panic(
			"The lengths of res and vec should be equal, got %v and %v",
			len(vec2Fr)+len(vec2Ext), len(vec1Fr)+len(vec1Ext),
		)
	}

	switch {
	case isVec1Fr && isVec2Fr:
		{
			vec1 := *unsafeCast[[]Fr, koalabear.Vector](&vec1Fr)
			vec2 := *unsafeCast[[]Fr, koalabear.Vector](&vec2Fr)
			res := vec1.InnerProduct(vec2)
			return Fr(res)
		}

	case isVec1Fr && isVec2Ext:
		var res, tmp Ext
		for i := range vec1Fr {
			tmp.MulByElement(&vec2Ext[i], &vec1Fr[i])
			res.Add(&res, &tmp)
		}
		return res

	case isVec1Ext && isVec2Fr:
		var res, tmp Ext
		for i := range vec1Ext {
			tmp.MulByElement(&vec1Ext[i], &vec2Fr[i])
			res.Add(&res, &tmp)
		}
		return res

	case isVec1Ext && isVec2Ext:
		var res, tmp Ext
		for i := range vec1Ext {
			tmp.Mul(&vec1Ext[i], &vec2Ext[i])
			res.Add(&res, &tmp)
		}
		return res

	default:
		utils.Panic("Invalid combination of types for InnerProduct: vec1=%T vec2=%T", vec1, vec2)
	}

	return nil // This is unreachable
}

// VecRand returns a random vector of the given length
func VecRand[F interface{ SetRandom() }](length int) []F {
	vec := make([]F, length)
	for i := range vec {
		vec[i].SetRandom()
	}
	return vec
}

// VecPrettify returns a string representation of the vector
func VecPrettify[F interface{ String() string }](a []F) string {
	res := "["

	for i := range a {
		// Discards the case first element when adding a comma
		if i > 0 {
			res += ", "
		}

		res += fmt.Sprintf("%v", a[i].String())
	}
	res += "]"

	return res
}

// Reverse the elements of a vector inplace
func Reverse[F any](v []F) {
	n := len(v) - 1
	for i := 0; i < len(v)/2; i++ {
		v[i], v[n-i] = v[n-i], v[i]
	}
}

// Repeat returns a vector of size n whose values are all equal to x.
func Repeat[F any](x F, n int) []F {
	res := make([]F, n)
	for i := range res {
		res[i] = x
	}
	return res
}

// IntToVec returns a vector instantiated from a list of integers.
func IntToVecFr[F interface{ SetInt64(int64) }](xs ...int) []Fr {
	res := make([]Fr, len(xs))
	for i, x := range xs {
		res[i].SetInt64(int64(x))
	}
	return res
}

// QuadsToVec returns a []Ext from a list of [4]int.
func QuadsToVecExt(xs ...[4]int) []Ext {
	res := make([]Ext, len(xs))
	for i, x := range xs {
		res[i].B0.A0.SetInt64(int64(x[0]))
		res[i].B0.A1.SetInt64(int64(x[1]))
		res[i].B1.A0.SetInt64(int64(x[2]))
		res[i].B1.A1.SetInt64(int64(x[3]))
	}
	return res
}

// VecExtends extends a vector to the desired length. If the new length is
// smaller, the function panics.
func VecExtends[F any](v []F, newLen int) []F {
	if newLen < len(v) {
		utils.Panic("newLen (%v) < len(v) (%v)", newLen, len(v))
	}
	res := make([]F, newLen)
	copy(res, v)
	return res
}

// Interleave interleaves a list of vectors into one. The input vectors may be
// a mixture of field extensions and field elements vectors. The output vector
// will be an array of []Fr if all inputs are []Fr and an array of []Ext
// otherwise.
func Interleave(vecs ...any) any {

	if len(vecs) == 0 {
		utils.Panic("the function was provided an empty list of vectors")
	}

	if len(vecs) == 1 {
		return vecs[0]
	}

	var (
		numVecs  = len(vecs)
		allAreFr = true
		length   = -1
	)

	for i := range vecs {
		switch v := vecs[i].(type) {
		case []Fr:
			if length == -1 {
				length = len(v)
			} else if length != len(v) {
				utils.Panic("length mismatch, %v != %v", len(v), length)
			}
		case []Ext:
			if length == -1 {
				length = len(v)
			} else if length != len(v) {
				utils.Panic("length mismatch, %v != %v", len(v), length)
			}
			allAreFr = false
		default:
			utils.Panic("the function was provided a vector of type %T", v)
		}
	}

	if allAreFr {
		res := make([]Fr, length*len(vecs))
		for j := range vecs {
			v := vecs[j].([]Fr)
			for i := 0; i < length; i++ {
				res[numVecs*i+j] = v[i]
			}
		}
		return res
	}

	res := make([]Ext, length*len(vecs))
	for j := range vecs {
		switch v := vecs[j].(type) {
		case []Fr:
			for i := 0; i < length; i++ {
				res[numVecs*i+j] = v[i].Lift()
			}
		case []Ext:
			for i := 0; i < length; i++ {
				res[numVecs*i+j] = v[i]
			}
		default:
			utils.Panic("the function was provided a vector of type %T", v)
		}
	}

	return res
}

// VecFill a vector `vec` in place with the given value `val`.
func VecFill[F any](v []F, val F) {
	for i := range v {
		v[i] = val
	}
}

// VecSum computes the sum of all elements in the vector.
func VecSum[F interface{ Add(*F, *F) }](v []F) (res F) {

	// In case F is koalabear, we can use the SIMD optimized implementation
	// gnark-crypto.
	if _, ok := any(v[0]).(Fr); ok {
		v := *unsafeCast[[]F, koalabear.Vector](&v)
		res := Fr(v.Sum())
		return any(res).(F)
	}

	for i := range v {
		res.Add(&res, &v[i])
	}
	return
}

// PowerVec allocates and returns a vector of size n consisting of consecutive
// powers of x, starting from x^0 = 1 and ending on x^{n-1}. The function panics
// if given x=0 and returns an empty vector if n=0.
func PowerVec[F interface {
	Mul(*F, *F)
	SetOne()
	IsZero() bool
}](x F, n int) []F {

	if x.IsZero() {
		utils.Panic("cannot build a power vec for x=0")
	}

	if n == 0 {
		return []F{}
	}

	res := make([]F, n)
	res[0].SetOne()

	for i := 1; i < n; i++ {
		res[i].Mul(&res[i-1], &x)
	}

	return res
}

// Equal returns true if the two vectors are equal.
func Equal[F comparable](v1, v2 []F) bool {

	if len(v1) != len(v2) {
		utils.Panic("a and b don't have the same length: %v %v", len(v1), len(v2))
	}

	for i := range v1 {
		if v1[i] != v2[i] {
			return false
		}
	}

	return true
}

// VecPseudoRand generates a vector of field element with a given size using the
// provided random number generator
func VecPseudoRand[F interface{ SetPseudoRand(*rand.Rand) }](rng *rand.Rand, size int) []F {
	slice := make([]F, size)
	for i := range slice {
		slice[i].SetPseudoRand(rng)
	}
	return slice
}
