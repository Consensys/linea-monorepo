//go:build !linux

package gpu

// SetCurrentDevice is a no-op on platforms without a stable OS-thread ID API.
// Linux CUDA builds use a per-thread implementation for multi-GPU workers.
func SetCurrentDevice(d *Device) {}

// CurrentDevice returns the default device on non-Linux platforms.
func CurrentDevice() *Device {
	return GetDevice()
}

// CurrentDeviceID reports the default device ID on non-Linux platforms.
func CurrentDeviceID() int {
	return 0
}

// SetCurrentDeviceID is a no-op on non-Linux platforms.
func SetCurrentDeviceID(id int) {}
