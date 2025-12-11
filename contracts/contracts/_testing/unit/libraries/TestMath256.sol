// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { Math256 } from "../../../lib/Math256.sol";

contract TestMath256 {
  function min(uint256 _a, uint256 _b) external pure returns (uint256 value) {
    return Math256.min(_a, _b);
  }

  function max(uint256 _a, uint256 _b) external pure returns (uint256 value) {
    return Math256.max(_a, _b);
  }

  function safeSub(uint256 _a, uint256 _b) external pure returns (uint256 value) {
    return Math256.safeSub(_a, _b);
  }

  function nextPow2(uint256 _x) external pure returns (uint256 value) {
    return Math256.nextPow2(_x);
  }

  function bitLength(uint256 _x) external pure returns (uint256 value) {
    return Math256.bitLength(_x);
  }
}
