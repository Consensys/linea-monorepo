// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ReentrancyGuard } from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

// Rewards Streamer with Multiplier Points
contract RewardsStreamerMP is ReentrancyGuard {
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();
    error StakingManager__InvalidLockingPeriod();
    error StakingManager__CannotRestakeWithLockedFunds();
    error StakingManager__TokensAreLocked();

    IERC20 public immutable STAKING_TOKEN;
    IERC20 public immutable REWARD_TOKEN;

    uint256 public constant SCALE_FACTOR = 1e18;
    uint256 public constant MP_RATE_PER_YEAR = 1e18;

    uint256 public constant MIN_LOCKING_PERIOD = 90 days;
    uint256 public constant MAX_LOCKING_PERIOD = 4 * 365 days;
    uint256 public constant MAX_MULTIPLIER = 4;

    uint256 public totalStaked;
    uint256 public totalMP;
    uint256 public totalMaxMP;
    uint256 public rewardIndex;
    uint256 public accountedRewards;
    uint256 public lastMPUpdatedTime;

    struct Account {
        uint256 stakedBalance;
        uint256 accountRewardIndex;
        uint256 accountMP;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address account => Account data) public accounts;

    constructor(address _stakingToken, address _rewardToken) {
        STAKING_TOKEN = IERC20(_stakingToken);
        REWARD_TOKEN = IERC20(_rewardToken);
        lastMPUpdatedTime = block.timestamp;
    }

    function stake(uint256 amount, uint256 lockPeriod) external nonReentrant {
        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        if (lockPeriod != 0 && (lockPeriod < MIN_LOCKING_PERIOD || lockPeriod > MAX_LOCKING_PERIOD)) {
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

        bool success = STAKING_TOKEN.transferFrom(msg.sender, address(this), amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        account.stakedBalance += amount;
        totalStaked += amount;

        uint256 initialMP = amount;
        uint256 potentialMP = amount * MAX_MULTIPLIER;
        uint256 bonusMP = 0;

        if (lockPeriod != 0) {
            uint256 lockMultiplier = (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR) / MAX_LOCKING_PERIOD;
            bonusMP = amount * lockMultiplier / SCALE_FACTOR;
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

    function unstake(uint256 amount) external nonReentrant {
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

        bool success = STAKING_TOKEN.transfer(msg.sender, amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        account.accountRewardIndex = rewardIndex;
    }

    function _updateGlobalState() internal {
        updateGlobalMP();
        updateRewardIndex();
    }

    function updateGlobalState() external {
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

    function updateAccountMP(address accountAddress) external {
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

    function getStakedBalance(address accountAddress) external view returns (uint256) {
        return accounts[accountAddress].stakedBalance;
    }

    function getPendingRewards(address accountAddress) external view returns (uint256) {
        return calculateAccountRewards(accountAddress);
    }

    function getAccount(address accountAddress) external view returns (Account memory) {
        return accounts[accountAddress];
    }
}
