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
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;

import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

public class TraceLineLimitTransactionSelectorTest {
  private static final int OVER_LINE_COUNT_LIMIT_CACHE_SIZE = 2;
  private static final String MODULE_LINE_LIMITS_RESOURCE_NAME = "/sequencer/line-limits.toml";
  private Map<String, Integer> lineCountLimits;
  private LineaTracerConfiguration lineaTracerConfiguration;
  private SelectorsStateManager selectorsStateManager;

  @TempDir static Path tempDir;
  static Path lineLimitsConfPath;

  @BeforeAll
  public static void beforeAll() throws IOException {
    lineLimitsConfPath = tempDir.resolve("line-limits.toml");
    Files.copy(
        TraceLineLimitTransactionSelectorTest.class.getResourceAsStream(
            MODULE_LINE_LIMITS_RESOURCE_NAME),
        lineLimitsConfPath);
  }

  @BeforeEach
  public void initialize() {
    lineaTracerConfiguration =
        LineaTracerConfiguration.builder()
            .moduleLimitsFilePath(lineLimitsConfPath.toString())
            .build();
    lineCountLimits =
        new HashMap<>(ModuleLineCountValidator.createLimitModules(lineaTracerConfiguration));
  }

  private TestableTraceLineLimitTransactionSelector newSelectorForNewBlock(
      final Map<String, Integer> lineCountLimits) {
    selectorsStateManager = new SelectorsStateManager();
    final var selector =
        new TestableTraceLineLimitTransactionSelector(
            selectorsStateManager,
            lineaTracerConfiguration,
            lineCountLimits,
            OVER_LINE_COUNT_LIMIT_CACHE_SIZE);
    selectorsStateManager.blockSelectionStarted();
    return selector;
  }

  @Test
  public void shouldSelectWhenBelowLimits() {
    final var transactionSelector = newSelectorForNewBlock(lineCountLimits);
    transactionSelector.resetCache();

    final var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000, 0);
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
    lineCountLimits.put("EXT", 5);
    final var transactionSelector = newSelectorForNewBlock(lineCountLimits);
    transactionSelector.resetCache();

    final var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000, 0);
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
    lineCountLimits.put("EXT", 5);
    var transactionSelector = newSelectorForNewBlock(lineCountLimits);
    transactionSelector.resetCache();

    var evaluationContext =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000, 0);
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
    transactionSelector = newSelectorForNewBlock(lineCountLimits);
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
    lineCountLimits.put("EXT", 5);
    final var transactionSelector = newSelectorForNewBlock(lineCountLimits);
    transactionSelector.resetCache();

    final TestTransactionEvaluationContext[] evaluationContexts =
        new TestTransactionEvaluationContext[OVER_LINE_COUNT_LIMIT_CACHE_SIZE + 1];
    for (int i = 0; i <= OVER_LINE_COUNT_LIMIT_CACHE_SIZE; i++) {
      var evaluationContext =
          mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000, 0);
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

    // only the last two txs must be in the over limit cache, since the first one was evicted
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
      final long gasLimit,
      final int payloadSize) {
    PendingTransaction pendingTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(transaction.getHash()).thenReturn(Hash.wrap(Bytes32.random()));
    when(transaction.getSize()).thenReturn(size);
    when(transaction.getGasLimit()).thenReturn(gasLimit);
    when(transaction.getPayload()).thenReturn(Bytes.repeat((byte) 1, payloadSize));
    when(pendingTransaction.getTransaction()).thenReturn(transaction);
    when(pendingTransaction.hasPriority()).thenReturn(hasPriority);
    return new TestTransactionEvaluationContext(
        mock(ProcessableBlockHeader.class), pendingTransaction, effectiveGasPrice, minGasPrice);
  }

  private class TestableTraceLineLimitTransactionSelector
      extends TraceLineLimitTransactionSelector {
    TestableTraceLineLimitTransactionSelector(
        final SelectorsStateManager selectorsStateManager,
        final LineaTracerConfiguration lineaTracerConfiguration,
        final Map<String, Integer> moduleLimits,
        final int overLimitCacheSize) {
      super(
          selectorsStateManager,
          BigInteger.ONE,
          moduleLimits,
          LineaTransactionSelectorConfiguration.builder()
              .overLinesLimitCacheSize(overLimitCacheSize)
              .build(),
          LineaL1L2BridgeSharedConfiguration.builder()
              .contract(Address.fromHexString("0xDEADBEEF"))
              .topic(Bytes.fromHexString("0x012345"))
              .build(),
          lineaTracerConfiguration);
    }

    void resetCache() {
      overLineCountLimitCache.clear();
    }

    boolean isOverLineCountLimitTxCached(final Hash txHash) {
      return overLineCountLimitCache.contains(txHash);
    }
  }
}
