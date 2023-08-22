// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { RLPReader } from "./Rlp.sol";

using RLPReader for RLPReader.RLPItem;
using RLPReader for RLPReader.Iterator;
using RLPReader for bytes;

/*
 * dev Thrown when the transaction data length is too short.
 */
error TransactionShort();

/*
 * dev Thrown when the transaction type is unknown.
 */
error UnknownTransactionType();

/**
 * @title Contract to decode RLP formatted transactions.
 * @author ConsenSys Software Inc.
 */
library TransactionDecoder {
  /**
   * @notice Decodes the transaction extracting the calldata.
   * @param _transaction The RLP transaction.
   * @return data Returns the transaction calldata as bytes.
   */
  function decodeTransaction(bytes calldata _transaction) internal pure returns (bytes memory) {
    if (_transaction.length < 1) {
      revert TransactionShort();
    }

    bytes1 version = _transaction[0];

    if (version == 0x01) {
      return _decodeEIP2930Transaction(_transaction);
    }

    if (version == 0x02) {
      return _decodeEIP1559Transaction(_transaction);
    }

    if (version >= 0xc0) {
      return _decodeLegacyTransaction(_transaction);
    }

    revert UnknownTransactionType();
  }

  /**
   * @notice Decodes the EIP1559 transaction extracting the calldata.
   * @param _transaction The RLP transaction.
   * @return data Returns the transaction calldata as bytes.
   */
  function _decodeEIP1559Transaction(bytes calldata _transaction) private pure returns (bytes memory data) {
    bytes memory txData = _transaction[1:]; // skip the version byte

    RLPReader.RLPItem memory rlp = txData._toRlpItem();
    RLPReader.Iterator memory it = rlp._iterator();

    data = it._skipTo(8)._toBytes();
  }

  /**
   * @notice Decodes the EIP29230 transaction extracting the calldata.
   * @param _transaction The RLP transaction.
   * @return data Returns the transaction calldata as bytes.
   */
  function _decodeEIP2930Transaction(bytes calldata _transaction) private pure returns (bytes memory data) {
    bytes memory txData = _transaction[1:]; // skip the version byte

    RLPReader.RLPItem memory rlp = txData._toRlpItem();
    RLPReader.Iterator memory it = rlp._iterator();

    data = it._skipTo(7)._toBytes();
  }

  /**
   * @notice Decodes the legacy transaction extracting the calldata.
   * @param _transaction The RLP transaction.
   * @return data Returns the transaction calldata as bytes.
   */
  function _decodeLegacyTransaction(bytes calldata _transaction) private pure returns (bytes memory data) {
    bytes memory txData = _transaction;

    RLPReader.RLPItem memory rlp = txData._toRlpItem();
    RLPReader.Iterator memory it = rlp._iterator();

    data = it._skipTo(6)._toBytes();
  }
}
