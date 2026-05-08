//go:build !cuda

// GPU type stubs for non-CUDA builds. Guard calls with gpu.Enabled.
package vortex

import (
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

// ─── KBVector ────────────────────────────────────────────────────────────────

type KBVector struct{}

func NewKBVector(_ *gpu.Device, _ int) (*KBVector, error)     { panic("gpu: cuda required") }
func (v *KBVector) Free()                                     {}
func (v *KBVector) Len() int                                  { return 0 }
func (v *KBVector) CopyFromHost(_ []koalabear.Element)        { panic("gpu: cuda required") }
func (v *KBVector) CopyToHost(_ []koalabear.Element)          { panic("gpu: cuda required") }
func (v *KBVector) Add(_, _ *KBVector)                        { panic("gpu: cuda required") }
func (v *KBVector) Sub(_, _ *KBVector)                        { panic("gpu: cuda required") }
func (v *KBVector) Mul(_, _ *KBVector)                        { panic("gpu: cuda required") }
func (v *KBVector) Scale(_ koalabear.Element)                 { panic("gpu: cuda required") }
func (v *KBVector) ScaleByPowers(_ koalabear.Element)         { panic("gpu: cuda required") }
func (v *KBVector) BitReverse()                               { panic("gpu: cuda required") }
func (v *KBVector) CopyFromDevice(_ *KBVector)                 { panic("gpu: cuda required") }
func (v *KBVector) DevicePtr() unsafe.Pointer                 { panic("gpu: cuda required") }

// ─── GPUFFTDomain ────────────────────────────────────────────────────────────

type GPUFFTDomain struct{}

func NewGPUFFTDomain(_ *gpu.Device, _ int) (*GPUFFTDomain, error) { panic("gpu: cuda required") }
func (f *GPUFFTDomain) Free()                                     {}
func (f *GPUFFTDomain) FFT(_ *KBVector)                           { panic("gpu: cuda required") }
func (f *GPUFFTDomain) FFTInverse(_ *KBVector)                    { panic("gpu: cuda required") }
func (f *GPUFFTDomain) CosetFFT(_ *KBVector, _ koalabear.Element) { panic("gpu: cuda required") }

// ─── GPUPoseidon2 ────────────────────────────────────────────────────────────

type GPUPoseidon2 struct{}

func NewGPUPoseidon2(_ *gpu.Device, _ int) (*GPUPoseidon2, error) { panic("gpu: cuda required") }
func (p *GPUPoseidon2) Free()                                     {}
func (p *GPUPoseidon2) CompressBatch(_ []koalabear.Element, _ int) []Hash {
	panic("gpu: cuda required")
}

// ─── GPUVortex ───────────────────────────────────────────────────────────────

type GPUVortex struct{}

func NewGPUVortex(_ *gpu.Device, _ *Params, _ int) (*GPUVortex, error) { panic("gpu: cuda required") }
func (gv *GPUVortex) Free()                                            {}
func (gv *GPUVortex) Commit(_ [][]koalabear.Element) (*CommitState, Hash, error) {
	panic("gpu: cuda required")
}

// ─── GPU helpers ─────────────────────────────────────────────────────────────

func GPULinCombE4(_ *gpu.Device, _ []*KBVector, _ fext.E4, _ int) []fext.E4 {
	panic("gpu: cuda required")
}
