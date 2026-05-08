package gpu

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
)

const (
	EnvDeviceID = "LINEA_PROVER_GPU_DEVICE_ID"

	// EnvAggregation is the master feature flag that opts the aggregation
	// pipeline (public-input wizard, BW6 aggregation, BN254 emulation) into
	// the GPU prover. Compression always uses the GPU when one is available;
	// aggregation only does so when this is set, since the aggregation GPU
	// path also depends on optional PI Vortex/quotient kernels that we want
	// to keep behind a flag for now.
	EnvAggregation = "LINEA_PROVER_GPU_AGGREGATION"
)

// HasDevice reports whether the binary was built with the cuda tag and a GPU
// device is reachable on this host. It is the canonical check for "should we
// route compute to the GPU?". Cheap (memoised behind sync.Once for device 0).
func HasDevice() bool {
	if !Enabled {
		return false
	}
	return GetDevice() != nil
}

// IsAggregationEnabled reports whether GPU dispatch should be used for the
// aggregation pipeline. Returns true only when both a GPU is available AND
// the operator has opted in via $LINEA_PROVER_GPU_AGGREGATION=1.
func IsAggregationEnabled() bool {
	if !HasDevice() {
		return false
	}
	switch os.Getenv(EnvAggregation) {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	}
	return false
}

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

// ConfiguredDeviceID parses LINEA_PROVER_GPU_DEVICE_ID. When unset, callers
// should use the default device routing.
func ConfiguredDeviceID() (id int, configured bool, err error) {
	raw := os.Getenv(EnvDeviceID)
	if raw == "" {
		return 0, false, nil
	}
	id, err = strconv.Atoi(raw)
	if err != nil {
		return 0, true, fmt.Errorf("invalid %s %q: %w", EnvDeviceID, raw, err)
	}
	if id < 0 {
		return 0, true, fmt.Errorf("%s must be non-negative, got %d", EnvDeviceID, id)
	}
	return id, true, nil
}

// DeviceFromEnvOrCurrent returns the explicitly configured GPU, when
// LINEA_PROVER_GPU_DEVICE_ID is set, or the GPU currently pinned to this OS
// thread. If neither is configured it falls back to device 0.
func DeviceFromEnvOrCurrent() (*Device, int, error) {
	id, configured, err := ConfiguredDeviceID()
	if err != nil {
		return nil, 0, err
	}
	if !configured {
		dev := CurrentDevice()
		if dev == nil {
			return nil, CurrentDeviceID(), nil
		}
		return dev, dev.DeviceID(), nil
	}
	if !Enabled {
		return nil, id, fmt.Errorf("%s=%d requires a binary built with the cuda tag", EnvDeviceID, id)
	}
	dev := GetDeviceN(id)
	if dev == nil {
		return nil, id, fmt.Errorf("GPU device %d is not available", id)
	}
	if err := dev.Bind(); err != nil {
		return nil, id, fmt.Errorf("bind GPU device %d: %w", id, err)
	}
	SetCurrentDevice(dev)
	SetCurrentDeviceID(id)
	return dev, id, nil
}

// PinConfiguredDevice locks the current goroutine to its OS thread and binds
// LINEA_PROVER_GPU_DEVICE_ID for process-level GPU work. The returned cleanup
// function must be called by the same goroutine.
func PinConfiguredDevice() (id int, configured bool, cleanup func(), err error) {
	cleanup = func() {}
	id, configured, err = ConfiguredDeviceID()
	if err != nil || !configured {
		return id, configured, cleanup, err
	}
	if !Enabled {
		return id, configured, cleanup,
			fmt.Errorf("%s=%d requires a binary built with the cuda tag", EnvDeviceID, id)
	}
	dev := GetDeviceN(id)
	if dev == nil {
		return id, configured, cleanup, fmt.Errorf("GPU device %d is not available", id)
	}
	runtime.LockOSThread()
	if err := dev.Bind(); err != nil {
		runtime.UnlockOSThread()
		return id, configured, cleanup, fmt.Errorf("bind GPU device %d: %w", id, err)
	}
	SetCurrentDevice(dev)
	SetCurrentDeviceID(id)
	return id, configured, func() {
		SetCurrentDevice(nil)
		SetCurrentDeviceID(0)
		runtime.UnlockOSThread()
	}, nil
}

func raiseGOGC() {
	const minGOGC = 1000
	if cur := debug.SetGCPercent(minGOGC); cur > minGOGC {
		debug.SetGCPercent(cur) // restore if already higher
	}
}
