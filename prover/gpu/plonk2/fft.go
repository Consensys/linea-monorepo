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

// FFTDomainSpec contains host-side NTT constants in raw AoS Montgomery layout.
//
// ForwardTwiddles and InverseTwiddles must contain Size()/2 field elements:
// omega^0, omega^1, ... and omega^-0, omega^-1, ... respectively. The inverse
// cardinality is one field element, 1/n.
type FFTDomainSpec struct {
	Curve           Curve
	Size            int
	ForwardTwiddles []uint64
	InverseTwiddles []uint64
	CardinalityInv  []uint64
}

// FFTDomain owns GPU-resident twiddle factors for one curve and domain size.
type FFTDomain struct {
	handle C.gnark_gpu_plonk2_ntt_domain_t
	dev    *gpu.Device
	curve  Curve
	size   int
}

// NewFFTDomain uploads a curve-specific NTT domain to the GPU.
func NewFFTDomain(dev *gpu.Device, spec FFTDomainSpec) (*FFTDomain, error) {
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	info, err := spec.Curve.validate()
	if err != nil {
		return nil, err
	}
	if !isPowerOfTwo(spec.Size) {
		return nil, fmt.Errorf("plonk2: domain size must be a positive power of two")
	}

	half := spec.Size / 2
	twiddleWords := half * info.ScalarLimbs
	if len(spec.ForwardTwiddles) != twiddleWords {
		return nil, fmt.Errorf(
			"plonk2: forward twiddle word count %d, want %d",
			len(spec.ForwardTwiddles),
			twiddleWords,
		)
	}
	if len(spec.InverseTwiddles) != twiddleWords {
		return nil, fmt.Errorf(
			"plonk2: inverse twiddle word count %d, want %d",
			len(spec.InverseTwiddles),
			twiddleWords,
		)
	}
	if len(spec.CardinalityInv) != info.ScalarLimbs {
		return nil, fmt.Errorf(
			"plonk2: inverse cardinality word count %d, want %d",
			len(spec.CardinalityInv),
			info.ScalarLimbs,
		)
	}

	var fwdPtr, invPtr *C.uint64_t
	if twiddleWords > 0 {
		fwdPtr = (*C.uint64_t)(unsafe.Pointer(&spec.ForwardTwiddles[0]))
		invPtr = (*C.uint64_t)(unsafe.Pointer(&spec.InverseTwiddles[0]))
	}

	var handle C.gnark_gpu_plonk2_ntt_domain_t
	if err := toError(C.gnark_gpu_plonk2_ntt_domain_create(
		devCtx(dev),
		cCurve(spec.Curve),
		C.size_t(spec.Size),
		fwdPtr,
		invPtr,
		(*C.uint64_t)(unsafe.Pointer(&spec.CardinalityInv[0])),
		&handle,
	)); err != nil {
		return nil, err
	}

	d := &FFTDomain{
		handle: handle,
		dev:    dev,
		curve:  spec.Curve,
		size:   spec.Size,
	}
	runtime.SetFinalizer(d, (*FFTDomain).Free)
	return d, nil
}

// Free releases GPU twiddle memory. It is safe to call multiple times.
func (d *FFTDomain) Free() {
	if d != nil && d.handle != nil {
		C.gnark_gpu_plonk2_ntt_domain_destroy(d.handle)
		d.handle = nil
		runtime.SetFinalizer(d, nil)
	}
}

func (d *FFTDomain) checkVector(v *FrVector) error {
	if d == nil || d.handle == nil || d.dev == nil || d.dev.Handle() == nil {
		return gpu.ErrDeviceClosed
	}
	if err := v.checkLive(); err != nil {
		return err
	}
	if d.dev != v.dev {
		return fmt.Errorf("plonk2: domain and vector are on different devices")
	}
	if d.curve != v.curve {
		return fmt.Errorf("plonk2: domain curve %s, vector curve %s", d.curve, v.curve)
	}
	if d.size != v.n {
		return fmt.Errorf("plonk2: domain size %d, vector length %d", d.size, v.n)
	}
	return nil
}

