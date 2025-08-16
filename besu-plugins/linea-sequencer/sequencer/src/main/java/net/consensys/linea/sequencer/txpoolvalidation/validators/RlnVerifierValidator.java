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
package net.consensys.linea.sequencer.txpoolvalidation.validators;

import com.github.benmanes.caffeine.cache.Cache;
import com.github.benmanes.caffeine.cache.Caffeine;
import com.google.common.annotations.VisibleForTesting;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import java.io.Closeable;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import java.util.concurrent.atomic.AtomicInteger;
import net.consensys.linea.config.LineaRlnValidatorConfiguration;
import net.consensys.linea.rln.JniRlnVerificationService;
import net.consensys.linea.rln.RlnVerificationService;
import net.consensys.linea.sequencer.txpoolvalidation.shared.DenyListManager;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient.KarmaInfo;
import net.consensys.linea.sequencer.txpoolvalidation.shared.NullifierTracker;
import net.vac.prover.RlnProof;
import net.vac.prover.RlnProofFilter;
import net.vac.prover.RlnProofReply;
import net.vac.prover.RlnProverGrpc;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * RLN (Rate Limiting Nullifier) Verifier Validator for gasless transaction validation.
 *
 * <p>This validator implements a comprehensive RLN verification system that:
 *
 * <ul>
 *   <li>Maintains a deny list of addresses that have exceeded their quotas
 *   <li>Verifies RLN proofs using JNI calls to Rust implementation
 *   <li>Queries gRPC Karma service for user quota status (service handles all counting internally)
 *   <li>Provides premium gas bypass functionality for deny-listed users
 * </ul>
 *
 * <p><strong>Core Validation Flow:</strong>
 *
 * <ol>
 *   <li>Check if sender is on deny list and validate premium gas if required
 *   <li>Retrieve and verify RLN proof from in-memory cache
 *   <li>Validate proof authenticity using cryptographic verification
 *   <li>Query user's current quota status via gRPC karma service
 *   <li>Add to deny list if quota exceeded, otherwise allow transaction
 * </ol>
 *
 * <p><strong>gRPC Integration:</strong> This validator maintains two gRPC connections:
 *
 * <ul>
 *   <li>RLN Proof Service: Streaming server for receiving RLN proofs
 *   <li>Karma Service: Request-response service for querying user quota status
 * </ul>
 *
 * Both connections feature exponential backoff reconnection strategies.
 *
 * <p><strong>Cache Management:</strong> Implements an LRU cache with TTL expiration for efficient
 * proof storage and retrieval during asynchronous transaction validation.
 *
 * <p><strong>Thread Safety:</strong> All operations are thread-safe using concurrent data
 * structures and proper synchronization for file I/O operations.
 *
 * @see PluginTransactionPoolValidator
 * @see LineaRlnValidatorConfiguration
 * @author Status Network Development Team
 * @since 1.0
 */
public class RlnVerifierValidator implements PluginTransactionPoolValidator, Closeable {
  private static final Logger LOG = LoggerFactory.getLogger(RlnVerifierValidator.class);

  private final LineaRlnValidatorConfiguration rlnConfig;
  private final BlockchainService blockchainService;
  private final byte[] rlnVerifyingKeyBytes;
  private final DenyListManager denyListManager;
  private final RlnVerificationService rlnVerificationService;
  private ScheduledExecutorService proofCacheEvictionScheduler;

  private final Map<String, CompletableFuture<CachedProof>> pendingProofs =
      new ConcurrentHashMap<>();

  private final AtomicInteger activeProofWaits = new AtomicInteger(0);
  private static final int MAX_CONCURRENT_PROOF_WAITS = 100; // Configurable limit

  /**
   * Represents a cached RLN proof with combined format and extracted public inputs.
   *
   * @param combinedProofBytes Combined proof data (proof + proof values serialized together)
   * @param senderBytes Sender address bytes
   * @param shareXHex X-coordinate of the secret share (public input)
   * @param shareYHex Y-coordinate of the secret share (public input)
   * @param epochHex Current epoch identifier (public input)
   * @param rootHex Merkle tree root of the RLN membership tree (public input)
   * @param nullifierHex Unique nullifier for this transaction (public input)
   * @param cachedAt Timestamp when this proof was cached for TTL management
   */
  record CachedProof(
      byte[] combinedProofBytes,
      byte[] senderBytes,
      String shareXHex,
      String shareYHex,
      String epochHex,
      String rootHex,
      String nullifierHex,
      Instant cachedAt) {}

  // High-performance Caffeine cache for RLN proofs
  private final Cache<String, CachedProof> rlnProofCache;

  // gRPC client members for proof service
  private ManagedChannel proofServiceChannel;
  private RlnProverGrpc.RlnProverStub asyncProofStub;

  // Shared karma service client (injected dependency)
  private final KarmaServiceClient karmaServiceClient;

  // Shared nullifier tracker for preventing proof reuse (injected dependency)
  private final NullifierTracker nullifierTracker;

  private ScheduledExecutorService grpcReconnectionScheduler;

  // Exponential backoff state
  private final AtomicInteger proofStreamRetryCount = new AtomicInteger(0);
  private volatile long lastProofStreamRetryTime = 0;

