package test_utils

import (
	"bytes"
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

	"github.com/consensys/gnark/frontend"
	snarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"

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

// ReaderHash is a wrapper around a hash.Hash that matches all writes with its input stream.
type ReaderHash struct {
	h hash.Hash
	r io.Reader
}

func (r *ReaderHash) Write(p []byte) (n int, err error) {
	if rd := hashReadWrite(r.r); !bytes.Equal(rd, p) {
		panic(fmt.Errorf("ReaderHash.Write: expected %x, encountered %x", rd, p))
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

func PrettyPrintHashesToFile(in, out string) {
	var t FakeTestingT
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)
	PrettyPrintHashes(in, f)
	require.NoError(t, err)
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
func GetZkevmWitness(req *execution.Request, cfg *config.Config) *zkevm.Witness {
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	return witness.ZkEVM
}

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *zkevm.ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := &config.TracesLimits{
		Add:                                  1 << 19,
		Bin:                                  1 << 18,
		Blake2Fmodexpdata:                    1 << 14,
		Blockdata:                            1 << 12,
		Blockhash:                            1 << 12,
		Ecdata:                               1 << 18,
		Euc:                                  1 << 16,
		Exp:                                  1 << 14,
		Ext:                                  1 << 20,
		Gas:                                  1 << 16,
		Hub:                                  1 << 21,
		Logdata:                              1 << 16,
		Loginfo:                              1 << 12,
		Mmio:                                 1 << 21,
		Mmu:                                  1 << 21,
		Mod:                                  1 << 17,
		Mul:                                  1 << 16,
		Mxp:                                  1 << 19,
		Oob:                                  1 << 18,
		Rlpaddr:                              1 << 12,
		Rlptxn:                               1 << 17,
		Rlptxrcpt:                            1 << 17,
		Rom:                                  1 << 22,
		Romlex:                               1 << 12,
		Shakiradata:                          1 << 15,
		Shf:                                  1 << 16,
		Stp:                                  1 << 14,
		Trm:                                  1 << 15,
		Txndata:                              1 << 14,
		Wcp:                                  1 << 18,
		Binreftable:                          1 << 20,
		Shfreftable:                          1 << 12,
		Instdecoder:                          1 << 9,
		PrecompileEcrecoverEffectiveCalls:    1 << 9,
		PrecompileSha2Blocks:                 1 << 9,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       1 << 10,
		PrecompileModexpEffectiveCalls4096:   1 << 4,
		PrecompileEcaddEffectiveCalls:        1 << 6,
		PrecompileEcmulEffectiveCalls:        1 << 6,
		PrecompileEcpairingEffectiveCalls:    1 << 4,
		PrecompileEcpairingMillerLoops:       1 << 4,
		PrecompileEcpairingG2MembershipCalls: 1 << 4,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 13,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    1 << 8,
		ShomeiMerkleProofs:                   1 << 14,
	}

	return zkevm.FullZKEVMWithSuite(traceLimits, zkevm.CompilationSuite{}, &config.Config{})
}

// GetAffinities returns a list of affinities for the following modules. This
// affinities regroup how the modules are grouped.
//
//	ecadd / ecmul / ecpairing
//	hub / hub.scp / hub.acp
//	everything related to keccak
func GetAffinities(z *zkevm.ZkEvm) [][]column.Natural {

	return [][]column.Natural{
		{
			z.Ecmul.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecadd.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecpair.AlignedFinalExpCircuit.IsActive.(column.Natural),
			z.Ecpair.AlignedG2MembershipData.IsActive.(column.Natural),
			z.Ecpair.AlignedMillerLoopCircuit.IsActive.(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("hub.HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.scp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.acp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.ccp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.envcp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.stkcp_PEEK_AT_STACK_POW_4").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("KECCAK_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("CLEANING_KECCAK_CleanLimb").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_KECCAK_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_KECCAK_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_IS_ACTIVE_").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_BLOCK_BASE_2_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_OVER_BLOCKS_TAGS_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("HASH_OUTPUT_Hash_Lo").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("SHA2_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_SHA2_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LENGTH_CONSISTENCY_SHA2_BYTE_LEN_0_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_SHA2_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("Coefficient_SHA2").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_IS_ACTIVE").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_IS_ACTIVE").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("mmio.CN_ABC").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmio.MMIO_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmu.STAMP").(column.Natural),
		},
	}
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

// CompareExportedFields checks if two values are equal, ignoring unexported fields, including in nested structs.
// It logs mismatched fields with their paths and values.
func CompareExportedFields(a, b interface{}) bool {
	return CompareExportedFieldsWithPath(a, b, "")
}

func CompareExportedFieldsWithPath(a, b interface{}, path string) bool {
	v1, v2 := reflect.ValueOf(a), reflect.ValueOf(b)

	// Ensure both values are valid
	if !v1.IsValid() || !v2.IsValid() {
		// Treat nil and zero values as equivalent
		if !v1.IsValid() && !v2.IsValid() {
			return true
		}
		if !v1.IsValid() {
			v1 = reflect.Zero(v2.Type())
		}
		if !v2.IsValid() {
			v2 = reflect.Zero(v1.Type())
		}
	}

	// Skip ignorable fields
	if serialization.IsIgnoreableType(v1.Type()) {
		logrus.Printf("Skipping comparison of Ignorable type:%s at %s\n", v1.Type().String(), path)
		return true
	}

	// Ensure same type
	if v1.Type() != v2.Type() {
		logrus.Printf("Mismatch at %s: types differ (v1: %v, v2: %v, types: %v, %v)\n", path, a, b, v1.Type(), v2.Type())
		return false
	}

	// Ignore Func
	if v1.Kind() == reflect.Func {
		return true
	}

	// Handle maps
	if v1.Kind() == reflect.Map {
		if v1.Len() != v2.Len() {
			if serialization.IsIgnoreableType(v1.Type()) {
				logrus.Printf("Skipping comparison of ignoreable types at %s\n", path)
				return true
			}
			logrus.Printf("Mismatch at %s: map lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1.Len(), v2.Len(), v1.Type())
			return false
		}
		for _, key := range v1.MapKeys() {
			value1 := v1.MapIndex(key)
			value2 := v2.MapIndex(key)
			if !value2.IsValid() {
				logrus.Printf("Mismatch at %s: key %v is missing in second map\n", path, key)
				return false
			}
			keyPath := fmt.Sprintf("%s[%v]", path, key)
			if !CompareExportedFieldsWithPath(value1.Interface(), value2.Interface(), keyPath) {
				return false
			}
		}
		// logrus.Infof("Comparing map at %s: len(v1)=%d, len(v2)=%d", path, v1.Len(), v2.Len())
		return true
	}

	// Handle pointers by dereferencing
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v1.IsNil() != v2.IsNil() {
			if serialization.IsIgnoreableType(v1.Type()) {
				logrus.Printf("Skipping comparison of ignoreable types at %s\n", path)
				return true
			}
			logrus.Printf("Mismatch at %s: nil status differs (v1: %v, v2: %v, type: %v)\n", path, a, b, v1.Type())
			return false
		}
		return CompareExportedFieldsWithPath(v1.Elem().Interface(), v2.Elem().Interface(), path)
	}

	// Handle structs
	if v1.Kind() == reflect.Struct {
		equal := true
		for i := 0; i < v1.NumField(); i++ {

			structField := v1.Type().Field(i)

			// When the field is has the omitted tag, we skip it there without
			// any warning.
			if tag, hasTag := structField.Tag.Lookup(serialization.SerdeStructTag); hasTag {
				if strings.Contains(tag, serialization.SerdeStructTagOmit) {
					continue
				}
			}

			// Skip unexported fields
			if !structField.IsExported() {
				continue
			}

			f1 := v1.Field(i)
			f2 := v2.Field(i)
			fieldName := structField.Name
			fieldPath := fieldName
			if path != "" {
				fieldPath = path + "." + fieldName
			}
			if !CompareExportedFieldsWithPath(f1.Interface(), f2.Interface(), fieldPath) {
				equal = false
			}
		}
		return equal
	}

	// Handle slices or arrays
	if v1.Kind() == reflect.Slice || v1.Kind() == reflect.Array {
		if v1.Len() != v2.Len() {
			logrus.Printf("Mismatch at %s: slice lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1, v2, v1.Type())
			return false
		}
		equal := true
		for i := 0; i < v1.Len(); i++ {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			if !CompareExportedFieldsWithPath(v1.Index(i).Interface(), v2.Index(i).Interface(), elemPath) {
				equal = false
			}
		}
		return equal
	}

	// For other types, use DeepEqual and log if mismatched
	if !reflect.DeepEqual(a, b) {
		logrus.Printf("Mismatch at %s: values differ (v1: %v, v2: %v, type_v1: %v type_v2: %v)\n", path, a, b, v1.Type(), v2.Type())
		panic("fail fast")
		return false
	}
	return true
}
