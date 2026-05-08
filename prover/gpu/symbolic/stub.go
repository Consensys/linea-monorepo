//go:build !cuda

// Stub types for non-CUDA builds. Guard calls with gpu.Enabled.
package symbolic

import (
	"unsafe"

	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/vortex"
)

type GPUSymProgram struct{}

func CompileSymGPU(_ *gpu.Device, _ *GPUProgram) (*GPUSymProgram, error) {
	panic("gpu: cuda required")
}
func (p *GPUSymProgram) Free() {}

// SymInput input descriptor tags.
const (
	SymInputKB      = 0
	SymInputConstE4 = 1
	SymInputRotKB   = 2
	SymInputE4Vec   = 3
)

type SymInput struct {
	Tag    int
	DPtr   unsafe.Pointer
	Offset int
	Val    [4]uint32
}

func SymInputFromVec(_ *vortex.KBVector) SymInput                { panic("gpu: cuda required") }
func SymInputFromRotatedVec(_ *vortex.KBVector, _ int) SymInput  { panic("gpu: cuda required") }
func SymInputFromE4Vec(_ *vortex.KBVector) SymInput              { panic("gpu: cuda required") }
func SymInputFromConst(_ fext.E4) SymInput                       { panic("gpu: cuda required") }

func EvalSymGPU(_ *gpu.Device, _ *GPUSymProgram, _ []SymInput, _ int) []fext.E4 {
	panic("gpu: cuda required")
}
