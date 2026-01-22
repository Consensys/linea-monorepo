package invalidity_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// Test values - these must match between the wizard proof and the FuncInputs
var (
	// Raw byte values for consistency
	stateRootHashBytes = [32]byte{0x01, 0x02, 0x03, 0x04}
	txHashBytes        = [32]byte{0x11, 0x22, 0x33, 0x44}
	fromAddressBytes   = [20]byte{0xAA, 0xBB, 0xCC, 0xDD}

	// Field element representations (for wizard public inputs)
	stateRootHashVal = *new(field.Element).SetBytes(stateRootHashBytes[:])
	txHashHiVal      = *new(field.Element).SetBytes(txHashBytes[:16])
	txHashLoVal      = *new(field.Element).SetBytes(txHashBytes[16:])
	fromAddressVal   = *new(field.Element).SetBytes(fromAddressBytes[:])
)

// testCase defines the parameters for a single test
type testCase struct {
	name                 string
	invalidityType       invalidity.InvalidityType
	hashBadPrecompileVal field.Element
	nbL2LogsVal          field.Element
}

var testCases = []testCase{
	{
		name:                 "BadPrecompile",
		invalidityType:       invalidity.BadPrecompile,
		hashBadPrecompileVal: field.NewElement(10), // non-zero to pass AssertIsDifferent
		nbL2LogsVal:          field.NewElement(5),  // doesn't matter for this case
	},
	{
		name:                 "TooManyLogs",
		invalidityType:       invalidity.TooManyLogs,
		hashBadPrecompileVal: field.NewElement(0),  // doesn't matter for this case
		nbL2LogsVal:          field.NewElement(20), // > MAX_L2_LOGS (16) to pass AssertIsLessOrEqual
	},
}

func TestBadPrecompileCircuit(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testZkevm, proof := mockZkevm(tc.hashBadPrecompileVal, tc.nbL2LogsVal)

			// FuncInputs uses the native Go types (common.Hash, types.EthAddress, etc.)
			assi := invalidity.AssigningInputs{
				InvalidityType:   tc.invalidityType,
				Zkevm:            testZkevm,
				ZkevmWizardProof: proof,
				FuncInputs: public_input.Invalidity{
					StateRootHash: types.Bytes32(stateRootHashBytes),
					TxHash:        common.Hash(txHashBytes),
					FromAddress:   types.EthAddress(fromAddressBytes),
				},
			}

			// define the circuit
			circuit := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadPrecompileCircuit{
					InvalidityType: int(tc.invalidityType),
				},
			}

			// allocate the circuit - need to pass the zkevm in config
			circuit.Allocate(invalidity.Config{
				Zkevm: testZkevm,
			})

			// compile the circuit
			cs, err := frontend.Compile(
				ecc.BLS12_377.ScalarField(),
				scs.NewBuilder,
				&circuit,
			)
			require.NoError(t, err)

			// assign the circuit
			assignment := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadPrecompileCircuit{
					InvalidityType: int(tc.invalidityType),
				},
			}
			assignment.Assign(assi)

			witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
			require.NoError(t, err)

			err = cs.IsSolved(witness)
			require.NoError(t, err)
		})
	}
}

// mockZkevm creates a minimal ZkEvm with just the public inputs needed for BadPrecompileCircuit
func mockZkevm(hashBadPrecompileVal, nbL2LogsVal field.Element) (*zkevm.ZkEvm, wizard.Proof) {
	// Create a minimal wizard with just the public inputs needed for BadPrecompileCircuit
	define := func(b *wizard.Builder) {
		// Create proof columns of size 1 for public inputs
		stateRootHashCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("STATE_ROOT_HASH_COL"), 1)
		txHashHiCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("TX_HASH_HI_COL"), 1)
		txHashLoCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("TX_HASH_LO_COL"), 1)
		fromAddressCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("FROM_ADDRESS_COL"), 1)
		hashBadPrecompileCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), 1)
		nbL2LogsCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("NB_L2_LOGS_COL"), 1)

		// Register them as public inputs with the names expected by BadPrecompileCircuit
		b.CompiledIOP.InsertPublicInput("StateRootHash", accessors.NewFromPublicColumn(stateRootHashCol, 0))
		b.CompiledIOP.InsertPublicInput("TxHash_Hi", accessors.NewFromPublicColumn(txHashHiCol, 0))
		b.CompiledIOP.InsertPublicInput("TxHash_Lo", accessors.NewFromPublicColumn(txHashLoCol, 0))
		b.CompiledIOP.InsertPublicInput("FromAddress", accessors.NewFromPublicColumn(fromAddressCol, 0))
		b.CompiledIOP.InsertPublicInput("HashBadPrecompile", accessors.NewFromPublicColumn(hashBadPrecompileCol, 0))
		b.CompiledIOP.InsertPublicInput("NbL2Logs", accessors.NewFromPublicColumn(nbL2LogsCol, 0))
	}

	// Compile with dummy compiler for testing
	comp := wizard.Compile(define, dummy.Compile)

	// Create a proof by assigning values
	prove := func(run *wizard.ProverRuntime) {
		run.AssignColumn(ifaces.ColID("STATE_ROOT_HASH_COL"), smartvectors.NewConstant(stateRootHashVal, 1))
		run.AssignColumn(ifaces.ColID("TX_HASH_HI_COL"), smartvectors.NewConstant(txHashHiVal, 1))
		run.AssignColumn(ifaces.ColID("TX_HASH_LO_COL"), smartvectors.NewConstant(txHashLoVal, 1))
		run.AssignColumn(ifaces.ColID("FROM_ADDRESS_COL"), smartvectors.NewConstant(fromAddressVal, 1))
		run.AssignColumn(ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), smartvectors.NewConstant(hashBadPrecompileVal, 1))
		run.AssignColumn(ifaces.ColID("NB_L2_LOGS_COL"), smartvectors.NewConstant(nbL2LogsVal, 1))
	}

	proof := wizard.Prove(comp, prove)

	zkevm := &zkevm.ZkEvm{
		WizardIOP: comp,
	}
	return zkevm, proof
}
