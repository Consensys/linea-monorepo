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
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import org.apache.commons.lang3.mutable.MutableBoolean;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@Slf4j
public class ProfitableTransactionSelector implements PluginTransactionSelector {
  @VisibleForTesting protected static Set<Hash> unprofitableCache = new LinkedHashSet<>();
  @VisibleForTesting protected static Wei prevMinGasPrice = Wei.MAX_WEI;

  private final LineaTransactionSelectorConfiguration txSelectorConf;
  private final LineaProfitabilityConfiguration profitabilityConf;
  private final TransactionProfitabilityCalculator transactionProfitabilityCalculator;

  private int unprofitableRetries;
  private MutableBoolean minGasPriceDecreased;

  public ProfitableTransactionSelector(
      final LineaTransactionSelectorConfiguration txSelectorConf,
      final LineaProfitabilityConfiguration profitabilityConf) {
    this.txSelectorConf = txSelectorConf;
    this.profitabilityConf = profitabilityConf;
    this.transactionProfitabilityCalculator =
        new TransactionProfitabilityCalculator(profitabilityConf);
  }

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
      final long gasLimit = transaction.getGasLimit();

      // check the upfront profitability using the gas limit of the tx
      if (!transactionProfitabilityCalculator.isProfitable(
          "PreProcessing",
          transaction,
          profitabilityConf.minMargin(),
          minGasPrice,
          evaluationContext.getTransactionGasPrice(),
          gasLimit)) {
        return TX_UNPROFITABLE_UPFRONT;
      }

      if (unprofitableCache.contains(transaction.getHash())) {
        // only retry unprofitable txs if the min gas price went down
        if (minGasPriceDecreased.isTrue()) {

          if (unprofitableRetries >= txSelectorConf.unprofitableRetryLimit()) {
            log.atTrace()
                .setMessage("Limit of unprofitable tx retries reached: {}/{}")
                .addArgument(unprofitableRetries)
                .addArgument(txSelectorConf.unprofitableRetryLimit());
            return TX_UNPROFITABLE_RETRY_LIMIT;
          }

          log.atTrace()
              .setMessage("Retrying unprofitable tx. Retry: {}/{}")
              .addArgument(unprofitableRetries)
              .addArgument(txSelectorConf.unprofitableRetryLimit());
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
      final long gasUsed = processingResult.getEstimateGasUsedByTransaction();

      if (!transactionProfitabilityCalculator.isProfitable(
          "PostProcessing",
          transaction,
          profitabilityConf.minMargin(),
          evaluationContext.getMinGasPrice(),
          evaluationContext.getTransactionGasPrice(),
          gasUsed)) {
        rememberUnprofitable(transaction);
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

  private void rememberUnprofitable(final Transaction transaction) {
    while (unprofitableCache.size() >= txSelectorConf.unprofitableCacheSize()) {
      final var it = unprofitableCache.iterator();
      if (it.hasNext()) {
        it.next();
        it.remove();
      }
    }
    unprofitableCache.add(transaction.getHash());
    log.atTrace().setMessage("unprofitableCache={}").addArgument(unprofitableCache::size).log();
  }
}
