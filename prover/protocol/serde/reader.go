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
	fmt.Printf("[DEBUG] Start Deserialize. Buffer Len: %d\n", len(b))

	if len(b) < int(SizeOf[FileHeader]()) {
		return fmt.Errorf("buffer too small: have %d, need %d", len(b), SizeOf[FileHeader]())
	}
	header := (*FileHeader)(unsafe.Pointer(&b[0]))

	fmt.Printf("[DEBUG] Header Magic: %x (Expected: %x)\n", header.Magic, Magic)
	fmt.Printf("[DEBUG] Payload Offset: %d\n", header.PayloadOff)

	if header.Magic != Magic {
		return fmt.Errorf("invalid magic bytes")
	}

	if Ref(header.PayloadOff).IsNull() {
		fmt.Println("[DEBUG] Payload Offset is Null. Returning zero value.")
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val.Elem().Set(reflect.Zero(val.Elem().Type()))
		}
		return nil
	}

	if Ref(header.PayloadOff) > Ref(len(b)) {
		return fmt.Errorf("payload offset %d out of bounds (len: %d)", header.PayloadOff, len(b))
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	ctx := &ReaderContext{
		data:   b,
		ptrMap: make(map[int64]reflect.Value),
	}

	fmt.Printf("[DEBUG] Starting reconstruction of Root Type: %v at Offset: %d\n", val.Elem().Type(), header.PayloadOff)
	return ctx.reconstruct(val.Elem(), int64(header.PayloadOff))
}

func (ctx *ReaderContext) reconstruct(target reflect.Value, offset int64) error {
	t := target.Type()
	k := target.Kind()

	// Logging Trace
	fmt.Printf(">> [Reconstruct] Type: %-20v | Kind: %-10v | Offset: %d\n", t, k, offset)

	// 1. Custom Registry
	if k == reflect.Struct || k == reflect.Ptr || k == reflect.Interface {
		if handler, ok := CustomRegistry[t]; ok {
			fmt.Printf("   -> Handled by Custom Registry: %v\n", t)
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
				return fmt.Errorf("ptr bigInt error: %w", err)
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
			return fmt.Errorf("field element out of bounds: off=%d size=%d len=%d", offset, size, len(ctx.data))
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
		fmt.Printf("   -> Slice Header: Offset=%d, Len=%d\n", fs.Offset, fs.Len)

		if fs.Offset.IsNull() {
			target.Set(reflect.Zero(t))
			return nil
		}
		// Validate slice data bounds
		// Note: We don't check full data range here because underlying types might be pointers
		// but we can check if the start exists.
		if int(fs.Offset) > len(ctx.data) {
			return fmt.Errorf("slice data start %d out of bounds %d", fs.Offset, len(ctx.data))
		}

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
		end := int64(fs.Offset) + int64(fs.Len)
		if end > int64(len(ctx.data)) {
			return fmt.Errorf("string data out of bounds: end=%d, len=%d", end, len(ctx.data))
		}
		strBytes := ctx.data[fs.Offset:end]
		target.SetString(string(strBytes))
		return nil
	}

	// 7. Interfaces
	if k == reflect.Interface {
		if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(ctx.data) {
			return fmt.Errorf("interface header out of bounds")
		}
		ih := (*InterfaceHeader)(unsafe.Pointer(&ctx.data[offset]))
		fmt.Printf("   -> Interface ID: %d, Offset: %d\n", ih.TypeID, ih.Offset)

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

		fmt.Printf("   -> Map Len: %d, StartOff: %d\n", fs.Len, currentOff)

		for i := 0; i < int(fs.Len); i++ {
			newKey := reflect.New(keyType).Elem()
			nextOff, err := ctx.readMapElement(newKey, currentOff)
			if err != nil {
				return fmt.Errorf("map key error at idx %d: %w", i, err)
			}
			currentOff = nextOff
			newVal := reflect.New(elemType).Elem()
			nextOff, err = ctx.readMapElement(newVal, currentOff)
			if err != nil {
				return fmt.Errorf("map val error at idx %d: %w", i, err)
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
			return fmt.Errorf("array out of bounds: off=%d total=%d len=%d", offset, totalSize, len(ctx.data))
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
		if offset+8 > int64(len(ctx.data)) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetInt(val64)
		return nil
	}
	if k == reflect.Uint {
		if offset+8 > int64(len(ctx.data)) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&ctx.data[offset]))
		target.SetUint(val64)
		return nil
	}
	size := int(t.Size())
	if offset+int64(size) > int64(len(ctx.data)) {
		return fmt.Errorf("primitive %v out of bounds", t)
	}
	basePtr := unsafe.Pointer(&ctx.data[0])
	srcPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	dstPtr := target.UnsafeAddr()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), size), unsafe.Slice((*byte)(srcPtr), size))
	return nil
}

func (ctx *ReaderContext) reconstructStruct(target reflect.Value, offset int64) error {
	currentOff := offset
	fmt.Printf("   [Struct] %v (NumFields: %d) starting at %d\n", target.Type(), target.NumField(), offset)

	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		t := target.Type().Field(i)

		if !t.IsExported() {
			continue
		}
		if strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		fmt.Printf("      -> Field[%d]: %-15s | Type: %-15v | CurrentOff: %d | DataLeft: %d\n",
			i, t.Name, t.Type, currentOff, int64(len(ctx.data))-currentOff)

		if t.Type == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
			if currentOff+8 > int64(len(ctx.data)) {
				return fmt.Errorf("frontend.Variable ref OOB")
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

		_, hasCustom := CustomRegistry[f.Type()]

		isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})
		if hasCustom || f.Kind() == reflect.Ptr ||
			f.Kind() == reflect.Slice || f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map ||
			f.Kind() == reflect.Func || isBigInt {

			if currentOff+8 > int64(len(ctx.data)) {
				return fmt.Errorf("ref pointer OOB for field %s", t.Name)
			}

			ref := *(*Ref)(unsafe.Pointer(&ctx.data[currentOff]))
			currentOff += 8
			if !ref.IsNull() {
				if err := ctx.reconstruct(f, int64(ref)); err != nil {
					return fmt.Errorf("field %s reconstruct error: %w", t.Name, err)
				}
			}
			continue
		}

		binSize := getBinarySize(f.Type())
		if err := ctx.reconstruct(f, currentOff); err != nil {
			return fmt.Errorf("field %s inline error: %w", t.Name, err)
		}
		currentOff += binSize
	}
	return nil
}

func (ctx *ReaderContext) readMapElement(target reflect.Value, offset int64) (int64, error) {
	t := target.Type()
	_, hasCustom := CustomRegistry[t]
	isRef := hasCustom || t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.String ||
		t.Kind() == reflect.Interface || t.Kind() == reflect.Map ||
		t.Kind() == reflect.Func ||
		t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) ||
		t == reflect.TypeOf((*frontend.Variable)(nil)).Elem()

	if isRef {
		if offset+8 > int64(len(ctx.data)) {
			return 0, fmt.Errorf("map element ref OOB")
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
	dataStart := int64(fs.Offset) + 1 // +1 for sign byte
	dataLen := int64(fs.Len)
	if dataStart+dataLen > int64(len(data)) {
		return fmt.Errorf("bigint data out of bounds: start=%d len=%d buf=%d", dataStart, dataLen, len(data))
	}
	bytes := data[dataStart : dataStart+dataLen]
	bi := new(big.Int).SetBytes(bytes)

	// Read sign bit
	signByte := data[int64(fs.Offset)]
	if signByte == 1 {
		bi.Neg(bi)
	}
	target.Set(reflect.ValueOf(*bi))
	return nil
}
