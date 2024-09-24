// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.26;

import { IPauseManager } from "../../interfaces/IPauseManager.sol";
import { IPermissionsManager } from "../../interfaces/IPermissionsManager.sol";

/**
 * @title LineaRollup interface for current functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface ILineaRollup {
  /**
   * @notice Initialization data structure for the LineaRollup contract.
   * @param initialStateRootHash The initial hash at migration used for proof verification.
   * @param initialL2BlockNumber The initial block number at migration.
   * @param genesisTimestamp The L2 genesis timestamp for first finalization.
   * @param defaultVerifier The default verifier for rollup proofs.
   * @param rateLimitPeriodInSeconds The period in which withdrawal amounts and fees will be accumulated.
   * @param rateLimitAmountInWei The limit allowed for withdrawing in the rate limit period.
   * @param roleAddresses The list of role addresses.
   * @param pauseTypeRoles The list of pause type roles.
   * @param unpauseTypeRoles The list of unpause type roles.
   */
  struct InitializationData {
    bytes32 initialStateRootHash;
    uint256 initialL2BlockNumber;
    uint256 genesisTimestamp;
    address defaultVerifier;
    uint256 rateLimitPeriodInSeconds;
    uint256 rateLimitAmountInWei;
    IPermissionsManager.RoleAddress[] roleAddresses;
    IPauseManager.PauseTypeRole[] pauseTypeRoles;
    IPauseManager.PauseTypeRole[] unpauseTypeRoles;
  }

  /**
   * @notice Supporting data for compressed calldata submission including compressed data.
   * @dev finalStateRootHash is used to set state root at the end of the data.
   * @dev firstBlockInData is the first block that is included in the data submitted.
   * @dev finalBlockInData is the last block that is included in the data submitted.
   * @dev snarkHash is the computed hash for compressed data (using a SNARK-friendly hash function) that aggregates per data submission to be used in public input.
   * @dev compressedData is the compressed transaction data. It contains ordered data for each L2 block - l2Timestamps, the encoded txData.
   */
  struct SubmissionDataV2 {
    bytes32 finalStateRootHash;
    uint256 firstBlockInData;
    uint256 finalBlockInData;
    bytes32 snarkHash;
    bytes compressedData;
  }

  /**
   * @notice Supporting data for compressed blob data submission.
   * @dev finalStateRootHash is used to set state root at the end of the data.
   * @dev firstBlockInData is the first block that is included in the data submitted.
   * @dev finalBlockInData is the last block that is included in the data submitted.
   * @dev snarkHash is the computed hash for compressed data (using a SNARK-friendly hash function) that aggregates per data submission to be used in public input.
   */
  struct SupportingSubmissionDataV2 {
    bytes32 finalStateRootHash;
    uint256 firstBlockInData;
    uint256 finalBlockInData;
    bytes32 snarkHash;
  }

  /**
   * @notice Shnarf data for validating a shnarf.
   * @dev parentShnarf is the parent computed shnarf.
   * @dev snarkHash is the computed hash for compressed data (using a SNARK-friendly hash function) that aggregates per data submission to be used in public input.
   * @dev finalStateRootHash is the final state root hash.
   * @dev dataEvaluationPoint is the data evaluation point.
   * @dev dataEvaluationClaim is the data evaluation claim.
   */
  struct ShnarfData {
    bytes32 parentShnarf;
    bytes32 snarkHash;
    bytes32 finalStateRootHash;
    bytes32 dataEvaluationPoint;
    bytes32 dataEvaluationClaim;
  }

  /**
   * @notice Data stucture for compressed blob data submission.
   * @dev submissionData The supporting data for blob data submission excluding the compressed data.
   * @dev dataEvaluationClaim The data evaluation claim.
   * @dev kzgCommitment The blob KZG commitment.
   * @dev kzgProof The blob KZG point proof.
   */
  struct BlobSubmissionData {
    SupportingSubmissionDataV2 submissionData;
    uint256 dataEvaluationClaim;
    bytes kzgCommitment;
    bytes kzgProof;
  }

  /**
   * @notice Supporting data for finalization with or without proof.
   * @dev NB: the dynamic sized fields are placed last on purpose for efficient keccaking on public input.
   * @dev parentStateRootHash is the expected last state root hash finalized.
   * @dev lastFinalizedShnarf is the last finalized shnarf for proof continuity checks.
   * @dev finalBlockInData is the final block finalizing until.
   * @dev shnarfData contains data about the last data submission's shnarf used in finalization.
   * @dev lastFinalizedTimestamp is the expected last finalized block's timestamp.
   * @dev finalTimestamp is the timestamp of the last block being finalized.
   * @dev lastFinalizedL1RollingHash is the last stored L2 computed rolling hash used in finalization.
   * @dev l1RollingHash is the calculated rolling hash on L2 that is expected to match L1 at l1RollingHashMessageNumber.
   * This value will be used along with the stored last finalized L2 calculated rolling hash in the public input.
   * @dev lastFinalizedL1RollingHashMessageNumber is the last stored L2 computed message number used in finalization.
   * @dev l1RollingHashMessageNumber is the calculated message number on L2 that is expected to match the existing L1 rolling hash.
   * This value will be used along with the stored last finalized L2 calculated message number in the public input.
   * @dev l2MerkleTreesDepth is the depth of all l2MerkleRoots.
   * @dev l2MerkleRoots is an array of L2 message merkle roots of depth l2MerkleTreesDepth between last finalized block and finalSubmissionData.finalBlockInData.
   * @dev l2MessagingBlocksOffsets indicates by offset from currentL2BlockNumber which L2 blocks contain MessageSent events.
   */
  struct FinalizationDataV2 {
    bytes32 parentStateRootHash;
    bytes32 lastFinalizedShnarf;
    uint256 finalBlockInData;
    ShnarfData shnarfData;
    uint256 lastFinalizedTimestamp;
    uint256 finalTimestamp;
    bytes32 lastFinalizedL1RollingHash;
    bytes32 l1RollingHash;
    uint256 lastFinalizedL1RollingHashMessageNumber;
    uint256 l1RollingHashMessageNumber;
    uint256 l2MerkleTreesDepth;
    bytes32[] l2MerkleRoots;
    bytes l2MessagingBlocksOffsets;
  }

  /**
   * @notice Emitted when a verifier is set for a particular proof type.
   * @param verifierAddress The indexed new verifier address being set.
   * @param proofType The indexed proof type/index that the verifier is mapped to.
   * @param verifierSetBy The index address who set the verifier at the mapping.
   * @param oldVerifierAddress Indicates the previous address mapped to the proof type.
   * @dev The verifier will be set by an account with the VERIFIER_SETTER_ROLE. Typically the Safe.
   * @dev The oldVerifierAddress can be the zero address.
   */
  event VerifierAddressChanged(
    address indexed verifierAddress,
    uint256 indexed proofType,
    address indexed verifierSetBy,
    address oldVerifierAddress
  );

  /**
   * @notice Emitted when compressed data is being submitted and verified succesfully on L1.
   * @param shnarf The indexed shnarf for the data being submitted.
   * @param startBlock The indexed L2 block number indicating which block the data starts from.
   * @param endBlock The indexed L2 block number indicating which block the data ends on.
   * @dev Please note, shnarf was previously dataHash and points to the shnarfFinalBlockNumbers mapping.
   */
  event DataSubmittedV2(bytes32 indexed shnarf, uint256 indexed startBlock, uint256 indexed endBlock);

  /**
   * @notice Emitted when L2 blocks have been finalized on L1.
   * @param lastBlockFinalized The indexed last L2 block that is finalized in the finalization.
   * @param startingRootHash The indexed initial (also last finalized) L2 state root hash that the finalization is from.
   * @param finalRootHash The indexed L2 state root hash that the current finalization is up until.
   * @param withProof Indicates if the finalization is proven or not.
   */
  event DataFinalized(
    uint256 indexed lastBlockFinalized,
    bytes32 indexed startingRootHash,
    bytes32 indexed finalRootHash,
    bool withProof
  );

  /**
   * @dev Thrown when the point evaluation precompile call return data field(s) are wrong.
   */
  error PointEvaluationResponseInvalid(uint256 fieldElements, uint256 blsCurveModulus);

  /**
   * @dev Thrown when the point evaluation precompile call return data length is wrong.
   */
  error PrecompileReturnDataLengthWrong(uint256 expected, uint256 actual);

  /**
   * @dev Thrown when the point evaluation precompile call returns false.
   */
  error PointEvaluationFailed();

  /**
   * @dev Thrown when the blobhash equals to the zero hash.
   */
  error EmptyBlobData();

  /**
   * @dev Thrown when the blobhash at an index equals to the zero hash.
   */
  error EmptyBlobDataAtIndex(uint256 index);

  /**
   * @dev Thrown when the data for multiple blobs' submission has length zero.
   */
  error BlobSubmissionDataIsMissing();

  /**
   * @dev Thrown when the starting block in the data item is out of sequence with the last block number.
   */
  error DataStartingBlockDoesNotMatch(uint256 expected, uint256 actual);

  /**
   * @dev Thrown when the current data was already submitted.
   */
  error DataAlreadySubmitted(bytes32 currentDataHash);

  /**
   * @dev Thrown when the last finalized shnarf does not match the parent finalizing from.
   */
  error LastFinalizedShnarfWrong(bytes32 expected, bytes32 actual);

  /**
   * @dev Thrown when submissionData is empty.
   */
  error EmptySubmissionData();

  /**
   * @dev Thrown when finalizationData.l1RollingHash does not exist on L1 (Feedback loop).
   */
  error L1RollingHashDoesNotExistOnL1(uint256 messageNumber, bytes32 rollingHash);

  /**
   * @dev Thrown when finalization state does not match.
   */
  error FinalizationStateIncorrect(bytes32 expected, bytes32 value);

  /**
   * @dev Thrown when the first block is greater than final block in submission data.
   */
  error FirstBlockGreaterThanFinalBlock(uint256 firstBlockNumber, uint256 finalBlockNumber);

  /**
   * @dev Thrown when the first block in data is less than or equal to the last finalized block during data submission.
   */
  error FirstBlockLessThanOrEqualToLastFinalizedBlock(uint256 firstBlockNumber, uint256 lastFinalizedBlock);

  /**
   * @dev Thrown when the final block number in finalization data is less than or equal to the last finalized block during finalization.
   */
  error FinalBlockNumberLessThanOrEqualToLastFinalizedBlock(uint256 finalBlockNumber, uint256 lastFinalizedBlock);

  /**
   * @dev Thrown when the final block state equals the zero hash during finalization.
   */
  error FinalBlockStateEqualsZeroHash();

  /**
   * @dev Thrown when final l2 block timestamp higher than current block.timestamp during finalization.
   */
  error FinalizationInTheFuture(uint256 l2BlockTimestamp, uint256 currentBlockTimestamp);

  /**
   * @dev Thrown when a rolling hash is provided without a corresponding message number.
   */
  error MissingMessageNumberForRollingHash(bytes32 rollingHash);

  /**
   * @dev Thrown when a message number is provided without a corresponding rolling hash.
   */
  error MissingRollingHashForMessageNumber(uint256 messageNumber);

  /**
   * @dev Thrown when the first byte is not zero.
   * @dev This is used explicitly with the four bytes in assembly 0x729eebce.
   */
  error FirstByteIsNotZero();

  /**
   * @dev Thrown when bytes length is not a multiple of 32.
   */
  error BytesLengthNotMultipleOf32();

  /**
   * @dev Thrown when the snarkhash is the zero hash.
   */
  error SnarkHashIsZeroHash();

  /**
   * @dev Thrown when parent stateRootHash does not match.
   */
  error ParentStateRootHashInvalid(bytes32 expected, bytes32 actual);

  /**
   * @dev Thrown when the block being finalized until does not match that of the shnarf data.
   */
  error FinalBlockDoesNotMatchShnarfFinalBlock(uint256 expected, uint256 actual);

  /**
   * @dev Thrown when the lengths of the shnarfs array and final block numbers array don't match.
   */
  error ShnarfAndFinalBlockNumberLengthsMismatched(uint256 shnarfsLength, uint256 finalBlockNumbers);

  /**
   * @dev Thrown when the computed shnarf does not match what is expected.
   */
  error FinalShnarfWrong(bytes32 expected, bytes32 value);

  /**
   * @notice Adds or updated the verifier contract address for a proof type.
   * @dev VERIFIER_SETTER_ROLE is required to execute.
   * @param _newVerifierAddress The address for the verifier contract.
   * @param _proofType The proof type being set/updated.
   */
  function setVerifierAddress(address _newVerifierAddress, uint256 _proofType) external;

  /**
   * @notice Unset the verifier contract address for a proof type.
   * @dev VERIFIER_SETTER_ROLE is required to execute.
   * @param _proofType The proof type being set/updated.
   */
  function unsetVerifierAddress(uint256 _proofType) external;

  /**
   * @notice Submit one or more EIP-4844 blobs.
   * @dev OPERATOR_ROLE is required to execute.
   * @dev This should be a blob carrying transaction.
   * @param _blobSubmissionData The data for blob submission including proofs and required polynomials.
   * @param _parentShnarf The parent shnarf used in continuity checks as it includes the parentStateRootHash in its computation.
   * @param _finalBlobShnarf The expected final shnarf post computation of all the blob shnarfs.
   */
  function submitBlobs(
    BlobSubmissionData[] calldata _blobSubmissionData,
    bytes32 _parentShnarf,
    bytes32 _finalBlobShnarf
  ) external;

  /**
   * @notice Submit blobs using compressed data via calldata.
   * @dev OPERATOR_ROLE is required to execute.
   * @param _submissionData The supporting data for compressed data submission including compressed data.
   * @param _parentShnarf The parent shnarf used in continuity checks as it includes the parentStateRootHash in its computation.
   * @param _expectedShnarf The expected shnarf post computation of all the submission.
   */
  function submitDataAsCalldata(
    SubmissionDataV2 calldata _submissionData,
    bytes32 _parentShnarf,
    bytes32 _expectedShnarf
  ) external;

  /**
   * @notice Finalize compressed blocks without proof.
   * @dev DEFAULT_ADMIN_ROLE is required to execute.
   * @param _finalizationData The full finalization data.
   */
  function finalizeBlocksWithoutProof(FinalizationDataV2 calldata _finalizationData) external;

  /**
   * @notice Finalize compressed blocks with proof.
   * @dev OPERATOR_ROLE is required to execute.
   * @param _aggregatedProof The aggregated proof.
   * @param _proofType The proof type.
   * @param _finalizationData The full finalization data.
   */
  function finalizeBlocksWithProof(
    bytes calldata _aggregatedProof,
    uint256 _proofType,
    FinalizationDataV2 calldata _finalizationData
  ) external;
}
