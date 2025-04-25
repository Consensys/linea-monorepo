// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.28;

contract NoFundReceiver {
  fallback() external payable {
    // forced empty revert for code coverage
    revert();
  }

  receive() external payable {
    revert();
  }
}
