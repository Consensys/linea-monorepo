// The bn254 package provides an implementation of a gnark circuit that can
// recursively verify a PLONK proof on the BW6 field with an arithmetization
// over the BN254 curve. This circuit is used as the final step before we
// submit a proof to Ethereum.
package emulation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bw6761"
	"github.com/consensys/gnark/std/math/emulated"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
)

// shorthand for the emulated types as this can get verbose very quickly with
// generics. `em` stands for emulated
type (
	emFr       = sw_bw6761.ScalarField
	emG1       = sw_bw6761.G1Affine
	emG2       = sw_bw6761.G2Affine
	emGT       = sw_bw6761.GTEl
	emProof    = emPlonk.Proof[emFr, emG1, emG2]
	emCircVkey = emPlonk.CircuitVerifyingKey[emFr, emG1]
	emBaseVKey = emPlonk.BaseVerifyingKey[emFr, emG1, emG2]
	emWitness  = emPlonk.Witness[emFr]
)

// The outer-circuits converts a proof over the BW6-761 field into a proof over
// the BN254 field.
type CircuitEmulation struct {
	CircuitVkeys []emCircVkey      `gnark:"-"`
	BaseVKey     emBaseVKey        `gnark:"-"`
	Proof        emProof           `gnark:",secret:"`
	Witness      emWitness         `gnark:",secret:"`
	PublicInput  frontend.Variable `gnark:",public"`
	CircuitID    frontend.Variable `gnark:",secret:"`
}

func (c *CircuitEmulation) Define(api frontend.API) error {

	verifier, err := emPlonk.NewVerifier[emFr, emG1, emG2, emGT](api)
	if err != nil {
		return fmt.Errorf("while instantiating the verifier: %w", err)
	}

	var (
		proofs    = []emProof{c.Proof}
		witnesses = []emWitness{c.Witness}
		switches  = []frontend.Variable{c.CircuitID}
	)

	err = verifier.AssertDifferentProofs(c.BaseVKey, c.CircuitVkeys, switches, proofs, witnesses, emPlonk.WithCompleteArithmetic())
	if err != nil {
		return fmt.Errorf("while asserting the proof are correct: %w", err)
	}

	// @alex: the intent here is to show that public input and witness
	// represent an equal number but not in the same representation.
	f, err := emulated.NewField[emFr](api)
	if err != nil {
		return err
	}

	var (
		piBits   = f.ToBits(&c.Witness.Public[0])
		piNative = api.FromBinary(piBits[:fr.Bits]...)
	)

	for i := fr.Bits; i < len(piBits); i++ {
		api.AssertIsEqual(piBits[i], 0)
	}

	api.AssertIsEqual(piNative, c.PublicInput)

	return nil
}
