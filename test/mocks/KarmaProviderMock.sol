// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IRewardProvider } from "../../src/interfaces/IRewardProvider.sol";

contract KarmaProviderMock is IRewardProvider {
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
        revert("Not implemented");
    }

    function rewardsBalanceOfAccount(address account) external view override returns (uint256) {
        return userKarmaShare[account];
    }

    function totalRewardsSupply() external view override returns (uint256) {
        return totalKarmaShares;
    }
}
