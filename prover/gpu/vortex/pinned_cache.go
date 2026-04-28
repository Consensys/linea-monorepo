// Reusable pinned host-buffer cache.
//
// Why this exists
// ───────────────
// `cudaMallocHost` is one of the most expensive CUDA APIs — at ~64 MiB
// it costs several milliseconds, at hundreds of MiB it can be 10-50+ ms.
// Callers that allocate a fresh pinned buffer on every frame
// (gpu/quotient.RunGPU) burn this cost every prover step.
//
// Profiling observation that motivated this:
//
//   gpu/quotient TIMING @ n=2^20, 16 base roots:
//     pack=19.6 ms  ← 64 MiB cudaMallocHost + parallel host copy
//     ifft=1.1 ms   (GPU)
//     symEval kernel=2.7 ms (GPU)
//
// The actual GPU work is ~5 ms. The 19.6 ms "pack" is dominated by the
// fresh pinned alloc, not the host-side copy. Caching the pinned buffer
// per (device-ID, capacity) eliminates that cost on the second and
// subsequent calls.
//
// Lifecycle
// ─────────
//   - GetPinned(deviceID, n) returns a slice of at least n elements.
//     The buffer is sticky to the (deviceID, capacity) pair.
//   - The returned slice is the cached buffer at its original capacity;
//     callers should slice it down to their needed length.
//   - ReleasePinnedCache(deviceID) frees all cached buffers for that
//     device. Pass deviceID < 0 to clear every cached buffer.
//
// Thread safety
// ─────────────
// Concurrent goroutines on different devices serialise on the cache
// mutex but otherwise run in parallel. A goroutine that calls
// GetPinned with the same (deviceID, capacity) twice will get the SAME
// buffer back — callers that need two simultaneous buffers must use
// distinct sizes or fall back to AllocPinned.

//go:build cuda

package vortex

import (
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear"
)

type pinnedKey struct {
	deviceID int
	capacity int // element count, not bytes
}

var (
	pinnedMu    sync.Mutex
	pinnedCache = map[pinnedKey][]koalabear.Element{}
)

// GetPinned returns a pinned host buffer of at least n koalabear.Element
// (4 bytes each), backed by cudaMallocHost so it can be H2D'd via
// cudaMemcpyAsync without a staging copy. Slice it down to your real
// length before use.
//
// Buffers are cached per (deviceID, capacity). The first call at a
// given (deviceID, capacity) pays a cudaMallocHost; subsequent calls
// reuse the existing buffer.
func GetPinned(deviceID, n int) []koalabear.Element {
	if n <= 0 {
		return nil
	}
	key := pinnedKey{deviceID: deviceID, capacity: n}
	pinnedMu.Lock()
	defer pinnedMu.Unlock()
	if buf, ok := pinnedCache[key]; ok {
		return buf
	}
	// Cache miss: allocate fresh and stash. AllocPinned panics on failure.
	buf := AllocPinned(n)
	pinnedCache[key] = buf
	return buf
}

// ReleasePinnedCache frees all cached pinned buffers for the given
// device. Use at logical boundaries to reclaim host RAM (pinned memory
// is page-locked and counts against the system's pinned-memory budget).
//
// Pass deviceID < 0 to release every cached buffer regardless of device.
func ReleasePinnedCache(deviceID int) {
	pinnedMu.Lock()
	defer pinnedMu.Unlock()
	for key, buf := range pinnedCache {
		if deviceID >= 0 && key.deviceID != deviceID {
			continue
		}
		FreePinned(buf)
		delete(pinnedCache, key)
	}
}
