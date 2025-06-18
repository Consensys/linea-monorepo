package main

import (
	"fmt"
	"math/big"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	fp_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	fr_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

type mapInputType int
type pairInputType int

const (
	mapTrivial    mapInputType = iota // 0
	mapInRange                        // element is in range of the field
	mapOutOfRange                     // element is not in range of the field
)

const (
	pairFullTrivial            pairInputType = iota // (0, 0)
	pairLeftTrivialValid                            // (0, Q)
	pairLeftTrivialInvalid                          // (0, Q') where Q' is not in G2
	pairRightTrivialValid                           // (P, 0)
	pairRightTrivialInvalid                         // (P', 0) where P' is not in G1
	pairNonTrivialLeftInvalid                       // (P, Q) where P is not in G1
	pairNonTrivialRightInvalid                      // (P, Q) where Q is not in G2
	pairFullInvalid                                 // (P, Q) where P and Q not in G1 and G2 respectively
	pairNonTrivial                                  // (P, Q)
)

type affine interface {
	bls12381.G1Affine | bls12381.G2Affine
}

func generateG1Trivial() bls12381.G1Affine {
	var p bls12381.G1Affine
	p.SetInfinity()
	return p
}

func generateG2Trivial() bls12381.G2Affine {
	var p bls12381.G2Affine
	p.SetInfinity()
	return p
}

func generateTrivial[T affine]() T {
	var res T
	switch vv := any(&res).(type) {
	case *bls12381.G1Affine:
		v := generateG1Trivial()
		*vv = v
	case *bls12381.G2Affine:
		v := generateG2Trivial()
		*vv = v
	}
	return res
}

func generateG1OnCurve() bls12381.G1Affine {
	var x fp_bls12381.Element
	x.MustSetRandom()
	P := bls12381.GeneratePointNotInG1(x)
	var pa bls12381.G1Affine
	pa.FromJacobian(&P)
	if !pa.IsOnCurve() {
		panic("generated point is not on curve")
	}
	if pa.IsInSubGroup() {
		panic("generated point is in subgroup")
	}
	return pa
}

func generateG2OnCurve() bls12381.G2Affine {
	var x, y fp_bls12381.Element
	x.MustSetRandom()
	y.MustSetRandom()
	e := bls12381.E2{A0: x, A1: y}
	P := bls12381.GeneratePointNotInG2(e)
	var pa bls12381.G2Affine
	pa.FromJacobian(&P)
	if !pa.IsOnCurve() {
		panic("generated point is not on curve")
	}
	if pa.IsInSubGroup() {
		panic("generated point is in subgroup")
	}
	return pa
}

func generateOnCurve[T affine]() T {
	var res T
	switch vv := any(&res).(type) {
	case *bls12381.G1Affine:
		v := generateG1OnCurve()
		*vv = v
	case *bls12381.G2Affine:
		v := generateG2OnCurve()
		*vv = v
	}
	return res
}

func generateG1InSubgroup() bls12381.G1Affine {
	var p bls12381.G1Affine
	var s fr_bls12381.Element
	s.MustSetRandom()
	p.ScalarMultiplicationBase(s.BigInt(new(big.Int)))
	return p
}

func generateG2InSubgroup() bls12381.G2Affine {
	var p bls12381.G2Affine
	var s fr_bls12381.Element
	s.MustSetRandom()
	p.ScalarMultiplicationBase(s.BigInt(new(big.Int)))
	return p
}

func generateInSubgroup[T affine]() T {
	var res T
	switch vv := any(&res).(type) {
	case *bls12381.G1Affine:
		v := generateG1InSubgroup()
		*vv = v
	case *bls12381.G2Affine:
		v := generateG2InSubgroup()
		*vv = v
	}
	return res
}

func generateG1Invalid() bls12381.G1Affine {
	var p bls12381.G1Affine
	for {
		p.X.SetRandom()
		p.Y.SetRandom()
		if !p.IsOnCurve() {
			return p
		}
	}
}

func generateG2Invalid() bls12381.G2Affine {
	var p bls12381.G2Affine
	for {
		p.X.A0.SetRandom()
		p.X.A1.SetRandom()
		p.Y.A0.SetRandom()
		p.Y.A1.SetRandom()
		if !p.IsOnCurve() {
			return p
		}
	}
}

func generateInvalid[T affine]() T {
	var res T
	switch vv := any(&res).(type) {
	case *bls12381.G1Affine:
		v := generateG1Invalid()
		*vv = v
	case *bls12381.G2Affine:
		v := generateG2Invalid()
		*vv = v
	}
	return res
}

func splitG1ToLimbs(p bls12381.G1Affine) []string {
	px := p.X.Bytes()
	py := p.Y.Bytes()
	limbs := []string{
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", px[0:16]),
		fmt.Sprintf("0x%x", px[16:32]),
		fmt.Sprintf("0x%x", px[32:48]),
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", py[0:16]),
		fmt.Sprintf("0x%x", py[16:32]),
		fmt.Sprintf("0x%x", py[32:48]),
	}
	return limbs
}

func splitG2ToLimbs(q bls12381.G2Affine) []string {
	qxre := q.X.A0.Bytes()
	qxim := q.X.A1.Bytes()
	qyre := q.Y.A0.Bytes()
	qyim := q.Y.A1.Bytes()
	limbs := []string{
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", qxim[0:16]),
		fmt.Sprintf("0x%x", qxim[16:32]),
		fmt.Sprintf("0x%x", qxim[32:48]),
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", qxre[0:16]),
		fmt.Sprintf("0x%x", qxre[16:32]),
		fmt.Sprintf("0x%x", qxre[32:48]),
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", qyim[0:16]),
		fmt.Sprintf("0x%x", qyim[16:32]),
		fmt.Sprintf("0x%x", qyim[32:48]),
		"0x00000000000000000000000000000000",
		fmt.Sprintf("0x%x", qyre[0:16]),
		fmt.Sprintf("0x%x", qyre[16:32]),
		fmt.Sprintf("0x%x", qyre[32:48]),
	}
	return limbs
}

func splitToLimbs[T affine](p T) []string {
	switch vv := any(p).(type) {
	case bls12381.G1Affine:
		return splitG1ToLimbs(vv)
	case bls12381.G2Affine:
		return splitG2ToLimbs(vv)
	default:
		panic(fmt.Sprintf("unknown type for splitting to limbs: %T", p))
	}
}

func splitScalarToLimbs(s fr_bls12381.Element) []string {
	sb := s.Bytes()
	limbs := []string{
		fmt.Sprintf("0x%x", sb[0:16]),
		fmt.Sprintf("0x%x", sb[16:32]),
	}
	return limbs
}

func formatBoolAsInt(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
