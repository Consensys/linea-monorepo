package test_utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/zkevm"

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

func LoadJson(t *testing.T, path string, v any) {
	in, err := os.Open(path)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(in).Decode(v))
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

// PrettyPrintHashes reads a file of hashes and prints them in a human-readable format.
// if out is nil, it will print to os.Stdout
func PrettyPrintHashes(in string, out io.Writer) {
	const printIndexes = false
	nbResets := 0

	v := make([]any, 0)
	f, err := os.Open(in)
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
			v = append(v, fmt.Sprintf("RESET #%d", nbResets))
			nbResets++
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

	if out == nil {
		out = os.Stdout
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(v); err != nil {
		panic(err)
	}
}
func spaceOutFromRight(s string) string {
	n := len(s) + (len(s)+15)/16 - 1
	if n < 0 {
		return ""
	}
	var bb strings.Builder
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

// GetRepoRootPath assumes that current working directory is within the repo
func GetRepoRootPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	const repoName = "linea-monorepo"
	i := strings.LastIndex(wd, repoName)
	if i == -1 {
		return "", errors.New("could not find repo root")
	}
	i += len(repoName)
	return wd[:i], nil
}

// GetZkevmWitness returns a [zkevm.Witness]
func GetZkevmWitness(req *execution.Request, cfg *config.Config) (*execution.Response, *zkevm.Witness) {
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	return &out, witness.ZkEVM
}

func PrettyPrint(v reflect.Value, indent int) string {
	if !v.IsValid() {
		return "<invalid>"
	}

	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}

	// Handle interfaces
	if v.Kind() == reflect.Interface {
		return PrettyPrint(v.Elem(), indent)
	}

	switch v.Kind() {
	case reflect.Struct:
		var b strings.Builder
		t := v.Type()
		ind := strings.Repeat("  ", indent)
		b.WriteString(fmt.Sprintf("%s%s {\n", ind, t.Name()))
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)
			fieldName := fieldType.Name
			if !field.CanInterface() {
				b.WriteString(fmt.Sprintf("%s  %s: <unexported>\n", ind, fieldName))
				continue
			}
			fieldStr := PrettyPrint(field, indent+1)
			b.WriteString(fmt.Sprintf("%s  %s: %s\n", ind, fieldName, fieldStr))
		}
		b.WriteString(fmt.Sprintf("%s}", ind))
		return b.String()

	case reflect.Slice, reflect.Array:
		var b strings.Builder
		b.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(PrettyPrint(v.Index(i), indent))
		}
		b.WriteString("]")
		return b.String()

	case reflect.Map:
		var b strings.Builder
		b.WriteString("{")
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			if !key.CanInterface() || !val.CanInterface() {
				continue
			}
			b.WriteString(fmt.Sprintf("%s: %s, ", PrettyPrint(key, indent), PrettyPrint(val, indent)))
		}
		b.WriteString("}")
		return b.String()

	default:
		if v.CanInterface() {
			return fmt.Sprintf("%#v", v.Interface())
		}
		return "<unexported>"
	}
}
