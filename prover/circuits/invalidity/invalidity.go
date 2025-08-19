package invalidity

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type CircuitInvalidity struct {
	// The sub circuits for the invalidity cases:
	// - bad transaction nonce
	// - bad transaction value
	// ...
	SubCircuit SubCircuit
	// the functional public inputs of the circuit.
	FuncInputs FunctionalPublicInputsGnark
	// the hash of the functional public inputs
	PublicInput frontend.Variable
}

// SubCircuit is the circuit for the invalidity case
type SubCircuit interface {
	Define(frontend.API) error       // define the constraints
	Allocate(Config)                 //  allocate the circuit
	Assign(AssigningInputs)          // generate assignment
	ExecutionCtx() frontend.Variable // returns the execution context (FinalStateRootHash) used in the subcircuit
}

// AssigningInputs collects the inputs used for the circuit assignment
type AssigningInputs struct {
	AccountTrieInputs AccountTrieInputs
	Transaction       *types.Transaction
	FuncInputs        public_input.Invalidity
	InvalidityType    InvalidityType
}

// Define the constraints
func (c *CircuitInvalidity) Define(api frontend.API) error {
	c.SubCircuit.Define(api)
	api.AssertIsEqual(c.SubCircuit.ExecutionCtx(), c.FuncInputs.SateRootHash)
	// @azam constraint on the hashing of functional public inputs
	return nil
}

// Allocate the circuit
func (c *CircuitInvalidity) Allocate(config Config) {
	// allocate the subCircuit
	c.SubCircuit.Allocate(config)
	// @azam: allocate the Functional Public Inputs
}

// Assign the circuit
func (c *CircuitInvalidity) Assign(assi AssigningInputs) {
	// assign the sub circuits
	c.SubCircuit.Assign(assi)
	// assign the Functional Public Inputs
	c.FuncInputs.Assign(assi.FuncInputs)
	// assign the public input
	c.PublicInput = assi.FuncInputs.Sum(nil)
	c.FuncInputs.ExpectedBlockNumber = assi.FuncInputs.ExpectedBlockHeight
	c.FuncInputs.SateRootHash = assi.FuncInputs.StateRootHash[:]
}

// MakeProof and solve the circuit.
func (c *CircuitInvalidity) MakeProof(setup circuits.Setup, assi AssigningInputs, FuncInputs *public_input.Invalidity) string {

	switch assi.InvalidityType {
	case BadNonce:
		c.SubCircuit = &BadNonceCircuit{}
	case BadBalance:
		c.SubCircuit = &BadBalanceCircuit{}
	default:
		panic("unsupported invalidity type")
	}

	c.Assign(assi)

	//@azam what options should I add?
	proof, err := circuits.ProveCheck(
		&setup,
		c,
		emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
		emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
	)

	if err != nil {
		panic(err)
	}

	logrus.Infof("generated circuit proof `%++v` for input `%v`", proof, c.PublicInput.(*big.Int).String())

	// Write the serialized proof
	return circuits.SerializeProofRaw(proof)
}

// Config collects the data used for the sub circuits allocation
type Config struct {
	// depth of the merkle tree for the account trie
	Depth int
}

type builder struct {
	config  Config
	circuit *CircuitInvalidity
}

func NewBuilder(config Config) *builder {
	return &builder{config: config}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return makeCS(b.config, b.circuit), nil
}

// compile  the circuit to the constraints
func makeCS(config Config, circuit *CircuitInvalidity) constraint.ConstraintSystem {

	circuit.Allocate(config)

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}

type InvalidityType uint8

const (
	BadNonce   InvalidityType = 0
	BadBalance InvalidityType = 1
)
