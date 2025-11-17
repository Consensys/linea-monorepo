// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;
interface IBlobSupportingLineaRollup {
  /**
   * @notice Data structure for compressed blob data submission.
   * @dev submissionData The supporting data for blob data submission excluding the compressed data.
   * @dev dataEvaluationClaim The data evaluation claim.
   * @dev kzgCommitment The blob KZG commitment.
   * @dev kzgProof The blob KZG point proof.
   */
  struct BlobSubmission {
    uint256 dataEvaluationClaim;
    bytes kzgCommitment;
    bytes kzgProof;
    bytes32 finalStateRootHash;
    bytes32 snarkHash;
  }

  /**
   * @notice Submit one or more EIP-4844 blobs.
   * @dev OPERATOR_ROLE is required to execute.
   * @dev This should be a blob carrying transaction.
   * @param _blobSubmissions The data for blob submission including proofs and required polynomials.
   * @param _parentShnarf The parent shnarf used in continuity checks as it includes the parentStateRootHash in its computation.
   * @param _finalBlobShnarf The expected final shnarf post computation of all the blob shnarfs.
   */
  function submitBlobs(
    BlobSubmission[] calldata _blobSubmissions,
    bytes32 _parentShnarf,
    bytes32 _finalBlobShnarf
  ) external;
}
