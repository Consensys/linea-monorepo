package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/fields_bls12381"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/math/bitslice"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

const (
	nbBits     = 16 // we use 128 bits limbs for the BLS12-381 field
	nbBits128  = 128
	nbBytes    = nbBits / 8 // 128 bits = 16 bytes
	nbBytes128 = nbBits128 / 8

	// BLS scalar field is 255 bits, and we use 2 limbs of 128 bits to represent
	nbFrLimbs    = 16 // (x_1, x_0) MSB order
	nbFrLimbs128 = 2

	// BLS base field is 381 bits, and we use 4 limbs of 128 bits to represent
	// it. However, the highest limb is always zero, but the arithmetization
	// keeps it for nice alignment. We pass it to the circuit but check
	// explicitly that its 0.
	nbFpLimbs    = 32 // (x_3, x_2, x_1, x_0) MSB order
	nbFpLimbs128 = 4

	nbG1Limbs    = 2 * nbFpLimbs  // (Ax, Ay)
	nbG2Limbs    = 4 * nbFpLimbs  // (BxIm, BxRe, ByIm, ByRe)
	nbGtLimbs    = 12 * nbFpLimbs // representation according to gnark - we don't use Gt in arithmetization, only in glue for accumulation
	nbG1Limbs128 = 2 * nbFpLimbs128
	nbG2Limbs128 = 4 * nbFpLimbs128
	nbGtLimbs128 = 12 * nbFpLimbs128
)

func nbLimbs(g Group) int {
	switch g {
	case G1:
		return nbG1Limbs
	case G2:
		return nbG2Limbs
	default:
		panic("unknown group for bls nbLimbs")
	}
}

func nbLimbs128(g Group) int {
	switch g {
	case G1:
		return nbG1Limbs128
	case G2:
		return nbG2Limbs128
	default:
		panic("unknown group for bls nbLimbs128")
	}
}

var fpParams sw_bls12381.BaseField
var frParams sw_bls12381.ScalarField

type scalarElementWizard struct {
	S [nbFrLimbs]frontend.Variable
}

func (c scalarElementWizard) ToElement(api frontend.API, fr *emulated.Field[sw_bls12381.ScalarField]) *sw_bls12381.Scalar {
	S16 := make([]frontend.Variable, 16)
	copy(S16[0:8], c.S[8:16])
	copy(S16[8:16], c.S[0:8])
	S := gnarkutil.EmulatedFromLimbSlice(api, fr, S16, 16)
	return S
}

type baseElementWizard struct {
	P [nbFpLimbs]frontend.Variable
}

func (c baseElementWizard) ToElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) *emulated.Element[sw_bls12381.BaseField] {
	P16 := make([]frontend.Variable, 24)
	copy(P16[0:8], c.P[24:32])
	copy(P16[8:16], c.P[16:24])
	copy(P16[16:24], c.P[8:16])
	return gnarkutil.EmulatedFromLimbSlice(api, fp, P16, 16)
}

type g1ElementWizard struct {
	P [nbG1Limbs]frontend.Variable
}

func (c g1ElementWizard) ToElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) *sw_bls12381.G1Affine {
	// gnark represents the BLS12-381 Fp element on 6 limbs of 64 bits.
	// Arithmetization uses 4 limbs of 128 bits, but the MSB limb is always 0.
	// Arithmetization layout for G1 point P:
	//  - limbs 0-32 are X
	//  - limbs 32-64 are Y
	//
	// Furthermore, arithmetization uses orders the limbs in HI-LO format:
	//  - limbs 0-16 are X_HI
	//  - limbs 16-32 are X_LO
	//  - limbs 32-48 are Y_HI
	//  - limbs 48-64 are Y_LO
	//
	// As BLS12-381 requires only 381 bits, then highest 0 limbs are expected to be 0:
	// - limbs 0-8 of X_HI are 0
	// - limbs 0-8 of Y_HI are 0
	for i := range 8 {
		api.AssertIsEqual(c.P[i], 0)
		api.AssertIsEqual(c.P[nbFpLimbs+i], 0)
	}
	PX16 := make([]frontend.Variable, 24)
	copy(PX16[0:8], c.P[24:32])
	copy(PX16[8:16], c.P[16:24])
	copy(PX16[16:24], c.P[8:16])
	PY16 := make([]frontend.Variable, 24)
	copy(PY16[0:8], c.P[56:64])
	copy(PY16[8:16], c.P[48:56])
	copy(PY16[16:24], c.P[40:48])
	PX := gnarkutil.EmulatedFromLimbSlice(api, fp, PX16, 16)
	PY := gnarkutil.EmulatedFromLimbSlice(api, fp, PY16, 16)
	P := sw_bls12381.G1Affine{
		X: *PX,
		Y: *PY,
	}
	return &P
}

type g2ElementWizard struct {
	Q [nbG2Limbs]frontend.Variable
}

