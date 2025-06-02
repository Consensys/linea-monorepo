package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/fields_bls12381"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/math/emulated"
)

type MultiMapToG1Circuit[C convertable[T], T element] struct {
	Instances []MapToInstance[C, T] `gnark:",public"`
}

func newMultiMapToG1Circuit(g group, limits *Limits) frontend.Circuit {
	switch g {
	case G1:
		res := &MultiMapToG1Circuit[g1ElementWizard, sw_bls12381.G1Affine]{
			Instances: make([]MapToInstance[g1ElementWizard, sw_bls12381.G1Affine], limits.NbG1MapToInputInstances),
		}
		for i := range res.Instances {
			res.Instances[i].Input = make([]emulated.Element[sw_bls12381.BaseField], 1)
		}
		return res
	case G2:
		res := &MultiMapToG1Circuit[g2ElementWizard, sw_bls12381.G2Affine]{
			Instances: make([]MapToInstance[g2ElementWizard, sw_bls12381.G2Affine], limits.NbG2MapToInputInstances),
		}
		for i := range res.Instances {
			res.Instances[i].Input = make([]emulated.Element[sw_bls12381.BaseField], 2)
		}
		return res
	default:
		panic(fmt.Sprintf("unknown group %s for bls map to G1 circuit", g.String()))
	}
}

type MapToInstance[C convertable[T], T element] struct {
	Input  []emulated.Element[sw_bls12381.BaseField] // len==1 for G1, len==2 for G2
	Mapped C
}

func (c *MultiMapToG1Circuit[C, T]) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	var t T
	switch any(t).(type) {
	case sw_bls12381.G1Affine:
		// TODO: when have fixed evmprecompiles interfaces then use that
		g1, err := sw_bls12381.NewG1(api)
		if err != nil {
			return fmt.Errorf("new G1: %w", err)
		}
		for i := range c.Instances {
			if len(c.Instances[i].Input) != 1 {
				return fmt.Errorf("expected 1 input for G1 map to G1, got %d", len(c.Instances[i].Input))
			}
			tMapped := c.Instances[i].Mapped.ToElement(api, fp)
			res := g1.MapToG1(c.Instances[i].Input[0])
			g1.AssertIsEqual(res, &tMapped)
		}
	case sw_bls12381.G2Affine:
		g2 := sw_bls12381.NewG2(api)
		for i := range c.Instances {
			if len(c.Instances[i].Input) != 2 {
				return fmt.Errorf("expected 2 inputs for G2 map to G1, got %d", len(c.Instances[i].Input))
			}
			tMapped := c.Instances[i].Mapped.ToElement(api, fp)
			res := g2.MapToG2(fields_bls12381.E2{
				A0: c.Instances[i].Input[0],
				A1: c.Instances[i].Input[1],
			})
			g2.AssertIsEqual(res, &tMapped)
		}
	default:
		return fmt.Errorf("unknown group %T for bls map to G1 circuit", t)
	}
	return nil
}
