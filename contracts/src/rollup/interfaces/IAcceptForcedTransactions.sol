// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

/**
 * @title Interface to manage forced transaction storage.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IAcceptForcedTransactions {
  /**
   * @dev Thrown when another forced transaction is expected on the computed L2 block.
   */
  error ForcedTransactionExistsForBlock(uint256 blockNumber);

  /**
   * @dev Thrown when trying to overwrite an existing forced transaction.
   */
  error ForcedTransactionExistsForTransactionNumber(uint256 forcedTransactionNumber);

  /**
   * @notice Provides fields for forced transaction.
   * @return finalizedState The last finalized state hash.
   * @return forcedTransactionNumber The forced transaction number to use.
   * @return previousForcedTransactionRollingHash The previous forced transaction rolling hash.
   * @return currentFinalizedL2BlockNumber The current finalized L2 block number.
   */
  function getNextForcedTransactionFields()
    external
    returns (
      bytes32 finalizedState,
      uint256 forcedTransactionNumber,
      bytes32 previousForcedTransactionRollingHash,
      uint256 currentFinalizedL2BlockNumber
    );

  /**
   * @notice Stores forced transaction details required for proving feedback loop.
   * @dev The forced transaction number is incremented for the next transaction post storage.
   * @param _forcedTransactionNumber The forced transaction number.
   * @param _forcedL2BlockNumber The maximum expected L2 block number the transaction will be processed by.
   * @param _forcedTransactionRollingHash The rolling hash for all the forced transaction fields.
   */
  function storeForcedTransaction(
    uint256 _forcedTransactionNumber,
    uint256 _forcedL2BlockNumber,
    bytes32 _forcedTransactionRollingHash
  ) external;
}
