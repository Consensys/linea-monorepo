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
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
			name: "BadBalance - zero balance account",
			account: types.Account{
				Nonce:          0,
				Balance:        big.NewInt(0),
				StorageRoot:    types.Bytes32{},
				MimcCodeHash:   emptyCodeHash(),
				KeccakCodeHash: keccakEmptyCodeHash(),
				CodeSize:       0,
			},
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
			name: "BadNonce - wrong nonce",
			account: types.Account{
				Nonce:          65,
				Balance:        big.NewInt(1e18), // 1 ETH
				StorageRoot:    types.Bytes32{},
				MimcCodeHash:   emptyCodeHash(),
				KeccakCodeHash: keccakEmptyCodeHash(),
				CodeSize:       0,
			},
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
			// Create a Shomei-style ReadNonZeroTrace
			readTrace := createReadNonZeroTrace(t, tc.account, tc.fromAddress, config)

			// Create the backend Request
			req := backend.Request{
				InvalidityType:     tc.invalidityType,
				AccountMerkleProof: readTrace.decodedTrace,
			}

			// Call AccountTrieInputs()
			accountTrieInputs, extractedAddress, err := req.AccountTrieInputs()
			require.NoError(t, err)
			require.Equal(t, types.EthAddress(tc.fromAddress), extractedAddress)

			// Verify the extracted data matches our test data
			require.Equal(t, tc.account.Nonce, accountTrieInputs.Account.Nonce)
			require.Equal(t, tc.account.Balance.Cmp(accountTrieInputs.Account.Balance), 0)
			require.Equal(t, readTrace.root, accountTrieInputs.Root)
			require.Equal(t, readTrace.leaf, accountTrieInputs.Leaf)

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
					TxHash:        common.Hash(crypto.Keccak256(rlpEncodedTx)),
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

// testReadNonZeroTrace holds the test data for a ReadNonZeroTrace
type testReadNonZeroTrace struct {
	decodedTrace statemanager.DecodedTrace
	root         types.Bytes32
	leaf         types.Bytes32
}

// createReadNonZeroTrace creates a mock ReadNonZeroTraceWS with valid Merkle proof
func createReadNonZeroTrace(t *testing.T, account types.Account, address common.Address, config *smt.Config) testReadNonZeroTrace {
	// Compute HKey = MiMC(address)
	hKey := hashWithMiMC(address.Bytes())

	// Compute HVal = MiMC(account)
	hVal := hashAccount(account)

	// Create leaf opening
	leafOpening := accumulator.LeafOpening{
		Prev: 0,
		Next: 1,
		HKey: hKey,
		HVal: hVal,
	}

	// Compute leaf = MiMC(leafOpening)
	leaf := leafOpening.Hash(config)

	// Create a simple SMT with just this leaf
	tree := smt.NewEmptyTree(config)
	tree.Update(0, leaf)
	root := tree.Root

	// Get the Merkle proof
	proof, err := tree.Prove(0)
	require.NoError(t, err)

	// Verify the proof is valid
	valid := proof.Verify(config, leaf, root)
	require.True(t, valid, "Merkle proof should be valid")

	// Create the ReadNonZeroTrace
	readTrace := statemanager.ReadNonZeroTraceWS{
		Type:         0, // ReadNonZero
		Location:     "0x",
		NextFreeNode: 2,
		Key:          types.EthAddress(address),
		Value:        account,
		SubRoot:      root,
		LeafOpening:  leafOpening,
		Proof:        proof,
	}

	// Wrap in DecodedTrace
	decodedTrace := statemanager.DecodedTrace{
		Type:       0,
		Location:   "0x",
		Underlying: readTrace,
	}

	return testReadNonZeroTrace{
		decodedTrace: decodedTrace,
		root:         root,
		leaf:         leaf,
	}
}

// hashWithMiMC computes MiMC hash of input bytes
func hashWithMiMC(data []byte) types.Bytes32 {
	hasher := mimc.NewMiMC()
	hasher.Write(data)
	return types.Bytes32(hasher.Sum(nil))
}

// hashAccount computes MiMC hash of account (used for HVal)
func hashAccount(a types.Account) types.Bytes32 {
	hasher := mimc.NewMiMC()
	a.WriteTo(hasher)
	return types.Bytes32(hasher.Sum(nil))
}

// emptyCodeHash returns the MiMC hash for empty code
func emptyCodeHash() types.Bytes32 {
	// Empty code hash for EOA
	return types.Bytes32{}
}

// keccakEmptyCodeHash returns the Keccak256 hash of empty code (EOA)
func keccakEmptyCodeHash() types.FullBytes32 {
	// Keccak256 of empty bytes: 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
	return types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
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
