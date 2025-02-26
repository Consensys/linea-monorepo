// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/**
 * @title IRewardDistributor
 * @notice Interface for Reward Distributor contract.
 * @dev This interface is necessary to unify reward distributor contracts.
 * @dev Karma token contract makes use of this to aggregate rewards.
 */
interface IRewardDistributor {
    /**
     * @notice Returns the total supply of rewards.
     * @return Total supply of rewards.
     */
    function totalRewardsSupply() external view returns (uint256);

    /**
     * @notice Returns the balance of rewards for a vault
     * @param account Address of the vault.
     * @return Balance of rewards for the vault.
     */
    function rewardsBalanceOf(address account) external view returns (uint256);

    /**
     * @notice Returns the balance of rewards for an account.
     * @param user Address of the account.
     * @return Balance of rewards for the account.
     */
    function rewardsBalanceOfAccount(address user) external view returns (uint256);
    function setReward(uint256 amount, uint256 duration) external;
}
