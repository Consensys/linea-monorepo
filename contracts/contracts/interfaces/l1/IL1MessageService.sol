// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.26;

/**
 * @title L1 Message Service interface for pre-existing functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */

interface IL1MessageService {
  /**
   * @param proof The proof array related to the claimed message.
   * @param messageNumber The message number of the claimed message.
   * @param leafIndex The leaf index related to the merkle proof of the message.
   * @param from The address of the original sender.
   * @param to The address the message is intended for.
   * @param fee The fee being paid for the message delivery.
   * @param value The value to be transferred to the destination address.
   * @param feeRecipient The recipient for the fee.
   * @param merkleRoot The merkle root of the claimed message.
   * @param data The calldata to pass to the recipient.
   */
  struct ClaimMessageWithProofParams {
    bytes32[] proof;
    uint256 messageNumber;
    uint32 leafIndex;
    address from;
    address to;
    uint256 fee;
    uint256 value;
    address payable feeRecipient;
    bytes32 merkleRoot;
    bytes data;
  }

  /**
   * @notice Emitted when initializing Linea Rollup contract with a system migration block.
   */
  event SystemMigrationBlockInitialized(uint256 systemMigrationBlock);
  /**
   * @dev Thrown when L2 merkle root does not exist.
   */
  error L2MerkleRootDoesNotExist();

  /**
   * @dev Thrown when the merkle proof is invalid.
   */
  error InvalidMerkleProof();

  /**
   * @dev Thrown when merkle depth doesn't match proof length.
   */
  error ProofLengthDifferentThanMerkleDepth(uint256 actual, uint256 expected);
}
