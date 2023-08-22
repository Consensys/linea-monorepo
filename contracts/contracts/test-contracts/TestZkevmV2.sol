// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { ZkEvmV2, TransactionDecoder, CodecV2 } from "../ZkEvmV2.sol";

contract TestZkEvmV2 is ZkEvmV2 {
  function addL1L2MessageHash(bytes calldata _rlpTx) external {
    bytes memory data = TransactionDecoder.decodeTransaction(_rlpTx);

    bytes32[] memory l1L2MessageHashes = CodecV2._extractXDomainAddHashes(data);

    for (uint256 i; i < l1L2MessageHashes.length; i++) {
      _addL1L2MessageHash(l1L2MessageHashes[i]);
    }
  }

  function extractMessageHashes(bytes calldata _rlpTx) external pure returns (bytes32[] memory) {
    bytes memory data = TransactionDecoder.decodeTransaction(_rlpTx);

    bytes32[] memory l1L2MessageHashes = CodecV2._extractXDomainAddHashes(data);

    return l1L2MessageHashes;
  }

  function processBlockTransactions(
    bytes[] calldata _transactions,
    uint16[] calldata _batchReceptionIndices
  ) external returns (bytes32 hashOfTxHashes) {
    return _processBlockTransactions(_transactions, _batchReceptionIndices);
  }

  function processMessageHashes(bytes32[] calldata _logs) external returns (bytes32 hashOfTxHashes) {
    return _processMessageHashes(_logs);
  }
}
