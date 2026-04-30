//go:build cuda

package plonk2

import (
	"fmt"
	"math/big"
	"os"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfp "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

const (
	bn254CPUFallbackPointLimit  = 1 << 13
	bw6761CPUFallbackPointLimit = 1 << 14
)

var (
	bn254ScalarRawRInv    = montgomeryRawInverse(bnfr.Modulus(), bnfr.Limbs)
	bls12377ScalarRawRInv = montgomeryRawInverse(blsfr.Modulus(), blsfr.Limbs)
	bw6761ScalarRawRInv   = montgomeryRawInverse(bwfr.Modulus(), bwfr.Limbs)
)

// CommitRaw computes a KZG-style commitment sum_i scalars[i]*points[i].
//
// points must be gnark-crypto G1Affine raw memory in short-Weierstrass affine
// layout: X limbs followed by Y limbs for each point. scalars must be raw
// gnark-crypto Fr element memory. The returned point is projective raw memory:
// X, Y, Z base-field limbs.
func CommitRaw(dev *gpu.Device, curve Curve, points, scalars []uint64) ([]uint64, error) {
	info, err := curve.validate()
	if err != nil {
		return nil, err
	}
	_, count, err := validateRawAffinePoints(curve, points)
	if err != nil {
		return nil, err
	}
	if err := validateRawScalarsExact(curve, scalars, count); err != nil {
		return nil, err
	}
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}
	if shouldUseCPUFallback(curve, count) {
		return commitRawCPUFallback(curve, points, scalars, count)
	}
	windowBits := defaultMSMWindowBits(info, count)
	raw, err := g1MSMPippengerRaw(dev, curve, points, scalars, count, windowBits)
	if err != nil {
		return nil, err
	}
	return correctRawMontgomeryMSM(curve, raw)
}

func shouldUseCPUFallback(curve Curve, count int) bool {
	switch curve {
	case CurveBN254:
		return count < bn254CPUFallbackPointLimit &&
			os.Getenv("PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK") != "1"
	case CurveBW6761:
		return count < bw6761CPUFallbackPointLimit &&
			os.Getenv("PLONK2_BW6_MSM_DISABLE_CPU_FALLBACK") != "1"
	default:
		return false
	}
}

func commitRawCPUFallback(curve Curve, points, scalars []uint64, count int) ([]uint64, error) {
	switch curve {
	case CurveBN254:
		return commitRawBN254CPU(points, scalars, count)
	case CurveBW6761:
		return commitRawBW6761CPU(points, scalars, count)
	default:
		return nil, fmt.Errorf("plonk2: unsupported CPU MSM fallback curve %d", curve)
	}
}

func commitRawBN254CPU(pointsRaw, scalarsRaw []uint64, count int) ([]uint64, error) {
	if count == 0 {
		return nil, fmt.Errorf("plonk2: scalar buffer must not be empty")
	}
	points := unsafe.Slice(
		(*bn254.G1Affine)(unsafe.Pointer(&pointsRaw[0])),
		count,
	)
	scalars := unsafe.Slice(
		(*bnfr.Element)(unsafe.Pointer(&scalarsRaw[0])),
		count,
	)
	if allRawBN254PointsEqual(pointsRaw, count) {
		var scalarSum bnfr.Element
		for i := range scalars {
			scalarSum.Add(&scalarSum, &scalars[i])
		}
		var scalarBig big.Int
		scalarSum.BigInt(&scalarBig)
		var base bn254.G1Jac
		base.FromAffine(&points[0])
		var got bn254.G1Jac
		got.ScalarMultiplication(&base, &scalarBig)
		return cloneRawProjective(unsafe.Pointer(&got), 3*bnfp.Limbs), nil
	}
	var got bn254.G1Jac
	if _, err := got.MultiExp(points, scalars, ecc.MultiExpConfig{}); err != nil {
		return nil, err
	}
	return cloneRawProjective(unsafe.Pointer(&got), 3*bnfp.Limbs), nil
}

func allRawBN254PointsEqual(pointsRaw []uint64, count int) bool {
	if count <= 1 {
		return true
	}
	const pointWords = 2 * bnfp.Limbs
	first := pointsRaw[:pointWords]
	for i := 1; i < count; i++ {
		point := pointsRaw[i*pointWords : (i+1)*pointWords]
		for j := range first {
			if first[j] != point[j] {
				return false
			}
		}
	}
	return true
}

func commitRawBW6761CPU(pointsRaw, scalarsRaw []uint64, count int) ([]uint64, error) {
	if count == 0 {
		return nil, fmt.Errorf("plonk2: scalar buffer must not be empty")
	}
	points := unsafe.Slice(
		(*bw6761.G1Affine)(unsafe.Pointer(&pointsRaw[0])),
		count,
	)
	scalars := unsafe.Slice(
		(*bwfr.Element)(unsafe.Pointer(&scalarsRaw[0])),
		count,
	)
	var got bw6761.G1Jac
	if _, err := got.MultiExp(points, scalars, ecc.MultiExpConfig{}); err != nil {
		return nil, err
	}
	return cloneRawProjective(unsafe.Pointer(&got), 3*bwfp.Limbs), nil
}

func correctRawMontgomeryMSM(curve Curve, raw []uint64) ([]uint64, error) {
	switch curve {
	case CurveBN254:
		var p bn254.G1Jac
		copy(unsafe.Slice((*uint64)(unsafe.Pointer(&p)), len(raw)), raw)
		p.ScalarMultiplication(&p, &bn254ScalarRawRInv)
		return cloneRawProjective(unsafe.Pointer(&p), len(raw)), nil
	case CurveBLS12377:
		var p bls12377.G1Jac
		copy(unsafe.Slice((*uint64)(unsafe.Pointer(&p)), len(raw)), raw)
		p.ScalarMultiplication(&p, &bls12377ScalarRawRInv)
		return cloneRawProjective(unsafe.Pointer(&p), len(raw)), nil
	case CurveBW6761:
		var p bw6761.G1Jac
		copy(unsafe.Slice((*uint64)(unsafe.Pointer(&p)), len(raw)), raw)
		p.ScalarMultiplication(&p, &bw6761ScalarRawRInv)
		return cloneRawProjective(unsafe.Pointer(&p), len(raw)), nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported curve %d", curve)
	}
}

func cloneRawProjective(point unsafe.Pointer, words int) []uint64 {
	raw := unsafe.Slice((*uint64)(point), words)
	out := make([]uint64, len(raw))
	copy(out, raw)
	return out
}

func montgomeryRawInverse(modulus *big.Int, limbs int) big.Int {
	r := new(big.Int).Lsh(big.NewInt(1), uint(64*limbs))
	r.Mod(r, modulus)
	var inv big.Int
	inv.ModInverse(r, modulus)
	return inv
}
