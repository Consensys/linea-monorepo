package vector

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/gnark/frontend"
)

/*
	Contains a set of utility functions
	relative to vector/polynomial operations
*/

// Deep-copies the input vector
func DeepCopy(pol []field.Element) []field.Element {
	return append([]field.Element{}, pol...)
}

// Multiply a vector by a scalar - in place. The result should be
// preallocated or it is going to panic.
func ScalarMul(res, vec []field.Element, scalar field.Element) {
	if len(res) != len(vec) {
		utils.Panic("The inputs should have the same length %v %v", len(res), len(vec))
	}

	for i := range vec {
		res[i].Mul(&vec[i], &scalar)
	}
}

// Returns the scalar (inner) product of two vectors
func ScalarProd(a, b []field.Element) field.Element {

	if len(b) != len(a) {
		utils.Panic("The inputs should have the same length %v %v", len(a), len(b))
	}

	var res, tmp field.Element
	for i := range a {
		tmp.Mul(&a[i], &b[i])
		res.Add(&res, &tmp)
	}
	return res
}

// Returns the scalar (inner) product of two vectors
func ParScalarProd(a, b []field.Element, numcpus int) field.Element {

	if numcpus == 0 {
		numcpus = runtime.GOMAXPROCS(0)
	}

	if len(b) != len(a) {
		utils.Panic("The inputs should have the same length %v %v", len(a), len(b))
	}

	var res field.Element
	lock := sync.Mutex{}

	parallel.Execute(len(a), func(start, stop int) {
		subRes := ScalarProd(a[start:stop], b[start:stop])
		lock.Lock()
		res.Add(&res, &subRes)
		lock.Unlock()
	}, numcpus)

	return res
}

// Applies the butterfly element-wise
func Butterfly(a, b []field.Element) {
	for i := range a {
		field.Butterfly(&a[i], &b[i])
	}
}

// Creates a random vector of size n
func Rand(n int) []field.Element {
	vec := make([]field.Element, n)
	for i := range vec {
		_, err := vec[i].SetRandom()
		// Just to enfore never having to deal with zeroes
		if err != nil {
			panic(err)
		}
	}
	return vec
}

// Splits a vector in smaller chunks
func SplitExact(vec []field.Element, numChunks int) [][]field.Element {

	// Sanity-check, it should divide
	if len(vec)%numChunks != 0 {
		utils.Panic("Can't split #%v in %v equal chunks", len(vec), numChunks)
	}

	chunkSize := len(vec) / numChunks

	res := make([][]field.Element, numChunks)
	for i := range res {
		res[i] = vec[i*chunkSize : (i+1)*chunkSize]
	}

	return res
}

