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

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import net.consensys.linea.sequencer.txselection.selectors.MaxBlockCallDataTransactionSelector;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class MaxBlockCallDataSizeTransactionSelectorTest {
  private static final int BLOCK_CALL_DATA_MAX_SIZE = 100;
  private static final int BLOCK_CALL_DATA_HALF_SIZE = 50;
  private static final int TX_CALL_DATA_SIZE = BLOCK_CALL_DATA_HALF_SIZE + 1;
  private PluginTransactionSelector transactionSelector;

  @BeforeEach
  public void initialize() {
    transactionSelector = new MaxBlockCallDataTransactionSelector(BLOCK_CALL_DATA_MAX_SIZE);
  }

  @Test
  public void shouldSelectTransactionWhen_BlockCallDataSize_IsLessThan_MaxBlockCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE);
    var mockTransaction2 = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE - 1);
    verifyTransactionSelection(
        transactionSelector, mockTransaction, TransactionSelectionResult.SELECTED);
    verifyTransactionSelection(
        transactionSelector, mockTransaction2, TransactionSelectionResult.SELECTED);
  }

  @Test
  public void shouldSelectTransactionWhen_BlockCallDataSize_IsEqualTo_MaxBlockCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE);
    var mockTransaction2 = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE);
    verifyTransactionSelection(
        transactionSelector, mockTransaction, TransactionSelectionResult.SELECTED);
    verifyTransactionSelection(
        transactionSelector, mockTransaction2, TransactionSelectionResult.SELECTED);
  }

  @Test
  public void
      shouldNotSelectTransactionWhen_BlockCallDataSize_IsGreaterThan_MaxBlockCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE);
    var mockTransaction2 = mockTransactionOfCallDataSize(BLOCK_CALL_DATA_HALF_SIZE + 1);
    verifyTransactionSelection(
        transactionSelector, mockTransaction, TransactionSelectionResult.SELECTED);
    verifyTransactionSelection(
        transactionSelector, mockTransaction2, TransactionSelectionResult.BLOCK_FULL);
  }

  @Test
  public void shouldNotSelectAdditionalTransactionOnceBlockIsFull() {
    var firstTransaction = mockTransactionOfCallDataSize(TX_CALL_DATA_SIZE);
    var secondTransaction = mockTransactionOfCallDataSize(TX_CALL_DATA_SIZE);
    var thirdTransaction = mockTransactionOfCallDataSize(TX_CALL_DATA_SIZE);

    verifyTransactionSelection(
        transactionSelector, firstTransaction, TransactionSelectionResult.SELECTED);
    verifyTransactionSelection(
        transactionSelector, secondTransaction, TransactionSelectionResult.BLOCK_FULL);
    verifyTransactionSelection(
        transactionSelector, thirdTransaction, TransactionSelectionResult.BLOCK_FULL);
  }

  private void verifyTransactionSelection(
      final PluginTransactionSelector selector,
      final PendingTransaction transaction,
      final TransactionSelectionResult expectedSelectionResult) {
    var selectionResult = selector.evaluateTransactionPreProcessing(transaction);
    assertThat(selectionResult).isEqualTo(expectedSelectionResult);
    notifySelector(selector, transaction, selectionResult);
  }

  private void notifySelector(
      final PluginTransactionSelector selector,
      final PendingTransaction transaction,
      final TransactionSelectionResult selectionResult) {
    if (selectionResult.equals(TransactionSelectionResult.SELECTED)) {
      selector.onTransactionSelected(transaction, mock(TransactionProcessingResult.class));
    } else {
      selector.onTransactionNotSelected(transaction, selectionResult);
    }
  }

  private PendingTransaction mockTransactionOfCallDataSize(final int size) {
    PendingTransaction mockTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(mockTransaction.getTransaction()).thenReturn(transaction);
    when(transaction.getPayload()).thenReturn(Bytes.repeat((byte) 1, size));
    return mockTransaction;
  }
}
