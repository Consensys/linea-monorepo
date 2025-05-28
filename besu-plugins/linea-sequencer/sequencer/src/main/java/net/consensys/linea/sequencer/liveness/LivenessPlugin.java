/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.liveness;

import com.google.auto.service.AutoService;
import java.io.IOException;
import java.math.BigInteger;
import java.time.Instant;
import java.util.Arrays;
import java.util.Collections;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.atomic.AtomicReference;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.config.LivenessPluginConfiguration;
import net.consensys.linea.metrics.LineaMetricCategory;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.metrics.Counter;
import org.hyperledger.besu.plugin.services.metrics.MetricCategory;
import org.web3j.abi.FunctionEncoder;
import org.web3j.abi.datatypes.Bool;
import org.web3j.abi.datatypes.Function;
import org.web3j.abi.datatypes.generated.Uint256;
import org.web3j.crypto.Credentials;
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.DefaultBlockParameterName;
import org.web3j.protocol.core.methods.response.EthGetTransactionCount;
import org.web3j.protocol.http.HttpService;
import org.web3j.utils.Numeric;

/**
 * The LivenessPlugin is responsible for monitoring the blockchain and sending transactions to
 * update the LineaSequencerUptimeFeed contract when the sequencer is down/up.
 *
 * <p>This plugin works by checking the timestamp of the last block and comparing it to the current
 * time. If the last block is older than a configurable threshold, it sends two transactions: 1. A
 * 'down' transaction with the timestamp of the last block 2. An 'up' transaction with the current
 * timestamp
 *
 * <p>These transactions help protocols like Aave to better handle liquidations during sequencer
 * downtime.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LivenessPlugin extends AbstractLineaRequiredPlugin
    implements BesuEvents.BlockAddedListener {

  private static final String PLUGIN_NAME = "LivenessPlugin";
  public static final BigInteger ZERO_TRANSACTION_VALUE = BigInteger.ZERO;
  private static final MetricCategory SEQUENCER_LIVENESS_CATEGORY =
      LineaMetricCategory.SEQUENCER_LIVENESS;

  private Web3j web3j;
  private Credentials credentials;
  private Counter uptimeTransactionsCounter;
  private final AtomicLong uptimeTransactionDownCount = new AtomicLong(0);
  private final AtomicLong uptimeTransactionUpCount = new AtomicLong(0);

  private ScheduledExecutorService scheduler;
  private final AtomicReference<BlockHeader> lastProcessedBlock = new AtomicReference<>();
  private final AtomicLong lastReportedTimestamp = new AtomicLong(0);
  private boolean isPluginEnabled = false;

  private LivenessPluginConfiguration livenessPluginConfiguration;

  @Override
  public void doRegister(final ServiceManager serviceManager) {
    log.info("Registering {} ...", PLUGIN_NAME);

    // Register metric category
    metricCategoryRegistry.addMetricCategory(SEQUENCER_LIVENESS_CATEGORY);

    log.info("{} registered successfully", PLUGIN_NAME);
  }

  @Override
  public void doStart() {
    log.info("Starting {} ...", PLUGIN_NAME);

    // Register for block events (besuEvents is now available from parent class)
    besuEvents.addBlockAddedListener(this);

    // Get configuration
    livenessPluginConfiguration = getConfiguration();

    // Update instance variables from configuration
    isPluginEnabled = livenessPluginConfiguration.enabled();

    if (!isPluginEnabled) {
      log.info("{} is disabled", PLUGIN_NAME);
      return;
    }

    // Initialize metrics if enabled
    if (livenessPluginConfiguration.metricCategoryEnabled()
        && metricCategoryRegistry.isMetricCategoryEnabled(SEQUENCER_LIVENESS_CATEGORY)) {
      // Legacy counter for backward compatibility
      uptimeTransactionsCounter =
          metricsSystem.createCounter(
              SEQUENCER_LIVENESS_CATEGORY,
              "uptime_transactions",
              "Number of sequencer uptime transactions sent");

      // Labeled gauge for better aggregation across instances
      final var labelledUptimeGauge =
          metricsSystem.createLabelledSuppliedGauge(
              SEQUENCER_LIVENESS_CATEGORY,
              "uptime_transactions_total",
              "Total number of sequencer uptime transactions sent by status",
              "status");
      labelledUptimeGauge.labels(uptimeTransactionDownCount::doubleValue, "down");
      labelledUptimeGauge.labels(uptimeTransactionUpCount::doubleValue, "up");
    }

    long checkIntervalMilliSeconds = livenessPluginConfiguration.checkIntervalMilliseconds();

    // Contract address validation handled by CLI options

    // Initialize Web3j client for Web3Signer (validation handled by CLI options)
    String signerUrl = livenessPluginConfiguration.signerUrl();
    String signerKeyId = livenessPluginConfiguration.signerKeyId();
    web3j = Web3j.build(new HttpService(signerUrl));

    log.info("Using Web3Signer with key ID: {}", signerKeyId);

    // TODO: initialize credentials from Web3Signer (for now, we'll use a placeholder)
    credentials =
        Credentials.create("0x0000000000000000000000000000000000000000000000000000000000000000");

    // Initialize with current block
    lastProcessedBlock.set(blockchainService.getChainHeadHeader());
    if (lastProcessedBlock.get() != null) {
      lastReportedTimestamp.set(lastProcessedBlock.get().getTimestamp());
    } else {
      log.warn("No blocks available in the blockchain");
    }

    // Run a first check
    checkBlockTimestampAndReport();

    // Start periodic check with named thread
    scheduler =
        Executors.newSingleThreadScheduledExecutor(
            r -> {
              Thread t = new Thread(r, "liveness-plugin-scheduler");
              t.setDaemon(true);
              return t;
            });

    scheduler.scheduleAtFixedRate(
        this::checkBlockTimestampAndReport,
        checkIntervalMilliSeconds,
        checkIntervalMilliSeconds,
        TimeUnit.MILLISECONDS);

    log.info(
        "{} started with configuration: maxBlockAgeMilliseconds={}, checkIntervalMilliseconds={}, contractAddress={}, signerUrl={}, gasLimit={}",
        PLUGIN_NAME,
        livenessPluginConfiguration.maxBlockAgeMilliseconds(),
        checkIntervalMilliSeconds,
        livenessPluginConfiguration.contractAddress(),
        signerUrl,
        livenessPluginConfiguration.gasLimit());
  }

  @Override
  public void stop() {
    log.info("Stopping {} ...", PLUGIN_NAME);

    if (scheduler != null) {
      scheduler.shutdown();
      try {
        if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
          scheduler.shutdownNow();
        }
      } catch (InterruptedException e) {
        log.error("Error terminating scheduler: {}", e.getMessage(), e);
        scheduler.shutdownNow();
        Thread.currentThread().interrupt();
      }
    }

    if (web3j != null) {
      web3j.shutdown();
    }

    super.stop();

    log.info("{} stopped", PLUGIN_NAME);
  }

  /**
   * Gets the plugin configuration.
   *
   * @return The LivenessPluginConfiguration
   */
  private LivenessPluginConfiguration getConfiguration() {
    return livenessPluginConfiguration();
  }

  @Override
  public void onBlockAdded(AddedBlockContext addedBlockContext) {
    if (!isPluginEnabled) return;

    BlockHeader newBlock = addedBlockContext.getBlockHeader();
    lastProcessedBlock.set(newBlock);

    // Reset the last reported timestamp when a new block is added
    lastReportedTimestamp.set(newBlock.getTimestamp());

    log.debug(
        "New block added: lastProcessedBlock ={}, lastReportedTimestamp ={}",
        lastProcessedBlock.get().getNumber(),
        lastReportedTimestamp.get());
  }

  /**
   * Checks the timestamp of the last block and reports downtime if necessary. This method is called
   * periodically by the scheduler.
   */
  private void checkBlockTimestampAndReport() {
    if (!isPluginEnabled) return;

    try {
      BlockHeader lastBlock = lastProcessedBlock.get();
      if (lastBlock == null) {
        log.warn("No blocks available in the blockchain");
        return;
      }

      long currentTimestamp = Instant.now().getEpochSecond();
      long lastBlockTimestamp = lastBlock.getTimestamp();
      long timeSinceLastBlock = currentTimestamp - lastBlockTimestamp;

      log.debug(
          "Checking block timestamp: lastBlockNumber={}, lastBlockTimestamp={}, currentTimestamp={}, timeSinceLastBlock={}s",
          lastBlock.getNumber(),
          lastBlockTimestamp,
          currentTimestamp,
          timeSinceLastBlock);

      // Check if we need to report downtime
      if (timeSinceLastBlock > livenessPluginConfiguration.maxBlockAgeMilliseconds()) {
        // Only report if we haven't reported recently or if significant time has passed
        long timeSinceLastReport = currentTimestamp - lastReportedTimestamp.get();
        if (timeSinceLastReport > livenessPluginConfiguration.maxBlockAgeMilliseconds()) {
          log.info(
              "Sequencer appears to have been down: lastBlockNumber={}, lastBlockTimestamp={}, timeSinceLastBlock={}s",
              lastBlock.getNumber(),
              lastBlockTimestamp,
              timeSinceLastBlock);

          sendSequencerUptimeTransaction(lastBlockTimestamp, currentTimestamp);
          lastReportedTimestamp.set(currentTimestamp);

          // Update lastProcessedBlock to current chain head to avoid re-triggering
          BlockHeader currentBlock = blockchainService.getChainHeadHeader();
          if (currentBlock != null) {
            lastProcessedBlock.set(currentBlock);
            log.debug(
                "Updated lastProcessedBlock to current chain head: {}", currentBlock.getNumber());
          }

          if (uptimeTransactionsCounter != null) {
            uptimeTransactionsCounter.inc(2); // We send two transactions (legacy counter)
            // Update labeled counters
            uptimeTransactionDownCount.incrementAndGet(); // down transaction
            uptimeTransactionUpCount.incrementAndGet(); // up transaction
          }
        }
      }
    } catch (RuntimeException e) {
      log.error("Unexpected error in checkBlockTimestampAndReport", e);
    }
  }

  /**
   * Creates the function call data for the LineaSequencerUptimeFeed contract.
   *
   * @param isUp true if the sequencer is up, false if it's down
   * @param timestamp the timestamp to report
   * @return the encoded function call data
   */
  private Bytes createFunctionCallData(boolean isUp, long timestamp) {
    Function function =
        new Function(
            "updateStatus",
            Arrays.asList(new Bool(isUp), new Uint256(timestamp)),
            Collections.emptyList());

    String encodedFunction = FunctionEncoder.encode(function);
    byte[] callDataBytes = Numeric.hexStringToByteArray(encodedFunction);
    return Bytes.wrap(callDataBytes);
  }

  /**
   * Submits a transaction to update the LineaSequencerUptimeFeed contract.
   *
   * @param isUp true if the sequencer is up, false if it's down
   * @param timestamp the timestamp to report
   * @throws IOException if there's an error submitting the transaction
   */
  private void submitUptimeTransaction(boolean isUp, long timestamp) throws IOException {
    Bytes callData = createFunctionCallData(isUp, timestamp);
    RawTransaction rawTransaction = createTransaction(callData);
    Transaction transaction = signTransaction(rawTransaction);
    submitTransaction(transaction);
  }

  /**
   * Sends two transactions to update the LineaSequencerUptimeFeed contract: 1. A 'down' transaction
   * with the timestamp of the last block 2. An 'up' transaction with the current timestamp
   *
   * @param lastBlockTimestamp the timestamp of the last block
   * @param currentTimestamp the current timestamp
   */
  private void sendSequencerUptimeTransaction(long lastBlockTimestamp, long currentTimestamp) {
    try {
      log.info(
          "Sending sequencer uptime transaction: lastBlockTimestamp={}, currentTimestamp={}",
          lastBlockTimestamp,
          currentTimestamp);

      // First transaction: mark as down with last block timestamp
      submitUptimeTransaction(false, lastBlockTimestamp);

      // Second transaction: mark as up with current timestamp
      submitUptimeTransaction(true, currentTimestamp);

      log.info("Sequencer uptime transactions submitted via JSON-RPC.");
    } catch (IOException e) {
      log.error("Error sending sequencer uptime transaction", e);
    }
  }

  /**
   * Creates a raw transaction to call the LineaSequencerUptimeFeed contract.
   *
   * @param callData the encoded function call data
   * @return the raw transaction
   * @throws IOException if there's an error creating the transaction
   */
  private RawTransaction createTransaction(Bytes callData) throws IOException {
    // Get nonce from the account (web3j and credentials are guaranteed to be non-null after
    // doStart())
    EthGetTransactionCount nonceResponse =
        web3j
            .ethGetTransactionCount(credentials.getAddress(), DefaultBlockParameterName.LATEST)
            .send();

    if (nonceResponse.hasError()) {
      throw new IOException("Error getting nonce: " + nonceResponse.getError().getMessage());
    }

    BigInteger nonce = nonceResponse.getTransactionCount();

    // TODO: get current gas price (currently using a 1 Gwei default value)
    Wei gasPrice = Wei.of(1_000_000_000L);

    // Create transaction
    return RawTransaction.createTransaction(
        nonce,
        gasPrice.getAsBigInteger(),
        BigInteger.valueOf(livenessPluginConfiguration.gasLimit()),
        Address.fromHexString(livenessPluginConfiguration.contractAddress()).toString(),
        ZERO_TRANSACTION_VALUE,
        callData.toHexString());
  }

  /**
   * Signs a raw transaction using the credentials.
   *
   * @param rawTransaction the raw transaction to sign
   * @return the signed transaction
   */
  private Transaction signTransaction(RawTransaction rawTransaction) {
    BigInteger chainId =
        blockchainService
            .getChainId()
            .orElseThrow(() -> new IllegalArgumentException("Chain id required"));
    byte[] signedMessage =
        TransactionEncoder.signMessage(rawTransaction, chainId.longValue(), credentials);
    String hexValue = Numeric.toHexString(signedMessage);
    return Transaction.readFrom(Bytes.fromHexString(hexValue));
  }

  /**
   * Submits a transaction to the transaction pool using eth_sendRawTransaction.
   *
   * @param transaction the transaction to submit
   */
  private void submitTransaction(Transaction transaction) {
    try {
      // Submit transaction to the transaction pool via eth_sendRawTransaction
      String transactionRlp = transaction.encoded().toHexString();

      // Use Web3j to submit the transaction properly
      org.web3j.protocol.core.methods.response.EthSendTransaction result =
          web3j.ethSendRawTransaction(transactionRlp).send();

      if (result.hasError()) {
        log.error("Failed to submit transaction: {}", result.getError().getMessage());
      } else {
        log.info("Transaction submitted successfully: {}", result.getTransactionHash());
      }
    } catch (Exception e) {
      log.error("Failed to submit transaction to transaction pool", e);
    }
  }
}
