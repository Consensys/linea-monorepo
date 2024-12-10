// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

contract GasLimitTestContract {
  uint256 internal _gasLimit;

  function getGasLimit() external returns (uint256 gasLimit) {
    return _gasLimit;
  }

  function setGasLimit() external {
    _gasLimit = block.gaslimit;
  }
}
