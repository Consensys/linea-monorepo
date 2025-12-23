package serde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/sirupsen/logrus"
)

// encoder: Holds the current encoding/serialization state
type encoder struct {
	// The growing slices of bytes(akin to "Heap") storing the actual serialized payload data.
	buf *bytes.Buffer

	// The write cursor pointing to the end of the buffer.
	offset int64

	// Maps a Go RAM address (uintptr -integer big enough to hold a mem. address
	// for low-level pointer arithmetic) to a File Offset (Ref).
	// Keeps track of memory addresses that have already been seen before.
	ptrMap map[uintptr]Ref

	// uuidMap deduplicates objects based on their logical ID (e.g. Column UUIDs).
	// This ensures that two different pointers to the "same" Column are serialized
	// as a single object reference.
	uuidMap map[string]Ref

	// idMap performs "String Interning" - mapping the raw string content to its File Offset (Ref).
	// If a colID/coinName/queryID is written once, subsequent occurrences just write the Ref (8 bytes).
	idMap map[string]Ref
}

func newEncoder() *encoder {
	// DEBUG: trace creation
	traceLog("Creating New Encoder")
	enc := &encoder{
		buf:     new(bytes.Buffer),
		offset:  0,
		ptrMap:  make(map[uintptr]Ref),
		uuidMap: make(map[string]Ref),
		idMap:   make(map[string]Ref),
	}

	// --- FIX START: Reserve Offset 0 ---
	// We write a single zero byte to ensure that 'offset' moves to 1.
	// This guarantees that Ref(0) is exclusively reserved for NULL pointers
	// and no valid object data ever exists at offset 0.
	enc.writeBytes([]byte{0})
	// --- FIX END ---

	return enc
}

// write writes the given data to the buffer and returns
// the start offset (beginning of cursor) at which the data was written
func (w *encoder) write(data any) int64 {
	start := w.offset
	val := normalizeIntegerSize(reflect.ValueOf(data))
	if err := binary.Write(w.buf, binary.LittleEndian, val); err != nil {
		panic(fmt.Errorf("binary.Write failed for type %T: %w", val, err))
	}
	w.offset += int64(binary.Size(val))
	return start
}

// writeBytes writes the given bytes to the buffer and returns
// the start offset (beginning of cursor) at which the bytes were written
func (w *encoder) writeBytes(b []byte) int64 {
	start := w.offset
	w.buf.Write(b)
	w.offset += int64(len(b))
	return start
}

