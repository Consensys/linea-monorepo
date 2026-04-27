//go:build cuda

package gpu

/*
#cgo LDFLAGS: -L${SRCDIR}/cuda/build -lgnark_gpu -L/usr/local/cuda/lib64 -lcudart -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/cuda/include

#include "gnark_gpu.h"
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// toError converts a C error code to a Go error.
func toError(code C.gnark_gpu_error_t) error {
	switch code {
	case C.GNARK_GPU_SUCCESS:
		return nil
	case C.GNARK_GPU_ERROR_CUDA:
		return &Error{Code: int(code), Message: "CUDA error"}
	case C.GNARK_GPU_ERROR_INVALID_ARG:
		return &Error{Code: int(code), Message: "invalid argument"}
	case C.GNARK_GPU_ERROR_OUT_OF_MEMORY:
		return &Error{Code: int(code), Message: "out of GPU memory"}
	case C.GNARK_GPU_ERROR_SIZE_MISMATCH:
		return &Error{Code: int(code), Message: "vector size mismatch"}
	default:
		return &Error{Code: int(code), Message: "unknown error"}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Device
// ─────────────────────────────────────────────────────────────────────────────

// Device manages GPU resources and operations.
// Create with New() and release with Close().
type Device struct {
	handle          C.gnark_gpu_context_t
	multiStreamInit bool
}

// New creates a Device on the specified GPU.
// The device must be closed when no longer needed.
func New(opts ...Option) (*Device, error) {
	cfg := config{deviceID: 0}
	for _, o := range opts {
		o(&cfg)
	}

	var handle C.gnark_gpu_context_t
	if err := toError(C.gnark_gpu_init(C.int(cfg.deviceID), &handle)); err != nil {
		return nil, err
	}

	d := &Device{handle: handle}
	runtime.SetFinalizer(d, (*Device).Close)
	return d, nil
}

// Close releases GPU resources associated with this device.
// It is safe to call Close multiple times.
func (d *Device) Close() error {
	if d.handle != nil {
		C.gnark_gpu_destroy(d.handle)
		d.handle = nil
		runtime.SetFinalizer(d, nil)
	}
	return nil
}

// Sync waits for all queued GPU operations to complete and returns
// any deferred error.
func (d *Device) Sync() error {
	if d.handle == nil {
		return ErrDeviceClosed
	}
	return toError(C.gnark_gpu_sync(d.handle))
}

// MemGetInfo returns free and total GPU memory in bytes.
func (d *Device) MemGetInfo() (free, total uint64, err error) {
	if d.handle == nil {
		return 0, 0, ErrDeviceClosed
	}
	var f, t C.size_t
	if err := toError(C.gnark_gpu_mem_get_info(d.handle, &f, &t)); err != nil {
		return 0, 0, err
	}
	return uint64(f), uint64(t), nil
}

// Closed reports whether the device has been closed.
func (d *Device) Closed() bool {
	return d.handle == nil
}

// Handle returns the opaque CUDA context handle as an unsafe.Pointer.
// Subpackages (plonk, vortex, symbolic) cast this back to their own
// C.gnark_gpu_context_t via their CGO import.
func (d *Device) Handle() unsafe.Pointer {
	return unsafe.Pointer(d.handle)
}

// CreateStream creates a CUDA stream at the given ID.
// Stream 0 (StreamCompute) is created automatically with the device.
func (d *Device) CreateStream(id StreamID) error {
	if d.handle == nil {
		return ErrDeviceClosed
	}
	return toError(C.gnark_gpu_create_stream(d.handle, C.int(id)))
}

// RecordEvent records an event on the specified stream.
// Another stream can later wait for this event via WaitEvent.
func (d *Device) RecordEvent(stream StreamID, event EventID) {
	if d.handle == nil {
		panic("gpu: RecordEvent on closed device")
	}
	if err := toError(C.gnark_gpu_record_event(d.handle, C.int(stream), C.int(event))); err != nil {
		panic("gpu: RecordEvent failed: " + err.Error())
	}
}

// WaitEvent makes the specified stream wait until the given event is recorded.
func (d *Device) WaitEvent(stream StreamID, event EventID) {
	if d.handle == nil {
		panic("gpu: WaitEvent on closed device")
	}
	if err := toError(C.gnark_gpu_wait_event(d.handle, C.int(stream), C.int(event))); err != nil {
		panic("gpu: WaitEvent failed: " + err.Error())
	}
}

// SyncStream waits for all operations on the specified stream to complete.
func (d *Device) SyncStream(stream StreamID) error {
	if d.handle == nil {
		return ErrDeviceClosed
	}
	return toError(C.gnark_gpu_sync_stream(d.handle, C.int(stream)))
}

// InitMultiStream creates the transfer and MSM streams.
// Call once before using multi-stream operations.
func (d *Device) InitMultiStream() error {
	if d.multiStreamInit {
		return nil
	}
	if err := d.CreateStream(StreamTransfer); err != nil {
		return err
	}
	if err := d.CreateStream(StreamMSM); err != nil {
		return err
	}
	d.multiStreamInit = true
	return nil
}
