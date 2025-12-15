package serde

import "unsafe"

const Magic = 0x5A45524F

type FileHeader struct {
	Magic       uint32
	Version     uint32
	PayloadType uint64
	PayloadOff  int64
	DataSize    int64
}

// InterfaceHeader must export fields for binary.Write
type InterfaceHeader struct {
	TypeID      uint16
	Indirection uint8
	Reserved    [5]uint8 // Explicit padding to 16 bytes
	Offset      Ref
}

type Ref int64

func (r Ref) IsNull() bool { return r == 0 }

type FileSlice struct {
	Offset Ref
	Len    int64
	Cap    int64
}

func SizeOf[T any]() int64 {
	var z T
	return int64(unsafe.Sizeof(z))
}
