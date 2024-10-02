// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

interface IXPProvider {
    function getTotalXP() external view returns (uint256);
    function getUserXP(address user) external view returns (uint256);
}
