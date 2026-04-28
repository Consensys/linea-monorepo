//go:build cuda

package plonk2

import (
	"fmt"
	"math/big"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
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
	if len(points) == 0 {
		return nil, fmt.Errorf("plonk2: point buffer must not be empty")
	}
	pointWords := 2 * info.BaseFieldLimbs
	if len(points)%pointWords != 0 {
		return nil, fmt.Errorf("plonk2: point word count %d is not a multiple of %d", len(points), pointWords)
	}
	count := len(points) / pointWords
	if len(scalars) != count*info.ScalarLimbs {
		return nil, fmt.Errorf("plonk2: scalar word count %d, want %d", len(scalars), count*info.ScalarLimbs)
	}
	windowBits := 16
	if info.ScalarBits > 320 {
		windowBits = 13
	}
	raw, err := g1MSMPippengerRaw(dev, curve, points, scalars, count, windowBits)
	if err != nil {
		return nil, err
	}
	return correctRawMontgomeryMSM(curve, raw)
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