// patch jumps back in-time to specific offset (written/reserved with zero bytes earlier)
// and overwrites the data (reserved with zeros) with the actual data
func (w *encoder) patch(offset int64, v any) {
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
func (w *encoder) writeSliceData(v reflect.Value) FileSlice {
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

// encode serializes a value into the writer buffer and returns a Ref
// (8-byte offset) pointing to root offset where the value was written.
// This essentially tells the decoder where the "entry point" of our object
// graph is located inside the file.
// Key responsibilities:
//  1. Handle pointer de-duplication so the same object is serialized only once.
//  2. Support circular references by registering the pointer address *before*
//     the actual bytes are written.
//  3. Delegate the actual byte-level serialization to linearize.
//
// For non-pointer values, this function simply forwards to linearize.
// For pointer values:
//   - nil pointers serialize to Ref(0)
//   - repeated pointers reuse the previously recorded Ref
//   - circular references are handled by pre-registering the current offset
//     before serializing the pointed-to value.

func encode(w *encoder, v reflect.Value) (Ref, error) {
	if !v.IsValid() {
		return 0, nil
	}

	// DEBUG: Trace Entry
	traceEnter("ENCODE", v)
	defer traceExit("ENCODE", nil)

	// 1. Handle De-Deuplication effectively
	var ptrAddr uintptr
	isRef := v.Kind() == reflect.Ptr
	if isRef {
		if v.IsNil() {
			traceLog("Encode: Pointer is Nil -> Ref(0)")
			return 0, nil
		}

		// Get the integer representation of the actual RAM address
		ptrAddr = v.Pointer()
		if ref, ok := w.ptrMap[ptrAddr]; ok {
			traceLog("Encode: Dedup Hit! Addr %x -> Ref %d", ptrAddr, ref)
			//logrus.Infof("Encode: Dedup Hit! Addr %x -> Ref %d", ptrAddr, ref)
			return ref, nil
		}

		// IMPORANT: We map it to the CURRENT offset (w.offset) because that's where we
		// are about to write it (in the future). This effectively handles CIRCULAR references.
		traceLog("Encode: New Pointer %x -> Will be Ref %d", ptrAddr, w.offset)
		w.ptrMap[ptrAddr] = Ref(w.offset)
	}

	// 2. WRITE THE DATA - Delegate to the router to actually serialize the bytes.
	ref, err := linearize(w, v)
	if err != nil {
		return 0, err
	}

	if isRef {
		traceLog("Encode: Finished Ptr %x -> Ref %d", ptrAddr, ref)
		w.ptrMap[ptrAddr] = ref
	}
	return ref, nil
}

func linearize(w *encoder, v reflect.Value) (Ref, error) {
	// DEBUG: Trace
	traceEnter("LINEARIZE", v)
	defer traceExit("LINEARIZE", nil)

	// Check Registry first for handling special types
	if handler, ok := customRegistry[v.Type()]; ok {
		traceLog("Using Custom Handler for %s", v.Type())
		return handler.marshall(w, v)
	}

	// If the type implements ifaces.Query, we use UUID deduplication
	// We check t.Kind() != Interface to ensure we are looking at the concrete type
	t := v.Type()
	if t.Kind() != reflect.Interface && t.Implements(TypeOfQuery) {
		traceLog("Type %s implements Query -> using UUID Dedup", t)
		return marshallQueryViaUUID(w, v)
	}

	if v.Type() == reflect.TypeOf(field.Element{}) {
		off := w.write(v.Interface())
		return Ref(off), nil
	}

	switch v.Kind() {
	// Handle standard Go Pointers - Deduplication happens in 'linearize' before this, so we just recurse.
	// We stick to `linearize` for guardrails - v is not a nil
	// and for the possiblity of the type being pointed at implmenting custom interfaces
	case reflect.Ptr:
		return encode(w, v.Elem())
	case reflect.Array:
		return linearizeArray(w, v)
	case reflect.Slice:
		if v.IsNil() {
			return 0, nil
		}

		elemType := v.Type().Elem()

		// --- FIX START ---
		// 1. Indirect Elements (Ptr, Interface, String, etc.) -> Must serialize as list of Refs.
		if isIndirectType(elemType) {
			traceLog("Slice is Indirect -> writeSliceOfIndirects")
			return writeSliceOfIndirects(w, v)
		}

		// 2. NON-POD Elements (e.g., Structs containing pointers).
		// We cannot use writeSliceData (unsafe memcopy) because it would write raw pointer addresses
		// to disk, which leads to corruption/drift upon loading.
		if !isPod(elemType) {
			traceLog("Slice is Struct-With-Pointers -> linearizeSliceSeq")
			return linearizeSliceSeq(w, v)
		}
		// --- FIX END ---

		// 3. POD Elements -> Safe for raw memcopy optimization.
		fs := w.writeSliceData(v)
		off := w.write(fs)
		return Ref(off), nil
	case reflect.String:
		str := v.String()
		dataOff := w.writeBytes([]byte(str))
		fs := FileSlice{Offset: Ref(dataOff), Len: int64(len(str)), Cap: int64(len(str))}
		off := w.write(fs)
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
		off := w.write(v.Interface())
		return Ref(off), nil
	}
}

// Define the interface type reflection once
var TypeOfQuery = reflect.TypeOf((*ifaces.Query)(nil)).Elem()

func marshallQueryViaUUID(enc *encoder, v reflect.Value) (Ref, error) {
	// 1. Extract the UUID
	// We know v implements Query, so we can cast it.
	q, ok := v.Interface().(ifaces.Query)
	if !ok {
		// Should not happen due to Implements() check, but safety first
		return 0, fmt.Errorf("value of type %s does not implement ifaces.Query", v.Type())
	}

	id := q.UUID().String()

	// 2. Check Logical Cache (UUID)
	if ref, ok := enc.uuidMap[id]; ok {
		traceLog("Query UUID Dedup Hit! %s -> Ref %d", id, ref)
		return ref, nil
	}

	// 3. Serialize
	// We strictly want to serialize the STRUCT content here, not the interface wrapper.
	// If we call encode(v) directly, we might loop if v is a pointer.
	// We fall back to linearizeStruct or the standard encoding flow for the *underlying* data.
	// However, since we intercepted this in linearize, we must be careful not to infinite loop.

	// STRATEGY:
	// We manually linearize the body based on its Kind.
	// This bypasses the 'linearize' check for Query interface to avoid recursion,
	// effectively treating it as a standard struct/pointer for the actual writing phase.

	var ref Ref
	var err error

	if v.Kind() == reflect.Ptr {
		// If it's a pointer, we dereference it to write the body,
		// OR we just let the standard struct handler do it.
		// A safer way is to call the internal helper that writes structs.
		// But since we don't want to duplicate logic, let's use a trick:
		// We cast it to a type that DOES NOT implement Query (e.g. generic struct)
		// effectively masking the method. But Go doesn't allow masking easily.

		// SIMPLEST WAY:
		// Recursively call encode on the Elem() (if ptr) or linearizeStruct (if struct).
		// Since 'encode' checks ptrMap (Memory Dedup), it's safe.
		// But we need to avoid re-entering 'marshallQueryViaUUID'.

		// If it is a Pointer, dereference it until we hit the Struct.
		// Then call linearizeStruct on the struct.
		elem := v
		for elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		// Reserve space and write fields
		ref, err = linearizeStruct(enc, elem)
	} else if v.Kind() == reflect.Struct {
		ref, err = linearizeStruct(enc, v)
	} else {
		return 0, fmt.Errorf("query implementation must be a struct or pointer to struct, got %v", v.Kind())
	}

	if err != nil {
		return 0, err
	}

	// 4. Register Reference
	enc.uuidMap[id] = ref
	return ref, nil
}

// linearizeStruct function works on a Reservation now and Patch later System.
// It does not append the struct fields one by one to the end of the file.
// Instead, it "allocates" the space first, and then goes back to fill it in.
// Explanation:
// linearizeStruct is recursive, the "deepest" objects (the ones pointed to incase of pointers) get their data finalized first
// while the "top-level" objects are still waiting to finish their patchStructBody loop. The Write Cursor (w.offset)
// always moves forward, ensuring that new objects never overwrite old ones.
func linearizeStruct(w *encoder, v reflect.Value) (Ref, error) {
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
	w.writeBytes(make([]byte, size))

	// PATCH - Go back and fill in the blank canvas with actual data.
	if err := patchStructBody(w, v, startOffset); err != nil {
		return 0, err
	}

	return Ref(startOffset), nil
}

func linearizeMap(w *encoder, v reflect.Value) (Ref, error) {
	if v.IsNil() {
		return 0, nil
	}
	var bodyBuf bytes.Buffer
	iter := v.MapRange()
	count := 0
	for iter.Next() {
		if err := encodeSeqItem(w, iter.Key(), &bodyBuf); err != nil {
			return 0, err
		}
		if err := encodeSeqItem(w, iter.Value(), &bodyBuf); err != nil {
			return 0, err
		}
		count++
	}
	dataOff := w.writeBytes(bodyBuf.Bytes())
	fs := FileSlice{Offset: Ref(dataOff), Len: int64(count), Cap: int64(count)}
	off := w.write(fs)
	return Ref(off), nil
}

func linearizeInterface(w *encoder, v reflect.Value) (Ref, error) {
	// DEBUG: Trace
	traceEnter("LIN_IFACE", v)
	defer traceExit("LIN_IFACE", nil)

	if v.IsNil() {
		traceLog("Interface is True Nil")
		return 0, nil
	}
	concreteVal := v.Elem()

	// DEBUG: Detect Typed Nil
	traceLog("Concrete Value: %s Kind: %v", concreteVal.Type(), concreteVal.Kind())
	if concreteVal.Kind() == reflect.Ptr && concreteVal.IsNil() {
		traceLog("!!! ALERT: TYPED NIL DETECTED !!! Type: %s", concreteVal.Type())
	}

	dataOff, err := encode(w, concreteVal)
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

	traceLog("Writing InterfaceHeader: TypeID=%d Ind=%d Offset=%d", typeID, indirection, dataOff)
	ih := InterfaceHeader{TypeID: typeID, PtrIndirection: uint8(indirection), Offset: dataOff}
	off := w.write(ih)
	return Ref(off), nil
}

func linearizeStructBodyMap(w *encoder, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)

		// 1. Skip unexported or omitted fields
		if strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}

		// Skip unexported fields while displaying a warning to the user incase of forgetfulness
		// of exporting any necessary field
		if !t.IsExported() {
			logrus.Warnf("field %v.%v is unexported", t.Type, t.Name)
			continue
		}
		if isIndirectType(t.Type) {
			ref, err := encode(w, f)
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

func patchStructBody(w *encoder, v reflect.Value, startOffset int64) error {
	// DEBUG: Trace
	traceEnter("PATCH_STRUCT", v, "StartOffset", startOffset)
	defer traceExit("PATCH_STRUCT", nil)

	currentFieldOff := int64(0)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		fType := f.Type()

		// 1. Skip unexported or omitted fields
		if strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}

		// Skip unexported fields while displaying a warning to the user incase of forgetfulness
		// of exporting any necessary field
		if !t.IsExported() {
			logrus.Warnf("field %v.%v is unexported", t.Type, t.Name)
			continue
		}

		// DEBUG: Trace Field
		traceLog("Patching Field: %s (Type: %s, Offset: %d)", t.Name, fType, startOffset+currentFieldOff)

		// 2. Handle References: Indirect types ( incl. Custom Types)
		// If the type is "indirect" (pointers, slices) OR it is a registered Custom Type
		// (e.g., frontend.Variable), we do NOT write the data here.
		// Instead, we 'linearize' it to the heap and patch the current struct field
		// with the returned 8-byte Reference ID (Ref).
		if isIndirectType(fType) {
			ref, err := encode(w, f)
			if err != nil {
				return err
			}
			w.patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8 // Refs are always 8 bytes
			continue
		}

		// 3. Handle Nested Structs (Inline)
		if f.Kind() == reflect.Struct {
			// Optimization: field.Element is treated as a primitive blob
			if fType == reflect.TypeOf(field.Element{}) {
				w.patch(startOffset+currentFieldOff, f.Interface())
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
		w.patch(startOffset+currentFieldOff, f.Interface())
		currentFieldOff += getBinarySize(fType)
	}
	return nil
}

func patchArray(w *encoder, v reflect.Value, startOffset int64) error {
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)

	// Pre-calculation: Determine if elements are "References"
	// (Pointers, Slices, or Registered Custom Types like frontend.Variable).
	// Since it's an array, this decision applies to EVERY element.
	isReference := isIndirectType(elemType)

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// Calculate the exact offset for this specific element
		offset := startOffset + (int64(i) * elemBinSize)

		// 1. Handle References (Heap Allocation)
		// If the element is a pointer, interface, or custom type, we delegate to
		// linearize() to write the data elsewhere and return a Ref ID.
		if isReference {
			ref, err := encode(w, elem)
			if err != nil {
				return err
			}
			w.patch(offset, ref)
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
				w.patch(offset, elem.Interface())
			} else {
				// Complex Struct: Recurse
				if err := patchStructBody(w, elem, offset); err != nil {
					return err
				}
			}

		default:
			// Primitives (int, uint, bool)
			// w.Patch automatically handles int/uint normalization now.
			w.patch(offset, elem.Interface())
		}
	}
	return nil
}

