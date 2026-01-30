package invalidity_test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	backend "github.com/consensys/linea-monorepo/prover/backend/invalidity"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// TestAccountTrieInputs verifies that the AccountTrieInputs() function produces
// valid inputs for the BadNonceBalanceCircuit.
//
// This test creates a mock Shomei ReadNonZeroTrace, wraps it in a Request,
// extracts AccountTrieInputs, and verifies the circuit can be solved.
func TestAccountTrieInputs(t *testing.T) {
	const maxRlpByteSize = 1024
	config := statemanager.MIMC_CONFIG

	// Create test accounts and transactions
	testCases := []struct {
		name           string
		account        types.Account
		tx             ethtypes.DynamicFeeTx
		fromAddress    common.Address
		invalidityType invalidity.InvalidityType
	}{
		{
			name:    "BadBalance - zero balance account",
			account: backend.CreateMockEOAAccount(0, big.NewInt(0)),
			tx: ethtypes.DynamicFeeTx{
				ChainID:   big.NewInt(59144),
				Nonce:     1,             // account nonce + 1 (valid nonce)
				Value:     big.NewInt(1), // but insufficient balance
				Gas:       21000,
				GasFeeCap: big.NewInt(1000000000),
			},
			fromAddress:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
			invalidityType: invalidity.BadBalance,
		},
		{
			name:    "BadNonce - wrong nonce",
			account: backend.CreateMockEOAAccount(65, big.NewInt(1e18)),
			tx: ethtypes.DynamicFeeTx{
				ChainID:   big.NewInt(59144),
				Nonce:     100,           // wrong nonce (should be 66)
				Value:     big.NewInt(0), // valid balance
				Gas:       21000,
				GasFeeCap: big.NewInt(1000000000),
			},
			fromAddress:    common.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01"),
			invalidityType: invalidity.BadNonce,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a Shomei-style ReadNonZeroTrace using the shared mock
			decodedTrace := backend.CreateMockAccountMerkleProof(tc.fromAddress, tc.account)

			// Create the backend Request
			req := backend.Request{
				InvalidityType:     tc.invalidityType,
				AccountMerkleProof: decodedTrace,
			}

			// Call AccountTrieInputs()
			accountTrieInputs, extractedAddress, err := req.AccountTrieInputs()
			require.NoError(t, err)
			require.Equal(t, types.EthAddress(tc.fromAddress), extractedAddress)

			// Verify the extracted data matches our test data
			require.Equal(t, tc.account.Nonce, accountTrieInputs.Account.Nonce)
			require.Equal(t, tc.account.Balance.Cmp(accountTrieInputs.Account.Balance), 0)

			// Now test that the circuit can be solved with these inputs
			tx := ethtypes.NewTx(&tc.tx)
			rlpEncodedTx := ethereum.EncodeTxForSigning(tx)

			// Create AssigningInputs using the extracted AccountTrieInputs
			assi := invalidity.AssigningInputs{
				AccountTrieInputs: accountTrieInputs,
				Transaction:       tx,
				FromAddress:       tc.fromAddress,
				MaxRlpByteSize:    maxRlpByteSize,
				InvalidityType:    tc.invalidityType,
				RlpEncodedTx:      rlpEncodedTx,
				FuncInputs: public_input.Invalidity{
					StateRootHash: accountTrieInputs.Root,
					TxHash:        ethereum.GetTxHash(tx),
					FromAddress:   types.EthAddress(tc.fromAddress),
				},
			}

			// Generate keccak proof
			kcomp, kproof := invalidity.MakeKeccakProofs(tx, maxRlpByteSize, dummy.Compile)
			assi.KeccakCompiledIOP = kcomp
			assi.KeccakProof = kproof

			// Allocate circuit
			circuit := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadNonceBalanceCircuit{},
			}
			circuit.Allocate(invalidity.Config{
				Depth:             config.Depth,
				KeccakCompiledIOP: kcomp,
				MaxRlpByteSize:    maxRlpByteSize,
			})

			// Compile circuit
			ccs, err := frontend.Compile(
				ecc.BLS12_377.ScalarField(),
				scs.NewBuilder,
				&circuit,
			)
			require.NoError(t, err)

			// Assign circuit
			assignment := invalidity.CircuitInvalidity{
				SubCircuit: &invalidity.BadNonceBalanceCircuit{},
			}
			assignment.Assign(assi)

			// Create witness and verify
			witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
			require.NoError(t, err)

			err = ccs.IsSolved(witness)
			require.NoError(t, err, "circuit should be satisfied with AccountTrieInputs from Request")
		})
	}
}

// TestAccountTrieInputsInvalidTrace verifies that AccountTrieInputs returns an error
// for invalid trace types.
func TestAccountTrieInputsInvalidTrace(t *testing.T) {
	// Test with nil underlying
	req := backend.Request{
		InvalidityType:     invalidity.BadNonce,
		AccountMerkleProof: statemanager.DecodedTrace{},
	}

	_, _, err := req.AccountTrieInputs()
	require.Error(t, err, "should error on nil underlying")

	// Test with wrong trace type (ReadZero instead of ReadNonZero)
	req2 := backend.Request{
		InvalidityType: invalidity.BadNonce,
		AccountMerkleProof: statemanager.DecodedTrace{
			Type:       1, // ReadZero
			Location:   "0x",
			Underlying: statemanager.ReadZeroTraceWS{},
		},
	}

	_, _, err = req2.AccountTrieInputs()
	require.Error(t, err, "should error on wrong trace type")
}
