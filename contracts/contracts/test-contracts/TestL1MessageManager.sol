// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { L1MessageManager } from "../messageService/l1/L1MessageManager.sol";

contract TestL1MessageManager is L1MessageManager {
  function addL2L1MessageHash(bytes32 _messageHash) external {
    _addL2L1MessageHash(_messageHash);
  }

  function updateL2L1MessageStatusToClaimed(bytes32 _messageHash) external {
    _updateL2L1MessageStatusToClaimed(_messageHash);
  }

  function addL1L2MessageHash(bytes32 _messageHash) external {
    _addL1L2MessageHash(_messageHash);
  }

  function updateL1L2MessageStatusToReceived(bytes32[] calldata _messageHashes) external {
    _updateL1L2MessageStatusToReceived(_messageHashes);
  }
}
