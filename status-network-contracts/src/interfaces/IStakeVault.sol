// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { IStakeManagerProxy } from "./IStakeManagerProxy.sol";

interface IStakeVault {
    function owner() external view returns (address);
    function stakeManager() external view returns (IStakeManagerProxy);
    function register() external;
    function lockUntil() external view returns (uint256);
    function depositedBalance() external view returns (uint256);
    function migrateFromVault(uint256 newLockUntil, uint256 newDepositedBalance) external;
}
