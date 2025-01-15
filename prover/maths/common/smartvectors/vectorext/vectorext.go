// vector offers a set of utility function relating to slices of field element
// and that are commonly used as part of the repo.
package vectorext

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"math/rand"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// DeepCopy deep-copies the input vector
func DeepCopy(pol []fext.Element) []fext.Element {
	return append([]fext.Element{}, pol...)
}

// ScalarMul multiplies a vector by a scalar - in place.
// The result should be preallocated or it is going to panic.
// res = vec is a valid parameter assignment.
func ScalarMul(res, vec []fext.Element, scalar fext.Element) {

	if len(res)+len(vec) == 0 {
		return
	}

	r := Vector(res)
	r.ScalarMul(Vector(vec), &scalar)
}

// ScalarProd returns the scalar (inner) product of a and b. The function panics
// if a and b do not have the same size. If they have both empty vectors, the
// function returns 0.
func ScalarProd(a, b []fext.Element) fext.Element {
	// The length checks is done by gnark-crypto already
	a_ := Vector(a)
	res := a_.InnerProduct(Vector(b))
	return res
}

// Rand creates a random vector of size n
func Rand(n int) []fext.Element {
	vec := make([]fext.Element, n)
	for i := range vec {
		_, err := vec[i].SetRandom()
		// Just to enfore never having to deal with zeroes
		if err != nil {
			panic(err)
		}
	}
	return vec
}

// MulElementWise multiplies two vectors element wise and write the result in
// res. res = a is a valid assignment.
func MulElementWise(res, a, b []fext.Element) {
	// The length checks is done by gnark-crypto already
	res_ := Vector(res)
	res_.Mul(Vector(a), Vector(b))
}

// Prettify returns a string representing `a` in a human-readable fashion
func Prettify(a []fext.Element) string {
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
func Reverse(v []fext.Element) {
	n := len(v) - 1
	for i := 0; i < len(v)/2; i++ {
		v[i], v[n-i] = v[n-i], v[i]
	}
}

// Repeat returns a vector of size n whose values are all equal to x.
func Repeat(x fext.Element, n int) []fext.Element {
	res := make([]fext.Element, n)
	for i := range res {
		res[i].Set(&x)
	}
	return res
}

// ForTest returns a vector instantiated from a list of integers.
func ForTest(xs ...int) []fext.Element {
	res := make([]fext.Element, len(xs))
	for i, x := range xs {
		res[i].SetInt64(int64(x))
	}
	return res
}

// Add adds two vectors `a` and `b` and put the result in `res`
// `res` must be pre-allocated by the caller and res, a and b must all have
// the same size.
// res == a or res == b or both is valid assignment.
func Add(res, a, b []fext.Element, extras ...[]fext.Element) {

	if len(res)+len(a)+len(b) == 0 {
		return
	}

	r := Vector(res)
	r.Add(a, b)

	for _, x := range extras {
		r.Add(r, Vector(x))
	}
}

func AddExt(res, a, b []fext.Element, extras ...[]fext.Element) {

	if len(res)+len(a)+len(b) == 0 {
		return
	}

	r := Vector(res)
	r.Add(a, b)

	for _, x := range extras {
		r.Add(r, Vector(x))
	}
}

// Sub substracts two vectors `a` and `b` and put the result in `res`
// `res` must be pre-allocated by the caller and res, a and b must all have
// the same size.
// res == a or res == b or both is valid assignment.
func Sub(res, a, b []fext.Element) {

	if len(res)+len(a)+len(b) == 0 {
		return
	}

	r := Vector(res)
	r.Sub(Vector(a), Vector(b))
}

// ZeroPad pads a vector to a given length.
// If the newLen is smaller than len(v), the function panic. It pads to the
// right (appending, not prepending)
// The resulting slice is allocated by the function, so it can be safely
// modified by the caller after the function returns.
func ZeroPad(v []fext.Element, newLen int) []fext.Element {
	if newLen < len(v) {
		utils.Panic("newLen (%v) < len(v) (%v)", newLen, len(v))
	}
	res := make([]fext.Element, newLen)
	copy(res, v)
	return res
}

// Interleave interleave two vectors:
//
//	(a, a, a, a), (b, b, b, b) -> (a, b, a, b, a, b, a, b)
//
// The vecs[i] vectors must all have the same length
func Interleave(vecs ...[]fext.Element) []fext.Element {
	numVecs := len(vecs)
	vecSize := len(vecs[0])

	// all vectors must have the same length
	for i := range vecs {
		if len(vecs[i]) != vecSize {
			utils.Panic("length mismatch, %v != %v", len(vecs[i]), vecSize)
		}
	}

	res := make([]fext.Element, numVecs*vecSize)
	for i := 0; i < vecSize; i++ {
		for j := 0; j < numVecs; j++ {
			res[i*numVecs+j] = vecs[j][i]
		}
	}

	return res
}

// Fill a vector `vec` in place with the given value `val`.
func Fill(v []fext.Element, val fext.Element) {
	for i := range v {
		v[i] = val
	}
}

// PowerVec allocates and returns a vector of size n consisting of consecutive
// powers of x, starting from x^0 = 1 and ending on x^{n-1}. The function panics
// if given x=0 and returns an empty vector if n=0.
func PowerVec(x fext.Element, n int) []fext.Element {
	if x == fext.Zero() {
		utils.Panic("cannot build a power vec for x=0")
	}

	if n == 0 {
		return []fext.Element{}
	}

	res := make([]fext.Element, n)
	res[0].SetOne()

	for i := 1; i < n; i++ {
		res[i].Mul(&res[i-1], &x)
	}

	return res
}

// IntoGnarkAssignment converts an array of field.Element into an array of
// frontend.Variable that can be used to assign a vector of frontend.Variable
// in a circuit or to generate a vector of constant in the circuit definition.
func IntoGnarkAssignment(msgData []fext.Element) []frontend.Variable {
	assignedMsg := []frontend.Variable{}
	for _, x := range msgData {
		assignedMsg = append(assignedMsg, frontend.Variable(x))
	}
	return assignedMsg
}

// Equal compares a and b and returns a boolean indicating whether they contain
// the same value. The function assumes that a and b have the same length. It
// panics otherwise.
func Equal(a, b []fext.Element) bool {

	if len(a) != len(b) {
		utils.Panic("a and b don't have the same length: %v %v", len(a), len(b))
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// PseudoRand generates a vector of field element with a given size using the
// provided random number generator
func PseudoRand(rng *rand.Rand, size int) []fext.Element {
	slice := make([]fext.Element, size)
	for i := range slice {
		slice[i] = fext.PseudoRand(rng)
	}
	return slice
}
