// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { IStakeManager } from "./interfaces/IStakeManager.sol";
import { TrustedCodehashAccess } from "./TrustedCodehashAccess.sol";

// Rewards Streamer with Multiplier Points
contract RewardsStreamerMP is UUPSUpgradeable, IStakeManager, TrustedCodehashAccess, ReentrancyGuardUpgradeable {
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();
    error StakingManager__InvalidLockingPeriod();
    error StakingManager__CannotRestakeWithLockedFunds();
    error StakingManager__TokensAreLocked();
    error StakingManager__AlreadyLocked();
    error StakingManager__EmergencyModeEnabled();

    IERC20 public STAKING_TOKEN;
    IERC20 public REWARD_TOKEN;

    uint256 public constant SCALE_FACTOR = 1e18;
    uint256 public constant MP_RATE_PER_YEAR = 1e18;

    uint256 public constant MIN_LOCKUP_PERIOD = 90 days;
    uint256 public constant MAX_LOCKUP_PERIOD = 4 * 365 days;
    uint256 public constant MAX_MULTIPLIER = 4;

    uint256 public totalStaked;
    uint256 public totalMP;
    uint256 public totalMaxMP;
    uint256 public rewardIndex;
    uint256 public accountedRewards;
    uint256 public lastMPUpdatedTime;
    bool public emergencyModeEnabled;

    struct Account {
        uint256 stakedBalance;
        uint256 accountRewardIndex;
        uint256 accountMP;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address account => Account data) public accounts;

    modifier onlyNotEmergencyMode() {
        if (emergencyModeEnabled) {
            revert StakingManager__EmergencyModeEnabled();
        }
        _;
    }

    constructor() {
        _disableInitializers();
    }

    function initialize(address _owner, address _stakingToken, address _rewardToken) public initializer {
        __TrustedCodehashAccess_init(_owner);
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        STAKING_TOKEN = IERC20(_stakingToken);
        REWARD_TOKEN = IERC20(_rewardToken);
        lastMPUpdatedTime = block.timestamp;
    }

    function _authorizeUpgrade(address) internal view override {
        _checkOwner();
    }

    function stake(uint256 amount, uint256 lockPeriod) external onlyTrustedCodehash onlyNotEmergencyMode nonReentrant {
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

        uint256 accountRewards = calculateAccountRewards(msg.sender);
        if (accountRewards > 0) {
            distributeRewards(msg.sender, accountRewards);
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

        account.accountMP += accountMP;
        totalMP += accountMP;

        account.maxMP += accountMaxMP;
        totalMaxMP += accountMaxMP;

        account.accountRewardIndex = rewardIndex;
        account.lastMPUpdateTime = block.timestamp;
    }

    function lock(uint256 lockPeriod) external onlyTrustedCodehash onlyNotEmergencyMode nonReentrant {
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
        account.accountMP += additionalBonusMP;
        account.maxMP += additionalBonusMP;

        // Update global state
        totalMP += additionalBonusMP;
        totalMaxMP += additionalBonusMP;

        account.accountRewardIndex = rewardIndex;
        account.lastMPUpdateTime = block.timestamp;
    }

    function unstake(uint256 amount) external onlyTrustedCodehash onlyNotEmergencyMode nonReentrant {
        Account storage account = accounts[msg.sender];
        if (amount > account.stakedBalance) {
            revert StakingManager__InsufficientBalance();
        }

        if (block.timestamp < account.lockUntil) {
            revert StakingManager__TokensAreLocked();
        }

        _updateGlobalState();
        _updateAccountMP(msg.sender);

        uint256 accountRewards = calculateAccountRewards(msg.sender);
        if (accountRewards > 0) {
            distributeRewards(msg.sender, accountRewards);
        }

        uint256 previousStakedBalance = account.stakedBalance;

        uint256 mpToReduce = (account.accountMP * amount * SCALE_FACTOR) / (previousStakedBalance * SCALE_FACTOR);
        uint256 maxMPToReduce = (account.maxMP * amount * SCALE_FACTOR) / (previousStakedBalance * SCALE_FACTOR);

        account.stakedBalance -= amount;
        account.accountMP -= mpToReduce;
        account.maxMP -= maxMPToReduce;
        totalMP -= mpToReduce;
        totalMaxMP -= maxMPToReduce;
        totalStaked -= amount;

        account.accountRewardIndex = rewardIndex;
    }

    function _updateGlobalState() internal {
        updateGlobalMP();
        updateRewardIndex();
    }

    function updateGlobalState() external onlyNotEmergencyMode {
        _updateGlobalState();
    }

    function updateGlobalMP() internal {
        if (totalMaxMP == 0) {
            lastMPUpdatedTime = block.timestamp;
            return;
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        if (timeDiff == 0) {
            return;
        }

        uint256 accruedMP = (timeDiff * totalStaked * MP_RATE_PER_YEAR) / (365 days * SCALE_FACTOR);
        if (totalMP + accruedMP > totalMaxMP) {
            accruedMP = totalMaxMP - totalMP;
        }

        // Adjust rewardIndex before updating totalMP
        uint256 previousTotalWeight = totalStaked + totalMP;
        totalMP += accruedMP;
        uint256 newTotalWeight = totalStaked + totalMP;

        if (previousTotalWeight != 0 && newTotalWeight != previousTotalWeight) {
            rewardIndex = (rewardIndex * previousTotalWeight) / newTotalWeight;
        }

        lastMPUpdatedTime = currentTime;
    }

    function updateRewardIndex() internal {
        uint256 totalWeight = totalStaked + totalMP;
        if (totalWeight == 0) {
            return;
        }

        uint256 rewardBalance = REWARD_TOKEN.balanceOf(address(this));
        uint256 newRewards = rewardBalance > accountedRewards ? rewardBalance - accountedRewards : 0;

        if (newRewards > 0) {
            rewardIndex += (newRewards * SCALE_FACTOR) / totalWeight;
            accountedRewards += newRewards;
        }
    }

    function _calculateBonusMP(uint256 amount, uint256 lockPeriod) internal view returns (uint256) {
        uint256 lockMultiplier = (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR) / MAX_LOCKUP_PERIOD;
        return amount * lockMultiplier / SCALE_FACTOR;
    }

    function _updateAccountMP(address accountAddress) internal {
        Account storage account = accounts[accountAddress];

        if (account.maxMP == 0 || account.stakedBalance == 0) {
            account.lastMPUpdateTime = block.timestamp;
            return;
        }

        uint256 timeDiff = block.timestamp - account.lastMPUpdateTime;
        if (timeDiff == 0) {
            return;
        }

        uint256 accruedMP = (timeDiff * account.stakedBalance * MP_RATE_PER_YEAR) / (365 days * SCALE_FACTOR);

        if (account.accountMP + accruedMP > account.maxMP) {
            accruedMP = account.maxMP - account.accountMP;
        }

        account.accountMP += accruedMP;
        account.lastMPUpdateTime = block.timestamp;
    }

    function updateAccountMP(address accountAddress) external onlyNotEmergencyMode {
        _updateAccountMP(accountAddress);
    }

    function calculateAccountRewards(address accountAddress) public view returns (uint256) {
        Account storage account = accounts[accountAddress];
        uint256 accountWeight = account.stakedBalance + account.accountMP;
        uint256 deltaRewardIndex = rewardIndex - account.accountRewardIndex;
        return (accountWeight * deltaRewardIndex) / SCALE_FACTOR;
    }

    function distributeRewards(address to, uint256 amount) internal {
        uint256 rewardBalance = REWARD_TOKEN.balanceOf(address(this));
        // If amount is higher than the contract's balance (for rounding error), transfer the balance.
        if (amount > rewardBalance) {
            amount = rewardBalance;
        }

        accountedRewards -= amount;

        bool success = REWARD_TOKEN.transfer(to, amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }
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

    function getPendingRewards(address accountAddress) external view returns (uint256) {
        return calculateAccountRewards(accountAddress);
    }

    function getAccount(address accountAddress) external view returns (Account memory) {
        return accounts[accountAddress];
    }
}
