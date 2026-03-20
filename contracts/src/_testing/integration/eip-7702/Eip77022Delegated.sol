// SPDX-License-Identifier: MIT
pragma solidity 0.8.33;

import { Eip7702TestNested } from "./Eip7702TestNested.sol";

/// @title Eip77022Delegated
/// @notice Test contract for EIP-7702 (Set Code) functionality with nested contract calls.
/// @author Consensys Software Inc.
/// @custom:security-contact security-report@linea.build
contract Eip77022Delegated {
  /// @notice Mapping storing a value for each address.
  mapping(address => uint256) public storedValue;

  /// @notice Emitted when a value is set for an address.
  /// @param sender The address that initiated the value change.
  /// @param value The new value set.
  event ValueSet(address indexed sender, uint256 value);

  /// @notice Sets a value for the caller and delegates a nested call to another contract.
  /// @param _newValue The value to store for the caller.
  /// @param _nestedContract The address of the nested contract to delegate to.
  function setValue(uint256 _newValue, address _nestedContract) external {
    storedValue[msg.sender] = _newValue;
    emit ValueSet(msg.sender, _newValue);
    Eip7702TestNested(_nestedContract).setValue{ value: 1000000000000000 }(_newValue);
  }

  /// @notice Retrieves the stored value for a given user.
  /// @param user The address to query.
  /// @return The stored value for the user.
  function getValue(address user) external view returns (uint256) {
    return storedValue[user];
  }
}
