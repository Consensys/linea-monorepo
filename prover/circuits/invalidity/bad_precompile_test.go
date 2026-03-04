package invalidity_test

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/maths/field"
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
	piStateRootHash    = field.RandomOctuplet()
	piTxHashBytes      = [32]byte{0x11, 0x22, 0x33, 0x44}
	piFromAddressBytes = [20]byte{0xAA, 0xBB, 0xCC, 0xDD}
)

// fixedInputs are the inputs that are fixed for different test cases
var fixedInputs = invalidityPI.FixedInputs{
	TxHashLimbs:    invalidity.CreateTxHashLimbs(piTxHashBytes),
	FromLimbs:      invalidity.CreateFromLimbs(piFromAddressBytes),
	StateRootLimbs: piStateRootHash,
	ColSize:        16,
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

			comp, proof = invalidity.MockZkevmPI(rng, inputs)

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
