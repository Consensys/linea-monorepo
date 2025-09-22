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

import com.google.common.annotations.VisibleForTesting;
import com.google.protobuf.ByteString;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import java.math.BigInteger;
import java.io.Closeable;
import java.io.IOException;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import net.consensys.linea.config.LineaRlnValidatorConfiguration;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient;
import net.consensys.linea.zktracer.LineCountingTracer;
import net.consensys.linea.zktracer.ZkCounter;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient.KarmaInfo;
import net.vac.prover.Address;
import net.vac.prover.RlnProverGrpc;
import net.vac.prover.SendTransactionReply;
import net.vac.prover.SendTransactionRequest;
import net.vac.prover.U256;
import net.vac.prover.Wei;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * RLN Prover Forwarder Validator for sending transactions to RLN prover service.
 *
 * <p>This validator implements a transaction forwarding system that:
 *
 * <ul>
 *   <li>Forwards local transactions to the RLN prover service for proof generation
 *   <li>Only validates transactions when running in RPC node mode (not sequencer mode)
 *   <li>Gracefully handles gRPC failures by falling back to default validation
 *   <li>Provides transaction statistics for monitoring and debugging
 * </ul>
 *
 * <p><strong>Core Validation Flow:</strong>
 *
 * <ol>
 *   <li>Check if transaction is local (only local transactions are forwarded)
 *   <li>Send transaction data to RLN prover service via gRPC
 *   <li>Wait for response from prover service
 *   <li>Allow or reject transaction based on prover response
 *   <li>Fall back to allowing transaction if gRPC service fails
 * </ol>
 *
 * <p><strong>gRPC Integration:</strong> This validator maintains a gRPC connection to the RLN
 * Prover Service for sending transaction data and receiving validation responses.
 *
 * <p><strong>Thread Safety:</strong> All operations are thread-safe using atomic counters for
 * statistics tracking.
 *
 * @see PluginTransactionPoolValidator
 * @see LineaRlnValidatorConfiguration
 * @author Status Network Development Team
 * @since 1.0
 */
public class RlnProverForwarderValidator implements PluginTransactionPoolValidator, Closeable {
  private static final Logger LOG = LoggerFactory.getLogger(RlnProverForwarderValidator.class);

  private final LineaRlnValidatorConfiguration rlnConfig;
  private final boolean enabled;

  // Statistics tracking
  private final AtomicInteger validationCallCount = new AtomicInteger(0);
  private final AtomicInteger localTransactionCount = new AtomicInteger(0);
  private final AtomicInteger peerTransactionCount = new AtomicInteger(0);
  private final AtomicInteger grpcSuccessCount = new AtomicInteger(0);
  private final AtomicInteger grpcFailureCount = new AtomicInteger(0);
  private final AtomicInteger karmaBypassCount = new AtomicInteger(0);

  // gRPC client components
  private final ManagedChannel channel;
  private final RlnProverGrpc.RlnProverBlockingStub blockingStub;

  // Karma service for gasless validation
  private final KarmaServiceClient karmaServiceClient;

  // Simulation dependencies for estimating gas used
  private final TransactionSimulationService transactionSimulationService;
  private final BlockchainService blockchainService;
  private final LineaTracerConfiguration tracerConfiguration;
  private final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration;

  /**
   * Creates a new RLN Prover Forwarder Validator with default gRPC channel management.
   *
   * @param rlnConfig Configuration for RLN validation including prover service endpoint
   * @param enabled Whether the validator is enabled (should be false in sequencer mode)
   * @param karmaServiceClient Service for checking karma eligibility for gasless transactions
   */
  public RlnProverForwarderValidator(
      LineaRlnValidatorConfiguration rlnConfig,
      boolean enabled,
      KarmaServiceClient karmaServiceClient,
      TransactionSimulationService transactionSimulationService,
      BlockchainService blockchainService,
      LineaTracerConfiguration tracerConfiguration,
      LineaL1L2BridgeSharedConfiguration l1L2BridgeSharedConfiguration) {
    this(rlnConfig,
        enabled,
        karmaServiceClient,
        transactionSimulationService,
        blockchainService,
        tracerConfiguration,
        l1L2BridgeSharedConfiguration,
        null);
  }

  /**
   * Backward-compatible constructor used by existing tests. New dependencies default to null.
   */
  public RlnProverForwarderValidator(
      LineaRlnValidatorConfiguration rlnConfig,
      boolean enabled,
      KarmaServiceClient karmaServiceClient) {
    this(rlnConfig, enabled, karmaServiceClient, null, null, null, null, null);
  }

