// File: serde/serde.go
package serde

import (
	"reflect"
	"unsafe"
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
