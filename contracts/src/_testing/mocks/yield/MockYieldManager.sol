// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.0;

/// @dev Mock YieldManager contract for unit testing of LineaRollup
contract MockYieldManager {
  function receiveFundsFromReserve() external payable {}
  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external {}
}