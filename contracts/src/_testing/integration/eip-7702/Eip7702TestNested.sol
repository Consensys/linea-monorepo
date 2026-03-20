// SPDX-License-Identifier: MIT
pragma solidity 0.8.33;

/// @title Eip7702TestNested
/// @notice Nested test contract that stores values delegated from the parent contract.
/// @author Consensys Software Inc.
/// @custom:security-contact security-report@linea.build
contract Eip7702TestNested {
  /// @notice Mapping storing a value for each address.
  mapping(address => uint256) public storedValue;

  /// @notice Emitted when a value is set for an address.
  /// @param sender The address that initiated the value change.
  /// @param value The new value set.
  event ValueSet(address indexed sender, uint256 value);

  /// @notice Sets a value for the caller with optional ETH transfer.
  /// @param newValue The value to store.
  function setValue(uint256 newValue) external payable {
    storedValue[msg.sender] = newValue;
    emit ValueSet(msg.sender, newValue);
  }

  /// @notice Retrieves the stored value for a given user.
  /// @param user The address to query.
  /// @return The stored value for the user.
  function getValue(address user) external view returns (uint256) {
    return storedValue[user];
  }
}
