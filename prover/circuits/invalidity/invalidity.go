package invalidity

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	wizardk "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	smtKoala "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	wizard "github.com/consensys/linea-monorepo/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type CircuitInvalidity struct {
	// The sub circuits for the invalidity cases:
	// - bad transaction nonce
	// - bad transaction value
	// ...
	SubCircuit SubCircuit `gnark:",secret"`
	// the functional public inputs of the circuit.
	FuncInputs FunctionalPublicInputsGnark `gnark:",secret"`
	// the hash of the functional public inputs
	PublicInput frontend.Variable `gnark:",public"`
}

// SubCircuit is the circuit for the invalidity case.
// It embeds frontend.Circuit (which provides Define) and adds FunctionalPIQGnark.
// Allocation and assignment are handled via concrete type methods.
type SubCircuit interface {
	frontend.Circuit
	FunctionalPIQGnark() FunctinalPIQGnark
}

// AssigningInputs collects the inputs used for the circuit assignment
type AssigningInputs struct {
	AccountTrieInputs AccountTrieInputs
	Transaction       *types.Transaction
	FuncInputs        public_input.Invalidity
	InvalidityType    InvalidityType
	FromAddress       common.Address
	RlpEncodedTx      []byte // the RLP encoded of the unsigned transaction
	KeccakCompiledIOP *wizardk.CompiledIOP
	KeccakProof       wizardk.Proof
	MaxRlpByteSize    int

	// inputs related to zkevm-wizard
	Zkevm            *zkevm.ZkEvm
	ZkevmWizardProof wizard.Proof

	StateRootHash linTypes.KoalaOctuplet // in case the Merkle proof is not provided. The state root hash is provided separately specially for the filtered address case.
}

// Define the constraints
func (c *CircuitInvalidity) Define(api frontend.API) error {
	// subCircuit constraints
	c.SubCircuit.Define(api)

	// Use the subcircuit's functional public inputs directly.
	// TxHash, FromAddress, StateRootHash, ToAddress come from the subcircuit via gnark:"-" fields;
	// TxNumber, ExpectedBlockNumber, FtxRollingHash are outer-circuit witness fields.
	c.FuncInputs.FunctinalPIQGnark = c.SubCircuit.FunctionalPIQGnark()

	//  constraint on the hashing of functional public inputs
	api.AssertIsEqual(c.PublicInput, c.FuncInputs.Sum(api))

	return nil
}

// Allocate the circuit
func (c *CircuitInvalidity) Allocate(config Config) {
	switch sub := c.SubCircuit.(type) {
	case *BadNonceBalanceCircuit:
		sub.Allocate(config)
	case *FilteredAddressCircuit:
		sub.Allocate(config)
	case *BadPrecompileCircuit:
		sub.Allocate(config)
	default:
		panic(fmt.Sprintf("unsupported subcircuit type: %T", c.SubCircuit))
	}
}

// Assign the circuit
func (c *CircuitInvalidity) Assign(assi AssigningInputs) {
	switch sub := c.SubCircuit.(type) {
	case *BadNonceBalanceCircuit:
		sub.Assign(assi)
	case *FilteredAddressCircuit:
		sub.Assign(assi)
	case *BadPrecompileCircuit:
		sub.Assign(assi)
	default:
		panic(fmt.Sprintf("unsupported subcircuit type: %T", c.SubCircuit))
	}
	// assign the Functional Public Inputs
	c.FuncInputs.Assign(assi.FuncInputs)
	// assign the public input
	c.PublicInput = assi.FuncInputs.Sum(nil)
}

