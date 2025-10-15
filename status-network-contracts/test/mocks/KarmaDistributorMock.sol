// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IRewardDistributor } from "../../src/interfaces/IRewardDistributor.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract KarmaDistributorMock is IRewardDistributor {
    // solhint-disable-next-line
    mapping(address => uint256) public userKarmaShare;

    uint256 public totalKarmaShares;

    uint256 public totalRewardsSupply;

    IERC20 rewardToken;

    constructor(IERC20 _rewardToken) {
        rewardToken = _rewardToken;
    }

    function setUserKarmaShare(address user, uint256 karma) external {
        userKarmaShare[user] = karma;
    }

    function setTotalKarmaShares(uint256 karma) external {
        totalKarmaShares = karma;
    }

    function rewardsBalanceOf(address) external pure override returns (uint256) {
        // solhint-disable-next-line
        revert("Not implemented");
    }

    // solhint-disable-next-line
    function setReward(uint256 amount, uint256) external override {
        totalRewardsSupply = amount;
    }

    function rewardsBalanceOfAccount(address account) external view override returns (uint256) {
        return userKarmaShare[account];
    }

    function redeemRewards(address account) external override returns (uint256) {
        uint256 amount = userKarmaShare[account];
        rewardToken.transfer(account, amount);
        return amount;
    }
}