  /**
   * Creates a new RLN Prover Forwarder Validator with default gRPC channel management (legacy
   * constructor for backward compatibility).
   *
   * @param rlnConfig Configuration for RLN validation including prover service endpoint
   * @param enabled Whether the validator is enabled (should be false in sequencer mode)
   */
  public RlnProverForwarderValidator(LineaRlnValidatorConfiguration rlnConfig, boolean enabled) {
    this(rlnConfig, enabled, null, null, null, null, null, null);
  }

  /**
   * Creates a new RLN Prover Forwarder Validator with optional pre-configured channel.
   *
   * <p>This constructor is primarily intended for testing scenarios where a mock gRPC channel needs
   * to be injected.
   *
   * @param rlnConfig Configuration for RLN validation
   * @param enabled Whether the validator is enabled
   * @param karmaServiceClient Service for checking karma eligibility for gasless transactions
   * @param providedChannel Optional pre-configured gRPC channel for testing
   */
  @VisibleForTesting
  RlnProverForwarderValidator(
      LineaRlnValidatorConfiguration rlnConfig,
      boolean enabled,
      KarmaServiceClient karmaServiceClient,
      TransactionSimulationService transactionSimulationService,
      BlockchainService blockchainService,
      LineaTracerConfiguration tracerConfiguration,
      LineaL1L2BridgeSharedConfiguration l1L2BridgeSharedConfiguration,
      ManagedChannel providedChannel) {
    this.rlnConfig = rlnConfig;
    this.enabled = enabled;
    this.karmaServiceClient = karmaServiceClient;
    this.transactionSimulationService = transactionSimulationService;
    this.blockchainService = blockchainService;
    this.tracerConfiguration = tracerConfiguration;
    this.l1L2BridgeConfiguration = l1L2BridgeSharedConfiguration;

    if (enabled) {
      if (providedChannel != null) {
        this.channel = providedChannel;
        LOG.info("Using pre-configured ManagedChannel for RLN Prover Forwarder.");
      } else {
        this.channel = createGrpcChannel();
        LOG.info(
            "RLN Prover Forwarder initialized for endpoint: {}:{}",
            rlnConfig.rlnProofServiceHost(),
            rlnConfig.rlnProofServicePort());
      }
      this.blockingStub = RlnProverGrpc.newBlockingStub(this.channel);
      LOG.info("RLN Prover Forwarder Validator is ENABLED.");
    } else {
      this.channel = null;
      this.blockingStub = null;
      LOG.info("RLN Prover Forwarder Validator is DISABLED (sequencer mode).");
    }
  }

  /**
   * Creates a gRPC channel based on configuration.
   *
   * @return The configured ManagedChannel
   */
  private ManagedChannel createGrpcChannel() {
    ManagedChannelBuilder<?> channelBuilder =
        ManagedChannelBuilder.forAddress(
            rlnConfig.rlnProofServiceHost(), rlnConfig.rlnProofServicePort());

    if (rlnConfig.rlnProofServiceUseTls()) {
      channelBuilder.useTransportSecurity();
    } else {
      channelBuilder.usePlaintext();
    }

    return channelBuilder.build();
  }

