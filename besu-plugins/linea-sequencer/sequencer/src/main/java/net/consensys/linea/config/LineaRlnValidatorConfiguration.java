/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.config;

import java.util.Optional;
import net.consensys.linea.plugins.LineaOptionsConfiguration;

/**
 * Configuration for the RLN (Rate Limiting Nullifier) Validator.
 *
 * @param rlnValidationEnabled Whether RLN validation is active.
 * @param verifyingKeyPath Path to the RLN verifying key file.
 * @param rlnProofServiceHost Hostname for the RLN Proof gRPC service.
 * @param rlnProofServicePort Port for the RLN Proof gRPC service.
 * @param rlnProofServiceUseTls Whether to use TLS for gRPC connection to proof service.
 * @param rlnProofCacheMaxSize Maximum number of proofs to keep in the in-memory cache.
 * @param rlnProofCacheExpirySeconds Time-to-live for proofs in the in-memory cache (in seconds).
 * @param rlnProofStreamRetries Max retries for establishing/re-establishing gRPC stream.
 * @param rlnProofStreamRetryIntervalMs Interval between gRPC stream retry attempts (in ms).
 * @param rlnProofLocalWaitTimeoutMs Timeout for waiting for a proof in local cache during
 *     validation (in ms).
 * @param sharedGaslessConfig Shared configuration including the deny list path and premium gas
 *     threshold.
 * @param karmaServiceHost Hostname for the Karma gRPC service.
 * @param karmaServicePort Port for the Karma gRPC service.
 * @param karmaServiceUseTls Whether to use TLS for gRPC connection to karma service.
 * @param karmaServiceTimeoutMs Timeout for karma service requests in milliseconds.
 * @param exponentialBackoffEnabled Whether to use exponential backoff for gRPC reconnections.
 * @param maxBackoffDelayMs Maximum backoff delay for gRPC reconnections in milliseconds.
 * @param defaultEpochForQuota Default epoch identifier.
 * @param rlnJniLibPath Optional explicit path to the rln_jni native library.
 */
public record LineaRlnValidatorConfiguration(
    boolean rlnValidationEnabled,
    String verifyingKeyPath,
    String rlnProofServiceHost,
    int rlnProofServicePort,
    boolean rlnProofServiceUseTls,
    long rlnProofCacheMaxSize,
    long rlnProofCacheExpirySeconds,
    int rlnProofStreamRetries,
    long rlnProofStreamRetryIntervalMs,
    long rlnProofLocalWaitTimeoutMs,
    LineaSharedGaslessConfiguration sharedGaslessConfig,
    String karmaServiceHost,
    int karmaServicePort,
    boolean karmaServiceUseTls,
    long karmaServiceTimeoutMs,
    boolean exponentialBackoffEnabled,
    long maxBackoffDelayMs,
    String defaultEpochForQuota,
    Optional<String> rlnJniLibPath)
    implements LineaOptionsConfiguration {

  public static LineaRlnValidatorConfiguration V1_DEFAULT =
      new LineaRlnValidatorConfiguration(
          false, // Disabled by default
          "/etc/linea/rln_verifying_key.bin",
          "localhost", // rlnProofServiceHost
          50051, // rlnProofServicePort
          false, // rlnProofServiceUseTls
          10000L, // rlnProofCacheMaxSize
          300L, // rlnProofCacheExpirySeconds (5 minutes)
          5, // rlnProofStreamRetries
          5000L, // rlnProofStreamRetryIntervalMs (5 seconds)
          200L, // rlnProofLocalWaitTimeoutMs (200ms)
          LineaSharedGaslessConfiguration.V1_DEFAULT,
          "localhost", // karmaServiceHost
          50053, // karmaServicePort
          false, // karmaServiceUseTls
          5000L, // karmaServiceTimeoutMs (5 seconds)
          true, // exponentialBackoffEnabled
          60000L, // maxBackoffDelayMs (60 seconds)
          "TIMESTAMP_1H", // defaultEpochForQuota
          Optional.empty() // rlnJniLibPath
          );

  // Accessor for deny list path for convenience
  public String denyListPath() {
    return sharedGaslessConfig.denyListPath();
  }

  // Accessor for premium gas price threshold in Wei for convenience (converting from GWei)
  public long premiumGasPriceThresholdWei() {
    return sharedGaslessConfig.premiumGasPriceThresholdGWei()
        * 1_000_000_000L; // Convert GWei to Wei
  }

  // Accessor for deny list refresh seconds for convenience
  public long denyListRefreshSeconds() {
    return sharedGaslessConfig.denyListRefreshSeconds();
  }

  // Accessor for deny list entry max age in minutes for convenience
  public long denyListEntryMaxAgeMinutes() {
    return sharedGaslessConfig.denyListEntryMaxAgeMinutes();
  }
}
