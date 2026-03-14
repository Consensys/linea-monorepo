package serde

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/sirupsen/logrus"
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

	// ancestors is a chain of read-only ptrMaps inherited from parent decoders
	// during parallel struct decode.  lookupPtr checks ptrMap first, then each
	// ancestor in order, giving O(1) amortised lookup with zero-copy inheritance
	// from the parent that ran the sequential "seed" phase.
	ancestors []map[int64]reflect.Value

	// depth tracks struct nesting; parallel fan-out is attempted at depth <= 2.
	depth int
}

// lookupPtr searches the decoder's own ptrMap, then walks the ancestor chain.
// Writes always go to the local ptrMap; ancestors are read-only.
func (dec *decoder) lookupPtr(offset int64) (reflect.Value, bool) {
	if val, ok := dec.ptrMap[offset]; ok {
		return val, true
	}
	for _, ancestor := range dec.ancestors {
		if val, ok := ancestor[offset]; ok {
			return val, true
		}
	}
	return reflect.Value{}, false
}

// parallelFieldThreshold is the minimum number of indirect (non-POD) fields a
// struct must have before we attempt to decode them in parallel.  For the
// execution-circuit, the top-level ZkEvm struct has 22 pointer fields and the
// CompiledIOP structs have 5-10 each; all well above this threshold.
const parallelFieldThreshold = 4

// fieldMeta describes a single exported struct field's position in the
// serialized data and its type characteristics.
type fieldMeta struct {
	index  int
	offset int64
	info   typeInfo
}

func (dec *decoder) decode(target reflect.Value, offset int64) error {
	traceEnter("DECODE", target)
	traceOffset("Reading From", offset)
	defer traceExit("DECODE", nil)

	t := target.Type()
	if handler, ok := customRegistry[t]; ok {
		traceLog("Using Custom Handler for %s", t)
		return handler.unmarshall(dec, target, offset)
	}

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
		return dec.decodePrimitive(target, offset)
	}
}

