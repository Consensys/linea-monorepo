// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IRewardDistributor } from "../../src/interfaces/IRewardDistributor.sol";

contract KarmaDistributorMock is IRewardDistributor {
    // solhint-disable-next-line
    mapping(address => uint256) public userKarmaShare;

    uint256 public totalKarmaShares;

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
    function setReward(uint256, uint256) external pure override { }

    function rewardsBalanceOfAccount(address account) external view override returns (uint256) {
        return userKarmaShare[account];
    }

    function totalRewardsSupply() external view override returns (uint256) {
        return totalKarmaShares;
    }
}
