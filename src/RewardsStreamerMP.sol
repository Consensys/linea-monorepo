// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

// Rewards Streamer with Multiplier Points
contract RewardsStreamerMP is ReentrancyGuard {
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();
    error StakingManager__InvalidLockingPeriod();
    error StakingManager__CannotRestakeWithLockedFunds();

    IERC20 public immutable stakingToken;
    IERC20 public immutable rewardToken;

    uint256 public constant SCALE_FACTOR = 1e18;
    uint256 public constant MP_RATE_PER_YEAR = 1;

    uint256 public constant MIN_LOCKING_PERIOD = 90 days;
    uint256 public constant MAX_LOCKING_PERIOD = (365 days) * 4;
    uint256 public constant MAX_MULTIPLIER = 4;

    uint256 public totalStaked;
    uint256 public rewardIndex;
    uint256 public accountedRewards;

    uint256 totalMP;
    uint256 potentialMP;
    uint256 lastMPUpdatedTime;

    struct UserInfo {
        uint256 stakedBalance;
        uint256 userRewardIndex;
        uint256 userMP;
        uint256 userPotentialMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address => UserInfo) public users;

    constructor(address _stakingToken, address _rewardToken) {
        stakingToken = IERC20(_stakingToken);
        rewardToken = IERC20(_rewardToken);
    }

    function stake(uint256 amount, uint256 lockPeriod) external nonReentrant {
        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        if (lockPeriod != 0 && (lockPeriod < MIN_LOCKING_PERIOD || lockPeriod > MAX_LOCKING_PERIOD)) {
            revert StakingManager__InvalidLockingPeriod();
        }

        updateRewardIndex();
        updateGlobalMP();
        updateUserMP(msg.sender);

        UserInfo storage user = users[msg.sender];
        if (user.lockUntil != 0 && user.lockUntil > block.timestamp) {
            revert StakingManager__CannotRestakeWithLockedFunds();
        }

        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            safeRewardTransfer(msg.sender, userRewards);
        }

        bool success = stakingToken.transferFrom(msg.sender, address(this), amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        user.stakedBalance += amount;
        totalStaked += amount;
        user.userRewardIndex = rewardIndex;

        // TODO: revisit initialMP calculation
        uint256 initialMP;
        uint256 userPotentialMP;
        if (lockPeriod == 0) {
            initialMP = amount;
            potentialMP = amount * 4;
        } else {
            uint256 maxAmount = amount * MAX_MULTIPLIER;
            initialMP = amount + (lockPeriod * maxAmount) / MAX_LOCKING_PERIOD;
            // TODO: this needs to be proportional,
            // not 8 only because the funds are locked.
            potentialMP = amount * 8;
        }

        user.userMP += initialMP;
        totalMP += initialMP;

        user.userPotentialMP = userPotentialMP;
        potentialMP += userPotentialMP;

        user.lastMPUpdateTime = block.timestamp;
    }

    function unstake(uint256 amount) external nonReentrant {
        UserInfo storage user = users[msg.sender];
        if (amount > user.stakedBalance) {
            revert StakingManager__InsufficientBalance();
        }

        updateRewardIndex();

        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            safeRewardTransfer(msg.sender, userRewards);
        }

        user.stakedBalance -= amount;
        totalStaked -= amount;

        bool success = stakingToken.transfer(msg.sender, amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        user.userRewardIndex = rewardIndex;
    }

    function updateRewardIndex() public {
        if (totalStaked == 0) {
            return;
        }

        uint256 rewardBalance = rewardToken.balanceOf(address(this));
        uint256 newRewards = rewardBalance > accountedRewards ? rewardBalance - accountedRewards : 0;

        if (newRewards > 0) {
            rewardIndex += (newRewards * SCALE_FACTOR) / totalStaked;
            accountedRewards += newRewards;
        }
    }

    function getStakedBalance(address userAddress) public view returns (uint256) {
        return users[userAddress].stakedBalance;
    }

    function getPendingRewards(address userAddress) public view returns (uint256) {
        return calculateUserRewards(userAddress);
    }

    function calculateUserRewards(address userAddress) public view returns (uint256) {
        UserInfo storage user = users[userAddress];
        return (user.stakedBalance * (rewardIndex - user.userRewardIndex)) / SCALE_FACTOR;
    }

    function safeRewardTransfer(address to, uint256 amount) internal {
        uint256 rewardBalance = rewardToken.balanceOf(address(this));
        // If amount is higher than the contract's balance (for rounding error), transfer the balance.
        if (amount > rewardBalance) {
            bool success = rewardToken.transfer(to, rewardBalance);
            if (!success) {
                revert StakingManager__TransferFailed();
            }
        } else {
            bool success = rewardToken.transfer(to, amount);
            if (!success) {
                revert StakingManager__TransferFailed();
            }
        }
    }

    function updateGlobalMP() internal {
        if (potentialMP == 0) {
            return;
        }

        uint256 currentTime = block.timestamp;
        uint256 timeDiff = currentTime - lastMPUpdatedTime;
        uint256 accruedMP = (timeDiff * totalStaked * MP_RATE_PER_YEAR) / (365 days);

        if (accruedMP > potentialMP) {
            accruedMP = potentialMP;
        }

        potentialMP -= accruedMP;
        totalMP += accruedMP;

        lastMPUpdatedTime = currentTime;
    }

    function updateUserMP(address userAddress) internal {
        UserInfo storage user = users[userAddress];
        if (user.userMP == 0) {
            return;
        }

        uint256 timeDiff = block.timestamp - user.lastMPUpdateTime;
        uint256 accruedMP = (timeDiff * user.stakedBalance * MP_RATE_PER_YEAR) / (365 days);

        if (accruedMP > user.userPotentialMP) {
            accruedMP = user.userPotentialMP;
        }

        user.userPotentialMP -= accruedMP;
        user.userMP += accruedMP;
        user.lastMPUpdateTime = block.timestamp;
    }
}
