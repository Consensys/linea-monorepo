// File: serde/reader.go
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
	data   []byte
	ptrMap map[int64]reflect.Value
}

func Deserialize(b []byte, v any) error {
	if len(b) < int(SizeOf[FileHeader]()) {
		return fmt.Errorf("buffer too small")
	}
	header := (*FileHeader)(unsafe.Pointer(&b[0]))
	if header.Magic != Magic {
		return fmt.Errorf("invalid magic bytes")
	}

	if Ref(header.PayloadOff).IsNull() {
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val.Elem().Set(reflect.Zero(val.Elem().Type()))
		}
		return nil
	}

	if Ref(header.PayloadOff) > Ref(len(b)) {
		return fmt.Errorf("payload offset out of bounds")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	ctx := &ReaderContext{
		data:   b,
		ptrMap: make(map[int64]reflect.Value),
	}

	return ctx.reconstruct(val.Elem(), int64(header.PayloadOff))
}

func (ctx *ReaderContext) reconstruct(target reflect.Value, offset int64) error {

	// 1. Custom Registry
	if handler, ok := CustomRegistry[target.Type()]; ok {
		return handler.Deserialize(ctx, target, offset)
	}

	// 2. Pointers
	if target.Kind() == reflect.Ptr {
		if val, ok := ctx.ptrMap[offset]; ok {
			target.Set(val)
			return nil
		}

		if target.Type() == reflect.TypeOf(&big.Int{}) {
			newItem := reflect.New(target.Type().Elem())
			ctx.ptrMap[offset] = newItem
			if err := reconstructBigInt(ctx.data, newItem.Elem(), offset); err != nil {
				return err
			}
			target.Set(newItem)
			return nil
		}

		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}

		ctx.ptrMap[offset] = target
		return ctx.reconstruct(target.Elem(), offset)
	}

	// 3. Big Ints
	if target.Type() == reflect.TypeOf(big.Int{}) {
		return reconstructBigInt(ctx.data, target, offset)
	}

	// 4. Field Elements
	if target.Type() == reflect.TypeOf(field.Element{}) {
		size := int(target.Type().Size())
		if offset < 0 || int(offset)+size > len(ctx.data) {
			return fmt.Errorf("field element out of bounds")
		}
		basePtr := unsafe.Pointer(&ctx.data[0])
		srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
		dstPtr := target.UnsafeAddr()
		copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
		return nil
	}

	// 5. Slices
	if target.Kind() == reflect.Slice {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
			return fmt.Errorf("slice header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
		if fs.Offset.IsNull() {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target.UnsafeAddr()))
		sh.Data = uintptr(unsafe.Pointer(&ctx.data[fs.Offset]))
		sh.Len = int(fs.Len)
		sh.Cap = int(fs.Cap)
		return nil
	}

	// 6. Strings
	if target.Kind() == reflect.String {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
			return fmt.Errorf("string header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
		if fs.Offset.IsNull() {
			return nil
		}
		strBytes := ctx.data[fs.Offset : int64(fs.Offset)+int64(fs.Len)]
		target.SetString(string(strBytes))
		return nil
	}

	// 7. Interfaces
	if target.Kind() == reflect.Interface {
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
		} else {
			concreteVal = reflect.New(concreteType).Elem()
		}
		if err := ctx.reconstruct(concreteVal, int64(ih.Offset)); err != nil {
			return err
		}
		target.Set(concreteVal)
		return nil
	}

	// 8. Maps
	if target.Kind() == reflect.Map {
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

	// 9. Structs
	if target.Kind() == reflect.Struct {
		return ctx.reconstructStruct(target, offset)
	}

	// 10. Arrays
	if target.Kind() == reflect.Array {
		t := target.Type()
		elemSize := getBinarySize(t.Elem())
		totalSize := elemSize * int64(t.Len())
		if offset < 0 || offset+totalSize > int64(len(ctx.data)) {
			return fmt.Errorf("array out of bounds")
		}
		if t.Elem() == reflect.TypeOf(field.Element{}) {
			srcPtr := unsafe.Pointer(&ctx.data[offset])
			dstPtr := target.UnsafeAddr()
			copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), int(totalSize)), unsafe.Slice((*byte)(srcPtr), int(totalSize)))
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

	// 11. Primitives
	if target.Kind() == reflect.Int {
		if offset < 0 || int(offset)+8 > len(ctx.data) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetInt(val64)
		return nil
	}
	if target.Kind() == reflect.Uint {
		if offset < 0 || int(offset)+8 > len(ctx.data) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetUint(val64)
		return nil
	}
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(ctx.data) {
		return fmt.Errorf("primitive out of bounds")
	}
	basePtr := unsafe.Pointer(&ctx.data[0])
	srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	dstPtr := target.UnsafeAddr()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
	return nil
}

