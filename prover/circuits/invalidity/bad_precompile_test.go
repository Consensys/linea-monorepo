package invalidity_test

import (
	"fmt"
	"math/big"
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
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// we dont need a cryptographic random number generator for testing
var rng = rand.New(rand.NewPCG(0, 0)) //nolint:gosec // G404: weak random is fine for tests

// Test data
var (
	// Raw byte values for consistency
	piStateRootHash    = field.RandomOctuplet()
	piTxHashBytes      = [32]byte{0x11, 0x22, 0x33, 0x44}
	piFromAddressBytes = [20]byte{0xAA, 0xBB, 0xCC, 0xDD}
)

// fixedInputs are the inputs that are fixed for different test cases
var fixedInputs = invalidityPI.FixedInputs{
	TxHashLimbs:    createTxHashLimbs(piTxHashBytes),
	FromLimbs:      createFromLimbs(piFromAddressBytes),
	StateRootLimbs: piStateRootHash,
	ColSize:        16,
}

// Helper functions to create limb arrays from byte arrays
func createTxHashLimbs(b [32]byte) [16]field.Element {
	var limbs [16]field.Element
	for i := 0; i < 16; i++ {
		limbs[i].SetBytes([]byte{b[i*2], b[i*2+1]})
	}
	return limbs
}

func createFromLimbs(b [20]byte) [10]field.Element {
	var limbs [10]field.Element
	for i := 0; i < 10; i++ {
		limbs[i].SetBytes([]byte{b[i*2], b[i*2+1]})
	}
	return limbs
}

// testCase defines a test case
type testCase struct {
	name             string
	invalidityType   invalidity.InvalidityType
	hasBadPrecompile bool
	nbL2LogsVal      int
}

var testCases = []testCase{
	{
		name:             "BadPrecompile_WithMockedAithColumns",
		invalidityType:   invalidity.BadPrecompile,
		hasBadPrecompile: true,
		nbL2LogsVal:      5,
	},
	{
		name:             "BadPrecompile_TooManyLogs_WithMockedArithColumns",
		invalidityType:   invalidity.BadPrecompile,
		hasBadPrecompile: true, // bad precompile
		nbL2LogsVal:      23,   // > MAX_L2_LOGS (16)
	},
	{
		name:             "TooManyLogs_WithMockedPI",
		invalidityType:   invalidity.TooManyLogs,
		hasBadPrecompile: false,
		nbL2LogsVal:      20, // > MAX_L2_LOGS (16)
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

			comp, proof = mockZkevmPI(inputs)

			testZkevm := &zkevm.ZkEvm{InitialCompiledIOP: comp}

			// Verify the wizard proof first
			err := wizard.Verify(comp, proof)
			require.NoError(t, err, "wizard verification failed")

			// Create a transaction with a To address
			txToAddr := common.HexToAddress("0xdeadbeefdeadb")
			tx := ethTypes.NewTx(&ethTypes.DynamicFeeTx{
				ChainID: big.NewInt(59144),
				To:      &txToAddr,
			})

			// Prepare the circuit inputs
			assi := invalidity.AssigningInputs{
				InvalidityType:   tc.invalidityType,
				Zkevm:            testZkevm,
				ZkevmWizardProof: proof,
				Transaction:      tx,
				FuncInputs: public_input.Invalidity{
					StateRootHash: piStateRootHash,
					TxHash:        common.Hash(piTxHashBytes),
					FromAddress:   types.EthAddress(piFromAddressBytes),
					ToAddress:     types.EthAddress(*tx.To()),
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

// mockZkevmPI creates a minimal ZkEvm with limb-based public inputs for BadPrecompileCircuit
func mockZkevmPI(in invalidityPI.Inputs) (*wizard.CompiledIOP, wizard.Proof) {
	define := func(b *wizard.Builder) {
		// Create proof columns for each limb (16 limbs for TxHash, 10 for FromAddress, 8 for StateRootHash)
		var (
			stateRootHashCols [8]ifaces.Column
			txHashCols        [16]ifaces.Column
			fromAddressCols   [10]ifaces.Column
		)

		// StateRootHash columns (8 KoalaBear elements)
		for i := 0; i < 8; i++ {
			stateRootHashCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("STATE_ROOT_HASH_%d", i), 1, true)
		}

		// TxHash columns (16 BE limbs)
		for i := 0; i < 16; i++ {
			txHashCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("TX_HASH_%d", i), 1, true)
		}

		// FromAddress columns (10 BE limbs)
		for i := 0; i < 10; i++ {
			fromAddressCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("FROM_ADDRESS_%d", i), 1, true)
		}

		// Scalar columns
		hasBadPrecompileCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), 1, true)
		nbL2LogsCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("NB_L2_LOGS_COL"), 1, true)

		// Register public inputs and build the extractor
		extractor := &invalidityPI.InvalidityPIExtractor{}

		// StateRootHash public inputs
		for i := 0; i < 8; i++ {
			extractor.StateRootHash[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("StateRootHash_BE_%d", i),
				accessors.NewFromPublicColumn(stateRootHashCols[i], 0),
			)
		}

		// TxHash public inputs
		for i := 0; i < 16; i++ {
			extractor.TxHash[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("TxHash_BE_%d", i),
				accessors.NewFromPublicColumn(txHashCols[i], 0),
			)
		}

		// FromAddress public inputs
		for i := 0; i < 10; i++ {
			extractor.FromAddress[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("From_BE_%d", i),
				accessors.NewFromPublicColumn(fromAddressCols[i], 0),
			)
		}

		// Scalar public inputs
		extractor.HasBadPrecompile = b.CompiledIOP.InsertPublicInput("HasBadPrecompile", accessors.NewFromPublicColumn(hasBadPrecompileCol, 0))
		extractor.NbL2Logs = b.CompiledIOP.InsertPublicInput("NbL2Logs", accessors.NewFromPublicColumn(nbL2LogsCol, 0))

		// Register the extractor in metadata
		b.CompiledIOP.ExtraData[invalidityPI.InvalidityPIExtractorMetadata] = extractor
	}

	comp := wizard.Compile(define, dummy.Compile)

	// Create a proof by assigning values
	prove := func(run *wizard.ProverRuntime) {

		// Create StateRootHash limbs
		for i := 0; i < 8; i++ {

			run.AssignColumn(ifaces.ColIDf("STATE_ROOT_HASH_%d", i), smartvectors.NewConstant(in.StateRootLimbs[i], 1))
		}

		// Create TxHash limbs in BE order (MSB first)
		for i := 0; i < 16; i++ {

			run.AssignColumn(ifaces.ColIDf("TX_HASH_%d", i), smartvectors.NewConstant(in.TxHashLimbs[i], 1))
		}

		// Create FromAddress limbs in BE order (MSB first)
		for i := 0; i < 10; i++ {

			run.AssignColumn(ifaces.ColIDf("FROM_ADDRESS_%d", i), smartvectors.NewConstant(in.FromLimbs[i], 1))
		}

		// Assign scalar values
		var hasBadPrecompileVal field.Element
		if in.CaseInputs.HasBadPrecompile {
			one := field.One()
			hasBadPrecompileVal = field.PseudoRand(rng)
			hasBadPrecompileVal.Add(&hasBadPrecompileVal, &one)
		} else {
			hasBadPrecompileVal = field.Zero()
		}

		run.AssignColumn(ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), smartvectors.NewConstant(hasBadPrecompileVal, 1))
		run.AssignColumn(ifaces.ColID("NB_L2_LOGS_COL"), smartvectors.NewConstant(field.NewElement(uint64(in.CaseInputs.NumL2Logs)), 1))
	}

	proof := wizard.Prove(comp, prove)

	return comp, proof
}
