/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.config;

import com.google.common.base.MoreObjects;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Positive;
import lombok.Getter;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine.Option;

@Getter
public class LivenessPluginCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "linea-liveness";

  public static final String ENABLED = "--plugin-linea-liveness-enabled";
  public static final boolean DEFAULT_ENABLED = false;

  public static final String MAX_BLOCK_AGE_SECONDS =
      "--plugin-linea-liveness-max-block-age-seconds";
  public static final long DEFAULT_MAX_BLOCK_AGE_SECONDS = 60;

  public static final String CONTRACT_ADDRESS = "--plugin-linea-liveness-contract-address";
  public static final String DEFAULT_CONTRACT_ADDRESS =
      "0x0000000000000000000000000000000000000000";

  public static final String SIGNER_URL = "--plugin-linea-liveness-signer-url";
  public static final String SIGNER_KEY_ID = "--plugin-linea-liveness-signer-key-id";
  public static final String SIGNER_ADDRESS = "--plugin-linea-liveness-signer-address";

  public static final String GAS_LIMIT = "--plugin-linea-liveness-gas-limit";
  public static final long DEFAULT_GAS_LIMIT = 100_000;

  public static final String GAS_PRICE = "--plugin-linea-liveness-gas-price";
  public static final long DEFAULT_GAS_PRICE = 7; // base fee per gas for L2

  public static final String METRICS_ENABLED = "--plugin-linea-liveness-metrics-enabled";
  public static final boolean DEFAULT_METRICS_ENABLED = false;

  @Option(
      names = {ENABLED},
      paramLabel = "<BOOLEAN>",
      description = "Enable the liveness plugin (default: ${DEFAULT-VALUE})",
      arity = "0..1")
  private boolean enabled = DEFAULT_ENABLED;

  @Positive
  @Option(
      names = {MAX_BLOCK_AGE_SECONDS},
      description =
          "Maximum age of the last block in seconds before reporting (default: ${DEFAULT-VALUE})",
      defaultValue = "" + DEFAULT_MAX_BLOCK_AGE_SECONDS)
  private long maxBlockAgeSeconds = DEFAULT_MAX_BLOCK_AGE_SECONDS;

  @NotBlank(message = "Contract address must not be blank")
  @Pattern(
      regexp = "^0x[a-fA-F0-9]{40}$",
      message = "Contract address must be a valid Ethereum address")
  @Option(
      names = {CONTRACT_ADDRESS},
      description = "Address of the LineaSequencerUptimeFeed contract (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = DEFAULT_CONTRACT_ADDRESS)
  private String contractAddress = DEFAULT_CONTRACT_ADDRESS;

  @NotBlank(message = "Web3Signer URL must not be blank")
  @Pattern(regexp = "^https?://.*", message = "Web3Signer URL must be a valid HTTP or HTTPS URL")
  @Option(
      names = {SIGNER_URL},
      description = "URL of the Web3Signer service, in charge of signing transactions",
      arity = "1")
  private String signerUrl;

  @NotBlank(message = "Web3Signer key ID must not be blank")
  @Pattern(
      regexp = "^[a-zA-Z0-9._-]+$",
      message =
          "Web3Signer key ID must contain only alphanumeric characters, dots, underscores, and hyphens")
  @Option(
      names = {SIGNER_KEY_ID},
      description =
          "Key ID to use with Web3Signer, the public key corresponding to the private key in charge of signing transactions",
      arity = "1")
  private String signerKeyId;

  @NotBlank(message = "Web3Signer address must not be blank")
  @Pattern(
      regexp = "^0x[a-fA-F0-9]{40}$",
      message = "Web3Signer address must be a valid Ethereum address")
  @Option(
      names = {SIGNER_ADDRESS},
      description = "Ethereum address corresponding to the Web3Signer key ID",
      arity = "1")
  private String signerAddress;

  @Positive
  @Option(
      names = {GAS_LIMIT},
      description = "Gas limit for transactions (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = "" + DEFAULT_GAS_LIMIT)
  private long gasLimit = DEFAULT_GAS_LIMIT;

  @Option(
      names = {GAS_PRICE},
      description = "Gas price in Wei for transactions. (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = "" + DEFAULT_GAS_PRICE)
  private long gasPrice = DEFAULT_GAS_PRICE;

  @Option(
      names = {METRICS_ENABLED},
      description = "Enable metrics for liveness monitoring (default: ${DEFAULT-VALUE})",
      arity = "0..1",
      defaultValue = "" + DEFAULT_METRICS_ENABLED)
  private boolean metricCategoryEnabled = DEFAULT_METRICS_ENABLED;

  private LivenessPluginCliOptions() {}

  /**
   * Create Liveness Plugin cli options.
   *
   * @return the Liveness Plugin cli options
   */
  public static LivenessPluginCliOptions create() {
    return new LivenessPluginCliOptions();
  }

  /**
   * Liveness Plugin CLI options from config.
   *
   * @param config the config
   * @return the Liveness Plugin CLI options
   */
  public static LivenessPluginCliOptions fromConfig(final LivenessPluginCliOptions config) {
    final LivenessPluginCliOptions options = create();
    options.enabled = config.enabled;
    options.maxBlockAgeSeconds = config.maxBlockAgeSeconds;
    options.contractAddress = config.contractAddress;
    options.signerUrl = config.signerUrl;
    options.signerKeyId = config.signerKeyId;
    options.signerAddress = config.signerAddress;
    options.gasLimit = config.gasLimit;
    options.gasPrice = config.gasPrice;
    options.metricCategoryEnabled = config.metricCategoryEnabled;
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  @Override
  public LivenessPluginConfiguration toDomainObject() {
    return LivenessPluginConfiguration.builder()
        .enabled(enabled)
        .maxBlockAgeSeconds(maxBlockAgeSeconds)
        .contractAddress(contractAddress)
        .signerUrl(signerUrl)
        .signerKeyId(signerKeyId)
        .signerAddress(signerAddress)
        .gasLimit(gasLimit)
        .gasPrice(gasPrice)
        .metricCategoryEnabled(metricCategoryEnabled)
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(ENABLED, enabled)
        .add(MAX_BLOCK_AGE_SECONDS, maxBlockAgeSeconds)
        .add(CONTRACT_ADDRESS, contractAddress)
        .add(SIGNER_URL, signerUrl)
        .add(SIGNER_KEY_ID, signerKeyId)
        .add(SIGNER_ADDRESS, signerAddress)
        .add(GAS_LIMIT, gasLimit)
        .add(GAS_PRICE, gasPrice)
        .add(METRICS_ENABLED, metricCategoryEnabled)
        .toString();
  }
}
