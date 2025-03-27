// SPDX-License-Identifier: Apache-2.0

/**
 * @title Library for RLP Encoding EIP-1559 transactions.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
pragma solidity ^0.8.28;

import { RlpWriter } from "./RlpWriter.sol";

library Eip1559RlpEncoder {
  /// @dev chainId is defined in the function and access list is not encoded.
  struct Eip1559Transaction {
    uint256 nonce;
    uint256 maxPriorityFeePerGas;
    uint256 maxFeePerGas;
    uint256 gasLimit;
    address to;
    uint256 value;
    bytes input;
    uint8 v;
    uint256 r;
    uint256 s;
  }

  function encodeEIP1559Tx(
    uint256 _chainId,
    Eip1559Transaction memory _transaction
  ) internal pure returns (bytes memory rlpEncodedTransaction, bytes32 transactionHash) {
    bytes[] memory fields = new bytes[](12);

    fields[0] = RlpWriter._encodeUint(_chainId);
    fields[1] = RlpWriter._encodeUint(_transaction.nonce);
    fields[2] = RlpWriter._encodeUint(_transaction.maxPriorityFeePerGas);
    fields[3] = RlpWriter._encodeUint(_transaction.maxFeePerGas);
    fields[4] = RlpWriter._encodeUint(_transaction.gasLimit);
    fields[5] = RlpWriter._encodeAddress(_transaction.to);
    fields[6] = RlpWriter._encodeUint(_transaction.value);
    fields[7] = RlpWriter._encodeBytes(_transaction.input);
    fields[8] = RlpWriter._encodeList(new bytes[](0)); // AccessList empty on purpose.
    fields[9] = RlpWriter._encodeUint(_transaction.v);
    fields[10] = RlpWriter._encodeUint(_transaction.r);
    fields[11] = RlpWriter._encodeUint(_transaction.s);

    bytes memory encodedList = RlpWriter._encodeList(fields);

    // Prepend type byte 0x02
    rlpEncodedTransaction = abi.encodePacked(hex"02", encodedList);
    transactionHash = keccak256(rlpEncodedTransaction);
  }
}
