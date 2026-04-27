//go:build !cuda

// Stub types for non-CUDA builds. All constructors and methods panic.
// Guard calls with gpu.Enabled.
package plonk

import (
	"context"
	"io"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bls12377kzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	kzgif "github.com/consensys/gnark-crypto/kzg"
	plonk377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bls12-377"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

// ─── FrVector ────────────────────────────────────────────────────────────────

type FrVector struct{}

func NewFrVector(_ *gpu.Device, _ int) (*FrVector, error)         { panic("gpu: cuda required") }
func (v *FrVector) Free()                                         {}
func (v *FrVector) Len() int                                      { return 0 }
func (v *FrVector) CopyFromHost(_ fr.Vector, _ ...gpu.StreamID)   { panic("gpu: cuda required") }
func (v *FrVector) CopyToHost(_ fr.Vector, _ ...gpu.StreamID)     { panic("gpu: cuda required") }
func (v *FrVector) CopyFromDevice(_ *FrVector, _ ...gpu.StreamID) { panic("gpu: cuda required") }
func (v *FrVector) SetZero(_ ...gpu.StreamID)                     { panic("gpu: cuda required") }
func (v *FrVector) Mul(_, _ *FrVector, _ ...gpu.StreamID)         { panic("gpu: cuda required") }
func (v *FrVector) Add(_, _ *FrVector, _ ...gpu.StreamID)         { panic("gpu: cuda required") }
func (v *FrVector) Sub(_, _ *FrVector, _ ...gpu.StreamID)         { panic("gpu: cuda required") }
func (v *FrVector) AddMul(_, _ *FrVector, _ ...gpu.StreamID)      { panic("gpu: cuda required") }
func (v *FrVector) AddScalarMul(_ *FrVector, _ fr.Element, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}
func (v *FrVector) ScalarMul(_ fr.Element, _ ...gpu.StreamID)     { panic("gpu: cuda required") }
func (v *FrVector) ScaleByPowers(_ fr.Element, _ ...gpu.StreamID) { panic("gpu: cuda required") }
func (v *FrVector) BatchInvert(_ *FrVector, _ ...gpu.StreamID)    { panic("gpu: cuda required") }
func (v *FrVector) PatchElements(_ int, _ []fr.Element)           { panic("gpu: cuda required") }

// ─── GPUFFTDomain ────────────────────────────────────────────────────────────

type GPUFFTDomain struct{}

func NewFFTDomain(_ *gpu.Device, _ int) (*GPUFFTDomain, error) { panic("gpu: cuda required") }
func (f *GPUFFTDomain) Size() int                              { return 0 }
func (f *GPUFFTDomain) Close()                                 {}
func (f *GPUFFTDomain) FFT(_ *FrVector, _ ...gpu.StreamID)     { panic("gpu: cuda required") }
func (f *GPUFFTDomain) FFTInverse(_ *FrVector, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}
func (f *GPUFFTDomain) BitReverse(_ *FrVector, _ ...gpu.StreamID) { panic("gpu: cuda required") }
func (f *GPUFFTDomain) CosetFFT(_ *FrVector, _ fr.Element, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}
func (f *GPUFFTDomain) CosetFFTInverse(_ *FrVector, _ fr.Element, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}

func Butterfly4Inverse(_, _, _, _ *FrVector, _, _ fr.Element) { panic("gpu: cuda required") }

// ─── G1TEPoint ───────────────────────────────────────────────────────────────

type G1TEPoint [12]uint64

func ConvertG1AffineToTE(_ []bls12377.G1Affine) []G1TEPoint     { panic("gpu: cuda required") }
func WriteG1TEPoints(_ io.Writer, _ []G1TEPoint) error          { panic("gpu: cuda required") }
func ReadG1TEPoints(_ io.Reader) ([]G1TEPoint, error)           { panic("gpu: cuda required") }
func WriteG1TEPointsRaw(_ io.Writer, _ []G1TEPoint) error       { panic("gpu: cuda required") }
func ReadG1TEPointsRaw(_ io.Reader, _ int) ([]G1TEPoint, error) { panic("gpu: cuda required") }

// ─── G1MSM ───────────────────────────────────────────────────────────────────

type G1MSMPoints struct{ N int }

func ConvertG1Points(_ []bls12377.G1Affine) (*G1MSMPoints, error)   { panic("gpu: cuda required") }
func PinG1TEPoints(_ []G1TEPoint) (*G1MSMPoints, error)             { panic("gpu: cuda required") }
func ReadG1TEPointsPinned(_ io.Reader, _ int) (*G1MSMPoints, error) { panic("gpu: cuda required") }
func (p *G1MSMPoints) Free()                                        {}
func (p *G1MSMPoints) Len() int                                     { return 0 }

type G1MSM struct{}

func NewG1MSM(_ *gpu.Device, _ *G1MSMPoints) (*G1MSM, error)         { panic("gpu: cuda required") }
func NewG1MSMN(_ *gpu.Device, _ *G1MSMPoints, _ int) (*G1MSM, error) { panic("gpu: cuda required") }
func (m *G1MSM) MultiExp(_ ...[]fr.Element) []bls12377.G1Jac         { panic("gpu: cuda required") }
func (m *G1MSM) OffloadPoints()                                      { panic("gpu: cuda required") }
func (m *G1MSM) ReloadPoints()                                       { panic("gpu: cuda required") }
func (m *G1MSM) Close()                                              {}

// ─── Kernels ─────────────────────────────────────────────────────────────────

func ZPrefixProduct(_ *gpu.Device, _, _, _ *FrVector)                 { panic("gpu: cuda required") }
func PolyEvalGPU(_ *gpu.Device, _ *FrVector, _ fr.Element) fr.Element { panic("gpu: cuda required") }
func PolyEvalFromHost(_ *gpu.Device, _ fr.Vector, _ fr.Element) fr.Element {
	panic("gpu: cuda required")
}
func ReduceBlindedCoset(_, _ *FrVector, _ []fr.Element, _ fr.Element) { panic("gpu: cuda required") }
func PlonkZComputeFactors(_, _, _ *FrVector, _ unsafe.Pointer, _, _, _, _ fr.Element, _ uint, _ *GPUFFTDomain) {
	panic("gpu: cuda required")
}
func ComputeL1Den(_ *FrVector, _ fr.Element, _ *GPUFFTDomain, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}
func PlonkGateAccum(_, _, _, _, _, _, _, _, _ *FrVector, _ fr.Element) { panic("gpu: cuda required") }
func PlonkPermBoundary(_, _, _, _, _, _, _, _, _ *FrVector, _, _, _, _, _, _, _ fr.Element, _ *GPUFFTDomain, _ ...gpu.StreamID) {
	panic("gpu: cuda required")
}
func DeviceAllocCopyInt64(_ *gpu.Device, _ []int64) (unsafe.Pointer, error) {
	panic("gpu: cuda required")
}
func DeviceFreePtr(_ unsafe.Pointer) {}

// ─── GPUProvingKey ───────────────────────────────────────────────────────────

type GPUProvingKey struct {
	Vk  *plonk377.VerifyingKey
	Kzg []G1TEPoint
}

func NewGPUProvingKey(_ []G1TEPoint, _ *plonk377.VerifyingKey) *GPUProvingKey {
	panic("gpu: cuda required")
}
func NewGPUProvingKeyFromPinned(_ *G1MSMPoints, _ *plonk377.VerifyingKey) *GPUProvingKey {
	panic("gpu: cuda required")
}
func (gpk *GPUProvingKey) Size() int                                     { return 0 }
func (gpk *GPUProvingKey) Prepare(_ *gpu.Device, _ *cs.SparseR1CS) error { panic("gpu: cuda required") }
func (gpk *GPUProvingKey) Close()                                        {}

func GPUProve(_ *gpu.Device, _ *GPUProvingKey, _ *cs.SparseR1CS, _ witness.Witness) (*plonk377.Proof, error) {
	panic("gpu: cuda required")
}
func GPUPreSolve(_ *gpu.Device, _ *GPUProvingKey, _ *cs.SparseR1CS, _ witness.Witness) (witness.Witness, error) {
	panic("gpu: cuda required")
}

// ─── SRSStore ────────────────────────────────────────────────────────────────

type SRSStore struct{}

func NewSRSStore(_ string) (*SRSStore, error) { panic("gpu: cuda required") }
func (s *SRSStore) GetSRS(_ context.Context, _ constraint.ConstraintSystem) (kzgif.SRS, kzgif.SRS, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) GetSRSCPU(_ context.Context, _ constraint.ConstraintSystem) (kzgif.SRS, kzgif.SRS, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) GetSRSGPU(_ context.Context, _ constraint.ConstraintSystem) ([]G1TEPoint, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) GetSRSGPUPinned(_ context.Context, _ constraint.ConstraintSystem) (*G1MSMPoints, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) LoadTEPoints(_ int, _ bool) ([]G1TEPoint, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) LoadTEPointsPinned(_ int, _ bool) (*G1MSMPoints, error) {
	panic("gpu: cuda required")
}
func (s *SRSStore) LoadPointsAffine(_ int, _ bool) ([]bls12377.G1Affine, error) {
	panic("gpu: cuda required")
}

// Suppress unused import warnings.
var _ = (*bls12377kzg.Digest)(nil)