// reconstructStruct is extracted to allow Custom handlers to fallback to standard struct decoding
func (ctx *ReaderContext) reconstructStruct(target reflect.Value, offset int64) error {
	currentOff := offset
	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		t := target.Type().Field(i)
		if !t.IsExported() {
			continue
		}
		if strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		if t.Type == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
			if currentOff+8 > int64(len(ctx.data)) {
				return fmt.Errorf("variable ref out of bounds")
			}
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

		isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})
		if f.Kind() == reflect.Ptr ||
			f.Kind() == reflect.Slice || f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map ||
			f.Kind() == reflect.Func ||
			isBigInt {
			if currentOff+8 > int64(len(ctx.data)) {
				return fmt.Errorf("struct ref field out of bounds")
			}
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
		if currentOff+binSize > int64(len(ctx.data)) {
			return fmt.Errorf("struct inline field out of bounds")
		}
		if err := ctx.reconstruct(f, currentOff); err != nil {
			return err
		}
		currentOff += binSize
	}
	return nil
}

// ... [readMapElement, reconstructBigInt, getBinarySize remain same] ...
func (ctx *ReaderContext) readMapElement(target reflect.Value, offset int64) (int64, error) {
	t := target.Type()
	isRef := t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.String ||
		t.Kind() == reflect.Interface || t.Kind() == reflect.Map ||
		t.Kind() == reflect.Func || // Handled
		t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) ||
		t == reflect.TypeOf((*frontend.Variable)(nil)).Elem()

	if isRef {
		if offset+8 > int64(len(ctx.data)) {
			return 0, fmt.Errorf("map element ref out of bounds")
		}
		ref := *(*Ref)(unsafe.Pointer(&ctx.data[offset]))
		if !ref.IsNull() {
			// Leverage ctx.reconstruct for recursion/dedup
			if err := ctx.reconstruct(target, int64(ref)); err != nil {
				return 0, err
			}
		}
		return offset + 8, nil
	}

	binSize := getBinarySize(t)
	if offset+binSize > int64(len(ctx.data)) {
		return 0, fmt.Errorf("map element inline out of bounds")
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

// func getBinarySize(t reflect.Type) int64 {
// 	if t == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
// 		return 8
// 	}
// 	if t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) {
// 		return 8
// 	}

// 	k := t.Kind()
// 	if k == reflect.Ptr || k == reflect.Slice ||
// 		k == reflect.String || k == reflect.Interface || k == reflect.Map {
// 		return 8
// 	}

// 	if k == reflect.Struct {
// 		if t == reflect.TypeOf(field.Element{}) {
// 			return int64(t.Size())
// 		}
// 		var sum int64
// 		for i := 0; i < t.NumField(); i++ {
// 			f := t.Field(i)
// 			if !f.IsExported() {
// 				continue
// 			}
// 			if strings.Contains(f.Tag.Get("serde"), "omit") {
// 				continue
// 			}
// 			sum += getBinarySize(f.Type)
// 		}
// 		return sum
// 	}

// 	// Array handling (recurse for elements)
// 	if k == reflect.Array {
// 		if t == reflect.TypeOf(field.Element{}) {
// 			return int64(t.Size())
// 		}
// 		elemSize := getBinarySize(t.Elem())
// 		return elemSize * int64(t.Len())
// 	}

// 	if k == reflect.Int || k == reflect.Uint {
// 		return 8
// 	}

// 	return int64(t.Size())
// }
