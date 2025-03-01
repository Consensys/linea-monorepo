package v0

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeBlockForCompression encodes a block for compression.
func EncodeBlockForCompression(block *types.Block, w io.Writer, encodingOptions ...encode.Option) error {
	if err := binary.Write(w, binary.LittleEndian, block.Time()); err != nil {
		return err
	}
	for _, tx := range block.Transactions() {
		if err := EncodeTxForCompression(tx, w, encodingOptions...); err != nil {
			return err
		}
	}
	return nil
}

// EncodeTxForCompression encodes a transaction for compression.
// this code is from zk-evm-monorepo/prover/... but doesn't include the chainID
func EncodeTxForCompression(tx *types.Transaction, w io.Writer, encodingOptions ...encode.Option) error {
	cfg := encode.NewConfig()
	for _, o := range encodingOptions {
		o(&cfg)
	}
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
			cfg.GetAddress(tx),
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
			cfg.GetAddress(tx),
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
			cfg.GetAddress(tx),
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
			cfg.GetAddress(tx),
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

// DecodeBlockFromUncompressed inverts [EncodeBlockForCompression]. It is primarily meant for
// testing and ensuring the encoding is bijective.
func DecodeBlockFromUncompressed(r *bytes.Reader) (encode.DecodedBlockData, error) {
	var decTimestamp uint64

	if err := binary.Read(r, binary.LittleEndian, &decTimestamp); err != nil {
		return encode.DecodedBlockData{}, fmt.Errorf("could not decode timestamp: %w", err)
	}

	decodedBlk := encode.DecodedBlockData{
		Timestamp: decTimestamp,
	}

	for r.Len() != 0 {
		var (
			from common.Address
		)
		if tx, err := DecodeTxFromUncompressed(r, &from); err != nil {
			return encode.DecodedBlockData{}, fmt.Errorf("could not decode transaction #%v: %w", len(decodedBlk.Txs), err)
		} else {
			decodedBlk.Txs = append(decodedBlk.Txs, tx)
			decodedBlk.Froms = append(decodedBlk.Froms, from)
		}
	}

	return decodedBlk, nil
}

func ReadTxAsRlp(r *bytes.Reader) (fields []any, _type uint8, err error) {
	firstByte, err := r.ReadByte()
	if err != nil {
		err = fmt.Errorf("could not read the first byte: %w", err)
		return
	}

	// According to the RLP rule, `0xc0 + x` or `0xf7` indicates that the current
	// item is a list and this is what's used to identify that the transaction is
	// a legacy transaction or an EIP-155 transaction.
	//
	// Note that 0xc0 would indicate an empty list and thus be an invalid tx.
	if firstByte == types.AccessListTxType || firstByte == types.DynamicFeeTxType {
		_type = firstByte
	} else {
		if firstByte > 0xc0 {
			// Set the byte-reader backward so that we can apply the rlp-decoder
			// over it.
			if err = r.UnreadByte(); err != nil {
				return
			}
			_type = 0
		} else {
			err = fmt.Errorf("unexpected first byte: %x", firstByte)
			return
		}
	}

	err = rlp.Decode(r, &fields)
	return
}

// DecodeTxFromUncompressed puts all the transaction data into the output, except for the from address,
// which will be put where the argument "from" is referencing
func DecodeTxFromUncompressed(r *bytes.Reader, from *common.Address) (types.TxData, error) {
	fields, _type, err := ReadTxAsRlp(r)
	if err != nil {
		return nil, err
	}
	decoders := [3]func([]any, *common.Address) (types.TxData, error){
		decodeLegacyTx,
		decodeAccessListTx,
		decodeDynamicFeeTx,
	}
	return decoders[_type](fields, from)
}

func decodeLegacyTx(fields []any, from *common.Address) (types.TxData, error) {

	if len(fields) != 7 {
		return nil, fmt.Errorf("unexpected number of field")
	}

	tx := new(types.LegacyTx)
	err := errors.Join(
		ethereum.TryCast(&tx.Nonce, fields[0], "nonce"),
		ethereum.TryCast(&tx.GasPrice, fields[1], "gas-price"),
		ethereum.TryCast(&tx.Gas, fields[2], "gas"),
		ethereum.TryCast(from, fields[3], "from"),
		ethereum.TryCast(&tx.To, fields[4], "to"),
		ethereum.TryCast(&tx.Value, fields[5], "value"),
		ethereum.TryCast(&tx.Data, fields[6], "data"),
	)
	return tx, err
}

func decodeAccessListTx(fields []any, from *common.Address) (types.TxData, error) {

	if len(fields) != 8 {
		return nil, fmt.Errorf("invalid number of field for a dynamic transaction")
	}

	tx := new(types.AccessListTx)
	err := errors.Join(
		ethereum.TryCast(&tx.Nonce, fields[0], "nonce"),
		ethereum.TryCast(&tx.GasPrice, fields[1], "gas-price"),
		ethereum.TryCast(&tx.Gas, fields[2], "gas"),
		ethereum.TryCast(from, fields[3], "from"),
		ethereum.TryCast(&tx.To, fields[4], "to"),
		ethereum.TryCast(&tx.Value, fields[5], "value"),
		ethereum.TryCast(&tx.Data, fields[6], "data"),
		ethereum.TryCast(&tx.AccessList, fields[7], "access-list"),
	)

	return tx, err
}

func decodeDynamicFeeTx(fields []any, from *common.Address) (types.TxData, error) {

	if len(fields) != 9 {
		return nil, fmt.Errorf("invalid number of field for a dynamic transaction")
	}

	tx := new(types.DynamicFeeTx)
	err := errors.Join(
		ethereum.TryCast(&tx.Nonce, fields[0], "nonce"),
		ethereum.TryCast(&tx.GasTipCap, fields[1], "gas-tip-cap"),
		ethereum.TryCast(&tx.GasFeeCap, fields[2], "gas-fee-cap"),
		ethereum.TryCast(&tx.Gas, fields[3], "gas"),
		ethereum.TryCast(from, fields[4], "from"),
		ethereum.TryCast(&tx.To, fields[5], "to"),
		ethereum.TryCast(&tx.Value, fields[6], "value"),
		ethereum.TryCast(&tx.Data, fields[7], "data"),
		ethereum.TryCast(&tx.AccessList, fields[8], "access-list"),
	)

	return tx, err
}
