// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { IL1MessageService } from "../../messaging/l1/interfaces/IL1MessageService.sol";

/**
 * @title Native yield extension module for the Linea L1MessageService.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface ILineaNativeYieldExtension {
  /**
   * @notice Emitted when ETH send from an authorized funder.
   * @param funder Address which sent ETH.
   * @param amount Donation amount.
   */
  event FundingReceived(address indexed funder, uint256 amount);

  /**
   * @notice Emitted when a permissionless ETH donation is received.
   * @param donor Address which sent the donation.
   * @param amount Donation amount.
   */
  event PermissionlessDonationReceived(address indexed donor, uint256 amount);

  /**
   * @notice Emitted when a YieldManager address is set.
   * @param oldYieldManagerAddress The previous YieldManager address.
   * @param newYieldManagerAddress The new YieldManager address.
   * @param caller Address which set the YieldManager address.
   */
  event YieldManagerChanged(
    address indexed oldYieldManagerAddress,
    address indexed newYieldManagerAddress,
    address indexed caller
  );

  /**
   * @notice Emitted when the L2YieldRecipient address is set.
   * @param oldL2YieldRecipientAddress The previous L2YieldRecipient address.
   * @param newL2YieldRecipientAddress The new L2YieldRecipient address.
   * @param caller Address which set the L2YieldRecipient address.
   */
  event L2YieldRecipientChanged(
    address indexed oldL2YieldRecipientAddress,
    address indexed newL2YieldRecipientAddress,
    address indexed caller
  );

  /**
   * @dev Thrown when the caller is not the YieldManager.
   */
  error CallerIsNotYieldManager();

  error LSTWithdrawalRequiresDeficit();

  /**
   * @notice Report native yield earned for L2 distribution by emitting a synthetic `MessageSent` event.
   * @dev Callable only by the registered YieldManager.
   * @param _amount The net earned yield.
   */
  function reportNativeYield(uint256 _amount, address _l2YieldRecipient) external;

  function claimMessageWithProofAndWithdrawLST(IL1MessageService.ClaimMessageWithProofParams calldata _params, address _yieldProvider) external;

  function isWithdrawLSTAllowed() external view returns (bool);

  /**
   * @notice Transfer ETH to the registered YieldManager.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Enforces that, after transfer, the L1MessageService balance remains ≥ the configured effective minimum reserve.
   * @param _amount Amount of ETH to transfer.
   */
  function transferFundsForNativeYield(uint256 _amount) external;

  /**
   * @notice Send ETH to this contract.
   * @dev FUNDER_ROLE is required to execute.
   */
  function fund() external payable;

  /**
   * @notice Set YieldManager address.
   * @dev YIELD_MANAGER_SETTER_ROLE is required to execute.
   * @param _newYieldManager YieldManager address.
   */
  function setYieldManager(address _newYieldManager) external;
}
