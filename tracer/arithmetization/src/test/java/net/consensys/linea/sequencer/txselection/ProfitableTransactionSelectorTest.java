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
package net.consensys.linea.sequencer.txselection;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.TX_UNPROFITABLE;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import net.consensys.linea.sequencer.txselection.selectors.ProfitableTransactionSelector;
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

  private PluginTransactionSelector transactionSelector;

  @BeforeEach
  public void initialize() {
    transactionSelector =
        new ProfitableTransactionSelector(
            VERIFICATION_GAS_COST, VERIFICATION_CAPACITY, GAS_PRICE_RATIO, MIN_MARGIN);
  }

  @Test
  public void shouldSelectWhenProfitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 100, Wei.of(1_100_000_000), Wei.of(1_000_000_000)),
        mockTransactionProcessingResult,
        SELECTED);
  }

  @Test
  public void shouldNotSelectWhenUnprofitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(false, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000)),
        mockTransactionProcessingResult,
        TX_UNPROFITABLE);
  }

  @Test
  public void shouldSelectPrevUnprofitableAfterGasPriceBump() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(
            false, 1000, Wei.of(1_100_000_000).multiply(9), Wei.of(1_000_000_000)),
        mockTransactionProcessingResult,
        SELECTED);
  }

  @Test
  public void shouldSelectPriorityTxEvenWhenUnprofitable() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(21000);
    verifyTransactionSelection(
        transactionSelector,
        mockEvaluationContext(true, 1000, Wei.of(1_100_000_000), Wei.of(1_000_000_000)),
        mockTransactionProcessingResult,
        SELECTED);
  }

  private void verifyTransactionSelection(
      final PluginTransactionSelector selector,
      final TestTransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult,
      final TransactionSelectionResult expectedSelectionResult) {
    var selectionResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);
    assertThat(selectionResult).isEqualTo(expectedSelectionResult);
    notifySelector(selector, evaluationContext, processingResult, selectionResult);
  }

  private TestTransactionEvaluationContext mockEvaluationContext(
      final boolean hasPriority,
      final int size,
      final Wei effectiveGasPrice,
      final Wei minGasPrice) {
    PendingTransaction pendingTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(transaction.getSize()).thenReturn(size);
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
}
