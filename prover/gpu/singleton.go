package gpu

import (
	"runtime/debug"
	"sync"
)

var (
	defaultDevOnce sync.Once
	defaultDev     *Device
)

// GetDevice returns a lazily-initialized default GPU device.
// Returns nil when GPU is not available (no CUDA build tag or init failure).
// Thread-safe; the device is created at most once per process.
//
// Side effect: raises GOGC to 1000 (if currently lower) when a device is
// successfully created. CUDA's pinned-memory allocations interact poorly
// with Go's GC at the default GOGC=100, adding seconds of overhead per
// GC cycle on large heaps.
func GetDevice() *Device {
	if !Enabled {
		return nil
	}
	defaultDevOnce.Do(func() {
		dev, err := New()
		if err != nil {
			return
		}
		defaultDev = dev
		// Raise GOGC to reduce GC frequency. CUDA pinned memory
		// causes GC pauses to be much longer than normal.
		const minGOGC = 1000
		if cur := debug.SetGCPercent(minGOGC); cur > minGOGC {
			debug.SetGCPercent(cur) // restore if already higher
		}
	})
	return defaultDev
}
