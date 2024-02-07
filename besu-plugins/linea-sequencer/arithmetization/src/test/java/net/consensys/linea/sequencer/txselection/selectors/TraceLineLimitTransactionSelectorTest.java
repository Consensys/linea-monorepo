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

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_MODULE_LINE_COUNT_OVERFLOW;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_MODULE_LINE_COUNT_OVERFLOW_CACHED;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.io.InputStream;
import java.util.Map;
import java.util.stream.Collectors;

import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.toml.Toml;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class TraceLineLimitTransactionSelectorTest {
  private static final int OVER_LINE_COUNT_LIMIT_CACHE_SIZE = 2;
  private TestableTraceLineLimitTransactionSelector transactionSelector;
  private Map<String, Integer> lineCountLimits;

  @BeforeEach
  public void initialize() {
    lineCountLimits = loadLineCountLimitConf();
    transactionSelector = newSelectorForNewBlock();
    transactionSelector.reset();
  }

  private TestableTraceLineLimitTransactionSelector newSelectorForNewBlock() {
    return new TestableTraceLineLimitTransactionSelector(
        lineCountLimits, "line-limits.toml", OVER_LINE_COUNT_LIMIT_CACHE_SIZE);
  }

  private Map<String, Integer> loadLineCountLimitConf() {
    try (final InputStream is =
        this.getClass().getResourceAsStream("/sequencer/line-limits.toml")) {
      return Toml.parse(is).getTable("traces-limits").toMap().entrySet().stream()
          .collect(Collectors.toMap(Map.Entry::getKey, e -> Math.toIntExact((long) e.getValue())));
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Test
  public void shouldSelectWhenBelowLimits() {
    final var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000);
    verifyTransactionSelection(
        transactionSelector,
        evaluationContext,
        mock(TransactionProcessingResult.class),
        SELECTED,
        SELECTED);
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isFalse();
  }

  @Test
  public void shouldNotSelectWhenOverLimits() {
    lineCountLimits.put("ADD", 1);
    final var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000);
    verifyTransactionSelection(
        transactionSelector,
        evaluationContext,
        mock(TransactionProcessingResult.class),
        SELECTED,
        TX_MODULE_LINE_COUNT_OVERFLOW);
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
  }

  @Test
  public void shouldNotReprocessedWhenOverLimits() {
    lineCountLimits.put("ADD", 1);
    var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000);
    verifyTransactionSelection(
        transactionSelector,
        evaluationContext,
        mock(TransactionProcessingResult.class),
        SELECTED,
        TX_MODULE_LINE_COUNT_OVERFLOW);

    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    transactionSelector = newSelectorForNewBlock();
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    // retrying the same tx should avoid reprocessing
    verifyTransactionSelection(
        transactionSelector,
        evaluationContext,
        mock(TransactionProcessingResult.class),
        TX_MODULE_LINE_COUNT_OVERFLOW_CACHED,
        null);
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
  }

  @Test
  public void shouldEvictWhenCacheIsFull() {
    lineCountLimits.put("ADD", 1);
    final TestTransactionEvaluationContext[] evaluationContexts =
        new TestTransactionEvaluationContext[OVER_LINE_COUNT_LIMIT_CACHE_SIZE + 1];
    for (int i = 0; i <= OVER_LINE_COUNT_LIMIT_CACHE_SIZE; i++) {
      var evaluationContext =
          mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000);
      verifyTransactionSelection(
          transactionSelector,
          evaluationContext,
          mock(TransactionProcessingResult.class),
          SELECTED,
          TX_MODULE_LINE_COUNT_OVERFLOW);
      evaluationContexts[i] = evaluationContext;
      assertThat(
              transactionSelector.isOverLineCountLimitTxCached(
                  evaluationContext.getPendingTransaction().getTransaction().getHash()))
          .isTrue();
    }

    // only the last two txs must be in the unprofitable cache, since the first one was evicted
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContexts[0].getPendingTransaction().getTransaction().getHash()))
        .isFalse();
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContexts[1].getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    assertThat(
            transactionSelector.isOverLineCountLimitTxCached(
                evaluationContexts[2].getPendingTransaction().getTransaction().getHash()))
        .isTrue();
  }

  private void verifyTransactionSelection(
      final TestableTraceLineLimitTransactionSelector selector,
      final TestTransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult,
      final TransactionSelectionResult expectedPreProcessingResult,
      final TransactionSelectionResult expectedPostProcessingResult) {
    var preProcessingResult = selector.evaluateTransactionPreProcessing(evaluationContext);
    assertThat(preProcessingResult).isEqualTo(expectedPreProcessingResult);
    if (preProcessingResult.equals(SELECTED)) {
      var postProcessingResult =
          selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);
      assertThat(postProcessingResult).isEqualTo(expectedPostProcessingResult);
      notifySelector(selector, evaluationContext, processingResult, postProcessingResult);
    } else {
      notifySelector(selector, evaluationContext, processingResult, preProcessingResult);
    }
  }

  private void notifySelector(
      final PluginTransactionSelector selector,
      final TestTransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult,
      final TransactionSelectionResult selectionResult) {
    if (selectionResult.equals(SELECTED)) {
      selector.onTransactionSelected(evaluationContext, processingResult);
    } else {
      selector.onTransactionNotSelected(evaluationContext, selectionResult);
    }
  }

  private TestTransactionEvaluationContext mockEvaluationContext(
      final boolean hasPriority,
      final int size,
      final Wei effectiveGasPrice,
      final Wei minGasPrice,
      final long gasLimit) {
    PendingTransaction pendingTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(transaction.getHash()).thenReturn(Hash.wrap(Bytes32.random()));
    when(transaction.getSize()).thenReturn(size);
    when(transaction.getGasLimit()).thenReturn(gasLimit);
    when(pendingTransaction.getTransaction()).thenReturn(transaction);
    when(pendingTransaction.hasPriority()).thenReturn(hasPriority);
    return new TestTransactionEvaluationContext(pendingTransaction, effectiveGasPrice, minGasPrice);
  }

  private class TestableTraceLineLimitTransactionSelector
      extends TraceLineLimitTransactionSelector {
    TestableTraceLineLimitTransactionSelector(
        final Map<String, Integer> moduleLimits,
        final String limitFilePath,
        final int overLimitCacheSize) {
      super(moduleLimits, limitFilePath, overLimitCacheSize);
    }

    void reset() {
      overLineCountLimitCache.clear();
    }

    boolean isOverLineCountLimitTxCached(final Hash txHash) {
      return overLineCountLimitCache.contains(txHash);
    }
  }
}
