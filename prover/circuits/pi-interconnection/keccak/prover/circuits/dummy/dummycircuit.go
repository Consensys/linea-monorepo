package dummy

import (
	"math/big"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	frbw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

// CircuitDummy, that is verified by only one public input. dummy.CircuitDummy can also
// be optionally given an ID (0 by default) so that a proof for circuit of ID
// x will be rejected by the verifier for the circuit of ID y.
type CircuitDummy struct {
	X frontend.Variable `gnark:",public"`

	// This relies on the fact that x^5 is a permutation in Fr of the bn254
	// curve.
	X5 frontend.Variable `gnark:",secret"`

	// Optional field, changes the circuit so that X5 = X^5 + ID. This
	// functionality allows generated different incompatible versions of the
	// dummy circuit.
	ID int `gnark:"-"`
}

func (c *CircuitDummy) Define(api frontend.API) error {
	x5 := api.Mul(c.X, c.X, c.X, c.X, c.X)
	committer, _ := api.(frontend.Committer)
	_, err := committer.Commit(c.X)
	if err != nil {
		panic(err)
	}
	api.AssertIsEqual(
		api.Add(x5, c.ID),
		c.X5,
	)
	return nil
}

// Generates a deterministic (and unsafe) setup. Take the circuit ID as an
// input to "specialize" the circuit. The circID parameter is only meaningful
// to generate test samples for the smart-contract.

// Generates an assignment for the circuit.
func Assign(id circuits.MockCircuitID, x_ any) *CircuitDummy {
	switch x := x_.(type) {
	case fr254.Element:
		var x5, c fr254.Element
		x5.Exp(x, big.NewInt(5))
		c.SetInt64(int64(id))
		x5.Add(&x5, &c)
		return &CircuitDummy{X: x, X5: x5, ID: int(id)}
	case fr377.Element:
		var x5, c fr377.Element
		x5.Exp(x, big.NewInt(5))
		c.SetInt64(int64(id))
		x5.Add(&x5, &c)
		return &CircuitDummy{X: x, X5: x5, ID: int(id)}
	case frbw6.Element:
		var x5, c frbw6.Element
		x5.Exp(x, big.NewInt(5))
		c.SetInt64(int64(id))
		x5.Add(&x5, &c)
		return &CircuitDummy{X: x, X5: x5, ID: int(id)}
	default:
		utils.Panic("unsupported element: %T", x_)
	}

	// @alex: unreachable
	return nil
}
