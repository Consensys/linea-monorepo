//go:build cuda

package plonk2

import (
	"fmt"
	"math/bits"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfp "github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	bnfp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfp "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/constraint"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

type genericProverState struct {
	dev    *gpu.Device
	curve  Curve
	n      int
	log2n  uint
	perm   *DeviceInt64
	fft    *FFTDomain
	kzg    *G1MSM
	lagKzg *G1MSM
	fixed  genericFixedPolynomials
	qcp    []*FrVector
	qcpLen int

	qkLagrangeTemplate []uint64
	scratch            *genericProofScratch
}

type genericFixedPolynomials struct {
	ql *FrVector
	qr *FrVector
	qm *FrVector
	qo *FrVector
	qk *FrVector
	s1 *FrVector
	s2 *FrVector
	s3 *FrVector
}

type genericTraceRaw struct {
	n           int
	permutation []int64
	ql          []uint64
	qr          []uint64
	qm          []uint64
	qo          []uint64
	qk          []uint64
	s1          []uint64
	s2          []uint64
	s3          []uint64
	qcp         [][]uint64
}

func newGenericProverState(
	dev *gpu.Device,
	ccs constraint.ConstraintSystem,
	pk gnarkplonk.ProvingKey,
) (*genericProverState, error) {
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if err := dev.Bind(); err != nil {
		return nil, fmt.Errorf("plonk2: bind CUDA device %d: %w", dev.DeviceID(), err)
	}
	curve, err := curveFromConstraintSystem(ccs)
	if err != nil {
		return nil, err
	}
	trace, err := genericTraceFromConstraintSystem(curve, ccs)
	if err != nil {
		return nil, err
	}
	canonicalSRS, lagrangeSRS, err := genericSRSRaw(curve, pk)
	if err != nil {
		return nil, err
	}
	if len(trace.permutation) != 3*trace.n {
		return nil, fmt.Errorf("plonk2: permutation length %d, want %d", len(trace.permutation), 3*trace.n)
	}

	state := &genericProverState{
		dev:                dev,
		curve:              curve,
		n:                  trace.n,
		log2n:              uint(bits.TrailingZeros(uint(trace.n))),
		qcpLen:             len(trace.qcp),
		qkLagrangeTemplate: trace.qk,
	}
	fail := func(label string, err error) (*genericProverState, error) {
		state.Close()
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	state.scratch, err = newGenericProofScratch(curve, trace.n, len(trace.qcp))
	if err != nil {
		return fail("allocate proof scratch", err)
	}

	state.kzg, err = NewG1MSM(dev, curve, canonicalSRS)
	if err != nil {
		return fail("create canonical SRS MSM", err)
	}
	state.lagKzg, err = NewG1MSM(dev, curve, lagrangeSRS)
	if err != nil {
		return fail("create lagrange SRS MSM", err)
	}
	state.fft, err = NewFFTDomain(dev, genericFFTSpec(curve, trace.n))
	if err != nil {
		return fail("create FFT domain", err)
	}
	state.perm, err = NewDeviceInt64(dev, trace.permutation)
	if err != nil {
		return fail("upload permutation", err)
	}
	if err := state.uploadFixedPolynomials(trace); err != nil {
		return fail("upload fixed polynomials", err)
	}
	if err := dev.Sync(); err != nil {
		return fail("sync generic prover preparation", err)
	}
	return state, nil
}

func (s *genericProverState) uploadFixedPolynomials(trace genericTraceRaw) error {
	var err error
	if s.fixed.ql, err = s.uploadCanonical(trace.ql); err != nil {
		return fmt.Errorf("ql: %w", err)
	}
	if s.fixed.qr, err = s.uploadCanonical(trace.qr); err != nil {
		return fmt.Errorf("qr: %w", err)
	}
	if s.fixed.qm, err = s.uploadCanonical(trace.qm); err != nil {
		return fmt.Errorf("qm: %w", err)
	}
	if s.fixed.qo, err = s.uploadCanonical(trace.qo); err != nil {
		return fmt.Errorf("qo: %w", err)
	}
	if s.fixed.qk, err = s.uploadCanonical(trace.qk); err != nil {
		return fmt.Errorf("qk: %w", err)
	}
	if s.fixed.s1, err = s.uploadCanonical(trace.s1); err != nil {
		return fmt.Errorf("s1: %w", err)
	}
	if s.fixed.s2, err = s.uploadCanonical(trace.s2); err != nil {
		return fmt.Errorf("s2: %w", err)
	}
	if s.fixed.s3, err = s.uploadCanonical(trace.s3); err != nil {
		return fmt.Errorf("s3: %w", err)
	}
	s.qcp = make([]*FrVector, len(trace.qcp))
	for i := range trace.qcp {
		s.qcp[i], err = s.uploadCanonical(trace.qcp[i])
		if err != nil {
			return fmt.Errorf("qcp[%d]: %w", i, err)
		}
	}
	return nil
}

func (s *genericProverState) uploadCanonical(lagrangeRaw []uint64) (*FrVector, error) {
	v, err := NewFrVector(s.dev, s.curve, s.n)
	if err != nil {
		return nil, err
	}
	if err := v.CopyFromHostRaw(lagrangeRaw); err != nil {
		v.Free()
		return nil, err
	}
	if err := s.fft.BitReverse(v); err != nil {
		v.Free()
		return nil, err
	}
	if err := s.fft.FFTInverse(v); err != nil {
		v.Free()
		return nil, err
	}
	return v, nil
}

func (s *genericProverState) commitLagrangeRaw(lagrangeRaw []uint64) ([]uint64, error) {
	return s.lagKzg.CommitRaw(lagrangeRaw)
}

func (s *genericProverState) commitLagrangeWaveRaw(batch [][]uint64) ([][]uint64, error) {
	return s.commitWaveRaw(s.lagKzg, batch)
}

func (s *genericProverState) commitCanonicalWaveRaw(batch [][]uint64) ([][]uint64, error) {
	return s.commitWaveRaw(s.kzg, batch)
}

func (s *genericProverState) pinLagrangeCommitmentWave() error {
	return s.pinCommitmentWave(s.lagKzg)
}

func (s *genericProverState) releaseLagrangeCommitmentWave() error {
	return s.releaseCommitmentWave(s.lagKzg)
}

func (s *genericProverState) pinCanonicalCommitmentWave() error {
	return s.pinCommitmentWave(s.kzg)
}

func (s *genericProverState) releaseCanonicalCommitmentWave() error {
	return s.releaseCommitmentWave(s.kzg)
}

func (s *genericProverState) commitWaveRaw(msm *G1MSM, batch [][]uint64) ([][]uint64, error) {
	if s == nil || s.dev == nil || msm == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if err := s.dev.Bind(); err != nil {
		return nil, fmt.Errorf("plonk2: bind CUDA device %d: %w", s.dev.DeviceID(), err)
	}
	return msm.commitRawBatch(batch)
}

func (s *genericProverState) pinCommitmentWave(msm *G1MSM) error {
	if s == nil || s.dev == nil || msm == nil {
		return gpu.ErrDeviceClosed
	}
	if err := s.dev.Bind(); err != nil {
		return fmt.Errorf("plonk2: bind CUDA device %d: %w", s.dev.DeviceID(), err)
	}
	return msm.PinWorkBuffers()
}

func (s *genericProverState) releaseCommitmentWave(msm *G1MSM) error {
	if s == nil || s.dev == nil || msm == nil {
		return gpu.ErrDeviceClosed
	}
	if err := s.dev.Bind(); err != nil {
		return fmt.Errorf("plonk2: bind CUDA device %d: %w", s.dev.DeviceID(), err)
	}
	return msm.ReleaseWorkBuffers()
}

func (s *genericProverState) Close() {
	if s == nil {
		return
	}
	for _, v := range []*FrVector{s.fixed.ql, s.fixed.qr, s.fixed.qm, s.fixed.qo, s.fixed.qk, s.fixed.s1, s.fixed.s2, s.fixed.s3} {
		if v != nil {
			v.Free()
		}
	}
	for _, v := range s.qcp {
		if v != nil {
			v.Free()
		}
	}
	if s.perm != nil {
		s.perm.Free()
	}
	if s.fft != nil {
		s.fft.Free()
	}
	if s.kzg != nil {
		s.kzg.Close()
	}
	if s.lagKzg != nil {
		s.lagKzg.Close()
	}
	*s = genericProverState{}
}

func (s *genericProverState) fixedPolynomialCount() int {
	if s == nil {
		return 0
	}
	return 8 + s.qcpLen
}

func genericTraceFromConstraintSystem(curve Curve, ccs constraint.ConstraintSystem) (genericTraceRaw, error) {
	switch curve {
	case CurveBN254:
		spr := ccs.(*csbn254.SparseR1CS)
		domain := bnfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bnfft.WithoutPrecompute())
		trace := bnplonk.NewTrace(spr, domain)
		qcp := make([][]uint64, len(trace.Qcp))
		for i := range trace.Qcp {
			qcp[i] = genericRawBN254Fr(trace.Qcp[i].Coefficients())
		}
		return genericTraceRaw{
			n:           int(domain.Cardinality),
			permutation: trace.S,
			ql:          genericRawBN254Fr(trace.Ql.Coefficients()),
			qr:          genericRawBN254Fr(trace.Qr.Coefficients()),
			qm:          genericRawBN254Fr(trace.Qm.Coefficients()),
			qo:          genericRawBN254Fr(trace.Qo.Coefficients()),
			qk:          genericRawBN254Fr(trace.Qk.Coefficients()),
			s1:          genericRawBN254Fr(trace.S1.Coefficients()),
			s2:          genericRawBN254Fr(trace.S2.Coefficients()),
			s3:          genericRawBN254Fr(trace.S3.Coefficients()),
			qcp:         qcp,
		}, nil
	case CurveBLS12377:
		spr := ccs.(*csbls12377.SparseR1CS)
		domain := blsfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), blsfft.WithoutPrecompute())
		trace := blsplonk.NewTrace(spr, domain)
		qcp := make([][]uint64, len(trace.Qcp))
		for i := range trace.Qcp {
			qcp[i] = genericRawBLS12377Fr(trace.Qcp[i].Coefficients())
		}
		return genericTraceRaw{
			n:           int(domain.Cardinality),
			permutation: trace.S,
			ql:          genericRawBLS12377Fr(trace.Ql.Coefficients()),
			qr:          genericRawBLS12377Fr(trace.Qr.Coefficients()),
			qm:          genericRawBLS12377Fr(trace.Qm.Coefficients()),
			qo:          genericRawBLS12377Fr(trace.Qo.Coefficients()),
			qk:          genericRawBLS12377Fr(trace.Qk.Coefficients()),
			s1:          genericRawBLS12377Fr(trace.S1.Coefficients()),
			s2:          genericRawBLS12377Fr(trace.S2.Coefficients()),
			s3:          genericRawBLS12377Fr(trace.S3.Coefficients()),
			qcp:         qcp,
		}, nil
	case CurveBW6761:
		spr := ccs.(*csbw6761.SparseR1CS)
		domain := bwfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bwfft.WithoutPrecompute())
		trace := bwplonk.NewTrace(spr, domain)
		qcp := make([][]uint64, len(trace.Qcp))
		for i := range trace.Qcp {
			qcp[i] = genericRawBW6761Fr(trace.Qcp[i].Coefficients())
		}
		return genericTraceRaw{
			n:           int(domain.Cardinality),
			permutation: trace.S,
			ql:          genericRawBW6761Fr(trace.Ql.Coefficients()),
			qr:          genericRawBW6761Fr(trace.Qr.Coefficients()),
			qm:          genericRawBW6761Fr(trace.Qm.Coefficients()),
			qo:          genericRawBW6761Fr(trace.Qo.Coefficients()),
			qk:          genericRawBW6761Fr(trace.Qk.Coefficients()),
			s1:          genericRawBW6761Fr(trace.S1.Coefficients()),
			s2:          genericRawBW6761Fr(trace.S2.Coefficients()),
			s3:          genericRawBW6761Fr(trace.S3.Coefficients()),
			qcp:         qcp,
		}, nil
	default:
		return genericTraceRaw{}, fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
}

