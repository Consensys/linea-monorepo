package bls

import (
	"fmt"
	"math/big"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

func set(nbL int, q *fp.Element, limbs []field.Element) {
	var buf []byte
	for i := range nbL/2 - 1 {
		lbts := limbs[i+1].Bytes()
		buf = append(buf, lbts[nbBytes:]...)
	}
	q.SetBytes(buf)
}

func nativeScalarMulAndSum(g group, currentAccumulator []field.Element, point []field.Element, scalar []field.Element) (nextAccumulator []field.Element) {
	nbL := nbLimbs(g)
	if len(scalar) != nbFrLimbs {
		panic(fmt.Sprintf("scalar must have exactly %d limbs, got %d", nbFrLimbs, len(scalar)))
	}
	if len(currentAccumulator) != nbL {
		panic(fmt.Sprintf("currentAccumulator must have exactly %d limbs, got %d", nbL, len(currentAccumulator)))
	}
	if len(point) != nbL {
		panic(fmt.Sprintf("point must have exactly %d limbs, got %d", nbL, len(point)))
	}
	var buf []byte
	for i := range nbFrLimbs {
		lbts := scalar[i].Bytes()
		buf = append(buf, lbts[nbBytes:]...)
	}
	s := new(big.Int).SetBytes(buf)

	switch g {
	case G1:
		var C, P, N bls12381.G1Affine
		set(nbL, &P.X, point[:nbL/2])
		set(nbL, &P.Y, point[nbL/2:])
		set(nbL, &C.X, currentAccumulator[:nbL/2])
		set(nbL, &C.Y, currentAccumulator[nbL/2:])
		N.ScalarMultiplication(&P, s)
		N.Add(&C, &N)
		NXBytes := N.X.Bytes()
		NYBytes := N.Y.Bytes()
		nextAccumulator = make([]field.Element, nbL)
		nextAccumulator[0].SetZero()
		nextAccumulator[nbL/2].SetZero()
		for i := 0; i < nbL/2-1; i++ {
			nextAccumulator[i+1].SetBytes(NXBytes[i*nbBytes : (i+1)*nbBytes])
			nextAccumulator[i+1+nbL/2].SetBytes(NYBytes[i*nbBytes : (i+1)*nbBytes])
		}
	case G2:
		var C, P, N bls12381.G2Affine
		set(nbL/2, &P.X.A1, point[:nbL/4])
		set(nbL/2, &P.X.A0, point[nbL/4:nbL/2])
		set(nbL/2, &P.Y.A1, point[nbL/2:3*nbL/4])
		set(nbL/2, &P.Y.A0, point[3*nbL/4:])
		set(nbL/2, &C.X.A1, currentAccumulator[:nbL/4])
		set(nbL/2, &C.X.A0, currentAccumulator[nbL/4:nbL/2])
		set(nbL/2, &C.Y.A1, currentAccumulator[nbL/2:3*nbL/4])
		set(nbL/2, &C.Y.A0, currentAccumulator[3*nbL/4:])
		N.ScalarMultiplication(&P, s)
		N.Add(&C, &N)
		NXA1Bytes := N.X.A1.Bytes()
		NXA0Bytes := N.X.A0.Bytes()
		NYA1Bytes := N.Y.A1.Bytes()
		NYA0Bytes := N.Y.A0.Bytes()
		nextAccumulator = make([]field.Element, nbL)
		nextAccumulator[0].SetZero()
		nextAccumulator[nbL/4].SetZero()
		nextAccumulator[nbL/2].SetZero()
		nextAccumulator[3*nbL/4].SetZero()
		for i := 0; i < nbL/4-1; i++ {
			nextAccumulator[i+1].SetBytes(NXA1Bytes[i*nbBytes : (i+1)*nbBytes])
			nextAccumulator[i+1+nbL/4].SetBytes(NXA0Bytes[i*nbBytes : (i+1)*nbBytes])
			nextAccumulator[i+1+nbL/2].SetBytes(NYA1Bytes[i*nbBytes : (i+1)*nbBytes])
			nextAccumulator[i+1+3*nbL/4].SetBytes(NYA0Bytes[i*nbBytes : (i+1)*nbBytes])
		}
	}
	return nextAccumulator
}