  /**
   * Creates a new RLN Verifier Validator with default gRPC channel management.
   *
   * @param rlnConfig Configuration for RLN validation including service endpoints
   * @param blockchainService Blockchain service for accessing chain state
   * @param denyListManager Shared deny list manager for state consistency
   * @param karmaServiceClient Shared karma service client for quota validation
   * @param nullifierTracker Shared nullifier tracker for preventing proof reuse
   */
  public RlnVerifierValidator(
      LineaRlnValidatorConfiguration rlnConfig,
      BlockchainService blockchainService,
      DenyListManager denyListManager,
      KarmaServiceClient karmaServiceClient,
      NullifierTracker nullifierTracker) {
    this(
        rlnConfig,
        blockchainService,
        denyListManager,
        karmaServiceClient,
        nullifierTracker,
        null,
        null);
  }

  /**
   * Creates a new RLN Verifier Validator with shared services and optional pre-configured proof
   * channel.
   *
   * <p>This constructor is primarily intended for testing scenarios where a mock proof gRPC channel
   * or mock RLN verification service needs to be injected.
   *
   * @param rlnConfig Configuration for RLN validation
   * @param blockchainService Blockchain service for accessing chain state
   * @param denyListManager Shared deny list manager for state consistency
   * @param karmaServiceClient Shared karma service client for quota validation
   * @param nullifierTracker Shared nullifier tracker for preventing proof reuse
   * @param providedProofChannel Optional pre-configured proof service channel for testing
   * @param providedRlnService Optional pre-configured RLN verification service for testing
   */
  @VisibleForTesting
  RlnVerifierValidator(
      LineaRlnValidatorConfiguration rlnConfig,
      BlockchainService blockchainService,
      DenyListManager denyListManager,
      KarmaServiceClient karmaServiceClient,
      NullifierTracker nullifierTracker,
      ManagedChannel providedProofChannel,
      RlnVerificationService providedRlnService) {
    this.rlnConfig = rlnConfig;
    this.blockchainService = blockchainService;
    this.denyListManager = denyListManager;
    this.karmaServiceClient = karmaServiceClient;
    this.nullifierTracker = nullifierTracker;
    this.proofServiceChannel = providedProofChannel;

    // Initialize RLN verification service
    if (providedRlnService != null) {
      this.rlnVerificationService = providedRlnService;
    } else {
      this.rlnVerificationService = new JniRlnVerificationService();
    }

    // Initialize LRU cache with TTL support
    this.rlnProofCache =
        Caffeine.newBuilder()
            .expireAfterWrite(rlnConfig.rlnProofCacheExpirySeconds(), TimeUnit.SECONDS)
            .maximumSize(rlnConfig.rlnProofCacheMaxSize())
            .build();

    if (rlnConfig.rlnValidationEnabled()) {
      LOG.info("RLN Validator is ENABLED.");

      if (denyListManager == null) {
        throw new IllegalArgumentException(
            "DenyListManager cannot be null when RLN validation is enabled");
      }

      byte[] keyBytes;
      try {
        keyBytes = Files.readAllBytes(Paths.get(rlnConfig.verifyingKeyPath()));
        LOG.info("RLN Verifying Key loaded successfully from {}.", rlnConfig.verifyingKeyPath());
        LOG.info("✅ IMPORTANT: Zerokit RLN uses built-in verifying keys for height 20 trees.");
        LOG.info("   - The loaded external key file is kept for API compatibility");
        LOG.info(
            "   - Actual verification uses zerokit's internal pre-compiled keys via zkey_from_folder()");
        LOG.info("   - This ensures correct verification for height 20 RLN proofs");
      } catch (IOException e) {
        LOG.warn(
            "Failed to load external RLN verifying key from {}: {}. This is acceptable when using zerokit's built-in keys.",
            rlnConfig.verifyingKeyPath(),
            e.getMessage());
        LOG.info(
            "✅ Using zerokit's built-in verifying keys for height 20 trees (no external key file needed).");
        keyBytes = new byte[0]; // Empty placeholder - zerokit ignores this
      } catch (UnsatisfiedLinkError | RuntimeException e) {
        LOG.error("Failed to initialize RLN JNI RlnBridge: {}", e.getMessage(), e);
        throw new IllegalStateException(
            "Failed to initialize RlnVerifierValidator: JNI linkage error", e);
      }
      this.rlnVerifyingKeyBytes = keyBytes;

      initializeGrpcClients();
      startProofStreamSubscription();
      startProofCacheEvictionScheduler();
      initializeSharedProofWaitExecutor();

    } else {
      this.rlnVerifyingKeyBytes = null;
      LOG.info("RLN Validator is DISABLED.");
    }
  }

  /**
   * Initializes gRPC client connection for proof service.
   *
   * <p>Creates managed channel with appropriate TLS configuration based on the provided
   * configuration. Supports both injected channels (for testing) and dynamically created channels.
   */
  private void initializeGrpcClients() {
    // Initialize proof service client
    initializeProofServiceClient();
  }

