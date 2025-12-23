package serde

import (
	"bytes"
	"fmt"
	"hash"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/sirupsen/logrus"
)

// --- REGISTRY DEFINITIONS ---

// indirectCodex handles types that are stored on the Heap (returning a Ref).
// Examples: BigInt, Strings, R1CS, Maps, Slices.
type indirectCodex struct {
	marshall   func(enc *encoder, v reflect.Value) (Ref, error)
	unmarshall func(dec *decoder, v reflect.Value, offset int64) error
}

// directCodex handles types that are stored Inline (Fixed Size).
// Examples: field.Element, small structs we want to control manually.
// These bypass the "Indirect" check, allowing arrays of these to be POD.
type directCodex struct {
	// binSize returns the on-disk size in bytes
	binSize func(t reflect.Type) int64
	// marshall writes data directly to the buffer/patch location.
	// It does NOT return a Ref, because the data is inline.
	marshall func(enc *encoder, v reflect.Value) error
	// patch writes data to a specific existing offset (used in Struct/Array patching)
	patch func(enc *encoder, offset int64, v reflect.Value) error
	// unmarshall reads data directly from the offset
	unmarshall func(dec *decoder, v reflect.Value, offset int64) error
}

var (
	// customIndirectRegistry stores types that result in a Reference (pointer).
	// isIndirectType() returns TRUE for these.
	customIndirectRegistry = map[reflect.Type]indirectCodex{}

	// customDirectRegistry stores types that are serialized inline.
	// isIndirectType() returns FALSE for these.
	customDirectRegistry = map[reflect.Type]directCodex{}
)

func registerIndirect(t reflect.Type, c indirectCodex) {
	customIndirectRegistry[t] = c
}

func registerDirect(t reflect.Type, c directCodex) {
	customDirectRegistry[t] = c
}

// --- INIT ---

func init() {

	// 1. DIRECT TYPES (Inline, Fixed Size) --------------------------

	// field.Element: The most critical Direct Type.
	// We register it here to bypass the BinaryMarshaler trap.
	registerDirect(reflect.TypeOf(field.Element{}), directCodex{
		binSize: func(_ reflect.Type) int64 { return 32 }, // [4]uint64 = 32 bytes
		marshall: func(enc *encoder, v reflect.Value) error {
			// Cast to [4]uint64 to FORCE raw Little-Endian write
			val := v.Interface().(field.Element)
			enc.write([4]uint64(val))
			return nil
		},
		patch: func(enc *encoder, offset int64, v reflect.Value) error {
			// Cast to [4]uint64 to FORCE raw Little-Endian write
			val := v.Interface().(field.Element)
			enc.patch(offset, [4]uint64(val))
			return nil
		},
		unmarshall: func(dec *decoder, v reflect.Value, offset int64) error {
			return dec.decodeFieldElement(v, offset)
		},
	})

	// 2. INDIRECT TYPES (Heap, Variable Size) -----------------------

	idCodex := indirectCodex{
		marshall:   marshallID,
		unmarshall: unmarshallID,
	}

	registerIndirect(reflect.TypeOf(ifaces.ColID("")), idCodex)
	registerIndirect(reflect.TypeOf(ifaces.QueryID("")), idCodex)
	registerIndirect(reflect.TypeOf(coin.Name("")), idCodex)

	// Value Types
	registerIndirect(reflect.TypeOf(big.Int{}), indirectCodex{
		marshall:   marshallBigInt,
		unmarshall: unmarshallBigInt})

	registerIndirect(reflect.TypeOf(&big.Int{}), indirectCodex{
		marshall:   marshallBigIntPtr,
		unmarshall: unmarshallBigInt})

	registerIndirect(reflect.TypeOf((*frontend.Variable)(nil)).Elem(), indirectCodex{
		marshall:   marshallFrontendVariable,
		unmarshall: unmarshallFrontendVariable,
	})

	// Gnark R1CS
	registerIndirect(reflect.TypeOf(cs.SparseR1CS{}), indirectCodex{
		marshall:   marshallPlonkCkt,
		unmarshall: unmarshallPlonkCkt})

	// Arithmetization
	registerIndirect(reflect.TypeOf(arithmetization.Arithmetization{}), indirectCodex{
		marshall:   marshallArithmetization,
		unmarshall: unmarshallArithmetization,
	})

	// RingSIS Key
	registerIndirect(reflect.TypeOf(ringsis.Key{}), indirectCodex{
		marshall:   marshallRingSisKey,
		unmarshall: unmarshallRingSisKey,
	})

	// FFT Domain
	registerIndirect(reflect.TypeOf(fft.Domain{}), indirectCodex{
		marshall:   marshallGnarkFFTDomain,
		unmarshall: unmarshallGnarkFFTDomain,
	})

	// Column Store
	registerIndirect(reflect.TypeOf(column.Store{}), indirectCodex{
		marshall:   marshallColumnStore,
		unmarshall: unmarshallColumnStore,
	})

	// Column Natural / Coin Info
	registerIndirect(reflect.TypeOf(column.Natural{}), indirectCodex{
		marshall:   marshallColumnNatural,
		unmarshall: unmarshallColumnNatural,
	})
	registerIndirect(reflect.TypeOf(coin.Info{}), indirectCodex{
		marshall:   marshallCoinInfo,
		unmarshall: unmarshallCoinInfo,
	})

	// Helpers
	registerIndirect(reflect.TypeOf(func() hash.Hash { return nil }), indirectCodex{
		marshall:   marshallAsEmpty,
		unmarshall: unmarshallHashGenerator,
	})
	registerIndirect(reflect.TypeOf(func() hashtypes.Hasher { return hashtypes.Hasher{} }), indirectCodex{
		marshall:   marshallAsEmpty,
		unmarshall: unmarshallHashTypeHasher,
	})
	registerIndirect(reflect.TypeOf(sync.Mutex{}), indirectCodex{
		marshall:   marshallAsEmpty,
		unmarshall: unmarshallAsZero,
	})
	registerIndirect(reflect.TypeOf(&sync.Mutex{}), indirectCodex{
		marshall:   marshallAsEmpty,
		unmarshall: unmarshallAsNewPtr,
	})
}

