// File: serde/custom.go
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

// CustomCodex defines how to serialize and deserialize a specific type.
type CustomCodex struct {
	Serialize func(w *Writer, v reflect.Value) (Ref, error)
	// CHANGED: Accept *ReaderContext instead of []byte to allow recursion
	Deserialize func(ctx *ReaderContext, v reflect.Value, offset int64) error
}

var CustomRegistry = map[reflect.Type]CustomCodex{}

func RegisterCustomType(t reflect.Type, c CustomCodex) {
	CustomRegistry[t] = c
}

func init() {
	// --- 1. Big Int ---
	RegisterCustomType(reflect.TypeOf(big.Int{}), CustomCodex{Serialize: serializeBigInt, Deserialize: deserializeBigInt})
	RegisterCustomType(reflect.TypeOf(&big.Int{}), CustomCodex{Serialize: serializeBigIntPtr, Deserialize: deserializeBigInt})

	// --- 2. Frontend Variable ---
	// frontend.Variable is often an alias for interface{}. We register the concrete type typically found inside.
	RegisterCustomType(reflect.TypeOf((*frontend.Variable)(nil)).Elem(), CustomCodex{
		Serialize:   serializeFrontendVariable,
		Deserialize: deserializeFrontendVariable,
	})

	// --- 3. Gnark R1CS ---
	RegisterCustomType(reflect.TypeOf(cs.SparseR1CS{}), CustomCodex{Serialize: serializeR1CS, Deserialize: deserializeR1CS})
	RegisterCustomType(reflect.TypeOf(&cs.SparseR1CS{}), CustomCodex{Serialize: serializeR1CSPtr, Deserialize: deserializeR1CS})

	// --- 4. Arithmetization ---
	RegisterCustomType(reflect.TypeOf(arithmetization.Arithmetization{}), CustomCodex{
		Serialize:   serializeArithmetization,
		Deserialize: deserializeArithmetization,
	})

	// --- 5. RingSIS Key (Pointer) ---
	RegisterCustomType(reflect.TypeOf(&ringsis.Key{}), CustomCodex{
		Serialize:   serializeRingSisKey,
		Deserialize: deserializeRingSisKey,
	})

	// --- 6. FFT Domain (Pointer) ---
	RegisterCustomType(reflect.TypeOf(&fft.Domain{}), CustomCodex{
		Serialize:   serializeGnarkFFTDomain,
		Deserialize: deserializeGnarkFFTDomain,
	})

	// --- 7. Smartvectors (Regular) ---
	// Treated as []field.Element
	RegisterCustomType(reflect.TypeOf(smartvectors.Regular{}), CustomCodex{
		Serialize:   serializeSmartVectorRegular,
		Deserialize: deserializeSmartVectorRegular,
	})

	// --- 8. Hash Generators / Types ---
	// These are functions or specialized structs that we serialize as "empty" or "default"
	RegisterCustomType(reflect.TypeOf(func() hash.Hash { return nil }), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeHashGenerator,
	})
	RegisterCustomType(reflect.TypeOf(func() hashtypes.Hasher { return hashtypes.Hasher{} }), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeHashTypeHasher,
	})

	// --- 9. Mutex ---
	// We serialize as nil/empty and deserialize as a fresh zero mutex
	RegisterCustomType(reflect.TypeOf(sync.Mutex{}), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeAsZero,
	})
	RegisterCustomType(reflect.TypeOf(&sync.Mutex{}), CustomCodex{
		Serialize:   serializeAsEmpty,
		Deserialize: deserializeAsNewPtr,
	})

	// --- Column Natural ---
	RegisterCustomType(reflect.TypeOf(column.Natural{}), CustomCodex{
		Serialize:   serializeColumnNatural,
		Deserialize: deserializeColumnNatural,
	})

	// --- Coin Info ---
	RegisterCustomType(reflect.TypeOf(coin.Info{}), CustomCodex{
		Serialize:   serializeCoinInfo,
		Deserialize: deserializeCoinInfo,
	})

	// --- Column Store (Pointer) ---
	RegisterCustomType(reflect.TypeOf(&column.Store{}), CustomCodex{
		Serialize:   serializeColumnStore,
		Deserialize: deserializeColumnStore,
	})
}

// ---------------- IMPLEMENTATIONS ----------------

// ---------------- COLUMN NATURAL ----------------

func serializeColumnNatural(w *Writer, v reflect.Value) (Ref, error) {
	nat := v.Interface().(column.Natural)
	packed := nat.Pack()
	return linearize(w, reflect.ValueOf(packed))
}

