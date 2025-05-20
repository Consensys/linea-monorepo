package main

/*
#cgo LDFLAGS: -L/opt/homebrew/Cellar/openjdk@21/21.0.6/libexec/openjdk.jdk/Contents/Home/lib/server -ljvm
#cgo CFLAGS: -I/opt/homebrew/opt/openjdk@21/include
#include <jni.h>
#include <stdlib.h>

inline jint GetArrayLength(JNIEnv *env, jbyteArray array) {
    return (*env)->GetArrayLength(env, array);
}

inline jbyte* GetByteArrayElements(JNIEnv *env, jbyteArray array, jboolean *isCopy) {
    return (*env)->GetByteArrayElements(env, array, isCopy);
}

inline jclass FindClass(JNIEnv *env, const char *name) {
    return (*env)->FindClass(env, name);
}

inline void ThrowNew(JNIEnv *env, jclass clazz, const char *message) {
    (*env)->ThrowNew(env, clazz, message);
}

inline jbyteArray NewByteArray(JNIEnv *env, jsize length) {
    return (*env)->NewByteArray(env, length);
}

inline void SetByteArrayRegion(JNIEnv *env, jbyteArray array, jsize start, jsize len, const jbyte *buf) {
    (*env)->SetByteArrayRegion(env, array, start, len, buf);
}

inline void ReleaseByteArrayElements(JNIEnv *env, jbyteArray array, jbyte *elements, jint mode) {
    (*env)->ReleaseByteArrayElements(env, array, elements, mode);
}
*/
import "C"

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"slices"
	"sync"
	"unsafe"

	blobv1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
)

//go:generate go build -tags nocorset -ldflags "-s -w" -buildmode=c-shared -o libcompressor.so libcompressor.go
func main() {}

type instance struct {
	sync.Mutex
	*blobv1.BlobMaker
}

const (
	errorJniClass            = "java/lang/Error"
	runtimeExceptionJniClass = "java/lang/RuntimeException"

	// TODO Make custom exceptions
	compressorErrorJniClass           = errorJniClass
	instanceNotFoundExceptionJniClass = runtimeExceptionJniClass
)

var (
	instances map[int]*instance
	mapLock   sync.Mutex
)

func throw(env *C.JNIEnv, errorClass string, err error) {
	errorClassC := C.CString(errorClass)
	defer C.free(unsafe.Pointer(errorClassC))

	errorMessageC := C.CString(err.Error())
	defer C.free(unsafe.Pointer(errorMessageC))

	// Use the JNIEnv pointer to call FindClass
	exceptionClass := C.jclass(C.FindClass(env, errorClassC))
	if unsafe.Pointer(exceptionClass) == nil {
		runtimeExceptionClass := C.CString("java/lang/RuntimeException")
		defer C.free(unsafe.Pointer(runtimeExceptionClass))

		C.ThrowNew(env, C.FindClass(env, runtimeExceptionClass), errorMessageC)
		return
	}

	C.ThrowNew(env, exceptionClass, errorMessageC)
}

func fromJniBytes(env *C.JNIEnv, bytes C.jbyteArray) []byte {
	// Get the length of the byte array
	length := C.GetArrayLength(env, bytes)

	// Get a pointer to the elements of the Java byte array
	ptr := C.GetByteArrayElements(env, bytes, nil)
	if ptr == nil {
		// If the byte array is null, throw an exception
		throw(env, compressorErrorJniClass, fmt.Errorf("byte array is null"))
		return nil
	}
	defer C.ReleaseByteArrayElements(env, bytes, ptr, 0) // Ensure the array is released
	// Convert the pointer to a Go byte slice
	return slices.Clone(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length))
}

func toJniBytes(env *C.JNIEnv, bytes []byte) C.jbyteArray {
	// Create a new Java byte array in the JVM
	jbyteArray := C.NewByteArray(env, C.jsize(len(bytes)))
	if unsafe.Pointer(jbyteArray) == nil {
		// If the array creation fails, throw an exception
		throw(env, runtimeExceptionJniClass, fmt.Errorf("failed to create Java byte array"))
		return jbyteArray
	}

	// Copy the compressed data into the Java byte array
	C.SetByteArrayRegion(env, jbyteArray, 0, C.jsize(len(bytes)), (*C.jbyte)(unsafe.Pointer(&bytes[0])))

	// Return the Java byte array
	return jbyteArray
}

