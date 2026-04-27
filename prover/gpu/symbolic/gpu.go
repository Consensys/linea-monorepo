// GPU evaluation of compiled symbolic programs via CUDA.
//
// Build constraint: requires CGO + CUDA (same as gpu.go).

//go:build cuda

package symbolic

/*
#cgo LDFLAGS: -L${SRCDIR}/../cuda/build -lgnark_gpu -L/usr/local/cuda/lib64 -lcudart -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/../cuda/include

#include "gnark_gpu_kb.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/vortex"

	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// devCtx casts the common gpu.Device handle back to the C type for CGO calls.
func devCtx(d *gpu.Device) C.gnark_gpu_context_t {
	return C.gnark_gpu_context_t(d.Handle())
}

func kbError(code C.kb_error_t) error {
	switch code {
	case C.KB_SUCCESS:
		return nil
	case C.KB_ERROR_CUDA:
		return fmt.Errorf("symbolic: CUDA error")
	case C.KB_ERROR_INVALID:
		return fmt.Errorf("symbolic: invalid argument")
	case C.KB_ERROR_OOM:
		return fmt.Errorf("symbolic: out of GPU memory")
	case C.KB_ERROR_SIZE:
		return fmt.Errorf("symbolic: size mismatch")
	default:
		return fmt.Errorf("symbolic: unknown error %d", int(code))
	}
}

func must(code C.kb_error_t) {
	if err := kbError(code); err != nil {
		panic(err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GPUSymProgram — compiled program handle (device-resident bytecode + consts)
// ─────────────────────────────────────────────────────────────────────────────

type GPUSymProgram struct {
	dev    *gpu.Device
	handle C.kb_sym_program_t
}

// CompileSymGPU uploads a compiled GPUProgram to the GPU.
func CompileSymGPU(dev *gpu.Device, pgm *GPUProgram) (*GPUSymProgram, error) {
	var h C.kb_sym_program_t

	var bcPtr, cPtr *C.uint32_t
	if len(pgm.Bytecode) > 0 {
		bcPtr = (*C.uint32_t)(unsafe.Pointer(&pgm.Bytecode[0]))
	}
	if len(pgm.Constants) > 0 {
		cPtr = (*C.uint32_t)(unsafe.Pointer(&pgm.Constants[0]))
	}

	if err := kbError(C.kb_sym_compile(
		devCtx(dev),
		bcPtr, C.uint32_t(len(pgm.Bytecode)),
		cPtr, C.uint32_t(len(pgm.Constants)/4),
		C.uint32_t(pgm.NumSlots),
		C.uint32_t(pgm.ResultSlot),
		&h,
	)); err != nil {
		return nil, err
	}

	return &GPUSymProgram{dev: dev, handle: h}, nil
}

func (p *GPUSymProgram) Free() {
	if p.handle != nil {
		C.kb_sym_free(p.handle)
		p.handle = nil
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SymInput — describes how the GPU reads one variable
// ─────────────────────────────────────────────────────────────────────────────

const (
	SymInputKB       = 0 // base field vector, embed into E4 as (val, 0, 0, 0)
	SymInputConstE4  = 1 // broadcast E4 constant to all threads
	SymInputRotKB    = 2 // rotated base field vector: d_ptr[(i+offset)%n]
	SymInputE4Vec    = 3 // E4 AoS vector: d_ptr[i*4..i*4+3]
	SymInputE4VecSOA = 4 // E4 SoA vector: d_ptr[c*n+i], c in [0..3]
	SymInputRotE4SOA = 5 // rotated E4 SoA: d_ptr[c*n+((i+off)%n)]
	SymInputRotE4AOS = 6 // rotated E4 AoS: d_ptr[((i+off)%n)*4..+3]
)

type SymInput struct {
	Tag    int            // SymInputKB, SymInputConstE4, SymInputRotKB, SymInputE4Vec
	DPtr   unsafe.Pointer // device pointer (nil for ConstE4)
	Offset int            // rotation offset
	Val    [4]uint32      // E4 constant value (Montgomery form)
}

// SymInputFromVec creates a base-field input from a device-resident KBVector.
func SymInputFromVec(v *vortex.KBVector) SymInput {
	return SymInput{
		Tag:  SymInputKB,
		DPtr: v.DevicePtr(),
	}
}

// SymInputFromRotatedVec creates a rotated base-field input.
func SymInputFromRotatedVec(v *vortex.KBVector, offset int) SymInput {
	return SymInput{
		Tag:    SymInputRotKB,
		DPtr:   v.DevicePtr(),
		Offset: offset,
	}
}

// SymInputFromE4Vec creates an E4 vector input from a device buffer of 4n uint32.
// The buffer layout is [b0.a0, b0.a1, b1.a0, b1.a1] × n elements.
func SymInputFromE4Vec(v *vortex.KBVector) SymInput {
	return SymInput{
		Tag:  SymInputE4Vec,
		DPtr: v.DevicePtr(),
	}
}

// SymInputFromE4SOA creates an E4 vector input in SoA layout.
// Layout per root is 4 contiguous vectors of size n:
// [b0.a0(0..n), b0.a1(0..n), b1.a0(0..n), b1.a1(0..n)].
func SymInputFromE4SOA(ptr unsafe.Pointer) SymInput {
	return SymInput{
		Tag:  SymInputE4VecSOA,
		DPtr: ptr,
	}
}

// SymInputFromRotE4SOA creates a rotated E4 vector input in SoA layout.
func SymInputFromRotE4SOA(ptr unsafe.Pointer, offset int) SymInput {
	return SymInput{
		Tag:    SymInputRotE4SOA,
		DPtr:   ptr,
		Offset: offset,
	}
}

// SymInputFromConst creates a constant E4 input (broadcast).
func SymInputFromConst(val fext.E4) SymInput {
	return SymInput{
		Tag: SymInputConstE4,
		Val: [4]uint32{
			uint32(val.B0.A0[0]), uint32(val.B0.A1[0]),
			uint32(val.B1.A0[0]), uint32(val.B1.A1[0]),
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// EvalSymGPU — evaluate compiled program over n elements → host E4 slice
// ─────────────────────────────────────────────────────────────────────────────

func EvalSymGPU(dev *gpu.Device, pgm *GPUSymProgram, inputs []SymInput, n int) []fext.E4 {
	// Build C input descriptors
	descs := make([]C.SymInputDesc, len(inputs))
	for i, inp := range inputs {
		descs[i].tag = C.uint32_t(inp.Tag)
		descs[i].offset = C.uint32_t(inp.Offset)
		descs[i].d_ptr = (*C.uint32_t)(inp.DPtr)
		descs[i].val[0] = C.uint32_t(inp.Val[0])
		descs[i].val[1] = C.uint32_t(inp.Val[1])
		descs[i].val[2] = C.uint32_t(inp.Val[2])
		descs[i].val[3] = C.uint32_t(inp.Val[3])
	}

	result := make([]fext.E4, n)

	var descPtr *C.SymInputDesc
	if len(descs) > 0 {
		descPtr = &descs[0]
	}

	must(C.kb_sym_eval(
		devCtx(dev),
		pgm.handle,
		descPtr, C.uint32_t(len(inputs)),
		C.uint32_t(n),
		(*C.uint32_t)(unsafe.Pointer(&result[0])),
	))

	return result
}
