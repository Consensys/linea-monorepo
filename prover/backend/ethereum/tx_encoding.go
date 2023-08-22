package ethereum

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
)

const (
	// Number of rlp encoded field of the transaction
	dynFeeNumField        int = 9
	accessListTxNumField  int = 8
	legacyTxNumField      int = 9
	unprotectedTxNumField int = 6
)

// Returns the transaction hash of the transaction
func GetTxHash(tx *types.Transaction) common.Hash {
	if !tx.Protected() {
		// The normal signer does not return the right value
		return getUnprotectedSigner().Hash(tx)
	}
	return getSigner(tx.ChainId()).Hash(tx)
}

// Encode the transaction so that the hash of the encoded bytes produces exactly
// the transaction hash.
func EncodeTxForSigning(tx *types.Transaction) (encodedTx []byte) {

	// Since there are different types of transactions in Ethereum.
	// We encode them differently
	var buffer bytes.Buffer

	switch {
	// LONDON with dynamic fees
	case tx.Type() == types.DynamicFeeTxType:
		buffer.Write([]byte{tx.Type()})
		rlp.Encode(&buffer, []interface{}{
			tx.ChainId(),
			tx.Nonce(),
			tx.GasTipCap(),
			tx.GasFeeCap(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		})
	// EIP2390 transaction with access-list
	case tx.Type() == types.AccessListTxType:
		buffer.Write([]byte{tx.Type()})
		rlp.Encode(&buffer, []interface{}{
			tx.ChainId(),
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		})
	// EIP155 signature with protection against replay
	case tx.Type() == types.LegacyTxType && tx.Protected():
		rlp.Encode(&buffer, []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.ChainId(), uint(0), uint(0),
		})
	// Homestead signature
	case tx.Type() == types.LegacyTxType && !tx.Protected():
		rlp.Encode(&buffer, []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
		})
	default:
		utils.Panic("Unknown type of transaction %v, %++v", tx.Type(), tx)
	}

	return buffer.Bytes()
}

// Decode transaction that were encoded using `EncodeTxForSigning`. The function
// fails under error.
func DecodeTxForSigning(b []byte) *types.Transaction {

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
		TryReadField(&tx.ChainID, &decoded, &errW, "chainID")
		TryReadField(&tx.Nonce, &decoded, &errW, "nonce")
		TryReadField(&tx.GasTipCap, &decoded, &errW, "gas-tip-cap")
		TryReadField(&tx.GasFeeCap, &decoded, &errW, "gas-fee-cap")
		TryReadField(&tx.Gas, &decoded, &errW, "gas")
		TryReadField(&tx.To, &decoded, &errW, "to")
		TryReadField(&tx.Value, &decoded, &errW, "value")
		TryReadField(&tx.Data, &decoded, &errW, "data")
		TryReadField(&tx.AccessList, &decoded, &errW, "access-list")

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
		TryReadField(&tx.ChainID, &decoded, &errW, "chainID")
		TryReadField(&tx.Nonce, &decoded, &errW, "nonce")
		TryReadField(&tx.GasPrice, &decoded, &errW, "gas-price")
		TryReadField(&tx.Gas, &decoded, &errW, "gas")
		TryReadField(&tx.To, &decoded, &errW, "to")
		TryReadField(&tx.Value, &decoded, &errW, "value")
		TryReadField(&tx.Data, &decoded, &errW, "data")
		TryReadField(&tx.AccessList, &decoded, &errW, "access-list")

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
	TryReadField(&tx.Nonce, &decoded, &errW, "nonce")
	TryReadField(&tx.GasPrice, &decoded, &errW, "gas-price")
	TryReadField(&tx.Gas, &decoded, &errW, "gas")
	TryReadField(&tx.To, &decoded, &errW, "to")
	TryReadField(&tx.Value, &decoded, &errW, "value")
	TryReadField(&tx.Data, &decoded, &errW, "data")

	if errW.Len() > 0 {
		logrus.Panicf("Failed deserializing as a legacy transaction : %v", errW.String())
	}

	logrus.Tracef("Succesfully decoded legacy transaction %++v", tx)
	return types.NewTx(&tx)
}

// Pop the head of `vals` attempt to cast it into `T`. If the casting succeeds, assign
// the casted value to `t`. Otherwise, write an error. if `val` is empty, it panics.
func TryReadField[T any](t *T, val *[]interface{}, w *strings.Builder, fieldName string) {

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
				TryReadField(&accessList[i], &list, w, fmt.Sprintf("%v-%v", fieldName, i))
			}
			iface = accessList
		case types.AccessTuple:
			tuple := types.AccessTuple{}
			TryReadField(&tuple.Address, &list, w, fmt.Sprintf("%v-%v", fieldName, "tuple-address"))
			TryReadField(&tuple.StorageKeys, &list, w, fmt.Sprintf("%v-%v", fieldName, "tuple-storage-keys"))
			iface = tuple
		case []common.Hash:
			hashes := make([]common.Hash, length)
			for i := range hashes {
				TryReadField(&hashes[i], &list, w, fmt.Sprintf("%v-%v", fieldName, i))
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
