package invalidity_test

import (
	"math/rand/v2"
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
	invalidityPI "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// we dont need a cryptographic random number generator for testing
var rng = rand.New(rand.NewPCG(0, 0)) //nolint:gosec // G404: weak random is fine for tests

// Test data
var (
	// Raw byte values for consistency
	piStateRootHashBytes = [32]byte{0x01, 0x02, 0x03, 0x04}
	piTxHashBytes        = [32]byte{0x11, 0x22, 0x33, 0x44}
	piFromAddressBytes   = [20]byte{0xAA, 0xBB, 0xCC, 0xDD}
)

// fixedInputs are the inputs that are fixed for different test cases
var fixedInputs = invalidityPI.FixedInputs{
	StateRootHash: *new(field.Element).SetBytes(piStateRootHashBytes[:]),
	TxHashHi:      *new(field.Element).SetBytes(piTxHashBytes[:16]),
	TxHashLo:      *new(field.Element).SetBytes(piTxHashBytes[16:]),
	AddressHi:     *new(field.Element).SetBytes(piFromAddressBytes[0:4]),
	AddressLo:     *new(field.Element).SetBytes(piFromAddressBytes[4:]),
	FromAddress:   *new(field.Element).SetBytes(piFromAddressBytes[:]),
	ColSize:       16,
}

// testCase defines a test case
type testCase struct {
	name             string
	invalidityType   invalidity.InvalidityType
	hasBadPrecompile bool
	nbL2LogsVal      int
	mockArithCols    bool // if true, use the mockZkevmArithCols, otherwise use the  mockZkevmPI
}

var testCases = []testCase{
	{
		name:             "BadPrecompile_WithMockedAithColumns",
		invalidityType:   invalidity.BadPrecompile,
		hasBadPrecompile: true,
		nbL2LogsVal:      5,
		mockArithCols:    true,
	},
	{
		name:             "BadPrecompile_TooManyLogs_WithMockedArithColumns",
		invalidityType:   invalidity.BadPrecompile,
		hasBadPrecompile: true,
		nbL2LogsVal:      23,
		mockArithCols:    true,
	},
	{
		name:             "TooManyLogs_WithMockedPI",
		invalidityType:   invalidity.TooManyLogs,
		hasBadPrecompile: false,
		nbL2LogsVal:      20, // > MAX_L2_LOGS (16)
		mockArithCols:    false,
	},
}

// TestInvalidityCircuitWithInvalidityPI tests the invalidity circuit
// by mocking the arithmetization columns registered as public inputs
func TestBadPrecompileCircuit(t *testing.T) {
	var (
		comp  *wizard.CompiledIOP
		proof wizard.Proof
	)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// prepare zkevm inputs
			inputs := invalidityPI.Inputs{
				FixedInputs: fixedInputs,
				CaseInputs: invalidityPI.CaseInputs{
					HasBadPrecompile: tc.hasBadPrecompile,
					NumL2Logs:        tc.nbL2LogsVal,
				},
			}
			if tc.mockArithCols {
				comp, proof = invalidityPI.MockZkevmArithCols(inputs)
			} else {
				comp, proof = mockZkevmPI(inputs)
			}

			testZkevm := &zkevm.ZkEvm{WizardIOP: comp}

			// Verify the wizard proof first
			err := wizard.Verify(comp, proof)
			require.NoError(t, err, "wizard verification failed")

			// Prepare the circuit inputs
			assi := invalidity.AssigningInputs{
				InvalidityType:   tc.invalidityType,
				Zkevm:            testZkevm,
				ZkevmWizardProof: proof,
				FuncInputs: public_input.Invalidity{
					StateRootHash: types.Bytes32(piStateRootHashBytes),
					TxHash:        common.Hash(piTxHashBytes),
					FromAddress:   types.EthAddress(piFromAddressBytes),
				},
			}

			// Define the circuit
			circuit := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadPrecompileCircuit{
					InvalidityType: tc.invalidityType,
				},
			}

			// Allocate the circuit
			circuit.Allocate(invalidity.Config{
				Zkevm: testZkevm,
			})

			// Compile the circuit
			cs, err := frontend.Compile(
				ecc.BLS12_377.ScalarField(),
				scs.NewBuilder,
				&circuit,
			)
			require.NoError(t, err, "circuit compilation failed")

			// Assign the circuit
			assignment := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadPrecompileCircuit{
					InvalidityType: tc.invalidityType,
				},
			}
			assignment.Assign(assi)

			// Create witness
			witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
			require.NoError(t, err, "witness creation failed")

			// Verify the circuit is satisfied
			err = cs.IsSolved(witness)
			require.NoError(t, err, "circuit is not satisfied")
		})
	}
}

