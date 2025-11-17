// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { IBlobSupportingLineaRollup } from "./interfaces/IBlobSupportingLineaRollup.sol";
import { ProvideLocalShnarf } from "./ProvideLocalShnarf.sol";
import { EfficientLeftRightKeccak } from "../libraries/EfficientLeftRightKeccak.sol";

/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract BlobSupportingLineaRollup is ProvideLocalShnarf, IBlobSupportingLineaRollup {
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
  ) external virtual whenTypeAndGeneralNotPaused(PauseType.BLOB_SUBMISSION) onlyRole(OPERATOR_ROLE) {
    _submitBlobs(_blobSubmissions, _parentShnarf, _finalBlobShnarf);
  }

  /**
   * @notice Submit one or more EIP-4844 blobs.
   * @param _blobSubmissions The data for blob submission including proofs and required polynomials.
   * @param _parentShnarf The parent shnarf used in continuity checks as it includes the parentStateRootHash in its computation.
   * @param _finalBlobShnarf The expected final shnarf post computation of all the blob shnarfs.
   */
  function _submitBlobs(
    BlobSubmission[] calldata _blobSubmissions,
    bytes32 _parentShnarf,
    bytes32 _finalBlobShnarf
  ) internal virtual {
    if (_blobSubmissions.length == 0) {
      revert BlobSubmissionDataIsMissing();
    }

    if (blobhash(_blobSubmissions.length) != EMPTY_HASH) {
      revert BlobSubmissionDataEmpty(_blobSubmissions.length);
    }

    if (_blobShnarfExists[_parentShnarf] == 0) {
      revert ParentBlobNotSubmitted(_parentShnarf);
    }

    /**
     * @dev validate we haven't submitted the last shnarf. There is a final check at the end of the function verifying,
     * that _finalBlobShnarf was computed correctly.
     * Note: As only the last shnarf is stored, we don't need to validate shnarfs,
     * computed for any previous blobs in the submission (if multiple are submitted).
     */
    if (_blobShnarfExists[_finalBlobShnarf] != 0) {
      revert DataAlreadySubmitted(_finalBlobShnarf);
    }

    bytes32 currentDataEvaluationPoint;
    bytes32 currentDataHash;

    /// @dev Assigning in memory saves a lot of gas vs. calldata reading.
    BlobSubmission memory blobSubmission;

    bytes32 computedShnarf = _parentShnarf;

    for (uint256 i; i < _blobSubmissions.length; i++) {
      blobSubmission = _blobSubmissions[i];

      currentDataHash = blobhash(i);

      if (currentDataHash == EMPTY_HASH) {
        revert EmptyBlobDataAtIndex(i);
      }

      bytes32 snarkHash = blobSubmission.snarkHash;

      currentDataEvaluationPoint = EfficientLeftRightKeccak._efficientKeccak(snarkHash, currentDataHash);

      _verifyPointEvaluation(
        currentDataHash,
        uint256(currentDataEvaluationPoint),
        blobSubmission.dataEvaluationClaim,
        blobSubmission.kzgCommitment,
        blobSubmission.kzgProof
      );

      computedShnarf = _computeShnarf(
        computedShnarf,
        snarkHash,
        blobSubmission.finalStateRootHash,
        currentDataEvaluationPoint,
        bytes32(blobSubmission.dataEvaluationClaim)
      );
    }

    if (_finalBlobShnarf != computedShnarf) {
      revert FinalShnarfWrong(_finalBlobShnarf, computedShnarf);
    }

    /// @dev use the last shnarf as the submission to store as technically it becomes the next parent shnarf.
    _blobShnarfExists[computedShnarf] = SHNARF_EXISTS_DEFAULT_VALUE;

    emit DataSubmittedV3(_parentShnarf, computedShnarf, blobSubmission.finalStateRootHash);
  }
}
