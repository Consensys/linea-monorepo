package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
)

type OnCurveInstance[C convertable[T], T element] struct {
	// P is the purported element
	P C
	// IsSuccess is 1 if the check is successful, 0 otherwise
	IsSuccess frontend.Variable
}

func (c OnCurveInstance[C, T]) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	switch v := any(c.P).(type) {
	case g1ElementWizard:
		P := v.ToElement(api, fp)
		res := pairing.IsOnCurve(&P)
		api.AssertIsEqual(res, c.IsSuccess)
		return nil
	case g2ElementWizard:
		Q := v.ToElement(api, fp)
		res := pairing.IsOnTwist(&Q)
		api.AssertIsEqual(res, c.IsSuccess)
		return nil
	default:
		return fmt.Errorf("unsupported element type %T for on-curve check", c.P)
	}
}

type NonGroupMembershipInstance[C convertable[T], T element] struct {
	// P is the purported element
	P C
}

func (c NonGroupMembershipInstance[C, T]) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	switch v := any(c.P).(type) {
	case g1ElementWizard:
		P := v.ToElement(api, fp)
		// We expect the element to be not on G1
		return evmprecompiles.ECPairBLSIsOnG1(api, &P, 0) // 0 means we expect it to be not on G1
	case g2ElementWizard:
		Q := v.ToElement(api, fp)
		// We expect the element to be not on G2
		return evmprecompiles.ECPairBLSIsOnG2(api, &Q, 0) // 0 means we expect it to be not on G2
	default:
		return fmt.Errorf("unsupported element type %T for non-group membership check", c.P)
	}
}

// -- circuit which performs multiple checks

type checkableInstance interface {
	OnCurveInstance[g1ElementWizard, sw_bls12381.G1Affine] | OnCurveInstance[g2ElementWizard, sw_bls12381.G2Affine] |
		NonGroupMembershipInstance[g1ElementWizard, sw_bls12381.G1Affine] | NonGroupMembershipInstance[g2ElementWizard, sw_bls12381.G2Affine]
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
