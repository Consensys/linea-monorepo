package serde

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const Magic = 0x5A45524F

// File Layout
// FileHeader describes the fixed-size header at the beginning of a serialized file.
// It provides enough metadata to validate the file, determine how to interpret the
// payload, and locate the serialized data region.
//
// All fields are written in a stable binary format and must remain backward-compatible
// across versions.
// IMPORTANT: DO NOT CHANGE THE LAYOUT (ordering of the fields) of this struct.
type FileHeader struct {
	// Magic is a constant identifier used to quickly validate that the file
	// is of the expected format (e.g. to reject random or corrupted input).
	Magic uint32

	// Version specifies the serialization format version.
	// It allows the reader to handle backward/forward compatibility
	// and apply version-specific decoding logic.
	Version uint32

	// PayloadType identifies the logical type of the serialized payload.
	// This is typically an application-defined enum or type ID that tells
	// the deserializer how to interpret the root object.
	PayloadType uint64

	// PayloadOff is the byte offset (from the start of the file/buffer)
	// where the payload begins.
	// This allows the header to be fixed-size while the payload layout evolves.
	PayloadOff int64

	// DataSize is the total size in bytes of the serialized payload data.
	// It can be used for bounds checking, mmap sizing, and integrity validation.
	DataSize int64
}

// InterfaceHeader represents the binary header for an interface and is specifically designed to be
// exactly 16 bytes to ensure the Offset field remains 8-byte aligned for efficient reading/writing.
//
// IMPORTANT:
// DO NOT CHANGE THE LAYOUT (ordering of the fields) of this struct.
// We perform `binary.Write` which is very literal i.e. writes the bytes of the struct fields in exact order and
// does not automatically insert padding for us like the Go compiler does automatically (inserts hidden padding bytes
// between fields in memory to align them) for structs in memory.
// We have to explicitly put the padding there if the file format requires a fixed size (like 16 bytes).
type InterfaceHeader struct {
	// 2 bytes: Unique identifier for the concrete type
	TypeID uint16

	// 1 byte: Number of pointer dereferences (e.g., ***T)
	PtrIndirection uint8

	// Reserved: 5 bytes of explicit padding.
	// Together with TypeID (2) and Indirection (1), these 5 bytes ensure
	// that the 'Offset' field starts exactly at the 8th byte.
	// This allows the 64-bit 'Ref' to be naturally aligned.
	// Total header size: 2 + 1 + 5 + 8 = 16 bytes.
	Reserved [5]uint8

	// 8 bytes: File or memory offset to the actual data
	Offset Ref
}

// Ref is a 8-byte offset in the serialized buffer
type Ref int64

func (r Ref) IsNull() bool { return r == 0 }

// FileSlice mirrors the struct of a Go slice header {Data uintptr, Len int, Cap int} conceptually
// for zero-copy slice reconstruction. Instead of Data (uintptr), we store the offset of the slice data
// in the serialized buffer. Notice the first two attributes shared similarties with Go string header
// representation {Data uintptr, Len int} and hence can also be used for zero-copy string construction.
//
// IMPORTANT: The ordering of the attributes MUST NOT be changed (Offset, Len, Cap) since the decoder relies on this format.
type FileSlice struct {
	// 8 Byte offset in the serialized buffer - beginning of the cursor
	Offset Ref

	// Number of elements in the slice
	Len int64

	// Original capacity (used to restore slice header)
	Cap int64
}

func SizeOf[T any]() int64 {
	var z T
	return int64(unsafe.Sizeof(z))
}

// databyte encapsulates the byte slice for the mapped memory.
type databyte struct {
	data []byte
}

// MappedFile represents a read-only memory-mapped file.
type MappedFile struct {
	file *os.File
	databyte
}

// openMappedFile opens a file and maps it into memory.
func openMappedFile(path string) (*MappedFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	size := info.Size()
	if size == 0 {
		f.Close()
		return &MappedFile{file: f, databyte: databyte{data: nil}}, nil
	}

	// PROT_READ: The memory is read-only to prevent accidental corruption.
	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("mmap failed: %w", err)
	}

	mf := &MappedFile{file: f, databyte: databyte{data: data}}
	runtime.SetFinalizer(mf, (*MappedFile).Close)
	return mf, nil
}

// Close unmaps the memory and closes the file descriptor.
func (mf *MappedFile) Close() error {
	if mf.data == nil {
		return mf.file.Close()
	}

	runtime.SetFinalizer(mf, nil)

	if err := syscall.Munmap(mf.data); err != nil {
		return fmt.Errorf("munmap failed: %w", err)
	}
	mf.data = nil
	return mf.file.Close()
}
