/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.forced;

import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.BAD_BALANCE;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.BAD_NONCE;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.BAD_PRECOMPILE;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.FILTERED_ADDRESSES;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.INCLUDED;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.OTHER;
import static net.consensys.linea.sequencer.forced.ForcedTransactionInclusionResult.TOO_MANY_LOGS;
import static net.consensys.linea.sequencer.txselection.LineaTransactionSelectionResult.DENIED_LOG_TOPIC;

import com.github.benmanes.caffeine.cache.Cache;
import com.github.benmanes.caffeine.cache.Caffeine;
import java.util.ArrayList;
import java.util.Deque;
import java.util.EnumMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicLong;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.metrics.LineaMetricCategory;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.MetricsSystem;
import org.hyperledger.besu.plugin.services.metrics.LabelledSuppliedMetric;
import org.hyperledger.besu.plugin.services.txselection.BlockTransactionSelectionService;

/**
 * Implementation of the forced transaction pool. Maintains a queue of pending forced transactions
 * and a cache of inclusion statuses.
 */
@Slf4j
public class LineaForcedTransactionPool implements ForcedTransactionPoolService {

  public static final int DEFAULT_STATUS_CACHE_SIZE = 10_000;

  private final Deque<ForcedTransaction> pendingQueue = new ConcurrentLinkedDeque<>();
  private final Cache<Hash, ForcedTransactionStatus> statusCache;
  private final Map<ForcedTransactionInclusionResult, AtomicLong> inclusionResultCounters =
      new EnumMap<>(ForcedTransactionInclusionResult.class);

  /**
   * Tracks the block number where a forced transaction failure occurred. Used to prevent processing
   * additional forced transactions in the same block, since only one invalidity proof can be
   * generated per block. Reset to -1 when moving to a new block. Required because there might be
   * multiple processForBlock calls for the same slot and block number
   */
  private volatile long failureBlockNumber = -1;

  public LineaForcedTransactionPool() {
    this(DEFAULT_STATUS_CACHE_SIZE, null);
  }

  public LineaForcedTransactionPool(final int statusCacheSize, final MetricsSystem metricsSystem) {
    this.statusCache = Caffeine.newBuilder().maximumSize(statusCacheSize).build();

    for (ForcedTransactionInclusionResult result : ForcedTransactionInclusionResult.values()) {
      inclusionResultCounters.put(result, new AtomicLong(0));
    }

    if (metricsSystem != null) {
      initMetrics(metricsSystem);
    }
  }

  private void initMetrics(final MetricsSystem metricsSystem) {
    metricsSystem.createGauge(
        LineaMetricCategory.SEQUENCER_FORCED_TX,
        "pool_size",
        "Number of pending forced transactions in the pool",
        this::pendingCount);

    metricsSystem.createGauge(
        LineaMetricCategory.SEQUENCER_FORCED_TX,
        "status_cache_size",
        "Number of entries in the forced transaction status cache",
        () -> (int) statusCache.estimatedSize());

    final LabelledSuppliedMetric inclusionResultGauge =
        metricsSystem.createLabelledSuppliedGauge(
            LineaMetricCategory.SEQUENCER_FORCED_TX,
            "inclusion_result",
            "Total count of forced transactions by inclusion result",
            "result");

    for (ForcedTransactionInclusionResult result : ForcedTransactionInclusionResult.values()) {
      inclusionResultGauge.labels(
          () -> inclusionResultCounters.get(result).doubleValue(), result.name());
    }
  }

  @Override
  public List<Hash> addForcedTransactions(final List<ForcedTransaction> transactions) {
    final List<Hash> hashes = new ArrayList<>(transactions.size());
    for (final ForcedTransaction tx : transactions) {
      pendingQueue.addLast(tx);
      hashes.add(tx.txHash());
      log.atDebug()
          .setMessage("action=add_forced_tx txHash={} deadline={}")
          .addArgument(tx.txHash()::toHexString)
          .addArgument(tx::deadline)
          .log();
    }
    log.atInfo()
        .setMessage("action=add_forced_txs count={} pendingQueueSize={}")
        .addArgument(transactions::size)
        .addArgument(pendingQueue::size)
        .log();
    return hashes;
  }

