// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { MessageServiceBase } from "../messaging/MessageServiceBase.sol";

contract L1GenericBridge is MessageServiceBase {
  // initialize etc

  uint256 private constant withdrawalDelay = 3 days; // can be made a config option.
  address public escrowAddress;

  enum BridingStatus {
    UNKNOWN,
    REQUESTED,
    COMPLETE
  }

  enum WithdrawalOption {
    WAIT,
    IMMEDIATE
  }

  mapping(bytes32 bridgingHash => BridingStatus status) public bridgingStatuses;

  // function here to set escrowAddress;

  // get next message number / nonce - this should come from the message service.
  uint256 nextMessageNumber;

  // add security
  function bridgeEth(address _to) external payable {
    messageService.sendMessage(
      remoteSender,
      0,
      abi.encodeCall(L2GenericBridge.receiveDepositedEth, (_to, msg.sender, msg.value, nextMessageNumber))
    );

    // Custom behaviour here
    // emit events etc
  }

  // add security
  function receiveDepositedEth(
    address _to,
    address _from,
    uint256 _value,
    uint256 _nonce,
    WithdrawalOption _option
  ) external onlyMessagingService onlyAuthorizedRemoteSender {
    // write optimized hashing here

    uint256 minWithdrawTime = block.timestamp;

    if (_option == WithdrawalOption.WAIT) {
      unchecked {
        minWithdrawTime += withdrawalDelay;
      }
    }

    bytes32 bridgingHash = keccak256(abi.encode(_to, _from, _value, _nonce, _option, minWithdrawTime));

    if (bridgingStatuses[bridgingHash] != BridingStatus.UNKNOWN) {
      // reentry not allowed
      revert();
    }

    // Transfer to the Recipient or customize the process.

    if (_option == WithdrawalOption.WAIT) {
      bridgingStatuses[bridgingHash] = BridingStatus.REQUESTED;
      // emit requested
    } else {
      // call escrowAddress for immediate transfer.
      bridgingStatuses[bridgingHash] = BridingStatus.COMPLETE;
    }
  }

  // add security for only an allowed operator (maybe even just the escrow address)
  function completeWithdrawal(
    address _to,
    address _from,
    uint256 _value,
    uint256 _nonce,
    WithdrawalOption _option,
    uint256 _minWithdrawTime
  ) external payable {
    if (block.timestamp < _minWithdrawTime) {
      revert("too early");
    }

    if (_value != msg.value) {
      revert("value is wrong");
    }

    bytes32 bridgingHash = keccak256(abi.encode(_to, _from, _value, _nonce, _option, _minWithdrawTime));

    if (bridgingStatuses[bridgingHash] != BridingStatus.REQUESTED) {
      // reentry not allowed
      revert();
    }

    bridgingStatuses[bridgingHash] = BridingStatus.COMPLETE;

    (bool callSuccess, bytes memory returnData) = _to.call{ value: _value }("");
    if (!callSuccess) {
      if (returnData.length > 0) {
        assembly {
          let data_size := mload(returnData)
          revert(add(32, returnData), data_size)
        }
      } else {
        //revert MessageSendingFailed(_to);
      }
    }
  }

  // function: add ETH for user withdrawal (admin only)
}

contract L2GenericBridge is MessageServiceBase {
  // initialize etc

  // get next message number / nonce;
  uint256 nextMessageNumber;

  // add security
  function bridgeEth(address _to, L1GenericBridge.WithdrawalOption _option) external payable {
    messageService.sendMessage(
      remoteSender,
      0,
      abi.encodeCall(L1GenericBridge.receiveDepositedEth, (_to, msg.sender, msg.value, nextMessageNumber, _option))
    );
  }

  // This is called by the message service (onlyMessagingService) with a check to make sure it came from the L1 Bridge (onlyAuthorizedRemoteSender)
  // The message service checks reentry already
  function receiveDepositedEth(
    address _to,
    address _from,
    uint256 _value,
    uint256 _nonce
  ) external onlyMessagingService onlyAuthorizedRemoteSender {
    // Transfer to the Recipient or customize the process.
  }
}
