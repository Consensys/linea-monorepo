package v1

import (
	"fmt"
	"github.com/consensys/zkevm-monorepo/prover/backend/ethereum"
	"io"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeBlockForCompression encodes a block for compression.
func EncodeBlockForCompression(block *types.Block, w io.Writer) error {
	rlpBlock := make([]interface{}, 1+block.Transactions().Len())
	rlpBlock[0] = block.Time()
	for i, tx := range block.Transactions() {
		rlpBlock[i+1] = trimTxForCompression(tx)
	}
	return rlp.Encode(w, rlpBlock)
}

// trimTxForCompression encodes a transaction for compression.
// this code is from zk-evm-monorepo/prover/... but doesn't include the chainID
func trimTxForCompression(tx *types.Transaction) []interface{} {
	switch {
	// LONDON with dynamic fees
	case tx.Type() == types.DynamicFeeTxType:
		return []interface{}{
			tx.Type(),
			tx.Nonce(),
			tx.GasTipCap(),
			tx.GasFeeCap(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		}
	// EIP2390 transaction with access-list
	case tx.Type() == types.AccessListTxType:
		return []interface{}{
			tx.Type(),
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		}
	// EIP155 signature with protection against replay
	case tx.Type() == types.LegacyTxType && tx.Protected():
		return []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			// tx.ChainId(), uint(0), uint(0),
		}
	// Homestead signature
	case tx.Type() == types.LegacyTxType && !tx.Protected(): // TODO Merge this with the previous case?
		return []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
		}

	default:
		panic(fmt.Sprintf("Unknown type of transaction %v, %++v", tx.Type(), tx))
	}
}
