package serde

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	cs "github.com/consensys/gnark/constraint"
	"github.com/consensys/linea-monorepo/prover/protocol/serde/core"
)

type View struct {
	data []byte
}

func NewView(data []byte) (*View, error) {
	if len(data) < int(unsafe.Sizeof(core.Header{})) {
		return nil, fmt.Errorf("file too short")
	}
	return &View{data: data}, nil
}

func (v *View) Deserialize(out any) error {
	header := core.BytesToStruct[core.Header](v.data)
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer")
	}
	return v.unmarshal(header.RootOffset, val.Elem())
}

func (v *View) unmarshal(off core.Offset, dest reflect.Value) error {
	if off == 0 {
		return nil
	}

	// 1. CUSTOM CODEX
	if codex, ok := getCodex(dest.Type()); ok {
		return codex.Des(v, off, dest)
	}

	// 2. CIRCUIT HANDLING
	if sys, ok := dest.Interface().(cs.ConstraintSystem); ok {
		sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
		blob := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
		reader := bytes.NewReader(blob)
		if _, err := sys.ReadFrom(reader); err != nil {
			return fmt.Errorf("failed to read circuit: %w", err)
		}
		return nil
	}

	switch dest.Kind() {
	case reflect.Slice:
		sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
		length := int(sh.Len)

		// Optimization: Zero Copy for Field Elements
		if strings.Contains(dest.Type().Elem().String(), "field.Element") {
			elemSize := int(dest.Type().Elem().Size())
			totalSize := length * elemSize
			blob := v.data[sh.Offset : sh.Offset+core.Offset(totalSize)]
			dest.Set(reflect.MakeSlice(dest.Type(), length, length))
			srcPtr := unsafe.Pointer(&blob[0])
			sliceHeaderPtr := dest.Addr().UnsafePointer()
			*(*unsafe.Pointer)(sliceHeaderPtr) = srcPtr
			return nil
		}

		dest.Set(reflect.MakeSlice(dest.Type(), length, length))
		offsetTable := v.data[sh.Offset:]
		for i := 0; i < length; i++ {
			elemOff := *(*core.Offset)(unsafe.Pointer(&offsetTable[i*core.SizeOfOffset]))
			if err := v.unmarshal(elemOff, dest.Index(i)); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
		length := int(sh.Len)
		dest.Set(reflect.MakeMapWithSize(dest.Type(), length))
		offsetTable := v.data[sh.Offset:]
		keyType := dest.Type().Key()
		valType := dest.Type().Elem()
		keySlot := reflect.New(keyType).Elem()
		valSlot := reflect.New(valType).Elem()
		for i := 0; i < length; i++ {
			start := i * 2 * core.SizeOfOffset
			keyOff := *(*core.Offset)(unsafe.Pointer(&offsetTable[start]))
			valOff := *(*core.Offset)(unsafe.Pointer(&offsetTable[start+core.SizeOfOffset]))
			keySlot.Set(reflect.Zero(keyType))
			valSlot.Set(reflect.Zero(valType))
			if err := v.unmarshal(keyOff, keySlot); err != nil {
				return err
			}
			if err := v.unmarshal(valOff, valSlot); err != nil {
				return err
			}
			dest.SetMapIndex(keySlot, valSlot)
		}
		return nil

	case reflect.Struct:
		n := dest.NumField()
		fieldOffsets := v.data[off:]
		for i := 0; i < n; i++ {
			if !dest.Type().Field(i).IsExported() {
				continue
			}
			fieldOff := *(*core.Offset)(unsafe.Pointer(&fieldOffsets[i*core.SizeOfOffset]))
			if err := v.unmarshal(fieldOff, dest.Field(i)); err != nil {
				return err
			}
		}
		return nil

	case reflect.Ptr:
		// Special handling for Pointer to Circuit
		if dest.Type().Elem().String() == "cs.SparseR1CS" {
			dest.Set(reflect.New(dest.Type().Elem()))
			sys := dest.Interface().(cs.ConstraintSystem)
			sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
			blob := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
			reader := bytes.NewReader(blob)
			if _, err := sys.ReadFrom(reader); err != nil {
				return fmt.Errorf("failed to read circuit: %w", err)
			}
			return nil
		}
		dest.Set(reflect.New(dest.Type().Elem()))
		return v.unmarshal(off, dest.Elem())

	case reflect.Interface:
		ifHeader := core.BytesToStruct[core.InterfaceHeader](v.data[off:])
		typ, ok := getTypeByID(ifHeader.TypeID)
		if !ok {
			return fmt.Errorf("unknown type id: %d", ifHeader.TypeID)
		}
		concrete := reflect.New(typ).Elem()
		if err := v.unmarshal(ifHeader.Offset, concrete); err != nil {
			return err
		}
		dest.Set(concrete)
		return nil

	case reflect.String:
		sh := core.BytesToStruct[core.SliceHeader](v.data[off:])
		strBytes := v.data[sh.Offset : sh.Offset+core.Offset(sh.Len)]
		dest.SetString(string(strBytes))
		return nil

	// Primitives
	case reflect.Int, reflect.Int64:
		val := core.ByteOrder.Uint64(v.data[off : off+8])
		dest.SetInt(int64(val))
		return nil
	case reflect.Uint, reflect.Uint64:
		val := core.ByteOrder.Uint64(v.data[off : off+8])
		dest.SetUint(val)
		return nil
	case reflect.Int32:
		val := core.ByteOrder.Uint32(v.data[off : off+4])
		dest.SetInt(int64(int32(val)))
		return nil
	case reflect.Uint32:
		val := core.ByteOrder.Uint32(v.data[off : off+4])
		dest.SetUint(uint64(val))
		return nil
	case reflect.Int16:
		val := core.ByteOrder.Uint16(v.data[off : off+2])
		dest.SetInt(int64(int16(val)))
		return nil
	case reflect.Uint16:
		val := core.ByteOrder.Uint16(v.data[off : off+2])
		dest.SetUint(uint64(val))
		return nil
	case reflect.Int8:
		val := v.data[off]
		dest.SetInt(int64(int8(val)))
		return nil
	case reflect.Uint8:
		val := v.data[off]
		dest.SetUint(uint64(val))
		return nil
	case reflect.Bool:
		val := v.data[off]
		dest.SetBool(val == 1)
		return nil
	}

	return nil
}
