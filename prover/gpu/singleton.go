package gpu

import (
	"os"
	"runtime/debug"
	"strconv"
	"sync"
)

var (
	defaultDevOnce sync.Once
	defaultDev     *Device

	devicesMu sync.Mutex
	devices   = map[int]*Device{}

	deviceCountOnce sync.Once
	deviceCount     int
)

// GetDevice returns a lazily-initialized default GPU device (device 0).
// Returns nil when GPU is not available (no CUDA build tag or init failure).
// Thread-safe; the device is created at most once per process.
//
// Side effect: raises GOGC to 1000 (if currently lower) when a device is
// successfully created. CUDA's pinned-memory allocations interact poorly
// with Go's GC at the default GOGC=100, adding seconds of overhead per
// GC cycle on large heaps.
func GetDevice() *Device {
	return GetDeviceN(0)
}

// GetDeviceN returns a lazily-initialized GPU device by ID, creating one if
// needed. Each id is initialized at most once. Returns nil when GPU is not
// available or device init fails.
//
// On multi-GPU hosts, callers route work to a specific device (e.g. by
// segment-index modulo DeviceCount()) and must keep all activity for a
// given segment on a single device — buffers and contexts do not migrate.
//
// The first successful device-init also raises GOGC to 1000; see GetDevice.
func GetDeviceN(id int) *Device {
	if !Enabled || id < 0 {
		return nil
	}

	if id == 0 {
		defaultDevOnce.Do(func() {
			dev, err := New(WithDeviceID(0))
			if err != nil {
				return
			}
			defaultDev = dev
			raiseGOGC()
		})
		return defaultDev
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()
	if d, ok := devices[id]; ok {
		return d
	}
	dev, err := New(WithDeviceID(id))
	if err != nil {
		return nil
	}
	devices[id] = dev
	raiseGOGC()
	return dev
}

// DeviceCount returns the number of GPUs the prover is configured to use,
// read from $LIMITLESS_GPU_COUNT (default 1). Returns 0 when GPU is disabled.
// Read once per process.
func DeviceCount() int {
	if !Enabled {
		return 0
	}
	deviceCountOnce.Do(func() {
		deviceCount = 1
		if v := os.Getenv("LIMITLESS_GPU_COUNT"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				deviceCount = n
			}
		}
	})
	return deviceCount
}

func raiseGOGC() {
	const minGOGC = 1000
	if cur := debug.SetGCPercent(minGOGC); cur > minGOGC {
		debug.SetGCPercent(cur) // restore if already higher
	}
}
