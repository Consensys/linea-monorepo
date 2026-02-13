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
// Returns true if the compressor was initialized, false otherwise.
// If false is returned, the TxError() method will return a string describing the error.
//
//export TxInit
func TxInit(dataLimit int, dictPath *C.char) bool {
	fPath := C.GoString(dictPath)
	return txInitGo(dataLimit, fPath)
}

func txInitGo(dataLimit int, dictPath string) bool {
	lock.Lock()
	defer lock.Unlock()
	txCompressor, lastError = blob_v1.NewTxCompressor(dataLimit, dictPath)

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

// TxWrite appends an RLP-encoded transaction to the compressed data.
// The Go code doesn't keep a pointer to the input slice and the caller is free to modify it.
// Returns true if the transaction was appended, false if it would exceed the limit.
//
// The input []byte is interpreted as an RLP encoded Transaction.
//
//export TxWrite
func TxWrite(input *C.char, inputLength C.int) (txAppended bool) {
	lock.Lock()
	defer lock.Unlock()
	rlpTx := unsafe.Slice((*byte)(unsafe.Pointer(input)), inputLength)
	txAppended, err := txCompressor.Write(rlpTx, false)
	if err != nil {
		lastError = err
		return false
	}

	return txAppended
}

// TxCanWrite checks if an RLP-encoded transaction can be appended without actually appending it.
// Returns true if the transaction could be appended, false otherwise.
//
//export TxCanWrite
func TxCanWrite(input *C.char, inputLength C.int) (canAppend bool) {
	lock.Lock()
	defer lock.Unlock()
	rlpTx := unsafe.Slice((*byte)(unsafe.Pointer(input)), inputLength)
	canAppend, err := txCompressor.CanWrite(rlpTx)
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
