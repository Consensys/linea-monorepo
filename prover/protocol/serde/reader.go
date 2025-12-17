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
	data   []byte
	ptrMap map[int64]reflect.Value
}

func (ctx *ReaderContext) reconstruct(target reflect.Value, offset int64) error {
	t := target.Type()
	k := target.Kind()

	// 1. Custom Registry
	if k == reflect.Struct || k == reflect.Ptr || k == reflect.Interface {
		if handler, ok := CustomRegistry[t]; ok {
			return handler.Deserialize(ctx, target, offset)
		}
	}

	// 2. Pointers
	if k == reflect.Ptr {
		if val, ok := ctx.ptrMap[offset]; ok {
			if val.Type().AssignableTo(t) {
				target.Set(val)
				return nil
			}
		}

		if t == reflect.TypeOf(&big.Int{}) {
			newItem := reflect.New(t.Elem())
			ctx.ptrMap[offset] = newItem
			if err := reconstructBigInt(ctx.data, newItem.Elem(), offset); err != nil {
				return err
			}
			target.Set(newItem)
			return nil
		}

		newPtr := reflect.New(t.Elem())
		ctx.ptrMap[offset] = newPtr
		target.Set(newPtr)
		return ctx.reconstruct(newPtr.Elem(), offset)
	}

	// 3. Big Ints
	if t == reflect.TypeOf(big.Int{}) {
		return reconstructBigInt(ctx.data, target, offset)
	}

	// 4. Field Elements
	if t == reflect.TypeOf(field.Element{}) {
		size := int(t.Size())
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
	if k == reflect.Slice {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
			return fmt.Errorf("slice header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
		if fs.Offset.IsNull() {
			target.Set(reflect.Zero(t))
			return nil
		}
		// Don't allocate a new slice. Just look at the data already sitting in the mapped file
		// at this specific offset
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target.UnsafeAddr()))
		sh.Data = uintptr(unsafe.Pointer(&ctx.data[fs.Offset]))
		sh.Len = int(fs.Len)
		sh.Cap = int(fs.Cap)
		return nil
	}

	// 6. Strings
	if k == reflect.String {
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
	if k == reflect.Interface {
		if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(ctx.data) {
			return fmt.Errorf("interface header out of bounds")
		}
		ih := (*InterfaceHeader)(unsafe.Pointer(&ctx.data[offset]))
		if ih.Offset.IsNull() {
			target.Set(reflect.Zero(t))
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

	// 8. Maps
	if k == reflect.Map {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(ctx.data) {
			return fmt.Errorf("map header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&ctx.data[offset]))
		if fs.Offset.IsNull() {
			target.Set(reflect.Zero(t))
			return nil
		}
		target.Set(reflect.MakeMapWithSize(t, int(fs.Len)))
		currentOff := int64(fs.Offset)
		keyType := t.Key()
		elemType := t.Elem()
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
	if k == reflect.Struct {
		return ctx.reconstructStruct(target, offset)
	}

	// 10. Arrays
	if k == reflect.Array {
		elemSize := getBinarySize(t.Elem())
		totalSize := elemSize * int64(target.Len())
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
	if k == reflect.Int {
		val64 := *(*int64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetInt(val64)
		return nil
	}
	if k == reflect.Uint {
		val64 := *(*uint64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetUint(val64)
		return nil
	}
	size := int(t.Size())
	basePtr := unsafe.Pointer(&ctx.data[0])
	srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	dstPtr := target.UnsafeAddr()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
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
