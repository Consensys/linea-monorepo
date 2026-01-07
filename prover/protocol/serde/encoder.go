package serde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

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
	traceLog("Creating New Encoder")
	enc := &encoder{
		buf:     new(bytes.Buffer),
		offset:  0,
		ptrMap:  make(map[uintptr]Ref),
		uuidMap: make(map[string]Ref),
		idMap:   make(map[string]Ref),
	}
	return enc
}

// write writes the given data (raw Primitives, FileSlice, Interface headers etc) to the buffer and returns
// the start offset where the data was written.
func (w *encoder) write(data any) int64 {
	start := w.offset
	switch v := data.(type) {
	// Normalize platform-dependent int/uint to fixed-size int64/uint64
	case int:
		var buf [8]byte
		*(*int64)(unsafe.Pointer(&buf[0])) = int64(v)
		w.buf.Write(buf[:])
		w.offset += 8
	case uint:
		var buf [8]byte
		*(*uint64)(unsafe.Pointer(&buf[0])) = uint64(v)
		w.buf.Write(buf[:])
		w.offset += 8
	case int64, uint64:
		w.buf.Write((*[8]byte)(unsafe.Pointer(&v))[:])
		w.offset += 8
	case int32, uint32:
		w.buf.Write((*[4]byte)(unsafe.Pointer(&v))[:])
		w.offset += 4
	case int16, uint16:
		w.buf.Write((*[2]byte)(unsafe.Pointer(&v))[:])
		w.offset += 2
	case int8, uint8, bool:
		w.buf.Write((*[1]byte)(unsafe.Pointer(&v))[:])
		w.offset += 1
	case Ref:
		w.buf.Write((*[8]byte)(unsafe.Pointer(&v))[:])
		w.offset += 8
	case FileSlice:
		w.buf.Write((*[24]byte)(unsafe.Pointer(&v))[:])
		w.offset += 24
	case InterfaceHeader:
		w.buf.Write((*[unsafe.Sizeof(v)]byte)(unsafe.Pointer(&v))[:])
		w.offset += int64(unsafe.Sizeof(v))
	default:
		// Fallback to reflection
		val := normalizeIntegerSize(reflect.ValueOf(data))
		binary.Write(w.buf, binary.LittleEndian, val)
		w.offset += int64(binary.Size(val))
	}
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

// patch overwrites previously reserved zero bytes (at offset) with actual data v.
func (w *encoder) patch(offset int64, v any) {
	bufSlice := w.buf.Bytes()
	switch val := v.(type) {
	case Ref:
		*(*Ref)(unsafe.Pointer(&bufSlice[offset])) = val
	case int64:
		*(*int64)(unsafe.Pointer(&bufSlice[offset])) = val
	case uint64:
		*(*uint64)(unsafe.Pointer(&bufSlice[offset])) = val
	case FileSlice:
		*(*FileSlice)(unsafe.Pointer(&bufSlice[offset])) = val
	case InterfaceHeader:
		*(*InterfaceHeader)(unsafe.Pointer(&bufSlice[offset])) = val
	default:
		// Fallback for complex types
		var tmp bytes.Buffer
		val = normalizeIntegerSize(reflect.ValueOf(v))
		binary.Write(&tmp, binary.LittleEndian, val)
		copy(bufSlice[offset:], tmp.Bytes())
	}
}

// writeSliceData snapshots the raw memory backing a POD (Plain-Old-Data) slice and
// records where it was written,so the slice can be reconstructed later as a zero-copy view
// into the serialized buffer.
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
	dataPtr := unsafe.Pointer(v.Pointer())

	// Reinterpret the slice backing array as a byte slice of length totalBytes.
	// This does NOT allocate or copy; it is a raw view over existing memory.
	dataBytes := unsafe.Slice((*byte)(dataPtr), totalBytes)
	offset := w.offset
	w.buf.Write(dataBytes)
	w.offset += int64(totalBytes)
	return FileSlice{Offset: Ref(offset), Len: int64(v.Len()), Cap: int64(v.Cap())}
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
	traceEnter("ENCODE", v)
	defer traceExit("ENCODE", nil)

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
			return ref, nil
		}

		// IMPORANT: We map it to the CURRENT offset (w.offset) because that's where we
		// are about to write it (in the future). This effectively handles CIRCULAR references.
		traceLog("Encode: New Pointer %x -> Will be Ref %d", ptrAddr, w.offset)
		w.ptrMap[ptrAddr] = Ref(w.offset)
	}

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
	traceEnter("LINEARIZE", v)
	defer traceExit("LINEARIZE", nil)

	// Custom handlers override default behaviour
	if handler, ok := customRegistry[v.Type()]; ok {
		traceLog("Using Custom Handler for %s", v.Type())
		return handler.marshall(w, v)
	}

	// UUID dedup for ifaces.Query
	t := v.Type()
	if t.Kind() != reflect.Interface && t.Implements(TypeOfQuery) {
		traceLog("Type %s implements Query -> using UUID Dedup", t)
		return marshallQuery(w, v)
	}

	switch v.Kind() {
	case reflect.Ptr:
		// Handle standard Go Pointers - Deduplication happens in 'encode' before this,
		// so we just recurse. We stick to `encode` for guardrails - v is not a nil and
		// for the possiblity of the type being pointed at implmenting custom interfaces.
		return encode(w, v.Elem())
	case reflect.Array:
		return linearizeArray(w, v)
	case reflect.Slice:
		if v.IsNil() {
			return 0, nil
		}
		elemType := v.Type().Elem()
		if isIndirectType(elemType) {
			traceLog("Slice is Indirect -> writeSliceOfIndirects")
			return writeIndirectSlice(w, v)
		}
		if !isPOD(elemType) {
			traceLog("Slice is Non-POD -> linearizeSliceSeq")
			return linearizeSliceSeq(w, v)
		}
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
	traceEnter("LIN_IFACE", v)
	defer traceExit("LIN_IFACE", nil)
	if v.IsNil() {
		traceLog("Interface is True Nil")
		return 0, nil
	}
	concreteVal := v.Elem()
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
	ih := InterfaceHeader{TypeID: typeID, PtrIndirection: uint8(indirection), Offset: dataOff}
	off := w.write(ih)
	return Ref(off), nil
}

