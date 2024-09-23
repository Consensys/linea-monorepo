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
    uint256 public potentialMP;
    uint256 public rewardIndex;
    uint256 public accountedRewards;
    uint256 public lastMPUpdatedTime;

    struct UserInfo {
        uint256 stakedBalance;
        uint256 userRewardIndex;
        uint256 userMP;
        uint256 userPotentialMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address account => UserInfo data) public users;

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
        updateUserMP(msg.sender);

        UserInfo storage user = users[msg.sender];
        if (user.lockUntil != 0 && user.lockUntil > block.timestamp) {
            revert StakingManager__CannotRestakeWithLockedFunds();
        }

        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            distributeRewards(msg.sender, userRewards);
        }

        bool success = STAKING_TOKEN.transferFrom(msg.sender, address(this), amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        user.stakedBalance += amount;
        totalStaked += amount;

        uint256 initialMP = amount;
        uint256 userPotentialMP = amount * MAX_MULTIPLIER;

        if (lockPeriod != 0) {
            uint256 lockMultiplier = (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR) / MAX_LOCKING_PERIOD;
            initialMP += amount * lockMultiplier / SCALE_FACTOR;
            userPotentialMP += (amount * lockMultiplier / SCALE_FACTOR);
            user.lockUntil = block.timestamp + lockPeriod;
        } else {
            user.lockUntil = 0;
        }

        user.userMP += initialMP;
        totalMP += initialMP;

        user.userPotentialMP += userPotentialMP;
        potentialMP += userPotentialMP;

        user.userRewardIndex = rewardIndex;
        user.lastMPUpdateTime = block.timestamp;
    }

    function unstake(uint256 amount) external nonReentrant {
        UserInfo storage user = users[msg.sender];
        if (amount > user.stakedBalance) {
            revert StakingManager__InsufficientBalance();
        }

        if (block.timestamp < user.lockUntil) {
            revert StakingManager__TokensAreLocked();
        }

        _updateGlobalState();
        updateUserMP(msg.sender);

        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            distributeRewards(msg.sender, userRewards);
        }

        uint256 previousStakedBalance = user.stakedBalance;
        user.stakedBalance -= amount;
        totalStaked -= amount;

        uint256 amountRatio = (amount * SCALE_FACTOR) / previousStakedBalance;
        uint256 mpToReduce = (user.userMP * amountRatio) / SCALE_FACTOR;
        uint256 potentialMPToReduce = (user.userPotentialMP * amountRatio) / SCALE_FACTOR;

        user.userMP -= mpToReduce;
        user.userPotentialMP -= potentialMPToReduce;
        totalMP -= mpToReduce;
        potentialMP -= potentialMPToReduce;

        bool success = STAKING_TOKEN.transfer(msg.sender, amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        user.userRewardIndex = rewardIndex;
    }

    function _updateGlobalState() internal {
        updateGlobalMP();
        updateRewardIndex();
    }

    function updateGlobalState() external {
        _updateGlobalState();
    }

    function updateGlobalMP() internal {
        if (potentialMP == 0) {
            lastMPUpdatedTime = block.timestamp;
            return;
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        if (timeDiff == 0) {
            return;
        }

        uint256 accruedMP = (timeDiff * totalStaked * MP_RATE_PER_YEAR) / (365 days * SCALE_FACTOR);
        if (accruedMP > potentialMP) {
            accruedMP = potentialMP;
        }

        // Adjust rewardIndex before updating totalMP
        uint256 previousTotalWeight = totalStaked + totalMP;
        totalMP += accruedMP;
        uint256 newTotalWeight = totalStaked + totalMP;

        if (previousTotalWeight != 0 && newTotalWeight != previousTotalWeight) {
            rewardIndex = (rewardIndex * previousTotalWeight) / newTotalWeight;
        }

        potentialMP -= accruedMP;
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

    function updateUserMP(address userAddress) internal {
        UserInfo storage user = users[userAddress];

        if (user.userPotentialMP == 0 || user.stakedBalance == 0) {
            user.lastMPUpdateTime = block.timestamp;
            return;
        }

        uint256 timeDiff = block.timestamp - user.lastMPUpdateTime;
        if (timeDiff == 0) {
            return;
        }

        uint256 accruedMP = (timeDiff * user.stakedBalance * MP_RATE_PER_YEAR) / (365 days * SCALE_FACTOR);

        if (accruedMP > user.userPotentialMP) {
            accruedMP = user.userPotentialMP;
        }

        user.userPotentialMP -= accruedMP;
        user.userMP += accruedMP;

        user.lastMPUpdateTime = block.timestamp;
    }

    function calculateUserRewards(address userAddress) public view returns (uint256) {
        UserInfo storage user = users[userAddress];
        uint256 userWeight = user.stakedBalance + user.userMP;
        uint256 deltaRewardIndex = rewardIndex - user.userRewardIndex;
        return (userWeight * deltaRewardIndex) / SCALE_FACTOR;
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

    function getStakedBalance(address userAddress) external view returns (uint256) {
        return users[userAddress].stakedBalance;
    }

    function getPendingRewards(address userAddress) external view returns (uint256) {
        return calculateUserRewards(userAddress);
    }

    function getUserInfo(address userAddress) external view returns (UserInfo memory) {
        return users[userAddress];
    }
}
