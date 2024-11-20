package test_utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark/frontend"
	snarkHash "github.com/consensys/gnark/std/hash"
	"hash"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakeTestingT struct{}

func (FakeTestingT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format+"\n", args...))
}

func (FakeTestingT) FailNow() {
	os.Exit(-1)
}

func RandIntN(n int) int {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	if n > math.MaxInt {
		panic("RandIntN: n too large")
	}
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n)) // #nosec G115 -- Above check precludes an overflow
}

func RandIntSliceN(length, n int) []int {
	res := make([]int, length)
	for i := range res {
		res[i] = RandIntN(n)
	}
	return res
}

type BytesEqualError struct {
	Index int
	error string
}

func (e *BytesEqualError) Error() string {
	return e.error
}

func LoadJson(t *testing.T, path string, v any) {
	in, err := os.Open(path)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(in).Decode(v))
}

// BytesEqual between byte slices a,b
// a readable error message would show in case of inequality
// TODO error options: block size, check forwards or backwards etc
func BytesEqual(expected, actual []byte) error {
	l := min(len(expected), len(actual))

	failure := 0
	for failure < l {
		if expected[failure] != actual[failure] {
			break
		}
		failure++
	}

	if len(expected) == len(actual) {
		return nil
	}

	// there is a mismatch
	var sb strings.Builder

	const (
		radius    = 40
		blockSize = 32
	)

	printCentered := func(b []byte) {

		for i := max(failure-radius, 0); i <= failure+radius; i++ {
			if i%blockSize == 0 && i != failure-radius {
				sb.WriteString("  ")
			}
			if i >= 0 && i < len(b) {
				sb.WriteString(hex.EncodeToString([]byte{b[i]})) // inefficient, but this whole error printing sub-procedure will not be run more than once
			} else {
				sb.WriteString("  ")
			}
		}
	}

	sb.WriteString(fmt.Sprintf("mismatch starting at byte %d\n", failure))

	sb.WriteString("expected: ")
	printCentered(expected)
	sb.WriteString("\n")

	sb.WriteString("actual:   ")
	printCentered(actual)
	sb.WriteString("\n")

	sb.WriteString("          ")
	for i := max(failure-radius, 0); i <= failure+radius; {
		if i%blockSize == 0 && i != failure-radius {
			s := strconv.Itoa(i)
			sb.WriteString("  ")
			sb.WriteString(s)
			i += len(s) / 2
			if len(s)%2 != 0 {
				sb.WriteString(" ")
				i++
			}
		} else {
			if i == failure {
				sb.WriteString("^^")
			} else {
				sb.WriteString("  ")
			}
			i++
		}
	}

	sb.WriteString("\n")

	return &BytesEqualError{
		Index: failure,
		error: sb.String(),
	}
}

func SlicesEqual[T any](expected, actual []T) error {
	if l1, l2 := len(expected), len(actual); l1 != l2 {
		return fmt.Errorf("length mismatch %d≠%d", l1, l2)
	}

	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			return fmt.Errorf("mismatch at #%d:\nexpected %v\nencountered %v", i, expected[i], actual[i])
		}
	}
	return nil
}

// WriterHash is a wrapper around a hash.Hash that duplicates all writes.
type WriterHash struct {
	h hash.Hash
	w io.Writer
}

func (w *WriterHash) Write(p []byte) (n int, err error) {
	if len(p) >= 65535 {
		panic("WriterHash.Write: too large")
	}
	if _, err = w.w.Write([]byte{byte(len(p) / 256), byte(len(p))}); err != nil {
		panic(err)
	}
	if _, err = w.w.Write(p); err != nil {
		panic(err)
	}
	return w.h.Write(p)
}

func (w *WriterHash) Sum(b []byte) []byte {
	if b != nil {
		panic("not supported")
	}
	return w.h.Sum(nil)
}

func (w *WriterHash) Reset() {
	if _, err := w.w.Write([]byte{255, 255}); err != nil {
		panic(err)
	}
	w.h.Reset()
}

func (w *WriterHash) Size() int {
	return w.h.Size()
}

func (w *WriterHash) BlockSize() int {
	return w.h.BlockSize()
}

func (w *WriterHash) CloseFile() {
	if err := w.w.(*os.File).Close(); err != nil {
		panic(err)
	}
}

func NewWriterHashToFile(h hash.Hash, path string) *WriterHash {
	w, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return &WriterHash{
		h: h,
		w: w,
	}
}

// ReaderHash is a wrapper around a hash.Hash that matches all writes with its input stream.
type ReaderHash struct {
	h hash.Hash
	r io.Reader
}