func linearizeArray(w *encoder, v reflect.Value) (Ref, error) {
	size := getBinarySize(v.Type())
	startOffset := w.offset
	w.writeBytes(make([]byte, size))
	if err := patchArray(w, v, startOffset); err != nil {
		return 0, err
	}
	return Ref(startOffset), nil
}

// linearizeSliceSeq: serializes non-POD slices element-by-element (used for structs that contain pointers)
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
			w.patch(currentOff, elem.Interface())
		}
		currentOff += getBinarySize(elem.Type())
	}
	fs := FileSlice{Offset: Ref(startOffset), Len: int64(v.Len()), Cap: int64(v.Cap())}
	off := w.write(fs)
	return Ref(off), nil
}

func writeIndirectSlice(w *encoder, v reflect.Value) (Ref, error) {
	n := v.Len()
	refs := make([]Ref, n)
	for i := 0; i < n; i++ {
		ref, err := encode(w, v.Index(i))
		if err != nil {
			return 0, err
		}
		refs[i] = ref
	}
	startOffset := w.offset
	if err := binary.Write(w.buf, binary.LittleEndian, refs); err != nil {
		return 0, err
	}
	w.offset += int64(n * 8)
	fs := FileSlice{Offset: Ref(startOffset), Len: int64(n), Cap: int64(v.Cap())}
	off := w.write(fs)
	return Ref(off), nil
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
	if v.Kind() == reflect.Struct {
		return linearizeStructBodyMap(w, v, buf)
	}
	if v.Kind() == reflect.Array {
		for j := 0; j < v.Len(); j++ {
			if err := encodeSeqItem(w, v.Index(j), buf); err != nil {
				return err
			}
		}
		return nil
	}
	val := normalizeIntegerSize(v)
	return binary.Write(buf, binary.LittleEndian, val)
}

// linearizeStructBodyMap - used to flatten an inline struct into a buffer (map/sequence usage)
func linearizeStructBodyMap(w *encoder, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		if strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}
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
			// Recurse for nested structs
			if err := linearizeStructBodyMap(w, f, buf); err != nil {
				return err
			}
			continue
		}
		val := normalizeIntegerSize(f)
		binary.Write(buf, binary.LittleEndian, val)
	}
	return nil
}

func patchStructBody(w *encoder, v reflect.Value, startOffset int64) error {
	traceEnter("PATCH_STRUCT", v, "StartOffset", startOffset)
	defer traceExit("PATCH_STRUCT", nil)
	currentFieldOff := int64(0)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		fType := f.Type()
		if strings.Contains(t.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}
		if !t.IsExported() {
			logrus.Warnf("field %v.%v is unexported", t.Type, t.Name)
			continue
		}
		traceLog("Patching Field: %s (Type: %s, Offset: %d)", t.Name, fType, startOffset+currentFieldOff)
		if isIndirectType(fType) {
			ref, err := encode(w, f)
			if err != nil {
				return err
			}
			w.patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}
		if f.Kind() == reflect.Struct {
			// Recurse to patch the inner struct's fields
			if err := patchStructBody(w, f, startOffset+currentFieldOff); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}
		if f.Kind() == reflect.Array {
			if err := patchArray(w, f, startOffset+currentFieldOff); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}
		w.patch(startOffset+currentFieldOff, f.Interface())
		currentFieldOff += getBinarySize(fType)
	}
	return nil
}

