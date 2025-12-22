package serde

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// decoder maps the continguous block into memory and "swizzles" (converts) file offsets into real memory addresses
// containing the actual payload data.
type decoder struct {

	// The entire memory-mapped file - the raw binary blob. The reader does not "read" from a stream,
	// rather it jumps around this byte slice using offsets.
	data []byte

	// The Deduplication Table - critical for maintaining referential integrity
	// For example: If object A and object B both point to object C, the file only stores object C once.
	// When the reader encounters the offset for C the first time, it deserializes it and stores the result in ptrMap.
	// When it encounters the offset for C the second time, it grabs the existing object from ptrMap.
	// Benefit: This preserves cycles (A -> B -> A) and saves memory.
	ptrMap map[int64]reflect.Value
}

func (dec *decoder) decode(target reflect.Value, offset int64) error {
	// DEBUG: Trace
	traceEnter("DECODE", target)
	traceOffset("Reading From", offset)
	defer traceExit("DECODE", nil)

	t := target.Type()

	// 1. Custom Registry
	if handler, ok := customRegistry[t]; ok {
		traceLog("Using Custom Handler for %s", t)
		return handler.unmarshall(dec, target, offset)
	}

	// 2. Main Type Switch
	switch target.Kind() {
	case reflect.Ptr:
		return dec.decodePtr(target, offset)
	case reflect.Slice:
		return dec.decodeSlice(target, offset)
	case reflect.String:
		return dec.decodeString(target, offset)
	case reflect.Map:
		return dec.decodeMap(target, offset)
	case reflect.Interface:
		return dec.decodeInterface(target, offset)
	case reflect.Struct:
		// Fallback for field.Element if not in registry
		if t == reflect.TypeOf(field.Element{}) {
			return dec.decodeFieldElement(target, offset)
		}
		return dec.decodeStruct(target, offset)
	case reflect.Array:
		return dec.decodeArray(target, offset)
	case reflect.Int, reflect.Int64:
		if offset+8 > int64(len(dec.data)) {
			return fmt.Errorf("int out of bounds")
		}
		val64 := *(*int64)(unsafe.Pointer(&dec.data[offset]))
		target.SetInt(val64)
		return nil
	case reflect.Uint, reflect.Uint64:
		if offset+8 > int64(len(dec.data)) {
			return fmt.Errorf("uint out of bounds")
		}
		val64 := *(*uint64)(unsafe.Pointer(&dec.data[offset]))
		target.SetUint(val64)
		return nil
	default:
		// Generic copy for other primitives (bool, byte, etc)
		return dec.decodePrimitive(target, offset)
	}
}

// --- Helpers ---

func (dec *decoder) decodePtr(target reflect.Value, offset int64) error {
	// DEBUG: Trace
	traceEnter("DEC_PTR", target)
	traceOffset("Ptr Ref", offset)
	defer traceExit("DEC_PTR", nil)

	t := target.Type()

	// Check Cache
	if val, ok := dec.ptrMap[offset]; ok {
		traceLog("Cache Hit for Ref %d", offset)
		if val.Type().AssignableTo(t) {
			target.Set(val)
			return nil
		}
	}

	// Generic Pointer - allocates fresh memory for the type the pointer points to (the "element").
	// i.e newPtr is of type `*t` where type `t` is the target's type - see first line, and allocates
	// memory to store type `t`, and saves it to the ptrMap (deduplication purposes).
	newPtr := reflect.New(t.Elem())
	traceLog("Allocated New Ptr: %v (Addr: %p) for Ref %d", newPtr.Type(), newPtr.Interface(), offset)
	dec.ptrMap[offset] = newPtr

	// Reconstruct recursively to actually fill that newly allocated memory with the data found at that
	// offset in the file.
	target.Set(newPtr)
	return dec.decode(newPtr.Elem(), offset)
}

