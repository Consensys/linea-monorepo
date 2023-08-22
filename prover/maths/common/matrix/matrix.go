package matrix

import (
	"fmt"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Deep-copies a double-slices
func DeepCopy(vec [][]field.Element) [][]field.Element {
	res := make([][]field.Element, len(vec))
	for i := range res {
		res[i] = vector.DeepCopy(vec[i])
	}
	return res
}

// Returns a matrix full of zeroes
func Zeroes(nSlices, sliceLen int) [][]field.Element {
	res := make([][]field.Element, nSlices)
	for sliceID := range res {
		res[sliceID] = make([]field.Element, sliceLen)
	}
	return res
}

// Rand returns a random matrix with given size
func Rand(numSlices, sliceLen int) [][]field.Element {
	res := make([][]field.Element, numSlices)
	for i := range res {
		res[i] = vector.Rand(sliceLen)
	}
	return res
}

// Transpose the matrix
func Transpose(m [][]field.Element) [][]field.Element {
	srcNumSlices := len(m)
	srcSliceLen := len(m[0])

	res := make([][]field.Element, srcSliceLen)
	for i := range res {
		res[i] = make([]field.Element, srcNumSlices)
		for j := range res[i] {
			res[i][j] = m[j][i]
		}
	}
	return res
}

// Returns true if the two matrices have the same dimensions
// Will also check if the two matrices are irregular
// An error is return if they are not
func SameDim(m, p [][]field.Element) error {

	if err := AssertRegular(m); err != nil {
		return fmt.Errorf("argument `m` is improper : %v", err)
	}

	if err := AssertRegular(p); err != nil {
		return fmt.Errorf("argument `p` is improper : %v", err)
	}

	if len(m) != len(p) {
		return fmt.Errorf("`p` and `m` don't have the same number of subslices %v %v", len(m), len(p))
	}

	if len(m[0]) != len(p[0]) {
		return fmt.Errorf("`p` and `m` 's subslices don't have the same length %v %v", len(m[0]), len(p[0]))
	}

	return nil
}

// Returns an error if the matrix as irregular dimensions
// Usefull for testing entries
func AssertRegular(m [][]field.Element) error {
	len_ := len(m[0])
	for i := range m {
		if len(m[i]) != len_ {
			return fmt.Errorf("the matrix has improper dimensions l0 = %v while l%v = %v", len_, i, len(m[i]))
		}
	}
	return nil
}

// Multiply all columns, each, by a different scalar
func ScalarMulSubslices(res, m [][]field.Element, scalars []field.Element) {

	// Sanity-checks on the dimensions
	if err := SameDim(res, m); err != nil {
		panic(err)
	}

	for i := range m {
		vector.ScalarMul(res[i], m[i], scalars[i])
	}
}

// Returns a matrix consisting of manytimes the same subslice
// The subslice is deep-copied
func RepeatSubslice(subslice []field.Element, n int) [][]field.Element {
	res := make([][]field.Element, n)
	for i := range res {
		res[i] = vector.DeepCopy(subslice)
	}
	return res
}

// Prettify a matrix into a string
func Prettify(m [][]field.Element) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i := range m {
		// skip the first entry to not have [,x ,...]
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(vector.Prettify(m[i]))
	}
	sb.WriteString("]")
	return sb.String()
}
