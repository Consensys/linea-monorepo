// The Writer takes a complex web of Go objects and flattens them into a single contiguous block of memory.
package serde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

//	Writer struct uses a bytes.Buffer to accumulate all data into one long string of bytes.
//
// This ensures that when you write to disk, the data is physically next to each other, which is much
// faster for hardware to read.
type Writer struct {
	buf    *bytes.Buffer
	offset int64
	ptrMap map[uintptr]Ref
}

func NewWriter() *Writer {
	return &Writer{
		buf:    new(bytes.Buffer),
		offset: 0,
		ptrMap: make(map[uintptr]Ref),
	}
}

func (w *Writer) Write(data any) int64 {
	start := w.offset
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Int:
		data = int64(v.Int())
	case reflect.Uint:
		data = uint64(v.Uint())
	}
	if err := binary.Write(w.buf, binary.LittleEndian, data); err != nil {
		panic(fmt.Errorf("binary.Write failed for type %T: %w", data, err))
	}
	w.offset += int64(binary.Size(data))
	return start
}

func (w *Writer) WriteBytes(b []byte) int64 {
	start := w.offset
	w.buf.Write(b)
	w.offset += int64(len(b))
	return start
}

func (w *Writer) WriteSlice(v reflect.Value) FileSlice {
	if v.Len() == 0 {
		return FileSlice{0, 0, 0}
	}
	totalBytes := v.Len() * int(v.Type().Elem().Size())
	dataPtr := unsafe.Pointer(v.Pointer())
	dataBytes := unsafe.Slice((*byte)(dataPtr), totalBytes)
	offset := w.offset
	w.buf.Write(dataBytes)
	w.offset += int64(totalBytes)
	return FileSlice{
		Offset: Ref(offset),
		Len:    int64(v.Len()),
		Cap:    int64(v.Cap()),
	}
}

func linearize(w *Writer, v reflect.Value) (Ref, error) {
	if !v.IsValid() {
		return 0, nil
	}

	// 1. Deduplication (Pointers)
	var ptrAddr uintptr
	isPtr := v.Kind() == reflect.Ptr
	if isPtr {
		if v.IsNil() {
			return 0, nil
		}
		ptrAddr = v.Pointer()

		if ref, ok := w.ptrMap[ptrAddr]; ok {
			return ref, nil
		}

		// Pre-map to handle cycles.
		// Relies on Structs using "Reserve and Patch" to honor this offset.
		w.ptrMap[ptrAddr] = Ref(w.offset)
	}

	// 2. Serialize
	ref, err := linearizeInternal(w, v)
	if err != nil {
		return 0, err
	}

	// 3. Validation / Update
	if isPtr {
		if ref != w.ptrMap[ptrAddr] {
			w.ptrMap[ptrAddr] = ref
		}
	}

	return ref, nil
}

func linearizeInternal(w *Writer, v reflect.Value) (Ref, error) {
	if handler, ok := CustomRegistry[v.Type()]; ok {
		return handler.Serialize(w, v)
	}

	if v.Kind() == reflect.Ptr {
		if v.Type() == reflect.TypeOf(&big.Int{}) {
			return writeBigInt(w, v.Interface().(*big.Int))
		}
		return linearize(w, v.Elem())
	}

	if v.Kind() == reflect.Func {
		return 0, nil
	}

	if v.Type() == reflect.TypeOf(field.Element{}) {
		off := w.Write(v.Interface())
		return Ref(off), nil
	}

	if v.Type() == reflect.TypeOf(big.Int{}) {
		if v.CanAddr() {
			return writeBigInt(w, v.Addr().Interface().(*big.Int))
		}
		bi := v.Interface().(big.Int)
		return writeBigInt(w, &bi)
	}

	if v.Kind() == reflect.Slice {
		if v.IsNil() {
			return 0, nil
		}
		fs := w.WriteSlice(v)
		off := w.Write(fs)
		return Ref(off), nil
	}

	if v.Kind() == reflect.String {
		str := v.String()
		dataOff := w.WriteBytes([]byte(str))
		fs := FileSlice{Offset: Ref(dataOff), Len: int64(len(str)), Cap: int64(len(str))}
		off := w.Write(fs)
		return Ref(off), nil
	}

	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return 0, nil
		}
		concreteVal := v.Elem()
		dataOff, err := linearize(w, concreteVal)
		if err != nil {
			return 0, err
		}
		baseType := concreteVal.Type()
		indirection := 0
		for baseType.Kind() == reflect.Ptr {
			if _, ok := TypeToID[baseType]; ok {
				break
			}
			baseType = baseType.Elem()
			indirection++
		}
		if _, ok := TypeToID[baseType]; !ok {
			if _, okPtr := TypeToID[reflect.PointerTo(baseType)]; okPtr {
				_ = okPtr
			}
		}
		typeID, ok := TypeToID[baseType]
		if !ok {
			return 0, fmt.Errorf("unregistered type: %v", concreteVal.Type())
		}
		ih := InterfaceHeader{
			TypeID:      typeID,
			Indirection: uint8(indirection),
			Offset:      dataOff,
		}
		off := w.Write(ih)
		return Ref(off), nil
	}

	if v.Kind() == reflect.Map {
		if v.IsNil() {
			return 0, nil
		}
		var bodyBuf bytes.Buffer
		iter := v.MapRange()
		count := 0
		for iter.Next() {
			if err := writeMapElement(w, iter.Key(), &bodyBuf); err != nil {
				return 0, err
			}
			if err := writeMapElement(w, iter.Value(), &bodyBuf); err != nil {
				return 0, err
			}
			count++
		}
		dataOff := w.WriteBytes(bodyBuf.Bytes())
		fs := FileSlice{Offset: Ref(dataOff), Len: int64(count), Cap: int64(count)}
		off := w.Write(fs)
		return Ref(off), nil
	}

	// Structs: Reserve and Patch
	if v.Kind() == reflect.Struct {
		size := getBinarySize(v.Type())
		startOffset := w.offset

		zeros := make([]byte, size)
		w.WriteBytes(zeros)

		if err := patchStructBody(w, v, startOffset); err != nil {
			return 0, err
		}
		return Ref(startOffset), nil
	}

	off := w.Write(v.Interface())
	return Ref(off), nil
}

