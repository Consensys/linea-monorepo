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
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine;

public class LineaRlnValidatorCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "RLN_VALIDATOR_CONFIG";

  // === ESSENTIAL OPTIONS (what operators actually need to configure) ===

  @CommandLine.Option(
      names = "--plugin-linea-rln-enabled",
      description = "Enable RLN validation for gasless transactions (default: ${DEFAULT-VALUE})",
      arity = "1")
  private boolean rlnValidationEnabled =
      LineaRlnValidatorConfiguration.V1_DEFAULT.rlnValidationEnabled();

  @CommandLine.Option(
      names = "--plugin-linea-rln-verifying-key",
      description = "Path to the RLN verifying key file (required when RLN is enabled)",
      arity = "1")
  private String verifyingKeyPath = LineaRlnValidatorConfiguration.V1_DEFAULT.verifyingKeyPath();

  @CommandLine.Option(
      names = "--plugin-linea-rln-proof-service",
      description = "RLN Proof service endpoint (host:port, default: ${DEFAULT-VALUE})",
      arity = "1")
  private String proofService = "localhost:50051";

  @CommandLine.Option(
      names = "--plugin-linea-rln-karma-service",
      description = "Karma service endpoint (host:port, default: ${DEFAULT-VALUE})",
      arity = "1")
  private String karmaService = "localhost:50052";

  @CommandLine.Option(
      names = "--plugin-linea-rln-deny-list-path",
      description = "Path to the gasless deny list file (default: ${DEFAULT-VALUE})",
      arity = "1")
  private String denyListPath = "/var/lib/besu/gasless-deny-list.txt";

  // === ADVANCED OPTIONS (most users won't need to change these) ===

  @CommandLine.Option(
      names = "--plugin-linea-rln-use-tls",
      description = "Use TLS for gRPC services (default: auto-detect based on ports)",
      arity = "1")
  private Optional<Boolean> useTls =
      Optional.empty(); // Auto-detect: false for :505x, true for :443/8443

  @CommandLine.Option(
      names = "--plugin-linea-rln-premium-gas-threshold-gwei",
      description = "Premium gas threshold in GWei to bypass deny list (default: ${DEFAULT-VALUE})",
      arity = "1")
  private long premiumGasThresholdGWei = 10L; // 10 GWei

  @CommandLine.Option(
      names = "--plugin-linea-rln-timeouts-ms",
      description = "Service timeout in milliseconds (default: ${DEFAULT-VALUE})",
      arity = "1")
  private long timeoutsMs = 5000L; // 5 seconds

  @CommandLine.Option(
      names = "--plugin-linea-rln-proof-wait-timeout-ms",
      description =
          "Timeout for waiting for RLN proof in cache during validation in milliseconds (default: ${DEFAULT-VALUE})",
      arity = "1")
  private long proofWaitTimeoutMs = 1000L; // 1 second (increased from 200ms)

  @CommandLine.Option(
      names = "--plugin-linea-rln-epoch-mode",
      description =
          "Epoch mode used to compute the RLN external nullifier (options: BLOCK, TIMESTAMP_1H, TEST, FIXED_FIELD_ELEMENT; default: ${DEFAULT-VALUE})",
      arity = "1")
  private String epochMode = LineaRlnValidatorConfiguration.V1_DEFAULT.defaultEpochForQuota();

  private LineaRlnValidatorCliOptions() {}

  public static LineaRlnValidatorCliOptions create() {
    return new LineaRlnValidatorCliOptions();
  }

  @Override
  public LineaRlnValidatorConfiguration toDomainObject() {
    // Parse service endpoints
    String[] proofParts = proofService.split(":");
    String proofHost = proofParts[0];
    int proofPort = Integer.parseInt(proofParts[1]);

    String[] karmaParts = karmaService.split(":");
    String karmaHost = karmaParts[0];
    int karmaPort = Integer.parseInt(karmaParts[1]);

    // Auto-detect TLS based on ports if not explicitly set
    boolean shouldUseTls =
        useTls.orElse(
            proofPort == 443 || proofPort == 8443 || karmaPort == 443 || karmaPort == 8443);

    // Create shared gasless config with simplified settings
    LineaSharedGaslessConfiguration sharedConfig =
        new LineaSharedGaslessConfiguration(
            denyListPath,
            60L, // 1 minute refresh interval (good default)
            premiumGasThresholdGWei,
            60L // 1 hour expiry (good default)
            );

    return new LineaRlnValidatorConfiguration(
        rlnValidationEnabled,
        verifyingKeyPath,
        proofHost,
        proofPort,
        shouldUseTls, // rlnProofServiceUseTls
        10000L, // rlnProofCacheMaxSize (good default)
        300L, // rlnProofCacheExpirySeconds (5 min, good default)
        5, // rlnProofStreamRetries (good default)
        5000L, // rlnProofStreamRetryIntervalMs (good default)
        proofWaitTimeoutMs, // rlnProofLocalWaitTimeoutMs (configurable via CLI)
        sharedConfig,
        karmaHost,
        karmaPort,
        shouldUseTls, // karmaServiceUseTls
        timeoutsMs, // karmaServiceTimeoutMs
        true, // exponentialBackoffEnabled (good default)
        60000L, // maxBackoffDelayMs (1 min, good default)
        epochMode, // defaultEpochForQuota (configurable via CLI)
        Optional.empty() // rlnJniLibPath (use system path)
        );
  }
}
