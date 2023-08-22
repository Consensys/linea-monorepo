// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { L1MessageService } from "./messageService/l1/L1MessageService.sol";
import { TransactionDecoder } from "./messageService/lib/TransactionDecoder.sol";
import { IZkEvmV2 } from "./interfaces/IZkEvmV2.sol";
import { IPlonkVerifier } from "./interfaces/IPlonkVerifier.sol";
import { CodecV2 } from "./messageService/lib/Codec.sol";

/**
 * @title Contract to manage cross-chain messaging on L1 and rollup proving.
 * @author ConsenSys Software Inc.
 */
contract ZkEvmV2 is IZkEvmV2, Initializable, AccessControlUpgradeable, L1MessageService {
  using TransactionDecoder for *;
  using CodecV2 for *;

  uint256 private constant MODULO_R = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
  bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

  uint256 public currentTimestamp;
  uint256 public currentL2BlockNumber;

  mapping(uint256 => bytes32) public stateRootHashes;
  mapping(uint256 => address) public verifiers;

  uint256[50] private __gap;

  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /**
   * @notice Initializes zkEvm and underlying service dependencies.
   * @dev DEFAULT_ADMIN_ROLE is set for the security council.
   * @dev OPERATOR_ROLE is set for operators.
   * @param _initialStateRootHash The initial hash at migration used for proof verification.
   * @param _initialL2BlockNumber The initial block number at migration.
   * @param _defaultVerifier The default verifier for rollup proofs.
   * @param _securityCouncil The address for the security council performing admin operations.
   * @param _operators The allowed rollup operators at initialization.
   * @param _rateLimitPeriodInSeconds The period in which withdrawal amounts and fees will be accumulated.
   * @param _rateLimitAmountInWei The limit allowed for withdrawing in the period.
   **/
  function initialize(
    bytes32 _initialStateRootHash,
    uint256 _initialL2BlockNumber,
    address _defaultVerifier,
    address _securityCouncil,
    address[] calldata _operators,
    uint256 _rateLimitPeriodInSeconds,
    uint256 _rateLimitAmountInWei
  ) public initializer {
    if (_defaultVerifier == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    for (uint256 i; i < _operators.length; ) {
      if (_operators[i] == address(0)) {
        revert ZeroAddressNotAllowed();
      }
      _grantRole(OPERATOR_ROLE, _operators[i]);
      unchecked {
        i++;
      }
    }

    _grantRole(DEFAULT_ADMIN_ROLE, _securityCouncil);

    __MessageService_init(_securityCouncil, _securityCouncil, _rateLimitPeriodInSeconds, _rateLimitAmountInWei);

    verifiers[0] = _defaultVerifier;
    currentL2BlockNumber = _initialL2BlockNumber;
    stateRootHashes[_initialL2BlockNumber] = _initialStateRootHash;
  }

  /**
   * @notice Adds or updates the verifier contract address for a proof type.
   * @dev DEFAULT_ADMIN_ROLE is required to execute.
   * @param _newVerifierAddress The address for the verifier contract.
   * @param _proofType The proof type being set/updated.
   **/
  function setVerifierAddress(address _newVerifierAddress, uint256 _proofType) external onlyRole(DEFAULT_ADMIN_ROLE) {
    if (_newVerifierAddress == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    emit VerifierAddressChanged(_newVerifierAddress, _proofType, msg.sender);

    verifiers[_proofType] = _newVerifierAddress;
  }

  /**
   * @notice Finalizes blocks without using a proof.
   * @dev DEFAULT_ADMIN_ROLE is required to execute.
   * @dev _blocksData[0].fromAddresses is a temporary workaround to pass bytes calldata
   * @param _blocksData The full BlockData collection - block, transaction and log data.
   **/
  function finalizeBlocksWithoutProof(
    BlockData[] calldata _blocksData
  ) external whenTypeNotPaused(GENERAL_PAUSE_TYPE) onlyRole(DEFAULT_ADMIN_ROLE) {
    _finalizeBlocks(_blocksData, _blocksData[0].fromAddresses, 0, bytes32(0), false);
  }

  /**
   * @notice Finalizes blocks using a proof.
   * @dev OPERATOR_ROLE is required to execute.
   * @dev If the verifier based on proof type is not found, it reverts.
   * @param _blocksData The full BlockData collection - block, transaction and log data.
   * @param _proof The proof to be verified with the proof type verifier contract.
   * @param _proofType The proof type to determine which verifier contract to use.
   * @param _parentStateRootHash The starting roothash for the last known block.
   **/
  function finalizeBlocks(
    BlockData[] calldata _blocksData,
    bytes calldata _proof,
    uint256 _proofType,
    bytes32 _parentStateRootHash
  )
    external
    whenTypeNotPaused(PROVING_SYSTEM_PAUSE_TYPE)
    whenTypeNotPaused(GENERAL_PAUSE_TYPE)
    onlyRole(OPERATOR_ROLE)
  {
    if (stateRootHashes[currentL2BlockNumber] != _parentStateRootHash) {
      revert StartingRootHashDoesNotMatch();
    }

    _finalizeBlocks(_blocksData, _proof, _proofType, _parentStateRootHash, true);
  }

  /**
   * @notice Finalizes blocks with or without using a proof depending on _shouldProve
   * @dev If the verifier based on proof type is not found, it reverts.
   * @param _blocksData The full BlockData collection - block, transaction and log data.
   * @param _proof The proof to be verified with the proof type verifier contract.
   * @param _proofType The proof type to determine which verifier contract to use.
   * @param _parentStateRootHash The starting roothash for the last known block.
   **/
  function _finalizeBlocks(
    BlockData[] calldata _blocksData,
    bytes calldata _proof,
    uint256 _proofType,
    bytes32 _parentStateRootHash,
    bool _shouldProve
  ) private {
    if (_blocksData.length == 0) {
      revert EmptyBlockDataArray();
    }

    uint256 currentBlockNumberTemp = currentL2BlockNumber;

    uint256 firstBlockNumber;
    unchecked {
      firstBlockNumber = currentBlockNumberTemp + 1;
    }

    uint256[] memory timestamps = new uint256[](_blocksData.length);
    bytes32[] memory blockHashes = new bytes32[](_blocksData.length);
    bytes32[] memory hashOfRootHashes;

    unchecked {
      hashOfRootHashes = new bytes32[](_blocksData.length + 1);
    }

    hashOfRootHashes[0] = _parentStateRootHash;

    bytes32 hashOfTxHashes;
    bytes32 hashOfMessageHashes;

    for (uint256 i; i < _blocksData.length; ) {
      BlockData calldata blockInfo = _blocksData[i];

      if (blockInfo.l2BlockTimestamp >= block.timestamp) {
        revert BlockTimestampError();
      }

      hashOfTxHashes = _processBlockTransactions(blockInfo.transactions, blockInfo.batchReceptionIndices);
      hashOfMessageHashes = _processMessageHashes(blockInfo.l2ToL1MsgHashes);

      unchecked {
        ++currentBlockNumberTemp;
      }

      blockHashes[i] = keccak256(
        abi.encodePacked(
          hashOfTxHashes,
          hashOfMessageHashes,
          keccak256(abi.encodePacked(blockInfo.batchReceptionIndices)),
          keccak256(blockInfo.fromAddresses)
        )
      );

      timestamps[i] = blockInfo.l2BlockTimestamp;

      unchecked {
        hashOfRootHashes[i + 1] = blockInfo.blockRootHash;
      }

      emit BlockFinalized(currentBlockNumberTemp, blockInfo.blockRootHash);

      unchecked {
        i++;
      }
    }

    unchecked {
      uint256 arrayIndex = _blocksData.length - 1;
      stateRootHashes[currentBlockNumberTemp] = _blocksData[arrayIndex].blockRootHash;
      currentTimestamp = _blocksData[arrayIndex].l2BlockTimestamp;
      currentL2BlockNumber = currentBlockNumberTemp;
    }

    if (_shouldProve) {
      uint256 publicInput = uint256(
        keccak256(
          abi.encode(
            keccak256(abi.encodePacked(blockHashes)),
            firstBlockNumber,
            keccak256(abi.encodePacked(timestamps)),
            keccak256(abi.encodePacked(hashOfRootHashes))
          )
        )
      );

      assembly {
        publicInput := mod(publicInput, MODULO_R)
      }

      _verifyProof(publicInput, _proofType, _proof, _parentStateRootHash);
    }
  }

  /**
   * @notice Hashes all transactions individually and then hashes the packed hash array.
   * @dev Updates the outbox status on L1 as received.
   * @param _transactions The transactions in a particular block.
   * @param _batchReceptionIndices The indexes where the transaction type is the L1->L2 achoring message hashes transaction.
   **/
  function _processBlockTransactions(
    bytes[] calldata _transactions,
    uint16[] calldata _batchReceptionIndices
  ) internal returns (bytes32 hashOfTxHashes) {
    bytes32[] memory transactionHashes = new bytes32[](_transactions.length);

    if (_transactions.length == 0) {
      revert EmptyBlock();
    }

    for (uint256 i; i < _batchReceptionIndices.length; ) {
      _updateL1L2MessageStatusToReceived(
        TransactionDecoder.decodeTransaction(_transactions[_batchReceptionIndices[i]])._extractXDomainAddHashes()
      );

      unchecked {
        i++;
      }
    }

    for (uint256 i; i < _transactions.length; ) {
      transactionHashes[i] = keccak256(_transactions[i]);

      unchecked {
        i++;
      }
    }
    hashOfTxHashes = keccak256(abi.encodePacked(transactionHashes));
  }

  /**
   * @notice Anchors message hashes and hashes the packed hash array.
   * @dev Also adds L2->L1 sent message hashes for later claiming.
   * @param _messageHashes The hashes in the message sent event logs.
   **/
  function _processMessageHashes(bytes32[] calldata _messageHashes) internal returns (bytes32 hashOfLogHashes) {
    for (uint256 i; i < _messageHashes.length; ) {
      _addL2L1MessageHash(_messageHashes[i]);

      unchecked {
        i++;
      }
    }
    hashOfLogHashes = keccak256(abi.encodePacked(_messageHashes));
  }

  /**
   * @notice Verifies the proof with locally computed public inputs.
   * @dev If the verifier based on proof type is not found, it reverts with InvalidProofType.
   * @param _publicInputHash The full BlockData collection - block, transaction and log data.
   * @param _proofType The proof type to determine which verifier contract to use.
   * @param _proof The proof to be verified with the proof type verifier contract.
   * @param _parentStateRootHash The beginning roothash to start with.
   **/
  function _verifyProof(
    uint256 _publicInputHash,
    uint256 _proofType,
    bytes calldata _proof,
    bytes32 _parentStateRootHash
  ) private {
    uint256[] memory input = new uint256[](1);
    input[0] = _publicInputHash;

    address verifierToUse = verifiers[_proofType];

    if (verifierToUse == address(0)) {
      revert InvalidProofType();
    }

    bool success = IPlonkVerifier(verifierToUse).Verify(_proof, input);
    if (!success) {
      revert InvalidProof();
    }

    emit BlocksVerificationDone(currentL2BlockNumber, _parentStateRootHash, stateRootHashes[currentL2BlockNumber]);
  }
}
