package v0

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeBlockForCompression encodes a block for compression.
func EncodeBlockForCompression(block *types.Block, w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, block.Time()); err != nil {
		return err
	}
	for _, tx := range block.Transactions() {
		if err := EncodeTxForCompression(tx, w); err != nil {
			return err
		}
	}
	return nil
}

// EncodeTxForCompression encodes a transaction for compression.
// this code is from zk-evm-monorepo/prover/... but doesn't include the chainID
func EncodeTxForCompression(tx *types.Transaction, w io.Writer) error {
	switch {
	// LONDON with dynamic fees
	case tx.Type() == types.DynamicFeeTxType:
		if _, err := w.Write([]byte{tx.Type()}); err != nil {
			return err
		}
		if err := rlp.Encode(w, []interface{}{
			// tx.ChainID(),
			tx.Nonce(),
			tx.GasTipCap(),
			tx.GasFeeCap(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		}); err != nil {
			return err
		}
	// EIP2390 transaction with access-list
	case tx.Type() == types.AccessListTxType:
		if _, err := w.Write([]byte{tx.Type()}); err != nil {
			return err
		}
		if err := rlp.Encode(w, []interface{}{
			// tx.ChainID(),
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
		}); err != nil {
			return err
		}
	// EIP155 signature with protection against replay
	case tx.Type() == types.LegacyTxType && tx.Protected():
		if err := rlp.Encode(w, []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
			// tx.ChainID(), uint(0), uint(0),
		}); err != nil {
			return err
		}
	// Homestead signature
	case tx.Type() == types.LegacyTxType && !tx.Protected():
		if err := rlp.Encode(w, []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			ethereum.GetFrom(tx),
			tx.To(),
			tx.Value(),
			tx.Data(),
		}); err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("Unknown type of transaction %v, %++v", tx.Type(), tx))
	}

	return nil
}