func patchStructBody(w *Writer, v reflect.Value, startOffset int64) error {
	currentFieldOff := int64(0)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)

		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		if t.Type == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
			bi := toBigInt(f)
			var ref Ref
			if bi != nil {
				var err error
				ref, err = writeBigInt(w, bi)
				if err != nil {
					return err
				}
			}
			patchBytes(w, startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}

		_, hasCustom := CustomRegistry[t.Type]
		isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})

		if hasCustom || f.Kind() == reflect.Ptr || f.Kind() == reflect.Slice ||
			f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map ||
			f.Kind() == reflect.Func || isBigInt {

			ref, err := linearize(w, f)
			if err != nil {
				return err
			}
			patchBytes(w, startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}

		if f.Kind() == reflect.Struct {
			if f.Type() == reflect.TypeOf(field.Element{}) {
				patchBytes(w, startOffset+currentFieldOff, f.Interface())
				currentFieldOff += int64(f.Type().Size())
			} else {
				if err := patchStructBody(w, f, startOffset+currentFieldOff); err != nil {
					return err
				}
				currentFieldOff += getBinarySize(f.Type())
			}
			continue
		}

		if f.Kind() == reflect.Array {
			if f.Type() == reflect.TypeOf(field.Element{}) {
				patchBytes(w, startOffset+currentFieldOff, f.Interface())
				currentFieldOff += int64(f.Type().Size())
			} else {
				if err := patchArray(w, f, startOffset+currentFieldOff); err != nil {
					return err
				}
				currentFieldOff += getBinarySize(f.Type())
			}
			continue
		}

		var val any
		k := f.Kind()

		switch k {
		case reflect.Int:
			val = int64(f.Int())
		case reflect.Uint:
			val = uint64(f.Uint())
		default:
			val = f.Interface()
		}

		patchBytes(w, startOffset+currentFieldOff, val)
		currentFieldOff += int64(binary.Size(val))
	}
	return nil
}

func patchArray(w *Writer, v reflect.Value, startOffset int64) error {
	currentElemOff := int64(0)
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// 1. Nested Arrays (Recursion)
		if elem.Kind() == reflect.Array {
			if err := patchArray(w, elem, startOffset+currentElemOff); err != nil {
				return err
			}
			currentElemOff += elemBinSize
			continue
		}

		// 2. Reference Types / Custom / Interfaces
		_, hasCustom := CustomRegistry[elemType]
		isBigInt := elemType == reflect.TypeOf(big.Int{}) || elemType == reflect.TypeOf(&big.Int{})

		if hasCustom || elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Slice ||
			elem.Kind() == reflect.String || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Map ||
			isBigInt {

			ref, err := linearize(w, elem)
			if err != nil {
				return err
			}
			patchBytes(w, startOffset+currentElemOff, ref)
			currentElemOff += 8 // Refs are always 8 bytes
			continue
		}

		// 3. Structs
		if elem.Kind() == reflect.Struct {
			if elemType == reflect.TypeOf(field.Element{}) {
				patchBytes(w, startOffset+currentElemOff, elem.Interface())
			} else {
				if err := patchStructBody(w, elem, startOffset+currentElemOff); err != nil {
					return err
				}
			}
			currentElemOff += elemBinSize
			continue
		}

		// 4. Primitives
		var val any
		k := elem.Kind()

		switch k {
		case reflect.Int:
			val = int64(elem.Int())
		case reflect.Uint:
			val = uint64(elem.Uint())
		default:
			val = elem.Interface()
		}

		patchBytes(w, startOffset+currentElemOff, val)
		currentElemOff += elemBinSize
	}
	return nil
}

