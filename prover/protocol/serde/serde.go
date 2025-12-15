// File: serde/serde.go
package serde

import (
	"math/big"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
)

var (
	SerdeStructTag = "serde"

	// Do not serialize fields with this tag
	SerdeStructTagOmit = "omit"

	// Serialize this but don't include it in test comparisions to prevent OOM
	SerdeStructTagTestOmit = "test_omit"
)

// Global type constants for reflection-based type checking.
// These define the reflect.Type of key protocol-specific types, used to identify
// special types during serialization and deserialization.
var (
	TypeOfColumnNatural = reflect.TypeOf(column.Natural{})
)

func Serialize(v any) ([]byte, error) {
	w := NewWriter()
	_ = w.Write(FileHeader{})
	rootOff, err := linearize(w, reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	finalHeader := FileHeader{
		Magic:       Magic,
		Version:     1,
		PayloadType: 0,
		PayloadOff:  int64(rootOff),
		DataSize:    w.offset,
	}
	b := w.buf.Bytes()
	*(*FileHeader)(unsafe.Pointer(&b[0])) = finalHeader
	return b, nil
}

func getBinarySize(t reflect.Type) int64 {
	// --- FIX: Check Custom Registry First ---
	// If a type has a custom handler, it is serialized as a Reference (8 bytes)
	if _, ok := CustomRegistry[t]; ok {
		return 8
	}

	if t == reflect.TypeOf((*frontend.Variable)(nil)).Elem() {
		return 8
	}
	if t == reflect.TypeOf(big.Int{}) || t == reflect.TypeOf(&big.Int{}) {
		return 8
	}

	k := t.Kind()
	if k == reflect.Ptr || k == reflect.Slice ||
		k == reflect.String || k == reflect.Interface || k == reflect.Map ||
		k == reflect.Func { // Added Func
		return 8
	}

	if k == reflect.Struct {
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}
		var sum int64
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if strings.Contains(f.Tag.Get("serde"), "omit") {
				continue
			}
			sum += getBinarySize(f.Type)
		}
		return sum
	}

	// Array handling (recurse for elements)
	if k == reflect.Array {
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}
		elemSize := getBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}

	if k == reflect.Int || k == reflect.Uint {
		return 8
	}

	return int64(t.Size())
}
