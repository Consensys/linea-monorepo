// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { IStakeManagerProxy } from "./IStakeManagerProxy.sol";

/**
 * @title IStakeVault
 * @notice Interface for a stake vault contract that interacts with a stake manager proxy.
 */
interface IStakeVault {
    /// @notice Struct to hold migration data
    struct MigrationData {
        /// @notice Timestamp until the funds are locked
        uint256 lockUntil;
        /// @notice Total amount deposited into the vault
        uint256 depositedBalance;
    }

    /**
     * @notice The owner of the vault
     * @return address The address of the owner
     */
    function owner() external view returns (address);

    /**
     * @notice The stake manager proxy associated with the vault
     * @return IStakeManagerProxy The stake manager proxy instance
     */
    function stakeManager() external view returns (IStakeManagerProxy);

    /**
     * @notice Registers the vault with the stake manager
     */
    function register() external;

    /**
     * @notice The timestamp until which the funds are locked
     * @return uint256 The timestamp until which the funds are locked
     */
    function lockUntil() external view returns (uint256);

    /**
     * @notice The total amount deposited into the vault
     * @return uint256 The total amount deposited into the vault
     */
    function depositedBalance() external view returns (uint256);

    /**
     * @notice Migrates funds from a previous vault
     * @param data The migration data containing lockUntil and depositedBalance
     */
    function migrateFromVault(MigrationData calldata data) external;
}
