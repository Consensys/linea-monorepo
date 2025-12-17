// File: serde/serde.go
package serde

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
)

var (
	SerdeStructTag         = "serde"
	SerdeStructTagOmit     = "omit"
	SerdeStructTagTestOmit = "test_omit"
)

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

func Deserialize(b []byte, v any) error {
	if len(b) < int(SizeOf[FileHeader]()) {
		return fmt.Errorf("buffer too small")
	}
	header := (*FileHeader)(unsafe.Pointer(&b[0]))
	if header.Magic != Magic {
		return fmt.Errorf("invalid magic bytes")
	}

	if Ref(header.PayloadOff).IsNull() {
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val.Elem().Set(reflect.Zero(val.Elem().Type()))
		}
		return nil
	}

	if Ref(header.PayloadOff) > Ref(len(b)) {
		return fmt.Errorf("payload offset out of bounds")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	ctx := &ReaderContext{
		data:   b,
		ptrMap: make(map[int64]reflect.Value),
	}

	return ctx.reconstruct(val.Elem(), int64(header.PayloadOff))
}

func getBinarySize(t reflect.Type) int64 {
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
		k == reflect.Func {
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
