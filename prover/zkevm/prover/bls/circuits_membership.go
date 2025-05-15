package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/math/emulated"
)

type C1CurveCheckInstance struct {
	// P is the purported C1 element
	P g1ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

func (c *C1CurveCheckInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	P := c.P.ToG1Element(api, fp)
	res := pairing.IsOnG1(&P)
	api.AssertIsEqual(res, c.IsSuccess)

	return nil
}

type G1GroupCheckInstance struct {
	// P is the purported G1 element
	P g1ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

type C2CurveCheckInstance struct {
	// Q is the purported C2 element
	Q g2ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

type G2GroupCheckInstance struct {
	// Q is the purported G2 element
	Q g2ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

type MultiC1CircuitCheckCircuit struct {
	Instance []C1CurveCheckInstance `gnark:",public"`
}

func newC1CurveCheckCircuit(nbInstances int) *MultiC1CircuitCheckCircuit {
	return &MultiC1CircuitCheckCircuit{
		Instance: make([]C1CurveCheckInstance, nbInstances),
	}
}

func (c *MultiC1CircuitCheckCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field emulation: %w", err)
	}
	pairing, err := sw_bls12381.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instance {
		if err := c.Instance[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("instance %d check: %w", i, err)
		}
	}
	return nil
}
