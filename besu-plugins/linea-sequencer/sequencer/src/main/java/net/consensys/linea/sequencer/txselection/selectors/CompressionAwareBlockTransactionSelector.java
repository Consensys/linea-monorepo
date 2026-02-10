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

import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import linea.blob.BlobCompressor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.utils.TransactionCompressor;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Block;
import org.hyperledger.besu.ethereum.core.BlockBody;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder;
import org.hyperledger.besu.ethereum.core.Difficulty;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions;
import org.hyperledger.besu.evm.log.LogsBloomFilter;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.AbstractStatefulPluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

/**
 * Stateful transaction selector that enforces a compressed block size limit (blob fit) using a
 * two-phase strategy:
 *
 * <p><b>Fast path:</b> accumulates {@code RawCompressedSize} per individual transaction. While the
 * cumulative sum is below {@code blobSizeLimit}, the block is guaranteed to fit because compressing
 * all transactions together always yields a smaller result than the sum of individually compressed
 * transactions.
 *
 * <p><b>Slow path:</b> once the cumulative sum reaches {@code blobSizeLimit}, builds a full block
 * RLP (using the pending block header data + all selected transactions + the candidate) and checks
 * with {@link BlobCompressor#canAppendBlock} whether it fits in the blob. This uses the native
 * compressor's actual block compression logic for maximum accuracy.
 *
 * <p>This approach maximises the number of transactions in a block: the fast path avoids expensive
 * full-block compression for the majority of transactions, while the slow path squeezes in extra
 * transactions that benefit from cross-transaction compression context.
 *
 * <p>All transaction types (regular, forced, bundle, liveness) flow through the same selector
 * pipeline and are tracked automatically.
 */
