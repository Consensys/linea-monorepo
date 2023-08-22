package ethereum

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// commonly used as a dummy "to" address
	TEST_ADDRESS   = common.HexToAddress("0xfeeddeadbeeffeeddeadbeeffeeddead01245678")
	TEST_ADDRESS_A = common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	TEST_HASH_F    = common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	TEST_HASH_A    = common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	CHAIN_ID       = big.NewInt(51)
)

// Test TxData
var testTxDatas = []types.TxData{
	// legacy transaction with dummy value
	&types.LegacyTx{
		Nonce:    2,
		GasPrice: big.NewInt(int64(123543135)),
		Gas:      4531112,
		To:       &TEST_ADDRESS,
		Value:    big.NewInt(int64(845315452)),
		Data:     common.Hex2Bytes("0xdeed8745a20f"),
	},
	// the same legacy transaction but we will
	// sign it with another signer.
	&types.LegacyTx{
		Nonce:    2,
		GasPrice: big.NewInt(int64(123543135)),
		Gas:      4531112,
		To:       &TEST_ADDRESS,
		Value:    big.NewInt(int64(845315452)),
		Data:     common.Hex2Bytes("0xdeed8745a20f"),
	},
	// access-list transaction
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
	// dynamic-fee transaction
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

// Test signers
var testSigners = []types.Signer{
	types.HomesteadSigner{},
	types.NewLondonSigner(CHAIN_ID),
	types.NewLondonSigner(CHAIN_ID),
	types.NewLondonSigner(CHAIN_ID),
}

// Test the consistency with ethereum signatures for unprotected legacy tx.
func TestTransactionSigning(t *testing.T) {

	os.Setenv("LAYER2_CHAIN_ID", CHAIN_ID.String())

	for i := range testSigners {
		// Fetch the right signer from the test vectors
		txData := testTxDatas[i]
		signer := testSigners[i]

		tx := types.NewTx(txData)

		// Sign the transaction
		txHash := signer.Hash(tx)
		privKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
		require.NoErrorf(t, err, "for signature #%v", i)
		sig, err := crypto.Sign(txHash[:], privKey)
		require.NoErrorf(t, err, "for signature #%v", i)
		tx, err = tx.WithSignature(signer, sig)
		require.NoErrorf(t, err, "for signature #%v", i)
		from, err := signer.Sender(tx)
		require.NoErrorf(t, err, "for signature #%v", i)

		// Check the transaction hash and that the v value is correct
		jsonSig := GetJsonSignature(tx)
		rlp := EncodeTxForSigning(tx)
		assert.True(t, jsonSig.V == "0x1b" || jsonSig.V == "0x1c", "V should be either 27 or 28 but was %v", jsonSig.V)
		assert.Equal(t, txHash, crypto.Keccak256Hash(rlp), "Mismatch of the tx hash")

		recovered := ECRecover(rlp, jsonSig)
		assert.Equal(t, from.Hex(), recovered.Hex(), "Mismatch of the recovered address")

		// Simulates the decoding of the transaction
		decodedTx := DecodeTxForSigning(rlp)

		assert.Equal(t, tx.To(), decodedTx.To())
		assert.Equal(t, tx.Nonce(), decodedTx.Nonce())
		assert.Equal(t, tx.Data(), decodedTx.Data())
		assert.Equal(t, tx.Value(), decodedTx.Value())
		assert.Equal(t, tx.Cost(), decodedTx.Cost())
	}

}
