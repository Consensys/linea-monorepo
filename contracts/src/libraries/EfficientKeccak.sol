// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.19;

/**
 * @title Contract to manage some efficient hashing functions.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library EfficientKeccak {
  /**
   * @notice Performs a gas optimized keccak hash for two bytes32 values.
   * @param _left Left value.
   * @param _right Right value.
   */
  function _efficientKeccak(bytes32 _left, bytes32 _right) internal pure returns (bytes32 value) {
    /// @solidity memory-safe-assembly
    assembly {
      mstore(0x00, _left)
      mstore(0x20, _right)
      value := keccak256(0x00, 0x40)
    }
  }

  /**
   * @notice Performs a gas optimized keccak hash for uint256 and address.
   * @param _left Left value.
   * @param _right Right value.
   */
  function _efficientKeccak(uint256 _left, address _right) internal pure returns (bytes32 value) {
    /// @solidity memory-safe-assembly
    assembly {
      mstore(0x00, _left)
      mstore(0x20, _right)
      value := keccak256(0x00, 0x40)
    }
  }

  /**
   * @notice Performs a gas optimized keccak hash for 5 words.
   * @param _v1 First value.
   * @param _v2 Second value.
   * @param _v3 Third value.
   * @param _v4 Fourth value.
   * @param _v5 Fifth value.
   */
  function _efficientKeccak(
    bytes32 _v1,
    bytes32 _v2,
    bytes32 _v3,
    bytes32 _v4,
    bytes32 _v5
  ) internal pure returns (bytes32 shnarf) {
    assembly {
      let mPtr := mload(0x40)
      mstore(mPtr, _v1)
      mstore(add(mPtr, 0x20), _v2)
      mstore(add(mPtr, 0x40), _v3)
      mstore(add(mPtr, 0x60), _v4)
      mstore(add(mPtr, 0x80), _v5)
      shnarf := keccak256(mPtr, 0xA0)
    }
  }
}
