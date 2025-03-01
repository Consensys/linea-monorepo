package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/bits"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/consensys/gnark/frontend"
	"golang.org/x/exp/constraints"
)

func IsPowerOfTwo[T ~int](n T) bool {
	return n&(n-1) == 0 && n > 0
}

func Abs(a int) int {
	mask := a >> (strconv.IntSize - 1)
	return (a ^ mask) - mask
}

func DivCeil(a, b int) int {
	res := a / b
	if b*res < a {
		return res + 1
	}
	return res
}

func DivExact(a, b int) int {
	if b == 0 {
		Panic("division by zero")
	}
	res := a / b
	if res*b != a {
		Panic("inexact division %d/%d", a, b)
	}
	return res
}

func AllReturnEqual[T, U any](fs func(T) U, args []T) (U, error) {
	if len(args) < 1 {
		Panic("Empty list of slice")
	}

	first := fs(args[0])

	for _, arg := range args[1:] {
		if curr := fs(arg); !reflect.DeepEqual(first, curr) {
			return first, fmt.Errorf("mismatch between %v %v, got %v != %v",
				args[0], arg, first, curr,
			)
		}
	}
	return first, nil
}

func NextPowerOfTwo[T ~int64 | ~uint64 | ~uintptr | ~int | ~uint](in T) T {
	if in < 0 || uint64(in) > 1<<62 {
		panic("input out of range")
	}
	if in == 0 {
		return 1
	}
	return 1 << (bits.Len64(uint64(in-1)))
}

func PositiveMod[T ~int](a, n T) T {
	res := a % n
	if res < 0 {
		return res + n
	}
	return res
}

func Join[T any](ts ...[]T) []T {
	total := 0
	for _, t := range ts {
		total += len(t)
	}
	res := make([]T, 0, total)
	for _, t := range ts {
		res = append(res, t...)
	}
	return res
}

func Log2Floor(a int) int {
	return bits.Len(uint(a)) - 1
}

func Log2Ceil(a int) int {
	floor := Log2Floor(a)
	if a != 1<<floor {
		return floor + 1
	}
	return floor
}

func GCD[T ~int](a, b T) T {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func Sha2SumHexOf(w io.WriterTo) string {
	hasher := sha256.New()
	if _, err := w.WriteTo(hasher); err != nil {
		Panic("hash write error: %v", err)
	}
	return HexEncodeToString(hasher.Sum(nil))
}

func Digest(src io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", fmt.Errorf("copy into hasher: %w", err)
	}
	return "0x" + hex.EncodeToString(h.Sum(nil)), nil
}

func RightPadWith[T any](s []T, n int, padWith T) []T {
	if len(s) > n {
		panic("input slice longer than desired padded length")
	}
	res := make([]T, n)
	copy(res, s)
	for i := len(s); i < n; i++ {
		res[i] = padWith
	}
	return res
}

func RightPad[T any](s []T, n int) []T {
	var padWith T
	return RightPadWith(s, n, padWith)
}

func RepeatSlice[T any](s []T, n int) []T {
	res := make([]T, n*len(s))
	for i := 0; i < n; i++ {
		copy(res[i*len(s):], s)
	}
	return res
}

func BigsToBytes(ins []*big.Int) []byte {
	res := make([]byte, len(ins))
	for i := range ins {
		if !ins[i].IsUint64() || ins[i].Uint64() > 0xFF {
			panic("value exceeds byte size")
		}
		res[i] = byte(ins[i].Uint64())
	}
	return res
}

func BigsToInts(ints []*big.Int) []int {
	res := make([]int, len(ints))
	for i := range ints {
		u := ints[i].Uint64()
		if !ints[i].IsUint64() || u > math.MaxInt {
			panic("overflow")
		}
		res[i] = int(u)
	}
	return res
}

func ToInt[T ~uint | ~uint64 | ~int64](i T) int {
	if i > math.MaxInt {
		panic("overflow")
	}
	return int(i)
}

func ToUint64[T constraints.Signed](i T) uint64 {
	if i < 0 {
		panic("negative")
	}
	return uint64(i)
}

func ToUint16[T ~int | ~uint](i T) uint16 {
	if i < 0 || i > math.MaxUint16 {
		panic("out of range")
	}
	return uint16(i)
}

func ToVariableSlice[X any](s []X) []frontend.Variable {
	res := make([]frontend.Variable, len(s))
	for i := range s {
		res[i] = s[i]
	}
	return res
}

func countInts[I constraints.Integer](s []I) []I {
	maxVal := Max(s...)
	counts := make([]I, maxVal+1)
	for _, x := range s {
		counts[x]++
	}
	return counts
}

func Partition[T any, I constraints.Integer](s []T, index []I) [][]T {
	if len(s) != len(index) {
		panic("s and index must have the same length")
	}
	if len(s) == 0 {
		return nil
	}
	
	maxIdx := Max(index...)
	partitions := make([][]T, maxIdx+1)
	counts := countInts(index)
	
	for i := range partitions {
		partitions[i] = make([]T, 0, counts[i])
	}
	
	for i, v := range s {
		partitions[index[i]] = append(partitions[index[i]], v)
	}
	return partitions
}

func Ite[T any](cond bool, ifSo, ifNot T) T {
	if cond {
		return ifSo
	}
	return ifNot
}

func RangeSlice[T constraints.Integer](length int, startingPoints ...T) []T {
	if len(startingPoints) == 0 {
		startingPoints = []T{0}
	}
	res := make([]T, length*len(startingPoints))
	for i, start := range startingPoints {
		base := i * length
		for j := 0; j < length; j++ {
			res[base+j] = start + T(j)
		}
	}
	return res
}

func ReadFromJSON(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func WriteToJSON(path string, v interface{}) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(v)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = fmt.Errorf("close error: %w", closeErr)
	}
	return err
}

func WriterstoEqual(expected, actual io.WriterTo) error {
	var bb bytes.Buffer
	if _, err := expected.WriteTo(&bb); err != nil {
		return err
	}
	ab := bb.Bytes()
	bb.Reset()
	if _, err := actual.WriteTo(&bb); err != nil {
		return err
	}
	return BytesEqual(ab, bb.Bytes())
}

func BytesEqual(expected, actual []byte) error {
	if bytes.Equal(expected, actual) {
		return nil
	}

	l := min(len(expected), len(actual))
	failure := 0
	
	for failure < l && expected[failure] == actual[failure] {
		failure++
	}

	var sb strings.Builder
	const radius = 40

	sb.WriteString(fmt.Sprintf("mismatch starting at byte %d\n", failure))

	printHex := func(b []byte) {
		for i := max(failure-radius, 0); i < min(failure+radius, len(b)); i++ {
			fmt.Fprintf(&sb, "%02x", b[i])
		}
	}

	sb.WriteString("expected: ")
	printHex(expected)
	sb.WriteString("\nactual:   ")
	printHex(actual)
	sb.WriteString("\n")

	return &BytesEqualError{
		Index: failure,
		error: sb.String(),
	}
}

type BytesEqualError struct {
	Index int
	error string
}

func (e *BytesEqualError) Error() string {
	return e.error
}

func ReadFromFile(path string, to io.ReaderFrom) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = to.ReadFrom(f)
	return errors.Join(err, f.Close())
}

func WriteToFile(path string, from io.WriterTo) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = from.WriteTo(f)
	return errors.Join(err, f.Close())
}
