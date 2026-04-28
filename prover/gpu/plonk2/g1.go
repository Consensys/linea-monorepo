//go:build cuda

package plonk2

/*
#include "gnark_gpu.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func g1AffineAddRaw(dev *gpu.Device, curve Curve, p, q []uint64) ([]uint64, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	pointWords := 2 * info.BaseFieldLimbs
	if len(p) != pointWords || len(q) != pointWords {
		return nil, fmt.Errorf("plonk2: G1 affine input words must be %d", pointWords)
	}

	out := make([]uint64, 3*info.BaseFieldLimbs)
	err = toError(C.gnark_gpu_plonk2_test_g1_affine_add(
		devCtx(dev),
		cCurve(curve),
		(*C.uint64_t)(unsafe.Pointer(&p[0])),
		(*C.uint64_t)(unsafe.Pointer(&q[0])),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func g1AffineDoubleRaw(dev *gpu.Device, curve Curve, p []uint64) ([]uint64, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	pointWords := 2 * info.BaseFieldLimbs
	if len(p) != pointWords {
		return nil, fmt.Errorf("plonk2: G1 affine input words must be %d", pointWords)
	}

	out := make([]uint64, 3*info.BaseFieldLimbs)
	err = toError(C.gnark_gpu_plonk2_test_g1_affine_double(
		devCtx(dev),
		cCurve(curve),
		(*C.uint64_t)(unsafe.Pointer(&p[0])),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func g1MSMNaiveRaw(dev *gpu.Device, curve Curve, points, scalars []uint64, count int) ([]uint64, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if count <= 0 {
		return nil, fmt.Errorf("plonk2: MSM count must be positive")
	}
	pointWords := count * 2 * info.BaseFieldLimbs
	scalarWords := count * info.ScalarLimbs
	if len(points) != pointWords {
		return nil, fmt.Errorf("plonk2: MSM point words %d, want %d", len(points), pointWords)
	}
	if len(scalars) != scalarWords {
		return nil, fmt.Errorf("plonk2: MSM scalar words %d, want %d", len(scalars), scalarWords)
	}

	out := make([]uint64, 3*info.BaseFieldLimbs)
	err = toError(C.gnark_gpu_plonk2_test_msm_naive(
		devCtx(dev),
		cCurve(curve),
		(*C.uint64_t)(unsafe.Pointer(&points[0])),
		(*C.uint64_t)(unsafe.Pointer(&scalars[0])),
		C.size_t(count),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func g1MSMPippengerRaw(
	dev *gpu.Device,
	curve Curve,
	points []uint64,
	scalars []uint64,
	count int,
	windowBits int,
) ([]uint64, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if count <= 0 {
		return nil, fmt.Errorf("plonk2: MSM count must be positive")
	}
	if windowBits <= 1 || windowBits > 24 {
		return nil, fmt.Errorf("plonk2: MSM window bits must be in [2,24]")
	}
	pointWords := count * 2 * info.BaseFieldLimbs
	scalarWords := count * info.ScalarLimbs
	if len(points) != pointWords {
		return nil, fmt.Errorf("plonk2: MSM point words %d, want %d", len(points), pointWords)
	}
	if len(scalars) != scalarWords {
		return nil, fmt.Errorf("plonk2: MSM scalar words %d, want %d", len(scalars), scalarWords)
	}

	out := make([]uint64, 3*info.BaseFieldLimbs)
	err = toError(C.gnark_gpu_plonk2_msm_pippenger(
		devCtx(dev),
		cCurve(curve),
		(*C.uint64_t)(unsafe.Pointer(&points[0])),
		(*C.uint64_t)(unsafe.Pointer(&scalars[0])),
		C.size_t(count),
		C.int(windowBits),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	))
	if err != nil {
		return nil, err
	}
	return out, nil
}
