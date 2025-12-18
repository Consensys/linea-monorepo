package serde

import (
	"bytes"
	"fmt"
	"hash"
	"math/big"
	"reflect"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
)

type customCodex struct {
	serialize   func(enc *encoder, v reflect.Value) (Ref, error)
	deserialize func(dec *decoder, v reflect.Value, offset int64) error
}

var customRegistry = map[reflect.Type]customCodex{}

func registerCustomType(t reflect.Type, c customCodex) {
	customRegistry[t] = c
}

func init() {
	// Value Types
	registerCustomType(reflect.TypeOf(big.Int{}), customCodex{serialize: serializeBigInt, deserialize: deserializeBigInt})
	registerCustomType(reflect.TypeOf(&big.Int{}), customCodex{serialize: serializeBigIntPtr, deserialize: deserializeBigInt})

	registerCustomType(reflect.TypeOf((*frontend.Variable)(nil)).Elem(), customCodex{
		serialize:   serializeFrontendVariable,
		deserialize: deserializeFrontendVariable,
	})

	// Gnark R1CS
	registerCustomType(reflect.TypeOf(cs.SparseR1CS{}), customCodex{serialize: serializeR1CS, deserialize: deserializeR1CS})

	// Arithmetization
	registerCustomType(reflect.TypeOf(arithmetization.Arithmetization{}), customCodex{
		serialize:   serializeArithmetization,
		deserialize: deserializeArithmetization,
	})

	// RingSIS Key
	registerCustomType(reflect.TypeOf(ringsis.Key{}), customCodex{
		serialize:   serializeRingSisKey,
		deserialize: deserializeRingSisKey,
	})

	// FFT Domain
	registerCustomType(reflect.TypeOf(fft.Domain{}), customCodex{
		serialize:   serializeGnarkFFTDomain,
		deserialize: deserializeGnarkFFTDomain,
	})

	// Column Store
	registerCustomType(reflect.TypeOf(column.Store{}), customCodex{
		serialize:   serializeColumnStore,
		deserialize: deserializeColumnStore,
	})

	// Column Natural / Coin Info
	registerCustomType(reflect.TypeOf(column.Natural{}), customCodex{
		serialize:   serializeColumnNatural,
		deserialize: deserializeColumnNatural,
	})
	registerCustomType(reflect.TypeOf(coin.Info{}), customCodex{
		serialize:   serializeCoinInfo,
		deserialize: deserializeCoinInfo,
	})

	// Helpers
	registerCustomType(reflect.TypeOf(smartvectors.Regular{}), customCodex{
		serialize:   serializeSmartVectorRegular,
		deserialize: deserializeSmartVectorRegular,
	})
	registerCustomType(reflect.TypeOf(func() hash.Hash { return nil }), customCodex{
		serialize:   serializeAsEmpty,
		deserialize: deserializeHashGenerator,
	})
	registerCustomType(reflect.TypeOf(func() hashtypes.Hasher { return hashtypes.Hasher{} }), customCodex{
		serialize:   serializeAsEmpty,
		deserialize: deserializeHashTypeHasher,
	})
	registerCustomType(reflect.TypeOf(sync.Mutex{}), customCodex{
		serialize:   serializeAsEmpty,
		deserialize: deserializeAsZero,
	})
	registerCustomType(reflect.TypeOf(&sync.Mutex{}), customCodex{
		serialize:   serializeAsEmpty,
		deserialize: deserializeAsNewPtr,
	})
}

// ---------------- IMPLEMENTATIONS ----------------