// getInstance finds the instance with the given ID.
// If not found, it will throw a RuntimeException.
// It will lock the instance to prevent concurrent access.
func getInstance(env *C.JNIEnv, instanceID C.int) *instance {
	mapLock.Lock()
	defer mapLock.Unlock()

	if inst, ok := instances[int(instanceID)]; ok {
		inst.Lock()
		return inst
	}
	throw(env, instanceNotFoundExceptionJniClass, fmt.Errorf("instance %d not found", int(instanceID)))
	return nil
}

// NewInstance initializes the compressor.
// The dataLimit argument is the maximum size of the compressed data.
// Returns the instance ID if successful, or -1 if an error occurred.
//
//export NewInstance
func NewInstance(env *C.JNIEnv, dataLimit C.int, dictPath *C.char) C.int {

	var bytes [4]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		throw(env, runtimeExceptionJniClass, err)
		return -1
	}

	compressor, err := blobv1.NewBlobMaker(int(dataLimit), C.GoString(dictPath))
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return -1
	}

	mapLock.Lock()
	defer mapLock.Unlock()

	id := int(binary.BigEndian.Uint32(bytes[:]))
	if _, ok := instances[id]; ok {
		throw(env, compressorErrorJniClass, fmt.Errorf("instance %d already exists", id))
		return -1
	}

	instances[id] = &instance{
		BlobMaker: compressor,
	}

	return C.int(id)
}

// Reset resets the compressor. Must be called between blobs.
//
//export Reset
func Reset(env *C.JNIEnv, instanceID C.int) {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	compressor.Reset()
}

// Write appends the input to the compressed data.
// The Go code doesn't keep a pointer to the input slice and the caller is free to modify it.
// Returns true if the chunk was appended, false if the chunk was discarded.
//
// The input []byte is interpreted as a RLP encoded Block.
//
//export Write
func Write(env *C.JNIEnv, instanceID C.int, rlpBlock C.jbyteArray) (chunkAppended bool) {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	chunkAppended, err := compressor.Write(fromJniBytes(env, rlpBlock), false)
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return false
	}

	return chunkAppended
}

// CanWrite behaves as Write, except that it doesn't append the input to the compressed data
// (but return true if it could)
//
//export CanWrite
func CanWrite(env *C.JNIEnv, instanceID C.int, rlpBlock C.jbyteArray) (chunkAppended bool) {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	chunkAppended, err := compressor.Write(fromJniBytes(env, rlpBlock), true)
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return false
	}

	return chunkAppended
}

// StartNewBatch starts a new batch; must be called between each batch in the blob.
//
//export StartNewBatch
func StartNewBatch(env *C.JNIEnv, instanceID C.int) {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	compressor.StartNewBatch()
}

// Len returns the length of the compressed data.
//
//export Len
func Len(env *C.JNIEnv, instanceID C.int) (length C.int) {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	return C.int(compressor.Len())
}

// Bytes returns the compressed data.
//
//export Bytes
func Bytes(env *C.JNIEnv, instanceID C.int) C.jbyteArray {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	return toJniBytes(env, compressor.Bytes())
}

// WorstCompressedBlockSize returns the size of the given block, as compressed by an "empty" blob maker.
// That is, with more context, blob maker could compress the block further, but this function
// returns the maximum size that can be achieved.
//
// The input is a RLP encoded block.
//
//export WorstCompressedBlockSize
func WorstCompressedBlockSize(env *C.JNIEnv, instanceID C.int, rlpBlock C.jbyteArray) C.int {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	_, n, err := compressor.WorstCompressedBlockSize(fromJniBytes(env, rlpBlock))
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return -1
	}
	return C.int(n)
}

// WorstCompressedTxSize returns the size of the given transaction, as compressed by an "empty" blob maker.
// That is, with more context, blob maker could compress the transaction further, but this function
// returns the maximum size that can be achieved.
//
// The input is a RLP encoded transaction.
//
//export WorstCompressedTxSize
func WorstCompressedTxSize(env *C.JNIEnv, instanceID C.int, rlpTx C.jbyteArray) C.int {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	n, err := compressor.WorstCompressedTxSize(fromJniBytes(env, rlpTx))
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return -1
	}
	return C.int(n)
}

// RawCompressedSize compresses the (raw) input and returns the length of the compressed data.
// The returned length account for the "padding" used by the blob maker to
// fit the data in field elements.
// Input size must be less than 256kB.
//
//export RawCompressedSize
func RawCompressedSize(env *C.JNIEnv, instanceID C.int, input C.jbyteArray) C.int {
	compressor := getInstance(env, instanceID)
	defer compressor.Unlock()

	n, err := compressor.RawCompressedSize(fromJniBytes(env, input))
	if err != nil {
		throw(env, compressorErrorJniClass, err)
		return -1
	}
	return C.int(n)
}
