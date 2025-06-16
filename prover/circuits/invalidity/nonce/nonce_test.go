package nonce

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// Commonly used as a dummy "to" address
	TEST_ADDRESS   = common.HexToAddress("0xfeeddeadbeeffeeddeadbeeffeeddead01245678")
	TEST_ADDRESS_A = common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	TEST_HASH_F    = common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	TEST_HASH_A    = common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	CHAIN_ID       = big.NewInt(51)
)

// Shared test data
var testTxDatas = []types.TxData{
	// Legacy transaction
	&types.LegacyTx{
		Nonce:    2,
		GasPrice: big.NewInt(int64(123543135)),
		Gas:      4531112,
		To:       &TEST_ADDRESS,
		Value:    big.NewInt(int64(845315452)),
		Data:     common.Hex2Bytes("0xdeed8745a20f"),
	},
	// Legacy transaction with additional signer
	&types.LegacyTx{
		Nonce:    2,
		GasPrice: big.NewInt(int64(123543135)),
		Gas:      4531112,
		To:       &TEST_ADDRESS,
		Value:    big.NewInt(int64(845315452)),
		Data:     common.Hex2Bytes("0xdeed8745a20f"),
	},
	// Access list transaction
	&types.AccessListTx{
		ChainID:  CHAIN_ID,
		Nonce:    2,
		GasPrice: big.NewInt(int64(123543135)),
		Gas:      4531112,
		To:       &TEST_ADDRESS,
		Value:    big.NewInt(int64(845315452)),
		Data:     common.Hex2Bytes("0xdeed8745a20f"),
		AccessList: types.AccessList{
			types.AccessTuple{Address: TEST_ADDRESS_A, StorageKeys: []common.Hash{TEST_HASH_A, TEST_HASH_F}},
		},
	},
	// Dynamic fee transaction
	&types.DynamicFeeTx{
		ChainID:   CHAIN_ID,
		Nonce:     2,
		GasTipCap: big.NewInt(int64(123543135)),
		GasFeeCap: big.NewInt(int64(112121212)),
		Gas:       4531112,
		To:        &TEST_ADDRESS,
		Value:     big.NewInt(int64(845315452)),
		Data:      common.Hex2Bytes("0xdeed8745a20f"),
		AccessList: types.AccessList{
			types.AccessTuple{Address: TEST_ADDRESS_A, StorageKeys: []common.Hash{TEST_HASH_A, TEST_HASH_F}},
		},
	},
}

// Shared test signers
var testSigners = []types.Signer{
	types.HomesteadSigner{},
	types.NewEIP155Signer(CHAIN_ID),
	types.NewEIP2930Signer(CHAIN_ID),
	types.NewLondonSigner(CHAIN_ID),
}

// TestExtractNonceFromRLP tests the ExtractNonceFromRLP function using unsigned RLP-encoded transactions.
func TestExtractNonceFromRLP(t *testing.T) {
	expectedNonce := uint64(2)

	for i, txData := range testTxDatas {
		t.Run(fmt.Sprintf("Transaction-%d", i), func(t *testing.T) {
			// Create a transaction object
			tx := types.NewTx(txData)

			// Encode the transaction into RLP format
			rlpBytes := ethereum.EncodeTxForSigning(tx)
			fmt.Printf("Unsigned RLP Bytes for Transaction-%d: 0x%x\n", i, rlpBytes)

			// Call the ExtractNonceFromRLP function
			nonce, err := ExtractNonceFromRLP(rlpBytes)
			require.NoError(t, err, "Failed to extract nonce from RLP")

			// Assert that the extracted nonce matches the expected value
			assert.Equal(t, expectedNonce, nonce, "Extracted nonce does not match expected value")
		})
	}
}

// TestExtractNonceFromSignedRLP tests the ExtractNonceFromRLP function using signed RLP-encoded transactions.
func TestExtractNonceFromSignedRLP(t *testing.T) {
	expectedNonce := uint64(2)

	for i := range testSigners {
		t.Run(fmt.Sprintf("SignedTransaction-%d", i), func(t *testing.T) {
			// Fetch the transaction data and signer
			txData := testTxDatas[i]
			signer := testSigners[i]
			tx := types.NewTx(txData)

			// Sign the transaction
			txHash := signer.Hash(tx)
			privKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
			require.NoErrorf(t, err, "Failed to generate private key for transaction #%d", i)
			sig, err := crypto.Sign(txHash[:], privKey)
			require.NoErrorf(t, err, "Failed to sign transaction #%d", i)
			tx, err = tx.WithSignature(signer, sig)
			require.NoErrorf(t, err, "Failed to apply signature to transaction #%d", i)

			// Encode the signed transaction into RLP format
			// Encode the transaction into RLP format
			rlpBytes := ethereum.EncodeTxForSigning(tx)
			fmt.Printf("Signed RLP Bytes for Transaction-%d: 0x%x\n", i, rlpBytes)

			// Call the ExtractNonceFromRLP function
			nonce, err := ExtractNonceFromRLP(rlpBytes)
			require.NoErrorf(t, err, "Failed to extract nonce from signed RLP for transaction #%d", i)

			// Assert that the extracted nonce matches the expected value
			assert.Equal(t, expectedNonce, nonce, "Extracted nonce does not match expected value for transaction #%d", i)
		})
	}
}
