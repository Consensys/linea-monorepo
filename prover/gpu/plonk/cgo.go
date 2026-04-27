//go:build cuda

// CGO directives and helpers for the plonk package.
package plonk

/*
#cgo LDFLAGS: -L${SRCDIR}/../cuda/build -lgnark_gpu -L/usr/local/cuda/lib64 -lcudart -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/../cuda/include

#include "gnark_gpu.h"
#include <stdlib.h>
*/
import "C"

import (
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// devCtx converts a gpu.Device to the C context handle type.
func devCtx(d *gpu.Device) C.gnark_gpu_context_t {
	return C.gnark_gpu_context_t(d.Handle())
}

// toError converts a C error code to a Go error.
func toError(code C.gnark_gpu_error_t) error {
	switch code {
	case C.GNARK_GPU_SUCCESS:
		return nil
	case C.GNARK_GPU_ERROR_CUDA:
		return &gpu.Error{Code: int(code), Message: "CUDA error"}
	case C.GNARK_GPU_ERROR_INVALID_ARG:
		return &gpu.Error{Code: int(code), Message: "invalid argument"}
	case C.GNARK_GPU_ERROR_OUT_OF_MEMORY:
		return &gpu.Error{Code: int(code), Message: "out of GPU memory"}
	case C.GNARK_GPU_ERROR_SIZE_MISMATCH:
		return &gpu.Error{Code: int(code), Message: "vector size mismatch"}
	default:
		return &gpu.Error{Code: int(code), Message: "unknown error"}
	}
}
