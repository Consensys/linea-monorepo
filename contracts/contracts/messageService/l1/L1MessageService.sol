// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/security/ReentrancyGuardUpgradeable.sol";

import { IMessageService } from "../../interfaces/IMessageService.sol";
import { IGenericErrors } from "../../interfaces/IGenericErrors.sol";
import { PauseManager } from "../lib/PauseManager.sol";
import { RateLimiter } from "../lib/RateLimiter.sol";
import { L1MessageManager } from "./L1MessageManager.sol";

/**
 * @title Contract to manage cross-chain messaging on L1.
 * @author ConsenSys Software Inc.
 */
abstract contract L1MessageService is
  Initializable,
  RateLimiter,
  L1MessageManager,
  ReentrancyGuardUpgradeable,
  PauseManager,
  IMessageService,
  IGenericErrors
{
  // @dev This is initialised to save user cost with existing slot.
  uint256 public nextMessageNumber;

  address private _messageSender;

  // Keep free storage slots for future implementation updates to avoid storage collision.
  uint256[50] private __gap;

  // @dev adding these should not affect storage as they are constants and are store in bytecode
  uint256 private constant REFUND_OVERHEAD_IN_GAS = 42000;

  /**
   * @notice Initialises underlying message service dependencies.
   * @dev _messageSender is initialised to a non-zero value for gas efficiency on claiming.
   * @param _limitManagerAddress The address owning the rate limiting management role.
   * @param _pauseManagerAddress The address owning the pause management role.
   * @param _rateLimitPeriod The period to rate limit against.
   * @param _rateLimitAmount The limit allowed for withdrawing the period.
   **/
  function __MessageService_init(
    address _limitManagerAddress,
    address _pauseManagerAddress,
    uint256 _rateLimitPeriod,
    uint256 _rateLimitAmount
  ) internal onlyInitializing {
    if (_limitManagerAddress == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    if (_pauseManagerAddress == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    __ERC165_init();
    __Context_init();
    __AccessControl_init();
    __RateLimiter_init(_rateLimitPeriod, _rateLimitAmount);

    _grantRole(RATE_LIMIT_SETTER_ROLE, _limitManagerAddress);
    _grantRole(PAUSE_MANAGER_ROLE, _pauseManagerAddress);

    nextMessageNumber = 1;
    _messageSender = address(123456789);
  }

  /**
   * @notice Adds a message for sending cross-chain and emits MessageSent.
   * @dev The message number is preset (nextMessageNumber) and only incremented at the end if successful for the next caller.
   * @dev This function should be called with a msg.value = _value + _fee. The fee will be paid on the destination chain.
   * @param _to The address the message is intended for.
   * @param _fee The fee being paid for the message delivery.
   * @param _calldata The calldata to pass to the recipient.
   **/
  function sendMessage(
    address _to,
    uint256 _fee,
    bytes calldata _calldata
  ) external payable whenTypeNotPaused(L1_L2_PAUSE_TYPE) whenTypeNotPaused(GENERAL_PAUSE_TYPE) {
    if (_to == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    if (_fee > msg.value) {
      revert ValueSentTooLow();
    }

    uint256 messageNumber = nextMessageNumber;
    uint256 valueSent = msg.value - _fee;

    bytes32 messageHash = keccak256(abi.encode(msg.sender, _to, _fee, valueSent, messageNumber, _calldata));

    // @dev Status check and revert is in the message manager
    _addL1L2MessageHash(messageHash);

    nextMessageNumber++;

    emit MessageSent(msg.sender, _to, _fee, valueSent, messageNumber, _calldata, messageHash);
  }

  /**
   * @notice Claims and delivers a cross-chain message.
   * @dev _feeRecipient can be set to address(0) to receive as msg.sender.
   * @dev _messageSender is set temporarily when claiming and reset post. Used in sender().
   * @dev _messageSender is reset to address(123456789) to be more gas efficient.
   * @param _from The address of the original sender.
   * @param _to The address the message is intended for.
   * @param _fee The fee being paid for the message delivery.
   * @param _value The value to be transferred to the destination address.
   * @param _feeRecipient The recipient for the fee.
   * @param _calldata The calldata to pass to the recipient.
   * @param _nonce The unique auto generated nonce used when sending the message.
   **/
  function claimMessage(
    address _from,
    address _to,
    uint256 _fee,
    uint256 _value,
    address payable _feeRecipient,
    bytes calldata _calldata,
    uint256 _nonce
  ) external nonReentrant distributeFees(_fee, _to, _calldata, _feeRecipient) {
    _requireTypeNotPaused(L2_L1_PAUSE_TYPE);
    _requireTypeNotPaused(GENERAL_PAUSE_TYPE);

    bytes32 messageHash = keccak256(abi.encode(_from, _to, _fee, _value, _nonce, _calldata));

    // @dev Status check and revert is in the message manager.
    _updateL2L1MessageStatusToClaimed(messageHash);

    _addUsedAmount(_fee + _value);

    _messageSender = _from;

    (bool callSuccess, bytes memory returnData) = _to.call{ value: _value }(_calldata);
    if (!callSuccess) {
      if (returnData.length > 0) {
        assembly {
          let data_size := mload(returnData)
          revert(add(32, returnData), data_size)
        }
      } else {
        revert MessageSendingFailed(_to);
      }
    }

    _messageSender = address(123456789);

    emit MessageClaimed(messageHash);
  }

  /**
   * @notice Claims and delivers a cross-chain message.
   * @dev _messageSender is set temporarily when claiming.
   **/
  function sender() external view returns (address) {
    return _messageSender;
  }

  /**
   * @notice Function to receive funds for liquidity purposes.
   **/
  receive() external payable virtual {}

  /**
   * @notice The unspent fee is refunded if applicable.
   * @param _feeInWei The fee paid for delivery in Wei.
   * @param _to The recipient of the message and gas refund.
   * @param _calldata The calldata of the message.
   **/
  modifier distributeFees(
    uint256 _feeInWei,
    address _to,
    bytes calldata _calldata,
    address _feeRecipient
  ) {
    //pre-execution
    uint256 startingGas = gasleft();
    _;
    //post-execution

    // we have a fee
    if (_feeInWei > 0) {
      // default postman fee
      uint256 deliveryFee = _feeInWei;

      // do we have empty calldata?
      if (_calldata.length == 0) {
        bool isDestinationEOA;

        assembly {
          isDestinationEOA := iszero(extcodesize(_to))
        }

        // are we calling an EOA
        if (isDestinationEOA) {
          // initial + cost to call and refund minus gasleft
          deliveryFee = (startingGas + REFUND_OVERHEAD_IN_GAS - gasleft()) * tx.gasprice;

          if (_feeInWei > deliveryFee) {
            payable(_to).send(_feeInWei - deliveryFee);
          } else {
            deliveryFee = _feeInWei;
          }
        }
      }

      address feeReceiver = _feeRecipient == address(0) ? msg.sender : _feeRecipient;

      bool callSuccess = payable(feeReceiver).send(deliveryFee);
      if (!callSuccess) {
        revert FeePaymentFailed(feeReceiver);
      }
    }
  }
}