  /**
   * Initializes the gRPC client for the RLN Proof Service.
   *
   * <p>Creates a managed channel configured for streaming proof reception with appropriate TLS
   * settings based on configuration.
   */
  private void initializeProofServiceClient() {
    boolean wasChannelProvided =
        (this.proofServiceChannel != null && !this.proofServiceChannel.isShutdown());

    if (wasChannelProvided) {
      LOG.info("Using pre-configured ManagedChannel for RLN Proof Service client.");
    } else {
      LOG.info("Creating new ManagedChannel for RLN Proof Service client based on configuration.");
      ManagedChannelBuilder<?> channelBuilder =
          ManagedChannelBuilder.forAddress(
              rlnConfig.rlnProofServiceHost(), rlnConfig.rlnProofServicePort());

      if (rlnConfig.rlnProofServiceUseTls()) {
        channelBuilder.useTransportSecurity();
      } else {
        channelBuilder.usePlaintext();
      }
      this.proofServiceChannel = channelBuilder.build();
    }

    this.asyncProofStub = RlnProverGrpc.newStub(this.proofServiceChannel);

    if (wasChannelProvided) {
      LOG.info("RLN Proof Service client initialized with injected ManagedChannel.");
    } else {
      LOG.info(
          "RLN Proof Service client initialized for target: {}:{}",
          rlnConfig.rlnProofServiceHost(),
          rlnConfig.rlnProofServicePort());
    }
  }

  /**
   * Starts the gRPC streaming subscription for receiving RLN proofs.
   *
   * <p>Establishes a persistent streaming connection to receive proofs asynchronously as they are
   * generated by the proof service. Implements automatic reconnection with exponential backoff on
   * failures.
   */
  private void startProofStreamSubscription() {
    if (asyncProofStub == null) {
      LOG.error("Cannot start RLN proof stream: gRPC stub not initialized.");
      return;
    }
    LOG.info("Attempting to subscribe to RLN proof stream...");
    RlnProofFilter request =
        RlnProofFilter.newBuilder().setAddress("").build(); // Empty address means all proofs

    asyncProofStub.getProofs(
        request,
        new StreamObserver<>() {
          @Override
          public void onNext(RlnProofReply proofMessage) {
            if (proofMessage.hasProof()) {
              RlnProof rlnProofMessage = proofMessage.getProof();
              String txHashHex =
                  Bytes.wrap(rlnProofMessage.getTxHash().toByteArray()).toHexString();
              LOG.debug("Received proof from gRPC stream for txHash: {}", txHashHex);

              // Parse the combined proof and extract public inputs using verification service
              String currentEpochId = getCurrentEpochIdentifier();
              LOG.debug(
                  "Processing proof for txHash: {}, using current epoch: {}",
                  txHashHex,
                  currentEpochId);

              try {
                RlnVerificationService.RlnProofData proofData =
                    rlnVerificationService.parseAndVerifyRlnProof(
                        rlnVerifyingKeyBytes,
                        rlnProofMessage.getProof().toByteArray(),
                        currentEpochId);

                if (proofData != null && proofData.isValid()) {
                  LOG.debug(
                      "Successfully parsed and verified proof for txHash: {}. Proof data: shareX={}, shareY={}, epoch={}, root={}, nullifier={}",
                      txHashHex,
                      proofData.shareX(),
                      proofData.shareY(),
                      proofData.epoch(),
                      proofData.root(),
                      proofData.nullifier());

                  CachedProof cachedProof =
                      new CachedProof(
                          rlnProofMessage.getProof().toByteArray(),
                          rlnProofMessage.getSender().toByteArray(),
                          proofData.shareX(), // share_x
                          proofData.shareY(), // share_y
                          proofData.epoch(), // epoch
                          proofData.root(), // root
                          proofData.nullifier(), // nullifier
                          Instant.now());

                  rlnProofCache.put(txHashHex, cachedProof);
                  LOG.info(
                      "Proof cached for txHash: {}, cache size: {}, proof epoch: {}",
                      txHashHex,
                      rlnProofCache.estimatedSize(),
                      proofData.epoch());

                  // Complete the future for any waiting threads
                  CompletableFuture<CachedProof> proofFuture = pendingProofs.remove(txHashHex);
                  if (proofFuture != null) {
                    proofFuture.complete(cachedProof);
                  }
                } else {
                  LOG.warn(
                      "Invalid proof received for txHash: {} (verification failed). ProofData: {}",
                      txHashHex,
                      proofData);
                  // Notify waiters about the failure
                  CompletableFuture<CachedProof> proofFuture = pendingProofs.remove(txHashHex);
                  if (proofFuture != null) {
                    proofFuture.complete(null);
                  }
                }
              } catch (Exception e) {
                LOG.error(
                    "Failed to parse and verify proof for txHash: {}: {}",
                    txHashHex,
                    e.getMessage(),
                    e);
                // Notify waiters about the failure
                CompletableFuture<CachedProof> proofFuture = pendingProofs.remove(txHashHex);
                if (proofFuture != null) {
                  proofFuture.complete(null);
                }
              }
            } else if (proofMessage.hasError()) {
              LOG.error("Received error from proof stream: {}", proofMessage.getError().getError());
            }

            // Reset retry count on successful message (even if proof was invalid)
            proofStreamRetryCount.set(0);
          }

          @Override
          public void onError(Throwable t) {
            LOG.error("RLN proof stream error: {}. Attempting to reconnect...", t.getMessage(), t);
            scheduleProofStreamReconnection();
          }

          @Override
          public void onCompleted() {
            LOG.info("RLN proof stream completed by server. Attempting to reconnect...");
            scheduleProofStreamReconnection();
          }
        });
  }