// --- Column Natural ---
func serializeColumnNatural(enc *encoder, v reflect.Value) (Ref, error) {
	nat := v.Interface().(column.Natural)
	packed := nat.Pack()
	return encode(enc, reflect.ValueOf(packed))
}
func deserializeColumnNatural(dec *decoder, v reflect.Value, offset int64) error {
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
func serializeCoinInfo(enc *encoder, v reflect.Value) (Ref, error) {
	c := v.Interface().(coin.Info)
	packed := asPackedCoin(c)
	return encode(enc, reflect.ValueOf(packed))
}
func deserializeCoinInfo(dec *decoder, v reflect.Value, offset int64) error {
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

// --- Column Store (STRUCT) ---
func serializeColumnStore(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	store := p.(*column.Store)
	packed := store.Pack()
	return encode(enc, reflect.ValueOf(packed))
}
func deserializeColumnStore(dec *decoder, v reflect.Value, offset int64) error {
	var packed column.PackedStore
	packedVal := reflect.ValueOf(&packed).Elem()
	if err := dec.decode(packedVal, offset); err != nil {
		return err
	}
	newStorePtr := packed.Unpack()
	v.Set(reflect.ValueOf(newStorePtr).Elem())
	return nil
}

// --- RingSIS Key (STRUCT) ---
func serializeRingSisKey(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	key := p.(*ringsis.Key)
	return Ref(enc.write(key.KeyGen.MaxNumFieldToHash)), nil
}
func deserializeRingSisKey(dec *decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+8 > len(dec.data) {
		return fmt.Errorf("ringsis key data out of bounds")
	}
	maxNum := *(*uint64)(unsafe.Pointer(&dec.data[offset]))
	key := ringsis.GenerateKey(ringsis.StdParams, int(maxNum))
	v.Set(reflect.ValueOf(key).Elem())
	return nil
}

// --- Gnark FFT Domain (STRUCT) ---
func serializeGnarkFFTDomain(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	domain := p.(*fft.Domain)
	var buf bytes.Buffer
	if _, err := domain.WriteTo(&buf); err != nil {
		return 0, err
	}
	off := enc.writeBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(enc.write(fs)), nil
}
func deserializeGnarkFFTDomain(dec *decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("fft domain header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}
	start := int64(fs.Offset)
	end := start + fs.Len
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
func serializeR1CS(enc *encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	csPtr := p.(*cs.SparseR1CS)
	var buf bytes.Buffer
	if _, err := csPtr.WriteTo(&buf); err != nil {
		return 0, err
	}
	off := enc.writeBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(enc.write(fs)), nil
}
func deserializeR1CS(dec *decoder, v reflect.Value, offset int64) error {
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
		return err
	}
	v.Set(reflect.ValueOf(csPtr).Elem())
	return nil
}

// --- Arithmetization ---
func serializeArithmetization(enc *encoder, v reflect.Value) (Ref, error) {
	var bodyBuf bytes.Buffer
	if err := linearizeStructBodyMap(enc, v, &bodyBuf); err != nil {
		return 0, err
	}
	off := enc.writeBytes(bodyBuf.Bytes())
	return Ref(off), nil
}
func deserializeArithmetization(dec *decoder, v reflect.Value, offset int64) error {
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
func serializeBigInt(enc *encoder, v reflect.Value) (Ref, error) {
	if v.Kind() == reflect.Ptr {
		return encodeBigInt(enc, v.Interface().(*big.Int))
	}
	if v.CanAddr() {
		return encodeBigInt(enc, v.Addr().Interface().(*big.Int))
	}
	bi := v.Interface().(big.Int)
	return encodeBigInt(enc, &bi)
}
func serializeBigIntPtr(enc *encoder, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	return encodeBigInt(enc, v.Interface().(*big.Int))
}
func deserializeBigInt(dec *decoder, v reflect.Value, offset int64) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeBigInt(dec.data, v.Elem(), offset)
	}
	return decodeBigInt(dec.data, v, offset)
}

func serializeFrontendVariable(enc *encoder, v reflect.Value) (Ref, error) {
	bi := toBigInt(v)
	if bi == nil {
		return 0, nil
	}
	return encodeBigInt(enc, bi)
}
func deserializeFrontendVariable(dec *decoder, v reflect.Value, offset int64) error {
	bi := new(big.Int)
	if err := decodeBigInt(dec.data, reflect.ValueOf(bi).Elem(), offset); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(bi))
	return nil
}
func serializeSmartVectorRegular(enc *encoder, v reflect.Value) (Ref, error) {
	// We expect a smartvectors.Regular type here
	sliceVal := v
	if v.Kind() == reflect.Interface {
		sliceVal = v.Elem()
	}

	fs := enc.writeSliceData(sliceVal)
	// Write the resulting FileSlice header and return its offset
	return Ref(enc.write(fs)), nil
}

func deserializeSmartVectorRegular(dec *decoder, v reflect.Value, offset int64) error {
	sliceType := reflect.SliceOf(reflect.TypeOf(field.Element{}))
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(sliceVal)
	if err := dec.decode(slicePtr.Elem(), offset); err != nil {
		return err
	}
	v.Set(slicePtr.Elem().Convert(v.Type()))
	return nil
}
func serializeAsEmpty(enc *encoder, v reflect.Value) (Ref, error) {
	// FIX: Write 1 byte so that the object has a unique offset/identity in the stream.
	return Ref(enc.write(byte(0))), nil
}

func deserializeHashGenerator(dec *decoder, v reflect.Value, offset int64) error {
	f := func() hash.Hash { return mimc.NewMiMC() }
	v.Set(reflect.ValueOf(f))
	return nil
}
func deserializeHashTypeHasher(dec *decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.ValueOf(hashtypes.MiMC))
	return nil
}
func deserializeAsZero(dec *decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.Zero(v.Type()))
	return nil
}
func deserializeAsNewPtr(dec *decoder, v reflect.Value, offset int64) error {
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

	// Note: We use enc.buf.Write (raw std. library write) to dump the "payload" (the sign byte + magnitude bytes)
	// directly onto the heap without any metadata. We then manually construct one FileSlice that points to that raw
	// dump. We use enc.write(fs) to write that header properly into the file structure.  This is done essentially to
	// avoid the "double headers" problem (i.e. to get the actual payload, the decoder would have to read a header to
	// find a header to find the bytes - thereby wasting 24 bytes of space and adding CPU overhead unnecessarily).
	enc.buf.WriteByte(byte(sign))
	enc.buf.Write(bytes)
	enc.offset += int64(1 + len(bytes)) // plus 1 for sign byte
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

	// Note: While encoding we write ths "sign" byte twice once in the encoder buffer and then in the FileSlice.Cap.
	// While it looks redundant, we favour this method for semantic Completeness of the Payload ([Sign Byte] + [Magnitude Bytes])
	// which is a complete, standalone binary representation of a Signed Integer. While decoding, we fetch the sign byte
	// from `FileSlice.Cap` because the decoder has already loaded the struct into the CPU cache while reading `Offset` and `Len`.
	// The `Cap` field is effectively already in a register. This avoids *peeking* into the data blob:
	// reading the sign byte would require following the pointer into the memory-mapped region (`dec.data[offset]`), performing an extra memory load,
	// and then creating a slice for the remaining data.
	if fs.Cap == 1 {
		bi.Neg(bi)
	}
	target.Set(reflect.ValueOf(*bi))
	return nil
}

func toBigInt(v reflect.Value) *big.Int {
	switch val := v.Interface().(type) {
	case int:
		return big.NewInt(int64(val))
	case int64:
		return big.NewInt(val)
	case uint64:
		return new(big.Int).SetUint64(val)
	case field.Element:
		bi := new(big.Int)
		val.BigInt(bi)
		return bi
	case big.Int:
		return &val
	case *big.Int:
		return val
	default:
		return nil
	}
}

// Helper to get pointer from struct value
func ptrFromStruct(v reflect.Value) (interface{}, error) {
	if !v.CanAddr() {
		return nil, fmt.Errorf("cannot address struct of type %v", v.Type())
	}
	return v.Addr().Interface(), nil
}
