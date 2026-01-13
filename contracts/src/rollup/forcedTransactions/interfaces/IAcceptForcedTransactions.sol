// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Interface to manage forced transaction storage functions.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IAcceptForcedTransactions {
  /**
   * @notice Emitted when the forced transaction fee is set.
   * @param forcedTransactionFeeInWei The forced transaction fee in wei.
   */
  event ForcedTransactionFeeSet(uint256 forcedTransactionFeeInWei);

  /**
   * @dev Thrown when another forced transaction is expected on the computed L2 block or the previous block number is higher than the submitted one.
   */
  error ForcedTransactionExistsForBlockOrIsTooLow(uint256 blockNumber);

  /**
   * @notice Provides state fields for forced transactions.
   * @return finalizedState The last finalized state hash.
   * @return previousForcedTransactionRollingHash The previous forced transaction rolling hash.
   * @return currentFinalizedL2BlockNumber The current finalized L2 block number.
   * @return forcedTransactionFeeAmount The forced transaction fee amount.
   */
  function getRequiredForcedTransactionFields()
    external
    returns (
      bytes32 finalizedState,
      bytes32 previousForcedTransactionRollingHash,
      uint256 currentFinalizedL2BlockNumber,
      uint256 forcedTransactionFeeAmount
    );

  /**
   * @notice Stores forced transaction details required for proving feedback loop.
   * @dev FORCED_TRANSACTION_SENDER_ROLE is required to store a forced transaction.
   * @dev The forced transaction number is incremented for the next transaction post storage.
   * @dev The forced transaction fee is sent in the same transaction and the gateway will revert if it is not met.
   * @param _forcedL2BlockNumber The maximum expected L2 block number the transaction will be processed by.
   * @param _forcedTransactionRollingHash The rolling hash for all the forced transaction fields.
   * @return forcedTransactionNumber The unique forced transaction number for the transaction.
   */
  function storeForcedTransaction(
    uint256 _forcedL2BlockNumber,
    bytes32 _forcedTransactionRollingHash
  ) external payable returns (uint256 forcedTransactionNumber);

  /**
   * @notice Sets the forced transaction fee.
   * @dev Only callable by an account with the FORCED_TRANSACTION_FEE_SETTER_ROLE.
   * @param _forcedTransactionFeeInWei The forced transaction fee in wei.
   */
  function setForcedTransactionFee(uint256 _forcedTransactionFeeInWei) external;
}
