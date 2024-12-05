// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { IStakeManager } from "./interfaces/IStakeManager.sol";
import { IStakeVault } from "./interfaces/IStakeVault.sol";
import { TrustedCodehashAccess } from "./TrustedCodehashAccess.sol";

// Rewards Streamer with Multiplier Points
contract RewardsStreamerMP is
    Initializable,
    UUPSUpgradeable,
    IStakeManager,
    TrustedCodehashAccess,
    ReentrancyGuardUpgradeable
{
    error StakingManager__InvalidVault();
    error StakingManager__VaultNotRegistered();
    error StakingManager__VaultAlreadyRegistered();
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();
    error StakingManager__InvalidLockingPeriod();
    error StakingManager__CannotRestakeWithLockedFunds();
    error StakingManager__TokensAreLocked();
    error StakingManager__AlreadyLocked();
    error StakingManager__EmergencyModeEnabled();
    error StakingManager__DurationCannotBeZero();

    IERC20 public STAKING_TOKEN;

    uint256 public constant SCALE_FACTOR = 1e18;
    uint256 public constant MP_RATE_PER_YEAR = 1;

    uint256 public constant YEAR = 365 days;
    uint256 public constant MIN_LOCKUP_PERIOD = 90 days;
    uint256 public constant MAX_LOCKUP_PERIOD = 4 * YEAR;
    uint256 public constant MAX_MULTIPLIER = 4;

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

    struct Account {
        uint256 stakedBalance;
        uint256 accountRewardIndex;
        uint256 mpAccrued;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address vault => Account data) public accounts;
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
    function registerVault() external onlyTrustedCodehash {
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
     * @param user The address of the user
     * @return The vaults owned by the user
     */
    function getUserVaults(address user) external view returns (address[] memory) {
        return vaults[user];
    }

    /**
     * @notice Get the total multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the multiplier points
     * @param user The address of the user
     * @return The total multiplier points for the user
     */
    function mpBalanceOfUser(address user) external view returns (uint256) {
        address[] memory userVaults = vaults[user];
        uint256 userTotalMP = 0;

        for (uint256 i = 0; i < userVaults.length; i++) {
            Account storage account = accounts[userVaults[i]];
            userTotalMP += account.mpAccrued + _getAccountPendingdMP(account);
        }
        return userTotalMP;
    }

    /**
     * @notice Get the total maximum multiplier points for a user
     * @dev Iterates over all vaults owned by the user and sums the maximum multiplier points
     * @param user The address of the user
     * @return The total maximum multiplier points for the user
     */
    function getUserTotalMaxMP(address user) external view returns (uint256) {
        address[] memory userVaults = vaults[user];
        uint256 userTotalMaxMP = 0;

        for (uint256 i = 0; i < userVaults.length; i++) {
            userTotalMaxMP += accounts[userVaults[i]].maxMP;
        }
        return userTotalMaxMP;
    }

    /**
     * @notice Get the total staked balance for a user
     * @dev Iterates over all vaults owned by the user and sums the staked balances
     * @param user The address of the user
     * @return The total staked balance for the user
     */
    function getUserTotalStakedBalance(address user) external view returns (uint256) {
        address[] memory userVaults = vaults[user];
        uint256 userTotalStake = 0;

        for (uint256 i = 0; i < userVaults.length; i++) {
            userTotalStake += accounts[userVaults[i]].stakedBalance;
        }
        return userTotalStake;
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

        if (lockPeriod != 0 && (lockPeriod < MIN_LOCKUP_PERIOD || lockPeriod > MAX_LOCKUP_PERIOD)) {
            revert StakingManager__InvalidLockingPeriod();
        }

        _updateGlobalState();
        _updateAccountMP(msg.sender);

        Account storage account = accounts[msg.sender];
        if (account.lockUntil != 0 && account.lockUntil > block.timestamp) {
            revert StakingManager__CannotRestakeWithLockedFunds();
        }

        account.stakedBalance += amount;
        totalStaked += amount;

        uint256 initialMP = amount;
        uint256 potentialMP = amount * MAX_MULTIPLIER;
        uint256 bonusMP = 0;

        if (lockPeriod != 0) {
            bonusMP = _calculateBonusMP(amount, lockPeriod);
            account.lockUntil = block.timestamp + lockPeriod;
        } else {
            account.lockUntil = 0;
        }

        uint256 accountMaxMP = initialMP + bonusMP + potentialMP;
        uint256 accountMP = initialMP + bonusMP;

        account.mpAccrued += accountMP;
        totalMPAccrued += accountMP;

        account.maxMP += accountMaxMP;
        totalMaxMP += accountMaxMP;

        account.accountRewardIndex = rewardIndex;
    }

    function lock(uint256 lockPeriod)
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
        nonReentrant
    {
        if (lockPeriod < MIN_LOCKUP_PERIOD || lockPeriod > MAX_LOCKUP_PERIOD) {
            revert StakingManager__InvalidLockingPeriod();
        }

        Account storage account = accounts[msg.sender];

        if (account.lockUntil > 0) {
            revert StakingManager__AlreadyLocked();
        }

        if (account.stakedBalance == 0) {
            revert StakingManager__InsufficientBalance();
        }

        _updateGlobalState();
        _updateAccountMP(msg.sender);

        uint256 additionalBonusMP = _calculateBonusMP(account.stakedBalance, lockPeriod);

        // Update account state
        account.lockUntil = block.timestamp + lockPeriod;
        account.mpAccrued += additionalBonusMP;
        account.maxMP += additionalBonusMP;

        // Update global state
        totalMPAccrued += additionalBonusMP;
        totalMaxMP += additionalBonusMP;

        account.accountRewardIndex = rewardIndex;
    }

    function unstake(uint256 amount)
        external
        onlyTrustedCodehash
        onlyNotEmergencyMode
        onlyRegisteredVault
        nonReentrant
    {
        Account storage account = accounts[msg.sender];
        if (amount > account.stakedBalance) {
            revert StakingManager__InsufficientBalance();
        }

        if (block.timestamp < account.lockUntil) {
            revert StakingManager__TokensAreLocked();
        }
        _unstake(amount, account, msg.sender);
    }

    function _unstake(uint256 amount, Account storage account, address accountAddress) internal {
        _updateGlobalState();
        _updateAccountMP(accountAddress);

        uint256 previousStakedBalance = account.stakedBalance;

        // solhint-disable-next-line
        uint256 mpToReduce = Math.mulDiv(account.mpAccrued, amount, previousStakedBalance);
        uint256 maxMPToReduce = Math.mulDiv(account.maxMP, amount, previousStakedBalance);

        account.stakedBalance -= amount;
        account.mpAccrued -= mpToReduce;
        account.maxMP -= maxMPToReduce;
        account.accountRewardIndex = rewardIndex;
        totalMPAccrued -= mpToReduce;
        totalMaxMP -= maxMPToReduce;
        totalStaked -= amount;
    }

    // @notice Allows an account to leave the system. This can happen when a
    //         user doesn't agree with an upgrade of the stake manager.
    // @dev This function is protected by whitelisting the codehash of the caller.
    //      This ensures `StakeVault`s will call this function only if they don't
    //      trust the `StakeManager` (e.g. in case of an upgrade).
    function leave() external onlyTrustedCodehash nonReentrant {
        Account storage account = accounts[msg.sender];

        if (account.stakedBalance > 0) {
            // calling `_unstake` to update accounting accordingly
            _unstake(account.stakedBalance, account, msg.sender);

            // further cleanup that isn't done in `_unstake`
            account.accountRewardIndex = 0;
            account.lockUntil = 0;
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
        (uint256 adjustedRewardIndex, uint256 newTotalMPAccrued) = _pendingTotalMPAccrued();
        if (newTotalMPAccrued > totalMPAccrued) {
            totalMPAccrued = newTotalMPAccrued;
            lastMPUpdatedTime = block.timestamp;
        }

        if (adjustedRewardIndex != rewardIndex) {
            rewardIndex = adjustedRewardIndex;
        }
    }

    function _pendingTotalMPAccrued() internal view returns (uint256, uint256) {
        uint256 adjustedRewardIndex = rewardIndex;

        if (totalMaxMP == 0) {
            return (adjustedRewardIndex, totalMPAccrued);
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        if (timeDiff == 0) {
            return (adjustedRewardIndex, totalMPAccrued);
        }

        uint256 accruedMP = (timeDiff * totalStaked * MP_RATE_PER_YEAR) / YEAR;
        if (totalMPAccrued + accruedMP > totalMaxMP) {
            accruedMP = totalMaxMP - totalMPAccrued;
        }

        uint256 newTotalMPAccrued = totalMPAccrued + accruedMP;

        // Adjust rewardIndex before updating totalMP
        uint256 previousTotalWeight = totalStaked + totalMPAccrued;
        uint256 newTotalWeight = totalStaked + newTotalMPAccrued;
        if (previousTotalWeight != 0 && newTotalWeight != previousTotalWeight) {
            adjustedRewardIndex = (rewardIndex * previousTotalWeight) / newTotalWeight;
        }

        return (adjustedRewardIndex, newTotalMPAccrued);
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
        (uint256 adjustedRewardIndex, uint256 newTotalMPAccrued) = _pendingTotalMPAccrued();
        uint256 totalWeight = totalStaked + newTotalMPAccrued;

        if (totalWeight == 0) {
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

        uint256 newRewardIndex = adjustedRewardIndex + Math.mulDiv(accruedRewards, SCALE_FACTOR, totalWeight);

        return (accruedRewards, newRewardIndex);
    }

    function _calculateBonusMP(uint256 amount, uint256 lockPeriod) internal pure returns (uint256) {
        return Math.mulDiv(amount, lockPeriod, YEAR);
    }

    function _getAccountPendingdMP(Account storage account) internal view returns (uint256) {
        if (account.maxMP == 0 || account.stakedBalance == 0) {
            return 0;
        }

        uint256 timeDiff = block.timestamp - account.lastMPUpdateTime;
        if (timeDiff == 0) {
            return 0;
        }

        uint256 accruedMP = Math.mulDiv(timeDiff * account.stakedBalance, MP_RATE_PER_YEAR, YEAR);

        if (account.mpAccrued + accruedMP > account.maxMP) {
            accruedMP = account.maxMP - account.mpAccrued;
        }
        return accruedMP;
    }

    function _updateAccountMP(address accountAddress) internal {
        Account storage account = accounts[accountAddress];
        uint256 accruedMP = _getAccountPendingdMP(account);

        account.mpAccrued += accruedMP;
        account.lastMPUpdateTime = block.timestamp;
    }

    function updateAccountMP(address accountAddress) external onlyNotEmergencyMode {
        _updateAccountMP(accountAddress);
    }

    function enableEmergencyMode() external onlyOwner {
        if (emergencyModeEnabled) {
            revert StakingManager__EmergencyModeEnabled();
        }
        emergencyModeEnabled = true;
    }

    function getStakedBalance(address accountAddress) public view returns (uint256) {
        return accounts[accountAddress].stakedBalance;
    }

    function getAccount(address accountAddress) external view returns (Account memory) {
        return accounts[accountAddress];
    }

    function totalRewardsSupply() public view returns (uint256) {
        return totalRewardsAccrued + _calculatePendingRewards();
    }

    function rewardsBalanceOf(address accountAddress) public view returns (uint256) {
        uint256 newRewardIndex;
        (, newRewardIndex) = _pendingRewardIndex();

        Account storage account = accounts[accountAddress];

        uint256 accountWeight = account.stakedBalance + _mpBalanceOf(accountAddress);
        uint256 deltaRewardIndex = newRewardIndex - account.accountRewardIndex;

        return (accountWeight * deltaRewardIndex) / SCALE_FACTOR;
    }

    function rewardsBalanceOfUser(address user) external view returns (uint256) {
        address[] memory userVaults = vaults[user];
        uint256 userTotalRewards = 0;

        for (uint256 i = 0; i < userVaults.length; i++) {
            userTotalRewards += rewardsBalanceOf(userVaults[i]);
        }

        return userTotalRewards;
    }

    function _mpBalanceOf(address accountAddress) internal view returns (uint256) {
        Account storage account = accounts[accountAddress];
        return account.mpAccrued + _getAccountPendingdMP(account);
    }

    function mpBalanceOf(address accountAddress) external view returns (uint256) {
        return _mpBalanceOf(accountAddress);
    }
}
