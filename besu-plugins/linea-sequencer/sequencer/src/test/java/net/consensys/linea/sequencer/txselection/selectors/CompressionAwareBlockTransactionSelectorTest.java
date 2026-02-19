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
import net.consensys.linea.utils.TxEncodingUtils;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class CompressionAwareBlockTransactionSelectorTest {

  /** A small data limit sufficient for test transactions. */
  private static final int TEST_DATA_LIMIT = 4 * 1024;

  private TestTransactionFactory txFactory;
  private TxCompressor txCompressor;

  @BeforeEach
  void setUp() {
    txFactory = new TestTransactionFactory();
    txCompressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, TEST_DATA_LIMIT);
    txCompressor.reset();
  }

  /**
   * Re-initializes the shared compressor with a new data limit. Calling {@code getInstance} with a
   * new limit re-runs {@code Init} on the underlying native singleton and returns a wrapper that
   * uses it.
   */
  private TxCompressor compressorWithLimit(final int dataLimit) {
    final var compressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, dataLimit);
    compressor.reset();
    return compressor;
  }

  @Test
  void selectsTransactionsWhenCompressedSizeBelowLimit() {
    final var selector = createSelector(TEST_DATA_LIMIT);

    final var tx1 = txFactory.createTransaction();
    final var tx2 = txFactory.createTransaction();
    final var tx3 = txFactory.createTransaction();

    verifySelection(selector, wrapTx(tx1), SELECTED);
    verifySelection(selector, wrapTx(tx2), SELECTED);
    verifySelection(selector, wrapTx(tx3), SELECTED);
  }

  @Test
  void rejectsWhenBlockCannotFitMoreTransactions() {
    final int tinyLimit = 40;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(tinyLimit);

    final var tx = txFactory.createTransaction();
    verifySelection(selector, wrapTx(tx), BLOCK_COMPRESSED_SIZE_OVERFLOW);
  }

  @Test
  void eventuallyFillsBlockAndRejects() {
    // Measure per-tx compressed size using stateless compression
    final var probeTx = txFactory.createTransaction();
    final int perTxCompressed = getStatelessCompressedSize(probeTx);

    // Use a limit that allows approximately 5 transactions
    final int blobSizeLimit = perTxCompressed * 5;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(blobSizeLimit);

    int selectedCount = 0;
    boolean rejected = false;
    for (int i = 0; i < 100 && !rejected; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      if (preResult.equals(SELECTED)) {
        final var processingResult = mock(TransactionProcessingResult.class);
        final var postResult =
            selector.evaluateTransactionPostProcessing(context, processingResult);
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
    // Measure per-tx compressed size using stateless compression
    final var probeTx = txFactory.createTransaction();
    final int perTxCompressed = getStatelessCompressedSize(probeTx);

    // Use a limit where stateless compression would allow exactly N txs,
    // but context-aware compression should allow more
    final int statelessTxCount = 3;
    final int blobSizeLimit = perTxCompressed * statelessTxCount + 1;

    txFactory = new TestTransactionFactory();
    final var selector = createSelector(blobSizeLimit);

    int selectedCount = 0;
    for (int i = 0; i < 100; i++) {
      final var tx = txFactory.createTransaction();
      final var context = wrapTx(tx);
      final var preResult = selector.evaluateTransactionPreProcessing(context);
      if (preResult.equals(SELECTED)) {
        final var processingResult = mock(TransactionProcessingResult.class);
        final var postResult =
            selector.evaluateTransactionPostProcessing(context, processingResult);
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

    // The TxCompressor maintains context across transactions, allowing more txs to fit
    // than if each was compressed independently (stateless).
    assertThat(selectedCount)
        .as(
            "Context-aware compression should allow more than %d txs that stateless would permit",
            statelessTxCount)
        .isGreaterThan(statelessTxCount);
  }

  /** Gets the stateless compressed size of a transaction (without context from previous txs). */
  private int getStatelessCompressedSize(final Transaction tx) {
    final var tempCompressor = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, 128 * 1024);
    final byte[] encoded = TxEncodingUtils.encodeForCompressor(tx);
    return tempCompressor.compressedSize(encoded);
  }

  /** Creates a selector with the compressor re-initialized to the given limit. */
  private CompressionAwareBlockTransactionSelector createSelector(final int blobSizeLimit) {
    txCompressor = compressorWithLimit(blobSizeLimit);
    return new CompressionAwareBlockTransactionSelector(txCompressor, blobSizeLimit);
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
