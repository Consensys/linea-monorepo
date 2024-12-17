// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IRewardProvider } from "../../src/interfaces/IRewardProvider.sol";

contract XPProviderMock is IRewardProvider {
    mapping(address => uint256) public userXPShare;

    uint256 public totalXPShares;

    function setUserXPShare(address user, uint256 xp) external {
        userXPShare[user] = xp;
    }

    function setTotalXPShares(uint256 xp) external {
        totalXPShares = xp;
    }

    function rewardsBalanceOf(address) external pure override returns (uint256) {
        revert("Not implemented");
    }

    function rewardsBalanceOfUser(address account) external view override returns (uint256) {
        return userXPShare[account];
    }

    function totalRewardsSupply() external view override returns (uint256) {
        return totalXPShares;
    }
}
