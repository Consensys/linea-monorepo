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

package net.consensys.linea.config;

import java.math.BigDecimal;

import com.google.common.base.MoreObjects;
import jakarta.validation.constraints.Positive;
import picocli.CommandLine;

/** The Linea profitability calculator CLI options. */
public class LineaProfitabilityCliOptions {
  public static final String FIXED_GAS_COST_WEI = "--plugin-linea-fixed-gas-cost-wei";
  public static final long DEFAULT_FIXED_GAS_COST_WEI = 0;

  public static final String VARIABLE_GAS_COST_WEI = "--plugin-linea-variable-gas-cost-wei";
  public static final long DEFAULT_VARIABLE_GAS_COST_WEI = 1_000_000_000;

  public static final String MIN_MARGIN = "--plugin-linea-min-margin";
  public static final BigDecimal DEFAULT_MIN_MARGIN = BigDecimal.ONE;

  public static final String ESTIMATE_GAS_MIN_MARGIN = "--plugin-linea-estimate-gas-min-margin";
  public static final BigDecimal DEFAULT_ESTIMATE_GAS_MIN_MARGIN = BigDecimal.ONE;

  public static final String TX_POOL_MIN_MARGIN = "--plugin-linea-tx-pool-min-margin";
  public static final BigDecimal DEFAULT_TX_POOL_MIN_MARGIN = BigDecimal.valueOf(0.5);

  public static final String TX_POOL_ENABLE_CHECK_API =
      "--plugin-linea-tx-pool-profitability-check-api-enabled";
  public static final boolean DEFAULT_TX_POOL_ENABLE_CHECK_API = true;

  public static final String TX_POOL_ENABLE_CHECK_P2P =
      "--plugin-linea-tx-pool-profitability-check-p2p-enabled";
  public static final boolean DEFAULT_TX_POOL_ENABLE_CHECK_P2P = false;

  public static final String EXTRA_DATA_PRICING_ENABLED =
      "--plugin-linea-extra-data-pricing-enabled";
  public static final boolean DEFAULT_EXTRA_DATA_PRICING_ENABLED = false;

  @Positive
  @CommandLine.Option(
      names = {FIXED_GAS_COST_WEI},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "Fixed gas cost in Wei (default: ${DEFAULT-VALUE})")
  private long fixedGasCostWei = DEFAULT_FIXED_GAS_COST_WEI;

  @Positive
  @CommandLine.Option(
      names = {VARIABLE_GAS_COST_WEI},
      hidden = true,
      paramLabel = "<INTEGER>",
      description = "Variable gas cost in Wei (default: ${DEFAULT-VALUE})")
  private long variableGasCostWei = DEFAULT_VARIABLE_GAS_COST_WEI;

  @Positive
  @CommandLine.Option(
      names = {MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description = "Minimum margin of a transaction to be selected (default: ${DEFAULT-VALUE})")
  private BigDecimal minMargin = DEFAULT_MIN_MARGIN;

  @Positive
  @CommandLine.Option(
      names = {ESTIMATE_GAS_MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description =
          "Recommend a specific gas price when using linea_estimateGas (default: ${DEFAULT-VALUE})")
  private BigDecimal estimateGasMinMargin = DEFAULT_ESTIMATE_GAS_MIN_MARGIN;

  @Positive
  @CommandLine.Option(
      names = {TX_POOL_MIN_MARGIN},
      hidden = true,
      paramLabel = "<FLOAT>",
      description =
          "The min margin an incoming tx must have to be accepted in the txpool (default: ${DEFAULT-VALUE})")
  private BigDecimal txPoolMinMargin = DEFAULT_TX_POOL_MIN_MARGIN;

  @CommandLine.Option(
      names = {TX_POOL_ENABLE_CHECK_API},
      arity = "0..1",
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description =
          "Enable the profitability check for txs received via API? (default: ${DEFAULT-VALUE})")
  private boolean txPoolCheckApiEnabled = DEFAULT_TX_POOL_ENABLE_CHECK_API;

  @CommandLine.Option(
      names = {TX_POOL_ENABLE_CHECK_P2P},
      arity = "0..1",
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description =
          "Enable the profitability check for txs received via p2p? (default: ${DEFAULT-VALUE})")
  private boolean txPoolCheckP2pEnabled = DEFAULT_TX_POOL_ENABLE_CHECK_P2P;

  @CommandLine.Option(
      names = {EXTRA_DATA_PRICING_ENABLED},
      arity = "0..1",
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description =
          "Enable setting pricing parameters via extra data field (default: ${DEFAULT-VALUE})")
  private boolean extraDataPricingEnabled = DEFAULT_EXTRA_DATA_PRICING_ENABLED;

  private LineaProfitabilityCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static LineaProfitabilityCliOptions create() {
    return new LineaProfitabilityCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  public static LineaProfitabilityCliOptions fromConfig(
      final LineaProfitabilityConfiguration config) {
    final LineaProfitabilityCliOptions options = create();
    options.fixedGasCostWei = config.fixedCostWei();
    options.variableGasCostWei = config.variableCostWei();
    options.minMargin = BigDecimal.valueOf(config.minMargin());
    options.estimateGasMinMargin = BigDecimal.valueOf(config.estimateGasMinMargin());
    options.txPoolMinMargin = BigDecimal.valueOf(config.txPoolMinMargin());
    options.txPoolCheckApiEnabled = config.txPoolCheckApiEnabled();
    options.txPoolCheckP2pEnabled = config.txPoolCheckP2pEnabled();
    options.extraDataPricingEnabled = config.extraDataPricingEnabled();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  public LineaProfitabilityConfiguration toDomainObject() {
    return LineaProfitabilityConfiguration.builder()
        .fixedCostWei(fixedGasCostWei)
        .variableCostWei(variableGasCostWei)
        .minMargin(minMargin.doubleValue())
        .estimateGasMinMargin(estimateGasMinMargin.doubleValue())
        .txPoolMinMargin(txPoolMinMargin.doubleValue())
        .txPoolCheckApiEnabled(txPoolCheckApiEnabled)
        .txPoolCheckP2pEnabled(txPoolCheckP2pEnabled)
        .extraDataPricingEnabled(extraDataPricingEnabled)
        .build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(FIXED_GAS_COST_WEI, fixedGasCostWei)
        .add(VARIABLE_GAS_COST_WEI, variableGasCostWei)
        .add(MIN_MARGIN, minMargin)
        .add(ESTIMATE_GAS_MIN_MARGIN, estimateGasMinMargin)
        .add(TX_POOL_MIN_MARGIN, txPoolMinMargin)
        .add(TX_POOL_ENABLE_CHECK_API, txPoolCheckApiEnabled)
        .add(TX_POOL_ENABLE_CHECK_P2P, txPoolCheckP2pEnabled)
        .add(EXTRA_DATA_PRICING_ENABLED, extraDataPricingEnabled)
        .toString();
  }
}
