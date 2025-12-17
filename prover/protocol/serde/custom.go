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

type CustomCodex struct {
	Serialize   func(enc *Encoder, v reflect.Value) (Ref, error)
	Deserialize func(dec *Decoder, v reflect.Value, offset int64) error
}

var CustomRegistry = map[reflect.Type]CustomCodex{}

func RegisterCustomType(t reflect.Type, c CustomCodex) {
	CustomRegistry[t] = c
}

func init() {
	// Value Types
	RegisterCustomType(reflect.TypeOf(big.Int{}), CustomCodex{Serialize: serializeBigInt, Deserialize: deserializeBigInt})
	RegisterCustomType(reflect.TypeOf(&big.Int{}), CustomCodex{Serialize: serializeBigIntPtr, Deserialize: deserializeBigInt})

	RegisterCustomType(reflect.TypeOf((*frontend.Variable)(nil)).Elem(), CustomCodex{
		Serialize:   serializeFrontendVariable,
		Deserialize: deserializeFrontendVariable,
	})

	// Gnark R1CS
	RegisterCustomType(reflect.TypeOf(cs.SparseR1CS{}), CustomCodex{Serialize: serializeR1CS, Deserialize: deserializeR1CS})

	// Arithmetization
	RegisterCustomType(reflect.TypeOf(arithmetization.Arithmetization{}), CustomCodex{
		Serialize:   serializeArithmetization,
		Deserialize: deserializeArithmetization,
	})

	// RingSIS Key
	RegisterCustomType(reflect.TypeOf(ringsis.Key{}), CustomCodex{
		Serialize:   serializeRingSisKey,
		Deserialize: deserializeRingSisKey,
	})

	// FFT Domain
	RegisterCustomType(reflect.TypeOf(fft.Domain{}), CustomCodex{
		Serialize:   serializeGnarkFFTDomain,
		Deserialize: deserializeGnarkFFTDomain,
	})

	// Column Store
	RegisterCustomType(reflect.TypeOf(column.Store{}), CustomCodex{
		Serialize:   serializeColumnStore,
		Deserialize: deserializeColumnStore,
	})

	// Column Natural / Coin Info
	RegisterCustomType(reflect.TypeOf(column.Natural{}), CustomCodex{
		Serialize:   serializeColumnNatural,
		Deserialize: deserializeColumnNatural,
	})
	RegisterCustomType(reflect.TypeOf(coin.Info{}), CustomCodex{
		Serialize:   serializeCoinInfo,
		Deserialize: deserializeCoinInfo,
	})

	// Helpers
	RegisterCustomType(reflect.TypeOf(smartvectors.Regular{}), CustomCodex{
		Serialize:   serializeSmartVectorRegular,
		Deserialize: deserializeSmartVectorRegular,
	})
	RegisterCustomType(reflect.TypeOf(func() hash.Hash { return nil }), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeHashGenerator,
	})
	RegisterCustomType(reflect.TypeOf(func() hashtypes.Hasher { return hashtypes.Hasher{} }), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeHashTypeHasher,
	})
	RegisterCustomType(reflect.TypeOf(sync.Mutex{}), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeAsZero,
	})
	RegisterCustomType(reflect.TypeOf(&sync.Mutex{}), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeAsNewPtr,
	})
}

// ---------------- IMPLEMENTATIONS ----------------

