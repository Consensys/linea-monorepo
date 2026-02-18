/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txselection.selectors;

import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.BLOCK_COMPRESSED_SIZE_OVERFLOW;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import linea.blob.GoBackedTxCompressor;
import linea.blob.TxCompressor;
import linea.blob.TxCompressorVersion;
import net.consensys.linea.utils.TestTransactionFactory;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class CompressionAwareBlockTransactionSelectorTest {

  /** A small data limit sufficient for test transactions. */
  private static final int TEST_DATA_LIMIT = 4 * 1024;

  private SelectorsStateManager selectorsStateManager;
  private TestTransactionFactory txFactory;
  private TxCompressor txCompressor;

  @BeforeEach
  void setUp() {
    selectorsStateManager = new SelectorsStateManager();
    txFactory = new TestTransactionFactory();
    txCompressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, TEST_DATA_LIMIT);
    txCompressor.reset();
  }

  @Test
  void selectsTransactionsWhenCompressedSizeBelowLimit() {
    final var selector = createSelector();

    final var tx1 = txFactory.createTransaction();
    final var tx2 = txFactory.createTransaction();
    final var tx3 = txFactory.createTransaction();

    verifySelection(selector, wrapTx(tx1), SELECTED);
    verifySelection(selector, wrapTx(tx2), SELECTED);
    verifySelection(selector, wrapTx(tx3), SELECTED);
  }

  @Test
  void rejectsWhenBlockCannotFitMoreTransactions() {
    // Use a very small limit that won't fit even one transaction
    final int tinyLimit = 100;
    txCompressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, tinyLimit);
    txCompressor.reset();

    final var selector = createSelector();

    final var tx = txFactory.createTransaction();
    verifySelection(selector, wrapTx(tx), BLOCK_COMPRESSED_SIZE_OVERFLOW);
  }

  @Test
  void eventuallyFillsBlockAndRejects() {
    final var selector = createSelector();

    int selectedCount = 0;
    boolean rejected = false;
    for (int i = 0; i < 100 && !rejected; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      if (preResult.equals(SELECTED)) {
        final var processingResult = mock(TransactionProcessingResult.class);
        final var postResult = selector.evaluateTransactionPostProcessing(context, processingResult);
        if (postResult.equals(SELECTED)) {
          selector.onTransactionSelected(context, processingResult);
          selectedCount++;
        } else {
          selector.onTransactionNotSelected(context, postResult);
          assertThat(postResult).isEqualTo(BLOCK_COMPRESSED_SIZE_OVERFLOW);
          rejected = true;
        }
      } else {
        selector.onTransactionNotSelected(context, preResult);
        assertThat(preResult).isEqualTo(BLOCK_COMPRESSED_SIZE_OVERFLOW);
        rejected = true;
      }
    }
    assertThat(selectedCount).isGreaterThanOrEqualTo(1);
    assertThat(rejected).isTrue();
  }

  @Test
  void compressorMaintainsContextAcrossTransactions() {
    // The TxCompressor should benefit from compression context across transactions,
    // allowing more transactions to fit than if each was compressed independently.
    final var selector = createSelector();

    int selectedCount = 0;
    for (int i = 0; i < 100; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      if (preResult.equals(SELECTED)) {
        final var processingResult = mock(TransactionProcessingResult.class);
        final var postResult = selector.evaluateTransactionPostProcessing(context, processingResult);
        if (postResult.equals(SELECTED)) {
          selector.onTransactionSelected(context, processingResult);
          selectedCount++;
        } else {
          break;
        }
      } else {
        break;
      }
    }

    // With 4KB limit and context-aware compression, we should fit multiple transactions
    assertThat(selectedCount)
        .as("TxCompressor should allow multiple transactions with context sharing")
        .isGreaterThan(1);
  }

  private CompressionAwareBlockTransactionSelector createSelector() {
    selectorsStateManager = new SelectorsStateManager();
    final var selector =
        new CompressionAwareBlockTransactionSelector(selectorsStateManager, txCompressor);
    selectorsStateManager.blockSelectionStarted();
    return selector;
  }

  private void verifySelection(
      final CompressionAwareBlockTransactionSelector selector,
      final TestTransactionEvaluationContext context,
      final TransactionSelectionResult expectedPreResult) {
    final var preResult = selector.evaluateTransactionPreProcessing(context);
    assertThat(preResult).isEqualTo(expectedPreResult);
    if (expectedPreResult.equals(SELECTED)) {
      final var processingResult = mock(TransactionProcessingResult.class);
      final var postResult = selector.evaluateTransactionPostProcessing(context, processingResult);
      assertThat(postResult).isEqualTo(SELECTED);
      selector.onTransactionSelected(context, processingResult);
    } else {
      selector.onTransactionNotSelected(context, preResult);
    }
  }

  private TestTransactionEvaluationContext wrapTx(final Transaction tx) {
    final PendingTransaction pendingTx = mock(PendingTransaction.class);
    when(pendingTx.getTransaction()).thenReturn(tx);
    return new TestTransactionEvaluationContext(pendingTx);
  }
}