func patchArray(w *encoder, v reflect.Value, startOffset int64) error {
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)
	isReference := isIndirectType(elemType)

	// Fast Path: Bulk Patch for POD types
	// If we have a plain array (like [4]uint64 - field.Element), we can copy the bytes directly
	// into the buffer at the correct offset, avoiding the loop entirely.
	if isPOD(elemType) && v.CanAddr() {
		totalBytes := elemBinSize * int64(v.Len())

		// Get raw bytes from the Go value
		// Note: v.UnsafeAddr() gives us the pointer to the array in memory
		srcPtr := unsafe.Pointer(v.UnsafeAddr())

		// This "tricks" Go into treating that memory address as a slice of bytes ([]byte) of length totalBytes.
		// Crucially, this step does not copy any data yet; it just creates a window (handle) through which we
		// can read the raw bytes of the array. This is the core of the optimization. It creates a []byte that points
		// directly to your POD-array (eg. fr.Element) in memory.
		srcBytes := unsafe.Slice((*byte)(srcPtr), totalBytes)

		// Get the target slice in the buffer
		// Note: w.buf is a bytes.Buffer, but we need random access.
		// We use .Bytes() which returns the unread portion, but we need absolute access.
		// Ideally, w.buf.Bytes() gives us the slice.
		// Warning: This assumes w.buf.Bytes() returns the underlying slice starting at 0 if we haven't read.
		// A safer way in your existing code structure is to access w.buf.Bytes() and index relative to 0.
		bufSlice := w.buf.Bytes()

		if startOffset+totalBytes > int64(len(bufSlice)) {
			return fmt.Errorf("patchArray: buffer overflow during bulk copy")
		}

		// Copy directly into the buffer
		copy(bufSlice[startOffset:], srcBytes)
		return nil
	}

	// Fall back - sequential patch
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
		switch elem.Kind() {
		case reflect.Array:
			if err := patchArray(w, elem, offset); err != nil {
				return err
			}
		case reflect.Struct:
			if err := patchStructBody(w, elem, offset); err != nil {
				return err
			}
		default:
			w.patch(offset, elem.Interface())
		}
	}
	return nil
}

// Type size cache to avoid repeated computation of binary sizes
// map[reflect.Type]int64
var binarySizeCache sync.Map

func getBinarySize(t reflect.Type) int64 {
	if cached, ok := binarySizeCache.Load(t); ok {
		return cached.(int64)
	}

	size := computeBinarySize(t) // Move current logic here
	binarySizeCache.Store(t, size)
	return size
}

// computeBinarySize: returns how many bytes the serialized representation of type `t` takes.
// NOTE: This is NOT the same as Go's in-memory size. The size returned here reflects
// how the value is represented on disk:
//
// - Fixed-size, pointer-free values are inlined.
// - Variable-size or heap-backed values are replaced by an 8-byte Ref.
//
// This function MUST remain perfectly consistent with the actual write logic;
// any mismatch will result in corrupted offsets or incorrect deserialization.
func computeBinarySize(t reflect.Type) int64 {
	// Indirect types (custom registries, ptrs, slices, strings etc) are types that have variable sizes
	// not known at compile time and hence their inline representation is a Ref (8-byte offset).
	// These values are written elsewhere and referenced inline by offset.
	if isIndirectType(t) {
		return 8
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
		var sum int64
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if strings.Contains(f.Tag.Get(serdeStructTag), serdeStructTagOmit) {
				continue
			}
			sum += computeBinarySize(f.Type)
		}
		return sum
	}
	// Arrays are fixed-size and serialized inline as repeated elements.
	if k == reflect.Array {
		elemSize := computeBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}
	return int64(t.Size())
}

// isIndirectType returns true if the given type is an indirect type.
// Indirect types are types that have variable sizes not known at compile time.
// This includes pointers, slices, strings, interfaces, maps, and functions.
// Direct types are types that have fixed sizes known at compile time.
// This includes structs, arrays, and primitive types (bool,int/uint8, int/uint (normalized), etc).
// Types handled inside of the Custom Registry are also considered indirect.
func isIndirectType(t reflect.Type) bool {
	if _, ok := customRegistry[t]; ok {
		return true
	}
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

// isPOD returns true if the given type is a "Plain Old Data" type which are zero-copy safe.
// Plain Old Data types are types that have fixed sizes known at compile time.
func isPOD(t reflect.Type) bool {
	if isIndirectType(t) {
		return false
	}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			if !isPOD(t.Field(i).Type) {
				return false
			}
		}
		return true
	}
	if t.Kind() == reflect.Array {
		return isPOD(t.Elem())
	}
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
		return v.Interface()
	}
}
