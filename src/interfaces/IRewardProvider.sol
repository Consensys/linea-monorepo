// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

interface IRewardProvider {
    function totalRewardsSupply() external view returns (uint256);
    function rewardsBalanceOf(address account) external view returns (uint256);
    function rewardsBalanceOfAccount(address user) external view returns (uint256);
}
