// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

interface IXPProvider {
    function getTotalXPContribution() external view returns (uint256);
    function getUserXPContribution(address user) external view returns (uint256);
}
