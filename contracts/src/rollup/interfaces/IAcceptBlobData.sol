// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

/**
 * @title IAcceptBlobData interface for shared data accepting errors and events.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IAcceptBlobData {
  /**
   * @dev Thrown when the current data was already submitted.
   */
  error DataAlreadySubmitted(bytes32 currentDataHash);

  /**
   * @dev Thrown when a shnarf does not exist for a parent blob.
   */
  error ParentBlobNotSubmitted(bytes32 shnarf);

  /**
   * @dev Thrown when the computed shnarf does not match what is expected.
   */
  error FinalShnarfWrong(bytes32 expected, bytes32 value);

  /**
   * @notice Emitted when compressed data is being submitted and verified succesfully on L1.
   * @dev The block range is indexed and parent shnarf included for state reconstruction simplicity.
   * @param parentShnarf The parent shnarf for the data being submitted.
   * @param shnarf The indexed shnarf for the data being submitted.
   * @param finalStateRootHash The L2 state root hash that the current blob submission ends on. NB: The last blob in the collection.
   */
  event DataSubmittedV3(bytes32 parentShnarf, bytes32 indexed shnarf, bytes32 finalStateRootHash);
}
