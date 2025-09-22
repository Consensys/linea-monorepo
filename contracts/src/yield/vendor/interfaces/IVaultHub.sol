// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity >=0.8.0;

interface IVaultHub {
    struct VaultConnection {
        // ### 1st slot
        /// @notice address of the vault owner
        address owner;
        /// @notice maximum number of stETH shares that can be minted by vault owner
        uint96 shareLimit;
        // ### 2nd slot
        /// @notice index of the vault in the list of vaults. Indexes is guaranteed to be stable only if there was no deletions.
        /// @dev vaultIndex is always greater than 0
        uint96 vaultIndex;
        /// @notice if true, vault is disconnected and fee is not accrued
        bool pendingDisconnect;
        /// @notice share of ether that is locked on the vault as an additional reserve
        /// e.g RR=30% means that for 1stETH minted 1/(1-0.3)=1.428571428571428571 ETH is locked on the vault
        uint16 reserveRatioBP;
        /// @notice if vault's reserve decreases to this threshold, it should be force rebalanced
        uint16 forcedRebalanceThresholdBP;
        /// @notice infra fee in basis points
        uint16 infraFeeBP;
        /// @notice liquidity fee in basis points
        uint16 liquidityFeeBP;
        /// @notice reservation fee in basis points
        uint16 reservationFeeBP;
        /// @notice if true, vault owner manually paused the beacon chain deposits
        bool isBeaconDepositsManuallyPaused;
        /// 64 bits gap
    }
}