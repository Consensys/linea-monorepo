package datatransfer

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// deriveName derives column names.
func deriveName(mainName string, ids ...int) ifaces.ColID {
	idStr := []string{}
	for i := range ids {
		idStr = append(idStr, strconv.Itoa(ids[i]))
	}
	return ifaces.ColIDf("%v_%v_%v", "DT", mainName, strings.Join(idStr, "_"))
}

// getZeroOnes receives n  and outputs the pattern  (0,..0,1,..,1) such that there are n elements 1.
func getZeroOnes(n field.Element, max int) (a []field.Element) {
	if n.Uint64() > uint64(max) {
		utils.Panic("%v should be smaller than %v", n.Uint64(), max)
	}
	for j := 0; j < max-int(n.Uint64()); j++ {
		a = append(a, field.Zero())

	}
	for i := max - int(n.Uint64()); i < max; i++ {
		a = append(a, field.One())

	}

	return a

}

func bytesAsBlockPtrUnsafe(s []byte) *keccak.Block {
	return (*keccak.Block)(unsafe.Pointer(&s[0]))
}

func baseRecomposeHandles(a []ifaces.Column, base any) *symbolic.Expression {
	res := symbolic.NewConstant(0)
	for k := len(a) - 1; k >= 0; k-- {
		res = symbolic.Add(symbolic.Mul(res, base), a[k])
	}
	return res
}

// It recompose the slices with the given length in a little-endian order
func baseRecomposeByLengthHandles(slices []ifaces.Column, base any, lenSlices []ifaces.Column) *symbolic.Expression {
	res := symbolic.NewConstant(0)
	for k := 0; k < len(slices); k++ {
		res = symbolic.Add(symbolic.Mul(res, base), symbolic.Mul(slices[k], lenSlices[k]))
	}
	return res
}

// it recompose the slices with the given length in big-endian, when the lengths of the slices are binary.
func baseRecomposeBinaryLen(slices []ifaces.Column, base any, lenSlices []ifaces.Column) *symbolic.Expression {
	res := symbolic.NewConstant(0)
	for k := len(slices) - 1; k >= 0; k-- {
		currBase := symbolic.Add(symbolic.Mul(base, lenSlices[k]), symbolic.Sub(1, lenSlices[k]))
		res = symbolic.Add(symbolic.Mul(res, currBase), slices[k])
	}
	return res
}

// Converts a 16bits fieldElement to a given base, the base should be given in field element
// form to save on expensive conversion.
func uInt16ToBaseX(x uint16, base *field.Element) field.Element {
	res := field.Zero()
	one := field.One()
	resIsZero := true

	for k := 16; k >= 0; k-- {
		// The test allows skipping useless field muls or testing
		// the entire field element.
		if !resIsZero {
			res.Mul(&res, base)
		}

		// Skips the field addition if the bit is zero
		bit := (x >> k) & 1
		if bit > 0 {
			res.Add(&res, &one)
			resIsZero = false
		}
	}

	return res
}

// It receives a set of numbers and cut/stitch them to the given length.
// Example; if the given length is 8 and the number of chunk is 3;
// then (15,3,6) is organized as (8,7,1,2,6),
// 15 --> 8,7, 3-->1,2 (because stitching 7 and 1 gives 8)
func cutAndStitch(nByte []field.Element, nbChunk, lenChunk int) (b [][]field.Element) {
	missing := uint64(lenChunk)
	b = make([][]field.Element, nbChunk)
	for i := range nByte {
		var a []field.Element
		curr := nByte[i].Uint64()
		for curr != 0 {
			if curr >= missing {
				a = append(a, field.NewElement(missing))
				curr = curr - missing
				missing = uint64(lenChunk)
			} else {
				a = append(a, field.NewElement(curr))
				missing = missing - curr
				curr = 0
			}
		}
		// message to hash is based on big endian, LCD should be big endian order.
		// Thus add the zeros at the beginning, if we have less than nbChunk.
		for len(a) < nbChunk {
			a = append([]field.Element{field.Zero()}, a...)
		}

		for j := 0; j < nbChunk; j++ {
			b[j] = append(b[j], a[j])
		}

	}
	return b
}

