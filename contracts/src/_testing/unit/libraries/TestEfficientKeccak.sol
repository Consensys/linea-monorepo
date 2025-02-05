// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.26;

import { EfficientKeccak } from "../../../libraries/EfficientKeccak.sol";

contract TestEfficientKeccak {
  function efficientKeccakLeftRight(bytes32 _left, bytes32 _right) external pure returns (bytes32 value) {
    return EfficientKeccak._efficientKeccak(_left, _right);
  }
  function efficientKeccak(
    bytes32 _v1,
    bytes32 _v2,
    bytes32 _v3,
    bytes32 _v4,
    bytes32 _v5
  ) external pure returns (bytes32 value) {
    return EfficientKeccak._efficientKeccak(_v1, _v2, _v3, _v4, _v5);
  }

  function efficientKeccakWithSomeUints(
    bytes32 _v1,
    uint256 _v2,
    uint256 _v3,
    bytes32 _v4,
    uint256 _v5
  ) external pure returns (bytes32 value) {
    return EfficientKeccak._efficientKeccak(_v1, bytes32(_v2), bytes32(_v3), _v4, bytes32(_v5));
  }
}
