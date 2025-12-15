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

func Serialize(v any) ([]byte, error) {
	w := NewWriter()
	_ = w.Write(FileHeader{})
	rootOff, err := linearize(w, reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	finalHeader := FileHeader{
		Magic:       Magic,
		Version:     1,
		PayloadType: 0,
		PayloadOff:  int64(rootOff),
		DataSize:    w.offset,
	}
	b := w.buf.Bytes()
	*(*FileHeader)(unsafe.Pointer(&b[0])) = finalHeader
	return b, nil
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
		return nil
	}
	if Ref(header.PayloadOff) >= Ref(len(b)) {
		return fmt.Errorf("payload offset out of bounds")
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}
	return reconstruct(b, val.Elem(), int64(header.PayloadOff))
}

func getBinarySize(t reflect.Type) int64 {
	if t == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
		return 8
	}
	if t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) {
		return 8
	}

	k := t.Kind()
	if k == reflect.Ptr || k == reflect.Slice || k == reflect.String || k == reflect.Interface || k == reflect.Map {
		return 8
	}

	if k == reflect.Struct {
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}
		var sum int64
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if strings.Contains(f.Tag.Get("serde"), "omit") {
				continue
			}
			sum += getBinarySize(f.Type)
		}
		return sum
	}

	// Array handling (recurse for elements)
	if k == reflect.Array {
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}
		elemSize := getBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}

	if k == reflect.Int || k == reflect.Uint {
		return 8
	}

	return int64(t.Size())
}

func reconstruct(data []byte, target reflect.Value, offset int64) error {
	// 1. Pointers
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		if target.Type() == reflect.TypeOf(&big.Int{}) {
			return reconstructBigInt(data, target.Elem(), offset)
		}
		return reconstruct(data, target.Elem(), offset)
	}

	// 2. Big Ints
	if target.Type() == reflect.TypeOf(big.Int{}) {
		return reconstructBigInt(data, target, offset)
	}

	// 3. Field Elements
	if target.Type() == reflect.TypeOf(field.Element{}) {
		size := int(target.Type().Size())
		if offset < 0 || int(offset)+size > len(data) {
			return fmt.Errorf("field element out of bounds")
		}
		basePtr := unsafe.Pointer(&data[0])
		srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
		dstPtr := target.UnsafeAddr()
		copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
		return nil
	}

	// 4. Slices
	if target.Kind() == reflect.Slice {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(data) {
			return fmt.Errorf("slice header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&data[offset]))
		if fs.Offset.IsNull() {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}
		if fs.Offset < 0 || int64(fs.Offset) >= int64(len(data)) {
			return fmt.Errorf("slice data start out of bounds")
		}
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target.UnsafeAddr()))
		sh.Data = uintptr(unsafe.Pointer(&data[fs.Offset]))
		sh.Len = int(fs.Len)
		sh.Cap = int(fs.Cap)
		return nil
	}

	// 5. Strings
	if target.Kind() == reflect.String {
		if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(data) {
			return fmt.Errorf("string header out of bounds")
		}
		fs := (*FileSlice)(unsafe.Pointer(&data[offset]))
		if fs.Offset.IsNull() {
			return nil
		}
		if int64(fs.Offset)+int64(fs.Len) > int64(len(data)) {
			return fmt.Errorf("string data out of bounds")
		}
		strBytes := data[fs.Offset : int64(fs.Offset)+int64(fs.Len)]
		target.SetString(string(strBytes))
		return nil
	}

	// 6. Interfaces
	if target.Kind() == reflect.Interface {
		if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(data) {
			return fmt.Errorf("interface header out of bounds")
		}
		ih := (*InterfaceHeader)(unsafe.Pointer(&data[offset]))
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
		if err := reconstruct(data, concreteVal, int64(ih.Offset)); err != nil {
			return err
		}
		target.Set(concreteVal)
		return nil
	}

	// 7. Maps
	if target.Kind() == reflect.Map {
		return nil
	}

	// 8. Structs
	if target.Kind() == reflect.Struct {
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
				if currentOff+8 > int64(len(data)) {
					return fmt.Errorf("variable ref out of bounds")
				}
				ref := *(*Ref)(unsafe.Pointer(&data[currentOff]))
				currentOff += 8
				if !ref.IsNull() {
					bi := new(big.Int)
					if err := reconstructBigInt(data, reflect.ValueOf(bi).Elem(), int64(ref)); err != nil {
						return err
					}
					f.Set(reflect.ValueOf(bi))
				}
				continue
			}

			isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})
			if f.Kind() == reflect.Ptr || f.Kind() == reflect.Slice || f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map || isBigInt {
				if currentOff+8 > int64(len(data)) {
					return fmt.Errorf("struct ref field out of bounds")
				}
				ref := *(*Ref)(unsafe.Pointer(&data[currentOff]))
				currentOff += 8
				if !ref.IsNull() {
					if f.Kind() == reflect.Ptr && !isBigInt {
						newObj := reflect.New(f.Type().Elem())
						if err := reconstruct(data, newObj.Elem(), int64(ref)); err != nil {
							return err
						}
						f.Set(newObj)
					} else {
						if err := reconstruct(data, f, int64(ref)); err != nil {
							return err
						}
					}
				}
				continue
			}

			binSize := getBinarySize(f.Type())
			if currentOff+binSize > int64(len(data)) {
				return fmt.Errorf("struct inline field out of bounds")
			}
			if err := reconstruct(data, f, currentOff); err != nil {
				return err
			}
			currentOff += binSize
		}
		return nil
	}

	// 9. Arrays
	// FIX: Explicit Array handling with correct unsafe.Pointer casting
	if target.Kind() == reflect.Array {
		binSize := getBinarySize(target.Type())
		if offset < 0 || int64(offset)+binSize > int64(len(data)) {
			return fmt.Errorf("array out of bounds")
		}
		srcPtr := unsafe.Pointer(&data[offset])
		dstPtr := target.UnsafeAddr()

		// FIX: dstPtr (uintptr) -> unsafe.Pointer -> *byte
		copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), int(binSize)), unsafe.Slice((*byte)(srcPtr), int(binSize)))
		return nil
	}

	// 10. Primitives
	if target.Kind() == reflect.Int {
		if offset < 0 || int(offset)+8 > len(data) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&data[offset]))
		target.SetInt(val64)
		return nil
	}
	if target.Kind() == reflect.Uint {
		if offset < 0 || int(offset)+8 > len(data) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&data[offset]))
		target.SetUint(val64)
		return nil
	}

	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(data) {
		return fmt.Errorf("primitive out of bounds")
	}
	basePtr := unsafe.Pointer(&data[0])
	srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	dstPtr := target.UnsafeAddr()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
	return nil
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