// ---------------- IMPLEMENTATIONS ----------------

// marshallID handles string deduplication.
func marshallID(enc *encoder, v reflect.Value) (Ref, error) {
	s := v.String()

	// 1. Check Cache (Deduplication)
	if ref, ok := enc.idMap[s]; ok {
		traceLog("ID cache Hit! '%s' -> Ref %d", s, ref)
		return ref, nil
	}

	// 2. Write Data (Payload)
	startOffset := enc.offset
	enc.buf.WriteString(s)
	enc.offset += int64(len(s))

	// 3. Write Header
	fs := FileSlice{
		Offset: Ref(startOffset),
		Len:    int64(len(s)),
		Cap:    int64(len(s)),
	}

	headerOff := enc.write(fs)
	ref := Ref(headerOff)

	// 4. Update Cache
	enc.idMap[s] = ref
	traceLog("String Intern New: '%s' -> Ref %d", s, ref)

	return ref, nil
}

func unmarshallID(dec *decoder, v reflect.Value, offset int64) error {
	return dec.decodeString(v, offset)
}

// --- Column Store (STRUCT) ---
func marshallColumnStore(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, nil
	}
	store := p.(*column.Store)
	packed := store.Pack()
	return encode(enc, reflect.ValueOf(packed))
}
func unmarshallColumnStore(dec *decoder, v reflect.Value, offset int64) error {
	var packed column.PackedStore
	packedVal := reflect.ValueOf(&packed).Elem()
	if err := dec.decode(packedVal, offset); err != nil {
		return err
	}
	newStorePtr := packed.Unpack()
	v.Set(reflect.ValueOf(newStorePtr).Elem())
	return nil
}

// --- Column Natural ---
func marshallColumnNatural(enc *encoder, v reflect.Value) (Ref, error) {
	nat := v.Interface().(column.Natural)

	// 1. Logical Deduplication by UUID
	id := nat.UUID().String()
	if ref, ok := enc.uuidMap[id]; ok {
		traceLog("UUID Dedup Hit! %s -> Ref %d", id, ref)
		return ref, nil
	}

	packed := nat.Pack()

	// Encode the packed struct
	ref, err := encode(enc, reflect.ValueOf(packed))
	if err != nil {
		return 0, err
	}

	// 2. Register the new reference
	enc.uuidMap[id] = ref

	return ref, nil
}
func unmarshallColumnNatural(dec *decoder, v reflect.Value, offset int64) error {
	var packed column.PackedNatural
	packedVal := reflect.ValueOf(&packed).Elem()
	if err := dec.decode(packedVal, offset); err != nil {
		return err
	}
	nat := packed.Unpack()
	v.Set(reflect.ValueOf(nat))
	return nil
}

// --- Coin Info ---
type PackedCoin struct {
	Type       int8   `serde:"t"`
	Size       int    `serde:"s"`
	UpperBound int32  `serde:"u"`
	Name       string `serde:"n"`
	Round      int    `serde:"r"`
}

