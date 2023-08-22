// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { CodecV2 } from "../messageService/lib/Codec.sol";
import { TransactionDecoder } from "../messageService/lib/TransactionDecoder.sol";

contract TestTransactionDecoder {
  using TransactionDecoder for *;

  function decodeTransactionAndHashes(bytes calldata _data) external pure returns (bytes32[] memory) {
    bytes memory transaction = TransactionDecoder.decodeTransaction(_data);
    return CodecV2._extractXDomainAddHashes(transaction);
  }
}
