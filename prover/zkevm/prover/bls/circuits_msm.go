package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
)

const (
	nbRowsPerG1Mul = nbFpLimbs + 3*nbG1Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
	nbRowsPerG2Mul = nbFpLimbs + 3*nbG2Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
)

type mulInstance[C convertable[T], T element] struct {
	CurrentAccumulator C                  `gnark:",public"`
	Scalar             sw_bls12381.Scalar `gnark:",public"`
	Point              C                  `gnark:",public"`
	NextAccumulator    C                  `gnark:",public"`
}

type multiMulCircuit[C convertable[T], T element] struct {
	Instances []mulInstance[C, T]
}

func (c *multiMulCircuit[C, T]) Define(api frontend.API) error {
	f, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	nbInstances := len(c.Instances)
	switch vv := any(c.Instances).(type) {
	case []mulInstance[g1ElementWizard, sw_bls12381.G1Affine]:
		for i := range nbInstances {
			tCurrent := vv[i].CurrentAccumulator.ToElement(api, f)
			tScalar := &c.Instances[i].Scalar
			tPoint := vv[i].Point.ToElement(api, f)
			tNext := vv[i].NextAccumulator.ToElement(api, f)
			if err := evmprecompiles.ECG1ScalarMulSumBLS(api, &tCurrent, &tPoint, tScalar, &tNext); err != nil {
				return fmt.Errorf("instance %d scalar mul sum: %w", i, err)
			}
		}
	case []mulInstance[g2ElementWizard, sw_bls12381.G2Affine]:
		for i := range nbInstances {
			tCurrent := vv[i].CurrentAccumulator.ToElement(api, f)
			tScalar := &c.Instances[i].Scalar
			tPoint := vv[i].Point.ToElement(api, f)
			tNext := vv[i].NextAccumulator.ToElement(api, f)
			if err := evmprecompiles.ECG2ScalarMulSumBLS(api, &tCurrent, &tPoint, tScalar, &tNext); err != nil {
				return fmt.Errorf("instance %d scalar mul sum: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("unknown group %T for bls msm circuit", vv)
	}
	return nil
}

func newMulCircuit(g group, limits *Limits) frontend.Circuit {
	switch g {
	case G1:
		return &multiMulCircuit[g1ElementWizard, sw_bls12381.G1Affine]{Instances: make([]mulInstance[g1ElementWizard, sw_bls12381.G1Affine], limits.NbG1MulInputInstances)}
	case G2:
		return &multiMulCircuit[g2ElementWizard, sw_bls12381.G2Affine]{Instances: make([]mulInstance[g2ElementWizard, sw_bls12381.G2Affine], limits.NbG2MulInputInstances)}
	default:
		panic(fmt.Sprintf("unknown group %s for bls msm circuit", g.String()))
	}
}