/*
func (dec *decoder) decodeSlice(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("slice header out of bounds")
	}

	// Read the File Header First by looking at the offset to find a `FileSlice` struct.
	// This is a small header stored in the binary file that tells us three things:
	// Where the actual data is (the Offset).
	// How much data there is (the Len and Cap)
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))

	// If the fs.Offset is marked as null, it sets the target slice to its zero value (a nil slice)
	// and exits early.
	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// Verify bounds of the underlying data start
	if int(fs.Offset) >= len(dec.data) {
		return fmt.Errorf("slice data offset out of bounds")
	}

	// --- FIX START ---
	// Check if we are decoding an Indirect Slice (e.g. []*T)
	if isIndirectType(target.Type().Elem()) {
		// We cannot use Zero-Copy here because the file contains int64 Offsets (Refs),
		// but the Go slice expects real memory pointers.

		// 1. Allocate a standard Go slice
		target.Set(reflect.MakeSlice(target.Type(), int(fs.Len), int(fs.Cap)))

		// 2. The data at fs.Offset is an array of Refs (int64).
		// We iterate over them and resolve each one.
		refArrayStart := int64(fs.Offset)

		// Safety check for the ref array bounds
		if refArrayStart+(fs.Len*8) > int64(len(dec.data)) {
			return fmt.Errorf("slice refs array out of bounds")
		}

		for i := 0; i < int(fs.Len); i++ {
			// Read the Ref at index i
			refOffset := refArrayStart + int64(i*8)
			ref := *(*Ref)(unsafe.Pointer(&dec.data[refOffset]))

			// If not null, recursively decode the object it points to
			if !ref.IsNull() {
				if err := dec.decode(target.Index(i), int64(ref)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	// --- FIX END ---

	// ZERO-COPY (The pointer "Swizzling" technique) SLICE
	// 1. Get the pointer to the raw data inside the memory map
	// i.e. calculate the memory address of the data inside the file
	dataPtr := unsafe.Pointer(&dec.data[fs.Offset])

	// 2. Manually construct the slice header to point to this data
	// We cast the target's address to a struct that mimics the Go slice layout.
	// This works for all slice types.
	sh := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(target.UnsafeAddr()))

	// Manually overwrite the slice fields to make the Go runtime think our slice is a normal slice,
	// but its underlying data is actually the memory-mapped file itself. This unsafe usage is permitted.
	sh.Data = uintptr(dataPtr)
	sh.Len = int(fs.Len)
	sh.Cap = int(fs.Cap)
	return nil
} */

// decodeMap: handles the reconstruction of Go maps which are complex internal hash table
// structures that cannot be zero-copied (unlike slices or arrays).
/*
func (dec *decoder) decodeMap(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("map header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// Since, we cannot zero-copy maps, and unlike arrays(value types, where memory is automatically allocated by Go compiler),
	// maps are referential types and we must allocate a new map in heap memory and insert every key-value pair individually.
	target.Set(reflect.MakeMapWithSize(target.Type(), int(fs.Len)))
	currentOff := int64(fs.Offset)
	keyType := target.Type().Key()
	ValType := target.Type().Elem()

	// In each iteration, we decode the corresponding key and value (call decodeMapElement twice),
	// and insert them into the map
	for i := 0; i < int(fs.Len); i++ {
		newKey := reflect.New(keyType).Elem()
		nextOff, err := dec.decodeSeqItem(newKey, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff

		newVal := reflect.New(ValType).Elem()
		nextOff, err = dec.decodeSeqItem(newVal, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
		target.SetMapIndex(newKey, newVal)
	}
	return nil
} */