func (d *FFTDomain) checkScalarRaw(generator []uint64) error {
	if d == nil || d.handle == nil {
		return gpu.ErrDeviceClosed
	}
	info, err := d.curve.validate()
	if err != nil {
		return err
	}
	if len(generator) != info.ScalarLimbs {
		return fmt.Errorf(
			"plonk2: generator word count %d, want %d",
			len(generator),
			info.ScalarLimbs,
		)
	}
	return nil
}

// Size returns the domain cardinality.
func (d *FFTDomain) Size() int {
	if d == nil {
		return 0
	}
	return d.size
}

// Curve returns the domain scalar field curve.
func (d *FFTDomain) Curve() Curve {
	if d == nil {
		return 0
	}
	return d.curve
}

// FFT runs an in-place forward DIF NTT. The output is bit-reversed.
func (d *FFTDomain) FFT(v *FrVector) error {
	if err := d.checkVector(v); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_ntt_forward(d.handle, v.handle))
}

// FFTInverse runs an in-place inverse DIT NTT. The input must be bit-reversed.
func (d *FFTDomain) FFTInverse(v *FrVector) error {
	if err := d.checkVector(v); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_ntt_inverse(d.handle, v.handle))
}

// BitReverse applies the in-place bit-reversal permutation.
func (d *FFTDomain) BitReverse(v *FrVector) error {
	if err := d.checkVector(v); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_ntt_bit_reverse(d.handle, v.handle))
}

// CosetFFT evaluates a natural-order coefficient vector on generator*H.
//
// The result is in natural order, matching gpu/plonk's public CosetFFT
// semantics. Internally this is ScaleByPowersRaw(generator), FFT, BitReverse.
func (d *FFTDomain) CosetFFT(v *FrVector, generator []uint64) error {
	if err := d.checkVector(v); err != nil {
		return err
	}
	if err := d.checkScalarRaw(generator); err != nil {
		return err
	}
	if err := v.ScaleByPowersRaw(generator); err != nil {
		return err
	}
	if err := d.FFT(v); err != nil {
		return err
	}
	return d.BitReverse(v)
}

// CosetFFTInverse recovers natural-order coefficients from coset evaluations.
//
// generatorInv must be the inverse of the forward coset generator.
func (d *FFTDomain) CosetFFTInverse(v *FrVector, generatorInv []uint64) error {
	if err := d.checkVector(v); err != nil {
		return err
	}
	if err := d.checkScalarRaw(generatorInv); err != nil {
		return err
	}
	if err := d.BitReverse(v); err != nil {
		return err
	}
	if err := d.FFTInverse(v); err != nil {
		return err
	}
	return v.ScaleByPowersRaw(generatorInv)
}

func (d *FFTDomain) transformBatch(plan NTTPlan, vectors []*FrVector, generator []uint64) error {
	if len(vectors) == 0 {
		return fmt.Errorf("plonk2: NTT batch must not be empty")
	}
	if len(vectors) != plan.BatchCount {
		return fmt.Errorf("plonk2: NTT batch count %d, plan expects %d", len(vectors), plan.BatchCount)
	}
	if d.curve != plan.Curve {
		return fmt.Errorf("plonk2: NTT plan curve %s, domain curve %s", plan.Curve, d.curve)
	}
	if d.size != plan.Size {
		return fmt.Errorf("plonk2: NTT plan size %d, domain size %d", plan.Size, d.size)
	}
	for i, v := range vectors {
		var err error
		switch plan.Direction {
		case nttDirectionForward:
			err = d.FFT(v)
		case nttDirectionInverse:
			err = d.FFTInverse(v)
		case nttDirectionCosetForward:
			err = d.CosetFFT(v, generator)
		case nttDirectionCosetInverse:
			err = d.CosetFFTInverse(v, generator)
		default:
			err = fmt.Errorf("plonk2: unsupported NTT direction %d", plan.Direction)
		}
		if err != nil {
			return fmt.Errorf("plonk2: NTT batch item %d: %w", i, err)
		}
	}
	return nil
}
