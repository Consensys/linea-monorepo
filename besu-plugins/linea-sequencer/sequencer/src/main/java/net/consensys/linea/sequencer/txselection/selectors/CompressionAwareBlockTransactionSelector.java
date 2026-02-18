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
import static org.hyperledger.besu.plugin.data.TransactionSelectionResult.SELECTED;

import linea.blob.TxCompressor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.AbstractStatefulPluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

/**
 * Stateful transaction selector that enforces a compressed block size limit (blob fit) using the
 * {@link TxCompressor}.
 *
 * <p>The {@link TxCompressor} is additive: it maintains compression context across transactions,
 * allowing efficient incremental size checks. Each {@link TxCompressor#canAppendTransaction} call
 * is fast because the compressor already holds the compressed state of all previously accepted
 * transactions.
 *
 * <p>For each candidate transaction, the selector checks if it fits without actually appending it.
 * After successful EVM execution, the transaction is permanently appended to the compressor.
 *
 * <p>All transaction types (regular, forced, bundle, liveness) flow through the same selector
 * pipeline and are tracked automatically.
 */
@Slf4j
public class CompressionAwareBlockTransactionSelector
    extends AbstractStatefulPluginTransactionSelector<
        CompressionAwareBlockTransactionSelector.CompressionState> {

  private final TxCompressor txCompressor;

  public CompressionAwareBlockTransactionSelector(
      final SelectorsStateManager selectorsStateManager, final TxCompressor txCompressor) {
    super(selectorsStateManager, new CompressionState(0, 0), CompressionState::duplicate);
    this.txCompressor = txCompressor;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final byte[] rlpEncodedTx = encodeTransaction(transaction);

    final CompressionState state = getWorkingState();
    final int compressedSizeBefore = txCompressor.getCompressedSize();

    // Check if the transaction can be appended without exceeding the limit
    final boolean canAppend = txCompressor.canAppendTransaction(rlpEncodedTx);

    if (!canAppend) {
      log.atTrace()
          .setMessage(
              "event=tx_selection decision=reject reason=block_compressed_size_overflow "
                  + "tx_hash={} tx_rlp_size={} compressed_size_before={}")
          .addArgument(transaction::getHash)
          .addArgument(rlpEncodedTx.length)
          .addArgument(compressedSizeBefore)
          .log();
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    log.atTrace()
        .setMessage(
            "event=tx_selection decision=tentative_select tx_hash={} tx_rlp_size={} compressed_size_before={}")
        .addArgument(transaction::getHash)
        .addArgument(rlpEncodedTx.length)
        .addArgument(compressedSizeBefore)
        .log();

    // Store the RLP-encoded transaction for post-processing
    setWorkingState(new CompressionState(compressedSizeBefore, rlpEncodedTx.length));
    return SELECTED;
  }

  private static byte[] encodeTransaction(final Transaction transaction) {
    final BytesValueRLPOutput rlpOutput = new BytesValueRLPOutput();
    transaction.writeTo(rlpOutput);
    return rlpOutput.encoded().toArray();
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final byte[] rlpEncodedTx = encodeTransaction(transaction);

    // Actually append the transaction to the compressor now that EVM execution succeeded
    final TxCompressor.AppendResult result = txCompressor.appendTransaction(rlpEncodedTx);

    if (!result.getTxAppended()) {
      // This should not happen if canAppendTransaction returned true in pre-processing,
      // but handle it defensively
      log.atWarn()
          .setMessage(
              "event=tx_selection decision=reject_post_processing reason=append_failed "
                  + "tx_hash={} compressed_size_before={} compressed_size_after={}")
          .addArgument(transaction::getHash)
          .addArgument(result::getCompressedSizeBefore)
          .addArgument(result::getCompressedSizeAfter)
          .log();
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    log.atTrace()
        .setMessage(
            "event=tx_selection decision=select tx_hash={} tx_rlp_size={} "
                + "compressed_size_before={} compressed_size_after={}")
        .addArgument(transaction::getHash)
        .addArgument(rlpEncodedTx.length)
        .addArgument(result::getCompressedSizeBefore)
        .addArgument(result::getCompressedSizeAfter)
        .log();

    setWorkingState(new CompressionState(result.getCompressedSizeAfter(), rlpEncodedTx.length));
    return SELECTED;
  }

  /**
   * State tracking compression progress. The compressor itself maintains the actual compressed
   * data; this state is used for logging and debugging.
   */
  record CompressionState(int compressedSize, int lastTxRlpSize) {
    static CompressionState duplicate(final CompressionState state) {
      return new CompressionState(state.compressedSize(), state.lastTxRlpSize());
    }
  }
}