func encodeSeqItem(w *encoder, v reflect.Value, buf *bytes.Buffer) error {
	t := v.Type()

	// Handle Indirect types (Interfaces, Pointers, Strings, Registered Custom Types)
	// If the type is "indirect" (variable size or reference or handles its own serialization (custom)),
	// we offload the heavy lifting to the main 'linearize' function which will return a Ref (offset ID),
	// which we write into the map buffer.
	if isIndirectType(t) {
		ref, err := encode(w, v)
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
			if err := encodeSeqItem(w, v.Index(j), buf); err != nil {
				return err
			}
		}
		return nil
	}

	// 4. PRIMITIVES
	val := normalizeIntegerSize(v)
	return binary.Write(buf, binary.LittleEndian, val)
}

// --- NEW HELPER FUNCTION ---
func writeSliceOfIndirects(w *encoder, v reflect.Value) (Ref, error) {
	// 1. Recursively encode all elements first.
	// This ensures their data is written to the heap and we get back valid Refs (offsets).
	refs := make([]Ref, v.Len())
	for i := 0; i < v.Len(); i++ {
		ref, err := encode(w, v.Index(i))
		if err != nil {
			return 0, err
		}
		refs[i] = ref
	}

	// 2. Write the array of Refs (int64s) to the buffer.
	// This creates a contiguous block of "pointers-by-offset" in the file.
	startOffset := w.offset
	if err := binary.Write(w.buf, binary.LittleEndian, refs); err != nil {
		return 0, err
	}
	w.offset += int64(len(refs) * 8)

	// 3. Write the FileSlice Header pointing to this array of Refs.
	fs := FileSlice{
		Offset: Ref(startOffset),
		Len:    int64(v.Len()),
		Cap:    int64(v.Cap()),
	}
	off := w.write(fs)
	return Ref(off), nil
}

