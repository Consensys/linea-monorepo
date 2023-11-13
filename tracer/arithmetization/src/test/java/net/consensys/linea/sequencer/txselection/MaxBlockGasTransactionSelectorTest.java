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

import static net.consensys.linea.sequencer.txselection.selectors.MaxBlockGasTransactionSelector.TRANSACTION_GAS_EXCEEDS_MAX_BLOCK_GAS;
import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import net.consensys.linea.sequencer.txselection.selectors.MaxBlockGasTransactionSelector;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class MaxBlockGasTransactionSelectorTest {
  private static final int MAX_GAS_PER_BLOCK = 1000;
  private static final int MAX_GAS_PER_BLOCK_20_PERCENTAGE = 200;
  private static final int MAX_GAS_PER_BLOCK_80_PERCENTAGE = 800;

  private PluginTransactionSelector transactionSelector;

  @BeforeEach
  public void initialize() {
    transactionSelector = new MaxBlockGasTransactionSelector(MAX_GAS_PER_BLOCK);
  }

  @Test
  public void shouldSelectWhen_GasUsedByTransaction_IsLessThan_MaxGasPerBlock() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(MAX_GAS_PER_BLOCK - 1);
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult,
        TransactionSelectionResult.SELECTED);
  }

  @Test
  public void shouldSelectWhen_GasUsedByTransaction_IsEqual_MaxGasPerBlock() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(MAX_GAS_PER_BLOCK);
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult,
        TransactionSelectionResult.SELECTED);
  }

  @Test
  public void shouldNotSelectWhen_GasUsedByTransaction_IsGreaterThan_MaxGasPerBlock() {
    var mockTransactionProcessingResult = mockTransactionProcessingResult(MAX_GAS_PER_BLOCK + 1);
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult,
        TransactionSelectionResult.invalid(TRANSACTION_GAS_EXCEEDS_MAX_BLOCK_GAS));
  }

  @Test
  public void shouldNotSelectWhen_CumulativeGasUsed_IsGreaterThan_MaxGasPerBlock() {
    // block empty, transaction 80% max gas, should select
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult(MAX_GAS_PER_BLOCK_80_PERCENTAGE),
        TransactionSelectionResult.SELECTED);

    // block 80% full, transaction 80% max gas, should not select
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult(MAX_GAS_PER_BLOCK_80_PERCENTAGE),
        TransactionSelectionResult.TX_TOO_LARGE_FOR_REMAINING_GAS);

    // block 80% full, transaction 20% max gas, should select
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction(),
        mockTransactionProcessingResult(MAX_GAS_PER_BLOCK_20_PERCENTAGE),
        TransactionSelectionResult.SELECTED);
  }

  private void verifyTransactionSelection(
      final PluginTransactionSelector selector,
      final PendingTransaction transaction,
      final TransactionProcessingResult processingResult,
      final TransactionSelectionResult expectedSelectionResult) {
    var selectionResult = selector.evaluateTransactionPostProcessing(transaction, processingResult);
    assertThat(selectionResult).isEqualTo(expectedSelectionResult);
    notifySelector(selector, transaction, processingResult, selectionResult);
  }

  private PendingTransaction mockTransaction() {
    PendingTransaction mockTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(mockTransaction.getTransaction()).thenReturn(transaction);
    return mockTransaction;
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
      final PendingTransaction transaction,
      final TransactionProcessingResult processingResult,
      final TransactionSelectionResult selectionResult) {
    if (selectionResult.equals(TransactionSelectionResult.SELECTED)) {
      selector.onTransactionSelected(transaction, processingResult);
    } else {
      selector.onTransactionNotSelected(transaction, selectionResult);
    }
  }
}
