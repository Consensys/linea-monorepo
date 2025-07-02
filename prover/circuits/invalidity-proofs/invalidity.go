package badnonce

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/core/types"
)

type CircuitInvadity struct {
	// The sub circuit for the invalidity case:
	// - bad transaction nonce
	// - bad transaction value
	// ...
	subCircuit SubCircuit
	// the functional public inputs of the circuit.
	FuncInputs FunctionalPublicInputsGnark
	// the hash of the functional public inputs
	PublicInput frontend.Variable
}

type SubCircuit interface {
	Allocate(Config) //  allocate the circuit
	// Compile()              // compile the circuit
	// Assign()               // generate assignment
	MakeProof(circuits.Setup, AssignInputs, public_input.Invalidity) string // set the witness and solve the circuit
}

type FunctionalPublicInputsGnark struct {
	TxHashMSB            frontend.Variable
	TxHashLSB            frontend.Variable
	FromAddress          frontend.Variable
	BlockHeight          frontend.Variable
	InitialStateRootHash frontend.Variable
	TimeStamp            frontend.Variable
}

type AssignInputs struct {
	Tree        smt.Tree
	Pos         int
	Account     Account
	LeafOpening accumulator.LeafOpening
	Transaction types.Transaction
	FuncInputs  public_input.Invalidity
}

// @azam check for TxHash
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHashMSB = pi.TxHash[:16]
	gpi.TxHashLSB = pi.TxHash[16:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.BlockHeight = pi.BlockHeight
	gpi.InitialStateRootHash = pi.InitialStateRootHash[:]
	gpi.TimeStamp = pi.TimeStamp
}
func (c CircuitInvadity) Define(api frontend.API) error {
	return nil
}

// Config collects the data used for circuit allocation
type Config struct {
	Depth int
}