// getBinarySize returns the number of bytes a value of type `T` will occupy
// in the serialized buffer according to serde layout rules.
//
// NOTE: This is NOT the same as Go's in-memory size. The size returned here reflects
// how the value is represented on disk:
//
// - Fixed-size, pointer-free values are inlined.
// - Variable-size or heap-backed values are replaced by an 8-byte Ref.
//
// This function MUST remain perfectly consistent with the actual write logic;
// any mismatch will result in corrupted offsets or incorrect deserialization.
func getBinarySize(t reflect.Type) int64 {

	// Indirect types (custom registries, ptrs, slices, strings etc) are types that have variable sizes
	// not known at compile time and hence their inline representation is a Ref (8-byte offset).
	// These values are written elsewhere and referenced inline by offset.
	if isIndirectType(t) {
		return 8 // Size of Ref (8-byte offset)
	}

	k := t.Kind()

	// Explicit handling of scalar types int/uint are platform-dependent in Go,
	// so we normalize them to 8 bytes in the serialized representation.
	if k == reflect.Int || k == reflect.Uint {
		return 8
	}

	// Structs are serialized inline by concatenating the serialized
	// representation of each exported, non-omitted field.
	if k == reflect.Struct {
		// Special-case POD types that are safe to inline as raw bytes.
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}

		var sum int64
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Skip unexported fields.
			if !f.IsExported() {
				continue
			}

			// Skip fields explicitly omitted from serialization.
			if strings.Contains(f.Tag.Get(serdeStructTag), serdeStructTagOmit) {
				continue
			}

			sum += getBinarySize(f.Type)
		}
		return sum
	}

	// Arrays are fixed-size and serialized inline as repeated elements.
	if k == reflect.Array {
		// field.Element arrays are treated as POD blobs.
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}

		elemSize := getBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}

	// Fallback for other fixed-size, pointer-free types
	// (e.g. bool, int/uint8, int/uint16, int/uint32, int/uint64, float32/64).
	return int64(t.Size())
}

