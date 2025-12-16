package serde

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"syscall"
	"unsafe"
)

const Magic = 0x5A45524F

// File Layout
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

// databyte encapsulates the byte slice for the mapped memory.
type databyte struct {
	data []byte
}

// MappedFile represents a read-only memory-mapped file.
type MappedFile struct {
	file *os.File
	databyte
}

// OpenMappedFile opens a file and maps it into memory.
func OpenMappedFile(path string) (*MappedFile, error) {
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

// UnsafeCastSlice reinterprets the raw bytes at offset as a slice of type T.
// Note: This requires Go 1.18+ generics support.
// The syntax below is designed to be executable in modern Go.
func UnsafeCastSlice[T any](mf *MappedFile, offset int64, count int) ([]T, error) {
	var zero T

	// Use reflect.TypeOf to safely get size of the element type.
	elemSize := int(reflect.TypeOf(zero).Size())
	totalBytes := elemSize * count

	// FIX: Corrected the boundary check logic and syntax.
	if offset < 0 || int(offset)+totalBytes > len(mf.data) {
		return nil, errors.New("cast out of bounds")
	}

	// Get pointer to the start of the data within the mmap region
	ptr := unsafe.Pointer(&mf.data[offset])

	// Create slice header using Go 1.17+ unsafe.Slice (best practice)
	return unsafe.Slice((*T)(ptr), count), nil
}