func asPackedCoin(c coin.Info) PackedCoin {
	return PackedCoin{Type: int8(c.Type), Size: c.Size, UpperBound: int32(c.UpperBound), Name: string(c.Name), Round: c.Round}
}
func marshallCoinInfo(enc *encoder, v reflect.Value) (Ref, error) {
	c := v.Interface().(coin.Info)
	id := c.UUID().String()
	if ref, ok := enc.uuidMap[id]; ok {
		return ref, nil
	}
	packed := asPackedCoin(c)

	ref, err := encode(enc, reflect.ValueOf(packed))
	if err != nil {
		return 0, err
	}
	enc.uuidMap[id] = ref
	return ref, nil
}
func unmarshallCoinInfo(dec *decoder, v reflect.Value, offset int64) error {
	var packed PackedCoin
	if err := dec.decode(reflect.ValueOf(&packed).Elem(), offset); err != nil {
		return err
	}
	sizes := []int{}
	if packed.Size > 0 {
		sizes = append(sizes, packed.Size)
	}
	if packed.UpperBound > 0 {
		sizes = append(sizes, int(packed.UpperBound))
	}
	unpacked := coin.NewInfo(coin.Name(packed.Name), coin.Type(packed.Type), packed.Round, sizes...)
	v.Set(reflect.ValueOf(unpacked))
	return nil
}

// --- RingSIS Key (STRUCT) ---
func marshallRingSisKey(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	key := p.(*ringsis.Key)
	return Ref(enc.write(key.KeyGen.MaxNumFieldToHash)), nil
}
func unmarshallRingSisKey(dec *decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+8 > len(dec.data) {
		return fmt.Errorf("ringsis key data out of bounds")
	}
	maxNum := *(*uint64)(unsafe.Pointer(&dec.data[offset]))
	key := ringsis.GenerateKey(ringsis.StdParams, int(maxNum))
	v.Set(reflect.ValueOf(key).Elem())
	return nil
}

// --- Gnark FFT Domain (STRUCT) ---
func marshallGnarkFFTDomain(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	domain := p.(*fft.Domain)
	startOffset := enc.offset

	// Zero allocation direct write
	n, err := domain.WriteTo(enc.buf)
	if err != nil {
		return 0, fmt.Errorf("fft domain write failed: %w", err)
	}

	enc.offset += n
	fs := FileSlice{Offset: Ref(startOffset), Len: n, Cap: n}
	headerOffset := enc.write(fs)
	return Ref(headerOffset), nil
}
func unmarshallGnarkFFTDomain(dec *decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("fft domain header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}

	start := int64(fs.Offset)
	end := start + int64(fs.Len)
	if start < 0 || end > int64(len(dec.data)) {
		return fmt.Errorf("fft domain data out of bounds")
	}

	d := &fft.Domain{}
	if _, err := d.ReadFrom(bytes.NewReader(dec.data[start:end])); err != nil {
		return err
	}

	v.Set(reflect.ValueOf(d).Elem())
	return nil
}

// --- Gnark R1CS (STRUCT) ---
func marshallPlonkCkt(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}

	csPtr, ok := p.(*cs.SparseR1CS)
	if !ok {
		return 0, fmt.Errorf("serializeR1CS: expected *cs.SparseR1CS, got %T", p)
	}

	startOffset := enc.offset
	n, err := csPtr.WriteTo(enc.buf)
	if err != nil {
		return 0, fmt.Errorf("R1CS write failed: %w", err)
	}

	enc.offset += n
	fs := FileSlice{Offset: Ref(startOffset), Len: n, Cap: n}
	off := enc.write(fs)

	return Ref(off), nil
}
func unmarshallPlonkCkt(dec *decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("R1CS header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}

	start := int64(fs.Offset)
	end := start + fs.Len
	if start < 0 || end > int64(len(dec.data)) {
		return fmt.Errorf("R1CS data out of bounds")
	}

	csPtr := &cs.SparseR1CS{}
	if _, err := csPtr.ReadFrom(bytes.NewReader(dec.data[start:end])); err != nil {
		return fmt.Errorf("failed to read R1CS: %w", err)
	}

	if v.Kind() == reflect.Ptr {
		v.Set(reflect.ValueOf(csPtr))
	} else {
		v.Set(reflect.ValueOf(csPtr).Elem())
	}
	return nil
}

// --- Arithmetization ---
func marshallArithmetization(enc *encoder, v reflect.Value) (Ref, error) {
	var bodyBuf bytes.Buffer
	if err := linearizeStructBodyMap(enc, v, &bodyBuf); err != nil {
		return 0, err
	}
	off := enc.writeBytes(bodyBuf.Bytes())
	return Ref(off), nil
}
func unmarshallArithmetization(dec *decoder, v reflect.Value, offset int64) error {
	if err := dec.decodeStruct(v, offset); err != nil {
		return err
	}
	arith := v.Addr().Interface().(*arithmetization.Arithmetization)
	var err error
	arith.BinaryFile, arith.Metadata, err = arithmetization.UnmarshalZkEVMBin(arith.ZkEVMBin)
	if err != nil {
		return err
	}
	arith.AirSchema, arith.LimbMapping = arithmetization.CompileZkevmBin(arith.BinaryFile, arith.Settings.OptimisationLevel)
	return nil
}

// --- Big Int, Frontend Variable, SmartVectors, Helpers ---
func marshallBigInt(enc *encoder, v reflect.Value) (Ref, error) {
	if v.Kind() == reflect.Ptr {
		return encodeBigInt(enc, v.Interface().(*big.Int))
	}
	if v.CanAddr() {
		return encodeBigInt(enc, v.Addr().Interface().(*big.Int))
	}
	bi := v.Interface().(big.Int)
	return encodeBigInt(enc, &bi)
}
func marshallBigIntPtr(enc *encoder, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	return encodeBigInt(enc, v.Interface().(*big.Int))
}
func unmarshallBigInt(dec *decoder, v reflect.Value, offset int64) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeBigInt(dec.data, v.Elem(), offset)
	}
	return decodeBigInt(dec.data, v, offset)
}

