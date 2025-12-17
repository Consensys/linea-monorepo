// Reader maps that block into memory and "swizzles" (converts) file offsets into real memory addresses
package serde

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

type ReaderContext struct {

	// The entire memory-mapped file - the raw binary blob. The reader does not "read" from a stream;
	// it jumps around this byte slice using offsets.
	data []byte

	// The Deduplication Table - critical for maintaining referential integrity
	// For example: If object A and object B both point to object C, the file only stores object C once.
	// When the reader encounters the offset for C the first time, it deserializes it and stores the result in ptrMap.
	// When it encounters the offset for C the second time, it grabs the existing object from ptrMap.
	// Benefit: This preserves cycles (A -> B -> A) and saves memory.
	ptrMap map[int64]reflect.Value
}

func (ctx *ReaderContext) reconstruct(target reflect.Value, offset int64) error {
	t := target.Type()

	// 1. Custom Registry
	if handler, ok := CustomRegistry[t]; ok {
		return handler.Deserialize(ctx, target, offset)
	}

	// 2. Main Type Switch
	switch target.Kind() {
	case reflect.Ptr:
		return ctx.reconstructPointer(target, offset)
	case reflect.Slice:
		return ctx.reconstructSlice(target, offset)
	case reflect.String:
		return ctx.reconstructString(target, offset)
	case reflect.Map:
		return ctx.reconstructMap(target, offset)
	case reflect.Interface:
		return ctx.reconstructInterface(target, offset)
	case reflect.Struct:
		// Fallback for field.Element if not in registry
		if t == reflect.TypeOf(field.Element{}) {
			return ctx.reconstructFieldElement(target, offset)
		}
		return ctx.reconstructStruct(target, offset)
	case reflect.Array:
		return ctx.reconstructArray(target, offset)
	case reflect.Int, reflect.Int64:
		if offset+8 > int64(len(ctx.data)) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetInt(val64)
		return nil
	case reflect.Uint, reflect.Uint64:
		if offset+8 > int64(len(ctx.data)) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetUint(val64)
		return nil
	default:
		// Generic copy for other primitives (bool, byte, etc)
		return ctx.reconstructPrimitive(target, offset)
	}
}

// --- Helpers ---

func (ctx *ReaderContext) reconstructPointer(target reflect.Value, offset int64) error {
	t := target.Type()

	// Check Cache
	if val, ok := ctx.ptrMap[offset]; ok {
		if val.Type().AssignableTo(t) {
			target.Set(val)
			return nil
		}
	}

	// Generic Pointer
	newPtr := reflect.New(t.Elem())
	ctx.ptrMap[offset] = newPtr
	target.Set(newPtr)
	return ctx.reconstruct(newPtr.Elem(), offset)
}

func (ctx *ReaderContext) reconstructSlice(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
		return fmt.Errorf("slice header out of bounds")
	}

	fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))

	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// Verify bounds of the underlying data start
	if int(fs.Offset) >= len(ctx.data) {
		return fmt.Errorf("slice data offset out of bounds")
	}

	// ZERO-COPY MAGIC
	// 1. Get the pointer to the raw data inside the memory map
	dataPtr := unsafe.Pointer(&ctx.data[fs.Offset])

	// 2. Manually construct the slice header to point to this data
	// We cast the target's address to a struct that mimics the Go slice layout.
	// This works for all slice types.
	sh := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(target.UnsafeAddr()))

	sh.Data = uintptr(dataPtr)
	sh.Len = int(fs.Len)
	sh.Cap = int(fs.Cap)

	return nil
}

