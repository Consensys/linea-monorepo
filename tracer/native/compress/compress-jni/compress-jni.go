package main

import "C"
import (
	"errors"
	"os"
	"sync"
	"unsafe"

	"github.com/consensys/compress/lzss"
)

var (
	compressor *lzss.Compressor
	lastError  error // last error that occurred
	lock       sync.Mutex
)

const compressionLevel = lzss.BestCompression

// Init initializes the compressor.
// Returns true if the compressor was initialized, false otherwise.
// If false is returned, the Error() method will return a string describing the error.
//
//export Init
func Init(dictPath *C.char) bool {
	fPath := C.GoString(dictPath)
	return initGo(fPath)
}

func initGo(dictPath string) bool {
	lock.Lock()
	defer lock.Unlock()

	// read the dictionary
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		lastError = err
		return false
	}

	compressor, lastError = lzss.NewCompressor(dict, compressionLevel)

	return lastError == nil
}

// Compress compresses the input and returns the length of the compressed data.
// If an error occurred, returns -1.
// User must call Error() to get the error message.
//
// This function is thread-safe.
//
//export CompressedSize
func CompressedSize(input *C.char, inputLength C.int) C.int {
	inputSlice := C.GoBytes(unsafe.Pointer(input), inputLength)

	n, err := compressor.CompressedSize256k(inputSlice)
	if err != nil {
		lock.Lock()
		lastError = errors.Join(lastError, err)
		lock.Unlock()
		return -1
	}
	if n > len(inputSlice) {
		// this simulates the fallback to "no compression"
		// this case may happen if the input is not compressible
		// in which case the compressed size is the input size + the header size
		n = len(inputSlice) + lzss.HeaderSize
	}
	return C.int(n)
}

// Error returns the last encountered error.
// If no error was encountered, returns nil.
//
//export Error
func Error() *C.char {
	lock.Lock()
	defer lock.Unlock()
	if lastError != nil {
		// this leaks memory, but since this represents a fatal error, it's probably ok.
		return C.CString(lastError.Error())
	}
	return nil
}

func main() {}