// Equal returns true iff the two vectors are equals
// Panic on two unequal length vectors
func Equal(a, b []field.Element) bool {

	if len(a) != len(b) {
		utils.Panic("a and b do not have the same length %v != %v", len(a), len(b))
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Multiply two vectors element wise and write the result in res
func MulElementWise(res, a, b []field.Element) {

	if len(res) != len(a) || len(b) != len(a) {
		utils.Panic("The inputs should have the same length %v %v %v", len(res), len(a), len(b))
	}

	for i := range a {
		res[i].Mul(&a[i], &b[i])
	}
}

// Pretty print a vector
func Prettify(a []field.Element) string {
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
func Reverse(v []field.Element) {
	n := len(v) - 1
	for i := 0; i < len(v)/2; i++ {
		v[i], v[n-i] = v[n-i], v[i]
	}
}

// Returns a vector of repeated values
func Repeat(x field.Element, n int) []field.Element {
	res := make([]field.Element, n)
	for i := range res {
		res[i] = x
	}
	return res
}

// For test returns a vector instantiated from a list of integers*
func ForTest(xs ...int) []field.Element {
	res := make([]field.Element, len(xs))
	for i, x := range xs {
		res[i].SetInt64(int64(x))
	}
	return res
}

// Returns an explanatory error if the vectors don't have the same length
func SameLength(u, v []field.Element) error {
	if len(u) != len(v) {
		return fmt.Errorf("the two vectors do not have the same length %v %v", len(u), len(v))
	}
	return nil
}

// Add two vectors `a` and `b` and put the result in `res`
// `res` must be pre-allocated
func Add(res, a, b []field.Element) {

	if len(res) != len(a) || len(b) != len(a) {
		utils.Panic("The inputs should have the same length %v %v %v", len(res), len(a), len(b))
	}

	for i := range res {
		res[i].Add(&a[i], &b[i])
	}
}

// Sub two vectors `a` and `b` and put the result in `res`
// `res` must be pre-allocated
func Sub(res, a, b []field.Element) {

	if len(res) != len(a) || len(b) != len(a) {
		utils.Panic("The inputs should have the same length %v %v %v", len(res), len(a), len(b))
	}

	for i := range res {
		res[i].Sub(&a[i], &b[i])
	}
}

// Sub all elements of a vector `a` by a field element`b` and put the result in `res`
// `res` must be pre-allocated
func SubScalarRight(res, a []field.Element, b field.Element) {
	for i := range res {
		res[i].Sub(&a[i], &b)
	}
}

// Returns the result of substracting a field element `a` by all elements of a vector `b`
// separately and put the result in `res`
// `res` must be pre-allocated
func SubScalarLeft(res []field.Element, a field.Element, b []field.Element) {
	for i := range res {
		res[i].Sub(&a, &b[i])
	}
}

// Zero pad a vector to a given length. Copies the original slice.
// If the newLen is smaller, panic. It pads to the right (appending, not prepending)
func ZeroPad(v []field.Element, newLen int) []field.Element {
	if newLen < len(v) {
		utils.Panic("newLen (%v) < len(v) (%v)", newLen, len(v))
	}
	res := make([]field.Element, newLen)
	copy(res, v)
	return res
}

// Clear a vector v in-place
func Clear(v []field.Element) {
	for i := range v {
		v[i].SetZero()
	}
}

// Fill a vector in place with the given value
func Fill(v []field.Element, val field.Element) {
	for i := range v {
		v[i] = val
	}
}

// FromStrings returns a vector from a list of strings
func FromStrings(strs ...string) []field.Element {
	res := make([]field.Element, len(strs))
	for i, s := range strs {
		res[i].SetString(s)
	}
	return res
}

func Marshal(vec []field.Element) []byte {
	// Marshal the elements in a vector of bytes
	bytes := make([]byte, 0, len(vec)*field.Bytes)
	for i := range vec {
		bytes = append(bytes, vec[i].Marshal()...)
	}
	return bytes
}

// Panics if the marshalled vector does not have exactly the
// expected size
func Unmarshal(vec []byte, howmany int) []field.Element {

	canDeserialize := len(vec) / field.Bytes
	if field.Bytes*canDeserialize < len(vec) {
		utils.Panic("#bytes (%v) does not divide the byte size of a field element (%v)", len(vec), field.Bytes)
	}

	if canDeserialize != howmany {
		utils.Panic("can deserialize %v fields, expected %v", canDeserialize, howmany)
	}

	res := make([]field.Element, howmany)
	for i := range res {
		res[i].SetBytes(vec[:field.Bytes])
		vec = vec[field.Bytes:]
	}

	if len(vec) > 0 {
		panic("should have drained the full vec. The current function is wrong")
	}

	return res
}

// Prepends a vector to the desired length
func ZeroPrependToSize(vec []field.Element, newSize int) []field.Element {
	if len(vec) > newSize {
		utils.Panic("newSize is smaller then the current size %v < %v", newSize, len(vec))
	}

	if len(vec) == newSize {
		return append([]field.Element{}, vec...)
	}

	res := make([]field.Element, newSize)
	copy(res[newSize-len(vec):], vec)
	return res
}

// Preprends with a given value to desired length
func PreprendToSize(vec []field.Element, newSize int, paddingVal field.Element) []field.Element {
	if len(vec) > newSize {
		utils.Panic("newSize is smaller then the current size %v < %v", newSize, len(vec))
	}

	if len(vec) == newSize {
		return append([]field.Element{}, vec...)
	}

	res := Repeat(paddingVal, newSize-len(vec))
	return append(res, vec...)
}

// Exponentiate a vector element-wise
func Exp(res, base []field.Element, exp int) {

	if len(res) != len(base) {
		utils.Panic("The inputs should have the same length %v %v", len(res), len(base))
	}

	if exp < 0 {
		panic("Unsupported")
	}

	if exp == 0 {
		for i := range res {
			res[i].SetOne()
		}
		return
	}

	if exp == 1 {
		copy(res, base)
		return
	}

	bigExp := big.NewInt(int64(exp))
	for i := range res {
		res[i].Exp(base[i], bigExp)
	}
}

/*
Given x and "n", returns a vector of n element (1, x, x^2, x^3, ..., x^{n-1})
*/
func PowerVec(x field.Element, n int) []field.Element {
	res := make([]field.Element, n)
	res[0].SetOne()
	for i := 1; i < n; i++ {
		res[i].Mul(&res[i-1], &x)
	}
	return res
}

/*
Converts an array of field.Element into an array of assignment
*/
func IntoGnarkAssignment(msgData []field.Element) []frontend.Variable {
	assignedMsg := []frontend.Variable{}
	for _, x := range msgData {
		assignedMsg = append(assignedMsg, frontend.Variable(x))
	}
	return assignedMsg
}

func ExplicitDeclarationStr(v []field.Element) string {
	entriesDecl := make([]string, 0, len(v))
	for _, w := range v {
		entriesDecl = append(entriesDecl, field.ExplicitDeclarationStr(w))
	}
	return fmt.Sprintf("[]field.Element{%v}", strings.Join(entriesDecl, ", "))
}
