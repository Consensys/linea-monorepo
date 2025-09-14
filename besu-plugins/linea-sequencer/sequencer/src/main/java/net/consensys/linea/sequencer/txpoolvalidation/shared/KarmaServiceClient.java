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
package net.consensys.linea.sequencer.txpoolvalidation.shared;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import java.io.Closeable;
import java.io.IOException;
import java.time.Instant;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;
import net.vac.prover.GetUserTierInfoReply;
import net.vac.prover.GetUserTierInfoRequest;
import net.vac.prover.RlnProverGrpc;
import net.vac.prover.UserTierInfoError;
import net.vac.prover.UserTierInfoResult;
import org.hyperledger.besu.datatypes.Address;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Shared gRPC client for Karma Service operations.
 *
 * <p>This client encapsulates all karma-related gRPC communication, including:
 *
 * <ul>
 *   <li>Channel management with proper TLS configuration
 *   <li>Request timeout handling
 *   <li>Comprehensive error handling for various gRPC failure scenarios
 *   <li>Resource cleanup and connection lifecycle management
 * </ul>
 *
 * <p>Used by both RLN validation and gas estimation components to avoid code duplication in karma
 * service interactions.
 *
 * @author Status Network Development Team
 * @since 1.0
 */
public class KarmaServiceClient implements Closeable {
  private static final Logger LOG = LoggerFactory.getLogger(KarmaServiceClient.class);

  /**
   * Represents user karma information retrieved from the Karma Service.
   *
   * @param tier User's karma tier (e.g., "Basic", "Active", "Regular")
   * @param epochTxCount Number of transactions used in current epoch
   * @param dailyQuota Daily transaction quota for this tier
   * @param epochId Current epoch identifier from karma service
   * @param karmaBalance User's total karma balance
   */
  public record KarmaInfo(
      String tier, int epochTxCount, int dailyQuota, String epochId, long karmaBalance) {}

  private final String serviceName;
  private final long timeoutMs;
  private ManagedChannel channel;
  private RlnProverGrpc.RlnProverBlockingStub baseStub;

  // Circuit breaker state
  private final AtomicInteger consecutiveFailures = new AtomicInteger(0);
  private final AtomicLong lastFailureTime = new AtomicLong(0);
  private final int failureThreshold = 5;
  private final long circuitBreakerRecoveryMs = 30000; // 30 seconds

  /**
   * Creates a new Karma Service client with the specified configuration.
   *
   * @param serviceName Name for logging and identification purposes
   * @param host Karma service host
   * @param port Karma service port
   * @param useTls Whether to use TLS for the connection
   * @param timeoutMs Request timeout in milliseconds
   */
  public KarmaServiceClient(
      String serviceName, String host, int port, boolean useTls, long timeoutMs) {
    this(serviceName, host, port, useTls, timeoutMs, null);
  }

  /**
   * Creates a new Karma Service client with optional pre-configured channel.
   *
   * <p>This constructor is primarily intended for testing scenarios where mock gRPC channels need
   * to be injected.
   *
   * @param serviceName Name for logging and identification purposes
   * @param host Karma service host (ignored if providedChannel is not null)
   * @param port Karma service port (ignored if providedChannel is not null)
   * @param useTls Whether to use TLS (ignored if providedChannel is not null)
   * @param timeoutMs Request timeout in milliseconds
   * @param providedChannel Optional pre-configured channel for testing
   */
  public KarmaServiceClient(
      String serviceName,
      String host,
      int port,
      boolean useTls,
      long timeoutMs,
      ManagedChannel providedChannel) {
    this.serviceName = serviceName;
    this.timeoutMs = timeoutMs;

    if (providedChannel != null && !providedChannel.isShutdown()) {
      LOG.info("{}: Using pre-configured ManagedChannel for Karma Service client.", serviceName);
      this.channel = providedChannel;
    } else {
      LOG.info(
          "{}: Creating new ManagedChannel for Karma Service client at {}:{}",
          serviceName,
          host,
          port);
      ManagedChannelBuilder<?> channelBuilder = ManagedChannelBuilder.forAddress(host, port);

      if (useTls) {
        channelBuilder.useTransportSecurity();
      } else {
        channelBuilder.usePlaintext();
      }

      this.channel = channelBuilder.build();
    }

    this.baseStub = RlnProverGrpc.newBlockingStub(this.channel);

    LOG.info("{}: Karma Service client initialized successfully", serviceName);
  }

