// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title FallbackOperator interface for current functions, structs, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IFallbackOperator {
  /**
   * @notice Emitted when the fallback operator role is granted.
   * @param caller The address that called the function granting the role.
   * @param fallbackOperator The fallback operator address that received the operator role.
   */
  event FallbackOperatorRoleGranted(address indexed caller, address indexed fallbackOperator);

  /**
   * @notice Emitted when the fallback operator role is set on the contract.
   * @param caller The address that set the fallback operator address.
   * @param fallbackOperator The fallback operator address.
   */
  event FallbackOperatorAddressSet(address indexed caller, address indexed fallbackOperator);

  /**
   * @notice Sets the fallback operator role to the specified address if six months have passed since the last finalization.
   * @dev Reverts if six months have not passed since the last finalization.
   * @param _messageNumber Last finalized L1 message number as part of the feedback loop.
   * @param _rollingHash Last finalized L1 rolling hash as part of the feedback loop.
   * @param _lastFinalizedTimestamp Last finalized L2 block timestamp.
   */
  function setFallbackOperator(uint256 _messageNumber, bytes32 _rollingHash, uint256 _lastFinalizedTimestamp) external;
}