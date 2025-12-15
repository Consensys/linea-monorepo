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

// ... [Writer struct and NewWriter remain unchanged] ...

type Writer struct {
	buf    *bytes.Buffer
	offset int64
}

func NewWriter() *Writer {
	return &Writer{
		buf:    new(bytes.Buffer),
		offset: 0,
	}
}

func (w *Writer) Write(data any) int64 {
	start := w.offset

	// Normalize ints
	switch v := data.(type) {
	case int:
		data = int64(v)
	case uint:
		data = uint64(v)
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

	// 1. Pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, nil
		}
		if v.Type() == reflect.TypeOf(&big.Int{}) {
			return writeBigInt(w, v.Interface().(*big.Int))
		}
		return linearize(w, v.Elem())
	}

	// 2. Field Elements
	if v.Type() == reflect.TypeOf(field.Element{}) {
		off := w.Write(v.Interface())
		return Ref(off), nil
	}

	// 3. Big Ints
	if v.Type() == reflect.TypeOf(big.Int{}) {
		if v.CanAddr() {
			return writeBigInt(w, v.Addr().Interface().(*big.Int))
		}
		bi := v.Interface().(big.Int)
		return writeBigInt(w, &bi)
	}

	// 4. Slices
	if v.Kind() == reflect.Slice {
		if v.IsNil() {
			return 0, nil
		}
		fs := w.WriteSlice(v)
		off := w.Write(fs)
		return Ref(off), nil
	}

	// 5. Strings
	if v.Kind() == reflect.String {
		str := v.String()
		dataOff := w.WriteBytes([]byte(str))
		fs := FileSlice{Offset: Ref(dataOff), Len: int64(len(str)), Cap: int64(len(str))}
		off := w.Write(fs)
		return Ref(off), nil
	}

	// 6. Interfaces
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

	// 7. Maps
	if v.Kind() == reflect.Map {
		if v.IsNil() {
			return 0, nil
		}

		var bodyBuf bytes.Buffer
		iter := v.MapRange()
		count := 0

		for iter.Next() {
			k := iter.Key()
			val := iter.Value()

			// Write Key
			if err := writeMapElement(w, k, &bodyBuf); err != nil {
				return 0, fmt.Errorf("failed to write map key: %w", err)
			}
			// Write Value
			if err := writeMapElement(w, val, &bodyBuf); err != nil {
				return 0, fmt.Errorf("failed to write map value: %w", err)
			}
			count++
		}

		// Write data blob
		dataOff := w.WriteBytes(bodyBuf.Bytes())

		// Write Header
		fs := FileSlice{
			Offset: Ref(dataOff),
			Len:    int64(count),
			Cap:    int64(count),
		}
		off := w.Write(fs)
		return Ref(off), nil
	}

	// 8. Structs
	if v.Kind() == reflect.Struct {
		var bodyBuf bytes.Buffer
		if err := linearizeStructBody(w, v, &bodyBuf); err != nil {
			return 0, err
		}
		off := w.WriteBytes(bodyBuf.Bytes())
		return Ref(off), nil
	}

	// 9. Primitives
	off := w.Write(v.Interface())
	return Ref(off), nil
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

	isRef := v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice ||
		v.Kind() == reflect.String || v.Kind() == reflect.Interface || v.Kind() == reflect.Map ||
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

	// Handle Array of POD or fallback
	if v.Kind() == reflect.Array {
		if t == reflect.TypeOf(field.Element{}) {
			return binary.Write(buf, binary.LittleEndian, v.Interface())
		}
		// Treat arrays like structs for inline writing
		for j := 0; j < v.Len(); j++ {
			if err := binary.Write(buf, binary.LittleEndian, v.Index(j).Interface()); err != nil {
				return fmt.Errorf("array element write failed: %w", err)
			}
		}
		return nil
	}

	// Primitives
	val := v.Interface()
	switch v := val.(type) {
	case int:
		val = int64(v)
	case uint:
		val = uint64(v)
	}
	return binary.Write(buf, binary.LittleEndian, val)
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

func linearizeStructBody(w *Writer, v reflect.Value, buf *bytes.Buffer) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)

		if !t.IsExported() {
			continue
		}
		if strings.Contains(t.Tag.Get("serde"), "omit") {
			continue
		}

		if err := writeMapElement(w, f, buf); err != nil {
			return fmt.Errorf("field %s: %w", t.Name, err)
		}
	}
	return nil
}