// decodeMap: handles the reconstruction of Go maps which are complex internal hash table
// structures that cannot be zero-copied (unlike slices or arrays).
func (dec *decoder) decodeMap(target reflect.Value, offset int64) error {
	// 1. CHECK CACHE with TYPE SAFETY
	if val, ok := dec.ptrMap[offset]; ok {
		// Only use the cached value if it matches the expected type.
		// This handles the "Aliasing" case where a Custom Handler (Type A)
		// reuses the offset for its internal representation (Type B).
		// We want to skip the cache for Type B to avoid a panic.
		if val.Type().AssignableTo(target.Type()) {
			target.Set(val)
			return nil
		}
	}

	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("map header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))

	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// 2. CREATE MAP
	newMap := reflect.MakeMapWithSize(target.Type(), int(fs.Len))

	// 3. REGISTER (Safe)
	// Only register if the slot is empty. If it's occupied (by a different type),
	// we assume the existing occupant is the "Primary" object (e.g. the Struct wrapping this Map)
	// and this Map is just a transient view. Overwriting it would break cycles for the Primary object.
	if _, exists := dec.ptrMap[offset]; !exists {
		dec.ptrMap[offset] = newMap
	}

	// Set the target immediately
	target.Set(newMap)

	currentOff := int64(fs.Offset)
	keyType := target.Type().Key()
	ValType := target.Type().Elem()

	// 4. FILL
	for i := 0; i < int(fs.Len); i++ {
		newKey := reflect.New(keyType).Elem()
		nextOff, err := dec.decodeSeqItem(newKey, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff

		newVal := reflect.New(ValType).Elem()
		nextOff, err = dec.decodeSeqItem(newVal, currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff

		newMap.SetMapIndex(newKey, newVal)
	}
	return nil
}

func (dec *decoder) decodeSlice(target reflect.Value, offset int64) error {
	// 1. CHECK CACHE with TYPE SAFETY
	if val, ok := dec.ptrMap[offset]; ok {
		if val.Type().AssignableTo(target.Type()) {
			target.Set(val)
			return nil
		}
		// If types mismatch, we ignore the cache hit and proceed to decode (Unwrapping scenario)
	}

	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("slice header out of bounds")
	}

	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))

	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	if int(fs.Offset) >= len(dec.data) {
		return fmt.Errorf("slice data offset out of bounds")
	}

	// CASE A: INDIRECT SLICE
	if isIndirectType(target.Type().Elem()) {
		newSlice := reflect.MakeSlice(target.Type(), int(fs.Len), int(fs.Cap))

		// Register (Safe)
		if _, exists := dec.ptrMap[offset]; !exists {
			dec.ptrMap[offset] = newSlice
		}

		target.Set(newSlice)

		refArrayStart := int64(fs.Offset)
		if refArrayStart+(fs.Len*8) > int64(len(dec.data)) {
			return fmt.Errorf("slice refs array out of bounds")
		}

		for i := 0; i < int(fs.Len); i++ {
			refOffset := refArrayStart + int64(i*8)
			ref := *(*Ref)(unsafe.Pointer(&dec.data[refOffset]))

			if !ref.IsNull() {
				if err := dec.decode(newSlice.Index(i), int64(ref)); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// CASE B: ZERO-COPY SLICE
	dataPtr := unsafe.Pointer(&dec.data[fs.Offset])
	sh := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(target.UnsafeAddr()))

	sh.Data = uintptr(dataPtr)
	sh.Len = int(fs.Len)
	sh.Cap = int(fs.Cap)

	// Register (Safe) - even zero-copy slices might be referenced in cycles/DeepCmp
	if _, exists := dec.ptrMap[offset]; !exists {
		dec.ptrMap[offset] = target
	}

	return nil
}
func (dec *decoder) decodeString(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("string header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		return nil
	}

	start := int64(fs.Offset)
	end := start + int64(fs.Len)
	if start < 0 || end > int64(len(dec.data)) {
		return fmt.Errorf("string content out of bounds")
	}

	// NOTE Std. string(bytes) copies. For zero-copy string construction - we' use unsafe.String
	// Since Go strings are immutable, the bytes passed to String must not be modified as long as
	// the returned string value exists.
	target.SetString(unsafe.String(&dec.data[start], fs.Len))
	return nil
}
func (dec *decoder) decodeInterface(target reflect.Value, offset int64) error {
	// DEBUG: Trace
	traceEnter("DEC_IFACE", target)
	traceOffset("Header At", offset)
	defer traceExit("DEC_IFACE", nil)

	if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(dec.data) {
		return fmt.Errorf("interface header out of bounds")
	}

	// InterfaceHeader inherently contains an Offset field. This Offset functions exactly like a Ref;
	// it tells the decoder precisely where the concrete data lives in the file. Whether the underlying
	// concrete type is a simple int (direct) or a complex string (indirect), the InterfaceHeader always
	// points to the start of that data. Hence, there is no need to distinguish between direct and indirect
	// types here unlike decodeMapElement.
	ih := (*InterfaceHeader)(unsafe.Pointer(&dec.data[offset]))

	// DEBUG: Log Header
	traceLog("Header Read: TypeID=%d Ind=%d Offset=%d", ih.TypeID, ih.PtrIndirection, ih.Offset)

	if int(ih.TypeID) < 0 || int(ih.TypeID) >= len(IDToType) {
		return fmt.Errorf("invalid type ID: %d", ih.TypeID)
	}
	concreteType := IDToType[ih.TypeID]
	for i := 0; i < int(ih.PtrIndirection); i++ {
		concreteType = reflect.PointerTo(concreteType)
	}

	// NOTE: We do NOT return early if ih.Offset.IsNull().
	// A null offset in a header simply means the *value* is nil (e.g. a nil pointer),
	// but the *type* information in the header is still valid and required.
	if ih.Offset.IsNull() {
		// Create a zero value of the concrete type (e.g., (*MyType)(nil))
		traceLog("Offset is Null -> Setting Zero Value (Typed Nil) for %s", concreteType)
		typedNil := reflect.Zero(concreteType)
		target.Set(typedNil)
		return nil
	}

	var concreteVal reflect.Value
	if concreteType.Kind() == reflect.Ptr {
		concreteVal = reflect.New(concreteType.Elem())
		ptrOffset := int64(ih.Offset)
		// Check cache for this specific interface pointer - ensures that if the interface points to an object
		// already reconstructed elsewhere, we reuse that instance (preserving referential integrity).
		if val, ok := dec.ptrMap[ptrOffset]; ok {
			if val.Type().AssignableTo(concreteType) {
				target.Set(val)
				return nil
			}
		}
		dec.ptrMap[ptrOffset] = concreteVal
		if err := dec.decode(concreteVal.Elem(), ptrOffset); err != nil {
			return err
		}
	} else {
		concreteVal = reflect.New(concreteType).Elem()
		if err := dec.decode(concreteVal, int64(ih.Offset)); err != nil {
			return err
		}
	}
	target.Set(concreteVal)
	return nil
}

// decodeArray: handles the reconstruction of Go arrays in a non-zero copy fashion.
//
// NOTE: Since an array's memory is part of the struct it belongs to, it is copied into the struct's memory.
// It is not "Zero-Copy" like a slice, which just points back to the file. This makes arrays safer for mutation
// but more expensive for very large datasets.
func (dec *decoder) decodeArray(target reflect.Value, offset int64) error {
	t := target.Type()

	// Calculate the size of single element and then the total size
	elemSize := getBinarySize(t.Elem())
	totalSize := elemSize * int64(target.Len())
	if offset < 0 || offset+totalSize > int64(len(dec.data)) {
		return fmt.Errorf("array out of bounds")
	}
	// Special "fast path" optimization
	if t.Elem() == reflect.TypeOf(field.Element{}) {
		return dec.decodeFieldElement(target, offset)
	}

	// Loop through the array using `decodeSeqItem` to "step" through the file bytes accurately,
	// even if the array elements are complex types like strings or pointers.
	currentOff := offset
	for i := 0; i < target.Len(); i++ {
		nextOff, err := dec.decodeSeqItem(target.Index(i), currentOff)
		if err != nil {
			return err
		}
		currentOff = nextOff
	}
	return nil
}

// decodeFieldElement: handles the reconstruction of field.Element ([4]uint64) in a non-zero copy fashion,
// but does a bulk-memory copy instead of iterating through each element. This is significantly faster
// than iterating through each element individually.
func (dec *decoder) decodeFieldElement(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(dec.data) {
		return fmt.Errorf("field element out of bounds")
	}
	var (
		srcPtr = unsafe.Pointer(&dec.data[offset])
		dstPtr = unsafe.Pointer(target.UnsafeAddr())
	)
	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

func (dec *decoder) decodePrimitive(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(dec.data) {
		return fmt.Errorf("primitive out of bounds")
	}

	// Fix WARNING: Cast uintptr -> unsafe.Pointer -> *byte
	var (
		srcPtr = unsafe.Pointer(&dec.data[offset])
		dstPtr = unsafe.Pointer(target.UnsafeAddr())
	)
	copy(
		unsafe.Slice((*byte)(dstPtr), size),
		unsafe.Slice((*byte)(srcPtr), size),
	)
	return nil
}

// decodeStruct: handles the reconstruction of Go structs. Since Structs are "Heterogenous"
// collection of fields, we loop through the struct schema (i.e. type definition) and
// decode each field accordingly as per `decodeSeqItem`.
func (dec *decoder) decodeStruct(target reflect.Value, offset int64) error {
	currentOffSet := offset
	t := target.Type()

	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		tf := t.Field(i)
		// Ignore unexported fields or fields with the "omit" tag
		if !tf.IsExported() || strings.Contains(tf.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}

		// Decode the current field and then calculate the offset for the next field
		// and then swap it with the `currentOffset`.
		nextOffset, err := dec.decodeSeqItem(f, currentOffSet)
		if err != nil {
			// Contextualize the error to help debugging
			return fmt.Errorf("failed to decode field '%s': %w", tf.Name, err)
		}
		currentOffSet = nextOffset
	}
	return nil
}

// decodeSeqItem: reads a single item (direct or indirect) and returns the
// offset of the next item so the caller knows where the next piece of data begins.
// Essentially, it decodes a single value at 'offset' and returns the offset
// for the next item in the sequence.
// NOTE: The core logic of this function is to handle the "stepping" of the offset while deciding
// whether to read data inline or follow a reference. Hence, explicit handling of direct and indirect
// types are required here even though the `getBinarySize()` returns the correct sizes for both types.
// This is because the `decode(target, X)` expects X to be the start of the object's header/data.
// For Direct types (e.g., int), the object sits right there at offset and hence can be directly decoded.
// For Indirect types (e.g., string), the actual object sits far away and the offset only contains a
// "signpost" (Ref) pointing to it.
//
// See `encodeSeqItem` mirroring the encoding logic.
func (dec *decoder) decodeSeqItem(target reflect.Value, offset int64) (int64, error) {
	// Bound check
	if offset < 0 || int(offset) >= len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: offset %d out of bounds (len: %d)", offset, len(dec.data))
	}

	t := target.Type()

	// For in-direct types, the map data section only stores an 8-byte Reference (offset).
	// The function reads that offset, jumps to that location to decode the actual object,
	// and moves the map cursor forward by exactly 8 bytes.
	if isIndirectType(t) {
		if int(offset)+8 > len(dec.data) {
			return 0, fmt.Errorf("decodeSeqItem: unable to read Ref at offset %d", offset)
		}
		ref := *(*Ref)(unsafe.Pointer(&dec.data[offset]))
		if !ref.IsNull() {
			if err := dec.decode(target, int64(ref)); err != nil {
				return 0, err
			}
		}
		return offset + 8, nil
	}

	// If the type is "direct" (size known at compile time), the data is stored "inline" within the map's data block.
	// The function decodes it at the current position and moves the cursor forward by the size of that type.
	binSize := getBinarySize(t)
	if int(offset)+int(binSize) > len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: unable to read binary data of size %d at offset %d", binSize, offset)
	}
	if err := dec.decode(target, offset); err != nil {
		return 0, err
	}
	return offset + binSize, nil
}
