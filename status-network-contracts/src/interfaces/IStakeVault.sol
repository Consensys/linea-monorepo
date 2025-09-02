// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { IStakeManagerProxy } from "./IStakeManagerProxy.sol";

interface IStakeVault {
    /// @notice Struct to hold migration data
    struct MigrationData {
        /// @notice Timestamp until the funds are locked
        uint256 lockUntil;
        /// @notice Total amount deposited into the vault
        uint256 depositedBalance;
    }

    function owner() external view returns (address);
    function stakeManager() external view returns (IStakeManagerProxy);
    function register() external;
    function lockUntil() external view returns (uint256);
    function depositedBalance() external view returns (uint256);
    function migrateFromVault(MigrationData calldata data) external;
}
