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

type Decoder struct {

	// The entire memory-mapped file - the raw binary blob. The reader does not "read" from a stream,
	// rather it jumps around this byte slice using offsets.
	data []byte

	// The Deduplication Table - critical for maintaining referential integrity
	// For example: If object A and object B both point to object C, the file only stores object C once.
	// When the reader encounters the offset for C the first time, it deserializes it and stores the result in ptrMap.
	// When it encounters the offset for C the second time, it grabs the existing object from ptrMap.
	// Benefit: This preserves cycles (A -> B -> A) and saves memory.
	ptrMap map[int64]reflect.Value
}

func (dec *Decoder) decode(target reflect.Value, offset int64) error {
	t := target.Type()

	// 1. Custom Registry
	if handler, ok := CustomRegistry[t]; ok {
		return handler.Deserialize(dec, target, offset)
	}

	// 2. Main Type Switch
	switch target.Kind() {
	case reflect.Ptr:
		return dec.decodePtr(target, offset)
	case reflect.Slice:
		return dec.decodeSlice(target, offset)
	case reflect.String:
		return dec.decodeString(target, offset)
	case reflect.Map:
		return dec.decodeMap(target, offset)
	case reflect.Interface:
		return dec.decodeInterface(target, offset)
	case reflect.Struct:
		// Fallback for field.Element if not in registry
		if t == reflect.TypeOf(field.Element{}) {
			return dec.decodeFieldElement(target, offset)
		}
		return dec.decodeStruct(target, offset)
	case reflect.Array:
		return dec.decodeArray(target, offset)
	case reflect.Int, reflect.Int64:
		if offset+8 > int64(len(dec.data)) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&dec.data[offset]))
		target.SetInt(val64)
		return nil
	case reflect.Uint, reflect.Uint64:
		if offset+8 > int64(len(dec.data)) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&dec.data[offset]))
		target.SetUint(val64)
		return nil
	default:
		// Generic copy for other primitives (bool, byte, etc)
		return dec.decodePrimitive(target, offset)
	}
}

// --- Helpers ---

func (dec *Decoder) decodePtr(target reflect.Value, offset int64) error {
	t := target.Type()

	// Check Cache
	if val, ok := dec.ptrMap[offset]; ok {
		if val.Type().AssignableTo(t) {
			target.Set(val)
			return nil
		}
	}

	// Generic Pointer - allocates fresh memory for the type the pointer points to (the "element").
	// i.e newPtr is of type `*t` where type `t` is the target's type - see first line, and allocates
	// memory to store type `t`, and saves it to the ptrMap (deduplication purposes).
	newPtr := reflect.New(t.Elem())
	dec.ptrMap[offset] = newPtr

	// Reconstruct recursively to actually fill that newly allocated memory with the data found at that
	// offset in the file.
	target.Set(newPtr)
	return dec.decode(newPtr.Elem(), offset)
}

func (dec *Decoder) decodeSlice(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("slice header out of bounds")
	}

	// Read the File Header First by looking at the offset to find a `FileSlice` struct.
	// This is a small header stored in the binary file that tells us three things:
	// Where the actual data is (the Offset).
	// How much data there is (the Len and Cap)
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))

	// If the fs.Offset is marked as null, it sets the target slice to its zero value (a nil slice)
	// and exits early.
	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// Verify bounds of the underlying data start
	if int(fs.Offset) >= len(dec.data) {
		return fmt.Errorf("slice data offset out of bounds")
	}

	// ZERO-COPY (The pointer "Swizzling" technique) SLICE
	// 1. Get the pointer to the raw data inside the memory map
	// i.e. calculate the memory address of the data inside the file
	dataPtr := unsafe.Pointer(&dec.data[fs.Offset])

	// 2. Manually construct the slice header to point to this data
	// We cast the target's address to a struct that mimics the Go slice layout.
	// This works for all slice types.
	sh := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(target.UnsafeAddr()))

	// Manually overwrite the slice fields to make the Go runtime think our slice is a normal slice,
	// but its underlying data is actually the memory-mapped file itself. This unsafe usage is permitted.
	sh.Data = uintptr(dataPtr)
	sh.Len = int(fs.Len)
	sh.Cap = int(fs.Cap)
	return nil
}

