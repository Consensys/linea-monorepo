package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
)

// -- checks for being on curve

type C1IsOnCurveInstance struct {
	// P is the purported C1 element
	P g1ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

func (c C1IsOnCurveInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	P := c.P.ToElement(api, fp)
	res := pairing.IsOnCurve(&P)
	api.AssertIsEqual(res, c.IsSuccess)
	return nil
}

// -- checks for small group non-membership. We perform membership checks inside the main circuits. This is only for asserting invalid inputs for MSM and PAIRING.

type G1NonMembershipInstance struct {
	// P is the purported G1 element
	P g1ElementWizard
}

func (c G1NonMembershipInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	P := c.P.ToElement(api, fp)
	return evmprecompiles.ECPairBLSIsOnG1(api, &P, 0) // 0 means we expect it to be not on G1
}

// -- checks for being on twist

type C2IsOnCurveInstance struct {
	// Q is the purported C2 element
	Q g2ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

func (c C2IsOnCurveInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	Q := c.Q.ToElement(api, fp)
	res := pairing.IsOnTwist(&Q)
	api.AssertIsEqual(res, c.IsSuccess)
	return nil
}

// -- checks for large group non-membership.
type G2NonMembershipInstance struct {
	// Q is the purported G2 element
	Q g2ElementWizard
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

func (c G2NonMembershipInstance) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	Q := c.Q.ToElement(api, fp)
	return evmprecompiles.ECPairBLSIsOnG2(api, &Q, 0) // 0 means we expect it to be not on G2
}

// -- circuit which performs multiple checks

type checkableInstance interface {
	C1IsOnCurveInstance | G2NonMembershipInstance | C2IsOnCurveInstance | G1NonMembershipInstance
	Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error
}

type multiCheckableCircuit[T checkableInstance] struct {
	Instance []T `gnark:",public"`
}

func newMultiCheckableCircuit[T checkableInstance](nbInstances int) *multiCheckableCircuit[T] {
	return &multiCheckableCircuit[T]{
		Instance: make([]T, nbInstances),
	}
}

func (c *multiCheckableCircuit[T]) Define(api frontend.API) error {
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
