package test_utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	if len(p) >= 65535 {
		panic("ReaderHash.Write: too large")
	}

	var ls [2]byte
	if _, err = r.r.Read(ls[:]); err != nil {
		panic(err)
	}

	buf := make([]byte, int(ls[0])*256+int(ls[1]))
	if _, err = r.r.Read(buf); err != nil {
		panic(err)
	}
	if !bytes.Equal(buf, p) {
		panic(fmt.Errorf("ReaderHash.Write: mismatch %x≠%x", buf, p))
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
	var ls [2]byte
	if _, err := r.r.Read(ls[:]); err != nil {
		panic(err)
	}
	if ls[0] != 255 || ls[1] != 255 {
		panic(fmt.Errorf("ReaderHash.Reset: unexpected %x", ls))
	}
	r.h.Reset()
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
