// vector offers a set of utility function relating to slices of field element
// and that are commonly used as part of the repo.
package vectorext

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
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

func ScalarProdByElement(a []fext.Element, b []field.Element) fext.Element {
	// The length checks is done by gnark-crypto already
	a_ := Vector(a)
	res := a_.InnerProductByElement(b)
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
		res[i].B0.A0.SetInt64(int64(x))
	}
	return res
}

// ForTestFromVect computes a vector of field extensions,
// where each field extension is populated using one vector of size [fext.ExtensionDegree]
func ForRandTestFromLen(len int) []fext.Element {
	res := make([]fext.Element, len)
	for i := range res {
		res[i].SetRandom()
	}
	return res
}

// ForTestFromVect computes a vector of field extensions,
// where each field extension is populated using one vector of size [fext.ExtensionDegree]
func ForTestFromVect(xs ...[4]int) []fext.Element {
	res := make([]fext.Element, len(xs))
	for i, x := range xs {
		res[i].B0.A0.SetInt64(int64(x[0]))
		res[i].B0.A1.SetInt64(int64(x[1]))
		res[i].B1.A0.SetInt64(int64(x[2]))
		res[i].B1.A1.SetInt64(int64(x[3]))
	}
	return res
}

// ForTestFromQuads groups the input into Quaternarys. Each Quaternary populates the first four
// coordinates of a field extension, and the function then
// returns a vector instantiated these field extension elements.
func ForTestFromQuads(xs ...int) []fext.Element {
	if len(xs)%4 != 0 {
		panic("ForTestFromQuads must receive a 4n-length input vector")
	}
	res := make([]fext.Element, len(xs)/4)
	for i := 0; i < len(res); i++ {
		res[i] = fext.NewFromInt(int64(xs[4*i]), int64(xs[4*i+1]), int64(xs[4*i+2]), int64(xs[4*i+3]))
	}
	return res
}

// ForTestCalculateQuadProduct performs the specific 4-element product calculation.
// It takes two 4-element slices (a_slice, b_slice).
// It returns a slice containing the four calculated results.
// This function is used for generating test results
func ForTestCalculateQuadProduct(a_slice []int, b_slice []int) []int {
	// Ensure input slices have the correct length to prevent out-of-bounds errors.
	// In a real application, you might want more robust error handling or panics.
	if len(a_slice) < 4 || len(b_slice) < 4 {
		fmt.Println("Error: Input slices must have at least 4 elements.")
		return nil
	}

	// Extract elements for clarity (optional, could use a_slice[0], etc. directly)
	a0, a1, a2, a3 := a_slice[0], a_slice[1], a_slice[2], a_slice[3]
	b0, b1, b2, b3 := b_slice[0], b_slice[1], b_slice[2], b_slice[3]

	// Calculate the four results based on the provided pattern
	result0 := a0*b0 + a1*b1*fext.RootPowers[1] + a2*b3*fext.RootPowers[1] + a3*b2*fext.RootPowers[1]
	result1 := a0*b1 + a1*b0 + a2*b2 + a3*b3*fext.RootPowers[1]
	result2 := a0*b2 + a1*b3*fext.RootPowers[1] + a2*b0 + a3*b1*fext.RootPowers[1]
	result3 := a0*b3 + a1*b2 + a2*b1 + a3*b0

	return []int{result0, result1, result2, result3}
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
	if x.IsZero() {
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
func IntoGnarkAssignment(msgData []fext.Element) []gnarkfext.Element {
	assignedMsg := []gnarkfext.Element{}
	for _, x := range msgData {
		assignedMsg = append(assignedMsg, gnarkfext.FromValue(x))
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