  /**
   * Fetches karma information for a user via gRPC Karma Service.
   *
   * <p>Retrieves current karma status including tier, quota, and usage information for the
   * specified user address. Includes proper error handling for gRPC failures and timeouts.
   *
   * @param userAddress The user address to query karma information for
   * @return Optional containing karma info if successful, empty on failure
   */
  public Optional<KarmaInfo> fetchKarmaInfo(Address userAddress) {
    if (baseStub == null) {
      LOG.warn("{}: Karma service not configured. Cannot fetch karma info.", serviceName);
      return Optional.empty();
    }

    // Circuit breaker: check if service is temporarily disabled due to failures
    if (isCircuitBreakerOpen()) {
      LOG.debug(
          "{}: Circuit breaker open, skipping karma service call for {}",
          serviceName,
          userAddress.toHexString());
      return Optional.empty();
    }

    // Convert Besu Address to protobuf Address
    net.vac.prover.Address protoAddress =
        net.vac.prover.Address.newBuilder()
            .setValue(com.google.protobuf.ByteString.copyFrom(userAddress.toArrayUnsafe()))
            .build();

    GetUserTierInfoRequest request =
        GetUserTierInfoRequest.newBuilder().setUser(protoAddress).build();

    try {
      LOG.debug(
          "{}: Fetching karma info for user {} via gRPC", serviceName, userAddress.toHexString());

      // Retry logic with exponential backoff
      GetUserTierInfoReply response = fetchKarmaInfoWithRetry(request);
      if (response == null) {
        return Optional.empty();
      }

      // Handle the oneof response structure
      if (response.hasRes()) {
        UserTierInfoResult result = response.getRes();

        // Validate response structure
        if (!validateUserTierInfoResult(result)) {
          LOG.warn(
              "{}: Invalid karma service response structure for user {}",
              serviceName,
              userAddress.toHexString());
          return Optional.empty();
        }

        // Extract tier info with additional validation
        String tierName = result.hasTier() ? result.getTier().getName() : "Unknown";
        int dailyQuota = result.hasTier() ? (int) result.getTier().getQuota() : 0;

        // Validate extracted values
        if (!isValidTierName(tierName) || dailyQuota < 0 || result.getTxCount() < 0) {
          LOG.warn(
              "{}: Invalid karma data for user {}: tier={}, quota={}, txCount={}",
              serviceName,
              userAddress.toHexString(),
              tierName,
              dailyQuota,
              result.getTxCount());
          return Optional.empty();
        }

        LOG.debug(
            "{}: Karma service response for {}: tier={}, epochTxCount={}, dailyQuota={}, epoch={}, epochSlice={}",
            serviceName,
            userAddress.toHexString(),
            tierName,
            result.getTxCount(),
            dailyQuota,
            result.getCurrentEpoch(),
            result.getCurrentEpochSlice());

        // Reset circuit breaker on successful response
        consecutiveFailures.set(0);

        return Optional.of(
            new KarmaInfo(
                tierName,
                (int) result.getTxCount(),
                dailyQuota,
                String.valueOf(
                    result.getCurrentEpoch()), // Convert epoch to string for compatibility
                0L)); // karma balance not in new schema, set to 0

      } else if (response.hasError()) {
        UserTierInfoError error = response.getError();
        LOG.warn(
            "{}: Karma service error for user {}: {}",
            serviceName,
            userAddress.toHexString(),
            error.getMessage());
        return Optional.empty();
      } else {
        LOG.warn(
            "{}: Karma service returned empty response for user {}",
            serviceName,
            userAddress.toHexString());
        return Optional.empty();
      }

    } catch (StatusRuntimeException e) {
      Status.Code code = e.getStatus().getCode();

      // Track failures for circuit breaker (except NOT_FOUND which is expected)
      if (code != Status.Code.NOT_FOUND) {
        recordFailure();
      }

      if (code == Status.Code.NOT_FOUND) {
        LOG.debug("{}: User {} not found in karma service", serviceName, userAddress.toHexString());
        return Optional.empty();
      } else if (code == Status.Code.DEADLINE_EXCEEDED) {
        LOG.warn(
            "{}: Karma service timeout for user {} - Status: {}, Description: {}, Cause: {}",
            serviceName,
            userAddress.toHexString(),
            e.getStatus().getCode(),
            e.getStatus().getDescription(),
            e.getCause());
        return Optional.empty();
      } else {
        LOG.error(
            "{}: Karma service gRPC error for user {} - Code: {}, Status: {}, Description: {}, Message: {}",
            serviceName,
            userAddress.toHexString(),
            code,
            e.getStatus(),
            e.getStatus().getDescription(),
            e.getMessage(),
            e);
        return Optional.empty();
      }
    } catch (Exception e) {
      recordFailure();
      LOG.error(
          "{}: Unexpected error calling karma service for user {}: {}",
          serviceName,
          userAddress.toHexString(),
          e.getMessage(),
          e);
      return Optional.empty();
    }
  }

  /**
   * Checks if the karma service client is available and properly configured.
   *
   * @return true if the client is ready to make requests, false otherwise
   */
  public boolean isAvailable() {
    return channel != null && !channel.isShutdown() && baseStub != null && !isCircuitBreakerOpen();
  }

  /**
   * Checks if the circuit breaker is currently open due to consecutive failures.
   *
   * @return true if circuit breaker is open and calls should be skipped
   */
  private boolean isCircuitBreakerOpen() {
    int failures = consecutiveFailures.get();
    if (failures < failureThreshold) {
      return false;
    }

    long lastFailure = lastFailureTime.get();
    long currentTime = Instant.now().toEpochMilli();

    // Check if recovery window has passed
    if (currentTime - lastFailure > circuitBreakerRecoveryMs) {
      LOG.info("{}: Circuit breaker recovery window passed, allowing retry", serviceName);
      // Reset on recovery attempt
      consecutiveFailures.set(0);
      return false;
    }

    return true;
  }

