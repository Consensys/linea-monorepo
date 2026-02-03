package dummy

import (
	"context"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	frbw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark/frontend"
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
func MakeUnsafeSetup(srsProvider circuits.SRSProvider, circID circuits.MockCircuitID, scalarField *big.Int) (circuits.Setup, error) {
	scs, err := MakeCS(circID, scalarField)
	if err != nil {
		return circuits.Setup{}, err
	}
	return circuits.MakeSetup(context.TODO(), "TODO", scs, srsProvider, nil)
}

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

// Generates a dummy proof for circuit ID with public input X.
func MakeProof(setup *circuits.Setup, x any, circID circuits.MockCircuitID) string {
	assignment := Assign(circID, x)

	opts := []any{}

	// Craft the options to make proof: for the bl12 case, we use the native
	// verifier as we know this proof will be recursively composed.
	if setup.CurveID() == ecc.BLS12_377 {
		opts = []any{
			emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), setup.Circuit.Field()),
			emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), setup.Circuit.Field()),
		}
	}

	proof, err := circuits.ProveCheck(setup, assignment, opts...)
	if err != nil {
		utils.Panic("while calling plonkutil.ProveCheck: %s", err.Error())
	}

	// x is from
	var xString string
	switch x_ := x.(type) {
	case fr254.Element:
		xString = "FrBn254:" + x_.String()
	case fr377.Element:
		xString = "FrBls12-377:" + x_.String()
	case frbw6.Element:
		xString = "FrBw6-761:" + x_.String()
	}

	logrus.Infof("generated dummy-circuit proof `%++v` for public input `%v`", proof, xString)

	if setup.CurveID() == ecc.BN254 {
		// Write the serialized proof
		return circuits.SerializeProofSolidityBn254(proof)
	}

	return circuits.SerializeProofRaw(proof)
}
