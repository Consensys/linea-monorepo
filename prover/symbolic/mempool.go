package symbolic

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
)

type chunkT = [MAX_CHUNK_SIZE]field.Element

/*
Reference count for the allocated arrays
*/
var rC sync.Map = sync.Map{}

var (
	chunkPool = sync.Pool{
		New: func() interface{} {
			var res chunkT
			return &res
		},
	}
)

// Clear the pool completely, shields against memory leaks
// Eg: if we forgot to dump a polynomial at some point, this will ensure the value get dumped eventually
// Returns how many polynomials were cleared that way
func ClearPool() int {
	res := 0
	rC.Range(func(k, _ interface{}) bool {
		switch ptr := k.(type) {
		case *chunkT:
			chunkPool.Put(ptr)
		default:
			utils.Panic("tried to clear %T", ptr)
		}
		res++
		return true
	})
	return res
}

// Returns the number of element in the pool
// Does not mutate it
func CountPool() int {
	res := 0
	rC.Range(func(_, _ interface{}) bool {
		res++
		return true
	})
	return res
}

// Tries to find a reusable MultiLin or allocate a new one
func Allocate(n int) []field.Element {
	if n > MAX_CHUNK_SIZE {
		utils.Panic("n %v is larger than the MAX_CHUNK_SIZE %v", n, MAX_CHUNK_SIZE)
	}

	ptr_ := chunkPool.Get().(*chunkT)

	one := int32(1)
	rC.Store(ptr_, &one) // remember we allocated the pointer is being used
	return (*ptr_)[:n]
}

// Decrease the reference count
func DecRef(arr []field.Element) {
	ptr_, err := ptr(arr)
	if err != nil {
		// An error means the ref is not from this pool
		return
	}
	// If the rC did not registers, then
	// either the array was allocated somewhere else and its fine to ignore
	// otherwise a double put and we MUST ignore
	refCnt_, ok := rC.Load(ptr_)
	if !ok {
		logrus.Tracef("not from the pool %p\n", ptr_)
		return
	}

	refCnt := refCnt_.(*int32)
	newRefCount := atomic.AddInt32(refCnt, -1)

	// fmt.Printf("decrement rc for %p - cnt (before) %v\n", ptr_, refCnt)
	if newRefCount <= 0 {
		// Drop the chunk
		rC.Delete(ptr_)
		chunkPool.Put(ptr_)
		return
	}
}

// Increase the reference count of a pooled object
func AddRef(arr []field.Element) {
	ptr_, err := ptr(arr)
	if err != nil {
		// An error means the ref is not from this pool
		return
	}
	// If the rC did not registers, then
	// either the array was allocated somewhere else and its fine to ignore
	// otherwise a double put and we MUST ignore
	refCnt_, ok := rC.Load(ptr_)
	if !ok {
		utils.Panic("not from the pool %p\n", ptr_)
	}

	// Just increase the reference count
	refCnt := refCnt_.(*int32)
	atomic.AddInt32(refCnt, 1)
}

// Get the pointer from the header of the slice
func ptr(s []field.Element) (*chunkT, error) {
	// Re-increase the array up to max capacity
	if cap(s) != MAX_CHUNK_SIZE {
		err := fmt.Errorf("can't cast to large array, the put array's is %v it should have capacity %v", cap(s), MAX_CHUNK_SIZE)
		logrus.Tracef("Error in `ptr` %v", err)
		return nil, err
	}
	return (*chunkT)(unsafe.Pointer(&s[0])), nil
}

// Dump MemStats into stdout
func DumpMemStats() {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	fmt.Printf("MEMSTATS mallocs %v - alloc %v - totalalloc %v\n", stats.Mallocs, stats.Alloc, stats.TotalAlloc)
}