func marshallFrontendVariable(enc *encoder, v reflect.Value) (Ref, error) {
	bi := toBigInt(v)
	if bi == nil {
		return 0, nil
	}
	return encodeBigInt(enc, bi)
}
func unmarshallFrontendVariable(dec *decoder, v reflect.Value, offset int64) error {
	bi := new(big.Int)
	if err := decodeBigInt(dec.data, reflect.ValueOf(bi).Elem(), offset); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(bi))
	return nil
}

func marshallAsEmpty(enc *encoder, v reflect.Value) (Ref, error) {
	return Ref(enc.write(byte(0))), nil
}

func unmarshallHashGenerator(dec *decoder, v reflect.Value, offset int64) error {
	f := func() hash.Hash { return mimc.NewMiMC() }
	v.Set(reflect.ValueOf(f))
	return nil
}
func unmarshallHashTypeHasher(dec *decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.ValueOf(hashtypes.MiMC))
	return nil
}
func unmarshallAsZero(dec *decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.Zero(v.Type()))
	return nil
}
func unmarshallAsNewPtr(dec *decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.New(v.Type().Elem()))
	return nil
}

func encodeBigInt(enc *encoder, b *big.Int) (Ref, error) {
	if b == nil {
		return 0, nil
	}
	bytes := b.Bytes()
	sign := int64(0)
	if b.Sign() < 0 {
		sign = 1
	}
	startOffset := enc.offset

	enc.buf.WriteByte(byte(sign))
	enc.buf.Write(bytes)
	enc.offset += int64(1 + len(bytes))
	fs := FileSlice{Offset: Ref(startOffset), Len: int64(len(bytes)), Cap: sign}
	off := enc.write(fs)
	return Ref(off), nil
}

func decodeBigInt(data []byte, target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(data) {
		return fmt.Errorf("bigint header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}
	dataStart := int64(fs.Offset) + 1
	dataLen := int64(fs.Len)
	if dataStart+dataLen > int64(len(data)) {
		return fmt.Errorf("bigint data out of bounds")
	}
	bytes := data[dataStart : dataStart+dataLen]
	bi := new(big.Int).SetBytes(bytes)

	if fs.Cap == 1 {
		bi.Neg(bi)
	}
	target.Set(reflect.ValueOf(*bi))
	return nil
}

func toBigInt(v reflect.Value) *big.Int {
	if !v.IsValid() || (v.Kind() == reflect.Interface && v.IsNil()) {
		return nil
	}

	switch val := v.Interface().(type) {
	case int:
		return big.NewInt(int64(val))
	case int8:
		return big.NewInt(int64(val))
	case int16:
		return big.NewInt(int64(val))
	case int32:
		return big.NewInt(int64(val))
	case int64:
		return big.NewInt(val)

	case uint:
		return new(big.Int).SetUint64(uint64(val))
	case uint8:
		return new(big.Int).SetUint64(uint64(val))
	case uint16:
		return new(big.Int).SetUint64(uint64(val))
	case uint32:
		return new(big.Int).SetUint64(uint64(val))
	case uint64:
		return new(big.Int).SetUint64(val)

	case field.Element:
		bi := new(big.Int)
		val.BigInt(bi)
		return bi
	case big.Int:
		return new(big.Int).Set(&val)
	case *big.Int:
		return new(big.Int).Set(val)

	case string:
		if bi, ok := new(big.Int).SetString(val, 10); ok {
			return bi
		}
		return nil

	default:
		if strings.Contains(v.Type().String(), "expr.Term") {
			logrus.Warn("******* Skipping expr.Term *******")
			return nil
		}
		return nil
	}
}

func ptrFromStruct(v reflect.Value) (interface{}, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		return v.Interface(), nil
	}

	if !v.CanAddr() {
		return nil, fmt.Errorf("cannot address value of type %v", v.Type())
	}
	return v.Addr().Interface(), nil
}