// --- Column Natural ---
func serializeColumnNatural(enc *Encoder, v reflect.Value) (Ref, error) {
	nat := v.Interface().(column.Natural)
	packed := nat.Pack()
	return linearize(enc, reflect.ValueOf(packed))
}
func deserializeColumnNatural(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeCoinInfo(enc *Encoder, v reflect.Value) (Ref, error) {
	c := v.Interface().(coin.Info)
	packed := asPackedCoin(c)
	return linearize(enc, reflect.ValueOf(packed))
}
func deserializeCoinInfo(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeColumnStore(enc *Encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	store := p.(*column.Store)
	packed := store.Pack()
	return linearize(enc, reflect.ValueOf(packed))
}
func deserializeColumnStore(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeRingSisKey(enc *Encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	key := p.(*ringsis.Key)
	return Ref(enc.Write(key.KeyGen.MaxNumFieldToHash)), nil
}
func deserializeRingSisKey(dec *Decoder, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+8 > len(dec.data) {
		return fmt.Errorf("ringsis key data out of bounds")
	}
	maxNum := *(*uint64)(unsafe.Pointer(&dec.data[offset]))
	key := ringsis.GenerateKey(ringsis.StdParams, int(maxNum))
	v.Set(reflect.ValueOf(key).Elem())
	return nil
}

// --- Gnark FFT Domain (STRUCT) ---
func serializeGnarkFFTDomain(enc *Encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	domain := p.(*fft.Domain)
	var buf bytes.Buffer
	if _, err := domain.WriteTo(&buf); err != nil {
		return 0, err
	}
	off := enc.WriteBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(enc.Write(fs)), nil
}
func deserializeGnarkFFTDomain(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeR1CS(enc *Encoder, v reflect.Value) (Ref, error) {
	p, err := ptrFromStruct(v)
	if err != nil {
		return 0, err
	}
	csPtr := p.(*cs.SparseR1CS)
	var buf bytes.Buffer
	if _, err := csPtr.WriteTo(&buf); err != nil {
		return 0, err
	}
	off := enc.WriteBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(enc.Write(fs)), nil
}
func deserializeR1CS(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeArithmetization(enc *Encoder, v reflect.Value) (Ref, error) {
	var bodyBuf bytes.Buffer
	if err := linearizeStructBodyMap(enc, v, &bodyBuf); err != nil {
		return 0, err
	}
	off := enc.WriteBytes(bodyBuf.Bytes())
	return Ref(off), nil
}
func deserializeArithmetization(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeBigInt(enc *Encoder, v reflect.Value) (Ref, error) {
	if v.Kind() == reflect.Ptr {
		return writeBigInt(enc, v.Interface().(*big.Int))
	}
	if v.CanAddr() {
		return writeBigInt(enc, v.Addr().Interface().(*big.Int))
	}
	bi := v.Interface().(big.Int)
	return writeBigInt(enc, &bi)
}
func serializeBigIntPtr(enc *Encoder, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	return writeBigInt(enc, v.Interface().(*big.Int))
}
func deserializeBigInt(dec *Decoder, v reflect.Value, offset int64) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeBigInt(dec.data, v.Elem(), offset)
	}
	return decodeBigInt(dec.data, v, offset)
}
func serializeFrontendVariable(enc *Encoder, v reflect.Value) (Ref, error) {
	bi := toBigInt(v)
	if bi == nil {
		return 0, nil
	}
	return writeBigInt(enc, bi)
}
func deserializeFrontendVariable(dec *Decoder, v reflect.Value, offset int64) error {
	bi := new(big.Int)
	if err := decodeBigInt(dec.data, reflect.ValueOf(bi).Elem(), offset); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(bi))
	return nil
}
func serializeSmartVectorRegular(enc *Encoder, v reflect.Value) (Ref, error) {
	// We expect a smartvectors.Regular type here
	sliceVal := v
	if v.Kind() == reflect.Interface {
		sliceVal = v.Elem()
	}

	fs := enc.writeSliceData(sliceVal)
	// Write the resulting FileSlice header and return its offset
	return Ref(enc.Write(fs)), nil
}

func deserializeSmartVectorRegular(dec *Decoder, v reflect.Value, offset int64) error {
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
func serializeAsEmpty(enc *Encoder, v reflect.Value) (Ref, error) {
	// FIX: Write 1 byte so that the object has a unique offset/identity in the stream.
	return Ref(enc.Write(byte(0))), nil
}

func deserializeHashGenerator(dec *Decoder, v reflect.Value, offset int64) error {
	f := func() hash.Hash { return mimc.NewMiMC() }
	v.Set(reflect.ValueOf(f))
	return nil
}
func deserializeHashTypeHasher(dec *Decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.ValueOf(hashtypes.MiMC))
	return nil
}
func deserializeAsZero(dec *Decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.Zero(v.Type()))
	return nil
}
func deserializeAsNewPtr(dec *Decoder, v reflect.Value, offset int64) error {
	v.Set(reflect.New(v.Type().Elem()))
	return nil
}

func writeBigInt(enc *Encoder, b *big.Int) (Ref, error) {
	if b == nil {
		return 0, nil
	}
	bytes := b.Bytes()
	sign := int64(0)
	if b.Sign() < 0 {
		sign = 1
	}
	start := enc.offset
	enc.buf.WriteByte(byte(sign))
	enc.buf.Write(bytes)
	enc.offset += int64(1 + len(bytes))
	fs := FileSlice{Offset: Ref(start), Len: int64(len(bytes)), Cap: sign}
	off := enc.Write(fs)
	return Ref(off), nil
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
