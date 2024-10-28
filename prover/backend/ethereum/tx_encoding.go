package ethereum

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
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

const (
	// Number of rlp encoded field of the transaction
	dynFeeNumField        int = 9
	accessListTxNumField  int = 8
	legacyTxNumField      int = 9
	unprotectedTxNumField int = 6
)

// DecodeTxFromBytes from a string of bytes. If the stream of bytes is larger
// than the transaction then the remaining bytes are discarded and only the
// first bytes are used to decode the transaction. The function returns the
// transactions and the number of bytes read.
func DecodeTxFromBytes(b *bytes.Reader) (tx types.TxData, err error) {

	var (
		firstByte byte
	)

	if b.Len() == 0 {
		return nil, fmt.Errorf("empty buffer")
	}

	if firstByte, err = b.ReadByte(); err != nil {
		return nil, fmt.Errorf("could not read the first byte: %w", err)
	}

	switch {
	case firstByte == types.DynamicFeeTxType:
		return decodeDynamicFeeTx(b)
	case firstByte == types.AccessListTxType:
		return decodeAccessListTx(b)
	// According to the RLP rule, `0xc0 + x` or `0xf7` indicates that the current
	// item is a list and this is what's used to identify that the transaction is
	// a legacy transaction or a EIP-155 transaction.
	//
	// Note that 0xc0 would indicate an empty list and thus be an invalid tx.
	case firstByte > 0xc0:
		// Set the byte-reader backward so that we can apply the rlp-decoder
		// over it.
		b.UnreadByte()
		return decodeLegacyTx(b)
	default:
		return nil, fmt.Errorf("unexpected first byte: %x", firstByte)
	}
}

// decodeDynamicFeeTx encodes a [types.DynamicFeeTx] into a [bytes.Reader] and
// returns an error if it did not pass.
func decodeDynamicFeeTx(b *bytes.Reader) (parsedTx *types.DynamicFeeTx, err error) {
	decTx := []any{}

	if err = rlp.Decode(b, &decTx); err != nil {
		return nil, fmt.Errorf("could not rlp decode transaction: %w", err)
	}

	if len(decTx) != dynFeeNumField {
		return nil, fmt.Errorf("invalid number of field for a dynamic transaction")
	}

	parsedTx = new(types.DynamicFeeTx)

	err = errors.Join(
		TryCast(&parsedTx.ChainID, decTx[0], "chainID"),
		TryCast(&parsedTx.Nonce, decTx[1], "nonce"),
		TryCast(&parsedTx.GasTipCap, decTx[2], "gas-tip-cap"),
		TryCast(&parsedTx.GasFeeCap, decTx[3], "gas-fee-cap"),
		TryCast(&parsedTx.Gas, decTx[4], "gas"),
		TryCast(&parsedTx.To, decTx[5], "to"),
		TryCast(&parsedTx.Value, decTx[6], "value"),
		TryCast(&parsedTx.Data, decTx[7], "data"),
		TryCast(&parsedTx.AccessList, decTx[8], "access-list"),
	)

	return
}

// decodeAccessListTx decodes a [types.AccessListTx] from a [bytes.Reader]
// and returns an error if it did not pass.
func decodeAccessListTx(b *bytes.Reader) (parsedTx *types.AccessListTx, err error) {

	decTx := []any{}

	if err := rlp.Decode(b, &decTx); err != nil {
		return nil, fmt.Errorf("could not rlp decode transaction: %w", err)
	}

	if len(decTx) != accessListTxNumField {
		return nil, fmt.Errorf("invalid number of field for a dynamic transaction")
	}

	parsedTx = new(types.AccessListTx)
	err = errors.Join(
		TryCast(&parsedTx.ChainID, decTx[0], "chainID"),
		TryCast(&parsedTx.Nonce, decTx[1], "nonce"),
		TryCast(&parsedTx.GasPrice, decTx[2], "gas-price"),
		TryCast(&parsedTx.Gas, decTx[3], "gas"),
		TryCast(&parsedTx.To, decTx[4], "to"),
		TryCast(&parsedTx.Value, decTx[5], "value"),
		TryCast(&parsedTx.Data, decTx[6], "data"),
		TryCast(&parsedTx.AccessList, decTx[7], "access-list"),
	)

	return
}

