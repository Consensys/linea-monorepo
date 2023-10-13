/*
 * Copyright ConsenSys AG.
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

import static net.consensys.linea.sequencer.txselection.selectors.MaxTransactionCallDataTransactionSelector.CALL_DATA_TOO_BIG_INVALID_REASON;
import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import net.consensys.linea.sequencer.txselection.selectors.MaxTransactionCallDataTransactionSelector;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class MaxTransactionCallDataSizeTransactionSelectorTest {
  private static final int MAX_TX_CALL_DATA_SIZE = 10;
  private PluginTransactionSelector transactionSelector;

  @BeforeEach
  public void initialize() {
    transactionSelector = new MaxTransactionCallDataTransactionSelector(MAX_TX_CALL_DATA_SIZE);
  }

  @Test
  public void shouldSelectTransactionWhen_CallDataSize_IsLessThan_MaxTxCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(MAX_TX_CALL_DATA_SIZE - 1);
    verifyTransactionSelection(
        transactionSelector, mockTransaction, TransactionSelectionResult.SELECTED);
  }

  @Test
  public void shouldSelectTransactionWhen_CallDataSize_IsEqualTo_MaxTxCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(MAX_TX_CALL_DATA_SIZE);
    verifyTransactionSelection(
        transactionSelector, mockTransaction, TransactionSelectionResult.SELECTED);
  }

  @Test
  public void shouldNotSelectTransactionWhen_CallDataSize_IsGreaterThan_MaxTxCallDataSize() {
    var mockTransaction = mockTransactionOfCallDataSize(MAX_TX_CALL_DATA_SIZE + 1);
    verifyTransactionSelection(
        transactionSelector,
        mockTransaction,
        TransactionSelectionResult.invalid(CALL_DATA_TOO_BIG_INVALID_REASON));
  }

  private void verifyTransactionSelection(
      final PluginTransactionSelector selector,
      final PendingTransaction transaction,
      final TransactionSelectionResult expectedSelectionResult) {
    var selectionResult = selector.evaluateTransactionPreProcessing(transaction);
    assertThat(selectionResult).isEqualTo(expectedSelectionResult);
  }

  private PendingTransaction mockTransactionOfCallDataSize(final int size) {
    PendingTransaction mockTransaction = mock(PendingTransaction.class);
    Transaction transaction = mock(Transaction.class);
    when(mockTransaction.getTransaction()).thenReturn(transaction);
    when(transaction.getPayload()).thenReturn(Bytes.repeat((byte) 1, size));
    return mockTransaction;
  }
}