@Slf4j
public class CompressionAwareBlockTransactionSelector
    extends AbstractStatefulPluginTransactionSelector<
        CompressionAwareBlockTransactionSelector.CompressionState> {

  private static final SecureRandom RANDOM = new SecureRandom();

  private final int blobSizeLimit;
  private final int compressedBlockHeaderOverhead;
  private final TransactionCompressor transactionCompressor;
  private final BlobCompressor blobCompressor;

  public CompressionAwareBlockTransactionSelector(
      final SelectorsStateManager selectorsStateManager,
      final int blobSizeLimit,
      final int compressedBlockHeaderOverhead,
      final TransactionCompressor transactionCompressor,
      final BlobCompressor blobCompressor) {
    super(
        selectorsStateManager,
        new CompressionState(0L, new ArrayList<>()),
        CompressionState::duplicate);
    this.blobSizeLimit = blobSizeLimit;
    this.compressedBlockHeaderOverhead = compressedBlockHeaderOverhead;
    this.transactionCompressor = transactionCompressor;
    this.blobCompressor = blobCompressor;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final int txCompressedSize =
        transactionCompressor.getCompressedSize(transaction);

    final CompressionState state = getWorkingState();
    final long newCumulative = state.cumulativeRawCompressedSize() + txCompressedSize;

    // Fast path: sum of per-tx compressed sizes is below the effective limit (blob limit minus
    // block header overhead). Since compressing all txs together always yields a smaller result
    // than the sum of individually compressed txs, the block is guaranteed to fit.
    final long fastPathLimit = blobSizeLimit - compressedBlockHeaderOverhead;
    if (newCumulative < fastPathLimit) {
      log.atTrace()
          .setMessage(
              "Fast path: tx {} compressed={}, cumulative would be {} < effectiveLimit {} "
                  + "(blobLimit={} - headerOverhead={})")
          .addArgument(transaction::getHash)
          .addArgument(txCompressedSize)
          .addArgument(newCumulative)
          .addArgument(fastPathLimit)
          .addArgument(blobSizeLimit)
          .addArgument(compressedBlockHeaderOverhead)
          .log();
      return SELECTED;
    }

    // Slow path: conservative estimate exceeded the limit.
    // Build full block RLP and check with canAppendBlock for actual compression.
    final ProcessableBlockHeader pendingHeader = evaluationContext.getPendingBlockHeader();
    final List<Transaction> tentativeTxs = new ArrayList<>(state.selectedTransactions());
    tentativeTxs.add(transaction);
    final byte[] blockRlp = buildBlockRlp(pendingHeader, tentativeTxs);

    blobCompressor.reset();
    final boolean fits = blobCompressor.canAppendBlock(blockRlp);

    if (!fits) {
      log.atTrace()
          .setMessage(
              "Slow path REJECT: tx {} would not fit in blob "
                  + "(cumulative per-tx estimate was {}, limit {})")
          .addArgument(transaction::getHash)
          .addArgument(newCumulative)
          .addArgument(blobSizeLimit)
          .log();
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    log.atTrace()
        .setMessage(
            "Slow path ACCEPT: tx {} fits in blob via full-block compression "
                + "(cumulative per-tx estimate was {}, limit {})")
        .addArgument(transaction::getHash)
        .addArgument(newCumulative)
        .addArgument(blobSizeLimit)
        .log();
    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final int txCompressedSize =
        transactionCompressor.getCompressedSize(transaction);

    final CompressionState state = getWorkingState();
    final List<Transaction> newTxs = new ArrayList<>(state.selectedTransactions());
    newTxs.add(transaction);
    setWorkingState(
        new CompressionState(state.cumulativeRawCompressedSize() + txCompressedSize, newTxs));

    return SELECTED;
  }

  /**
   * Builds an RLP-encoded block from the pending block header and the list of selected
   * transactions, using Besu's standard {@link Block}/{@link BlockHeader}/{@link BlockBody}
   * encoding. Placeholder header fields use random bytes (worst case for compression, giving a
   * conservative estimate).
   *
   * @param pendingHeader the pending block header (provides number, timestamp, gasLimit)
   * @param transactions the transactions to include in the block
   * @return the RLP-encoded block
   */
  private static byte[] buildBlockRlp(
      final ProcessableBlockHeader pendingHeader, final List<Transaction> transactions) {
    final BlockHeader header =
        BlockHeaderBuilder.create()
            .parentHash(randomHash())
            .ommersHash(randomHash())
            .coinbase(Address.wrap(Bytes.wrap(randomBytes(20))))
            .stateRoot(randomHash())
            .transactionsRoot(randomHash())
            .receiptsRoot(randomHash())
            .logsBloom(LogsBloomFilter.fromHexString(Bytes.wrap(randomBytes(256)).toHexString()))
            .difficulty(Difficulty.of(RANDOM.nextLong(Long.MAX_VALUE)))
            .number(pendingHeader.getNumber())
            .gasLimit(pendingHeader.getGasLimit())
            .gasUsed(RANDOM.nextLong(Long.MAX_VALUE))
            .timestamp(pendingHeader.getTimestamp())
            .extraData(Bytes.wrap(randomBytes(32)))
            .mixHash(randomHash())
            .nonce(RANDOM.nextLong())
            .baseFee(Wei.of(RANDOM.nextLong(Long.MAX_VALUE)))
            .blockHeaderFunctions(new MainnetBlockHeaderFunctions())
            .buildBlockHeader();

    final BlockBody body = new BlockBody(transactions, Collections.emptyList());
    final Block block = new Block(header, body);
    return block.toRlp().toArray();
  }

  private static Hash randomHash() {
    return Hash.wrap(Bytes32.wrap(randomBytes(32)));
  }

  private static byte[] randomBytes(final int length) {
    final byte[] bytes = new byte[length];
    RANDOM.nextBytes(bytes);
    return bytes;
  }

  /**
   * State tracking cumulative per-tx compressed size and the list of selected transactions. The
   * list is needed for the slow-path full-block compression check when the cumulative estimate
   * exceeds the blob size limit.
   */
  record CompressionState(long cumulativeRawCompressedSize, List<Transaction> selectedTransactions) {
    static CompressionState duplicate(final CompressionState state) {
      return new CompressionState(
          state.cumulativeRawCompressedSize(), new ArrayList<>(state.selectedTransactions()));
    }
  }
}
