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
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_UNPROFITABLE_MIN_GAS_PRICE_NOT_DECREASED;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_UNPROFITABLE_RETRY_LIMIT;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_UNPROFITABLE_UPFRONT;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import java.util.LinkedHashSet;
import java.util.Set;

import com.google.common.annotations.VisibleForTesting;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.mutable.MutableBoolean;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;
import org.slf4j.spi.LoggingEventBuilder;

@Slf4j
@RequiredArgsConstructor
public class ProfitableTransactionSelector implements PluginTransactionSelector {
  @VisibleForTesting protected static Set<Hash> unprofitableCache = new LinkedHashSet<>();
  @VisibleForTesting protected static Wei prevMinGasPrice = Wei.MAX_WEI;

  private final int verificationGasCost;
  private final int verificationCapacity;
  private final int gasPriceRatio;
  private final double minMargin;
  private final int adjustTxSize;
  private final int unprofitableCacheSize;
  private final int unprofitableRetryLimit;
  private int unprofitableRetries;
  private MutableBoolean minGasPriceDecreased;

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext) {

    final Wei minGasPrice = evaluationContext.getMinGasPrice();

    // update prev min gas price only if it is a new block
    if (minGasPriceDecreased == null) {
      minGasPriceDecreased = new MutableBoolean(minGasPrice.lessThan(prevMinGasPrice));
      prevMinGasPrice = minGasPrice;
    }

    if (!evaluationContext.getPendingTransaction().hasPriority()) {
      final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();
      final double effectiveGasPrice =
          evaluationContext.getTransactionGasPrice().getAsBigInteger().doubleValue();
      final long gasLimit = transaction.getGasLimit();

      // check the upfront profitability using the gas limit of the tx
      if (!isProfitable(
          "PreProcessing",
          transaction,
          minGasPrice.getAsBigInteger().doubleValue(),
          effectiveGasPrice,
          gasLimit)) {
        return TX_UNPROFITABLE_UPFRONT;
      }

      if (unprofitableCache.contains(transaction.getHash())) {
        // only retry unprofitable txs if the min gas price went down
        if (minGasPriceDecreased.isTrue()) {

          if (unprofitableRetries >= unprofitableRetryLimit) {
            log.atTrace()
                .setMessage("Limit of unprofitable tx retries reached: {}/{}")
                .addArgument(unprofitableRetries)
                .addArgument(unprofitableRetryLimit);
            return TX_UNPROFITABLE_RETRY_LIMIT;
          }

          log.atTrace()
              .setMessage("Retrying unprofitable tx. Retry: {}/{}")
              .addArgument(unprofitableRetries)
              .addArgument(unprofitableRetryLimit);
          unprofitableCache.remove(transaction.getHash());
          unprofitableRetries++;

        } else {
          log.atTrace()
              .setMessage(
                  "Current block minGasPrice {} is higher than previous block {}, skipping unprofitable txs retry")
              .addArgument(minGasPrice::toHumanReadableString)
              .addArgument(prevMinGasPrice::toHumanReadableString)
              .log();
          return TX_UNPROFITABLE_MIN_GAS_PRICE_NOT_DECREASED;
        }
      }
    }

    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {

    if (!evaluationContext.getPendingTransaction().hasPriority()) {
      final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();
      final double minGasPrice = evaluationContext.getMinGasPrice().getAsBigInteger().doubleValue();
      final double effectiveGasPrice =
          evaluationContext.getTransactionGasPrice().getAsBigInteger().doubleValue();
      final long gasUsed = processingResult.getEstimateGasUsedByTransaction();

      if (!isProfitable("PostProcessing", transaction, minGasPrice, effectiveGasPrice, gasUsed)) {
        registerAsUnProfitable(transaction);
        return TX_UNPROFITABLE;
      }
    }
    return SELECTED;
  }

  @Override
  public void onTransactionSelected(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {
    unprofitableCache.remove(evaluationContext.getPendingTransaction().getTransaction().getHash());
  }

  @Override
  public void onTransactionNotSelected(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionSelectionResult transactionSelectionResult) {
    if (transactionSelectionResult.discard()) {
      unprofitableCache.remove(
          evaluationContext.getPendingTransaction().getTransaction().getHash());
    }
  }

  private boolean isProfitable(
      final String step,
      final Transaction transaction,
      final double minGasPrice,
      final double effectiveGasPrice,
      final long gas) {
    final double revenue = effectiveGasPrice * gas;

    final double l1GasPrice = minGasPrice * gasPriceRatio;
    final int serializedSize = Math.max(0, transaction.getSize() + adjustTxSize);
    final double verificationGasCostSlice =
        (((double) serializedSize) / verificationCapacity) * verificationGasCost;
    final double cost = l1GasPrice * verificationGasCostSlice;

    final double margin = revenue / cost;

    if (margin < minMargin) {
      log(
          log.atDebug(),
          step,
          transaction,
          margin,
          effectiveGasPrice,
          gas,
          minGasPrice,
          l1GasPrice,
          serializedSize,
          adjustTxSize);
      return false;
    } else {
      log(
          log.atTrace(),
          step,
          transaction,
          margin,
          effectiveGasPrice,
          gas,
          minGasPrice,
          l1GasPrice,
          serializedSize,
          adjustTxSize);
      return true;
    }
  }

  private void registerAsUnProfitable(final Transaction transaction) {
    if (unprofitableCache.size() >= unprofitableCacheSize) {
      final var it = unprofitableCache.iterator();
      if (it.hasNext()) {
        it.next();
        it.remove();
      }
    }
    unprofitableCache.add(transaction.getHash());
  }

  private void log(
      final LoggingEventBuilder leb,
      final String step,
      final Transaction transaction,
      final double margin,
      final double effectiveGasPrice,
      final long gasUsed,
      final double minGasPrice,
      final double l1GasPrice,
      final int serializedSize,
      final int adjustTxSize) {
    leb.setMessage(
            "Step {}. Transaction {} has a margin of {}, minMargin={}, verificationCapacity={}, "
                + "verificationGasCost={}, gasPriceRatio={}, effectiveGasPrice={}, gasUsed={}, minGasPrice={}, "
                + "l1GasPrice={}, serializedSize={}, adjustTxSize={}")
        .addArgument(step)
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
        .addArgument(adjustTxSize)
        .log();
  }
}
