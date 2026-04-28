//go:build cuda

package plonk2

/*
#include "gnark_gpu.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

var (
	errEmptyInt64DeviceSlice = errors.New("plonk2: int64 device slice must not be empty")
	errDeviceMismatch        = errors.New("plonk2: device mismatch")
	errPermutationTooSmall   = errors.New("plonk2: permutation device slice is too small")
)

// DeviceInt64 owns a device-resident int64 array.
type DeviceInt64 struct {
	dev *gpu.Device
	ptr unsafe.Pointer
	n   int
}

// NewDeviceInt64 uploads host int64 data to the selected GPU.
func NewDeviceInt64(dev *gpu.Device, values []int64) (*DeviceInt64, error) {
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if len(values) == 0 {
		return nil, errEmptyInt64DeviceSlice
	}
	var ptr unsafe.Pointer
	if err := toError(C.gnark_gpu_device_alloc_copy_int64(
		devCtx(dev),
		(*C.int64_t)(unsafe.Pointer(&values[0])),
		C.size_t(len(values)),
		&ptr,
	)); err != nil {
		return nil, err
	}
	out := &DeviceInt64{dev: dev, ptr: ptr, n: len(values)}
	runtime.SetFinalizer(out, (*DeviceInt64).Free)
	return out, nil
}

// Free releases the device memory. It is safe to call multiple times.
func (p *DeviceInt64) Free() {
	if p != nil && p.ptr != nil {
		C.gnark_gpu_device_free_ptr(p.ptr)
		p.ptr = nil
		runtime.SetFinalizer(p, nil)
	}
}

func (p *DeviceInt64) checkLive(dev *gpu.Device, minLen int) error {
	if p == nil || p.ptr == nil || p.dev == nil || p.dev.Handle() == nil {
		return gpu.ErrDeviceClosed
	}
	if p.dev != dev {
		return errDeviceMismatch
	}
	if p.n < minLen {
		return errPermutationTooSmall
	}
	return nil
}