func genericSRSRaw(curve Curve, pk gnarkplonk.ProvingKey) (canonical, lagrange []uint64, err error) {
	switch curve {
	case CurveBN254:
		p := pk.(*bnplonk.ProvingKey)
		return genericRawBN254G1(p.Kzg.G1), genericRawBN254G1(p.KzgLagrange.G1), nil
	case CurveBLS12377:
		p := pk.(*blsplonk.ProvingKey)
		return genericRawBLS12377G1(p.Kzg.G1), genericRawBLS12377G1(p.KzgLagrange.G1), nil
	case CurveBW6761:
		p := pk.(*bwplonk.ProvingKey)
		return genericRawBW6761G1(p.Kzg.G1), genericRawBW6761G1(p.KzgLagrange.G1), nil
	default:
		return nil, nil, fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
}

func genericFFTSpec(curve Curve, n int) FFTDomainSpec {
	switch curve {
	case CurveBN254:
		domain := bnfft.NewDomain(uint64(n))
		fwd, inv := genericTwiddlesBN254(n, domain.Generator, domain.GeneratorInv)
		return FFTDomainSpec{
			Curve:           CurveBN254,
			Size:            n,
			ForwardTwiddles: genericRawBN254Fr(fwd),
			InverseTwiddles: genericRawBN254Fr(inv),
			CardinalityInv:  genericRawBN254Fr([]bnfr.Element{domain.CardinalityInv}),
		}
	case CurveBLS12377:
		domain := blsfft.NewDomain(uint64(n))
		fwd, inv := genericTwiddlesBLS12377(n, domain.Generator, domain.GeneratorInv)
		return FFTDomainSpec{
			Curve:           CurveBLS12377,
			Size:            n,
			ForwardTwiddles: genericRawBLS12377Fr(fwd),
			InverseTwiddles: genericRawBLS12377Fr(inv),
			CardinalityInv:  genericRawBLS12377Fr([]blsfr.Element{domain.CardinalityInv}),
		}
	case CurveBW6761:
		domain := bwfft.NewDomain(uint64(n))
		fwd, inv := genericTwiddlesBW6761(n, domain.Generator, domain.GeneratorInv)
		return FFTDomainSpec{
			Curve:           CurveBW6761,
			Size:            n,
			ForwardTwiddles: genericRawBW6761Fr(fwd),
			InverseTwiddles: genericRawBW6761Fr(inv),
			CardinalityInv:  genericRawBW6761Fr([]bwfr.Element{domain.CardinalityInv}),
		}
	default:
		panic(fmt.Sprintf("unsupported curve %s", curve))
	}
}

func genericTwiddlesBN254(n int, generator, generatorInv bnfr.Element) ([]bnfr.Element, []bnfr.Element) {
	fwd := make([]bnfr.Element, n/2)
	inv := make([]bnfr.Element, n/2)
	if len(fwd) > 0 {
		fwd[0].SetOne()
		inv[0].SetOne()
	}
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func genericTwiddlesBLS12377(n int, generator, generatorInv blsfr.Element) ([]blsfr.Element, []blsfr.Element) {
	fwd := make([]blsfr.Element, n/2)
	inv := make([]blsfr.Element, n/2)
	if len(fwd) > 0 {
		fwd[0].SetOne()
		inv[0].SetOne()
	}
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func genericTwiddlesBW6761(n int, generator, generatorInv bwfr.Element) ([]bwfr.Element, []bwfr.Element) {
	fwd := make([]bwfr.Element, n/2)
	inv := make([]bwfr.Element, n/2)
	if len(fwd) > 0 {
		fwd[0].SetOne()
		inv[0].SetOne()
	}
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func genericRawBN254Fr(v []bnfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*bnfr.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func genericRawBLS12377Fr(v []blsfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*blsfr.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func genericRawBW6761Fr(v []bwfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*bwfr.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func genericRawBN254G1(points []bn254.G1Affine) []uint64 {
	if len(points) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&points[0])), len(points)*2*bnfp.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func genericRawBLS12377G1(points []bls12377.G1Affine) []uint64 {
	if len(points) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&points[0])), len(points)*2*blsfp.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func genericRawBW6761G1(points []bw6761.G1Affine) []uint64 {
	if len(points) == 0 {
		return nil
	}
	raw := unsafe.Slice((*uint64)(unsafe.Pointer(&points[0])), len(points)*2*bwfp.Limbs)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}