  /**
   * Validates a transaction by forwarding it to the RLN prover service.
   *
   * <p>This is the main validation entry point that forwards local transactions to the RLN prover
   * service for proof generation.
   *
   * <p><strong>Validation Logic:</strong>
   *
   * <ol>
   *   <li>Skip validation if validator is disabled (sequencer mode)
   *   <li>Only forward local transactions, allow peer transactions without forwarding
   *   <li>Send transaction data to RLN prover service via gRPC
   *   <li>Return validation result based on prover response
   *   <li>Fall back to allowing transaction if gRPC fails
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

    int callCount = validationCallCount.incrementAndGet();

    LOG.debug("*** RLN PROVER FORWARDER VALIDATION #{} ***", callCount);
    LOG.debug("Transaction Hash: {}", transaction.getHash().toHexString());
    LOG.debug("Transaction Sender: {}", transaction.getSender().toHexString());
    LOG.debug("Is Local: {}", isLocal);
    LOG.debug("Has Priority: {}", hasPriority);
    LOG.debug("Validator Enabled: {}", enabled);

    // Skip validation if disabled (sequencer mode)
    if (!enabled) {
      LOG.debug("RLN Prover Forwarder is disabled, skipping validation");
      return Optional.empty();
    }

    // Only validate local transactions via gRPC
    if (!isLocal) {
      peerTransactionCount.incrementAndGet();
      LOG.debug("Skipping gRPC forwarding for peer transaction");
      return Optional.empty(); // Accept peer transactions without gRPC forwarding
    }

    localTransactionCount.incrementAndGet();
    LOG.debug(
        "Forwarding local transaction to RLN prover: {} from {} (legacyGasPrice={}, maxFee={}, maxPrio={}, chainId={})",
        transaction.getHash().toHexString(),
        transaction.getSender().toHexString(),
        transaction.getGasPrice().map(Object::toString).orElse("-"),
        transaction.getMaxFeePerGas().map(Object::toString).orElse("-"),
        transaction.getMaxPriorityFeePerGas().map(Object::toString).orElse("-"),
        transaction.getChainId().map(Object::toString).orElse("-"));

    // GASLESS KARMA CHECK: Check if user is eligible for gasless transactions
    if (karmaServiceClient != null && karmaServiceClient.isAvailable()) {
      try {
        Optional<KarmaInfo> karmaInfoOpt =
            karmaServiceClient.fetchKarmaInfo(transaction.getSender());

        if (karmaInfoOpt.isPresent()) {
          KarmaInfo karmaInfo = karmaInfoOpt.get();
          boolean hasQuotaAvailable = karmaInfo.epochTxCount() < karmaInfo.dailyQuota();
          boolean isEligibleTier =
              !"Unknown".equals(karmaInfo.tier()) && karmaInfo.dailyQuota() > 0;

          LOG.debug(
              "Karma check for sender {}: Tier={}, TxCount={}, Quota={}, HasQuota={}, IsEligibleTier={}",
              transaction.getSender().toHexString(),
              karmaInfo.tier(),
              karmaInfo.epochTxCount(),
              karmaInfo.dailyQuota(),
              hasQuotaAvailable,
              isEligibleTier);

          if (hasQuotaAvailable && isEligibleTier) {
            // User has available karma quota - prioritize for gasless but still validate through
            // prover
            karmaBypassCount.incrementAndGet();
            LOG.info(
                "⚡ GASLESS PRIORITY: Sender {} has tier '{}' with available quota ({}/{}). Prioritizing gasless transaction {} for prover validation",
                transaction.getSender().toHexString(),
                karmaInfo.tier(),
                karmaInfo.epochTxCount(),
                karmaInfo.dailyQuota(),
                transaction.getHash().toHexString());
            // Continue with prover validation but with priority handling
          } else {
            LOG.debug(
                "Sender {} does not qualify for gasless bypass. HasQuota={}, IsEligibleTier={}",
                transaction.getSender().toHexString(),
                hasQuotaAvailable,
                isEligibleTier);
          }
        } else {
          LOG.debug("No karma info found for sender {}", transaction.getSender().toHexString());
        }
      } catch (Exception e) {
        LOG.warn(
            "Failed to check karma for sender {}: {}",
            transaction.getSender().toHexString(),
            e.getMessage());
      }
    }

    // Continue with normal RLN prover forwarding if not eligible for gasless bypass
    try {
      SendTransactionRequest.Builder requestBuilder = SendTransactionRequest.newBuilder();

      // Set transaction hash
      requestBuilder.setTransactionHash(ByteString.copyFrom(transaction.getHash().toArrayUnsafe()));

      // Set sender address
      requestBuilder.setSender(
          Address.newBuilder()
              .setValue(ByteString.copyFrom(transaction.getSender().toArrayUnsafe()))
              .build());

      // Set gas price if available
      transaction
          .getGasPrice()
          .ifPresent(
              gasPrice ->
                  requestBuilder.setGasPrice(
                      Wei.newBuilder()
                          .setValue(ByteString.copyFrom(gasPrice.getAsBigInteger().toByteArray()))
                          .build()));

      // Set chain ID if available
      transaction
          .getChainId()
          .ifPresent(
              chainId ->
                  requestBuilder.setChainId(
                      U256.newBuilder()
                          .setValue(ByteString.copyFrom(chainId.toByteArray()))
                          .build()));

      // Provide an estimated gas units value. As an initial implementation,
      // simulate execution to estimate gas used when possible; fallback to tx gas limit.
      long estimatedGasUsed = estimateGasUsed(transaction);
      LOG.debug(
          "Estimated gas used for tx {}: {}",
          transaction.getHash().toHexString(),
          estimatedGasUsed);
      requestBuilder.setEstimatedGasUsed(estimatedGasUsed);

      SendTransactionRequest request = requestBuilder.build();

      LOG.debug(
          "Sending transaction to RLN prover: txHash={}, sender={}, chainId={}",
          transaction.getHash().toHexString(),
          transaction.getSender().toHexString(),
          transaction.getChainId().map(Object::toString).orElse("-"));
      SendTransactionReply reply = blockingStub.sendTransaction(request);

      if (reply.getResult()) {
        grpcSuccessCount.incrementAndGet();
        LOG.debug("RLN prover accepted transaction {}", transaction.getHash());
        return Optional.empty(); // Transaction is valid
      } else {
        grpcFailureCount.incrementAndGet();
        LOG.warn("RLN prover rejected transaction {}", transaction.getHash());
        return Optional.of("RLN prover rejected transaction");
      }

    } catch (final Exception e) {
      grpcFailureCount.incrementAndGet();
      LOG.warn(
          "gRPC forwarding failed for transaction {}, falling back to default validation: {}",
          transaction.getHash(),
          e.getMessage());
      // Graceful fallback: accept the transaction if gRPC fails
      return Optional.empty();
    }
  }

  private LineCountingTracer createLineCountingTracer(
      final ProcessableBlockHeader pendingBlockHeader, final BigInteger chainId) {
    var lineCountingTracer =
        tracerConfiguration != null && tracerConfiguration.isLimitless()
            ? new ZkCounter(l1L2BridgeConfiguration)
            : new ZkTracer(net.consensys.linea.zktracer.Fork.LONDON, l1L2BridgeConfiguration, chainId);
    lineCountingTracer.traceStartConflation(1L);
    lineCountingTracer.traceStartBlock(pendingBlockHeader, pendingBlockHeader.getCoinbase());
    return lineCountingTracer;
  }

  private long estimateGasUsed(final Transaction transaction) {
    try {
      // Fast-path: simple ETH transfer with empty calldata
      if (transaction.getTo().isPresent()
          && transaction.getPayload().isEmpty()
          && transaction.getValue().getAsBigInteger().signum() > 0) {
        return 21_000L;
      }

      if (transactionSimulationService == null || blockchainService == null) {
        return transaction.getGasLimit();
      }

      final var pendingBlockHeader = transactionSimulationService.simulatePendingBlockHeader();
      final var chainId = blockchainService.getChainId().orElse(BigInteger.ZERO);
      final var tracer = createLineCountingTracer(pendingBlockHeader, chainId);
      final var maybeSimulationResults =
          transactionSimulationService.simulate(
              transaction, java.util.Optional.empty(), pendingBlockHeader, tracer, false, true);

      if (maybeSimulationResults.isPresent()) {
        final var sim = maybeSimulationResults.get();
        if (sim.isSuccessful()) {
          return sim.result().getEstimateGasUsedByTransaction();
        }
      }
    } catch (final Exception ignored) {
      // fall through to fallback below
    }
    return transaction.getGasLimit();
  }

  /**
   * Closes the gRPC channel and cleans up resources.
   *
   * @throws IOException if there are issues during resource cleanup
   */
  @Override
  public void close() throws IOException {
    if (channel != null && !channel.isShutdown()) {
      LOG.info("Shutting down RLN Prover Forwarder gRPC channel...");
      channel.shutdown();
      try {
        if (!channel.awaitTermination(5, TimeUnit.SECONDS)) {
          channel.shutdownNow();
        }
        LOG.info("RLN Prover Forwarder gRPC channel shutdown complete.");
      } catch (InterruptedException e) {
        channel.shutdownNow();
        Thread.currentThread().interrupt();
        LOG.warn("Interrupted while shutting down gRPC channel", e);
      }
    }
  }

