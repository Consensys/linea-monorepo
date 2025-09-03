// SPDX-License-Identifier: MIT

pragma solidity 0.8.26;

import { StakeVault } from "../../src/StakeVault.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract StakeVaultHarness is StakeVault {
  constructor(IERC20 _stakingToken) StakeVault(_stakingToken) {}

  function isInitializing() public view returns (bool) {
    return _isInitializing();
  }

  function getInitializedVersion() public view returns (uint8) {
    return _getInitializedVersion();
  }
}