func (ctx *ReaderContext) reconstructString(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
		return fmt.Errorf("string header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}

	start := int64(fs.Offset)
	end := start + int64(fs.Len)
	if start < 0 || end > int64(len(ctx.data)) {
		return fmt.Errorf("string content out of bounds")
	}

	// Zero-copy string construction
	strBytes := ctx.data[start:end]
	// Standard string(bytes) copies. To avoid copy, we'd need unsafe.String (Go 1.20+)
	// target.SetString(unsafe.String(&ctx.data[start], fs.Len))
	target.SetString(string(strBytes))
	return nil
}

func (ctx *ReaderContext) reconstructMap(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
		return fmt.Errorf("map header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	target.Set(reflect.MakeMapWithSize(target.Type(), int(fs.Len)))
	currentOff := int64(fs.Offset)
	keyType := target.Type().Key()
	elemType := target.Type().Elem()

	for i := 0; i < int(fs.Len); i++ {
		newKey := reflect.New(keyType).Elem()
		nextOff, err := ctx.readMapElement(newKey, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff

		newVal := reflect.New(elemType).Elem()
		nextOff, err = ctx.readMapElement(newVal, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
		target.SetMapIndex(newKey, newVal)
	}
	return nil
}

func (ctx *ReaderContext) reconstructInterface(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(ctx.data) {
		return fmt.Errorf("interface header out of bounds")
	}
	ih := (*InterfaceHeader)(unsafe.Pointer(&ctx.data[offset]))
	if ih.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}
	if int(ih.TypeID) >= len(IDToType) {
		return fmt.Errorf("invalid type ID: %d", ih.TypeID)
	}

	concreteType := IDToType[ih.TypeID]
	for i := 0; i < int(ih.Indirection); i++ {
		concreteType = reflect.PointerTo(concreteType)
	}

	var concreteVal reflect.Value
	if concreteType.Kind() == reflect.Ptr {
		concreteVal = reflect.New(concreteType.Elem())
		ptrOffset := int64(ih.Offset)
		// Check cache for this specific interface pointer
		if val, ok := ctx.ptrMap[ptrOffset]; ok {
			if val.Type().AssignableTo(concreteType) {
				target.Set(val)
				return nil
			}
		}
		ctx.ptrMap[ptrOffset] = concreteVal
		if err := ctx.reconstruct(concreteVal.Elem(), ptrOffset); err != nil {
			return err
		}
	} else {
		concreteVal = reflect.New(concreteType).Elem()
		if err := ctx.reconstruct(concreteVal, int64(ih.Offset)); err != nil {
			return err
		}
	}
	target.Set(concreteVal)
	return nil
}

func (ctx *ReaderContext) reconstructArray(target reflect.Value, offset int64) error {
	t := target.Type()
	elemSize := getBinarySize(t.Elem())
	totalSize := elemSize * int64(target.Len())
	if offset < 0 || offset+totalSize > int64(len(ctx.data)) {
		return fmt.Errorf("array out of bounds")
	}

	// Optimization: Field Elements are just copied bytes
	if t.Elem() == reflect.TypeOf(field.Element{}) {
		srcPtr := unsafe.Pointer(&ctx.data[offset])

		// FIX: Cast uintptr -> unsafe.Pointer
		dstPtr := unsafe.Pointer(target.UnsafeAddr())

		// Now we can cast unsafe.Pointer -> *byte safely
		copy(
			unsafe.Slice((*byte)(dstPtr), int(totalSize)),
			unsafe.Slice((*byte)(srcPtr), int(totalSize)),
		)
		return nil
	}

	currentOff := offset
	for i := 0; i < target.Len(); i++ {
		nextOff, err := ctx.readMapElement(target.Index(i), currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
	}
	return nil
}

func (ctx *ReaderContext) reconstructFieldElement(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(ctx.data) {
		return fmt.Errorf("field element out of bounds")
	}
	srcPtr := unsafe.Pointer(&ctx.data[offset]) // Fixed: Direct index
	dstPtr := unsafe.Pointer(target.UnsafeAddr())

	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

func (ctx *ReaderContext) reconstructPrimitive(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(ctx.data) {
		return fmt.Errorf("primitive out of bounds")
	}

	srcPtr := unsafe.Pointer(&ctx.data[offset])
	// Fix: Cast uintptr -> unsafe.Pointer -> *byte
	dstPtr := unsafe.Pointer(target.UnsafeAddr())

	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

func (ctx *ReaderContext) reconstructStruct(target reflect.Value, offset int64) error {
	currentOff := offset
	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		t := target.Type().Field(i)
		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		if t.Type == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
			ref := *(*Ref)(unsafe.Pointer(&ctx.data[currentOff]))
			currentOff += 8
			if !ref.IsNull() {
				bi := new(big.Int)
				if err := reconstructBigInt(ctx.data, reflect.ValueOf(bi).Elem(), int64(ref)); err != nil {
					return err
				}
				f.Set(reflect.ValueOf(bi))
			}
			continue
		}

		_, hasCustom := CustomRegistry[f.Type()]

		isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})
		if hasCustom || f.Kind() == reflect.Ptr ||
			f.Kind() == reflect.Slice || f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map ||
			f.Kind() == reflect.Func || isBigInt {

			ref := *(*Ref)(unsafe.Pointer(&ctx.data[currentOff]))
			currentOff += 8
			if !ref.IsNull() {
				if err := ctx.reconstruct(f, int64(ref)); err != nil {
					return err
				}
			}
			continue
		}

		binSize := getBinarySize(f.Type())
		if err := ctx.reconstruct(f, currentOff); err != nil {
			return err
		}
		currentOff += binSize
	}
	return nil
}

func (ctx *ReaderContext) readMapElement(target reflect.Value, offset int64) (int64, error) {
	// 1. ADD BOUNDS CHECK HERE
	if offset < 0 || int(offset) >= len(ctx.data) {
		return 0, fmt.Errorf("readMapElement: offset %d out of bounds (len: %d)", offset, len(ctx.data))
	}

	t := target.Type()
	_, hasCustom := CustomRegistry[t]
	isRef := hasCustom || t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.String ||
		t.Kind() == reflect.Interface || t.Kind() == reflect.Map ||
		t.Kind() == reflect.Func ||
		t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) ||
		t == reflect.TypeOf((*frontend.Variable)(nil)).Elem()

	if isRef {
		// 2. Add bounds check for the 8-byte reference read
		if int(offset)+8 > len(ctx.data) {
			return 0, fmt.Errorf("readMapElement: unable to read Ref at offset %d", offset)
		}

		ref := *(*Ref)(unsafe.Pointer(&ctx.data[offset]))
		if !ref.IsNull() {
			if err := ctx.reconstruct(target, int64(ref)); err != nil {
				return 0, err
			}
		}
		return offset + 8, nil
	}

	binSize := getBinarySize(t)

	// 3. Add bounds check for binary data read
	if int(offset)+int(binSize) > len(ctx.data) {
		return 0, fmt.Errorf("readMapElement: unable to read binary data of size %d at offset %d", binSize, offset)
	}

	if err := ctx.reconstruct(target, offset); err != nil {
		return 0, err
	}
	return offset + binSize, nil
}

func reconstructBigInt(data []byte, target reflect.Value, offset int64) error {
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