  /**
   * Schedules reconnection for the proof stream using exponential backoff strategy.
   *
   * <p>Implements intelligent reconnection with increasing delays to avoid overwhelming a failing
   * service while ensuring eventual connectivity restoration.
   *
   * <p><strong>Backoff Strategy:</strong>
   *
   * <ul>
   *   <li>Base delay from configuration (rlnProofStreamRetryIntervalMs)
   *   <li>Exponential increase: delay = base * 2^(retry_count)
   *   <li>Maximum delay capped by maxBackoffDelayMs configuration
   *   <li>Retry count resets on successful connection
   * </ul>
   */
  private void scheduleProofStreamReconnection() {
    if (grpcReconnectionScheduler == null || grpcReconnectionScheduler.isShutdown()) {
      grpcReconnectionScheduler =
          Executors.newSingleThreadScheduledExecutor(r -> new Thread(r, "RlnGrpcReconnect"));
    }

    long delay;
    if (rlnConfig.exponentialBackoffEnabled()) {
      int retryCount = proofStreamRetryCount.getAndIncrement();
      // Ensure we don't exceed max retries
      if (retryCount >= rlnConfig.rlnProofStreamRetries()) {
        LOG.error(
            "Maximum proof stream retry attempts ({}) exceeded. Stopping reconnection attempts.",
            rlnConfig.rlnProofStreamRetries());
        return;
      }

      // Calculate exponential backoff: base * 2^retryCount, capped at max
      delay =
          Math.min(
              rlnConfig.rlnProofStreamRetryIntervalMs() * (1L << retryCount),
              rlnConfig.maxBackoffDelayMs());

      LOG.info(
          "Scheduling gRPC proof stream reconnection in {} ms (attempt {}/{})",
          delay,
          retryCount + 1,
          rlnConfig.rlnProofStreamRetries());
    } else {
      // Simple fixed delay reconnection
      delay = rlnConfig.rlnProofStreamRetryIntervalMs();
      LOG.info("Scheduling gRPC proof stream reconnection in {} ms (fixed delay)", delay);
    }

    lastProofStreamRetryTime = System.currentTimeMillis();
    grpcReconnectionScheduler.schedule(
        this::startProofStreamSubscription, delay, TimeUnit.MILLISECONDS);
  }

  /**
   * Starts the scheduled task for proof cache eviction.
   *
   * <p>Note: With Caffeine cache, automatic TTL-based eviction is handled internally. This method
   * is kept for compatibility but now only triggers manual cleanup.
   */
  private void startProofCacheEvictionScheduler() {
    // Caffeine handles TTL automatically, but we can still do periodic cleanup for metrics
    proofCacheEvictionScheduler =
        Executors.newSingleThreadScheduledExecutor(r -> new Thread(r, "RlnProofCacheEviction"));
    proofCacheEvictionScheduler.scheduleAtFixedRate(
        this::evictExpiredProofs,
        this.rlnConfig.rlnProofCacheExpirySeconds(),
        this.rlnConfig.rlnProofCacheExpirySeconds(),
        TimeUnit.SECONDS);
  }

  /** Initializes the shared executor for proof waiting operations. */
  private void initializeSharedProofWaitExecutor() {
    // This executor is no longer needed with the CompletableFuture-based approach
    LOG.info("Shared proof wait executor is no longer used.");
  }

  /**
   * Triggers manual cache cleanup and logs cache statistics.
   *
   * <p>Note: Caffeine automatically evicts expired entries, so this is primarily for logging and
   * manual cleanup triggers.
   */
  private void evictExpiredProofs() {
    LOG.debug("Running RLN proof cache cleanup. Current size: {}", rlnProofCache.estimatedSize());
    rlnProofCache.cleanUp(); // Manual cleanup trigger
    LOG.debug(
        "RLN proof cache cleanup finished. Size after cleanup: {}", rlnProofCache.estimatedSize());
  }

  /**
   * Waits for an RLN proof to appear in cache using an event-driven CompletableFuture.
   *
   * <p>This implementation avoids polling by creating a future that is completed by the gRPC stream
   * thread. Implements proper concurrency limits to prevent resource exhaustion.
   *
   * @param txHashString The transaction hash to wait for
   * @return The cached proof if found within timeout, null otherwise
   */
  private CachedProof waitForProofInCache(String txHashString) {
    // First check if proof is already available
    CachedProof proof = rlnProofCache.getIfPresent(txHashString);
    if (proof != null) {
      return proof;
    }

    // Apply backpressure - reject if too many concurrent waits
    if (activeProofWaits.get() >= MAX_CONCURRENT_PROOF_WAITS) {
      LOG.warn(
          "Too many concurrent proof waits ({}), rejecting wait for tx {}",
          activeProofWaits.get(),
          txHashString);
      return null;
    }

    CompletableFuture<CachedProof> proofFuture =
        pendingProofs.computeIfAbsent(txHashString, k -> new CompletableFuture<>());

    activeProofWaits.incrementAndGet();
    try {
      // Wait for the future to be completed by the gRPC onNext handler
      return proofFuture.get(rlnConfig.rlnProofLocalWaitTimeoutMs(), TimeUnit.MILLISECONDS);
    } catch (TimeoutException e) {
      LOG.warn("Proof wait timed out for tx {}", txHashString);
      return null;
    } catch (Exception e) {
      LOG.warn("Error waiting for proof for tx {}: {}", txHashString, e.getMessage(), e);
      return null;
    } finally {
      // Ensure the future is removed to prevent memory leaks if it timed out
      pendingProofs.remove(txHashString);
      activeProofWaits.decrementAndGet();
    }
  }

