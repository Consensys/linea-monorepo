// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract RewardsStreamer is ReentrancyGuard {
    error StakingManager__AmountCannotBeZero();
    error StakingManager__TransferFailed();
    error StakingManager__InsufficientBalance();

    IERC20 public immutable stakingToken;
    IERC20 public immutable rewardToken;

    uint256 public constant SCALE_FACTOR = 1e18;

    uint256 public totalStaked;
    uint256 public rewardIndex;
    uint256 public accountedRewards;

    struct UserInfo {
        uint256 stakedBalance;
        uint256 userRewardIndex;
    }

    mapping(address => UserInfo) public users;

    constructor(address _stakingToken, address _rewardToken) {
        stakingToken = IERC20(_stakingToken);
        rewardToken = IERC20(_rewardToken);
    }

    function stake(uint256 amount) external nonReentrant {
        if (amount == 0) {
            revert StakingManager__AmountCannotBeZero();
        }

        updateRewardIndex();

        UserInfo storage user = users[msg.sender];
        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            distributeRewards(msg.sender, userRewards);
        }

        bool success = stakingToken.transferFrom(msg.sender, address(this), amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        user.stakedBalance += amount;
        totalStaked += amount;
        user.userRewardIndex = rewardIndex;
    }

    function unstake(uint256 amount) external nonReentrant {
        UserInfo storage user = users[msg.sender];
        if (amount > user.stakedBalance) {
            revert StakingManager__InsufficientBalance();
        }

        updateRewardIndex();

        uint256 userRewards = calculateUserRewards(msg.sender);
        if (userRewards > 0) {
            distributeRewards(msg.sender, userRewards);
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

    // send the rewards and updates accountedRewards
    function distributeRewards(address to, uint256 amount) internal {
        uint256 rewardBalance = rewardToken.balanceOf(address(this));
        // If amount is higher than the contract's balance (for rounding error), transfer the balance.
        if (amount > rewardBalance) {
            amount = rewardBalance;
        }

        bool success = rewardToken.transfer(to, amount);
        if (!success) {
            revert StakingManager__TransferFailed();
        }

        accountedRewards -= amount;
    }

    function getUserInfo(address userAddress) public view returns (UserInfo memory) {
        return users[userAddress];
    }
}
