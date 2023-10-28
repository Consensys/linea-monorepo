/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;

/**
 * This class implements TransactionSelector and is used to select transactions based on their call
 * data size. If the call data size of a transaction exceeds the maximum limit, the transaction is
 * not selected. The maximum limit for the call data size is defined in the LineaConfiguration.
 */
@Slf4j
@RequiredArgsConstructor
public class MaxTransactionCallDataTransactionSelector implements PluginTransactionSelector {

  private final int maxTxCallDataSize;
  /** The reason for invalidation if the call data size is too big. */
  public static String CALL_DATA_TOO_BIG_INVALID_REASON = "Transaction Call Data is too big";

  /**
   * Evaluates a pending transaction based on its call data size.
   *
   * @param pendingTransaction The transaction to be evaluated.
   * @return The result of the transaction selection.
   */
  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final PendingTransaction pendingTransaction) {
    final Transaction transaction = pendingTransaction.getTransaction();

    var transactionCallDataSize = transaction.getPayload().size();
    if (transactionCallDataSize > maxTxCallDataSize) {
      log.warn(
          "Not adding transaction {} because callData size {} is too big",
          transaction,
          transactionCallDataSize);
      return TransactionSelectionResult.invalid(CALL_DATA_TOO_BIG_INVALID_REASON);
    }
    return TransactionSelectionResult.SELECTED;
  }

  /**
   * No evaluation is performed post-processing.
   *
   * @param pendingTransaction The processed transaction.
   * @param processingResult The result of the transaction processing.
   * @return Always returns SELECTED.
   */
  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final PendingTransaction pendingTransaction,
      final TransactionProcessingResult processingResult) {
    // Evaluation done in pre-processing, no action needed here.
    return TransactionSelectionResult.SELECTED;
  }
}