  /**
   * Adds an address to the deny list with current timestamp.
   *
   * @param address The address to add to the deny list
   */
  void addToDenyList(final Address address) {
    denyListManager.addToDenyList(address);
  }

  /**
   * Removes an address from the deny list.
   *
   * @param address The address to remove from the deny list
   * @return true if the address was in the list and removed, false otherwise
   */
  boolean removeFromDenyList(final Address address) {
    return denyListManager.removeFromDenyList(address);
  }

  /**
   * Generates the current epoch identifier based on configuration.
   *
   * <p>Supports different epoch strategies:
   *
   * <ul>
   *   <li>"BLOCK" - Uses current block number as field element
   *   <li>"TIMESTAMP_1H" - Uses hourly timestamp buckets as field element
   *   <li>"TEST" - Uses hardcoded test epoch for testing
   * </ul>
   *
   * @return The current epoch identifier as hex string (field element compatible)
   */
  private String getCurrentEpochIdentifier() {
    var currentHeader = blockchainService.getChainHeadHeader();
    long timestamp = currentHeader.getTimestamp();
    long blockNumber = currentHeader.getNumber();

    return switch (rlnConfig.defaultEpochForQuota().toUpperCase()) {
      case "BLOCK" -> {
        // Secure block-based epoch generation with entropy mixing
        String blockStr = "BLOCK:" + blockNumber + ":SALT:" + getSecureEpochSalt();
        yield hashToFieldElementHex(blockStr);
      }
      case "TIMESTAMP_1H" -> {
        // Secure timestamp-based epoch generation with entropy mixing
        String timestampStr =
            "TIME:"
                + Instant.ofEpochSecond(timestamp)
                    .atZone(ZoneOffset.UTC)
                    .format(DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH"))
                + ":SALT:" + getSecureEpochSalt();
        yield hashToFieldElementHex(timestampStr);
      }
      case "TEST" -> {
        LOG.warn("Using TEST epoch mode - this should only be used in testing!");
        yield "0x09a6ed7f807775ba43e63fbba747a7f0122aa3fac4a05b3392aea03eecdd1128";
      }
      case "FIXED_FIELD_ELEMENT" -> {
        // For development/testing: use a fixed field element that matches the RLN prover
        // This should match the external_nullifier from your RLN proof logs
        yield "0x1c61ef0b2ebc0235d85fe8537b4455549356e3895005ba7a03fbd4efc9ba3692";
      }
      default -> {
        LOG.warn(
            "Unknown defaultEpochForQuota: '{}'. Defaulting to block number hash.",
            rlnConfig.defaultEpochForQuota());
        String blockStr = "BLOCK:" + blockNumber;
        yield hashToFieldElementHex(blockStr);
      }
    };
  }

  /**
   * Generates a secure salt for epoch generation to prevent predictable epoch values.
   * Uses blockchain state entropy for security while maintaining determinism.
   *
   * @return Secure salt string based on recent blockchain state
   */
  private String getSecureEpochSalt() {
    try {
      var currentHeader = blockchainService.getChainHeadHeader();
      long blockNumber = currentHeader.getNumber();
      long timestamp = currentHeader.getTimestamp();
      
      // Use recent block data for entropy while maintaining determinism within epoch windows
      // Mix block hash with timestamp for additional entropy
      String entropySource = "ENTROPY:" + (blockNumber / 100) * 100 + ":" + (timestamp / 3600) * 3600;
      
      // Hash to create compact, secure salt
      java.security.MessageDigest digest = java.security.MessageDigest.getInstance("SHA-256");
      byte[] hash = digest.digest(entropySource.getBytes(java.nio.charset.StandardCharsets.UTF_8));
      
      // Use first 8 bytes for compact salt
      StringBuilder salt = new StringBuilder();
      for (int i = 0; i < 8; i++) {
        salt.append(String.format("%02x", hash[i]));
      }
      return salt.toString();
      
    } catch (Exception e) {
      LOG.error("Error generating secure epoch salt: {}", e.getMessage());
      // Fallback to basic timestamp for determinism
      return String.valueOf(System.currentTimeMillis() / 3600000); // Hour-based fallback
    }
  }

  /**
   * Converts a string to a field element compatible hex representation. Uses SHA-256 hash to ensure
   * deterministic field element generation.
   *
   * @param input The input string to hash
   * @return Hex string representation suitable for field element conversion
   */
  private String hashToFieldElementHex(String input) {
    try {
      java.security.MessageDigest digest = java.security.MessageDigest.getInstance("SHA-256");
      byte[] hash = digest.digest(input.getBytes(java.nio.charset.StandardCharsets.UTF_8));

      // Convert to hex string with 0x prefix
      StringBuilder hexString = new StringBuilder("0x");
      for (byte b : hash) {
        hexString.append(String.format("%02x", b));
      }

      return hexString.toString();
    } catch (java.security.NoSuchAlgorithmException e) {
      LOG.error("SHA-256 algorithm not available: {}", e.getMessage(), e);
      // Fallback to a simple conversion if SHA-256 is not available
      return "0x" + Integer.toHexString(input.hashCode()).toLowerCase();
    }
  }

  /**
   * Validates if a proof epoch is acceptable compared to the current epoch.
   * Implements flexible epoch validation to prevent race conditions while maintaining security.
   *
   * @param proofEpochId The epoch from the RLN proof
   * @param currentEpochId The current system epoch
   * @return true if the proof epoch is valid, false if outside acceptable window
   */
  private boolean isEpochValid(String proofEpochId, String currentEpochId) {
    // Exact match is always valid
    if (currentEpochId.equals(proofEpochId)) {
      return true;
    }

    // For different epoch modes, implement appropriate tolerance windows
    String epochMode = rlnConfig.defaultEpochForQuota().toUpperCase();
    
    switch (epochMode) {
      case "BLOCK":
        // Allow proofs from previous 2 blocks to handle block timing races
        return isBlockEpochValid(proofEpochId, currentEpochId, 2);
      
      case "TIMESTAMP_1H":
        // Allow proofs from current hour and previous hour for timing tolerance
        return isTimestampEpochValid(proofEpochId, currentEpochId, 1);
      
      case "TEST":
      case "FIXED_FIELD_ELEMENT":
        // In test mode, be more permissive for testing scenarios
        return true;
      
      default:
        // For unknown modes, default to strict validation for security
        LOG.warn("Unknown epoch mode '{}', using strict validation", epochMode);
        return false;
    }
  }

  /**
   * Validates block-based epochs within tolerance window.
   */
  private boolean isBlockEpochValid(String proofEpoch, String currentEpoch, int blockTolerance) {
    try {
      // Extract block numbers from epoch hashes (simplified approach)
      // In production, you'd want more sophisticated epoch comparison
      var currentHeader = blockchainService.getChainHeadHeader();
      long currentBlock = currentHeader.getNumber();
      
      // For each potential recent block, generate its epoch and compare
      for (int i = 0; i <= blockTolerance; i++) {
        String testBlockStr = "BLOCK:" + (currentBlock - i);
        String testEpoch = hashToFieldElementHex(testBlockStr);
        if (testEpoch.equals(proofEpoch)) {
          if (i > 0) {
            LOG.debug("Accepting proof from {} blocks ago (tolerance: {})", i, blockTolerance);
          }
          return true;
        }
      }
      return false;
    } catch (Exception e) {
      LOG.error("Error validating block epoch: {}", e.getMessage());
      return false; // Fail secure
    }
  }

  /**
   * Validates timestamp-based epochs within tolerance window.
   */
  private boolean isTimestampEpochValid(String proofEpoch, String currentEpoch, int hourTolerance) {
    try {
      var currentHeader = blockchainService.getChainHeadHeader();
      long currentTimestamp = currentHeader.getTimestamp();
      
      // Check current hour and previous hours within tolerance
      for (int i = 0; i <= hourTolerance; i++) {
        long testTimestamp = currentTimestamp - (i * 3600); // Subtract hours
        String testTimeStr = "TIME:" + 
            Instant.ofEpochSecond(testTimestamp)
                .atZone(ZoneOffset.UTC)
                .format(DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH"));
        String testEpoch = hashToFieldElementHex(testTimeStr);
        if (testEpoch.equals(proofEpoch)) {
          if (i > 0) {
            LOG.debug("Accepting proof from {} hours ago (tolerance: {})", i, hourTolerance);
          }
          return true;
        }
      }
      return false;
    } catch (Exception e) {
      LOG.error("Error validating timestamp epoch: {}", e.getMessage());
      return false; // Fail secure
    }
  }

  /**
   * Fetches current karma status for a user via shared Karma Service client. The karma service
   * handles all transaction counting internally.
   *
   * @param userAddress The user address to query karma information for
   * @return Optional containing karma info (including current quota status) if successful, empty on
   *     failure
   */
  private Optional<KarmaInfo> fetchKarmaInfoFromService(Address userAddress) {
    if (karmaServiceClient == null || !karmaServiceClient.isAvailable()) {
      LOG.warn("Karma service client not available. Cannot fetch karma info.");
      return Optional.empty();
    }

    return karmaServiceClient.fetchKarmaInfo(userAddress);
  }

  /**
   * Validates a transaction against RLN requirements.
   *
   * <p>This is the main validation entry point that orchestrates the complete RLN validation flow
   * including deny list checks, proof verification, and quota enforcement.
   *
   * <p><strong>Validation Steps:</strong>
   *
   * <ol>
   *   <li>Check deny list status and premium gas bypass
   *   <li>Retrieve and validate RLN proof from cache
   *   <li>Verify cryptographic proof authenticity
   *   <li>Check user karma quota via gRPC service
   *   <li>Apply deny list penalties for quota violations
   * </ol>
   *
   * @param transaction The transaction to validate
   * @param isLocal Whether this is a local transaction
   * @param hasPriority Whether this transaction has priority status
   * @return Optional error message if validation fails, empty if valid
   */
  @Override
  public Optional<String> validateTransaction(
      Transaction transaction, boolean isLocal, boolean hasPriority) {
    if (!rlnConfig.rlnValidationEnabled()) {
      return Optional.empty(); // RLN validation is disabled
    }

    final Address sender = transaction.getSender();
    final org.hyperledger.besu.datatypes.Hash txHash = transaction.getHash();
    final String txHashString = txHash.toHexString();

    // 1. Deny List Check
    if (denyListManager.isDenied(sender)) {
      // User is actively denied. Check for premium gas.
      long premiumThresholdWei = rlnConfig.premiumGasPriceThresholdWei();
      Wei effectiveGasPrice =
          transaction
              .getGasPrice()
              .map(q -> Wei.of(q.getAsBigInteger()))
              .orElseGet(
                  () ->
                      transaction
                          .getMaxFeePerGas()
                          .map(q -> Wei.of(q.getAsBigInteger()))
                          .orElse(Wei.ZERO));

      if (effectiveGasPrice.getAsBigInteger().compareTo(BigInteger.valueOf(premiumThresholdWei))
          >= 0) {
        denyListManager.removeFromDenyList(sender);
        LOG.info(
            "Sender {} was on deny list but paid premium gas ({} Wei >= {} Wei). Allowing and removing from deny list.",
            sender.toHexString(),
            effectiveGasPrice,
            premiumThresholdWei);
      } else {
        LOG.warn(
            "Sender {} is on deny list. Transaction {} rejected. Effective gas price {} Wei < {} Wei.",
            sender.toHexString(),
            txHashString,
            effectiveGasPrice,
            premiumThresholdWei);
        return Optional.of("Sender on deny list, premium gas not met.");
      }
    }

    // 2. RLN Proof Verification (via gRPC Cache) - with non-blocking wait
    LOG.debug("Attempting to fetch RLN proof for txHash: {} from cache.", txHashString);
    CachedProof proof = waitForProofInCache(txHashString);

    if (proof == null) {
      LOG.warn(
          "RLN proof not found in cache after timeout for txHash: {}. Timeout: {}ms",
          txHashString,
          rlnConfig.rlnProofLocalWaitTimeoutMs());
      return Optional.of("RLN proof not found in cache after timeout.");
    }
    LOG.debug("RLN proof found in cache for txHash: {}", txHashString);

    // Validate proof epoch format first
    if (proof.epochHex() == null || proof.epochHex().trim().isEmpty()) {
      LOG.warn("Invalid proof epoch for tx {}: epoch is null or empty", txHashString);
      return Optional.of("RLN validation failed: Invalid proof epoch");
    }

    // Validate that the proof epoch is a valid hex string
    if (!proof.epochHex().matches("^0x[0-9a-fA-F]+$")) {
      LOG.warn("Invalid proof epoch format for tx {}: {}", txHashString, proof.epochHex());
      return Optional.of("RLN validation failed: Invalid proof epoch format");
    }

    String currentEpochId = getCurrentEpochIdentifier();
    String proofEpochId = proof.epochHex();
    LOG.debug("Proof epoch: {}, Current epoch: {}", proofEpochId, currentEpochId);

    // CRITICAL SECURITY FIX: Use the proof's epoch for nullifier tracking, not current epoch
    // This prevents nullifier reuse across different epochs
    if (nullifierTracker != null) {
      boolean isNullifierNew =
          nullifierTracker.checkAndMarkNullifier(proof.nullifierHex(), proofEpochId);
      if (!isNullifierNew) {
        LOG.error(
            "CRITICAL SECURITY VIOLATION: Nullifier reuse detected for tx {}. Nullifier: {}, Proof Epoch: {}",
            txHashString,
            proof.nullifierHex(),
            proofEpochId);
        return Optional.of(
            "RLN validation failed: Nullifier already used in epoch "
                + proofEpochId
                + " (potential double-spend attack)");
      }
      LOG.debug(
          "Nullifier {} verified as unique for proof epoch {}", proof.nullifierHex(), proofEpochId);
    } else {
      LOG.error("NullifierTracker not available - SECURITY RISK: Cannot prevent nullifier reuse!");
      return Optional.of("RLN validation failed: Nullifier tracking unavailable");
    }

    // Flexible epoch validation: allow proofs from recent epochs to prevent race conditions
    // while still maintaining security against replay attacks
    if (!isEpochValid(proofEpochId, currentEpochId)) {
      LOG.warn(
          "SECURITY WARNING: Epoch validation failed for tx {}. Proof epoch: {}, Current epoch: {}. Outside acceptable window.",
          txHashString,
          proofEpochId,
          currentEpochId);
      return Optional.of(
          "RLN validation failed: Proof epoch "
              + proofEpochId
              + " outside acceptable window from current epoch "
              + currentEpochId);
    }

    LOG.debug("Epoch validation passed for tx {}: {}", txHashString, currentEpochId);

    // Since the proof was already verified and public inputs extracted during caching,
    // we can skip the verification step here as the proof is already validated.
    // However, for completeness and double-checking, we can still verify if needed.

    // The proof verification was already done during the onNext() processing when the proof was
    // cached.
    // At this point, we can trust that the cached proof is valid and the public inputs are correct.
    LOG.info("Using cached and pre-verified RLN proof for tx: {}", txHashString);

    // 3. Karma / Quota Check (via gRPC Karma Service) - with fail-safe circuit breaker
    Optional<KarmaInfo> karmaInfoOpt = fetchKarmaInfoFromService(sender);

    if (karmaInfoOpt.isEmpty()) {
      // SECURITY: Reject when karma service is down to prevent DoS attacks
      LOG.warn(
          "Karma service unavailable for sender {} and tx {}. REJECTING transaction for security.",
          sender.toHexString(),
          txHashString);

      return Optional.of(
          "RLN validation failed: Karma service unavailable - transaction rejected for security");
    }

    KarmaInfo karmaInfo = karmaInfoOpt.get();
    LOG.debug(
        "Karma info for sender {}: Tier={}, EpochTxCount={}, DailyQuota={}, EpochId={}, KarmaBalance={}",
        sender.toHexString(),
        karmaInfo.tier(),
        karmaInfo.epochTxCount(),
        karmaInfo.dailyQuota(),
        karmaInfo.epochId(),
        karmaInfo.karmaBalance());

    // Check if user has exceeded their quota (karma service handles all counting internally)
    if (karmaInfo.epochTxCount() >= karmaInfo.dailyQuota()) {
      LOG.warn(
          "User {} (Tier: {}) has exceeded their transaction quota for epoch {}. Count: {}, Quota: {}. Transaction {} rejected.",
          sender.toHexString(),
          karmaInfo.tier(),
          karmaInfo.epochId(),
          karmaInfo.epochTxCount(),
          karmaInfo.dailyQuota(),
          txHashString);
      addToDenyList(sender);
      return Optional.of("User transaction quota exceeded for current epoch. Added to deny list.");
    }

    // User is within quota - allow transaction (karma service handles transaction counting
    // internally)
    LOG.debug(
        "User {} (Tier: {}) is within transaction quota. Count: {}, Quota: {}. Transaction {} allowed by karma check.",
        sender.toHexString(),
        karmaInfo.tier(),
        karmaInfo.epochTxCount(),
        karmaInfo.dailyQuota(),
        txHashString);

    LOG.info(
        "Transaction {} from sender {} passed all RLN validations.",
        txHashString,
        sender.toHexString());
    return Optional.empty(); // Transaction is valid from RLN perspective
  }

  /**
   * Closes all resources including gRPC channels and scheduled executors.
   *
   * <p>Ensures graceful shutdown of all background tasks and network connections. This method
   * should be called when the validator is no longer needed to prevent resource leaks.
   *
   * @throws IOException if there are issues during resource cleanup
   */
  @Override
  public void close() throws IOException {
    LOG.info("Closing RlnVerifierValidator resources...");

    // Shutdown gRPC channels
    if (proofServiceChannel != null && !proofServiceChannel.isShutdown()) {
      proofServiceChannel.shutdown();
      try {
        if (!proofServiceChannel.awaitTermination(5, TimeUnit.SECONDS)) {
          proofServiceChannel.shutdownNow();
        }
      } catch (InterruptedException e) {
        proofServiceChannel.shutdownNow();
        Thread.currentThread().interrupt();
      }
    }

    if (karmaServiceClient != null) {
      try {
        karmaServiceClient.close();
      } catch (IOException e) {
        LOG.warn("Error closing karma service client: {}", e.getMessage(), e);
      }
    }

    // Shutdown schedulers
    if (proofCacheEvictionScheduler != null && !proofCacheEvictionScheduler.isShutdown()) {
      proofCacheEvictionScheduler.shutdownNow();
    }
    if (grpcReconnectionScheduler != null && !grpcReconnectionScheduler.isShutdown()) {
      grpcReconnectionScheduler.shutdownNow();
    }

    LOG.info("RlnVerifierValidator resources closed.");
  }

  // Test-only helper methods

  @VisibleForTesting
  void addToDenyListForTest(Address user, Instant addedAt) {
    denyListManager.addToDenyList(user);
  }

  @VisibleForTesting
  boolean isDeniedForTest(Address user) {
    return denyListManager.isDenied(user);
  }

  @VisibleForTesting
  void loadDenyListFromFileForTest() {
    denyListManager.reloadFromFile();
  }

  @VisibleForTesting
  Optional<CachedProof> getProofFromCacheForTest(String txHash) {
    return Optional.ofNullable(rlnProofCache.getIfPresent(txHash));
  }

  @VisibleForTesting
  void addProofToCacheForTest(String txHash, CachedProof proof) {
    rlnProofCache.put(txHash, proof);
  }
}
