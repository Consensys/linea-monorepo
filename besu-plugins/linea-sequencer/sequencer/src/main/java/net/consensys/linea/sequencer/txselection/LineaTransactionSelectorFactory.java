/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.sequencer.txselection;

import java.time.Instant;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Supplier;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bundles.BundlePoolService;
import net.consensys.linea.bundles.TransactionBundle;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.jsonrpc.JsonRpcManager;
import net.consensys.linea.metrics.HistogramMetrics;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.liveness.LivenessService;
import net.consensys.linea.sequencer.txselection.selectors.LineaTransactionSelector;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.txselection.BlockTransactionSelectionService;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelectorFactory;
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager;

/**
 * Represents a factory for creating transaction selectors. Note that a new instance of the
 * transaction selector is created everytime a new block creation time is started.
 *
 * <p>Also provides an entrypoint for bundle transactions
 */
@Slf4j
public class LineaTransactionSelectorFactory implements PluginTransactionSelectorFactory {
  private final BlockchainService blockchainService;
  private final Optional<JsonRpcManager> rejectedTxJsonRpcManager;
  private final LineaTransactionSelectorConfiguration txSelectorConfiguration;
  private final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration;
  private final LineaProfitabilityConfiguration profitabilityConfiguration;
  private final LineaTracerConfiguration tracerConfiguration;
  private final Optional<HistogramMetrics> maybeProfitabilityMetrics;
  private final BundlePoolService bundlePoolService;
  private final Optional<LivenessService> livenessService;
  private final InvalidTransactionByLineCountCache invalidTransactionByLineCountCache;
  private final AtomicReference<LineaTransactionSelector> currSelector = new AtomicReference<>();

  public LineaTransactionSelectorFactory(
      final BlockchainService blockchainService,
      final LineaTransactionSelectorConfiguration txSelectorConfiguration,
      final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration,
      final LineaProfitabilityConfiguration profitabilityConfiguration,
      final LineaTracerConfiguration tracerConfiguration,
      final Optional<LivenessService> livenessService,
      final Optional<JsonRpcManager> rejectedTxJsonRpcManager,
      final Optional<HistogramMetrics> maybeProfitabilityMetrics,
      final BundlePoolService bundlePoolService,
      final InvalidTransactionByLineCountCache invalidTransactionByLineCountCache) {
    this.blockchainService = blockchainService;
    this.txSelectorConfiguration = txSelectorConfiguration;
    this.l1L2BridgeConfiguration = l1L2BridgeConfiguration;
    this.profitabilityConfiguration = profitabilityConfiguration;
    this.tracerConfiguration = tracerConfiguration;
    this.rejectedTxJsonRpcManager = rejectedTxJsonRpcManager;
    this.maybeProfitabilityMetrics = maybeProfitabilityMetrics;
    this.bundlePoolService = bundlePoolService;
    this.livenessService = livenessService;
    this.invalidTransactionByLineCountCache = invalidTransactionByLineCountCache;
  }

  @Override
  public PluginTransactionSelector create(final SelectorsStateManager selectorsStateManager) {
    final var selector =
        new LineaTransactionSelector(
            selectorsStateManager,
            blockchainService,
            txSelectorConfiguration,
            l1L2BridgeConfiguration,
            profitabilityConfiguration,
            tracerConfiguration,
            rejectedTxJsonRpcManager,
            maybeProfitabilityMetrics,
            invalidTransactionByLineCountCache);
    currSelector.set(selector);
    return selector;
  }

