// File: serde/writer.go
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

// Writer: Holds the current encoding/serialization state
type Writer struct {
	// The growing array of bytes (the "Heap").
	buf *bytes.Buffer

	// The write cursor. Points to the end of the buffer.
	offset int64

	// Maps a Go RAM address (uintptr -integer big enough to hold a mem. address
	// for low-level pointer arithmetic) to a File Offset (Ref).
	// Keeps track of memory addresses that have already been seen before.
	ptrMap map[uintptr]Ref
}

func NewWriter() *Writer {
	return &Writer{
		buf:    new(bytes.Buffer),
		offset: 0,
		ptrMap: make(map[uintptr]Ref), // Initialize the deduplication map.
	}
}

// Write writes the given data to the buffer and returns
// the start offset (beginning of cursor) at which the data was written
func (w *Writer) Write(data any) int64 {
	start := w.offset
	v := reflect.ValueOf(data)

	// Normalize `int` and `uint` to `int64` and `uint64`
	// since Go native `int` size depends on the CPU-architecture (32-but vs 64-bit).
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

// WriteBytes writes the given bytes to the buffer and returns
// the start offset (beginning of cursor) at which the bytes were written
func (w *Writer) WriteBytes(b []byte) int64 {
	start := w.offset
	w.buf.Write(b)
	w.offset += int64(len(b))
	return start
}

// Patch jumps back to specific offset (written/reserved earlier) and overwrites
// the data (reserved with zeros) with the actual data
func (w *Writer) Patch(offset int64, v any) {
	var tmp bytes.Buffer
	val := v
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Int {
		val = int64(rv.Int())
	} else if rv.Kind() == reflect.Uint {
		val = uint64(rv.Uint())
	}

	binary.Write(&tmp, binary.LittleEndian, val)
	encoded := tmp.Bytes()
	bufSlice := w.buf.Bytes()

	// Sanity check: sum of offset and encoded length should not exceed the buffer length
	if int(offset)+len(encoded) > len(bufSlice) {
		panic(fmt.Errorf("patch out of bounds"))
	}
	copy(bufSlice[offset:], encoded)
}

// writeSliceData snapshots the raw memory backing a POD (plain-old-data- primitive) slice and
// records where it was written,so the slice can be reconstructed later as a zero-copy view into the serialized buffer.
// IMPORTANT:
//   - This function performs a raw memory copy using unsafe.
//   - It is ONLY valid for slices whose element type is fixed-size and contains
//     no pointers (i.e. plain-old-data / POD types).
//   - Examples of valid element types: uint64, float64, field.Element, structs
//     composed solely of such fields.
//   - INVALID for: []string, []map, []interface{}, []*T, []big.Int, etc.
//
// The returned FileSlice does NOT contain the slice data itself; it records:
//   - Offset: byte offset in the serialized buffer where the slice data begins
//   - Len:    number of elements in the slice
//   - Cap:    original slice capacity (needed to faithfully reconstruct the
//     slice header during deserialization).
func (w *Writer) writeSliceData(v reflect.Value) FileSlice {
	if v.Len() == 0 {
		return FileSlice{0, 0, 0}
	}

	// Compute the total number of bytes occupied by the slice backing array.
	// This assumes that each element has a fixed size and that the slice
	// memory layout is tightly packed.
	totalBytes := v.Len() * int(v.Type().Elem().Size())

	// Obtain a pointer to the first element of the slice backing array.
	// v.Pointer() is equivalent to the Data field of a slice header - valid for non-empty slice and if
	// the underlying element type is not a pointer-containing type.
	dataPtr := unsafe.Pointer(v.Pointer())

	// Reinterpret the slice backing array as a byte slice of length totalBytes.
	// This does NOT allocate or copy; it is a raw view over existing memory.
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

// linearize serializes a value into the writer buffer and returns a Ref
// (byte offset) pointing to where the value was written.
//
// Key responsibilities:
//  1. Handle pointer de-duplication so the same object is serialized only once.
//  2. Support circular references by registering the pointer address *before*
//     the actual bytes are written.
//  3. Delegate the actual byte-level serialization to linearizeValue.
//
// For non-pointer values, this function simply forwards to linearizeValue.
// For pointer values:
//   - nil pointers serialize to Ref(0)
//   - repeated pointers reuse the previously recorded Ref
//   - circular references are handled by pre-registering the current offset
//     before serializing the pointed-to value.
func linearize(w *Writer, v reflect.Value) (Ref, error) {
	if !v.IsValid() {
		return 0, nil
	}

	// 1. Handle De-Deuplication effectively
	var ptrAddr uintptr
	isPtr := v.Kind() == reflect.Ptr
	if isPtr {
		if v.IsNil() {
			return 0, nil
		}

		// Get the integer representation of the actual RAM address
		ptrAddr = v.Pointer()
		if ref, ok := w.ptrMap[ptrAddr]; ok {
			return ref, nil
		}

		// IMPORANT: We map it to the CURRENT offset (w.offset) because that's where we
		// are about to write it (in the future). This effectively handles CIRCULAR references.
		w.ptrMap[ptrAddr] = Ref(w.offset)
	}

	// 2. WRITE THE DATA - Delegate to the router to actually serialize the bytes.
	ref, err := linearizeValue(w, v)
	if err != nil {
		return 0, err
	}

	if isPtr {
		w.ptrMap[ptrAddr] = ref
	}
	return ref, nil
}

func linearizeValue(w *Writer, v reflect.Value) (Ref, error) {

	// Check Registry first for handling special types
	if handler, ok := CustomRegistry[v.Type()]; ok {
		return handler.Serialize(w, v)
	}

	if v.Type() == reflect.TypeOf(field.Element{}) {
		off := w.Write(v.Interface())
		return Ref(off), nil
	}

	switch v.Kind() {
	// Handle standard Go Pointers - Deduplication happens in 'linearize' before this, so we just recurse.
	// We stick to `linearize` for guardrails - v is not a nil
	// and for the possiblity of the type being pointed at implmenting custom interfaces
	case reflect.Ptr:
		return linearize(w, v.Elem())
	case reflect.Slice:
		if v.IsNil() {
			return 0, nil
		}
		fs := w.writeSliceData(v)
		off := w.Write(fs)
		return Ref(off), nil
	case reflect.String:
		str := v.String()
		dataOff := w.WriteBytes([]byte(str))
		fs := FileSlice{Offset: Ref(dataOff), Len: int64(len(str)), Cap: int64(len(str))}
		off := w.Write(fs)
		return Ref(off), nil
	case reflect.Interface:
		return linearizeInterface(w, v)
	case reflect.Map:
		return linearizeMap(w, v)
	case reflect.Struct:
		size := getBinarySize(v.Type())
		startOffset := w.offset
		w.WriteBytes(make([]byte, size)) // Reserve
		if err := patchStructBody(w, v, startOffset); err != nil {
			return 0, err
		}
		return Ref(startOffset), nil
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return 0, fmt.Errorf("unsupported type: %v cannot be serialized", v.Type())
	default:
		off := w.Write(v.Interface())
		return Ref(off), nil
	}
}

func patchStructBody(w *Writer, v reflect.Value, startOffset int64) error {
	currentFieldOff := int64(0)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)

		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		// Gnark Variable
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
			w.Patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}

		if isIndirectType(f.Type()) {
			ref, err := linearize(w, f)
			if err != nil {
				return err
			}
			w.Patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}

		if f.Kind() == reflect.Struct {
			if f.Type() == reflect.TypeOf(field.Element{}) {
				w.Patch(startOffset+currentFieldOff, f.Interface())
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
				w.Patch(startOffset+currentFieldOff, f.Interface())
				currentFieldOff += int64(f.Type().Size())
			} else {
				if err := patchArray(w, f, startOffset+currentFieldOff); err != nil {
					return err
				}
				currentFieldOff += getBinarySize(f.Type())
			}
			continue
		}

		// Primitives
		w.Patch(startOffset+currentFieldOff, f.Interface())
		// Use getBinarySize to safely get size of int/uint/etc.
		currentFieldOff += getBinarySize(f.Type())
	}
	return nil
}

