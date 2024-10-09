// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IXPProvider } from "../../src/interfaces/IXPProvider.sol";

contract XPProviderMock is IXPProvider {
    mapping(address => uint256) public userXPShare;

    uint256 public totalXPShares;

    function setUserXPShare(address user, uint256 xp) external {
        userXPShare[user] = xp;
    }

    function setTotalXPShares(uint256 xp) external {
        totalXPShares = xp;
    }

    function getUserXPShare(address account) external view override returns (uint256) {
        return userXPShare[account];
    }

    function getTotalXPShares() external view override returns (uint256) {
        return totalXPShares;
    }
}