func deserializeColumnNatural(ctx *ReaderContext, v reflect.Value, offset int64) error {
	var packed column.PackedNatural
	packedVal := reflect.ValueOf(&packed).Elem()

	// Use ctx.reconstruct
	if err := ctx.reconstruct(packedVal, offset); err != nil {
		return err
	}

	nat := packed.Unpack()
	v.Set(reflect.ValueOf(nat))
	return nil
}

// ---------------- COIN INFO ----------------

// PackedCoin is a compact representation of coin.Info used for serialization.
type PackedCoin struct {
	Type       int8   `serde:"t"`
	Size       int    `serde:"s"`
	UpperBound int32  `serde:"u"`
	Name       string `serde:"n"`
	Round      int    `serde:"r"`
}

// asPackedCoin converts a coin.Info to a PackedCoin.
func asPackedCoin(c coin.Info) PackedCoin {
	return PackedCoin{
		Type:       int8(c.Type),
		Size:       c.Size,
		UpperBound: int32(c.UpperBound),
		Name:       string(c.Name),
		Round:      c.Round,
	}
}

func serializeCoinInfo(w *Writer, v reflect.Value) (Ref, error) {
	c := v.Interface().(coin.Info)
	packed := asPackedCoin(c)
	return linearize(w, reflect.ValueOf(packed))
}

func deserializeCoinInfo(ctx *ReaderContext, v reflect.Value, offset int64) error {
	var packed PackedCoin
	// Use ctx.reconstruct
	if err := ctx.reconstruct(reflect.ValueOf(&packed).Elem(), offset); err != nil {
		return err
	}

	sizes := []int{}
	if packed.Size > 0 {
		sizes = append(sizes, packed.Size)
	}
	if packed.UpperBound > 0 {
		sizes = append(sizes, int(packed.UpperBound))
	}

	unpacked := coin.NewInfo(
		coin.Name(packed.Name),
		coin.Type(packed.Type),
		packed.Round,
		sizes...,
	)

	v.Set(reflect.ValueOf(unpacked))
	return nil
}

// ---------------- COLUMN STORE ----------------

func serializeColumnStore(w *Writer, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	store := v.Interface().(*column.Store)
	packed := store.Pack()
	return linearize(w, reflect.ValueOf(packed))
}

func deserializeColumnStore(ctx *ReaderContext, v reflect.Value, offset int64) error {
	var packed column.PackedStore
	packedVal := reflect.ValueOf(&packed).Elem()

	// Use ctx.reconstruct
	if err := ctx.reconstruct(packedVal, offset); err != nil {
		return err
	}

	store := packed.Unpack()
	v.Set(reflect.ValueOf(store))
	return nil
}

// --- Big Int ---
func serializeBigInt(w *Writer, v reflect.Value) (Ref, error) {
	if v.Kind() == reflect.Ptr {
		return writeBigInt(w, v.Interface().(*big.Int))
	}
	if v.CanAddr() {
		return writeBigInt(w, v.Addr().Interface().(*big.Int))
	}
	bi := v.Interface().(big.Int)
	return writeBigInt(w, &bi)
}
func serializeBigIntPtr(w *Writer, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	return writeBigInt(w, v.Interface().(*big.Int))
}
func deserializeBigInt(ctx *ReaderContext, v reflect.Value, offset int64) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return reconstructBigInt(ctx.data, v.Elem(), offset)
	}
	return reconstructBigInt(ctx.data, v, offset)
}

// --- Frontend Variable ---
func serializeFrontendVariable(w *Writer, v reflect.Value) (Ref, error) {
	bi := toBigInt(v)
	if bi == nil {
		return 0, nil
	}
	return writeBigInt(w, bi)
}
func deserializeFrontendVariable(ctx *ReaderContext, v reflect.Value, offset int64) error {
	bi := new(big.Int)
	if err := reconstructBigInt(ctx.data, reflect.ValueOf(bi).Elem(), offset); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(bi))
	return nil
}

// --- Arithmetization ---
func serializeArithmetization(w *Writer, v reflect.Value) (Ref, error) {
	var bodyBuf bytes.Buffer
	if err := linearizeStructBody(w, v, &bodyBuf); err != nil {
		return 0, err
	}
	off := w.WriteBytes(bodyBuf.Bytes())
	return Ref(off), nil
}
func deserializeArithmetization(ctx *ReaderContext, v reflect.Value, offset int64) error {
	// CHANGED: Call reconstructStruct to avoid infinite recursion into this handler
	if err := ctx.reconstructStruct(v, offset); err != nil {
		return err
	}

	arith := v.Addr().Interface().(*arithmetization.Arithmetization)
	var err error
	arith.BinaryFile, arith.Metadata, err = arithmetization.UnmarshalZkEVMBin(arith.ZkEVMBin)
	if err != nil {
		return fmt.Errorf("arithmetization unmarshal bin failed: %w", err)
	}
	arith.AirSchema, arith.LimbMapping = arithmetization.CompileZkevmBin(arith.BinaryFile, arith.Settings.OptimisationLevel)
	return nil
}