// mockZkevmPI creates a minimal ZkEvm with just the public inputs needed for BadPrecompileCircuit, it directly mocks the public inputs.
func mockZkevmPI(in invalidityPI.Inputs) (*wizard.CompiledIOP, wizard.Proof) {
	// Create a minimal wizard with just the public inputs needed for BadPrecompileCircuit
	define := func(b *wizard.Builder) {
		// Create proof columns of size 1 for public inputs
		stateRootHashCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("STATE_ROOT_HASH_COL"), 1)
		txHashHiCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("TX_HASH_HI_COL"), 1)
		txHashLoCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("TX_HASH_LO_COL"), 1)
		fromAddressCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("FROM_ADDRESS_COL"), 1)
		hasBadPrecompileCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), 1)
		nbL2LogsCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("NB_L2_LOGS_COL"), 1)

		// Register them as public inputs with the names expected by BadPrecompileCircuit
		b.CompiledIOP.InsertPublicInput("StateRootHash", accessors.NewFromPublicColumn(stateRootHashCol, 0))
		b.CompiledIOP.InsertPublicInput("TxHash_Hi", accessors.NewFromPublicColumn(txHashHiCol, 0))
		b.CompiledIOP.InsertPublicInput("TxHash_Lo", accessors.NewFromPublicColumn(txHashLoCol, 0))
		b.CompiledIOP.InsertPublicInput("FromAddress", accessors.NewFromPublicColumn(fromAddressCol, 0))
		b.CompiledIOP.InsertPublicInput("HasBadPrecompile", accessors.NewFromPublicColumn(hasBadPrecompileCol, 0))
		b.CompiledIOP.InsertPublicInput("NbL2Logs", accessors.NewFromPublicColumn(nbL2LogsCol, 0))
	}

	// Compile with dummy compiler for testing
	comp := wizard.Compile(define, dummy.Compile)

	// Create a proof by assigning values
	prove := func(run *wizard.ProverRuntime) {

		// if HashBadPrecompile is true, set the value to non-zerorandom value, otherwise set it to 0
		var hashBadPrecompileVal field.Element
		if in.CaseInputs.HasBadPrecompile {

			one := field.One()
			hashBadPrecompileVal = field.PseudoRand(rng)
			hashBadPrecompileVal.Add(&hashBadPrecompileVal, &one) // add 1 to avoid zero value
		} else {
			hashBadPrecompileVal = field.Zero()
		}

		run.AssignColumn(ifaces.ColID("STATE_ROOT_HASH_COL"), smartvectors.NewConstant(in.StateRootHash, 1))
		run.AssignColumn(ifaces.ColID("TX_HASH_HI_COL"), smartvectors.NewConstant(in.TxHashHi, 1))
		run.AssignColumn(ifaces.ColID("TX_HASH_LO_COL"), smartvectors.NewConstant(in.TxHashLo, 1))
		run.AssignColumn(ifaces.ColID("FROM_ADDRESS_COL"), smartvectors.NewConstant(in.FromAddress, 1))
		run.AssignColumn(ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), smartvectors.NewConstant(hashBadPrecompileVal, 1))
		run.AssignColumn(ifaces.ColID("NB_L2_LOGS_COL"), smartvectors.NewConstant(field.NewElement(uint64(in.NumL2Logs)), 1))
	}

	proof := wizard.Prove(comp, prove)

	return comp, proof
}
