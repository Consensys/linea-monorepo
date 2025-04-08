// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

import { RlpEncoder } from "../../../libraries/RlpEncoder.sol";

/**
 * @title Interface to manage forced transactions on L1.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IForcedTransactionGateway {
  struct LastFinalizedState {
    uint256 timestamp;
    uint256 messageNumber;
    bytes32 messageRollingHash;
    uint256 forcedTransactionNumber;
    bytes32 forcedTransactionRollingHash;
  }

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

  event ForcedTransactionAdded(
    address from,
    uint256 expectedBlockNumber,
    bytes32 forcedTransactionRollingHash,
    bytes rlpEncodedUnsignedTransaction,
    bytes rlpEncodedSignedTransaction
  );

  error MaxGasLimitExceeded();
  error CalldataInputLengthLimitExceeded();
  error GasParamatersContainZero(uint256 maxFeePerGas, uint256 maxPriorityFeePerGas);
  error MaxPriorityFeePerGasHigherThanMaxFee(uint256 maxFeePerGas, uint256 maxPriorityFeePerGas);
  error YParityGreaterThanOne(uint256 yParity);
  error SignerDoesNotMatchFrom(address expected, address actual);
  error ToAddressTooLow();

  /**
   * @dev Thrown when finalization state does not match.
   */
  error FinalizationStateIncorrect(bytes32 expected, bytes32 value);

  function submitForcedTransaction(
    Eip1559Transaction memory _forcedTransaction,
    LastFinalizedState calldata _lastFinalizedState
  ) external;
}
