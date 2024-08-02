// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ZkEvm rollup interface for pre-existing functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IZkEvmV2 {
  /**
   * @notice Emitted when a L2 blocks have been finalized on L1.
   * @param lastBlockFinalized The indexed L2 block number the finalization is up until.
   * @param startingRootHash The state root hash the finalization started from. This is the last finalized block's state root.
   * @param finalRootHash The L2 block state root hash the finalization ended on.
   */
  event BlocksVerificationDone(uint256 indexed lastBlockFinalized, bytes32 startingRootHash, bytes32 finalRootHash);

  /**
   * @dev Thrown when the starting rootHash does not match the existing state
   */
  error StartingRootHashDoesNotMatch();

  /**
   * @dev Thrown when zk proof is empty bytes
   */
  error ProofIsEmpty();

  /**
   * @dev Thrown when zk proof type is invalid
   */
  error InvalidProofType();

  /**
   * @dev Thrown when zk proof is invalid
   */
  error InvalidProof();
}
