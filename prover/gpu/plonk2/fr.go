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

// FrVector is a GPU-resident vector of scalar-field elements for one curve.
//
// Host copies use gnark-crypto's in-memory AoS Montgomery representation:
// [e0.limb0, e0.limb1, ..., e1.limb0, ...]. GPU storage is SoA by limb.
type FrVector struct {
	handle C.gnark_gpu_plonk2_fr_vector_t
	dev    *gpu.Device
	curve  Curve
	limbs  int
	n      int
}

// NewFrVector allocates n scalar-field elements on dev.
func NewFrVector(dev *gpu.Device, curve Curve, n int) (*FrVector, error) {
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, fmt.Errorf("plonk2: vector length must be positive")
	}

	var handle C.gnark_gpu_plonk2_fr_vector_t
	if err := toError(C.gnark_gpu_plonk2_fr_vector_alloc(
		devCtx(dev),
		cCurve(curve),
		C.size_t(n),
		&handle,
	)); err != nil {
		return nil, err
	}

	v := &FrVector{
		handle: handle,
		dev:    dev,
		curve:  curve,
		limbs:  info.ScalarLimbs,
		n:      n,
	}
	runtime.SetFinalizer(v, (*FrVector).Free)
	return v, nil
}

// Free releases the vector's GPU memory. It is safe to call multiple times.
func (v *FrVector) Free() {
	if v != nil && v.handle != nil {
		C.gnark_gpu_plonk2_fr_vector_free(v.handle)
		v.handle = nil
		runtime.SetFinalizer(v, nil)
	}
}

func (v *FrVector) checkLive() error {
	if v == nil || v.handle == nil || v.dev == nil || v.dev.Handle() == nil {
		return gpu.ErrDeviceClosed
	}
	return nil
}

func (v *FrVector) checkScalarRaw(scalar []uint64) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if len(scalar) != v.limbs {
		return fmt.Errorf("plonk2: scalar word count %d, want %d", len(scalar), v.limbs)
	}
	return nil
}

// Len returns the vector length in field elements.
func (v *FrVector) Len() int {
	if v == nil {
		return 0
	}
	return v.n
}

// Curve returns the vector's scalar-field curve.
func (v *FrVector) Curve() Curve {
	if v == nil {
		return 0
	}
	return v.curve
}

// Limbs returns the number of uint64 limbs per scalar element.
func (v *FrVector) Limbs() int {
	if v == nil {
		return 0
	}
	return v.limbs
}

// RawWords returns Len()*Limbs(), the expected size of raw host buffers.
func (v *FrVector) RawWords() int {
	if v == nil {
		return 0
	}
	return v.n * v.limbs
}

// CopyFromHostRaw copies raw AoS Montgomery words to the GPU.
func (v *FrVector) CopyFromHostRaw(src []uint64) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if len(src) != v.RawWords() {
		return fmt.Errorf("plonk2: host word count %d, want %d", len(src), v.RawWords())
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_copy_to_device(
		v.handle,
		(*C.uint64_t)(unsafe.Pointer(&src[0])),
		C.size_t(v.n),
	))
}

// CopyToHostRaw copies GPU data into raw AoS Montgomery words.
func (v *FrVector) CopyToHostRaw(dst []uint64) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if len(dst) != v.RawWords() {
		return fmt.Errorf("plonk2: host word count %d, want %d", len(dst), v.RawWords())
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_copy_to_host(
		v.handle,
		(*C.uint64_t)(unsafe.Pointer(&dst[0])),
		C.size_t(v.n),
	))
}

// CopyFromDevice copies src into v without a host roundtrip.
func (v *FrVector) CopyFromDevice(src *FrVector) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if err := src.checkLive(); err != nil {
		return err
	}
	if v.dev != src.dev {
		return fmt.Errorf("plonk2: vectors are on different devices")
	}
	if v.curve != src.curve {
		return fmt.Errorf("plonk2: vector curve mismatch")
	}
	if v.n != src.n {
		return fmt.Errorf("plonk2: vector length mismatch")
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_copy_d2d(
		devCtx(v.dev),
		v.handle,
		src.handle,
	))
}

