package limitless

import (
	"runtime"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

// pinGPU locks the calling goroutine to its current OS thread and registers
// a GPU device for it, derived from the segment slot index modulo the GPU
// count. This is the only mechanism by which the GPU dispatch sites
// (gpu/quotient, gpu/vortex) discover which device to talk to on multi-GPU
// hosts. When GPU is disabled or only one device is configured the pin is
// effectively a no-op (device 0 in both branches).
//
// CUDA's "current device" is per-OS-thread state. Bind() issues
// cudaSetDevice on this thread so subsequent allocations + kernel launches
// land on the chosen device. Without Bind() everything silently falls
// through to device 0 — the multi-GPU bug we fixed in 2026-04-27 was
// exactly this: NewGPUVortex(dev1, ...) returned a "device-1" pipeline whose
// memory was actually allocated on device 0.
//
// Each segment goroutine must call unpinGPU on exit to release the OS
// thread back to the scheduler and clear the per-thread map entry.
func pinGPU(segmentIdx int) {
	if !gpu.Enabled {
		return
	}
	n := gpu.DeviceCount()
	if n <= 0 {
		n = 1
	}
	id := segmentIdx % n
	dev := gpu.GetDeviceN(id)
	if dev == nil {
		return
	}
	runtime.LockOSThread()
	if err := dev.Bind(); err != nil {
		// If Bind fails, the goroutine may still hit GPU 0; leave the
		// thread locked but don't surface SetCurrentDevice so callers
		// don't think they're on the requested GPU.
		return
	}
	gpu.SetCurrentDevice(dev)
	gpu.SetCurrentDeviceID(id)
}

func unpinGPU() {
	if !gpu.Enabled {
		return
	}
	gpu.SetCurrentDevice(nil)
	gpu.SetCurrentDeviceID(0)
	runtime.UnlockOSThread()
}
