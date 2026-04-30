/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txpoolvalidation.validators;

import static net.consensys.linea.config.TransactionGasLimitCap.EIP_7825_MAX_TRANSACTION_GAS_LIMIT;
import static net.consensys.linea.config.TransactionGasLimitCap.effectiveMaxTxGasLimit;
import static net.consensys.linea.config.TransactionGasLimitCap.gasLimitExceededMessage;

import java.util.Optional;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;

/**
 * Validator that checks if the gas limit is below the configured max amount. This means that max
 * gas limit of a transaction could be less than the block gas limit.
 */
@Slf4j
public class GasLimitValidator implements PluginTransactionPoolValidator {
  final int configuredMaxTxGasLimit;
  final int effectiveMaxTxGasLimit;

  public GasLimitValidator(final int maxTxGasLimit) {
    this.configuredMaxTxGasLimit = maxTxGasLimit;
    this.effectiveMaxTxGasLimit = effectiveMaxTxGasLimit(maxTxGasLimit);

    if (maxTxGasLimit > EIP_7825_MAX_TRANSACTION_GAS_LIMIT) {
      log.warn(
          "Configured maximum transaction gas limit {} exceeds EIP-7825 maximum transaction gas "
              + "limit {}; using effective maximum transaction gas limit {}",
          maxTxGasLimit,
          EIP_7825_MAX_TRANSACTION_GAS_LIMIT,
          effectiveMaxTxGasLimit);
    }
  }

  @Override
  public Optional<String> validateTransaction(
      final Transaction transaction, final boolean isLocal, final boolean hasPriority) {
    if (transaction.getGasLimit() > effectiveMaxTxGasLimit) {
      final String errMsg =
          gasLimitExceededMessage(transaction.getGasLimit(), configuredMaxTxGasLimit);
      log.debug(errMsg);
      return Optional.of(errMsg);
    }
    return Optional.empty();
  }
}
