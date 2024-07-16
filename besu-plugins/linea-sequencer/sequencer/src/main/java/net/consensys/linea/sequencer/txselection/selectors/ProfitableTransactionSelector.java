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
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

/**
 * This class implements TransactionSelector and provides a specific implementation for evaluating
 * if the transaction is profitable, according to the current config and the min margin defined for
 * this context. Profitability check is done upfront using the gas limit, to avoid processing the
 * transaction at all, and if it passes it is done after the processing this time using the actual
 * gas used by the transaction. This selector keeps a cache of the unprofitable transactions to
 * avoid reprocessing all of them everytime, and only allows for a configurable limited number of
 * unprofitable transaction to be retried on every new block creation.
 */
@Slf4j
public class ProfitableTransactionSelector implements PluginTransactionSelector {
  @VisibleForTesting protected static Set<Hash> unprofitableCache = new LinkedHashSet<>();

  private final LineaTransactionSelectorConfiguration txSelectorConf;
  private final LineaProfitabilityConfiguration profitabilityConf;
  private final TransactionProfitabilityCalculator transactionProfitabilityCalculator;
  private final Wei baseFee;

  private int unprofitableRetries;

  public ProfitableTransactionSelector(
      final BlockchainService blockchainService,
      final LineaTransactionSelectorConfiguration txSelectorConf,
      final LineaProfitabilityConfiguration profitabilityConf) {
    this.txSelectorConf = txSelectorConf;
    this.profitabilityConf = profitabilityConf;
    this.transactionProfitabilityCalculator =
        new TransactionProfitabilityCalculator(profitabilityConf);
    this.baseFee =
        blockchainService
            .getNextBlockBaseFee()
            .orElseThrow(() -> new RuntimeException("We only support a base fee market"));
  }

  /**
   * Evaluates a transaction before processing. Checks if it is profitable using its gas limit. If
   * the transaction was found to be unprofitable during a previous block creation process, it is
   * retried, since the gas price market could now make it profitable, but only a configurable
   * amount of these transactions is retried each time, to avoid that they could potentially consume
   * all the time allocated to block creation.
   *
   * @param evaluationContext The current selection context.
   * @return TX_UNPROFITABLE_UPFRONT if the transaction is not profitable upfront,
   *     TX_UNPROFITABLE_RETRY_LIMIT if the transaction was already found to be unprofitable, and
   *     there are no more slot to retry past unprofitable transactions during this block creation
   *     process, otherwise SELECTED.
   */
  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext) {

    final Wei minGasPrice = evaluationContext.getMinGasPrice();

    if (!evaluationContext.getPendingTransaction().hasPriority()) {
      final Transaction transaction = evaluationContext.getPendingTransaction().getTransaction();
      final long gasLimit = transaction.getGasLimit();

      // check the upfront profitability using the gas limit of the tx
      if (!transactionProfitabilityCalculator.isProfitable(
          "PreProcessing",
          transaction,
          profitabilityConf.minMargin(),
          baseFee,
          evaluationContext.getTransactionGasPrice(),
          gasLimit,
          minGasPrice)) {
        return TX_UNPROFITABLE_UPFRONT;
      }

      if (unprofitableCache.contains(transaction.getHash())) {
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
      }
    }

    return SELECTED;
  }

  /**
   * Evaluates a transaction post-processing. Checks if it is profitable according to its gas used.
   * If unprofitable, the transaction is added to the unprofitable cache, to be retried in the
   * future, since gas price market fluctuations can make it profitable again.
   *
   * @param evaluationContext The current selection context.
   * @return TX_UNPROFITABLE if the transaction is not profitable after execution, otherwise
   *     SELECTED.
   */
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
          baseFee,
          evaluationContext.getTransactionGasPrice(),
          gasUsed,
          evaluationContext.getMinGasPrice())) {
        rememberUnprofitable(transaction);
        return TX_UNPROFITABLE;
      }
    }
    return SELECTED;
  }

  /**
   * If the transaction has been selected for block inclusion, then we remove it from the
   * unprofitable cache.
   *
   * @param evaluationContext The current selection context
   * @param processingResult The result of processing the selected transaction.
   */
  @Override
  public void onTransactionSelected(
      final TransactionEvaluationContext<? extends PendingTransaction> evaluationContext,
      final TransactionProcessingResult processingResult) {
    unprofitableCache.remove(evaluationContext.getPendingTransaction().getTransaction().getHash());
  }

  /**
   * If the transaction has not been selected and has been discarded from the transaction pool, then
   * we remove it from the unprofitable cache.
   *
   * @param evaluationContext The current selection context
   * @param transactionSelectionResult The transaction selection result
   */
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
