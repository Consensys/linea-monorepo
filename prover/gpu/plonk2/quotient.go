//go:build cuda

package plonk2

/*
#include "gnark_gpu.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// Butterfly4Inverse applies the inverse size-4 DFT used by the decomposed
// quotient iFFT in the PlonK prover.
func Butterfly4Inverse(b0, b1, b2, b3 *FrVector, omega4Inv, quarter []uint64) error {
	if err := checkSameShape4(b0, b1, b2, b3); err != nil {
		return err
	}
	if err := b0.checkScalarRaw(omega4Inv); err != nil {
		return err
	}
	if err := b0.checkScalarRaw(quarter); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_fr_vector_butterfly4_inverse(
		devCtx(b0.dev),
		b0.handle,
		b1.handle,
		b2.handle,
		b3.handle,
		(*C.uint64_t)(unsafe.Pointer(&omega4Inv[0])),
		(*C.uint64_t)(unsafe.Pointer(&quarter[0])),
	))
}

// ReduceBlindedCoset folds the blinding tail of a canonical polynomial into
// its first n coefficients before coset evaluation.
//
// Given src = p[0:n] and tail = p[n:], it computes
// dst[i] = src[i] + tail[i]*cosetPowN for i < len(tail), and dst[i] = src[i]
// otherwise. tail is raw AoS Montgomery data.
func ReduceBlindedCoset(dst, src *FrVector, tail, cosetPowN []uint64) error {
	if err := dst.checkLive(); err != nil {
		return err
	}
	if err := src.checkLive(); err != nil {
		return err
	}
	if dst.dev != src.dev {
		return fmt.Errorf("plonk2: vectors are on different devices")
	}
	if dst.curve != src.curve {
		return fmt.Errorf("plonk2: vector curve mismatch")
	}
	if dst.n != src.n {
		return fmt.Errorf("plonk2: vector length mismatch")
	}
	if len(tail)%dst.limbs != 0 {
		return fmt.Errorf("plonk2: tail word count %d is not a multiple of %d", len(tail), dst.limbs)
	}
	tailLen := len(tail) / dst.limbs
	if tailLen > dst.n {
		return fmt.Errorf("plonk2: tail length %d exceeds vector length %d", tailLen, dst.n)
	}
	if tailLen == 0 {
		return dst.CopyFromDevice(src)
	}
	if err := dst.checkScalarRaw(cosetPowN); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_reduce_blinded_coset(
		devCtx(dst.dev),
		dst.handle,
		src.handle,
		(*C.uint64_t)(unsafe.Pointer(&tail[0])),
		C.size_t(tailLen),
		(*C.uint64_t)(unsafe.Pointer(&cosetPowN[0])),
	))
}

// ComputeL1Den computes cosetGen*omega^i - 1 over the domain evaluation
// order. The output is suitable for BatchInvert before boundary accumulation.
func ComputeL1Den(out *FrVector, domain *FFTDomain, cosetGen []uint64) error {
	if err := domain.checkVector(out); err != nil {
		return err
	}
	if err := domain.checkScalarRaw(cosetGen); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_compute_l1_den(
		domain.handle,
		out.handle,
		(*C.uint64_t)(unsafe.Pointer(&cosetGen[0])),
	))
}

// PlonkGateAccum fuses the gate part of the PlonK quotient numerator:
//
//	result = (result + ql*l + qr*r + qm*l*r + qo*o + qk) * zhKInv
//
// result is expected to already contain the permutation and boundary
// contributions for the same coset.
func PlonkGateAccum(
	result, ql, qr, qm, qo, qk, l, r, o *FrVector,
	zhKInv []uint64,
) error {
	if err := checkSameShape(result, ql, qr, qm, qo, qk, l, r, o); err != nil {
		return err
	}
	if err := result.checkScalarRaw(zhKInv); err != nil {
		return err
	}
	return toError(C.gnark_gpu_plonk2_gate_accum(
		devCtx(result.dev),
		result.handle,
		ql.handle,
		qr.handle,
		qm.handle,
		qo.handle,
		qk.handle,
		l.handle,
		r.handle,
		o.handle,
		(*C.uint64_t)(unsafe.Pointer(&zhKInv[0])),
	))
}

// PlonkPermBoundary computes the permutation and first-Lagrange boundary part
// of the PlonK quotient numerator:
//
//	result = alpha * ((Z_shift*prod_sigma - Z*prod_id) +
//		alpha*(Z-1)*L1)
//
// where L1 is provided as l1Scalar*L1DenInv on the coset evaluation domain.
func PlonkPermBoundary(
	result, l, r, o, z, s1, s2, s3, l1DenInv *FrVector,
	domain *FFTDomain,
	alpha, beta, gamma, l1Scalar, cosetShift, cosetShiftSq, cosetGen []uint64,
) error {
	if err := checkSameShape(result, l, r, o, z, s1, s2, s3, l1DenInv); err != nil {
		return err
	}
	if err := domain.checkVector(result); err != nil {
		return err
	}
	scalars := [][]uint64{
		alpha,
		beta,
		gamma,
		l1Scalar,
		cosetShift,
		cosetShiftSq,
		cosetGen,
	}
	params := make([]uint64, 0, len(scalars)*result.limbs)
	for _, scalar := range scalars {
		if err := result.checkScalarRaw(scalar); err != nil {
			return err
		}
		params = append(params, scalar...)
	}
	return toError(C.gnark_gpu_plonk2_perm_boundary(
		devCtx(result.dev),
		result.handle,
		l.handle,
		r.handle,
		o.handle,
		z.handle,
		s1.handle,
		s2.handle,
		s3.handle,
		l1DenInv.handle,
		(*C.uint64_t)(unsafe.Pointer(&params[0])),
		domain.handle,
	))
}

// PlonkZComputeFactors computes the numerator and denominator factors for the
// permutation product polynomial Z. On success, l contains numerators and r
// contains denominators; o is read-only.
func PlonkZComputeFactors(
	l, r, o *FrVector,
	perm *DeviceInt64,
	domain *FFTDomain,
	beta, gamma, cosetShift, cosetShiftSq []uint64,
	log2n uint,
) error {
	if err := checkSameShape(l, r, o); err != nil {
		return err
	}
	if err := domain.checkVector(l); err != nil {
		return err
	}
	if err := perm.checkLive(l.dev, 3*l.n); err != nil {
		return err
	}
	scalars := [][]uint64{beta, gamma, cosetShift, cosetShiftSq}
	params := make([]uint64, 0, len(scalars)*l.limbs)
	for _, scalar := range scalars {
		if err := l.checkScalarRaw(scalar); err != nil {
			return err
		}
		params = append(params, scalar...)
	}
	return toError(C.gnark_gpu_plonk2_z_compute_factors(
		devCtx(l.dev),
		l.handle,
		r.handle,
		o.handle,
		perm.ptr,
		(*C.uint64_t)(unsafe.Pointer(&params[0])),
		C.uint(log2n),
		domain.handle,
	))
}

// ZPrefixProduct computes z[i] = product(ratio[0:i]) with an exclusive shift:
// z[0] = 1 and z[i] = ratio[0]*...*ratio[i-1].
func ZPrefixProduct(z, ratio, temp *FrVector) error {
	if err := checkSameShape(z, ratio, temp); err != nil {
		return err
	}
	numChunksMax := (ratio.n + 1023) / 1024
	chunkProducts := make([]uint64, numChunksMax*ratio.limbs)
	var numChunks C.size_t
	if err := toError(C.gnark_gpu_plonk2_z_prefix_phase1(
		devCtx(z.dev),
		z.handle,
		ratio.handle,
		(*C.uint64_t)(unsafe.Pointer(&chunkProducts[0])),
		&numChunks,
	)); err != nil {
		return err
	}

	nc := int(numChunks)
	scannedPrefixes := make([]uint64, nc*ratio.limbs)
	switch ratio.curve {
	case CurveBN254:
		scanChunkProductsBN254(chunkProducts, scannedPrefixes, nc)
	case CurveBLS12377:
		scanChunkProductsBLS12377(chunkProducts, scannedPrefixes, nc)
	case CurveBW6761:
		scanChunkProductsBW6761(chunkProducts, scannedPrefixes, nc)
	default:
		return fmt.Errorf("plonk2: unsupported curve %s", ratio.curve)
	}

	return toError(C.gnark_gpu_plonk2_z_prefix_phase3(
		devCtx(z.dev),
		z.handle,
		temp.handle,
		(*C.uint64_t)(unsafe.Pointer(&scannedPrefixes[0])),
		C.size_t(nc),
	))
}

func scanChunkProductsBN254(chunkProducts, scannedPrefixes []uint64, n int) {
	copy(scannedPrefixes[:4], chunkProducts[:4])
	for i := 1; i < n; i++ {
		var prev, cur, product bnfr.Element
		copy(prev[:], scannedPrefixes[(i-1)*4:i*4])
		copy(cur[:], chunkProducts[i*4:(i+1)*4])
		product.Mul(&prev, &cur)
		copy(scannedPrefixes[i*4:(i+1)*4], product[:])
	}
}

func scanChunkProductsBLS12377(chunkProducts, scannedPrefixes []uint64, n int) {
	copy(scannedPrefixes[:4], chunkProducts[:4])
	for i := 1; i < n; i++ {
		var prev, cur, product blsfr.Element
		copy(prev[:], scannedPrefixes[(i-1)*4:i*4])
		copy(cur[:], chunkProducts[i*4:(i+1)*4])
		product.Mul(&prev, &cur)
		copy(scannedPrefixes[i*4:(i+1)*4], product[:])
	}
}

func scanChunkProductsBW6761(chunkProducts, scannedPrefixes []uint64, n int) {
	copy(scannedPrefixes[:6], chunkProducts[:6])
	for i := 1; i < n; i++ {
		var prev, cur, product bwfr.Element
		copy(prev[:], scannedPrefixes[(i-1)*6:i*6])
		copy(cur[:], chunkProducts[i*6:(i+1)*6])
		product.Mul(&prev, &cur)
		copy(scannedPrefixes[i*6:(i+1)*6], product[:])
	}
}

func checkSameShape4(a, b, c, d *FrVector) error {
	return checkSameShape(a, b, c, d)
}

func checkSameShape(a *FrVector, others ...*FrVector) error {
	if err := a.checkLive(); err != nil {
		return err
	}
	for _, v := range others {
		if err := v.checkLive(); err != nil {
			return err
		}
		if a.dev != v.dev {
			return fmt.Errorf("plonk2: vectors are on different devices")
		}
		if a.curve != v.curve {
			return fmt.Errorf("plonk2: vector curve mismatch")
		}
		if a.n != v.n {
			return fmt.Errorf("plonk2: vector length mismatch")
		}
	}
	return nil
}
