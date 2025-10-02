// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

interface IRollupFeeVault {
  /**
   * @dev Thrown when a parameter is the zero address.
   */
  error ZeroAddressNotAllowed();

  /**
   * @dev Thrown when the operating costs transfer fails.
   */
  error OperatingCostsTransferFailed();

  /**
   * @dev Thrown when the burn of ETH fails.
   */
  error EthBurnFailed();

  /**
   * @dev Thrown when the timestamps are not in sequence.
   */
  error TimestampsNotInSequence();

  /**
   * @dev Thrown when the end timestamp is not greater than the start timestamp.
   */
  error EndTimestampMustBeGreaterThanStartTimestamp();

  /**
   * @dev Thrown when the operating costs are zero.
   */
  error ZeroOperatingCosts();

  /**
   * @dev Thrown when the contract balance is insufficient.
   */
  error InsufficientBalance();

  /**
   * @dev Emitted when an invoice is processed.
   * @dev If amountRequested < amountPaid, the difference is previous unpaid invoice amount.
   * @param startTimestamp The start timestamp of the invoicing period.
   * @param endTimestamp The end timestamp of the invoicing period.
   * @param amountPaid The amount that was paid.
   * @param amountRequested The amount that was requested.
   */
  event InvoiceProcessed(
    uint256 indexed startTimestamp,
    uint256 indexed endTimestamp,
    uint256 amountPaid,
    uint256 amountRequested
  );

  /**
   * @dev Emitted when ETH is burned, swapped, and bridged.
   * @param ethBurnt The amount of ETH that was burned.
   * @param lineaTokensBridged The amount of LINEA tokens that were bridged.
   */
  event EthBurntSwappedAndBridged(uint256 ethBurnt, uint256 lineaTokensBridged);

  /**
   * @dev Emitted when the L1 burner contract address is updated.
   * @param newL1BurnerContract The new L1 burner contract address.
   */
  event L1BurnerContractUpdated(address newL1BurnerContract);

  /**
   * @dev Emitted when the DEX contract address is updated.
   * @param newDex The new DEX contract address.
   */
  event DexUpdated(address newDex);

  /**
   * @dev Emitted when ETH is received.
   * @param amount The amount of ETH received.
   */
  event EthReceived(uint256 amount);

  /**
   * @dev Emitted when the operating costs are updated.
   * @param newOperatingCosts The new operating costs value.
   */
  event OperatingCostsUpdated(uint256 newOperatingCosts);
}
