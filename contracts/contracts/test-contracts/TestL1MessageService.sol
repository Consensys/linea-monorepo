// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { L1MessageService } from "../messageService/l1/L1MessageService.sol";

contract TestL1MessageService is L1MessageService {
  /**
   * @dev Thrown when the message has already been received.
   */
  error MessageAlreadyReceived(bytes32 messageHash);

  address public originalSender;
  bool private reentryDone;

  function initialize(
    address _limitManagerAddress,
    address _pauserManagerAddress,
    uint256 _rateLimitPeriod,
    uint256 _rateLimitAmount
  ) public initializer {
    _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
    __MessageService_init(_limitManagerAddress, _pauserManagerAddress, _rateLimitPeriod, _rateLimitAmount);
  }

  function tryInitialize(
    address _limitManagerAddress,
    address _pauserManagerAddress,
    uint256 _rateLimitPeriod,
    uint256 _rateLimitAmount
  ) external {
    __MessageService_init(_limitManagerAddress, _pauserManagerAddress, _rateLimitPeriod, _rateLimitAmount);
  }

  // @dev - the this. sendMessage is because the function is an "external" call and not wrapped
  function canSendMessage(address _to, uint256 _fee, bytes calldata _calldata) external payable {
    this.sendMessage{ value: msg.value }(_to, _fee, _calldata);
  }

  function addL2L1MessageHash(bytes32 _messageHash) external {
    if (inboxL2L1MessageStatus[_messageHash] != INBOX_STATUS_UNKNOWN) {
      revert MessageAlreadyReceived(_messageHash);
    }

    inboxL2L1MessageStatus[_messageHash] = INBOX_STATUS_RECEIVED;
  }

  function setSender() external payable {
    (bool success, bytes memory data) = msg.sender.call(abi.encodeWithSignature("sender()"));
    if (success) {
      originalSender = abi.decode(data, (address));
    }
  }

  function sendNewMessage() external payable {
    this.sendMessage{ value: 1 wei }(address(this), 1, "0x");
  }

  function doReentry() external payable {
    address originalAddress;

    (bool success, bytes memory data) = msg.sender.call(abi.encodeWithSignature("sender()"));
    if (success) {
      originalAddress = abi.decode(data, (address));
    }

    if (!reentryDone) {
      (bool succeeded, bytes memory dataInner) = msg.sender.call(
        abi.encodeWithSignature(
          "claimMessage(address,address,uint256,uint256,address,bytes,uint256)",
          originalAddress,
          originalAddress,
          0.05 ether,
          1 ether,
          address(0),
          abi.encodeWithSignature("doReentry()", 1)
        )
      );

      if (succeeded) {
        reentryDone = true;
      } else {
        if (dataInner.length > 0) {
          assembly {
            let data_size := mload(dataInner)
            revert(add(32, dataInner), data_size)
          }
        } else {
          revert("Function call reverted");
        }
      }
    }
  }

  function addFunds() external payable {}
}