// SetZero sets every element to zero.
func (v *FrVector) SetZero() error {
	if err := v.checkLive(); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_set_zero(devCtx(v.dev), v.handle))
}

func (v *FrVector) checkBinaryInputs(a, b *FrVector) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if err := a.checkLive(); err != nil {
		return err
	}
	if err := b.checkLive(); err != nil {
		return err
	}
	if v.dev != a.dev || a.dev != b.dev {
		return fmt.Errorf("plonk2: vectors are on different devices")
	}
	if v.curve != a.curve || a.curve != b.curve {
		return fmt.Errorf("plonk2: vector curve mismatch")
	}
	if v.n != a.n || a.n != b.n {
		return fmt.Errorf("plonk2: vector length mismatch")
	}
	return nil
}

// Add computes v = a + b.
func (v *FrVector) Add(a, b *FrVector) error {
	if err := v.checkBinaryInputs(a, b); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_add(devCtx(v.dev), v.handle, a.handle, b.handle))
}

// Sub computes v = a - b.
func (v *FrVector) Sub(a, b *FrVector) error {
	if err := v.checkBinaryInputs(a, b); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_sub(devCtx(v.dev), v.handle, a.handle, b.handle))
}

// Mul computes v = a * b in Montgomery form.
func (v *FrVector) Mul(a, b *FrVector) error {
	if err := v.checkBinaryInputs(a, b); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_mul(devCtx(v.dev), v.handle, a.handle, b.handle))
}

// AddMul computes v = v + a*b in Montgomery form.
func (v *FrVector) AddMul(a, b *FrVector) error {
	if err := v.checkBinaryInputs(a, b); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_addmul(devCtx(v.dev), v.handle, a.handle, b.handle))
}

// ScalarMulRaw computes v = v * scalar in Montgomery form.
func (v *FrVector) ScalarMulRaw(scalar []uint64) error {
	if err := v.checkScalarRaw(scalar); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_scalar_mul(
		devCtx(v.dev),
		v.handle,
		(*C.uint64_t)(unsafe.Pointer(&scalar[0])),
	))
}

// AddScalarMulRaw computes v = v + a*scalar in Montgomery form.
func (v *FrVector) AddScalarMulRaw(a *FrVector, scalar []uint64) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if err := a.checkLive(); err != nil {
		return err
	}
	if v.dev != a.dev {
		return fmt.Errorf("plonk2: vectors are on different devices")
	}
	if v.curve != a.curve {
		return fmt.Errorf("plonk2: vector curve mismatch")
	}
	if v.n != a.n {
		return fmt.Errorf("plonk2: vector length mismatch")
	}
	if err := v.checkScalarRaw(scalar); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_add_scalar_mul(
		devCtx(v.dev),
		v.handle,
		a.handle,
		(*C.uint64_t)(unsafe.Pointer(&scalar[0])),
	))
}

// BatchInvert computes v[i] = 1/v[i] in place.
//
// temp must be a live vector with the same curve, device, and length. The
// current CUDA implementation uses per-element exponentiation and reserves the
// temp parameter for the optimized Montgomery batch-inversion replacement.
func (v *FrVector) BatchInvert(temp *FrVector) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if err := temp.checkLive(); err != nil {
		return err
	}
	if v.dev != temp.dev {
		return fmt.Errorf("plonk2: vectors are on different devices")
	}
	if v.curve != temp.curve {
		return fmt.Errorf("plonk2: vector curve mismatch")
	}
	if v.n != temp.n {
		return fmt.Errorf("plonk2: vector length mismatch")
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_batch_invert(
		devCtx(v.dev),
		v.handle,
		temp.handle,
	))
}

// ScaleByPowersRaw computes v[i] = v[i] * generator^i in place.
func (v *FrVector) ScaleByPowersRaw(generator []uint64) error {
	if err := v.checkScalarRaw(generator); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_scale_by_powers(
		devCtx(v.dev),
		v.handle,
		(*C.uint64_t)(unsafe.Pointer(&generator[0])),
	))
}