  @Override
  public void processForBlock(
      final long blockNumber,
      final long blockTimestamp,
      final BlockTransactionSelectionService blockTransactionSelectionService) {

    if (pendingQueue.isEmpty()) {
      return;
    }

    // If we had a failure in a previous block, reset the flag for this new block
    if (failureBlockNumber != -1 && failureBlockNumber < blockNumber) {
      failureBlockNumber = -1;
    }

    // Skip if we already had a failure in this block (only one invalidity proof per block)
    if (failureBlockNumber == blockNumber) {
      log.atDebug()
          .setMessage("action=skip_forced_txs_already_failed blockNumber={} failureBlockNumber={}")
          .addArgument(blockNumber)
          .addArgument(failureBlockNumber)
          .log();
      return;
    }

    log.atDebug()
        .setMessage("action=process_forced_txs_start blockNumber={} pendingCount={}")
        .addArgument(blockNumber)
        .addArgument(pendingQueue::size)
        .log();

    int index = 0;
    final int initialSize = pendingQueue.size();

    for (int i = 0; i < initialSize; i++) {
      final ForcedTransaction ftx = pendingQueue.peekFirst();
      if (ftx == null) {
        break;
      }

      log.atTrace()
          .setMessage("action=evaluate_forced_tx txHash={} index={} blockNumber={}")
          .addArgument(ftx.txHash()::toHexString)
          .addArgument(index)
          .addArgument(blockNumber)
          .log();

      final PendingTransaction pendingTx =
          new org.hyperledger.besu.ethereum.eth.transactions.PendingTransaction.Local(
              ftx.transaction());

      final TransactionSelectionResult result =
          blockTransactionSelectionService.evaluatePendingTransaction(pendingTx);

      if (result.selected()) {
        pendingQueue.pollFirst();
        recordStatus(ftx, blockNumber, blockTimestamp, INCLUDED);
        log.atInfo()
            .setMessage("action=forced_tx_included txHash={} blockNumber={} index={}")
            .addArgument(ftx.txHash()::toHexString)
            .addArgument(blockNumber)
            .addArgument(index)
            .log();
        index++;
      } else {
        final ForcedTransactionInclusionResult inclusionResult = mapToInclusionResult(result);

        // Mark that we had a failure in this block - prevents processing more FTXs in same block
        failureBlockNumber = blockNumber;

        if (index == 0 && !inclusionResult.shouldRetry()) {
          // Final rejection at index 0 with known reason - remove and record status
          pendingQueue.pollFirst();
          recordStatus(ftx, blockNumber, blockTimestamp, inclusionResult);
          log.atInfo()
              .setMessage(
                  "action=forced_tx_rejected txHash={} blockNumber={} index=0 selectionResult={} inclusionResult={}")
              .addArgument(ftx.txHash()::toHexString)
              .addArgument(blockNumber)
              .addArgument(result)
              .addArgument(inclusionResult)
              .log();
        } else {
          // Either index > 0 or unknown reason (OTHER) - retry in next block
          log.atInfo()
              .setMessage(
                  "action=forced_tx_retry_scheduled txHash={} blockNumber={} index={} selectionResult={} inclusionResult={}")
              .addArgument(ftx.txHash()::toHexString)
              .addArgument(blockNumber)
              .addArgument(index)
              .addArgument(result)
              .addArgument(inclusionResult)
              .log();
        }
        break;
      }
    }

    log.atDebug()
        .setMessage("action=process_forced_txs_end blockNumber={} remainingPending={}")
        .addArgument(blockNumber)
        .addArgument(pendingQueue::size)
        .log();
  }

  @Override
  public Optional<ForcedTransactionStatus> getInclusionStatus(final Hash txHash) {
    return Optional.ofNullable(statusCache.getIfPresent(txHash));
  }

  @Override
  public int pendingCount() {
    return pendingQueue.size();
  }

  private void recordStatus(
      final ForcedTransaction ftx,
      final long blockNumber,
      final long blockTimestamp,
      final ForcedTransactionInclusionResult result) {
    final ForcedTransactionStatus status =
        new ForcedTransactionStatus(
            ftx.txHash(), ftx.transaction().getSender(), blockNumber, blockTimestamp, result);
    statusCache.put(ftx.txHash(), status);
    inclusionResultCounters.get(result).incrementAndGet();
  }

  /**
   * Maps a Besu TransactionSelectionResult to a ForcedTransactionInclusionResult.
   *
   * <p>First checks against known Linea-specific result constants using equality, then falls back
   * to string matching for standard Besu results. Returns OTHER for unrecognized results, which
   * triggers a retry in the next block.
   */
  private ForcedTransactionInclusionResult mapToInclusionResult(
      final TransactionSelectionResult result) {

    // Check against known Linea-specific result constants using equality
    if (result.equals(DENIED_LOG_TOPIC)) {
      return FILTERED_ADDRESSES;
    }
    final String resultString = result.toString().toUpperCase();

    if (resultString.contains("NONCE")) {
      return BAD_NONCE;
    }
    if (resultString.contains("BALANCE") || resultString.contains("UPFRONT_COST")) {
      return BAD_BALANCE;
    }
    if (resultString.contains("PRECOMPILE")) {
      return BAD_PRECOMPILE;
    }
    if (resultString.contains("TOO_MANY_LOGS")) {
      return TOO_MANY_LOGS;
    }
    if (resultString.contains("DENIED")
        || resultString.contains("FILTERED")
        || resultString.contains("BLOCKED")) {
      return FILTERED_ADDRESSES;
    }

    log.atWarn()
        .setMessage("action=unknown_selection_result result={} willRetry=true")
        .addArgument(resultString)
        .log();
    return OTHER;
  }
}
