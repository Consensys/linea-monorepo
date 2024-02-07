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
import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import net.consensys.linea.config.LineaTransactionSelectorCliOptions;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class ProfitableTransactionSelectorTest {
  private static final int VERIFICATION_GAS_COST = 1_200_000;
  private static final int VERIFICATION_CAPACITY = 90_000;
  private static final int GAS_PRICE_RATIO = 15;
  private static final double MIN_MARGIN = 1.0;
  private static final int ADJUST_TX_SIZE = -45;
  private static final int UNPROFITABLE_CACHE_SIZE = 2;
  private static final int UNPROFITABLE_RETRY_LIMIT = 1;
  private final LineaTransactionSelectorConfiguration conf =
      LineaTransactionSelectorCliOptions.create().toDomainObject().toBuilder()
          .gasPriceRatio(GAS_PRICE_RATIO)
          .adjustTxSize(ADJUST_TX_SIZE)
          .minMargin(MIN_MARGIN)
          .unprofitableCacheSize(UNPROFITABLE_CACHE_SIZE)
          .unprofitableRetryLimit(UNPROFITABLE_RETRY_LIMIT)
          .verificationCapacity(VERIFICATION_CAPACITY)
          .verificationGasCost(VERIFICATION_GAS_COST)
          .build();
  private TestableProfitableTransactionSelector transactionSelector;

  @BeforeEach
  public void initialize() {
    transactionSelector = newSelectorForNewBlock();
    transactionSelector.reset();
  }

  private TestableProfitableTransactionSelector newSelectorForNewBlock() {
    return new TestableProfitableTransactionSelector(conf);
  }

  @Test
  public void shouldSelectWhenProfitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000),
        mockTransactionProcessingResult,
        SELECTED,
        SELECTED);
  }

  @Test
  public void shouldSelectWhenProfitableWithAdjustedSize() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 150, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000),
        mockTransactionProcessingResult,
        SELECTED,
        SELECTED);
  }

  @Test
  public void shouldNotSelectWhenUnprofitableUpfront() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000),
        mockTransactionProcessingResult,
        TX_UNPROFITABLE_UPFRONT,
        null);
  }

  @Test
  public void shouldNotSelectWhenUnprofitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 210000),
        mockTransactionProcessingResult,
        SELECTED,
        TX_UNPROFITABLE);
  }

  @Test
  public void shouldSelectPrevUnprofitableAfterGasPriceBump() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(
            false, 1000, Wei.of(1_100_000_000).multiply(9), Wei.of(1_000_000_000), 210000),
        mockTransactionProcessingResult,
        SELECTED,
        SELECTED);
  }

  @Test
  public void shouldSelectPriorityTxEvenWhenUnprofitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(true, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 21000),
        mockTransactionProcessingResult,
        SELECTED,
        SELECTED);
  }

  @Test
  public void shouldRetryUnprofitableTxWhenBelowLimit() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    var mockEvaluationContext =
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 210000);
    // first try
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext,
        mockTransactionProcessingResult,
        SELECTED,
        TX_UNPROFITABLE);
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    // simulate another block
    newSelectorForNewBlock();
    // we should remember of the unprofitable tx
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    // second try because we are below the retry limit
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext,
        mockTransactionProcessingResult,
        SELECTED,
        TX_UNPROFITABLE);
  }

  @Test
  public void shouldEvictWhenUnprofitableCacheIsFull() {
    final TestTransactionEvaluationContext[] evaluationContexts =
        new TestTransactionEvaluationContext[UNPROFITABLE_CACHE_SIZE + 1];
    for (int i = 0; i <= UNPROFITABLE_CACHE_SIZE; i++) {
      var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
      var mockEvaluationContext =
          mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000), 210000);
      evaluationContexts[i] = mockEvaluationContext;
      verifyTransactionSelection(
          transactionSelector,
          mockEvaluationContext,
          mockTransactionProcessingResult,
          SELECTED,
          TX_UNPROFITABLE);
      assertThat(
              transactionSelector.isUnprofitableTxCached(
                  mockEvaluationContext.getPendingTransaction().getTransaction().getHash()))
          .isTrue();
    }
    // only the last two txs must be in the unprofitable cache, since the first one was evicted
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                evaluationContexts[0].getPendingTransaction().getTransaction().getHash()))
        .isFalse();
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                evaluationContexts[1].getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                evaluationContexts[2].getPendingTransaction().getTransaction().getHash()))
        .isTrue();
  }

  @Test
  public void shouldNotRetryUnprofitableTxWhenRetryLimitReached() {
    var minGasPriceBlock1 = Wei.of(1_000_000_000);
    var mockTransactionProcessingResult1 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext1 =
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), minGasPriceBlock1, 210000);
    // first try of first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1,
        mockTransactionProcessingResult1,
        SELECTED,
        TX_UNPROFITABLE);

    var mockTransactionProcessingResult2 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext2 =
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), minGasPriceBlock1, 210000);
    // first try of second tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext2,
        mockTransactionProcessingResult2,
        SELECTED,
        TX_UNPROFITABLE);

    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext2.getPendingTransaction().getTransaction().getHash()))
        .isTrue();

    // simulate another block
    transactionSelector = newSelectorForNewBlock();
    // we need to decrease the min gas price in order to allow a retry
    var minGasPriceBlock2 = Wei.of(1_000_000_000).subtract(1);

    // we should remember of the unprofitable txs for the new block
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext2.getPendingTransaction().getTransaction().getHash()))
        .isTrue();

    // second try of the first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1.setMinGasPrice(minGasPriceBlock2),
        mockTransactionProcessingResult1,
        SELECTED,
        TX_UNPROFITABLE);

    // second try of the second tx is not retried since we reached the retry limit
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext2.setMinGasPrice(minGasPriceBlock2),
        mockTransactionProcessingResult2,
        TX_UNPROFITABLE_RETRY_LIMIT,
        null);
  }

  @Test
  public void shouldNotRetryUnprofitableTxWhenMinGasPriceNotDecreased() {
    var minGasPriceBlock1 = Wei.of(1_000_000_000);
    var mockTransactionProcessingResult1 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext1 =
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), minGasPriceBlock1, 210000);
    // first try of first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1,
        mockTransactionProcessingResult1,
        SELECTED,
        TX_UNPROFITABLE);

    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();

    // simulate another block
    transactionSelector = newSelectorForNewBlock();
    // we keep the min gas price the same to avoid retry
    var minGasPriceBlock2 = minGasPriceBlock1;

    // we should remember of the unprofitable txs for the new block
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();

    // second try of the first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1.setMinGasPrice(minGasPriceBlock2),
        mockTransactionProcessingResult1,
        TX_UNPROFITABLE_MIN_GAS_PRICE_NOT_DECREASED,
        null);
  }

  @Test
  public void profitableAndUnprofitableTxsMix() {
    var minGasPriceBlock1 = Wei.of(1_000_000_000);
    var mockTransactionProcessingResult1 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext1 =
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), minGasPriceBlock1, 210000);
    // first try of first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1,
        mockTransactionProcessingResult1,
        SELECTED,
        TX_UNPROFITABLE);

    var mockTransactionProcessingResult2 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext2 =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), minGasPriceBlock1, 210000);
    // first try of second tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext2,
        mockTransactionProcessingResult2,
        SELECTED,
        SELECTED);

    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext2.getPendingTransaction().getTransaction().getHash()))
        .isFalse();

    // simulate another block
    transactionSelector = newSelectorForNewBlock();
    // we keep the min gas price the same to avoid retry
    var minGasPriceBlock2 = minGasPriceBlock1;

    // we should remember of the unprofitable txs for the new block
    assertThat(
            transactionSelector.isUnprofitableTxCached(
                mockEvaluationContext1.getPendingTransaction().getTransaction().getHash()))
        .isTrue();

    // second try of the first tx
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext1.setMinGasPrice(minGasPriceBlock2),
        mockTransactionProcessingResult1,
        TX_UNPROFITABLE_MIN_GAS_PRICE_NOT_DECREASED,
        null);

    var mockTransactionProcessingResult3 = mockTransactionProcessingResult(21000);
    var mockEvaluationContext3 =
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), minGasPriceBlock1, 21000);

    // new profitable tx is selected
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext3,
        mockTransactionProcessingResult3,
        SELECTED,
        SELECTED);
  }

  private void verifyTransactionSelection(
      final ProfitableTransactionSelector selector,
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

  private TransactionProcessingResult mockTransactionProcessingResult(long gasUsedByTransaction) {
    TransactionProcessingResult mockTransactionProcessingResult =
        mock(TransactionProcessingResult.class);
    when(mockTransactionProcessingResult.getEstimateGasUsedByTransaction())
        .thenReturn(gasUsedByTransaction);
    return mockTransactionProcessingResult;
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

  private static class TestableProfitableTransactionSelector extends ProfitableTransactionSelector {

    TestableProfitableTransactionSelector(final LineaTransactionSelectorConfiguration conf) {
      super(conf);
    }

    boolean isUnprofitableTxCached(final Hash txHash) {
      return unprofitableCache.contains(txHash);
    }

    void reset() {
      prevMinGasPrice = Wei.MAX_WEI;
      unprofitableCache.clear();
    }
  }
}