func (r *ReaderHash) Write(p []byte) (n int, err error) {
	if rd := hashReadWrite(r.r); !bytes.Equal(rd, p) {
		panic(fmt.Errorf("ReaderHash.Write: mismatch %x≠%x", rd, p))
	}

	return r.h.Write(p)
}

func (r *ReaderHash) Sum(b []byte) []byte {
	if b != nil {
		panic("not supported")
	}
	return r.h.Sum(nil)
}

func (r *ReaderHash) Reset() {
	hashReadReset(r.r)
	r.h.Reset()
}

func hashReadWrite(r io.Reader) []byte {

	var ls [2]byte
	if _, err := r.Read(ls[:]); err != nil {
		panic(err)
	}

	buf := make([]byte, int(ls[0])*256+int(ls[1]))
	if len(buf) == 65535 {
		panic("ReaderHash.Write: Reset expected")
	}
	if _, err := r.Read(buf); err != nil {
		panic(err)
	}

	return buf
}

func hashReadReset(r io.Reader) {
	var ls [2]byte
	if _, err := r.Read(ls[:]); err != nil {
		panic(err)
	}
	if ls[0] != 255 || ls[1] != 255 {
		panic(fmt.Errorf("ReaderHash.Reset: unexpected %x", ls))
	}
}

func (r *ReaderHash) Size() int {
	return r.h.Size()
}

func (r *ReaderHash) BlockSize() int {
	return r.h.BlockSize()
}

func NewReaderHashFromFile(h hash.Hash, path string) *ReaderHash {
	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return &ReaderHash{
		h: h,
		r: r,
	}
}

func (r *ReaderHash) CloseFile() {
	if err := r.r.(*os.File).Close(); err != nil {
		panic(err)
	}
}

// ReaderHashSnark is a wrapper around a FieldHasher that matches all writes with its input stream.
type ReaderHashSnark struct {
	h   snarkHash.FieldHasher
	r   io.Reader
	api frontend.API
}

func NewReaderHashSnarkFromFile(api frontend.API, h snarkHash.FieldHasher, path string) snarkHash.FieldHasher {
	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return &ReaderHashSnark{
		h:   h,
		r:   r,
		api: api,
	}
}

func (r *ReaderHashSnark) Sum() frontend.Variable {
	return r.h.Sum()
}

func (r *ReaderHashSnark) Write(data ...frontend.Variable) {
	r.h.Write(data...)

	for i := 0; i < len(data); {
		buf := hashReadWrite(r.r)
		for len(buf) != 0 {
			n := min(len(buf), (r.api.Compiler().FieldBitLen()+7)/8)
			r.api.AssertIsEqual(data[i], buf[:n])
			buf = buf[n:]
			i++
		}
	}
}

func (r *ReaderHashSnark) Reset() {
	hashReadReset(r.r)
	r.h.Reset()
}

func (r *ReaderHashSnark) CloseFile() {
	if err := r.r.(*os.File).Close(); err != nil {
		panic(err)
	}
}

func PrettyPrintHashes(filename string) {
	const printIndexes = false

	v := make([]any, 0)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	var (
		length [2]byte
		buf    []byte
	)

	var i int
	for _, err = f.Read(length[:]); err == nil; _, err = f.Read(length[:]) {
		l := int(length[0])*256 + int(length[1])
		if l == 65535 { // a reset
			v = append(v, "RESET")
			i = 0
			continue
		}
		if l > len(buf) {
			buf = make([]byte, l)
		}

		if _, err = f.Read(buf[:l]); err != nil {
			break
		}

		prettyBuf := spaceOutFromRight(hex.EncodeToString(buf[:l]))
		if printIndexes {
			v = append(v, fmt.Sprintf("%d: 0x%s", i, prettyBuf))
		} else {
			v = append(v, "0x"+prettyBuf)
		}

		i++
	}
	if err != io.EOF {
		panic(err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func spaceOutFromRight(s string) string {
	var bb strings.Builder
	n := len(s) + (len(s)+15)/16 - 1
	bb.Grow(n)
	remainder := len(s) % 16
	first := true
	if remainder != 0 {
		bb.WriteString(s[:remainder])
		s = s[remainder:]
		first = false
	}
	for len(s) > 0 {
		if !first {
			bb.WriteByte(' ')
		}
		bb.WriteString(s[:16])
		s = s[16:]
		first = false
	}

	if bb.Len() != n {
		panic("incorrect size estimation")
	}
	return bb.String()
}