func patchArray(w *Writer, v reflect.Value, startOffset int64) error {
	currentElemOff := int64(0)
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Array {
			if err := patchArray(w, elem, startOffset+currentElemOff); err != nil {
				return err
			}
		} else if isIndirectType(elemType) {
			ref, err := linearize(w, elem)
			if err != nil {
				return err
			}
			w.Patch(startOffset+currentElemOff, ref)
		} else if elem.Kind() == reflect.Struct {
			if elemType == reflect.TypeOf(field.Element{}) {
				w.Patch(startOffset+currentElemOff, elem.Interface())
			} else {
				if err := patchStructBody(w, elem, startOffset+currentElemOff); err != nil {
					return err
				}
			}
		} else {
			w.Patch(startOffset+currentElemOff, elem.Interface())
		}
		currentElemOff += elemBinSize
	}
	return nil
}

func linearizeMap(w *Writer, v reflect.Value) (Ref, error) {
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

func linearizeInterface(w *Writer, v reflect.Value) (Ref, error) {
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
	typeID, ok := TypeToID[baseType]
	if !ok {
		return 0, fmt.Errorf("encounterd unregistered concrete type: %v", concreteVal.Type())
	}
	ih := InterfaceHeader{TypeID: typeID, Indirection: uint8(indirection), Offset: dataOff}
	off := w.Write(ih)
	return Ref(off), nil
}

func isIndirectType(t reflect.Type) bool {
	if _, ok := CustomRegistry[t]; ok {
		return true
	}
	if t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) {
		return true
	}
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
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

	// Handle INDIRECT/CUSTOM TYPES (Interfaces, Pointers, Strings, Registered Custom Types)
	// If the type is "indirect" (variable size or reference) OR handles its own serialization (custom),
	// we offload the heavy lifting to the main 'linearize' function whichwill return a Ref (offset ID),
	//  which we write into the map buffer.
	_, isCustom := CustomRegistry[t]
	if isIndirectType(t) || isCustom {
		ref, err := linearize(w, v)
		if err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, ref)
	}

	// 2. STRUCTS
	// We handle structs that are stored "inline" (not pointers).
	if v.Kind() == reflect.Struct {
		// Optimization: field.Element is small/fixed, write it directly.
		if t == reflect.TypeOf(field.Element{}) {
			return binary.Write(buf, binary.LittleEndian, v.Interface())
		}
		// Complex structs: flatten their fields into the buffer.
		return linearizeStructBodyMap(w, v, buf)
	}

	// 3. ARRAYS are fixed size, so we iterate and write elements recursively.
	if v.Kind() == reflect.Array {
		for j := 0; j < v.Len(); j++ {
			if err := writeMapElement(w, v.Index(j), buf); err != nil {
				return err
			}
		}
		return nil
	}

	// 4. PRIMITIVES
	// Normalize platform-dependent integers (int, uint) to fixed 64-bit sizes
	// to ensure the binary output is identical on 32-bit and 64-bit machines.
	var val any
	switch v.Kind() {
	case reflect.Int:
		val = int64(v.Int())
	case reflect.Uint:
		val = uint64(v.Uint())
	default:
		val = v.Interface()
	}

	return binary.Write(buf, binary.LittleEndian, val)
}

func linearizeStructBodyMap(w *Writer, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		if !t.IsExported() || strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}
		if isIndirectType(t.Type) {
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
				if err := linearizeStructBodyMap(w, f, buf); err != nil {
					return err
				}
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
		binary.Write(buf, binary.LittleEndian, val)
	}
	return nil
}
