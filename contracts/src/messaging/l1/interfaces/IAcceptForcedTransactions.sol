// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

interface IAcceptForcedTransactions {
  error ForcedTransactionExistsForBlock(uint256 blockNumber);
  function currentFinalizedState() external returns (bytes32 lastFinalizedState);

  function getLineaRollupProvidedFields()
    external
    returns (uint256 forcedTransactionNumber, bytes32 forcedTransactionRollingHash, uint256 currentL2BlockNumber);

  function storeForcedTransaction(
    uint256 _forcedTransactionNumber,
    uint256 _forcedL2BlockNumber,
    bytes32 _forcedTransactionRollingHash
  ) external;
}