func (dec *Decoder) decodeString(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("string header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}

	start := int64(fs.Offset)
	end := start + int64(fs.Len)
	if start < 0 || end > int64(len(dec.data)) {
		return fmt.Errorf("string content out of bounds")
	}

	// Zero-copy string construction - we'd need unsafe.String (Go 1.20+)
	// NOTE Std. string(bytes) copies.
	target.SetString(unsafe.String(&dec.data[start], fs.Len))
	return nil
}

func (dec *Decoder) decodeMap(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("map header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
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
		nextOff, err := dec.decodeMapElement(newKey, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff

		newVal := reflect.New(elemType).Elem()
		nextOff, err = dec.decodeMapElement(newVal, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
		target.SetMapIndex(newKey, newVal)
	}
	return nil
}

func (dec *Decoder) decodeInterface(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(dec.data) {
		return fmt.Errorf("interface header out of bounds")
	}
	ih := (*InterfaceHeader)(unsafe.Pointer(&dec.data[offset]))
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
		if val, ok := dec.ptrMap[ptrOffset]; ok {
			if val.Type().AssignableTo(concreteType) {
				target.Set(val)
				return nil
			}
		}
		dec.ptrMap[ptrOffset] = concreteVal
		if err := dec.decode(concreteVal.Elem(), ptrOffset); err != nil {
			return err
		}
	} else {
		concreteVal = reflect.New(concreteType).Elem()
		if err := dec.decode(concreteVal, int64(ih.Offset)); err != nil {
			return err
		}
	}
	target.Set(concreteVal)
	return nil
}

func (dec *Decoder) decodeArray(target reflect.Value, offset int64) error {
	t := target.Type()
	elemSize := getBinarySize(t.Elem())
	totalSize := elemSize * int64(target.Len())
	if offset < 0 || offset+totalSize > int64(len(dec.data)) {
		return fmt.Errorf("array out of bounds")
	}

	// Optimization: Field Elements are just copied bytes
	if t.Elem() == reflect.TypeOf(field.Element{}) {
		srcPtr := unsafe.Pointer(&dec.data[offset])

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
		nextOff, err := dec.decodeMapElement(target.Index(i), currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
	}
	return nil
}

func (dec *Decoder) decodeFieldElement(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(dec.data) {
		return fmt.Errorf("field element out of bounds")
	}
	srcPtr := unsafe.Pointer(&dec.data[offset]) // Fixed: Direct index
	dstPtr := unsafe.Pointer(target.UnsafeAddr())

	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

func (dec *Decoder) decodePrimitive(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(dec.data) {
		return fmt.Errorf("primitive out of bounds")
	}

	srcPtr := unsafe.Pointer(&dec.data[offset])
	// Fix: Cast uintptr -> unsafe.Pointer -> *byte
	dstPtr := unsafe.Pointer(target.UnsafeAddr())

	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

func (dec *Decoder) decodeStruct(target reflect.Value, offset int64) error {
	currentOff := offset
	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		t := target.Type().Field(i)
		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		if t.Type == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
			ref := *(*Ref)(unsafe.Pointer(&dec.data[currentOff]))
			currentOff += 8
			if !ref.IsNull() {
				bi := new(big.Int)
				if err := decodeBigInt(dec.data, reflect.ValueOf(bi).Elem(), int64(ref)); err != nil {
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

			ref := *(*Ref)(unsafe.Pointer(&dec.data[currentOff]))
			currentOff += 8
			if !ref.IsNull() {
				if err := dec.decode(f, int64(ref)); err != nil {
					return err
				}
			}
			continue
		}

		binSize := getBinarySize(f.Type())
		if err := dec.decode(f, currentOff); err != nil {
			return err
		}
		currentOff += binSize
	}
	return nil
}

func (dec *Decoder) decodeMapElement(target reflect.Value, offset int64) (int64, error) {
	// 1. ADD BOUNDS CHECK HERE
	if offset < 0 || int(offset) >= len(dec.data) {
		return 0, fmt.Errorf("readMapElement: offset %d out of bounds (len: %d)", offset, len(dec.data))
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
		if int(offset)+8 > len(dec.data) {
			return 0, fmt.Errorf("readMapElement: unable to read Ref at offset %d", offset)
		}

		ref := *(*Ref)(unsafe.Pointer(&dec.data[offset]))
		if !ref.IsNull() {
			if err := dec.decode(target, int64(ref)); err != nil {
				return 0, err
			}
		}
		return offset + 8, nil
	}

	binSize := getBinarySize(t)

	// 3. Add bounds check for binary data read
	if int(offset)+int(binSize) > len(dec.data) {
		return 0, fmt.Errorf("readMapElement: unable to read binary data of size %d at offset %d", binSize, offset)
	}

	if err := dec.decode(target, offset); err != nil {
		return 0, err
	}
	return offset + binSize, nil
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