  // Statistics methods for monitoring and testing

  /**
   * Get the total number of validation calls.
   *
   * @return Total validation call count
   */
  public int getValidationCallCount() {
    return validationCallCount.get();
  }

  /**
   * Get the number of local transactions processed.
   *
   * @return Local transaction count
   */
  public int getLocalTransactionCount() {
    return localTransactionCount.get();
  }

  /**
   * Get the number of peer transactions processed.
   *
   * @return Peer transaction count
   */
  public int getPeerTransactionCount() {
    return peerTransactionCount.get();
  }

  /**
   * Get the number of successful gRPC calls.
   *
   * @return gRPC success count
   */
  public int getGrpcSuccessCount() {
    return grpcSuccessCount.get();
  }

  /**
   * Get the number of failed gRPC calls.
   *
   * @return gRPC failure count
   */
  public int getGrpcFailureCount() {
    return grpcFailureCount.get();
  }

  /**
   * Get the number of transactions that bypassed validation due to karma eligibility.
   *
   * @return Karma bypass count
   */
  public int getKarmaBypassCount() {
    return karmaBypassCount.get();
  }

  /**
   * Get the gRPC service endpoint.
   *
   * @return Endpoint in format "host:port"
   */
  public String getEndpoint() {
    if (rlnConfig != null) {
      return rlnConfig.rlnProofServiceHost() + ":" + rlnConfig.rlnProofServicePort();
    }
    return "unknown";
  }

  /**
   * Check if the validator is enabled.
   *
   * @return true if enabled, false otherwise
   */
  public boolean isEnabled() {
    return enabled;
  }
}
