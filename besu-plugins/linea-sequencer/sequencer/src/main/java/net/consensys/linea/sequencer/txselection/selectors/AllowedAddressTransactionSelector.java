/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_FILTERED_ADDRESSES;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import java.util.Optional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;

/**
 * Transaction selector that rejects transactions based on sender or recipient address validation.
 * Delegates to {@link
 * net.consensys.linea.sequencer.txpoolvalidation.validators.DeniedAddressValidator} for the actual
 * validation logic.
 */
@Slf4j
@RequiredArgsConstructor
public class AllowedAddressTransactionSelector implements PluginTransactionSelector {

  private final PluginTransactionPoolValidator denylistValidator;

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();

    final Optional<String> validationError =
        denylistValidator.validateTransaction(transaction, false, false);
    if (validationError.isPresent()) {
      log.atTrace()
          .setMessage("action=reject_filtered_address txHash={} reason={}")
          .addArgument(transaction::getHash)
          .addArgument(validationError::get)
          .log();
      return TX_FILTERED_ADDRESSES;
    }

    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    return SELECTED;
  }
}
