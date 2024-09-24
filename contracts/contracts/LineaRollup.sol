// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { L1MessageService } from "./messageService/l1/L1MessageService.sol";
import { ZkEvmV2 } from "./ZkEvmV2.sol";
import { ILineaRollup } from "./interfaces/l1/ILineaRollup.sol";

import { Utils } from "./lib/Utils.sol";
/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LineaRollup is AccessControlUpgradeable, ZkEvmV2, L1MessageService, ILineaRollup {
  using Utils for *;

  bytes32 public constant VERIFIER_SETTER_ROLE = keccak256("VERIFIER_SETTER_ROLE");
  bytes32 public constant GENESIS_SHNARF =
    keccak256(
      abi.encode(
        EMPTY_HASH,
        EMPTY_HASH,
        0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd,
        EMPTY_HASH,
        EMPTY_HASH
      )
    );

  bytes32 internal constant EMPTY_HASH = 0x0;
  uint256 internal constant BLS_CURVE_MODULUS =
    52435875175126190479447740508185965837690552500527637822603658699938581184513;
  address internal constant POINT_EVALUATION_PRECOMPILE_ADDRESS = address(0x0a);
  uint256 internal constant POINT_EVALUATION_RETURN_DATA_LENGTH = 64;
  uint256 internal constant POINT_EVALUATION_FIELD_ELEMENTS_LENGTH = 4096;

  uint256 internal constant SIX_MONTHS_IN_SECONDS = 15768000; // 365 / 2 * 86400

  /// @dev DEPRECATED in favor of the single shnarfFinalBlockNumbers mapping.
  mapping(bytes32 dataHash => bytes32 finalStateRootHash) public dataFinalStateRootHashes;
  /// @dev DEPRECATED in favor of the single shnarfFinalBlockNumbers mapping.
  mapping(bytes32 dataHash => bytes32 parentHash) public dataParents;
  /// @dev DEPRECATED in favor of the single shnarfFinalBlockNumbers mapping.
  mapping(bytes32 dataHash => bytes32 shnarfHash) public dataShnarfHashes;
  /// @dev DEPRECATED in favor of the single shnarfFinalBlockNumbers mapping.
  mapping(bytes32 dataHash => uint256 startingBlock) public dataStartingBlock;
  /// @dev DEPRECATED in favor of the single shnarfFinalBlockNumbers mapping.
  mapping(bytes32 dataHash => uint256 endingBlock) public dataEndingBlock;

  /// @dev DEPRECATED in favor of currentFinalizedState hash.
  uint256 public currentL2StoredL1MessageNumber;
  /// @dev DEPRECATED in favor of currentFinalizedState hash.
  bytes32 public currentL2StoredL1RollingHash;

  bytes32 public currentFinalizedShnarf;

  /**
   * @dev NB: THIS IS THE ONLY MAPPING BEING USED FOR DATA SUBMISSION TRACKING.
   */
  mapping(bytes32 shnarf => uint256 finalBlockNumber) public shnarfFinalBlockNumbers;

  /// @dev Hash of the L2 computed L1 message number, rolling hash and finalized timestamp.
  bytes32 public currentFinalizedState;

  /// @dev The address of the gateway operator.
  /// @dev This address is granted the OPERATOR_ROLE after six months of finalization inactivity by the current operators.
  address public gatewayOperator;

  /// @dev Total contract storage is 11 slots.

  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /**
   * @notice Initializes LineaRollup and underlying service dependencies - used for new networks only.
   * @dev DEFAULT_ADMIN_ROLE is set for the security council.
   * @dev OPERATOR_ROLE is set for operators.
   * @dev Note: This is used for new testnets and local/CI testing, and will not replace existing proxy based contracts.
   * @param _initialStateRootHash The initial hash at migration used for proof verification.
   * @param _initialL2BlockNumber The initial block number at migration.
   * @param _defaultVerifier The default verifier for rollup proofs.
   * @param _securityCouncil The address for the security council performing admin operations.
   * @param _operators The allowed rollup operators at initialization.
   * @param _rateLimitPeriodInSeconds The period in which withdrawal amounts and fees will be accumulated.
   * @param _rateLimitAmountInWei The limit allowed for withdrawing in the rate limit period.
   * @param _genesisTimestamp The L2 genesis timestamp for first finalization.
   */
  function initialize(
    bytes32 _initialStateRootHash,
    uint256 _initialL2BlockNumber,
    address _defaultVerifier,
    address _securityCouncil,
    address[] calldata _operators,
    uint256 _rateLimitPeriodInSeconds,
    uint256 _rateLimitAmountInWei,
    uint256 _genesisTimestamp,
    address _gatewayOperator
  ) external initializer {
    if (_defaultVerifier == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    for (uint256 i; i < _operators.length; ++i) {
      if (_operators[i] == address(0)) {
        revert ZeroAddressNotAllowed();
      }
      _grantRole(OPERATOR_ROLE, _operators[i]);
    }

    _grantRole(DEFAULT_ADMIN_ROLE, _securityCouncil);
    _grantRole(VERIFIER_SETTER_ROLE, _securityCouncil);

    __MessageService_init(_securityCouncil, _securityCouncil, _rateLimitPeriodInSeconds, _rateLimitAmountInWei);

    verifiers[0] = _defaultVerifier;

    gatewayOperator = _gatewayOperator;

    currentL2BlockNumber = _initialL2BlockNumber;
    stateRootHashes[_initialL2BlockNumber] = _initialStateRootHash;

    shnarfFinalBlockNumbers[GENESIS_SHNARF] = _initialL2BlockNumber;

    currentFinalizedShnarf = GENESIS_SHNARF;
    currentFinalizedState = _computeLastFinalizedState(0, EMPTY_HASH, _genesisTimestamp);
  }

  /**
   * @notice Initializes and sets the gateway operator.
   * @dev This will be included with whatever other initializer upgrades we need to perform on branch merging.
   * @param _gatewayOperator The address of the gateway operator.
   */
  function initializeGatewayOperator(address _gatewayOperator) external reinitializer(6) {
    gatewayOperator = _gatewayOperator;
  }

  /**
   * @notice Adds or updates the verifier contract address for a proof type.
   * @dev VERIFIER_SETTER_ROLE is required to execute.
   * @param _newVerifierAddress The address for the verifier contract.
   * @param _proofType The proof type being set/updated.
   */
  function setVerifierAddress(address _newVerifierAddress, uint256 _proofType) external onlyRole(VERIFIER_SETTER_ROLE) {
    if (_newVerifierAddress == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    emit VerifierAddressChanged(_newVerifierAddress, _proofType, msg.sender, verifiers[_proofType]);

    verifiers[_proofType] = _newVerifierAddress;
  }

  /**
   * @notice Sets the gateway operator role to the specified address if six months have passed since the last finalization.
   * @dev Reverts if six months have not passed since the last finalization.
   * @param _messageNumber Last finalized L1 message number as part of the feedback loop.
   * @param _rollingHash Last finalized L1 rolling hash as part of the feedback loop.
   * @param _lastFinalizedTimestamp Last finalized L2 block timestamp.
   */
  function setGatewayOperator(uint256 _messageNumber, bytes32 _rollingHash, uint256 _lastFinalizedTimestamp) external {
    if (block.timestamp < _lastFinalizedTimestamp + SIX_MONTHS_IN_SECONDS) {
      revert LastFinalizationTimeNotLapsed();
    }
    if (currentFinalizedState != _computeLastFinalizedState(_messageNumber, _rollingHash, _lastFinalizedTimestamp)) {
      revert FinalizationStateIncorrect(
        currentFinalizedState,
        _computeLastFinalizedState(_messageNumber, _rollingHash, _lastFinalizedTimestamp)
      );
    }

    address gatewayOperatorAddress = gatewayOperator;

    _grantRole(OPERATOR_ROLE, gatewayOperatorAddress);
    emit GatewayOperatorRoleGranted(msg.sender, gatewayOperatorAddress);
  }

  /**
   * @notice Unset the verifier contract address for a proof type.
   * @dev VERIFIER_SETTER_ROLE is required to execute.
   * @param _proofType The proof type being set/updated.
   */
  function unsetVerifierAddress(uint256 _proofType) external onlyRole(VERIFIER_SETTER_ROLE) {
    emit VerifierAddressChanged(address(0), _proofType, msg.sender, verifiers[_proofType]);

    delete verifiers[_proofType];
  }

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
  ) external whenTypeAndGeneralNotPaused(PROVING_SYSTEM_PAUSE_TYPE) onlyRole(OPERATOR_ROLE) {
    uint256 blobSubmissionLength = _blobSubmissionData.length;

    if (blobSubmissionLength == 0) {
      revert BlobSubmissionDataIsMissing();
    }

    bytes32 currentDataEvaluationPoint;
    bytes32 currentDataHash;
    uint256 lastFinalizedBlockNumber = currentL2BlockNumber;

    /// @dev Assigning in memory saves a lot of gas vs. calldata reading.
    BlobSubmissionData memory blobSubmissionData;

    bytes32 computedShnarf = _parentShnarf;

    uint256 blobFinalBlockNumber = shnarfFinalBlockNumbers[computedShnarf];

    for (uint256 i; i < blobSubmissionLength; i++) {
      blobSubmissionData = _blobSubmissionData[i];

      currentDataHash = blobhash(i);

      if (currentDataHash == EMPTY_HASH) {
        revert EmptyBlobDataAtIndex(i);
      }

      _validateSubmissionData(blobSubmissionData.submissionData, blobFinalBlockNumber, lastFinalizedBlockNumber);

      currentDataEvaluationPoint = Utils._efficientKeccak(blobSubmissionData.submissionData.snarkHash, currentDataHash);

      _verifyPointEvaluation(
        currentDataHash,
        uint256(currentDataEvaluationPoint),
        blobSubmissionData.dataEvaluationClaim,
        blobSubmissionData.kzgCommitment,
        blobSubmissionData.kzgProof
      );

      computedShnarf = _computeShnarf(
        computedShnarf,
        blobSubmissionData.submissionData.snarkHash,
        blobSubmissionData.submissionData.finalStateRootHash,
        currentDataEvaluationPoint,
        bytes32(blobSubmissionData.dataEvaluationClaim)
      );

      blobFinalBlockNumber = blobSubmissionData.submissionData.finalBlockInData;
    }

    if (_finalBlobShnarf != computedShnarf) {
      revert FinalShnarfWrong(_finalBlobShnarf, computedShnarf);
    }

    /**
     * @dev validate we haven't submitted the last shnarf.
     * Note: As only the last shnarf is stored, we don't need to validate shnarfs,
     * computed for any previous blobs in the submission (if multiple are submitted).
     */
    if (shnarfFinalBlockNumbers[computedShnarf] != 0) {
      revert DataAlreadySubmitted(computedShnarf);
    }

    /// @dev use the last shnarf as the submission to store as technically it becomes the next parent shnarf.
    shnarfFinalBlockNumbers[computedShnarf] = blobFinalBlockNumber;

    emit DataSubmittedV2(computedShnarf, _blobSubmissionData[0].submissionData.firstBlockInData, blobFinalBlockNumber);
  }

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
  ) external whenTypeAndGeneralNotPaused(PROVING_SYSTEM_PAUSE_TYPE) onlyRole(OPERATOR_ROLE) {
    if (_submissionData.compressedData.length == 0) {
      revert EmptySubmissionData();
    }

    SupportingSubmissionDataV2 memory submissionData = SupportingSubmissionDataV2({
      finalStateRootHash: _submissionData.finalStateRootHash,
      firstBlockInData: _submissionData.firstBlockInData,
      finalBlockInData: _submissionData.finalBlockInData,
      snarkHash: _submissionData.snarkHash
    });

    bytes32 currentDataHash = keccak256(_submissionData.compressedData);

    _validateSubmissionData(submissionData, shnarfFinalBlockNumbers[_parentShnarf], currentL2BlockNumber);

    bytes32 dataEvaluationPoint = Utils._efficientKeccak(_submissionData.snarkHash, currentDataHash);
    bytes32 computedShnarf = _computeShnarf(
      _parentShnarf,
      _submissionData.snarkHash,
      _submissionData.finalStateRootHash,
      dataEvaluationPoint,
      _calculateY(_submissionData.compressedData, dataEvaluationPoint)
    );

    if (_expectedShnarf != computedShnarf) {
      revert FinalShnarfWrong(_expectedShnarf, computedShnarf);
    }

    if (shnarfFinalBlockNumbers[computedShnarf] != 0) {
      revert DataAlreadySubmitted(computedShnarf);
    }

    shnarfFinalBlockNumbers[computedShnarf] = _submissionData.finalBlockInData;

    emit DataSubmittedV2(computedShnarf, _submissionData.firstBlockInData, _submissionData.finalBlockInData);
  }

  /**
   * @notice Internal function to validate submission data.
   * @param _submissionData The supporting data for compressed data submission excluding compressed data.
   * @param _parentFinalBlockNumber The final block number for the parent blob.
   * @param _lastFinalizedBlockNumber The last finalized block number.
   */
  function _validateSubmissionData(
    SupportingSubmissionDataV2 memory _submissionData,
    uint256 _parentFinalBlockNumber,
    uint256 _lastFinalizedBlockNumber
  ) internal pure {
    if (_submissionData.finalStateRootHash == EMPTY_HASH) {
      revert FinalBlockStateEqualsZeroHash();
    }

    if (_submissionData.snarkHash == EMPTY_HASH) {
      revert SnarkHashIsZeroHash();
    }

    // for it to be equal the number would have to wrap round twice in overflow..
    unchecked {
      if (_parentFinalBlockNumber + 1 != _submissionData.firstBlockInData) {
        revert DataStartingBlockDoesNotMatch(_parentFinalBlockNumber + 1, _submissionData.firstBlockInData);
      }
    }

    if (_submissionData.firstBlockInData <= _lastFinalizedBlockNumber) {
      revert FirstBlockLessThanOrEqualToLastFinalizedBlock(_submissionData.firstBlockInData, _lastFinalizedBlockNumber);
    }

    if (_submissionData.firstBlockInData > _submissionData.finalBlockInData) {
      revert FirstBlockGreaterThanFinalBlock(_submissionData.firstBlockInData, _submissionData.finalBlockInData);
    }
  }

  /**
   * @notice Internal function to compute and save the finalization state.
   * @dev Using assembly this way is cheaper gas wise.
   * @param _messageNumber Is the last L2 computed L1 message number in the finalization.
   * @param _rollingHash Is the last L2 computed L1 rolling hash in the finalization.
   * @param _timestamp The final timestamp in the finalization.
   */
  function _computeLastFinalizedState(
    uint256 _messageNumber,
    bytes32 _rollingHash,
    uint256 _timestamp
  ) internal pure returns (bytes32 hashedFinalizationState) {
    assembly {
      let mPtr := mload(0x40)
      mstore(mPtr, _messageNumber)
      mstore(add(mPtr, 0x20), _rollingHash)
      mstore(add(mPtr, 0x40), _timestamp)
      hashedFinalizationState := keccak256(mPtr, 0x60)
    }
  }

  /**
   * @notice Internal function to compute the shnarf more efficiently.
   * @dev Using assembly this way is cheaper gas wise.
   * @param _parentShnarf The shnarf of the parent data item.
   * @param _snarkHash Is the computed hash for compressed data (using a SNARK-friendly hash function) that aggregates per data submission to be used in public input.
   * @param _finalStateRootHash The final state root hash of the data being submitted.
   * @param _dataEvaluationPoint The data evaluation point.
   * @param _dataEvaluationClaim The data evaluation claim.
   */
  function _computeShnarf(
    bytes32 _parentShnarf,
    bytes32 _snarkHash,
    bytes32 _finalStateRootHash,
    bytes32 _dataEvaluationPoint,
    bytes32 _dataEvaluationClaim
  ) internal pure returns (bytes32 shnarf) {
    assembly {
      let mPtr := mload(0x40)
      mstore(mPtr, _parentShnarf)
      mstore(add(mPtr, 0x20), _snarkHash)
      mstore(add(mPtr, 0x40), _finalStateRootHash)
      mstore(add(mPtr, 0x60), _dataEvaluationPoint)
      mstore(add(mPtr, 0x80), _dataEvaluationClaim)
      shnarf := keccak256(mPtr, 0xA0)
    }
  }

  /**
   * @notice Performs point evaluation for the compressed blob.
   * @dev _dataEvaluationPoint is modular reduced to be lower than the BLS_CURVE_MODULUS for precompile checks.
   * @param _currentDataHash The current blob versioned hash.
   * @param _dataEvaluationPoint The data evaluation point.
   * @param _dataEvaluationClaim The data evaluation claim.
   * @param _kzgCommitment The blob KZG commitment.
   * @param _kzgProof The blob KZG point proof.
   */
  function _verifyPointEvaluation(
    bytes32 _currentDataHash,
    uint256 _dataEvaluationPoint,
    uint256 _dataEvaluationClaim,
    bytes memory _kzgCommitment,
    bytes memory _kzgProof
  ) internal view {
    assembly {
      _dataEvaluationPoint := mod(_dataEvaluationPoint, BLS_CURVE_MODULUS)
    }

    (bool success, bytes memory returnData) = POINT_EVALUATION_PRECOMPILE_ADDRESS.staticcall(
      abi.encodePacked(_currentDataHash, _dataEvaluationPoint, _dataEvaluationClaim, _kzgCommitment, _kzgProof)
    );

    if (!success) {
      revert PointEvaluationFailed();
    }

    if (returnData.length != POINT_EVALUATION_RETURN_DATA_LENGTH) {
      revert PrecompileReturnDataLengthWrong(POINT_EVALUATION_RETURN_DATA_LENGTH, returnData.length);
    }

    uint256 fieldElements;
    uint256 blsCurveModulus;
    assembly {
      fieldElements := mload(add(returnData, 32))
      blsCurveModulus := mload(add(returnData, POINT_EVALUATION_RETURN_DATA_LENGTH))
    }
    if (fieldElements != POINT_EVALUATION_FIELD_ELEMENTS_LENGTH || blsCurveModulus != BLS_CURVE_MODULUS) {
      revert PointEvaluationResponseInvalid(fieldElements, blsCurveModulus);
    }
  }

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
  ) external whenTypeAndGeneralNotPaused(PROVING_SYSTEM_PAUSE_TYPE) onlyRole(OPERATOR_ROLE) {
    if (_aggregatedProof.length == 0) {
      revert ProofIsEmpty();
    }

    uint256 lastFinalizedBlockNumber = currentL2BlockNumber;

    if (stateRootHashes[lastFinalizedBlockNumber] != _finalizationData.parentStateRootHash) {
      revert StartingRootHashDoesNotMatch();
    }

    bytes32 lastFinalizedShnarf = currentFinalizedShnarf;

    if (_finalizationData.lastFinalizedShnarf != lastFinalizedShnarf) {
      revert LastFinalizedShnarfWrong(lastFinalizedShnarf, _finalizationData.lastFinalizedShnarf);
    }

    bytes32 finalShnarf = _finalizeBlocks(_finalizationData, lastFinalizedBlockNumber, true);

    uint256 publicInput = _computePublicInput(
      _finalizationData,
      lastFinalizedShnarf,
      finalShnarf,
      lastFinalizedBlockNumber
    );

    _verifyProof(
      publicInput,
      _proofType,
      _aggregatedProof,
      _finalizationData.parentStateRootHash,
      _finalizationData.finalBlockInData,
      _finalizationData.shnarfData.finalStateRootHash
    );
  }

  /**
   * @notice Finalize compressed blocks without proof.
   * @dev DEFAULT_ADMIN_ROLE is required to execute.
   * @param _finalizationData The full finalization data.
   */
  function finalizeBlocksWithoutProof(
    FinalizationDataV2 calldata _finalizationData
  ) external whenTypeNotPaused(GENERAL_PAUSE_TYPE) onlyRole(DEFAULT_ADMIN_ROLE) {
    _finalizeBlocks(_finalizationData, currentL2BlockNumber, false);
  }

  /**
   * @notice Internal function to finalize compressed blocks.
   * @param _finalizationData The full finalization data.
   * @param _lastFinalizedBlock The last finalized block.
   * @param _withProof If we are finalizing with a proof.
   * @return finalShnarf The final computed shnarf in finalizing.
   */
  function _finalizeBlocks(
    FinalizationDataV2 calldata _finalizationData,
    uint256 _lastFinalizedBlock,
    bool _withProof
  ) internal returns (bytes32 finalShnarf) {
    if (_finalizationData.finalBlockInData <= _lastFinalizedBlock) {
      revert FinalBlockNumberLessThanOrEqualToLastFinalizedBlock(
        _finalizationData.finalBlockInData,
        _lastFinalizedBlock
      );
    }

    _validateL2ComputedRollingHash(_finalizationData.l1RollingHashMessageNumber, _finalizationData.l1RollingHash);

    if (
      _computeLastFinalizedState(
        _finalizationData.lastFinalizedL1RollingHashMessageNumber,
        _finalizationData.lastFinalizedL1RollingHash,
        _finalizationData.lastFinalizedTimestamp
      ) != currentFinalizedState
    ) {
      revert FinalizationStateIncorrect(
        _computeLastFinalizedState(
          _finalizationData.lastFinalizedL1RollingHashMessageNumber,
          _finalizationData.lastFinalizedL1RollingHash,
          _finalizationData.lastFinalizedTimestamp
        ),
        currentFinalizedState
      );
    }

    if (_finalizationData.finalTimestamp >= block.timestamp) {
      revert FinalizationInTheFuture(_finalizationData.finalTimestamp, block.timestamp);
    }

    if (_finalizationData.shnarfData.finalStateRootHash == EMPTY_HASH) {
      revert FinalBlockStateEqualsZeroHash();
    }

    finalShnarf = _computeShnarf(
      _finalizationData.shnarfData.parentShnarf,
      _finalizationData.shnarfData.snarkHash,
      _finalizationData.shnarfData.finalStateRootHash,
      _finalizationData.shnarfData.dataEvaluationPoint,
      _finalizationData.shnarfData.dataEvaluationClaim
    );

    if (shnarfFinalBlockNumbers[finalShnarf] != _finalizationData.finalBlockInData) {
      revert FinalBlockDoesNotMatchShnarfFinalBlock(
        _finalizationData.finalBlockInData,
        shnarfFinalBlockNumbers[finalShnarf]
      );
    }

    _addL2MerkleRoots(_finalizationData.l2MerkleRoots, _finalizationData.l2MerkleTreesDepth);
    _anchorL2MessagingBlocks(_finalizationData.l2MessagingBlocksOffsets, _lastFinalizedBlock);

    stateRootHashes[_finalizationData.finalBlockInData] = _finalizationData.shnarfData.finalStateRootHash;

    currentL2BlockNumber = _finalizationData.finalBlockInData;

    currentFinalizedShnarf = finalShnarf;

    currentFinalizedState = _computeLastFinalizedState(
      _finalizationData.l1RollingHashMessageNumber,
      _finalizationData.l1RollingHash,
      _finalizationData.finalTimestamp
    );

    emit DataFinalized(
      _finalizationData.finalBlockInData,
      _finalizationData.parentStateRootHash,
      _finalizationData.shnarfData.finalStateRootHash,
      _withProof
    );
  }

  /**
   * @notice Internal function to validate l1 rolling hash.
   * @param _rollingHashMessageNumber Message number associated with the rolling hash as computed on L2.
   * @param _rollingHash L1 rolling hash as computed on L2.
   */
  function _validateL2ComputedRollingHash(uint256 _rollingHashMessageNumber, bytes32 _rollingHash) internal view {
    if (_rollingHashMessageNumber == 0) {
      if (_rollingHash != EMPTY_HASH) {
        revert MissingMessageNumberForRollingHash(_rollingHash);
      }
    } else {
      if (_rollingHash == EMPTY_HASH) {
        revert MissingRollingHashForMessageNumber(_rollingHashMessageNumber);
      }
      if (rollingHashes[_rollingHashMessageNumber] != _rollingHash) {
        revert L1RollingHashDoesNotExistOnL1(_rollingHashMessageNumber, _rollingHash);
      }
    }
  }

  /**
   * @notice Internal function to calculate Y for public input generation.
   * @param _data Compressed data from submission data.
   * @param _dataEvaluationPoint The data evaluation point.
   * @dev Each chunk of 32 bytes must start with a 0 byte.
   * @dev The dataEvaluationPoint value is modulo-ed down during the computation and scalar field checking is not needed.
   * @dev There is a hard constraint in the circuit to enforce the polynomial degree limit (4096), which will also be enforced with EIP-4844.
   * @return compressedDataComputedY The Y calculated value using the Horner method.
   */
  function _calculateY(
    bytes calldata _data,
    bytes32 _dataEvaluationPoint
  ) internal pure returns (bytes32 compressedDataComputedY) {
    if (_data.length % 0x20 != 0) {
      revert BytesLengthNotMultipleOf32();
    }

    bytes4 errorSelector = ILineaRollup.FirstByteIsNotZero.selector;
    assembly {
      for {
        let i := _data.length
      } gt(i, 0) {

      } {
        i := sub(i, 0x20)
        let chunk := calldataload(add(_data.offset, i))
        if iszero(iszero(and(chunk, 0xFF00000000000000000000000000000000000000000000000000000000000000))) {
          let ptr := mload(0x40)
          mstore(ptr, errorSelector)
          revert(ptr, 0x4)
        }
        compressedDataComputedY := addmod(
          mulmod(compressedDataComputedY, _dataEvaluationPoint, BLS_CURVE_MODULUS),
          chunk,
          BLS_CURVE_MODULUS
        )
      }
    }
  }

  /**
   * @notice Compute the public input.
   * @dev Using assembly this way is cheaper gas wise.
   * @dev NB: the dynamic sized fields are placed last in _finalizationData on purpose to optimise hashing ranges.
   * @dev Computing the public input as the following:
   * keccak256(
   *  abi.encode(
   *     _lastFinalizedShnarf,
   *     _finalShnarf,
   *     _finalizationData.lastFinalizedTimestamp,
   *     _finalizationData.finalTimestamp,
   *     _lastFinalizedBlockNumber,
   *     _finalizationData.finalBlockInData,
   *     _finalizationData.lastFinalizedL1RollingHash,
   *     _finalizationData.l1RollingHash,
   *     _finalizationData.lastFinalizedL1RollingHashMessageNumber,
   *     _finalizationData.l1RollingHashMessageNumber,
   *     _finalizationData.l2MerkleTreesDepth,
   *     keccak256(
   *         abi.encodePacked(_finalizationData.l2MerkleRoots)
   *     )
   *   )
   * )
   * Data is found at the following offsets:
   * 0x00    parentStateRootHash
   * 0x20    lastFinalizedShnarf
   * 0x40    finalBlockInData
   * 0x60    shnarfData.parentShnarf
   * 0x80    shnarfData.snarkHash
   * 0xa0    shnarfData.finalStateRootHash
   * 0xc0    shnarfData.dataEvaluationPoint
   * 0xe0    shnarfData.dataEvaluationClaim
   * 0x100   lastFinalizedTimestamp
   * 0x120   finalTimestamp
   * 0x140   lastFinalizedL1RollingHash
   * 0x160   l1RollingHash
   * 0x180   lastFinalizedL1RollingHashMessageNumber
   * 0x1a0   l1RollingHashMessageNumber
   * 0x1c0   l2MerkleTreesDepth
   * 0x1e0   l2MerkleRootsLengthLocation
   * 0x200   l2MessagingBlocksOffsetsLengthLocation
   * 0x220   l2MerkleRootsLength
   * 0x240   l2MerkleRoots
   * Dynamic l2MessagingBlocksOffsetsLength (location depends on where l2MerkleRoots ends)
   * Dynamic l2MessagingBlocksOffsets (location depends on where l2MerkleRoots ends)
   * @param _finalizationData The full finalization data.
   * @param _lastFinalizedShnarf The last finalized shnarf.
   * @param _finalShnarf The final shnarf in the finalization.
   * @param _lastFinalizedBlockNumber The last finalized block number.
   */
  function _computePublicInput(
    FinalizationDataV2 calldata _finalizationData,
    bytes32 _lastFinalizedShnarf,
    bytes32 _finalShnarf,
    uint256 _lastFinalizedBlockNumber
  ) private pure returns (uint256 publicInput) {
    assembly {
      let mPtr := mload(0x40)
      mstore(mPtr, _lastFinalizedShnarf)
      mstore(add(mPtr, 0x20), _finalShnarf)

      /**
       * _finalizationData.lastFinalizedTimestamp
       * _finalizationData.finalTimestamp
       */
      calldatacopy(add(mPtr, 0x40), add(_finalizationData, 0x100), 0x40)

      mstore(add(mPtr, 0x80), _lastFinalizedBlockNumber)

      // _finalizationData.finalBlockInData
      calldatacopy(add(mPtr, 0xA0), add(_finalizationData, 0x40), 0x20)

      /**
       * _finalizationData.lastFinalizedL1RollingHash
       * _finalizationData.l1RollingHash
       * _finalizationData.lastFinalizedL1RollingHashMessageNumber
       * _finalizationData.l1RollingHashMessageNumber
       * _finalizationData.l2MerkleTreesDepth
       */
      calldatacopy(add(mPtr, 0xC0), add(_finalizationData, 0x140), 0xA0)

      /**
       * @dev Note the following in hashing the _finalizationData.l2MerkleRoots array:
       * The second memory pointer and free pointer are offset by 0x20 to temporarily hash the array outside the scope of working memory,
       * as we need the space left for the array hash to be stored at 0x160.
       */
      let mPtrMerkleRoot := add(mPtr, 0x180)
      let merkleRootsLen := calldataload(add(_finalizationData, 0x220))
      calldatacopy(mPtrMerkleRoot, add(_finalizationData, 0x240), mul(merkleRootsLen, 0x20))
      let l2MerkleRootsHash := keccak256(mPtrMerkleRoot, mul(merkleRootsLen, 0x20))
      mstore(add(mPtr, 0x160), l2MerkleRootsHash)

      publicInput := mod(keccak256(mPtr, 0x180), MODULO_R)
    }
  }
}