  /** Records a service failure for circuit breaker tracking. */
  private void recordFailure() {
    int failures = consecutiveFailures.incrementAndGet();
    lastFailureTime.set(Instant.now().toEpochMilli());

    if (failures == failureThreshold) {
      LOG.warn("{}: Circuit breaker opened after {} consecutive failures", serviceName, failures);
    } else if (failures > failureThreshold) {
      LOG.debug("{}: Circuit breaker remains open, failure count: {}", serviceName, failures);
    }
  }

  /**
   * Fetches karma info with retry logic and exponential backoff.
   *
   * @param request The gRPC request to retry
   * @return The response if successful, null if all retries failed
   */
  private GetUserTierInfoReply fetchKarmaInfoWithRetry(GetUserTierInfoRequest request) {
    final int maxRetries = 3;
    final long baseDelayMs = 100;

    for (int attempt = 0; attempt < maxRetries; attempt++) {
      try {
        // Create a new stub with deadline for each attempt
        RlnProverGrpc.RlnProverBlockingStub stubWithDeadline =
            baseStub.withDeadlineAfter(timeoutMs, TimeUnit.MILLISECONDS);

        GetUserTierInfoReply response = stubWithDeadline.getUserTierInfo(request);

        // Success - reset circuit breaker and return
        consecutiveFailures.set(0);
        return response;

      } catch (StatusRuntimeException e) {
        boolean shouldRetry = isRetriableError(e.getStatus().getCode());

        if (!shouldRetry || attempt == maxRetries - 1) {
          // Non-retriable error or final attempt - give up
          throw e;
        }

        // Exponential backoff for retriable errors
        long delayMs = baseDelayMs * (1L << attempt); // 100ms, 200ms, 400ms
        LOG.debug(
            "{}: Retriable error on attempt {}, retrying in {}ms: {}",
            serviceName,
            attempt + 1,
            delayMs,
            e.getStatus().getCode());

        try {
          Thread.sleep(delayMs);
        } catch (InterruptedException ie) {
          Thread.currentThread().interrupt();
          throw new StatusRuntimeException(
              Status.CANCELLED.withDescription("Interrupted during retry"));
        }
      }
    }

    return null; // Should never reach here
  }

  /**
   * Determines if a gRPC error is retriable.
   *
   * @param statusCode The gRPC status code
   * @return true if the error should be retried
   */
  private boolean isRetriableError(Status.Code statusCode) {
    return switch (statusCode) {
      case UNAVAILABLE, DEADLINE_EXCEEDED, RESOURCE_EXHAUSTED, ABORTED -> true;
      case NOT_FOUND, INVALID_ARGUMENT, PERMISSION_DENIED, UNAUTHENTICATED -> false;
      default -> false; // Conservative approach for unknown errors
    };
  }

  /**
   * Validates the structure and content of UserTierInfoResult from karma service.
   *
   * @param result The result to validate
   * @return true if the result is valid and safe to use
   */
  private boolean validateUserTierInfoResult(UserTierInfoResult result) {
    if (result == null) {
      return false;
    }

    // Check for reasonable bounds on transaction count
    if (result.getTxCount() < 0 || result.getTxCount() > 1_000_000) {
      return false;
    }

    // Validate tier information if present
    if (result.hasTier()) {
      var tier = result.getTier();
      if (tier.getQuota() < 0 || tier.getQuota() > 10_000) {
        return false;
      }
      if (tier.getName() == null || tier.getName().trim().isEmpty()) {
        return false;
      }
    }

    return true;
  }

  /**
   * Validates tier name for security and reasonable values.
   *
   * @param tierName The tier name to validate
   * @return true if the tier name is valid
   */
  private boolean isValidTierName(String tierName) {
    if (tierName == null || tierName.trim().isEmpty()) {
      return false;
    }

    // Allow only alphanumeric characters and basic punctuation
    if (!tierName.matches("^[a-zA-Z0-9_\\-\\s]+$")) {
      return false;
    }

    // Reasonable length limits
    if (tierName.length() > 50) {
      return false;
    }

    return true;
  }

  /**
   * Closes the gRPC channel and releases all resources.
   *
   * <p>This method should be called when the client is no longer needed to prevent resource leaks.
   *
   * @throws IOException if there are issues during resource cleanup
   */
  @Override
  public void close() throws IOException {
    if (channel != null && !channel.isShutdown()) {
      LOG.info("{}: Shutting down Karma Service gRPC channel", serviceName);
      channel.shutdown();
      try {
        if (!channel.awaitTermination(5, TimeUnit.SECONDS)) {
          channel.shutdownNow();
        }
      } catch (InterruptedException e) {
        channel.shutdownNow();
        Thread.currentThread().interrupt();
      }
      LOG.info("{}: Karma Service gRPC channel shut down", serviceName);
    }
  }
}
