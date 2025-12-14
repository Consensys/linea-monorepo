package serde

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/serde/core"
)

type Codex struct {
	Ser func(w *Writer, v reflect.Value) (core.Offset, error)
	Des func(v *View, off core.Offset, dest reflect.Value) error
}

var (
	codexRegistry = make(map[reflect.Type]Codex)
	codexLock     sync.RWMutex
)

func RegisterCodex(t reflect.Type, c Codex) {
	codexLock.Lock()
	defer codexLock.Unlock()
	codexRegistry[t] = c
}

func getCodex(t reflect.Type) (Codex, bool) {
	codexLock.RLock()
	defer codexLock.RUnlock()
	c, ok := codexRegistry[t]
	return c, ok
}

func init() {
	// 1. Field Element (Raw Bytes)
	RegisterCodex(reflect.TypeOf(field.Element{}), Codex{
		Ser: func(w *Writer, v reflect.Value) (core.Offset, error) {
			if !v.CanAddr() {
				tmp := reflect.New(v.Type()).Elem()
				tmp.Set(v)
				v = tmp
			}
			size := int(v.Type().Size())
			ptr := unsafe.Pointer(v.UnsafeAddr())
			bytes := unsafe.Slice((*byte)(ptr), size)
			return w.writeBytes(bytes), nil
		},
		Des: func(v *View, off core.Offset, dest reflect.Value) error {
			size := int(dest.Type().Size())
			srcBytes := v.data[off : int(off)+size]
			dstPtr := unsafe.Pointer(dest.UnsafeAddr())
			copy(unsafe.Slice((*byte)(dstPtr), size), srcBytes)
			return nil
		},
	})

	// 2. Big Int
	RegisterCodex(reflect.TypeOf(big.Int{}), Codex{
		Ser: func(w *Writer, v reflect.Value) (core.Offset, error) {
			bi := v.Interface().(big.Int)
			b := bi.Bytes()
			dataOff := w.writeBytes(b)
			return w.writeSliceHeader(uint32(len(b)), dataOff), nil
		},
		Des: func(v *View, off core.Offset, dest reflect.Value) error {
			sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
			blob := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
			bi := new(big.Int).SetBytes(blob)
			dest.Set(reflect.ValueOf(*bi))
			return nil
		},
	})

	// 3. Frontend Variable
	RegisterCodex(reflect.TypeOf(frontend.Variable(nil)), Codex{
		Ser: func(w *Writer, v reflect.Value) (core.Offset, error) {
			bi := new(big.Int)
			switch val := v.Interface().(type) {
			case int:
				bi.SetInt64(int64(val))
			case uint64:
				bi.SetUint64(val)
			case *big.Int:
				bi = val
			case big.Int:
				bi = &val
			case field.Element:
				neg := new(field.Element).Neg(&val)
				if neg.IsUint64() {
					bi.SetInt64(-int64(neg.Uint64()))
				} else {
					val.BigInt(bi)
				}
			default:
				return 0, fmt.Errorf("unsupported frontend.Variable type: %T", val)
			}
			b := bi.Bytes()
			dataOff := w.writeBytes(b)
			return w.writeSliceHeader(uint32(len(b)), dataOff), nil
		},
		Des: func(v *View, off core.Offset, dest reflect.Value) error {
			sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
			blob := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
			bi := new(big.Int).SetBytes(blob)
			dest.Set(reflect.ValueOf(bi))
			return nil
		},
	})

	// 4. Gnark FFT Domain
	RegisterCodex(reflect.TypeOf(fft.Domain{}), Codex{
		Ser: func(w *Writer, v reflect.Value) (core.Offset, error) {
			domain := v.Interface().(fft.Domain)
			var buf bytes.Buffer
			if _, err := domain.WriteTo(&buf); err != nil {
				return 0, err
			}
			dataOff := w.writeBytes(buf.Bytes())
			return w.writeSliceHeader(uint32(buf.Len()), dataOff), nil
		},
		Des: func(v *View, off core.Offset, dest reflect.Value) error {
			sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
			blob := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
			dom := dest.Addr().Interface().(*fft.Domain)
			_, err := dom.ReadFrom(bytes.NewReader(blob))
			return err
		},
	})
}
