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

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_TOO_LARGE_FOR_REMAINING_USER_GAS;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@Slf4j
@RequiredArgsConstructor
public class MaxBlockGasTransactionSelector implements PluginTransactionSelector {

  private final long maxGasPerBlock;
  private long cumulativeBlockGasUsed;

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {

    final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();
    final long gasUsedByTransaction = processingResult.getEstimateGasUsedByTransaction();

    if (gasUsedByTransaction > maxGasPerBlock) {
      log.atTrace()
          .setMessage(
              "Not selecting transaction {}, its gas used {} is greater than max user gas per block {},"
                  + " removing it from the txpool")
          .addArgument(transaction::getHash)
          .addArgument(gasUsedByTransaction)
          .addArgument(maxGasPerBlock)
          .log();
      return TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS;
    }

    if (isTransactionExceedingMaxBlockGasLimit(gasUsedByTransaction)) {
      log.atTrace()
          .setMessage(
              "Not selecting transaction {}, its cumulative block gas used {} greater than max user gas per block {},"
                  + " skipping it")
          .addArgument(transaction::getHash)
          .addArgument(cumulativeBlockGasUsed)
          .addArgument(maxGasPerBlock)
          .log();
      return TX_TOO_LARGE_FOR_REMAINING_USER_GAS;
    }
    return SELECTED;
  }

  private boolean isTransactionExceedingMaxBlockGasLimit(long transactionGasUsed) {
    try {
      return Math.addExact(cumulativeBlockGasUsed, transactionGasUsed) > maxGasPerBlock;
    } catch (final ArithmeticException e) {
      // Overflow won't occur as cumulativeBlockGasUsed won't exceed Long.MAX_VALUE
      return true;
    }
  }

  @Override
  public void onTransactionSelected(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {
    cumulativeBlockGasUsed =
        Math.addExact(cumulativeBlockGasUsed, processingResult.getEstimateGasUsedByTransaction());
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext) {
    // Evaluation done in post-processing, no action needed here.
    return SELECTED;
  }
}
