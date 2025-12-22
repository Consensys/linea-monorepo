// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { MessageHashing } from "../messageService/lib/MessageHashing.sol";

contract TestMessageHashing {
  function hashMessage(
    address _from,
    address _to,
    uint256 _fee,
    uint256 _valueSent,
    uint256 _messageNumber,
    bytes calldata _calldata
  ) external pure returns (bytes32 messageHash) {
    return MessageHashing._hashMessage(_from, _to, _fee, _valueSent, _messageNumber, _calldata);
  }

  function hashMessageWithEmptyCalldata(
    address _from,
    address _to,
    uint256 _fee,
    uint256 _valueSent,
    uint256 _messageNumber
  ) external pure returns (bytes32 messageHash) {
    return MessageHashing._hashMessageWithEmptyCalldata(_from, _to, _fee, _valueSent, _messageNumber);
  }
}

