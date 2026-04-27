// Package gpu provides GPU device management, stream scheduling, and error
// handling for CUDA-accelerated cryptographic operations (PlonK prover, Vortex
// commitment, symbolic expression evaluation).
//
// Build tags:
//
//	cuda     — links against CUDA runtime; full GPU acceleration
//	!cuda    — stub types that panic; compile-only on non-GPU machines
//
// Usage:
//
//	if gpu.Enabled {
//	    dev, _ := gpu.New()
//	    defer dev.Close()
//	    // GPU path ...
//	} else {
//	    // CPU path ...
//	}
//
// Runtime model (cuda build):
//
//	Go code        C API               CUDA stream(s)
//	-------        -----               ----------------
//	New()   -----> gnark_gpu_init  --> context + default stream
//	ops()   -----> enqueue kernels --> async execution
//	Sync()  -----> sync call       --> wait until stream idle
//	Close() -----> destroy         --> free all device resources
//
// Error contract:
//   - C API returns error codes; ToError maps them to stable Go errors.
//   - Compute-path methods panic on programmer misuse (size/device mismatch),
//     matching gnark-crypto style.
package gpu

import (
	"errors"
	"fmt"
)

// ─────────────────────────────────────────────────────────────────────────────
// Errors
// ─────────────────────────────────────────────────────────────────────────────

// ErrDeviceClosed is returned when operating on a closed device.
var ErrDeviceClosed = errors.New("gpu: device closed")

// Error represents a gnark-gpu error.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("gpu: error %d: %s", e.Code, e.Message)
}

// ─────────────────────────────────────────────────────────────────────────────
// CUDA Streams & Events
//
// Streams allow overlapping GPU operations (compute, H2D, D2H) for pipeline
// parallelism. Events provide cross-stream synchronization.
//
//	┌─────────────┐   ┌─────────────┐   ┌─────────────┐
//	│  Compute(0) │   │ Transfer(1) │   │   MSM(2)    │
//	│  FFT, gates │   │  H2D / D2H  │   │  sort+accum │
//	└──────┬──────┘   └──────┬──────┘   └──────┬──────┘
//	       │                 │                  │
//	       ├─── event ──────►│                  │
//	       │                 ├─── event ───────►│
//	       │                 │                  │
//
// ─────────────────────────────────────────────────────────────────────────────

// StreamID identifies a CUDA stream within a Device context.
type StreamID int

const (
	StreamCompute  StreamID = 0 // default compute stream (always available)
	StreamTransfer StreamID = 1 // dedicated H2D/D2H transfer stream
	StreamMSM      StreamID = 2 // dedicated MSM pipeline stream
)

// EventID identifies a CUDA event for cross-stream synchronization.
type EventID int

// ─────────────────────────────────────────────────────────────────────────────
// Device configuration
// ─────────────────────────────────────────────────────────────────────────────

type config struct {
	deviceID int
}

// Option configures a Device.
type Option func(*config)

// WithDeviceID selects which GPU to use (default 0).
func WithDeviceID(id int) Option {
	return func(c *config) {
		c.deviceID = id
	}
}
