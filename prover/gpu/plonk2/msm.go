//go:build cuda

package plonk2

/*
#include "gnark_gpu.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

// G1MSM keeps a short-Weierstrass affine SRS resident on the GPU.
type G1MSM struct {
	dev        *gpu.Device
	curve      Curve
	info       CurveInfo
	handle     C.gnark_gpu_plonk2_msm_t
	points     int
	windowBits int
	runPlan    MSMRunPlan
}

// NewG1MSM uploads affine short-Weierstrass points and returns a reusable MSM.
func NewG1MSM(dev *gpu.Device, curve Curve, points []uint64) (*G1MSM, error) {
	return newG1MSMWithWindowBits(dev, curve, points, 0)
}

func newG1MSMWithWindowBits(dev *gpu.Device, curve Curve, points []uint64, windowBits int) (*G1MSM, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	_, count, err := validateRawAffinePoints(curve, points)
	if err != nil {
		return nil, err
	}
	if windowBits == 0 {
		runPlan, err := defaultMSMRunPlan(MSMRunPlanConfig{
			Curve:  curve,
			Points: count,
		})
		if err != nil {
			return nil, err
		}
		windowBits = runPlan.WindowBits
	}
	runPlan, err := defaultMSMRunPlan(MSMRunPlanConfig{
		Curve:      curve,
		Points:     count,
		WindowBits: windowBits,
		SharedBase: true,
	})
	if err != nil {
		return nil, err
	}
	if windowBits <= 1 || windowBits > 24 {
		return nil, fmt.Errorf("plonk2: window bits must be in [2,24]")
	}

	var handle C.gnark_gpu_plonk2_msm_t
	err = toError(C.gnark_gpu_plonk2_msm_create(
		devCtx(dev),
		cCurve(curve),
		(*C.uint64_t)(unsafe.Pointer(&points[0])),
		C.size_t(count),
		C.int(windowBits),
		&handle,
	))
	if err != nil {
		return nil, err
	}

	msm := &G1MSM{
		dev:        dev,
		curve:      curve,
		info:       info,
		handle:     handle,
		points:     count,
		windowBits: windowBits,
		runPlan:    runPlan,
	}
	runtime.SetFinalizer(msm, (*G1MSM).Close)
	return msm, nil
}

// Close releases GPU resources held by m.
func (m *G1MSM) Close() {
	if m == nil || m.handle == nil {
		return
	}
	C.gnark_gpu_plonk2_msm_destroy(m.handle)
	m.handle = nil
	runtime.SetFinalizer(m, nil)
}

// PinWorkBuffers keeps MSM scratch buffers resident across commitments.
//
// The resident SRS points always stay on the GPU while the handle is live.
// Work buffers cover scalar staging, CUB sort storage, bucket accumulators,
// and per-window results. Reusing them avoids per-commit allocation churn, but
// large SRS handles can reserve substantial VRAM.
func (m *G1MSM) PinWorkBuffers() error {
	if m == nil || m.handle == nil || m.dev == nil || m.dev.Handle() == nil {
		return gpu.ErrDeviceClosed
	}
	return toError(C.gnark_gpu_plonk2_msm_pin_work_buffers(m.handle))
}

// ReleaseWorkBuffers releases reusable MSM scratch buffers.
//
// CommitRaw reallocates them lazily if needed. This mirrors gpu/plonk's memory
// lifecycle and lets prover phases reclaim bucket/sort memory while quotient
// kernels are running.
func (m *G1MSM) ReleaseWorkBuffers() error {
	if m == nil || m.handle == nil || m.dev == nil || m.dev.Handle() == nil {
		return gpu.ErrDeviceClosed
	}
	return toError(C.gnark_gpu_plonk2_msm_release_work_buffers(m.handle))
}

// CommitRaw computes a KZG-style commitment using the resident base points.
func (m *G1MSM) CommitRaw(scalars []uint64) ([]uint64, error) {
	if m == nil || m.handle == nil || m.dev == nil || m.dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if len(scalars) == 0 {
		return nil, fmt.Errorf("plonk2: scalar buffer must not be empty")
	}
	count, err := validateRawScalarsAtMost(m.curve, scalars, m.points)
	if err != nil {
		return nil, err
	}

	raw := make([]uint64, 3*m.info.BaseFieldLimbs)
	err := toError(C.gnark_gpu_plonk2_msm_run(
		m.handle,
		(*C.uint64_t)(unsafe.Pointer(&scalars[0])),
		C.size_t(count),
		(*C.uint64_t)(unsafe.Pointer(&raw[0])),
	))
	if err != nil {
		return nil, err
	}
	return correctRawMontgomeryMSM(m.curve, raw)
}

// commitRawBatch computes a private shared-base commitment wave.
//
// The implementation intentionally preserves the existing single-commit CUDA
// path. Callers that care about amortizing setup costs should pin work buffers
// before calling and release them after the whole wave.
func (m *G1MSM) commitRawBatch(scalarBatch [][]uint64) ([][]uint64, error) {
	if len(scalarBatch) == 0 {
		return nil, fmt.Errorf("plonk2: scalar batch must not be empty")
	}
	out := make([][]uint64, len(scalarBatch))
	for i := range scalarBatch {
		if len(scalarBatch[i]) == 0 {
			return nil, fmt.Errorf("plonk2: scalar batch item %d must not be empty", i)
		}
		commitment, err := m.CommitRaw(scalarBatch[i])
		if err != nil {
			return nil, fmt.Errorf("plonk2: scalar batch item %d: %w", i, err)
		}
		out[i] = commitment
	}
	return out, nil
}
