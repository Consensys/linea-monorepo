package ethereum

import (
	"bytes"

	"github.com/consensys/zkevm-monorepo/prover/utils"
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
