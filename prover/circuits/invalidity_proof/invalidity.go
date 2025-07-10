package invalidity_proof

import (
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type CircuitInvalidity struct {
	// The sub circuits for the invalidity cases:
	// - bad transaction nonce
	// - bad transaction value
	// ...
	SubCircuits SubCircuit
	// the functional public inputs of the circuit.
	FuncInputs FunctionalPublicInputsGnark
	// the hash of the functional public inputs
	PublicInput frontend.Variable
}

// SubCircuit is the circuit for the invalidity case
type SubCircuit interface {
	Define(frontend.API) error // define the constraints
	Allocate(Config)           //  allocate the circuit
	Assign(AssigningInputs)    // generate assignment
}

// AssigningInputs collects the inputs used for the circuit assignment
type AssigningInputs struct {
	AccountTrieInputs AccountTrieInputs
	Transaction       *types.Transaction
	FuncInputs        public_input.Invalidity
	// the address of the sender
	// gateway contract on L1 extract it via ecrevovery over the signature
	FromAddress EthAddress
}

// Define the constraints
func (c *CircuitInvalidity) Define(api frontend.API) error {
	c.SubCircuits.Define(api)
	// @azam constraint on the hashing of functional public inputs
	return nil
}

// Allocate the circuit
func (c *CircuitInvalidity) Allocate(config Config) {
	// allocate the subCircuit
	c.SubCircuits.Allocate(config)
	// @azam: allocate the Functional Public Inputs
}

// Assign the circuit
func (c *CircuitInvalidity) Assign(assi AssigningInputs) {
	// assign the sub circuits
	c.SubCircuits.Assign(assi)
	// assign the Functional Public Inputs
	c.FuncInputs.Assign(assi.FuncInputs)
	// assign the public input
	c.PublicInput = assi.FuncInputs.Sum(nil)

}

// MakeProof and solve the circuit.
func (c *CircuitInvalidity) MakeProof(setup circuits.Setup, assi AssigningInputs, FuncInputs public_input.Invalidity) string {
	c.Assign(assi)

	//@azam what options should I add?
	proof, err := circuits.ProveCheck(
		&setup,
		c,
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
