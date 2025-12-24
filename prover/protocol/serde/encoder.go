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

// ---------------------------------------------------------------------------
// Encoder
// ---------------------------------------------------------------------------

// encoder: Holds the current encoding/serialization state
type encoder struct {
	buf     *bytes.Buffer
	offset  int64
	ptrMap  map[uintptr]Ref
	uuidMap map[string]Ref
	idMap   map[string]Ref
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

	// Reserve Offset 0 for NULL references explicitly.
	// enc.writeBytes([]byte{0})
	return enc
}

// write writes the given data (primitive, struct, header types) to the buffer and returns
// the start offset where the data was written.
func (w *encoder) write(data any) int64 {
	start := w.offset
	val := normalizeIntegerSize(reflect.ValueOf(data))
	if err := binary.Write(w.buf, binary.LittleEndian, val); err != nil {
		panic(fmt.Errorf("binary.Write failed for type %T: %w", val, err))
	}
	w.offset += int64(binary.Size(val))
	return start
}

func (w *encoder) writeBytes(b []byte) int64 {
	start := w.offset
	w.buf.Write(b)
	w.offset += int64(len(b))
	return start
}

// patch overwrites previously reserved zero bytes (at offset) with actual data v.
func (w *encoder) patch(offset int64, v any) {
	var tmp bytes.Buffer
	val := normalizeIntegerSize(reflect.ValueOf(v))
	binary.Write(&tmp, binary.LittleEndian, val)
	encoded := tmp.Bytes()
	bufSlice := w.buf.Bytes()
	if int(offset)+len(encoded) > len(bufSlice) {
		panic(fmt.Errorf("patch out of bounds"))
	}
	copy(bufSlice[offset:], encoded)
}

// writeSliceData writes raw bytes of a POD slice into the buffer and returns a FileSlice
// describing where the data lives. Unsafe raw-copy — only valid for POD element types.
func (w *encoder) writeSliceData(v reflect.Value) FileSlice {
	if v.Len() == 0 {
		return FileSlice{0, 0, 0}
	}
	totalBytes := v.Len() * int(v.Type().Elem().Size())
	dataPtr := unsafe.Pointer(v.Pointer())
	dataBytes := unsafe.Slice((*byte)(dataPtr), totalBytes)
	offset := w.offset
	w.buf.Write(dataBytes)
	w.offset += int64(totalBytes)
	return FileSlice{Offset: Ref(offset), Len: int64(v.Len()), Cap: int64(v.Cap())}
}

// encode: entry point that handles pointer deduplication and circular refs.
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
		ptrAddr = v.Pointer()
		if ref, ok := w.ptrMap[ptrAddr]; ok {
			traceLog("Encode: Dedup Hit! Addr %x -> Ref %d", ptrAddr, ref)
			return ref, nil
		}
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
		return marshallQueryViaUUID(w, v)
	}

	switch v.Kind() {
	case reflect.Ptr:
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
			return writeSliceOfIndirects(w, v)
		}
		if !isPod(elemType) {
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

var TypeOfQuery = reflect.TypeOf((*ifaces.Query)(nil)).Elem()

func marshallQueryViaUUID(enc *encoder, v reflect.Value) (Ref, error) {
	q, ok := v.Interface().(ifaces.Query)
	if !ok {
		return 0, fmt.Errorf("value of type %s does not implement ifaces.Query", v.Type())
	}
	id := q.UUID().String()
	if ref, ok := enc.uuidMap[id]; ok {
		traceLog("Query UUID Dedup Hit! %s -> Ref %d", id, ref)
		return ref, nil
	}

	// Linearize the underlying value (struct or pointer to struct) using existing helpers.
	var ref Ref
	var err error
	if v.Kind() == reflect.Ptr {
		// Dereference pointers until we reach the concrete
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
	enc.uuidMap[id] = ref
	return ref, nil
}

// linearizeStruct: Reserve then Patch model for structs. Returns Ref pointing to start.
func linearizeStruct(w *encoder, v reflect.Value) (Ref, error) {
	size := getBinarySize(v.Type())
	startOffset := w.offset
	w.writeBytes(make([]byte, size))
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
	fs := FileSlice{Offset: Ref(startOffset), Len: int64(v.Len()), Cap: int64(v.Cap())}
	off := w.write(fs)
	return Ref(off), nil
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

// getBinarySize: returns how many bytes the serialized representation of `t` takes.
func getBinarySize(t reflect.Type) int64 {
	if isIndirectType(t) {
		return 8
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

func isIndirectType(t reflect.Type) bool {
	if _, ok := customRegistry[t]; ok {
		return true
	}
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

func isPod(t reflect.Type) bool {
	if isIndirectType(t) {
		return false
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
