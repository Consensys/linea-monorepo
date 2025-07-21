package bls

import (
	"fmt"
	"math/big"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

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

	buf = nil
	var resBytesX, resBytesY []byte
	switch g {
	case G1:
		var C, P, N bls12381.G1Affine
		for i := range nbFrLimbs {
			lbts := point[i].Bytes()
			buf = append(buf, lbts[nbBytes:]...)
		}
		P.SetBytes(buf)
		buf = nil
		for i := range nbL {
			lbts := currentAccumulator[i].Bytes()
			buf = append(buf, lbts[nbBytes:]...)
		}
		C.SetBytes(buf)
		N.ScalarMultiplication(&P, s)
		N.Add(&C, &N)
		NXBytes := N.X.Bytes()
		NYBytes := N.Y.Bytes()
		resBytesX = NXBytes[:]
		resBytesY = NYBytes[:]
	case G2:
		var C, P, N bls12381.G2Affine
		for i := range nbFrLimbs {
			lbts := point[i].Bytes()
			buf = append(buf, lbts[nbBytes:]...)
		}
		P.SetBytes(buf)
		buf = nil
		for i := range nbL {
			lbts := currentAccumulator[i].Bytes()
			buf = append(buf, lbts[nbBytes:]...)
		}
		C.SetBytes(buf)
		N.ScalarMultiplication(&P, s)
		N.Add(&C, &N)
		NXA1Bytes := N.X.A1.Bytes()
		NXA0Bytes := N.X.A0.Bytes()
		NYA1Bytes := N.Y.A1.Bytes()
		NYA0Bytes := N.Y.A0.Bytes()
		resBytesX = make([]byte, 0, 2*len(NXA1Bytes))
		resBytesX = append(resBytesX, NXA1Bytes[:]...)
		resBytesX = append(resBytesX, NXA0Bytes[:]...)
		resBytesY = make([]byte, 0, 2*len(NYA1Bytes))
		resBytesY = append(resBytesY, NYA1Bytes[:]...)
		resBytesY = append(resBytesY, NYA0Bytes[:]...)
	}
	nextAccumulator = make([]field.Element, nbL)
	nextAccumulator[0].SetZero()
	nextAccumulator[nbL/2].SetZero()
	for i := 0; i < nbL/2-1; i++ {
		nextAccumulator[i+1].SetBytes(resBytesX[i*nbBytes : (i+1)*nbBytes])
		nextAccumulator[i+1+nbL/2].SetBytes(resBytesY[i*nbBytes : (i+1)*nbBytes])
	}
	return nextAccumulator
}
