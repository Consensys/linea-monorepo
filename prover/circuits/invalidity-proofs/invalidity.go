package badnonce

import (
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

const (
	NBSubCircuit = 2
)

type CircuitInvalidity struct {
	// The sub circuit for the invalidity case:
	// - bad transaction nonce
	// - bad transaction value
	// ...
	subCircuits [NBSubCircuit]SubCircuit
	// the functional public inputs of the circuit.
	FuncInputs FunctionalPublicInputsGnark
	// the hash of the functional public inputs
	PublicInput frontend.Variable
}

type SubCircuit interface {
	Define(frontend.API) error
	Allocate(Config)        //  allocate the circuit
	Assign(AssigningInputs) // generate assignment
}

// AssigningInputs collects the inputs used for the circuit assignment
type AssigningInputs struct {
	Tree        *smt.Tree
	Pos         int
	Account     Account
	LeafOpening accumulator.LeafOpening
	Transaction *types.Transaction
	FuncInputs  public_input.Invalidity
}

func (c CircuitInvalidity) Define(api frontend.API) error {
	for i := range c.subCircuits {
		c.subCircuits[i].Define(api)
	}
	return nil
}

func (c CircuitInvalidity) Allocate(config Config) {
	// allocate the subCircuit
	for i := range c.subCircuits {
		c.subCircuits[i].Allocate(config)
	}
	// allocate the Functional Public Inputs
}

func (c CircuitInvalidity) Assign(assi AssigningInputs) CircuitInvalidity {
	// assign the sub circuits
	for i := range c.subCircuits {
		c.subCircuits[i].Assign(assi)
	}
	// assign the Functional Public Inputs
	c.FuncInputs.Assign(assi.FuncInputs)
	// assign the public input
	c.PublicInput = new(big.Int).SetBytes(assi.FuncInputs.Sum(nil))
	return c
}

func (c CircuitInvalidity) MakeProof(setup circuits.Setup, assi AssigningInputs, FuncInputs public_input.Invalidity) string {
	assignment := c.Assign(assi)

	//@azam what options should I add?
	proof, err := circuits.ProveCheck(
		&setup,
		&assignment,
	)

	if err != nil {
		panic(err)
	}

	logrus.Infof("generated circuit proof `%++v` for input `%v`", proof, assignment.PublicInput.(*big.Int).String())

	// Write the serialized proof
	return circuits.SerializeProofRaw(proof)
}

// Config collects the data used for choosing the subcircuit and its allocation
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
