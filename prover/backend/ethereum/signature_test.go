package ethereum

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
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

		recovered := ecRecover(rlp, jsonSig)
		assert.Equal(t, from.Hex(), recovered.Hex(), "Mismatch of the recovered address")

		// Simulates the decoding of the transaction
		decodedTx := decodeTxForSigning(rlp)

		assert.Equal(t, tx.To(), decodedTx.To())
		assert.Equal(t, tx.Nonce(), decodedTx.Nonce())
		assert.Equal(t, tx.Data(), decodedTx.Data())
		assert.Equal(t, tx.Value(), decodedTx.Value())
		assert.Equal(t, tx.Cost(), decodedTx.Cost())
	}

}

// ecRecovebackend/ethereum/signature.gor the signature from properly encoded transaction
func ecRecover(msg []byte, sig Signature) common.Address {

	// Ethereum signatures are signed using Keccak
	msgHash := crypto.Keccak256Hash(msg)
	pubKey, _, err := RecoverPublicKey(msgHash, sig)
	if err != nil {
		// Can happen if the signature is incorrect. Since we use this function
		// only for testing purpose, it is fine to panic.
		panic(err)
	}

	res := common.Address{}
	res.SetBytes(crypto.Keccak256(pubKey[:])[12:])
	return res
}

const (
	// Number of rlp encoded field of the transaction
	dynFeeNumField        int = 9
	accessListTxNumField  int = 8
	legacyTxNumField      int = 9
	unprotectedTxNumField int = 6
)

// decodeTxForSigning transaction that were encoded using `EncodeTxForSigning`. The function
// fails under error.
func decodeTxForSigning(b []byte) *types.Transaction {

	// Early check that we are not passed a
	if len(b) == 0 {
		panic("Empty byte string. Can't decode it into a transaction")
	}

	// Attempt decoding as a dynamic fee transaction
	decoded := []interface{}{}
	err := rlp.DecodeBytes(b[1:], &decoded)

	// Attempt decoding the transaction as a dynamic fee transaction
	if err == nil && b[0] == types.DynamicFeeTxType && len(decoded) == dynFeeNumField {
		tx := types.DynamicFeeTx{}
		errW := strings.Builder{}

		// Check the decoding field after field
		tryReadField(&tx.ChainID, &decoded, &errW, "chainID")
		tryReadField(&tx.Nonce, &decoded, &errW, "nonce")
		tryReadField(&tx.GasTipCap, &decoded, &errW, "gas-tip-cap")
		tryReadField(&tx.GasFeeCap, &decoded, &errW, "gas-fee-cap")
		tryReadField(&tx.Gas, &decoded, &errW, "gas")
		tryReadField(&tx.To, &decoded, &errW, "to")
		tryReadField(&tx.Value, &decoded, &errW, "value")
		tryReadField(&tx.Data, &decoded, &errW, "data")
		tryReadField(&tx.AccessList, &decoded, &errW, "access-list")

		errString := errW.String()

		if len(errString) == 0 {
			logrus.Tracef("Succesfully decoded transaction %++v", tx)
			// Successfully decoded the dynamic fee transaction
			return types.NewTx(&tx)
		}

		logrus.Debugf("Failed attempt to decode as a  dynamic transaction %v", errString)
	}

	// Attempt decoding the transaction as an access-list transaction
	if err == nil && b[0] == types.AccessListTxType && len(decoded) == accessListTxNumField {
		tx := types.AccessListTx{}
		errW := strings.Builder{}

		// Check the decoding field after field
		tryReadField(&tx.ChainID, &decoded, &errW, "chainID")
		tryReadField(&tx.Nonce, &decoded, &errW, "nonce")
		tryReadField(&tx.GasPrice, &decoded, &errW, "gas-price")
		tryReadField(&tx.Gas, &decoded, &errW, "gas")
		tryReadField(&tx.To, &decoded, &errW, "to")
		tryReadField(&tx.Value, &decoded, &errW, "value")
		tryReadField(&tx.Data, &decoded, &errW, "data")
		tryReadField(&tx.AccessList, &decoded, &errW, "access-list")

		errString := errW.String()

		if len(errString) == 0 {
			logrus.Tracef("Succesfully decoded access-list transaction %++v", tx)
			return types.NewTx(&tx)
		}

		logrus.Debugf("Failed attempt to decode as an access-list tx %v", errString)
	}

	logrus.Tracef("Attempting to decode transaction as a legacy transaction")

	// From this point, we assume this is an unprefixed rlp encoded legacy transaction
	err = rlp.DecodeBytes(b, &decoded)
	if err != nil {
		logrus.Panicf("Could not decode legacy transaction %v", err)
	}

	// Attempt decoding as a legacy transaction
	if len(decoded) != legacyTxNumField && len(decoded) != unprotectedTxNumField {
		logrus.Panicf(
			"Wrong size for a legacy transaction (%v) should be either %v or %v",
			len(decoded), legacyTxNumField, unprotectedTxNumField,
		)
	}

	tx := types.LegacyTx{}
	errW := strings.Builder{}

	// Attempts casting field after field
	// - There is no "ChainID" field for LegacyTransactions
	// - The two last fields are not used
	tryReadField(&tx.Nonce, &decoded, &errW, "nonce")
	tryReadField(&tx.GasPrice, &decoded, &errW, "gas-price")
	tryReadField(&tx.Gas, &decoded, &errW, "gas")
	tryReadField(&tx.To, &decoded, &errW, "to")
	tryReadField(&tx.Value, &decoded, &errW, "value")
	tryReadField(&tx.Data, &decoded, &errW, "data")

	if errW.Len() > 0 {
		logrus.Panicf("Failed deserializing as a legacy transaction : %v", errW.String())
	}

	logrus.Tracef("Succesfully decoded legacy transaction %++v", tx)
	return types.NewTx(&tx)
}

