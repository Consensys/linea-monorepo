// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.28;
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

  IAcceptForcedTransactions public immutable LINEA_ROLLUP;
  uint256 public immutable DESTINATION_CHAIN_ID;
  uint256 public immutable L2_BLOCK_BUFFER;
  uint256 public immutable MAX_GAS_LIMIT;
  uint256 public immutable MAX_INPUT_LENGTH_LIMIT;

  constructor(
    address _lineaRollup,
    uint256 _destinationChainId,
    uint256 _l2BlockBuffer,
    uint256 _maxGasLimit,
    uint256 _maxInputLengthBuffer
  ) {
    LINEA_ROLLUP = IAcceptForcedTransactions(_lineaRollup);
    DESTINATION_CHAIN_ID = _destinationChainId;
    L2_BLOCK_BUFFER = _l2BlockBuffer;
    MAX_GAS_LIMIT = _maxGasLimit;
    MAX_INPUT_LENGTH_LIMIT = _maxInputLengthBuffer;
  }

  function submitForcedTransaction(
    Eip1559Transaction memory _forcedTransaction,
    LastFinalizedState calldata _lastFinalizedState
  ) external {
    // gas limit check
    if (_forcedTransaction.gasLimit > MAX_GAS_LIMIT) {
      revert MaxGasLimitExceeded();
    }

    // calldata length check
    if (_forcedTransaction.input.length > MAX_INPUT_LENGTH_LIMIT) {
      revert CalldataInputLengthLimitExceeded();
    }

    // 0 value gas fields
    if (_forcedTransaction.maxPriorityFeePerGas == 0 || _forcedTransaction.maxFeePerGas == 0) {
      revert GasParamatersContainZero(_forcedTransaction.maxFeePerGas, _forcedTransaction.maxPriorityFeePerGas);
    }

    // priority fee must not be more than max fee per gas
    if (_forcedTransaction.maxPriorityFeePerGas > _forcedTransaction.maxFeePerGas) {
      revert MaxPriorityFeePerGasHigherThanMaxFee(
        _forcedTransaction.maxFeePerGas,
        _forcedTransaction.maxPriorityFeePerGas
      );
    }

    // no point if parity is not 0 or 1
    if (_forcedTransaction.yParity > 1) {
      revert YParityGreaterThanOne(_forcedTransaction.yParity);
    }

    // exclude precompiles and address 0
    if (_forcedTransaction.to < address(21)) {
      revert ToAddressTooLow();
    }

    // while less than ideal naming, it saves gas by doing a single query
    // gets all LineaRollup fields (nb: increments the counter)
    (
      bytes32 currentFinalizedState,
      uint256 forcedTransactionNumber,
      bytes32 previousForcedTransactionRollingHash,
      uint256 currentFinalizedL2BlockNumber
    ) = LINEA_ROLLUP.getLineaRollupProvidedFields();

    // validate state is correct in order to use the timestamp.. we might need a better way than this.
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

    // all 12 fields for sequencer raw submission
    bytes[] memory signedTransactionFields = new bytes[](12);
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

    // clone for RLP encoding just the unsigned transaction payload fields
    bytes[] memory unsignedTransactionFields = new bytes[](9);
    for (uint256 i; i < 9; i++) {
      unsignedTransactionFields[i] = signedTransactionFields[i];
    }

    // RLP encode the unsigned transaction payload fields
    bytes memory rlpEncodedUnsignedTransaction = abi.encodePacked(
      hex"02",
      RlpEncoder._encodeList(unsignedTransactionFields)
    );

    // Hash the RLP encoded insigned transaction to get
    bytes32 hashedPayload = keccak256(rlpEncodedUnsignedTransaction);

    // recover from for prover
    address signer;
    unchecked {
      signer = ecrecover(
        hashedPayload,
        _forcedTransaction.yParity + 27,
        bytes32(_forcedTransaction.r),
        bytes32(_forcedTransaction.s)
      );
    }

    // COMPUTE BLOCK NUMBER - TO BE DISCUSSED
    uint256 expectedBlockNumber;
    unchecked {
      // last L2 block + seconds between then and now to get "current" block and then 3 days of 1 second
      expectedBlockNumber =
        currentFinalizedL2BlockNumber +
        block.timestamp -
        _lastFinalizedState.timestamp +
        L2_BLOCK_BUFFER; // we assume a 1 second block time
    }

    // compute a rolling mimc hash
    bytes32 forcedTransactionRollingHash = _computeForcedTransactionRollingHash(
      previousForcedTransactionRollingHash,
      hashedPayload,
      expectedBlockNumber,
      signer
    );

    // store the computed rolling hash validating there isn't already an existing forced transaction in the same block
    LINEA_ROLLUP.storeForcedTransaction(forcedTransactionNumber, expectedBlockNumber, forcedTransactionRollingHash);

    emit ForcedTransactionAdded(
      signer,
      expectedBlockNumber,
      forcedTransactionRollingHash,
      rlpEncodedUnsignedTransaction,
      abi.encodePacked(hex"02", RlpEncoder._encodeList(signedTransactionFields))
    );
  }

  function _computeForcedTransactionRollingHash(
    bytes32 _previousRollingHash,
    bytes32 _hashedPayload,
    uint256 _expectedBlockNumber,
    address _from
  ) internal pure returns (bytes32 forcedTransactionRollingHash) {
    bytes memory mimcPayload;

    assembly {
      let mostSignificantBytes := shr(128, _hashedPayload)
      let leadSignificantBytes := and(_hashedPayload, 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF)

      mimcPayload := mload(0x40)
      mstore(mimcPayload, 0xA0)
      mstore(add(mimcPayload, 0x20), _previousRollingHash)
      mstore(add(mimcPayload, 0x40), mostSignificantBytes)
      mstore(add(mimcPayload, 0x60), leadSignificantBytes)
      mstore(add(mimcPayload, 0x80), _expectedBlockNumber)
      mstore(add(mimcPayload, 0xA0), _from)
      mstore(0x40, add(mimcPayload, 0xC0))
    }

    forcedTransactionRollingHash = Mimc.hash(mimcPayload);
  }
}
