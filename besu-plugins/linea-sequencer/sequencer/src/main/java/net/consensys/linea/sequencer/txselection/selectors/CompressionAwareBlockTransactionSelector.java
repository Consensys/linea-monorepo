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

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Random;

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

  /**
   * Fixed seed for deterministic placeholder generation. Using a fixed seed ensures consistent
   * compression behavior across runs, making the slow-path check reproducible and predictable.
   */
  private static final long PLACEHOLDER_SEED = 0xDEADBEEFL;

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
    if (blobSizeLimit <= compressedBlockHeaderOverhead) {
      throw new IllegalArgumentException(
          "blobSizeLimit ("
              + blobSizeLimit
              + ") must be greater than compressedBlockHeaderOverhead ("
              + compressedBlockHeaderOverhead
              + ")");
    }
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

    // Fast path: sum of per-tx compressed sizes is at or below the effective limit (blob limit
    // minus block header overhead). Since compressing all txs together always yields a smaller
    // result than the sum of individually compressed txs, the block is guaranteed to fit.
    final long fastPathLimit = blobSizeLimit - compressedBlockHeaderOverhead;
    if (newCumulative <= fastPathLimit) {
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
   * encoding. Placeholder header fields use deterministic pseudo-random bytes (seeded with a fixed
   * value) to ensure reproducible compression behavior while still providing varied data that
   * compresses similarly to real headers.
   *
   * @param pendingHeader the pending block header (provides number, timestamp, gasLimit)
   * @param transactions the transactions to include in the block
   * @return the RLP-encoded block
   */
  private static byte[] buildBlockRlp(
      final ProcessableBlockHeader pendingHeader, final List<Transaction> transactions) {
    // Use a new Random instance with fixed seed for each call to ensure deterministic output
    final Random random = new Random(PLACEHOLDER_SEED);

    final BlockHeader header =
        BlockHeaderBuilder.create()
            .parentHash(randomHash(random))
            .ommersHash(randomHash(random))
            .coinbase(Address.wrap(Bytes.wrap(randomBytes(random, 20))))
            .stateRoot(randomHash(random))
            .transactionsRoot(randomHash(random))
            .receiptsRoot(randomHash(random))
            .logsBloom(LogsBloomFilter.fromHexString(Bytes.wrap(randomBytes(random, 256)).toHexString()))
            .difficulty(Difficulty.of(random.nextLong(Long.MAX_VALUE)))
            .number(pendingHeader.getNumber())
            .gasLimit(pendingHeader.getGasLimit())
            .gasUsed(random.nextLong(Long.MAX_VALUE))
            .timestamp(pendingHeader.getTimestamp())
            .extraData(Bytes.wrap(randomBytes(random, 32)))
            .mixHash(randomHash(random))
            .nonce(random.nextLong())
            .baseFee(Wei.of(random.nextLong(Long.MAX_VALUE)))
            .blockHeaderFunctions(new MainnetBlockHeaderFunctions())
            .buildBlockHeader();

    final BlockBody body = new BlockBody(transactions, Collections.emptyList());
    final Block block = new Block(header, body);
    return block.toRlp().toArray();
  }

  private static Hash randomHash(final Random random) {
    return Hash.wrap(Bytes32.wrap(randomBytes(random, 32)));
  }

  private static byte[] randomBytes(final java.util.Random random, final int length) {
    final byte[] bytes = new byte[length];
    random.nextBytes(bytes);
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