func patchBytes(w *Writer, offset int64, v any) {
	var tmp bytes.Buffer
	binary.Write(&tmp, binary.LittleEndian, v)
	encoded := tmp.Bytes()

	bufSlice := w.buf.Bytes()
	if int(offset)+len(encoded) > len(bufSlice) {
		panic(fmt.Errorf("patch out of bounds"))
	}
	copy(bufSlice[offset:], encoded)
}

func writeBigInt(w *Writer, b *big.Int) (Ref, error) {
	if b == nil {
		return 0, nil
	}
	bytes := b.Bytes()
	sign := int64(0)
	if b.Sign() < 0 {
		sign = 1
	}
	start := w.offset
	w.buf.WriteByte(byte(sign))
	w.buf.Write(bytes)
	w.offset += int64(1 + len(bytes))
	fs := FileSlice{Offset: Ref(start), Len: int64(len(bytes)), Cap: sign}
	off := w.Write(fs)
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

func writeMapElement(w *Writer, v reflect.Value, buf *bytes.Buffer) error {
	t := v.Type()
	if t == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
		bi := toBigInt(v)
		if bi == nil {
			binary.Write(buf, binary.LittleEndian, Ref(0))
		} else {
			ref, err := writeBigInt(w, bi)
			if err != nil {
				return err
			}
			binary.Write(buf, binary.LittleEndian, ref)
		}
		return nil
	}

	_, hasCustom := CustomRegistry[t]

	isRef := hasCustom || v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice ||
		v.Kind() == reflect.String || v.Kind() == reflect.Interface || v.Kind() == reflect.Map ||
		v.Kind() == reflect.Func ||
		t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{})

	if isRef {
		ref, err := linearize(w, v)
		if err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, ref)
	}

	if v.Kind() == reflect.Struct {
		if t == reflect.TypeOf(field.Element{}) {
			return binary.Write(buf, binary.LittleEndian, v.Interface())
		}
		return linearizeStructBody(w, v, buf)
	}

	if v.Kind() == reflect.Array {
		if t == reflect.TypeOf(field.Element{}) {
			return binary.Write(buf, binary.LittleEndian, v.Interface())
		}
		for j := 0; j < v.Len(); j++ {
			if err := writeMapElement(w, v.Index(j), buf); err != nil {
				return fmt.Errorf("array element write failed: %w", err)
			}
		}
		return nil
	}

	var val any
	k := v.Kind()
	if k == reflect.Int {
		val = int64(v.Int())
	} else if k == reflect.Uint {
		val = uint64(v.Uint())
	} else {
		val = v.Interface()
	}

	return binary.Write(buf, binary.LittleEndian, val)
}

func linearizeStructBody(w *Writer, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		_, hasCustom := CustomRegistry[t.Type]
		isBigInt := f.Type() == reflect.TypeOf(big.Int{}) || f.Type() == reflect.TypeOf(&big.Int{})

		if hasCustom || f.Kind() == reflect.Ptr || f.Kind() == reflect.Slice ||
			f.Kind() == reflect.String || f.Kind() == reflect.Interface || f.Kind() == reflect.Map ||
			f.Kind() == reflect.Func || isBigInt {
			ref, err := linearize(w, f)
			if err != nil {
				return err
			}
			binary.Write(buf, binary.LittleEndian, ref)
			continue
		}

		if f.Kind() == reflect.Struct {
			if f.Type() == reflect.TypeOf(field.Element{}) {
				binary.Write(buf, binary.LittleEndian, f.Interface())
			} else {
				if err := linearizeStructBody(w, f, buf); err != nil {
					return err
				}
			}
			continue
		}

		// For Array in Maps, recursing linearizeStructBodyOld isn't quite right for arrays,
		// but typically Maps don't hold Arrays of complex structs in these circuits.
		// If needed, an inner loop here similar to writeMapElement's array handler would be added.

		var val any
		k := f.Kind()
		if k == reflect.Int {
			val = int64(f.Int())
		} else if k == reflect.Uint {
			val = uint64(f.Uint())
		} else {
			val = f.Interface()
		}
		binary.Write(buf, binary.LittleEndian, val)
	}
	return nil
}
