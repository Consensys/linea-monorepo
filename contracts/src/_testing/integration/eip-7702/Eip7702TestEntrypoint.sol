// SPDX-License-Identifier: MIT
pragma solidity 0.8.33;

import { Eip77022Delegated } from "./Eip77022Delegated.sol";

/// @title Eip7702TestEntrypoint
/// @notice Outer test contract that orchestrates calls through nested and delegated addresses via EIP-7702.
/// @author Consensys Software Inc.
/// @custom:security-contact security-report@linea.build
contract Eip7702TestEntrypoint {
  /// @notice Emitted when a value is set through delegation.
  /// @param sender The original sender triggering the delegation.
  /// @param value The value set.
  event ValueSet(address indexed sender, uint256 value);

  /// @notice Delegates a setValue call through an EIP-7702 delegated address to a nested contract.
  /// @param _newValue The value to store.
  /// @param _delegatingAddress The EOA address that has delegated code.
  /// @param _nestedContract The address of the nested contract to call.
  function setValue(uint256 _newValue, address _delegatingAddress, address _nestedContract) external {
    // Calls the EOA with delegated code to set the value in the nested contract.
    Eip77022Delegated(_delegatingAddress).setValue(_newValue, _nestedContract);
    emit ValueSet(msg.sender, _newValue);
  }
}
