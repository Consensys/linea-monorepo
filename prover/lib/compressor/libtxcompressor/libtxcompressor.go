package main

import "C"

import (
	"sync"
	"unsafe"

	blob_v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
)

//go:generate go build -tags nocorset -ldflags "-s -w" -buildmode=c-shared -o libtxcompressor.so libtxcompressor.go
func main() {}

var (
	txCompressor *blob_v1.TxCompressor
	lastError    error
	lock         sync.Mutex
)

// TxInit initializes the transaction compressor.
// The dataLimit argument is the maximum size of the compressed data.
// The caller should account for blob overhead (~100 bytes) when setting this limit.
// enableRecompress controls whether the compressor attempts recompression when
// incremental compression exceeds the limit. Set to false for faster operation.
// Returns true if the compressor was initialized, false otherwise.
// If false is returned, the TxError() method will return a string describing the error.
//
//export TxInit
func TxInit(dataLimit int, dictPath *C.char, enableRecompress bool) bool {
	fPath := C.GoString(dictPath)
	return txInitGo(dataLimit, fPath, enableRecompress)
}

func txInitGo(dataLimit int, dictPath string, enableRecompress bool) bool {
	lock.Lock()
	defer lock.Unlock()
	txCompressor, lastError = blob_v1.NewTxCompressor(dataLimit, dictPath, enableRecompress)

	return lastError == nil
}

// TxReset resets the compressor to its initial state.
// Must be called between each block being built.
//
//export TxReset
func TxReset() {
	lock.Lock()
	defer lock.Unlock()

	lastError = nil

	txCompressor.Reset()
}

// TxWriteRaw appends pre-encoded transaction data to the compressed data.
// The input should be: from address (20 bytes) + RLP-encoded transaction for signing.
// This is the fast path that avoids RLP decoding and signature recovery on the Go side.
// The Go code doesn't keep a pointer to the input slice and the caller is free to modify it.
// Returns true if the transaction was appended, false if it would exceed the limit.
//
//export TxWriteRaw
func TxWriteRaw(input *C.char, inputLength C.int) (txAppended bool) {
	lock.Lock()
	defer lock.Unlock()
	txData := unsafe.Slice((*byte)(unsafe.Pointer(input)), inputLength)
	txAppended, err := txCompressor.WriteRaw(txData, false)
	if err != nil {
		lastError = err
		return false
	}

	return txAppended
}

// TxCanWriteRaw checks if pre-encoded transaction data can be appended without actually appending it.
// The input should be: from address (20 bytes) + RLP-encoded transaction for signing.
// Returns true if the transaction could be appended, false otherwise.
//
//export TxCanWriteRaw
func TxCanWriteRaw(input *C.char, inputLength C.int) (canAppend bool) {
	lock.Lock()
	defer lock.Unlock()
	txData := unsafe.Slice((*byte)(unsafe.Pointer(input)), inputLength)
	canAppend, err := txCompressor.CanWriteRaw(txData)
	if err != nil {
		lastError = err
		return false
	}

	return canAppend
}

// TxLen returns the current length of the compressed data.
//
//export TxLen
func TxLen() (length int) {
	lock.Lock()
	defer lock.Unlock()
	return txCompressor.Len()
}

// TxWritten returns the number of uncompressed bytes written to the compressor.
//
//export TxWritten
func TxWritten() (written int) {
	lock.Lock()
	defer lock.Unlock()
	return txCompressor.Written()
}

// TxBytes returns the compressed data.
// The caller is responsible for allocating the memory for the output slice.
// Length of the output slice must equal the value returned by TxLen().
//
//export TxBytes
func TxBytes(dataOut *C.char) {
	lock.Lock()
	defer lock.Unlock()
	compressed := txCompressor.Bytes()

	outSlice := unsafe.Slice((*byte)(unsafe.Pointer(dataOut)), len(compressed))
	copy(outSlice, compressed)
}

// TxRawCompressedSize compresses the (raw) input statelessly and returns the length of the compressed data.
// The returned length accounts for the "padding" used to fit the data in field elements.
// Input size must be less than 256kB.
// If an error occurred, returns -1.
//
// This function is stateless and does not affect the compressor's internal state.
// It is useful for estimating the compressed size of a transaction for profitability calculations.
//
//export TxRawCompressedSize
func TxRawCompressedSize(input *C.char, inputLength C.int) (compressedSize int) {
	lock.Lock()
	defer lock.Unlock()
	data := unsafe.Slice((*byte)(unsafe.Pointer(input)), inputLength)
	size, err := txCompressor.RawCompressedSize(data)
	if err != nil {
		lastError = err
		return -1
	}
	return size
}

// TxError returns the last encountered error.
// If no error was encountered, returns nil.
//
//export TxError
func TxError() *C.char {
	lock.Lock()
	defer lock.Unlock()
	if lastError != nil {
		// this leaks memory, but since this represents a fatal error, it's probably ok.
		return C.CString(lastError.Error())
	}
	return nil
}