// Pop the head of `vals` attempt to cast it into `T`. If the casting succeeds, assign
// the casted value to `t`. Otherwise, write an error. if `val` is empty, it panics.
func tryReadField[T any](t *T, val *[]interface{}, w *strings.Builder, fieldName string) {

	// Sanity-check
	if len(*val) == 0 {
		panic("attempted to drain an empty slice")
	}

	// Drain the first value from the slice, so that the next call
	// reads the next value.
	defer func() { *val = (*val)[1:] }()

	// Runtime type checking without reflect
	var ifaceT interface{} = *t

	// The rlp encoding is not "type-aware", if the underlying field is an access-list,
	// it will decode into []interface{} (and we recursively parse it) otherwise, it
	// always decode to `[]byte`
	if list, ok := (*val)[0].([]interface{}); ok {

		length := len(list)
		var iface interface{}

		switch (ifaceT).(type) {
		case types.AccessList:
			accessList := make(types.AccessList, length)
			for i := range accessList {
				tryReadField(&accessList[i], &list, w, fmt.Sprintf("%v-%v", fieldName, i))
			}
			iface = accessList
		case types.AccessTuple:
			tuple := types.AccessTuple{}
			tryReadField(&tuple.Address, &list, w, fmt.Sprintf("%v-%v", fieldName, "tuple-address"))
			tryReadField(&tuple.StorageKeys, &list, w, fmt.Sprintf("%v-%v", fieldName, "tuple-storage-keys"))
			iface = tuple
		case []common.Hash:
			hashes := make([]common.Hash, length)
			for i := range hashes {
				tryReadField(&hashes[i], &list, w, fmt.Sprintf("%v-%v", fieldName, i))
			}
			iface = hashes
		default:
			logrus.Panicf("unsupported type %T for %v", ifaceT, fieldName)
		}

		*t = iface.(T)
		return
	}

	head := (*val)[0].([]byte)

	switch ifaceT.(type) {
	case *common.Address:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as an hexstring encoded address.
		address := common.BytesToAddress(head)
		var ifaceAddress interface{} = &address
		*t = ifaceAddress.(T)
	case common.Address:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as an hexstring encoded address.
		address := common.BytesToAddress(head)
		var ifaceAddress interface{} = address
		*t = ifaceAddress.(T)
	case common.Hash:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as an hexstring encoded address.
		hash := common.BytesToHash(head)
		var ifaceH interface{} = hash
		*t = ifaceH.(T)
	case *big.Int:
		var parsedBigInt big.Int
		parsedBigInt.SetBytes(head)
		*t = interface{}(&parsedBigInt).(T)
	case uint64:
		// The encoding of uint64 can use less than 8 bytes. For this
		// reason we go through a big integer.
		var parsedBigInt big.Int
		parsedBigInt.SetBytes(head)
		var intVal interface{} = parsedBigInt.Uint64()
		*t = intVal.(T)
	case []byte:
		*t = interface{}(head).(T)
	default:
		// Unsupported type - accumulate the error
		err := fmt.Errorf("could not decode %v (value %s, type %T) as type %T", fieldName, head, head, *t)
		if w.Len() > 0 {
			// separator between the error messages
			w.WriteString(", ")
		}
		w.WriteString(err.Error())
	}
}
