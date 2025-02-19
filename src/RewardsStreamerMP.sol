// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { IStakeManager } from "./interfaces/IStakeManager.sol";
import { IStakeVault } from "./interfaces/IStakeVault.sol";
import { IRewardProvider } from "./interfaces/IRewardProvider.sol";
import { TrustedCodehashAccess } from "./TrustedCodehashAccess.sol";
import { StakeMath } from "./math/StakeMath.sol";

// Rewards Streamer with Multiplier Points
contract RewardsStreamerMP is
    Initializable,
    UUPSUpgradeable,
    IStakeManager,
    TrustedCodehashAccess,
    ReentrancyGuardUpgradeable,
    IRewardProvider,
    StakeMath
{
    error StakingManager__InvalidVault();
    error StakingManager__VaultNotRegistered();
    error StakingManager__VaultAlreadyRegistered();
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();
    error StakingManager__LockingPeriodCannotBeZero();
    error StakingManager__CannotRestakeWithLockedFunds();
    error StakingManager__TokensAreLocked();
    error StakingManager__AlreadyLocked();
    error StakingManager__EmergencyModeEnabled();
    error StakingManager__DurationCannotBeZero();

    IERC20 public STAKING_TOKEN;

    uint256 public constant SCALE_FACTOR = 1e18;

    uint256 public totalStaked;
    uint256 public totalMPAccrued;
    uint256 public totalMaxMP;
    uint256 public rewardIndex;
    uint256 public lastMPUpdatedTime;
    bool public emergencyModeEnabled;

    uint256 public totalRewardsAccrued;
    uint256 public rewardAmount;
    uint256 public lastRewardTime;
    uint256 public rewardStartTime;
    uint256 public rewardEndTime;

    struct VaultData {
        uint256 stakedBalance;
        uint256 rewardIndex;
        uint256 mpAccrued;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address vault => VaultData data) public vaultData;
    mapping(address owner => address[] vault) public vaults;
    mapping(address vault => address owner) public vaultOwners;

    modifier onlyRegisteredVault() {
        if (vaultOwners[msg.sender] == address(0)) {
            revert StakingManager__VaultNotRegistered();
        }
        _;
    }

    modifier onlyNotEmergencyMode() {
        if (emergencyModeEnabled) {
            revert StakingManager__EmergencyModeEnabled();
        }
        _;
    }

    constructor() {
        _disableInitializers();
    }

    function initialize(address _owner, address _stakingToken) public initializer {
        __TrustedCodehashAccess_init(_owner);
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        STAKING_TOKEN = IERC20(_stakingToken);
        lastMPUpdatedTime = block.timestamp;
    }

    function _authorizeUpgrade(address) internal view override {
        _checkOwner();
    }

    /**
     * @notice Registers a vault with its owner. Called by the vault itself during initialization.
     * @dev Only callable by contracts with trusted codehash
     */
    function registerVault() external onlyTrustedCodehash onlyNotEmergencyMode {
        address vault = msg.sender;
        address owner = IStakeVault(vault).owner();

        if (vaultOwners[vault] != address(0)) {
            revert StakingManager__VaultAlreadyRegistered();
        }

        // Verify this is a legitimate vault by checking it points to stakeManager
        if (address(IStakeVault(vault).stakeManager()) != address(this)) {
            revert StakingManager__InvalidVault();
        }

        vaultOwners[vault] = owner;
        vaults[owner].push(vault);
    }

    /**
     * @notice Get the vaults owned by a user
     * @param account The address of the user
     * @return The vaults owned by the user
     */
    function getAccountVaults(address account) external view returns (address[] memory) {
        return vaults[account];
    }

    /**
     * @notice Get the total multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the multiplier points
     * @param account The address of the user
     * @return The total multiplier points for the user
     */
    function mpBalanceOfAccount(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalMP = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            VaultData storage vault = vaultData[accountVaults[i]];
            accountTotalMP += vault.mpAccrued + _getVaultPendingMP(vault);
        }
        return accountTotalMP;
    }

    /**
     * @notice Get the total maximum multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the maximum multiplier points
     * @param account The address of the user
     * @return The total maximum multiplier points for the user
     */
    function getAccountTotalMaxMP(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalMaxMP = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalMaxMP += vaultData[accountVaults[i]].maxMP;
        }
        return accountTotalMaxMP;
    }

    /**
     * @notice Get the total staked balance for a user
     * @dev Iterates over all vaults owned by the user and sums the staked balances
     * @param account The address of the user
     * @return The total staked balance for the user
     */
    function getAccountTotalStakedBalance(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalStake = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalStake += vaultData[accountVaults[i]].stakedBalance;
        }
        return accountTotalStake;
    }

    function stake(
        uint256 amount,
        uint256 lockPeriod
    )
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
        nonReentrant
    {
        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        _updateGlobalState();
        _updateVaultMP(msg.sender, true);

        VaultData storage vault = vaultData[msg.sender];
        if (vault.lockUntil != 0 && vault.lockUntil > block.timestamp) {
            revert StakingManager__CannotRestakeWithLockedFunds();
        }
        (uint256 _deltaMpTotal, uint256 _deltaMPMax, uint256 _newLockEnd) =
            _calculateStake(vault.stakedBalance, vault.maxMP, vault.lockUntil, block.timestamp, amount, lockPeriod);

        vault.stakedBalance += amount;
        totalStaked += amount;

        if (lockPeriod != 0) {
            vault.lockUntil = _newLockEnd;
        } else {
            vault.lockUntil = 0;
        }

        vault.mpAccrued += _deltaMpTotal;
        totalMPAccrued += _deltaMpTotal;

        vault.maxMP += _deltaMPMax;
        totalMaxMP += _deltaMPMax;

        vault.rewardIndex = rewardIndex;
    }

    function lock(uint256 lockPeriod)
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
        nonReentrant
    {
        VaultData storage vault = vaultData[msg.sender];

        if (vault.lockUntil > 0) {
            revert StakingManager__AlreadyLocked();
        }

        if (lockPeriod == 0) {
            revert StakingManager__LockingPeriodCannotBeZero();
        }

        _updateGlobalState();
        _updateVaultMP(msg.sender, true);
        (uint256 deltaMp, uint256 newLockEnd) =
            _calculateLock(vault.stakedBalance, vault.maxMP, vault.lockUntil, block.timestamp, lockPeriod);

        // Update account state
        vault.lockUntil = newLockEnd;
        vault.mpAccrued += deltaMp;
        vault.maxMP += deltaMp;

        // Update global state
        totalMPAccrued += deltaMp;
        totalMaxMP += deltaMp;

        vault.rewardIndex = rewardIndex;
    }

    function unstake(uint256 amount)
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
        nonReentrant
    {
        VaultData storage vault = vaultData[msg.sender];
        _unstake(amount, vault, msg.sender);
    }

    function _unstake(uint256 amount, VaultData storage vault, address vaultAddress) internal {
        _updateGlobalState();
        _updateVaultMP(vaultAddress, true);

        (uint256 _deltaMpTotal, uint256 _deltaMpMax) = _calculateUnstake(
            vault.stakedBalance, vault.lockUntil, block.timestamp, vault.mpAccrued, vault.maxMP, amount
        );
        vault.stakedBalance -= amount;
        vault.mpAccrued -= _deltaMpTotal;
        vault.maxMP -= _deltaMpMax;
        vault.rewardIndex = rewardIndex;
        totalMPAccrued -= _deltaMpTotal;
        totalMaxMP -= _deltaMpMax;
        totalStaked -= amount;
    }

    // @notice Allows an account to leave the system. This can happen when a
    //         user doesn't agree with an upgrade of the stake manager.
    // @dev This function is protected by whitelisting the codehash of the caller.
    //      This ensures `StakeVault`s will call this function only if they don't
    //      trust the `StakeManager` (e.g. in case of an upgrade).
    function leave() external onlyTrustedCodehash nonReentrant {
        VaultData storage vault = vaultData[msg.sender];

        if (vault.stakedBalance > 0) {
            //updates lockuntil to allow unstake early
            vault.lockUntil = block.timestamp;
            // calling `_unstake` to update accounting accordingly
            _unstake(vault.stakedBalance, vault, msg.sender);

            // further cleanup that isn't done in `_unstake`
            vault.rewardIndex = 0;
            vault.lockUntil = 0;
        }
    }

    function _updateGlobalState() internal {
        updateGlobalMP();
        updateRewardIndex();
    }

    function updateGlobalState() external onlyNotEmergencyMode {
        _updateGlobalState();
    }

    function updateGlobalMP() internal {
        uint256 newTotalMPAccrued = _pendingTotalMPAccrued();
        if (newTotalMPAccrued > totalMPAccrued) {
            totalMPAccrued = newTotalMPAccrued;
            lastMPUpdatedTime = block.timestamp;
        }
    }

    function _pendingTotalMPAccrued() internal view returns (uint256) {
        if (totalMaxMP == 0) {
            return totalMPAccrued;
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        if (timeDiff == 0) {
            return totalMPAccrued;
        }

        uint256 accruedMP = _accrueMP(totalStaked, timeDiff);
        if (totalMPAccrued + accruedMP > totalMaxMP) {
            accruedMP = totalMaxMP - totalMPAccrued;
        }

        uint256 newTotalMPAccrued = totalMPAccrued + accruedMP;

        return newTotalMPAccrued;
    }

    function setReward(uint256 amount, uint256 duration) external onlyOwner {
        if (duration == 0) {
            revert StakingManager__DurationCannotBeZero();
        }

        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        // this will call _updateRewardIndex and update the totalRewardsAccrued
        _updateGlobalState();

        // in case _updateRewardIndex returns earlier,
        // we still update the lastRewardTime
        lastRewardTime = block.timestamp;
        rewardAmount = amount;
        rewardStartTime = block.timestamp;
        rewardEndTime = block.timestamp + duration;
    }

    function _calculatePendingRewards() internal view returns (uint256) {
        if (rewardEndTime <= rewardStartTime) {
            // No active reward period
            return 0;
        }

        uint256 currentTime = block.timestamp < rewardEndTime ? block.timestamp : rewardEndTime;

        if (currentTime <= lastRewardTime) {
            // No new rewards have accrued since lastRewardTime
            return 0;
        }

        uint256 timeElapsed = currentTime - lastRewardTime;
        uint256 duration = rewardEndTime - rewardStartTime;

        if (duration == 0) {
            // Prevent division by zero
            return 0;
        }

        uint256 accruedRewards = Math.mulDiv(timeElapsed, rewardAmount, duration);
        return accruedRewards;
    }

    function updateRewardIndex() internal {
        uint256 accruedRewards;
        uint256 newRewardIndex;

        (accruedRewards, newRewardIndex) = _pendingRewardIndex();
        totalRewardsAccrued += accruedRewards;

        if (newRewardIndex > rewardIndex) {
            rewardIndex = newRewardIndex;
            lastRewardTime = block.timestamp < rewardEndTime ? block.timestamp : rewardEndTime;
        }
    }

    function pendingRewardIndex() external view returns (uint256) {
        uint256 newRewardIndex;
        (, newRewardIndex) = _pendingRewardIndex();
        return newRewardIndex;
    }

    function _pendingRewardIndex() internal view returns (uint256, uint256) {
        uint256 totalShares = totalStaked;

        if (totalShares == 0) {
            return (0, rewardIndex);
        }

        uint256 currentTime = block.timestamp;
        uint256 applicableTime = rewardEndTime > currentTime ? currentTime : rewardEndTime;
        uint256 elapsedTime = applicableTime - lastRewardTime;

        if (elapsedTime == 0) {
            return (0, rewardIndex);
        }

        uint256 accruedRewards = _calculatePendingRewards();
        if (accruedRewards == 0) {
            return (0, rewardIndex);
        }

        uint256 newRewardIndex = rewardIndex + Math.mulDiv(accruedRewards, SCALE_FACTOR, totalShares);

        return (accruedRewards, newRewardIndex);
    }

    function _getVaultPendingMP(VaultData storage vault) internal view returns (uint256) {
        if (block.timestamp == vault.lastMPUpdateTime) {
            return 0;
        }
        if (vault.maxMP == 0 || vault.stakedBalance == 0) {
            return 0;
        }

        uint256 deltaMpTotal = _calculateAccrual(
            vault.stakedBalance, vault.mpAccrued, vault.maxMP, vault.lastMPUpdateTime, block.timestamp
        );

        return deltaMpTotal;
    }

    function _updateVaultMP(address vaultAddress, bool forceMPUpdate) internal {
        VaultData storage vault = vaultData[vaultAddress];
        uint256 accruedMP = _getVaultPendingMP(vault);
        if (accruedMP > 0 || forceMPUpdate) {
            vault.mpAccrued += accruedMP;
            vault.lastMPUpdateTime = block.timestamp;
        }
    }

    function updateVaultMP(address vaultAddress) external onlyNotEmergencyMode {
        _updateVaultMP(vaultAddress, false);
    }

    function enableEmergencyMode() external onlyOwner onlyNotEmergencyMode {
        emergencyModeEnabled = true;
    }

    function getStakedBalance(address vaultAddress) public view returns (uint256) {
        return vaultData[vaultAddress].stakedBalance;
    }

    function getVault(address vaultAddress) external view returns (VaultData memory) {
        return vaultData[vaultAddress];
    }

    function totalRewardsSupply() public view returns (uint256) {
        return totalRewardsAccrued + _calculatePendingRewards();
    }

    function rewardsBalanceOf(address vaultAddress) public view returns (uint256) {
        uint256 newRewardIndex;
        (, newRewardIndex) = _pendingRewardIndex();

        VaultData storage vault = vaultData[vaultAddress];

        uint256 accountShares = vault.stakedBalance;
        uint256 deltaRewardIndex = newRewardIndex - vault.rewardIndex;

        return (accountShares * deltaRewardIndex) / SCALE_FACTOR;
    }

    function rewardsBalanceOfAccount(address account) external view returns (uint256) {
        address[] memory accountVaults = vaults[account];
        uint256 accountTotalRewards = 0;

        for (uint256 i = 0; i < accountVaults.length; i++) {
            accountTotalRewards += rewardsBalanceOf(accountVaults[i]);
        }

        return accountTotalRewards;
    }

    function _mpBalanceOf(address vaultAddress) internal view returns (uint256) {
        VaultData storage vault = vaultData[vaultAddress];
        return vault.mpAccrued + _getVaultPendingMP(vault);
    }

    function mpBalanceOf(address vaultAddress) external view returns (uint256) {
        return _mpBalanceOf(vaultAddress);
    }
}
