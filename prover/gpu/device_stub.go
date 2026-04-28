//go:build !cuda

package gpu

import "unsafe"

// Device is a stub for non-CUDA builds. All methods panic.
// Guard with gpu.Enabled before calling.
type Device struct{}

func New(opts ...Option) (*Device, error) { panic("gpu: requires cuda build tag") }
func (d *Device) Close() error           { panic("gpu: requires cuda build tag") }
func (d *Device) Sync() error            { panic("gpu: requires cuda build tag") }
func (d *Device) Closed() bool           { return true }
func (d *Device) Handle() unsafe.Pointer { panic("gpu: requires cuda build tag") }
func (d *Device) DeviceID() int          { return 0 }
func (d *Device) Bind() error            { panic("gpu: requires cuda build tag") }

func (d *Device) MemGetInfo() (free, total uint64, err error) {
	panic("gpu: requires cuda build tag")
}

func (d *Device) CreateStream(id StreamID) error       { panic("gpu: requires cuda build tag") }
func (d *Device) RecordEvent(stream StreamID, event EventID) { panic("gpu: requires cuda build tag") }
func (d *Device) WaitEvent(stream StreamID, event EventID)   { panic("gpu: requires cuda build tag") }
func (d *Device) SyncStream(stream StreamID) error      { panic("gpu: requires cuda build tag") }
func (d *Device) InitMultiStream() error                { panic("gpu: requires cuda build tag") }