  public void selectPendingTransactions(
      final BlockTransactionSelectionService bts, final ProcessableBlockHeader pendingBlockHeader) {

    // do not use directly this atomic boolean but always use the lambda below,
    // to check for interrupt, because they check if the thread is actually interrupted.
    final AtomicBoolean interruptRecorded = new AtomicBoolean(false);

    final Supplier<Boolean> isSelectionInterrupted =
        () -> {
          if (interruptRecorded.get()) {
            return true;
          }
          if (Thread.currentThread().isInterrupted()) {
            log.info("Bundle selection interrupted");
            interruptRecorded.set(true);
            return true;
          }
          return false;
        };

    // check and send liveness bundle if any
    checkAndSendLivenessBundle(bts, pendingBlockHeader.getNumber());

    if (isSelectionInterrupted.get()) return;

    final var bundlesByBlockNumber =
        bundlePoolService.getBundlesByBlockNumber(pendingBlockHeader.getNumber());

    log.atDebug()
        .setMessage("Bundle pool stats: total={}, for block #{}={}")
        .addArgument(bundlePoolService::size)
        .addArgument(pendingBlockHeader::getNumber)
        .addArgument(bundlesByBlockNumber::size)
        .log();

    if (isSelectionInterrupted.get()) return;

    final var selectionStartedAt = System.nanoTime();

    bundlesByBlockNumber.stream()
        .takeWhile(unused -> !isSelectionInterrupted.get())
        .forEach(
            bundle -> {
              final var bundleStartedAt = System.nanoTime();
              log.trace("Starting evaluation of bundle {}", bundle.bundleIdentifier());

              var maybeBadBundleRes =
                  bundle.pendingTransactions().stream()
                      .takeWhile(unused -> !isSelectionInterrupted.get())
                      .map(bts::evaluatePendingTransaction)
                      .filter(evalRes -> !evalRes.selected())
                      .findFirst();

              final var now = System.nanoTime();
              final var cumulativeBundleSelectionTime = now - selectionStartedAt;
              final var currentBundleSelectionTime = now - bundleStartedAt;

              if (isSelectionInterrupted.get()) {
                log.atDebug()
                    .setMessage(
                        "Bundle selection interrupted while processing bundle {},"
                            + " elapsed time: current bundle {}ms, cumulative {}ms")
                    .addArgument(bundle::bundleIdentifier)
                    .addArgument(() -> nanosToMillis(currentBundleSelectionTime))
                    .addArgument(() -> nanosToMillis(cumulativeBundleSelectionTime))
                    .log();
                rollback(bts);
              } else {
                if (maybeBadBundleRes.isPresent()) {
                  log.atDebug()
                      .setMessage(
                          "Failed bundle {}, reason {}, elapsed time: current bundle {}ms, cumulative {}ms")
                      .addArgument(bundle::bundleIdentifier)
                      .addArgument(maybeBadBundleRes::get)
                      .addArgument(() -> nanosToMillis(currentBundleSelectionTime))
                      .addArgument(() -> nanosToMillis(cumulativeBundleSelectionTime))
                      .log();
                  rollback(bts);
                } else {
                  log.atDebug()
                      .setMessage(
                          "Selected bundle {}, elapsed time: current bundle {}ms, cumulative {}ms")
                      .addArgument(bundle::bundleIdentifier)
                      .addArgument(() -> nanosToMillis(currentBundleSelectionTime))
                      .addArgument(() -> nanosToMillis(cumulativeBundleSelectionTime))
                      .log();
                  commit(bts);
                }
              }
            });
    currSelector.set(null);
  }

  private void checkAndSendLivenessBundle(
      BlockTransactionSelectionService bts, long pendingBlockNumber) {
    final long headBlockTimestamp = blockchainService.getChainHeadHeader().getTimestamp();

    Optional<TransactionBundle> livenessBundle =
        livenessService.isPresent()
            ? livenessService
                .get()
                .checkBlockTimestampAndBuildBundle(
                    Instant.now().getEpochSecond(), headBlockTimestamp, pendingBlockNumber)
            : Optional.empty();

    if (livenessBundle.isPresent()) {
      log.trace("Starting evaluation of liveness bundle {}", livenessBundle.get());
      var badBundleRes =
          livenessBundle.get().pendingTransactions().stream()
              .map(bts::evaluatePendingTransaction)
              .filter(evalRes -> !evalRes.selected())
              .findFirst();

      if (badBundleRes.isPresent()) {
        log.debug("Failed liveness bundle {}, reason {}", livenessBundle.get(), badBundleRes);
        livenessService.get().updateUptimeMetrics(false, headBlockTimestamp);
        rollback(bts);
      } else {
        log.debug("Selected liveness bundle {}", livenessBundle.get());
        livenessService.get().updateUptimeMetrics(true, headBlockTimestamp);
        commit(bts);
      }
    }
  }

  private void commit(final BlockTransactionSelectionService bts) {
    currSelector.get().getOperationTracer().commitTransactionBundle();
    bts.commit();
  }

  private void rollback(final BlockTransactionSelectionService bts) {
    currSelector.get().getOperationTracer().popTransactionBundle();
    bts.rollback();
  }

  private long nanosToMillis(final long nanos) {
    return TimeUnit.NANOSECONDS.toMillis(nanos);
  }
}
