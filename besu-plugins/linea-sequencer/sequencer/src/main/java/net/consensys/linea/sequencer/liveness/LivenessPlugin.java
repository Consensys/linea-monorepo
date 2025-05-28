/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.liveness;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.auto.service.AutoService;
import java.io.IOException;
import java.math.BigInteger;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.Instant;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
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
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.DefaultBlockParameterName;
import org.web3j.protocol.core.methods.response.EthGetTransactionCount;
import org.web3j.protocol.http.HttpService;
import org.web3j.utils.Numeric;

/**
 * The LivenessPlugin is monitoring the blockchain and sending transactions to update the
 * LineaSequencerUptimeFeed contract when the sequencer is down/up.
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
  private HttpClient httpClient;
  private ObjectMapper objectMapper;
  private String signerKeyId;
  private Counter uptimeTransactionsCounter;
  private Counter transactionRetryCounter;
  private Counter transactionFailureCounter;
  private final AtomicLong uptimeTransactionDownCount = new AtomicLong(0);
  private final AtomicLong uptimeTransactionUpCount = new AtomicLong(0);

  private ScheduledExecutorService scheduler;
  private final AtomicReference<BlockHeader> lastProcessedBlock = new AtomicReference<>();
  private final AtomicLong lastReportedTimestamp = new AtomicLong(0);
  private boolean isPluginEnabled = false;

  private LivenessPluginConfiguration livenessPluginConfiguration;

  // Constants for error classification
  private static final Set<String> NON_RETRIABLE_ERROR_KEYWORDS =
      Set.of(
          "insufficient funds",
          "gas limit",
          "gas price",
          "nonce",
          "invalid",
          "malformed",
          "unauthorized",
          "forbidden");

  private static final Set<String> RETRIABLE_ERROR_KEYWORDS =
      Set.of("timeout", "connection", "network", "unavailable", "busy", "overloaded", "rate limit");

  private static final Set<Integer> RETRIABLE_ERROR_CODES =
      Set.of(
          -32000, // Generic server error
          -32001, // Resource unavailable
          -32002, // Resource not found
          -32003 // Transaction rejected (maybe temporary)
          );

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
      uptimeTransactionsCounter =
          metricsSystem.createCounter(
              SEQUENCER_LIVENESS_CATEGORY,
              "uptime_transactions",
              "Number of sequencer uptime transactions sent");

      transactionRetryCounter =
          metricsSystem.createCounter(
              SEQUENCER_LIVENESS_CATEGORY,
              "transaction_retries",
              "Number of transaction submission retry attempts");

      transactionFailureCounter =
          metricsSystem.createCounter(
              SEQUENCER_LIVENESS_CATEGORY,
              "transaction_failures",
              "Number of transaction submission failures after all retries");

      // Labeled gauge for better aggregation across instances
      final var labelledUptimeGauge =
          metricsSystem.createLabelledSuppliedGauge(
              SEQUENCER_LIVENESS_CATEGORY,
              "uptime_transactions",
              "Total number of sequencer uptime transactions sent by status",
              "status");
      labelledUptimeGauge.labels(uptimeTransactionDownCount::doubleValue, "down");
      labelledUptimeGauge.labels(uptimeTransactionUpCount::doubleValue, "up");
    }

    long checkIntervalMilliSeconds = livenessPluginConfiguration.checkIntervalMilliseconds();

    // Contract address validation handled by CLI options

    // Initialize Web3j client for Web3Signer (validation handled by CLI options)
    String signerUrl = livenessPluginConfiguration.signerUrl();
    signerKeyId = livenessPluginConfiguration.signerKeyId();
    web3j = Web3j.build(new HttpService(signerUrl));

    log.info("Using Web3Signer with key ID: {}", signerKeyId);

    // Initialize HTTP client and JSON mapper for Web3Signer API calls
    httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(30)).build();
    objectMapper = new ObjectMapper();

    // Initialize with current block
    lastProcessedBlock.set(blockchainService.getChainHeadHeader());
    if (lastProcessedBlock.get() != null) {
      lastReportedTimestamp.set(
          lastProcessedBlock.get().getTimestamp() * 1000); // Convert to milliseconds
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
        "{} started with configuration: maxBlockAgeMilliseconds={}, checkIntervalMilliseconds={}, contractAddress={}, signerUrl={}, signerAddress={}, gasLimit={}, maxRetryAttempts={}, retryDelayMilliseconds={}",
        PLUGIN_NAME,
        livenessPluginConfiguration.maxBlockAgeMilliseconds(),
        checkIntervalMilliSeconds,
        livenessPluginConfiguration.contractAddress(),
        signerUrl,
        livenessPluginConfiguration.signerAddress(),
        livenessPluginConfiguration.gasLimit(),
        livenessPluginConfiguration.maxRetryAttempts(),
        livenessPluginConfiguration.retryDelayMilliseconds());
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

    if (httpClient != null) {
      // HttpClient doesn't have an explicit close method, but resources are automatically released
      httpClient = null;
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
    lastReportedTimestamp.set(newBlock.getTimestamp() * 1000); // Convert to milliseconds

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
        // Try to get current chain head
        lastBlock = blockchainService.getChainHeadHeader();
        if (lastBlock != null) {
          lastProcessedBlock.set(lastBlock);
          log.debug("Retrieved chain head header: {}", lastBlock.getNumber());
        } else {
          // No blocks available - treat this as potential downtime
          long currentTimestamp = Instant.now().toEpochMilli();
          long timeSinceLastReport = currentTimestamp - lastReportedTimestamp.get();

          // If we haven't reported recently and significant time has passed since plugin start
          if (timeSinceLastReport > livenessPluginConfiguration.maxBlockAgeMilliseconds()) {
            log.warn("No blocks available in the blockchain - potential sequencer downtime");

            // Report downtime using current timestamp as both down and up timestamps
            // This indicates the sequencer has been down since the last report
            boolean[] transactionResults =
                sendSequencerUptimeTransaction(
                    Math.max(
                            lastReportedTimestamp.get(),
                            currentTimestamp
                                - livenessPluginConfiguration.maxBlockAgeMilliseconds())
                        / 1000,
                    currentTimestamp / 1000);
            boolean downTransactionSuccess = transactionResults[0];
            boolean upTransactionSuccess = transactionResults[1];

            // Only update lastReportedTimestamp if at least one transaction succeeded
            if (downTransactionSuccess || upTransactionSuccess) {
              lastReportedTimestamp.set(currentTimestamp);
            }

            // Update metrics for successful transactions
            updateUptimeMetrics(downTransactionSuccess, upTransactionSuccess);
          }
          return;
        }
      }

      long currentTimestamp = Instant.now().toEpochMilli();
      long lastBlockTimestamp = lastBlock.getTimestamp() * 1000; // Convert seconds to milliseconds
      long timeSinceLastBlock = currentTimestamp - lastBlockTimestamp;

      log.debug(
          "Checking block timestamp: lastBlockNumber={}, lastBlockTimestamp={}ms, currentTimestamp={}ms, timeSinceLastBlock={}ms",
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
              "Sequencer appears to have been down: lastBlockNumber={}, lastBlockTimestamp={}ms, timeSinceLastBlock={}ms",
              lastBlock.getNumber(),
              lastBlockTimestamp,
              timeSinceLastBlock);

          boolean[] transactionResults =
              sendSequencerUptimeTransaction(lastBlockTimestamp / 1000, currentTimestamp / 1000);
          boolean downTransactionSuccess = transactionResults[0];
          boolean upTransactionSuccess = transactionResults[1];

          // Only update lastReportedTimestamp if at least one transaction succeeded
          if (downTransactionSuccess || upTransactionSuccess) {
            lastReportedTimestamp.set(currentTimestamp);
          }

          // Update lastProcessedBlock to current chain head to avoid re-triggering
          BlockHeader currentBlock = blockchainService.getChainHeadHeader();
          if (currentBlock != null) {
            lastProcessedBlock.set(currentBlock);
            log.debug(
                "Updated lastProcessedBlock to current chain head: {}", currentBlock.getNumber());
          }

          // Only increment metrics for successful transactions
          updateUptimeMetrics(downTransactionSuccess, upTransactionSuccess);
        }
      }
    } catch (RuntimeException e) {
      log.error("Unexpected error in checkBlockTimestampAndReport", e);
    }
  }

  /**
   * Creates the function call data for the LineaSequencerUptimeFeed contract.
   *
   * @param isUp true if the sequencer is up, false if it is down
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
   * @param isUp true if the sequencer is up, false if it is down
   * @param timestamp the timestamp to report
   * @return true if the transaction was successfully submitted, false otherwise
   * @throws IOException if there's an error creating, signing, or submitting the transaction after
   *     all retries
   */
  private boolean submitUptimeTransaction(boolean isUp, long timestamp) throws IOException {
    Bytes callData = createFunctionCallData(isUp, timestamp);
    RawTransaction rawTransaction = createTransaction(callData);
    Transaction transaction = signTransaction(rawTransaction);
    return submitTransaction(transaction);
  }

  /**
   * Sends two transactions to update the LineaSequencerUptimeFeed contract: 1. A 'down' transaction
   * with the timestamp of the last block 2. An 'up' transaction with the current timestamp
   *
   * @param lastBlockTimestamp the timestamp of the last block
   * @param currentTimestamp the current timestamp
   * @return an array with two elements: [downTransactionSuccess, upTransactionSuccess]
   */
  private boolean[] sendSequencerUptimeTransaction(long lastBlockTimestamp, long currentTimestamp) {
    boolean downTransactionSuccess = false;
    boolean upTransactionSuccess = false;

    try {
      log.info(
          "Sending sequencer uptime transaction: lastBlockTimestamp={}, currentTimestamp={}",
          lastBlockTimestamp,
          currentTimestamp);

      // First transaction: mark as down with last block timestamp
      downTransactionSuccess = submitUptimeTransaction(false, lastBlockTimestamp);

      // Second transaction: mark as up with current timestamp
      upTransactionSuccess = submitUptimeTransaction(true, currentTimestamp);

      if (downTransactionSuccess && upTransactionSuccess) {
        log.info("Sequencer uptime transactions submitted via JSON-RPC.");
      } else {
        log.warn(
            "Some sequencer uptime transactions failed: down={}, up={}",
            downTransactionSuccess,
            upTransactionSuccess);
      }
    } catch (IOException e) {
      log.error("Error sending sequencer uptime transaction", e);
    }

    return new boolean[] {downTransactionSuccess, upTransactionSuccess};
  }

  /**
   * Creates a raw transaction to call the LineaSequencerUptimeFeed contract.
   *
   * @param callData the encoded function call data
   * @return the raw transaction
   * @throws IOException if there's an error creating the transaction
   */
  private RawTransaction createTransaction(Bytes callData) throws IOException {
    // Get nonce from the Web3Signer account address
    String signerAddress = livenessPluginConfiguration.signerAddress();
    EthGetTransactionCount nonceResponse =
        web3j.ethGetTransactionCount(signerAddress, DefaultBlockParameterName.LATEST).send();

    if (nonceResponse.hasError()) {
      throw new IOException("Error getting nonce: " + nonceResponse.getError().getMessage());
    }

    BigInteger nonce = nonceResponse.getTransactionCount();

    // Validate nonce is not null and is non-negative
    if (nonce == null) {
      throw new IOException("Received null nonce from Web3j for address: " + signerAddress);
    }

    if (nonce.compareTo(BigInteger.ZERO) < 0) {
      throw new IOException(
          "Received invalid negative nonce: " + nonce + " for address: " + signerAddress);
    }

    log.debug("Retrieved valid nonce: {} for address: {}", nonce, signerAddress);

    // Get gas price - either configured value or fetch dynamically
    Wei gasPrice = getGasPrice();

    // Validate and get gas limit
    long gasLimit = getValidatedGasLimit();

    // Create transaction
    return RawTransaction.createTransaction(
        nonce,
        gasPrice.getAsBigInteger(),
        BigInteger.valueOf(gasLimit),
        Address.fromHexString(livenessPluginConfiguration.contractAddress()).toString(),
        ZERO_TRANSACTION_VALUE,
        callData.toHexString());
  }

  /**
   * Gets the gas price for transactions. If configured gas price is 0 (dynamic), fetches the
   * current gas price from the network. Otherwise, uses the configured value.
   *
   * @return the gas price in Wei
   */
  private Wei getGasPrice() {
    long configuredGasPriceGwei = livenessPluginConfiguration.gasPriceGwei();

    if (configuredGasPriceGwei > 0) {
      // Use configured gas price
      long gasPriceWei = configuredGasPriceGwei * 1_000_000_000L; // Convert Gwei to Wei
      log.debug(
          "Using configured gas price: {} Gwei ({} Wei)", configuredGasPriceGwei, gasPriceWei);
      return Wei.of(gasPriceWei);
    }

    // Fetch dynamic gas price from network
    try {
      BigInteger networkGasPrice = web3j.ethGasPrice().send().getGasPrice();

      if (networkGasPrice == null || networkGasPrice.compareTo(BigInteger.ZERO) <= 0) {
        // Fallback to default if network price is invalid
        long fallbackGasPriceWei = 1_000_000_000L; // 1 Gwei fallback
        log.warn(
            "Invalid network gas price ({}), using fallback: {} Wei",
            networkGasPrice,
            fallbackGasPriceWei);
        return Wei.of(fallbackGasPriceWei);
      }

      // Add 10% buffer to ensure the transaction goes through quickly
      BigInteger bufferedGasPrice =
          networkGasPrice.multiply(BigInteger.valueOf(110)).divide(BigInteger.valueOf(100));
      log.debug(
          "Using dynamic gas price: {} Wei (network: {} Wei + 10% buffer)",
          bufferedGasPrice, networkGasPrice);
      return Wei.of(bufferedGasPrice);

    } catch (Exception e) {
      // Fallback to default on any error
      long fallbackGasPriceWei = 1_000_000_000L; // 1 Gwei fallback
      log.warn(
          "Error fetching dynamic gas price, using fallback: {} Wei. Error: {}",
          fallbackGasPriceWei,
          e.getMessage());
      return Wei.of(fallbackGasPriceWei);
    }
  }

  /**
   * Validates and returns the gas limit for transactions. Ensures the gas limit is positive and
   * within reasonable bounds.
   *
   * @return the validated gas limit
   * @throws IOException if the gas limit is invalid
   */
  private long getValidatedGasLimit() throws IOException {
    long configuredGasLimit = livenessPluginConfiguration.gasLimit();

    // Minimum gas limit for a contract call (21_000 for simple transfer plus some overhead)
    long minimumGasLimit = 21000L;
    // Maximum reasonable gas limit (to prevent accidentally high values)
    long maximumGasLimit = 10_000_000L;

    if (configuredGasLimit <= 0) {
      throw new IOException("Gas limit must be positive, but was: " + configuredGasLimit);
    }

    if (configuredGasLimit < minimumGasLimit) {
      log.warn(
          "Configured gas limit ({}) is below minimum ({}), using minimum",
          configuredGasLimit,
          minimumGasLimit);
      return minimumGasLimit;
    }

    if (configuredGasLimit > maximumGasLimit) {
      log.warn(
          "Configured gas limit ({}) exceeds maximum ({}), using maximum",
          configuredGasLimit,
          maximumGasLimit);
      return maximumGasLimit;
    }

    log.debug("Using validated gas limit: {}", configuredGasLimit);
    return configuredGasLimit;
  }

  /**
   * Signs a raw transaction using Web3Signer.
   *
   * @param rawTransaction the raw transaction to sign
   * @return the signed transaction
   */
  private String signTransactionWithWeb3Signer(RawTransaction rawTransaction) throws IOException {
    try {
      // Get the unsigned serialized transaction
      String unsignedTransactionHex =
          Numeric.toHexString(TransactionEncoder.encode(rawTransaction));

      // Prepare the request body for Web3Signer
      Map<String, String> requestBody = new HashMap<>();
      requestBody.put("data", unsignedTransactionHex);
      String jsonBody = objectMapper.writeValueAsString(requestBody);

      // Create HTTP request to Web3Signer
      String web3SignerUrl = livenessPluginConfiguration.signerUrl();
      String endpoint = web3SignerUrl + "/api/v1/eth1/sign/" + signerKeyId;

      HttpRequest request =
          HttpRequest.newBuilder()
              .uri(URI.create(endpoint))
              .header("Content-Type", "application/json")
              .timeout(Duration.ofSeconds(30))
              .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
              .build();

      // Send request and get response
      HttpResponse<String> response =
          httpClient.send(request, HttpResponse.BodyHandlers.ofString());

      if (response.statusCode() != 200) {
        String responseBody = response.body();
        String bodyDescription = responseBody != null ? responseBody : "<null>";
        throw new IOException(
            "Web3Signer API call failed with status: "
                + response.statusCode()
                + ", body: "
                + bodyDescription);
      }

      // The response should be the signed transaction hex string
      String responseBody = response.body();
      if (responseBody == null) {
        throw new IOException("Web3Signer API returned null response body");
      }

      String signedTransactionHex = responseBody.trim();

      if (signedTransactionHex.isEmpty()) {
        throw new IOException("Web3Signer API returned empty response body");
      }

      // Remove quotes if present (some APIs return quoted strings)
      if (signedTransactionHex.startsWith("\"") && signedTransactionHex.endsWith("\"")) {
        signedTransactionHex = signedTransactionHex.substring(1, signedTransactionHex.length() - 1);
      }

      log.debug("Successfully signed transaction with Web3Signer");
      return signedTransactionHex;

    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      throw new IOException("Web3Signer request was interrupted", e);
    }
  }

  /**
   * Signs a raw transaction using Web3Signer.
   *
   * @param rawTransaction the raw transaction to sign
   * @return the signed transaction
   * @throws IOException if signing fails, or the signed transaction is invalid
   */
  private Transaction signTransaction(RawTransaction rawTransaction) throws IOException {
    String signedTransactionHex = signTransactionWithWeb3Signer(rawTransaction);

    // Additional validation layer (should not be needed due to signTransactionWithWeb3Signer
    // validation,
    // but provides defense in depth)
    if (signedTransactionHex.trim().isEmpty()) {
      throw new IOException("Signed transaction hex is null or empty");
    }

    try {
      return Transaction.readFrom(Bytes.fromHexString(signedTransactionHex));
    } catch (IllegalArgumentException e) {
      throw new IOException("Failed to parse signed transaction hex: " + e.getMessage(), e);
    } catch (Exception e) {
      throw new IOException("Unexpected error parsing signed transaction: " + e.getMessage(), e);
    }
  }

  /**
   * Submits a transaction to the transaction pool using eth_sendRawTransaction with retry logic.
   * This method implements robust error handling including null validation, exponential backoff
   * retry logic, and comprehensive error reporting.
   *
   * @param transaction the transaction to submit
   * @return true if the transaction was successfully submitted, false otherwise
   * @throws IllegalArgumentException if transaction is null
   * @throws IOException if all retry attempts fail for critical operations
   */
  private boolean submitTransaction(Transaction transaction) throws IOException {
    // Null validation
    if (transaction == null) {
      IllegalArgumentException e = new IllegalArgumentException("Transaction cannot be null");
      log.error("Transaction submission failed: transaction is null");
      if (transactionFailureCounter != null) {
        transactionFailureCounter.inc();
      }
      throw e;
    }

    final int maxRetries = livenessPluginConfiguration.maxRetryAttempts();
    final long baseDelayMs = livenessPluginConfiguration.retryDelayMilliseconds();

    log.debug(
        "Starting transaction submission with {} max retries and {}ms base delay",
        maxRetries,
        baseDelayMs);

    Exception lastException = null;

    for (int attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        log.debug("Transaction submission attempt {} of {}", attempt + 1, maxRetries + 1);

        // Submit transaction to the transaction pool via eth_sendRawTransaction
        String transactionRlp = transaction.encoded().toHexString();

        if (transactionRlp == null || transactionRlp.isEmpty()) {
          throw new IOException("Failed to encode transaction: encoded RLP is null or empty");
        }

        // Use Web3j to submit the transaction properly
        org.web3j.protocol.core.methods.response.EthSendTransaction result =
            web3j.ethSendRawTransaction(transactionRlp).send();

        // Validate the response
        if (result == null) {
          throw new IOException("Received null response from eth_sendRawTransaction");
        }

        if (result.hasError()) {
          String errorMessage =
              result.getError() != null ? result.getError().getMessage() : "Unknown error";
          int errorCode = result.getError() != null ? result.getError().getCode() : -1;

          // Check if this is a retriable error
          if (isRetriableError(errorCode, errorMessage)) {
            throw new IOException(
                String.format(
                    "Retriable error from eth_sendRawTransaction (code: %d): %s",
                    errorCode, errorMessage));
          } else {
            // Non-retriable error - fail immediately
            log.error(
                "Non-retriable error from eth_sendRawTransaction (code: {}): {}",
                errorCode,
                errorMessage);
            if (transactionFailureCounter != null) {
              transactionFailureCounter.inc();
            }
            return false;
          }
        }

        // Success case
        String txHash = result.getTransactionHash();
        if (txHash == null || txHash.isEmpty()) {
          log.warn("Transaction submitted but received null/empty transaction hash");
        } else {
          log.info("Transaction submitted successfully: {}", txHash);
        }

        return true;

      } catch (Exception e) {
        lastException = e;

        if (attempt < maxRetries) {
          // Calculate exponential backoff delay: baseDelay * 2^attempt
          long delayMs = baseDelayMs * (1L << attempt);
          // Cap the delay at 30 seconds to prevent excessively long waits
          delayMs = Math.min(delayMs, 30_000L);

          log.warn(
              "Transaction submission attempt {} failed (will retry in {}ms): {}",
              attempt + 1,
              delayMs,
              e.getMessage());

          if (transactionRetryCounter != null) {
            transactionRetryCounter.inc();
          }

          try {
            Thread.sleep(delayMs);
          } catch (InterruptedException ie) {
            Thread.currentThread().interrupt();
            log.error("Transaction submission interrupted during retry delay");
            throw new IOException("Transaction submission interrupted", ie);
          }
        } else {
          // Final attempt failed
          log.error(
              "Transaction submission failed after {} attempts. Final error: {}",
              maxRetries + 1,
              e.getMessage(),
              e);

          if (transactionFailureCounter != null) {
            transactionFailureCounter.inc();
          }
        }
      }
    }

    // All retries exhausted
    String errorMsg =
        String.format("Transaction submission failed after %d attempts", maxRetries + 1);
    if (lastException != null) {
      throw new IOException(errorMsg, lastException);
    } else {
      throw new IOException(errorMsg);
    }
  }

  /**
   * Determines if an error from eth_sendRawTransaction is retriable. Some errors indicate temporary
   * network issues that may resolve on retry, while others indicate permanent issues that retrying
   * won't fix.
   *
   * @param errorCode the error code from the JSON-RPC response
   * @param errorMessage the error message from the JSON-RPC response
   * @return true if the error should be retried, false if it is permanent
   */
  private boolean isRetriableError(int errorCode, String errorMessage) {
    if (errorMessage == null) {
      return false;
    }

    String lowerMessage = errorMessage.toLowerCase();

    // Check for non-retriable errors first (permanent failures)
    if (containsAnyKeyword(lowerMessage, NON_RETRIABLE_ERROR_KEYWORDS)) {
      return false;
    }

    // Check for retriable errors
    return containsAnyKeyword(lowerMessage, RETRIABLE_ERROR_KEYWORDS)
        || RETRIABLE_ERROR_CODES.contains(errorCode);
  }

  /**
   * Checks if the message contains any of the specified keywords as substrings. Uses Stream API for
   * concise and potentially parallel execution.
   */
  private boolean containsAnyKeyword(String message, Set<String> keywords) {
    return keywords.stream().anyMatch(message::contains);
  }

  /**
   * Updates uptime metrics based on transaction success status. This method consolidates metrics
   * update logic to avoid code duplication.
   *
   * @param downTransactionSuccess true if the down transaction was successful
   * @param upTransactionSuccess true if the up transaction was successful
   */
  private void updateUptimeMetrics(boolean downTransactionSuccess, boolean upTransactionSuccess) {
    if (uptimeTransactionsCounter != null) {
      int successfulTransactions = 0;
      if (downTransactionSuccess) {
        uptimeTransactionDownCount.incrementAndGet();
        successfulTransactions++;
      }
      if (upTransactionSuccess) {
        uptimeTransactionUpCount.incrementAndGet();
        successfulTransactions++;
      }
      if (successfulTransactions > 0) {
        uptimeTransactionsCounter.inc(successfulTransactions);
      }
    }
  }
}
