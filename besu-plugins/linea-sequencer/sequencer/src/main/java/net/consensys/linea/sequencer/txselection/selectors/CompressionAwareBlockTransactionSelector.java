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
import net.consensys.linea.utils.TxEncodingUtils;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

/**
 * Stateful transaction selector that enforces a compressed block size limit (blob fit) using a
 * two-phase strategy optimized for performance:
 *
 * <p><b>Fast path:</b> Uses stateless per-transaction compressed size estimates. While the
 * cumulative sum of individually compressed transaction sizes is below the limit, the block is
 * guaranteed to fit because compressing all transactions together always yields a smaller result
 * than the sum of individually compressed transactions. This path avoids expensive JNI calls to the
 * Go compressor for the majority of transactions.
 *
 * <p><b>Slow path:</b> Once the cumulative estimate reaches the limit, uses the additive {@link
 * TxCompressor} to check if the transaction actually fits. The compressor maintains compression
 * context across transactions, so this check is accurate and can accept transactions that the fast
 * path would conservatively reject.
 *
 * <p>This approach maximizes both performance and block utilization: the fast path avoids expensive
 * compression calls for most transactions, while the slow path squeezes in extra transactions that
 * benefit from cross-transaction compression context.
 *
 * <p>All transaction types (regular, forced, bundle, liveness) flow through the same selector
 * pipeline and are tracked automatically.
 */
@Slf4j
public class CompressionAwareBlockTransactionSelector implements PluginTransactionSelector {

  private final TxCompressor txCompressor;
  private final int blobSizeLimit;

  private long cumulativeEstimatedSize = 0;
  private boolean inSlowPath = false;

  /**
   * Creates a new compression-aware block transaction selector.
   *
   * @param txCompressor the transaction compressor to use for slow path checks
   * @param blobSizeLimit the maximum compressed size limit in bytes (should match the limit
   *     configured in the TxCompressor)
   */
  public CompressionAwareBlockTransactionSelector(
      final TxCompressor txCompressor, final int blobSizeLimit) {
    this.txCompressor = txCompressor;
    this.blobSizeLimit = blobSizeLimit;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final byte[] from = TxEncodingUtils.getSenderBytes(transaction);
    final byte[] rlpForSigning = TxEncodingUtils.encodeForSigning(transaction);
    final byte[] txData = TxEncodingUtils.encodeForCompressor(transaction);

    if (!inSlowPath) {
      // Fast path: use stateless per-tx compressed size estimate
      final int txEstimatedSize = txCompressor.compressedSize(txData);
      final long newCumulative = cumulativeEstimatedSize + txEstimatedSize;

      if (newCumulative <= blobSizeLimit) {
        log.atTrace()
            .setMessage(
                "event=tx_selection path=fast decision=tentative_select "
                    + "tx_hash={} tx_rlp_size={} tx_estimated_compressed={} cumulative_estimate={}")
            .addArgument(transaction::getHash)
            .addArgument(rlpForSigning.length)
            .addArgument(txEstimatedSize)
            .addArgument(newCumulative)
            .log();
        return SELECTED;
      }

      // Fast path limit exceeded, switch to slow path
      log.atDebug()
          .setMessage(
              "event=tx_selection path=switch_to_slow "
                  + "tx_hash={} cumulative_estimate={} limit={}")
          .addArgument(transaction::getHash)
          .addArgument(newCumulative)
          .addArgument(blobSizeLimit)
          .log();
      inSlowPath = true;
    }

    // Slow path: use additive compressor for accurate check
    final int compressedSizeBefore = txCompressor.getCompressedSize();
    final boolean canAppend = txCompressor.canAppendTransaction(from, rlpForSigning);

    if (!canAppend) {
      log.atTrace()
          .setMessage(
              "event=tx_selection path=slow decision=reject reason=block_compressed_size_overflow "
                  + "tx_hash={} tx_rlp_size={} compressed_size_before={}")
          .addArgument(transaction::getHash)
          .addArgument(rlpForSigning.length)
          .addArgument(compressedSizeBefore)
          .log();
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    log.atTrace()
        .setMessage(
            "event=tx_selection path=slow decision=tentative_select "
                + "tx_hash={} tx_rlp_size={} compressed_size_before={}")
        .addArgument(transaction::getHash)
        .addArgument(rlpForSigning.length)
        .addArgument(compressedSizeBefore)
        .log();

    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final byte[] from = TxEncodingUtils.getSenderBytes(transaction);
    final byte[] rlpForSigning = TxEncodingUtils.encodeForSigning(transaction);
    final byte[] txData = TxEncodingUtils.encodeForCompressor(transaction);

    if (!inSlowPath) {
      // Fast path: update cumulative estimate and append to compressor
      final int txEstimatedSize = txCompressor.compressedSize(txData);
      cumulativeEstimatedSize += txEstimatedSize;

      // Also append to compressor to maintain state for when we switch to slow path
      final TxCompressor.AppendResult result = txCompressor.appendTransaction(from, rlpForSigning);

      log.atTrace()
          .setMessage(
              "event=tx_selection path=fast decision=select "
                  + "tx_hash={} tx_rlp_size={} tx_estimated_compressed={} cumulative_estimate={} "
                  + "actual_compressed_before={} actual_compressed_after={}")
          .addArgument(transaction::getHash)
          .addArgument(rlpForSigning.length)
          .addArgument(txEstimatedSize)
          .addArgument(cumulativeEstimatedSize)
          .addArgument(result::getCompressedSizeBefore)
          .addArgument(result::getCompressedSizeAfter)
          .log();

      return SELECTED;
    }

    // Slow path: append to compressor
    final TxCompressor.AppendResult result = txCompressor.appendTransaction(from, rlpForSigning);

    if (!result.getTxAppended()) {
      // This should not happen if canAppendTransaction returned true in pre-processing
      log.atWarn()
          .setMessage(
              "event=tx_selection path=slow decision=reject_post_processing reason=append_failed "
                  + "tx_hash={} compressed_size_before={} compressed_size_after={}")
          .addArgument(transaction::getHash)
          .addArgument(result::getCompressedSizeBefore)
          .addArgument(result::getCompressedSizeAfter)
          .log();
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    log.atTrace()
        .setMessage(
            "event=tx_selection path=slow decision=select "
                + "tx_hash={} tx_rlp_size={} compressed_size_before={} compressed_size_after={}")
        .addArgument(transaction::getHash)
        .addArgument(rlpForSigning.length)
        .addArgument(result::getCompressedSizeBefore)
        .addArgument(result::getCompressedSizeAfter)
        .log();

    return SELECTED;
  }
}