// decodeLegacyTx decodes a [types.LegacyTx] from a [bytes.Reader] and returns
// an error if it did not pass.
//
// Note that when the transaction is an EIP-155 transaction, the chain-ID is
// not decoded although it could. The reason is that it is complicated to set
// it in the returned element as it "included" in the signature and we don't
// encode the signature.
func decodeLegacyTx(b *bytes.Reader) (parsedTx *types.LegacyTx, err error) {

	decTx := []any{}

	if err = rlp.Decode(b, &decTx); err != nil {
		return nil, fmt.Errorf("could not rlp decode transaction: %w", err)
	}

	if len(decTx) != legacyTxNumField && len(decTx) != unprotectedTxNumField {
		return nil, fmt.Errorf("unexpected number of field")
	}

	parsedTx = new(types.LegacyTx)
	err = errors.Join(
		TryCast(&parsedTx.Nonce, decTx[0], "nonce"),
		TryCast(&parsedTx.GasPrice, decTx[1], "gas-price"),
		TryCast(&parsedTx.Gas, decTx[2], "gas"),
		TryCast(&parsedTx.To, decTx[3], "to"),
		TryCast(&parsedTx.Value, decTx[4], "value"),
		TryCast(&parsedTx.Data, decTx[5], "data"),
	)

	return
}

// TryCast will attempt to set t with the underlying value of `from` will return
// an error if the type does not match. The explainer string is used to generate
// the error if any.
func TryCast[T any](into *T, from any, explainer string) error {

	if into == nil || from == nil {
		return fmt.Errorf("from or into is/are nil")
	}

	// The rlp encoding is not "type-aware", if the underlying field is an
	// access-list, it will decode into []interface{} (and we recursively parse
	// it) otherwise, it always decodes to `[]byte`
	if list, ok := (from).([]interface{}); ok {

		var (
			length = len(list)
			err    error
		)

		switch any(*into).(type) {

		case types.AccessList:
			accessList := make(types.AccessList, length)
			for i := range accessList {
				err = errors.Join(
					err,
					TryCast(&accessList[i], list[i], fmt.Sprintf("%v[%v]", explainer, i)),
				)
			}
			*into = (any(accessList)).(T)
			return err

		case types.AccessTuple:
			tuple := types.AccessTuple{}
			err = errors.Join(
				TryCast(&tuple.Address, list[0], fmt.Sprintf("%v.%v", explainer, "address")),
				TryCast(&tuple.StorageKeys, list[1], fmt.Sprintf("%v.%v", explainer, "storage-key")),
			)
			*into = (any(tuple)).(T)
			return err

		case []common.Hash:
			hashes := make([]common.Hash, length)
			for i := range hashes {
				TryCast(&hashes[i], list[i], fmt.Sprintf("%v[%v]", explainer, i))
			}
			*into = (any(hashes)).(T)
			return err

		default:
			return fmt.Errorf("unsupported type %T for %v", *into, explainer)
		}
	}

	var (
		intoAny   = any(*into)
		fromBytes = from.([]byte)
	)

	switch intoAny.(type) {
	case *common.Address:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as a hex string encoded address.
		address := common.BytesToAddress(fromBytes)
		*into = any(&address).(T)
	case common.Address:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as an hexstring encoded address.
		address := common.BytesToAddress(fromBytes)
		*into = any(address).(T)
	case common.Hash:
		// Parse the bytes as an UTF8 string (= direct casting in go).
		// Then, the string as a hexstring encoded address.
		hash := common.BytesToHash(fromBytes)
		*into = any(hash).(T)
	case *big.Int:
		var parsedBigInt big.Int
		parsedBigInt.SetBytes(fromBytes)
		*into = any(&parsedBigInt).(T)
	case uint64:
		// The encoding of uint64 can use less than 8 bytes. For this
		// reason we go through a big integer.
		var parsedBigInt big.Int
		parsedBigInt.SetBytes(fromBytes)
		*into = any(parsedBigInt.Uint64()).(T)
	case []byte:
		*into = any(fromBytes).(T)
	default:
		// Unsupported type - accumulate the error
		return fmt.Errorf("could not decode %v (value %s, type %T) as type %T", explainer, from, from, *into)
	}

	return nil
}
