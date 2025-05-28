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

import jakarta.validation.constraints.Positive;
import lombok.Getter;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine.Option;

@Getter
public class LivenessPluginCliOptions implements LineaCliOptions {
  public static final String CONFIG_KEY = "liveness-config";

  public static final String ENABLED_OPTION = "--plugin-linea-liveness-enabled";
  public static final boolean DEFAULT_ENABLED = false;

  public static final String MAX_BLOCK_AGE_SECONDS_OPTION =
      "--plugin-linea-liveness-max-block-age-seconds";
  public static final long DEFAULT_MAX_BLOCK_AGE_SECONDS = 60;

  public static final String CHECK_INTERVAL_SECONDS_OPTION =
      "--plugin-linea-liveness-check-interval-seconds";
  public static final long DEFAULT_CHECK_INTERVAL_SECONDS = 10;

  public static final String CONTRACT_ADDRESS_OPTION = "--plugin-linea-liveness-contract-address";
  public static final String DEFAULT_CONTRACT_ADDRESS =
      "0x0000000000000000000000000000000000000000";

  public static final String SIGNER_URL_OPTION = "--plugin-linea-liveness-signer-url";
  public static final String SIGNER_KEY_ID_OPTION = "--plugin-linea-liveness-signer-key-id";

  public static final String GAS_LIMIT_OPTION = "--plugin-linea-liveness-gas-limit";
  public static final long DEFAULT_GAS_LIMIT = 100_000;

  public static final String METRIC_CATEGORY_OPTION =
      "--plugin-linea-liveness-metric-category-enabled";
  public static final boolean DEFAULT_METRIC_CATEGORY = false;

  @Option(
      names = {ENABLED_OPTION},
      description = "Enable the liveness plugin (default: ${DEFAULT-VALUE})",
      arity = "0..1",
      defaultValue = "" + DEFAULT_ENABLED)
  private boolean enabled = DEFAULT_ENABLED;

  @Positive
  @Option(
      names = {MAX_BLOCK_AGE_SECONDS_OPTION},
      description =
          "Maximum age of the last block in seconds before reporting (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = "" + DEFAULT_MAX_BLOCK_AGE_SECONDS)
  private long maxBlockAgeSeconds = DEFAULT_MAX_BLOCK_AGE_SECONDS;

  @Positive
  @Option(
      names = {CHECK_INTERVAL_SECONDS_OPTION},
      description = "Interval in seconds between checks (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = "" + DEFAULT_CHECK_INTERVAL_SECONDS)
  private long checkIntervalSeconds = DEFAULT_CHECK_INTERVAL_SECONDS;

  @Option(
      names = {CONTRACT_ADDRESS_OPTION},
      description = "Address of the LineaSequencerUptimeFeed contract (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = DEFAULT_CONTRACT_ADDRESS)
  private String contractAddress = DEFAULT_CONTRACT_ADDRESS;

  @Option(
      names = {SIGNER_URL_OPTION},
      description = "URL of the Web3Signer service",
      arity = "1")
  private String signerUrl;

  @Option(
      names = {SIGNER_KEY_ID_OPTION},
      description = "Key ID to use with Web3Signer",
      arity = "1")
  private String signerKeyId;

  @Positive
  @Option(
      names = {GAS_LIMIT_OPTION},
      description = "Gas limit for transactions (default: ${DEFAULT-VALUE})",
      arity = "1",
      defaultValue = "" + DEFAULT_GAS_LIMIT)
  private long gasLimit = DEFAULT_GAS_LIMIT;

  @Option(
      names = {METRIC_CATEGORY_OPTION},
      description = "Enable metrics for liveness monitoring (default: ${DEFAULT-VALUE})",
      arity = "0..1",
      defaultValue = "" + DEFAULT_METRIC_CATEGORY)
  private boolean metricCategoryEnabled = DEFAULT_METRIC_CATEGORY;

  public static LivenessPluginCliOptions create() {
    return new LivenessPluginCliOptions();
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
        .checkIntervalSeconds(checkIntervalSeconds)
        .contractAddress(contractAddress)
        .signerUrl(signerUrl)
        .signerKeyId(signerKeyId)
        .gasLimit(gasLimit)
        .metricCategoryEnabled(metricCategoryEnabled)
        .build();
  }
}
