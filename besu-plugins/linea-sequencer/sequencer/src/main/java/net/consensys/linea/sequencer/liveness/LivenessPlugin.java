/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package net.consensys.linea.sequencer.liveness;

import java.io.IOException;
import java.math.BigInteger;
import java.time.Instant;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.atomic.AtomicReference;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.config.LineaNodeType;
import net.consensys.linea.config.LineaRejectedTxReportingConfiguration;
import net.consensys.linea.jsonrpc.JsonRpcManager;
import net.consensys.linea.jsonrpc.JsonRpcRequestBuilder;
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
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
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
  private static final BigInteger DEFAULT_CHAIN_ID = new BigInteger("59144");
  public static final BigInteger ZERO_TRANSACTION_VALUE = BigInteger.ZERO;
  private static final MetricCategory SEQUENCER_LIVENESS_CATEGORY =
      LineaMetricCategory.SEQUENCER_LIVENESS;

  private BlockchainService blockchainService;
  private Web3j web3j;
  private Credentials credentials;

  private ScheduledExecutorService scheduler;
  private final AtomicReference<BlockHeader> lastProcessedBlock = new AtomicReference<>();
  private final AtomicLong lastReportedTimestamp = new AtomicLong(0);
  private boolean isPluginEnabled = false;

  // Configuration options
  private long maxBlockAgeSeconds;
  private Address livenessStateContractAddress;
  private long gasLimit;
  private BigInteger chainId;

  private JsonRpcManager jsonRpcManager;
  private Counter uptimeTransactionsCounter;

  @Override
  public void doRegister(final ServiceManager serviceManager) {
    log.info("Registering {} ...", PLUGIN_NAME);

    try {
      // Register CLI options
      serviceManager
          .getService(PicoCLIOptions.class)
          .orElseThrow(() -> new RuntimeException("Failed to obtain PicoCLIOptions"))
          .addPicoCLIOptions(
              LivenessPluginCliOptions.CONFIG_KEY, LivenessPluginCliOptions.create());

      // Get required services
      blockchainService =
          serviceManager
              .getService(BlockchainService.class)
              .orElseThrow(() -> new RuntimeException("Failed to obtain BlockchainService"));

      // Register for block events if available
      serviceManager
          .getService(BesuEvents.class)
          .ifPresent(events -> events.addBlockAddedListener(this));

      // Register metric category
      metricCategoryRegistry.addMetricCategory(SEQUENCER_LIVENESS_CATEGORY);

      final LineaRejectedTxReportingConfiguration lineaRejectedTxReportingConfiguration =
          rejectedTxReportingConfiguration();

      jsonRpcManager =
          new JsonRpcManager(
              PLUGIN_NAME, besuConfiguration.getDataPath(), lineaRejectedTxReportingConfiguration);

      log.info("{} registered successfully", PLUGIN_NAME);
    } catch (Exception e) {
      log.error("{} registration failed: {}", PLUGIN_NAME, e.getMessage(), e);
      throw e;
    }
  }

  @Override
  public void doStart() {
    log.info("Starting {} ...", PLUGIN_NAME);

    try {
      // Get configuration
      final LivenessPluginConfiguration config = getConfiguration();

      // Update instance variables from configuration
      isPluginEnabled = config.isEnabled();

      if (!isPluginEnabled) {
        log.info("{} is disabled", PLUGIN_NAME);
        return;
      }

      // Initialize metrics if enabled
      if (config.isMetricCategoryEnabled()
          && metricCategoryRegistry != null
          && metricsSystem != null
          && metricCategoryRegistry.isMetricCategoryEnabled(SEQUENCER_LIVENESS_CATEGORY)) {
        uptimeTransactionsCounter =
            metricsSystem.createCounter(
                SEQUENCER_LIVENESS_CATEGORY,
                "uptime_transactions",
                "Number of sequencer uptime transactions sent");
      }

      maxBlockAgeSeconds = config.getMaxBlockAgeSeconds();
      long checkIntervalSeconds = config.getCheckIntervalSeconds();

      // Validate contract address
      if (config.getContractAddress() == null
          || config.getContractAddress().isEmpty()
          || config
              .getContractAddress()
              .equals(LivenessPluginCliOptions.DEFAULT_CONTRACT_ADDRESS)) {
        throw new IllegalArgumentException(
            "Invalid contract address configuration for LivenessPlugin");
      }
      livenessStateContractAddress = Address.fromHexString(config.getContractAddress());

      // Validate signer configuration
      String signerUrl = config.getSignerUrl();
      String signerKeyId = config.getSignerKeyId();
      if (signerUrl == null
          || signerUrl.isEmpty()
          || signerKeyId == null
          || signerKeyId.isEmpty()) {
        throw new IllegalArgumentException("Web3Signer URL and key ID must be provided");
      }

      gasLimit = config.getGasLimit();
      chainId = blockchainService.getChainId().orElse(DEFAULT_CHAIN_ID);

      // Initialize Web3j client for Web3Signer
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
          checkIntervalSeconds,
          checkIntervalSeconds,
          TimeUnit.SECONDS);

      log.info(
          "{} started with configuration: maxBlockAgeSeconds={}, checkIntervalSeconds={}, contractAddress={}, signerUrl={}, gasLimit={}",
          PLUGIN_NAME,
          maxBlockAgeSeconds,
          checkIntervalSeconds,
          livenessStateContractAddress,
          signerUrl,
          gasLimit);

    } catch (Exception e) {
      log.error("{} startup failed: {}", PLUGIN_NAME, e.getMessage(), e);
      throw e; // Let AbstractLineaRequiredPlugin handle this exception
    }
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

    if (jsonRpcManager != null) {
      jsonRpcManager.shutdown();
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
        "New block added: number={}, timestamp={}", newBlock.getNumber(), newBlock.getTimestamp());
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
      if (timeSinceLastBlock > maxBlockAgeSeconds) {
        // Only report if we haven't reported recently or if significant time has passed
        long timeSinceLastReport = currentTimestamp - lastReportedTimestamp.get();
        if (timeSinceLastReport > maxBlockAgeSeconds) {
          log.info(
              "Sequencer appears to have been down: lastBlockNumber={}, lastBlockTimestamp={}, timeSinceLastBlock={}s",
              lastBlock.getNumber(),
              lastBlockTimestamp,
              timeSinceLastBlock);

          sendSequencerUptimeTransaction(lastBlockTimestamp, currentTimestamp);
          lastReportedTimestamp.set(currentTimestamp);

          if (uptimeTransactionsCounter != null) {
            uptimeTransactionsCounter.inc(2); // We send two transactions
          }
        }
      }
    } catch (Exception e) {
      log.error("Error in checkBlockTimestampAndReport", e);
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
    } catch (Exception e) {
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
    // Get nonce from the account
    BigInteger nonce = BigInteger.ZERO;
    if (web3j != null && credentials != null) {
      EthGetTransactionCount ethGetTransactionCount =
          web3j
              .ethGetTransactionCount(
                  credentials.getAddress(),
                  org.web3j.protocol.core.DefaultBlockParameterName.LATEST)
              .send();
      if (ethGetTransactionCount.hasError()) {
        throw new IOException(
            "Error getting nonce: " + ethGetTransactionCount.getError().getMessage());
      }
      nonce = ethGetTransactionCount.getTransactionCount();
    }

    // TODO: get current gas price (currently using a 1 Gwei default value)
    Wei gasPrice = Wei.of(1_000_000_000L);

    // Create transaction
    return RawTransaction.createTransaction(
        nonce,
        gasPrice.getAsBigInteger(),
        BigInteger.valueOf(gasLimit),
        livenessStateContractAddress.toString(),
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
    byte[] signedMessage =
        TransactionEncoder.signMessage(rawTransaction, chainId.longValue(), credentials);
    String hexValue = Numeric.toHexString(signedMessage);
    return Transaction.readFrom(Bytes.fromHexString(hexValue));
  }

  /**
   * Submits a transaction to the network using JSON-RPC.
   *
   * @param transaction the transaction to submit
   */
  private void submitTransaction(Transaction transaction) {
    try {
      String jsonRpcCall =
          JsonRpcRequestBuilder.generateSaveRejectedTxJsonRpc(
              LineaNodeType.SEQUENCER,
              transaction,
              Instant.now(),
              Optional.empty(),
              "Sequencer uptime transaction",
              List.of());
      jsonRpcManager.submitNewJsonRpcCallAsync(jsonRpcCall);
      log.info("Transaction submitted via JSON-RPC: {}", transaction.getHash());
    } catch (Exception e) {
      log.error("Failed to submit transaction via JSON-RPC", e);
    }
  }
}