func (c g2ElementWizard) ToElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) *sw_bls12381.G2Affine {
	// see comments in g1ElementWizard.ToElement regarding limb ordering and zero
	// padding.
	// Arithmetization layout for G2 point Q:
	//  - limbs 0-64 are X = (X_im, X_re)
	//  - limbs 64-128 are Y = (Y_im, Y_re)
	for i := range 8 {
		api.AssertIsEqual(c.Q[i], 0)
		api.AssertIsEqual(c.Q[nbFpLimbs+i], 0)
		api.AssertIsEqual(c.Q[2*nbFpLimbs+i], 0)
		api.AssertIsEqual(c.Q[3*nbFpLimbs+i], 0)
	}
	QXA16 := make([]frontend.Variable, 24)
	QXB16 := make([]frontend.Variable, 24)
	QYA16 := make([]frontend.Variable, 24)
	QYB16 := make([]frontend.Variable, 24)
	// 0-32 are XA
	copy(QXA16[0:8], c.Q[24:32])
	copy(QXA16[8:16], c.Q[16:24])
	copy(QXA16[16:24], c.Q[8:16])
	// 32-64 are XB
	copy(QXB16[0:8], c.Q[56:64])
	copy(QXB16[8:16], c.Q[48:56])
	copy(QXB16[16:24], c.Q[40:48])
	// 64-96 are YA
	copy(QYA16[0:8], c.Q[88:96])
	copy(QYA16[8:16], c.Q[80:88])
	copy(QYA16[16:24], c.Q[72:80])
	// 96-128 are YB
	copy(QYB16[0:8], c.Q[120:128])
	copy(QYB16[8:16], c.Q[112:120])
	copy(QYB16[16:24], c.Q[104:112])

	QXA := gnarkutil.EmulatedFromLimbSlice(api, fp, QXA16, 16)
	QXB := gnarkutil.EmulatedFromLimbSlice(api, fp, QXB16, 16)
	QYA := gnarkutil.EmulatedFromLimbSlice(api, fp, QYA16, 16)
	QYB := gnarkutil.EmulatedFromLimbSlice(api, fp, QYB16, 16)
	QX := fields_bls12381.E2{
		A0: *QXA,
		A1: *QXB,
	}

	QY := fields_bls12381.E2{
		A0: *QYA,
		A1: *QYB,
	}

	var Q sw_bls12381.G2Affine
	Q.P.X = QX
	Q.P.Y = QY

	return &Q
}

type gtElementWizard struct {
	T [nbGtLimbs]frontend.Variable
}

func (c gtElementWizard) ToElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) *sw_bls12381.GTEl {
	A0Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A1Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A2Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A3Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A4Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A5Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A6Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A7Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A8Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A9Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A10Limbs := make([]frontend.Variable, fpParams.NbLimbs())
	A11Limbs := make([]frontend.Variable, fpParams.NbLimbs())

	// assert that the MSB limb is 0 in arithmetization
	for i := range 12 {
		api.AssertIsEqual(c.T[i*nbFpLimbs], 0)
	}

	for i := range nbFpLimbs - 1 {
		A0Limbs[len(A0Limbs)-(2*i+2)], A0Limbs[len(A0Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[i+1], 64, bitslice.WithNbDigits(128))
		A1Limbs[len(A1Limbs)-(2*i+2)], A1Limbs[len(A1Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A2Limbs[len(A2Limbs)-(2*i+2)], A2Limbs[len(A2Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[2*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A3Limbs[len(A3Limbs)-(2*i+2)], A3Limbs[len(A3Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[3*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A4Limbs[len(A4Limbs)-(2*i+2)], A4Limbs[len(A4Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[4*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A5Limbs[len(A5Limbs)-(2*i+2)], A5Limbs[len(A5Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[5*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A6Limbs[len(A6Limbs)-(2*i+2)], A6Limbs[len(A6Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[6*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A7Limbs[len(A7Limbs)-(2*i+2)], A7Limbs[len(A7Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[7*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A8Limbs[len(A8Limbs)-(2*i+2)], A8Limbs[len(A8Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[8*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A9Limbs[len(A9Limbs)-(2*i+2)], A9Limbs[len(A9Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[9*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A10Limbs[len(A10Limbs)-(2*i+2)], A10Limbs[len(A10Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[10*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
		A11Limbs[len(A11Limbs)-(2*i+2)], A11Limbs[len(A11Limbs)-(2*i+1)] = bitslice.Partition(api, c.T[11*nbFpLimbs+i+1], 64, bitslice.WithNbDigits(128))
	}

	pairing, err := sw_bls12381.NewPairing(api)
	if err != nil {
		panic(fmt.Sprintf("new pairing: %v", err))
	}
	T := pairing.Ext12.FromTower([12]*emulated.Element[sw_bls12381.BaseField]{
		fp.NewElement(A0Limbs),
		fp.NewElement(A1Limbs),
		fp.NewElement(A2Limbs),
		fp.NewElement(A3Limbs),
		fp.NewElement(A4Limbs),
		fp.NewElement(A5Limbs),
		fp.NewElement(A6Limbs),
		fp.NewElement(A7Limbs),
		fp.NewElement(A8Limbs),
		fp.NewElement(A9Limbs),
		fp.NewElement(A10Limbs),
		fp.NewElement(A11Limbs),
	})

	return T
}

type element interface {
	sw_bls12381.G1Affine | sw_bls12381.G2Affine | sw_bls12381.GTEl
}

type convertable[T element] interface {
	g1ElementWizard | g2ElementWizard | gtElementWizard
	ToElement(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField]) *T
}
