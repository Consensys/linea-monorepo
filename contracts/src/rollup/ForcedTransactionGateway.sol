// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.28;
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IAcceptForcedTransactions } from "./interfaces/IAcceptForcedTransactions.sol";
import { IForcedTransactionGateway } from "./interfaces/IForcedTransactionGateway.sol";
import { Mimc } from "../libraries/Mimc.sol";
import { RlpEncoder } from "../libraries/RlpEncoder.sol";
import { FinalizedStateHashing } from "../libraries/FinalizedStateHashing.sol";
/**
 * @title Contract to manage forced transactions on L1.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract ForcedTransactionGateway is IForcedTransactionGateway {
  using Mimc for *;
  using RlpEncoder for *;
  using FinalizedStateHashing for *;

  uint256 private constant UNSIGNED_TRANSACTION_FIELD_LENGTH = 9;
  uint256 private constant SIGNED_TRANSACTION_FIELD_LENGTH = 12;
  address private constant PRECOMPILE_ADDRESS_LIMIT = address(21);

  /// @notice Contains the destination address to store the forced transactions on.
  IAcceptForcedTransactions public immutable LINEA_ROLLUP;

  /// @notice Contains the destination chain ID used in the RLP encoding.
  uint256 public immutable DESTINATION_CHAIN_ID;

  /// @notice Contains the buffer for computing the L2 block the transaction will be processed by.
  uint256 public immutable L2_BLOCK_BUFFER;

  /// @notice Contains the maximum gas allowed for a forced transaction.
  uint256 public immutable MAX_GAS_LIMIT;

  /// @notice Contains the maximum calldata length allowed for a forced transaction.
  uint256 public immutable MAX_INPUT_LENGTH_LIMIT;

  constructor(
    address _lineaRollup,
    uint256 _destinationChainId,
    uint256 _l2BlockBuffer,
    uint256 _maxGasLimit,
    uint256 _maxInputLengthBuffer
  ) {
    if (_lineaRollup == address(0)) {
      revert IGenericErrors.ZeroAddressNotAllowed();
    }

    if (_destinationChainId == 0) {
      revert IGenericErrors.ZeroValueNotAllowed();
    }

    if (_l2BlockBuffer == 0) {
      revert IGenericErrors.ZeroValueNotAllowed();
    }

    if (_maxGasLimit == 0) {
      revert IGenericErrors.ZeroValueNotAllowed();
    }

    if (_maxInputLengthBuffer == 0) {
      revert IGenericErrors.ZeroValueNotAllowed();
    }

    LINEA_ROLLUP = IAcceptForcedTransactions(_lineaRollup);
    DESTINATION_CHAIN_ID = _destinationChainId;
    L2_BLOCK_BUFFER = _l2BlockBuffer;
    MAX_GAS_LIMIT = _maxGasLimit;
    MAX_INPUT_LENGTH_LIMIT = _maxInputLengthBuffer;
  }

  /**
   * @notice Function to submit forced transactions.
   * @param _forcedTransaction The fields required for the transaction excluding chainId.
   * @param _lastFinalizedState The last finalized state validated to use the timestamp in block number calculation.
   */
  function submitForcedTransaction(
    Eip1559Transaction memory _forcedTransaction,
    LastFinalizedState calldata _lastFinalizedState
  ) external {
    if (_forcedTransaction.gasLimit > MAX_GAS_LIMIT) {
      revert MaxGasLimitExceeded();
    }

    if (_forcedTransaction.input.length > MAX_INPUT_LENGTH_LIMIT) {
      revert CalldataInputLengthLimitExceeded();
    }

    if (_forcedTransaction.maxPriorityFeePerGas == 0 || _forcedTransaction.maxFeePerGas == 0) {
      revert GasFeeParametersContainZero(_forcedTransaction.maxFeePerGas, _forcedTransaction.maxPriorityFeePerGas);
    }

    if (_forcedTransaction.maxPriorityFeePerGas > _forcedTransaction.maxFeePerGas) {
      revert MaxPriorityFeePerGasHigherThanMaxFee(
        _forcedTransaction.maxFeePerGas,
        _forcedTransaction.maxPriorityFeePerGas
      );
    }

    if (_forcedTransaction.yParity > 1) {
      revert YParityGreaterThanOne(_forcedTransaction.yParity);
    }

    if (_forcedTransaction.to < address(PRECOMPILE_ADDRESS_LIMIT)) {
      revert ToAddressTooLow();
    }

    (
      bytes32 currentFinalizedState,
      bytes32 previousForcedTransactionRollingHash,
      uint256 currentFinalizedL2BlockNumber
    ) = LINEA_ROLLUP.getRequiredForcedTransactionFields();

    if (
      currentFinalizedState !=
      FinalizedStateHashing._computeLastFinalizedState(
        _lastFinalizedState.messageNumber,
        _lastFinalizedState.messageRollingHash,
        _lastFinalizedState.forcedTransactionNumber,
        _lastFinalizedState.forcedTransactionRollingHash,
        _lastFinalizedState.timestamp
      )
    ) {
      revert FinalizationStateIncorrect(
        currentFinalizedState,
        FinalizedStateHashing._computeLastFinalizedState(
          _lastFinalizedState.messageNumber,
          _lastFinalizedState.messageRollingHash,
          _lastFinalizedState.forcedTransactionNumber,
          _lastFinalizedState.forcedTransactionRollingHash,
          _lastFinalizedState.timestamp
        )
      );
    }

    bytes[] memory signedTransactionFields = new bytes[](SIGNED_TRANSACTION_FIELD_LENGTH);
    signedTransactionFields[0] = RlpEncoder._encodeUint(DESTINATION_CHAIN_ID);
    signedTransactionFields[1] = RlpEncoder._encodeUint(_forcedTransaction.nonce);
    signedTransactionFields[2] = RlpEncoder._encodeUint(_forcedTransaction.maxPriorityFeePerGas);
    signedTransactionFields[3] = RlpEncoder._encodeUint(_forcedTransaction.maxFeePerGas);
    signedTransactionFields[4] = RlpEncoder._encodeUint(_forcedTransaction.gasLimit);
    signedTransactionFields[5] = RlpEncoder._encodeAddress(_forcedTransaction.to);
    signedTransactionFields[6] = RlpEncoder._encodeUint(_forcedTransaction.value);
    signedTransactionFields[7] = RlpEncoder._encodeBytes(_forcedTransaction.input);
    signedTransactionFields[8] = RlpEncoder._encodeAccessList(_forcedTransaction.accessList);
    signedTransactionFields[9] = RlpEncoder._encodeUint(_forcedTransaction.yParity);
    signedTransactionFields[10] = RlpEncoder._encodeUint(_forcedTransaction.r);
    signedTransactionFields[11] = RlpEncoder._encodeUint(_forcedTransaction.s);

    bytes memory rlpEncodedUnsignedTransaction = abi.encodePacked(
      hex"02",
      RlpEncoder._encodeList(signedTransactionFields, UNSIGNED_TRANSACTION_FIELD_LENGTH)
    );

    bytes32 hashedPayload = keccak256(rlpEncodedUnsignedTransaction);

    address signer;
    unchecked {
      signer = ecrecover(
        hashedPayload,
        _forcedTransaction.yParity + 27,
        bytes32(_forcedTransaction.r),
        bytes32(_forcedTransaction.s)
      );
    }

    uint256 expectedBlockNumber;
    unchecked {
      /// @dev The computation uses 1s block time making block number and seconds interchangable.
      expectedBlockNumber =
        currentFinalizedL2BlockNumber +
        block.timestamp -
        _lastFinalizedState.timestamp +
        L2_BLOCK_BUFFER;
    }

    bytes32 forcedTransactionRollingHash = _computeForcedTransactionRollingHash(
      previousForcedTransactionRollingHash,
      hashedPayload,
      expectedBlockNumber,
      signer
    );

    emit ForcedTransactionAdded(
      LINEA_ROLLUP.storeForcedTransaction(expectedBlockNumber, forcedTransactionRollingHash),
      signer,
      expectedBlockNumber,
      forcedTransactionRollingHash,
      rlpEncodedUnsignedTransaction,
      abi.encodePacked(hex"02", RlpEncoder._encodeList(signedTransactionFields))
    );
  }

  /**
   * @notice Function to compute the forced transaction rolling hash with Mimc.
   * @dev _hashedUnsignedPayload is split into two 16 bytes limbs, each represented in MSB over uint256 (e.g. 0x00..<16 bytes>..00yy..<16 bytes>..yy),
   * to ensure that each limb fits on a BLS12-377 field element without overflowing, preventing potential resulting collisions.
   * @param _previousRollingHash The previous rolling hash.
   * @param _hashedUnsignedPayload The keccak256 hashed unsigned payload.
   * @param _expectedBlockNumber The expected block number.
   * @param _signer The transaction signer address.
   */
  function _computeForcedTransactionRollingHash(
    bytes32 _previousRollingHash,
    bytes32 _hashedUnsignedPayload,
    uint256 _expectedBlockNumber,
    address _signer
  ) internal pure returns (bytes32 forcedTransactionRollingHash) {
    bytes memory mimcPayload;

    assembly {
      let mostSignificantBytes := shr(128, _hashedUnsignedPayload)
      let leastSignificantBytes := and(_hashedUnsignedPayload, 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF)

      mimcPayload := mload(0x40)
      mstore(mimcPayload, 0xA0)
      mstore(add(mimcPayload, 0x20), _previousRollingHash)
      mstore(add(mimcPayload, 0x40), mostSignificantBytes)
      mstore(add(mimcPayload, 0x60), leastSignificantBytes)
      mstore(add(mimcPayload, 0x80), _expectedBlockNumber)
      mstore(add(mimcPayload, 0xA0), _signer)
      mstore(0x40, add(mimcPayload, 0xC0))
    }

    forcedTransactionRollingHash = Mimc.hash(mimcPayload);
  }
}
