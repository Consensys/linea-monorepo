// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

// See contracts/COMPILERS.md
pragma solidity >=0.8.0;

/**
 * @title Dashboard
 * @notice This contract is a UX-layer for StakingVault and meant to be used as its owner.
 * This contract improves the vault UX by bundling all functions from the StakingVault and VaultHub
 * in this single contract. It provides administrative functions for managing the StakingVault,
 * including funding, withdrawing, minting, burning, and rebalancing operations.
 */
interface IDashboard {
  // ==================== View Functions ====================

  /**
   * @notice Returns the stETH share limit of the vault
   */
  function shareLimit() external view returns (uint256);

  /**
   * @notice Returns the number of stETH shares minted
   */
  function liabilityShares() external view returns (uint256);

  /**
   * @notice Returns the total value of the vault in ether.
   */
  function totalValue() external view returns (uint256);

  /**
   * @notice Returns the overall unsettled obligations of the vault in ether
   * @dev includes the node operator fee
   */
  function unsettledObligations() external view returns (uint256);

  /**
   * @notice Returns the amount of ether that can be instantly withdrawn from the staking vault.
   * @dev This is the amount of ether that is not locked in the StakingVault and not reserved for fees and obligations.
   */
  function withdrawableValue() external view returns (uint256);

  // ==================== Vault Management Functions ====================

  /**
   * @notice Disconnects the underlying StakingVault from the hub and passing its ownership to Dashboard.
   *         After receiving the final report, one can call reconnectToVaultHub() to reconnect to the hub
   *         or abandonDashboard() to transfer the ownership to a new owner.
   */
  function voluntaryDisconnect() external;

  /**
   * @notice Accepts the ownership over the StakingVault transferred from VaultHub on disconnect
   * and immediately transfers it to a new pending owner. This new owner will have to accept the ownership
   * on the StakingVault contract.
   * @param _newOwner The address to transfer the StakingVault ownership to.
   */
  function abandonDashboard(address _newOwner) external;

  /**
   * @notice Accepts the ownership over the StakingVault and connects to VaultHub. Can be called to reconnect
   *         to the hub after voluntaryDisconnect()
   */
  function reconnectToVaultHub() external;

  /**
   * @notice Funds the staking vault with ether
   */
  function fund() external payable;

  /**
   * @notice Withdraws ether from the staking vault to a recipient
   * @param _recipient Address of the recipient
   * @param _ether Amount of ether to withdraw
   */
  function withdraw(address _recipient, uint256 _ether) external;

  /**
   * @notice Mints stETH tokens backed by the vault to the recipient.
   * !NB: this will revert with`VaultHub.ZeroArgument("_amountOfShares")` if the amount of stETH is less than 1 share
   * @param _recipient Address of the recipient
   * @param _amountOfStETH Amount of stETH to mint
   */
  function mintStETH(address _recipient, uint256 _amountOfStETH) external payable;

  /**
   * @notice Rebalances StakingVault by withdrawing ether to VaultHub corresponding to shares amount provided
   * @param _shares amount of shares to rebalance
   */
  function rebalanceVaultWithShares(uint256 _shares) external;

  /**
   * @notice Rebalances the vault by transferring ether given the shares amount
   * @param _ether amount of ether to rebalance
   */
  function rebalanceVaultWithEther(uint256 _ether) external payable;

  /**
   * @notice Pauses beacon chain deposits on the StakingVault.
   */
  function pauseBeaconChainDeposits() external;

  /**
   * @notice Resumes beacon chain deposits on the StakingVault.
   */
  function resumeBeaconChainDeposits() external;

  /**
   * @notice Initiates a withdrawal from validator(s) on the beacon chain using EIP-7002 triggerable withdrawals
   *         Both partial withdrawals (disabled for if vault is unhealthy) and full validator exits are supported.
   * @param _pubkeys Concatenated validator external keys (48 bytes each).
   * @param _amounts Withdrawal amounts in wei for each validator key and must match _pubkeys length.
   *         Set amount to 0 for a full validator exit.
   *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation
   * @param _refundRecipient Address to receive any fee refunds, if zero, refunds go to msg.sender.
   * @dev    A withdrawal fee must be paid via msg.value.
   *         Use `StakingVault.calculateValidatorWithdrawalFee()` to determine the required fee for the current block.
   */
  function triggerValidatorWithdrawals(
    bytes calldata _pubkeys,
    uint64[] calldata _amounts,
    address _refundRecipient
  ) external payable;

  function nodeOperatorDisbursableFee() external view returns (uint256);

  function disburseNodeOperatorFee() external;
}
