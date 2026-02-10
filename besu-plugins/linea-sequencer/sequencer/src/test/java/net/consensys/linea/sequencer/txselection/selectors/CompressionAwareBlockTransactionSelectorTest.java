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

import linea.blob.BlobCompressorVersion;
import linea.blob.GoBackedBlobCompressor;
import net.consensys.linea.utils.CachingTransactionCompressor;
import net.consensys.linea.utils.TestTransactionFactory;
import net.consensys.linea.utils.TransactionCompressor;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class CompressionAwareBlockTransactionSelectorTest {

  /** A small data limit sufficient for test transactions. */
  private static final int TEST_DATA_LIMIT = 4 * 1024;

  /** Block header overhead used in tests. */
  private static final int TEST_HEADER_OVERHEAD = 512;

  private static final TransactionCompressor TX_COMPRESSOR = new CachingTransactionCompressor();

  private SelectorsStateManager selectorsStateManager;
  private TestTransactionFactory txFactory;

  @BeforeEach
  void setUp() {
    selectorsStateManager = new SelectorsStateManager();
    txFactory = new TestTransactionFactory();
  }

  /**
   * Re-initializes the shared compressor with a new data limit. Calling {@code getInstance} with a
   * new limit re-runs {@code Init} on the underlying native singleton and returns a wrapper that
   * uses it.
   */
  private static GoBackedBlobCompressor compressorWithLimit(final int dataLimit) {
    return GoBackedBlobCompressor.getInstance(BlobCompressorVersion.V1_2, dataLimit);
  }

  @Test
  void fastPathSelectsTransactionsWhenCumulativeCompressedSizeBelowLimit() {
    final int blobSizeLimit = TEST_DATA_LIMIT;
    final var selector = createSelector(blobSizeLimit);

    final var tx1 = txFactory.createTransaction();
    final var tx2 = txFactory.createTransaction();
    final var tx3 = txFactory.createTransaction();

    verifySelection(selector, wrapTx(tx1), SELECTED);
    verifySelection(selector, wrapTx(tx2), SELECTED);
    verifySelection(selector, wrapTx(tx3), SELECTED);
  }

  @Test
  void rejectsWhenBlockCannotFitMoreTransactions() {
    final var probeTx = txFactory.createTransaction();
    final int perTxCompressed = TX_COMPRESSOR.getCompressedSize(probeTx);

    // Use a small header overhead so we can test with a tight blob size limit.
    // The blob size limit must be > header overhead, but we want it smaller than one tx.
    final int smallHeaderOverhead = 10;
    final int blobSizeLimit = smallHeaderOverhead + perTxCompressed - 1;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(blobSizeLimit, smallHeaderOverhead);

    final var tx = txFactory.createTransaction();
    verifySelection(selector, wrapTx(tx), BLOCK_COMPRESSED_SIZE_OVERFLOW);
  }

  @Test
  void eventuallyFillsBlockAndRejects() {
    final var probeTx = txFactory.createTransaction();
    final int perTxCompressed = TX_COMPRESSOR.getCompressedSize(probeTx);

    // Use a small header overhead so we can test with a tight blob size limit.
    final int smallHeaderOverhead = 10;
    final int blobSizeLimit = smallHeaderOverhead + perTxCompressed * 5;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(blobSizeLimit, smallHeaderOverhead);

    int selectedCount = 0;
    boolean rejected = false;
    for (int i = 0; i < 100 && !rejected; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      final var processingResult = mock(TransactionProcessingResult.class);
      selector.evaluateTransactionPostProcessing(context, processingResult);
      if (preResult.equals(SELECTED)) {
        selector.onTransactionSelected(context, processingResult);
        selectedCount++;
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
  void slowPathMaximisesTransactionsAboveFastPathLimit() {
    // Measure per-tx compressed size and compute how many txs the fast path alone would allow.
    // The fast path budget is: blobSizeLimit - headerOverhead.
    // With the two-phase selector, the slow path should allow strictly more txs.
    final var probeTx = txFactory.createTransaction();
    final int perTxCompressed = TX_COMPRESSOR.getCompressedSize(probeTx);

    // Use a limit where the fast path alone allows exactly N txs:
    //   fastPathLimit = blobSizeLimit - headerOverhead
    //   N = floor(fastPathLimit / perTxCompressed)
    // We pick blobSizeLimit so N is a small known number.
    final int fastPathTxCount = 3;
    final int blobSizeLimit = perTxCompressed * fastPathTxCount + TEST_HEADER_OVERHEAD + 1;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(blobSizeLimit);

    int selectedCount = 0;
    for (int i = 0; i < 100; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      final var processingResult = mock(TransactionProcessingResult.class);
      selector.evaluateTransactionPostProcessing(context, processingResult);
      if (preResult.equals(SELECTED)) {
        selector.onTransactionSelected(context, processingResult);
        selectedCount++;
      } else {
        selector.onTransactionNotSelected(context, preResult);
        break;
      }
    }

    // The slow path (canAppendBlock) should allow strictly more than the fast path alone.
    // Fast path alone would stop after fastPathTxCount txs; slow path should squeeze in more.
    assertThat(selectedCount)
        .as(
            "Slow path should allow more than %d txs that the fast path alone would permit",
            fastPathTxCount)
        .isGreaterThan(fastPathTxCount);
  }

  /**
   * Creates a selector with the compressor re-initialized to the given limit, mirroring production
   * behavior where the compressor's dataLimit matches the selector's blobSizeLimit.
   */
  private CompressionAwareBlockTransactionSelector createSelector(final int blobSizeLimit) {
    return createSelector(blobSizeLimit, TEST_HEADER_OVERHEAD);
  }

  private CompressionAwareBlockTransactionSelector createSelector(
      final int blobSizeLimit, final int headerOverhead) {
    final var compressor = compressorWithLimit(blobSizeLimit);
    selectorsStateManager = new SelectorsStateManager();
    final var selector =
        new CompressionAwareBlockTransactionSelector(
            selectorsStateManager, blobSizeLimit, headerOverhead, TX_COMPRESSOR, compressor);
    selectorsStateManager.blockSelectionStarted();
    return selector;
  }

  private void verifySelection(
      final CompressionAwareBlockTransactionSelector selector,
      final TestTransactionEvaluationContext context,
      final TransactionSelectionResult expectedPreResult) {
    final var preResult = selector.evaluateTransactionPreProcessing(context);
    assertThat(preResult).isEqualTo(expectedPreResult);
    final var processingResult = mock(TransactionProcessingResult.class);
    selector.evaluateTransactionPostProcessing(context, processingResult);
    if (expectedPreResult.equals(SELECTED)) {
      selector.onTransactionSelected(context, processingResult);
    } else {
      selector.onTransactionNotSelected(context, preResult);
    }
  }

  private TestTransactionEvaluationContext wrapTx(final Transaction tx) {
    final PendingTransaction pendingTx = mock(PendingTransaction.class);
    when(pendingTx.getTransaction()).thenReturn(tx);
    return new TestTransactionEvaluationContext(mockBlockHeader(), pendingTx);
  }

  private ProcessableBlockHeader mockBlockHeader() {
    final ProcessableBlockHeader header = mock(ProcessableBlockHeader.class);
    when(header.getNumber()).thenReturn(1L);
    when(header.getTimestamp()).thenReturn(1700000000L);
    when(header.getCoinbase()).thenReturn(Address.ZERO);
    when(header.getGasLimit()).thenReturn(30_000_000L);
    when(header.getParentHash()).thenReturn(Hash.wrap(Bytes32.ZERO));
    return header;
  }
}
