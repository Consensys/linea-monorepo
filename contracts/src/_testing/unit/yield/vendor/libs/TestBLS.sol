// SPDX-License-Identifier: MIT
pragma solidity ^0.8.25;

import { BLS12_381 } from "../../../../../yield/libs/vendor/lido/BLS.sol";

contract TestBLS {
  function sha256Pair(bytes32 left, bytes32 right) external view returns (bytes32) {
    return BLS12_381.sha256Pair(left, right);
  }
}
