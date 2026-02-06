package invalidity

import (
	"encoding/json"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

// SubCircuit is the circuit for the invalidity case
type SubCircuit interface {
	Define(frontend.API) error                           // define the constraints
	Allocate(Config)                                     //  allocate the circuit
	Assign(AssigningInputs)                              // generate assignment
	FunctionalPublicInputs() FunctionalPublicInputsGnark // returns the functional public inputs used in the subcircuit
}

// AssigningInputs collects the inputs used for the circuit assignment
type AssigningInputs struct {
	AccountTrieInputs AccountTrieInputs
	Transaction       *types.Transaction
	FuncInputs        public_input.Invalidity
	InvalidityType    InvalidityType
	FromAddress       common.Address
	RlpEncodedTx      []byte // the RLP encoded of the unsigned transaction
	KeccakCompiledIOP *wizard.CompiledIOP
	KeccakProof       wizard.Proof
	MaxRlpByteSize    int

	// inputs related to zkevm-wizard
	Zkevm            *zkevm.ZkEvm
	ZkevmWizardProof wizard.Proof
}

// Define the constraints
func (c *CircuitInvalidity) Define(api frontend.API) error {
	// subCircuit constraints
	c.SubCircuit.Define(api)

	// constraints on the consistence of functional public inputs
	// note that any FPI solely related to FtxFtxRollingHash is not checked here,
	// since they are used in the interconnection circuit, and not in the subcircuit.
	subCircuitFPI := c.SubCircuit.FunctionalPublicInputs()
	api.AssertIsEqual(
		api.Sub(c.FuncInputs.TxHash[0], subCircuitFPI.TxHash[0]),
		0,
	)
	api.AssertIsEqual(
		api.Sub(c.FuncInputs.TxHash[1], subCircuitFPI.TxHash[1]),
		0,
	)

	// it is failing here
	api.AssertIsEqual(
		api.Sub(c.FuncInputs.StateRootHash[0], subCircuitFPI.StateRootHash[0]),
		0,
	)
	api.AssertIsEqual(
		api.Sub(c.FuncInputs.StateRootHash[1], subCircuitFPI.StateRootHash[1]),
		0,
	)
	api.AssertIsEqual(
		api.Sub(c.FuncInputs.FromAddress, subCircuitFPI.FromAddress),
		0,
	)

	//  constraint on the hashing of functional public inputs
	api.AssertIsEqual(c.PublicInput, c.FuncInputs.Sum(api))

	return nil
}

// Allocate the circuit
func (c *CircuitInvalidity) Allocate(config Config) {
	// allocate the subCircuit
	c.SubCircuit.Allocate(config)
}

// Assign the circuit
func (c *CircuitInvalidity) Assign(assi AssigningInputs) {
	// assign the sub circuits
	c.SubCircuit.Assign(assi)
	// assign the Functional Public Inputs
	c.FuncInputs.Assign(assi.FuncInputs)
	// assign the public input
	c.PublicInput = assi.FuncInputs.Sum(nil)
}

// MakeProof and solve the circuit.
func (c *CircuitInvalidity) MakeProof(
	setup circuits.Setup,
	assi AssigningInputs,
	compilationSuite ...func(*wizard.CompiledIOP),
) string {

	switch assi.InvalidityType {
	case BadNonce, BadBalance:
		c.SubCircuit = &BadNonceBalanceCircuit{}
		assi.KeccakCompiledIOP, assi.KeccakProof = MakeKeccakProofs(assi.Transaction, assi.MaxRlpByteSize, compilationSuite...)
	case BadPrecompile, TooManyLogs:
		//c.SubCircuit = &BadPrecompileCircuit{}
		// zkevm wizard proof is already assigned
	case FilteredAddressFrom, FilteredAddressTo:
		panic(fmt.Sprintf("InvalidityType %s is not yet implemented", assi.InvalidityType))
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

	//	logrus.Infof("generated circuit proof `%++v` for input `%v`", proof, c.PublicInput.(*big.Int).String())

	// Write the serialized proof
	return circuits.SerializeProofRaw(proof)
}

// Config collects the data used for the sub circuits allocation
type Config struct {
	// depth of the merkle tree for the account trie
	Depth             int
	KeccakCompiledIOP *wizard.CompiledIOP
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

/*
// UpdateFtxRollingHash updates the ftxRollingHash
func UpdateFtxRollingHash(
	prevFtxRollingHash linTypes.KoalaOctuplet,
	txPayload *types.Transaction,
	expectedBlockHeight int,
	fromAddress linTypes.EthAddress,
) linTypes.KoalaOctuplet {
	signer := types.NewLondonSigner(txPayload.ChainId())
	txHash := signer.Hash(txPayload)

	hasher := mimc.NewMiMC()

	hasher.Write(prevFtxRollingHash[:])
	hasher.Write(txHash[:LIMB_SIZE])
	hasher.Write(txHash[LIMB_SIZE:])
	linTypes.WriteInt64On32Bytes(hasher, int64(expectedBlockHeight))
	hasher.Write(fromAddress[:])

	sum := hasher.Sum(nil)
	return [32]byte(sum)
}
*/
