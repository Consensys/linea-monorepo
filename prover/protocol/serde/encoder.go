package serde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Encoder: Holds the current encoding/serialization state
type Encoder struct {
	// The growing array of bytes (the "Heap").
	buf *bytes.Buffer

	// The write cursor. Points to the end of the buffer.
	offset int64

	// Maps a Go RAM address (uintptr -integer big enough to hold a mem. address
	// for low-level pointer arithmetic) to a File Offset (Ref).
	// Keeps track of memory addresses that have already been seen before.
	ptrMap map[uintptr]Ref
}

func NewWriter() *Encoder {
	return &Encoder{
		buf:    new(bytes.Buffer),
		offset: 0,
		ptrMap: make(map[uintptr]Ref), // Initialize the deduplication map.
	}
}

// Write writes the given data to the buffer and returns
// the start offset (beginning of cursor) at which the data was written
func (w *Encoder) Write(data any) int64 {
	start := w.offset
	val := normalizeIntegerSize(reflect.ValueOf(data))
	if err := binary.Write(w.buf, binary.LittleEndian, val); err != nil {
		panic(fmt.Errorf("binary.Write failed for type %T: %w", val, err))
	}
	w.offset += int64(binary.Size(val))
	return start
}

// WriteBytes writes the given bytes to the buffer and returns
// the start offset (beginning of cursor) at which the bytes were written
func (w *Encoder) WriteBytes(b []byte) int64 {
	start := w.offset
	w.buf.Write(b)
	w.offset += int64(len(b))
	return start
}