// linearizeArray handles the serialization of Go arrays (e.g., [10]MyStruct).
// Unlike slices (which are variable length), arrays are fixed-size values.
// We must serialize them by reserving space and then filling it element-by-element
// to ensure that any pointers inside the elements are correctly resolved to Refs.
func linearizeArray(w *encoder, v reflect.Value) (Ref, error) {
	// 1. PREDICT: Calculate exactly how many bytes this array will occupy on disk.
	// This uses getBinarySize() which accounts for Refs (8 bytes) vs Inline data.
	size := getBinarySize(v.Type())

	// 2. SNAPSHOT: Capture the current cursor position. This is where our array begins.
	startOffset := w.offset

	// 3. RESERVE: Write zero bytes to reserve the contiguous block of memory.
	w.writeBytes(make([]byte, size))

	// 4. PATCH: Iterate through the array elements and write the actual data
	// (or References) into the reserved space. We reuse the existing patchArray helper.
	if err := patchArray(w, v, startOffset); err != nil {
		return 0, err
	}

	return Ref(startOffset), nil
}

// linearizeSliceSeq mimics linearizeArray but for Slices.
// It serializes elements sequentially into the buffer (heap).
// Used for "Non-POD" slices (e.g., slice of structs containing pointers).
func linearizeSliceSeq(w *encoder, v reflect.Value) (Ref, error) {
	totalSize := int64(0)
	for i := 0; i < v.Len(); i++ {
		totalSize += getBinarySize(v.Index(i).Type())
	}

	startOffset := w.offset
	w.writeBytes(make([]byte, totalSize))

	currentOff := startOffset
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		// Reuse patch logic (generic patch handles structs/arrays/primitives)
		switch elem.Kind() {
		case reflect.Struct:
			if err := patchStructBody(w, elem, currentOff); err != nil {
				return 0, err
			}
		case reflect.Array:
			if err := patchArray(w, elem, currentOff); err != nil {
				return 0, err
			}
		default:
			// Primitive
			w.patch(currentOff, elem.Interface())
		}
		currentOff += getBinarySize(elem.Type())
	}

	fs := FileSlice{
		Offset: Ref(startOffset),
		Len:    int64(v.Len()),
		Cap:    int64(v.Cap()),
	}
	off := w.write(fs)
	return Ref(off), nil
}

