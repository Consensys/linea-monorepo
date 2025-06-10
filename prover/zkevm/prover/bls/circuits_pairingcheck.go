package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
)

type MultiMillerLoopMulCircuit struct {
	Instances []MillerLoopMulInstance `gnark:",public"`
}

func newMultiMillerLoopMulCircuit(limits *Limits) *MultiMillerLoopMulCircuit {
	return &MultiMillerLoopMulCircuit{
		Instances: make([]MillerLoopMulInstance, limits.NbMillerLoopInputInstances),
	}
}

func (c *MultiMillerLoopMulCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	pairing, err := sw_bls12381.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("check instance %d: %w", i, err)
		}
	}
	return nil
}

type MillerLoopMulInstance struct {
	Prev    gtElementWizard
	P       g1ElementWizard
	Q       g2ElementWizard
	Current gtElementWizard
}

func (c *MillerLoopMulInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	prev := c.Prev.ToElement(api, fp)
	p := c.P.ToElement(api, fp)
	q := c.Q.ToElement(api, fp)
	current := c.Current.ToElement(api, fp)

	return evmprecompiles.ECPairBLSMillerLoopAndMul(api, &prev, &p, &q, &current)
}

type MultiMillerLoopFinalExpCircuit struct {
	Instances []MillerLoopFinalExpInstance `gnark:",public"`
}

func newMultiMillerLoopFinalExpCircuit(limits *Limits) *MultiMillerLoopFinalExpCircuit {
	return &MultiMillerLoopFinalExpCircuit{
		Instances: make([]MillerLoopFinalExpInstance, limits.NbFinalExpInputInstances),
	}
}

func (c *MultiMillerLoopFinalExpCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	pairing, err := sw_bls12381.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("check instance %d: %w", i, err)
		}
	}
	return nil
}

type MillerLoopFinalExpInstance struct {
	Prev     gtElementWizard
	P        g1ElementWizard
	Q        g2ElementWizard
	Expected [2]frontend.Variable
}

func (c *MillerLoopFinalExpInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	prev := c.Prev.ToElement(api, fp)
	p := c.P.ToElement(api, fp)
	q := c.Q.ToElement(api, fp)

	api.AssertIsEqual(c.Expected[0], 0)
	return evmprecompiles.ECPairBLSMillerLoopAndFinalExpCheck(api, &prev, &p, &q, c.Expected[1])
}
