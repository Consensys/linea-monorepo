// File: serde/serde.go
package serde

import (
	"reflect"
	"unsafe"

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