// --- RingSIS Key ---
func serializeRingSisKey(w *Writer, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	key := v.Interface().(*ringsis.Key)
	return Ref(w.Write(key.KeyGen.MaxNumFieldToHash)), nil
}
func deserializeRingSisKey(ctx *ReaderContext, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+8 > len(ctx.data) {
		return fmt.Errorf("ringsis key data out of bounds")
	}
	maxNum := *(*uint64)(unsafe.Pointer(&ctx.data[offset]))

	key := ringsis.GenerateKey(ringsis.StdParams, int(maxNum))
	v.Set(reflect.ValueOf(key))
	return nil
}

// --- Gnark FFT Domain ---
func serializeGnarkFFTDomain(w *Writer, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	domain := v.Interface().(*fft.Domain)
	var buf bytes.Buffer
	if _, err := domain.WriteTo(&buf); err != nil {
		return 0, err
	}
	off := w.WriteBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(w.Write(fs)), nil
}
func deserializeGnarkFFTDomain(ctx *ReaderContext, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
		return fmt.Errorf("fft domain header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}
	start := int64(fs.Offset)
	end := start + fs.Len
	if start < 0 || end > int64(len(ctx.data)) {
		return fmt.Errorf("fft domain data out of bounds")
	}

	d := &fft.Domain{}
	if _, err := d.ReadFrom(bytes.NewReader(ctx.data[start:end])); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(d))
	return nil
}

// --- SmartVectors ---
func serializeSmartVectorRegular(w *Writer, v reflect.Value) (Ref, error) {
	sliceVal := reflect.ValueOf(v.Interface().(smartvectors.Regular))
	fs := w.WriteSlice(sliceVal)
	return Ref(w.Write(fs)), nil
}
func deserializeSmartVectorRegular(ctx *ReaderContext, v reflect.Value, offset int64) error {
	sliceType := reflect.SliceOf(reflect.TypeOf(field.Element{}))
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(sliceVal)

	// Use ctx.reconstruct
	if err := ctx.reconstruct(slicePtr.Elem(), offset); err != nil {
		return err
	}
	v.Set(slicePtr.Elem().Convert(v.Type()))
	return nil
}

// --- Gnark R1CS ---
func serializeR1CS(w *Writer, v reflect.Value) (Ref, error) {
	var csPtr *cs.SparseR1CS
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, nil
		}
		csPtr = v.Interface().(*cs.SparseR1CS)
	} else if v.CanAddr() {
		csPtr = v.Addr().Interface().(*cs.SparseR1CS)
	} else {
		return 0, fmt.Errorf("cannot take address of R1CS")
	}

	var buf bytes.Buffer
	if _, err := csPtr.WriteTo(&buf); err != nil {
		return 0, fmt.Errorf("R1CS WriteTo failed: %w", err)
	}
	off := w.WriteBytes(buf.Bytes())
	fs := FileSlice{Offset: Ref(off), Len: int64(buf.Len()), Cap: int64(buf.Len())}
	return Ref(w.Write(fs)), nil
}
func serializeR1CSPtr(w *Writer, v reflect.Value) (Ref, error) {
	return serializeR1CS(w, v)
}
func deserializeR1CS(ctx *ReaderContext, v reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
		return fmt.Errorf("R1CS header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}
	start := int64(fs.Offset)
	end := start + fs.Len
	if start < 0 || end > int64(len(ctx.data)) {
		return fmt.Errorf("R1CS data out of bounds")
	}

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("target for R1CS must be a pointer")
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	csPtr := v.Interface().(*cs.SparseR1CS)
	if _, err := csPtr.ReadFrom(bytes.NewReader(ctx.data[start:end])); err != nil {
		return err
	}
	return nil
}

// --- Helpers for Hash / Mutex ---
func serializeAsEmpty(w *Writer, v reflect.Value) (Ref, error) {
	return Ref(w.Write(struct{}{})), nil
}
func deserializeHashGenerator(ctx *ReaderContext, v reflect.Value, offset int64) error {
	f := func() hash.Hash { return mimc.NewMiMC() }
	v.Set(reflect.ValueOf(f))
	return nil
}
func deserializeHashTypeHasher(ctx *ReaderContext, v reflect.Value, offset int64) error {
	v.Set(reflect.ValueOf(hashtypes.MiMC))
	return nil
}
func deserializeAsZero(ctx *ReaderContext, v reflect.Value, offset int64) error {
	v.Set(reflect.Zero(v.Type()))
	return nil
}
func deserializeAsNewPtr(ctx *ReaderContext, v reflect.Value, offset int64) error {
	v.Set(reflect.New(v.Type().Elem()))
	return nil
}
