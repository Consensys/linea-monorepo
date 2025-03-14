// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ITrustedCodehashAccess } from "./ITrustedCodehashAccess.sol";
import { IStakeConstants } from "./IStakeConstants.sol";

/**
 * @title IStakeManager
 * @notice StakeManager interface for staking and unstaking funds.
 * @dev StakeManager is a contract that allows users to stake and unstake funds
 *      for a determined period of time. It also allows users to lock their
 *      funds for a determined period of time.
 */
interface IStakeManager is ITrustedCodehashAccess, IStakeConstants {
    /// @notice Emitted when a vault isn't registered.
    error StakingManager__VaultNotRegistered();
    /// @notice Emitted when a vault is already registered.
    error StakingManager__VaultAlreadyRegistered();
    /// @notice Emitted when the vault is invalid
    error StakingManager__InvalidVault();
    /// @notice Emitted when the amount to stake is zero.
    error StakingManager__AmountCannotBeZero();
    /// @notice Emitted when the lock period is not zero.
    error StakingManager__InvalidLockPeriod();
    /// @notice Emitted when emergency mode is enabled.
    error StakingManager__EmergencyModeEnabled();
    /// @notice Emitted trying to migrate to non empty vault
    error StakingManager__MigrationTargetHasFunds();
    /// @notice Emitted when the caller is not the owner of the vault.
    error StakingManager__Unauthorized();
    /// @notice Emitted when the duration is zero.
    error StakingManager__DurationCannotBeZero();
    /// @notice Emitted when there are insufficient funds to stake.
    error StakingManager__InsufficientBalance();
    /// @notice Emitted when the reward period has not ended.
    error StakingManager__RewardPeriodNotEnded();

    /// @notice Emitted when a vault is registered.
    event VaultRegistered(address indexed vault, address indexed owner);
    /// @notice Emitted when a vault is migrated.
    event VaultMigrated(address indexed from, address indexed to);
    /// @notice Emitted when funds are staked.
    event Staked(address indexed vault, uint256 amount, uint256 lockPeriod);
    /// @notice Emitted when funds are locked.
    event Locked(address indexed vault, uint256 lockPeriod, uint256 lockUntil);
    /// @notice Emitted when funds are unstaked.
    event Unstaked(address indexed vault, uint256 amount);
    /// @notice Emitted when emergency mode is enabled.
    event EmergencyModeEnabled();
    /// @notice Emitted when an account leaves the system
    event AccountLeft(address indexed vault);
    /// @notice Emited when accounts compound their MP
    event Compound(address indexed vault, uint256 amount);

    function registerVault() external;
    function stake(uint256 _amount, uint256 _seconds) external;
    function lock(uint256 _seconds) external;
    function unstake(uint256 _amount) external;
    function leave() external;
    function migrateToVault(address migrateTo) external;
    function compound(address _vault) external;

    function emergencyModeEnabled() external view returns (bool);
    function totalStaked() external view returns (uint256);
    function totalMPAccrued() external view returns (uint256);
    function totalMaxMP() external view returns (uint256);
    function getStakedBalance(address _vault) external view returns (uint256 _balance);

    function STAKING_TOKEN() external view returns (IERC20);
}