// MakeProof and solve the circuit.
func (c *CircuitInvalidity) MakeProof(
	setup circuits.Setup,
	assi AssigningInputs,
) string {

	switch assi.InvalidityType {
	case BadNonce, BadBalance:
		c.SubCircuit = &BadNonceBalanceCircuit{}
	case BadPrecompile, TooManyLogs:
		c.SubCircuit = &BadPrecompileCircuit{}
	case FilteredAddressFrom, FilteredAddressTo:
		c.SubCircuit = &FilteredAddressCircuit{}
	default:
		panic("unsupported invalidity type")
	}

	c.Assign(assi)

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

// CheckOnly compiles the circuit, assigns it, and verifies the constraint
// system is satisfied without generating a real proof. Mirrors MakeProof but
// only checks constraint satisfaction. Used in partial mode.
func (c *CircuitInvalidity) CheckOnly(assi AssigningInputs) error {
	switch assi.InvalidityType {
	case BadNonce, BadBalance:
		c.SubCircuit = &BadNonceBalanceCircuit{}
	case BadPrecompile, TooManyLogs:
		c.SubCircuit = &BadPrecompileCircuit{}
	case FilteredAddressFrom, FilteredAddressTo:
		c.SubCircuit = &FilteredAddressCircuit{}
	default:
		return fmt.Errorf("unsupported invalidity type: %d", assi.InvalidityType)
	}

	c.Allocate(Config{
		KeccakCompiledIOP: assi.KeccakCompiledIOP,
		Depth:             smtKoala.DefaultDepth,
		MaxRlpByteSize:    assi.MaxRlpByteSize,
		Zkevm:             assi.Zkevm,
	})

	ccs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		c,
	)
	if err != nil {
		return fmt.Errorf("circuit compilation failed: %w", err)
	}

	assignment := CircuitInvalidity{}
	switch assi.InvalidityType {
	case BadNonce, BadBalance:
		assignment.SubCircuit = &BadNonceBalanceCircuit{}
	case BadPrecompile, TooManyLogs:
		assignment.SubCircuit = &BadPrecompileCircuit{}
	case FilteredAddressFrom, FilteredAddressTo:
		assignment.SubCircuit = &FilteredAddressCircuit{}
	}
	assignment.Assign(assi)

	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	if err != nil {
		return fmt.Errorf("witness creation failed: %w", err)
	}

	if err := ccs.IsSolved(witness); err != nil {
		return fmt.Errorf("circuit constraint check failed: %w", err)
	}

	return nil
}

// Config collects the data used for the sub circuits allocation
type Config struct {
	// depth of the merkle tree for the account trie
	Depth             int
	KeccakCompiledIOP *wizardk.CompiledIOP
	MaxRlpByteSize    int
	Zkevm             *zkevm.ZkEvm
}

type builder struct {
	config  Config
	circuit *CircuitInvalidity
	comp    *wizard.CompiledIOP
}

func NewBuilder(config Config) *builder {
	return &builder{config: config}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return makeCS(b.config, b.circuit, b.comp), nil
}

// compile  the circuit to the constraints
func makeCS(config Config, circuit *CircuitInvalidity, comp *wizard.CompiledIOP) constraint.ConstraintSystem {

	circuit.Allocate(config)

	scs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}

type InvalidityType uint8

const (
	BadNonce            InvalidityType = 0
	BadBalance          InvalidityType = 1
	BadPrecompile       InvalidityType = 2
	TooManyLogs         InvalidityType = 3
	FilteredAddressFrom InvalidityType = 4
	FilteredAddressTo   InvalidityType = 5
)

// String returns the string representation of the InvalidityType
func (t InvalidityType) String() string {
	switch t {
	case BadNonce:
		return "BadNonce"
	case BadBalance:
		return "BadBalance"
	case BadPrecompile:
		return "BadPrecompile"
	case TooManyLogs:
		return "TooManyLogs"
	case FilteredAddressFrom:
		return "FilteredAddressFrom"
	case FilteredAddressTo:
		return "FilteredAddressTo"
	default:
		return "Unknown"
	}
}

// MarshalJSON converts InvalidityType to its string representation for JSON
func (t InvalidityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON converts a JSON string to InvalidityType
func (t *InvalidityType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("InvalidityType must be a string: %w", err)
	}
	switch s {
	case "BadNonce":
		*t = BadNonce
	case "BadBalance":
		*t = BadBalance
	case "BadPrecompile":
		*t = BadPrecompile
	case "TooManyLogs":
		*t = TooManyLogs
	case "FilteredAddressFrom":
		*t = FilteredAddressFrom
	case "FilteredAddressTo":
		*t = FilteredAddressTo
	default:
		return fmt.Errorf("unknown InvalidityType: %s", s)
	}
	return nil
}
