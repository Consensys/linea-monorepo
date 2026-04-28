//go:build !cuda

package plonk2

import "github.com/consensys/linea-monorepo/prover/gpu"

type FrVector struct{}
type DeviceInt64 struct{}
type G1MSM struct{}

func NewFrVector(_ *gpu.Device, _ Curve, _ int) (*FrVector, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

func NewDeviceInt64(_ *gpu.Device, _ []int64) (*DeviceInt64, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

func (v *FrVector) Free()                            {}
func (p *DeviceInt64) Free()                         {}
func (v *FrVector) Len() int                         { return 0 }
func (v *FrVector) Curve() Curve                     { return 0 }
func (v *FrVector) Limbs() int                       { return 0 }
func (v *FrVector) RawWords() int                    { return 0 }
func (v *FrVector) CopyFromHostRaw(_ []uint64) error { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) CopyToHostRaw(_ []uint64) error   { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) CopyFromDevice(_ *FrVector) error { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) SetZero() error                   { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) Add(_, _ *FrVector) error         { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) Sub(_, _ *FrVector) error         { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) Mul(_, _ *FrVector) error         { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) AddMul(_, _ *FrVector) error      { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) ScalarMulRaw(_ []uint64) error    { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) AddScalarMulRaw(_ *FrVector, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}
func (v *FrVector) BatchInvert(_ *FrVector) error { panic("gpu/plonk2: requires cuda build tag") }
func (v *FrVector) ScaleByPowersRaw(_ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func Butterfly4Inverse(_, _, _, _ *FrVector, _, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func ReduceBlindedCoset(_, _ *FrVector, _, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func ComputeL1Den(_ *FrVector, _ *FFTDomain, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func PlonkGateAccum(_, _, _, _, _, _, _, _, _ *FrVector, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func PlonkPermBoundary(
	_, _, _, _, _, _, _, _, _ *FrVector,
	_ *FFTDomain,
	_, _, _, _, _, _, _ []uint64,
) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func PlonkZComputeFactors(
	_, _, _ *FrVector,
	_ *DeviceInt64,
	_ *FFTDomain,
	_, _, _, _ []uint64,
	_ uint,
) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func ZPrefixProduct(_, _, _ *FrVector) error {
	panic("gpu/plonk2: requires cuda build tag")
}

func CommitRaw(_ *gpu.Device, _ Curve, _, _ []uint64) ([]uint64, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

func NewG1MSM(_ *gpu.Device, _ Curve, _ []uint64) (*G1MSM, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

func (m *G1MSM) Close() {}

func (m *G1MSM) PinWorkBuffers() error {
	panic("gpu/plonk2: requires cuda build tag")
}

func (m *G1MSM) ReleaseWorkBuffers() error {
	panic("gpu/plonk2: requires cuda build tag")
}

func (m *G1MSM) CommitRaw(_ []uint64) ([]uint64, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

type FFTDomainSpec struct {
	Curve           Curve
	Size            int
	ForwardTwiddles []uint64
	InverseTwiddles []uint64
	CardinalityInv  []uint64
}

type FFTDomain struct{}

func NewFFTDomain(_ *gpu.Device, _ FFTDomainSpec) (*FFTDomain, error) {
	panic("gpu/plonk2: requires cuda build tag")
}

func (d *FFTDomain) Free()                 {}
func (d *FFTDomain) Size() int             { return 0 }
func (d *FFTDomain) Curve() Curve          { return 0 }
func (d *FFTDomain) FFT(_ *FrVector) error { panic("gpu/plonk2: requires cuda build tag") }
func (d *FFTDomain) FFTInverse(_ *FrVector) error {
	panic("gpu/plonk2: requires cuda build tag")
}
func (d *FFTDomain) BitReverse(_ *FrVector) error {
	panic("gpu/plonk2: requires cuda build tag")
}
func (d *FFTDomain) CosetFFT(_ *FrVector, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}
func (d *FFTDomain) CosetFFTInverse(_ *FrVector, _ []uint64) error {
	panic("gpu/plonk2: requires cuda build tag")
}