// It receives the length of the slices and decompose the element to the slices with the given lengths.
// decomposition is in little Endian. Zeroes are added at the beginning if we have less than three slices.
func decomposeByLength(a field.Element, lenA int, givenLen []int) (slices []field.Element) {

	//sanity check
	s := 0
	for i := range givenLen {
		s = s + givenLen[i]
	}
	if s != lenA {
		utils.Panic("input can not be decomposed to the given lengths")
	}

	b := a.Bytes()
	bytes := b[32-lenA:]
	slices = make([]field.Element, len(givenLen))
	for i := range givenLen {
		if givenLen[i] == 0 {
			slices[i] = field.Zero()
		} else {
			b := bytes[:givenLen[i]]
			x := 0
			s := uint64(0)
			for j := 0; j < givenLen[i]; j++ {
				s = s | uint64(b[j])<<x
				x = x + 8
			}
			slices[i] = field.NewElement(s)
			bytes = bytes[givenLen[i]:]
		}
	}

	return slices

}

// It receives a column and indicates where the accumulation reaches the target value
// It panics if at any point the accumulation goes beyond the target value.
func AccReachedTargetValue(column []field.Element, targetVal int) (reachedTarget []field.Element) {
	s := 0
	for j := range column {
		s = s + int(column[j].Uint64())
		if s > targetVal {
			utils.Panic("Should not reach a value larger than target value")
		}
		if s == targetVal {
			reachedTarget = append(reachedTarget, field.One())
			s = 0
		} else {
			reachedTarget = append(reachedTarget, field.Zero())
		}

	}
	return reachedTarget
}

// It receives multiple matrices and a filter, it returns the spaghetti form of the matrices
func makeSpaghetti(filter [][]field.Element, matrix ...[][]field.Element) (spaghetti [][]field.Element) {
	spaghetti = make([][]field.Element, len(matrix))

	// populate spaghetties
	for i := range filter[0] {
		for j := range filter {
			if filter[j][i].Uint64() == 1 {
				for k := range matrix {
					spaghetti[k] = append(spaghetti[k], matrix[k][j][i])
				}
			}

		}
	}
	return spaghetti
}

// It receives a vector and a filter and fold it to a  matrix.
func makeMatrix(filter [][]field.Element, myVector []field.Element) (matrix [][]field.Element) {

	matrix = make([][]field.Element, len(filter))
	for j := range matrix {
		matrix[j] = make([]field.Element, len(filter[0]))
	}
	// populate matrix
	k := 0
	for i := range filter[0] {
		for j := range filter {
			if filter[j][i].Uint64() == 1 && k < len(myVector) {
				matrix[j][i] = myVector[k]
				k++
			}

		}
	}
	return matrix
}

// It receives the length of the slices and decompose the element to the slices with the given lengths.
// decomposition is in little Endian. Zeroes are added at the beginning if we have less than three slices.
func decomposeByLengthFr(a field.Element, lenA int, givenLen []field.Element) (slices []field.Element) {

	//sanity check
	s := 0
	for i := range givenLen {
		s = s + int(givenLen[i].Uint64())
	}
	if s != lenA {
		utils.Panic("input can not be decomposed to the given lengths")
	}

	b := a.Bytes()
	bytes := b[32-lenA:]
	slices = make([]field.Element, len(givenLen))
	for i := range givenLen {
		if givenLen[i] == field.Zero() {
			slices[i] = field.Zero()
		} else {
			b := bytes[:givenLen[i].Uint64()]
			x := 0
			s := uint64(0)
			for j := 0; j < int(givenLen[i].Uint64()); j++ {
				s = s | uint64(b[j])<<x
				x = x + 8
			}
			slices[i] = field.NewElement(s)
			bytes = bytes[givenLen[i].Uint64():]
		}
	}

	return slices

}

// It converts a slices of uint4 from  Bing-endian to little-endian.
// Since BE to LE is over Bytes;
//
//	for uint4 two adjacent elements keep their order
//
// e.g., (a0,a1,a2,a3,....,a14,a15)
// is converted to (a14,a15, ....,a2,a3,a0,a1)
func SlicesBeToLeUint4(s []field.Element) []field.Element {
	i := 0
	var a []field.Element
	for i < len(s) {
		a = append(a, s[len(s)-1-i-1])
		a = append(a, s[len(s)-1-i])
		i = i + 2
	}
	return a
}

// It converts a slices of uint4 from  Bing-endian to little-endian.
func SlicesBeToLeHandle(s []ifaces.Column) []ifaces.Column {
	i := 0
	var a []ifaces.Column
	for i < len(s) {
		a = append(a, s[len(s)-1-i-1])
		a = append(a, s[len(s)-1-i])
		i = i + 2
	}
	return a
}