func (dec *decoder) decodePtr(target reflect.Value, offset int64) error {
	traceEnter("DEC_PTR", target)
	traceOffset("Ptr Ref", offset)
	defer traceExit("DEC_PTR", nil)
	if offset == 0 {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}
	t := target.Type()
	if val, ok := dec.lookupPtr(offset); ok {
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
	if int(fs.Offset) >= len(dec.data) {
		return fmt.Errorf("slice data offset out of bounds")
	}

	elemType := target.Type().Elem()
	info := getTypeInfo(elemType)
	// Indirect element handling: stored as array of Refs. We cannot use Zero-Copy here because
	// the file contains int64 Offsets (Refs), but the Go slice expects real memory pointers.
	if info.isIndirect {
		// Allocate a standard Go slice. The data at fs.Offset is an array of Refs (int64).
		// We iterate over them and resolve each one.
		target.Set(reflect.MakeSlice(target.Type(), int(fs.Len), int(fs.Cap)))
		refArrayStart := int64(fs.Offset)
		if refArrayStart+(fs.Len*8) > int64(len(dec.data)) {
			return fmt.Errorf("slice refs array out of bounds")
		}
		for i := 0; i < int(fs.Len); i++ {
			// Read the Ref at index i
			refOffset := refArrayStart + int64(i*8)
			ref := *(*Ref)(unsafe.Pointer(&dec.data[refOffset]))
			if !ref.IsNull() {
				if err := dec.decode(target.Index(i), int64(ref)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	// For non-indirect element types we need to decide between zero-copy swizzling and sequential decode.
	hasPtrs := false
	if elemType.Kind() == reflect.Struct || elemType.Kind() == reflect.Array {
		// Conservative assumption: treat complex structs/arrays as non-swizzle unless they are pure POD.
		if !info.isPOD {
			hasPtrs = true
		}
	}
	if hasPtrs {
		target.Set(reflect.MakeSlice(target.Type(), int(fs.Len), int(fs.Cap)))
		currentOff := int64(fs.Offset)
		for i := 0; i < int(fs.Len); i++ {
			nextOff, err := dec.decodeSeqItem(target.Index(i), currentOff)
			if err != nil {
				return err
			}
			currentOff = nextOff
		}
		return nil
	}

	// Bound-check
	totalBytes := int64(fs.Len) * info.binSize
	if int64(fs.Offset)+totalBytes > int64(len(dec.data)) {
		return fmt.Errorf("POD slice data (size %d) exceeds buffer bounds", totalBytes)
	}

	// ZERO-COPY swizzle into mmap bytes for POD element types.
	// For this, we get the pointer to the raw data inside the memory map
	// i.e. calculate the memory address of the data inside the file.
	dataPtr := unsafe.Pointer(&dec.data[fs.Offset])

	// Manually construct the slice header to point to this data
	// We cast the target's address to a struct that mimics the Go slice layout
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
}

// decodeMap: handles the reconstruction of Go maps which are complex internal hash table
// structures that cannot be zero-copied (unlike slices or arrays).
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
	traceEnter("DEC_IFACE", target)
	traceOffset("Header At", offset)
	defer traceExit("DEC_IFACE", nil)
	if offset < 0 || int(offset)+int(SizeOf[InterfaceHeader]()) > len(dec.data) {
		return fmt.Errorf("interface header out of bounds")
	}

	// InterfaceHeader inherently contains an Offset field. This Offset functions exactly like a Ref;
	// it tells the decoder precisely where the concrete data lives in the file.
	// NOTE: Whether the underlying concrete type is a simple int (direct) or a complex string (indirect), the
	// InterfaceHeader always points to the start of that data.
	ih := (*InterfaceHeader)(unsafe.Pointer(&dec.data[offset]))
	traceLog("Header Read: TypeID=%d Ind=%d Offset=%d", ih.TypeID, ih.PtrIndirection, ih.Offset)

	// Composite types (slice, map, array) are encoded with a sentinel TypeID.
	if ih.TypeID == compositeTypeID {
		return dec.decodeBoxedComposite(target, ih)
	}

	if int(ih.TypeID) >= len(IDToType) {
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
		traceLog("Offset is Null -> Setting Zero Value (Typed Nil) for %s", concreteType)
		// Create a zero value of the concrete type (e.g., (*MyType)(nil))
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
		if val, ok := dec.lookupPtr(ptrOffset); ok {
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

// resolveCompositeTypeField reconstructs a reflect.Type from the 2-byte TypeID
// sub-field stored in InterfaceHeader.Reserved for boxed composite descriptors.
// It is the inverse of encodeCompositeTypeField.
//
//   - compositeElemEmptyStruct (0xFFFF) → struct{}
//   - otherwise: bits 14–15 = pointer indirection, bits 0–13 = base TypeID
func resolveCompositeTypeField(field uint16) (reflect.Type, error) {
	if field == compositeElemEmptyStruct {
		return emptyStructType, nil
	}
	indirection := int(field >> compositeIndirectionShift)
	baseID := field & compositeTypeMask
	if int(baseID) >= len(IDToType) {
		return nil, fmt.Errorf("decodeBoxedComposite: base TypeID %d out of range (registry size %d)", baseID, len(IDToType))
	}
	t := IDToType[baseID]
	for i := 0; i < indirection; i++ {
		t = reflect.PointerTo(t)
	}
	return t, nil
}

// decodeBoxedComposite reconstructs a slice, map, array, or struct{} that was
// stored inside an interface value and encoded with the compositeTypeID sentinel.
// The type description is read from the InterfaceHeader.Reserved bytes:
//
//	Reserved[0]   = composite kind (compositeKindSlice / Map / Array / EmptyStruct)
//	Reserved[1:3] = primary TypeID field  (elem for slice/array; key for map; unused for EmptyStruct)
//	Reserved[3:5] = secondary field       (val TypeID field for map; raw array length; unused otherwise)
func (dec *decoder) decodeBoxedComposite(target reflect.Value, ih *InterfaceHeader) error {
	kind := ih.Reserved[0]

	// compositeKindEmptyStruct carries no type or data: just set the zero struct{}.
	if kind == compositeKindEmptyStruct {
		target.Set(reflect.Zero(emptyStructType))
		return nil
	}

	primaryField := uint16(ih.Reserved[1]) | uint16(ih.Reserved[2])<<8
	secondaryRaw := uint16(ih.Reserved[3]) | uint16(ih.Reserved[4])<<8

	primaryType, err := resolveCompositeTypeField(primaryField)
	if err != nil {
		return fmt.Errorf("decodeBoxedComposite primary: %w", err)
	}

	var compositeType reflect.Type
	switch kind {
	case compositeKindSlice:
		compositeType = reflect.SliceOf(primaryType)

	case compositeKindMap:
		valType, err := resolveCompositeTypeField(secondaryRaw)
		if err != nil {
			return fmt.Errorf("decodeBoxedComposite map value: %w", err)
		}
		compositeType = reflect.MapOf(primaryType, valType)

	case compositeKindArray:
		// secondaryRaw is the raw array length, not a TypeID field.
		compositeType = reflect.ArrayOf(int(secondaryRaw), primaryType)

	default:
		return fmt.Errorf("decodeBoxedComposite: unknown composite kind %d", kind)
	}

	if ih.Offset.IsNull() {
		target.Set(reflect.Zero(compositeType))
		return nil
	}

	concreteVal := reflect.New(compositeType).Elem()
	if err := dec.decode(concreteVal, int64(ih.Offset)); err != nil {
		return err
	}
	target.Set(concreteVal)
	return nil
}

// decodeArray: handles the reconstruction of Go arrays in a non-zero copy fashion.
//
// NOTE: Since an array's memory is part of the struct it belongs to, it is copied into the struct's memory.
// It is not "Zero-Copy" like a slice, which just points back to the file. This makes arrays safer for mutation
func (dec *decoder) decodeArray(target reflect.Value, offset int64) error {
	t := target.Type()
	info := getTypeInfo(t.Elem())
	elemSize := info.binSize
	totalSize := elemSize * int64(target.Len())
	if offset < 0 || offset+totalSize > int64(len(dec.data)) {
		return fmt.Errorf("array out of bounds")
	}

	// Optimization: Bulk Copy for POD types. If the element type is POD
	// (eg field.Element [4]uint64), we can copy the whole block at once.
	if info.isPOD {
		dstPtr := unsafe.Pointer(target.UnsafeAddr())
		srcPtr := unsafe.Pointer(&dec.data[offset])
		// unsafe.Slice allows us to treat the pointers as byte slices for the copy
		// and copy all bytes in one go
		copy(unsafe.Slice((*byte)(dstPtr), totalSize), unsafe.Slice((*byte)(srcPtr), totalSize))
		return nil
	}

	// Fallback: Sequential Decode (for pointers, strings, etc.)
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

func (dec *decoder) decodePrimitive(target reflect.Value, offset int64) error {
	size := int(target.Type().Size())
	if offset < 0 || int(offset)+size > len(dec.data) {
		return fmt.Errorf("primitive out of bounds")
	}
	var (
		srcPtr = unsafe.Pointer(&dec.data[offset])
		dstPtr = unsafe.Pointer(target.UnsafeAddr())
	)
	copy(unsafe.Slice((*byte)(dstPtr), size), unsafe.Slice((*byte)(srcPtr), size))
	return nil
}

// decodeStruct handles the reconstruction of Go structs.  It pre-computes
// every field's file offset, then chooses between sequential and parallel
// decode based on struct depth and the number of indirect fields.
//
// Parallel fan-out at depth <= 2:
//
//	depth 1 = ZkEvm struct:      InitialCompiledIOP decoded sequentially first
//	                               (seeds ptrMap), remaining 21 fields in parallel.
//	depth 2 = CompiledIOP struct: Columns decoded sequentially first (seeds
//	                               ptrMap), QueriesNoParams / SubProvers / etc
//	                               decoded in parallel.
//
// Each parallel goroutine inherits the parent's ptrMap via a read-only
// ancestor chain so it benefits from all previously decoded shared pointers.
func (dec *decoder) decodeStruct(target reflect.Value, offset int64) error {
	dec.depth++
	defer func() { dec.depth-- }()

	// ---- Pre-compute field offsets ----
	// We always need these to know where each field starts in the file,
	// regardless of whether we take the parallel or sequential path.
	numFields := target.NumField()
	fields := make([]fieldMeta, 0, numFields)
	currentOffSet := offset
	for i := 0; i < numFields; i++ {
		tf := target.Type().Field(i)
		if !tf.IsExported() || strings.Contains(tf.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}
		info := getTypeInfo(tf.Type)
		fields = append(fields, fieldMeta{index: i, offset: currentOffSet, info: info})
		currentOffSet += info.binSize
	}

	// ---- Choose parallel vs sequential path ----
	// Parallel fan-out at depth <= 2:
	//   depth 1 = ZkEvm struct:      first field (InitialCompiledIOP) sequential,
	//                                 remaining 21 fields in parallel.
	//   depth 2 = CompiledIOP struct: first field (Columns) sequential,
	//                                 QueriesNoParams / SubProvers / etc in parallel.
	indirectCount := 0
	for _, fm := range fields {
		if !fm.info.isPOD {
			indirectCount++
		}
	}

	if dec.depth <= 2 && indirectCount >= parallelFieldThreshold {
		return dec.decodeStructParallel(target, fields)
	}
	return dec.decodeStructSequential(target, fields)
}

// decodeStructSequential is the original field-by-field decode path.
func (dec *decoder) decodeStructSequential(target reflect.Value, fields []fieldMeta) error {
	for _, fm := range fields {
		tf := target.Type().Field(fm.index)
		f := target.Field(fm.index)

		if fm.info.isPOD && f.CanAddr() {
			dstPtr := unsafe.Pointer(f.UnsafeAddr())
			srcPtr := unsafe.Pointer(&dec.data[fm.offset])
			copy(unsafe.Slice((*byte)(dstPtr), fm.info.binSize),
				unsafe.Slice((*byte)(srcPtr), fm.info.binSize))
			continue
		}

		if _, err := dec.decodeSeqItem(f, fm.offset); err != nil {
			return fmt.Errorf("failed to decode field '%s': %w", tf.Name, err)
		}
	}
	return nil
}

// decodeStructParallel uses a two-phase strategy:
//
//  1. POD fields are bulk-copied (fast memcpy).
//  2. The FIRST indirect field is decoded sequentially with the current decoder,
//     seeding the ptrMap with shared data (e.g. Columns in CompiledIOP,
//     InitialCompiledIOP in ZkEvm).
//  3. All remaining indirect fields are decoded in parallel.  Each goroutine
//     gets its own decoder whose ancestor chain includes the parent's ptrMap,
//     giving it read-only access to every pointer decoded so far.  New pointers
//     go into each child's local ptrMap — no locks required.
func (dec *decoder) decodeStructParallel(target reflect.Value, fields []fieldMeta) error {
	// Phase 1 — POD fields: fast, no allocation.
	for _, fm := range fields {
		if !fm.info.isPOD {
			continue
		}
		f := target.Field(fm.index)
		if f.CanAddr() {
			dstPtr := unsafe.Pointer(f.UnsafeAddr())
			srcPtr := unsafe.Pointer(&dec.data[fm.offset])
			copy(unsafe.Slice((*byte)(dstPtr), fm.info.binSize),
				unsafe.Slice((*byte)(srcPtr), fm.info.binSize))
		}
	}

	// Collect indirect fields.
	indirectFields := make([]fieldMeta, 0, len(fields))
	for _, fm := range fields {
		if !fm.info.isPOD {
			indirectFields = append(indirectFields, fm)
		}
	}
	if len(indirectFields) == 0 {
		return nil
	}

	// Phase 2 — First indirect field: decode sequentially to seed the ptrMap
	// with shared objects that the remaining fields will reference.
	{
		first := indirectFields[0]
		tf := target.Type().Field(first.index)
		f := target.Field(first.index)
		if _, err := dec.decodeSeqItem(f, first.offset); err != nil {
			return fmt.Errorf("failed to decode field '%s': %w", tf.Name, err)
		}
		logrus.Infof("[serde] field %s.%s decoded (sequential-seed)", target.Type().Name(), tf.Name)
		indirectFields = indirectFields[1:]
	}
	if len(indirectFields) == 0 {
		return nil
	}

	// Phase 3 — Remaining indirect fields in parallel.
	// Build the ancestor chain for children: parent's ptrMap + parent's ancestors.
	childAncestors := make([]map[int64]reflect.Value, 0, 1+len(dec.ancestors))
	childAncestors = append(childAncestors, dec.ptrMap)
	childAncestors = append(childAncestors, dec.ancestors...)

	type errResult struct {
		fieldName string
		err       error
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		errList []errResult
	)

	for _, fm := range indirectFields {
		wg.Add(1)
		go func(fm fieldMeta) {
			defer wg.Done()

			tf := target.Type().Field(fm.index)

			// Each goroutine gets its own decoder with a fresh local ptrMap
			// and read-only access to the parent's accumulated entries via
			// the ancestor chain.
			child := &decoder{
				data:      dec.data,
				ptrMap:    make(map[int64]reflect.Value, 1024),
				ancestors: childAncestors,
				depth:     dec.depth, // inherit depth so nested structs can also parallelize
			}

			f := target.Field(fm.index)
			if _, err := child.decodeSeqItem(f, fm.offset); err != nil {
				mu.Lock()
				errList = append(errList, errResult{tf.Name, err})
				mu.Unlock()
			}

			logrus.Infof("[serde] field %s.%s decoded (parallel)", target.Type().Name(), tf.Name)
		}(fm)
	}

	wg.Wait()

	if len(errList) > 0 {
		first := errList[0]
		return fmt.Errorf("failed to decode field '%s': %w", first.fieldName, first.err)
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
	if offset < 0 || int(offset) >= len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: offset %d out of bounds (len: %d)", offset, len(dec.data))
	}
	t := target.Type()
	info := getTypeInfo(t)
	// For in-direct types, the map data section only stores an 8-byte Reference (offset).
	// The function reads that offset, jumps to that location to decode the actual object,
	// and moves the map cursor forward by exactly 8 bytes.
	if info.isIndirect {
		if int(offset)+8 > len(dec.data) {
			return 0, fmt.Errorf("decodeSeqItem: unable to read Ref at offset %d", offset)
		}
		ref := *(*Ref)(unsafe.Pointer(&dec.data[offset]))
		if !ref.IsNull() {
			if err := dec.decode(target, int64(ref)); err != nil {
				return 0, err
			}
		} else {
			target.Set(reflect.Zero(target.Type()))
		}
		return offset + 8, nil
	}
	// If the type is "direct" (size known at compile time), the data is stored "inline" within the map's data block.
	// The function decodes it at the current position and moves the cursor forward by the size of that type.
	binSize := info.binSize
	if int(offset)+int(binSize) > len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: unable to read binary data of size %d at offset %d", binSize, offset)
	}
	if err := dec.decode(target, offset); err != nil {
		return 0, err
	}
	return offset + binSize, nil
}
