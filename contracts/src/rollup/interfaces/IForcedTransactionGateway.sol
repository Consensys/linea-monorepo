// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { RlpEncoder } from "../../libraries/RlpEncoder.sol";

/**
 * @title Interface to manage forced transactions on L1.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IForcedTransactionGateway {
  /**
   * @notice Supporting data for the last finalized rollup state.
   * @dev timestamp The last finalized timestamp.
   * @dev messageNumber The L2 computed L1 message number.
   * @dev messageRollingHash The L2 computed L1 message rolling hash.
   * @dev forcedTransactionNumber The last finalized forced transaction processed on L2.
   * @dev forcedTransactionRollingHash The last finalized forced transaction's rolling hash processed on L2.
   */
  struct LastFinalizedState {
    uint256 timestamp;
    uint256 messageNumber;
    bytes32 messageRollingHash;
    uint256 forcedTransactionNumber;
    bytes32 forcedTransactionRollingHash;
  }

  /**
   * @notice Supporting data for an EIP-1559 transaction.
   * @dev nonce The nonce for the transaction belonging to the signer.
   * @dev maxPriorityFeePerGas The max priority fee per gas.
   * @dev maxFeePerGas The max fee per gas.
   * @dev gasLimit The transaction's gas limit.
   * @dev to The destination address of the transaction.
   * @dev value The Ether value to transfer.
   * @dev input The calldata input to send to the "to" address.
   * @dev accessList The access list for the transaction.
   * @dev yParity The signature's yParity.
   * @dev r The r portion of the signature.
   * @dev s The s portion of the signature.
   */
  struct Eip1559Transaction {
    uint256 nonce;
    uint256 maxPriorityFeePerGas;
    uint256 maxFeePerGas;
    uint256 gasLimit;
    address to;
    uint256 value;
    bytes input;
    RlpEncoder.AccessList[] accessList;
    uint8 yParity;
    uint256 r;
    uint256 s;
  }

  /**
   * @notice Emitted when a forced transaction is added.
   * @param forcedTransactionNumber The indexed forced transaction number.
   * @param from The recovered signer's from address.
   * @param expectedBlockNumber The maximum expected L2 block number processing will occur by.
   * @param forcedTransactionRollingHash The computed rolling Mimc based hash.
   * @param rlpEncodedSignedTransaction The RLP encoded type 02 transaction payload including signature.
   */
  event ForcedTransactionAdded(
    uint256 indexed forcedTransactionNumber,
    address from,
    uint256 expectedBlockNumber,
    bytes32 forcedTransactionRollingHash,
    bytes rlpEncodedSignedTransaction
  );

  /**
   * @dev Thrown when the max gas limit configured will be exceeded.
   */
  error MaxGasLimitExceeded();

  /**
   * @dev Thrown when the gas limit configured is less than the minimum 21000.
   */
  error GasLimitTooLow();

  /**
   * @dev Thrown when the input length on the transaction is too long.
   */
  error CalldataInputLengthLimitExceeded();

  /**
   * @dev Thrown when one of the gas fee parameters are zero.
   */
  error GasFeeParametersContainZero(uint256 maxFeePerGas, uint256 maxPriorityFeePerGas);

  /**
   * @dev Thrown when the max priority fee per gas is higher than the max fee per gas.
   */
  error MaxPriorityFeePerGasHigherThanMaxFee(uint256 maxFeePerGas, uint256 maxPriorityFeePerGas);

  /**
   * @dev Thrown when the yParity is not 0 or 1.
   */
  error YParityGreaterThanOne(uint256 yParity);

  /**
   * @dev Thrown when the to address is the zero address or a precompile.
   */
  error ToAddressTooLow();

  /**
   * @dev Thrown when finalization state does not match.
   */
  error FinalizationStateIncorrect(bytes32 expected, bytes32 value);

  /**
   * @notice Function to submit forced transactions.
   * @param _forcedTransaction The fields required for the transaction excluding chainId.
   * @param _lastFinalizedState The last finalized state validated to use the timestamp in block number calculation.
   */
  function submitForcedTransaction(
    Eip1559Transaction memory _forcedTransaction,
    LastFinalizedState calldata _lastFinalizedState
  ) external;
}
