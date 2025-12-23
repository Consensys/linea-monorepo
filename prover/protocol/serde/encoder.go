package serde

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

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
	// enc.writeBytes([]byte{0})
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
	return FileSlice{
		Offset: Ref(offset),
		Len:    int64(v.Len()),
		Cap:    int64(v.Cap()),
	}
}

// encode serializes a value into the writer buffer and returns a Ref (8-byte offset)
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

	// 1. Check Indirect Registry (Big Objects, Heap)
	if handler, ok := customIndirectRegistry[v.Type()]; ok {
		traceLog("Using Indirect Handler for %s", v.Type())
		return handler.marshall(w, v)
	}

	// 2. Check Direct Registry (Small Objects, Inline)
	// Even though they are inline, we might have linearize called on them via Interface{}.
	// We just write them and return Ref(offset).
	if handler, ok := customDirectRegistry[v.Type()]; ok {
		traceLog("Using Direct Handler for %s", v.Type())
		start := w.offset
		if err := handler.marshall(w, v); err != nil {
			return 0, err
		}
		return Ref(start), nil
	}

	// If the type implements ifaces.Query, we use UUID deduplication
	// We check t.Kind() != Interface to ensure we are looking at the concrete type
	t := v.Type()
	if t.Kind() != reflect.Interface && t.Implements(TypeOfQuery) {
		traceLog("Type %s implements Query -> using UUID Dedup", t)
		return marshallQueryViaUUID(w, v)
	}

	switch v.Kind() {
	// Handle standard Go Pointers - Deduplication happens in 'linearize' before this, so we just recurse.
	case reflect.Ptr:
		return encode(w, v.Elem())
	case reflect.Array:
		return linearizeArray(w, v)
	case reflect.Slice:
		if v.IsNil() {
			return 0, nil
		}

		elemType := v.Type().Elem()

		// 1. Indirect Elements (Ptr, Interface, String, etc.) -> Must serialize as list of Refs.
		if isIndirectType(elemType) {
			traceLog("Slice is Indirect -> writeSliceOfIndirects")
			return writeSliceOfIndirects(w, v)
		}

		// 2. NON-POD Elements (e.g., Structs containing pointers).
		if !isPod(elemType) {
			traceLog("Slice is Struct-With-Pointers -> linearizeSliceSeq")
			return linearizeSliceSeq(w, v)
		}

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
	q, ok := v.Interface().(ifaces.Query)
	if !ok {
		return 0, fmt.Errorf("value of type %s does not implement ifaces.Query", v.Type())
	}

	id := q.UUID().String()

	// 2. Check Logical Cache (UUID)
	if ref, ok := enc.uuidMap[id]; ok {
		traceLog("Query UUID Dedup Hit! %s -> Ref %d", id, ref)
		return ref, nil
	}

	// 3. Serialize
	var ref Ref
	var err error

	if v.Kind() == reflect.Ptr {
		elem := v
		for elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
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
func linearizeStruct(w *encoder, v reflect.Value) (Ref, error) {
	// PREDICT: Calculate total size of the struct in bytes.
	size := getBinarySize(v.Type())

	// SNAPSHOT: Capture the current cursor position.
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
		if !t.IsExported() {
			logrus.Warnf("field %v.%v is unexported", t.Type, t.Name)
			continue
		}

		// 2. Indirect Types
		if isIndirectType(t.Type) {
			ref, err := encode(w, f)
			if err != nil {
				return err
			}
			binary.Write(buf, binary.LittleEndian, ref)
			continue
		}

		// 3. Direct Types (Custom)
		// Since we are writing to a temp buffer, we use the Marshaller
		if handler, ok := customDirectRegistry[f.Type()]; ok {
			// Create a mini-encoder pointing to this buffer?
			// Simpler: The handler's Marshall uses w.write which uses w.buf.
			// Here we are writing to 'buf' (a separate buffer for map bodies).
			// We can assume direct types are "Direct" and just write bytes.
			// But to respect the handler logic (which might do casts), we need the bytes.
			// HACK: for this specific function (used in Maps/Arithmetization),
			// we rely on the fact that Direct types are usually fixed binary blobs.
			// We can invoke the handler on a temporary encoder.
			tmpEnc := &encoder{buf: buf, offset: 0}
			if err := handler.marshall(tmpEnc, f); err != nil {
				return err
			}
			continue
		}

		// 4. Structs (Recursive)
		if f.Kind() == reflect.Struct {
			if err := linearizeStructBodyMap(w, f, buf); err != nil {
				return err
			}
			continue
		}

		// 5. Primitives
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
		if !t.IsExported() {
			logrus.Warnf("field %v.%v is unexported", t.Type, t.Name)
			continue
		}

		traceLog("Patching Field: %s (Type: %s, Offset: %d)", t.Name, fType, startOffset+currentFieldOff)

		// 2. Handle References: Indirect types
		if isIndirectType(fType) {
			ref, err := encode(w, f)
			if err != nil {
				return err
			}
			w.patch(startOffset+currentFieldOff, ref)
			currentFieldOff += 8
			continue
		}

		// 3. Handle Direct Custom Types (Inline, Fixed Size, Special Logic)
		if handler, ok := customDirectRegistry[fType]; ok {
			if err := handler.patch(w, startOffset+currentFieldOff, f); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}

		// 4. Handle Nested Structs (Inline)
		if f.Kind() == reflect.Struct {
			// Recurse to patch the inner struct's fields
			if err := patchStructBody(w, f, startOffset+currentFieldOff); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}

		// 5. Handle Arrays (Inline)
		if f.Kind() == reflect.Array {
			if err := patchArray(w, f, startOffset+currentFieldOff); err != nil {
				return err
			}
			currentFieldOff += getBinarySize(fType)
			continue
		}

		// 6. Handle Primitives (int, uint, bool)
		w.patch(startOffset+currentFieldOff, f.Interface())
		currentFieldOff += getBinarySize(fType)
	}
	return nil
}

func patchArray(w *encoder, v reflect.Value, startOffset int64) error {
	elemType := v.Type().Elem()
	elemBinSize := getBinarySize(elemType)
	isReference := isIndirectType(elemType)

	// Check for Custom Direct Type for the element (e.g. Array of field.Element)
	directHandler, isDirect := customDirectRegistry[elemType]

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		offset := startOffset + (int64(i) * elemBinSize)

		if isReference {
			ref, err := encode(w, elem)
			if err != nil {
				return err
			}
			w.patch(offset, ref)
			continue
		}

		if isDirect {
			if err := directHandler.patch(w, offset, elem); err != nil {
				return err
			}
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

func encodeSeqItem(w *encoder, v reflect.Value, buf *bytes.Buffer) error {
	t := v.Type()

	if isIndirectType(t) {
		ref, err := encode(w, v)
		if err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, ref)
	}

	// Direct Custom Types (Inline) inside a Sequence
	if handler, ok := customDirectRegistry[t]; ok {
		// Create temp encoder to write to the sequence buffer
		tmpEnc := &encoder{buf: buf, offset: 0}
		return handler.marshall(tmpEnc, v)
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

// --- NEW HELPER FUNCTION ---
func writeSliceOfIndirects(w *encoder, v reflect.Value) (Ref, error) {
	refs := make([]Ref, v.Len())
	for i := 0; i < v.Len(); i++ {
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
	w.offset += int64(len(refs) * 8)

	fs := FileSlice{
		Offset: Ref(startOffset),
		Len:    int64(v.Len()),
		Cap:    int64(v.Cap()),
	}
	off := w.write(fs)
	return Ref(off), nil
}

func getBinarySize(t reflect.Type) int64 {

	// 1. Indirect types (Heap) -> always 8 bytes (Ref)
	if isIndirectType(t) {
		return 8
	}

	// 2. Direct Custom Types (Inline, known size)
	if handler, ok := customDirectRegistry[t]; ok {
		return handler.binSize(t)
	}

	k := t.Kind()

	if k == reflect.Int || k == reflect.Uint {
		return 8
	}

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
			sum += getBinarySize(f.Type)
		}
		return sum
	}

	if k == reflect.Array {
		elemSize := getBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}

	return int64(t.Size())
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

func linearizeSliceSeq(w *encoder, v reflect.Value) (Ref, error) {
	totalSize := int64(0)
	for i := 0; i < v.Len(); i++ {
		totalSize += getBinarySize(v.Index(i).Type())
	}

	startOffset := w.offset
	w.writeBytes(make([]byte, totalSize))

	currentOff := startOffset
	elemType := v.Type().Elem()
	directHandler, isDirect := customDirectRegistry[elemType]

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		if isDirect {
			if err := directHandler.patch(w, currentOff, elem); err != nil {
				return 0, err
			}
		} else {
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
// Indirect types are types that have variable sizes not known at compile time
// or are handled by the Indirect Custom Registry.
func isIndirectType(t reflect.Type) bool {
	if _, ok := customIndirectRegistry[t]; ok {
		return true
	}

	// Note: We do NOT check customDirectRegistry here.
	// If it's in Direct registry, it is NOT indirect.

	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

// isPod (Plain Old Data) returns true only if the type contains NO references.
func isPod(t reflect.Type) bool {
	if isIndirectType(t) {
		return false
	}

	// Direct Custom types (like field.Element) are POD if they are in the registry.
	if _, ok := customDirectRegistry[t]; ok {
		return true
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			if !isPod(t.Field(i).Type) {
				return false
			}
		}
		return true
	}

	if t.Kind() == reflect.Array {
		return isPod(t.Elem())
	}

	return true
}

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
