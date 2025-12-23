package serde

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// decoder maps the continguous block into memory and "swizzles" (converts) file offsets into real memory addresses
type decoder struct {
	data   []byte
	ptrMap map[int64]reflect.Value
}

func (dec *decoder) decode(target reflect.Value, offset int64) error {
	// DEBUG: Trace
	traceEnter("DECODE", target)
	traceOffset("Reading From", offset)
	defer traceExit("DECODE", nil)

	t := target.Type()

	// 1. Indirect Custom Registry (Refs)
	if handler, ok := customIndirectRegistry[t]; ok {
		traceLog("Using Indirect Handler for %s", t)
		return handler.unmarshall(dec, target, offset)
	}

	// 2. Direct Custom Registry (Inline)
	if handler, ok := customDirectRegistry[t]; ok {
		traceLog("Using Direct Handler for %s", t)
		return handler.unmarshall(dec, target, offset)
	}

	// 3. Main Type Switch
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

// --- Helpers ---

func (dec *decoder) decodePtr(target reflect.Value, offset int64) error {
	traceEnter("DEC_PTR", target)
	traceOffset("Ptr Ref", offset)
	defer traceExit("DEC_PTR", nil)

	if offset == 0 {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	t := target.Type()

	// Check Cache
	if val, ok := dec.ptrMap[offset]; ok {
		traceLog("Cache Hit for Ref %d", offset)
		if val.Type().AssignableTo(t) {
			target.Set(val)
			return nil
		}
	}

	newPtr := reflect.New(t.Elem())
	traceLog("Allocated New Ptr: %v (Addr: %p) for Ref %d", newPtr.Type(), newPtr.Interface(), offset)
	dec.ptrMap[offset] = newPtr

	target.Set(newPtr)
	return dec.decode(newPtr.Elem(), offset)
}

func (dec *decoder) decodeSlice(target reflect.Value, offset int64) error {
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

	// 1. Indirect Elements -> Slice of Refs
	if isIndirectType(target.Type().Elem()) {
		target.Set(reflect.MakeSlice(target.Type(), int(fs.Len), int(fs.Cap)))
		refArrayStart := int64(fs.Offset)

		if refArrayStart+(fs.Len*8) > int64(len(dec.data)) {
			return fmt.Errorf("slice refs array out of bounds")
		}

		for i := 0; i < int(fs.Len); i++ {
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

	// 2. Non-POD Slices (Structs with Ptrs) -> Sequential Decode
	elemType := target.Type().Elem()
	hasPtrs := false
	if !isPod(elemType) {
		hasPtrs = true
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

	// 3. ZERO-COPY (POD) -> Swizzling
	dataPtr := unsafe.Pointer(&dec.data[fs.Offset])

	sh := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(target.UnsafeAddr()))

	sh.Data = uintptr(dataPtr)
	sh.Len = int(fs.Len)
	sh.Cap = int(fs.Cap)
	return nil
}

func (dec *decoder) decodeMap(target reflect.Value, offset int64) error {
	if offset < 0 || int(offset)+int(SizeOf[FileSlice]()) > len(dec.data) {
		return fmt.Errorf("map header out of bounds")
	}
	fs := (*FileSlice)(unsafe.Pointer(&dec.data[offset]))
	if fs.Offset.IsNull() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

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
	if start < 0 || start+fs.Len > int64(len(dec.data)) {
		return fmt.Errorf("string content out of bounds")
	}

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

	ih := (*InterfaceHeader)(unsafe.Pointer(&dec.data[offset]))
	traceLog("Header Read: TypeID=%d Ind=%d Offset=%d", ih.TypeID, ih.PtrIndirection, ih.Offset)

	if int(ih.TypeID) < 0 || int(ih.TypeID) >= len(IDToType) {
		return fmt.Errorf("invalid type ID: %d", ih.TypeID)
	}
	concreteType := IDToType[ih.TypeID]
	for i := 0; i < int(ih.PtrIndirection); i++ {
		concreteType = reflect.PointerTo(concreteType)
	}

	if ih.Offset.IsNull() {
		traceLog("Offset is Null -> Setting Zero Value (Typed Nil) for %s", concreteType)
		typedNil := reflect.Zero(concreteType)
		target.Set(typedNil)
		return nil
	}

	var concreteVal reflect.Value
	if concreteType.Kind() == reflect.Ptr {
		concreteVal = reflect.New(concreteType.Elem())
		ptrOffset := int64(ih.Offset)
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

func (dec *decoder) decodeArray(target reflect.Value, offset int64) error {
	t := target.Type()

	elemSize := getBinarySize(t.Elem())
	totalSize := elemSize * int64(target.Len())
	if offset < 0 || offset+totalSize > int64(len(dec.data)) {
		return fmt.Errorf("array out of bounds")
	}

	// Optimization check: If element is POD (including Direct Custom), we might be able to fast path?
	// For simplicity, we use decodeSeqItem which handles custom direct logic.
	// But if it is field.Element, we want the fast unsafe copy if possible.
	// Check Direct Registry:
	if _, ok := customDirectRegistry[t.Elem()]; ok {
		// Even if direct, we loop, because Unmarshall might do something logic-heavy.
		// However, for field.Element specifically, the unmarshaller calls decodeFieldElement which is fast copy.
	}

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

func (dec *decoder) decodeStruct(target reflect.Value, offset int64) error {
	currentOffSet := offset
	t := target.Type()

	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		tf := t.Field(i)
		if !tf.IsExported() || strings.Contains(tf.Tag.Get(serdeStructTag), serdeStructTagOmit) {
			continue
		}

		nextOffset, err := dec.decodeSeqItem(f, currentOffSet)
		if err != nil {
			return fmt.Errorf("failed to decode field '%s': %w", tf.Name, err)
		}
		currentOffSet = nextOffset
	}
	return nil
}

func (dec *decoder) decodeSeqItem(target reflect.Value, offset int64) (int64, error) {
	if offset < 0 || int(offset) >= len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: offset %d out of bounds (len: %d)", offset, len(dec.data))
	}

	t := target.Type()

	// 1. Indirect Custom Types (Refs)
	if isIndirectType(t) {
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

	// 2. Direct Custom Types (Inline)
	if handler, ok := customDirectRegistry[t]; ok {
		// We need to know the size to advance the offset
		binSize := handler.binSize(t)
		if int(offset)+int(binSize) > len(dec.data) {
			return 0, fmt.Errorf("decodeSeqItem: unable to read Direct Custom Type of size %d at offset %d", binSize, offset)
		}
		if err := handler.unmarshall(dec, target, offset); err != nil {
			return 0, err
		}
		return offset + binSize, nil
	}

	// 3. Standard Direct Types (Inline)
	binSize := getBinarySize(t)
	if int(offset)+int(binSize) > len(dec.data) {
		return 0, fmt.Errorf("decodeSeqItem: unable to read binary data of size %d at offset %d", binSize, offset)
	}
	if err := dec.decode(target, offset); err != nil {
		return 0, err
	}
	return offset + binSize, nil
}
