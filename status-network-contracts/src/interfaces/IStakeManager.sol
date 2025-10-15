// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

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
    error StakeManager__VaultNotRegistered();
    /// @notice Emitted when a vault is already registered.
    error StakeManager__VaultAlreadyRegistered();
    /// @notice Emitted when the vault is invalid
    error StakeManager__InvalidVault();
    /// @notice Emitted when the amount to stake is zero.
    error StakeManager__AmountCannotBeZero();
    /// @notice Emitted when emergency mode is enabled.
    error StakeManager__EmergencyModeEnabled();
    /// @notice Emitted trying to migrate to non empty vault
    error StakeManager__MigrationTargetHasFunds();
    /// @notice Emitted when the caller is not the owner of the vault.
    error StakeManager__Unauthorized();
    /// @notice Emitted when the duration is zero.
    error StakeManager__DurationCannotBeZero();
    /// @notice Emitted when there are insufficient funds to stake.
    error StakeManager__InsufficientBalance();
    /// @notice Emitted when the reward period has not ended.
    error StakeManager__RewardPeriodNotEnded();
    /// @notice Emitted when trying to unstake and funds are locked
    error StakeManager__FundsLocked();
    /// @notice Emitted when transfering rewards fails
    error StakeManager__RewardTransferFailed();

    /// @notice Emitted when a reward supplier is set
    event RewardsSupplierSet(address indexed supplier);
    /// @notice Emitted when a reward is set
    event RewardSet(uint256 amount, uint256 duration, uint256 startTime, uint256 endTime);
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
    event VaultLeft(address indexed vault);
    /// @notice Emited when accounts update their vaults
    event VaultUpdated(address indexed vault, uint256 rewardsAccrued, uint256 mpAccrued);
    /// @notice Emitted when rewards are redeemed
    event RewardsRedeemed(address indexed vault, uint256 amount);

    /**
     * @notice Registers a vault with its owner. Called by the vault itself during initialization.
     */
    function registerVault() external;

    /**
     * @notice Allows users to stake and start accruing MPs.
     * @param _amount The amount of tokens to stake
     * @param _seconds The duration to lock the stake
     * @param _currentLockUntil The current lock end time of the vault
     * @return _lockUntil The new lock end time of the vault
     */
    function stake(uint256 _amount, uint256 _seconds, uint256 _currentLockUntil) external returns (uint256 _lockUntil);

    /**
     * @notice Allows users to lock their staked balance for a specified duration.
     * @param _seconds The duration to lock the stake
     * @param _currentLockUntil The current lock end time of the vault
     * @return _lockUntil The new lock end time of the vault
     */
    function lock(uint256 _seconds, uint256 _currentLockUntil) external returns (uint256 _lockUntil);

    /**
     * @notice Allows users to unstake their staked balance.
     * @param _amount The amount of tokens to unstake
     */
    function unstake(uint256 _amount) external;

    /**
     * @notice Allows an account to leave the system.
     */
    function leave() external;

    /**
     * @notice Migrate the staked balance of a vault to another vault.
     * @param migrateTo The address of the vault to migrate to.
     */
    function migrateToVault(address migrateTo) external;

    /**
     * @notice Allows users to claim their accrued rewards.
     * @param vaultAddress The address of the vault to update.
     */
    function updateVault(address vaultAddress) external;

    /**
     * @notice Flag whether emergency mode is enabled.
     * @return bool True if emergency mode is enabled, false otherwise.
     */
    function emergencyModeEnabled() external view returns (bool);

    /**
     * @notice Returns the total staked balance across all vaults.
     * @return uint256 The total staked balance.
     */
    function totalStaked() external view returns (uint256);

    /**
     * @notice Returns the total MP accrued across all vaults.
     * @return uint256 The total MP accrued.
     */
    function totalMPAccrued() external view returns (uint256);

    /**
     * @notice Returns the total max MP across all vaults.
     * @return uint256 The total max MP.
     */
    function totalMaxMP() external view returns (uint256);

    /**
     * @notice Returns the staked balance of a vault.
     * @param _vault The address of the vault to query.
     * @return _balance The staked balance of the vault.
     */
    function stakedBalanceOf(address _vault) external view returns (uint256 _balance);

    /**
     * @notice Returns the staking token
     * @return IERC20 The staking token
     */
    function STAKING_TOKEN() external view returns (IERC20);
}
