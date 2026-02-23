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
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
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
 * with {@link BlobCompressor#reset} + {@link BlobCompressor#appendBlock} whether it fits in the
 * blob. This uses the native compressor's actual block compression logic for maximum accuracy.
 *
 * <p>The slow-path compression ({@code reset} + {@code appendBlock}) runs on a dedicated
 * single-threaded background executor so that it overlaps with EVM execution. Pre-processing
 * returns {@code SELECTED} optimistically; post-processing collects the result. If the block is
 * full, post-processing returns {@code BLOCK_COMPRESSED_SIZE_OVERFLOW} and a sentinel value is
 * written to the state so that all subsequent pre-processings short-circuit immediately without
 * spawning further work.
 *
 * <p>This approach maximises the number of transactions in a block: the fast path avoids expensive
 * full-block compression for the majority of transactions, while the slow path squeezes in extra
 * transactions that benefit from cross-transaction compression context.
 *
 * <p>All transaction types (regular, forced, bundle, liveness) flow through the same selector
 * pipeline and are tracked automatically.
 */
@Slf4j
public class CompressionAwareTransactionSelector
    extends AbstractStatefulPluginTransactionSelector<
        CompressionAwareTransactionSelector.CompressionState> {

  /**
   * Fixed seed for deterministic placeholder generation. Using a fixed seed ensures consistent
   * compression behavior across runs, making the slow-path check reproducible and predictable.
   */
  private static final long PLACEHOLDER_SEED = 0xDEADBEEFL;

  /**
   * Sentinel value for {@code cumulativeCompressedSize} that signals the block is definitively
   * full. Pre-processing immediately returns {@code BLOCK_COMPRESSED_SIZE_OVERFLOW} when this value
   * is detected, preventing further async submissions and wasted EVM executions.
   */
  private static final long BLOCK_FULL_SENTINEL = Long.MAX_VALUE;

  private final long fastPathLimit;
  private final TransactionCompressor transactionCompressor;
  private final BlobCompressor blobCompressor;

  /**
   * Shared single-threaded executor for slow-path compression. Declared static so that one daemon
   * thread is created once for the JVM lifetime regardless of how many blocks are built. Using a
   * single thread guarantees that {@code blobCompressor.reset()} and {@code
   * blobCompressor.appendBlock()} are never called concurrently, even when a cancelled task's
   * native call has not yet returned. Block building in Besu is sequential, so a single background
   * thread is sufficient.
   *
   * <p><b>Thread-safety note:</b> after pre-processing submits the async task, downstream selectors
   * in the same pipeline (e.g. {@code ProfitableTransactionSelector}) may call {@code
   * blobCompressor.compressedSize()} concurrently via {@code CachingTransactionCompressor}. This is
   * safe because {@code compressedSize()} is a stateless read that compresses an input byte array
   * without touching the block-accumulation state managed by {@code reset()}/{@code appendBlock()}.
   * These two operations work on independent parts of the native compressor's state.
   */
  private static final ExecutorService COMPRESSION_EXECUTOR =
      Executors.newSingleThreadExecutor(
          r -> {
            final Thread t = new Thread(r, "compression-slow-path");
            t.setDaemon(true);
            return t;
          });

  /**
   * Pending result of an async slow-path compression check. Non-null only between a slow-path
   * pre-processing call and its corresponding post-processing call (or {@code
   * onTransactionNotSelected}). Accessed only from the block-building thread, but declared {@code
   * volatile} for safe publication to the background thread.
   */
  private volatile Future<Boolean> pendingSlowPathFuture;

  public CompressionAwareTransactionSelector(
      final SelectorsStateManager selectorsStateManager,
      final int blobSizeLimit,
      final int compressedBlockHeaderOverhead,
      final TransactionCompressor transactionCompressor,
      final BlobCompressor blobCompressor) {
    super(
        selectorsStateManager,
        new CompressionState(0L, new ArrayList<>()),
        CompressionState::duplicate);
    this.fastPathLimit = blobSizeLimit - compressedBlockHeaderOverhead;
    if (fastPathLimit <= 0) {
      throw new IllegalArgumentException("fastPathLimit must be positive, got " + fastPathLimit);
    }
    this.transactionCompressor = transactionCompressor;
    this.blobCompressor = blobCompressor;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();

    final CompressionState state = getWorkingState();

    // Block-full sentinel: a previous slow-path check confirmed overflow. Stop immediately
    // without submitting further work or running EVM execution.
    if (state.cumulativeCompressedSize() == BLOCK_FULL_SENTINEL) {
      return BLOCK_COMPRESSED_SIZE_OVERFLOW;
    }

    final int txCompressedSize = transactionCompressor.getCompressedSize(transaction);
    long newConservativeCumulative = state.cumulativeCompressedSize() + txCompressedSize;

    // Fast path: sum of per-tx compressed sizes is at or below the effective limit (blob limit
    // minus block header overhead). Since compressing all txs together always yields a smaller
    // result than the sum of individually compressed txs, the block is guaranteed to fit.
    if (newConservativeCumulative <= fastPathLimit) {
      log.atTrace()
          .setMessage(
              "event=tx_selection path=fast decision=select tx_hash={} tx_compressed_size={} cumulative_conservative={} fast_path_limit={}")
          .addArgument(transaction::getHash)
          .addArgument(txCompressedSize)
          .addArgument(newConservativeCumulative)
          .addArgument(fastPathLimit)
          .log();
      setWorkingState(new CompressionState(newConservativeCumulative, state.selectedTransactions));
      return SELECTED;
    }

    // Slow path: conservative estimate exceeded the limit.
    // Build the block RLP on this thread (cheap, CPU-only), then submit the native
    // reset+appendBlock to the background executor so it can overlap with EVM execution.
    final ProcessableBlockHeader pendingHeader = evaluationContext.getPendingBlockHeader();
    final List<Transaction> tentativeTxs = new ArrayList<>(state.selectedTransactions());
    tentativeTxs.add(transaction);
    final byte[] blockRlp = buildBlockRlp(pendingHeader, tentativeTxs);

    log.atTrace()
        .setMessage(
            "event=tx_selection path=slow decision=pending tx_hash={} fast_path_limit={} tentative_tx_count={} tentative_block_rlp_size={}")
        .addArgument(transaction::getHash)
        .addArgument(fastPathLimit)
        .addArgument(tentativeTxs::size)
        .addArgument(blockRlp.length)
        .log();

    pendingSlowPathFuture =
        COMPRESSION_EXECUTOR.submit(
            () -> {
              blobCompressor.reset();
              return blobCompressor.appendBlock(blockRlp).getBlockAppended();
            });

    setWorkingState(new CompressionState(fastPathLimit, state.selectedTransactions));
    return SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final Transaction transaction =
        (Transaction) evaluationContext.getPendingTransaction().getTransaction();
    final CompressionState state = getWorkingState();

    final Future<Boolean> future = pendingSlowPathFuture;
    if (future != null) {
      pendingSlowPathFuture = null;

      final boolean blockAppended;
      try {
        final long waitStartNs = System.nanoTime();
        blockAppended = future.get();
        final long waitNs = System.nanoTime() - waitStartNs;
        if (waitNs > 0) {
          log.atDebug()
              .setMessage("event=tx_selection path=slow compression_wait_us={} tx_hash={}")
              .addArgument(waitNs / 1_000)
              .addArgument(transaction::getHash)
              .log();
        }
      } catch (final InterruptedException e) {
        Thread.currentThread().interrupt();
        log.atWarn()
            .setMessage(
                "event=tx_selection path=slow interrupted waiting for compression tx_hash={}")
            .addArgument(transaction::getHash)
            .log();
        return BLOCK_COMPRESSED_SIZE_OVERFLOW;
      } catch (final ExecutionException e) {
        log.atWarn()
            .setMessage("event=tx_selection path=slow compression failed tx_hash={}")
            .addArgument(transaction::getHash)
            .setCause(e.getCause())
            .log();
        return BLOCK_COMPRESSED_SIZE_OVERFLOW;
      }

      if (!blockAppended) {
        log.atTrace()
            .setMessage(
                "event=tx_selection path=slow decision=reject reason=block_compressed_size_overflow tx_hash={} fast_path_limit={}")
            .addArgument(transaction::getHash)
            .addArgument(fastPathLimit)
            .log();
        setWorkingState(new CompressionState(BLOCK_FULL_SENTINEL, state.selectedTransactions()));
        return BLOCK_COMPRESSED_SIZE_OVERFLOW;
      }

      log.atTrace()
          .setMessage("event=tx_selection path=slow decision=select tx_hash={} fast_path_limit={}")
          .addArgument(transaction::getHash)
          .addArgument(fastPathLimit)
          .log();
    }

    final List<Transaction> newTxs = new ArrayList<>(state.selectedTransactions());
    newTxs.add(transaction);
    setWorkingState(new CompressionState(state.cumulativeCompressedSize(), newTxs));
    return SELECTED;
  }

  @Override
  public void onTransactionNotSelected(
      final TransactionEvaluationContext evaluationContext,
      final TransactionSelectionResult result) {
    final Future<Boolean> future = pendingSlowPathFuture;
    if (future != null) {
      pendingSlowPathFuture = null;
      // cancel(false): do not interrupt a running native JNI call. The single-threaded executor
      // will naturally serialise any subsequent task behind any still-running compression.
      future.cancel(false);
    }
    super.onTransactionNotSelected(evaluationContext, result);
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
            .logsBloom(
                LogsBloomFilter.fromHexString(Bytes.wrap(randomBytes(random, 256)).toHexString()))
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
   * State tracking the cumulative compressed size and the list of selected transactions. The list
   * is needed for the slow-path full-block compression check when the cumulative estimate exceeds
   * the blob size limit.
   *
   * <p>{@code cumulativeCompressedSize} holds the conservative sum of individually-compressed tx
   * sizes while the fast path is active. Once the slow path is triggered for the first time it is
   * set to {@code fastPathLimit} as a sentinel, ensuring all subsequent transactions go through the
   * slow path. When a slow-path check confirms overflow it is set to {@link Long#MAX_VALUE} (the
   * "block full" sentinel), causing all subsequent pre-processings to short-circuit immediately.
   */
  record CompressionState(long cumulativeCompressedSize, List<Transaction> selectedTransactions) {
    static CompressionState duplicate(final CompressionState state) {
      return new CompressionState(
          state.cumulativeCompressedSize(), new ArrayList<>(state.selectedTransactions()));
    }
  }
}
