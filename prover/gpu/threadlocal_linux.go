//go:build linux

package gpu

import (
	"sync"

	"golang.org/x/sys/unix"
)

// SetCurrentDevice associates the current OS thread with a GPU device.
// The caller is expected to have called runtime.LockOSThread() so the
// goroutine doesn't migrate. Pass nil to clear.
//
// Used to pin each segment-prover goroutine to a specific GPU on multi-GPU
// hosts. GPU dispatch sites should call CurrentDevice() to honour the
// per-thread choice; falls back to GetDevice() when unset.
func SetCurrentDevice(d *Device) {
	tid := unix.Gettid()
	if d == nil {
		threadDeviceMu.Lock()
		delete(threadDevice, tid)
		threadDeviceMu.Unlock()
		return
	}
	threadDeviceMu.Lock()
	threadDevice[tid] = d
	threadDeviceMu.Unlock()
}

// CurrentDevice returns the device pinned to the current OS thread via
// SetCurrentDevice. Falls back to GetDevice() when unset.
func CurrentDevice() *Device {
	tid := unix.Gettid()
	threadDeviceMu.RLock()
	d := threadDevice[tid]
	threadDeviceMu.RUnlock()
	if d != nil {
		return d
	}
	return GetDevice()
}

// CurrentDeviceID returns the index passed to GetDeviceN for the current
// OS thread's device, or 0 when unset / multi-device disabled. Used by the
// GPU phase tracer.
func CurrentDeviceID() int {
	tid := unix.Gettid()
	threadDeviceMu.RLock()
	id := threadDeviceID[tid]
	threadDeviceMu.RUnlock()
	return id
}

// SetCurrentDeviceID is the lower-level setter used together with
// SetCurrentDevice when tracing wants to know the index, not the handle.
func SetCurrentDeviceID(id int) {
	tid := unix.Gettid()
	threadDeviceMu.Lock()
	threadDeviceID[tid] = id
	threadDeviceMu.Unlock()
}

var (
	threadDeviceMu sync.RWMutex
	threadDevice   = map[int]*Device{}
	threadDeviceID = map[int]int{}
)
