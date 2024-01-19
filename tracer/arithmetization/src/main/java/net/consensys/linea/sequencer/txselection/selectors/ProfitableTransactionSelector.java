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

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_UNPROFITABLE;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;
import org.slf4j.spi.LoggingEventBuilder;

@Slf4j
@RequiredArgsConstructor
public class ProfitableTransactionSelector implements PluginTransactionSelector {
  private final int verificationGasCost;
  private final int verificationCapacity;
  private final int gasPriceRatio;
  private final double minMargin;

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {

    if (!evaluationContext.getPendingTransaction().hasPriority()) {
      final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();

      final double effectiveGasPrice =
          evaluationContext.getTransactionGasPrice().getAsBigInteger().doubleValue();
      final long gasUsed = processingResult.getEstimateGasUsedByTransaction();
      final double revenue = effectiveGasPrice * gasUsed;

      final double minGasPrice = evaluationContext.getMinGasPrice().getAsBigInteger().doubleValue();
      final double l1GasPrice = minGasPrice * gasPriceRatio;
      final int serializedSize = transaction.getSize();
      final double verificationGasCostSlice =
          (((double) serializedSize) / verificationCapacity) * verificationGasCost;
      final double cost = l1GasPrice * verificationGasCostSlice;

      final double margin = revenue / cost;

      log(
          log.atTrace(),
          transaction,
          margin,
          effectiveGasPrice,
          gasUsed,
          minGasPrice,
          l1GasPrice,
          serializedSize);

      if (margin < minMargin) {
        log(
            log.atDebug(),
            transaction,
            margin,
            effectiveGasPrice,
            gasUsed,
            minGasPrice,
            l1GasPrice,
            serializedSize);
        return TX_UNPROFITABLE;
      }
    }
    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext) {
    return SELECTED;
  }

  private void log(
      final LoggingEventBuilder leb,
      final Transaction transaction,
      final double margin,
      final double effectiveGasPrice,
      final long gasUsed,
      final double minGasPrice,
      final double l1GasPrice,
      final int serializedSize) {
    leb.setMessage(
            "Transaction {} has a margin of {}, minMargin={}, verificationCapacity={}, "
                + "verificationGasCost={}, gasPriceRatio={}, effectiveGasPrice={}, gasUsed={}, minGasPrice={}, "
                + "l1GasPrice={}, serializedSize={}")
        .addArgument(transaction::getHash)
        .addArgument(margin)
        .addArgument(minMargin)
        .addArgument(verificationCapacity)
        .addArgument(verificationGasCost)
        .addArgument(gasPriceRatio)
        .addArgument(effectiveGasPrice)
        .addArgument(gasUsed)
        .addArgument(minGasPrice)
        .addArgument(l1GasPrice)
        .addArgument(serializedSize)
        .log();
  }
}