// isIndirectType returns true if the given type is an indirect type.
// Indirect types are types that have variable sizes not known at compile time.
// This includes pointers, slices, strings, interfaces, maps, and functions.
// Direct types are types that have fixed sizes known at compile time.
// This includes arrays, and primitive types (bool,int/uint8, int/uint (normalized), etc).
// Types handled inside of the Custom Registry are also considered indirect.
// NOTE: structs can be either direct or indirect depending on their fields. Hence when
// dealing with structs, we use isPod() to determine if they are direct or indirect.
func isIndirectType(t reflect.Type) bool {
	if _, ok := customRegistry[t]; ok {
		return true
	}

	// Indirect types are types that have variable sizes - not known at compile time
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

// isPod (Plain Old Data) returns true only if the type contains NO references
// (pointers, slices, strings, interfaces, maps, custom types) recursively.
// It is used to determine if unsafe raw memory copy is safe.
func isPod(t reflect.Type) bool {
	// If it is Indirect (ptr, slice, map, interface, string, custom), it is NOT POD.
	if isIndirectType(t) {
		return false
	}

	// Structs are POD only if all fields are POD
	if t.Kind() == reflect.Struct {
		if t == reflect.TypeOf(field.Element{}) {
			return true
		}
		for i := 0; i < t.NumField(); i++ {
			if !isPod(t.Field(i).Type) {
				return false
			}
		}
		return true
	}

	// Arrays are POD only if elements are POD
	if t.Kind() == reflect.Array {
		return isPod(t.Elem())
	}

	// Primitives (Int, Float, Bool) are POD.
	// Note: 'String' is caught by isIndirectType check above.
	return true
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