// Patch jumps back in-time to specific offset (written/reserved with zero bytes earlier)
// and overwrites the data (reserved with zeros) with the actual data
func (w *Encoder) Patch(offset int64, v any) {
	var tmp bytes.Buffer
	val := normalizeIntegerSize(reflect.ValueOf(v))
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
func (w *Encoder) writeSliceData(v reflect.Value) FileSlice {
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
func linearize(w *Encoder, v reflect.Value) (Ref, error) {
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

func linearizeValue(w *Encoder, v reflect.Value) (Ref, error) {

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
		return linearizeStruct(w, v)
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return 0, fmt.Errorf("unsupported type: %v cannot be serialized", v.Type())
	default:
		off := w.Write(v.Interface())
		return Ref(off), nil
	}
}

// linearizeStruct function works on a Reservation now and Patch later System.
// It does not append the struct fields one by one to the end of the file.
// Instead, it "allocates" the space first, and then goes back to fill it in.
// Explanation:
// linearizeStruct is recursive, the "deepest" objects (the ones pointed to incase of pointers) get their data finalized first
// while the "top-level" objects are still waiting to finish their patchStructBody loop. The Write Cursor (w.offset)
// always moves forward, ensuring that new objects never overwrite old ones.
func linearizeStruct(w *Encoder, v reflect.Value) (Ref, error) {
	// PREDICT: Calculate total size of the struct in bytes.
	// CRITICAL: This calculation MUST perfectly match the bytes written by 'patchStructBody'.
	// If getBinarySize returns X, patchStructBody must write exactly X bytes.
	// Any mismatch will cause data corruption or a panic in w.Patch.
	size := getBinarySize(v.Type())

	// SNAPSHOT: Capture the current cursor position. This 'startOffset' will point to the beginning of our reserved "slot"
	// Remember before reserving space, let us capture the current offset (i.e cursor position of the writer)
	// This is important because in the reserve step, when we write zero bytes, the writter offset
	// will be changed (mutates to the end of the reserved space). However, while coming back in time
	// we need to know the current offset (beginning of cursor).
	startOffset := w.offset

	// RESERVE - Write 'size' bytes of zeros. We now have a blank canvas in the file.
	w.WriteBytes(make([]byte, size))

	// PATCH - Go back and fill in the blank canvas with actual data.
	if err := patchStructBody(w, v, startOffset); err != nil {
		return 0, err
	}

	return Ref(startOffset), nil
}

func linearizeMap(w *Encoder, v reflect.Value) (Ref, error) {
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

func linearizeInterface(w *Encoder, v reflect.Value) (Ref, error) {
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

func linearizeStructBodyMap(w *Encoder, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		if !t.IsExported() || strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
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
		val := normalizeIntegerSize(f)
		binary.Write(buf, binary.LittleEndian, val)
	}

	return nil
}

func patchStructBody(w *Encoder, v reflect.Value, startOffset int64) error {
	currentFieldOff := int64(0)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		fType := f.Type()

		// 1. Skip unexported or omitted fields
		if !t.IsExported() || strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}

		// 2. Handle References: Indirect types & Custom Types
		//
		// IF the type is "indirect" (pointers, slices) OR it is a registered Custom Type
		// (e.g., frontend.Variable), we do NOT write the data here.
		// Instead, we 'linearize' it to the heap and patch the current struct field
		// with the returned 8-byte Reference ID (Ref).
		_, isCustom := CustomRegistry[fType]
		if isIndirectType(fType) || isCustom {
			ref, err := linearize(w, f)
			if err != nil {
				return err
			}
			w.Patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8 // Refs are always 8 bytes
			continue
		}

		// 3. Handle Nested Structs (Inline)
		if f.Kind() == reflect.Struct {
			// Optimization: field.Element is treated as a primitive blob
			if fType == reflect.TypeOf(field.Element{}) {
				w.Patch(startOffset+currentFieldOff, f.Interface())
			} else {
				// Recurse to patch the inner struct's fields
				if err := patchStructBody(w, f, startOffset+currentFieldOff); err != nil {
					return err
				}
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}

		// 4. Handle Arrays (Inline)
		if f.Kind() == reflect.Array {
			// Delegate to patchArray helper
			if err := patchArray(w, f, startOffset+currentFieldOff); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}

		// 5. Handle Primitives (int, uint, bool)
		// Note: w.Patch calls 'normalizeValue' internally, so int/uint safety is handled.
		w.Patch(startOffset+currentFieldOff, f.Interface())
		currentFieldOff += getBinarySize(fType)
	}
	return nil
}

func patchArray(w *Encoder, v reflect.Value, startOffset int64) error {
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)

	// Pre-calculation: Determine if elements are "References"
	// (Pointers, Slices, or Registered Custom Types like frontend.Variable).
	// Since it's an array, this decision applies to EVERY element.
	_, isCustom := CustomRegistry[elemType]
	isReference := isIndirectType(elemType) || isCustom

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// Calculate the exact offset for this specific element
		offset := startOffset + (int64(i) * elemBinSize)

		// 1. Handle References (Heap Allocation)
		// If the element is a pointer, interface, or custom type, we delegate to
		// linearize() to write the data elsewhere and return a Ref ID.
		if isReference {
			ref, err := linearize(w, elem)
			if err != nil {
				return err
			}
			w.Patch(offset, ref)
			continue
		}

		// 2. Handle Inline Data
		switch elem.Kind() {
		case reflect.Array:
			// Nested Array: Recurse
			if err := patchArray(w, elem, offset); err != nil {
				return err
			}

		case reflect.Struct:
			// Optimization: field.Element is a "primitive" struct
			if elemType == reflect.TypeOf(field.Element{}) {
				w.Patch(offset, elem.Interface())
			} else {
				// Complex Struct: Recurse
				if err := patchStructBody(w, elem, offset); err != nil {
					return err
				}
			}

		default:
			// Primitives (int, uint, bool)
			// w.Patch automatically handles int/uint normalization now.
			w.Patch(offset, elem.Interface())
		}
	}
	return nil
}

func writeMapElement(w *Encoder, v reflect.Value, buf *bytes.Buffer) error {
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
	val := normalizeIntegerSize(v)
	return binary.Write(buf, binary.LittleEndian, val)
}

func isIndirectType(t reflect.Type) bool {
	if _, ok := CustomRegistry[t]; ok {
		return true
	}

	// Indirect types are types that have variable sizes - not known at compile time
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

// normalizeIntegerSize converts platform-dependent types (int, uint) to fixed-size
// equivalents (int64, uint64) - 64 bit values. This ensures that the binary representation
// is consistent across different CPU architectures (32-bit vs 64-bit).
func normalizeIntegerSize(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Int:
		return int64(v.Int())
	case reflect.Uint:
		return uint64(v.Uint())
	default:
		// For all other types (int64, float64, structs, etc.),
		// return the interface as-is.
		return v.Interface()
	}
}
